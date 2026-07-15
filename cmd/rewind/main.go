package main

import (
	"flag"
	"fmt"
	"os"
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

	repo, err := detectRepo(target)
	if err != nil {
		return err
	}

	fmt.Printf("%s: repositorio detectado em %s\n", appName, repo.Root)
	if repo.Branch != "" {
		fmt.Printf("branch atual: %s\n", repo.Branch)
	} else {
		fmt.Println("branch atual: (sem commits ainda)")
	}
	return nil
}
