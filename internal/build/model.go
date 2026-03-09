package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmPlanMsg struct{}
type CancelPlanMsg struct{}
type StartBuildMsg struct{}
type CancelBuildMsg struct{}

type BuildCompleteMsg struct {
	Summary *Summary
}

type Model struct {
	Status           status.RepoStatus
	Platforms        []platform.Platform
	Selected         map[string]bool
	nixState         engine.NixState
	nixBlockedReason string
	Mode             run.Mode
	ForceRebuild     map[string]bool
	Summary         *Summary
	BuildCancel     context.CancelFunc
	Focused         int
	Width           int
	Height          int
	parallelTracker components.ParallelTracker
	metricsBar      components.MetricsBar
	commandDisplay  components.CommandDisplay
	logViewer       *components.LogViewer
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, selected map[string]bool) *Model {
	m := &Model{
		Status:       st,
		Platforms:    platforms,
		Selected:     selected,
		nixState:     engine.ProbeNix(st.Root),
		Mode:         run.ModeSmart,
		ForceRebuild: make(map[string]bool),
		Width:        80,
		Height:       24,
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
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateComponentDimensions()
		return m, nil
	case tea.KeyMsg:
		s := msg.String()
		if m.Summary != nil && len(m.Summary.Rows) > 0 {
			if m.logViewer != nil {
				switch s {
				case "]": // next platform
					m.Focused++
					m.Focused = m.focusedIndex(len(m.Summary.Rows))
					m.syncLogViewerContent()
					m.refreshBuildComponents()
					return m, nil
				case "[": // prev platform
					m.Focused--
					m.Focused = m.focusedIndex(len(m.Summary.Rows))
					m.syncLogViewerContent()
					m.refreshBuildComponents()
					return m, nil
				case "j", "k", "down", "up", "pgup", "pgdown", "e", "n", "N":
					return m, m.logViewer.Update(msg)
				}
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
		m.initLogViewer()
		m.refreshBuildComponents()
		m.updateComponentDimensions()
		fetchMetrics := func() tea.Msg {
			_ = m.metricsBar.FetchMetrics()
			return nil
		}
		if m.logViewer != nil {
			return m, tea.Batch(m.logViewer.Init(), fetchMetrics)
		}
		return m, fetchMetrics
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

func (m *Model) summaryRowsToBuildLanes() []components.BuildLane {
	if m.Summary == nil {
		return nil
	}
	lanes := make([]components.BuildLane, len(m.Summary.Rows))
	for i, r := range m.Summary.Rows {
		st := primitives.StatusSuccess
		progress := 1.0
		switch r.Status {
		case RowStatusSuccess:
			st = primitives.StatusSuccess
		case RowStatusFailure:
			st = primitives.StatusError
		case RowStatusSkipped:
			st = primitives.StatusUnavailable
			progress = 0
		case RowStatusCancelled:
			st = primitives.StatusBlocked
			progress = 0
		default:
			st = primitives.StatusUnavailable
		}
		lanes[i] = components.BuildLane{
			Platform: r.PlatformID,
			Progress: progress,
			Status:   st,
			Current:  string(r.Status),
		}
	}
	return lanes
}

func (m *Model) initLogViewer() {
	pid := m.focusedLogPlatformID()
	if pid == "" {
		return
	}
	content := m.logContentForPlatform(pid)
	m.logViewer = components.NewLogViewer(content)
}

func (m *Model) syncLogViewerContent() {
	if m.logViewer == nil {
		return
	}
	pid := m.focusedLogPlatformID()
	if pid == "" {
		return
	}
	content := m.logContentForPlatform(pid)
	m.logViewer.SetContent(content)
}

func (m *Model) refreshBuildComponents() {
	if m.Summary == nil {
		return
	}
	m.parallelTracker.Lanes = m.summaryRowsToBuildLanes()
	pid := m.focusedLogPlatformID()
	m.commandDisplay = m.buildCommandDisplay(pid)
}

func (m *Model) updateComponentDimensions() {
	if m.Width <= 0 {
		return
	}
	leftWidth := (m.Width * 3) / 10
	middleWidth := (m.Width * 3) / 10
	rightWidth := m.Width - leftWidth - middleWidth - 2

	m.parallelTracker.Width = leftWidth
	m.parallelTracker.Height = m.Height
	m.metricsBar.Width = middleWidth
	if m.logViewer != nil {
		m.logViewer.SetWidth(rightWidth)
		if m.Height > 0 {
			m.logViewer.SetHeight(m.Height - 2)
		}
	}
}

func (m *Model) View() string {
	if m.Summary != nil && len(m.Summary.Rows) > 0 {
		return m.viewLogResults()
	}
	return m.viewPreflightPlan()
}

func (m *Model) viewLogResults() string {
	leftWidth := (m.Width * 3) / 10
	middleWidth := (m.Width * 3) / 10
	rightWidth := m.Width - leftWidth - middleWidth - 2

	leftPane := m.parallelTracker.Render()
	middlePane := lipgloss.JoinVertical(lipgloss.Left,
		m.metricsBar.Render(),
		"",
		m.commandDisplay.Render(),
	)

	var rightPane string
	if m.logViewer != nil {
		rightPane = m.logViewer.View()
	} else {
		rightPane = "(no log)"
	}

	border := styles.Border
	return lipgloss.JoinHorizontal(lipgloss.Top,
		border.Width(leftWidth).Render(leftPane),
		border.Width(middleWidth).Render(middlePane),
		border.Width(rightWidth).Render(rightPane),
	)
}

func (m *Model) buildCommandDisplay(platformID string) components.CommandDisplay {
	cd := components.CommandDisplay{
		Command:   "",
		Executing: false,
	}
	if platformID == "" {
		return cd
	}
	for _, p := range m.Platforms {
		if p.ID == platformID {
			scriptPath := filepath.Join(m.Status.Root, p.ScriptRelPath)
			cd.Command = "sh " + scriptPath
			break
		}
	}
	if m.Summary != nil {
		for _, r := range m.Summary.Rows {
			if r.PlatformID == platformID {
				switch r.Status {
				case RowStatusSuccess:
					exit0 := 0
					cd.ExitCode = &exit0
				case RowStatusFailure:
					exit1 := 1
					cd.ExitCode = &exit1
				}
				break
			}
		}
	}
	return cd
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
