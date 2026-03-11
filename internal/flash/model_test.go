package flash

import (
	"strings"
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

func TestModel_ViewPlatformSelect(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputRelPath: "pkg/mp255-ulrpm"},
		{ID: "cassia-x2000", OutputRelPath: "pkg/cassia-x2000"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect
	m.Width = 80
	m.Height = 24

	view := m.View()

	// Should contain platform names
	if !strings.Contains(view, "mp255-ulrpm") {
		t.Error("View should contain mp255-ulrpm")
	}
	if !strings.Contains(view, "cassia-x2000") {
		t.Error("View should contain cassia-x2000")
	}
}

func TestModel_KeyboardNavigation(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "platform1"},
		{ID: "platform2"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Press 'j' to move down
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model.cursor != 1 {
		t.Errorf("After 'j', cursor = %d, want 1", model.cursor)
	}

	// Press 'k' to move up
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model2.cursor != 0 {
		t.Errorf("After 'k', cursor = %d, want 0", model2.cursor)
	}
}

func TestModel_PlatformSelectEnter(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputRelPath: "pkg/mp255-ulrpm"},
		{ID: "cassia-x2000", OutputRelPath: "pkg/cassia-x2000"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect
	m.cursor = 1 // Select cassia

	// Press Enter
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")})
	_, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	// Should return a command that produces PlatformSelectedMsg
	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	msg := cmd()
	psMsg, ok := msg.(PlatformSelectedMsg)
	if !ok {
		t.Fatalf("Command returned %T, want PlatformSelectedMsg", msg)
	}

	if psMsg.PlatformID != "cassia-x2000" {
		t.Errorf("PlatformSelectedMsg.PlatformID = %v, want cassia-x2000", psMsg.PlatformID)
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
