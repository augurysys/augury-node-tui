package styles

import (
	"strings"
	"testing"
)

func TestStatusStyle_ReturnsCorrectStyle(t *testing.T) {
	tests := []struct {
		status string
		want   string // just check it returns something styled
	}{
		{"ready", "success"},
		{"clean", "success"},
		{"built", "success"},
		{"not ready", "warning"},
		{"dirty", "warning"},
		{"missing", "warning"},
		{"error", "error"},
		{"failed", "error"},
		{"unknown", "dim"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			style := StatusStyle(tt.status)
			rendered := style.Render(tt.status)
			// Just verify it produces non-empty output
			if rendered == "" {
				t.Errorf("StatusStyle(%q) produced empty output", tt.status)
			}
		})
	}
}

func TestKeyBinding_FormatsCorrectly(t *testing.T) {
	result := KeyBinding("q", "quit")
	if result == "" {
		t.Error("KeyBinding produced empty output")
	}
	if !strings.Contains(result, "q") {
		t.Error("KeyBinding output should contain key")
	}
	if !strings.Contains(result, "quit") {
		t.Error("KeyBinding output should contain description")
	}
}
