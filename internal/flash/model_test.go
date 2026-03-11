package flash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

// setupValidFlashRoot creates a temp dir with MP255 and SWUpdate prerequisites
// so CanFlash passes. Returns (root, platforms).
func setupValidFlashRoot(t *testing.T) (string, []platform.Platform) {
	t.Helper()
	root := t.TempDir()

	// MP255: release dir + deploy.sh
	mp255Release := filepath.Join(root, "pkg", "mp255-ulrpm")
	if err := os.MkdirAll(mp255Release, 0755); err != nil {
		t.Fatal(err)
	}
	deployDir := filepath.Join(root, "yocto")
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		t.Fatal(err)
	}
	deploySh := filepath.Join(deployDir, "deploy.sh")
	if err := os.WriteFile(deploySh, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// SWUpdate: image path (dir with .swu) + augury_update
	cassiaPath := filepath.Join(root, "pkg", "cassia-x2000")
	if err := os.MkdirAll(cassiaPath, 0755); err != nil {
		t.Fatal(err)
	}
	swuPath := filepath.Join(cassiaPath, "image.swu")
	if err := os.WriteFile(swuPath, []byte("fake swu"), 0644); err != nil {
		t.Fatal(err)
	}
	auguryDir := filepath.Join(root, "common", "otsn")
	if err := os.MkdirAll(auguryDir, 0755); err != nil {
		t.Fatal(err)
	}
	auguryUpdate := filepath.Join(auguryDir, "augury_update")
	if err := os.WriteFile(auguryUpdate, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputRelPath: "pkg/mp255-ulrpm"},
		{ID: "cassia-x2000", OutputRelPath: "pkg/cassia-x2000"},
	}
	return root, platforms
}

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
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
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

func TestModel_PlatformSelection(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Select MP255 (supports method selection)
	msg := PlatformSelectedMsg{PlatformID: "mp255-ulrpm"}
	m2, _ := m.Update(msg)
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model.state != stateMethodSelect {
		t.Errorf("After selecting MP255, state = %v, want %v", model.state, stateMethodSelect)
	}
	if model.adapter == nil {
		t.Error("adapter should be set for MP255")
	}
	if model.adapter.PlatformType() != PlatformTypeMP255 {
		t.Errorf("adapter.PlatformType() = %v, want %v", model.adapter.PlatformType(), PlatformTypeMP255)
	}

	// Select Cassia (goes straight to flashing)
	m = NewModel(st, platforms)
	m.state = statePlatformSelect
	m.cursor = 1 // cassia

	msg2 := PlatformSelectedMsg{PlatformID: "cassia-x2000"}
	m3, _ := m.Update(msg2)
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model2.state != stateFlashing {
		t.Errorf("After selecting Cassia, state = %v, want %v", model2.state, stateFlashing)
	}
	if model2.adapter == nil {
		t.Error("adapter should be set for Cassia")
	}
	if model2.adapter.PlatformType() != PlatformTypeSWUpdate {
		t.Errorf("adapter.PlatformType() = %v, want %v", model2.adapter.PlatformType(), PlatformTypeSWUpdate)
	}
}

func TestModel_PlatformSelectionErrors(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputRelPath: "pkg/mp255-ulrpm"},
		{ID: "unknown-device", OutputRelPath: "pkg/unknown"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}

	t.Run("platform not found", func(t *testing.T) {
		m := NewModel(st, platforms)
		m.state = statePlatformSelect

		msg := PlatformSelectedMsg{PlatformID: "nonexistent"}
		m2, _ := m.Update(msg)
		model, ok := m2.(*Model)
		if !ok {
			t.Fatal("Update did not return *Model")
		}

		if model.state != stateError {
			t.Errorf("After unknown platform, state = %v, want stateError", model.state)
		}
		if model.err == nil {
			t.Error("err should be set for unknown platform")
		}
	})

	t.Run("unsupported platform type", func(t *testing.T) {
		m := NewModel(st, platforms)
		m.state = statePlatformSelect

		msg := PlatformSelectedMsg{PlatformID: "unknown-device"}
		m2, _ := m.Update(msg)
		model, ok := m2.(*Model)
		if !ok {
			t.Fatal("Update did not return *Model")
		}

		if model.state != stateError {
			t.Errorf("After unsupported type, state = %v, want stateError", model.state)
		}
		if model.err == nil {
			t.Error("err should be set for unsupported type")
		}
	})
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

func TestModel_ViewMethodSelect(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Select MP255 to reach method select state
	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "mp255-ulrpm"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	model.Width = 80
	model.Height = 24

	view := model.View()

	if !strings.Contains(view, "Choose Flash Method") {
		t.Error("View should contain 'Choose Flash Method'")
	}
	if !strings.Contains(view, "mp255-ulrpm") {
		t.Error("View should contain platform name mp255-ulrpm")
	}
	if !strings.Contains(view, "Official UUU (automated)") {
		t.Error("View should contain UUU method name")
	}
	if !strings.Contains(view, "Manual/Rescue (step-by-step)") {
		t.Error("View should contain Manual method name")
	}
}

func TestModel_MethodSelectKeyboardNavigation(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Select MP255 to reach method select state
	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "mp255-ulrpm"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	// Press 'j' to move down
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model2.cursor != 1 {
		t.Errorf("After 'j', cursor = %d, want 1", model2.cursor)
	}

	// Press 'k' to move up
	m4, _ := model2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model3, ok := m4.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model3.cursor != 0 {
		t.Errorf("After 'k', cursor = %d, want 0", model3.cursor)
	}
}

func TestModel_MethodSelectEnter(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "mp255-ulrpm"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	model.cursor = 1 // Select Manual/Rescue

	// Press Enter
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model2.state != stateFlashing {
		t.Errorf("After Enter, state = %v, want %v", model2.state, stateFlashing)
	}
	if model2.selectedMethod != "manual" {
		t.Errorf("selectedMethod = %v, want manual", model2.selectedMethod)
	}
}

func TestModel_MethodSelectNumberKey(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "mp255-ulrpm"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	// Press '1' to select UUU
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model2.state != stateFlashing {
		t.Errorf("After '1', state = %v, want %v", model2.state, stateFlashing)
	}
	if model2.selectedMethod != "uuu" {
		t.Errorf("selectedMethod = %v, want uuu", model2.selectedMethod)
	}
}

func TestModel_MethodSelectEsc(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "mp255-ulrpm"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	// Press Esc to go back
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model2.state != statePlatformSelect {
		t.Errorf("After Esc, state = %v, want %v", model2.state, statePlatformSelect)
	}
	if model2.cursor != 0 {
		t.Errorf("After Esc, cursor = %d, want 0", model2.cursor)
	}
}

func TestModel_ViewFlashing_MP255(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Select MP255 → method select → select UUU → flashing
	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "mp255-ulrpm"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	model2.Width = 80
	model2.Height = 24

	view := model2.View()

	if !strings.Contains(view, "Flashing Firmware") {
		t.Error("View should contain 'Flashing Firmware'")
	}
	if !strings.Contains(view, "mp255-ulrpm") {
		t.Error("View should contain platform name mp255-ulrpm")
	}
	if !strings.Contains(view, "Method: uuu") {
		t.Error("View should contain method uuu")
	}
	if !strings.Contains(view, "Deploy.sh integration coming soon") {
		t.Error("View should contain MP255 step description")
	}
	if !strings.Contains(view, "Flashing will be implemented in next tasks") {
		t.Error("View should contain placeholder message")
	}
	if !strings.Contains(view, "Ready to flash") {
		t.Error("View should contain context 'Ready to flash'")
	}
}

func TestModel_ViewFlashing_SWUpdate(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Select Cassia (SWUpdate) → goes straight to flashing
	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "cassia-x2000"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	model.Width = 80
	model.Height = 24

	view := model.View()

	if !strings.Contains(view, "Flashing Firmware") {
		t.Error("View should contain 'Flashing Firmware'")
	}
	if !strings.Contains(view, "cassia-x2000") {
		t.Error("View should contain platform name cassia-x2000")
	}
	if !strings.Contains(view, "Verify firmware image") {
		t.Error("View should contain SWUpdate step descriptions")
	}
	if !strings.Contains(view, "Flash firmware to device") {
		t.Error("View should contain SWUpdate step descriptions")
	}
	if !strings.Contains(view, "Reboot device to apply firmware") {
		t.Error("View should contain SWUpdate step descriptions")
	}
}

func TestModel_FlashingEsc_MP255(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "mp255-ulrpm"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	// Press Esc from flashing → back to method select
	m4, _ := model2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model3, ok := m4.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model3.state != stateMethodSelect {
		t.Errorf("After Esc from flashing (MP255), state = %v, want %v", model3.state, stateMethodSelect)
	}
	if model3.selectedMethod != "" {
		t.Errorf("selectedMethod should be cleared, got %q", model3.selectedMethod)
	}
}

func TestModel_FlashingEsc_SWUpdate(t *testing.T) {
	root, platforms := setupValidFlashRoot(t)
	st := status.RepoStatus{Root: root}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	m2, _ := m.Update(PlatformSelectedMsg{PlatformID: "cassia-x2000"})
	model, ok := m2.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	// Press Esc from flashing → back to platform select
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model2, ok := m3.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}

	if model2.state != statePlatformSelect {
		t.Errorf("After Esc from flashing (SWUpdate), state = %v, want %v", model2.state, statePlatformSelect)
	}
	if model2.adapter != nil {
		t.Error("adapter should be cleared when going back from SWUpdate flashing")
	}
}
