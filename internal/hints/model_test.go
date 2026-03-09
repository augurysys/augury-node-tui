package hints

import (
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
)

func TestHintsScreen_UsesComponents(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	platforms := platform.Registry()
	m := NewModel(st, platforms)
	view := m.View()

	// Should use Card component (has borders)
	if !strings.Contains(view, "─") {
		t.Error("Hints screen should use Card component")
	}

	// Should use KeyHint components (bracket format)
	if !strings.Contains(view, "[") || !strings.Contains(view, "]") {
		t.Error("Hints screen should use KeyHint components")
	}
}
