package caches

import (
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
)

func TestCachesModel_TabsRenderReadOnlyDataSources(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())

	view := m.View()
	if view == "" {
		t.Fatal("View must not be empty")
	}
	if !strings.Contains(strings.ToLower(view), "build") && !strings.Contains(strings.ToLower(view), "unit") {
		t.Errorf("View must render build-unit tab or data; got %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "platform") {
		t.Errorf("View must render platform tab or data; got %q", view)
	}
}

func TestCachesModel_ActiveTabSwitches(t *testing.T) {
	st := status.RepoStatus{Root: "/repo", Branch: "main", SHA: "x"}
	m := NewModel(st, platform.Registry())

	initial := m.ActiveTab()
	m.NextTab()
	if m.ActiveTab() == initial {
		t.Error("NextTab must change active tab")
	}
}
