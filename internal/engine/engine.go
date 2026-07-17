package engine

import (
	"context"
	"sort"
	"time"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/stats"
)

type LanguageShare struct {
	Name  string
	Lines int
}

type WorldState struct {
	Cursor       time.Time
	LinesAdded   int
	LinesDeleted int
	CommitCount  int
}

type Bucket struct {
	Added   int
	Deleted int
	Commits int
}

type Engine struct {
	commits    []extract.CommitEvent
	timestamps []time.Time
	addPrefix  []int
	delPrefix  []int
	meta       extract.RepoMeta
}

func Build(ctx context.Context, src extract.EventSource) (*Engine, error) {
	stream, err := src.Stream(ctx)
	if err != nil {
		return nil, err
	}

	var commits []extract.CommitEvent
	for event := range stream {
		if commit, ok := event.(extract.CommitEvent); ok {
			commits = append(commits, commit)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sortChronologically(commits)
	return newEngine(commits, src.Meta()), nil
}

func (e *Engine) At(cursor time.Time) WorldState {
	count := upperBound(e.timestamps, cursor)
	return WorldState{
		Cursor:       cursor,
		LinesAdded:   e.addPrefix[count],
		LinesDeleted: e.delPrefix[count],
		CommitCount:  count,
	}
}

func (e *Engine) Series(from, to time.Time, buckets int) []Bucket {
	if buckets <= 0 {
		return nil
	}

	series := make([]Bucket, buckets)
	span := to.Sub(from)
	previous := lowerBound(e.timestamps, from)

	for i := 0; i < buckets; i++ {
		boundary := to
		if span > 0 && i < buckets-1 {
			frac := float64(i+1) / float64(buckets)
			boundary = from.Add(time.Duration(float64(span) * frac))
		}

		next := upperBound(e.timestamps, boundary)
		series[i] = Bucket{
			Added:   e.addPrefix[next] - e.addPrefix[previous],
			Deleted: e.delPrefix[next] - e.delPrefix[previous],
			Commits: next - previous,
		}
		previous = next
	}
	return series
}

func (e *Engine) Log(cursor time.Time, limit int) []extract.CommitEvent {
	if limit <= 0 {
		return nil
	}

	end := upperBound(e.timestamps, cursor)
	start := end - limit
	if start < 0 {
		start = 0
	}

	recent := make([]extract.CommitEvent, 0, end-start)
	for i := end - 1; i >= start; i-- {
		recent = append(recent, e.commits[i])
	}
	return recent
}

type Granularity int

const (
	ByCommit Granularity = iota
	ByDay
	ByWeek
)

func (g Granularity) Label() string {
	switch g {
	case ByDay:
		return "dia"
	case ByWeek:
		return "semana"
	default:
		return "commit"
	}
}

func (g Granularity) Coarser() Granularity {
	if g < ByWeek {
		return g + 1
	}
	return g
}

func (g Granularity) Finer() Granularity {
	if g > ByCommit {
		return g - 1
	}
	return g
}

func (e *Engine) Step(cursor time.Time, granularity Granularity, forward bool) time.Time {
	switch granularity {
	case ByDay:
		return e.clampToHistory(shiftDays(cursor, 1, forward))
	case ByWeek:
		return e.clampToHistory(shiftDays(cursor, 7, forward))
	default:
		if forward {
			return e.Next(cursor)
		}
		return e.Prev(cursor)
	}
}

func (e *Engine) clampToHistory(t time.Time) time.Time {
	if len(e.timestamps) == 0 {
		return t
	}
	first := e.timestamps[0]
	last := e.timestamps[len(e.timestamps)-1]
	if t.Before(first) {
		return first
	}
	if t.After(last) {
		return last
	}
	return t
}

func shiftDays(cursor time.Time, days int, forward bool) time.Time {
	if !forward {
		days = -days
	}
	return cursor.AddDate(0, 0, days)
}

func (e *Engine) Next(cursor time.Time) time.Time {
	idx := upperBound(e.timestamps, cursor)
	if idx < len(e.timestamps) {
		return e.timestamps[idx]
	}
	return cursor
}

func (e *Engine) Prev(cursor time.Time) time.Time {
	idx := lowerBound(e.timestamps, cursor) - 1
	if idx >= 0 {
		return e.timestamps[idx]
	}
	return cursor
}

func (e *Engine) Languages(cursor time.Time) []LanguageShare {
	totals := make(map[string]int)
	count := upperBound(e.timestamps, cursor)
	for i := 0; i < count; i++ {
		for _, file := range e.commits[i].Files {
			if file.Language == "" {
				continue
			}
			totals[file.Language] += file.LinesAdded - file.LinesDeleted
		}
	}

	shares := make([]LanguageShare, 0, len(totals))
	for name, lines := range totals {
		if lines <= 0 {
			continue
		}
		shares = append(shares, LanguageShare{Name: name, Lines: lines})
	}
	sort.Slice(shares, func(i, j int) bool {
		if shares[i].Lines == shares[j].Lines {
			return shares[i].Name < shares[j].Name
		}
		return shares[i].Lines > shares[j].Lines
	})
	return shares
}

type DAGSummary struct {
	Commits int
	Merges  int
	Tips    int
}

func (e *Engine) DAGStats(cursor time.Time) DAGSummary {
	count := upperBound(e.timestamps, cursor)
	subset := e.commits[:count]

	isParent := make(map[string]bool)
	merges := 0
	for _, commit := range subset {
		if len(commit.Parents) >= 2 {
			merges++
		}
		for _, parent := range commit.Parents {
			isParent[parent] = true
		}
	}

	tips := 0
	for _, commit := range subset {
		if !isParent[commit.Hash] {
			tips++
		}
	}
	return DAGSummary{Commits: count, Merges: merges, Tips: tips}
}

func (e *Engine) Heatmap(cursor time.Time) [7][24]int {
	var grid [7][24]int
	count := upperBound(e.timestamps, cursor)
	for i := 0; i < count; i++ {
		t := e.commits[i].Timestamp
		grid[int(t.Weekday())][t.Hour()]++
	}
	return grid
}

func (e *Engine) Summary(cursor time.Time) stats.Summary {
	count := upperBound(e.timestamps, cursor)
	return stats.Compute(e.commits[:count])
}

func (e *Engine) Meta() extract.RepoMeta {
	return e.meta
}

func newEngine(commits []extract.CommitEvent, meta extract.RepoMeta) *Engine {
	n := len(commits)
	engine := &Engine{
		commits:    commits,
		timestamps: make([]time.Time, n),
		addPrefix:  make([]int, n+1),
		delPrefix:  make([]int, n+1),
		meta:       meta,
	}
	for i, commit := range commits {
		engine.timestamps[i] = commit.Timestamp
		engine.addPrefix[i+1] = engine.addPrefix[i] + commit.LinesAdded
		engine.delPrefix[i+1] = engine.delPrefix[i] + commit.LinesDeleted
	}
	return engine
}

func sortChronologically(commits []extract.CommitEvent) {
	sort.SliceStable(commits, func(i, j int) bool {
		if commits[i].Timestamp.Equal(commits[j].Timestamp) {
			return commits[i].Hash < commits[j].Hash
		}
		return commits[i].Timestamp.Before(commits[j].Timestamp)
	})
}

func upperBound(timestamps []time.Time, t time.Time) int {
	return sort.Search(len(timestamps), func(i int) bool {
		return timestamps[i].After(t)
	})
}

func lowerBound(timestamps []time.Time, t time.Time) int {
	return sort.Search(len(timestamps), func(i int) bool {
		return !timestamps[i].Before(t)
	})
}
