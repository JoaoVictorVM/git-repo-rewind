package scenes

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
)

const statsLabelWidth = 22

type StatsCard struct{}

func (StatsCard) Render(f Frame) string {
	if f.Width < 1 || f.Height < 1 {
		return ""
	}
	th := f.Theme

	summary := f.Engine.Summary(f.Cursor)
	if summary.TotalCommits == 0 {
		return lipgloss.NewStyle().
			Foreground(th.Muted).
			Width(f.Width).
			Height(f.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("sem dados para o resumo ate aqui")
	}

	meta := f.Engine.Meta()
	label := lipgloss.NewStyle().Foreground(th.Muted)
	value := lipgloss.NewStyle().Foreground(th.Accent).Bold(true)
	row := func(name, val string) string {
		return label.Render(fmt.Sprintf("%-*s", statsLabelWidth, name)) + value.Render(val)
	}

	lines := []string{
		lipgloss.NewStyle().Foreground(th.Accent).Bold(true).Render("resumo · " + repoTitle(meta.Name)),
		"",
		row("commits", fmt.Sprintf("%d", summary.TotalCommits)),
		row("autores", fmt.Sprintf("%d", summary.Authors)),
		row("dia mais produtivo", fmt.Sprintf("%s (%d commits)", summary.BusiestDay.Date, summary.BusiestDay.Commits)),
		row("maior sequência", fmt.Sprintf("%d dias seguidos", summary.LongestStreak)),
		row("arquivo mais tocado", fmt.Sprintf("%s (%d)", truncate(summary.TopFile.Path, 28), summary.TopFile.Commits)),
		row("linguagens", languageSummary(f.Engine.Languages(f.Cursor))),
		row("período", fmt.Sprintf("%s → %s", meta.FirstCommit.Format("2006-01-02"), meta.LastCommit.Format("2006-01-02"))),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.Accent).
		Padding(0, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))

	return lipgloss.NewStyle().
		Width(f.Width).
		Height(f.Height).
		MaxHeight(f.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(box)
}

func repoTitle(name string) string {
	if name == "" {
		return "repositorio"
	}
	return name
}

func languageSummary(shares []engine.LanguageShare) string {
	if len(shares) == 0 {
		return "—"
	}
	total := 0
	for _, share := range shares {
		total += share.Lines
	}

	parts := make([]string, 0, 3)
	for i, share := range shares {
		if i >= 3 {
			break
		}
		pct := 0
		if total > 0 {
			pct = int(math.Round(float64(share.Lines) / float64(total) * 100))
		}
		parts = append(parts, fmt.Sprintf("%s %d%%", share.Name, pct))
	}
	return strings.Join(parts, " · ")
}
