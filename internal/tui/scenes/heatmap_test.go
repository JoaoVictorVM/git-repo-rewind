package scenes_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/theme"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/tui/scenes"
)

func heatmapEngine(t *testing.T) *engine.Engine {
	t.Helper()
	when := time.Date(2026, 1, 6, 14, 0, 0, 0, time.UTC)
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: when, Hash: "c1"},
			extract.CommitEvent{Timestamp: when.Add(time.Hour), Hash: "c2"},
		},
		meta: extract.RepoMeta{Name: "demo", TotalCommits: 2, FirstCommit: when, LastCommit: when.Add(time.Hour)},
	}
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return eng
}

func TestHeatmapFillsDimensions(t *testing.T) {
	eng := heatmapEngine(t)
	view := scenes.Heatmap{}.Render(scenes.Frame{
		Engine: eng, Cursor: eng.Meta().LastCommit, Theme: theme.Default(), Width: 80, Height: 20,
	})

	lines := strings.Split(view, "\n")
	if len(lines) != 20 {
		t.Fatalf("cena tem %d linhas, quer 20", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > 80 {
			t.Errorf("linha %d excede a largura: %d", i, w)
		}
	}
}

func TestHeatmapShowsGridChrome(t *testing.T) {
	eng := heatmapEngine(t)
	view := scenes.Heatmap{}.Render(scenes.Frame{
		Engine: eng, Cursor: eng.Meta().LastCommit, Theme: theme.Default(), Width: 80, Height: 20,
	})

	for _, want := range []string{"atividade", "Dom", "Seg", "Sab", "12", "menos", "mais"} {
		if !strings.Contains(view, want) {
			t.Errorf("cena nao contem %q\n---\n%s", want, view)
		}
	}
}
