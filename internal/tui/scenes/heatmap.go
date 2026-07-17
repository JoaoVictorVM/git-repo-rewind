package scenes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/theme"
)

const heatCellWidth = 2

var weekdayLabels = [7]string{"Dom", "Seg", "Ter", "Qua", "Qui", "Sex", "Sab"}

var heatRamp = []rune("░▒▓█")

type Heatmap struct{}

func (Heatmap) Title() string { return "Heatmap" }

func (Heatmap) Render(f Frame) string {
	if f.Width < 1 || f.Height < 1 {
		return ""
	}

	grid := f.Engine.Heatmap(f.Cursor)
	peak := heatPeak(grid)
	th := f.Theme

	parts := []string{
		lipgloss.NewStyle().Foreground(th.Accent).Bold(true).Render("atividade · dia da semana × hora"),
		"",
		renderHourAxis(th),
	}
	for d := 0; d < 7; d++ {
		parts = append(parts, renderHeatRow(weekdayLabels[d], grid[d], peak, th))
	}
	parts = append(parts, "", renderHeatLegend(th))

	block := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.NewStyle().
		Width(f.Width).
		Height(f.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(block)
}

func heatPeak(grid [7][24]int) int {
	peak := 0
	for d := range grid {
		for _, count := range grid[d] {
			if count > peak {
				peak = count
			}
		}
	}
	return peak
}

func renderHourAxis(th theme.Theme) string {
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", 4))
	for h := 0; h < 24; h++ {
		if h%3 == 0 {
			b.WriteString(fmt.Sprintf("%02d", h))
		} else {
			b.WriteString(strings.Repeat(" ", heatCellWidth))
		}
	}
	return lipgloss.NewStyle().Foreground(th.Muted).Render(b.String())
}

func renderHeatRow(label string, hours [24]int, peak int, th theme.Theme) string {
	gutter := lipgloss.NewStyle().Foreground(th.Muted).Render(fmt.Sprintf("%-4s", label))
	var b strings.Builder
	for h := 0; h < 24; h++ {
		b.WriteString(heatCell(hours[h], peak, th))
	}
	return gutter + b.String()
}

func heatCell(count, peak int, th theme.Theme) string {
	if count <= 0 {
		return lipgloss.NewStyle().Foreground(th.Muted).Render(strings.Repeat("·", heatCellWidth))
	}
	idx := (count*len(heatRamp) + peak - 1) / peak
	if idx < 1 {
		idx = 1
	}
	if idx > len(heatRamp) {
		idx = len(heatRamp)
	}
	return lipgloss.NewStyle().Foreground(th.Added).Render(strings.Repeat(string(heatRamp[idx-1]), heatCellWidth))
}

func renderHeatLegend(th theme.Theme) string {
	muted := lipgloss.NewStyle().Foreground(th.Muted)
	added := lipgloss.NewStyle().Foreground(th.Added)
	var ramp strings.Builder
	for _, r := range heatRamp {
		ramp.WriteString(added.Render(string(r)))
	}
	return muted.Render("menos ") + ramp.String() + muted.Render(" mais")
}
