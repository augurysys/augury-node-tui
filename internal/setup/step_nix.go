package setup

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type NixStepModel struct {
	nixInstalled        bool
	experimentalEnabled bool
	daemonOk            bool
	state               string
	fixError            error
	confirmed           bool
}

func NewNixStep() *NixStepModel {
	return &NixStepModel{
		state: "checking",
	}
}

func (m *NixStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		return NixHealthCheckMsg{
			NixInstalled:        CheckNixInstalled(),
			ExperimentalEnabled: CheckNixExperimentalFeatures(),
			DaemonOk:            CheckDaemonSocket(),
		}
	}
}

func (m *NixStepModel) Update(msg tea.Msg) (*NixStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case NixHealthCheckMsg:
		m.nixInstalled = msg.NixInstalled.Available
		m.experimentalEnabled = msg.ExperimentalEnabled.Available
		m.daemonOk = msg.DaemonOk.Available
		m.state = "ready"

		if m.AllChecksPassed() {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil

	case NixFixResultMsg:
		m.state = "ready"
		if msg.Err != nil {
			m.fixError = msg.Err
			return m, nil
		}
		m.experimentalEnabled = true
		if m.AllChecksPassed() {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil

	case tea.KeyMsg:
		if m.state != "ready" {
			return m, nil
		}

		switch msg.String() {
		case "f":
			if !m.experimentalEnabled {
				m.state = "fixing"
				return m, func() tea.Msg {
					err := AutoFixNixConfig()
					return NixFixResultMsg{Err: err}
				}
			}
		case "s":
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *NixStepModel) View() string {
	var lines []string

	header := styles.Header.Render("⚙️  Nix Configuration")
	lines = append(lines, header)
	lines = append(lines, "")

	if m.state == "checking" {
		lines = append(lines, "  "+styles.Dim.Render("Checking Nix installation..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	if m.state == "fixing" {
		lines = append(lines, "  "+styles.Info.Render("🔧 Applying auto-fix..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	status1 := "✓"
	style1 := styles.Success
	if !m.nixInstalled {
		status1 = "✗"
		style1 = styles.Error
	}
	lines = append(lines, "  "+style1.Render(status1+" Nix installed"))

	status2 := "✓"
	style2 := styles.Success
	if !m.experimentalEnabled {
		status2 = "✗"
		style2 = styles.Error
	}
	lines = append(lines, "  "+style2.Render(status2+" Experimental features enabled"))

	status3 := "✓"
	style3 := styles.Success
	if !m.daemonOk {
		status3 = "✗"
		style3 = styles.Error
	}
	lines = append(lines, "  "+style3.Render(status3+" Nix daemon accessible"))

	lines = append(lines, "")

	if m.fixError != nil {
		lines = append(lines, "  "+styles.Error.Render("✗ Auto-fix failed: "+m.fixError.Error()))
		lines = append(lines, "")
	}

	if !m.experimentalEnabled {
		lines = append(lines, "  "+styles.KeyBinding("f", "Auto-fix experimental features")+"  "+styles.KeyBinding("s", "Skip"))
	} else if m.AllChecksPassed() {
		lines = append(lines, "  "+styles.Success.Render("✓ All checks passed"))
	} else {
		lines = append(lines, "  "+styles.KeyBinding("s", "Skip")+"  "+styles.KeyBinding("q", "Cancel"))
	}

	return styles.Border.Render(strings.Join(lines, "\n"))
}

func (m *NixStepModel) AllChecksPassed() bool {
	return m.nixInstalled && m.experimentalEnabled && m.daemonOk
}

func (m *NixStepModel) Confirmed() bool {
	return m.confirmed
}

type NixHealthCheckMsg struct {
	NixInstalled        HealthCheckResult
	ExperimentalEnabled HealthCheckResult
	DaemonOk            HealthCheckResult
}

type NixFixResultMsg struct {
	Err error
}

type NextStepMsg struct{}
