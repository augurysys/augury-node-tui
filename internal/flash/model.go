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

	case PlatformSelectedMsg:
		return m.handlePlatformSelected(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case statePlatformSelect:
		return m.handlePlatformSelectKeys(msg)
	case stateMethodSelect:
		return m.handleMethodSelectKeys(msg)
	case stateFlashing:
		return m.handleFlashingKeys(msg)
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
	if m.adapter == nil {
		return m, nil
	}

	methods := m.adapter.GetMethods()

	switch msg.String() {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Parse number key
		idx := int(msg.Runes[0] - '1')
		if idx >= 0 && idx < len(methods) {
			m.selectedMethod = methods[idx].ID
			m.state = stateFlashing
			return m, nil
		}
	case "j", "down":
		if m.cursor < len(methods)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.cursor >= 0 && m.cursor < len(methods) {
			m.selectedMethod = methods[m.cursor].ID
			m.state = stateFlashing
			return m, nil
		}
	case "esc":
		m.state = statePlatformSelect
		m.cursor = 0
	}

	return m, nil
}

func (m *Model) handleFlashingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if m.selectedMethod != "" {
			m.selectedMethod = ""
			m.state = stateMethodSelect
		} else {
			m.adapter = nil
			m.state = statePlatformSelect
			m.cursor = 0
		}
	}
	return m, nil
}

func (m *Model) handlePlatformSelected(msg PlatformSelectedMsg) (tea.Model, tea.Cmd) {
	m.selectedPlatform = msg.PlatformID

	// Find platform
	var selectedPlatform *platform.Platform
	for i := range m.Platforms {
		if m.Platforms[i].ID == msg.PlatformID {
			selectedPlatform = &m.Platforms[i]
			break
		}
	}

	if selectedPlatform == nil {
		m.state = stateError
		m.err = fmt.Errorf("platform not found: %s", msg.PlatformID)
		return m, nil
	}

	outputPath := filepath.Join(m.Status.Root, selectedPlatform.OutputRelPath)

	// Detect platform type and create adapter
	ptype := DetectPlatformType(*selectedPlatform)
	switch ptype {
	case PlatformTypeMP255:
		m.adapter = NewMP255Adapter(m.Status.Root, selectedPlatform.ID, outputPath)
		// Validate prerequisites
		if err := m.adapter.CanFlash(outputPath); err != nil {
			m.state = stateError
			m.err = err
			return m, nil
		}
		// MP255 needs method selection
		m.state = stateMethodSelect
		m.cursor = 0 // Reset cursor for method selection

	case PlatformTypeSWUpdate:
		adapter, err := NewSWUpdateAdapter(m.Status.Root, selectedPlatform.ID, outputPath)
		if err != nil {
			m.state = stateError
			m.err = err
			return m, nil
		}
		m.adapter = adapter
		// Validate prerequisites
		if err := m.adapter.CanFlash(outputPath); err != nil {
			m.state = stateError
			m.err = err
			return m, nil
		}
		// SWUpdate goes straight to flashing
		m.state = stateFlashing

	default:
		m.state = stateError
		m.err = fmt.Errorf("unsupported platform type: %s", ptype)
	}

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
	if m.adapter == nil {
		return "No adapter"
	}

	methods := m.adapter.GetMethods()
	if len(methods) == 0 {
		return "No methods available"
	}

	content := styles.Title.Render("Choose Flash Method") + "\n\n"
	content += fmt.Sprintf("Platform: %s\n\n", m.selectedPlatform)
	content += "Select method:\n\n"

	for i, method := range methods {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		content += fmt.Sprintf("%s %d) %s\n", cursor, i+1, method.Name)
	}

	if m.cursor >= 0 && m.cursor < len(methods) {
		content += fmt.Sprintf("\n%s\n", methods[m.cursor].Description)
	}

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash", m.selectedPlatform},
		Context:    "",
		Content:    content,
		ActionKeys: []components.KeyBinding{
			{Key: "1/2", Label: "choose"},
			{Key: "enter", Label: "confirm"},
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

func (m *Model) viewFlashing() string {
	if m.adapter == nil {
		return "No adapter"
	}

	steps := m.adapter.GetSteps(m.selectedMethod)

	content := styles.Title.Render("Flashing Firmware") + "\n\n"
	content += fmt.Sprintf("Platform: %s\n", m.selectedPlatform)
	if m.selectedMethod != "" {
		content += fmt.Sprintf("Method: %s\n\n", m.selectedMethod)
	} else {
		content += "\n"
	}

	for i, step := range steps {
		content += fmt.Sprintf("%d. %s\n", i+1, step.Description)
	}

	content += "\n\nFlashing will be implemented in next tasks."

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash", m.selectedPlatform},
		Context:    "Ready to flash",
		Content:    content,
		ActionKeys: []components.KeyBinding{},
		NavKeys: []components.KeyBinding{
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}

	return layout.Render()
}

func (m *Model) viewComplete() string {
	content := styles.Title.Render("Flash Complete!") + "\n\n"
	content += fmt.Sprintf("Platform: %s\n", m.selectedPlatform)
	if m.selectedMethod != "" {
		content += fmt.Sprintf("Method: %s\n", m.selectedMethod)
	}
	content += "\n"
	content += styles.Success.Render("✓ Firmware flashed successfully")

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash", m.selectedPlatform},
		Context:    "Success",
		Content:    content,
		ActionKeys: []components.KeyBinding{},
		NavKeys: []components.KeyBinding{
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}

	return layout.Render()
}

func (m *Model) viewError() string {
	errMsg := "Unknown error"
	if m.err != nil {
		errMsg = m.err.Error()
	}

	content := styles.Title.Render("Flash Error") + "\n\n"
	content += styles.Error.Render(errMsg)

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash"},
		Context:    "Error",
		Content:    content,
		ActionKeys: []components.KeyBinding{},
		NavKeys: []components.KeyBinding{
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}

	return layout.Render()
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
