package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type InstallStepModel struct {
	repoRoot         string
	builtBinary      string
	alreadyInstalled bool
	state            string
	confirmed        bool
	buildErr         string
	clipboardStatus  string
	installErr       string
}

func NewInstallStep(repoRoot string) *InstallStepModel {
	return &InstallStepModel{
		repoRoot: repoRoot,
		state:    "building",
	}
}

func (m *InstallStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		binDir := filepath.Join(m.repoRoot, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return BinaryBuiltMsg{Err: fmt.Sprintf("create bin dir: %v", err)}
		}

		binaryPath := filepath.Join(binDir, "augury-node-tui")

		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/augury-node-tui")
		cmd.Dir = m.repoRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			return BinaryBuiltMsg{Err: fmt.Sprintf("build failed: %v\n%s", err, output)}
		}

		targetPath := "/usr/local/bin/augury-node-tui"
		target, err := os.Readlink(targetPath)
		alreadyInstalled := err == nil && target == binaryPath

		return BinaryBuiltMsg{
			Binary:           binaryPath,
			AlreadyInstalled: alreadyInstalled,
		}
	}
}

func (m *InstallStepModel) Update(msg tea.Msg) (*InstallStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case BinaryBuiltMsg:
		if msg.Err != "" {
			m.state = "error"
			m.buildErr = msg.Err
			return m, nil
		}
		m.builtBinary = msg.Binary
		m.alreadyInstalled = msg.AlreadyInstalled
		m.state = "ready"

		if m.alreadyInstalled {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil

	case ClipboardCopiedMsg:
		if msg.Success {
			m.clipboardStatus = "Command copied to clipboard"
		} else {
			m.clipboardStatus = "Failed to copy to clipboard"
		}
		return m, nil

	case InstallCompleteMsg:
		if msg.Err != "" {
			m.installErr = msg.Err
		} else {
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
			cmd := "sudo ln -sf " + m.builtBinary + " /usr/local/bin/augury-node-tui"
			m.clipboardStatus = ""
			return m, copyToClipboard(cmd)
		case "i":
			m.installErr = ""
			return m, m.autoInstall()
		case "r":
			m.clipboardStatus = ""
			m.installErr = ""
			m.state = "building"
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

func (m *InstallStepModel) autoInstall() tea.Cmd {
	return func() tea.Msg {
		targetPath := "/usr/local/bin/augury-node-tui"
		cmd := exec.Command("sudo", "ln", "-sf", m.builtBinary, targetPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return InstallCompleteMsg{Err: fmt.Sprintf("install failed: %v\n%s", err, output)}
		}
		return InstallCompleteMsg{}
	}
}

func (m *InstallStepModel) View() string {
	var lines []string

	header := styles.Header.Render("Binary Installation")
	lines = append(lines, header)
	lines = append(lines, "")

	if m.state == "building" {
		lines = append(lines, "  "+styles.Dim.Render("Building binary..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	if m.state == "error" {
		lines = append(lines, "  "+styles.Error.Render("Build failed:"))
		lines = append(lines, "  "+m.buildErr)
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("r", "Retry")+"  "+styles.KeyBinding("s", "Skip")+"  "+styles.KeyBinding("q", "Quit"))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	lines = append(lines, "  Built: "+styles.Success.Render(m.builtBinary))
	lines = append(lines, "")

	if m.alreadyInstalled {
		lines = append(lines, "  "+styles.Success.Render("Already installed to /usr/local/bin"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Continue"))
	} else {
		targetPath := "/usr/local/bin/augury-node-tui"
		lines = append(lines, "  Installation options:")
		lines = append(lines, "")
		lines = append(lines, "  1. System-wide (requires sudo):")
		cmd := "sudo ln -sf " + m.builtBinary + " " + targetPath
		lines = append(lines, "     "+styles.Dim.Render(cmd))
		lines = append(lines, "")
		lines = append(lines, "  2. User PATH: Add this to your shell config:")
		lines = append(lines, "     "+styles.Dim.Render("export PATH=\""+filepath.Dir(m.builtBinary)+":$PATH\""))
		lines = append(lines, "")

		if m.clipboardStatus != "" {
			lines = append(lines, "  "+styles.Info.Render(m.clipboardStatus))
			lines = append(lines, "")
		}

		if m.installErr != "" {
			lines = append(lines, "  "+styles.Error.Render(m.installErr))
			lines = append(lines, "")
		}

		keys := []string{
			styles.KeyBinding("i", "Auto-install"),
			styles.KeyBinding("c", "Copy cmd"),
			styles.KeyBinding("s", "Skip"),
			styles.KeyBinding("q", "Quit"),
		}
		lines = append(lines, "  "+strings.Join(keys, "  "))
	}

	return styles.Border.Render(strings.Join(lines, "\n"))
}

func (m *InstallStepModel) Confirmed() bool {
	return m.confirmed
}

type BinaryBuiltMsg struct {
	Binary           string
	AlreadyInstalled bool
	Err              string
}

type InstallCompleteMsg struct {
	Err string
}
