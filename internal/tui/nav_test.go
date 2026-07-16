package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

type listSource struct {
	events []extract.Event
	meta   extract.RepoMeta
}

func (s listSource) Stream(ctx context.Context) (<-chan extract.Event, error) {
	out := make(chan extract.Event)
	go func() {
		defer close(out)
		for _, event := range s.events {
			select {
			case <-ctx.Done():
				return
			case out <- event:
			}
		}
	}()
	return out, nil
}

func (s listSource) Meta() extract.RepoMeta {
	return s.meta
}

func navDay(n int) time.Time {
	return time.Date(2026, 4, n, 12, 0, 0, 0, time.UTC)
}

func threeCommitModel(t *testing.T) Model {
	t.Helper()
	src := listSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: navDay(1), Hash: "c1", LinesAdded: 5},
			extract.CommitEvent{Timestamp: navDay(2), Hash: "c2", LinesAdded: 5},
			extract.CommitEvent{Timestamp: navDay(3), Hash: "c3", LinesAdded: 5},
		},
		meta: extract.RepoMeta{Name: "demo", DefaultBranch: "main", TotalCommits: 3, FirstCommit: navDay(1), LastCommit: navDay(3)},
	}
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	sized, _ := New(eng).Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return settle(sized).(Model)
}

func settle(model tea.Model) tea.Model {
	for i := 0; i < 600; i++ {
		next, cmd := model.Update(tickMsg(time.Now()))
		model = next
		if cmd == nil {
			break
		}
	}
	return model
}

func press(model tea.Model, key string) (tea.Model, tea.Cmd) {
	return model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func TestCursorStartsAtEnd(t *testing.T) {
	m := threeCommitModel(t)
	if !m.cursor.Equal(navDay(3)) {
		t.Errorf("cursor inicial = %v, quer %v", m.cursor, navDay(3))
	}
}

func TestCursorStepsBackAndForward(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	model, _ = press(model, "h")
	if got := model.(Model).cursor; !got.Equal(navDay(2)) {
		t.Fatalf("apos h: cursor = %v, quer day2", got)
	}
	model, _ = press(model, "h")
	if got := model.(Model).cursor; !got.Equal(navDay(1)) {
		t.Fatalf("apos hh: cursor = %v, quer day1", got)
	}
	model, cmd := press(model, "h")
	if got := model.(Model).cursor; !got.Equal(navDay(1)) {
		t.Fatalf("h no inicio deveria ficar parado, obteve %v", got)
	}
	if cmd != nil {
		t.Error("movimento sem mudanca nao deveria agendar tick")
	}

	model, _ = press(model, "l")
	if got := model.(Model).cursor; !got.Equal(navDay(2)) {
		t.Fatalf("apos l: cursor = %v, quer day2", got)
	}
}

func TestCursorJumpsToExtremes(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	model, cmd := press(model, "g")
	if got := model.(Model).cursor; !got.Equal(navDay(1)) {
		t.Fatalf("apos g: cursor = %v, quer day1", got)
	}
	if cmd == nil {
		t.Error("pular para o inicio deveria reiniciar a animacao")
	}

	model = settle(model)
	model, _ = press(model, "G")
	if got := model.(Model).cursor; !got.Equal(navDay(3)) {
		t.Fatalf("apos G: cursor = %v, quer day3", got)
	}
}
