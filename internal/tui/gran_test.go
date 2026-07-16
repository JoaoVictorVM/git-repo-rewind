package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
)

func TestGranularityKeysCycle(t *testing.T) {
	var model tea.Model = threeCommitModel(t)
	if got := model.(Model).granularity; got != engine.ByCommit {
		t.Fatalf("granularidade inicial = %v, quer commit", got.Label())
	}

	model, _ = press(model, "+")
	if got := model.(Model).granularity; got != engine.ByDay {
		t.Fatalf("apos +: %v, quer dia", got.Label())
	}
	model, _ = press(model, "+")
	model, _ = press(model, "+")
	if got := model.(Model).granularity; got != engine.ByWeek {
		t.Fatalf("apos +++: %v, quer semana (travado)", got.Label())
	}
	model, _ = press(model, "-")
	if got := model.(Model).granularity; got != engine.ByDay {
		t.Fatalf("apos -: %v, quer dia", got.Label())
	}
}

func TestDayGranularityChangesStep(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	model, _ = press(model, "+")
	model, _ = press(model, "h")
	if got := model.(Model).cursor; !got.Equal(navDay(2)) {
		t.Errorf("passo de um dia a partir de day3 = %v, quer day2", got)
	}
}
