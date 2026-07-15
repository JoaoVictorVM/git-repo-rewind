package scenes

import (
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
)

func newGrid(width, height int) [][]rune {
	grid := make([][]rune, height)
	for r := range grid {
		grid[r] = make([]rune, width)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}
	return grid
}

func peaks(series []engine.Bucket) (int, int) {
	add, del := 0, 0
	for _, bucket := range series {
		if bucket.Added > add {
			add = bucket.Added
		}
		if bucket.Deleted > del {
			del = bucket.Deleted
		}
	}
	return add, del
}

func scaleHeight(value, peak, rows int) int {
	if value <= 0 || peak <= 0 || rows <= 0 {
		return 0
	}
	bars := int(math.Ceil(float64(value) / float64(peak) * float64(rows)))
	if bars > rows {
		bars = rows
	}
	return bars
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

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func rule(width int) string {
	if width < 1 {
		return ""
	}
	return strings.Repeat("─", width)
}

func firstLine(text string) string {
	if idx := strings.IndexByte(text, '\n'); idx >= 0 {
		return text[:idx]
	}
	return text
}

func truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= width {
		return text
	}
	runes := []rune(text)
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}
