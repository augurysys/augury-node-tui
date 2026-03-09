package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogViewer_Creation(t *testing.T) {
	content := "Line 1\nLine 2\nerror: something bad\nLine 4"
	viewer := NewLogViewer(content)

	if viewer == nil {
		t.Fatal("NewLogViewer returned nil")
	}

	view := viewer.View()
	if !strings.Contains(view, "Line 1") {
		t.Error("LogViewer should render content")
	}
}

func TestLogViewer_JumpToFirstError(t *testing.T) {
	content := "Line 1\nLine 2\nerror: problem here\nLine 4"
	viewer := NewLogViewer(content)

	// Jump to first error
	viewer.JumpToFirstError()

	// Error should be detected
	if len(viewer.Errors()) == 0 {
		t.Error("Should detect error in content")
	}
}

func TestLogViewer_Navigation(t *testing.T) {
	content := strings.Repeat("Line\n", 100)
	viewer := NewLogViewer(content)
	viewer.SetHeight(20)

	// Simulate scroll down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	viewer.Update(msg)

	view := viewer.View()
	if view == "" {
		t.Error("LogViewer should render after navigation")
	}
}

func TestLogViewer_ErrorNavigation(t *testing.T) {
	content := "Line 1\nerror: first problem\nLine 3\nerror: second problem\nLine 5"
	viewer := NewLogViewer(content)

	// Should detect 2 errors
	if len(viewer.Errors()) != 2 {
		t.Fatalf("Expected 2 errors, got: %d", len(viewer.Errors()))
	}

	// Jump to first error
	viewer.JumpToFirstError()
	view := viewer.View()

	// Status bar should show 1/2
	if !strings.Contains(view, "1/2") {
		t.Error("Status bar should show error 1/2")
	}

	// Next error
	viewer.NextError()
	view = viewer.View()

	// Status bar should show 2/2
	if !strings.Contains(view, "2/2") {
		t.Error("Status bar should show error 2/2 after NextError")
	}

	// Previous error (wraps to last)
	viewer.PrevError()
	view = viewer.View()

	// Should show 1/2 again
	if !strings.Contains(view, "1/2") {
		t.Error("Status bar should show error 1/2 after PrevError")
	}
}
