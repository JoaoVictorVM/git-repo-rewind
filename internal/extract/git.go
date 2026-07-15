package extract

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var _ EventSource = (*GitExtractor)(nil)

type GitExtractor struct {
	commits []CommitEvent
	meta    RepoMeta
}

func NewGitExtractor(path string) (*GitExtractor, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("abrir repositorio em %q: %w", path, err)
	}

	commits, err := collectCommits(repo)
	if err != nil {
		return nil, err
	}

	return &GitExtractor{commits: commits, meta: buildMeta(repo, commits)}, nil
}

func (g *GitExtractor) Stream(ctx context.Context) (<-chan Event, error) {
	out := make(chan Event)
	go func() {
		defer close(out)
		for _, commit := range g.commits {
			select {
			case <-ctx.Done():
				return
			case out <- commit:
			}
		}
	}()
	return out, nil
}

func (g *GitExtractor) Meta() RepoMeta {
	return g.meta
}

func collectCommits(repo *git.Repository) ([]CommitEvent, error) {
	head, err := repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("resolver HEAD: %w", err)
	}

	iter, err := repo.Log(&git.LogOptions{From: head.Hash(), Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, fmt.Errorf("percorrer historico: %w", err)
	}
	defer iter.Close()

	var commits []CommitEvent
	err = iter.ForEach(func(commit *object.Commit) error {
		event, err := toCommitEvent(commit)
		if err != nil {
			return err
		}
		commits = append(commits, event)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ler commits: %w", err)
	}

	sortChronologically(commits)
	return commits, nil
}

func toCommitEvent(commit *object.Commit) (CommitEvent, error) {
	stats, err := commit.Stats()
	if err != nil {
		return CommitEvent{}, fmt.Errorf("stats do commit %s: %w", commit.Hash, err)
	}

	added, deleted := 0, 0
	files := make([]string, 0, len(stats))
	for _, stat := range stats {
		added += stat.Addition
		deleted += stat.Deletion
		files = append(files, stat.Name)
	}

	parents := make([]string, 0, commit.NumParents())
	for _, parent := range commit.ParentHashes {
		parents = append(parents, parent.String())
	}

	return CommitEvent{
		Timestamp:    commit.Committer.When,
		Hash:         commit.Hash.String(),
		Author:       commit.Author.Name,
		Email:        commit.Author.Email,
		Message:      strings.TrimSpace(commit.Message),
		LinesAdded:   added,
		LinesDeleted: deleted,
		Files:        files,
		Parents:      parents,
	}, nil
}

func sortChronologically(commits []CommitEvent) {
	sort.SliceStable(commits, func(i, j int) bool {
		if commits[i].Timestamp.Equal(commits[j].Timestamp) {
			return commits[i].Hash < commits[j].Hash
		}
		return commits[i].Timestamp.Before(commits[j].Timestamp)
	})
}

func buildMeta(repo *git.Repository, commits []CommitEvent) RepoMeta {
	meta := RepoMeta{
		Name:          repoName(repo),
		DefaultBranch: defaultBranch(repo),
		TotalCommits:  len(commits),
	}
	if len(commits) > 0 {
		meta.FirstCommit = commits[0].Timestamp
		meta.LastCommit = commits[len(commits)-1].Timestamp
	}
	return meta
}

func repoName(repo *git.Repository) string {
	worktree, err := repo.Worktree()
	if err != nil {
		return ""
	}
	return filepath.Base(worktree.Filesystem.Root())
}

func defaultBranch(repo *git.Repository) string {
	head, err := repo.Head()
	if err != nil || !head.Name().IsBranch() {
		return ""
	}
	return head.Name().Short()
}
