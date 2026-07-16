package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/tui/scenes"
)

var minimapRamp = []rune("▁▂▃▄▅▆▇█")

type Model struct {
	engine      *engine.Engine
	meta        extract.RepoMeta
	cursor      time.Time
	width       int
	height      int
	granularity engine.Granularity
	addedAnim   counterAnim
	deletedAnim counterAnim
	animating   bool
	playing     bool
	playGen     int
}

func New(eng *engine.Engine) Model {
	meta := eng.Meta()
	return Model{
		engine:      eng,
		meta:        meta,
		cursor:      meta.LastCommit,
		addedAnim:   newCounterAnim(),
		deletedAnim: newCounterAnim(),
		animating:   meta.TotalCommits > 0,
	}
}

func (m Model) Init() tea.Cmd {
	if m.animating {
		return tick()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m.advanceAnimation()
	case playTickMsg:
		return m.advancePlay(int(msg))
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ", "space":
			return m.togglePlay()
		case "l", "right":
			return m.moveCursor(m.engine.Step(m.cursor, m.granularity, true))
		case "h", "left":
			return m.moveCursor(m.engine.Step(m.cursor, m.granularity, false))
		case "g":
			return m.moveCursor(m.meta.FirstCommit)
		case "G":
			return m.moveCursor(m.meta.LastCommit)
		case "+", "=":
			m.granularity = m.granularity.Coarser()
			return m, nil
		case "-", "_":
			m.granularity = m.granularity.Finer()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) moveCursor(to time.Time) (Model, tea.Cmd) {
	if to.Equal(m.cursor) {
		return m, nil
	}
	m.cursor = to
	if !m.animating {
		m.animating = true
		return m, tick()
	}
	return m, nil
}

func (m Model) advanceAnimation() (tea.Model, tea.Cmd) {
	if !m.animating {
		return m, nil
	}

	state := m.engine.At(m.cursor)
	addTarget := float64(state.LinesAdded)
	delTarget := float64(state.LinesDeleted)
	m.addedAnim.update(addTarget)
	m.deletedAnim.update(delTarget)

	if m.addedAnim.settled(addTarget) && m.deletedAnim.settled(delTarget) {
		m.addedAnim.snap(addTarget)
		m.deletedAnim.snap(delTarget)
		m.animating = false
		return m, nil
	}
	return m, tick()
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	header := m.renderHeader()
	footer := m.renderFooter()
	body := m.renderBody(m.height - lipgloss.Height(header) - lipgloss.Height(footer))
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) renderHeader() string {
	branch := m.meta.DefaultBranch
	if branch == "" {
		branch = "—"
	}
	date := "—"
	if !m.cursor.IsZero() {
		date = m.cursor.Format("2006-01-02 15:04")
	}

	title := lipgloss.NewStyle().Bold(true).Render("rewind · " + repoLabel(m.meta))
	info := fmt.Sprintf("branch %s · cursor %s · passo %s", branch, date, m.granularity.Label())
	if m.playing {
		info += " · ▶"
	}
	bar := spread(title, info, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, bar, rule(m.width))
}

func (m Model) renderBody(height int) string {
	if height < 1 {
		height = 1
	}
	if m.meta.TotalCommits == 0 {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("repositorio sem commits ainda")
	}
	counters := scenes.Counters{
		Added:   m.addedAnim.value(),
		Deleted: m.deletedAnim.value(),
		Commits: m.engine.At(m.cursor).CommitCount,
	}
	return scenes.Timeline{}.Render(m.engine, m.cursor, counters, m.width, height)
}

func (m Model) renderFooter() string {
	play := "play"
	if m.playing {
		play = "pausar"
	}
	hints := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("space %s · h/l mover · +/- passo · g/G extremos · q sair", play))
	summary := fmt.Sprintf("%d commits · %s", m.meta.TotalCommits, rangeLabel(m.meta))
	return lipgloss.JoinVertical(lipgloss.Left,
		rule(m.width),
		m.renderMinimap(m.width),
		spread(hints, summary, m.width),
	)
}

func (m Model) renderMinimap(width int) string {
	if width < 1 {
		return ""
	}
	if m.meta.TotalCommits == 0 {
		return lipgloss.NewStyle().Faint(true).Render(strings.Repeat("·", width))
	}

	counts := m.bucketCounts(width)
	peak := 0
	for _, count := range counts {
		if count > peak {
			peak = count
		}
	}

	cells := make([]rune, width)
	for i, count := range counts {
		cells[i] = spark(count, peak)
	}

	col := cursorColumn(m.cursor, m.meta.FirstCommit, m.meta.LastCommit, width)
	marker := lipgloss.NewStyle().Reverse(true).Render(string(cells[col]))
	return string(cells[:col]) + marker + string(cells[col+1:])
}

func (m Model) bucketCounts(width int) []int {
	counts := make([]int, width)
	span := m.meta.LastCommit.Sub(m.meta.FirstCommit)
	if span <= 0 {
		counts[width-1] = m.meta.TotalCommits
		return counts
	}

	previous := 0
	for i := 0; i < width; i++ {
		frac := float64(i+1) / float64(width)
		boundary := m.meta.FirstCommit.Add(time.Duration(float64(span) * frac))
		total := m.engine.At(boundary).CommitCount
		counts[i] = total - previous
		previous = total
	}
	return counts
}

func cursorColumn(cursor, first, last time.Time, width int) int {
	if width <= 1 {
		return 0
	}
	span := last.Sub(first)
	if span <= 0 {
		return width - 1
	}
	frac := float64(cursor.Sub(first)) / float64(span)
	frac = math.Max(0, math.Min(1, frac))
	return int(math.Round(frac * float64(width-1)))
}

func spark(count, peak int) rune {
	if count <= 0 {
		return ' '
	}
	if peak <= 0 {
		return minimapRamp[0]
	}
	idx := (count*len(minimapRamp) + peak - 1) / peak
	if idx < 1 {
		idx = 1
	}
	if idx > len(minimapRamp) {
		idx = len(minimapRamp)
	}
	return minimapRamp[idx-1]
}

func repoLabel(meta extract.RepoMeta) string {
	if meta.Name == "" {
		return "repositorio"
	}
	return meta.Name
}

func rangeLabel(meta extract.RepoMeta) string {
	if meta.TotalCommits == 0 {
		return "sem historico"
	}
	return fmt.Sprintf("%s → %s",
		meta.FirstCommit.Format("2006-01-02"),
		meta.LastCommit.Format("2006-01-02"))
}

func rule(width int) string {
	if width < 1 {
		return ""
	}
	return strings.Repeat("─", width)
}

func spread(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return lipgloss.NewStyle().Width(width).Render(left)
	}
	return left + strings.Repeat(" ", gap) + right
}
