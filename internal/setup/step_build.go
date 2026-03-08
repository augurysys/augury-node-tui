package setup

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type BuildStepModel struct {
	repoPath    string
	state       string
	buildOutput string
	buildError  string
	confirmed   bool
}

func NewBuildStep(repoPath string) *BuildStepModel {
	return &BuildStepModel{
		repoPath: repoPath,
		state:   "idle",
	}
}

func (m *BuildStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Stub: just set state to building for now
		return BuildOutputMsg{Output: ""}
	}
}

func (m *BuildStepModel) Update(msg tea.Msg) (*BuildStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case BuildOutputMsg:
		m.state = "building"
		m.buildOutput += msg.Output
		return m, nil

	case BuildCompleteMsg:
		if msg.Success {
			m.confirmed = true
			m.state = "complete"
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		m.buildError = msg.Error
		m.state = "failed"
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "r":
			if m.state == "failed" {
				m.state = "building"
				m.buildError = ""
				m.buildOutput = ""
				return m, func() tea.Msg {
					// Stub: will implement actual build in orchestration
					return BuildCompleteMsg{Success: true}
				}
			}
		}
	}
	return m, nil
}

func (m *BuildStepModel) View() string {
	var lines []string

	header := styles.Header.Render("🔨 Nix Build")
	lines = append(lines, header)
	lines = append(lines, "")

	if m.state == "building" || m.state == "complete" {
		lines = append(lines, "  "+styles.Dim.Render("Building")+"...")
		if m.buildOutput != "" {
			lines = append(lines, "")
			lines = append(lines, "  "+styles.Dim.Render(m.buildOutput))
		}
		if m.state == "building" {
			lines = append(lines, "")
			lines = append(lines, "  "+styles.KeyBinding("q", "Cancel"))
		}
	}

	if m.state == "failed" {
		lines = append(lines, "  "+styles.Error.Render("Build failed"))
		if m.buildError != "" {
			lines = append(lines, "")
			lines = append(lines, "  "+styles.Error.Render(m.buildError))
		}
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("r", "Retry")+" • "+styles.KeyBinding("q", "Quit"))
	}

	return styles.Border.Render(strings.Join(lines, "\n"))
}

func (m *BuildStepModel) Confirmed() bool {
	return m.confirmed
}

type BuildOutputMsg struct {
	Output string
}

type BuildCompleteMsg struct {
	Success bool
	Error   string
}
