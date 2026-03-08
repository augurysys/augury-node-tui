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
	cwd, _ := os.Getwd()
	root, err := workspace.ResolveRoot("", "", cwd)
	if err != nil {
		root = cwd
	}
	st, err := status.Collect(root)
	if err != nil {
		dirty := make(map[string]bool)
		for _, p := range status.RequiredPaths {
			dirty[p] = false
		}
		st = status.RepoStatus{Root: root, Branch: "?", SHA: "?", Dirty: dirty}
	}
	m := app.NewModel(st, platform.Registry(), 2*time.Second)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
