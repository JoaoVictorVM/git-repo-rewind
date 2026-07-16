package scenes_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/tui/scenes"
)

type fakeSource struct {
	events []extract.Event
	meta   extract.RepoMeta
}

func (f fakeSource) Stream(ctx context.Context) (<-chan extract.Event, error) {
	out := make(chan extract.Event)
	go func() {
		defer close(out)
		for _, event := range f.events {
			select {
			case <-ctx.Done():
				return
			case out <- event:
			}
		}
	}()
	return out, nil
}

func (f fakeSource) Meta() extract.RepoMeta {
	return f.meta
}

func demoEngine(t *testing.T) *engine.Engine {
	t.Helper()
	first := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	second := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: first, Hash: "aaaaaaa1", Author: "Ana", Message: "primeiro", LinesAdded: 20, LinesDeleted: 2},
			extract.CommitEvent{Timestamp: second, Hash: "bbbbbbb2", Author: "Bruno", Message: "segundo\ncorpo", LinesAdded: 8, LinesDeleted: 10},
		},
		meta: extract.RepoMeta{Name: "demo", DefaultBranch: "main", TotalCommits: 2, FirstCommit: first, LastCommit: second},
	}
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return eng
}

func TestTimelineFillsDimensions(t *testing.T) {
	eng := demoEngine(t)
	view := scenes.Timeline{}.Render(eng, eng.Meta().LastCommit, scenes.Counters{Added: 28, Deleted: 12, Commits: 2}, 80, 20)

	lines := strings.Split(view, "\n")
	if len(lines) != 20 {
		t.Fatalf("cena tem %d linhas, quer 20", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > 80 {
			t.Errorf("linha %d excede largura: %d", i, w)
		}
	}
}

func TestTimelineShowsCountersAndLog(t *testing.T) {
	eng := demoEngine(t)
	view := scenes.Timeline{}.Render(eng, eng.Meta().LastCommit, scenes.Counters{Added: 28, Deleted: 12, Commits: 2}, 80, 20)

	for _, want := range []string{"+28", "removidas", "2 commits", "bbbbbbb", "Bruno", "segundo"} {
		if !strings.Contains(view, want) {
			t.Errorf("cena nao contem %q\n---\n%s", want, view)
		}
	}
	if strings.Contains(view, "corpo") {
		t.Errorf("log deveria mostrar so o assunto, nao o corpo\n%s", view)
	}
}

func TestTimelineDrawsCursorLine(t *testing.T) {
	eng := demoEngine(t)
	view := scenes.Timeline{}.Render(eng, eng.Meta().FirstCommit, scenes.Counters{Added: 20, Deleted: 2, Commits: 1}, 80, 20)
	if !strings.Contains(view, "│") {
		t.Errorf("esperava linha vertical do cursor\n%s", view)
	}
}
