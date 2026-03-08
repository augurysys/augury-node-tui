package hydration

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

const keyLegend = "D dry-run | H hydrate | b/esc back"

type DryRunRow struct {
	PlatformID    string
	LocalPresent  bool
	PlannedSource string
}

type Model struct {
	Status    status.RepoStatus
	Platforms []platform.Platform
	Selected  map[string]bool
	nixState  engine.NixState
	rowStatus map[string]string
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, selected map[string]bool) *Model {
	return &Model{
		Status:    st,
		Platforms: platforms,
		Selected:  selected,
		nixState:  engine.ProbeNix(st.Root),
		rowStatus: make(map[string]string),
	}
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
		k := strings.ToUpper(msg.String())
		if k == "D" {
			return m, m.dispatchDryRun()
		}
		if k == "H" {
			return m, m.dispatchHydration()
		}
		return m, nil
	case jobResultMsg:
		m.handleJobResult(msg.Job, msg.Request)
		return m, nil
	}
	return m, nil
}

type jobResultMsg struct {
	Job     engine.Job
	Request engine.ActionRequest
}

func (m *Model) selectedPlatformIDs() []string {
	var ids []string
	for _, p := range m.Platforms {
		if m.Selected[p.ID] {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func (m *Model) dispatchDryRun() tea.Cmd {
	ids := m.selectedPlatformIDs()
	if len(ids) == 0 {
		return nil
	}
	cap := engine.ResolveCapability(m.Status.Root, engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetDryRun})
	if !cap.Available {
		for _, id := range ids {
			m.rowStatus[id] = "not-available"
		}
		return nil
	}
	if blocked, _ := engine.IsActionBlockedByNix(engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetDryRun}, m.nixState); blocked {
		for _, id := range ids {
			m.rowStatus[id] = "blocked"
		}
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(ids))
	for _, id := range ids {
		req := engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetDryRun, PlatformID: id}
		cmds = append(cmds, m.dispatchAction(req))
	}
	return tea.Batch(cmds...)
}

func (m *Model) dispatchHydration() tea.Cmd {
	ids := m.selectedPlatformIDs()
	if len(ids) == 0 {
		return nil
	}
	cap := engine.ResolveCapability(m.Status.Root, engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetRun})
	if !cap.Available {
		for _, id := range ids {
			m.rowStatus[id] = "not-available"
		}
		return nil
	}
	if blocked, _ := engine.IsActionBlockedByNix(engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetRun}, m.nixState); blocked {
		for _, id := range ids {
			m.rowStatus[id] = "blocked"
		}
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(ids))
	for _, id := range ids {
		req := engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetRun, PlatformID: id}
		cmds = append(cmds, m.dispatchAction(req))
	}
	return tea.Batch(cmds...)
}

func (m *Model) dispatchAction(req engine.ActionRequest) tea.Cmd {
	root := m.Status.Root
	return func() tea.Msg {
		job := engine.ExecuteAction(context.Background(), root, req)
		return jobResultMsg{Job: job, Request: req}
	}
}

func (m *Model) handleJobResult(job engine.Job, req engine.ActionRequest) {
	key := req.PlatformID
	if key == "" {
		key = "global"
	}
	m.rowStatus[key] = string(job.State)
	if job.Reason != "" {
		m.rowStatus[key] += ": " + job.Reason
	}
}

func (m *Model) RowStatus(platformID string) string {
	return m.rowStatus[platformID]
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
	b.WriteString(keyLegend + "\n")
	b.WriteString("platform | script | source | status\n")
	for _, r := range rows {
		script := "no"
		if r.LocalPresent {
			script = "yes"
		}
		s := m.RowStatus(r.PlatformID)
		b.WriteString(r.PlatformID + " | " + script + " | " + r.PlannedSource + " | " + s + "\n")
	}
	return b.String()
}
