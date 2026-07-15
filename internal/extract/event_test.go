package extract

import (
	"testing"
	"time"
)

func TestEventsExposeTimestamp(t *testing.T) {
	ts := time.Date(2026, 7, 15, 10, 30, 0, 0, time.UTC)
	events := []Event{
		CommitEvent{Timestamp: ts},
		FileChangeEvent{Timestamp: ts},
		BranchEvent{Timestamp: ts},
	}
	for _, e := range events {
		if !e.At().Equal(ts) {
			t.Errorf("%T.At() = %v, quer %v", e, e.At(), ts)
		}
	}
}
