# Consistent UI/UX Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement universal 3-panel layout and consistent styling across all TUI screens to match the setup wizard aesthetic.

**Architecture:** Create reusable `ScreenLayout` component that enforces top bar + content + bottom help structure. Enhance `DataTable` with full row highlighting and larger checkboxes. Migrate all 8 main screens to use the new components.

**Tech Stack:** Go, Bubbletea, Lipgloss, existing component system

---

## Task 1: Add New Styles

**Files:**
- Modify: `internal/styles/theme.go`

**Step 1: Add checkbox styles**

Add to `internal/styles/theme.go`:

```go
// Checkbox styles
var (
    CheckboxUnselected = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    CheckboxSelected   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)
```

**Step 2: Add row highlight style**

Add to `internal/styles/theme.go`:

```go
// Row highlight style
var RowHighlight = lipgloss.NewStyle().
    Background(lipgloss.Color("236")).
    Foreground(lipgloss.Color("255"))
```

**Step 3: Add separator style**

Add to `internal/styles/theme.go`:

```go
// Separator style
var Separator = lipgloss.NewStyle().
    Foreground(lipgloss.Color("240"))
```

**Step 4: Add top bar styles**

Add to `internal/styles/theme.go`:

```go
// Top bar styles
var (
    TopBar = lipgloss.NewStyle().
        Foreground(lipgloss.Color("39")).
        Bold(true)
    
    TopBarContext = lipgloss.NewStyle().
        Foreground(lipgloss.Color("246"))
)
```

**Step 5: Build to verify no errors**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go build ./...`
Expected: SUCCESS (no compilation errors)

**Step 6: Commit**

```bash
git add internal/styles/theme.go
git commit -m "feat: add styles for consistent UI/UX

- Checkbox styles (selected/unselected)
- Row highlight style for DataTable
- Separator style for panel dividers
- Top bar styles (breadcrumb and context)"
```

---

## Task 2: Create ScreenLayout Component

**Files:**
- Create: `internal/components/screen_layout.go`
- Create: `internal/components/screen_layout_test.go`

**Step 1: Write failing test**

Create `internal/components/screen_layout_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/components -run TestScreenLayout -v`
Expected: FAIL with "undefined: ScreenLayout"

**Step 3: Write minimal implementation**

Create `internal/components/screen_layout.go`:

```go
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
    Breadcrumb    []string
    Context       string
    Content       string
    ActionKeys    []KeyBinding
    NavKeys       []KeyBinding
    Width         int
    Height        int
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/components -run TestScreenLayout -v`
Expected: PASS

**Step 5: Build to verify no errors**

Run: `go build ./...`
Expected: SUCCESS

**Step 6: Commit**

```bash
git add internal/components/screen_layout.go internal/components/screen_layout_test.go
git commit -m "feat: add ScreenLayout component for consistent 3-panel UI

Provides reusable component enforcing top bar → content → bottom help
structure. Supports breadcrumb navigation, context display, and hybrid
help panel (action keys + nav keys)."
```

---

## Task 3: Enhance DataTable with Full Row Highlighting

**Files:**
- Modify: `internal/components/datatable.go`
- Modify: `internal/components/datatable_test.go`

**Step 1: Write failing test**

Add to `internal/components/datatable_test.go`:

```go
func TestDataTable_RowHighlight(t *testing.T) {
    columns := []Column{
        {Header: "Name", Width: 10, Sortable: false, Renderer: func(row interface{}) string {
            return row.(string)
        }},
    }
    
    table := NewDataTable(columns)
    table.SetRows([]interface{}{"row1", "row2", "row3"})
    table.SetWidth(80)
    table.cursorRow = 1 // Highlight second row
    
    output := table.View()
    
    // Check that highlighting is applied (would need to inspect styling)
    // For now just verify it doesn't panic
    if output == "" {
        t.Error("Expected non-empty output")
    }
}
```

**Step 2: Run test to verify it doesn't panic**

Run: `go test ./internal/components -run TestDataTable_RowHighlight -v`
Expected: PASS (cursorRow field exists in current implementation)

**Step 3: Update row rendering to apply highlight**

In `internal/components/datatable.go`, find the `View()` method and update the row rendering logic:

```go
// In View() method, when rendering data rows:
for i, row := range dt.rows {
    var cells []string
    for _, col := range dt.columns {
        content := col.Renderer(row)
        if col.Width == -1 {
            cells = append(cells, content)
        } else {
            cells = append(cells, truncate(content, col.Width))
        }
    }
    
    rowStr := strings.Join(cells, "  ")
    
    // Apply highlight if this is the cursor row
    if i == dt.cursorRow {
        rowStr = styles.RowHighlight.Render(rowStr)
    }
    
    lines = append(lines, rowStr)
}
```

**Step 4: Build and manually test**

Run: `go build ./...`
Expected: SUCCESS

**Step 5: Commit**

```bash
git add internal/components/datatable.go internal/components/datatable_test.go
git commit -m "feat: add full row highlighting to DataTable

Apply RowHighlight style to entire row when cursor is on it, making
navigation more visible."
```

---

## Task 4: Update DataTable Checkbox Rendering

**Files:**
- Modify: `internal/home/model.go`

**Step 1: Update checkbox column width**

In `internal/home/model.go`, find `initPlatformTable()` and update the checkbox column:

```go
columns := []components.Column{
    {Header: "☐", Width: 5, Sortable: false, Renderer: m.renderCheckbox}, // Changed from 3 to 5
    {Header: "Platform", Width: 20, Sortable: true, Renderer: m.renderPlatformID},
    {Header: "State", Width: 12, Sortable: true, Renderer: m.renderState},
    {Header: "Path", Width: -1, Sortable: true, Renderer: m.renderOutputPath},
}
```

**Step 2: Update checkbox renderer**

In `internal/home/model.go`, update `renderCheckbox()`:

```go
func (m *Model) renderCheckbox(row interface{}) string {
    e := row.(PlatformEntry)
    if e.Selected {
        return styles.CheckboxSelected.Render("[●]")
    }
    return styles.CheckboxUnselected.Render("[ · ]")
}
```

**Step 3: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/home/model.go
git commit -m "feat: update Home screen to use larger checkboxes

Changed from ☐/☑ (width 3) to [ · ]/[●] (width 5) for better visibility."
```

---

## Task 5: Migrate Home Screen to ScreenLayout

**Files:**
- Modify: `internal/home/model.go`

**Step 1: Add helper method to build action keys**

Add to `internal/home/model.go`:

```go
func (m *Model) buildActionKeys() []components.KeyBinding {
    var keys []components.KeyBinding
    
    // Dynamic keys based on selection
    keys = append(keys, components.KeyBinding{Key: "space", Label: "select"})
    
    selectedCount := m.countSelected()
    if selectedCount > 0 {
        keys = append(keys, components.KeyBinding{Key: "b", Label: "build"})
        keys = append(keys, components.KeyBinding{Key: "h", Label: "hydrate"})
    }
    
    keys = append(keys, components.KeyBinding{Key: "c", Label: "caches"})
    keys = append(keys, components.KeyBinding{Key: "v", Label: "validations"})
    keys = append(keys, components.KeyBinding{Key: "o", Label: "hints"})
    keys = append(keys, components.KeyBinding{Key: "p", Label: "pipeline"})
    
    return keys
}

func (m *Model) countSelected() int {
    count := 0
    for _, selected := range m.Selected {
        if selected {
            count++
        }
    }
    return count
}
```

**Step 2: Update View() to use ScreenLayout**

Replace the `View()` method in `internal/home/model.go`:

```go
func (m *Model) View() string {
    layout := components.ScreenLayout{
        Breadcrumb: []string{"🚀 Home"},
        Context:    m.buildContext(),
        Content:    m.renderContent(),
        ActionKeys: m.buildActionKeys(),
        NavKeys: []components.KeyBinding{
            {Key: "j/k", Label: "navigate"},
            {Key: "q", Label: "quit"},
        },
        Width:  m.Width,
        Height: m.Height,
    }
    return layout.Render()
}

func (m *Model) buildContext() string {
    parts := []string{m.Status.Branch}
    
    selectedCount := m.countSelected()
    if selectedCount > 0 {
        parts = append(parts, fmt.Sprintf("%d selected", selectedCount))
    }
    
    return strings.Join(parts, "  •  ")
}

func (m *Model) renderContent() string {
    var sections []string
    
    // Repo status card (bordered box - already using Card component)
    repoCard := primitives.Card{
        Title:   "📁 Repository",
        Content: m.renderRepoStatus(),
        Style:   primitives.CardNormal,
    }
    width := m.Width
    if width <= 0 {
        width = 80
    }
    sections = append(sections, repoCard.Render(width))
    
    // Platform table section (no border)
    platformHeader := styles.Header.Render("🎯 Platforms")
    hint := styles.Dim.Render(" (j/k: navigate • space: toggle)")
    if m.DeveloperDownloads == nil {
        hint += "  " + styles.Warning.Render("⚠ developer-downloads unavailable")
    }
    platformSection := platformHeader + hint + "\n" + m.platformTable.View()
    sections = append(sections, platformSection)
    
    return strings.Join(sections, "\n\n")
}
```

**Step 3: Remove old renderKeyHelp method**

Delete the `renderKeyHelp()` method from `internal/home/model.go` (no longer needed).

**Step 4: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 5: Test manually**

Run: `./augury-node-tui`
Expected: Home screen displays with new 3-panel layout, larger checkboxes, row highlighting

**Step 6: Commit**

```bash
git add internal/home/model.go
git commit -m "feat: migrate Home screen to ScreenLayout component

Uses new consistent 3-panel layout with:
- Top bar: breadcrumb + context (branch, selection count)
- Content: repo card (bordered) + platform table (no border)
- Bottom help: context-aware actions + universal nav"
```

---

## Task 6: Migrate Build Screen to ScreenLayout

**Files:**
- Modify: `internal/build/model.go`

**Step 1: Add helper methods**

Add to `internal/build/model.go`:

```go
func (m *Model) buildActionKeys() []components.KeyBinding {
    var keys []components.KeyBinding
    
    // Dynamic based on build state
    if m.isBuilding() {
        keys = append(keys, components.KeyBinding{Key: "c", Label: "cancel"})
    } else if m.hasFailed() {
        keys = append(keys, components.KeyBinding{Key: "r", Label: "retry"})
    }
    
    return keys
}

func (m *Model) buildContext() string {
    if m.isBuilding() {
        return fmt.Sprintf("%s  •  building %d platforms", m.branch, m.platformCount)
    }
    return m.branch
}
```

**Step 2: Update View() to use ScreenLayout**

Replace the `View()` method:

```go
func (m *Model) View() string {
    layout := components.ScreenLayout{
        Breadcrumb: []string{"🚀 Home", "Build"},
        Context:    m.buildContext(),
        Content:    m.renderContent(),
        ActionKeys: m.buildActionKeys(),
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

**Step 3: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/build/model.go
git commit -m "feat: migrate Build screen to ScreenLayout

Uses consistent 3-panel layout with dynamic action keys based on
build state (cancel when building, retry when failed)."
```

---

## Task 7: Migrate CI Dashboard to ScreenLayout

**Files:**
- Modify: `internal/ci/model.go`

**Step 1: Add helper methods**

Add to `internal/ci/model.go`:

```go
func (m *Model) buildActionKeys() []components.KeyBinding {
    var keys []components.KeyBinding
    
    switch m.state {
    case stateReady:
        keys = append(keys, components.KeyBinding{Key: "enter", Label: "view log"})
        keys = append(keys, components.KeyBinding{Key: "r", Label: "refresh"})
    case stateViewingLog:
        // LogViewer has its own keybindings
    }
    
    return keys
}

func (m *Model) buildContext() string {
    parts := []string{m.branch}
    
    if m.pipeline != nil {
        parts = append(parts, "workflow running")
    }
    
    return strings.Join(parts, "  •  ")
}
```

**Step 2: Update View() for non-LogViewer states**

Update the `View()` method to use ScreenLayout when not viewing logs:

```go
func (m *Model) View() string {
    // Keep full-screen LogViewer as-is
    if m.state == stateViewingLog {
        return m.logViewer.View()
    }
    
    layout := components.ScreenLayout{
        Breadcrumb: []string{"🚀 Home", "Pipeline"},
        Context:    m.buildContext(),
        Content:    m.renderContent(),
        ActionKeys: m.buildActionKeys(),
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
```

**Step 3: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/ci/model.go
git commit -m "feat: migrate CI Dashboard to ScreenLayout

Uses consistent 3-panel layout. LogViewer remains full-screen when
viewing individual job logs."
```

---

## Task 8: Migrate Remaining Screens (Caches, Validations, Hints, Hydration)

**Files:**
- Modify: `internal/caches/model.go`
- Modify: `internal/validations/model.go`
- Modify: `internal/hints/model.go`
- Modify: `internal/hydration/model.go`

**Step 1: Migrate Caches screen**

Similar pattern to Home screen - update `View()` in `internal/caches/model.go`:

```go
func (m *Model) View() string {
    layout := components.ScreenLayout{
        Breadcrumb: []string{"🚀 Home", "Caches"},
        Context:    m.buildContext(),
        Content:    m.renderContent(),
        ActionKeys: m.buildActionKeys(),
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

**Step 2: Migrate Validations screen**

Update `internal/validations/model.go` similarly.

**Step 3: Migrate Hints screen**

Update `internal/hints/model.go` similarly.

**Step 4: Migrate Hydration screen**

Update `internal/hydration/model.go` similarly (similar to Build screen).

**Step 5: Build to verify all**

Run: `go build ./...`
Expected: SUCCESS

**Step 6: Test each screen manually**

Run: `./augury-node-tui`
Navigate to each screen (c, v, o, h) and verify layout consistency

**Step 7: Commit**

```bash
git add internal/caches/model.go internal/validations/model.go internal/hints/model.go internal/hydration/model.go
git commit -m "feat: migrate remaining screens to ScreenLayout

Migrated Caches, Validations, Hints, and Hydration screens to use
consistent 3-panel layout. All main screens now have uniform UI/UX."
```

---

## Task 9: Remove ASCII Diagram (Optional Cleanup)

**Files:**
- Modify: `internal/home/model.go`

**Step 1: Remove diagram logic from Home screen**

In `internal/home/model.go`, remove the diagram rendering from `renderContent()`:

```go
func (m *Model) renderContent() string {
    var sections []string
    
    // Remove this block:
    // if m.Width >= diagram.MinDiagramWidth {
    //     sections = append(sections, diagram.PlatformFlow(m.Platforms))
    // }
    
    // Repo status card
    repoCard := primitives.Card{
        Title:   "📁 Repository",
        Content: m.renderRepoStatus(),
        Style:   primitives.CardNormal,
    }
    // ... rest of method
}
```

**Step 2: Build and test**

Run: `go build ./... && ./augury-node-tui`
Expected: Home screen cleaner without diagram clutter

**Step 3: Commit**

```bash
git add internal/home/model.go
git commit -m "refactor: remove ASCII diagram from Home screen

Removes visual clutter to match setup wizard's clean aesthetic.
Focus on platform table and essential information only."
```

---

## Task 10: Documentation and Final Testing

**Files:**
- Create: `docs/ui-ux-migration.md`

**Step 1: Document the UI/UX system**

Create `docs/ui-ux-migration.md`:

```markdown
# UI/UX Migration Complete

All main screens now use the consistent 3-panel layout via `ScreenLayout` component.

## Components

- `ScreenLayout`: Enforces top bar → content → bottom help structure
- Enhanced `DataTable`: Full row highlighting, larger checkboxes
- Style system: Consistent colors, borders, and typography

## Screen Inventory

- ✅ Home: Platform selection with repo status
- ✅ Build: Build execution with dynamic actions
- ✅ CI Dashboard: Pipeline/jobs with log viewing
- ✅ Caches: Cache management
- ✅ Validations: Validation results
- ✅ Hints: Developer hints
- ✅ Hydration: Hydration execution

## Adding New Screens

Use `ScreenLayout` for all new screens:

```go
func (m *Model) View() string {
    layout := components.ScreenLayout{
        Breadcrumb: []string{"🚀 Home", "NewScreen"},
        Context:    "context here",
        Content:    m.renderContent(),
        ActionKeys: []components.KeyBinding{
            {Key: "key", Label: "action"},
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

## Style Guidelines

- Use bordered boxes (Card component) for status/info sections
- Keep interactive areas (tables, lists) without borders
- Top bar: compact, single line with breadcrumb + abbreviated context
- Bottom help: context-aware actions on left, universal nav on right
```

**Step 2: Final integration test**

Run: `./augury-node-tui`
Navigate through all screens systematically:
- Home (select platforms, verify checkboxes and highlighting)
- Build (verify dynamic action keys)
- Pipeline (verify table consistency)
- Caches, Validations, Hints (verify layout consistency)
- Hydration (verify similar to Build)

Expected: All screens have consistent layout, styling, and behavior

**Step 3: Run all tests**

Run: `go test ./...`
Expected: All tests PASS

**Step 4: Commit documentation**

```bash
git add docs/ui-ux-migration.md
git commit -m "docs: document completed UI/UX migration

Added migration guide and usage documentation for ScreenLayout
component and consistent styling system."
```

**Step 5: Final summary commit (optional)**

```bash
git commit --allow-empty -m "feat: complete consistent UI/UX migration

All 8 main screens now use unified 3-panel layout with:
- ScreenLayout component for structure enforcement
- Enhanced DataTable with row highlighting
- Larger checkboxes [●]/[ · ] (width 5)
- Strategic box usage (status in boxes, tables without)
- Context-aware bottom help panel
- Consistent styling throughout

Matches setup wizard aesthetic: clean, focused, no clutter."
```

---

## Success Criteria

- [ ] All 8 main screens use `ScreenLayout` component
- [ ] DataTable supports full row highlighting
- [ ] Checkboxes are `[ · ]` / `[●]` format (width 5)
- [ ] Top bar shows breadcrumb + context on every screen
- [ ] Bottom help panel is persistent and context-aware
- [ ] Strategic boxes used consistently
- [ ] No inline key help (all in bottom panel)
- [ ] All tests pass
- [ ] Manual testing confirms consistent look and feel
