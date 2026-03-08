package home

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type ReplaySplashMsg struct{}

type NavigateMsg struct {
	Route string
}

type Model struct {
	Status    status.RepoStatus
	Platforms []platform.Platform
	Selected  map[string]bool
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	sel := make(map[string]bool)
	for _, p := range platforms {
		sel[p.ID] = false
	}
	return &Model{Status: st, Platforms: platforms, Selected: sel}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "a":
			return m, func() tea.Msg { return ReplaySplashMsg{} }
		case "q":
			return m, tea.Quit
		case "b":
			return m, func() tea.Msg { return NavigateMsg{Route: "build"} }
		case "h":
			return m, func() tea.Msg { return NavigateMsg{Route: "hydrate"} }
		case "c":
			return m, func() tea.Msg { return NavigateMsg{Route: "caches"} }
		case "v":
			return m, func() tea.Msg { return NavigateMsg{Route: "validations"} }
		case "o":
			return m, func() tea.Msg { return NavigateMsg{Route: "hints"} }
		}
	}
	return m, nil
}

func (m *Model) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("root: %s\n", m.Status.Root))
	b.WriteString(fmt.Sprintf("branch: %s\n", m.Status.Branch))
	b.WriteString(fmt.Sprintf("sha: %s\n", m.Status.SHA))
	b.WriteString("paths:\n")
	for _, p := range status.RequiredPaths {
		dirty := m.Status.Dirty[p]
		label := "clean"
		if dirty {
			label = "dirty"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", p, label))
	}
	b.WriteString("platforms:\n")
	for _, p := range m.Platforms {
		sel := " "
		if m.Selected[p.ID] {
			sel = "x"
		}
		b.WriteString(fmt.Sprintf("  [%s] %s\n", sel, p.ID))
	}
	b.WriteString("b build h hydrate c caches v validations o hints a replay q quit\n")
	return b.String()
}

func (m *Model) TogglePlatform(id string) {
	if _, ok := m.Selected[id]; ok {
		m.Selected[id] = !m.Selected[id]
	}
}

func (m *Model) IsPlatformSelected(id string) bool {
	return m.Selected[id]
}
