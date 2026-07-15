package engine

import (
	"context"
	"testing"
	"time"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

type fakeSource struct {
	events []extract.Event
	meta   extract.RepoMeta
}

func (f fakeSource) Stream(ctx context.Context) (<-chan extract.Event, error) {
	out := make(chan extract.Event)
	go func() {
		defer close(out)
		for _, event := range f.events {
			select {
			case <-ctx.Done():
				return
			case out <- event:
			}
		}
	}()
	return out, nil
}

func (f fakeSource) Meta() extract.RepoMeta {
	return f.meta
}

func day(n int) time.Time {
	return time.Date(2026, 1, n, 12, 0, 0, 0, time.UTC)
}

func buildEngine(t *testing.T, src fakeSource) *Engine {
	t.Helper()
	engine, err := Build(context.Background(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return engine
}

func TestAtAccumulatesUpToCursor(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(3), Hash: "c3", LinesAdded: 3, LinesDeleted: 3},
			extract.CommitEvent{Timestamp: day(1), Hash: "c1", LinesAdded: 10, LinesDeleted: 1},
			extract.CommitEvent{Timestamp: day(2), Hash: "c2", LinesAdded: 5, LinesDeleted: 2},
		},
	}
	engine := buildEngine(t, src)

	cases := []struct {
		name               string
		cursor             time.Time
		count              int
		wantAdded, wantDel int
	}{
		{"antes do primeiro", day(1).Add(-time.Hour), 0, 0, 0},
		{"exatamente no primeiro", day(1), 1, 10, 1},
		{"entre commits", day(2).Add(6 * time.Hour), 2, 15, 3},
		{"exatamente no segundo", day(2), 2, 15, 3},
		{"no ultimo", day(3), 3, 18, 6},
		{"depois do ultimo", day(9), 3, 18, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := engine.At(tc.cursor)
			if state.CommitCount != tc.count {
				t.Errorf("CommitCount = %d, quer %d", state.CommitCount, tc.count)
			}
			if state.LinesAdded != tc.wantAdded || state.LinesDeleted != tc.wantDel {
				t.Errorf("+%d -%d, quer +%d -%d", state.LinesAdded, state.LinesDeleted, tc.wantAdded, tc.wantDel)
			}
			if !state.Cursor.Equal(tc.cursor) {
				t.Errorf("Cursor = %v, quer %v", state.Cursor, tc.cursor)
			}
		})
	}
}

func TestAtIgnoresNonCommitEvents(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(1), Hash: "c1", LinesAdded: 4, LinesDeleted: 0},
			extract.BranchEvent{Timestamp: day(1), Action: extract.BranchCreated, Name: "feature"},
		},
	}
	engine := buildEngine(t, src)

	state := engine.At(day(2))
	if state.CommitCount != 1 || state.LinesAdded != 4 {
		t.Errorf("estado = %+v, quer 1 commit e +4", state)
	}
}

func TestEmptyEngine(t *testing.T) {
	engine := buildEngine(t, fakeSource{})

	state := engine.At(day(5))
	if state.CommitCount != 0 || state.LinesAdded != 0 || state.LinesDeleted != 0 {
		t.Errorf("estado vazio esperado, obteve %+v", state)
	}
}

func TestMetaPassthrough(t *testing.T) {
	meta := extract.RepoMeta{Name: "demo", DefaultBranch: "main", TotalCommits: 1, FirstCommit: day(1), LastCommit: day(1)}
	src := fakeSource{
		events: []extract.Event{extract.CommitEvent{Timestamp: day(1), Hash: "c1"}},
		meta:   meta,
	}
	engine := buildEngine(t, src)

	if engine.Meta() != meta {
		t.Errorf("Meta = %+v, quer %+v", engine.Meta(), meta)
	}
}
