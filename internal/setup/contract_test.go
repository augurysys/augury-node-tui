package setup

import (
	"path/filepath"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// Contract: WizardModel implements tea.Model
var _ tea.Model = (*WizardModel)(nil)

// Contract: All step models have View and Confirmed
var (
	_ interface {
		View() string
		Confirmed() bool
	} = (*RootStep)(nil)
	_ interface {
		Init() tea.Cmd
		View() string
		Confirmed() bool
	} = (*NixStepModel)(nil)
	_ interface {
		Init() tea.Cmd
		View() string
		Confirmed() bool
	} = (*GroupsStepModel)(nil)
	_ interface {
		Init() tea.Cmd
		View() string
		Confirmed() bool
	} = (*InstallStepModel)(nil)
	_ interface {
		Init() tea.Cmd
		View() string
		Confirmed() bool
	} = (*BuildStepModel)(nil)
	_ interface {
		Init() tea.Cmd
		View() string
		Confirmed() bool
	} = (*SuccessStepModel)(nil)
)

func TestContract_ConfigFileFormat(t *testing.T) {
	// Config struct must have expected TOML fields for round-trip
	cfg := config.Config{
		AuguryNodeRoot:   "/path/to/augury-node",
		BinaryInstalled:  true,
		NixVerified:      true,
		SetupCompletedAt: "2026-03-08T12:00:00Z",
		CompletedSteps:   []string{"root", "nix"},
		SkippedSteps:     []string{"groups"},
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}
	loaded, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if loaded.AuguryNodeRoot != cfg.AuguryNodeRoot {
		t.Errorf("augury_node_root: got %q, want %q", loaded.AuguryNodeRoot, cfg.AuguryNodeRoot)
	}
	if loaded.BinaryInstalled != cfg.BinaryInstalled {
		t.Errorf("binary_installed: got %v, want %v", loaded.BinaryInstalled, cfg.BinaryInstalled)
	}
	if loaded.NixVerified != cfg.NixVerified {
		t.Errorf("nix_verified: got %v, want %v", loaded.NixVerified, cfg.NixVerified)
	}
	if len(loaded.CompletedSteps) != len(cfg.CompletedSteps) {
		t.Errorf("completed_steps length: got %d, want %d", len(loaded.CompletedSteps), len(cfg.CompletedSteps))
	}
	if len(loaded.SkippedSteps) != len(cfg.SkippedSteps) {
		t.Errorf("skipped_steps length: got %d, want %d", len(loaded.SkippedSteps), len(cfg.SkippedSteps))
	}
}

func TestContract_MessageTypes(t *testing.T) {
	// Message types must be usable as tea.Msg
	var _ tea.Msg = NextStepMsg{}
	var _ tea.Msg = LaunchMainTUIMsg{}
	var _ tea.Msg = RootConfirmedMsg{}
	var _ tea.Msg = NixHealthCheckMsg{}
	var _ tea.Msg = NixFixResultMsg{}
	var _ tea.Msg = GroupCheckMsg{}
	var _ tea.Msg = InstallCheckMsg{}
	var _ tea.Msg = BuildOutputMsg{}
	var _ tea.Msg = BuildCompleteMsg{}
}
