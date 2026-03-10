package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/augurysys/augury-node-tui/internal/styles"
)

type RootStep struct {
	detectedPath    string
	userInput       string
	confirmed       bool
	matches         []string
	menuActive      bool
	menuIndex       int
	inputBeforeMenu string
}

func NewRootStep(detectedPath string) *RootStep {
	return &RootStep{
		detectedPath: detectedPath,
		userInput:    detectedPath,
	}
}

func (s *RootStep) Update(msg tea.Msg) (*RootStep, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if s.menuActive {
				s.exitMenu(true)
				return s, nil
			}
			path := strings.TrimSpace(s.userInput)
			if path == "" {
				return s, nil
			}
			expandedPath := expandHome(path)
			s.confirmed = true
			return s, func() tea.Msg { return RootConfirmedMsg{Path: expandedPath} }
		case tea.KeyEsc:
			if s.menuActive {
				s.exitMenu(false)
			}
			return s, nil
		case tea.KeyTab:
			s.tabComplete()
		case tea.KeyShiftTab:
			if s.menuActive {
				s.cyclePrev()
			}
		case tea.KeyUp:
			if s.menuActive {
				s.cyclePrev()
			}
		case tea.KeyDown:
			if s.menuActive {
				s.cycleNext()
			}
		case tea.KeyRunes:
			if len(msg.Runes) == 1 && msg.Runes[0] == 'q' && s.userInput == "" && !s.menuActive {
				return s, tea.Quit
			}
			if s.menuActive {
				s.exitMenu(true)
			}
			s.matches = nil
			s.userInput += string(msg.Runes)
		case tea.KeyBackspace:
			if s.menuActive {
				s.exitMenu(true)
			}
			s.matches = nil
			runes := []rune(s.userInput)
			if len(runes) > 0 {
				s.userInput = string(runes[:len(runes)-1])
			}
		}
	}
	return s, nil
}

func (s *RootStep) tabComplete() {
	if s.menuActive {
		s.cycleNext()
		return
	}

	if len(s.matches) > 0 {
		s.enterMenu()
		return
	}

	s.computeMatches()
	if len(s.matches) == 0 {
		return
	}

	if len(s.matches) == 1 {
		s.userInput = contractHome(s.matches[0])
		s.matches = nil
		return
	}

	input := s.userInput
	if input == "" {
		input = "."
	}
	expanded := expandHome(input)
	lcp := longestCommonPrefix(s.matches)

	if len(lcp) > len(expanded) {
		s.userInput = contractHome(lcp)
		return
	}

	s.enterMenu()
}

func (s *RootStep) computeMatches() {
	input := s.userInput
	if input == "" {
		input = "."
	}
	expanded := expandHome(input)

	dir := expanded
	prefix := ""
	info, err := os.Stat(expanded)
	if err != nil || !info.IsDir() {
		dir = filepath.Dir(expanded)
		prefix = filepath.Base(expanded)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		s.matches = nil
		return
	}

	var matches []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}
		if prefix == "" || strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			full := filepath.Join(dir, name)
			if e.IsDir() {
				full += "/"
			}
			matches = append(matches, full)
		}
	}

	s.matches = matches
}

func (s *RootStep) enterMenu() {
	if len(s.matches) == 0 {
		return
	}
	s.inputBeforeMenu = s.userInput
	s.menuActive = true
	s.menuIndex = 0
	s.userInput = contractHome(s.matches[0])
}

func (s *RootStep) exitMenu(accept bool) {
	if !accept {
		s.userInput = s.inputBeforeMenu
	}
	s.menuActive = false
	s.menuIndex = -1
	s.matches = nil
	s.inputBeforeMenu = ""
}

func (s *RootStep) cycleNext() {
	if !s.menuActive || len(s.matches) == 0 {
		return
	}
	s.menuIndex = (s.menuIndex + 1) % len(s.matches)
	s.userInput = contractHome(s.matches[s.menuIndex])
}

func (s *RootStep) cyclePrev() {
	if !s.menuActive || len(s.matches) == 0 {
		return
	}
	s.menuIndex--
	if s.menuIndex < 0 {
		s.menuIndex = len(s.matches) - 1
	}
	s.userInput = contractHome(s.matches[s.menuIndex])
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func contractHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	if path == home {
		return "~"
	}
	return path
}

func (s *RootStep) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Step 1: Augury Node Root"))
	b.WriteString("\n\n")

	if s.detectedPath != "" {
		b.WriteString(styles.Success.Render("Auto-detected: "))
		b.WriteString(s.detectedPath)
		b.WriteString("\n\n")
	}

	b.WriteString("Enter path to augury-node repository:\n")
	b.WriteString(styles.Border.Render(s.userInput))
	b.WriteString("\n\n")

	if s.menuActive && len(s.matches) > 0 {
		maxDisplay := 10
		displayCount := len(s.matches)
		hasMore := false
		if displayCount > maxDisplay {
			displayCount = maxDisplay
			hasMore = true
		}

		for i := 0; i < displayCount; i++ {
			name := filepath.Base(s.matches[i])
			if i == s.menuIndex {
				b.WriteString(styles.ItemSelected.Render(name))
			} else {
				b.WriteString(styles.Item.Render(name))
			}
			b.WriteString("\n")
		}

		if hasMore {
			remaining := len(s.matches) - maxDisplay
			b.WriteString(styles.Dim.Render(fmt.Sprintf("  ... and %d more", remaining)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	helpText := "Tab to complete"
	if s.menuActive {
		helpText = "Tab/Shift+Tab to cycle  Enter to select  Esc to cancel"
	}
	b.WriteString(styles.Dim.Render(helpText))

	return b.String()
}

func (s *RootStep) GetRootPath() string {
	return s.userInput
}

func (s *RootStep) Confirmed() bool {
	return s.confirmed
}

type RootConfirmedMsg struct {
	Path string
}
