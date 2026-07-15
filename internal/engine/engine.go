package engine

import (
	"context"
	"sort"
	"time"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

type WorldState struct {
	Cursor       time.Time
	LinesAdded   int
	LinesDeleted int
	CommitCount  int
}

type Engine struct {
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
	count := sort.Search(len(e.timestamps), func(i int) bool {
		return e.timestamps[i].After(cursor)
	})
	return WorldState{
		Cursor:       cursor,
		LinesAdded:   e.addPrefix[count],
		LinesDeleted: e.delPrefix[count],
		CommitCount:  count,
	}
}

func (e *Engine) Meta() extract.RepoMeta {
	return e.meta
}

func newEngine(commits []extract.CommitEvent, meta extract.RepoMeta) *Engine {
	n := len(commits)
	engine := &Engine{
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
