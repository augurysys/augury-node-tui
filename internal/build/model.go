package build

import (
	"path/filepath"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmPlanMsg struct{}
type CancelPlanMsg struct{}
type StartBuildMsg struct{}
type NavigateBackMsg struct{}

type Model struct {
	Status   status.RepoStatus
	Platforms []platform.Platform
	Selected map[string]bool
	Mode     run.Mode
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, selected map[string]bool) *Model {
	m := &Model{
		Status:    st,
		Platforms: platforms,
		Selected:  selected,
		Mode:      run.ModeSmart,
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
	switch msg.(type) {
	case ConfirmPlanMsg:
		return m, func() tea.Msg { return StartBuildMsg{} }
	case CancelPlanMsg:
		return m, func() tea.Msg { return NavigateBackMsg{} }
	}
	return m, nil
}

func (m *Model) View() string {
	return ""
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
