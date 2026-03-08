package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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
