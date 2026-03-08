package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSplashInitialState(t *testing.T) {
	m := NewSplashModel(2 * time.Second)
	if m.Dismissed {
		t.Error("initial state should show splash, not dismissed")
	}
	view := m.View()
	if view == "" {
		t.Error("initial view should render splash content")
	}
}

func TestSplashDismissesOnKeyEvent(t *testing.T) {
	m := NewSplashModel(2 * time.Second)
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	_ = cmd
	m = model.(*SplashModel)
	if !m.Dismissed {
		t.Error("splash should dismiss on key event")
	}
}

func TestSplashAutoDismissesAfterTimeout(t *testing.T) {
	m := NewSplashModel(10 * time.Millisecond)
	model, cmd := m.Update(TimeoutMsg{})
	_ = cmd
	m = model.(*SplashModel)
	if !m.Dismissed {
		t.Error("splash should auto-dismiss after timeout")
	}
}

func TestSplashReplayReturnsToSplashState(t *testing.T) {
	m := NewSplashModel(2 * time.Second)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = model.(*SplashModel)
	if !m.Dismissed {
		t.Fatal("splash should be dismissed first")
	}
	model, _ = m.Update(ReplayMsg{})
	m = model.(*SplashModel)
	if m.Dismissed {
		t.Error("replay should return to splash state")
	}
}

func TestSplashCompactFallbackWhenTerminalTooSmall(t *testing.T) {
	m := NewSplashModel(2 * time.Second)
	m.Width = 20
	m.Height = 5
	view := m.View()
	if view == "" {
		t.Error("compact view should not be empty")
	}
	// Compact fallback should be shorter than full splash
	if len(view) > 200 {
		t.Error("compact view should be shorter for small terminals")
	}
}
