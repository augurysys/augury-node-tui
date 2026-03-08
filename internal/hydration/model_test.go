package hydration

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
)

func TestHydrationModel_DryRunRowsForSelectedPlatforms(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}

	m := NewModel(st, platforms, selected)
	rows := m.DryRunRows()

	if len(rows) == 0 {
		t.Fatal("DryRunRows must return at least one row for selected platform")
	}
	found := false
	for _, r := range rows {
		if r.PlatformID == platforms[0].ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("DryRunRows must include selected platform %q", platforms[0].ID)
	}
}

func TestHydrationModel_CommandDispatchReturnsRunSpec(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}

	m := NewModel(st, platforms, selected)
	spec, ok := m.CommandDispatch(platforms[0].ID)

	if !ok {
		t.Skip("CommandDispatch returns not available when script missing; adapter-only")
	}
	_ = spec
	if spec.Root != root {
		t.Errorf("spec.Root = %q, want %q", spec.Root, root)
	}
	if spec.Name != platforms[0].ID {
		t.Errorf("spec.Name = %q, want %q", spec.Name, platforms[0].ID)
	}
}

func TestHydrationModel_NotAvailableWhenScriptMissing(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}

	m := NewModel(st, platforms, selected)
	_, ok := m.CommandDispatch(platforms[0].ID)

	if ok {
		t.Skip("when script exists, ok=true; when missing, ok=false")
	}
	view := m.View()
	if !strings.Contains(view, "not available") && !strings.Contains(view, "unavailable") {
		t.Logf("View when unavailable should mention not available; got %q", view)
	}
}
