# Flash Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add firmware flashing to augury-node-tui for MP255 and SWUpdate platforms with interactive dev workflow.

**Architecture:** Adapter pattern for platform-specific flash logic (MP255/SWUpdate), unified state machine in Flash screen, ScreenLayout UI integration. MP255 wraps `deploy.sh`, SWUpdate wraps `augury_update`.

**Tech Stack:** Go, Bubbletea, exec.Command for process management, ScreenLayout component, DataTable for platform selection.

---

## Task 1: Flash Adapter Interface & Types

**Files:**
- Create: `internal/flash/adapter.go`
- Create: `internal/flash/types.go`

**Step 1: Write adapter interface test**

Create `internal/flash/adapter_test.go`:

```go
package flash

import (
	"context"
	"testing"
)

func TestAdapterInterface(t *testing.T) {
	// Compile-time check that concrete types implement interface
	var _ FlashAdapter = (*mockAdapter)(nil)
}

type mockAdapter struct {
	platformType string
}

func (m *mockAdapter) PlatformType() string { return m.platformType }
func (m *mockAdapter) SupportsMethodSelection() bool { return false }
func (m *mockAdapter) GetMethods() []FlashMethod { return nil }
func (m *mockAdapter) GetSteps(method string) []FlashStep { return nil }
func (m *mockAdapter) ExecuteStep(ctx context.Context, step FlashStep) (string, error) {
	return "", nil
}
func (m *mockAdapter) CanFlash(imagePath string) error { return nil }
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/flash/...`
Expected: FAIL - "package flash: no Go files"

**Step 3: Define types**

Create `internal/flash/types.go`:

```go
package flash

// PromptType defines the kind of user interaction required
type PromptType int

const (
	PromptConfirm PromptType = iota // Press Enter to continue
	PromptChoice                     // Select 1/2/3...
	PromptManual                     // Follow instructions, then confirm
)

// FlashMethod represents a flashing method (e.g., UUU, Manual/Rescue)
type FlashMethod struct {
	ID          string
	Name        string
	Description string
}

// FlashStep represents one interactive step in the flash process
type FlashStep struct {
	ID          string
	Description string
	PromptType  PromptType
	Choices     []string // For PromptChoice
}
```

**Step 4: Define adapter interface**

Create `internal/flash/adapter.go`:

```go
package flash

import "context"

// FlashAdapter defines the interface for platform-specific flash implementations
type FlashAdapter interface {
	// PlatformType returns "mp255" or "swupdate"
	PlatformType() string

	// SupportsMethodSelection returns true if platform has multiple flash methods
	SupportsMethodSelection() bool

	// GetMethods returns available flash methods (empty if no selection needed)
	GetMethods() []FlashMethod

	// GetSteps returns the interactive steps for the given method
	GetSteps(method string) []FlashStep

	// ExecuteStep runs one step and returns output/error
	ExecuteStep(ctx context.Context, step FlashStep) (output string, err error)

	// CanFlash validates the image path and platform prerequisites
	CanFlash(imagePath string) error
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/flash/...`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/flash/
git commit -m "feat(flash): add adapter interface and types"
```

---

## Task 2: Platform Detection Logic

**Files:**
- Create: `internal/flash/platform.go`
- Create: `internal/flash/platform_test.go`

**Step 1: Write platform detection test**

Create `internal/flash/platform_test.go`:

```go
package flash

import (
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
)

func TestDetectPlatformType(t *testing.T) {
	tests := []struct {
		name     string
		platform platform.Platform
		want     string
	}{
		{
			name:     "mp255 platform",
			platform: platform.Platform{ID: "mp255-ulrpm"},
			want:     "mp255",
		},
		{
			name:     "cassia platform",
			platform: platform.Platform{ID: "cassia-x2000"},
			want:     "swupdate",
		},
		{
			name:     "moxa platform",
			platform: platform.Platform{ID: "moxa-uc3100"},
			want:     "swupdate",
		},
		{
			name:     "unknown platform",
			platform: platform.Platform{ID: "unknown-device"},
			want:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectPlatformType(tt.platform)
			if got != tt.want {
				t.Errorf("DetectPlatformType() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/flash/... -v`
Expected: FAIL - "undefined: DetectPlatformType"

**Step 3: Implement platform detection**

Create `internal/flash/platform.go`:

```go
package flash

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/platform"
)

// DetectPlatformType returns the flash adapter type for a platform
func DetectPlatformType(p platform.Platform) string {
	// MP255 platforms
	if strings.HasPrefix(p.ID, "mp255") {
		return "mp255"
	}

	// SWUpdate platforms
	if strings.Contains(p.ID, "cassia") || strings.Contains(p.ID, "moxa") {
		return "swupdate"
	}

	return "unknown"
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/flash/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/flash/platform.go internal/flash/platform_test.go
git commit -m "feat(flash): add platform type detection"
```

---

## Task 3: Flash Screen State Machine & Messages

**Files:**
- Create: `internal/flash/messages.go`
- Create: `internal/flash/model.go` (initial)
- Create: `internal/flash/model_test.go`

**Step 1: Define messages**

Create `internal/flash/messages.go`:

```go
package flash

// PlatformSelectedMsg is sent when user selects a platform
type PlatformSelectedMsg struct {
	PlatformID string
}

// MethodSelectedMsg is sent when user selects a flash method
type MethodSelectedMsg struct {
	MethodID string
}

// StepCompleteMsg is sent when a flash step completes
type StepCompleteMsg struct {
	StepID string
	Output string
}

// FlashErrorMsg is sent when flash fails
type FlashErrorMsg struct {
	Err error
}

// FlashCompleteMsg is sent when flash succeeds
type FlashCompleteMsg struct{}

// CancelFlashMsg is sent when user cancels
type CancelFlashMsg struct{}
```

**Step 2: Write state machine test**

Create `internal/flash/model_test.go`:

```go
package flash

import (
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
)

func TestModel_StateTransitions(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputPath: "/path/to/mp255"},
		{ID: "cassia-x2000", OutputPath: "/path/to/cassia"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)

	// Initial state
	if m.state != stateIdle {
		t.Errorf("Initial state = %v, want %v", m.state, stateIdle)
	}

	// Can transition to platform select
	m2, _ := m.Update(nil) // First render
	if m2.(*Model).state != statePlatformSelect {
		t.Errorf("After first Update, state = %v, want %v", m2.(*Model).state, statePlatformSelect)
	}
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/flash/... -v`
Expected: FAIL - "undefined: NewModel"

**Step 4: Create initial model with states**

Create `internal/flash/model.go`:

```go
package flash

import (
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateIdle state = iota
	statePlatformSelect
	stateMethodSelect
	stateFlashing
	stateComplete
	stateError
)

// Model is the flash screen model
type Model struct {
	Status       status.RepoStatus
	Platforms    []platform.Platform
	Width        int
	Height       int
	state        state
	selectedPlatform string
	selectedMethod   string
	adapter      FlashAdapter
	err          error
}

// NewModel creates a new flash model
func NewModel(st status.RepoStatus, platforms []platform.Platform) *Model {
	return &Model{
		Status:    st,
		Platforms: platforms,
		state:     stateIdle,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Transition from idle to platform select on first update
	if m.state == stateIdle {
		m.state = statePlatformSelect
	}

	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	return "Flash screen placeholder"
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/flash/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/flash/messages.go internal/flash/model.go internal/flash/model_test.go
git commit -m "feat(flash): add state machine and messages"
```

---

## Task 4: SWUpdate Adapter (Simpler One First)

**Files:**
- Create: `internal/flash/swupdate_adapter.go`
- Create: `internal/flash/swupdate_adapter_test.go`

**Step 1: Write adapter test**

Create `internal/flash/swupdate_adapter_test.go`:

```go
package flash

import (
	"context"
	"testing"
)

func TestSWUpdateAdapter_PlatformType(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	if adapter.PlatformType() != "swupdate" {
		t.Errorf("PlatformType() = %v, want %v", adapter.PlatformType(), "swupdate")
	}
}

func TestSWUpdateAdapter_SupportsMethodSelection(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	if adapter.SupportsMethodSelection() != false {
		t.Error("SupportsMethodSelection() should be false for SWUpdate")
	}
}

func TestSWUpdateAdapter_GetSteps(t *testing.T) {
	adapter := NewSWUpdateAdapter("/tmp/root", "cassia-x2000", "/path/to/image.swu")
	steps := adapter.GetSteps("")
	
	if len(steps) != 3 {
		t.Errorf("GetSteps() returned %d steps, want 3", len(steps))
	}

	// Verify step IDs
	expectedIDs := []string{"verify", "flash", "reboot"}
	for i, step := range steps {
		if step.ID != expectedIDs[i] {
			t.Errorf("Step %d ID = %v, want %v", i, step.ID, expectedIDs[i])
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/flash/... -run TestSWUpdateAdapter -v`
Expected: FAIL - "undefined: NewSWUpdateAdapter"

**Step 3: Implement SWUpdate adapter**

Create `internal/flash/swupdate_adapter.go`:

```go
package flash

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SWUpdateAdapter wraps augury_update for SWUpdate-based platforms
type SWUpdateAdapter struct {
	root      string
	platform  string
	imagePath string
}

// NewSWUpdateAdapter creates a new SWUpdate adapter
func NewSWUpdateAdapter(root, platform, imagePath string) *SWUpdateAdapter {
	return &SWUpdateAdapter{
		root:      root,
		platform:  platform,
		imagePath: imagePath,
	}
}

// PlatformType returns "swupdate"
func (a *SWUpdateAdapter) PlatformType() string {
	return "swupdate"
}

// SupportsMethodSelection returns false (no method choice for SWUpdate)
func (a *SWUpdateAdapter) SupportsMethodSelection() bool {
	return false
}

// GetMethods returns empty slice (no methods to choose)
func (a *SWUpdateAdapter) GetMethods() []FlashMethod {
	return nil
}

// GetSteps returns the three-step SWUpdate flow
func (a *SWUpdateAdapter) GetSteps(method string) []FlashStep {
	return []FlashStep{
		{
			ID:          "verify",
			Description: "Verify firmware image",
			PromptType:  PromptConfirm,
		},
		{
			ID:          "flash",
			Description: "Flash firmware to device",
			PromptType:  PromptConfirm,
		},
		{
			ID:          "reboot",
			Description: "Reboot device to apply firmware",
			PromptType:  PromptConfirm,
		},
	}
}

// ExecuteStep runs one step of the flash process
func (a *SWUpdateAdapter) ExecuteStep(ctx context.Context, step FlashStep) (string, error) {
	augury_update := filepath.Join(a.root, "common/otsn/augury_update")

	var cmd *exec.Cmd
	switch step.ID {
	case "verify":
		cmd = exec.CommandContext(ctx, augury_update, "verify", a.imagePath)
	case "flash":
		cmd = exec.CommandContext(ctx, augury_update, "apply_update", a.imagePath)
	case "reboot":
		return "Reboot required. Please power cycle the device.", nil
	default:
		return "", fmt.Errorf("unknown step: %s", step.ID)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// CanFlash validates prerequisites
func (a *SWUpdateAdapter) CanFlash(imagePath string) error {
	// Check image file exists
	if _, err := os.Stat(imagePath); err != nil {
		return fmt.Errorf("image file not found: %w", err)
	}

	// Check augury_update exists
	augury_update := filepath.Join(a.root, "common/otsn/augury_update")
	if _, err := os.Stat(augury_update); err != nil {
		return fmt.Errorf("augury_update not found: %w", err)
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/flash/... -run TestSWUpdateAdapter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/flash/swupdate_adapter.go internal/flash/swupdate_adapter_test.go
git commit -m "feat(flash): add SWUpdate adapter"
```

---

## Task 5: MP255 Adapter Stub

**Files:**
- Create: `internal/flash/mp255_adapter.go`
- Create: `internal/flash/mp255_adapter_test.go`

**Step 1: Write basic adapter tests**

Create `internal/flash/mp255_adapter_test.go`:

```go
package flash

import (
	"testing"
)

func TestMP255Adapter_PlatformType(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	if adapter.PlatformType() != "mp255" {
		t.Errorf("PlatformType() = %v, want %v", adapter.PlatformType(), "mp255")
	}
}

func TestMP255Adapter_SupportsMethodSelection(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	if adapter.SupportsMethodSelection() != true {
		t.Error("SupportsMethodSelection() should be true for MP255")
	}
}

func TestMP255Adapter_GetMethods(t *testing.T) {
	adapter := NewMP255Adapter("/tmp/root", "mp255-ulrpm", "/path/to/release")
	methods := adapter.GetMethods()
	
	if len(methods) != 2 {
		t.Errorf("GetMethods() returned %d methods, want 2", len(methods))
	}

	expectedIDs := []string{"uuu", "manual"}
	for i, method := range methods {
		if method.ID != expectedIDs[i] {
			t.Errorf("Method %d ID = %v, want %v", i, method.ID, expectedIDs[i])
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/flash/... -run TestMP255Adapter -v`
Expected: FAIL - "undefined: NewMP255Adapter"

**Step 3: Implement MP255 adapter stub**

Create `internal/flash/mp255_adapter.go`:

```go
package flash

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// MP255Adapter wraps deploy.sh for MP255 platforms
type MP255Adapter struct {
	root       string
	platform   string
	releasePath string
}

// NewMP255Adapter creates a new MP255 adapter
func NewMP255Adapter(root, platform, releasePath string) *MP255Adapter {
	return &MP255Adapter{
		root:       root,
		platform:   platform,
		releasePath: releasePath,
	}
}

// PlatformType returns "mp255"
func (a *MP255Adapter) PlatformType() string {
	return "mp255"
}

// SupportsMethodSelection returns true (UUU vs Manual choice)
func (a *MP255Adapter) SupportsMethodSelection() bool {
	return true
}

// GetMethods returns UUU and Manual/Rescue methods
func (a *MP255Adapter) GetMethods() []FlashMethod {
	return []FlashMethod{
		{
			ID:          "uuu",
			Name:        "Official UUU (automated)",
			Description: "Uses install_linux_fw_uuu.sh with serial automation",
		},
		{
			ID:          "manual",
			Name:        "Manual/Rescue (step-by-step)",
			Description: "DFU + Fastboot + UMS with manual control",
		},
	}
}

// GetSteps returns placeholder steps (will be implemented later)
func (a *MP255Adapter) GetSteps(method string) []FlashStep {
	// TODO: Parse deploy.sh for actual steps
	return []FlashStep{
		{
			ID:          "placeholder",
			Description: "Deploy.sh integration coming soon",
			PromptType:  PromptConfirm,
		},
	}
}

// ExecuteStep runs one step (stub for now)
func (a *MP255Adapter) ExecuteStep(ctx context.Context, step FlashStep) (string, error) {
	return "MP255 flashing not yet implemented", fmt.Errorf("not implemented")
}

// CanFlash validates prerequisites
func (a *MP255Adapter) CanFlash(imagePath string) error {
	// Check release directory exists
	if _, err := os.Stat(imagePath); err != nil {
		return fmt.Errorf("release directory not found: %w", err)
	}

	// Check deploy.sh exists
	deployScript := filepath.Join(a.root, "yocto/deploy.sh")
	if _, err := os.Stat(deployScript); err != nil {
		return fmt.Errorf("deploy.sh not found: %w", err)
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/flash/... -run TestMP255Adapter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/flash/mp255_adapter.go internal/flash/mp255_adapter_test.go
git commit -m "feat(flash): add MP255 adapter stub"
```

---

## Task 6: Flash Screen Platform Selection UI

**Files:**
- Modify: `internal/flash/model.go`
- Modify: `internal/flash/model_test.go`

**Step 1: Add platform selection to view test**

Add to `internal/flash/model_test.go`:

```go
func TestModel_ViewPlatformSelect(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputPath: "/path/to/mp255"},
		{ID: "cassia-x2000", OutputPath: "/path/to/cassia"},
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/flash/... -run TestModel_ViewPlatformSelect -v`
Expected: FAIL - view doesn't contain platform names

**Step 3: Implement platform selection UI**

Update `internal/flash/model.go` imports:

```go
import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)
```

Add to `Model` struct:

```go
cursor int // For platform/method selection
```

Update `View()` method:

```go
// View renders the UI
func (m *Model) View() string {
	switch m.state {
	case statePlatformSelect:
		return m.viewPlatformSelect()
	case stateMethodSelect:
		return m.viewMethodSelect()
	case stateFlashing:
		return m.viewFlashing()
	case stateComplete:
		return m.viewComplete()
	case stateError:
		return m.viewError()
	default:
		return "Loading..."
	}
}

func (m *Model) viewPlatformSelect() string {
	content := styles.Title.Render("Platform Selection") + "\n\n"
	content += "Select platform to flash:\n\n"

	for i, p := range m.Platforms {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		ptype := DetectPlatformType(p)
		content += fmt.Sprintf("%s %s  [%s]\n", cursor, p.ID, ptype)
	}

	content += fmt.Sprintf("\nImage: %s\n", m.imagePath())

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash"},
		Context:    "",
		Content:    content,
		ActionKeys: []components.KeyBinding{
			{Key: "enter", Label: "select"},
		},
		NavKeys: []components.KeyBinding{
			{Key: "j/k", Label: "navigate"},
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}

	return layout.Render()
}

func (m *Model) viewMethodSelect() string {
	return "Method selection coming soon"
}

func (m *Model) viewFlashing() string {
	return "Flashing..."
}

func (m *Model) viewComplete() string {
	return "Flash complete!"
}

func (m *Model) viewError() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	return "Unknown error"
}

func (m *Model) imagePath() string {
	if m.cursor < 0 || m.cursor >= len(m.Platforms) {
		return ""
	}
	return m.Platforms[m.cursor].OutputPath
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/flash/... -run TestModel_ViewPlatformSelect -v`
Expected: PASS

**Step 5: Add keyboard navigation test**

Add to `internal/flash/model_test.go`:

```go
func TestModel_KeyboardNavigation(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "platform1"},
		{ID: "platform2"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Press 'j' to move down
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model := m2.(*Model)
	if model.cursor != 1 {
		t.Errorf("After 'j', cursor = %d, want 1", model.cursor)
	}

	// Press 'k' to move up
	m3, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model2 := m3.(*Model)
	if model2.cursor != 0 {
		t.Errorf("After 'k', cursor = %d, want 0", model2.cursor)
	}
}
```

**Step 6: Implement keyboard navigation**

Update `Update()` in `internal/flash/model.go`:

```go
// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Transition from idle to platform select on first update
	if m.state == stateIdle {
		m.state = statePlatformSelect
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case statePlatformSelect:
		return m.handlePlatformSelectKeys(msg)
	case stateMethodSelect:
		return m.handleMethodSelectKeys(msg)
	default:
		return m, nil
	}
}

func (m *Model) handlePlatformSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.Platforms)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		// Select platform
		if m.cursor >= 0 && m.cursor < len(m.Platforms) {
			return m, func() tea.Msg {
				return PlatformSelectedMsg{PlatformID: m.Platforms[m.cursor].ID}
			}
		}
	}
	return m, nil
}

func (m *Model) handleMethodSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement method selection keys
	return m, nil
}
```

**Step 7: Run test to verify it passes**

Run: `go test ./internal/flash/... -run TestModel_KeyboardNavigation -v`
Expected: PASS

**Step 8: Commit**

```bash
git add internal/flash/model.go internal/flash/model_test.go
git commit -m "feat(flash): add platform selection UI and navigation"
```

---

## Task 7: Platform Selection Handler

**Files:**
- Modify: `internal/flash/model.go`
- Modify: `internal/flash/model_test.go`

**Step 1: Write platform selection test**

Add to `internal/flash/model_test.go`:

```go
func TestModel_PlatformSelection(t *testing.T) {
	platforms := []platform.Platform{
		{ID: "mp255-ulrpm", OutputPath: "/path/to/mp255"},
		{ID: "cassia-x2000", OutputPath: "/path/to/cassia"},
	}

	st := status.RepoStatus{Root: "/tmp/test"}
	m := NewModel(st, platforms)
	m.state = statePlatformSelect

	// Select MP255 (supports method selection)
	msg := PlatformSelectedMsg{PlatformID: "mp255-ulrpm"}
	m2, _ := m.Update(msg)
	model := m2.(*Model)

	if model.state != stateMethodSelect {
		t.Errorf("After selecting MP255, state = %v, want %v", model.state, stateMethodSelect)
	}

	// Select Cassia (goes straight to flashing)
	m = NewModel(st, platforms)
	m.state = statePlatformSelect
	m.cursor = 1 // cassia

	msg2 := PlatformSelectedMsg{PlatformID: "cassia-x2000"}
	m3, _ := m.Update(msg2)
	model2 := m3.(*Model)

	if model2.state != stateFlashing {
		t.Errorf("After selecting Cassia, state = %v, want %v", model2.state, stateFlashing)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/flash/... -run TestModel_PlatformSelection -v`
Expected: FAIL - PlatformSelectedMsg not handled

**Step 3: Handle platform selection**

Update `Update()` in `internal/flash/model.go`:

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Transition from idle to platform select on first update
	if m.state == stateIdle {
		m.state = statePlatformSelect
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case PlatformSelectedMsg:
		return m.handlePlatformSelected(msg)
	}

	return m, nil
}
```

Add handler method:

```go
func (m *Model) handlePlatformSelected(msg PlatformSelectedMsg) (tea.Model, tea.Cmd) {
	m.selectedPlatform = msg.PlatformID

	// Find platform
	var selectedPlatform *platform.Platform
	for i := range m.Platforms {
		if m.Platforms[i].ID == msg.PlatformID {
			selectedPlatform = &m.Platforms[i]
			break
		}
	}

	if selectedPlatform == nil {
		m.state = stateError
		m.err = fmt.Errorf("platform not found: %s", msg.PlatformID)
		return m, nil
	}

	// Detect platform type and create adapter
	ptype := DetectPlatformType(*selectedPlatform)
	switch ptype {
	case "mp255":
		m.adapter = NewMP255Adapter(m.Status.Root, selectedPlatform.ID, selectedPlatform.OutputPath)
		// MP255 needs method selection
		m.state = stateMethodSelect
		m.cursor = 0 // Reset cursor for method selection

	case "swupdate":
		m.adapter = NewSWUpdateAdapter(m.Status.Root, selectedPlatform.ID, selectedPlatform.OutputPath)
		// SWUpdate goes straight to flashing
		m.state = stateFlashing

	default:
		m.state = stateError
		m.err = fmt.Errorf("unsupported platform type: %s", ptype)
	}

	return m, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/flash/... -run TestModel_PlatformSelection -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/flash/model.go internal/flash/model_test.go
git commit -m "feat(flash): handle platform selection and adapter creation"
```

---

## Task 8: App Integration - Home Key Handler

**Files:**
- Modify: `internal/home/model.go`
- Modify: `internal/app/model.go`

**Step 1: Add flash nav to home model**

Update `Update()` in `internal/home/model.go` to handle `f` key:

Find the key handler section (around line 200):

```go
case "f":
	return m, func() tea.Msg { return nav.GoToFlash{} }
```

Add to imports in `internal/home/model.go`:

```go
import (
	// existing imports...
	"github.com/augurysys/augury-node-tui/internal/nav"
)
```

**Step 2: Add flash message to nav**

Add to `internal/nav/nav.go`:

```go
// GoToFlash navigates to flash screen
type GoToFlash struct{}
```

**Step 3: Update home buildActionKeys**

Update `buildActionKeys()` in `internal/home/model.go` to include flash key:

```go
func (m *Model) buildActionKeys() []components.KeyBinding {
	var keys []components.KeyBinding

	// ... existing keys ...

	keys = append(keys, components.KeyBinding{Key: "f", Label: "flash"})
	
	// ... rest of keys ...

	return keys
}
```

**Step 4: Add flash to app router**

Update `internal/app/model.go` imports:

```go
import (
	// existing imports...
	"github.com/augurysys/augury-node-tui/internal/flash"
)
```

Add flash field to `Model`:

```go
type Model struct {
	// ... existing fields ...
	flash *flash.Model
}
```

Update `Update()` in `internal/app/model.go` to handle `nav.GoToFlash`:

```go
case nav.GoToFlash:
	if m.flash == nil {
		m.flash = flash.NewModel(m.status, m.platforms)
	}
	m.screen = screenFlash
	return m, m.flash.Init()
```

Add `screenFlash` to screen constants:

```go
const (
	screenHome screen = iota
	screenBuild
	screenHydration
	screenCaches
	screenValidations
	screenHints
	screenCI
	screenFlash // Add this
)
```

Update view routing in `View()`:

```go
func (m *Model) View() string {
	switch m.screen {
	case screenHome:
		return m.home.View()
	case screenBuild:
		return m.build.View()
	case screenFlash:
		return m.flash.View()
	// ... other cases ...
	default:
		return "Unknown screen"
	}
}
```

Update message propagation in `Update()`:

```go
case tea.WindowSizeMsg:
	m.Width = msg.Width
	m.Height = msg.Height

	// Propagate to all screens
	if hm, _ := m.home.Update(msg); hm != nil {
		m.home = hm.(*home.Model)
	}
	// ... other screens ...
	if m.flash != nil {
		if fm, _ := m.flash.Update(msg); fm != nil {
			m.flash = fm.(*flash.Model)
		}
	}
```

**Step 5: Build and test**

Run: `go build ./cmd/augury-node-tui`
Expected: SUCCESS

Run TUI manually: `./augury-node-tui`
Test: Press `f` on home, should see flash screen

**Step 6: Commit**

```bash
git add internal/home/model.go internal/app/model.go internal/nav/nav.go
git commit -m "feat(flash): integrate flash screen into app navigation"
```

---

## Task 9: Method Selection UI (MP255)

**Files:**
- Modify: `internal/flash/model.go`

**Step 1: Implement method selection view**

Add `viewMethodSelect()` in `internal/flash/model.go`:

```go
func (m *Model) viewMethodSelect() string {
	if m.adapter == nil {
		return "No adapter"
	}

	methods := m.adapter.GetMethods()
	if len(methods) == 0 {
		return "No methods available"
	}

	content := styles.Title.Render("Choose Flash Method") + "\n\n"
	content += fmt.Sprintf("Platform: %s\n\n", m.selectedPlatform)
	content += "Select method:\n\n"

	for i, method := range methods {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		content += fmt.Sprintf("%s %d) %s\n", cursor, i+1, method.Name)
	}

	if m.cursor >= 0 && m.cursor < len(methods) {
		content += fmt.Sprintf("\n%s\n", methods[m.cursor].Description)
	}

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash", m.selectedPlatform},
		Context:    "",
		Content:    content,
		ActionKeys: []components.KeyBinding{
			{Key: "1/2", Label: "choose"},
		},
		NavKeys: []components.KeyBinding{
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}

	return layout.Render()
}
```

**Step 2: Implement method selection keys**

Update `handleMethodSelectKeys()`:

```go
func (m *Model) handleMethodSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	methods := m.adapter.GetMethods()

	switch msg.String() {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Parse number key
		idx := int(msg.Runes[0] - '1')
		if idx >= 0 && idx < len(methods) {
			m.selectedMethod = methods[idx].ID
			m.state = stateFlashing
			return m, nil
		}
	case "j", "down":
		if m.cursor < len(methods)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.cursor >= 0 && m.cursor < len(methods) {
			m.selectedMethod = methods[m.cursor].ID
			m.state = stateFlashing
			return m, nil
		}
	case "esc":
		m.state = statePlatformSelect
		m.cursor = 0
	}

	return m, nil
}
```

**Step 3: Build and test**

Run: `go build ./cmd/augury-node-tui`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/flash/model.go
git commit -m "feat(flash): add method selection UI for MP255"
```

---

## Task 10: Basic Flashing View (Placeholder)

**Files:**
- Modify: `internal/flash/model.go`

**Step 1: Implement flashing view**

Update `viewFlashing()` in `internal/flash/model.go`:

```go
func (m *Model) viewFlashing() string {
	if m.adapter == nil {
		return "No adapter"
	}

	steps := m.adapter.GetSteps(m.selectedMethod)
	
	content := styles.Title.Render("Flashing Firmware") + "\n\n"
	content += fmt.Sprintf("Platform: %s\n", m.selectedPlatform)
	if m.selectedMethod != "" {
		content += fmt.Sprintf("Method: %s\n\n", m.selectedMethod)
	} else {
		content += "\n"
	}

	for i, step := range steps {
		content += fmt.Sprintf("%d. %s\n", i+1, step.Description)
	}

	content += "\n\nFlashing will be implemented in next tasks."

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Flash", m.selectedPlatform},
		Context:    "Ready to flash",
		Content:    content,
		ActionKeys: []components.KeyBinding{
			{Key: "enter", Label: "start (placeholder)"},
		},
		NavKeys: []components.KeyBinding{
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}

	return layout.Render()
}
```

**Step 2: Build and test**

Run: `go build ./cmd/augury-node-tui`
Expected: SUCCESS

Test flow: `f` → select mp255 → select method → see placeholder

**Step 3: Commit**

```bash
git add internal/flash/model.go
git commit -m "feat(flash): add placeholder flashing view"
```

---

## Task 11: Documentation & Manual Testing

**Files:**
- Create: `internal/flash/README.md`
- Update: `docs/keybindings.md`

**Step 1: Document flash module**

Create `internal/flash/README.md`:

```markdown
# Flash Module

Firmware flashing for MP255 and SWUpdate platforms.

## Architecture

- `model.go`: Main Bubbletea model with state machine
- `adapter.go`: Platform-agnostic flash interface
- `mp255_adapter.go`: Wraps deploy.sh for MP255
- `swupdate_adapter.go`: Wraps augury_update for SWUpdate
- `platform.go`: Platform type detection

## States

1. `stateIdle` - Initial state
2. `statePlatformSelect` - Choose platform
3. `stateMethodSelect` - Choose flash method (MP255 only)
4. `stateFlashing` - Execute flash steps
5. `stateComplete` - Success
6. `stateError` - Failure

## Adding New Platforms

1. Implement `FlashAdapter` interface
2. Update `DetectPlatformType()` to recognize platform
3. Add adapter creation in `handlePlatformSelected()`

## Testing

```bash
# Unit tests
go test ./internal/flash/...

# Integration test (manual)
./augury-node-tui
# Press 'f' → select platform → follow prompts
```

## Current Status

**Implemented:**
- Platform selection UI
- Method selection UI (MP255)
- Adapter pattern foundation
- SWUpdate adapter (basic)
- MP255 adapter (stub)

**TODO:**
- MP255 deploy.sh integration
- Step execution with output streaming
- Error handling and retry
- Log file capture
- Cancel support
```

**Step 2: Update keybindings doc**

Add to `docs/keybindings.md`:

```markdown
## Flash Screen

**Access:** Press `f` from Home

### Platform Selection
- `j/k` or `↑/↓`: Navigate platforms
- `Enter`: Select platform
- `Esc`: Back to Home
- `q`: Quit

### Method Selection (MP255 only)
- `1/2`: Choose method by number
- `j/k` or `↑/↓`: Navigate methods
- `Enter`: Select method
- `Esc`: Back to platform selection

### Flashing
- `Enter`: Continue to next step
- `c`: Cancel flash
- `Esc`: Abort and return
```

**Step 3: Manual testing checklist**

Create testing checklist in commit message:

```bash
git add internal/flash/README.md docs/keybindings.md
git commit -m "docs(flash): add module documentation

Manual testing checklist:
- [ ] Navigate to flash screen with 'f' key
- [ ] Platform list shows all built platforms
- [ ] j/k navigation works
- [ ] Selecting MP255 shows method selection
- [ ] Selecting SWUpdate shows flashing view
- [ ] Method selection navigation works
- [ ] Back/esc navigation works correctly
- [ ] Window resize doesn't break layout
"
```

---

## Summary

**Completed:**
1. ✅ Adapter interface and types
2. ✅ Platform detection logic
3. ✅ Flash screen state machine
4. ✅ SWUpdate adapter (basic flow)
5. ✅ MP255 adapter (stub with methods)
6. ✅ Platform selection UI
7. ✅ Platform selection handler
8. ✅ App integration (navigation)
9. ✅ Method selection UI
10. ✅ Placeholder flashing view
11. ✅ Documentation

**Remaining Work (Future Tasks):**
- MP255 deploy.sh integration (parsing prompts, executing)
- Step execution with streaming output
- Error handling and display
- Cancel/abort functionality
- Log file capture
- Complete/success view
- SWUpdate adapter enhancements (verify/flash execution)

**Success Criteria Met:**
- Flash screen accessible via `f` key ✅
- Platform selection works ✅
- MP255 method selection works ✅
- SWUpdate direct path works ✅
- Unified ScreenLayout UI ✅
- Foundation for step execution ✅

**Next Phase:** Implement actual flashing logic (MP255 deploy.sh wrapper, SWUpdate execution, output streaming).
