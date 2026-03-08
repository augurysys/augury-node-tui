package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestConfig_ReadWriteRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := Config{
		AuguryNodeRoot:   "/home/user/augury-node",
		BinaryInstalled:  true,
		NixVerified:      true,
		SetupCompletedAt:  "2026-03-08T19:45:00Z",
		CompletedSteps:   []string{"root", "nix"},
		SkippedSteps:     []string{"groups"},
	}

	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	loaded, err := Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !reflect.DeepEqual(cfg, loaded) {
		t.Errorf("config mismatch:\nwant: %+v\ngot:  %+v", cfg, loaded)
	}
}

func TestConfig_DefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath failed: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Error("DefaultPath should return absolute path")
	}
	if !strings.Contains(path, ".config/augury-node-tui") {
		t.Errorf("DefaultPath should be in .config/augury-node-tui; got %q", path)
	}
}

func TestConfig_ReadNonexistent(t *testing.T) {
	_, err := Read("/nonexistent/config.toml")
	if err == nil {
		t.Error("Read should return error for nonexistent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Read should return os.IsNotExist error; got %v", err)
	}
}
