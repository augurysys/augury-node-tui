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

func TestHydrationModel_BlockedReasonSurfacedInViewAndState(t *testing.T) {
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
	blockedReason := "nix develop failed: flake not found"
	m.SetNixState(engine.NixState{Ready: false, Reason: blockedReason})

	_, _ = m.Update(keyMsg("D"))
	rowStatus := m.RowStatus(platID)
	view := m.View()

	if !strings.Contains(rowStatus, "blocked") {
		t.Errorf("RowStatus = %q, want to contain blocked", rowStatus)
	}
	if !strings.Contains(rowStatus, blockedReason) {
		t.Errorf("RowStatus = %q, want to contain reason %q", rowStatus, blockedReason)
	}
	if !strings.Contains(view, blockedReason) {
		t.Errorf("View must contain blocked reason %q; got %q", blockedReason, view)
	}

	m2 := NewModel(st, platforms, selected)
	m2.SetNixState(engine.NixState{Ready: false, Reason: blockedReason})
	_, _ = m2.Update(keyMsg("H"))
	rowStatusH := m2.RowStatus(platID)
	if !strings.Contains(rowStatusH, blockedReason) {
		t.Errorf("H blocked: RowStatus = %q, want to contain reason %q", rowStatusH, blockedReason)
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

func TestHydrationModel_LowercaseDAndHAccepted(t *testing.T) {
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

	_, cmdD := m.Update(keyMsg("d"))
	if cmdD == nil {
		t.Error("lowercase d must trigger dry-run command")
	}
	m2 := NewModel(st, platforms, selected)
	_, cmdH := m2.Update(keyMsg("h"))
	if cmdH == nil {
		t.Error("lowercase h must trigger hydration command")
	}
}

func TestHydrationScreen_UsesComponents(t *testing.T) {
	tmp := t.TempDir()
	st := status.RepoStatus{Root: tmp, Branch: "main", SHA: "abc1234"}
	platforms := platform.Registry()
	selected := map[string]bool{}
	if len(platforms) > 0 {
		selected[platforms[0].ID] = true
	}
	m := NewModel(st, platforms, selected)
	m.SetArtifacts([]Artifact{
		{Name: "artifact1", Status: "cached", Progress: 100, Total: 100},
		{Name: "artifact2", Status: "downloading", Progress: 50, Total: 100},
	})

	view := m.View()

	// Should use DataTable (column separators │ or ║)
	if !strings.Contains(view, "│") && !strings.Contains(view, "║") {
		t.Error("Hydration screen should use DataTable")
	}

	// Should render artifact names
	if !strings.Contains(view, "artifact1") {
		t.Error("Should render artifact names")
	}

	// Should use ProgressBar for in-progress artifacts
	if !strings.Contains(view, "█") && !strings.Contains(view, "░") {
		t.Error("Should show progress bars for downloading artifacts")
	}
}

func TestHydrationModel_ViewShowsKeyLegend(t *testing.T) {
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
	selected := map[string]bool{platforms[0].ID: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "abc1234"}
	m := NewModel(st, platforms, selected)

	view := m.View()
	if !strings.Contains(view, "D dry-run") || !strings.Contains(view, "H hydrate") {
		t.Errorf("View must show key legend (D dry-run, H hydrate); got %q", view)
	}
	if !strings.Contains(view, "back") {
		t.Errorf("View must show back key hint; got %q", view)
	}
}
