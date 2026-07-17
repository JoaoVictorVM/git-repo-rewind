package extract

import "time"

type Event interface {
	At() time.Time
	isEvent()
}

type FileChange struct {
	Path         string
	Language     string
	LinesAdded   int
	LinesDeleted int
}

type CommitEvent struct {
	Timestamp    time.Time
	Hash         string
	Author       string
	Email        string
	Message      string
	LinesAdded   int
	LinesDeleted int
	Files        []FileChange
	Parents      []string
}

func (e CommitEvent) At() time.Time { return e.Timestamp }

func (CommitEvent) isEvent() {}

type FileChangeEvent struct {
	Timestamp    time.Time
	CommitHash   string
	Path         string
	Language     string
	LinesAdded   int
	LinesDeleted int
}

func (e FileChangeEvent) At() time.Time { return e.Timestamp }

func (FileChangeEvent) isEvent() {}

type BranchAction int

const (
	BranchCreated BranchAction = iota
	BranchMerged
)

type BranchEvent struct {
	Timestamp time.Time
	Action    BranchAction
	Name      string
	Commit    string
	Parents   []string
}

func (e BranchEvent) At() time.Time { return e.Timestamp }

func (BranchEvent) isEvent() {}
