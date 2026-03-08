package setup

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/augurysys/augury-node-tui/internal/styles"
)

type RootStep struct {
	detectedPath string
	userInput    string
	confirmed    bool
}

func NewRootStep(detectedPath string) *RootStep {
	return &RootStep{
		detectedPath: detectedPath,
		userInput:    detectedPath,
	}
}

func (s *RootStep) Update(msg tea.Msg) (*RootStep, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			s.confirmed = true
			return s, func() tea.Msg { return RootConfirmedMsg{Path: s.GetRootPath()} }
		case tea.KeyRunes:
			s.userInput += string(msg.Runes)
		case tea.KeyBackspace:
			if len(s.userInput) > 0 {
				s.userInput = s.userInput[:len(s.userInput)-1]
			}
		}
	}
	return s, nil
}

func (s *RootStep) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Step 1: Augury Node Root"))
	b.WriteString("\n\n")

	if s.detectedPath != "" {
		b.WriteString(styles.Success.Render("✓ Auto-detected: "))
		b.WriteString(s.detectedPath)
		b.WriteString("\n\n")
	}

	b.WriteString("Enter path to augury-node repository:\n")
	b.WriteString(styles.Border.Render(s.userInput))
	b.WriteString("\n\n")
	b.WriteString(styles.Dim.Render("Press Enter to confirm"))

	return b.String()
}

func (s *RootStep) GetRootPath() string {
	return s.userInput
}

type RootConfirmedMsg struct {
	Path string
}
