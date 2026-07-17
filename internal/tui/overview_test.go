package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOverviewShowsAllScenesAtOnce(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	single := model.View()
	if strings.Contains(single, "atividade") || strings.Contains(single, "árvore de commits") {
		t.Fatalf("modo cena unica nao deveria mostrar outras cenas\n%s", single)
	}

	model, _ = press(model, "o")
	grid := model.View()
	for _, want := range []string{"atividade", "árvore de commits", "adicionadas"} {
		if !strings.Contains(grid, want) {
			t.Errorf("overview deveria mostrar %q simultaneamente\n%s", want, grid)
		}
	}
}

func TestOverviewToggleReturnsToScene(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	model, _ = press(model, "o")
	if !model.(Model).overview {
		t.Fatal("o deveria ligar o overview")
	}

	model, _ = press(model, "o")
	if model.(Model).overview {
		t.Fatal("o deveria desligar o overview")
	}
	if strings.Contains(model.View(), "atividade") {
		t.Errorf("de volta na timeline nao deveria mostrar o heatmap\n%s", model.View())
	}
}

func TestOverviewFillsTerminalHeight(t *testing.T) {
	var model tea.Model = threeCommitModel(t)
	model, _ = press(model, "o")
	view := model.View()
	if got := strings.Count(view, "\n") + 1; got != 24 {
		t.Errorf("overview tem %d linhas, quer 24", got)
	}
}
