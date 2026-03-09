package components

import (
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/ansi"
	"github.com/mattn/go-runewidth"
)

func TestMetricsBar_Render_AllMetrics(t *testing.T) {
	m := MetricsBar{
		CPU:        0.82,
		Memory:     0.65,
		Disk:       0.48,
		HotProcess: "gcc (3 threads)",
		Width:      80,
	}

	rendered := m.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "CPU") || !strings.Contains(clean, "MEM") || !strings.Contains(clean, "DISK") {
		t.Errorf("Should contain CPU, MEM, DISK labels, got: %s", clean)
	}
	if !strings.Contains(clean, "82%") || !strings.Contains(clean, "65%") || !strings.Contains(clean, "48%") {
		t.Errorf("Should show percentages for all metrics, got: %s", clean)
	}
	if !strings.Contains(clean, "gcc") {
		t.Errorf("Should show hot process, got: %s", clean)
	}
	if !strings.Contains(clean, "█") || !strings.Contains(clean, "░") {
		t.Errorf("Should have progress bars (█░), got: %s", clean)
	}
}

func TestMetricsBar_Render_ZeroMetrics(t *testing.T) {
	m := MetricsBar{
		CPU:        0,
		Memory:     0,
		Disk:       0,
		HotProcess: "",
		Width:      80,
	}

	rendered := m.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "0%") {
		t.Errorf("Zero metrics should show 0%%, got: %s", clean)
	}
	// Bar should be all unfilled
	unfilled := strings.Count(clean, "░")
	if unfilled == 0 {
		t.Errorf("Zero metrics should have unfilled blocks, got: %s", clean)
	}
}

func TestMetricsBar_Render_FullMetrics(t *testing.T) {
	m := MetricsBar{
		CPU:        1.0,
		Memory:    1.0,
		Disk:      1.0,
		HotProcess: "stress",
		Width:     80,
	}

	rendered := m.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "100%") {
		t.Errorf("Full metrics should show 100%%, got: %s", clean)
	}
	filled := strings.Count(clean, "█")
	if filled < 3 {
		t.Errorf("Full metrics should have filled bars, got: %s", clean)
	}
}

func TestMetricsBar_Render_SequentialColors_Under50(t *testing.T) {
	// 48% should use dim (Overlay0) - we can't easily assert color, but we verify bar renders
	m := MetricsBar{
		CPU:        0.48,
		Memory:     0.48,
		Disk:       0.48,
		HotProcess: "idle",
		Width:      80,
	}

	rendered := m.Render()
	if rendered == "" {
		t.Error("Should render with values under 50%")
	}
	if !strings.Contains(ansi.StripAnsi(rendered), "48%") {
		t.Error("Should show 48%")
	}
}

func TestMetricsBar_Render_SequentialColors_50to80(t *testing.T) {
	m := MetricsBar{
		CPU:        0.65,
		Memory:     0.65,
		Disk:       0.65,
		HotProcess: "build",
		Width:      80,
	}

	rendered := m.Render()
	if rendered == "" {
		t.Error("Should render with values 50-80%")
	}
	if !strings.Contains(ansi.StripAnsi(rendered), "65%") {
		t.Error("Should show 65%")
	}
}

func TestMetricsBar_Render_SequentialColors_Over80(t *testing.T) {
	m := MetricsBar{
		CPU:        0.95,
		Memory:     0.95,
		Disk:       0.95,
		HotProcess: "gcc",
		Width:      80,
	}

	rendered := m.Render()
	if rendered == "" {
		t.Error("Should render with values over 80%")
	}
	if !strings.Contains(ansi.StripAnsi(rendered), "95%") {
		t.Error("Should show 95%")
	}
}

func TestMetricsBar_Render_WidthResponsive(t *testing.T) {
	m := MetricsBar{
		CPU:        0.82,
		Memory:     0.65,
		Disk:       0.48,
		HotProcess: "very-long-process-name-that-exceeds-width",
		Width:      40,
	}

	rendered := m.Render()
	clean := ansi.StripAnsi(rendered)
	displayWidth := runewidth.StringWidth(clean)

	if displayWidth > 42 {
		t.Errorf("Output should respect Width, got %d display width: %s", displayWidth, clean)
	}
}

func TestMetricsBar_Render_EmptyHotProcess(t *testing.T) {
	m := MetricsBar{
		CPU:        0.5,
		Memory:     0.5,
		Disk:       0.5,
		HotProcess: "",
		Width:      80,
	}

	rendered := m.Render()
	clean := ansi.StripAnsi(rendered)

	if !strings.Contains(clean, "Hot") {
		t.Errorf("Should contain Hot label even when empty, got: %s", clean)
	}
}

func TestMetricsBar_Render_ClampsValues(t *testing.T) {
	m := MetricsBar{
		CPU:        1.5,
		Memory:    -0.1,
		Disk:       2.0,
		HotProcess: "test",
		Width:      80,
	}

	rendered := m.Render()
	if rendered == "" {
		t.Error("Should not panic with out-of-range values")
	}
	// Should not have more filled blocks than bar width allows
	clean := ansi.StripAnsi(rendered)
	lines := strings.Split(clean, " ")
	for _, seg := range lines {
		filled := strings.Count(seg, "█")
		if filled > 10 {
			t.Errorf("Bar should be clamped, segment has %d filled: %s", filled, seg)
		}
	}
}

func TestMetricsBar_Render_OutputFormat(t *testing.T) {
	m := MetricsBar{
		CPU:        0.82,
		Memory:     0.65,
		Disk:       0.48,
		HotProcess: "gcc (3 threads)",
		Width:      120,
	}

	rendered := m.Render()
	clean := ansi.StripAnsi(rendered)

	// Expected format: CPU: ████░ 82%  MEM: ███░░ 65%  DISK: ██░░░ 48%  Hot: gcc (3 threads)
	if !strings.Contains(clean, "CPU:") {
		t.Errorf("Should start with CPU:, got: %s", clean)
	}
	if !strings.Contains(clean, "MEM:") {
		t.Errorf("Should contain MEM:, got: %s", clean)
	}
	if !strings.Contains(clean, "DISK:") {
		t.Errorf("Should contain DISK:, got: %s", clean)
	}
	if !strings.Contains(clean, "Hot:") {
		t.Errorf("Should contain Hot:, got: %s", clean)
	}
	if !strings.Contains(clean, "gcc (3 threads)") {
		t.Errorf("Should show full hot process, got: %s", clean)
	}
}
