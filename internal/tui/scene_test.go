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

func TestPlaceholderSceneRenders(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	if strings.Contains(model.View(), "em construção") {
		t.Fatalf("timeline nao deveria ser um placeholder")
	}

	model, _ = press(model, "3")
	view := model.View()
	if !strings.Contains(view, "Branches") {
		t.Errorf("tab-strip deveria listar Branches\n%s", view)
	}
	if !strings.Contains(view, "em construção") {
		t.Errorf("cena Branches deveria mostrar placeholder\n%s", view)
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
