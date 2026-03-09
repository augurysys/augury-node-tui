package components

import (
	"regexp"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/mattn/go-runewidth"
)

// stripAnsi removes ANSI escape sequences for content assertions
func stripAnsi(s string) string {
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

func TestParallelTracker_EmptyLanes(t *testing.T) {
	p := ParallelTracker{
		Lanes:  []BuildLane{},
		Width:  80,
		Height: 10,
	}

	rendered := p.Render()

	if rendered != "" {
		t.Errorf("Empty lanes should render empty string, got: %q", rendered)
	}
}

func TestParallelTracker_SingleRunningLane(t *testing.T) {
	p := ParallelTracker{
		Lanes: []BuildLane{
			{
				Platform: "node2",
				Progress: 0.82,
				Status:   primitives.StatusRunning,
				Current:  "gcc-wrapper-13.2.0",
			},
		},
		Width:  80,
		Height: 10,
	}

	rendered := p.Render()
	clean := stripAnsi(rendered)

	if !strings.Contains(clean, "node2") {
		t.Errorf("Should contain platform name, got: %s", clean)
	}
	if !strings.Contains(clean, "82%") {
		t.Errorf("Should show 82%% progress, got: %s", clean)
	}
	if !strings.Contains(clean, "gcc-wrapper-13.2.0") {
		t.Errorf("Should show current package, got: %s", clean)
	}
	if !strings.Contains(clean, "▶") {
		t.Errorf("Running lane should have ▶ icon, got: %s", clean)
	}
	if !strings.Contains(clean, "█") || !strings.Contains(clean, "░") {
		t.Errorf("Should have progress bar (█░), got: %s", clean)
	}
}

func TestParallelTracker_SingleQueuedLane(t *testing.T) {
	p := ParallelTracker{
		Lanes: []BuildLane{
			{
				Platform: "cassia-x2000",
				Progress: 0,
				Status:   primitives.StatusBlocked,
				Current:  "",
			},
		},
		Width:  80,
		Height: 10,
	}

	rendered := p.Render()
	clean := stripAnsi(rendered)

	if !strings.Contains(clean, "cassia-x2000") {
		t.Errorf("Should contain platform name, got: %s", clean)
	}
	if !strings.Contains(clean, "queued") {
		t.Errorf("Queued lane should show 'queued', got: %s", clean)
	}
	if !strings.Contains(clean, "□") {
		t.Errorf("Queued lane should have □ icon, got: %s", clean)
	}
}

func TestParallelTracker_MultipleLanes(t *testing.T) {
	p := ParallelTracker{
		Lanes: []BuildLane{
			{
				Platform: "node2",
				Progress: 0.82,
				Status:   primitives.StatusRunning,
				Current:  "gcc-wrapper-13.2.0",
			},
			{
				Platform: "moxa-uc3100",
				Progress: 0.32,
				Status:   primitives.StatusRunning,
				Current:  "glibc-2.38",
			},
			{
				Platform: "cassia-x2000",
				Progress: 0,
				Status:   primitives.StatusBlocked,
				Current:  "",
			},
		},
		Width:  80,
		Height: 10,
	}

	rendered := p.Render()
	clean := stripAnsi(rendered)

	// All platforms present
	if !strings.Contains(clean, "node2") || !strings.Contains(clean, "moxa-uc3100") || !strings.Contains(clean, "cassia-x2000") {
		t.Errorf("Should contain all platform names, got: %s", clean)
	}
	// Running lanes have progress
	if !strings.Contains(clean, "82%") || !strings.Contains(clean, "32%") {
		t.Errorf("Running lanes should show percentages, got: %s", clean)
	}
	// Queued lane
	if !strings.Contains(clean, "queued") {
		t.Errorf("Queued lane should show 'queued', got: %s", clean)
	}
	// One line per lane
	lines := strings.Split(strings.TrimSpace(rendered), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (one per lane), got %d: %v", len(lines), lines)
	}
}

func TestParallelTracker_ProgressBarStyling(t *testing.T) {
	// 82% should be mostly filled (green zone)
	p := ParallelTracker{
		Lanes: []BuildLane{
			{
				Platform: "test",
				Progress: 0.82,
				Status:   primitives.StatusRunning,
				Current:  "pkg",
			},
		},
		Width:  80,
		Height: 10,
	}

	rendered := p.Render()
	// Should have more filled than unfilled blocks at 82%
	filled := strings.Count(stripAnsi(rendered), "█")
	unfilled := strings.Count(stripAnsi(rendered), "░")
	if filled < unfilled {
		t.Errorf("82%% progress should have more filled than unfilled blocks (filled=%d, unfilled=%d)", filled, unfilled)
	}
}

func TestParallelTracker_ZeroProgress(t *testing.T) {
	p := ParallelTracker{
		Lanes: []BuildLane{
			{
				Platform: "test",
				Progress: 0,
				Status:   primitives.StatusRunning,
				Current:  "first-pkg",
			},
		},
		Width:  80,
		Height: 10,
	}

	rendered := p.Render()
	clean := stripAnsi(rendered)

	if !strings.Contains(clean, "0%") {
		t.Errorf("Zero progress should show 0%%, got: %s", clean)
	}
	// Bar should be all unfilled
	unfilled := strings.Count(clean, "░")
	if unfilled == 0 {
		t.Errorf("Zero progress should have unfilled blocks, got: %s", clean)
	}
}

func TestParallelTracker_WidthResponsive(t *testing.T) {
	p := ParallelTracker{
		Lanes: []BuildLane{
			{
				Platform: "very-long-platform-name",
				Progress: 0.5,
				Status:   primitives.StatusRunning,
				Current:  "very-long-current-package-name",
			},
		},
		Width:  40,
		Height: 10,
	}

	rendered := p.Render()
	lines := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")

	for _, line := range lines {
		clean := stripAnsi(line)
		displayWidth := runewidth.StringWidth(clean)
		if displayWidth > 42 { // Width 40 + small slack
			t.Errorf("Line should respect width, got %d display width: %s", displayWidth, clean)
		}
	}
}

func TestParallelTracker_HeightLimitsVisibleLanes(t *testing.T) {
	lanes := make([]BuildLane, 10)
	for i := 0; i < 10; i++ {
		lanes[i] = BuildLane{
			Platform: "platform-" + string(rune('a'+i)),
			Progress: 0.5,
			Status:   primitives.StatusRunning,
			Current:  "pkg",
		}
	}

	p := ParallelTracker{
		Lanes:  lanes,
		Width:  80,
		Height: 3,
	}

	rendered := p.Render()
	lines := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")

	if len(lines) > 3 {
		t.Errorf("Height 3 should limit to 3 lines, got %d", len(lines))
	}
}

func TestParallelTracker_StatusIconMapping(t *testing.T) {
	tests := []struct {
		status       primitives.Status
		expectedIcon string
	}{
		{primitives.StatusRunning, "▶"},
		{primitives.StatusBlocked, "□"},
		{primitives.StatusUnavailable, "□"},
	}

	for _, tt := range tests {
		p := ParallelTracker{
			Lanes: []BuildLane{
				{
					Platform: "test",
					Progress: 0,
					Status:   tt.status,
					Current:  "",
				},
			},
			Width:  80,
			Height: 10,
		}

		rendered := p.Render()
		clean := stripAnsi(rendered)

		if !strings.Contains(clean, tt.expectedIcon) {
			t.Errorf("Status %v should show icon %q, got: %s", tt.status, tt.expectedIcon, clean)
		}
	}
}

func TestParallelTracker_OutputFormat(t *testing.T) {
	p := ParallelTracker{
		Lanes: []BuildLane{
			{
				Platform: "node2",
				Progress: 0.82,
				Status:   primitives.StatusRunning,
				Current:  "gcc-wrapper-13.2.0",
			},
		},
		Width:  80,
		Height: 10,
	}

	rendered := p.Render()
	clean := stripAnsi(rendered)

	// Expected format: "▶ node2       ████████░░ 82%  gcc-wrapper-13.2.0"
	// Check order: icon, platform, bar, pct, current
	iconMatch := regexp.MustCompile(`^▶\s+node2\s+`)
	if !iconMatch.MatchString(clean) {
		t.Errorf("Line should start with '▶ node2 ', got: %s", clean)
	}
	if !regexp.MustCompile(`\d+%`).MatchString(clean) {
		t.Errorf("Should contain percentage, got: %s", clean)
	}
	if !strings.Contains(clean, "gcc-wrapper-13.2.0") {
		t.Errorf("Should contain current package, got: %s", clean)
	}
}
