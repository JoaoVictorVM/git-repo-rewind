package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoVictorVM/git-repo-rewind/internal/engine"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/extract"
	"github.com/JoaoVictorVM/git-repo-rewind/internal/tui"
)

const appName = "rewind"

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", appName, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet(appName, flag.ContinueOnError)
	showVersion := fs.Bool("version", false, "mostra a versao e sai")
	repoPath := fs.String("path", ".", "caminho do repositorio git a explorar")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Printf("%s %s\n", appName, version)
		return nil
	}

	target := *repoPath
	if arg := fs.Arg(0); arg != "" {
		target = arg
	}

	extractor, err := extract.NewGitExtractor(target)
	if err != nil {
		return err
	}

	eng, err := engine.Build(context.Background(), extractor)
	if err != nil {
		return err
	}

	program := tea.NewProgram(tui.New(eng), tea.WithAltScreen())
	_, err = program.Run()
	return err
}
