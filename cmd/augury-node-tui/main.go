package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/augurysys/augury-node-tui/internal/app"
	"github.com/augurysys/augury-node-tui/internal/config"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/setup"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/workspace"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		runSetupWizard()
		return
	}

	rootFlag := flag.String("root", "", "path to augury-node repository")
	flag.Parse()

	var configPath string
	if cfgPathStr, err := config.DefaultPath(); err == nil {
		if cfg, err := config.Read(cfgPathStr); err == nil {
			configPath = cfg.AuguryNodeRoot
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	root, err := workspace.ResolveRoot(*rootFlag, configPath, cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "augury-node-tui: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run from an augury-node repo, configure via 'augury-node-tui setup', or pass --root.\n")
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
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "augury-node-tui: %v\n", err)
		os.Exit(1)
	}
}

func runSetupWizard() {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	reconfigure := fs.Bool("reconfigure", false, "run wizard even if config exists")
	_ = fs.Parse(os.Args[2:])

	wizard := setup.NewWizard(*reconfigure)
	p := tea.NewProgram(wizard, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Setup wizard error: %v\n", err)
		os.Exit(1)
	}

	if w, ok := finalModel.(*setup.WizardModel); ok && w.LaunchMainRequested() {
		fmt.Println("Launching main TUI...")
		// Could re-exec or run main TUI here
		// For now just print message
	}
}
