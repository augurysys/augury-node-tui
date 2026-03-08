package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func TestLog_TabSwitchTogglesFullAndError(t *testing.T) {
	tmp := t.TempDir()
	root := tmp
	logDir := filepath.Join(root, "tmp", "augury-node-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	logPath := filepath.Join(logDir, platforms[0].ID+".log")
	if err := os.WriteFile(logPath, []byte("line1\nerror: failed\nline3"), 0644); err != nil {
		t.Fatal(err)
	}
	selected := map[string]bool{platforms[0].ID: true}
	m := NewModel(status.RepoStatus{Root: root, Branch: "main", SHA: "x"}, platforms, selected)
	m.Summary = &Summary{
		Rows: []SummaryRow{{PlatformID: platforms[0].ID, Status: RowStatusFailure}},
	}
	m.Focused = 0
	if m.LogTab != "full" {
		t.Errorf("initial LogTab want full, got %q", m.LogTab)
	}
	child, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = child.(*Model)
	if m.LogTab != "error" {
		t.Errorf("after t: LogTab want error, got %q", m.LogTab)
	}
	child, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = child.(*Model)
	if m.LogTab != "full" {
		t.Errorf("after t again: LogTab want full, got %q", m.LogTab)
	}
}

func TestLog_JumpToErrorSwitchesToErrorTab(t *testing.T) {
	tmp := t.TempDir()
	logDir := filepath.Join(tmp, "tmp", "augury-node-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	if err := os.WriteFile(filepath.Join(logDir, platforms[0].ID+".log"), []byte("a\nerror: x\nb"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, map[string]bool{platforms[0].ID: true})
	m.Summary = &Summary{Rows: []SummaryRow{{PlatformID: platforms[0].ID, Status: RowStatusFailure}}}
	m.Focused = 0
	child, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = child.(*Model)
	if m.LogTab != "error" {
		t.Errorf("after e: LogTab want error, got %q", m.LogTab)
	}
}

func TestLog_NavigationKeysScrollLogView(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(logDir, pid+".log"), []byte("L1\nL2\nL3\nL4\nL5"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, map[string]bool{pid: true})
	m.Summary = &Summary{Rows: []SummaryRow{{PlatformID: pid, Status: RowStatusFailure}}}
	m.Focused = 0
	off0 := m.scrollOffsetForPlatform(pid)
	child, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = child.(*Model)
	child, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = child.(*Model)
	off1 := m.scrollOffsetForPlatform(pid)
	if off1 <= off0 {
		t.Errorf("j/down should increase scroll offset: was %d, now %d", off0, off1)
	}
	child, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = child.(*Model)
	off2 := m.scrollOffsetForPlatform(pid)
	if off2 >= off1 {
		t.Errorf("k/up should decrease scroll offset: was %d, now %d", off1, off2)
	}
}

func TestLog_TabKeySwitchesLikeT(t *testing.T) {
	tmp := t.TempDir()
	logDir := filepath.Join(tmp, "tmp", "augury-node-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	if err := os.WriteFile(filepath.Join(logDir, platforms[0].ID+".log"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, map[string]bool{platforms[0].ID: true})
	m.Summary = &Summary{Rows: []SummaryRow{{PlatformID: platforms[0].ID, Status: RowStatusFailure}}}
	m.Focused = 0
	child, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = child.(*Model)
	if m.LogTab != "error" {
		t.Errorf("tab key: LogTab want error, got %q", m.LogTab)
	}
}

func TestLog_PgUpPgDownScroll(t *testing.T) {
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
	logContent := strings.Repeat("line\n", 30)
	if err := os.WriteFile(filepath.Join(logDir, pid+".log"), []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, map[string]bool{pid: true})
	m.Summary = &Summary{Rows: []SummaryRow{{PlatformID: pid, Status: RowStatusFailure}}}
	m.Focused = 0
	child, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = child.(*Model)
	child, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = child.(*Model)
	if m.scrollOffsetForPlatform(pid) < 10 {
		t.Errorf("pgdown should increase scroll by 10; got %d", m.scrollOffsetForPlatform(pid))
	}
	child, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = child.(*Model)
	if m.scrollOffsetForPlatform(pid) >= 20 {
		t.Errorf("pgup should decrease scroll; got %d", m.scrollOffsetForPlatform(pid))
	}
}

func TestLog_ErrorTabShowsFirstErrorContext(t *testing.T) {
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
	logContent := "pre1\npre2\nerror: build failed\npost1\npost2"
	if err := os.WriteFile(filepath.Join(logDir, pid+".log"), []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, map[string]bool{pid: true})
	m.Summary = &Summary{Rows: []SummaryRow{{PlatformID: pid, Status: RowStatusFailure}}}
	m.Focused = 0
	m.LogTab = "error"
	view := m.View()
	if !strings.Contains(view, "error: build failed") {
		t.Errorf("error tab view must contain first error context; got %q", view)
	}
	if !strings.Contains(view, "pre2") || !strings.Contains(view, "post1") {
		t.Errorf("error tab view must contain context around error; got %q", view)
	}
}

func TestLog_ScrollClampedToValidRange(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(logDir, pid+".log"), []byte("L1\nL2\nL3"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(status.RepoStatus{Root: tmp, Branch: "main", SHA: "x"}, platforms, map[string]bool{pid: true})
	m.Summary = &Summary{Rows: []SummaryRow{{PlatformID: pid, Status: RowStatusFailure}}}
	m.Focused = 0
	child, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = child.(*Model)
	if m.scrollOffsetForPlatform(pid) != 0 {
		t.Errorf("k at top must not go negative; got %d", m.scrollOffsetForPlatform(pid))
	}
	for i := 0; i < 20; i++ {
		child, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = child.(*Model)
	}
	if m.scrollOffsetForPlatform(pid) > 2 {
		t.Errorf("scroll must clamp to max (2 lines); got %d", m.scrollOffsetForPlatform(pid))
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
	m.Summary = &Summary{Rows: []SummaryRow{{PlatformID: pid, Status: RowStatusFailure}}}
	m.Focused = 0
	view := m.View()
	if !strings.Contains(view, "Build results") {
		t.Errorf("View must show Build results when Summary exists; got %q", view)
	}
	if !strings.Contains(view, "log line 1") {
		t.Errorf("View must render log content; got %q", view)
	}
	if !strings.Contains(view, "--- log ---") {
		t.Errorf("View must show log section; got %q", view)
	}
}
