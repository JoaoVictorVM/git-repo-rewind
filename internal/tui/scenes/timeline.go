package scenes

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

const (
	minLogRows   = 3
	maxLogRows   = 8
	minChartRows = 3
	shortHashLen = 7
	authorWidth  = 14
)

type Timeline struct{}

func (Timeline) Render(eng *engine.Engine, cursor time.Time, width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}

	counters := renderCounters(eng.At(cursor), width)
	countersRows := lipgloss.Height(counters)

	logRows := clamp((height-countersRows-1)/3, minLogRows, maxLogRows)
	chartRows := height - countersRows - logRows - 1
	if chartRows < minChartRows {
		chartRows = minChartRows
		logRows = height - countersRows - chartRows - 1
	}
	if logRows < 0 {
		logRows = 0
	}

	meta := eng.Meta()
	chart := renderChart(
		eng.Series(meta.FirstCommit, meta.LastCommit, width),
		cursorColumn(cursor, meta.FirstCommit, meta.LastCommit, width),
		width, chartRows,
	)
	log := renderLog(eng.Log(cursor, logRows), width, logRows)

	view := lipgloss.JoinVertical(lipgloss.Left, counters, chart, rule(width), log)
	return lipgloss.NewStyle().Width(width).Height(height).MaxHeight(height).Render(view)
}

func renderCounters(state engine.WorldState, width int) string {
	added := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("+%d", state.LinesAdded))
	deleted := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("−%d", state.LinesDeleted))
	label := fmt.Sprintf("%s adicionadas   %s removidas   ·   %d commits", added, deleted, state.CommitCount)
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(label)
}

func renderChart(series []engine.Bucket, cursor, width, height int) string {
	if height < 1 {
		height = 1
	}

	up := height / 2
	if up < 1 {
		up = 1
	}
	down := height - up
	peakAdd, peakDel := peaks(series)

	grid := newGrid(width, height)
	for c := 0; c < width && c < len(series); c++ {
		for k := 0; k < scaleHeight(series[c].Added, peakAdd, up); k++ {
			grid[up-1-k][c] = '█'
		}
		for k := 0; k < scaleHeight(series[c].Deleted, peakDel, down); k++ {
			grid[up+k][c] = '█'
		}
	}

	if cursor >= 0 && cursor < width {
		for r := 0; r < height; r++ {
			if grid[r][cursor] == ' ' {
				grid[r][cursor] = '│'
			}
		}
	}

	lines := make([]string, height)
	for r := range grid {
		lines[r] = string(grid[r])
	}
	return strings.Join(lines, "\n")
}

func renderLog(commits []extract.CommitEvent, width, rows int) string {
	lines := make([]string, 0, rows)
	for _, commit := range commits {
		lines = append(lines, renderLogLine(commit, width))
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return strings.Join(lines[:rows], "\n")
}

func renderLogLine(commit extract.CommitEvent, width int) string {
	hash := commit.Hash
	if len(hash) > shortHashLen {
		hash = hash[:shortHashLen]
	}
	prefix := fmt.Sprintf("%s  %-*s  ", hash, authorWidth, truncate(commit.Author, authorWidth))
	subject := truncate(firstLine(commit.Message), width-lipgloss.Width(prefix))
	styled := lipgloss.NewStyle().Faint(true).Render(hash) +
		fmt.Sprintf("  %-*s  ", authorWidth, truncate(commit.Author, authorWidth)) + subject
	return styled
}
