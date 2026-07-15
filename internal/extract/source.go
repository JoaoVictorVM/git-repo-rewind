package extract

import (
	"context"
	"time"
)

type RepoMeta struct {
	Name          string
	DefaultBranch string
	TotalCommits  int
	FirstCommit   time.Time
	LastCommit    time.Time
}

type EventSource interface {
	Stream(ctx context.Context) (<-chan Event, error)
	Meta() RepoMeta
}
