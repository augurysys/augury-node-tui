package ci

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelNoToken(t *testing.T) {
	m := NewModel("", "gh/org/repo", "main", "/tmp/root")
	if m.state != stateNoToken {
		t.Errorf("state = %v, want stateNoToken", m.state)
	}
}

func TestNewModelWithToken(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	if m.state != stateLoading {
		t.Errorf("state = %v, want stateLoading", m.state)
	}
	if m.client == nil {
		t.Error("client should not be nil")
	}
}

func TestUpdatePipelineLoaded(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	p := &Pipeline{ID: "p1", Number: 10, State: "success"}

	result, cmd := m.Update(PipelineLoadedMsg{Pipeline: p, Slug: "gh/org/repo"})
	model := result.(*Model)
	if model.pipeline == nil || model.pipeline.ID != "p1" {
		t.Error("pipeline should be set")
	}
	if cmd == nil {
		t.Error("should return cmd to fetch jobs")
	}
}

func TestUpdateJobsLoaded(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	m.Width = 120
	m.Height = 40

	jobs := []Job{
		{ID: "j1", Name: "lint", Status: "success", JobNumber: 1},
		{ID: "j2", Name: "build", Status: "failed", JobNumber: 2},
	}

	result, _ := m.Update(JobsLoadedMsg{Jobs: jobs})
	model := result.(*Model)
	if model.state != stateReady {
		t.Errorf("state = %v, want stateReady", model.state)
	}
	if len(model.jobs) != 2 {
		t.Errorf("jobs count = %d, want 2", len(model.jobs))
	}
	if model.jobsTable == nil {
		t.Error("jobsTable should be initialized")
	}
}

func TestUpdateCIError(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")

	result, _ := m.Update(CIErrorMsg{Err: fmt.Errorf("test error")})
	model := result.(*Model)
	if model.state != stateError {
		t.Errorf("state = %v, want stateError", model.state)
	}
	if model.errMsg != "test error" {
		t.Errorf("errMsg = %q, want %q", model.errMsg, "test error")
	}
}

func TestKeyRefreshInReady(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	m.state = stateReady

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := result.(*Model)
	if model.state != stateLoading {
		t.Errorf("state = %v, want stateLoading", model.state)
	}
	if cmd == nil {
		t.Error("should return fetchPipeline cmd")
	}
}

func TestKeyRetryInError(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	m.state = stateError

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := result.(*Model)
	if model.state != stateLoading {
		t.Errorf("state = %v, want stateLoading", model.state)
	}
	if cmd == nil {
		t.Error("should return fetchPipeline cmd")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "-"},
		{90 * time.Second, "1:30"},
		{5*time.Minute + 22*time.Second, "5:22"},
		{65 * time.Minute, "65:00"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
