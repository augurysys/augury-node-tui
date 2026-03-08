package hints

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Status    status.RepoStatus
	Platforms []platform.Platform
}

func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	return &Model{Status: st, Platforms: platforms}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	var b strings.Builder
	b.WriteString("Config / Flow Hints\n")
	b.WriteString("Static cards per platform:\n\n")
	for _, p := range m.Platforms {
		b.WriteString("--- " + p.ID + " ---\n")
		b.WriteString(fmt.Sprintf("  script: %s\n", p.ScriptRelPath))
		b.WriteString(fmt.Sprintf("  output: %s\n", p.OutputRelPath))
		b.WriteString("  cache: branch-scoped, artifact hydration on demand\n")
		b.WriteString("\n")
	}
	return b.String()
}
