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

func TestSeriesBucketsDeltas(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(1), Hash: "c1", LinesAdded: 10, LinesDeleted: 1},
			extract.CommitEvent{Timestamp: day(2), Hash: "c2", LinesAdded: 5, LinesDeleted: 2},
			extract.CommitEvent{Timestamp: day(3), Hash: "c3", LinesAdded: 3, LinesDeleted: 3},
		},
	}
	engine := buildEngine(t, src)

	got := engine.Series(day(1), day(3), 3)
	want := []Bucket{
		{Added: 10, Deleted: 1, Commits: 1},
		{Added: 5, Deleted: 2, Commits: 1},
		{Added: 3, Deleted: 3, Commits: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("Series retornou %d buckets, quer %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("bucket %d = %+v, quer %+v", i, got[i], want[i])
		}
	}
}

func TestSeriesIncludesBoundaryCommits(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(1), Hash: "c1", LinesAdded: 4},
			extract.CommitEvent{Timestamp: day(2), Hash: "c2", LinesAdded: 6},
		},
	}
	engine := buildEngine(t, src)

	got := engine.Series(day(1), day(2), 1)
	if len(got) != 1 || got[0].Added != 10 || got[0].Commits != 2 {
		t.Errorf("bucket unico = %+v, quer {Added:10 Commits:2}", got[0])
	}
}

func TestLogReturnsRecentFirst(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(1), Hash: "c1"},
			extract.CommitEvent{Timestamp: day(2), Hash: "c2"},
			extract.CommitEvent{Timestamp: day(3), Hash: "c3"},
		},
	}
	engine := buildEngine(t, src)

	got := engine.Log(day(3), 2)
	if len(got) != 2 || got[0].Hash != "c3" || got[1].Hash != "c2" {
		t.Errorf("Log(day3, 2) = %v, quer [c3 c2]", hashes(got))
	}
	if before := engine.Log(day(1).Add(-time.Hour), 5); len(before) != 0 {
		t.Errorf("Log antes do primeiro deveria ser vazio, obteve %v", hashes(before))
	}
	if none := engine.Log(day(3), 0); none != nil {
		t.Errorf("Log com limite 0 deveria ser nil, obteve %v", hashes(none))
	}
}

func hashes(commits []extract.CommitEvent) []string {
	out := make([]string, len(commits))
	for i, c := range commits {
		out[i] = c.Hash
	}
	return out
}

func TestNextAndPrevStepByCommit(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(1), Hash: "c1"},
			extract.CommitEvent{Timestamp: day(2), Hash: "c2"},
			extract.CommitEvent{Timestamp: day(3), Hash: "c3"},
		},
	}
	engine := buildEngine(t, src)

	if got := engine.Next(day(1)); !got.Equal(day(2)) {
		t.Errorf("Next(day1) = %v, quer day2", got)
	}
	if got := engine.Next(day(2).Add(6 * time.Hour)); !got.Equal(day(3)) {
		t.Errorf("Next entre commits = %v, quer day3", got)
	}
	if got := engine.Next(day(3)); !got.Equal(day(3)) {
		t.Errorf("Next no ultimo deveria ficar parado, obteve %v", got)
	}
	if got := engine.Prev(day(3)); !got.Equal(day(2)) {
		t.Errorf("Prev(day3) = %v, quer day2", got)
	}
	if got := engine.Prev(day(1)); !got.Equal(day(1)) {
		t.Errorf("Prev no primeiro deveria ficar parado, obteve %v", got)
	}
}

func TestGranularityCoarserFiner(t *testing.T) {
	if got := ByCommit.Coarser(); got != ByDay {
		t.Errorf("Coarser(commit) = %v, quer dia", got.Label())
	}
	if got := ByDay.Coarser(); got != ByWeek {
		t.Errorf("Coarser(dia) = %v, quer semana", got.Label())
	}
	if got := ByWeek.Coarser(); got != ByWeek {
		t.Errorf("Coarser(semana) deveria travar em semana, obteve %v", got.Label())
	}
	if got := ByWeek.Finer(); got != ByDay {
		t.Errorf("Finer(semana) = %v, quer dia", got.Label())
	}
	if got := ByCommit.Finer(); got != ByCommit {
		t.Errorf("Finer(commit) deveria travar em commit, obteve %v", got.Label())
	}
}

func TestStepByGranularity(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(1), Hash: "c1"},
			extract.CommitEvent{Timestamp: day(10), Hash: "c2"},
			extract.CommitEvent{Timestamp: day(20), Hash: "c3"},
		},
	}
	engine := buildEngine(t, src)

	if got := engine.Step(day(1), ByCommit, true); !got.Equal(day(10)) {
		t.Errorf("passo commit para frente = %v, quer day10", got)
	}
	if got := engine.Step(day(10), ByDay, true); !got.Equal(day(11)) {
		t.Errorf("passo dia para frente = %v, quer day11", got)
	}
	if got := engine.Step(day(10), ByWeek, false); !got.Equal(day(3)) {
		t.Errorf("passo semana para tras = %v, quer day3", got)
	}
	if got := engine.Step(day(20), ByWeek, true); !got.Equal(day(20)) {
		t.Errorf("passo alem do fim deveria clampar em day20, obteve %v", got)
	}
	if got := engine.Step(day(1), ByDay, false); !got.Equal(day(1)) {
		t.Errorf("passo antes do inicio deveria clampar em day1, obteve %v", got)
	}
}

func TestNextPrevEmptyEngine(t *testing.T) {
	engine := buildEngine(t, fakeSource{})
	cursor := day(5)
	if got := engine.Next(cursor); !got.Equal(cursor) {
		t.Errorf("Next em motor vazio mudou o cursor: %v", got)
	}
	if got := engine.Prev(cursor); !got.Equal(cursor) {
		t.Errorf("Prev em motor vazio mudou o cursor: %v", got)
	}
}

func TestHeatmapAccumulatesByWeekdayHour(t *testing.T) {
	c1 := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	c2 := time.Date(2026, 1, 6, 14, 0, 0, 0, time.UTC)
	c3 := time.Date(2026, 1, 5, 9, 30, 0, 0, time.UTC)
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: c1, Hash: "c1"},
			extract.CommitEvent{Timestamp: c2, Hash: "c2"},
			extract.CommitEvent{Timestamp: c3, Hash: "c3"},
		},
	}
	engine := buildEngine(t, src)

	upToC3 := engine.Heatmap(c3)
	if got := upToC3[int(c1.Weekday())][9]; got != 2 {
		t.Errorf("celula (dia de c1, 9h) ate c3 = %d, quer 2", got)
	}
	if got := upToC3[int(c2.Weekday())][14]; got != 0 {
		t.Errorf("celula de c2 nao deveria acender antes do cursor, obteve %d", got)
	}

	full := engine.Heatmap(c2)
	if got := full[int(c2.Weekday())][14]; got != 1 {
		t.Errorf("celula de c2 ate o fim = %d, quer 1", got)
	}
	if got := full[int(c1.Weekday())][9]; got != 2 {
		t.Errorf("celula de c1/c3 ate o fim = %d, quer 2", got)
	}
}

func TestLanguagesAggregatesNetLines(t *testing.T) {
	src := fakeSource{
		events: []extract.Event{
			extract.CommitEvent{Timestamp: day(1), Hash: "c1", Files: []extract.FileChange{
				{Path: "a.go", Language: "Go", LinesAdded: 100, LinesDeleted: 0},
				{Path: "r.md", Language: "Markdown", LinesAdded: 20, LinesDeleted: 0},
			}},
			extract.CommitEvent{Timestamp: day(2), Hash: "c2", Files: []extract.FileChange{
				{Path: "a.go", Language: "Go", LinesAdded: 10, LinesDeleted: 5},
				{Path: "b.py", Language: "Python", LinesAdded: 3, LinesDeleted: 3},
				{Path: "bin", Language: "", LinesAdded: 999, LinesDeleted: 0},
			}},
		},
	}
	engine := buildEngine(t, src)

	atC1 := engine.Languages(day(1))
	if len(atC1) != 2 || atC1[0].Name != "Go" || atC1[0].Lines != 100 {
		t.Fatalf("no c1 esperava Go liderando com 100, obteve %+v", atC1)
	}

	full := engine.Languages(day(2))
	if len(full) != 2 {
		t.Fatalf("esperava 2 linguagens (Python liquido 0 e sem-linguagem sao descartados), obteve %+v", full)
	}
	if full[0].Name != "Go" || full[0].Lines != 105 {
		t.Errorf("Go liquido = %+v, quer 105", full[0])
	}
	if full[1].Name != "Markdown" || full[1].Lines != 20 {
		t.Errorf("segunda linguagem = %+v, quer Markdown 20", full[1])
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
