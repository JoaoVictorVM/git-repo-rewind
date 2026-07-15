package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

type repoInfo struct {
	Root   string
	Branch string
}

func detectRepo(path string) (repoInfo, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return repoInfo{}, fmt.Errorf("nenhum repositorio git encontrado a partir de %q: %w", path, err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return repoInfo{}, fmt.Errorf("nao foi possivel abrir a worktree: %w", err)
	}

	info := repoInfo{Root: worktree.Filesystem.Root()}

	if head, err := repo.Head(); err == nil && head.Name().IsBranch() {
		info.Branch = head.Name().Short()
	}

	return info, nil
}
