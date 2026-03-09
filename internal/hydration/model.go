package hydration

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

const keyLegend = "D dry-run | H hydrate | b/esc back"

// Artifact represents a hydration artifact for display in the table
type Artifact struct {
	Name     string
	Status   string
	Progress int
	Total    int
}

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
	artifacts []Artifact // optional override; when set, used instead of DryRunRows
	Width     int
	Height    int
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, selected map[string]bool) *Model {
	return &Model{
		Status:    st,
		Platforms: platforms,
		Selected:  selected,
		nixState:  engine.ProbeNix(st.Root),
		rowStatus: make(map[string]string),
		Width:     80,
		Height:    20,
	}
}

func (m *Model) SetArtifacts(a []Artifact) {
	m.artifacts = a
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
		return m, nil
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
	if blocked, reason := engine.IsActionBlockedByNix(engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetDryRun}, m.nixState); blocked {
		s := "blocked"
		if reason != "" {
			s = "blocked: " + reason
		}
		for _, id := range ids {
			m.rowStatus[id] = s
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
	if blocked, reason := engine.IsActionBlockedByNix(engine.ActionRequest{Kind: engine.KindHydration, Target: engine.TargetRun}, m.nixState); blocked {
		s := "blocked"
		if reason != "" {
			s = "blocked: " + reason
		}
		for _, id := range ids {
			m.rowStatus[id] = s
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
	b.WriteString(keyLegend + "\n")

	rows := m.buildRows()
	if len(rows) == 0 {
		b.WriteString("No platforms selected.\n")
		return b.String()
	}

	// When not using artifacts, check availability from DryRunRows
	if len(m.artifacts) == 0 {
		drRows := m.DryRunRows()
		available := false
		for _, r := range drRows {
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
	}

	columns := []components.Column{
		{
			Header: "Artifact",
			Width:  40,
			Renderer: func(r interface{}) string {
				return r.(Artifact).Name
			},
		},
		{
			Header: "Status",
			Width:  15,
			Renderer: func(r interface{}) string {
				art := r.(Artifact)
				primary := primaryStatus(art.Status)
				badge := primitives.StatusBadge{
					Label:  primary,
					Status: statusFromString(primary),
				}
				return badge.Render()
			},
		},
		{
			Header: "Message",
			Width:  40,
			Renderer: func(r interface{}) string {
				art := r.(Artifact)
				_, msg := splitStatusMessage(art.Status)
				return msg
			},
		},
		{
			Header: "Progress",
			Width:  40,
			Renderer: func(r interface{}) string {
				art := r.(Artifact)
				if art.Status == "downloading" || art.Status == "hydrating" {
					bar := primitives.ProgressBar{
						Current: art.Progress,
						Total:   art.Total,
						Width:   35,
						Label:   "",
					}
					return bar.Render()
				}
				return ""
			},
		},
	}

	table := components.NewDataTable(columns)
	table.SetWidth(m.Width)
	if m.Height > 0 {
		table.SetHeight(m.Height)
	} else {
		table.SetHeight(20)
	}

	rowInterfaces := make([]interface{}, len(rows))
	for i, a := range rows {
		rowInterfaces[i] = a
	}
	table.SetRows(rowInterfaces)

	b.WriteString(table.View())
	return b.String()
}

func (m *Model) buildRows() []Artifact {
	if len(m.artifacts) > 0 {
		return m.artifacts
	}
	var rows []Artifact
	for _, r := range m.DryRunRows() {
		s := m.RowStatus(r.PlatformID)
		if s == "" {
			s = "pending"
		}
		rows = append(rows, Artifact{
			Name:     r.PlatformID,
			Status:   s,
			Progress: 0,
			Total:    0,
		})
	}
	return rows
}

func primaryStatus(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, ": "); idx >= 0 {
		return s[:idx]
	}
	return s
}

func splitStatusMessage(s string) (string, string) {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, ": "); idx >= 0 {
		return s[:idx], strings.TrimSpace(s[idx+2:])
	}
	return s, ""
}

func statusFromString(s string) primitives.Status {
	switch strings.ToLower(s) {
	case "cached", "complete", "ok":
		return primitives.StatusSuccess
	case "downloading", "hydrating", "pending":
		return primitives.StatusRunning
	case "error", "failed":
		return primitives.StatusError
	case "blocked":
		return primitives.StatusBlocked
	case "missing", "unavailable", "not-available":
		return primitives.StatusUnavailable
	default:
		return primitives.StatusUnavailable
	}
}
