package flash

import (
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModel_StateTransitions(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputRelPath: "pkg/mp255-ulrpm"},
		{ID: "cassia-x2000", OutputRelPath: "pkg/cassia-x2000"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)

	// Initial state
	if m.state != stateIdle {
		t.Errorf("Initial state = %v, want %v", m.state, stateIdle)
	}

	// Can transition to platform select
	m2, _ := m.Update(nil)
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model.state != statePlatformSelect {
		t.Errorf("After first Update, state = %v, want %v", model.state, statePlatformSelect)
	}
}

func TestModel_StateStability(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputRelPath: "pkg/mp255-ulrpm"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)

	// First update: idle → platform select
	m2, _ := m.Update(nil)
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model.state != statePlatformSelect {
		t.Errorf("After first Update, state = %v, want %v", model.state, statePlatformSelect)
	}

	// Second update: should stay in platform select
	m3, _ := model.Update(nil)
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model2.state != statePlatformSelect {
		t.Errorf("After second Update, state = %v, want %v", model2.state, statePlatformSelect)
	}
}

func TestModel_WindowResize(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputRelPath: "pkg/mp255-ulrpm"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)

	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	m2, _ := m.Update(msg)
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model.Width != 100 {
		t.Errorf("Width = %d, want 100", model.Width)
	}
	if model.Height != 40 {
		t.Errorf("Height = %d, want 40", model.Height)
	}
}
