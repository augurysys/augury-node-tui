package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepNix_DisplaysChecks(t *testing.T) {
	step := NewNixStep()
	step.nixInstalled = true
	step.experimentalEnabled = false
	step.state = "ready"

	view := step.View()
	if !strings.Contains(view, "Nix installed") && !strings.Contains(view, "✓") {
		t.Error("View should show Nix installed status")
	}
	if !strings.Contains(view, "Experimental features") && !strings.Contains(view, "✗") {
		t.Error("View should show experimental features status")
	}
}

func TestStepNix_AutoFixExecutes(t *testing.T) {
	step := NewNixStep()
	step.nixInstalled = true
	step.experimentalEnabled = false
	step.state = "ready"

	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	if step.state != "fixing" {
		t.Error("Pressing 'f' should trigger auto-fix")
	}
	if cmd == nil {
		t.Error("Should return command to execute fix")
	}
}

func TestStepNix_AllChecksPassAutoAdvances(t *testing.T) {
	step := NewNixStep()

	// Simulate health check with all checks passing
	healthMsg := NixHealthCheckMsg{
		NixInstalled:        HealthCheckResult{Available: true},
		ExperimentalEnabled: HealthCheckResult{Available: true},
		DaemonOk:            HealthCheckResult{Available: true},
	}

	step, cmd := step.Update(healthMsg)

	if !step.AllChecksPassed() {
		t.Error("Should report all checks passed")
	}

	if !step.Confirmed() {
		t.Error("Should auto-confirm when all checks pass")
	}

	if cmd == nil {
		t.Fatal("Should return command to advance")
	}

	// Verify the command returns NextStepMsg
	msg := cmd()
	if _, ok := msg.(NextStepMsg); !ok {
		t.Errorf("Command should return NextStepMsg, got %T", msg)
	}
}
