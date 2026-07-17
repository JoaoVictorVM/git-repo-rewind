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

func statsEngine(t *testing.T) *engine.Engine {
	t.Helper()
	base := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: base, Hash: "c1", Email: "ana@x", Files: []extract.FileChange{{Path: "a.go", Language: "Go", LinesAdded: 40}}},
			extract.CommitEvent{Timestamp: base.Add(24 * time.Hour), Hash: "c2", Email: "bruno@x", Files: []extract.FileChange{{Path: "a.go", Language: "Go", LinesAdded: 10}}},
		},
		meta: extract.RepoMeta{Name: "demo", TotalCommits: 2, FirstCommit: base, LastCommit: base.Add(24 * time.Hour)},
	}
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return eng
}

func TestStatsCardFillsDimensions(t *testing.T) {
	eng := statsEngine(t)
	view := scenes.StatsCard{}.Render(scenes.Frame{
		Engine: eng, Cursor: eng.Meta().LastCommit, Theme: theme.Default(), Width: 80, Height: 20,
	})

	lines := strings.Split(view, "\n")
	if len(lines) != 20 {
		t.Fatalf("card tem %d linhas, quer 20", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > 80 {
			t.Errorf("linha %d excede a largura: %d", i, w)
		}
	}
}

func TestStatsCardShowsSummary(t *testing.T) {
	eng := statsEngine(t)
	view := scenes.StatsCard{}.Render(scenes.Frame{
		Engine: eng, Cursor: eng.Meta().LastCommit, Theme: theme.Default(), Width: 80, Height: 20,
	})

	for _, want := range []string{"resumo", "commits", "autores", "dia mais produtivo", "sequência", "arquivo mais tocado", "Go", "a.go"} {
		if !strings.Contains(view, want) {
			t.Errorf("card nao contem %q\n---\n%s", want, view)
		}
	}
}
