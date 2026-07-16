package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

func TestCounterAnimConverges(t *testing.T) {
	anim := newCounterAnim()
	for i := 0; i < 600; i++ {
		anim.update(100)
	}
	if got := anim.value(); got != 100 {
		t.Errorf("valor final = %d, quer 100", got)
	}
}

func TestCounterAnimDoesNotOvershoot(t *testing.T) {
	anim := newCounterAnim()
	for i := 0; i < 600; i++ {
		anim.update(100)
		if anim.pos > 100+settleEpsilon {
			t.Fatalf("overshoot na iteracao %d: pos = %v", i, anim.pos)
		}
	}
}

func TestCounterAnimSettles(t *testing.T) {
	anim := newCounterAnim()
	settledAt := -1
	for i := 0; i < 600; i++ {
		anim.update(50)
		if anim.settled(50) {
			settledAt = i
			break
		}
	}
	if settledAt < 0 {
		t.Fatalf("animacao nao assentou em 600 passos: pos=%v vel=%v", anim.pos, anim.vel)
	}
}

type animSource struct {
	added, deleted int
	when           time.Time
}

func (s animSource) Stream(ctx context.Context) (<-chan extract.Event, error) {
	out := make(chan extract.Event, 1)
	out <- extract.CommitEvent{Timestamp: s.when, Hash: "c1", Author: "Ana", Message: "commit", LinesAdded: s.added, LinesDeleted: s.deleted}
	close(out)
	return out, nil
}

func (s animSource) Meta() extract.RepoMeta {
	return extract.RepoMeta{Name: "demo", DefaultBranch: "main", TotalCommits: 1, FirstCommit: s.when, LastCommit: s.when}
}

func TestModelAnimatesCountersToTarget(t *testing.T) {
	when := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
	eng, err := engine.Build(context.Background(), animSource{added: 137, deleted: 40, when: when})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var model tea.Model = New(eng)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if cmd := model.Init(); cmd == nil {
		t.Fatal("Init deveria iniciar a animacao")
	}

	initial := model.View()
	if strings.Contains(initial, "+137") {
		t.Errorf("contador nao deveria ja estar no alvo antes de animar\n%s", initial)
	}

	var cmd tea.Cmd
	for i := 0; i < 600; i++ {
		model, cmd = model.Update(tickMsg(time.Now()))
		if cmd == nil {
			break
		}
	}
	if cmd != nil {
		t.Fatal("animacao deveria ter assentado e parado de agendar ticks")
	}

	final := model.View()
	if !strings.Contains(final, "+137") || !strings.Contains(final, "−40") {
		t.Errorf("contadores nao chegaram ao alvo\n%s", final)
	}
}
