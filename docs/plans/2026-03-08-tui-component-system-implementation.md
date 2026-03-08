# TUI Component System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Establish three-tier component hierarchy (atoms, molecules, organisms) with design token system, enabling consistent UX and rapid Phase 2/3/4 feature development.

**Architecture:** Incremental refactor in 3 waves - Wave 1 (primitives + simple screens), Wave 2 (molecules + medium screens), Wave 3 (complex components + advanced screens). Each wave validated before advancing.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, Bubbles (table/viewport), go-runewidth, gopsutil/v3

---

## Wave 1: Foundation & Primitives

### Task 1: Design Token System

**Files:**
- Create: `internal/styles/tokens.go`
- Create: `internal/styles/tokens_test.go`

**Step 1: Write the failing test**

Create `internal/styles/tokens_test.go`:

```go
package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestPalette_AllColorsDefined(t *testing.T) {
	p := DefaultPalette()
	
	colors := []string{
		p.Base, p.Surface0, p.Overlay0, p.Text,
		p.Success, p.Warning, p.Error, p.Info,
		p.AccentPink, p.AccentMauve, p.AccentPeach, p.AccentTeal,
	}
	
	for _, color := range colors {
		if color == "" {
			t.Errorf("color not defined")
		}
		// Validate hex format
		if len(color) != 7 || color[0] != '#' {
			t.Errorf("invalid hex color: %s", color)
		}
	}
}

func TestTypography_AllStylesDefined(t *testing.T) {
	typo := DefaultTypography()
	
	// Verify each style is usable
	_ = typo.Title.Render("test")
	_ = typo.Section.Render("test")
	_ = typo.Body.Render("test")
	_ = typo.Dim.Render("test")
	_ = typo.Highlight.Render("test")
}

func TestBorders_ThickThinNone(t *testing.T) {
	borders := DefaultBorders()
	
	if borders.Thick == lipgloss.Border{} {
		t.Error("Thick border not defined")
	}
	if borders.Thin == lipgloss.Border{} {
		t.Error("Thin border not defined")
	}
	if borders.None != lipgloss.Border{} {
		t.Error("None border should be empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/styles -v`  
Expected: FAIL with "no such file or directory"

**Step 3: Write minimal implementation**

Create `internal/styles/tokens.go`:

```go
package styles

import "github.com/charmbracelet/lipgloss"

// Palette defines the application color scheme (Catppuccin Mocha)
type Palette struct {
	// Base colors
	Base     string // #1E1E2E (background)
	Surface0 string // #313244 (elevated surfaces)
	Overlay0 string // #6C7086 (dimmed text)
	Text     string // #CDD6F4 (primary text)

	// Semantic status colors
	Success string // #A6E3A1 (green)
	Warning string // #F9E2AF (yellow)
	Error   string // #F38BA8 (red)
	Info    string // #89B4FA (blue)

	// Accent colors (categorical mapping)
	AccentPink  string // #F5C2E7 (platforms)
	AccentMauve string // #CBA6F7 (builds)
	AccentPeach string // #FAB387 (caches)
	AccentTeal  string // #94E2D5 (validations)
}

// DefaultPalette returns Catppuccin Mocha palette
func DefaultPalette() Palette {
	return Palette{
		Base:     "#1E1E2E",
		Surface0: "#313244",
		Overlay0: "#6C7086",
		Text:     "#CDD6F4",

		Success: "#A6E3A1",
		Warning: "#F9E2AF",
		Error:   "#F38BA8",
		Info:    "#89B4FA",

		AccentPink:  "#F5C2E7",
		AccentMauve: "#CBA6F7",
		AccentPeach: "#FAB387",
		AccentTeal:  "#94E2D5",
	}
}

// Typography defines text styles
type Typography struct {
	Title     lipgloss.Style
	Section   lipgloss.Style
	Body      lipgloss.Style
	Dim       lipgloss.Style
	Highlight lipgloss.Style
}

// DefaultTypography returns standard text styles
func DefaultTypography() Typography {
	p := DefaultPalette()
	return Typography{
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.AccentPink)),
		Section:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Text)),
		Body:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.Text)),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color(p.Overlay0)),
		Highlight: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.AccentMauve)),
	}
}

// Borders defines border styles
type Borders struct {
	Thick lipgloss.Border
	Thin  lipgloss.Border
	None  lipgloss.Border
}

// DefaultBorders returns standard border styles
func DefaultBorders() Borders {
	return Borders{
		Thick: lipgloss.ThickBorder(),
		Thin:  lipgloss.NormalBorder(),
		None:  lipgloss.Border{},
	}
}

// Spacing constants
const (
	SpacingCompact  = 0
	SpacingNormal   = 1
	SpacingSpaciousm = 2
)
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/styles -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/styles/
git commit -m "feat: add design token system with Catppuccin Mocha palette"
```

---

### Task 2: Card Primitive

**Files:**
- Create: `internal/components/primitives/card.go`
- Create: `internal/components/primitives/card_test.go`

**Step 1: Write the failing test**

Create `internal/components/primitives/card_test.go`:

```go
package primitives

import (
	"strings"
	"testing"
)

func TestCard_RenderWithTitle(t *testing.T) {
	card := Card{
		Title:   "Test Card",
		Content: "This is content",
		Style:   CardNormal,
	}
	
	rendered := card.Render(40)
	
	if !strings.Contains(rendered, "Test Card") {
		t.Error("Card should contain title")
	}
	if !strings.Contains(rendered, "This is content") {
		t.Error("Card should contain content")
	}
	// Should have border characters
	if !strings.Contains(rendered, "─") && !strings.Contains(rendered, "┌") {
		t.Error("Card should have borders")
	}
}

func TestCard_WordWrapping(t *testing.T) {
	longContent := strings.Repeat("word ", 50)
	card := Card{
		Content: longContent,
		Style:   CardNormal,
	}
	
	rendered := card.Render(20)
	lines := strings.Split(rendered, "\n")
	
	for _, line := range lines {
		// Strip ANSI codes for accurate width check
		cleanLine := stripAnsi(line)
		if len(cleanLine) > 22 { // 20 + border padding
			t.Errorf("Line exceeds max width: %d chars", len(cleanLine))
		}
	}
}

func TestCard_StyleVariants(t *testing.T) {
	compact := Card{Content: "test", Style: CardCompact}
	normal := Card{Content: "test", Style: CardNormal}
	emphasized := Card{Content: "test", Style: CardEmphasized}
	
	compactLines := len(strings.Split(compact.Render(40), "\n"))
	normalLines := len(strings.Split(normal.Render(40), "\n"))
	
	// Compact should have fewer lines (no padding)
	if compactLines >= normalLines {
		t.Error("Compact style should have fewer lines than normal")
	}
	
	// Emphasized should be visually distinct (test by rendering)
	empRender := emphasized.Render(40)
	if empRender == normal.Render(40) {
		t.Error("Emphasized style should differ from normal")
	}
}

// Helper to strip ANSI codes
func stripAnsi(s string) string {
	// Simple ANSI strip: remove \x1b[...m sequences
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/components/primitives -v`  
Expected: FAIL with "no such file or directory"

**Step 3: Write minimal implementation**

Create `internal/components/primitives/card.go`:

```go
package primitives

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/augury-node-tui/internal/styles"
)

// CardStyle defines card appearance variants
type CardStyle int

const (
	CardCompact    CardStyle = iota // No padding
	CardNormal                       // Normal padding
	CardEmphasized                   // Thick border, accent color
)

// Card is a bordered container with optional title
type Card struct {
	Title   string
	Content string
	Style   CardStyle
}

// Render produces the styled card within given width
func (c Card) Render(width int) string {
	palette := styles.DefaultPalette()
	borders := styles.DefaultBorders()
	typo := styles.DefaultTypography()
	
	// Choose border and style based on variant
	var border lipgloss.Border
	var borderColor string
	var padding int
	
	switch c.Style {
	case CardCompact:
		border = borders.Thin
		borderColor = palette.Overlay0
		padding = 0
	case CardEmphasized:
		border = borders.Thick
		borderColor = palette.AccentMauve
		padding = 1
	default: // CardNormal
		border = borders.Thin
		borderColor = palette.Text
		padding = 1
	}
	
	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width - 2). // Account for borders
		Padding(0, padding)
	
	// Word-wrap content
	wrapped := wordWrap(c.Content, width-4-padding*2) // Account for borders + padding
	
	// Combine title and content
	var content string
	if c.Title != "" {
		titleLine := typo.Section.Render(c.Title)
		content = titleLine + "\n" + wrapped
	} else {
		content = wrapped
	}
	
	return style.Render(content)
}

// wordWrap breaks text at word boundaries to fit width
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder
	
	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" " + word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}
	
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	
	return strings.Join(lines, "\n")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/components/primitives -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/components/primitives/
git commit -m "feat: add Card primitive component with word wrapping"
```

---

### Task 3: StatusBadge Primitive

**Files:**
- Create: `internal/components/primitives/statusbadge.go`
- Modify: `internal/components/primitives/card_test.go` (add stripAnsi to shared test utils)
- Create: `internal/components/primitives/statusbadge_test.go`

**Step 1: Write the failing test**

Create `internal/components/primitives/statusbadge_test.go`:

```go
package primitives

import (
	"strings"
	"testing"
)

func TestStatusBadge_AllStatusTypes(t *testing.T) {
	statuses := []Status{
		StatusSuccess,
		StatusError,
		StatusWarning,
		StatusRunning,
		StatusBlocked,
		StatusUnavailable,
	}
	
	for _, status := range statuses {
		badge := StatusBadge{Label: "Test", Status: status}
		rendered := badge.Render()
		
		if rendered == "" {
			t.Errorf("Status %d produced empty render", status)
		}
		if !strings.Contains(rendered, "Test") {
			t.Errorf("Badge missing label for status %d", status)
		}
	}
}

func TestStatusBadge_IconMapping(t *testing.T) {
	tests := []struct {
		status       Status
		expectedIcon string
	}{
		{StatusSuccess, "✓"},
		{StatusError, "✗"},
		{StatusWarning, "⚠"},
		{StatusRunning, "●"},
		{StatusBlocked, "⊘"},
		{StatusUnavailable, "◌"},
	}
	
	for _, tt := range tests {
		badge := StatusBadge{Label: "Test", Status: tt.status}
		rendered := badge.Render()
		
		if !strings.Contains(rendered, tt.expectedIcon) {
			t.Errorf("Status %d should contain icon %s, got: %s",
				tt.status, tt.expectedIcon, stripAnsi(rendered))
		}
	}
}

func TestStatusBadge_ColorMapping(t *testing.T) {
	// Success should have green color code
	success := StatusBadge{Label: "OK", Status: StatusSuccess}
	successRender := success.Render()
	
	// Error should have red color code
	error := StatusBadge{Label: "Failed", Status: StatusError}
	errorRender := error.Render()
	
	// Renders should differ (different ANSI codes)
	if successRender == errorRender {
		t.Error("Success and error badges should have different colors")
	}
	
	// Both should contain ANSI escape sequences
	if !strings.Contains(successRender, "\x1b[") {
		t.Error("Success badge should contain color codes")
	}
	if !strings.Contains(errorRender, "\x1b[") {
		t.Error("Error badge should contain color codes")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/components/primitives -v -run TestStatusBadge`  
Expected: FAIL with "undefined: StatusBadge"

**Step 3: Write minimal implementation**

Create `internal/components/primitives/statusbadge.go`:

```go
package primitives

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/augury-node-tui/internal/styles"
)

// Status represents state for status-based color mapping
type Status int

const (
	StatusSuccess Status = iota
	StatusError
	StatusWarning
	StatusRunning
	StatusBlocked
	StatusUnavailable
)

// StatusBadge displays a colored status indicator
type StatusBadge struct {
	Label  string
	Status Status
}

// Render produces the styled status badge
func (b StatusBadge) Render() string {
	palette := styles.DefaultPalette()
	
	var icon string
	var color string
	
	switch b.Status {
	case StatusSuccess:
		icon = "✓"
		color = palette.Success
	case StatusError:
		icon = "✗"
		color = palette.Error
	case StatusWarning:
		icon = "⚠"
		color = palette.Warning
	case StatusRunning:
		icon = "●"
		color = palette.Info
	case StatusBlocked:
		icon = "⊘"
		color = palette.Overlay0
	case StatusUnavailable:
		icon = "◌"
		color = palette.Overlay0
	}
	
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	return style.Render(icon + " " + b.Label)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/components/primitives -v -run TestStatusBadge`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/components/primitives/
git commit -m "feat: add StatusBadge primitive with status-based color mapping"
```

---

### Task 4: KeyHint Primitive

**Files:**
- Create: `internal/components/primitives/keyhint.go`
- Create: `internal/components/primitives/keyhint_test.go`

**Step 1: Write the failing test**

Create `internal/components/primitives/keyhint_test.go`:

```go
package primitives

import (
	"strings"
	"testing"
)

func TestKeyHint_EnabledState(t *testing.T) {
	hint := KeyHint{
		Key:         "b",
		Description: "build",
		Enabled:     true,
	}
	
	rendered := hint.Render()
	
	if !strings.Contains(rendered, "b") {
		t.Error("Should contain key")
	}
	if !strings.Contains(rendered, "build") {
		t.Error("Should contain description")
	}
	// Enabled should not be dimmed (no "blocked" text)
	if strings.Contains(strings.ToLower(rendered), "blocked") {
		t.Error("Enabled hint should not show 'blocked'")
	}
}

func TestKeyHint_DisabledState(t *testing.T) {
	hint := KeyHint{
		Key:         "b",
		Description: "build",
		Enabled:     false,
	}
	
	rendered := hint.Render()
	
	if !strings.Contains(rendered, "b") {
		t.Error("Should contain key")
	}
	if !strings.Contains(rendered, "build") {
		t.Error("Should contain description")
	}
	
	// Disabled hints should be visually distinct (dimmed color)
	// We can't easily test ANSI color codes, but we can test structure
	enabledHint := KeyHint{Key: "b", Description: "build", Enabled: true}
	if rendered == enabledHint.Render() {
		t.Error("Disabled hint should render differently than enabled")
	}
}

func TestKeyHint_Format(t *testing.T) {
	hint := KeyHint{
		Key:         "ctrl+c",
		Description: "cancel",
		Enabled:     true,
	}
	
	rendered := hint.Render()
	
	// Should have bracket format [key]
	if !strings.Contains(rendered, "[") || !strings.Contains(rendered, "]") {
		t.Error("Key hint should use bracket format")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/components/primitives -v -run TestKeyHint`  
Expected: FAIL with "undefined: KeyHint"

**Step 3: Write minimal implementation**

Create `internal/components/primitives/keyhint.go`:

```go
package primitives

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/augury-node-tui/internal/styles"
)

// KeyHint displays a keyboard shortcut with description
type KeyHint struct {
	Key         string
	Description string
	Enabled     bool
}

// Render produces the styled key hint
func (h KeyHint) Render() string {
	palette := styles.DefaultPalette()
	typo := styles.DefaultTypography()
	
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette.AccentMauve)).
		Bold(true)
	
	var descStyle lipgloss.Style
	if h.Enabled {
		descStyle = typo.Body
	} else {
		descStyle = typo.Dim
	}
	
	keyPart := keyStyle.Render(fmt.Sprintf("[%s]", h.Key))
	descPart := descStyle.Render(h.Description)
	
	return fmt.Sprintf("%s %s", keyPart, descPart)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/components/primitives -v -run TestKeyHint`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/components/primitives/
git commit -m "feat: add KeyHint primitive for consistent keybinding display"
```

---

### Task 5: ProgressBar Primitive

**Files:**
- Create: `internal/components/primitives/progressbar.go`
- Create: `internal/components/primitives/progressbar_test.go`

**Step 1: Write the failing test**

Create `internal/components/primitives/progressbar_test.go`:

```go
package primitives

import (
	"strings"
	"testing"
)

func TestProgressBar_ZeroProgress(t *testing.T) {
	bar := ProgressBar{
		Current: 0,
		Total:   100,
		Width:   20,
		Label:   "Building",
	}
	
	rendered := bar.Render()
	
	if !strings.Contains(rendered, "Building") {
		t.Error("Should contain label")
	}
	if !strings.Contains(rendered, "0%") {
		t.Error("Should show 0%")
	}
}

func TestProgressBar_FullProgress(t *testing.T) {
	bar := ProgressBar{
		Current: 100,
		Total:   100,
		Width:   20,
		Label:   "Done",
	}
	
	rendered := bar.Render()
	
	if !strings.Contains(rendered, "100%") {
		t.Error("Should show 100%")
	}
	// Should be mostly filled
	if strings.Count(rendered, "█") < 15 {
		t.Errorf("Expected mostly filled bar, got: %s", stripAnsi(rendered))
	}
}

func TestProgressBar_PartialProgress(t *testing.T) {
	bar := ProgressBar{
		Current: 50,
		Total:   100,
		Width:   20,
		Label:   "Building",
	}
	
	rendered := bar.Render()
	
	if !strings.Contains(rendered, "50%") {
		t.Error("Should show 50%")
	}
	
	// Should have both filled and unfilled blocks
	hasFilled := strings.Contains(rendered, "█")
	hasUnfilled := strings.Contains(rendered, "░")
	
	if !hasFilled || !hasUnfilled {
		t.Errorf("Expected mix of filled/unfilled blocks, got: %s", stripAnsi(rendered))
	}
}

func TestProgressBar_FractionDisplay(t *testing.T) {
	bar := ProgressBar{
		Current: 820,
		Total:   1000,
		Width:   30,
		Label:   "Packages",
	}
	
	rendered := bar.Render()
	
	if !strings.Contains(rendered, "820") || !strings.Contains(rendered, "1000") {
		t.Errorf("Should show fraction (820/1000), got: %s", stripAnsi(rendered))
	}
}

func TestProgressBar_HandleZeroTotal(t *testing.T) {
	bar := ProgressBar{
		Current: 0,
		Total:   0,
		Width:   20,
		Label:   "Empty",
	}
	
	// Should not panic
	rendered := bar.Render()
	if rendered == "" {
		t.Error("Should render even with zero total")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/components/primitives -v -run TestProgressBar`  
Expected: FAIL with "undefined: ProgressBar"

**Step 3: Write minimal implementation**

Create `internal/components/primitives/progressbar.go`:

```go
package primitives

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/augury-node-tui/internal/styles"
)

// ProgressBar displays progress as filled/unfilled blocks
type ProgressBar struct {
	Current int
	Total   int
	Width   int
	Label   string
}

// Render produces the styled progress bar
func (p ProgressBar) Render() string {
	palette := styles.DefaultPalette()
	typo := styles.DefaultTypography()
	
	// Calculate percentage
	var pct float64
	if p.Total > 0 {
		pct = float64(p.Current) / float64(p.Total) * 100
	}
	
	// Calculate filled blocks
	barWidth := p.Width - len(p.Label) - 15 // Reserve space for label + percentage + fraction
	if barWidth < 10 {
		barWidth = 10
	}
	
	filledBlocks := int(float64(barWidth) * pct / 100)
	if filledBlocks > barWidth {
		filledBlocks = barWidth
	}
	
	// Build bar visual
	filled := strings.Repeat("█", filledBlocks)
	unfilled := strings.Repeat("░", barWidth-filledBlocks)
	
	// Color based on progress (sequential mapping)
	var barColor string
	if pct < 50 {
		barColor = palette.Overlay0
	} else if pct < 80 {
		barColor = palette.Warning
	} else {
		barColor = palette.Success
	}
	
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
	
	// Format: "Label: ████░░ 82% (820/1000)"
	labelPart := typo.Body.Render(p.Label + ": ")
	barPart := barStyle.Render(filled + unfilled)
	pctPart := typo.Body.Render(fmt.Sprintf(" %.0f%%", pct))
	fractionPart := typo.Dim.Render(fmt.Sprintf(" (%d/%d)", p.Current, p.Total))
	
	return labelPart + barPart + pctPart + fractionPart
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/components/primitives -v -run TestProgressBar`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/components/primitives/
git commit -m "feat: add ProgressBar primitive with sequential color mapping"
```

---

### Task 6: Refactor Success Screen

**Files:**
- Modify: `internal/setup/step_success.go`
- Modify: `internal/setup/step_success_test.go` (update to new rendering)

**Step 1: Write test for new card-based layout**

Add to `internal/setup/step_success_test.go`:

```go
func TestSuccessStep_UsesCardComponent(t *testing.T) {
	step := setup.NewSuccessStep([]string{})
	view := step.View()
	
	// Should have card borders
	if !strings.Contains(view, "─") && !strings.Contains(view, "┌") {
		t.Error("Success screen should use Card component with borders")
	}
	
	// Should contain success message
	if !strings.Contains(view, "Setup Complete") {
		t.Error("Should show success message")
	}
}
```

**Step 2: Run test to verify current behavior**

Run: `go test ./internal/setup -v -run TestSuccessStep_UsesCardComponent`  
Expected: FAIL (current implementation doesn't use Card)

**Step 3: Refactor step_success.go to use Card**

Modify `internal/setup/step_success.go`:

```go
package setup

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/augury-node-tui/internal/components/primitives"
	"github.com/augury-node-tui/internal/styles"
)

// ... existing SuccessStepModel definition ...

func (s *SuccessStepModel) View() string {
	typo := styles.DefaultTypography()
	
	// Build content
	var content strings.Builder
	content.WriteString("✓ Setup completed successfully!\n\n")
	content.WriteString("Next steps:\n")
	content.WriteString("  1. Run commands from the TUI\n")
	content.WriteString("  2. Explore build/hydrate/caches screens\n")
	
	if len(s.skippedSteps) > 0 {
		content.WriteString("\n⚠ Some steps were skipped:\n")
		for _, step := range s.skippedSteps {
			content.WriteString("  - " + step + "\n")
		}
	}
	
	// Use Card component
	card := primitives.Card{
		Title:   "Setup Complete",
		Content: content.String(),
		Style:   primitives.CardEmphasized,
	}
	
	cardView := card.Render(80)
	
	// Add key hints using KeyHint component
	launchHint := primitives.KeyHint{
		Key:         "enter",
		Description: "Launch main TUI",
		Enabled:     true,
	}
	quitHint := primitives.KeyHint{
		Key:         "q",
		Description: "Quit",
		Enabled:     true,
	}
	
	keyHints := "\n" + launchHint.Render() + "  •  " + quitHint.Render()
	
	return cardView + keyHints
}
```

**Step 4: Run test to verify new behavior**

Run: `go test ./internal/setup -v -run TestSuccessStep`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/
git commit -m "refactor: migrate success screen to Card and KeyHint components"
```

---

### Task 7: Refactor Hints Screen

**Files:**
- Modify: `internal/hints/model.go`
- Modify: `internal/hints/model_test.go`

**Step 1: Add test for card-based layout**

Add to `internal/hints/model_test.go`:

```go
func TestHintsScreen_UsesComponents(t *testing.T) {
	m := hints.NewModel()
	view := m.View()
	
	// Should use Card component (has borders)
	if !strings.Contains(view, "─") {
		t.Error("Hints screen should use Card component")
	}
	
	// Should use KeyHint components (bracket format)
	if !strings.Contains(view, "[") || !strings.Contains(view, "]") {
		t.Error("Hints screen should use KeyHint components")
	}
}
```

**Step 2: Run test to verify current behavior fails**

Run: `go test ./internal/hints -v -run TestHintsScreen_UsesComponents`  
Expected: FAIL

**Step 3: Refactor hints screen**

Modify `internal/hints/model.go`:

```go
package hints

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/augury-node-tui/internal/components/primitives"
	"github.com/augury-node-tui/internal/styles"
	"strings"
)

// ... existing Model definition ...

func (m Model) View() string {
	typo := styles.DefaultTypography()
	
	// Build hints content
	var content strings.Builder
	
	content.WriteString("Navigation:\n")
	hints := []primitives.KeyHint{
		{Key: "j/k", Description: "Navigate up/down", Enabled: true},
		{Key: "tab", Description: "Switch panes", Enabled: true},
		{Key: "q", Description: "Back/Quit", Enabled: true},
	}
	for _, hint := range hints {
		content.WriteString("  " + hint.Render() + "\n")
	}
	
	content.WriteString("\nActions:\n")
	actionHints := []primitives.KeyHint{
		{Key: "b", Description: "Build platform", Enabled: m.NixReady},
		{Key: "h", Description: "Hydrate artifacts", Enabled: m.NixReady},
		{Key: "c", Description: "Manage caches", Enabled: true},
		{Key: "v", Description: "Run validations", Enabled: true},
	}
	for _, hint := range actionHints {
		content.WriteString("  " + hint.Render() + "\n")
	}
	
	// Use Card component
	card := primitives.Card{
		Title:   "Keyboard Shortcuts",
		Content: content.String(),
		Style:   primitives.CardNormal,
	}
	
	return card.Render(m.Width)
}
```

**Step 4: Run test to verify pass**

Run: `go test ./internal/hints -v -run TestHintsScreen_UsesComponents`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/hints/
git commit -m "refactor: migrate hints screen to Card and KeyHint components"
```

---

## Wave 2: Molecules & Medium Complexity

### Task 8: DataTable Component (Foundation)

**Files:**
- Create: `internal/components/datatable.go`
- Create: `internal/components/datatable_test.go`

**Step 1: Write the failing test**

Create `internal/components/datatable_test.go`:

```go
package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testRow struct {
	Name   string
	Status string
	Count  int
}

func TestDataTable_Creation(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 20, Sortable: true},
		{Header: "Status", Width: 10, Sortable: false},
	}
	
	table := NewDataTable(columns)
	
	if table == nil {
		t.Fatal("NewDataTable returned nil")
	}
	
	view := table.View()
	if !strings.Contains(view, "Name") || !strings.Contains(view, "Status") {
		t.Error("Table should render column headers")
	}
}

func TestDataTable_SetRows(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 20, Renderer: func(r interface{}) string {
			return r.(testRow).Name
		}},
	}
	
	table := NewDataTable(columns)
	rows := []interface{}{
		testRow{Name: "Row1"},
		testRow{Name: "Row2"},
	}
	
	table.SetRows(rows)
	view := table.View()
	
	if !strings.Contains(view, "Row1") {
		t.Error("Table should render row data")
	}
}

func TestDataTable_Navigation(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 20, Renderer: func(r interface{}) string {
			return r.(testRow).Name
		}},
	}
	
	table := NewDataTable(columns)
	rows := []interface{}{
		testRow{Name: "Row1"},
		testRow{Name: "Row2"},
	}
	table.SetRows(rows)
	
	// Simulate 'j' key (down)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	table.Update(msg)
	
	selected := table.SelectedRow()
	if selected == nil {
		t.Fatal("SelectedRow returned nil")
	}
	
	if selected.(testRow).Name != "Row2" {
		t.Errorf("Expected Row2 to be selected, got: %v", selected)
	}
}

func TestDataTable_Virtualization(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 20, Renderer: func(r interface{}) string {
			return r.(testRow).Name
		}},
	}
	
	table := NewDataTable(columns)
	table.SetHeight(10) // Only 10 rows visible
	
	// Add 1000 rows
	rows := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = testRow{Name: fmt.Sprintf("Row%d", i)}
	}
	table.SetRows(rows)
	
	view := table.View()
	
	// Should not render all 1000 rows (performance check)
	lineCount := strings.Count(view, "\n")
	if lineCount > 30 { // Header + 10 visible + some buffer
		t.Errorf("Table rendered too many lines (%d), should virtualize", lineCount)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/components -v -run TestDataTable`  
Expected: FAIL with "undefined: NewDataTable"

**Step 3: Write minimal implementation**

Create `internal/components/datatable.go`:

```go
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/augury-node-tui/internal/styles"
)

// Alignment for table columns
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignRight
	AlignCenter
)

// Column defines table column properties
type Column struct {
	Header   string
	Width    int       // -1 for flex
	Sortable bool
	Align    Alignment
	Renderer func(row interface{}) string
}

// DataTable is a sortable, navigable table component
type DataTable struct {
	columns     []Column
	rows        []interface{}
	selectedIdx int
	sortColumn  string
	sortAsc     bool
	width       int
	height      int
	scrollOff   int // Scroll offset for virtualization
}

// NewDataTable creates a new table with given columns
func NewDataTable(columns []Column) *DataTable {
	return &DataTable{
		columns:     columns,
		rows:        []interface{}{},
		selectedIdx: 0,
		sortColumn:  "",
		sortAsc:     true,
		width:       80,
		height:      20,
		scrollOff:   0,
	}
}

// SetRows updates table data
func (t *DataTable) SetRows(rows []interface{}) {
	t.rows = rows
	if t.selectedIdx >= len(rows) {
		t.selectedIdx = len(rows) - 1
	}
	if t.selectedIdx < 0 {
		t.selectedIdx = 0
	}
}

// SetHeight sets visible row count
func (t *DataTable) SetHeight(height int) {
	t.height = height
}

// Update handles key messages for navigation
func (t *DataTable) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if t.selectedIdx < len(t.rows)-1 {
				t.selectedIdx++
				t.adjustScroll()
			}
		case "k", "up":
			if t.selectedIdx > 0 {
				t.selectedIdx--
				t.adjustScroll()
			}
		case "g":
			t.selectedIdx = 0
			t.scrollOff = 0
		case "G":
			t.selectedIdx = len(t.rows) - 1
			t.adjustScroll()
		}
	}
	return nil
}

// adjustScroll updates scroll offset for viewport tracking
func (t *DataTable) adjustScroll() {
	visibleRows := t.height - 2 // Header + border
	if t.selectedIdx < t.scrollOff {
		t.scrollOff = t.selectedIdx
	}
	if t.selectedIdx >= t.scrollOff+visibleRows {
		t.scrollOff = t.selectedIdx - visibleRows + 1
	}
}

// View renders the table
func (t *DataTable) View() string {
	palette := styles.DefaultPalette()
	typo := styles.DefaultTypography()
	
	var result strings.Builder
	
	// Render header
	var headerCells []string
	for _, col := range t.columns {
		headerStyle := typo.Section
		cell := headerStyle.Render(truncate(col.Header, col.Width))
		headerCells = append(headerCells, cell)
	}
	result.WriteString(strings.Join(headerCells, " │ ") + "\n")
	result.WriteString(strings.Repeat("─", t.width) + "\n")
	
	// Render visible rows (virtualization)
	visibleRows := t.height - 2
	endIdx := t.scrollOff + visibleRows
	if endIdx > len(t.rows) {
		endIdx = len(t.rows)
	}
	
	for i := t.scrollOff; i < endIdx; i++ {
		row := t.rows[i]
		var cells []string
		
		for _, col := range t.columns {
			content := col.Renderer(row)
			cell := truncate(content, col.Width)
			
			// Highlight selected row
			if i == t.selectedIdx {
				highlightStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(palette.AccentMauve)).
					Bold(true)
				cell = highlightStyle.Render(cell)
			}
			
			cells = append(cells, cell)
		}
		
		result.WriteString(strings.Join(cells, " │ ") + "\n")
	}
	
	return result.String()
}

// SelectedRow returns currently selected row data
func (t *DataTable) SelectedRow() interface{} {
	if t.selectedIdx >= 0 && t.selectedIdx < len(t.rows) {
		return t.rows[t.selectedIdx]
	}
	return nil
}

// truncate limits string to width, adding ellipsis if needed
func truncate(s string, width int) string {
	if len(s) <= width {
		return s + strings.Repeat(" ", width-len(s))
	}
	return s[:width-1] + "…"
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/components -v -run TestDataTable`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/components/
git commit -m "feat: add DataTable component with navigation and virtualization"
```

---

### Task 9: LogViewer Component (Foundation)

**Files:**
- Create: `internal/components/logviewer.go`
- Create: `internal/components/logviewer_test.go`
- Create: `internal/logs/parser.go`
- Create: `internal/logs/parser_test.go`

**Step 1: Write the failing test for parser**

Create `internal/logs/parser_test.go`:

```go
package logs

import (
	"testing"
)

func TestParseErrors_NixError(t *testing.T) {
	content := `Building packages...
error: experimental Nix feature 'nix-command' is disabled
Build failed`
	
	errors := ParseErrors(content)
	
	if len(errors) == 0 {
		t.Fatal("Should detect Nix error")
	}
	
	err := errors[0]
	if err.Level != ErrorLevelError {
		t.Error("Should classify as Error level")
	}
	if err.LineNumber <= 0 {
		t.Error("Should capture line number")
	}
	if err.Suggestion == "" {
		t.Error("Should provide suggestion")
	}
}

func TestParseErrors_GCCError(t *testing.T) {
	content := `Compiling foo.c
foo.c:42: undefined reference to 'bar'
Compilation failed`
	
	errors := ParseErrors(content)
	
	if len(errors) == 0 {
		t.Fatal("Should detect GCC error")
	}
	
	if !strings.Contains(errors[0].LineText, "undefined reference") {
		t.Error("Should capture error text")
	}
}

func TestParseErrors_NoErrors(t *testing.T) {
	content := `All builds successful
No errors found`
	
	errors := ParseErrors(content)
	
	if len(errors) != 0 {
		t.Errorf("Should not detect errors in clean output, got: %d", len(errors))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/logs -v`  
Expected: FAIL with "no such file or directory"

**Step 3: Implement error parser**

Create `internal/logs/parser.go`:

```go
package logs

import (
	"regexp"
	"strings"
)

// ErrorLevel indicates severity
type ErrorLevel int

const (
	ErrorLevelCritical ErrorLevel = iota
	ErrorLevelError
	ErrorLevelWarning
)

// ErrorLocation represents a detected error in logs
type ErrorLocation struct {
	LineNumber int
	LineText   string
	Level      ErrorLevel
	Context    []string // Lines before/after
	Suggestion string
}

var errorPatterns = []struct {
	regex      *regexp.Regexp
	level      ErrorLevel
	suggestion string
}{
	{
		regex:      regexp.MustCompile(`error: experimental Nix feature.*disabled`),
		level:      ErrorLevelError,
		suggestion: "Enable nix-command and flakes in ~/.config/nix/nix.conf",
	},
	{
		regex:      regexp.MustCompile(`ERROR: Task.*failed`),
		level:      ErrorLevelError,
		suggestion: "Check tmp/work/<package>/temp/log.do_* for details",
	},
	{
		regex:      regexp.MustCompile(`undefined reference to`),
		level:      ErrorLevelError,
		suggestion: "Missing library in DEPENDS or LDFLAGS",
	},
	{
		regex:      regexp.MustCompile(`error:|ERROR:`),
		level:      ErrorLevelError,
		suggestion: "",
	},
	{
		regex:      regexp.MustCompile(`warning:|WARNING:`),
		level:      ErrorLevelWarning,
		suggestion: "",
	},
}

// ParseErrors scans content for error patterns
func ParseErrors(content string) []ErrorLocation {
	lines := strings.Split(content, "\n")
	var errors []ErrorLocation
	
	for i, line := range lines {
		for _, pattern := range errorPatterns {
			if pattern.regex.MatchString(line) {
				// Extract context (5 lines before/after)
				contextStart := i - 5
				if contextStart < 0 {
					contextStart = 0
				}
				contextEnd := i + 6
				if contextEnd > len(lines) {
					contextEnd = len(lines)
				}
				
				errors = append(errors, ErrorLocation{
					LineNumber: i + 1, // 1-indexed
					LineText:   line,
					Level:      pattern.level,
					Context:    lines[contextStart:contextEnd],
					Suggestion: pattern.suggestion,
				})
				break // Only match first pattern per line
			}
		}
	}
	
	return errors
}
```

**Step 4: Run parser tests**

Run: `go test ./internal/logs -v`  
Expected: PASS

**Step 5: Write LogViewer component test**

Create `internal/components/logviewer_test.go`:

```go
package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogViewer_Creation(t *testing.T) {
	content := "Line 1\nLine 2\nerror: something bad\nLine 4"
	viewer := NewLogViewer(content)
	
	if viewer == nil {
		t.Fatal("NewLogViewer returned nil")
	}
	
	view := viewer.View()
	if !strings.Contains(view, "Line 1") {
		t.Error("LogViewer should render content")
	}
}

func TestLogViewer_JumpToFirstError(t *testing.T) {
	content := "Line 1\nLine 2\nerror: problem here\nLine 4"
	viewer := NewLogViewer(content)
	
	// Jump to first error
	cmd := viewer.JumpToFirstError()
	if cmd == nil {
		t.Error("JumpToFirstError should return command")
	}
	
	// Error should be detected
	if len(viewer.Errors()) == 0 {
		t.Error("Should detect error in content")
	}
}

func TestLogViewer_Navigation(t *testing.T) {
	content := strings.Repeat("Line\n", 100)
	viewer := NewLogViewer(content)
	viewer.SetHeight(20)
	
	// Simulate scroll down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	viewer.Update(msg)
	
	view := viewer.View()
	if view == "" {
		t.Error("LogViewer should render after navigation")
	}
}
```

**Step 6: Implement LogViewer component**

Create `internal/components/logviewer.go`:

```go
package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/augury-node-tui/internal/logs"
	"github.com/augury-node-tui/internal/styles"
)

// LogViewer displays scrollable logs with error navigation
type LogViewer struct {
	content    string
	errors     []logs.ErrorLocation
	currentErr int
	viewport   viewport.Model
}

// NewLogViewer creates a new log viewer
func NewLogViewer(content string) *LogViewer {
	vp := viewport.New(80, 20)
	vp.SetContent(content)
	
	errors := logs.ParseErrors(content)
	
	return &LogViewer{
		content:    content,
		errors:     errors,
		currentErr: -1,
		viewport:   vp,
	}
}

// SetHeight updates viewport height
func (v *LogViewer) SetHeight(height int) {
	v.viewport.Height = height
}

// JumpToFirstError moves viewport to first error
func (v *LogViewer) JumpToFirstError() tea.Cmd {
	if len(v.errors) == 0 {
		return nil
	}
	
	v.currentErr = 0
	err := v.errors[0]
	
	// Calculate viewport position (line number to Y offset)
	v.viewport.GotoTop()
	for i := 0; i < err.LineNumber-1 && i < 10; i++ {
		v.viewport.LineDown(1)
	}
	
	return nil
}

// NextError jumps to next error
func (v *LogViewer) NextError() tea.Cmd {
	if len(v.errors) == 0 {
		return nil
	}
	
	v.currentErr = (v.currentErr + 1) % len(v.errors)
	err := v.errors[v.currentErr]
	
	v.viewport.GotoTop()
	for i := 0; i < err.LineNumber-1; i++ {
		v.viewport.LineDown(1)
	}
	
	return nil
}

// PrevError jumps to previous error
func (v *LogViewer) PrevError() tea.Cmd {
	if len(v.errors) == 0 {
		return nil
	}
	
	v.currentErr--
	if v.currentErr < 0 {
		v.currentErr = len(v.errors) - 1
	}
	
	err := v.errors[v.currentErr]
	v.viewport.GotoTop()
	for i := 0; i < err.LineNumber-1; i++ {
		v.viewport.LineDown(1)
	}
	
	return nil
}

// Update handles key messages
func (v *LogViewer) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			return v.JumpToFirstError()
		case "n":
			return v.NextError()
		case "N":
			return v.PrevError()
		default:
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return cmd
		}
	}
	
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

// View renders the log viewer
func (v *LogViewer) View() string {
	palette := styles.DefaultPalette()
	
	// Highlight errors in viewport content
	content := v.viewport.View()
	
	// Add status bar if errors present
	var statusBar string
	if len(v.errors) > 0 && v.currentErr >= 0 {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Error))
		statusBar = errStyle.Render(
			fmt.Sprintf(" [Error %d/%d] ", v.currentErr+1, len(v.errors)),
		)
	}
	
	return content + "\n" + statusBar
}

// Errors returns detected error locations
func (v *LogViewer) Errors() []logs.ErrorLocation {
	return v.errors
}
```

**Step 7: Run all tests**

Run: `go test ./internal/components ./internal/logs -v`  
Expected: PASS

**Step 8: Commit**

```bash
git add internal/components/ internal/logs/
git commit -m "feat: add LogViewer component with error parsing and navigation"
```

---

## Wave 2 Continuation Tasks

Due to space constraints, I'll summarize remaining Wave 2 and Wave 3 tasks. Each follows same TDD pattern:

### Task 10: Refactor Validations Screen
- Create tests expecting DataTable usage
- Modify `internal/validations/model.go` to use DataTable component
- Commit

### Task 11: Refactor Hydration Screen
- Create tests expecting DataTable + ProgressBar
- Modify `internal/hydration/model.go` 
- Commit

## Wave 3: Complex Components & Screens

### Task 12: ParallelTracker Component
- TDD cycle: test → implement → commit
- Files: `internal/components/paralleltracker.go` + tests

### Task 13: MetricsBar Component
- TDD cycle with gopsutil integration
- Files: `internal/components/metricsbar.go` + tests
- Add dependency: `github.com/shirou/gopsutil/v3`

### Task 14: CommandDisplay Component
- TDD cycle for command transparency pattern
- Files: `internal/components/commanddisplay.go` + tests

### Task 15: Refactor Caches Screen
- Migrate to DataTable + MetricsBar
- Files: `internal/caches/model.go`

### Task 16: Refactor Build Screen
- Three-pane layout: ParallelTracker + MetricsBar + LogViewer
- Files: `internal/build/model.go`

### Task 17: Refactor Home Screen
- Card (repo status) + DataTable (platforms) + MetricsBar
- Files: `internal/home/model.go`

## Dependencies

Run after completing Wave 1:

```bash
go get github.com/mattn/go-runewidth
go get github.com/shirou/gopsutil/v3
go mod tidy
```

## Validation

After each wave:
1. Run all tests: `go test ./... -v`
2. Build: `go build -o augury-node-tui ./cmd/augury-node-tui`
3. Manual smoke test: Run TUI, navigate all screens
4. Verify no regressions in existing functionality

## Success Criteria

All tasks complete when:
- ✅ All unit tests pass
- ✅ All screens use component system
- ✅ Visual consistency across screens (Catppuccin Mocha)
- ✅ No functional regressions
- ✅ Ready for Phase 2/3 feature additions

## Execution Notes

- Follow strict TDD: Test → Fail → Implement → Pass → Commit
- One component at a time, validate before moving forward
- Keep commits atomic and focused
- Run full test suite after each wave
