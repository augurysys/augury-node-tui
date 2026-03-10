package components

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

type KeyBinding struct {
	Key   string
	Label string
}

type ScreenLayout struct {
	Breadcrumb []string
	Context    string
	Content    string
	ActionKeys []KeyBinding
	NavKeys    []KeyBinding
	Width      int
	Height     int
}

func (s *ScreenLayout) Render() string {
	var sections []string

	// Top bar
	sections = append(sections, s.renderTopBar())

	// Separator
	sections = append(sections, s.renderSeparator())

	// Content
	sections = append(sections, s.Content)

	// Separator
	sections = append(sections, s.renderSeparator())

	// Bottom help
	sections = append(sections, s.renderBottomHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (s *ScreenLayout) renderTopBar() string {
	breadcrumb := strings.Join(s.Breadcrumb, " → ")
	breadcrumbStyled := styles.TopBar.Render(breadcrumb)

	if s.Context != "" {
		contextStyled := styles.TopBarContext.Render("  •  " + s.Context)
		return breadcrumbStyled + contextStyled
	}

	return breadcrumbStyled
}

func (s *ScreenLayout) renderSeparator() string {
	width := s.Width
	if width <= 0 {
		width = 80
	}
	return styles.Separator.Render(strings.Repeat("─", width))
}

func (s *ScreenLayout) renderBottomHelp() string {
	var parts []string

	// Left side: context-aware action keys
	for i, kb := range s.ActionKeys {
		if i > 0 {
			parts = append(parts, "  ")
		}
		parts = append(parts, styles.KeyBinding(kb.Key, kb.Label))
	}

	// Separator between actions and nav
	if len(s.ActionKeys) > 0 && len(s.NavKeys) > 0 {
		parts = append(parts, "  •  ")
	}

	// Right side: universal nav keys
	for i, kb := range s.NavKeys {
		if i > 0 {
			parts = append(parts, "  ")
		}
		parts = append(parts, styles.KeyBinding(kb.Key, kb.Label))
	}

	return styles.KeyHelp.Render(strings.Join(parts, ""))
}
