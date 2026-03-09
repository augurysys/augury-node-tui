package validations

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/visual/diagram"
	tea "github.com/charmbracelet/bubbletea"
)

var presets = []string{"all", "shellcheck-only", "bats-only", "parse-test-only"}

var presetCommands = map[string]struct {
	script string
	args   []string
}{
	"all":             {"scripts/validate-all.sh", nil},
	"shellcheck-only":  {"scripts/validate-shellcheck.sh", nil},
	"bats-only":       {"scripts/validate-bats.sh", nil},
	"parse-test-only": {"scripts/validate-parse-test.sh", nil},
}

var keyToPreset = map[string]string{
	"1": "all",
	"2": "shellcheck-only",
	"3": "bats-only",
	"4": "parse-test-only",
}

var presetToRequest = map[string]engine.ActionRequest{
	"all":             engine.ValidationsAll,
	"shellcheck-only": engine.ValidationsShellcheck,
	"bats-only":       engine.ValidationsBats,
	"parse-test-only": engine.ValidationsParse,
}

type jobResultMsg struct {
	Job     engine.Job
	Request engine.ActionRequest
}

// Validation represents a single validation row for display
type Validation struct {
	Name    string
	Status  string
	Message string
}

type Model struct {
	Status       status.RepoStatus
	nixState     engine.NixState
	presetStatus map[string]string
	validations  []Validation // optional override; when set, used instead of presetStatus
	Width        int
	Height       int
}

func NewModel(st status.RepoStatus) *Model {
	return &Model{
		Status:       st,
		nixState:     engine.ProbeNix(st.Root),
		presetStatus: make(map[string]string),
	}
}

func (m *Model) SetNixState(nix engine.NixState) {
	m.nixState = nix
}

func (m *Model) SetValidations(v []Validation) {
	m.validations = v
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
		k := msg.String()
		if preset, ok := keyToPreset[k]; ok {
			return m, m.dispatchPreset(preset)
		}
		return m, nil
	case jobResultMsg:
		m.handleJobResult(msg.Job, msg.Request)
		return m, nil
	}
	return m, nil
}

func (m *Model) dispatchPreset(preset string) tea.Cmd {
	req, ok := presetToRequest[preset]
	if !ok {
		return nil
	}
	cap := engine.ResolveCapability(m.Status.Root, req)
	if !cap.Available {
		m.presetStatus[preset] = "not-available"
		if cap.Reason != "" {
			m.presetStatus[preset] += ": " + cap.Reason
		}
		return nil
	}
	if blocked, reason := engine.IsActionBlockedByNix(req, m.nixState); blocked {
		m.presetStatus[preset] = "blocked"
		if reason != "" {
			m.presetStatus[preset] += ": " + reason
		}
		return nil
	}
	return m.dispatchAction(req)
}

func (m *Model) dispatchAction(req engine.ActionRequest) tea.Cmd {
	root := m.Status.Root
	return func() tea.Msg {
		job := engine.ExecuteAction(context.Background(), root, req)
		return jobResultMsg{Job: job, Request: req}
	}
}

func (m *Model) handleJobResult(job engine.Job, req engine.ActionRequest) {
	preset := requestToPreset(req)
	if preset == "" {
		return
	}
	m.presetStatus[preset] = string(job.State)
	if job.Reason != "" {
		m.presetStatus[preset] += ": " + job.Reason
	}
	if job.LogPath != "" {
		m.presetStatus[preset] += " (log: " + filepath.Base(job.LogPath) + ")"
	}
}

func requestToPreset(req engine.ActionRequest) string {
	for p, r := range presetToRequest {
		if r.ID() == req.ID() {
			return p
		}
	}
	return ""
}

func (m *Model) CommandForPreset(preset string) (run.RunSpec, bool) {
	cmd, ok := presetCommands[preset]
	if !ok {
		return run.RunSpec{}, false
	}
	scriptPath := filepath.Join(m.Status.Root, cmd.script)
	if _, err := os.Stat(scriptPath); err != nil {
		return run.RunSpec{}, false
	}
	args := []string{scriptPath}
	if len(cmd.args) > 0 {
		args = append(args, cmd.args...)
	}
	return run.RunSpec{
		Name:    preset,
		Root:    m.Status.Root,
		Mode:    run.ModeSmart,
		Command: "sh",
		Args:    args,
	}, true
}

func (m *Model) View() string {
	var b strings.Builder
	if m.Width >= diagram.MinDiagramWidth {
		b.WriteString(diagram.ValidationPipeline())
		b.WriteString("\n")
	}
	b.WriteString("Validations\n")
	b.WriteString("1 all | 2 shellcheck | 3 bats | 4 parse-test\n")
	b.WriteString("Presets: all, shellcheck-only, bats-only, parse-test-only\n")

	rows := m.buildRows()
	columns := []components.Column{
		{
			Header: "Name",
			Width:  30,
			Renderer: func(r interface{}) string {
				return r.(Validation).Name
			},
		},
		{
			Header: "Status",
			Width:  15,
			Renderer: func(r interface{}) string {
				v := r.(Validation)
				badge := primitives.StatusBadge{
					Label:  v.Status,
					Status: statusFromString(v.Status),
				}
				return badge.Render()
			},
		},
		{
			Header: "Message",
			Width:  50,
			Renderer: func(r interface{}) string {
				return r.(Validation).Message
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
	for i, v := range rows {
		rowInterfaces[i] = v
	}
	table.SetRows(rowInterfaces)

	b.WriteString(table.View())
	return b.String()
}

func (m *Model) buildRows() []Validation {
	if len(m.validations) > 0 {
		return m.validations
	}
	rows := make([]Validation, 0, len(presets))
	for _, p := range presets {
		statusStr := m.presetStatus[p]
		if statusStr == "" {
			_, ok := m.CommandForPreset(p)
			if ok {
				statusStr = "available"
			} else {
				statusStr = "not-available"
			}
		}
		rows = append(rows, Validation{Name: p, Status: statusStr, Message: ""})
	}
	return rows
}

func statusFromString(s string) primitives.Status {
	switch strings.ToLower(s) {
	case "pass", "ok", "success", "available":
		return primitives.StatusSuccess
	case "fail", "error", "failed", "not-available":
		return primitives.StatusError
	case "warn", "warning":
		return primitives.StatusWarning
	case "running", "pending":
		return primitives.StatusRunning
	case "blocked":
		return primitives.StatusBlocked
	default:
		return primitives.StatusUnavailable
	}
}
