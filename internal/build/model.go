package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/logs"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmPlanMsg struct{}
type CancelPlanMsg struct{}
type StartBuildMsg struct{}
type CancelBuildMsg struct{}

type BuildCompleteMsg struct {
	Summary *Summary
}

const errorContextBefore = 5
const errorContextAfter = 5

type Model struct {
	Status          status.RepoStatus
	Platforms       []platform.Platform
	Selected        map[string]bool
	nixState        engine.NixState
	nixBlockedReason string
	Mode            run.Mode
	ForceRebuild    map[string]bool
	Summary         *Summary
	BuildCancel     context.CancelFunc
	Focused         int
	LogTab          string
	LogScrollOffset map[string]int
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, selected map[string]bool) *Model {
	m := &Model{
		Status:          st,
		Platforms:       platforms,
		Selected:        selected,
		nixState:        engine.ProbeNix(st.Root),
		Mode:            run.ModeSmart,
		ForceRebuild:    make(map[string]bool),
		LogTab:          "full",
		LogScrollOffset: make(map[string]int),
	}
	if m.Selected == nil {
		m.Selected = make(map[string]bool)
	}
	return m
}

func (m *Model) SetNixState(nix engine.NixState) {
	m.nixState = nix
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		if m.Summary != nil && m.focusedLogPlatformID() != "" {
			pid := m.focusedLogPlatformID()
			switch s {
			case "t", "tab":
				m.resetScrollForPlatform(pid)
				if m.LogTab == "full" {
					m.LogTab = "error"
				} else {
					m.LogTab = "full"
				}
				return m, nil
			case "e":
				m.resetScrollForPlatform(pid)
				m.LogTab = "error"
				return m, nil
			case "j", "down":
				m.adjustScroll(pid, 1)
				return m, nil
			case "k", "up":
				m.adjustScroll(pid, -1)
				return m, nil
			case "pgup":
				m.adjustScroll(pid, -10)
				return m, nil
			case "pgdown":
				m.adjustScroll(pid, 10)
				return m, nil
			}
		}
		switch s {
		case "enter":
			m.nixBlockedReason = ""
			return m, func() tea.Msg { return ConfirmPlanMsg{} }
		case "f":
			plan := m.Plan()
			var entries []PlanEntry
			if len(plan.Entries) > 0 {
				entries = plan.Entries
			} else {
				for _, p := range m.Platforms {
					if m.Selected[p.ID] {
						entries = append(entries, PlanEntry{PlatformID: p.ID})
					}
				}
			}
			if len(entries) > 0 {
				idx := m.focusedIndex(len(entries))
				m.ToggleForceRebuild(entries[idx].PlatformID)
			}
			return m, nil
		case "m":
			m.CycleMode()
			return m, nil
		case "j", "down":
			m.Focused++
			return m, nil
		case "k", "up":
			m.Focused--
			return m, nil
		}
	case ConfirmPlanMsg:
		return m, func() tea.Msg { return StartBuildMsg{} }
	case CancelPlanMsg:
		return m, func() tea.Msg { return nav.NavigateBackMsg{} }
	case StartBuildMsg:
		req := engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetBuild}
		if blocked, reason := engine.IsActionBlockedByNix(req, m.nixState); blocked {
			m.nixBlockedReason = reason
			return m, nil
		}
		specs := m.RunSpecs()
		if len(specs) == 0 {
			return m, nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.BuildCancel = cancel
		return m, func() tea.Msg {
			return BuildCompleteMsg{Summary: ExecuteSequential(ctx, specs)}
		}
	case BuildCompleteMsg:
		m.Summary = msg.Summary
		m.BuildCancel = nil
		return m, nil
	case CancelBuildMsg:
		if m.BuildCancel != nil {
			m.BuildCancel()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) focusedIndex(n int) int {
	if n <= 0 {
		return 0
	}
	return (m.Focused%n + n) % n
}

func (m *Model) focusedLogPlatformID() string {
	if m.Summary == nil || len(m.Summary.Rows) == 0 {
		return ""
	}
	idx := m.focusedIndex(len(m.Summary.Rows))
	return m.Summary.Rows[idx].PlatformID
}

func (m *Model) logPathForPlatform(platformID string) string {
	return filepath.Join(m.Status.Root, "tmp", "augury-node-tui", platformID+".log")
}

func (m *Model) logContentForPlatform(platformID string) string {
	data, err := os.ReadFile(m.logPathForPlatform(platformID))
	if err != nil {
		return ""
	}
	return string(data)
}

func (m *Model) displayLogContent(platformID string) string {
	raw := m.logContentForPlatform(platformID)
	if m.LogTab == "error" {
		lineIdx, ok := logs.FindFirstErrorLine(raw)
		if !ok {
			return raw
		}
		return logs.ExtractContextAround(raw, lineIdx, errorContextBefore, errorContextAfter)
	}
	return raw
}

func (m *Model) resetScrollForPlatform(platformID string) {
	if m.LogScrollOffset == nil {
		m.LogScrollOffset = make(map[string]int)
	}
	m.LogScrollOffset[platformID] = 0
}

func (m *Model) scrollOffsetForPlatform(platformID string) int {
	if m.LogScrollOffset == nil {
		return 0
	}
	return m.LogScrollOffset[platformID]
}

func (m *Model) adjustScroll(platformID string, delta int) {
	if m.LogScrollOffset == nil {
		m.LogScrollOffset = make(map[string]int)
	}
	content := m.displayLogContent(platformID)
	lines := strings.Split(content, "\n")
	maxScroll := 0
	if len(lines) > 1 {
		maxScroll = len(lines) - 1
	}
	cur := m.LogScrollOffset[platformID]
	cur += delta
	if cur < 0 {
		cur = 0
	}
	if cur > maxScroll {
		cur = maxScroll
	}
	m.LogScrollOffset[platformID] = cur
}

func (m *Model) View() string {
	if m.Summary != nil && len(m.Summary.Rows) > 0 {
		return m.viewLogResults()
	}
	return m.viewPreflightPlan()
}

func (m *Model) viewLogResults() string {
	var b strings.Builder
	pid := m.focusedLogPlatformID()
	b.WriteString("Build results | t/tab switch full/error | e jump to error | j/k pgup/pgdown scroll | Esc/b back\n")
	b.WriteString(fmt.Sprintf("tab: %s | platform: %s\n", m.LogTab, pid))
	for _, r := range m.Summary.Rows {
		cur := " "
		if r.PlatformID == pid {
			cur = ">"
		}
		b.WriteString(fmt.Sprintf(" %s %s %s\n", cur, r.PlatformID, r.Status))
	}
	content := m.displayLogContent(pid)
	lines := strings.Split(content, "\n")
	offset := m.scrollOffsetForPlatform(pid)
	if offset >= len(lines) {
		offset = 0
		if m.LogScrollOffset == nil {
			m.LogScrollOffset = make(map[string]int)
		}
		m.LogScrollOffset[pid] = offset
	}
	visible := lines[offset:]
	b.WriteString("--- log ---\n")
	b.WriteString(strings.Join(visible, "\n"))
	if len(visible) == 0 && content != "" {
		b.WriteString("(scroll up)")
	}
	return b.String()
}

func (m *Model) viewPreflightPlan() string {
	plan := m.Plan()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Build pre-flight | mode: %s\n", plan.Mode))
	if m.nixBlockedReason != "" {
		b.WriteString(fmt.Sprintf("Blocked: %s\n", m.nixBlockedReason))
	}
	b.WriteString("m cycle mode | f toggle force rebuild | Enter confirm | Esc/b back\n")
	b.WriteString("platforms in plan:\n")
	var entries []PlanEntry
	if len(plan.Entries) > 0 {
		entries = plan.Entries
	} else {
		for _, p := range m.Platforms {
			if m.Selected[p.ID] {
				entries = append(entries, PlanEntry{PlatformID: p.ID})
			}
		}
	}
	focused := m.focusedIndex(len(entries))
	for i, e := range entries {
		force := ""
		if plan.ForceRebuild[e.PlatformID] {
			force = " [force]"
		}
		cur := " "
		if i == focused {
			cur = ">"
		}
		present := "?"
		if e.LocalArtifactPresent != nil {
			if *e.LocalArtifactPresent {
				present = "yes"
			} else {
				present = "no"
			}
		}
		b.WriteString(fmt.Sprintf(" %s %s %s%s\n", cur, e.PlatformID, present, force))
	}
	if len(entries) == 0 {
		b.WriteString("  (none selected)\n")
	}
	return b.String()
}

func (m *Model) Plan() *Plan {
	var selected []platform.Platform
	for _, p := range m.Platforms {
		if m.Selected[p.ID] {
			selected = append(selected, p)
		}
	}
	return BuildPlan(m.Status.Root, selected, m.Mode, m.ForceRebuild)
}

func (m *Model) ToggleForceRebuild(id string) {
	if m.ForceRebuild == nil {
		m.ForceRebuild = make(map[string]bool)
	}
	m.ForceRebuild[id] = !m.ForceRebuild[id]
}

func (m *Model) CycleMode() {
	switch m.Mode {
	case run.ModeSmart:
		m.Mode = run.ModeClean
	case run.ModeClean:
		m.Mode = run.ModeValidationOnly
	case run.ModeValidationOnly:
		m.Mode = run.ModeSmart
	default:
		m.Mode = run.ModeSmart
	}
}

func (m *Model) RunSpecs() []run.RunSpec {
	if m.Mode == run.ModeValidationOnly {
		return nil
	}
	var specs []run.RunSpec
	for _, p := range m.Platforms {
		if !m.Selected[p.ID] {
			continue
		}
		scriptPath := filepath.Join(m.Status.Root, p.ScriptRelPath)
		specs = append(specs, run.RunSpec{
			Name:    p.ID,
			Root:    m.Status.Root,
			Mode:    m.Mode,
			Command: "sh",
			Args:    []string{scriptPath},
		})
	}
	return specs
}
