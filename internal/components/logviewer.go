package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/augurysys/augury-node-tui/internal/logs"
	"github.com/augurysys/augury-node-tui/internal/styles"
)

// LogViewer displays scrollable logs with error navigation
type LogViewer struct {
	content    string
	errors     []logs.ErrorLocation
	currentErr int
	viewport   viewport.Model
}

// NewLogViewer creates a new log viewer
func NewLogViewer(content string) *LogViewer {
	vp := viewport.New(80, 20)
	vp.SetContent(content)

	errors := logs.ParseErrors(content)

	return &LogViewer{
		content:    content,
		errors:     errors,
		currentErr: -1,
		viewport:   vp,
	}
}

// SetHeight updates viewport height
func (v *LogViewer) SetHeight(height int) {
	v.viewport.Height = height
}

// JumpToFirstError moves viewport to first error
func (v *LogViewer) JumpToFirstError() tea.Cmd {
	if len(v.errors) == 0 {
		return nil
	}

	v.currentErr = 0
	err := v.errors[0]

	// Scroll to error line (YOffset is 0-indexed line number)
	v.viewport.GotoTop()
	v.viewport.SetYOffset(max(0, err.LineNumber-1))

	return func() tea.Msg { return nil }
}

// NextError jumps to next error
func (v *LogViewer) NextError() tea.Cmd {
	if len(v.errors) == 0 {
		return nil
	}

	v.currentErr = (v.currentErr + 1) % len(v.errors)
	err := v.errors[v.currentErr]

	v.viewport.GotoTop()
	v.viewport.SetYOffset(max(0, err.LineNumber-1))

	return nil
}

// PrevError jumps to previous error
func (v *LogViewer) PrevError() tea.Cmd {
	if len(v.errors) == 0 {
		return nil
	}

	v.currentErr--
	if v.currentErr < 0 {
		v.currentErr = len(v.errors) - 1
	}

	err := v.errors[v.currentErr]
	v.viewport.GotoTop()
	v.viewport.SetYOffset(max(0, err.LineNumber-1))

	return nil
}

// Update handles key messages
func (v *LogViewer) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			return v.JumpToFirstError()
		case "n":
			return v.NextError()
		case "N":
			return v.PrevError()
		default:
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return cmd
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

// View renders the log viewer
func (v *LogViewer) View() string {
	palette := styles.DefaultPalette()

	// Get viewport content
	content := v.viewport.View()

	// Add status bar if errors present
	var statusBar string
	if len(v.errors) > 0 && v.currentErr >= 0 {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Error))
		statusBar = errStyle.Render(
			fmt.Sprintf(" [Error %d/%d] ", v.currentErr+1, len(v.errors)),
		)
	}

	return content + "\n" + statusBar
}

// Errors returns detected error locations
func (v *LogViewer) Errors() []logs.ErrorLocation {
	return v.errors
}
