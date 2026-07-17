package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestThemeToggleCycles(t *testing.T) {
	var model tea.Model = threeCommitModel(t)
	if got := model.(Model).theme.Name; got != "default" {
		t.Fatalf("tema inicial = %q, quer default", got)
	}

	model, _ = press(model, "T")
	if got := model.(Model).theme.Name; got != "nerv" {
		t.Fatalf("apos Shift+T: tema = %q, quer nerv", got)
	}

	model, _ = press(model, "T")
	if got := model.(Model).theme.Name; got != "default" {
		t.Fatalf("apos Shift+T novamente: tema = %q, quer default (ciclo)", got)
	}
}
