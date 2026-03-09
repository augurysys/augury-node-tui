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
	cmd := viewer.JumpToFirstError()
	if cmd == nil {
		t.Error("JumpToFirstError should return command")
	}

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
