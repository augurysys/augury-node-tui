package flash

import (
	"fmt"
	"path/filepath"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/styles"
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
	cursor           int // For platform/method selection
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

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case statePlatformSelect:
		return m.handlePlatformSelectKeys(msg)
	case stateMethodSelect:
		return m.handleMethodSelectKeys(msg)
	default:
		return m, nil
	}
}

func (m *Model) handlePlatformSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.Platforms)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		// Select platform
		if m.cursor >= 0 && m.cursor < len(m.Platforms) {
			return m, func() tea.Msg {
				return PlatformSelectedMsg{PlatformID: m.Platforms[m.cursor].ID}
			}
		}
	}
	return m, nil
}

func (m *Model) handleMethodSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement method selection keys
	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	switch m.state {
	case statePlatformSelect:
		return m.viewPlatformSelect()
	case stateMethodSelect:
		return m.viewMethodSelect()
	case stateFlashing:
		return m.viewFlashing()
	case stateComplete:
		return m.viewComplete()
	case stateError:
		return m.viewError()
	default:
		return "Loading..."
	}
}

func (m *Model) viewPlatformSelect() string {
	content := styles.Title.Render("Platform Selection") + "\n\n"

	if len(m.Platforms) == 0 {
		content += "No platforms available.\n"
		content += "Build a platform first.\n"
	} else {
		content += "Select platform to flash:\n\n"

		for i, p := range m.Platforms {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			ptype := DetectPlatformType(p)
			content += fmt.Sprintf("%s %s  [%s]\n", cursor, p.ID, ptype)
		}

		content += fmt.Sprintf("\nImage: %s\n", m.imagePath())
	}

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash"},
		Context:    "",
		Content:    content,
		ActionKeys: []components.KeyBinding{
			{Key: "enter", Label: "select"},
		},
		NavKeys: []components.KeyBinding{
			{Key: "j/k", Label: "navigate"},
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}

	return layout.Render()
}

func (m *Model) viewMethodSelect() string {
	return "Method selection coming soon"
}

func (m *Model) viewFlashing() string {
	return "Flashing..."
}

func (m *Model) viewComplete() string {
	return "Flash complete!"
}

func (m *Model) viewError() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	return "Unknown error"
}

func (m *Model) imagePath() string {
	if m.cursor < 0 || m.cursor >= len(m.Platforms) {
		return ""
	}
	p := m.Platforms[m.cursor]
	if p.OutputRelPath == "" {
		return ""
	}
	return filepath.Join(m.Status.Root, p.OutputRelPath)
}
