# Onboarding Wizard Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build interactive setup wizard (`augury-node-tui setup`) that guides users through one-time configuration with auto-fixes, sudo guidance, and live Nix progress.

**Architecture:** CLI flag routing to setup vs. normal mode, wizard uses Bubble Tea with step-based UI models, config persisted to `~/.config/augury-node-tui/config.toml`, health checks in separate package for reusability.

**Tech Stack:** Go 1.24, Bubble Tea, Lip Gloss (existing), TOML config parsing (github.com/pelletier/go-toml/v2), clipboard (atotto/clipboard), existing internal/styles theme.

---

## Task 0: Add Config File Support

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Test: `internal/config/config_test.go`

**Step 1: Write the failing test**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_ReadWriteRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	
	cfg := Config{
		AuguryNodeRoot:    "/home/user/augury-node",
		BinaryInstalled:   true,
		NixVerified:       true,
		SetupCompletedAt:  "2026-03-08T19:45:00Z",
		CompletedSteps:    []string{"root", "nix"},
		SkippedSteps:      []string{"groups"},
	}
	
	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	
	loaded, err := Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	
	if loaded.AuguryNodeRoot != cfg.AuguryNodeRoot {
		t.Errorf("Root mismatch: got %q, want %q", loaded.AuguryNodeRoot, cfg.AuguryNodeRoot)
	}
	if loaded.BinaryInstalled != cfg.BinaryInstalled {
		t.Error("BinaryInstalled mismatch")
	}
}

func TestConfig_DefaultPath(t *testing.T) {
	path := DefaultPath()
	if !filepath.IsAbs(path) {
		t.Error("DefaultPath should return absolute path")
	}
	if !strings.Contains(path, ".config/augury-node-tui") {
		t.Errorf("DefaultPath should be in .config/augury-node-tui; got %q", path)
	}
}

func TestConfig_ReadNonexistent(t *testing.T) {
	_, err := Read("/nonexistent/config.toml")
	if err == nil {
		t.Error("Read should return error for nonexistent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Read should return os.IsNotExist error; got %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/... -v`  
Expected: FAIL with "package config is not in GOROOT"

**Step 3: Write minimal implementation**

Create `internal/config/config.go`:
```go
package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	AuguryNodeRoot   string   `toml:"augury_node_root"`
	BinaryInstalled  bool     `toml:"binary_installed"`
	NixVerified      bool     `toml:"nix_verified"`
	SetupCompletedAt string   `toml:"setup_completed_at"`
	CompletedSteps   []string `toml:"completed_steps"`
	SkippedSteps     []string `toml:"skipped_steps"`
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "augury-node-tui", "config.toml")
}

func Read(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Write(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

**Step 4: Add dependency and run test**

Run: `go get github.com/pelletier/go-toml/v2 && go test ./internal/config/... -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
go.mod go.sum
git commit -m "feat(config): add TOML config read/write support

Adds config file persistence at ~/.config/augury-node-tui/config.toml
for storing augury-node root path and setup completion state."
```

---

## Task 1: Add System Health Check Functions

**Files:**
- Create: `internal/setup/health.go`
- Create: `internal/setup/health_test.go`
- Test: `internal/setup/health_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHealth_CheckNixInstalled(t *testing.T) {
	// This checks if 'nix' command exists in PATH
	result := CheckNixInstalled()
	// Should return true on systems with Nix, false otherwise
	if result && exec.Command("nix", "--version").Run() != nil {
		t.Error("CheckNixInstalled returned true but nix command failed")
	}
}

func TestHealth_CheckNixExperimentalFeatures(t *testing.T) {
	// Mock test - check if it reads nix.conf
	dir := t.TempDir()
	confPath := filepath.Join(dir, "nix.conf")
	os.WriteFile(confPath, []byte("experimental-features = nix-command flakes\n"), 0644)
	
	result := checkNixConfFile(confPath)
	if !result {
		t.Error("Should detect experimental features in config")
	}
}

func TestHealth_CheckNixUsersGroup(t *testing.T) {
	// Check if user is in nix-users group
	result := CheckNixUsersGroup()
	// Result depends on system, just verify it doesn't panic
	_ = result
}

func TestHealth_FindAuguryNodeRoot(t *testing.T) {
	// Test auto-detection
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "scripts/devices"), 0755)
	os.MkdirAll(filepath.Join(dir, "scripts/lib"), 0755)
	os.MkdirAll(filepath.Join(dir, "pkg"), 0755)
	
	subdir := filepath.Join(dir, "some/nested/path")
	os.MkdirAll(subdir, 0755)
	
	found, err := FindAuguryNodeRoot(subdir)
	if err != nil {
		t.Fatalf("Should find root from nested path: %v", err)
	}
	if found != dir {
		t.Errorf("Found %q, want %q", found, dir)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -v`  
Expected: FAIL with "package setup is not in GOROOT"

**Step 3: Write minimal implementation**

Create `internal/setup/health.go`:
```go
package setup

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/workspace"
)

func CheckNixInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

func CheckNixExperimentalFeatures() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	userConf := filepath.Join(home, ".config", "nix", "nix.conf")
	if checkNixConfFile(userConf) {
		return true
	}
	return checkNixConfFile("/etc/nix/nix.conf")
}

func checkNixConfFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "nix-command") && strings.Contains(content, "flakes")
}

func CheckNixUsersGroup() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	gids, err := u.GroupIds()
	if err != nil {
		return false
	}
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)
		if err != nil {
			continue
		}
		if g.Name == "nix-users" || g.Name == "nixbld" {
			return true
		}
	}
	return false
}

func FindAuguryNodeRoot(cwd string) (string, error) {
	return workspace.ResolveRoot("", "", cwd)
}

func AutoFixNixConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "nix")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "nix.conf")
	
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "experimental-features") {
		content += "\nexperimental-features = nix-command flakes\n"
		return os.WriteFile(path, []byte(content), 0644)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/health.go internal/setup/health_test.go
git commit -m "feat(setup): add system health check functions

Checks for Nix installation, experimental features, group membership,
and augury-node root detection with auto-fix for Nix config."
```

---

## Task 2: Add Setup Wizard Root Detection Step

**Files:**
- Create: `internal/setup/step_root.go`
- Create: `internal/setup/step_root_test.go`
- Test: `internal/setup/step_root_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepRoot_AutoDetectDisplaysPath(t *testing.T) {
	step := NewRootStep("/detected/augury-node")
	view := step.View()
	
	if !strings.Contains(view, "/detected/augury-node") {
		t.Errorf("View should show detected path; got %q", view)
	}
	if !strings.Contains(view, "Found") || !strings.Contains(view, "✓") {
		t.Error("View should show success indicator for detected path")
	}
}

func TestStepRoot_ManualEntry(t *testing.T) {
	step := NewRootStep("")
	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	
	if step.mode != "manual" {
		t.Error("Pressing 'c' should switch to manual entry mode")
	}
}

func TestStepRoot_Confirmed(t *testing.T) {
	step := NewRootStep("/valid/path")
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyEnter})
	
	if !step.Confirmed() {
		t.Error("Enter should confirm the step")
	}
	if cmd == nil {
		t.Error("Should return command to proceed")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestStepRoot -v`  
Expected: FAIL with "undefined: NewRootStep"

**Step 3: Write minimal implementation**

Create `internal/setup/step_root.go`:
```go
package setup

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/augurysys/augury-node-tui/internal/workspace"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type RootStepModel struct {
	detectedPath string
	manualPath   string
	mode         string
	confirmed    bool
	textInput    textinput.Model
	validationErr error
}

func NewRootStep(detected string) *RootStepModel {
	ti := textinput.New()
	ti.Placeholder = "/path/to/augury-node"
	ti.Width = 50
	return &RootStepModel{
		detectedPath: detected,
		mode:         "auto",
		textInput:    ti,
	}
}

func (m *RootStepModel) Init() tea.Cmd {
	return nil
}

func (m *RootStepModel) Update(msg tea.Msg) (*RootStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == "manual" {
			switch msg.Type {
			case tea.KeyEnter:
				path := strings.TrimSpace(m.textInput.Value())
				if err := workspace.ValidateRoot(path); err != nil {
					m.validationErr = err
					return m, nil
				}
				m.manualPath = path
				m.confirmed = true
				return m, func() tea.Msg { return NextStepMsg{} }
			case tea.KeyEsc:
				m.mode = "auto"
				m.textInput.Reset()
				m.validationErr = nil
				return m, nil
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}
		
		switch msg.String() {
		case "c":
			m.mode = "manual"
			m.textInput.Focus()
			return m, textinput.Blink
		case "enter":
			if m.detectedPath != "" {
				m.confirmed = true
				return m, func() tea.Msg { return NextStepMsg{} }
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *RootStepModel) View() string {
	var sections []string
	
	header := styles.Header.Render("📁 Augury-Node Repository")
	sections = append(sections, header)
	
	if m.mode == "auto" && m.detectedPath != "" {
		status := fmt.Sprintf("  %s  %s",
			styles.Success.Render("✓ Found at:"),
			styles.Highlight.Render(m.detectedPath))
		sections = append(sections, status)
		sections = append(sections, "")
		sections = append(sections, "  "+styles.Dim.Render("This will be saved to config."))
		sections = append(sections, "")
		sections = append(sections, "  "+styles.KeyBinding("enter", "Continue")+"  "+styles.KeyBinding("c", "Change Path"))
	} else if m.mode == "auto" && m.detectedPath == "" {
		sections = append(sections, "  "+styles.Warning.Render("✗ Not auto-detected"))
		sections = append(sections, "")
		sections = append(sections, "  "+styles.Dim.Render("Press 'c' to enter path manually"))
		sections = append(sections, "")
		sections = append(sections, "  "+styles.KeyBinding("c", "Enter Path")+"  "+styles.KeyBinding("q", "Cancel"))
	} else {
		sections = append(sections, "  "+styles.Dim.Render("Enter augury-node repository path:"))
		sections = append(sections, "")
		sections = append(sections, "  "+m.textInput.View())
		if m.validationErr != nil {
			sections = append(sections, "")
			sections = append(sections, "  "+styles.Error.Render("✗ "+m.validationErr.Error()))
		}
		sections = append(sections, "")
		sections = append(sections, "  "+styles.KeyBinding("enter", "Validate")+"  "+styles.KeyBinding("esc", "Back"))
	}
	
	content := strings.Join(sections, "\n")
	return styles.Border.Render(content)
}

func (m *RootStepModel) Confirmed() bool {
	return m.confirmed
}

func (m *RootStepModel) SelectedPath() string {
	if m.manualPath != "" {
		return m.manualPath
	}
	return m.detectedPath
}

type NextStepMsg struct{}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -run TestStepRoot -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_root.go internal/setup/step_root_test.go
git commit -m "feat(setup): add root detection step UI

Interactive step for detecting or manually entering augury-node path
with validation and confirmation."
```

---

## Task 3: Add Nix Configuration Step

**Files:**
- Create: `internal/setup/step_nix.go`
- Create: `internal/setup/step_nix_test.go`
- Test: `internal/setup/step_nix_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepNix_DisplaysChecks(t *testing.T) {
	step := NewNixStep()
	step.nixInstalled = true
	step.experimentalEnabled = false
	
	view := step.View()
	if !strings.Contains(view, "Nix installed") && !strings.Contains(view, "✓") {
		t.Error("View should show Nix installed status")
	}
	if !strings.Contains(view, "Experimental features") && !strings.Contains(view, "✗") {
		t.Error("View should show experimental features status")
	}
}

func TestStepNix_AutoFixExecutes(t *testing.T) {
	step := NewNixStep()
	step.nixInstalled = true
	step.experimentalEnabled = false
	
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	
	if step.state != "fixing" {
		t.Error("Pressing 'f' should trigger auto-fix")
	}
	if cmd == nil {
		t.Error("Should return command to execute fix")
	}
}

func TestStepNix_AllChecksPassAutoAdvances(t *testing.T) {
	step := NewNixStep()
	step.nixInstalled = true
	step.experimentalEnabled = true
	step.daemonOk = true
	
	if !step.AllChecksPassed() {
		t.Error("Should report all checks passed")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestStepNix -v`  
Expected: FAIL with "undefined: NewNixStep"

**Step 3: Write minimal implementation**

Create `internal/setup/step_nix.go`:
```go
package setup

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type NixStepModel struct {
	nixInstalled        bool
	experimentalEnabled bool
	daemonOk            bool
	state               string
	fixError            error
	confirmed           bool
}

func NewNixStep() *NixStepModel {
	return &NixStepModel{
		state: "checking",
	}
}

func (m *NixStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		return NixHealthCheckMsg{
			NixInstalled:        CheckNixInstalled(),
			ExperimentalEnabled: CheckNixExperimentalFeatures(),
			DaemonOk:            checkDaemonSocket(),
		}
	}
}

func (m *NixStepModel) Update(msg tea.Msg) (*NixStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case NixHealthCheckMsg:
		m.nixInstalled = msg.NixInstalled
		m.experimentalEnabled = msg.ExperimentalEnabled
		m.daemonOk = msg.DaemonOk
		m.state = "ready"
		
		if m.AllChecksPassed() {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil
		
	case NixFixResultMsg:
		m.state = "ready"
		if msg.Err != nil {
			m.fixError = msg.Err
			return m, nil
		}
		m.experimentalEnabled = true
		if m.AllChecksPassed() {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil
		
	case tea.KeyMsg:
		if m.state != "ready" {
			return m, nil
		}
		
		switch msg.String() {
		case "f":
			if !m.experimentalEnabled {
				m.state = "fixing"
				return m, func() tea.Msg {
					err := AutoFixNixConfig()
					return NixFixResultMsg{Err: err}
				}
			}
		case "s":
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *NixStepModel) View() string {
	var lines []string
	
	header := styles.Header.Render("⚙️  Nix Configuration")
	lines = append(lines, header)
	lines = append(lines, "")
	
	if m.state == "checking" {
		lines = append(lines, "  "+styles.Dim.Render("Checking Nix installation..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}
	
	if m.state == "fixing" {
		lines = append(lines, "  "+styles.Info.Render("🔧 Applying auto-fix..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}
	
	status1 := "✓"
	style1 := styles.Success
	if !m.nixInstalled {
		status1 = "✗"
		style1 = styles.Error
	}
	lines = append(lines, "  "+style1.Render(status1+" Nix installed"))
	
	status2 := "✓"
	style2 := styles.Success
	if !m.experimentalEnabled {
		status2 = "✗"
		style2 = styles.Error
	}
	lines = append(lines, "  "+style2.Render(status2+" Experimental features enabled"))
	
	status3 := "✓"
	style3 := styles.Success
	if !m.daemonOk {
		status3 = "✗"
		style3 = styles.Warning
	}
	lines = append(lines, "  "+style3.Render(status3+" Daemon accessible"))
	
	lines = append(lines, "")
	
	if m.fixError != nil {
		lines = append(lines, "  "+styles.Error.Render("Fix failed: "+m.fixError.Error()))
		lines = append(lines, "")
	}
	
	if !m.experimentalEnabled && m.nixInstalled {
		lines = append(lines, "  "+styles.Info.Render("Auto-fix available:"))
		lines = append(lines, "  "+styles.Dim.Render("Add to ~/.config/nix/nix.conf"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("f", "Auto-Fix")+"  "+styles.KeyBinding("s", "Skip"))
	} else if m.AllChecksPassed() {
		lines = append(lines, "  "+styles.Success.Render("All checks passed!"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Continue"))
	} else {
		lines = append(lines, "  "+styles.KeyBinding("s", "Skip")+"  "+styles.KeyBinding("q", "Cancel"))
	}
	
	content := strings.Join(lines, "\n")
	return styles.Border.Render(content)
}

func (m *NixStepModel) AllChecksPassed() bool {
	return m.nixInstalled && m.experimentalEnabled && m.daemonOk
}

func (m *NixStepModel) Confirmed() bool {
	return m.confirmed
}

func checkDaemonSocket() bool {
	_, err := os.Stat("/nix/var/nix/daemon-socket/socket")
	return err == nil
}

type NixHealthCheckMsg struct {
	NixInstalled        bool
	ExperimentalEnabled bool
	DaemonOk            bool
}

type NixFixResultMsg struct {
	Err error
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -run TestStepNix -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_nix.go internal/setup/step_nix_test.go
git commit -m "feat(setup): add Nix configuration step UI

Step checks Nix installation, experimental features, daemon access
with auto-fix option for experimental features."
```

---

## Task 4: Add Group Membership Step (Interactive Guide)

**Files:**
- Create: `internal/setup/step_groups.go`
- Create: `internal/setup/step_groups_test.go`
- Test: `internal/setup/step_groups_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepGroups_NotInGroupShowsCommands(t *testing.T) {
	step := NewGroupsStep()
	step.inNixUsers = false
	
	view := step.View()
	if !strings.Contains(view, "sudo usermod") {
		t.Error("View should show sudo command when not in group")
	}
	if !strings.Contains(view, "newgrp") {
		t.Error("View should show newgrp command")
	}
}

func TestStepGroups_AlreadyInGroupAutoAdvances(t *testing.T) {
	step := NewGroupsStep()
	step.inNixUsers = true
	
	if !step.Confirmed() {
		t.Error("Should auto-confirm if already in group")
	}
}

func TestStepGroups_RecheckUpdatesStatus(t *testing.T) {
	step := NewGroupsStep()
	step.inNixUsers = false
	
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	
	if step.state != "rechecking" {
		t.Error("Pressing 'r' should trigger recheck")
	}
	if cmd == nil {
		t.Error("Should return command to recheck")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestStepGroups -v`  
Expected: FAIL with "undefined: NewGroupsStep"

**Step 3: Write minimal implementation**

Create `internal/setup/step_groups.go`:
```go
package setup

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type GroupsStepModel struct {
	inNixUsers bool
	state      string
	confirmed  bool
}

func NewGroupsStep() *GroupsStepModel {
	return &GroupsStepModel{
		state: "checking",
	}
}

func (m *GroupsStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		return GroupCheckMsg{InNixUsers: CheckNixUsersGroup()}
	}
}

func (m *GroupsStepModel) Update(msg tea.Msg) (*GroupsStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case GroupCheckMsg:
		m.inNixUsers = msg.InNixUsers
		m.state = "ready"
		
		if m.inNixUsers {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil
		
	case tea.KeyMsg:
		if m.state != "ready" {
			return m, nil
		}
		
		switch msg.String() {
		case "r":
			m.state = "rechecking"
			return m, func() tea.Msg {
				return GroupCheckMsg{InNixUsers: CheckNixUsersGroup()}
			}
		case "s":
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		case "c":
			return m, copyToClipboard("sudo usermod -aG nix-users $USER && newgrp nix-users")
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *GroupsStepModel) View() string {
	var lines []string
	
	header := styles.Header.Render("👥 Permissions Setup")
	lines = append(lines, header)
	lines = append(lines, "")
	
	if m.state == "checking" || m.state == "rechecking" {
		lines = append(lines, "  "+styles.Dim.Render("Checking group membership..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}
	
	if m.inNixUsers {
		lines = append(lines, "  "+styles.Success.Render("✓ In nix-users group"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.Dim.Render("Permissions OK"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Continue"))
	} else {
		lines = append(lines, "  "+styles.Error.Render("✗ Not in nix-users group"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.Dim.Render("Run this command in a new terminal:"))
		lines = append(lines, "")
		
		cmd := styles.Info.Render("  sudo usermod -aG nix-users $USER")
		lines = append(lines, "  "+cmd)
		cmd2 := styles.Info.Render("  newgrp nix-users")
		lines = append(lines, "  "+cmd2)
		
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("c", "Copy")+"  "+styles.KeyBinding("r", "Re-check")+"  "+styles.KeyBinding("s", "Skip"))
	}
	
	content := strings.Join(lines, "\n")
	return styles.Border.Render(content)
}

func (m *GroupsStepModel) Confirmed() bool {
	return m.confirmed
}

type GroupCheckMsg struct {
	InNixUsers bool
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		// Will implement with atotto/clipboard in next task
		return ClipboardCopiedMsg{Success: true}
	}
}

type ClipboardCopiedMsg struct {
	Success bool
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -run TestStepGroups -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_groups.go internal/setup/step_groups_test.go
git commit -m "feat(setup): add group membership guide step

Interactive guide for adding user to nix-users group with
copy-to-clipboard and recheck functionality."
```

---

## Task 5: Add Binary Installation Step

**Files:**
- Create: `internal/setup/step_install.go`
- Create: `internal/setup/step_install_test.go`
- Test: `internal/setup/step_install_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepInstall_DisplaysTargetPath(t *testing.T) {
	step := NewInstallStep("/current/binary")
	view := step.View()
	
	if !strings.Contains(view, "/usr/local/bin") {
		t.Error("View should show target installation path")
	}
}

func TestStepInstall_AlreadyInstalledSkips(t *testing.T) {
	step := NewInstallStep("/current/binary")
	step.alreadyInstalled = true
	
	if !step.Confirmed() {
		t.Error("Should auto-confirm if already installed")
	}
}

func TestStepInstall_ShowsSudoCommand(t *testing.T) {
	step := NewInstallStep("/current/binary")
	step.alreadyInstalled = false
	
	view := step.View()
	if !strings.Contains(view, "sudo") {
		t.Error("View should show sudo requirement")
	}
	if !strings.Contains(view, "ln -sf") {
		t.Error("View should show symlink command")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestStepInstall -v`  
Expected: FAIL with "undefined: NewInstallStep"

**Step 3: Write minimal implementation**

Create `internal/setup/step_install.go`:
```go
package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

const targetBinaryPath = "/usr/local/bin/augury-node-tui"

type InstallStepModel struct {
	sourceBinary     string
	alreadyInstalled bool
	state            string
	confirmed        bool
}

func NewInstallStep(binaryPath string) *InstallStepModel {
	return &InstallStepModel{
		sourceBinary: binaryPath,
		state:        "checking",
	}
}

func (m *InstallStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		target, err := os.Readlink(targetBinaryPath)
		installed := err == nil && target == m.sourceBinary
		return InstallCheckMsg{AlreadyInstalled: installed}
	}
}

func (m *InstallStepModel) Update(msg tea.Msg) (*InstallStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case InstallCheckMsg:
		m.alreadyInstalled = msg.AlreadyInstalled
		m.state = "ready"
		
		if m.alreadyInstalled {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil
		
	case tea.KeyMsg:
		if m.state != "ready" {
			return m, nil
		}
		
		switch msg.String() {
		case "c":
			cmd := fmt.Sprintf("sudo ln -sf %s %s", m.sourceBinary, targetBinaryPath)
			return m, copyToClipboard(cmd)
		case "r":
			m.state = "rechecking"
			return m, func() tea.Msg {
				target, err := os.Readlink(targetBinaryPath)
				installed := err == nil && target == m.sourceBinary
				return InstallCheckMsg{AlreadyInstalled: installed}
			}
		case "s":
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *InstallStepModel) View() string {
	var lines []string
	
	header := styles.Header.Render("📦 Binary Installation")
	lines = append(lines, header)
	lines = append(lines, "")
	
	if m.state == "checking" || m.state == "rechecking" {
		lines = append(lines, "  "+styles.Dim.Render("Checking installation..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}
	
	if m.alreadyInstalled {
		lines = append(lines, "  "+styles.Success.Render("✓ Already installed at:"))
		lines = append(lines, "  "+styles.Highlight.Render(targetBinaryPath))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Continue"))
	} else {
		lines = append(lines, "  "+styles.Dim.Render("Install to: "+targetBinaryPath))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.Warning.Render("This requires sudo permission."))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.Dim.Render("Run in a new terminal:"))
		
		cmd := fmt.Sprintf("sudo ln -sf %s %s", m.sourceBinary, targetBinaryPath)
		cmdBox := "  " + styles.Info.Render(cmd)
		lines = append(lines, cmdBox)
		
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("c", "Copy")+"  "+styles.KeyBinding("r", "Re-check")+"  "+styles.KeyBinding("s", "Skip"))
	}
	
	content := strings.Join(lines, "\n")
	return styles.Border.Render(content)
}

func (m *InstallStepModel) Confirmed() bool {
	return m.confirmed
}

func (m *InstallStepModel) WasInstalled() bool {
	return m.alreadyInstalled || !m.confirmed
}

type InstallCheckMsg struct {
	AlreadyInstalled bool
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -run TestStepInstall -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_install.go internal/setup/step_install_test.go
git commit -m "feat(setup): add binary installation step

Guides user through installing binary to /usr/local/bin with
sudo command display, copy-to-clipboard, and recheck."
```

---

## Task 6: Add Nix Build Progress Step

**Files:**
- Create: `internal/setup/step_nixbuild.go`
- Create: `internal/setup/step_nixbuild_test.go`
- Test: `internal/setup/step_nixbuild_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepNixBuild_DisplaysProgress(t *testing.T) {
	step := NewNixBuildStep("/augury-node")
	step.state = "building"
	step.progress = BuildProgress{
		Built:      12,
		Total:      20,
		Downloaded: "1.2 GB",
		Current:    "gcc-wrapper-13.2.0",
	}
	
	view := step.View()
	if !strings.Contains(view, "12/20") {
		t.Error("View should show build progress")
	}
	if !strings.Contains(view, "gcc-wrapper") {
		t.Error("View should show current package")
	}
}

func TestStepNixBuild_SkipOption(t *testing.T) {
	step := NewNixBuildStep("/augury-node")
	step.state = "ready"
	
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	
	if !step.skipped {
		t.Error("Pressing 's' should mark as skipped")
	}
	if !step.Confirmed() {
		t.Error("Skip should confirm the step")
	}
}

func TestStepNixBuild_CompletionAutoAdvances(t *testing.T) {
	step := NewNixBuildStep("/augury-node")
	step.state = "building"
	
	step, cmd := step.Update(BuildCompleteMsg{Success: true})
	
	if !step.completed {
		t.Error("BuildCompleteMsg should mark as completed")
	}
	if !step.Confirmed() {
		t.Error("Completion should confirm the step")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestStepNixBuild -v`  
Expected: FAIL with "undefined: NewNixBuildStep"

**Step 3: Write minimal implementation**

Create `internal/setup/step_nixbuild.go`:
```go
package setup

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type BuildProgress struct {
	Built      int
	Total      int
	Downloaded string
	Current    string
	Elapsed    time.Duration
}

type NixBuildStepModel struct {
	root      string
	state     string
	progress  BuildProgress
	startTime time.Time
	completed bool
	skipped   bool
	buildErr  error
	confirmed bool
}

func NewNixBuildStep(root string) *NixBuildStepModel {
	return &NixBuildStepModel{
		root:  root,
		state: "ready",
	}
}

func (m *NixBuildStepModel) Init() tea.Cmd {
	return nil
}

func (m *NixBuildStepModel) Update(msg tea.Msg) (*NixBuildStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == "ready" {
			switch msg.String() {
			case "enter", "b":
				m.state = "building"
				m.startTime = time.Now()
				return m, m.startBuild()
			case "s":
				m.skipped = true
				m.confirmed = true
				return m, func() tea.Msg { return NextStepMsg{} }
			case "q":
				return m, tea.Quit
			}
		} else if m.state == "building" {
			switch msg.String() {
			case "s":
				m.skipped = true
				m.confirmed = true
				return m, func() tea.Msg { return NextStepMsg{} }
			}
		}
		
	case BuildProgressMsg:
		m.progress = msg.Progress
		m.progress.Elapsed = time.Since(m.startTime)
		return m, m.waitForProgress()
		
	case BuildCompleteMsg:
		m.state = "complete"
		m.buildErr = msg.Err
		m.completed = msg.Success
		m.confirmed = true
		return m, func() tea.Msg { return NextStepMsg{} }
	}
	
	return m, nil
}

func (m *NixBuildStepModel) startBuild() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, "nix", "develop", m.root+"/.#dev-env", "--command", "true")
		
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return BuildCompleteMsg{Success: false, Err: err}
		}
		
		if err := cmd.Start(); err != nil {
			return BuildCompleteMsg{Success: false, Err: err}
		}
		
		progressChan := make(chan BuildProgress)
		go parseNixProgress(stderr, progressChan)
		
		go func() {
			for prog := range progressChan {
				// Send progress updates
				_ = prog
			}
		}()
		
		err = cmd.Wait()
		return BuildCompleteMsg{Success: err == nil, Err: err}
	}
}

func (m *NixBuildStepModel) waitForProgress() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return BuildProgressMsg{Progress: m.progress}
	})
}

func (m *NixBuildStepModel) View() string {
	var lines []string
	
	header := styles.Header.Render("🔨 Building Nix Development Environment")
	lines = append(lines, header)
	lines = append(lines, "")
	
	if m.state == "ready" {
		lines = append(lines, "  "+styles.Dim.Render("This prepares tools for building augury-node platforms."))
		lines = append(lines, "  "+styles.Dim.Render("First run takes 5-10 minutes as packages download."))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Start Build")+"  "+styles.KeyBinding("s", "Skip"))
	} else if m.state == "building" {
		lines = append(lines, "  "+styles.Info.Render("Downloading and building packages..."))
		lines = append(lines, "")
		
		if m.progress.Total > 0 {
			pct := (m.progress.Built * 100) / m.progress.Total
			bar := renderProgressBar(pct)
			lines = append(lines, fmt.Sprintf("  %s %d%% (%d/%d packages)",
				bar, pct, m.progress.Built, m.progress.Total))
		}
		
		if m.progress.Current != "" {
			lines = append(lines, "")
			lines = append(lines, "  "+styles.Dim.Render("Current: "+m.progress.Current))
		}
		
		if m.progress.Downloaded != "" {
			lines = append(lines, "  "+styles.Dim.Render("Downloaded: "+m.progress.Downloaded))
		}
		
		if m.progress.Elapsed > 0 {
			elapsed := m.progress.Elapsed.Round(time.Second)
			lines = append(lines, "  "+styles.Dim.Render("Elapsed: "+elapsed.String()))
		}
		
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("s", "Skip"))
	} else if m.state == "complete" {
		if m.buildErr != nil {
			lines = append(lines, "  "+styles.Error.Render("✗ Build failed"))
			lines = append(lines, "")
			lines = append(lines, "  "+styles.Dim.Render("You can build manually later:"))
			lines = append(lines, "  "+styles.Info.Render("nix develop "+m.root+"/.#dev-env"))
		} else {
			lines = append(lines, "  "+styles.Success.Render("✓ Environment ready"))
		}
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Continue"))
	}
	
	content := strings.Join(lines, "\n")
	return styles.Border.Render(content)
}

func (m *NixBuildStepModel) Confirmed() bool {
	return m.confirmed
}

func renderProgressBar(percent int) string {
	width := 20
	filled := (percent * width) / 100
	bar := strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)
	return "[" + bar + "]"
}

func parseNixProgress(r io.Reader, out chan<- BuildProgress) {
	defer close(out)
	
	re := regexp.MustCompile(`\[(\d+)/(\d+) built.*?(\d+\.\d+ [GM]iB) DL\]`)
	currentRe := regexp.MustCompile(`building (.+)`)
	
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		
		if matches := re.FindStringSubmatch(line); len(matches) >= 4 {
			var built, total int
			fmt.Sscanf(matches[1], "%d", &built)
			fmt.Sscanf(matches[2], "%d", &total)
			
			prog := BuildProgress{
				Built:      built,
				Total:      total,
				Downloaded: matches[3],
			}
			
			if curr := currentRe.FindStringSubmatch(line); len(curr) >= 2 {
				prog.Current = curr[1]
			}
			
			out <- prog
		}
	}
}

type BuildProgressMsg struct {
	Progress BuildProgress
}

type BuildCompleteMsg struct {
	Success bool
	Err     error
}

type InstallCheckMsg struct {
	AlreadyInstalled bool
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -run TestStepNixBuild -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_nixbuild.go internal/setup/step_nixbuild_test.go
git commit -m "feat(setup): add Nix environment build step

Live progress display for first-time Nix environment build with
package count, download size, current package, and elapsed time."
```

---

## Task 7: Add Success Screen Step

**Files:**
- Create: `internal/setup/step_success.go`
- Create: `internal/setup/step_success_test.go`
- Test: `internal/setup/step_success_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"strings"
	"testing"
)

func TestStepSuccess_DisplaysSummary(t *testing.T) {
	step := NewSuccessStep(true, true, true)
	view := step.View()
	
	if !strings.Contains(view, "Setup Complete") || !strings.Contains(view, "✅") {
		t.Error("View should show success message")
	}
	if !strings.Contains(view, "augury-node-tui") {
		t.Error("View should mention the command")
	}
}

func TestStepSuccess_ShowsSkippedWarnings(t *testing.T) {
	step := NewSuccessStep(true, false, false)
	step.skippedSteps = []string{"groups", "install"}
	
	view := step.View()
	if !strings.Contains(view, "skipped") {
		t.Error("View should mention skipped steps")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestStepSuccess -v`  
Expected: FAIL with "undefined: NewSuccessStep"

**Step 3: Write minimal implementation**

Create `internal/setup/step_success.go`:
```go
package setup

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type SuccessStepModel struct {
	nixReady      bool
	binaryInstalled bool
	groupsOk      bool
	skippedSteps  []string
	launchRequested bool
}

func NewSuccessStep(nixReady, binary, groups bool) *SuccessStepModel {
	return &SuccessStepModel{
		nixReady:        nixReady,
		binaryInstalled: binary,
		groupsOk:        groups,
	}
}

func (m *SuccessStepModel) Init() tea.Cmd {
	return nil
}

func (m *SuccessStepModel) Update(msg tea.Msg) (*SuccessStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "l":
			m.launchRequested = true
			return m, func() tea.Msg { return LaunchMainTUIMsg{} }
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *SuccessStepModel) View() string {
	var lines []string
	
	header := styles.Success.Render("✅ Setup Complete!")
	lines = append(lines, "  "+header)
	lines = append(lines, "")
	
	lines = append(lines, "  "+styles.Dim.Render("augury-node-tui is ready to use."))
	lines = append(lines, "")
	
	lines = append(lines, "  "+styles.Dim.Render("Configuration saved to:"))
	lines = append(lines, "  "+styles.Highlight.Render("~/.config/augury-node-tui/config.toml"))
	lines = append(lines, "")
	
	if len(m.skippedSteps) > 0 {
		lines = append(lines, "  "+styles.Warning.Render("⚠ Some steps were skipped:"))
		for _, step := range m.skippedSteps {
			lines = append(lines, "    "+styles.Dim.Render("- "+step))
		}
		lines = append(lines, "")
	}
	
	lines = append(lines, "  "+styles.Dim.Render("Quick start:"))
	lines = append(lines, "  "+styles.Info.Render("$ augury-node-tui"))
	lines = append(lines, "")
	
	lines = append(lines, "  "+styles.KeyBinding("l", "Launch Now")+"  "+styles.KeyBinding("q", "Exit"))
	
	content := strings.Join(lines, "\n")
	return styles.Border.Render(content)
}

func (m *SuccessStepModel) LaunchRequested() bool {
	return m.launchRequested
}

type LaunchMainTUIMsg struct{}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -run TestStepSuccess -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_success.go internal/setup/step_success_test.go
git commit -m "feat(setup): add success completion step

Success screen with summary, skipped steps warning, and option
to immediately launch main TUI."
```

---

## Task 8: Add Wizard Orchestration

**Files:**
- Create: `internal/setup/wizard.go`
- Create: `internal/setup/wizard_test.go`
- Test: `internal/setup/wizard_test.go`

**Step 1: Write the failing test**

```go
package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWizard_InitStartsAtRootStep(t *testing.T) {
	w := NewWizard()
	if w.currentStep != 0 {
		t.Error("Wizard should start at step 0")
	}
}

func TestWizard_NextStepMsgAdvances(t *testing.T) {
	w := NewWizard()
	w.currentStep = 0
	
	w, _ = w.Update(NextStepMsg{})
	
	if w.currentStep != 1 {
		t.Errorf("NextStepMsg should advance step; got %d, want 1", w.currentStep)
	}
}

func TestWizard_ViewShowsProgressIndicator(t *testing.T) {
	w := NewWizard()
	w.currentStep = 2
	
	view := w.View()
	if !strings.Contains(view, "Step") || !strings.Contains(view, "/6") {
		t.Error("View should show step progress indicator")
	}
}

func TestWizard_LaunchMainTUIExits(t *testing.T) {
	w := NewWizard()
	w.currentStep = 5
	
	w, cmd := w.Update(LaunchMainTUIMsg{})
	
	if w.launchMain != true {
		t.Error("LaunchMainTUIMsg should set launchMain flag")
	}
	if cmd == nil {
		t.Error("Should return quit command")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestWizard -v`  
Expected: FAIL with "undefined: NewWizard"

**Step 3: Write minimal implementation**

Create `internal/setup/wizard.go`:
```go
package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/config"
	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type WizardModel struct {
	currentStep   int
	stepRoot      *RootStepModel
	stepNix       *NixStepModel
	stepGroups    *GroupsStepModel
	stepInstall   *InstallStepModel
	stepNixBuild  *NixBuildStepModel
	stepSuccess   *SuccessStepModel
	config        config.Config
	launchMain    bool
}

func NewWizard() *WizardModel {
	cwd, _ := os.Getwd()
	detected, _ := FindAuguryNodeRoot(cwd)
	
	binaryPath, _ := os.Executable()
	
	return &WizardModel{
		currentStep:  0,
		stepRoot:     NewRootStep(detected),
		stepNix:      NewNixStep(),
		stepGroups:   NewGroupsStep(),
		stepInstall:  NewInstallStep(binaryPath),
		stepNixBuild: nil,
		stepSuccess:  nil,
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return m.stepRoot.Init()
}

func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NextStepMsg:
		return m.advanceStep()
		
	case LaunchMainTUIMsg:
		m.launchMain = true
		return m, tea.Quit
		
	case tea.KeyMsg:
		if msg.String() == "q" && m.currentStep == 0 {
			return m, tea.Quit
		}
	}
	
	return m.updateCurrentStep(msg)
}

func (m *WizardModel) updateCurrentStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch m.currentStep {
	case 0:
		m.stepRoot, cmd = m.stepRoot.Update(msg)
	case 1:
		m.stepNix, cmd = m.stepNix.Update(msg)
	case 2:
		m.stepGroups, cmd = m.stepGroups.Update(msg)
	case 3:
		m.stepInstall, cmd = m.stepInstall.Update(msg)
	case 4:
		m.stepNixBuild, cmd = m.stepNixBuild.Update(msg)
	case 5:
		m.stepSuccess, cmd = m.stepSuccess.Update(msg)
	}
	
	return m, cmd
}

func (m *WizardModel) advanceStep() (tea.Model, tea.Cmd) {
	if m.currentStep == 0 && m.stepRoot.Confirmed() {
		root := m.stepRoot.SelectedPath()
		m.config.AuguryNodeRoot = root
		m.config.CompletedSteps = append(m.config.CompletedSteps, "root")
		m.saveConfig()
	}
	
	if m.currentStep == 1 && m.stepNix.Confirmed() {
		m.config.NixVerified = m.stepNix.AllChecksPassed()
		m.config.CompletedSteps = append(m.config.CompletedSteps, "nix")
		m.saveConfig()
		
		if !m.stepNix.AllChecksPassed() {
			m.config.SkippedSteps = append(m.config.SkippedSteps, "nix")
		}
	}
	
	if m.currentStep == 2 && m.stepGroups.Confirmed() {
		m.config.CompletedSteps = append(m.config.CompletedSteps, "groups")
		m.saveConfig()
		
		if !m.stepGroups.inNixUsers {
			m.config.SkippedSteps = append(m.config.SkippedSteps, "groups")
		}
	}
	
	if m.currentStep == 3 && m.stepInstall.Confirmed() {
		m.config.BinaryInstalled = m.stepInstall.alreadyInstalled
		m.config.CompletedSteps = append(m.config.CompletedSteps, "install")
		m.saveConfig()
		
		if !m.stepInstall.WasInstalled() {
			m.config.SkippedSteps = append(m.config.SkippedSteps, "install")
		}
		
		m.stepNixBuild = NewNixBuildStep(m.config.AuguryNodeRoot)
	}
	
	if m.currentStep == 4 && m.stepNixBuild.Confirmed() {
		m.config.CompletedSteps = append(m.config.CompletedSteps, "nixbuild")
		m.config.SetupCompletedAt = time.Now().Format(time.RFC3339)
		m.saveConfig()
		
		if m.stepNixBuild.skipped {
			m.config.SkippedSteps = append(m.config.SkippedSteps, "nixbuild")
		}
		
		m.stepSuccess = NewSuccessStep(
			m.stepNixBuild.completed,
			m.stepInstall.alreadyInstalled,
			m.stepGroups.inNixUsers,
		)
		m.stepSuccess.skippedSteps = m.config.SkippedSteps
	}
	
	m.currentStep++
	
	var cmd tea.Cmd
	switch m.currentStep {
	case 1:
		cmd = m.stepNix.Init()
	case 2:
		cmd = m.stepGroups.Init()
	case 3:
		cmd = m.stepInstall.Init()
	case 4:
		cmd = m.stepNixBuild.Init()
	case 5:
		cmd = m.stepSuccess.Init()
	}
	
	return m, cmd
}

func (m *WizardModel) saveConfig() {
	path := config.DefaultPath()
	config.Write(path, m.config)
}

func (m *WizardModel) View() string {
	var sections []string
	
	title := styles.Title.Render(fmt.Sprintf("🧙 Setup Wizard [Step %d/6]", m.currentStep+1))
	sections = append(sections, title)
	
	var stepView string
	switch m.currentStep {
	case 0:
		stepView = m.stepRoot.View()
	case 1:
		stepView = m.stepNix.View()
	case 2:
		stepView = m.stepGroups.View()
	case 3:
		stepView = m.stepInstall.View()
	case 4:
		stepView = m.stepNixBuild.View()
	case 5:
		stepView = m.stepSuccess.View()
	}
	
	sections = append(sections, stepView)
	
	return strings.Join(sections, "\n")
}

func (m *WizardModel) LaunchMainRequested() bool {
	return m.launchMain
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -run TestWizard -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/wizard.go internal/setup/wizard_test.go
git commit -m "feat(setup): add wizard orchestration

Main wizard model that sequences steps, manages config persistence,
and handles step transitions with progress indicator."
```

---

## Task 9: Add Clipboard Support

**Files:**
- Modify: `internal/setup/step_groups.go`
- Modify: `internal/setup/step_install.go`
- Test: `internal/setup/step_groups_test.go`

**Step 1: Write the failing test**

Add to `internal/setup/step_groups_test.go`:
```go
func TestStepGroups_CopyReturnsSuccess(t *testing.T) {
	step := NewGroupsStep()
	step.inNixUsers = false
	
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	
	if cmd == nil {
		t.Error("Copy should return command")
	}
	
	msg := cmd()
	if _, ok := msg.(ClipboardCopiedMsg); !ok {
		t.Errorf("Copy should return ClipboardCopiedMsg; got %T", msg)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/setup/... -run TestStepGroups_Copy -v`  
Expected: FAIL (clipboard not implemented)

**Step 3: Add dependency and implement**

Run: `go get github.com/atotto/clipboard`

Update `copyToClipboard` function in `step_groups.go` and `step_install.go`:
```go
func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		return ClipboardCopiedMsg{Success: err == nil, Text: text}
	}
}
```

Add import: `"github.com/atotto/clipboard"`

Update both step models to handle `ClipboardCopiedMsg`:
```go
case ClipboardCopiedMsg:
	if msg.Success {
		// Could show brief success indicator
	}
	return m, nil
```

Update `ClipboardCopiedMsg` type:
```go
type ClipboardCopiedMsg struct {
	Success bool
	Text    string
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/setup/... -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_groups.go internal/setup/step_install.go internal/setup/step_groups_test.go go.mod go.sum
git commit -m "feat(setup): implement clipboard copy for sudo commands

Uses atotto/clipboard for cross-platform clipboard support.
Pressing 'c' copies command to clipboard for easy pasting."
```

---

## Task 10: Wire Setup Command to Main

**Files:**
- Modify: `cmd/augury-node-tui/main.go`
- Test: Manual testing (CLI flag routing)

**Step 1: Update main.go to route commands**

Modify `cmd/augury-node-tui/main.go`:
```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/augurysys/augury-node-tui/internal/app"
	"github.com/augurysys/augury-node-tui/internal/config"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/setup"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/workspace"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		runSetupWizard()
		return
	}
	
	runMainTUI()
}

func runSetupWizard() {
	w := setup.NewWizard()
	p := tea.NewProgram(w, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup wizard error: %v\n", err)
		os.Exit(1)
	}
	
	wizard := model.(*setup.WizardModel)
	if wizard.LaunchMainRequested() {
		runMainTUI()
	}
}

func runMainTUI() {
	cfg, err := config.Read(config.DefaultPath())
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "augury-node-tui: no configuration found\n")
			fmt.Fprintf(os.Stderr, "Run 'augury-node-tui setup' first.\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "augury-node-tui: config error: %v\n", err)
		os.Exit(1)
	}
	
	root := cfg.AuguryNodeRoot
	if err := workspace.ValidateRoot(root); err != nil {
		fmt.Fprintf(os.Stderr, "augury-node-tui: configured root invalid: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'augury-node-tui setup --reconfigure' to fix.\n")
		os.Exit(1)
	}
	
	st, err := status.Collect(root)
	if err != nil {
		dirty := make(map[string]bool)
		for _, p := range status.RequiredPaths {
			dirty[p] = false
		}
		st = status.RepoStatus{Root: root, Branch: "?", SHA: "?", Dirty: dirty}
	}
	
	m := app.NewModel(st, platform.Registry(), 2*time.Second)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "augury-node-tui: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Test manually**

Run:
```bash
cd /tmp
go build -o augury-node-tui /home/ngurfinkel/Repos/augury-node-tui/cmd/augury-node-tui

# Should show error asking to run setup
./augury-node-tui

# Should launch wizard
./augury-node-tui setup
```

Expected:
- First command: Error message about running setup
- Second command: Wizard launches

**Step 3: Commit**

```bash
git add cmd/augury-node-tui/main.go
git commit -m "feat(cli): wire setup wizard to main command

Routes 'augury-node-tui setup' to wizard, reads config for normal mode.
Shows helpful error if config missing."
```

---

## Task 11: Add Integration Tests for Wizard Flow

**Files:**
- Create: `integration/setup_wizard_test.go`
- Test: `integration/setup_wizard_test.go`

**Step 1: Write the test**

```go
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/setup"
	tea "github.com/charmbracelet/bubbletea"
)

func TestWizard_RootStepAutoDetect(t *testing.T) {
	root := setupFakeAuguryNode(t)
	os.Chdir(root)
	
	w := setup.NewWizard()
	view := w.View()
	
	if !strings.Contains(view, root) {
		t.Errorf("Wizard should auto-detect root %q; view:\n%s", root, view)
	}
}

func TestWizard_FullFlowSimulation(t *testing.T) {
	root := setupFakeAuguryNode(t)
	os.Chdir(root)
	
	w := setup.NewWizard()
	
	w, _ = w.Update(w.Init()())
	
	w, _ = w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.(*setup.WizardModel).CurrentStep() != 1 {
		t.Error("Should advance to Nix step after root confirmation")
	}
	
	w, _ = w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if w.(*setup.WizardModel).CurrentStep() != 2 {
		t.Error("Should advance to groups step after Nix skip")
	}
}

func TestWizard_ConfigPersistence(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.toml")
	
	t.Setenv("HOME", filepath.Dir(configDir))
	
	root := setupFakeAuguryNode(t)
	w := setup.NewWizard()
	
	w, _ = w.Update(setup.NextStepMsg{})
	
	cfg, err := config.Read(configPath)
	if err != nil {
		t.Fatalf("Config should be saved after step completion: %v", err)
	}
	
	if cfg.AuguryNodeRoot == "" {
		t.Error("Config should have root path saved")
	}
}
```

**Step 2: Run test to verify behavior**

Run: `go test ./integration/... -run TestWizard -v`  
Expected: PASS (or minor adjustments needed)

**Step 3: Commit**

```bash
git add integration/setup_wizard_test.go
git commit -m "test(setup): add integration tests for wizard flow

Tests auto-detection, step progression, config persistence,
and full wizard simulation."
```

---

## Task 12: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `docs/configuration.md`
- Create: `docs/setup-wizard.md`

**Step 1: Update README.md**

Replace Quick Start section:
```markdown
## Quick Start

### First Time Setup

```bash
# 1. Build the TUI
cd /path/to/augury-node-tui
go build -o augury-node-tui ./cmd/augury-node-tui

# 2. Run setup wizard (from anywhere)
./augury-node-tui setup
```

The setup wizard will guide you through:
- Finding your augury-node repository
- Configuring Nix with experimental features
- Setting up permissions (nix-users group)
- Installing binary to PATH
- Building Nix development environment (first run only)

After setup completes, just run:
```bash
augury-node-tui
```

### Manual Setup

If you prefer manual setup, see [docs/configuration.md](docs/configuration.md).

## Re-running Setup

```bash
augury-node-tui setup --reconfigure
```
```

**Step 2: Create setup wizard documentation**

Create `docs/setup-wizard.md`:
```markdown
# Setup Wizard Guide

## Overview

The setup wizard (`augury-node-tui setup`) provides interactive guided configuration for first-time users.

## Steps

### 1. Augury-Node Root Detection

Auto-detects augury-node repository by searching ancestor directories.
If not found, prompts for manual entry.

**Keys:**
- `enter` - Confirm detected path
- `c` - Change to manual entry
- `q` - Cancel setup

### 2. Nix Configuration

Checks for:
- Nix installation (`nix` command in PATH)
- Experimental features enabled
- Daemon socket accessibility

**Auto-fixes:**
- Adds `experimental-features = nix-command flakes` to `~/.config/nix/nix.conf`

**Keys:**
- `f` - Auto-fix experimental features
- `s` - Skip (continue anyway)
- `q` - Cancel

### 3. Group Membership

Checks if user is in `nix-users` group.

**Guided action:**
Displays commands to add user to group:
```bash
sudo usermod -aG nix-users $USER
newgrp nix-users
```

**Keys:**
- `c` - Copy commands to clipboard
- `r` - Re-check group membership
- `s` - Skip
- `q` - Cancel

### 4. Binary Installation

Installs symlink to `/usr/local/bin/augury-node-tui`.

**Guided action:**
Displays command:
```bash
sudo ln -sf /source/path /usr/local/bin/augury-node-tui
```

**Keys:**
- `c` - Copy command to clipboard
- `r` - Re-check installation
- `s` - Skip (use full path instead)
- `q` - Cancel

### 5. Nix Environment Build

Triggers first-time Nix development environment build.

Shows live progress:
- Package count (built/total)
- Download size
- Current package being built
- Elapsed time

**Keys:**
- `s` - Skip (build manually later with `nix develop .#dev-env`)

### 6. Success

Shows completion summary and any skipped steps.

**Keys:**
- `l` - Launch main TUI immediately
- `q` - Exit

## Config File

Setup wizard saves configuration to:
```
~/.config/augury-node-tui/config.toml
```

Contents:
```toml
augury_node_root = "/home/user/Repos/augury-node"
binary_installed = true
nix_verified = true
setup_completed_at = "2026-03-08T19:45:00Z"
completed_steps = ["root", "nix", "groups", "install", "nixbuild"]
skipped_steps = []
```

## Troubleshooting

**Setup fails mid-way:**
- Config is saved incrementally
- Re-run `augury-node-tui setup` to resume

**Config is invalid:**
- Run `augury-node-tui setup --reconfigure` to start over

**Still having issues:**
- Delete `~/.config/augury-node-tui/config.toml`
- Run `augury-node-tui setup` for fresh start
```

**Step 3: Commit**

```bash
git add README.md docs/setup-wizard.md docs/configuration.md
git commit -m "docs: update for setup wizard workflow

Replaces manual setup instructions with setup wizard guide.
Adds comprehensive wizard documentation."
```

---

## Task 13: Add --reconfigure Flag Support

**Files:**
- Modify: `cmd/augury-node-tui/main.go`
- Test: Manual testing

**Step 1: Update main to handle --reconfigure**

Modify `main()`:
```go
func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "setup" {
			reconfigure := len(os.Args) > 2 && os.Args[2] == "--reconfigure"
			runSetupWizard(reconfigure)
			return
		}
	}
	
	runMainTUI()
}

func runSetupWizard(reconfigure bool) {
	if reconfigure {
		path := config.DefaultPath()
		os.Remove(path)
	}
	
	w := setup.NewWizard()
	p := tea.NewProgram(w, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup wizard error: %v\n", err)
		os.Exit(1)
	}
	
	wizard := model.(*setup.WizardModel)
	if wizard.LaunchMainRequested() {
		runMainTUI()
	}
}
```

**Step 2: Test manually**

Run:
```bash
augury-node-tui setup --reconfigure
```

Expected: Wizard starts fresh (ignores existing config)

**Step 3: Commit**

```bash
git add cmd/augury-node-tui/main.go
git commit -m "feat(cli): add --reconfigure flag for setup wizard

Allows re-running setup from scratch by deleting existing config.
Usage: augury-node-tui setup --reconfigure"
```

---

## Task 14: Add Contract Tests for Setup Wizard

**Files:**
- Modify: `docs/contract_test.go`
- Test: `docs/contract_test.go`

**Step 1: Write the failing tests**

Add to `docs/contract_test.go`:
```go
func TestSetupWizard_READMEMentionsSetupCommand(t *testing.T) {
	readme := readFileOrFail(t, "../README.md")
	if !strings.Contains(readme, "augury-node-tui setup") {
		t.Error("README must mention 'augury-node-tui setup' command")
	}
}

func TestSetupWizard_DocsExist(t *testing.T) {
	files := []string{
		"../docs/setup-wizard.md",
		"../docs/plans/2026-03-08-onboarding-wizard-design.md",
		"../docs/plans/2026-03-08-onboarding-wizard-implementation.md",
	}
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("Setup wizard doc should exist: %s", f)
		}
	}
}

func TestSetupWizard_ConfigFileSchema(t *testing.T) {
	wizardDoc := readFileOrFail(t, "../docs/setup-wizard.md")
	required := []string{
		"augury_node_root",
		"binary_installed",
		"nix_verified",
		"setup_completed_at",
		"completed_steps",
		"skipped_steps",
	}
	for _, field := range required {
		if !strings.Contains(wizardDoc, field) {
			t.Errorf("Setup wizard docs must document config field: %s", field)
		}
	}
}
```

**Step 2: Run test to verify they fail** (if docs not yet updated)

Run: `go test ./docs/... -v`  
Expected: FAIL if any docs missing

**Step 3: Ensure docs are complete** (from Task 12)

Already completed in Task 12.

**Step 4: Run test to verify they pass**

Run: `go test ./docs/... -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add docs/contract_test.go
git commit -m "test(docs): add contract tests for setup wizard

Ensures README mentions setup command and wizard docs exist
with complete config schema documentation."
```

---

## Testing & Validation

**After all tasks complete:**

1. **Unit tests pass:**
   ```bash
   go test ./internal/setup/... -v
   go test ./internal/config/... -v
   ```

2. **Integration tests pass:**
   ```bash
   go test ./integration/... -v
   ```

3. **Contract tests pass:**
   ```bash
   go test ./docs/... -v
   ```

4. **Manual end-to-end test:**
   ```bash
   # Clean slate
   rm ~/.config/augury-node-tui/config.toml
   
   # Run setup from random directory
   cd /tmp
   /path/to/augury-node-tui setup
   
   # Verify each step works
   # After completion, verify main TUI launches
   augury-node-tui
   ```

5. **Edge case testing:**
   - Setup from non-augury-node directory
   - Setup with Nix already configured
   - Setup with binary already installed
   - Skip all optional steps
   - Interrupt wizard mid-way, resume

---

## Dependencies

New dependencies to add:
```bash
go get github.com/pelletier/go-toml/v2
go get github.com/atotto/clipboard
go get github.com/charmbracelet/bubbles/textinput
```

Existing dependencies (already available):
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- internal/styles (theme)
- internal/workspace (validation)

---

## Commit Strategy

Small, focused commits after each task:
- Task 0: Config support
- Task 1: Health checks
- Tasks 2-7: Individual step UIs
- Task 8: Wizard orchestration
- Task 9: Clipboard support
- Task 10: CLI routing
- Task 11: Integration tests
- Task 12: Documentation
- Task 13: Reconfigure flag
- Task 14: Contract tests

Each commit is independently testable and reviewable.
