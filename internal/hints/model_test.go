package hints

import (
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
)

func TestHintsModel_RendersStaticPerPlatformFlowCards(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}

	m := NewModel(st, platforms)
	view := m.View()

	if view == "" {
		t.Fatal("View must not be empty")
	}
	for _, p := range platforms {
		if !strings.Contains(view, p.ID) {
			t.Errorf("View must include platform %q in flow cards; got %q", p.ID, view)
		}
	}
}

func TestHintsModel_FlowCardsIncludeScriptAndOutputPaths(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	platforms := platform.Registry()
	if len(platforms) == 0 {
		t.Fatal("need at least one platform")
	}

	m := NewModel(st, platforms)
	view := m.View()

	p := platforms[0]
	if !strings.Contains(view, p.ScriptRelPath) {
		t.Errorf("flow card for %q must show script path %q", p.ID, p.ScriptRelPath)
	}
	if !strings.Contains(view, p.OutputRelPath) {
		t.Errorf("flow card for %q must show output path %q", p.ID, p.OutputRelPath)
	}
}
