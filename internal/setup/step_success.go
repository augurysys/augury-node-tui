package setup

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	tea "github.com/charmbracelet/bubbletea"
)

type SuccessStepModel struct {
	state       string
	confirmed   bool
	skippedSteps []string
}

func NewSuccessStep(skippedSteps []string) *SuccessStepModel {
	return &SuccessStepModel{
		state:        "ready",
		confirmed:    true,
		skippedSteps: skippedSteps,
	}
}

func (m *SuccessStepModel) Init() tea.Cmd {
	return nil
}

func (m *SuccessStepModel) Update(msg tea.Msg) (*SuccessStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *SuccessStepModel) View() string {
	// Build content
	var content strings.Builder
	content.WriteString("✓ Setup completed successfully!\n\n")
	content.WriteString("Next steps:\n")
	content.WriteString("  1. Run commands from the TUI\n")
	content.WriteString("  2. Explore build/hydrate/caches screens\n")

	if len(m.skippedSteps) > 0 {
		content.WriteString("\n⚠ Some steps were skipped:\n")
		for _, step := range m.skippedSteps {
			content.WriteString("  - " + step + "\n")
		}
	}

	// Use Card component
	card := primitives.Card{
		Title:   "Setup Complete",
		Content: content.String(),
		Style:   primitives.CardEmphasized,
	}

	cardView := card.Render(80)

	// Add key hints using KeyHint component
	launchHint := primitives.KeyHint{
		Key:         "enter",
		Description: "Launch main TUI",
		Enabled:     true,
	}
	quitHint := primitives.KeyHint{
		Key:         "q",
		Description: "Quit",
		Enabled:     true,
	}

	keyHints := "\n" + launchHint.Render() + "  •  " + quitHint.Render()

	return cardView + keyHints
}

func (m *SuccessStepModel) Confirmed() bool {
	return m.confirmed
}
