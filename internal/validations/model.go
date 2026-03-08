package validations

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/run"
	"github.com/augurysys/augury-node-tui/internal/status"
)

var presets = []string{"all", "shellcheck-only", "bats-only", "parse-test-only"}

var presetCommands = map[string]struct {
	script string
	args   []string
}{
	"all":             {"scripts/validate-all.sh", nil},
	"shellcheck-only":  {"scripts/validate-shellcheck.sh", nil},
	"bats-only":       {"scripts/validate-bats.sh", nil},
	"parse-test-only": {"scripts/validate-parse-test.sh", nil},
}

type Model struct {
	Status status.RepoStatus
}

func NewModel(st status.RepoStatus) *Model {
	return &Model{Status: st}
}

func (m *Model) CommandForPreset(preset string) (run.RunSpec, bool) {
	cmd, ok := presetCommands[preset]
	if !ok {
		return run.RunSpec{}, false
	}
	scriptPath := filepath.Join(m.Status.Root, cmd.script)
	if _, err := os.Stat(scriptPath); err != nil {
		return run.RunSpec{}, false
	}
	args := []string{scriptPath}
	if len(cmd.args) > 0 {
		args = append(args, cmd.args...)
	}
	return run.RunSpec{
		Name:    preset,
		Root:    m.Status.Root,
		Mode:    run.ModeSmart,
		Command: "sh",
		Args:    args,
	}, true
}

func (m *Model) View() string {
	var b strings.Builder
	b.WriteString("Validations\n")
	b.WriteString("Presets: all, shellcheck-only, bats-only, parse-test-only\n")
	for _, p := range presets {
		_, ok := m.CommandForPreset(p)
		avail := "not available"
		if ok {
			avail = "available"
		}
		b.WriteString("  " + p + ": " + avail + "\n")
	}
	return b.String()
}
