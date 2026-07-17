package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabCyclesScenes(t *testing.T) {
	var model tea.Model = threeCommitModel(t)
	if got := model.(Model).active; got != 0 {
		t.Fatalf("cena inicial = %d, quer 0", got)
	}

	want := []int{1, 2, 3, 0}
	for _, expected := range want {
		model, _ = press(model, "tab")
		if got := model.(Model).active; got != expected {
			t.Fatalf("apos tab: cena = %d, quer %d", got, expected)
		}
	}
}

func TestNumberKeysSelectScene(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	model, _ = press(model, "3")
	if got := model.(Model).active; got != 2 {
		t.Fatalf("apos 3: cena = %d, quer 2 (branches)", got)
	}

	model, _ = press(model, "1")
	if got := model.(Model).active; got != 0 {
		t.Fatalf("apos 1: cena = %d, quer 0 (timeline)", got)
	}
}

func TestSceneKeysSwitchBody(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	model, _ = press(model, "3")
	if !strings.Contains(model.View(), "árvore de commits") {
		t.Errorf("tecla 3 deveria mostrar a cena de branches\n%s", model.View())
	}

	model, _ = press(model, "2")
	if !strings.Contains(model.View(), "atividade") {
		t.Errorf("tecla 2 deveria mostrar o heatmap\n%s", model.View())
	}
}

func TestTabStripListsAllScenes(t *testing.T) {
	view := threeCommitModel(t).View()
	for _, name := range []string{"Timeline", "Heatmap", "Branches", "Linguagens"} {
		if !strings.Contains(view, name) {
			t.Errorf("tab-strip nao contem %q", name)
		}
	}
}
