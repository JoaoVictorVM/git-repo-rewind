package scenes

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/theme"
)

const (
	languageNameWidth = 14
	languagePctWidth  = 4
	languageMaxWidth  = 72
)

type Languages struct{}

func (Languages) Title() string { return "Linguagens" }

func (Languages) Render(f Frame) string {
	if f.Width < 1 || f.Height < 1 {
		return ""
	}
	th := f.Theme

	shares := f.Engine.Languages(f.Cursor)
	if len(shares) == 0 {
		return lipgloss.NewStyle().
			Foreground(th.Muted).
			Width(f.Width).
			Height(f.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("nenhuma linguagem detectada ate aqui")
	}

	total := 0
	for _, share := range shares {
		total += share.Lines
	}

	barWidth := contentWidth(f.Width) - languageNameWidth - languagePctWidth - 2
	if barWidth < 1 {
		barWidth = 1
	}
	maxRows := f.Height - 2
	if maxRows < 1 {
		maxRows = 1
	}

	parts := []string{lipgloss.NewStyle().Foreground(th.Accent).Bold(true).Render("linguagens · composição no cursor"), ""}
	parts = append(parts, languageRows(shares, total, maxRows, barWidth, th)...)

	block := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.NewStyle().
		Width(f.Width).
		Height(f.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(block)
}

func languageRows(shares []engine.LanguageShare, total, maxRows, barWidth int, th theme.Theme) []string {
	shown := shares
	other := 0
	if len(shares) > maxRows {
		shown = shares[:maxRows-1]
		for _, share := range shares[maxRows-1:] {
			other += share.Lines
		}
	}

	rows := make([]string, 0, maxRows)
	for _, share := range shown {
		rows = append(rows, languageRow(share.Name, share.Lines, total, barWidth, th))
	}
	if other > 0 {
		rows = append(rows, languageRow("outras", other, total, barWidth, th))
	}
	return rows
}

func languageRow(name string, lines, total, barWidth int, th theme.Theme) string {
	frac := 0.0
	if total > 0 {
		frac = float64(lines) / float64(total)
	}
	filled := int(math.Round(frac * float64(barWidth)))
	if filled > barWidth {
		filled = barWidth
	}

	label := fmt.Sprintf("%-*s", languageNameWidth, truncate(name, languageNameWidth))
	bar := lipgloss.NewStyle().Foreground(th.Accent).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(th.Muted).Render(strings.Repeat("░", barWidth-filled))
	pct := lipgloss.NewStyle().Foreground(th.Muted).Render(fmt.Sprintf("%*d%%", languagePctWidth-1, int(math.Round(frac*100))))
	return label + " " + bar + " " + pct
}

func contentWidth(width int) int {
	if width > languageMaxWidth {
		return languageMaxWidth
	}
	return width
}
