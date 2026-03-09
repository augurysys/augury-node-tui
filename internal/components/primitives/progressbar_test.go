package primitives

import (
	"strings"
	"testing"
)

func TestProgressBar_ZeroProgress(t *testing.T) {
	bar := ProgressBar{
		Current: 0,
		Total:   100,
		Width:   20,
		Label:   "Building",
	}

	rendered := bar.Render()

	if !strings.Contains(rendered, "Building") {
		t.Error("Should contain label")
	}
	if !strings.Contains(rendered, "0%") {
		t.Error("Should show 0%")
	}
}

func TestProgressBar_FullProgress(t *testing.T) {
	bar := ProgressBar{
		Current: 100,
		Total:   100,
		Width:   20,
		Label:   "Done",
	}

	rendered := bar.Render()

	if !strings.Contains(rendered, "100%") {
		t.Error("Should show 100%")
	}
	// Should be mostly filled
	if strings.Count(rendered, "█") < 15 {
		t.Errorf("Expected mostly filled bar, got: %s", stripAnsi(rendered))
	}
}

func TestProgressBar_PartialProgress(t *testing.T) {
	bar := ProgressBar{
		Current: 50,
		Total:   100,
		Width:   20,
		Label:   "Building",
	}

	rendered := bar.Render()

	if !strings.Contains(rendered, "50%") {
		t.Error("Should show 50%")
	}

	// Should have both filled and unfilled blocks
	hasFilled := strings.Contains(rendered, "█")
	hasUnfilled := strings.Contains(rendered, "░")

	if !hasFilled || !hasUnfilled {
		t.Errorf("Expected mix of filled/unfilled blocks, got: %s", stripAnsi(rendered))
	}
}

func TestProgressBar_FractionDisplay(t *testing.T) {
	bar := ProgressBar{
		Current: 820,
		Total:   1000,
		Width:   30,
		Label:   "Packages",
	}

	rendered := bar.Render()

	if !strings.Contains(rendered, "820") || !strings.Contains(rendered, "1000") {
		t.Errorf("Should show fraction (820/1000), got: %s", stripAnsi(rendered))
	}
}

func TestProgressBar_HandleZeroTotal(t *testing.T) {
	bar := ProgressBar{
		Current: 0,
		Total:   0,
		Width:   20,
		Label:   "Empty",
	}

	// Should not panic
	rendered := bar.Render()
	if rendered == "" {
		t.Error("Should render even with zero total")
	}
}

func TestProgressBar_NegativeCurrent(t *testing.T) {
	bar := ProgressBar{
		Current: -10, // pct=-10%, filledBlocks becomes negative -> strings.Repeat panics
		Total:   100,
		Width:   20,
		Label:   "Negative",
	}

	// Should not panic (strings.Repeat panics with negative count)
	rendered := bar.Render()
	if rendered == "" {
		t.Error("Should render even with negative current")
	}
}

func TestProgressBar_CurrentExceedsTotal(t *testing.T) {
	bar := ProgressBar{
		Current: 150,
		Total:   100,
		Width:   20,
		Label:   "Over",
	}

	// Should not panic
	rendered := bar.Render()
	if rendered == "" {
		t.Error("Should render even when current exceeds total")
	}

	// Should show actual percentage (>100%)
	if !strings.Contains(rendered, "150%") {
		t.Error("Should show actual percentage")
	}

	// Bar visual should be fully filled (clamped), not overflow
	filledCount := strings.Count(stripAnsi(rendered), "█")
	if filledCount == 0 {
		t.Error("Bar should be filled when current exceeds total")
	}
	if filledCount > 20 {
		t.Errorf("Bar should be clamped to width (20), got %d filled blocks", filledCount)
	}
}
