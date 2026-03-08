package caches

import (
	"context"
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/visual/diagram"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	TabBuildUnit  = 0
	TabPlatform   = 1
	TabCount      = 2
)

type Model struct {
	Status              status.RepoStatus
	Platforms           []platform.Platform
	nixState            engine.NixState
	activeTab           int
	selectedIndex       int
	disabledReason      string
	confirmShown        bool
	pendingConfirmReq   engine.ActionRequest
	rowStatus           map[string]string
	Width               int
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	return &Model{
		Status:        st,
		Platforms:     platforms,
		nixState:      engine.ProbeNix(st.Root),
		activeTab:     TabBuildUnit,
		selectedIndex: 0,
		rowStatus:     make(map[string]string),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
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
	if len(m.Platforms) == 0 {
		return ""
	}
	i := m.selectedIndex
	if i < 0 {
		i = 0
	}
	if i >= len(m.Platforms) {
		i = len(m.Platforms) - 1
	}
	return m.Platforms[i].ID
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
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
			return m, nil
		}
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

func (m *Model) View() string {
	var b strings.Builder
	if m.Width >= diagram.MinDiagramWidth {
		b.WriteString(diagram.CacheTopology(m.activeTab))
		b.WriteString("\n")
	}
	b.WriteString("Caches\n")
	b.WriteString(fmt.Sprintf("Tab: %s\n", m.ActiveTabName()))
	if m.disabledReason != "" {
		b.WriteString(fmt.Sprintf("Disabled: %s\n", m.disabledReason))
	}
	if m.confirmShown {
		b.WriteString("Confirm destructive action? (y/n)\n")
		return b.String()
	}
	b.WriteString("Build-unit: B=build R=pull D=delete | Platform: P=pull U=push X=clean\n")
	if s := m.rowStatus["global"]; s != "" {
		b.WriteString(fmt.Sprintf("Global: [%s]\n", s))
	}
	for _, p := range m.Platforms {
		line := fmt.Sprintf("  %s -> %s", p.ID, p.OutputRelPath)
		if s := m.rowStatus[p.ID]; s != "" {
			line += " [" + s + "]"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}
