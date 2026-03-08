package hydration

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type DryRunRow struct {
	PlatformID    string
	LocalPresent  bool
	PlannedSource string
}

type Model struct {
	Status    status.RepoStatus
	Platforms []platform.Platform
	Selected  map[string]bool
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, selected map[string]bool) *Model {
	return &Model{Status: st, Platforms: platforms, Selected: selected}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) DryRunRows() []DryRunRow {
	var rows []DryRunRow
	for _, p := range m.Platforms {
		if !m.Selected[p.ID] {
			continue
		}
		scriptPath := filepath.Join(m.Status.Root, "scripts", "hydrate")
		localPresent := false
		if _, err := os.Stat(scriptPath); err == nil {
			localPresent = true
		}
		plannedSource := "branch+sha"
		if !localPresent {
			plannedSource = "not available"
		}
		rows = append(rows, DryRunRow{
			PlatformID:    p.ID,
			LocalPresent:  localPresent,
			PlannedSource: plannedSource,
		})
	}
	return rows
}

func (m *Model) CommandDispatch(platformID string) (run.RunSpec, bool) {
	p, ok := platform.ByID(platformID)
	if !ok {
		return run.RunSpec{}, false
	}
	scriptPath := filepath.Join(m.Status.Root, "scripts", "hydrate")
	if _, err := os.Stat(scriptPath); err != nil {
		return run.RunSpec{}, false
	}
	return run.RunSpec{
		Name:    p.ID,
		Root:    m.Status.Root,
		Mode:    run.ModeSmart,
		Command: "sh",
		Args:    []string{scriptPath, "--platform", p.ID},
	}, true
}

func (m *Model) View() string {
	var b strings.Builder
	rows := m.DryRunRows()
	if len(rows) == 0 {
		b.WriteString("No platforms selected.\n")
		return b.String()
	}
	available := false
	for _, r := range rows {
		if r.PlannedSource != "not available" {
			available = true
			break
		}
	}
	if !available {
		b.WriteString("Hydration not available.\n")
		b.WriteString("scripts/hydrate not found in target repo.\n")
		return b.String()
	}
	b.WriteString("platform | local | source\n")
	for _, r := range rows {
		local := "no"
		if r.LocalPresent {
			local = "yes"
		}
		b.WriteString(r.PlatformID + " | " + local + " | " + r.PlannedSource + "\n")
	}
	return b.String()
}
