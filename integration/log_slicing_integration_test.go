package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/build"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

const deterministicLogWithError = `L0
L1
L2
L3
L4
error: something failed
L6
L7
L8
L9
L10
`

func TestLogSlicing_FirstErrorExtractionFromFixtureLog(t *testing.T) {
	root := setupFixtureRoot(t)
	logDir := filepath.Join(root, "tmp", "augury-node-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}
	pid := platforms[0].ID
	if err := os.WriteFile(filepath.Join(logDir, pid+".log"), []byte(deterministicLogWithError), 0644); err != nil {
		t.Fatal(err)
	}

	selected := map[string]bool{pid: true}
	st := status.RepoStatus{Root: root, Branch: "main", SHA: "x"}
	m := build.NewModel(st, platforms, selected)
	m.Summary = &build.Summary{
		Rows: []build.SummaryRow{{PlatformID: pid, Status: build.RowStatusFailure}},
	}
	m.Focused = 0

	child, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = child.(*build.Model)

	view := m.View()
	if !strings.Contains(view, "error: something failed") {
		t.Errorf("view must contain first error line; got %q", view)
	}
	if !strings.Contains(view, "L4") || !strings.Contains(view, "L6") {
		t.Errorf("view must contain context around first error (L4, L6); got %q", view)
	}
}
