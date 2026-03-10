package setup

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type CircleCIStepModel struct {
	currentToken string
	userInput    string
	confirmed    bool
	skipped      bool
}

func NewCircleCIStep() *CircleCIStepModel {
	return &CircleCIStepModel{}
}

func NewCircleCIStepWithCurrent(currentToken string) *CircleCIStepModel {
	return &CircleCIStepModel{
		currentToken: currentToken,
	}
}

func (s *CircleCIStepModel) Init() tea.Cmd {
	return nil
}

func (s *CircleCIStepModel) Update(msg tea.Msg) (*CircleCIStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			token := strings.TrimSpace(s.userInput)
			s.confirmed = true
			if token == "" {
				if s.currentToken != "" {
					s.userInput = s.currentToken
				} else {
					s.skipped = true
				}
			}
			return s, func() tea.Msg { return NextStepMsg{} }
		case tea.KeyCtrlQ:
			return s, tea.Quit
		case tea.KeyRunes:
			if len(msg.Runes) == 1 && msg.Runes[0] == 'q' && s.userInput == "" && s.currentToken == "" {
				return s, tea.Quit
			}
			s.userInput += string(msg.Runes)
		case tea.KeyBackspace:
			runes := []rune(s.userInput)
			if len(runes) > 0 {
				s.userInput = string(runes[:len(runes)-1])
			}
		}
	}
	return s, nil
}

func (s *CircleCIStepModel) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Step 6: CircleCI Token (Optional)"))
	b.WriteString("\n\n")

	if s.currentToken != "" {
		b.WriteString(styles.Info.Render("Current: "))
		b.WriteString(strings.Repeat("*", len(s.currentToken)))
		b.WriteString("\n")
		b.WriteString(styles.Dim.Render("Press Enter on empty field to keep current value"))
		b.WriteString("\n\n")
	}

	b.WriteString("Enter your CircleCI personal API token for CI dashboard access.\n")
	if s.currentToken == "" {
		b.WriteString(styles.Dim.Render("Leave blank and press Enter to skip."))
	}
	b.WriteString("\n\n")

	masked := strings.Repeat("*", len(s.userInput))
	b.WriteString(styles.Border.Render(masked))

	return b.String()
}

func (s *CircleCIStepModel) Confirmed() bool {
	return s.confirmed
}

func (s *CircleCIStepModel) Skipped() bool {
	return s.skipped
}

func (s *CircleCIStepModel) Token() string {
	return strings.TrimSpace(s.userInput)
}
