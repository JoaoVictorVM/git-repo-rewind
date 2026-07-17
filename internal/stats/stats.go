package stats

import (
	"fmt"
	"sort"
	"time"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
)

type DayStat struct {
	Date    string
	Commits int
}

type FileStat struct {
	Path    string
	Commits int
}

type Summary struct {
	TotalCommits  int
	Authors       int
	BusiestDay    DayStat
	LongestStreak int
	TopFile       FileStat
}

func Compute(commits []extract.CommitEvent) Summary {
	summary := Summary{TotalCommits: len(commits)}
	if len(commits) == 0 {
		return summary
	}

	authors := make(map[string]bool)
	perDay := make(map[string]int)
	perFile := make(map[string]int)
	activeDays := make(map[int]bool)

	for _, commit := range commits {
		authors[authorKey(commit)] = true

		year, month, day := commit.Timestamp.Date()
		perDay[fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)]++
		activeDays[dayOrdinal(year, month, day)] = true

		touched := make(map[string]bool)
		for _, file := range commit.Files {
			if touched[file.Path] {
				continue
			}
			touched[file.Path] = true
			perFile[file.Path]++
		}
	}

	summary.Authors = len(authors)
	summary.BusiestDay = busiestDay(perDay)
	summary.TopFile = mostTouchedFile(perFile)
	summary.LongestStreak = longestStreak(activeDays)
	return summary
}

func authorKey(commit extract.CommitEvent) string {
	if commit.Email != "" {
		return commit.Email
	}
	return commit.Author
}

func dayOrdinal(year int, month time.Month, day int) int {
	return int(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix() / 86400)
}

func busiestDay(perDay map[string]int) DayStat {
	best := DayStat{}
	for date, commits := range perDay {
		if commits > best.Commits || (commits == best.Commits && (best.Date == "" || date < best.Date)) {
			best = DayStat{Date: date, Commits: commits}
		}
	}
	return best
}

func mostTouchedFile(perFile map[string]int) FileStat {
	best := FileStat{}
	for path, commits := range perFile {
		if commits > best.Commits || (commits == best.Commits && (best.Path == "" || path < best.Path)) {
			best = FileStat{Path: path, Commits: commits}
		}
	}
	return best
}

func longestStreak(activeDays map[int]bool) int {
	if len(activeDays) == 0 {
		return 0
	}

	ordinals := make([]int, 0, len(activeDays))
	for ordinal := range activeDays {
		ordinals = append(ordinals, ordinal)
	}
	sort.Ints(ordinals)

	longest, current := 1, 1
	for i := 1; i < len(ordinals); i++ {
		if ordinals[i] == ordinals[i-1]+1 {
			current++
		} else {
			current = 1
		}
		if current > longest {
			longest = current
		}
	}
	return longest
}
