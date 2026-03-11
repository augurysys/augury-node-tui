package flash

import (
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateIdle state = iota
	statePlatformSelect
	stateMethodSelect
	stateFlashing
	stateComplete
	stateError
)

// Model is the flash screen model
type Model struct {
	Status           status.RepoStatus
	Platforms        []platform.Platform
	Width            int
	Height           int
	state            state
	selectedPlatform string
	selectedMethod   string
	adapter          FlashAdapter
	err              error
}

// NewModel creates a new flash model
func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	return &Model{
		Status:    st,
		Platforms: platforms,
		state:     stateIdle,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Transition from idle to platform select on first update
	if m.state == stateIdle {
		m.state = statePlatformSelect
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	return "Flash screen placeholder"
}
