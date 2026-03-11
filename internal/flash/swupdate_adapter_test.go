package flash

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var _ FlashAdapter = (*SWUpdateAdapter)(nil)

// mustNewSWUpdateAdapter creates an adapter with a valid resolvable path for tests
// that don't need to validate image/augury_update. Creates a temp .swu file if basePath
// is empty.
func mustNewSWUpdateAdapter(t *testing.T, root, platform, basePath string) *SWUpdateAdapter {
	if basePath == "" {
		tmpDir := t.TempDir()
		swuPath := filepath.Join(tmpDir, "test.swu")
		if err := os.WriteFile(swuPath, []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}
		basePath = swuPath
	}
	adapter, err := NewSWUpdateAdapter(root, platform, basePath)
	if err != nil {
		t.Fatalf("NewSWUpdateAdapter() error = %v", err)
	}
	return adapter
}

func TestResolveSWUFile(t *testing.T) {
	t.Run("file ending in .swu returns as-is", func(t *testing.T) {
		tmpDir := t.TempDir()
		swuPath := filepath.Join(tmpDir, "image.swu")
		if err := os.WriteFile(swuPath, []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ResolveSWUFile(swuPath)
		if err != nil {
			t.Fatalf("ResolveSWUFile() error = %v, want nil", err)
		}
		if got != swuPath {
			t.Errorf("ResolveSWUFile() = %v, want %v", got, swuPath)
		}
	})

	t.Run("directory with no .swu files returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create directory with no .swu files
		if err := os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ResolveSWUFile(tmpDir)
		if err == nil {
			t.Error("ResolveSWUFile() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "no .swu files found in directory") {
			t.Errorf("ResolveSWUFile() error = %v, want 'no .swu files found in directory'", err)
		}
	})

	t.Run("directory with one .swu file returns that path", func(t *testing.T) {
		tmpDir := t.TempDir()
		swuPath := filepath.Join(tmpDir, "single.swu")
		if err := os.WriteFile(swuPath, []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ResolveSWUFile(tmpDir)
		if err != nil {
			t.Fatalf("ResolveSWUFile() error = %v, want nil", err)
		}
		if got != swuPath {
			t.Errorf("ResolveSWUFile() = %v, want %v", got, swuPath)
		}
	})

	t.Run("directory with multiple .swu files returns latest by ModTime", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldPath := filepath.Join(tmpDir, "old.swu")
		newPath := filepath.Join(tmpDir, "new.swu")
		if err := os.WriteFile(oldPath, []byte("old"), 0644); err != nil {
			t.Fatal(err)
		}
		// Set old.swu to have older ModTime
		past := time.Now().Add(-time.Hour)
		if err := os.Chtimes(oldPath, past, past); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(newPath, []byte("new"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ResolveSWUFile(tmpDir)
		if err != nil {
			t.Fatalf("ResolveSWUFile() error = %v, want nil", err)
		}
		if got != newPath {
			t.Errorf("ResolveSWUFile() = %v, want %v (latest by ModTime)", got, newPath)
		}
	})

	t.Run("invalid path returns error", func(t *testing.T) {
		_, err := ResolveSWUFile("/nonexistent/path/12345")
		if err == nil {
			t.Error("ResolveSWUFile() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "invalid path") {
			t.Errorf("ResolveSWUFile() error = %v, want 'invalid path'", err)
		}
	})
}

func TestSWUpdateAdapter_PlatformType(t *testing.T) {
	adapter := mustNewSWUpdateAdapter(t, "/tmp/root", "cassia-x2000", "")
	if adapter.PlatformType() != "swupdate" {
		t.Errorf("PlatformType() = %v, want %v", adapter.PlatformType(), "swupdate")
	}
}

func TestSWUpdateAdapter_SupportsMethodSelection(t *testing.T) {
	adapter := mustNewSWUpdateAdapter(t, "/tmp/root", "cassia-x2000", "")
	if adapter.SupportsMethodSelection() != false {
		t.Error("SupportsMethodSelection() should be false for SWUpdate")
	}
}

func TestSWUpdateAdapter_GetSteps(t *testing.T) {
	adapter := mustNewSWUpdateAdapter(t, "/tmp/root", "cassia-x2000", "")
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
	adapter := mustNewSWUpdateAdapter(t, "/tmp/root", "cassia-x2000", "")
	methods := adapter.GetMethods()
	if methods != nil {
		t.Errorf("GetMethods() = %v, want nil", methods)
	}
}

func TestSWUpdateAdapter_ExecuteStepReboot(t *testing.T) {
	adapter := mustNewSWUpdateAdapter(t, "/tmp/root", "cassia-x2000", "")
	output, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "reboot"})

	if err != nil {
		t.Errorf("ExecuteStep(reboot) error = %v, want nil", err)
	}
	if !strings.Contains(output, "Reboot required") {
		t.Errorf("ExecuteStep(reboot) output = %q, want message about reboot", output)
	}
}

func TestSWUpdateAdapter_ExecuteStepUnknown(t *testing.T) {
	adapter := mustNewSWUpdateAdapter(t, "/tmp/root", "cassia-x2000", "")
	_, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "unknown-step"})

	if err == nil {
		t.Error("ExecuteStep(unknown-step) should return error")
	}
	if !strings.Contains(err.Error(), "unknown step") {
		t.Errorf("ExecuteStep(unknown-step) error = %v, want 'unknown step'", err)
	}
}

func TestSWUpdateAdapter_CanFlash(t *testing.T) {
	t.Run("invalid path returns error from NewSWUpdateAdapter", func(t *testing.T) {
		_, err := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/nonexistent/image.swu")
		if err == nil {
			t.Error("NewSWUpdateAdapter should return error for nonexistent path")
		}
		if !strings.Contains(err.Error(), "invalid path") {
			t.Errorf("NewSWUpdateAdapter error = %v, want 'invalid path'", err)
		}
	})

	t.Run("missing augury_update", func(t *testing.T) {
		tmpDir := t.TempDir()
		imagePath := filepath.Join(tmpDir, "test.swu")
		if err := os.WriteFile(imagePath, []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}

		adapter, err := NewSWUpdateAdapter("/nonexistent/root", "cassia-x2000", imagePath)
		if err != nil {
			t.Fatalf("NewSWUpdateAdapter() error = %v", err)
		}
		err = adapter.CanFlash("")
		if err == nil {
			t.Error("CanFlash should return error for missing augury_update")
		}
		if !strings.Contains(err.Error(), "augury_update not found") {
			t.Errorf("CanFlash error = %v, want 'augury_update not found'", err)
		}
	})
}

func TestNewSWUpdateAdapter_ResolvesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	swuPath := filepath.Join(tmpDir, "firmware.swu")
	if err := os.WriteFile(swuPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	adapter, err := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", tmpDir)
	if err != nil {
		t.Fatalf("NewSWUpdateAdapter() error = %v", err)
	}
	// Adapter should have resolved to the .swu file
	if adapter.imagePath != swuPath {
		t.Errorf("adapter.imagePath = %v, want %v", adapter.imagePath, swuPath)
	}
}
