package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepGroups_NotInGroupShowsCommands(t *testing.T) {
	step := NewGroupsStep()
	step.inNixUsers = false
	step.state = "ready"

	view := step.View()
	if !strings.Contains(view, "sudo usermod") {
		t.Error("View should show sudo command when not in group")
	}
	if !strings.Contains(view, "newgrp") {
		t.Error("View should show newgrp command")
	}
}

func TestStepGroups_AlreadyInGroupAutoAdvances(t *testing.T) {
	step := NewGroupsStep()

	// Simulate check showing user is in group
	checkMsg := GroupCheckMsg{InNixUsers: true}
	step, cmd := step.Update(checkMsg)

	if !step.Confirmed() {
		t.Error("Should auto-confirm if already in group")
	}
	if cmd == nil {
		t.Fatal("Should return command to advance")
	}

	// Verify command returns NextStepMsg
	msg := cmd()
	if _, ok := msg.(NextStepMsg); !ok {
		t.Errorf("Command should return NextStepMsg, got %T", msg)
	}
}

func TestStepGroups_RecheckUpdatesStatus(t *testing.T) {
	step := NewGroupsStep()
	step.inNixUsers = false
	step.state = "ready"

	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if step.state != "rechecking" {
		t.Error("Pressing 'r' should trigger recheck")
	}
	if cmd == nil {
		t.Error("Should return command to recheck")
	}
}
