package setup

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type BuildStepModel struct {
	repoPath    string
	state       string
	buildOutput []string
	buildError  string
	confirmed   bool
	width       int
	lastLine    string
	spinner     int
	outputChan  chan string
}

func NewBuildStep(repoPath string) *BuildStepModel {
	return &BuildStepModel{
		repoPath:    repoPath,
		state:       "idle",
		buildOutput: []string{},
		width:       80,
	}
}

func (m *BuildStepModel) Init() tea.Cmd {
	m.outputChan = make(chan string, 100)
	return tea.Batch(
		m.runBuild(m.outputChan),
		m.listenForOutput(m.outputChan),
		m.tickSpinner(),
	)
}

func (m *BuildStepModel) runBuild(outputChan chan<- string) tea.Cmd {
	return func() tea.Msg {
		defer close(outputChan)

		cmd := exec.Command("nix", "build", "--print-build-logs")
		cmd.Dir = m.repoPath

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return BuildCompleteMsg{Success: false, Error: fmt.Sprintf("failed to create stdout pipe: %v", err)}
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return BuildCompleteMsg{Success: false, Error: fmt.Sprintf("failed to create stderr pipe: %v", err)}
		}

		if err := cmd.Start(); err != nil {
			return BuildCompleteMsg{Success: false, Error: fmt.Sprintf("failed to start nix build: %v", err)}
		}

		go streamOutput(stdout, outputChan)
		go streamOutput(stderr, outputChan)

		err = cmd.Wait()

		if err != nil {
			return BuildCompleteMsg{Success: false, Error: fmt.Sprintf("nix build failed: %v", err)}
		}
		return BuildCompleteMsg{Success: true}
	}
}

func (m *BuildStepModel) listenForOutput(outputChan <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-outputChan
		if !ok {
			return nil
		}
		return BuildOutputMsg{Output: line}
	}
}

func streamOutput(r io.Reader, ch chan<- string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ch <- scanner.Text()
	}
}

func (m *BuildStepModel) tickSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

func (m *BuildStepModel) Update(msg tea.Msg) (*BuildStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case SpinnerTickMsg:
		if m.state == "building" {
			m.spinner = (m.spinner + 1) % 4
			return m, m.tickSpinner()
		}

	case BuildOutputMsg:
		m.state = "building"
		line := strings.TrimSpace(msg.Output)
		if line != "" {
			m.lastLine = line
			if len(m.buildOutput) > 100 {
				m.buildOutput = m.buildOutput[1:]
			}
			m.buildOutput = append(m.buildOutput, line)
		}
		return m, m.listenForOutput(m.outputChan)

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
				m.state = "idle"
				m.buildError = ""
				m.buildOutput = []string{}
				m.lastLine = ""
				m.outputChan = make(chan string, 100)
				return m, tea.Batch(
					m.runBuild(m.outputChan),
					m.listenForOutput(m.outputChan),
					m.tickSpinner(),
				)
			}
		case "s":
			if m.state == "failed" {
				m.confirmed = true
				return m, func() tea.Msg { return NextStepMsg{} }
			}
		}
	}
	return m, nil
}

func (m *BuildStepModel) View() string {
	var b strings.Builder

	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinnerFrames[m.spinner%len(spinnerFrames)]

	header := styles.Header.Render("Nix Build")
	b.WriteString(header)
	b.WriteString("\n\n")

	maxWidth := m.width - 8
	if maxWidth < 40 {
		maxWidth = 40
	}

	switch m.state {
	case "idle":
		b.WriteString("  " + styles.Dim.Render("Preparing build...") + "\n")

	case "building":
		b.WriteString("  " + styles.Info.Render(spinner) + " " + styles.Bold.Render("Building augury-node with Nix...") + "\n")
		b.WriteString("  " + styles.Dim.Render("This may take several minutes on first run") + "\n\n")

		if m.lastLine != "" {
			displayLine := m.lastLine
			if len(displayLine) > maxWidth {
				displayLine = displayLine[:maxWidth-3] + "..."
			}
			b.WriteString("  " + styles.Dim.Render(displayLine) + "\n")
		}

		tailLines := 8
		startIdx := len(m.buildOutput) - tailLines
		if startIdx < 0 {
			startIdx = 0
		}
		if len(m.buildOutput) > 0 {
			b.WriteString("\n")
			for i := startIdx; i < len(m.buildOutput); i++ {
				line := m.buildOutput[i]
				if len(line) > maxWidth {
					line = line[:maxWidth-3] + "..."
				}
				b.WriteString("  " + styles.Dim.Render(line) + "\n")
			}
		}

		b.WriteString("\n  " + styles.KeyBinding("q", "Cancel") + "\n")

	case "complete":
		b.WriteString("  " + styles.Success.Render("✓ Build complete") + "\n")

	case "failed":
		b.WriteString("  " + styles.Error.Render("✗ Build failed") + "\n\n")
		if m.buildError != "" {
			errorLines := strings.Split(m.buildError, "\n")
			for _, line := range errorLines {
				if len(line) > maxWidth {
					line = line[:maxWidth-3] + "..."
				}
				b.WriteString("  " + styles.Error.Render(line) + "\n")
			}
		}

		tailLines := 10
		startIdx := len(m.buildOutput) - tailLines
		if startIdx < 0 {
			startIdx = 0
		}
		if len(m.buildOutput) > 0 {
			b.WriteString("\n  " + styles.Dim.Render("Last output lines:") + "\n")
			for i := startIdx; i < len(m.buildOutput); i++ {
				line := m.buildOutput[i]
				if len(line) > maxWidth {
					line = line[:maxWidth-3] + "..."
				}
				b.WriteString("  " + styles.Dim.Render(line) + "\n")
			}
		}

		b.WriteString("\n  " + styles.KeyBinding("r", "Retry") + "  " + styles.KeyBinding("s", "Skip") + "  " + styles.KeyBinding("q", "Quit") + "\n")
	}

	borderStyle := styles.Border
	if m.width > 0 {
		borderStyle = borderStyle.Width(maxWidth)
	}

	return borderStyle.Render(b.String())
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

type SpinnerTickMsg struct{}
