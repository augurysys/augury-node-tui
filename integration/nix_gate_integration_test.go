package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/augurysys/augury-node-tui/internal/app"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNix_BlockedMode_ActionsDisabled(t *testing.T) {
	root := setupFixtureRoot(t)
	nixBlocked := engine.NixState{Ready: false, Reason: "nix develop failed: flake not found"}
	st, _ := status.Collect(root)
	if st.Root == "" {
		st = status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	}
	m := app.NewModelWithNix(st, platform.Registry(), 10*time.Millisecond, nixBlocked)

	model, cmd := m.Update(ui.TimeoutMsg{})
	if cmd != nil {
		model, _ = model.(*app.Model).Update(cmd())
	}
	m = model.(*app.Model)
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cmd != nil {
		model, _ = model.(*app.Model).Update(cmd())
	}
	m = model.(*app.Model)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")})
	m = model.(*app.Model)

	view := m.View()
	if !strings.Contains(view, nixBlocked.Reason) {
		t.Errorf("view must contain blocked reason when nix not ready; got %q", view)
	}
}

func TestNix_ReadyMode_ActionsExecute(t *testing.T) {
	root := setupFixtureRoot(t)
	tmp := t.TempDir()
	fakeNix := filepath.Join(tmp, "nix")
	if err := os.WriteFile(fakeNix, []byte("#!/bin/sh\n[ \"$1\" = develop ] && exit 0\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", tmp+string(filepath.ListSeparator)+origPath)
	defer t.Setenv("PATH", origPath)

	st, _ := status.Collect(root)
	if st.Root == "" {
		st = status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	}
	m := app.NewModelWithNix(st, platform.Registry(), 10*time.Millisecond, engine.NixState{Ready: true, Reason: ""})

	model, cmd := m.Update(ui.TimeoutMsg{})
	if cmd != nil {
		model, _ = model.(*app.Model).Update(cmd())
	}
	m = model.(*app.Model)
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cmd != nil {
		model, _ = model.(*app.Model).Update(cmd())
	}
	m = model.(*app.Model)
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")})
	m = model.(*app.Model)

	if cmd == nil {
		t.Error("action B when nix ready: expected cmd, got nil")
	}
	view := m.View()
	if strings.Contains(view, "nix develop failed") || strings.Contains(view, "flake not found") {
		t.Errorf("view must not show blocked reason when nix ready; got %q", view)
	}
}
