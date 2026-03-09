package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

func TestBuildModel_ModeSelectionDefaultsToSmart(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry(), nil)
	if m.Mode != run.ModeSmart {
		t.Errorf("mode selection must default to Smart; got %q", m.Mode)
	}
}

func TestBuildModel_CleanModeMapsToCLEAN1(t *testing.T) {
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platforms, selected)
	m.Mode = run.ModeClean
	specs := m.RunSpecs()
	if len(specs) == 0 {
		t.Fatal("RunSpecs must include selected platform")
	}
	for _, s := range specs {
		if s.Mode != run.ModeClean {
			t.Errorf("RunSpec for clean mode must have Mode=ModeClean; got %q", s.Mode)
		}
	}
}

func TestBuildModel_ConfirmEmitsConfirmMsg(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry(), nil)
	_, cmd := m.Update(ConfirmPlanMsg{})
	if cmd == nil {
		t.Fatal("ConfirmPlanMsg must return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(StartBuildMsg); !ok {
		t.Errorf("confirm must produce StartBuildMsg; got %T", msg)
	}
}

func TestBuildModel_CancelEmitsCancelMsg(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry(), nil)
	_, cmd := m.Update(CancelPlanMsg{})
	if cmd == nil {
		t.Fatal("CancelPlanMsg must return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(nav.NavigateBackMsg); !ok {
		t.Errorf("cancel must produce NavigateBackMsg; got %T", msg)
	}
}

func TestBuildModel_PlanBuiltFromSelectedPlatformsOnly(t *testing.T) {
	platforms := platform.Registry()
	if len(platforms) < 2 {
		t.Fatal("need at least two platforms")
	}
	selected := map[string]bool{platforms[0].ID: true, platforms[1].ID: false}
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platforms, selected)
	plan := m.Plan()
	if len(plan.Entries) != 1 {
		t.Errorf("plan must include only selected platforms; got %d entries, want 1", len(plan.Entries))
	}
	if len(plan.Entries) > 0 && plan.Entries[0].PlatformID != platforms[0].ID {
		t.Errorf("plan entry must be selected platform %q; got %q", platforms[0].ID, plan.Entries[0].PlatformID)
	}
}

func TestBuildModel_PlanUsesModelForceRebuildState(t *testing.T) {
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platforms, selected)
	m.ToggleForceRebuild(platforms[0].ID)
	plan := m.Plan()
	if !plan.ForceRebuild[platforms[0].ID] {
		t.Error("plan must use model-level force rebuild state")
	}
}

func TestBuildModel_CycleMode(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry(), nil)
	if m.Mode != run.ModeSmart {
		t.Errorf("initial mode must be Smart; got %q", m.Mode)
	}
	m.CycleMode()
	if m.Mode != run.ModeClean {
		t.Errorf("after first cycle: want Clean; got %q", m.Mode)
	}
	m.CycleMode()
	if m.Mode != run.ModeValidationOnly {
		t.Errorf("after second cycle: want ValidationOnly; got %q", m.Mode)
	}
	m.CycleMode()
	if m.Mode != run.ModeSmart {
		t.Errorf("after third cycle: want Smart; got %q", m.Mode)
	}
}

func TestBuildModel_ViewRendersPreflightPlan(t *testing.T) {
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platforms, selected)
	view := m.View()
	if view == "" {
		t.Fatal("View must not be empty")
	}
	if !strings.Contains(view, "Build pre-flight") {
		t.Errorf("View should contain pre-flight header; got %q", view)
	}
	if !strings.Contains(view, platforms[0].ID) {
		t.Errorf("View should contain selected platform; got %q", view)
	}
}

func TestBuildModel_EnterEmitsConfirmPlanMsg(t *testing.T) {
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platform.Registry(), nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter must return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(ConfirmPlanMsg); !ok {
		t.Errorf("Enter must produce ConfirmPlanMsg; got %T", msg)
	}
}

func TestLog_ViewRendersLogWhenSummaryExists(t *testing.T) {
	tmp := t.TempDir()
	logDir := filepath.Join(tmp, "tmp", "augury-node-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	pid := platforms[0].ID
	if err := os.WriteFile(filepath.Join(logDir, pid+".log"), []byte("log line 1\nlog line 2"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, map[string]bool{pid: true})
	child, _ := m.Update(BuildCompleteMsg{
		Summary: &Summary{Rows: []SummaryRow{{PlatformID: pid, Status: RowStatusFailure}}},
	})
	m = child.(*Model)
	view := m.View()
	if !strings.Contains(view, "log line 1") {
		t.Errorf("View must render log content; got %q", view)
	}
	if !strings.Contains(view, pid) {
		t.Errorf("View must show platform in ParallelTracker; got %q", view)
	}
}

func TestLog_PlatformSwitchWithBracketKeys(t *testing.T) {
	tmp := t.TempDir()
	logDir := filepath.Join(tmp, "tmp", "augury-node-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) < 2 {
		t.Skip("need at least two platforms for platform switch test")
	}
	p0, p1 := platforms[0].ID, platforms[1].ID
	if err := os.WriteFile(filepath.Join(logDir, p0+".log"), []byte("platform0 log"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, p1+".log"), []byte("platform1 log"), 0644); err != nil {
		t.Fatal(err)
	}
	selected := map[string]bool{p0: true, p1: true}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, selected)
	child, _ := m.Update(BuildCompleteMsg{
		Summary: &Summary{
			Rows: []SummaryRow{
				{PlatformID: p0, Status: RowStatusSuccess},
				{PlatformID: p1, Status: RowStatusSuccess},
			},
		},
	})
	m = child.(*Model)
	view0 := m.View()
	if !strings.Contains(view0, "platform0 log") {
		t.Errorf("initial view must show platform0 log; got %q", view0)
	}
	child, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	m = child.(*Model)
	view1 := m.View()
	if !strings.Contains(view1, "platform1 log") {
		t.Errorf("after ]: view must show platform1 log; got %q", view1)
	}
	child, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	m = child.(*Model)
	view2 := m.View()
	if !strings.Contains(view2, "platform0 log") {
		t.Errorf("after [: view must show platform0 log; got %q", view2)
	}
}

func TestBuildModel_StartBuildBlockedWhenNixNotReady(t *testing.T) {
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	m := NewModel(status.RepoStatus{Root: "/x", Branch: "main", SHA: "x"}, platforms, selected)
	m.SetNixState(engine.NixState{Ready: false, Reason: "nix not available"})

	_, cmd := m.Update(StartBuildMsg{})
	if cmd != nil {
		t.Error("StartBuildMsg when nix not ready: expected no cmd, got cmd")
	}
	view := m.View()
	if !strings.Contains(view, "nix not available") {
		t.Errorf("View must show blocked reason when nix not ready; got %q", view)
	}
}

func TestBuildModel_StartBuildAllowedWhenNixReady(t *testing.T) {
	root := t.TempDir()
	scriptsDevices := root + "/scripts/devices"
	if err := os.MkdirAll(scriptsDevices, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptsDevices+"/node2-build.sh", []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	selected := map[string]bool{platforms[0].ID: true}
	m := NewModel(status.RepoStatus{Root: root, Branch: "main", SHA: "x"}, platforms, selected)
	m.SetNixState(engine.NixState{Ready: true, Reason: ""})

	_, cmd := m.Update(StartBuildMsg{})
	if cmd == nil {
		t.Fatal("StartBuildMsg when nix ready: expected cmd, got nil")
	}
	_ = cmd
}
