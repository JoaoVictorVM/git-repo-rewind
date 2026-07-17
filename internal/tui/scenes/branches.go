package scenes

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/theme"
)

type Branches struct{}

func (Branches) Title() string { return "Branches" }

func (Branches) Render(f Frame) string {
	if f.Width < 1 || f.Height < 1 {
		return ""
	}
	th := f.Theme

	stats := f.Engine.DAGStats(f.Cursor)
	if stats.Commits == 0 {
		return lipgloss.NewStyle().
			Foreground(th.Muted).
			Width(f.Width).
			Height(f.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("sem historico ate aqui")
	}

	header := lipgloss.NewStyle().Foreground(th.Accent).Bold(true).Render("árvore de commits") +
		lipgloss.NewStyle().Foreground(th.Muted).Render(
			fmt.Sprintf("   %d commits · %d merges · %d ramos", stats.Commits, stats.Merges, stats.Tips))

	available := f.Height - 2
	if available < 1 {
		available = 1
	}
	visible := (available + 1) / 2
	commits := f.Engine.Log(f.Cursor, visible)

	parts := append([]string{header, ""}, renderCommitGraph(commits, th, f.Width)...)
	block := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.NewStyle().Width(f.Width).Height(f.Height).MaxHeight(f.Height).Render(block)
}

func renderCommitGraph(commits []extract.CommitEvent, th theme.Theme, width int) []string {
	nodeStyle := lipgloss.NewStyle().Foreground(th.Accent).Bold(true)
	mergeStyle := lipgloss.NewStyle().Foreground(th.Added).Bold(true)
	lineStyle := lipgloss.NewStyle().Foreground(th.Muted)

	rows := make([]string, 0, len(commits)*2)
	for i, commit := range commits {
		glyph := nodeStyle.Render("●")
		if len(commit.Parents) >= 2 {
			glyph = mergeStyle.Render("◆")
		}

		hash := commit.Hash
		if len(hash) > shortHashLen {
			hash = hash[:shortHashLen]
		}
		prefix := fmt.Sprintf("● %s  ", hash)
		subject := truncate(firstLine(commit.Message), width-lipgloss.Width(prefix))
		rows = append(rows, glyph+" "+lineStyle.Render(hash)+"  "+subject)

		if i < len(commits)-1 {
			rows = append(rows, lineStyle.Render("│"))
		}
	}
	return rows
}
