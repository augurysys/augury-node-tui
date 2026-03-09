package components

import (
	"strings"
	"testing"
	"time"

	"github.com/augurysys/augury-node-tui/internal/ansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestCommandDisplay_Render_Executing(t *testing.T) {
	c := CommandDisplay{
		Command:     "nix develop .#dev-env --command scripts/devices/node2-build.sh",
		Description: "Building...",
		Executing:   true,
	}

	rendered := c.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "Running:") {
		t.Errorf("Executing state should contain 'Running:', got: %s", clean)
	}
	if !strings.Contains(clean, "nix develop .#dev-env --command scripts/devices/node2-build.sh") {
		t.Errorf("Should show command, got: %s", clean)
	}
	if !strings.Contains(clean, "[●]") {
		t.Errorf("Should show running icon [●], got: %s", clean)
	}
	if !strings.Contains(clean, "Building...") {
		t.Errorf("Should show description, got: %s", clean)
	}
}

func TestCommandDisplay_Render_Executing_EmptyDescription(t *testing.T) {
	c := CommandDisplay{
		Command:     "echo hello",
		Description: "",
		Executing:   true,
	}

	rendered := c.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "Running:") {
		t.Errorf("Should contain 'Running:', got: %s", clean)
	}
	if !strings.Contains(clean, "echo hello") {
		t.Errorf("Should show command, got: %s", clean)
	}
	// Empty description - we may or may not show [●]; spec says "Show description when executing"
	// So we show [●] with empty or minimal text
	if !strings.Contains(clean, "●") {
		t.Errorf("Should show running icon, got: %s", clean)
	}
}

func TestCommandDisplay_Render_Success(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(oldProfile) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	exit0 := 0
	c := CommandDisplay{
		Command:     "nix develop .#dev-env --command scripts/devices/node2-build.sh",
		Description: "Building...",
		Executing:   false,
		ExitCode:    &exit0,
		Duration:    ptrDuration(2*time.Minute + 34*time.Second),
	}

	rendered := c.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "✓") {
		t.Errorf("Success should show ✓, got: %s", clean)
	}
	if !strings.Contains(clean, "nix develop .#dev-env --command scripts/devices/node2-build.sh") {
		t.Errorf("Should show command, got: %s", clean)
	}
	if !strings.Contains(clean, "exit 0") {
		t.Errorf("Should show exit 0, got: %s", clean)
	}
	if !strings.Contains(clean, "2m34s") {
		t.Errorf("Should show duration 2m34s, got: %s", clean)
	}
	// Should have color codes for success (green)
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("Success should have color codes")
	}
}

func TestCommandDisplay_Render_Error(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(oldProfile) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	exit1 := 1
	c := CommandDisplay{
		Command:     "make build",
		Description: "Building...",
		Executing:   false,
		ExitCode:    &exit1,
		Duration:    ptrDuration(5 * time.Second),
	}

	rendered := c.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "✗") {
		t.Errorf("Error should show ✗, got: %s", clean)
	}
	if !strings.Contains(clean, "make build") {
		t.Errorf("Should show command, got: %s", clean)
	}
	if !strings.Contains(clean, "exit 1") {
		t.Errorf("Should show exit 1, got: %s", clean)
	}
	if !strings.Contains(clean, "5s") {
		t.Errorf("Should show duration 5s, got: %s", clean)
	}
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("Error should have color codes")
	}
}

func TestCommandDisplay_Render_Success_NoDuration(t *testing.T) {
	exit0 := 0
	c := CommandDisplay{
		Command:     "echo done",
		Description: "",
		Executing:   false,
		ExitCode:    &exit0,
		Duration:    nil,
	}

	rendered := c.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "✓") {
		t.Errorf("Should show ✓, got: %s", clean)
	}
	if !strings.Contains(clean, "exit 0") {
		t.Errorf("Should show exit 0, got: %s", clean)
	}
	// Should not have duration when nil
	if strings.Contains(clean, "0s") || strings.Contains(clean, "m") {
		// "0s" could appear from formatting; "m" from "exit 0" - be careful
		// Actually "exit 0" doesn't have "m". Let's check we don't have ", 0s" or ", Xs"
		// Format when duration is nil: "(exit 0)" without duration
		if strings.Contains(clean, ", ") && strings.HasSuffix(strings.TrimSpace(clean), "s") {
			t.Errorf("Should not show duration when nil, got: %s", clean)
		}
	}
}

func TestCommandDisplay_Render_NilExitCode_Complete(t *testing.T) {
	c := CommandDisplay{
		Command:     "unknown",
		Description: "",
		Executing:   false,
		ExitCode:    nil,
	}

	rendered := c.Render()
	clean := ansi.StripAnsi(rendered)

	// When ExitCode is nil and complete, we need to handle - show ? or default to error
	// Spec says "based on exit code" - we'll show ? with dim for unknown
	if !strings.Contains(clean, "unknown") {
		t.Errorf("Should show command, got: %s", clean)
	}
	// Should not panic
	if rendered == "" {
		t.Error("Should produce some output")
	}
}

func TestCommandDisplay_Render_EmptyCommand(t *testing.T) {
	c := CommandDisplay{
		Command:     "",
		Description: "Building...",
		Executing:   true,
	}

	rendered := c.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "Running:") {
		t.Errorf("Should contain Running:, got: %s", clean)
	}
	if !strings.Contains(clean, "Building...") {
		t.Errorf("Should show description, got: %s", clean)
	}
}

func TestCommandDisplay_Render_ColorMapping(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(oldProfile) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	exit0 := 0
	exit1 := 1

	success := CommandDisplay{Command: "ok", Executing: false, ExitCode: &exit0}
	errorDisplay := CommandDisplay{Command: "fail", Executing: false, ExitCode: &exit1}

	successRendered := success.Render()
	errorRendered := errorDisplay.Render()

	if successRendered == errorRendered {
		t.Error("Success and error should have different output (colors)")
	}
	if !strings.Contains(successRendered, "\x1b[") {
		t.Error("Success should have ANSI color codes")
	}
	if !strings.Contains(errorRendered, "\x1b[") {
		t.Error("Error should have ANSI color codes")
	}
}

func ptrDuration(d time.Duration) *time.Duration {
	return &d
}
