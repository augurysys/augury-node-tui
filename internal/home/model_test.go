package home

import (
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHomeModel_ViewRendersRepoMetadata(t *testing.T) {
	st := status.RepoStatus{
		Root:   "/repo",
		Branch: "main",
		SHA:    "abc1234",
	}
	m := NewModel(st, platform.Registry())
	view := m.View()
	if view == "" {
		t.Fatal("View must not be empty")
	}
	if !strings.Contains(view, "/repo") {
		t.Errorf("View should contain root; got %q", view)
	}
	if !strings.Contains(view, "main") {
		t.Errorf("View should contain branch; got %q", view)
	}
	if !strings.Contains(view, "abc1234") {
		t.Errorf("View should contain sha; got %q", view)
	}
}

func TestHomeModel_ViewRendersDirtyIndicators(t *testing.T) {
	st := status.RepoStatus{
		Root:   "/repo",
		Branch: "main",
		SHA:    "abc1234",
		Dirty:  map[string]bool{"common/": true, "submodules/halo-node/": false},
	}
	m := NewModel(st, platform.Registry())
	view := m.View()
	if !strings.Contains(view, "common/") {
		t.Errorf("View should show common/ path; got %q", view)
	}
	if !strings.Contains(view, "dirty") || !strings.Contains(view, "clean") {
		t.Errorf("View should show dirty/clean indicators; got %q", view)
	}
}

func TestHomeModel_PlatformSelectDeselect(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	id := platforms[0].ID

	if m.IsPlatformSelected(id) {
		t.Error("platform should not be selected initially")
	}
	m.TogglePlatform(id)
	if !m.IsPlatformSelected(id) {
		t.Error("platform should be selected after toggle")
	}
	m.TogglePlatform(id)
	if m.IsPlatformSelected(id) {
		t.Error("platform should be deselected after second toggle")
	}
}

func TestHomeModel_KeyA_EmitsReplaySplash(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	_ = model
	if cmd == nil {
		t.Fatal("pressing 'a' must return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(nav.ReplaySplashMsg); !ok {
		t.Errorf("cmd must produce ReplaySplashMsg; got %T", msg)
	}
}

func TestHomeModel_KeyQ_Quits(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("pressing 'q' must return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("cmd must produce QuitMsg; got %T", msg)
	}
}

func TestHomeModel_KeyB_EmitsNavigate(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if cmd == nil {
		t.Fatal("pressing 'b' must return a cmd")
	}
	msg := cmd()
	nm, ok := msg.(nav.NavigateMsg)
	if !ok || nm.Route != "build" {
		t.Errorf("cmd must produce NavigateMsg{Route: build}; got %T %v", msg, msg)
	}
}

func TestHomeModel_KeyH_EmitsNavigate(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if cmd == nil {
		t.Fatal("pressing 'h' must return a cmd")
	}
	msg := cmd()
	nm, ok := msg.(nav.NavigateMsg)
	if !ok || nm.Route != "hydrate" {
		t.Errorf("cmd must produce NavigateMsg{Route: hydrate}; got %T", msg)
	}
}

func TestHomeModel_KeyC_EmitsNavigate(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cmd == nil {
		t.Fatal("pressing 'c' must return a cmd")
	}
	msg := cmd()
	nm, ok := msg.(nav.NavigateMsg)
	if !ok || nm.Route != "caches" {
		t.Errorf("cmd must produce NavigateMsg{Route: caches}; got %T", msg)
	}
}

func TestHomeModel_KeyV_EmitsNavigate(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if cmd == nil {
		t.Fatal("pressing 'v' must return a cmd")
	}
	msg := cmd()
	nm, ok := msg.(nav.NavigateMsg)
	if !ok || nm.Route != "validations" {
		t.Errorf("cmd must produce NavigateMsg{Route: validations}; got %T", msg)
	}
}

func TestHomeModel_KeyO_EmitsNavigate(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	if cmd == nil {
		t.Fatal("pressing 'o' must return a cmd")
	}
	msg := cmd()
	nm, ok := msg.(nav.NavigateMsg)
	if !ok || nm.Route != "hints" {
		t.Errorf("cmd must produce NavigateMsg{Route: hints}; got %T", msg)
	}
}

func TestHomeModel_KeyJ_MovesFocusDown(t *testing.T) {
	platforms := platform.Registry()
	if len(platforms) < 2 {
		t.Skip("need at least 2 platforms")
	}
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platforms)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = model.(*Model)
	if m.Focused != 1 {
		t.Errorf("j should move focus to 1; got %d", m.Focused)
	}
}

func TestHomeModel_KeySpace_TogglesPlatform(t *testing.T) {
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platforms)
	id := platforms[0].ID
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = model.(*Model)
	if !m.IsPlatformSelected(id) {
		t.Error("space should toggle platform selection")
	}
}

