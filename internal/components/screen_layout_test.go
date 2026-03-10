package components

import (
	"strings"
	"testing"
)

func TestScreenLayout_BasicRender(t *testing.T) {
	layout := ScreenLayout{
		Breadcrumb: []string{"Home"},
		Context:    "master  •  3 selected",
		Content:    "Test content",
		ActionKeys: []KeyBinding{
			{Key: "space", Label: "select"},
		},
		NavKeys: []KeyBinding{
			{Key: "q", Label: "quit"},
		},
		Width:  80,
		Height: 24,
	}

	result := layout.Render()

	if !strings.Contains(result, "Home") {
		t.Error("Expected breadcrumb 'Home' in output")
	}
	if !strings.Contains(result, "master") {
		t.Error("Expected context 'master' in output")
	}
	if !strings.Contains(result, "Test content") {
		t.Error("Expected content in output")
	}
	if !strings.Contains(result, "space") {
		t.Error("Expected action key 'space' in output")
	}
}

func TestScreenLayout_MultiBreadcrumb(t *testing.T) {
	layout := ScreenLayout{
		Breadcrumb: []string{"Home", "Build"},
		Context:    "building",
		Content:    "Content",
		Width:      80,
		Height:     24,
	}

	result := layout.Render()

	if !strings.Contains(result, "Home") || !strings.Contains(result, "Build") {
		t.Error("Expected multi-level breadcrumb in output")
	}
	if !strings.Contains(result, "→") {
		t.Error("Expected breadcrumb separator '→'")
	}
}
