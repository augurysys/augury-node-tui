package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckNixInstalled(t *testing.T) {
	result := CheckNixInstalled()
	if !result.Available {
		t.Skip("Nix not installed, skipping test")
	}
	if result.Error != nil {
		t.Errorf("Nix installed but got error: %v", result.Error)
	}
}

func TestCheckNixExperimentalFeatures(t *testing.T) {
	if !CheckNixInstalled().Available {
		t.Skip("Nix not installed")
	}
	result := CheckNixExperimentalFeatures()
	if !result.Available {
		t.Fatalf("Experimental features not enabled: %s", result.Message)
	}
	if result.Error != nil {
		t.Errorf("CheckNixExperimentalFeatures failed: %v", result.Error)
	}
}

func TestCheckNixGroup(t *testing.T) {
	result := CheckNixGroup()
	if !result.Available {
		t.Fatalf("User not in nix-users group: %s", result.Message)
	}
	if result.Error != nil {
		t.Errorf("CheckNixGroup failed: %v", result.Error)
	}
}

func TestCheckDaemonSocket(t *testing.T) {
	if !CheckNixInstalled().Available {
		t.Skip("Nix not installed")
	}
	result := CheckDaemonSocket()
	if !result.Available && result.Error == nil {
		t.Skip("Nix daemon not running (common in CI)")
	}
	if !result.Available {
		t.Fatalf("Daemon socket not accessible: %s", result.Message)
	}
	if result.Error != nil {
		t.Errorf("CheckDaemonSocket failed: %v", result.Error)
	}
}

func TestAutoFixNixConfig(t *testing.T) {
	// Save original HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Use temp dir as HOME
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)

	// Run auto-fix
	if err := AutoFixNixConfig(); err != nil {
		t.Fatalf("AutoFixNixConfig failed: %v", err)
	}

	// Check config was created
	confPath := filepath.Join(tempHome, ".config", "nix", "nix.conf")
	data, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "nix-command") {
		t.Error("Config should contain 'nix-command'")
	}
	if !strings.Contains(content, "flakes") {
		t.Error("Config should contain 'flakes'")
	}
}
