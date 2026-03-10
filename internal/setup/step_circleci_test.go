package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCircleCIStep_EmptySkips(t *testing.T) {
	s := NewCircleCIStep()

	s, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !s.Confirmed() {
		t.Error("should be confirmed")
	}
	if !s.Skipped() {
		t.Error("empty input should be skipped")
	}
	if s.Token() != "" {
		t.Errorf("token should be empty, got %q", s.Token())
	}
	if cmd == nil {
		t.Error("should return NextStepMsg cmd")
	}
}

func TestCircleCIStep_WithToken(t *testing.T) {
	s := NewCircleCIStep()

	for _, r := range "my-token-123" {
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	s, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !s.Confirmed() {
		t.Error("should be confirmed")
	}
	if s.Skipped() {
		t.Error("non-empty input should not be skipped")
	}
	if s.Token() != "my-token-123" {
		t.Errorf("token = %q, want %q", s.Token(), "my-token-123")
	}
	if cmd == nil {
		t.Error("should return NextStepMsg cmd")
	}
}

func TestCircleCIStep_ViewMasksInput(t *testing.T) {
	s := NewCircleCIStep()
	for _, r := range "secret" {
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	view := s.View()
	if !strings.Contains(view, "******") {
		t.Error("view should mask input with asterisks")
	}
	if strings.Contains(view, "secret") {
		t.Error("view should not show raw token")
	}
}
