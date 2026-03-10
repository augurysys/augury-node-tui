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
	targetRepoRoot   string
	tuiRepoRoot      string
	builtBinary      string
	alreadyInstalled bool
	state            string
	confirmed        bool
	buildErr         string
	clipboardStatus  string
	installErr       string
	width            int
}

func NewInstallStep(targetRepoRoot string) *InstallStepModel {
	tuiRepoRoot, _ := os.Getwd()
	return &InstallStepModel{
		targetRepoRoot: targetRepoRoot,
		tuiRepoRoot:    tuiRepoRoot,
		state:          "building",
		width:          80,
	}
}

func (m *InstallStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		binDir := filepath.Join(m.targetRepoRoot, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return BinaryBuiltMsg{Err: fmt.Sprintf("create bin dir: %v", err)}
		}

		binaryPath := filepath.Join(binDir, "augury-node-tui")

		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/augury-node-tui")
		cmd.Dir = m.tuiRepoRoot
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
	case tea.WindowSizeMsg:
		m.width = msg.Width

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
		m.state = "ready"
		if msg.Err != "" {
			m.installErr = msg.Err
		} else {
			targetPath := "/usr/local/bin/augury-node-tui"
			if target, err := os.Readlink(targetPath); err == nil && target == m.builtBinary {
				m.alreadyInstalled = true
				m.confirmed = true
				return m, func() tea.Msg { return NextStepMsg{} }
			} else {
				m.installErr = "Symlink verification failed after install"
			}
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
			m.state = "installing"
			targetPath := "/usr/local/bin/augury-node-tui"
			return m, tea.ExecProcess(
				exec.Command("sudo", "ln", "-sf", m.builtBinary, targetPath),
				func(err error) tea.Msg {
					if err != nil {
						return InstallCompleteMsg{Err: fmt.Sprintf("installation failed: %v", err)}
					}
					return InstallCompleteMsg{}
				},
			)
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


func (m *InstallStepModel) View() string {
	var lines []string

	header := styles.Header.Render("Binary Installation")
	lines = append(lines, header)
	lines = append(lines, "")

	maxWidth := m.width - 8
	if maxWidth < 40 {
		maxWidth = 40
	}

	if m.state == "building" {
		lines = append(lines, "  "+styles.Dim.Render("Building binary..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	if m.state == "installing" {
		lines = append(lines, "  "+styles.Info.Render("Installing to /usr/local/bin..."))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.Dim.Render("TUI suspended for sudo prompt"))
		lines = append(lines, "  "+styles.Dim.Render("Enter your password in the terminal"))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	if m.state == "error" {
		lines = append(lines, "  "+styles.Error.Render("Build failed:"))
		lines = append(lines, "")
		
		errorLines := strings.Split(m.buildErr, "\n")
		for _, errLine := range errorLines {
			for len(errLine) > 0 {
				if len(errLine) <= maxWidth {
					lines = append(lines, "  "+styles.Dim.Render(errLine))
					break
				}
				lines = append(lines, "  "+styles.Dim.Render(errLine[:maxWidth]))
				errLine = errLine[maxWidth:]
			}
		}
		
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	binaryPath := m.builtBinary
	if len(binaryPath) > maxWidth-10 {
		binaryPath = "..." + binaryPath[len(binaryPath)-(maxWidth-13):]
	}
	lines = append(lines, "  Built: "+styles.Success.Render(binaryPath))
	lines = append(lines, "")

	if m.alreadyInstalled {
		lines = append(lines, "  "+styles.Success.Render("Already installed to /usr/local/bin"))
	} else {
		targetPath := "/usr/local/bin/augury-node-tui"
		lines = append(lines, "  Installation options:")
		lines = append(lines, "")
		lines = append(lines, "  1. System-wide (requires sudo):")
		
		cmd := "sudo ln -sf " + m.builtBinary + " " + targetPath
		if len(cmd) > maxWidth-5 {
			lines = append(lines, "     "+styles.Dim.Render("sudo ln -sf \\"))
			lines = append(lines, "       "+styles.Dim.Render(m.builtBinary+" \\"))
			lines = append(lines, "       "+styles.Dim.Render(targetPath))
		} else {
			lines = append(lines, "     "+styles.Dim.Render(cmd))
		}
		
		lines = append(lines, "")
		lines = append(lines, "  2. User PATH: Add this to your shell config:")
		pathCmd := "export PATH=\"" + filepath.Dir(m.builtBinary) + ":$PATH\""
		if len(pathCmd) > maxWidth-5 {
			lines = append(lines, "     "+styles.Dim.Render("export PATH=\""+filepath.Dir(m.builtBinary)+":$PATH\""))
		} else {
			lines = append(lines, "     "+styles.Dim.Render(pathCmd))
		}
		lines = append(lines, "")

		if m.clipboardStatus != "" {
			status := m.clipboardStatus
			if len(status) > maxWidth {
				status = status[:maxWidth-3] + "..."
			}
			lines = append(lines, "  "+styles.Info.Render(status))
			lines = append(lines, "")
		}

		if m.installErr != "" {
			errLines := strings.Split(m.installErr, "\n")
			for _, errLine := range errLines {
				for len(errLine) > 0 {
					if len(errLine) <= maxWidth {
						lines = append(lines, "  "+styles.Error.Render(errLine))
						break
					}
					lines = append(lines, "  "+styles.Error.Render(errLine[:maxWidth]))
					errLine = errLine[maxWidth:]
				}
			}
			lines = append(lines, "")
		}
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
