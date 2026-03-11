package caches

import (
	"context"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/visual/diagram"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	TabBuildUnit = 0
	TabPlatform  = 1
	TabCount     = 2
)

// CacheEntry represents a single cache row for the DataTable.
type CacheEntry struct {
	PlatformID    string
	PlatformLabel string
	OutputPath    string
	Type          string
	LocalState    string
	LocalPresent  bool
	RemoteState   string
	RemotePresent bool
	Size          string
}

type Model struct {
	Status            status.RepoStatus
	Platforms         []platform.Platform
	nixState          engine.NixState
	activeTab         int
	disabledReason    string
	confirmShown       bool
	pendingConfirmReq  engine.ActionRequest
	rowStatus         map[string]string
	Width             int
	Height            int
	cacheTable        *components.DataTable
	metricsBar        components.MetricsBar
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	m := &Model{
		Status:        st,
		Platforms:     platforms,
		nixState:      engine.ProbeNix(st.Root),
		activeTab:     TabBuildUnit,
		rowStatus:     make(map[string]string),
		metricsBar:    components.MetricsBar{},
	}
	m.initCacheTable()
	return m
}

func (m *Model) initCacheTable() {
	columns := []components.Column{
		{Header: "Platform", Width: 20, Sortable: true, Renderer: m.renderPlatform},
		{Header: "Type", Width: 15, Sortable: true, Renderer: m.renderCacheType},
		{Header: "Local", Width: 10, Sortable: true, Renderer: m.renderLocalState},
		{Header: "Remote", Width: 10, Sortable: true, Renderer: m.renderRemoteState},
		{Header: "Size", Width: 12, Sortable: true, Align: components.AlignRight, Renderer: m.renderSize},
		{Header: "Actions", Width: -1, Sortable: false, Renderer: m.renderActions},
	}
	m.cacheTable = components.NewDataTable(columns)
	m.cacheTable.SetRows(m.fetchCacheData())
}

func (m *Model) renderPlatform(row interface{}) string {
	return row.(CacheEntry).PlatformLabel
}

func (m *Model) renderCacheType(row interface{}) string {
	return row.(CacheEntry).Type
}

func (m *Model) renderLocalState(row interface{}) string {
	cache := row.(CacheEntry)
	st := primitives.StatusSuccess
	if !cache.LocalPresent {
		st = primitives.StatusUnavailable
	}
	return primitives.StatusBadge{Label: cache.LocalState, Status: st}.Render()
}

func (m *Model) renderRemoteState(row interface{}) string {
	cache := row.(CacheEntry)
	st := primitives.StatusSuccess
	if !cache.RemotePresent {
		st = primitives.StatusUnavailable
	}
	return primitives.StatusBadge{Label: cache.RemoteState, Status: st}.Render()
}

func (m *Model) renderSize(row interface{}) string {
	return row.(CacheEntry).Size
}

func (m *Model) renderActions(row interface{}) string {
	if m.activeTab == TabBuildUnit {
		return "B=build R=pull D=delete"
	}
	return "P=pull U=push X=clean"
}

func (m *Model) fetchCacheData() []interface{} {
	var entries []CacheEntry
	if m.activeTab == TabBuildUnit {
		for _, p := range m.Platforms {
			entries = append(entries, m.cacheEntryForPlatform(p, "build-unit"))
		}
		entries = append(entries, m.cacheEntryForGlobal())
	} else {
		for _, p := range m.Platforms {
			entries = append(entries, m.cacheEntryForPlatform(p, "platform"))
		}
	}
	rows := make([]interface{}, len(entries))
	for i, e := range entries {
		rows[i] = e
	}
	return rows
}

func (m *Model) cacheEntryForGlobal() CacheEntry {
	s := m.rowStatus["global"]
	if s == "" {
		s = "pending"
	}
	localPresent := strings.HasPrefix(strings.ToLower(s), "cached") ||
		strings.HasPrefix(strings.ToLower(s), "success") ||
		strings.HasPrefix(strings.ToLower(s), "complete")
	return CacheEntry{
		PlatformID:    "",
		PlatformLabel: "Global",
		OutputPath:    "",
		Type:          "build-unit",
		LocalState:    primaryStatus(s),
		LocalPresent:  localPresent,
		RemoteState:   "-",
		RemotePresent: false,
		Size:          "-",
	}
}

func (m *Model) cacheEntryForPlatform(p platform.Platform, cacheType string) CacheEntry {
	s := m.rowStatus[p.ID]
	if s == "" {
		s = "pending"
	}
	localPresent := strings.HasPrefix(strings.ToLower(s), "cached") ||
		strings.HasPrefix(strings.ToLower(s), "success") ||
		strings.HasPrefix(strings.ToLower(s), "complete")
	return CacheEntry{
		PlatformID:    p.ID,
		PlatformLabel: p.ID,
		OutputPath:    p.OutputRelPath,
		Type:          cacheType,
		LocalState:    primaryStatus(s),
		LocalPresent:  localPresent,
		RemoteState:   "-",
		RemotePresent: false,
		Size:          "-",
	}
}

func primaryStatus(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, ": "); idx >= 0 {
		return s[:idx]
	}
	return s
}

func (m *Model) Init() tea.Cmd {
	m.initCacheTable()
	return func() tea.Msg {
		_ = m.metricsBar.FetchMetrics()
		return nil
	}
}

func (m *Model) SetNixState(nix engine.NixState) {
	m.nixState = nix
}

func (m *Model) DisabledReason() string {
	return m.disabledReason
}

func (m *Model) ConfirmShown() bool {
	return m.confirmShown
}

func (m *Model) PendingConfirmAction() engine.ActionRequest {
	return m.pendingConfirmReq
}

func (m *Model) selectedPlatformID() string {
	row := m.cacheTable.SelectedRow()
	if row == nil {
		return ""
	}
	return row.(CacheEntry).PlatformID
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.metricsBar.Width = msg.Width
		m.cacheTable.SetWidth(m.Width)
		if m.Height > 0 {
			reservedHeight := 14
			tableHeight := msg.Height - reservedHeight
			if tableHeight < 5 {
				tableHeight = 5
			}
			m.cacheTable.SetHeight(tableHeight)
		} else {
			m.cacheTable.SetHeight(20)
		}
		return m, nil
	case tea.KeyMsg:
		k := msg.String()
		if m.confirmShown {
			if k == "y" || k == "Y" {
				m.confirmShown = false
				req := m.pendingConfirmReq
				m.pendingConfirmReq = engine.ActionRequest{}
				return m, m.dispatchAction(req)
			}
			if k == "n" || k == "N" || k == "esc" {
				m.confirmShown = false
				m.pendingConfirmReq = engine.ActionRequest{}
				return m, nil
			}
			return m, nil
		}
		m.disabledReason = ""
		if k == "tab" || k == "t" {
			m.NextTab()
			m.cacheTable.SetRows(m.fetchCacheData())
			return m, nil
		}
		m.cacheTable.Update(msg)
		platID := m.selectedPlatformID()
		req, ok := ActionForKey(m.activeTab, k, platID)
		if !ok {
			return m, nil
		}
		cap := engine.ResolveCapability(m.Status.Root, req)
		if !cap.Available {
			m.disabledReason = cap.Reason
			return m, nil
		}
		if blocked, reason := engine.IsActionBlockedByNix(req, m.nixState); blocked {
			m.disabledReason = reason
			return m, nil
		}
		if IsDestructive(req) {
			m.confirmShown = true
			m.pendingConfirmReq = req
			return m, nil
		}
		return m, m.dispatchAction(req)
	case jobResultMsg:
		m.handleJobResult(msg.Job, msg.Request)
		m.cacheTable.SetRows(m.fetchCacheData())
		return m, nil
	}
	return m, nil
}

func (m *Model) dispatchAction(req engine.ActionRequest) tea.Cmd {
	root := m.Status.Root
	return func() tea.Msg {
		job := engine.ExecuteAction(context.Background(), root, req)
		return jobResultMsg{Job: job, Request: req}
	}
}

type jobResultMsg struct {
	Job     engine.Job
	Request engine.ActionRequest
}

func (m *Model) handleJobResult(job engine.Job, req engine.ActionRequest) {
	key := req.PlatformID
	if key == "" {
		key = "global"
	}
	m.rowStatus[key] = string(job.State)
	if job.Reason != "" {
		m.rowStatus[key] += ": " + job.Reason
	}
}

func (m *Model) ActiveTab() int {
	return m.activeTab
}

func (m *Model) ActiveTabName() string {
	switch m.activeTab {
	case TabBuildUnit:
		return "build-unit"
	case TabPlatform:
		return "platform"
	default:
		return "build-unit"
	}
}

func (m *Model) NextTab() {
	m.activeTab = (m.activeTab + 1) % TabCount
}

func (m *Model) RowStatus(platformID string) string {
	return m.rowStatus[platformID]
}

func (m *Model) buildContext() string {
	parts := []string{m.Status.Branch, m.ActiveTabName()}
	if m.disabledReason != "" {
		parts = append(parts, "disabled: "+m.disabledReason)
	}
	if m.confirmShown {
		parts = append(parts, "confirm (y/n)")
	}
	return strings.Join(parts, "  •  ")
}

func (m *Model) buildActionKeys() []components.KeyBinding {
	keys := []components.KeyBinding{
		{Key: "tab", Label: "switch tab"},
	}
	if m.activeTab == TabBuildUnit {
		keys = append(keys,
			components.KeyBinding{Key: "B", Label: "build"},
			components.KeyBinding{Key: "R", Label: "pull"},
			components.KeyBinding{Key: "D", Label: "delete"},
		)
	} else {
		keys = append(keys,
			components.KeyBinding{Key: "P", Label: "pull"},
			components.KeyBinding{Key: "U", Label: "push"},
			components.KeyBinding{Key: "X", Label: "clean"},
		)
	}
	return keys
}

func (m *Model) renderContent() string {
	var b strings.Builder
	b.WriteString(m.metricsBar.Render())
	b.WriteString("\n")
	if m.Width >= diagram.MinDiagramWidth {
		b.WriteString(diagram.CacheTopology(m.activeTab))
		b.WriteString("\n")
	}
	if m.confirmShown {
		b.WriteString("Confirm destructive action? (y/n)\n")
		return b.String()
	}
	b.WriteString(m.cacheTable.View())
	return b.String()
}

func (m *Model) View() string {
	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Caches"},
		Context:    m.buildContext(),
		Content:    m.renderContent(),
		ActionKeys: m.buildActionKeys(),
		NavKeys: []components.KeyBinding{
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}
	return layout.Render()
}
