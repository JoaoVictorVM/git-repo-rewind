package extract

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func writeAndCommit(t *testing.T, wt *git.Worktree, dir, name, content, message string, when time.Time) plumbing.Hash {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("escrever %s: %v", name, err)
	}
	if _, err := wt.Add(name); err != nil {
		t.Fatalf("add %s: %v", name, err)
	}
	sig := &object.Signature{Name: "Ana Dev", Email: "ana@example.com", When: when}
	hash, err := wt.Commit(message, &git.CommitOptions{Author: sig, Committer: sig})
	if err != nil {
		t.Fatalf("commit %q: %v", message, err)
	}
	return hash
}

func drainEvents(t *testing.T, extractor *GitExtractor) []CommitEvent {
	t.Helper()
	stream, err := extractor.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	var commits []CommitEvent
	for event := range stream {
		commit, ok := event.(CommitEvent)
		if !ok {
			t.Fatalf("evento inesperado no stream: %T", event)
		}
		commits = append(commits, commit)
	}
	return commits
}

func TestGitExtractorEmitsCommitsChronologically(t *testing.T) {
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}

	first := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	second := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	h1 := writeAndCommit(t, wt, dir, "a.txt", "l1\nl2\nl3\n", "primeiro commit", first)
	writeAndCommit(t, wt, dir, "a.txt", "l1\nl2\nl3\nl4\nl5\n", "segundo commit", second)

	extractor, err := NewGitExtractor(dir)
	if err != nil {
		t.Fatalf("NewGitExtractor: %v", err)
	}

	commits := drainEvents(t, extractor)
	if len(commits) != 2 {
		t.Fatalf("esperava 2 commits, obteve %d", len(commits))
	}

	c1, c2 := commits[0], commits[1]
	if !c1.Timestamp.Equal(first) || !c2.Timestamp.Equal(second) {
		t.Fatalf("ordem cronologica quebrada: %v depois %v", c1.Timestamp, c2.Timestamp)
	}
	if c1.LinesAdded != 3 || c1.LinesDeleted != 0 {
		t.Errorf("commit 1: +%d -%d, quer +3 -0", c1.LinesAdded, c1.LinesDeleted)
	}
	if c2.LinesAdded != 2 || c2.LinesDeleted != 0 {
		t.Errorf("commit 2: +%d -%d, quer +2 -0", c2.LinesAdded, c2.LinesDeleted)
	}
	if len(c1.Files) != 1 || c1.Files[0] != "a.txt" {
		t.Errorf("commit 1 files = %v, quer [a.txt]", c1.Files)
	}
	if c1.Author != "Ana Dev" || c1.Email != "ana@example.com" {
		t.Errorf("autoria = %q <%q>", c1.Author, c1.Email)
	}
	if c1.Message != "primeiro commit" {
		t.Errorf("mensagem = %q", c1.Message)
	}
	if len(c1.Parents) != 0 {
		t.Errorf("commit raiz nao deveria ter pais: %v", c1.Parents)
	}
	if len(c2.Parents) != 1 || c2.Parents[0] != h1.String() {
		t.Errorf("commit 2 parents = %v, quer [%s]", c2.Parents, h1)
	}
}

func TestGitExtractorMeta(t *testing.T) {
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}

	first := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)
	last := time.Date(2026, 3, 5, 20, 0, 0, 0, time.UTC)
	writeAndCommit(t, wt, dir, "a.txt", "x\n", "c1", first)
	writeAndCommit(t, wt, dir, "b.txt", "y\n", "c2", last)

	extractor, err := NewGitExtractor(dir)
	if err != nil {
		t.Fatalf("NewGitExtractor: %v", err)
	}

	meta := extractor.Meta()
	if meta.TotalCommits != 2 {
		t.Errorf("TotalCommits = %d, quer 2", meta.TotalCommits)
	}
	if !meta.FirstCommit.Equal(first) || !meta.LastCommit.Equal(last) {
		t.Errorf("range = [%v, %v], quer [%v, %v]", meta.FirstCommit, meta.LastCommit, first, last)
	}
	if meta.Name != filepath.Base(dir) {
		t.Errorf("Name = %q, quer %q", meta.Name, filepath.Base(dir))
	}
	if meta.DefaultBranch == "" {
		t.Errorf("DefaultBranch vazio para repo com commits")
	}
}

func TestGitExtractorEmptyRepo(t *testing.T) {
	dir := t.TempDir()
	if _, err := git.PlainInit(dir, false); err != nil {
		t.Fatalf("PlainInit: %v", err)
	}

	extractor, err := NewGitExtractor(dir)
	if err != nil {
		t.Fatalf("NewGitExtractor em repo vazio: %v", err)
	}

	if got := extractor.Meta().TotalCommits; got != 0 {
		t.Errorf("TotalCommits = %d, quer 0", got)
	}
	if commits := drainEvents(t, extractor); len(commits) != 0 {
		t.Errorf("esperava stream vazio, obteve %d commits", len(commits))
	}
}
