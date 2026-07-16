package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func playingModelAtStart(t *testing.T) tea.Model {
	t.Helper()
	var model tea.Model = threeCommitModel(t)
	model, _ = press(model, "g")
	return settle(model)
}

func TestSpaceTogglesAutoplay(t *testing.T) {
	model := playingModelAtStart(t)

	model, cmd := press(model, " ")
	if !model.(Model).playing {
		t.Fatal("space deveria ligar o autoplay")
	}
	if cmd == nil {
		t.Fatal("ligar autoplay deveria agendar um playTick")
	}

	model, cmd = press(model, " ")
	if model.(Model).playing {
		t.Fatal("space deveria desligar o autoplay")
	}
	if cmd != nil {
		t.Fatal("desligar autoplay nao deveria agendar tick")
	}
}

func TestAutoplayAdvancesCursor(t *testing.T) {
	model := playingModelAtStart(t)
	if got := model.(Model).cursor; !got.Equal(navDay(1)) {
		t.Fatalf("cursor deveria comecar em day1, obteve %v", got)
	}

	model, _ = press(model, " ")
	gen := model.(Model).playGen

	model, cmd := model.Update(playTickMsg(gen))
	if got := model.(Model).cursor; !got.Equal(navDay(2)) {
		t.Fatalf("apos um tick de autoplay cursor = %v, quer day2", got)
	}
	if cmd == nil {
		t.Fatal("autoplay deveria reagendar o proximo passo")
	}
}

func TestAutoplayStopsAtEnd(t *testing.T) {
	var model tea.Model = threeCommitModel(t)
	if got := model.(Model).cursor; !got.Equal(navDay(3)) {
		t.Fatalf("cursor deveria estar no fim, obteve %v", got)
	}

	model, _ = press(model, " ")
	gen := model.(Model).playGen

	model, cmd := model.Update(playTickMsg(gen))
	if model.(Model).playing {
		t.Fatal("autoplay deveria pausar ao chegar no fim")
	}
	if cmd != nil {
		t.Fatal("nao deveria reagendar tick apos o fim")
	}
	if got := model.(Model).cursor; !got.Equal(navDay(3)) {
		t.Fatalf("cursor nao deveria passar do fim, obteve %v", got)
	}
}

func TestAutoplayIgnoresStaleGeneration(t *testing.T) {
	model := playingModelAtStart(t)

	model, _ = press(model, " ")
	staleGen := model.(Model).playGen
	model, _ = press(model, " ")
	model, _ = press(model, " ")

	before := model.(Model).cursor
	model, cmd := model.Update(playTickMsg(staleGen))
	if got := model.(Model).cursor; !got.Equal(before) {
		t.Fatalf("tick de geracao antiga nao deveria mover o cursor: %v -> %v", before, got)
	}
	if cmd != nil {
		t.Fatal("tick de geracao antiga nao deveria reagendar")
	}
}
