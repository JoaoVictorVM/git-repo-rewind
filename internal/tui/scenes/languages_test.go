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

func languagesEngine(t *testing.T) *engine.Engine {
	t.Helper()
	when := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: when, Hash: "c1", Files: []extract.FileChange{
				{Path: "a.go", Language: "Go", LinesAdded: 80},
				{Path: "b.py", Language: "Python", LinesAdded: 20},
			}},
		},
		meta: extract.RepoMeta{Name: "demo", TotalCommits: 1, FirstCommit: when, LastCommit: when},
	}
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return eng
}

func TestLanguagesFillsDimensions(t *testing.T) {
	eng := languagesEngine(t)
	view := scenes.Languages{}.Render(scenes.Frame{
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

func TestLanguagesShowsBars(t *testing.T) {
	eng := languagesEngine(t)
	view := scenes.Languages{}.Render(scenes.Frame{
		Engine: eng, Cursor: eng.Meta().LastCommit, Theme: theme.Default(), Width: 80, Height: 20,
	})

	for _, want := range []string{"linguagens", "Go", "Python", "80%", "20%", "█"} {
		if !strings.Contains(view, want) {
			t.Errorf("cena nao contem %q\n---\n%s", want, view)
		}
	}
}

func TestLanguagesEmptyState(t *testing.T) {
	when := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	src := fakeSource{
		events: []extract.Event{extract.CommitEvent{Timestamp: when, Hash: "c1"}},
		meta:   extract.RepoMeta{Name: "demo", TotalCommits: 1, FirstCommit: when, LastCommit: when},
	}
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	view := scenes.Languages{}.Render(scenes.Frame{
		Engine: eng, Cursor: when, Theme: theme.Default(), Width: 80, Height: 20,
	})
	if !strings.Contains(view, "nenhuma linguagem") {
		t.Errorf("esperava estado vazio\n%s", view)
	}
}
