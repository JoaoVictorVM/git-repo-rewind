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
	"github.com/JoaoVictorVM/git-repo-rewind/internal/theme"
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
	theme       theme.Theme
	themes      []theme.Theme
	themeIndex  int
	sceneList   []scenes.Scene
	active      int
	addedAnim   counterAnim
	deletedAnim counterAnim
	animating   bool
	playing     bool
	playGen     int
	overview    bool
	showStats   bool
	showHelp    bool
}

func New(eng *engine.Engine) Model {
	meta := eng.Meta()
	return Model{
		engine: eng,
		meta:   meta,
		cursor: meta.LastCommit,
		theme:  theme.Presets()[0],
		themes: theme.Presets(),
		sceneList: []scenes.Scene{
			scenes.Timeline{},
			scenes.Heatmap{},
			scenes.Branches{},
			scenes.Languages{},
		},
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
		case "o":
			m.overview = !m.overview
			return m, nil
		case "s":
			m.showStats = !m.showStats
			return m, nil
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "T", "shift+t":
			m.themeIndex = (m.themeIndex + 1) % len(m.themes)
			m.theme = m.themes[m.themeIndex]
			return m, nil
		case "tab":
			m.active = (m.active + 1) % len(m.sceneList)
			return m, nil
		case "1", "2", "3", "4":
			if idx := int(msg.String()[0] - '1'); idx < len(m.sceneList) {
				m.active = idx
			}
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

	title := lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).Render("rewind · " + repoLabel(m.meta))
	info := fmt.Sprintf("branch %s · cursor %s · passo %s", branch, date, m.granularity.Label())
	if m.overview {
		info += " · " + lipgloss.NewStyle().Foreground(m.theme.Accent).Render("grid")
	}
	if m.playing {
		info += " · " + lipgloss.NewStyle().Foreground(m.theme.Accent).Render("▶")
	}
	bar := spread(title, info, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, bar, m.renderTabs(), m.styledRule())
}

func (m Model) renderTabs() string {
	labels := make([]string, len(m.sceneList))
	for i, scene := range m.sceneList {
		style := lipgloss.NewStyle().Foreground(m.theme.Muted)
		if i == m.active {
			style = lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true)
		}
		labels[i] = style.Render(fmt.Sprintf("%d %s", i+1, scene.Title()))
	}
	return strings.Join(labels, "   ")
}

func (m Model) styledRule() string {
	return lipgloss.NewStyle().Foreground(m.theme.Muted).Render(rule(m.width))
}

func (m Model) renderBody(height int) string {
	if height < 1 {
		height = 1
	}
	if m.showHelp {
		return m.renderHelp(height)
	}
	if m.meta.TotalCommits == 0 {
		return lipgloss.NewStyle().
			Foreground(m.theme.Muted).
			Width(m.width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("repositorio sem commits — nada para explorar ainda")
	}
	if m.showStats {
		return scenes.StatsCard{}.Render(m.frame(m.width, height))
	}
	if m.overview {
		return m.renderOverview(height)
	}
	return m.sceneList[m.active].Render(m.frame(m.width, height))
}

func (m Model) frame(width, height int) scenes.Frame {
	return scenes.Frame{
		Engine: m.engine,
		Cursor: m.cursor,
		Counters: scenes.Counters{
			Added:   m.addedAnim.value(),
			Deleted: m.deletedAnim.value(),
			Commits: m.engine.At(m.cursor).CommitCount,
		},
		Theme:  m.theme,
		Width:  width,
		Height: height,
	}
}

func (m Model) renderOverview(height int) string {
	leftW := (m.width - 1) / 2
	rightW := m.width - 1 - leftW
	topH := (height - 1) / 2
	bottomH := height - 1 - topH

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderCell(m.sceneList[0], leftW, topH),
		m.verticalDivider(topH),
		m.renderCell(m.sceneList[1], rightW, topH),
	)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderCell(m.sceneList[2], leftW, bottomH),
		m.verticalDivider(bottomH),
		m.renderCell(m.sceneList[3], rightW, bottomH),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, m.styledRule(), bottom)
}

func (m Model) renderCell(scene scenes.Scene, width, height int) string {
	title := lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).Render(scene.Title())
	body := scene.Render(m.frame(width, height-1))
	cell := lipgloss.JoinVertical(lipgloss.Left, title, body)
	return lipgloss.NewStyle().Width(width).Height(height).MaxHeight(height).Render(cell)
}

func (m Model) renderHelp(height int) string {
	key := lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true)
	desc := lipgloss.NewStyle().Foreground(m.theme.Muted)
	row := func(keys, action string) string {
		return key.Render(fmt.Sprintf("%-12s", keys)) + desc.Render(action)
	}

	lines := []string{
		key.Render("atalhos"),
		"",
		row("h/l  ←/→", "mover o cursor"),
		row("g / G", "início / fim da história"),
		row("+ / -", "granularidade (commit/dia/semana)"),
		row("space", "play/pause do autoplay"),
		row("Tab 1–4", "trocar de cena"),
		row("o", "modo overview (grid 2×2)"),
		row("s", "card de resumo"),
		row("Shift+T", "trocar de tema"),
		row("?", "mostrar/esconder esta ajuda"),
		row("q  Ctrl+C", "sair"),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Padding(0, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))

	return lipgloss.NewStyle().
		Width(m.width).
		Height(height).
		MaxHeight(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(box)
}

func (m Model) verticalDivider(height int) string {
	style := lipgloss.NewStyle().Foreground(m.theme.Muted)
	lines := make([]string, height)
	for i := range lines {
		lines[i] = style.Render("│")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderFooter() string {
	play := "play"
	if m.playing {
		play = "pausar"
	}
	muted := lipgloss.NewStyle().Foreground(m.theme.Muted)
	hints := muted.Render(
		fmt.Sprintf("space %s · h/l mover · ? ajuda · q sair", play))
	summary := muted.Render(fmt.Sprintf("%d commits · %s", m.meta.TotalCommits, rangeLabel(m.meta)))
	return lipgloss.JoinVertical(lipgloss.Left,
		m.styledRule(),
		m.renderMinimap(m.width),
		spread(hints, summary, m.width),
	)
}

func (m Model) renderMinimap(width int) string {
	if width < 1 {
		return ""
	}
	muted := lipgloss.NewStyle().Foreground(m.theme.Muted)
	if m.meta.TotalCommits == 0 {
		return muted.Render(strings.Repeat("·", width))
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
	marker := lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).Render(string(cells[col]))
	return muted.Render(string(cells[:col])) + marker + muted.Render(string(cells[col+1:]))
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
