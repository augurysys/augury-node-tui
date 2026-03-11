package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func stubStatus() status.RepoStatus {
	return status.RepoStatus{Root: "/x", Branch: "main", SHA: "abc1234"}
}

func TestApp_SplashTransitionsToHomeOnTimeout(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 10*time.Millisecond)
	if m.Route() != "splash" {
		t.Fatalf("initial route should be splash; got %q", m.Route())
	}
	model, cmd := m.Update(ui.TimeoutMsg{})
	if cmd != nil {
		_ = cmd()
	}
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("splash timeout should transition to home; got %q", m.Route())
	}
}

func TestApp_SplashTransitionsToHomeOnKey(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("splash key should transition to home; got %q", m.Route())
	}
}

func TestApp_HomeKeyB_TransitionsToBuild(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if cmd == nil {
		t.Fatal("home b must return cmd")
	}
	model, _ = model.(*Model).Update(cmd())
	m = model.(*Model)
	if m.Route() != "build" {
		t.Errorf("home b should transition to build; got %q", m.Route())
	}
}

func TestApp_HomeKeyH_TransitionsToHydrate(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if cmd == nil {
		t.Fatal("home h must return cmd")
	}
	model, _ = model.(*Model).Update(cmd())
	m = model.(*Model)
	if m.Route() != "hydrate" {
		t.Errorf("home h should transition to hydrate; got %q", m.Route())
	}
}

func TestApp_HomeKeyC_TransitionsToCaches(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cmd == nil {
		t.Fatal("home c must return cmd")
	}
	model, _ = model.(*Model).Update(cmd())
	m = model.(*Model)
	if m.Route() != "caches" {
		t.Errorf("home c should transition to caches; got %q", m.Route())
	}
}

func TestApp_HomeKeyV_TransitionsToValidations(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if cmd == nil {
		t.Fatal("home v must return cmd")
	}
	model, _ = model.(*Model).Update(cmd())
	m = model.(*Model)
	if m.Route() != "validations" {
		t.Errorf("home v should transition to validations; got %q", m.Route())
	}
}

func TestApp_HomeKeyO_TransitionsToHints(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	if cmd == nil {
		t.Fatal("home o must return cmd")
	}
	model, _ = model.(*Model).Update(cmd())
	m = model.(*Model)
	if m.Route() != "hints" {
		t.Errorf("home o should transition to hints; got %q", m.Route())
	}
}

func TestApp_ReturnFromBuild_GoesToHome(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "build"
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("b from build should go to home; got %q", m.Route())
	}
}

func TestApp_ReturnFromHydrate_GoesToHome(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "hydrate"
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("b from hydrate should go to home; got %q", m.Route())
	}
}

func TestApp_ReturnFromCaches_GoesToHome(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "caches"
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("b from caches should go to home; got %q", m.Route())
	}
}

func TestApp_CachesConfirmModal_EscCancelsNotRouteHome(t *testing.T) {
	root := t.TempDir()
	scriptsDev := root + "/scripts/dev"
	if err := os.MkdirAll(scriptsDev, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDev+"/delete-build-unit-cache.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.WriteFile(tmp+"/nix", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmp+string(filepath.ListSeparator)+os.Getenv("PATH"))

	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry(), 2*time.Second)
	m.route = "caches"

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m = model.(*Model)
	if !m.caches.ConfirmShown() {
		t.Fatal("D must show confirm modal")
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(*Model)
	if m.Route() != "caches" {
		t.Errorf("esc with confirm shown must stay on caches (cancel modal); got route %q", m.Route())
	}
	if m.caches.ConfirmShown() {
		t.Error("esc must dismiss confirm modal")
	}
}

func TestApp_ReturnFromValidations_GoesToHome(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "validations"
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("b from validations should go to home; got %q", m.Route())
	}
}

func TestApp_ReturnFromHints_GoesToHome(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "hints"
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("b from hints should go to home; got %q", m.Route())
	}
}

func TestApp_GoToFlash_TransitionsToFlashRoute(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, _ := m.Update(nav.GoToFlash{})
	m = model.(*Model)
	if m.Route() != "flash" {
		t.Errorf("After GoToFlash, route = %v, want flash", m.Route())
	}
	if m.flash == nil {
		t.Error("flash model should be initialized")
	}
}

func TestApp_FlashBack_ReturnsToHome(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, _ := m.Update(nav.GoToFlash{})
	m = model.(*Model)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = model.(*Model)
	if m.Route() != "home" {
		t.Errorf("After b from flash, route = %v, want home", m.Route())
	}
}

func TestApp_HomeKeyA_ReplaysSplash(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		t.Fatal("home a must return cmd")
	}
	model, _ = model.(*Model).Update(cmd())
	m = model.(*Model)
	if m.Route() != "splash" {
		t.Errorf("home a should replay splash; got %q", m.Route())
	}
}

func TestApp_WindowSizePropagatesToDiagramRoutes(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "splash"
	model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = model.(*Model)

	// Navigate to home without any further WindowSizeMsg
	m.route = "home"
	view := m.View()
	if view == "" {
		t.Fatal("view must not be empty")
	}
	if !containsRune(view, '\u250c') {
		t.Errorf("home view must include diagram (box-drawing) after WindowSizeMsg propagated; got %q", view)
	}

	// Navigate to caches
	m.route = "caches"
	view = m.View()
	if !containsRune(view, '\u250c') {
		t.Errorf("caches view must include diagram after WindowSizeMsg propagated; got %q", view)
	}

	// Navigate to validations
	m.route = "validations"
	view = m.View()
	if !containsRune(view, '\u250c') {
		t.Errorf("validations view must include diagram after WindowSizeMsg propagated; got %q", view)
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func TestApp_HomeKeyQ_Quits(t *testing.T) {
	m := NewModel(stubStatus(), platform.Registry(), 2*time.Second)
	m.route = "home"
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("home q must return quit cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("cmd must produce QuitMsg; got %T", msg)
	}
}

func TestNixGate_ActionKeypressesBlockedWhenNixNotReady(t *testing.T) {
	root := t.TempDir()
	scriptsDevices := root + "/scripts/devices"
	if err := os.MkdirAll(scriptsDevices, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDevices+"/node2-build.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	nixNotReady := engine.NixState{Ready: false, Reason: "nix not available"}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModelWithNix(st, platform.Registry(), 2*time.Second, nixNotReady)
	m.route = "caches"

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")})
	m = model.(*Model)

	if m.caches.DisabledReason() != nixNotReady.Reason {
		t.Errorf("action B when nix not ready: DisabledReason = %q, want %q", m.caches.DisabledReason(), nixNotReady.Reason)
	}
}

func TestNixGate_ActionKeypressesAllowedWhenNixReady(t *testing.T) {
	root := t.TempDir()
	scriptsDevices := root + "/scripts/devices"
	if err := os.MkdirAll(scriptsDevices, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDevices+"/node2-build.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.WriteFile(tmp+"/nix", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmp+string(filepath.ListSeparator)+os.Getenv("PATH"))

	nixReady := engine.NixState{Ready: true, Reason: ""}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModelWithNix(st, platform.Registry(), 2*time.Second, nixReady)
	m.route = "caches"

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")})
	m = model.(*Model)

	if m.caches.DisabledReason() != "" {
		t.Errorf("action B when nix ready: DisabledReason = %q, want empty", m.caches.DisabledReason())
	}
	if cmd == nil {
		t.Error("action B when nix ready: expected cmd, got nil")
	}
	_ = cmd
}

func TestNixGate_BlockedReasonSurfacedInUIState(t *testing.T) {
	root := t.TempDir()
	scriptsDevices := root + "/scripts/devices"
	if err := os.MkdirAll(scriptsDevices, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDevices+"/node2-build.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	blockedReason := "nix develop failed: flake not found"
	nixNotReady := engine.NixState{Ready: false, Reason: blockedReason}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModelWithNix(st, platform.Registry(), 2*time.Second, nixNotReady)
	m.route = "caches"

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")})
	m = model.(*Model)
	view := m.View()

	if m.caches.DisabledReason() != blockedReason {
		t.Errorf("DisabledReason = %q, want %q", m.caches.DisabledReason(), blockedReason)
	}
	if !strings.Contains(view, blockedReason) {
		t.Errorf("view must contain blocked reason %q; got %q", blockedReason, view)
	}
}

func TestNixGate_BuildBlockedWhenNixNotReady(t *testing.T) {
	root := t.TempDir()
	scriptsDevices := root + "/scripts/devices"
	if err := os.MkdirAll(scriptsDevices, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDevices+"/node2-build.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	nixNotReady := engine.NixState{Ready: false, Reason: "nix not available"}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	platforms := platform.Registry()
	selected := map[string]bool{}
	for _, p := range platforms {
		selected[p.ID] = true
		break
	}
	m := NewModelWithNix(st, platforms, 2*time.Second, nixNotReady)
	m.route = "build"
	m.build.Selected = selected

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*Model)
	if cmd != nil {
		model, cmd = m.Update(cmd())
		m = model.(*Model)
	}
	if cmd != nil {
		model, cmd = m.Update(cmd())
		m = model.(*Model)
	}
	view := m.View()
	if cmd != nil {
		t.Error("build start when nix not ready: expected no cmd, got cmd")
	}
	if !strings.Contains(view, "nix not available") {
		t.Errorf("build view must contain blocked reason when nix not ready; got %q", view)
	}
}

func TestNixGate_BuildAllowedWhenNixReady(t *testing.T) {
	root := t.TempDir()
	scriptsDevices := root + "/scripts/devices"
	if err := os.MkdirAll(scriptsDevices, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDevices+"/node2-build.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	nixReady := engine.NixState{Ready: true, Reason: ""}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	platforms := platform.Registry()
	selected := map[string]bool{}
	for _, p := range platforms {
		selected[p.ID] = true
		break
	}
	m := NewModelWithNix(st, platforms, 2*time.Second, nixReady)
	m.route = "build"
	m.build.Selected = selected

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*Model)
	if cmd != nil {
		model, cmd = m.Update(cmd())
		m = model.(*Model)
	}
	if cmd != nil {
		model, cmd = m.Update(cmd())
		m = model.(*Model)
	}
	if cmd == nil {
		t.Error("build start when nix ready: expected cmd, got nil")
	}
	_ = cmd
}
