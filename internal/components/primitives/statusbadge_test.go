package primitives

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestStatusBadge_AllStatusTypes(t *testing.T) {
	statuses := []Status{
		StatusSuccess,
		StatusError,
		StatusWarning,
		StatusRunning,
		StatusBlocked,
		StatusUnavailable,
	}

	for _, status := range statuses {
		badge := StatusBadge{Label: "Test", Status: status}
		rendered := badge.Render()

		if rendered == "" {
			t.Errorf("Status %d produced empty render", status)
		}
		if !strings.Contains(rendered, "Test") {
			t.Errorf("Badge missing label for status %d", status)
		}
	}
}

func TestStatusBadge_IconMapping(t *testing.T) {
	tests := []struct {
		status       Status
		expectedIcon string
	}{
		{StatusSuccess, "✓"},
		{StatusError, "✗"},
		{StatusWarning, "⚠"},
		{StatusRunning, "●"},
		{StatusBlocked, "⊘"},
		{StatusUnavailable, "◌"},
	}

	for _, tt := range tests {
		badge := StatusBadge{Label: "Test", Status: tt.status}
		rendered := badge.Render()

		if !strings.Contains(rendered, tt.expectedIcon) {
			t.Errorf("Status %d should contain icon %s, got: %s",
				tt.status, tt.expectedIcon, stripAnsi(rendered))
		}
	}
}

func TestStatusBadge_ColorMapping(t *testing.T) {
	// Force color output so ANSI codes are emitted in non-TTY (e.g. go test)
	lipgloss.SetColorProfile(termenv.TrueColor)

	// Success should have green color code
	success := StatusBadge{Label: "OK", Status: StatusSuccess}
	successRender := success.Render()

	// Error should have red color code
	errBadge := StatusBadge{Label: "Failed", Status: StatusError}
	errorRender := errBadge.Render()

	// Renders should differ (different ANSI codes)
	if successRender == errorRender {
		t.Error("Success and error badges should have different colors")
	}

	// Both should contain ANSI escape sequences
	if !strings.Contains(successRender, "\x1b[") {
		t.Error("Success badge should contain color codes")
	}
	if !strings.Contains(errorRender, "\x1b[") {
		t.Error("Error badge should contain color codes")
	}
}
