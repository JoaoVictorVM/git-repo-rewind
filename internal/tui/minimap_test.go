package tui

import (
	"testing"
	"time"
)

func TestSparkRamp(t *testing.T) {
	cases := []struct {
		count, peak int
		want        rune
	}{
		{0, 10, ' '},
		{1, 8, '▁'},
		{8, 8, '█'},
		{10, 10, '█'},
		{5, 10, '▄'},
	}
	for _, tc := range cases {
		if got := spark(tc.count, tc.peak); got != tc.want {
			t.Errorf("spark(%d, %d) = %q, quer %q", tc.count, tc.peak, got, tc.want)
		}
	}
}

func TestCursorColumn(t *testing.T) {
	first := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	last := first.Add(100 * time.Hour)

	cases := []struct {
		name   string
		cursor time.Time
		width  int
		want   int
	}{
		{"inicio", first, 10, 0},
		{"fim", last, 10, 9},
		{"meio", first.Add(50 * time.Hour), 11, 5},
		{"antes do range fica em zero", first.Add(-time.Hour), 10, 0},
		{"depois do range fica no fim", last.Add(time.Hour), 10, 9},
		{"span zero vai pro fim", first, 10, 9},
	}
	for _, tc := range cases {
		end := last
		if tc.name == "span zero vai pro fim" {
			end = first
		}
		if got := cursorColumn(tc.cursor, first, end, tc.width); got != tc.want {
			t.Errorf("%s: cursorColumn = %d, quer %d", tc.name, got, tc.want)
		}
	}
}
