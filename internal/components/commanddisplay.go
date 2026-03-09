package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

// CommandDisplay shows actual commands being executed (lazygit pattern) for command transparency.
type CommandDisplay struct {
	Command     string
	Description string
	Executing   bool
	ExitCode    *int
	Duration    *time.Duration
}

// Render produces the styled command display output.
// While executing: "Running: <command>\n[●] <description>"
// After completion: "✓ <command> (exit 0, 2m34s)" or "✗ <command> (exit 1, 5s)"
func (c CommandDisplay) Render() string {
	palette := styles.DefaultPalette()

	if c.Executing {
		return c.renderExecuting(palette)
	}
	return c.renderComplete(palette)
}

func (c CommandDisplay) renderExecuting(palette styles.Palette) string {
	runningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Info))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Info))

	line1 := runningStyle.Render("Running: " + c.Command)
	desc := c.Description
	if desc == "" {
		desc = "..."
	}
	line2 := descStyle.Render("[●] " + desc)
	return line1 + "\n" + line2
}

func (c CommandDisplay) renderComplete(palette styles.Palette) string {
	var icon string
	var color string

	if c.ExitCode == nil {
		icon = "?"
		color = palette.Overlay0
	} else if *c.ExitCode == 0 {
		icon = "✓"
		color = palette.Success
	} else {
		icon = "✗"
		color = palette.Error
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	// Format: "✓ command (exit 0, 2m34s)" or "✓ command (exit 0)" when no duration
	suffix := c.formatSuffix()
	text := icon + " " + c.Command
	if suffix != "" {
		text += " " + suffix
	}
	return style.Render(text)
}

func (c CommandDisplay) formatSuffix() string {
	if c.ExitCode == nil {
		return ""
	}
	parts := []string{fmt.Sprintf("exit %d", *c.ExitCode)}
	if c.Duration != nil && *c.Duration > 0 {
		parts = append(parts, c.formatDuration(*c.Duration))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (c CommandDisplay) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return d.Round(time.Millisecond).String()
}
