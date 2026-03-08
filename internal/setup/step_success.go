package setup

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type SuccessStepModel struct {
	state     string
	confirmed bool
}

func NewSuccessStep() *SuccessStepModel {
	return &SuccessStepModel{
		state:     "ready",
		confirmed: true,
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
	var lines []string

	header := styles.Header.Render("✓ Setup Complete")
	lines = append(lines, header)
	lines = append(lines, "")
	lines = append(lines, "  "+styles.Success.Render("✓")+" All setup steps completed successfully.")
	lines = append(lines, "")
	lines = append(lines, "  "+styles.Dim.Render("Next steps:"))
	lines = append(lines, "")
	lines = append(lines, "  Run "+styles.Info.Render("augury-node-tui")+" to start the application.")
	lines = append(lines, "")
	lines = append(lines, "  "+styles.KeyBinding("enter", "Quit"))

	return styles.Border.Render(strings.Join(lines, "\n"))
}

func (m *SuccessStepModel) Confirmed() bool {
	return m.confirmed
}
