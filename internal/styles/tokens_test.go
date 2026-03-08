package styles

import (
	"testing"
)

func TestPalette_ColorsAreDefined(t *testing.T) {
	if Palette.Base == "" {
		t.Error("Palette.Base should be defined")
	}
	if Palette.Success == "" {
		t.Error("Palette.Success should be defined")
	}
	if Palette.Error == "" {
		t.Error("Palette.Error should be defined")
	}
}

func TestPalette_StatusColors(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		expected string
	}{
		{"Success is green", Palette.Success, "#A6E3A1"},
		{"Error is red", Palette.Error, "#F38BA8"},
		{"Warning is yellow", Palette.Warning, "#F9E2AF"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color != tt.expected {
				t.Errorf("got %s, want %s", tt.color, tt.expected)
			}
		})
	}
}
