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

func branchesEngine(t *testing.T) *engine.Engine {
	t.Helper()
	base := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: base, Hash: "root0000", Message: "commit inicial"},
			extract.CommitEvent{Timestamp: base.Add(time.Hour), Hash: "feat1111", Message: "feature", Parents: []string{"root0000"}},
			extract.CommitEvent{Timestamp: base.Add(2 * time.Hour), Hash: "merg2222", Message: "Merge feature", Parents: []string{"root0000", "feat1111"}},
		},
		meta: extract.RepoMeta{Name: "demo", TotalCommits: 3, FirstCommit: base, LastCommit: base.Add(2 * time.Hour)},
	}
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return eng
}

func TestBranchesFillsDimensions(t *testing.T) {
	eng := branchesEngine(t)
	view := scenes.Branches{}.Render(scenes.Frame{
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

func TestBranchesShowsGraphAndStats(t *testing.T) {
	eng := branchesEngine(t)
	view := scenes.Branches{}.Render(scenes.Frame{
		Engine: eng, Cursor: eng.Meta().LastCommit, Theme: theme.Default(), Width: 80, Height: 20,
	})

	for _, want := range []string{"árvore de commits", "3 commits", "1 merges", "merg222", "Merge feature", "◆", "●"} {
		if !strings.Contains(view, want) {
			t.Errorf("cena nao contem %q\n---\n%s", want, view)
		}
	}
}
