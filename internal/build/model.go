package build

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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

type Model struct {
	Status         status.RepoStatus
	Platforms      []platform.Platform
	Selected       map[string]bool
	Mode           run.Mode
	ForceRebuild   map[string]bool
	Summary        *Summary
	BuildCancel    context.CancelFunc
	Focused        int
	LogTab         string
	LogScrollOffset int
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, selected map[string]bool) *Model {
	m := &Model{
		Status:         st,
		Platforms:      platforms,
		Selected:      selected,
		Mode:          run.ModeSmart,
		ForceRebuild:  make(map[string]bool),
		LogTab:        "full",
	}
	if m.Selected == nil {
		m.Selected = make(map[string]bool)
	}
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		if m.Summary != nil && m.focusedLogPlatformID() != "" {
			switch s {
			case "t":
				if m.LogTab == "full" {
					m.LogTab = "error"
				} else {
					m.LogTab = "full"
				}
				return m, nil
			case "e":
				m.LogTab = "error"
				return m, nil
			case "j", "down":
				m.LogScrollOffset++
				if m.LogScrollOffset < 0 {
					m.LogScrollOffset = 0
				}
				return m, nil
			case "k", "up":
				m.LogScrollOffset--
				if m.LogScrollOffset < 0 {
					m.LogScrollOffset = 0
				}
				return m, nil
			}
		}
		switch s {
		case "enter":
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

func (m *Model) View() string {
	plan := m.Plan()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Build pre-flight | mode: %s\n", plan.Mode))
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
