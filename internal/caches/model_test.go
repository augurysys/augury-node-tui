package caches

import (
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

func TestCachesModel_BuildUnitTabActionKeys(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need platforms")
	}
	platID := platforms[0].ID

	for _, key := range []string{"B", "R", "D"} {
		_, _ = m.Update(keyMsg(key))
		req, ok := ActionForKey(TabBuildUnit, key, platID)
		if !ok {
			t.Errorf("key %q must map to action on build-unit tab", key)
		}
		if req.Kind == "" {
			t.Errorf("key %q must resolve to non-zero ActionRequest", key)
		}
	}
}

func TestCachesModel_PlatformCacheTabActionKeys(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())
	m.NextTab()
	if m.ActiveTab() != TabPlatform {
		t.Fatal("must be on platform tab")
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need platforms")
	}
	platID := platforms[0].ID

	for _, key := range []string{"P", "U", "X"} {
		_, _ = m.Update(keyMsg(key))
		req, ok := ActionForKey(TabPlatform, key, platID)
		if !ok {
			t.Errorf("key %q must map to action on platform-cache tab", key)
		}
		if req.Kind == "" {
			t.Errorf("key %q must resolve to non-zero ActionRequest", key)
		}
	}
}

func TestCachesModel_DisabledActionWithCapabilityReason(t *testing.T) {
	root := t.TempDir()
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())

	_, _ = m.Update(keyMsg("R"))
	if m.DisabledReason() == "" {
		t.Error("pressing R with no scripts must set DisabledReason with capability reason")
	}
	view := m.View()
	if !strings.Contains(view, m.DisabledReason()) && m.DisabledReason() != "" {
		t.Error("View must surface DisabledReason when action is disabled")
	}
}

func TestCachesModel_DestructiveActionConfirmationRequired(t *testing.T) {
	root := t.TempDir()
	scriptsDev := root + "/scripts/dev"
	if err := os.MkdirAll(scriptsDev, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDev+"/delete-build-unit-cache.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	fakeNix := tmp + "/nix"
	if err := os.WriteFile(fakeNix, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmp+string(filepath.ListSeparator)+os.Getenv("PATH"))

	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())

	_, _ = m.Update(keyMsg("D"))
	if !m.ConfirmShown() {
		t.Error("pressing D (destructive) must show confirm modal before execution")
	}
	if m.PendingConfirmAction().Target == "" {
		t.Error("PendingConfirmAction must be set when confirm shown for D")
	}

	cm, _ := m.Update(keyMsg("n"))
	c := cm.(*Model)
	if c.ConfirmShown() {
		t.Error("pressing n must dismiss confirm modal without executing")
	}
}

func TestCachesModel_GlobalActionStatusRenderedInView(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())
	m.Update(jobResultMsg{
		Job:     engine.Job{State: engine.JobStateSuccess},
		Request: engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetPull},
	})
	view := m.View()
	if !strings.Contains(view, "success") {
		t.Error("view must render global action status when set")
	}
	if !strings.Contains(view, "Global") {
		t.Error("view must include Global label for platform-agnostic action status")
	}
}

func TestCachesModel_DestructiveXConfirmationRequired(t *testing.T) {
	root := t.TempDir()
	scriptsDev := root + "/scripts/dev"
	if err := os.MkdirAll(scriptsDev, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDev+"/clean-platform-cache.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	fakeNix := tmp + "/nix"
	if err := os.WriteFile(fakeNix, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmp+string(filepath.ListSeparator)+os.Getenv("PATH"))

	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())
	m.NextTab()
	if m.ActiveTab() != TabPlatform {
		t.Fatal("must be on platform tab")
	}

	_, _ = m.Update(keyMsg("X"))
	if !m.ConfirmShown() {
		t.Error("pressing X (platform-cache clean) must show confirm modal before execution")
	}
}

func TestCachesModel_TabsRenderReadOnlyDataSources(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())

	view := m.View()
	if view == "" {
		t.Fatal("View must not be empty")
	}
	if !strings.Contains(strings.ToLower(view), "build") && !strings.Contains(strings.ToLower(view), "unit") {
		t.Errorf("View must render build-unit tab or data; got %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "platform") {
		t.Errorf("View must render platform tab or data; got %q", view)
	}
}

func TestDiagram_CachesViewIncludesDiagramWhenWideEnough(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := m.View()
	if !strings.Contains(view, "\u250c") {
		t.Errorf("View must include cache topology diagram (box-drawing) when width >= 60; got %q", view)
	}
}

func TestDiagram_CachesViewExcludesDiagramWhenTooNarrow(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 24})
	view := m.View()
	if strings.Contains(view, "\u250c") {
		t.Errorf("View should not include box-drawing diagram when width < 60")
	}
}

func TestCachesModel_ActiveTabSwitches(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())

	initial := m.ActiveTab()
	m.NextTab()
	if m.ActiveTab() == initial {
		t.Error("NextTab must change active tab")
	}
}
