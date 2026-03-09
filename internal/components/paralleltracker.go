package components

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// BuildLane represents a single concurrent build lane (platform)
type BuildLane struct {
	Platform string
	Progress float64 // 0.0 to 1.0
	Status   primitives.Status
	Current  string // Current package/step
}

// ParallelTracker displays concurrent build lanes (parallel progress tracks)
type ParallelTracker struct {
	Lanes  []BuildLane
	Width  int
	Height int
}

const (
	barWidthBlocks = 10
	platformWidth  = 14
	currentWidth   = 24
)

// Render produces the styled parallel tracker output
func (p ParallelTracker) Render() string {
	if len(p.Lanes) == 0 {
		return ""
	}

	palette := styles.DefaultPalette()
	typo := styles.DefaultTypography()

	// Limit visible lanes by Height
	visibleCount := len(p.Lanes)
	if p.Height > 0 && visibleCount > p.Height {
		visibleCount = p.Height
	}

	var lines []string
	for i := 0; i < visibleCount; i++ {
		lane := p.Lanes[i]
		line := renderLane(lane, p.Width, palette, typo)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func renderLane(lane BuildLane, maxWidth int, palette styles.Palette, typo styles.Typography) string {
	icon, iconColor := statusIconAndColor(lane.Status, palette)

	// Allocate width: icon(2) + platform + space(1) + bar + pct(5) + space(1) + current
	const iconW, pctW, spaceW = 2, 5, 2
	remaining := maxWidth - iconW - spaceW - pctW - spaceW
	if remaining < 10 {
		remaining = 40 // fallback
	}
	platformW := platformWidth
	barW := barWidthBlocks
	currentW := currentWidth
	if platformW+barW+currentW > remaining {
		// Scale down: prefer bar, then platform, then current
		barW = remaining / 3
		if barW < 5 {
			barW = 5
		}
		platformW = (remaining - barW) / 2
		currentW = remaining - barW - platformW
		if currentW < 5 {
			currentW = 5
			platformW = remaining - barW - currentW
		}
	}

	platform := truncateToWidth(lane.Platform, platformW)

	if lane.Status == primitives.StatusBlocked || lane.Status == primitives.StatusUnavailable {
		// Queued: "□ platform-name queued"
		iconPart := lipgloss.NewStyle().Foreground(lipgloss.Color(iconColor)).Render(icon + " ")
		platformPart := typo.Body.Render(platform)
		queuedPart := typo.Dim.Render(" queued")
		line := iconPart + platformPart + queuedPart
		return truncateLineToWidth(line, maxWidth)
	}

	// Running: "▶ platform  ████████░░ 82%  current-pkg"
	pct := lane.Progress * 100
	filledBlocks := int(float64(barW) * lane.Progress)
	if filledBlocks < 0 {
		filledBlocks = 0
	}
	if filledBlocks > barW {
		filledBlocks = barW
	}
	filled := strings.Repeat("█", filledBlocks)
	unfilled := strings.Repeat("░", barW-filledBlocks)

	// Progress bar color (sequential: <50% dim, 50-80% yellow, >80% green)
	var barColor string
	if pct < 50 {
		barColor = palette.Overlay0
	} else if pct < 80 {
		barColor = palette.Warning
	} else {
		barColor = palette.Success
	}
	barPart := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render(filled + unfilled)

	iconPart := lipgloss.NewStyle().Foreground(lipgloss.Color(iconColor)).Render(icon + " ")
	platformPart := typo.Body.Render(platform)
	pctPart := typo.Body.Render(fmt.Sprintf(" %.0f%%", pct))
	currentPart := typo.Dim.Render(" " + truncateToWidth(lane.Current, currentW))

	line := iconPart + platformPart + " " + barPart + pctPart + currentPart
	return truncateLineToWidth(line, maxWidth)
}

// truncateLineToWidth truncates a line (possibly with ANSI) to max display width
func truncateLineToWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	ellipsisW := runewidth.StringWidth("…")
	maxVisible := width - ellipsisW

	var visibleWidth, bytePos int
	for bytePos < len(s) {
		// Skip ANSI escape sequences
		if s[bytePos] == '\x1b' && bytePos+1 < len(s) && s[bytePos+1] == '[' {
			bytePos += 2
			for bytePos < len(s) && s[bytePos] != 'm' {
				bytePos++
			}
			if bytePos < len(s) {
				bytePos++
			}
			continue
		}

		r, size := utf8.DecodeRuneInString(s[bytePos:])
		if size == 0 {
			break
		}
		w := runewidth.RuneWidth(r)
		if visibleWidth+w > maxVisible {
			return s[:bytePos] + "…"
		}
		visibleWidth += w
		bytePos += size
	}
	return s
}

func statusIconAndColor(status primitives.Status, palette styles.Palette) (string, string) {
	switch status {
	case primitives.StatusRunning:
		return "▶", palette.Info
	case primitives.StatusBlocked, primitives.StatusUnavailable:
		return "□", palette.Overlay0
	default:
		return "▶", palette.Overlay0
	}
}

func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	// Strip ANSI for width calculation
	displayWidth := runewidth.StringWidth(stripAnsiForWidth(s))
	if displayWidth <= width {
		return s
	}
	return runewidth.Truncate(s, width-1, "…")
}

// stripAnsiForWidth removes ANSI codes for runewidth calculation
func stripAnsiForWidth(s string) string {
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
