package validations

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/status"
)

func TestValidationsModel_PresetsMapToExactCommands(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}

	m := NewModel(st)

	presets := []string{"all", "shellcheck-only", "bats-only", "parse-test-only"}
	for _, preset := range presets {
		cmd, ok := m.CommandForPreset(preset)
		if !ok {
			t.Logf("preset %q returns not available when script missing", preset)
			continue
		}
		if cmd.Root != root {
			t.Errorf("preset %q: cmd.Root = %q, want %q", preset, cmd.Root, root)
		}
		if cmd.Command == "" {
			t.Errorf("preset %q: cmd.Command must not be empty", preset)
		}
	}
}

func TestValidationsModel_UnknownPresetReturnsNotAvailable(t *testing.T) {
	st := status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}
	m := NewModel(st)

	_, ok := m.CommandForPreset("unknown-preset")
	if ok {
		t.Error("unknown preset must return ok=false")
	}
}

func TestValidationsModel_ViewListsPresets(t *testing.T) {
	st := status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}
	m := NewModel(st)

	view := m.View()
	if view == "" {
		t.Fatal("View must not be empty")
	}
	for _, p := range []string{"all", "shellcheck", "bats", "parse-test"} {
		if !strings.Contains(strings.ToLower(view), strings.ToLower(p)) {
			t.Errorf("View must mention preset %q; got %q", p, view)
		}
	}
}
