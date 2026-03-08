package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepBuild_DisplaysProgress(t *testing.T) {
	step := NewBuildStep("/augury-node")
	step.state = "building"
	step.buildOutput = "Building package..."

	view := step.View()
	if !strings.Contains(view, "Building") {
		t.Error("View should show build status")
	}
	if !strings.Contains(view, "Building package...") {
		t.Error("View should show build output")
	}
}

func TestStepBuild_BuildCompleteAdvances(t *testing.T) {
	step := NewBuildStep("/augury-node")
	buildMsg := BuildCompleteMsg{Success: true}
	step, cmd := step.Update(buildMsg)

	if !step.Confirmed() {
		t.Error("Should confirm on successful build")
	}
	if cmd == nil {
		t.Fatal("Should return command to advance")
	}
	msg := cmd()
	if _, ok := msg.(NextStepMsg); !ok {
		t.Errorf("Command should return NextStepMsg, got %T", msg)
	}
}

func TestStepBuild_BuildFailureShowsError(t *testing.T) {
	step := NewBuildStep("/augury-node")
	buildMsg := BuildCompleteMsg{Success: false, Error: "build failed"}
	step, _ = step.Update(buildMsg)

	view := step.View()
	if !strings.Contains(view, "build failed") {
		t.Error("View should show build error")
	}
}

func TestStepBuild_QuitDuringBuild(t *testing.T) {
	step := NewBuildStep("/augury-node")
	step.state = "building"

	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("q should return quit command")
	}
	// Note: tea.Quit is a function, we can't easily assert it matches
	// but we verify a command is returned
}
