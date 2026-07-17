package stats

import (
	"testing"
	"time"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

func commit(day int, email string, files ...string) extract.CommitEvent {
	changes := make([]extract.FileChange, len(files))
	for i, path := range files {
		changes[i] = extract.FileChange{Path: path, LinesAdded: 1}
	}
	return extract.CommitEvent{
		Timestamp: time.Date(2026, 1, day, 10, 0, 0, 0, time.UTC),
		Hash:      email + string(rune('0'+day)),
		Email:     email,
		Files:     changes,
	}
}

func TestComputeSummary(t *testing.T) {
	commits := []extract.CommitEvent{
		commit(1, "ana@x", "a.go"),
		commit(2, "ana@x", "a.go", "b.go"),
		commit(2, "bruno@x", "a.go"),
		commit(3, "ana@x", "c.go"),
		commit(5, "ana@x", "a.go"),
	}

	summary := Compute(commits)

	if summary.TotalCommits != 5 {
		t.Errorf("TotalCommits = %d, quer 5", summary.TotalCommits)
	}
	if summary.Authors != 2 {
		t.Errorf("Authors = %d, quer 2", summary.Authors)
	}
	if summary.BusiestDay.Date != "2026-01-02" || summary.BusiestDay.Commits != 2 {
		t.Errorf("BusiestDay = %+v, quer 2026-01-02 com 2", summary.BusiestDay)
	}
	if summary.LongestStreak != 3 {
		t.Errorf("LongestStreak = %d, quer 3 (dias 1,2,3)", summary.LongestStreak)
	}
	if summary.TopFile.Path != "a.go" || summary.TopFile.Commits != 4 {
		t.Errorf("TopFile = %+v, quer a.go com 4", summary.TopFile)
	}
}

func TestComputeEmpty(t *testing.T) {
	summary := Compute(nil)
	if summary.TotalCommits != 0 || summary.Authors != 0 || summary.LongestStreak != 0 {
		t.Errorf("resumo vazio esperado, obteve %+v", summary)
	}
}
