package build

import (
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
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
	if _, ok := msg.(NavigateBackMsg); !ok {
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
