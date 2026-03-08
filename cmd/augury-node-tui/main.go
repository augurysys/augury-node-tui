package main

import (
	"fmt"
	"os"
	"time"

	"github.com/augurysys/augury-node-tui/internal/app"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/workspace"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	root, err := workspace.ResolveRoot("", "", cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "augury-node-tui: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run from an augury-node repo or pass --root.\n")
		os.Exit(1)
	}
	st, err := status.Collect(root)
	if err != nil {
		dirty := make(map[string]bool)
		for _, p := range status.RequiredPaths {
			dirty[p] = false
		}
		st = status.RepoStatus{Root: root, Branch: "?", SHA: "?", Dirty: dirty}
	}
	// NewModel probes nix at init; app enforces mandatory nix gating for executable actions
	m := app.NewModel(st, platform.Registry(), 2*time.Second)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
