package flash

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var _ FlashAdapter = (*MP255Adapter)(nil)

func TestMP255Adapter_PlatformType(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	if adapter.PlatformType() != PlatformTypeMP255 {
		t.Errorf("PlatformType() = %v, want %v", adapter.PlatformType(), PlatformTypeMP255)
	}
}

func TestMP255Adapter_SupportsMethodSelection(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	if adapter.SupportsMethodSelection() != true {
		t.Error("SupportsMethodSelection() should be true for MP255")
	}
}

func TestMP255Adapter_GetMethods(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	methods := adapter.GetMethods()

	if len(methods) != 2 {
		t.Errorf("GetMethods() returned %d methods, want 2", len(methods))
	}

	expectedIDs := []string{"uuu", "manual"}
	for i, method := range methods {
		if method.ID != expectedIDs[i] {
			t.Errorf("Method %d ID = %v, want %v", i, method.ID, expectedIDs[i])
		}
	}
}

func TestMP255Adapter_GetSteps(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	steps := adapter.GetSteps("uuu")

	if len(steps) != 1 {
		t.Errorf("GetSteps() returned %d steps, want 1 (placeholder)", len(steps))
	}
	if len(steps) > 0 && steps[0].ID != "placeholder" {
		t.Errorf("GetSteps()[0].ID = %v, want placeholder", steps[0].ID)
	}
}

func TestMP255Adapter_ExecuteStep(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	output, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "placeholder"})

	if err == nil {
		t.Error("ExecuteStep should return error (not implemented)")
	}
	if !strings.Contains(output, "MP255 flashing not yet implemented") {
		t.Errorf("ExecuteStep output = %q, want message about not implemented", output)
	}
}

func TestMP255Adapter_CanFlash(t *testing.T) {
	t.Run("missing release directory", func(t *testing.T) {
		adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
		err := adapter.CanFlash("/nonexistent/release")
		if err == nil {
			t.Error("CanFlash should return error for missing release directory")
		}
		if !strings.Contains(err.Error(), "release directory not found") {
			t.Errorf("CanFlash error = %v, want 'release directory not found'", err)
		}
	})

	t.Run("missing deploy.sh", func(t *testing.T) {
		tmpDir := t.TempDir()
		releaseDir := filepath.Join(tmpDir, "release")
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			t.Fatal(err)
		}

		adapter := NewMP255Adapter("/nonexistent/root", "mp255-ulrpm", releaseDir)
		err := adapter.CanFlash(releaseDir)
		if err == nil {
			t.Error("CanFlash should return error for missing deploy.sh")
		}
		if !strings.Contains(err.Error(), "deploy.sh not found") {
			t.Errorf("CanFlash error = %v, want 'deploy.sh not found'", err)
		}
	})
}
