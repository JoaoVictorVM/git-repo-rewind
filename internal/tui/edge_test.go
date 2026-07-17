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

func modelFrom(t *testing.T, src listSource) tea.Model {
	t.Helper()
	eng, err := engine.Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	sized, _ := New(eng).Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return sized
}

func TestEmptyRepoShowsMessageWithoutCrashing(t *testing.T) {
	model := modelFrom(t, listSource{meta: extract.RepoMeta{Name: "vazio"}})

	if !strings.Contains(model.View(), "sem commits") {
		t.Errorf("esperava aviso de repo vazio\n%s", model.View())
	}

	for _, key := range []string{"l", "h", "g", "G", " ", "2", "o", "s", "+"} {
		model, _ = press(model, key)
	}
	if !strings.Contains(model.View(), "sem commits") {
		t.Errorf("repo vazio deveria seguir estavel apos teclas\n%s", model.View())
	}
}

func TestHelpWorksEvenOnEmptyRepo(t *testing.T) {
	model := modelFrom(t, listSource{meta: extract.RepoMeta{Name: "vazio"}})

	model, _ = press(model, "?")
	view := model.View()
	if !strings.Contains(view, "atalhos") || !strings.Contains(view, "mover o cursor") {
		t.Errorf("ajuda deveria funcionar mesmo em repo vazio\n%s", view)
	}
}

func TestHelpToggle(t *testing.T) {
	var model tea.Model = threeCommitModel(t)

	model, _ = press(model, "?")
	if !strings.Contains(model.View(), "atalhos") {
		t.Errorf("? deveria abrir a ajuda\n%s", model.View())
	}

	model, _ = press(model, "?")
	if strings.Contains(model.View(), "trocar de tema") {
		t.Errorf("? deveria fechar a ajuda\n%s", model.View())
	}
}

func TestSingleCommitRendersWithoutCrashing(t *testing.T) {
	when := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	src := listSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: when, Hash: "solo1234", Message: "unico", LinesAdded: 9},
		},
		meta: extract.RepoMeta{Name: "solo", DefaultBranch: "main", TotalCommits: 1, FirstCommit: when, LastCommit: when},
	}
	model := modelFrom(t, src)

	if got := model.(Model).cursor; !got.Equal(when) {
		t.Fatalf("cursor deveria estar no unico commit, obteve %v", got)
	}
	for _, key := range []string{"h", "l", "g", "G", "2", "3", "4", "1", "s", "o"} {
		model, _ = press(model, key)
		if strings.Count(model.View(), "\n")+1 != 24 {
			t.Fatalf("view degenerou apos %q", key)
		}
	}
}
