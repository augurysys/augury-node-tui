package validations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

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

func TestValidationsModel_KeyMappingForPresetExecution(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	scriptsDir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, script := range []string{"validate-all.sh", "validate-shellcheck.sh", "validate-bats.sh", "validate-parse-test.sh"} {
		if err := os.WriteFile(filepath.Join(scriptsDir, script), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	fakeNixDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(fakeNixDir, "nix"), []byte("#!/bin/sh\nexec sh -c \"echo ready\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeNixDir+string(filepath.ListSeparator)+os.Getenv("PATH"))
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	keyToReq := map[string]engine.ActionRequest{
		"1": engine.ValidationsAll,
		"2": engine.ValidationsShellcheck,
		"3": engine.ValidationsBats,
		"4": engine.ValidationsParse,
	}
	for key, wantReq := range keyToReq {
		m := NewModel(st)
		_, cmd := m.Update(keyMsg(key))
		if cmd == nil {
			t.Errorf("key %q must trigger a command", key)
			continue
		}
		msg := cmd()
		if msg == nil {
			t.Errorf("key %q command must produce a message", key)
			continue
		}
		jrm, ok := msg.(jobResultMsg)
		if !ok {
			t.Errorf("key %q: expected jobResultMsg, got %T", key, msg)
			continue
		}
		if jrm.Request.ID() != wantReq.ID() {
			t.Errorf("key %q: request ID = %q, want %q", key, jrm.Request.ID(), wantReq.ID())
		}
	}
}

func TestValidationsModel_ActionBlockedWhenNixNotReady(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	scriptsDir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "validate-all.sh"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModel(st)

	nix := engine.ProbeNix(root)
	if nix.Ready {
		t.Skip("nix is ready in test env; cannot test blocked state")
	}
	_, _ = m.Update(keyMsg("1"))
	view := m.View()
	if !strings.Contains(view, "blocked") {
		t.Errorf("View must show blocked when nix not ready; got %q", view)
	}
}

func TestValidationsModel_NotAvailableWhenRequiredScriptsMissing(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModel(st)

	_, _ = m.Update(keyMsg("1"))
	view := m.View()
	if !strings.Contains(view, "not-available") {
		t.Errorf("View must show not-available when required scripts missing; got %q", view)
	}
}

func TestDiagram_ValidationsViewIncludesDiagramWhenWideEnough(t *testing.T) {
	st := status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}
	m := NewModel(st)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := m.View()
	if !strings.Contains(view, "\u250c") {
		t.Errorf("View must include validation pipeline diagram (box-drawing) when width >= 60; got %q", view)
	}
}

func TestDiagram_ValidationsViewExcludesDiagramWhenTooNarrow(t *testing.T) {
	st := status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}
	m := NewModel(st)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 24})
	view := m.View()
	if strings.Contains(view, "\u250c") {
		t.Errorf("View should not include box-drawing diagram when width < 60")
	}
}

func TestValidationsModel_ResultSummaryUpdatesAfterRun(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	scriptsDir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "validate-all.sh"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	fakeNixDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(fakeNixDir, "nix"), []byte("#!/bin/sh\nexec sh -c \"echo ready\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeNixDir+string(filepath.ListSeparator)+os.Getenv("PATH"))
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := NewModel(st)

	_, cmd := m.Update(keyMsg("1"))
	if cmd == nil {
		t.Fatal("key 1 must trigger a command")
	}
	msg := cmd()
	if msg == nil {
		t.Fatal("command must produce a message")
	}
	m2, _ := m.Update(msg)
	view := m2.(*Model).View()
	if !strings.Contains(view, "success") {
		t.Errorf("View must render run result status success after run; got %q", view)
	}
}

func TestValidationsScreen_UsesDataTable(t *testing.T) {
	st := status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}
	m := NewModel(st)
	// Add some test validation data
	m.SetValidations([]Validation{
		{Name: "Test1", Status: "pass", Message: "OK"},
		{Name: "Test2", Status: "fail", Message: "Error"},
	})

	view := m.View()

	// Should use DataTable component (check for table formatting)
	if !strings.Contains(view, "│") {
		t.Error("Validations screen should use DataTable with column separators")
	}

	// Should render validation data
	if !strings.Contains(view, "Test1") || !strings.Contains(view, "Test2") {
		t.Error("Should render validation names")
	}
}
