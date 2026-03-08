package validations

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
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

type Model struct {
	Status       status.RepoStatus
	presetStatus map[string]string
}

func NewModel(st status.RepoStatus) *Model {
	return &Model{
		Status:       st,
		presetStatus: make(map[string]string),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	nix := engine.ProbeNix(m.Status.Root)
	if blocked, reason := engine.IsActionBlockedByNix(req, nix); blocked {
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
	b.WriteString("Validations\n")
	b.WriteString("1 all | 2 shellcheck | 3 bats | 4 parse-test\n")
	b.WriteString("Presets: all, shellcheck-only, bats-only, parse-test-only\n")
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
		b.WriteString("  " + p + ": " + statusStr + "\n")
	}
	return b.String()
}
