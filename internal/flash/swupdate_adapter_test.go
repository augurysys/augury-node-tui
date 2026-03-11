package flash

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var _ FlashAdapter = (*SWUpdateAdapter)(nil)

func TestSWUpdateAdapter_PlatformType(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	if adapter.PlatformType() != "swupdate" {
		t.Errorf("PlatformType() = %v, want %v", adapter.PlatformType(), "swupdate")
	}
}

func TestSWUpdateAdapter_SupportsMethodSelection(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	if adapter.SupportsMethodSelection() != false {
		t.Error("SupportsMethodSelection() should be false for SWUpdate")
	}
}

func TestSWUpdateAdapter_GetSteps(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	steps := adapter.GetSteps("")

	if len(steps) != 3 {
		t.Errorf("GetSteps() returned %d steps, want 3", len(steps))
	}

	// Verify step IDs
	expectedIDs := []string{"verify", "flash", "reboot"}
	for i, step := range steps {
		if step.ID != expectedIDs[i] {
			t.Errorf("Step %d ID = %v, want %v", i, step.ID, expectedIDs[i])
		}
	}
}

func TestSWUpdateAdapter_GetMethods(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	methods := adapter.GetMethods()
	if methods != nil {
		t.Errorf("GetMethods() = %v, want nil", methods)
	}
}

func TestSWUpdateAdapter_ExecuteStepReboot(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	output, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "reboot"})

	if err != nil {
		t.Errorf("ExecuteStep(reboot) error = %v, want nil", err)
	}
	if !strings.Contains(output, "Reboot required") {
		t.Errorf("ExecuteStep(reboot) output = %q, want message about reboot", output)
	}
}

func TestSWUpdateAdapter_ExecuteStepUnknown(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	_, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "unknown-step"})

	if err == nil {
		t.Error("ExecuteStep(unknown-step) should return error")
	}
	if !strings.Contains(err.Error(), "unknown step") {
		t.Errorf("ExecuteStep(unknown-step) error = %v, want 'unknown step'", err)
	}
}

func TestSWUpdateAdapter_CanFlash(t *testing.T) {
	// We cannot test success case without creating real files,
	// so we test error cases only

	t.Run("missing image", func(t *testing.T) {
		adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/nonexistent/image.swu")
		err := adapter.CanFlash("")
		if err == nil {
			t.Error("CanFlash should return error for missing image")
		}
		if !strings.Contains(err.Error(), "image file not found") {
			t.Errorf("CanFlash error = %v, want 'image file not found'", err)
		}
	})

	t.Run("missing augury_update", func(t *testing.T) {
		// Create temp image file
		tmpDir := t.TempDir()
		imagePath := filepath.Join(tmpDir, "test.swu")
		if err := os.WriteFile(imagePath, []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}

		// Root doesn't have augury_update
		adapter := NewSWUpdateAdapter("/nonexistent/root", "cassia-x2000", imagePath)
		err := adapter.CanFlash("")
		if err == nil {
			t.Error("CanFlash should return error for missing augury_update")
		}
		if !strings.Contains(err.Error(), "augury_update not found") {
			t.Errorf("CanFlash error = %v, want 'augury_update not found'", err)
		}
	})
}
