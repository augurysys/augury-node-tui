package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepRoot_AutoDetectDisplaysPath(t *testing.T) {
	step := NewRootStep("/detected/augury-node")
	view := step.View()

	if !strings.Contains(view, "/detected/augury-node") {
		t.Error("View should display detected path")
	}
}

func TestStepRoot_UserInputUpdatesPath(t *testing.T) {
	step := NewRootStep("")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/custom/path")})

	if step.GetRootPath() != "/custom/path" {
		t.Errorf("Path should be '/custom/path', got %q", step.GetRootPath())
	}
}

func TestStepRoot_EnterConfirms(t *testing.T) {
	step := NewRootStep("/some/path")

	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return a command")
	}
	if step.GetRootPath() != "/some/path" {
		t.Error("Path should be preserved")
	}

	// Run the command and verify the message
	msg := cmd()
	confirmMsg, ok := msg.(RootConfirmedMsg)
	if !ok {
		t.Fatalf("Command should return RootConfirmedMsg, got %T", msg)
	}
	if confirmMsg.Path != "/some/path" {
		t.Errorf("Message path should be '/some/path', got %q", confirmMsg.Path)
	}
}

func TestStepRoot_BackspaceHandlesUTF8(t *testing.T) {
	step := NewRootStep("")

	// Type "café"
	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("café")})
	if step.GetRootPath() != "café" {
		t.Errorf("Should have 'café', got %q", step.GetRootPath())
	}

	// Backspace should remove 'é', not corrupt the string
	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if step.GetRootPath() != "caf" {
		t.Errorf("Backspace should remove 'é', got %q", step.GetRootPath())
	}
}

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	dirs := []string{"apple", "apricot", "banana", "cherry"}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	files := []string{"app.txt", "ark.log"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return tmpDir
}

func TestTabComplete_SingleMatch(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "ban")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	expected := filepath.Join(tmpDir, "banana") + "/"
	if step.GetRootPath() != expected {
		t.Errorf("Tab should complete unique match. Got %q, want %q", step.GetRootPath(), expected)
	}
	if step.menuActive {
		t.Error("Menu should not be active for unique match")
	}
	if len(step.matches) != 0 {
		t.Error("Matches should be cleared after unique completion")
	}
}

func TestTabComplete_MultipleMatches_LCP(t *testing.T) {
	tmpDir := t.TempDir()

	os.Mkdir(filepath.Join(tmpDir, "config"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "config-backup"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "config-old"), 0755)

	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "con")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	expected := filepath.Join(tmpDir, "config")
	if step.GetRootPath() != expected {
		t.Errorf("First Tab should fill LCP to 'config'. Got %q, want %q", step.GetRootPath(), expected)
	}
	if step.menuActive {
		t.Error("Menu should not be active after first Tab when LCP makes progress")
	}
	if len(step.matches) != 3 {
		t.Errorf("Should have 3 matches (config/, config-backup/, config-old/), got %d", len(step.matches))
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !step.menuActive {
		t.Error("Second Tab should enter menu mode after LCP was filled")
	}
}

func TestTabComplete_MenuCycle(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "a")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !step.menuActive {
		t.Fatal("Menu should enter immediately when no LCP progress")
	}
	if step.menuIndex != 0 {
		t.Errorf("Menu should start at index 0, got %d", step.menuIndex)
	}

	inputBeforeMenu := step.inputBeforeMenu
	if inputBeforeMenu == "" {
		t.Error("inputBeforeMenu should be saved")
	}

	numMatches := len(step.matches)
	match0 := step.GetRootPath()

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})
	if step.menuIndex != 1 {
		t.Errorf("Tab should cycle forward to 1, got index %d", step.menuIndex)
	}
	match1 := step.GetRootPath()
	if match0 == match1 {
		t.Error("Tab should change the displayed path")
	}

	for i := 2; i < numMatches; i++ {
		step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})
		if step.menuIndex != i {
			t.Errorf("Tab should cycle to index %d, got %d", i, step.menuIndex)
		}
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})
	if step.menuIndex != 0 {
		t.Errorf("Tab should wrap around to 0 after %d matches, got %d", numMatches, step.menuIndex)
	}
	if step.GetRootPath() != match0 {
		t.Error("After wrapping, should return to first match")
	}
}

func TestTabComplete_ShiftTabReverse(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "a")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !step.menuActive || step.menuIndex != 0 {
		t.Fatalf("Setup failed: menu should be active at index 0, got menuActive=%v, menuIndex=%d", step.menuActive, step.menuIndex)
	}

	numMatches := len(step.matches)
	if numMatches < 2 {
		t.Fatalf("Need at least 2 matches for this test, got %d", numMatches)
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	expectedLast := numMatches - 1
	if step.menuIndex != expectedLast {
		t.Errorf("Shift+Tab from 0 should wrap to last (%d), got %d", expectedLast, step.menuIndex)
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	expectedPrev := numMatches - 2
	if step.menuIndex != expectedPrev {
		t.Errorf("Shift+Tab should cycle backward to %d, got %d", expectedPrev, step.menuIndex)
	}
}

func TestTabComplete_ArrowKeys(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "a")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !step.menuActive {
		t.Fatal("Setup failed: menu should be active")
	}

	numMatches := len(step.matches)
	if numMatches < 2 {
		t.Fatalf("Need at least 2 matches, got %d", numMatches)
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyDown})
	if step.menuIndex != 1 {
		t.Errorf("Down should cycle forward to 1, got %d", step.menuIndex)
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyUp})
	if step.menuIndex != 0 {
		t.Errorf("Up should cycle backward to 0, got %d", step.menuIndex)
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyUp})
	expectedLast := numMatches - 1
	if step.menuIndex != expectedLast {
		t.Errorf("Up from 0 should wrap to %d, got %d", expectedLast, step.menuIndex)
	}
}

func TestTabComplete_EscReverts(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	originalInput := filepath.Join(tmpDir, "a")
	step.userInput = originalInput

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !step.menuActive {
		t.Fatal("Setup failed: menu should be active")
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})
	if step.GetRootPath() == originalInput {
		t.Error("After cycling, input should have changed")
	}

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if step.menuActive {
		t.Error("Menu should be closed after Esc")
	}
	if step.GetRootPath() != originalInput {
		t.Errorf("Esc should revert to original input. Got %q, want %q", step.GetRootPath(), originalInput)
	}
	if len(step.matches) != 0 {
		t.Error("Matches should be cleared after Esc")
	}
}

func TestTabComplete_EnterAccepts(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "a")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})
	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !step.menuActive || step.menuIndex != 1 {
		t.Fatalf("Setup failed: menu should be active at index 1, got menuActive=%v, menuIndex=%d", step.menuActive, step.menuIndex)
	}

	acceptedPath := step.GetRootPath()

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if step.menuActive {
		t.Error("Menu should be closed after Enter")
	}
	if step.GetRootPath() != acceptedPath {
		t.Error("Enter should keep the highlighted path")
	}
	if len(step.matches) != 0 {
		t.Error("Matches should be cleared after Enter")
	}
}

func TestTabComplete_CharExitsMenu(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "a")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !step.menuActive {
		t.Fatal("Setup failed: menu should be active")
	}

	currentPath := step.GetRootPath()

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if step.menuActive {
		t.Error("Typing a character should exit menu")
	}
	expectedPath := currentPath + "x"
	if step.GetRootPath() != expectedPath {
		t.Errorf("Character should be appended. Got %q, want %q", step.GetRootPath(), expectedPath)
	}
	if len(step.matches) != 0 {
		t.Error("Matches should be cleared after typing")
	}
}

func TestTabComplete_BackspaceExitsMenu(t *testing.T) {
	tmpDir := setupTestDir(t)
	step := NewRootStep("")
	step.userInput = filepath.Join(tmpDir, "a")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !step.menuActive {
		t.Fatal("Setup failed: menu should be active")
	}

	currentPath := step.GetRootPath()

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if step.menuActive {
		t.Error("Backspace should exit menu")
	}

	expectedPath := currentPath[:len(currentPath)-1]
	if step.GetRootPath() != expectedPath {
		t.Errorf("Backspace should remove last char. Got %q, want %q", step.GetRootPath(), expectedPath)
	}
	if len(step.matches) != 0 {
		t.Error("Matches should be cleared after backspace")
	}
}
