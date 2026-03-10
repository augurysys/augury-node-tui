package setup

import (
	"strings"
	"testing"
)

func TestStepInstall_DisplaysTargetPath(t *testing.T) {
	step := NewInstallStep(t.TempDir())
	step.builtBinary = "/tmp/test/bin/augury-node-tui"
	step.state = "ready"
	view := step.View()

	if !strings.Contains(view, "/usr/local/bin") {
		t.Error("View should show target installation path")
	}
}

func TestStepInstall_AlreadyInstalledAutoAdvances(t *testing.T) {
	step := NewInstallStep(t.TempDir())
	builtMsg := BinaryBuiltMsg{Binary: "/tmp/binary", AlreadyInstalled: true}
	step, cmd := step.Update(builtMsg)

	if !step.Confirmed() {
		t.Error("Should auto-confirm if already installed")
	}
	if cmd == nil {
		t.Fatal("Should return command to advance")
	}
	msg := cmd()
	if _, ok := msg.(NextStepMsg); !ok {
		t.Errorf("Command should return NextStepMsg, got %T", msg)
	}
}

func TestStepInstall_ShowsSudoCommand(t *testing.T) {
	step := NewInstallStep(t.TempDir())
	step.builtBinary = "/tmp/test/bin/augury-node-tui"
	step.alreadyInstalled = false
	step.state = "ready"

	view := step.View()
	if !strings.Contains(view, "sudo") {
		t.Error("View should show sudo requirement")
	}
	if !strings.Contains(view, "ln -sf") {
		t.Error("View should show symlink command")
	}
}
