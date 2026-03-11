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
		t.Errorf("GetSteps() returned %d steps, want 1", len(steps))
	}
	if len(steps) > 0 && steps[0].ID != "flash" {
		t.Errorf("GetSteps()[0].ID = %v, want flash", steps[0].ID)
	}
	if len(steps) > 0 && !strings.Contains(steps[0].Description, "deploy.sh") {
		t.Errorf("GetSteps()[0].Description = %q, want to contain deploy.sh", steps[0].Description)
	}
}

func TestMP255Adapter_ExecuteStep(t *testing.T) {
	t.Run("runs deploy.sh with correct args and captures output", func(t *testing.T) {
		tmpDir := t.TempDir()
		yoctoDir := filepath.Join(tmpDir, "yocto")
		if err := os.MkdirAll(yoctoDir, 0755); err != nil {
			t.Fatal(err)
		}
		// Test script: receives --method 1 /path/to/release, echoes METHOD and RELEASE
		deployScript := `#!/bin/bash
echo "METHOD=$2"
echo "RELEASE=$3"
exit 0
`
		deployPath := filepath.Join(yoctoDir, "deploy.sh")
		if err := os.WriteFile(deployPath, []byte(deployScript), 0755); err != nil {
			t.Fatal(err)
		}
		releaseDir := filepath.Join(tmpDir, "release")
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			t.Fatal(err)
		}

		adapter := NewMP255Adapter(tmpDir, "mp255-ulrpm", releaseDir)
		// Set method via GetSteps (adapter stores it)
		adapter.GetSteps("uuu")
		output, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "flash"})

		if err != nil {
			t.Errorf("ExecuteStep error = %v, want nil", err)
		}
		if !strings.Contains(output, "METHOD=1") {
			t.Errorf("ExecuteStep output = %q, want to contain METHOD=1 (uuu -> 1)", output)
		}
		if !strings.Contains(output, "RELEASE="+releaseDir) {
			t.Errorf("ExecuteStep output = %q, want to contain RELEASE=%s", output, releaseDir)
		}
	})

	t.Run("manual method maps to 2", func(t *testing.T) {
		tmpDir := t.TempDir()
		yoctoDir := filepath.Join(tmpDir, "yocto")
		if err := os.MkdirAll(yoctoDir, 0755); err != nil {
			t.Fatal(err)
		}
		deployScript := `#!/bin/bash
echo "METHOD=$2"
exit 0
`
		if err := os.WriteFile(filepath.Join(yoctoDir, "deploy.sh"), []byte(deployScript), 0755); err != nil {
			t.Fatal(err)
		}
		releaseDir := filepath.Join(tmpDir, "release")
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			t.Fatal(err)
		}

		adapter := NewMP255Adapter(tmpDir, "mp255-ulrpm", releaseDir)
		adapter.GetSteps("manual")
		output, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "flash"})

		if err != nil {
			t.Errorf("ExecuteStep error = %v, want nil", err)
		}
		if !strings.Contains(output, "METHOD=2") {
			t.Errorf("ExecuteStep output = %q, want to contain METHOD=2 (manual -> 2)", output)
		}
	})

	t.Run("propagates script exit error", func(t *testing.T) {
		tmpDir := t.TempDir()
		yoctoDir := filepath.Join(tmpDir, "yocto")
		if err := os.MkdirAll(yoctoDir, 0755); err != nil {
			t.Fatal(err)
		}
		deployScript := `#!/bin/bash
echo "script failed"
exit 1
`
		if err := os.WriteFile(filepath.Join(yoctoDir, "deploy.sh"), []byte(deployScript), 0755); err != nil {
			t.Fatal(err)
		}
		releaseDir := filepath.Join(tmpDir, "release")
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			t.Fatal(err)
		}

		adapter := NewMP255Adapter(tmpDir, "mp255-ulrpm", releaseDir)
		adapter.GetSteps("uuu")
		output, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "flash"})

		if err == nil {
			t.Error("ExecuteStep should return error when script exits 1")
		}
		if !strings.Contains(output, "script failed") {
			t.Errorf("ExecuteStep output = %q, want to contain script output", output)
		}
	})

	t.Run("unknown step returns error", func(t *testing.T) {
		adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
		_, err := adapter.ExecuteStep(context.Background(), FlashStep{ID: "unknown-step"})

		if err == nil {
			t.Error("ExecuteStep(unknown-step) should return error")
		}
		if !strings.Contains(err.Error(), "unknown step") {
			t.Errorf("ExecuteStep error = %v, want 'unknown step'", err)
		}
	})

	t.Run("context cancellation stops execution", func(t *testing.T) {
		tmpDir := t.TempDir()
		yoctoDir := filepath.Join(tmpDir, "yocto")
		if err := os.MkdirAll(yoctoDir, 0755); err != nil {
			t.Fatal(err)
		}
		// Script that sleeps so we can cancel
		deployScript := `#!/bin/bash
sleep 60
exit 0
`
		if err := os.WriteFile(filepath.Join(yoctoDir, "deploy.sh"), []byte(deployScript), 0755); err != nil {
			t.Fatal(err)
		}
		releaseDir := filepath.Join(tmpDir, "release")
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		adapter := NewMP255Adapter(tmpDir, "mp255-ulrpm", releaseDir)
		adapter.GetSteps("uuu")
		_, err := adapter.ExecuteStep(ctx, FlashStep{ID: "flash"})

		if err == nil {
			t.Error("ExecuteStep should return error when context is cancelled")
		}
		if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "killed") {
			t.Errorf("ExecuteStep error = %v, want context-related error", err)
		}
	})
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
