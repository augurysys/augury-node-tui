package caches

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	TabBuildUnit  = 0
	TabPlatform   = 1
	TabCount      = 2
)

type Model struct {
	Status    status.RepoStatus
	Platforms []platform.Platform
	activeTab int
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	return &Model{Status: st, Platforms: platforms, activeTab: TabBuildUnit}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "tab" || msg.String() == "t" {
			m.NextTab()
		}
	}
	return m, nil
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

func (m *Model) View() string {
	var b strings.Builder
	b.WriteString("Caches (read-only)\n")
	b.WriteString(fmt.Sprintf("Tab: %s\n", m.ActiveTabName()))
	b.WriteString("Build-unit cache: submodules fingerprint + local/remote presence\n")
	b.WriteString("(no destructive actions in MVP)\n")
	b.WriteString("Platform cache: Buildroot, Yocto, Go, Cargo\n")
	for _, p := range m.Platforms {
		b.WriteString(fmt.Sprintf("  %s -> %s\n", p.ID, p.OutputRelPath))
	}
	return b.String()
}
