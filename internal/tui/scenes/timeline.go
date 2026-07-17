package scenes

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/theme"
)

const (
	minLogRows   = 3
	maxLogRows   = 8
	minChartRows = 3
	shortHashLen = 7
	authorWidth  = 14
)

type Counters struct {
	Added   int
	Deleted int
	Commits int
}

type Frame struct {
	Engine   *engine.Engine
	Cursor   time.Time
	Counters Counters
	Theme    theme.Theme
	Width    int
	Height   int
}

type Scene interface {
	Title() string
	Render(f Frame) string
}

type Timeline struct{}

func (Timeline) Title() string { return "Timeline" }

func (Timeline) Render(f Frame) string {
	width, height := f.Width, f.Height
	if width < 1 || height < 1 {
		return ""
	}
	th := f.Theme

	countersLine := renderCounters(f.Counters, th, width)
	countersRows := lipgloss.Height(countersLine)

	logRows := clamp((height-countersRows-1)/3, minLogRows, maxLogRows)
	chartRows := height - countersRows - logRows - 1
	if chartRows < minChartRows {
		chartRows = minChartRows
		logRows = height - countersRows - chartRows - 1
	}
	if logRows < 0 {
		logRows = 0
	}

	meta := f.Engine.Meta()
	chart := renderChart(
		f.Engine.Series(meta.FirstCommit, meta.LastCommit, width),
		cursorColumn(f.Cursor, meta.FirstCommit, meta.LastCommit, width),
		th, width, chartRows,
	)
	log := renderLog(f.Engine.Log(f.Cursor, logRows), th, width, logRows)
	separator := lipgloss.NewStyle().Foreground(th.Muted).Render(rule(width))

	view := lipgloss.JoinVertical(lipgloss.Left, countersLine, chart, separator, log)
	return lipgloss.NewStyle().Width(width).Height(height).MaxHeight(height).Render(view)
}

func renderCounters(counters Counters, th theme.Theme, width int) string {
	added := lipgloss.NewStyle().Foreground(th.Added).Bold(true).Render(fmt.Sprintf("+%d", counters.Added))
	deleted := lipgloss.NewStyle().Foreground(th.Deleted).Bold(true).Render(fmt.Sprintf("−%d", counters.Deleted))
	commits := lipgloss.NewStyle().Foreground(th.Muted).Render(fmt.Sprintf("%d commits", counters.Commits))
	label := fmt.Sprintf("%s adicionadas   %s removidas   ·   %s", added, deleted, commits)
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(label)
}

func renderChart(series []engine.Bucket, cursor int, th theme.Theme, width, height int) string {
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
		base := th.Added
		if r >= up {
			base = th.Deleted
		}
		lines[r] = styleChartRow(grid[r], cursor, base, th.Accent)
	}
	return strings.Join(lines, "\n")
}

func renderLog(commits []extract.CommitEvent, th theme.Theme, width, rows int) string {
	lines := make([]string, 0, rows)
	for _, commit := range commits {
		lines = append(lines, renderLogLine(commit, th, width))
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return strings.Join(lines[:rows], "\n")
}

func renderLogLine(commit extract.CommitEvent, th theme.Theme, width int) string {
	hash := commit.Hash
	if len(hash) > shortHashLen {
		hash = hash[:shortHashLen]
	}
	author := truncate(commit.Author, authorWidth)
	prefix := fmt.Sprintf("%s  %-*s  ", hash, authorWidth, author)
	subject := truncate(firstLine(commit.Message), width-lipgloss.Width(prefix))
	return lipgloss.NewStyle().Foreground(th.Muted).Render(hash) +
		fmt.Sprintf("  %-*s  ", authorWidth, author) + subject
}
