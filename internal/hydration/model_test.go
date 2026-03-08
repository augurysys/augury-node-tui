package hydration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

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
	scriptsDir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	hydratePath := filepath.Join(scriptsDir, "hydrate")
	if err := os.WriteFile(hydratePath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}

	m := NewModel(st, platforms, selected)
	spec, ok := m.CommandDispatch(platforms[0].ID)

	if !ok {
		t.Fatal("CommandDispatch must return ok=true when scripts/hydrate exists")
	}
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
		t.Fatal("CommandDispatch must return ok=false when scripts/hydrate is missing")
	}
	view := m.View()
	if !strings.Contains(view, "not available") && !strings.Contains(view, "unavailable") {
		t.Errorf("View when unavailable must mention not available; got %q", view)
	}
}

func TestHydrationModel_DTriggersDryRunActionPreview(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	scriptsDir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "hydrate"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	fakeNixDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(fakeNixDir, "nix"), []byte("#!/bin/sh\nexec sh -c \"echo ready\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeNixDir+string(filepath.ListSeparator)+os.Getenv("PATH"))
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	platID := platforms[0].ID
	selected := map[string]bool{platID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}
	m := NewModel(st, platforms, selected)

	_, cmd := m.Update(keyMsg("D"))
	if cmd == nil {
		t.Fatal("D must trigger a command (dry-run action)")
	}
	msg := cmd()
	if msg == nil {
		t.Error("D command must produce a message")
	}
	_ = engine.ExecuteAction(context.Background(), root, engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetDryRun, PlatformID: platID})
}

func TestHydrationModel_HTriggersHydrationActionForSelectedPlatforms(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	scriptsDir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "hydrate"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	fakeNixDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(fakeNixDir, "nix"), []byte("#!/bin/sh\nexec sh -c \"echo ready\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeNixDir+string(filepath.ListSeparator)+os.Getenv("PATH"))
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	platID := platforms[0].ID
	selected := map[string]bool{platID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}
	m := NewModel(st, platforms, selected)

	_, cmd := m.Update(keyMsg("H"))
	if cmd == nil {
		t.Fatal("H must trigger a command (hydration action)")
	}
	msg := cmd()
	if msg == nil {
		t.Error("H command must produce a message")
	}
}

func TestHydrationModel_BlockedStateWhenNixNotReady(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	scriptsDir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "hydrate"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	platID := platforms[0].ID
	selected := map[string]bool{platID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}
	m := NewModel(st, platforms, selected)

	nix := engine.ProbeNix(root)
	if nix.Ready {
		t.Skip("nix is ready in test env; cannot test blocked state")
	}
	_, _ = m.Update(keyMsg("D"))
	view := m.View()
	if !strings.Contains(view, "blocked") {
		t.Errorf("View must show blocked when nix not ready; got %q", view)
	}
}

func TestHydrationModel_NotAvailableStateWhenScriptMissing(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	platID := platforms[0].ID
	selected := map[string]bool{platID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}
	m := NewModel(st, platforms, selected)

	_, _ = m.Update(keyMsg("H"))
	view := m.View()
	if !strings.Contains(view, "not available") && !strings.Contains(view, "not-available") {
		t.Errorf("View must show not-available when script missing; got %q", view)
	}
}
