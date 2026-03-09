package hints

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Status    status.RepoStatus
	Platforms []platform.Platform
	NixReady  bool
	Width     int
	nixState  engine.NixState
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	return &Model{Status: st, Platforms: platforms}
}

func (m *Model) SetNixState(nix engine.NixState) {
	m.nixState = nix
	m.NixReady = nix.Ready
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		return m, nil
	}
	return m, nil
}

func (m *Model) View() string {
	// Build hints content
	var content strings.Builder

	content.WriteString("Navigation:\n")
	hints := []primitives.KeyHint{
		{Key: "j/k", Description: "Navigate up/down", Enabled: true},
		{Key: "tab", Description: "Switch panes", Enabled: true},
		{Key: "q", Description: "Back/Quit", Enabled: true},
	}
	for _, hint := range hints {
		content.WriteString("  " + hint.Render() + "\n")
	}

	content.WriteString("\nActions:\n")
	actionHints := []primitives.KeyHint{
		{Key: "b", Description: "Build platform", Enabled: m.NixReady},
		{Key: "h", Description: "Hydrate artifacts", Enabled: m.NixReady},
		{Key: "c", Description: "Manage caches", Enabled: true},
		{Key: "v", Description: "Run validations", Enabled: true},
	}
	for _, hint := range actionHints {
		content.WriteString("  " + hint.Render() + "\n")
	}

	// Use Card component
	card := primitives.Card{
		Title:   "Keyboard Shortcuts",
		Content: content.String(),
		Style:   primitives.CardNormal,
	}

	width := m.Width
	if width <= 0 {
		width = 80
	}
	return card.Render(width)
}
