package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStatsCardToggle(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	if strings.Contains(model.View(), "dia mais produtivo") {
		t.Fatal("timeline nao deveria mostrar o resumo")
	}

	model, _ = press(model, "s")
	if !model.(Model).showStats {
		t.Fatal("s deveria ligar o resumo")
	}
	if !strings.Contains(model.View(), "dia mais produtivo") {
		t.Errorf("resumo deveria estar visivel\n%s", model.View())
	}

	model, _ = press(model, "s")
	if model.(Model).showStats {
		t.Fatal("s deveria desligar o resumo")
	}
	if strings.Contains(model.View(), "dia mais produtivo") {
		t.Errorf("resumo deveria sumir\n%s", model.View())
	}
}
