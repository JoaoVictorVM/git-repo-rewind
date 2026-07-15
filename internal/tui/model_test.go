package tui_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/tui"
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

func buildModel(t *testing.T, src fakeSource) tui.Model {
	t.Helper()
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return tui.New(eng)
}

func demoSource() fakeSource {
	day := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
	return fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day, Hash: "abc123", LinesAdded: 12, LinesDeleted: 3},
		},
		meta: extract.RepoMeta{
			Name:          "demo",
			DefaultBranch: "main",
			TotalCommits:  1,
			FirstCommit:   day,
			LastCommit:    day,
		},
	}
}

func sized(m tui.Model) tui.Model {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(tui.Model)
}

func TestViewRendersChrome(t *testing.T) {
	view := sized(buildModel(t, demoSource())).View()

	for _, want := range []string{"rewind", "demo", "main", "2026-05-10", "q sair"} {
		if !strings.Contains(view, want) {
			t.Errorf("view nao contem %q\n---\n%s", want, view)
		}
	}
}

func TestViewFillsTerminalHeight(t *testing.T) {
	view := sized(buildModel(t, demoSource())).View()
	if got := strings.Count(view, "\n") + 1; got != 24 {
		t.Errorf("view tem %d linhas, quer 24", got)
	}
}

func TestEmptyRepoView(t *testing.T) {
	view := sized(buildModel(t, fakeSource{meta: extract.RepoMeta{Name: "vazio"}})).View()
	if !strings.Contains(view, "sem commits") {
		t.Errorf("esperava aviso de repo vazio\n---\n%s", view)
	}
}

func TestQuitKeys(t *testing.T) {
	m := sized(buildModel(t, demoSource()))
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyCtrlC},
	} {
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatalf("tecla %v deveria retornar comando", key)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("tecla %v: esperava tea.QuitMsg, obteve %T", key, cmd())
		}
	}
}
