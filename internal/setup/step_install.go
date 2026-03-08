package setup

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type InstallStepModel struct {
	sourceBinary     string
	alreadyInstalled bool
	state            string
	confirmed        bool
}

func NewInstallStep(binaryPath string) *InstallStepModel {
	return &InstallStepModel{
		sourceBinary: binaryPath,
		state:        "checking",
	}
}

func (m *InstallStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Stub: check if /usr/local/bin/augury-node-tui symlink exists and points to source
		return InstallCheckMsg{AlreadyInstalled: false}
	}
}

func (m *InstallStepModel) Update(msg tea.Msg) (*InstallStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case InstallCheckMsg:
		m.alreadyInstalled = msg.AlreadyInstalled
		m.state = "ready"

		if m.alreadyInstalled {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil

	case tea.KeyMsg:
		if m.state != "ready" {
			return m, nil
		}

		switch msg.String() {
		case "c":
			// Copy command to clipboard
			return m, nil
		case "r":
			// Re-check
			return m, m.Init()
		case "s":
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *InstallStepModel) View() string {
	var lines []string

	header := styles.Header.Render("📦 Binary Installation")
	lines = append(lines, header)
	lines = append(lines, "")

	if m.state == "checking" {
		lines = append(lines, "  "+styles.Dim.Render("Checking installation..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	targetPath := "/usr/local/bin/augury-node-tui"
	lines = append(lines, "  Target: "+styles.Info.Render(targetPath))
	lines = append(lines, "")

	if m.alreadyInstalled {
		lines = append(lines, "  "+styles.Success.Render("✓ Already installed"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Continue"))
	} else {
		lines = append(lines, "  "+styles.Dim.Render("Run this command to install:"))
		lines = append(lines, "")
		cmd := "sudo ln -sf " + m.sourceBinary + " " + targetPath
		lines = append(lines, "  "+styles.Info.Render(cmd))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("c", "Copy")+"  "+styles.KeyBinding("r", "Re-check")+"  "+styles.KeyBinding("s", "Skip")+"  "+styles.KeyBinding("q", "Quit"))
	}

	return styles.Border.Render(strings.Join(lines, "\n"))
}

func (m *InstallStepModel) Confirmed() bool {
	return m.confirmed
}

type InstallCheckMsg struct {
	AlreadyInstalled bool
}
