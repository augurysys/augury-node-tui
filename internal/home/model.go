package home

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/data/developerdownloads"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/augurysys/augury-node-tui/internal/visual/diagram"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Status                 status.RepoStatus
	Platforms              []platform.Platform
	Selected               map[string]bool
	Focused                int
	DeveloperDownloads     *developerdownloads.Index
	DeveloperDownloadsErr  error
	Width                  int
	nixState               engine.NixState
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	sel := make(map[string]bool)
	for _, p := range platforms {
		sel[p.ID] = false
	}
	idx, err := developerdownloads.ReadAt(st.Root)
	return &Model{Status: st, Platforms: platforms, Selected: sel, DeveloperDownloads: idx, DeveloperDownloadsErr: err}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		return m, nil
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
	var sections []string

	// Title
	title := styles.Title.Render("🚀 Augury Node Builder")
	sections = append(sections, title)

	// Diagram (if wide enough)
	if m.Width >= diagram.MinDiagramWidth {
		sections = append(sections, diagram.PlatformFlow(m.Platforms))
	}

	// Repository Status Section
	repoSection := m.renderRepoStatus()
	sections = append(sections, styles.Border.Render(repoSection))

	// Platforms Section
	platformsSection := m.renderPlatforms()
	sections = append(sections, styles.Section.Render(platformsSection))

	// Key Bindings
	keyHelp := m.renderKeyHelp()
	sections = append(sections, styles.KeyHelp.Render(keyHelp))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderRepoStatus() string {
	var lines []string

	// Header
	header := styles.Header.Render("📁 Repository")
	lines = append(lines, header)

	// Root
	lines = append(lines, fmt.Sprintf("  %s  %s",
		styles.Dim.Render("Root:"),
		styles.Highlight.Render(m.Status.Root)))

	// Branch
	lines = append(lines, fmt.Sprintf("  %s %s  %s %s",
		styles.Dim.Render("Branch:"),
		styles.Info.Render(m.Status.Branch),
		styles.Dim.Render("SHA:"),
		styles.Dim.Render(m.Status.SHA[:min(8, len(m.Status.SHA))])))

	// Nix Status
	var nixStatus string
	if m.nixState.Ready {
		nixStatus = fmt.Sprintf("  %s  %s",
			styles.Dim.Render("Nix:"),
			styles.Success.Render("✓ ready"))
	} else {
		reason := friendlyNixReason(m.nixState.Reason)
		if reason == "" {
			reason = "check setup"
		}
		nixStatus = fmt.Sprintf("  %s  %s\n        %s",
			styles.Dim.Render("Nix:"),
			styles.Error.Render("✗ not ready"),
			styles.Warning.Render("→ "+reason))
	}
	lines = append(lines, nixStatus)

	// Paths
	pathsClean := 0
	pathsDirty := 0
	for _, p := range status.RequiredPaths {
		if m.Status.Dirty[p] {
			pathsDirty++
		} else {
			pathsClean++
		}
	}
	var pathStatus string
	if pathsDirty > 0 {
		pathStatus = fmt.Sprintf("%s %d dirty, %d clean",
			styles.Warning.Render("⚠"),
			pathsDirty, pathsClean)
	} else {
		pathStatus = styles.Success.Render(fmt.Sprintf("✓ %d paths clean", pathsClean))
	}
	lines = append(lines, fmt.Sprintf("  %s  %s",
		styles.Dim.Render("Paths:"),
		pathStatus))

	return strings.Join(lines, "\n")
}

func (m *Model) renderPlatforms() string {
	var lines []string

	// Header
	header := styles.Header.Render("🎯 Platforms")
	hint := styles.Dim.Render(" (j/k: navigate • space: toggle)")
	lines = append(lines, header+hint)

	// Developer downloads status
	if m.DeveloperDownloads == nil {
		lines = append(lines, "  "+styles.Warning.Render("⚠ developer-downloads unavailable"))
	}

	// Platform list
	for i, p := range m.Platforms {
		line := m.renderPlatformItem(i, p)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderPlatformItem(index int, p platform.Platform) string {
	// Selection checkbox
	checkbox := "☐"
	if m.Selected[p.ID] {
		checkbox = "☑"
	}

	// Platform ID
	idText := p.ID

	// Source state indicator
	var stateText string
	if m.DeveloperDownloads != nil {
		state := m.DeveloperDownloads.SourceState(p.ID)
		if state != "" {
			stateStyle := styles.StatusStyle(string(state))
			switch state {
			case "built":
				stateText = stateStyle.Render("● built")
			case "hydrated":
				stateText = stateStyle.Render("● hydrated")
			case "missing":
				stateText = stateStyle.Render("○ missing")
			default:
				stateText = styles.Dim.Render("○ " + string(state))
			}
		}
	}

	// Build the line
	content := fmt.Sprintf("  %s %s %s", checkbox, idText, stateText)

	// Apply selection styling
	if index == m.Focused {
		return styles.ItemSelected.Render("▶ " + content)
	}
	return styles.Item.Render("  " + content)
}

func (m *Model) renderKeyHelp() string {
	keys := []string{
		styles.KeyBinding("b", "build"),
		styles.KeyBinding("h", "hydrate"),
		styles.KeyBinding("c", "caches"),
		styles.KeyBinding("v", "validations"),
		styles.KeyBinding("o", "hints"),
		styles.KeyBinding("a", "replay"),
		styles.KeyBinding("r", "refresh"),
		styles.KeyBinding("q", "quit"),
	}
	return strings.Join(keys, "  •  ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *Model) TogglePlatform(id string) {
	if _, ok := m.Selected[id]; ok {
		m.Selected[id] = !m.Selected[id]
	}
}

func (m *Model) IsPlatformSelected(id string) bool {
	return m.Selected[id]
}

func (m *Model) SetNixState(nix engine.NixState) {
	m.nixState = nix
}

func friendlyNixReason(reason string) string {
	if reason == "" {
		return ""
	}
	lower := strings.ToLower(reason)
	if strings.Contains(lower, "experimental") && strings.Contains(lower, "nix-command") {
		return "run: nix develop .#dev-env (enable experimental features)"
	}
	if strings.Contains(lower, "permission denied") && strings.Contains(lower, "daemon-socket") {
		return "run: sudo systemctl restart nix-daemon (or reboot)"
	}
	if strings.Contains(lower, "timed out") || strings.Contains(lower, "timeout") {
		return "probe timed out"
	}
	if strings.Contains(lower, "not found") || strings.Contains(lower, "command not found") {
		return "command not found in PATH"
	}
	if len(reason) > 50 {
		return reason[:47] + "..."
	}
	return reason
}
