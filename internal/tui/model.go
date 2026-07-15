package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

type Model struct {
	engine *engine.Engine
	meta   extract.RepoMeta
	cursor time.Time
	width  int
	height int
}

func New(eng *engine.Engine) Model {
	meta := eng.Meta()
	return Model{engine: eng, meta: meta, cursor: meta.LastCommit}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
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
	info := fmt.Sprintf("branch %s · cursor %s", branch, date)
	bar := spread(title, info, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, bar, rule(m.width))
}

func (m Model) renderBody(height int) string {
	if height < 1 {
		height = 1
	}
	content := "repositorio sem commits ainda"
	if m.meta.TotalCommits > 0 {
		state := m.engine.At(m.cursor)
		content = fmt.Sprintf("%d commits até o cursor\n+%d   −%d linhas",
			state.CommitCount, state.LinesAdded, state.LinesDeleted)
	}
	return lipgloss.NewStyle().
		Width(m.width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

func (m Model) renderFooter() string {
	hints := lipgloss.NewStyle().Faint(true).Render("q sair")
	summary := fmt.Sprintf("%d commits · %s", m.meta.TotalCommits, rangeLabel(m.meta))
	return lipgloss.JoinVertical(lipgloss.Left, rule(m.width), spread(hints, summary, m.width))
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
