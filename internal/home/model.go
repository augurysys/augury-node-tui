package home

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/data/developerdownloads"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Status             status.RepoStatus
	Platforms          []platform.Platform
	Selected           map[string]bool
	Focused            int
	DeveloperDownloads *developerdownloads.Index
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	sel := make(map[string]bool)
	for _, p := range platforms {
		sel[p.ID] = false
	}
	idx, _ := developerdownloads.ReadAt(st.Root)
	return &Model{Status: st, Platforms: platforms, Selected: sel, DeveloperDownloads: idx}
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
			return m, func() tea.Msg { return nav.ReplaySplashMsg{} }
		case "q":
			return m, tea.Quit
		case "b":
			return m, func() tea.Msg { return nav.NavigateMsg{Route: "build"} }
		case "h":
			return m, func() tea.Msg { return nav.NavigateMsg{Route: "hydrate"} }
		case "c":
			return m, func() tea.Msg { return nav.NavigateMsg{Route: "caches"} }
		case "v":
			return m, func() tea.Msg { return nav.NavigateMsg{Route: "validations"} }
		case "o":
			return m, func() tea.Msg { return nav.NavigateMsg{Route: "hints"} }
		case "j", "down":
			m.Focused = (m.Focused + 1) % max(1, len(m.Platforms))
			return m, nil
		case "k", "up":
			m.Focused = (m.Focused - 1 + len(m.Platforms)) % max(1, len(m.Platforms))
			return m, nil
		case " ", "space", "enter":
			if len(m.Platforms) > 0 {
				id := m.Platforms[m.Focused].ID
				m.TogglePlatform(id)
				return m, nil
			}
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
	b.WriteString("platforms: j/k up/down space/enter toggle\n")
	if m.DeveloperDownloads == nil {
		b.WriteString("developer-downloads: unavailable\n")
	}
	for i, p := range m.Platforms {
		sel := " "
		if m.Selected[p.ID] {
			sel = "x"
		}
		cur := " "
		if i == m.Focused {
			cur = ">"
		}
		line := fmt.Sprintf(" %s [%s] %s", cur, sel, p.ID)
		if m.DeveloperDownloads != nil {
			state := m.DeveloperDownloads.SourceState(p.ID)
			if state != "" {
				line += fmt.Sprintf(" (%s)", state)
			}
		}
		b.WriteString(line + "\n")
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
