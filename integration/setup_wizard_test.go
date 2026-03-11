package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/config"
	"github.com/augurysys/augury-node-tui/internal/setup"
	"github.com/augurysys/augury-node-tui/internal/workspace"
	tea "github.com/charmbracelet/bubbletea"
)

// setupFakeAuguryNode creates a minimal augury-node structure in a temp dir.
// Returns the absolute path to the root. Valid for workspace.ValidateRoot.
func setupFakeAuguryNode(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	for _, dir := range []string{"scripts/devices", "scripts/lib", "pkg"} {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0755); err != nil {
			t.Fatalf("create dir %s: %v", dir, err)
		}
	}
	abs, err := filepath.Abs(tmp)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	return abs
}

func TestSetupWizard_RootStepAutoDetection(t *testing.T) {
	root := setupFakeAuguryNode(t)
	if err := workspace.ValidateRoot(root); err != nil {
		t.Fatalf("fixture must be valid: %v", err)
	}

	// From root: FindAuguryNodeRoot should find it
	found, err := setup.FindAuguryNodeRoot(root)
	if err != nil {
		t.Fatalf("FindAuguryNodeRoot: %v", err)
	}
	if found != root {
		t.Errorf("FindAuguryNodeRoot from root: got %q, want %q", found, root)
	}

	// From subdir: should find ancestor
	subdir := filepath.Join(root, "scripts", "devices")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	found, err = setup.FindAuguryNodeRoot(subdir)
	if err != nil {
		t.Fatalf("FindAuguryNodeRoot from subdir: %v", err)
	}
	if found != root {
		t.Errorf("FindAuguryNodeRoot from subdir: got %q, want %q", found, root)
	}

	// Wizard created from root should show detected path in view
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	w := setup.NewWizard(false)
	view := w.View()
	if view == "" || !strings.Contains(view, root) {
		t.Errorf("root step view should display detected path %q; view: %q", root, view)
	}
}

func TestSetupWizard_FullFlowSimulation(t *testing.T) {
	root := setupFakeAuguryNode(t)
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	// Redirect config to temp dir
	configDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", configDir)
	defer os.Setenv("HOME", origHome)

	w := setup.NewWizard(false)
	if w.CurrentStep() != 0 {
		t.Fatalf("wizard should start at step 0, got %d", w.CurrentStep())
	}

	// Advance through wizard: RootConfirmedMsg -> NixHealthCheckMsg (cascades through
	// groups + install via Init cmds) -> BuildCompleteMsg -> success
	var model tea.Model = w
	runUpdates := func(msg tea.Msg) {
		var cmd tea.Cmd
		model, cmd = model.(*setup.WizardModel).Update(msg)
		for cmd != nil {
			model, cmd = model.(*setup.WizardModel).Update(cmd())
		}
	}

	runUpdates(setup.RootConfirmedMsg{Path: root})
	w = model.(*setup.WizardModel)
	if w.CurrentStep() != 1 {
		t.Fatalf("after root confirm: step %d, want 1", w.CurrentStep())
	}

	runUpdates(setup.NixHealthCheckMsg{
		NixInstalled:        setup.HealthCheckResult{Available: true},
		ExperimentalEnabled: setup.HealthCheckResult{Available: true},
		DaemonOk:            setup.HealthCheckResult{Available: true},
	})
	w = model.(*setup.WizardModel)
	// Nix+Groups auto-advance via Init cascade; Install Init returns AlreadyInstalled: false
	if w.CurrentStep() != 3 {
		t.Fatalf("after nix (cascade): step %d, want 3", w.CurrentStep())
	}

	runUpdates(setup.BinaryBuiltMsg{Binary: "/tmp/augury-node-tui", AlreadyInstalled: true, Err: ""})
	w = model.(*setup.WizardModel)
	if w.CurrentStep() != 4 {
		t.Fatalf("after install: step %d, want 4", w.CurrentStep())
	}

	runUpdates(setup.BuildCompleteMsg{Success: true})
	w = model.(*setup.WizardModel)
	if w.CurrentStep() != 5 {
		t.Fatalf("after build: step %d, want 5", w.CurrentStep())
	}

	view := w.View()
	if view == "" {
		t.Error("success step view should not be empty")
	}
}

func TestSetupWizard_ConfigPersistence(t *testing.T) {
	root := setupFakeAuguryNode(t)
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	configDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", configDir)
	defer os.Setenv("HOME", origHome)

	w := setup.NewWizard(false)

	// Advance through root step (persists config)
	model, cmd := w.Update(setup.RootConfirmedMsg{Path: root})
	for cmd != nil {
		model, cmd = model.(*setup.WizardModel).Update(cmd())
	}
	w = model.(*setup.WizardModel)

	// Config should exist
	path, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	cfg, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read config: %v", err)
	}
	if cfg.AuguryNodeRoot != root {
		t.Errorf("config augury_node_root: got %q, want %q", cfg.AuguryNodeRoot, root)
	}
	if len(cfg.CompletedSteps) == 0 || cfg.CompletedSteps[0] != "root" {
		t.Errorf("config completed_steps: got %v, want [root]", cfg.CompletedSteps)
	}
}
