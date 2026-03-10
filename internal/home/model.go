package home

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/data/developerdownloads"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/augurysys/augury-node-tui/internal/visual/diagram"
	tea "github.com/charmbracelet/bubbletea"
)

// PlatformEntry holds platform data for the DataTable
type PlatformEntry struct {
	ID         string
	State      string // source state: built, hydrated, missing, or ""
	OutputPath string
	Selected   bool
}

type Model struct {
	Status                status.RepoStatus
	Platforms             []platform.Platform
	Selected              map[string]bool
	DeveloperDownloads    *developerdownloads.Index
	DeveloperDownloadsErr error
	Width                 int
	Height                int
	nixState              engine.NixState
	platformTable         *components.DataTable
	metricsBar            components.MetricsBar
	showMetrics           bool
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	sel := make(map[string]bool)
	for _, p := range platforms {
		sel[p.ID] = false
	}
	idx, err := developerdownloads.ReadAt(st.Root)
	m := &Model{
		Status:                st,
		Platforms:             platforms,
		Selected:              sel,
		DeveloperDownloads:    idx,
		DeveloperDownloadsErr: err,
		metricsBar:            components.MetricsBar{},
	}
	m.initPlatformTable()
	return m
}

func (m *Model) initPlatformTable() {
	columns := []components.Column{
		{Header: "☐", Width: 5, Sortable: false, Renderer: m.renderCheckbox},
		{Header: "Platform", Width: 20, Sortable: true, Renderer: m.renderPlatformID},
		{Header: "State", Width: 12, Sortable: true, Renderer: m.renderState},
		{Header: "Path", Width: -1, Sortable: true, Renderer: m.renderOutputPath},
	}
	m.platformTable = components.NewDataTable(columns)
	m.platformTable.SetRows(m.fetchPlatformData())
}

func (m *Model) renderCheckbox(row interface{}) string {
	e := row.(PlatformEntry)
	if e.Selected {
		return styles.CheckboxSelected.Render("[●]")
	}
	return styles.CheckboxUnselected.Render("[ · ]")
}

func (m *Model) renderPlatformID(row interface{}) string {
	return row.(PlatformEntry).ID
}

func (m *Model) renderState(row interface{}) string {
	e := row.(PlatformEntry)
	if e.State == "" {
		return "-"
	}
	st := primitives.StatusSuccess
	switch e.State {
	case "built":
		st = primitives.StatusSuccess
	case "hydrated":
		st = primitives.StatusRunning
	case "missing":
		st = primitives.StatusUnavailable
	default:
		st = primitives.StatusWarning
	}
	return primitives.StatusBadge{Label: e.State, Status: st}.Render()
}

func (m *Model) renderOutputPath(row interface{}) string {
	return row.(PlatformEntry).OutputPath
}

func (m *Model) buildActionKeys() []components.KeyBinding {
	var keys []components.KeyBinding

	// Dynamic keys based on selection
	keys = append(keys, components.KeyBinding{Key: "space", Label: "select"})

	selectedCount := m.countSelected()
	if selectedCount > 0 {
		keys = append(keys, components.KeyBinding{Key: "b", Label: "build"})
		keys = append(keys, components.KeyBinding{Key: "h", Label: "hydrate"})
	}

	keys = append(keys, components.KeyBinding{Key: "c", Label: "caches"})
	keys = append(keys, components.KeyBinding{Key: "v", Label: "validations"})
	keys = append(keys, components.KeyBinding{Key: "o", Label: "hints"})
	keys = append(keys, components.KeyBinding{Key: "p", Label: "pipeline"})

	return keys
}

func (m *Model) countSelected() int {
	count := 0
	for _, selected := range m.Selected {
		if selected {
			count++
		}
	}
	return count
}

// fetchPlatformData converts platforms to table rows.
// Note: Platform table is refreshed only on toggle. If DeveloperDownloads changes
// externally (e.g. after build/hydrate), the table state will be stale until the
// next toggle or refresh.
func (m *Model) fetchPlatformData() []interface{} {
	rows := make([]interface{}, 0, len(m.Platforms))
	for _, p := range m.Platforms {
		state := ""
		if m.DeveloperDownloads != nil {
			state = string(m.DeveloperDownloads.SourceState(p.ID))
		}
		rows = append(rows, PlatformEntry{
			ID:         p.ID,
			State:      state,
			OutputPath: p.OutputRelPath,
			Selected:   m.Selected[p.ID],
		})
	}
	return rows
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.metricsBar.Width = msg.Width
		m.platformTable.SetWidth(msg.Width)

		if msg.Height > 0 {
			// Calculate remaining height for table (title, diagram, card, metrics, header, key help)
			reservedHeight := 22
			if msg.Width < diagram.MinDiagramWidth {
				reservedHeight -= 8 // No diagram if narrow
			}
			if !m.showMetrics {
				reservedHeight -= 1 // No metrics bar if disabled
			}
			tableHeight := msg.Height - reservedHeight
			if tableHeight < 5 {
				tableHeight = 5
			}
			m.platformTable.SetHeight(tableHeight)
		} else {
			m.platformTable.SetHeight(20)
		}
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
		case "p":
			return m, func() tea.Msg { return nav.NavigateMsg{Route: "ci"} }
		case "j", "down", "k", "up":
			m.platformTable.Update(msg)
			return m, nil
		case " ", "space", "enter":
			row := m.platformTable.SelectedRow()
			if row != nil {
				id := row.(PlatformEntry).ID
				m.TogglePlatform(id)
				m.platformTable.SetRows(m.fetchPlatformData())
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *Model) View() string {
	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home"},
		Context:    m.buildContext(),
		Content:    m.renderContent(),
		ActionKeys: m.buildActionKeys(),
		NavKeys: []components.KeyBinding{
			{Key: "j/k", Label: "navigate"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}
	return layout.Render()
}

func (m *Model) buildContext() string {
	parts := []string{m.Status.Branch}

	selectedCount := m.countSelected()
	if selectedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d selected", selectedCount))
	}

	return strings.Join(parts, "  •  ")
}

func (m *Model) renderContent() string {
	var sections []string

	// Repo status card (bordered box - already using Card component)
	repoCard := primitives.Card{
		Title:   "📁 Repository",
		Content: m.renderRepoStatus(),
		Style:   primitives.CardNormal,
	}
	width := m.Width
	if width <= 0 {
		width = 80
	}
	sections = append(sections, repoCard.Render(width))

	// Platform table section (no border)
	platformHeader := styles.Header.Render("🎯 Platforms")
	hint := styles.Dim.Render(" (j/k: navigate • space: toggle)")
	if m.DeveloperDownloads == nil {
		hint += "  " + styles.Warning.Render("⚠ developer-downloads unavailable")
	}
	platformSection := platformHeader + hint + "\n" + m.platformTable.View()
	sections = append(sections, platformSection)

	return strings.Join(sections, "\n\n")
}

func (m *Model) renderRepoStatus() string {
	lines := []string{
		fmt.Sprintf("Root: %s", m.Status.Root),
		fmt.Sprintf("Branch: %s  SHA: %s", m.Status.Branch, m.Status.SHA[:min(8, len(m.Status.SHA))]),
	}

	nixStatus := primitives.StatusSuccess
	if !m.nixState.Ready {
		nixStatus = primitives.StatusError
	}
	nixBadge := primitives.StatusBadge{
		Label:  nixStatusLabel(m.nixState),
		Status: nixStatus,
	}
	lines = append(lines, nixBadge.Render())
	if !m.nixState.Ready {
		if reason := friendlyNixReason(m.nixState.Reason); reason != "" {
			lines = append(lines, "  → "+reason)
		}
	}

	pathsClean := 0
	pathsDirty := 0
	for _, p := range status.RequiredPaths {
		if m.Status.Dirty[p] {
			pathsDirty++
		} else {
			pathsClean++
		}
	}
	pathStatus := primitives.StatusSuccess
	pathLabel := fmt.Sprintf("%d paths clean", pathsClean)
	if pathsDirty > 0 {
		pathStatus = primitives.StatusWarning
		pathLabel = fmt.Sprintf("%d dirty, %d clean", pathsDirty, pathsClean)
	}
	pathBadge := primitives.StatusBadge{Label: pathLabel, Status: pathStatus}
	lines = append(lines, pathBadge.Render())

	return strings.Join(lines, "\n")
}

func nixStatusLabel(nix engine.NixState) string {
	if nix.Ready {
		return "ready"
	}
	return "not ready"
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
