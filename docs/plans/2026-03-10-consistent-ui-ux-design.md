# Consistent UI/UX Design for augury-node-tui

**Date:** 2026-03-10  
**Status:** Approved  
**Goal:** Make all TUI screens consistent with the setup wizard's clean, focused aesthetic

## Overview

The setup wizard demonstrates the ideal UI/UX for the TUI: persistent top and bottom panels, clean bordered boxes, minimal clutter, and excellent visual hierarchy. This design extends that pattern to all main application screens.

## Design Principles

1. **Universal 3-panel layout**: Every screen uses top bar → content → bottom help
2. **Strategic boxes**: Bordered boxes for status/info, no borders for interactive areas
3. **Compact context**: Single-line top bar with breadcrumb + abbreviated context
4. **Hybrid help panel**: Context-aware actions (left) + universal nav (right)
5. **Enhanced visibility**: Larger checkboxes, full row highlighting for tables
6. **Consistent styling**: Centralized layout logic, shared visual components

## Architecture

### 3-Panel Layout Structure

```
┌─────────────────────────────────────────────┐
│ 🚀 Home → Build  •  master  •  3 selected   │ ← Top Bar (1 line, adaptive)
├─────────────────────────────────────────────┤
│                                             │
│  ╭──────────────────────╮                  │
│  │ 📁 Repository        │  ← Bordered Box  │
│  │ Status info here     │     (strategic)  │
│  ╰──────────────────────╯                  │
│                                             │
│  🎯 Platforms                               │ ← Interactive Area
│  ┌──────────────────────────────────────┐  │    (no border)
│  │ [ · ]  Platform  State   Path        │  │
│  │ [●]  ccimx8x   built    result/...   │  │ ← highlighted row (full width)
│  │ [ · ]  ccimx93   built    result/... │  │
│  └──────────────────────────────────────┘  │
│                                             │
├─────────────────────────────────────────────┤
│ space select  b build  •  j/k nav  q quit   │ ← Bottom Help (hybrid)
└─────────────────────────────────────────────┘
```

**Panel responsibilities:**
- **Top bar**: Breadcrumb navigation + abbreviated context (branch, selections, state)
- **Content area**: Screen-specific content with strategic box usage
- **Bottom help**: Left side = context-aware actions, right side = universal navigation

## Components

### 1. ScreenLayout Component

New reusable component that enforces the 3-panel structure:

```go
type ScreenLayout struct {
    // Top bar configuration
    Breadcrumb    []string           // ["Home", "Build"]
    Context       string             // "master  •  3 selected"
    
    // Content area
    Content       string             // The main screen content (rendered by child)
    
    // Bottom help configuration
    ActionKeys    []KeyBinding       // Left side: context-aware actions
    NavKeys       []KeyBinding       // Right side: universal nav
    
    // Dimensions
    Width         int
    Height        int
}

type KeyBinding struct {
    Key   string  // "space"
    Label string  // "select"
}

func (s *ScreenLayout) Render() string
```

**Usage pattern:**
```go
func (m *Model) View() string {
    layout := ScreenLayout{
        Breadcrumb: []string{"Home"},
        Context:    fmt.Sprintf("%s  •  %d selected", m.Status.Branch, m.countSelected()),
        Content:    m.renderContent(),
        ActionKeys: m.buildActionKeys(), // Dynamic based on state
        NavKeys:    []KeyBinding{
            {Key: "j/k", Label: "navigate"},
            {Key: "q", Label: "quit"},
        },
        Width:  m.Width,
        Height: m.Height,
    }
    return layout.Render()
}
```

### 2. Enhanced DataTable

Updates to `internal/components/DataTable`:

**Full row highlighting:**
- Track cursor position (already exists)
- When rendering, apply `styles.RowHighlight` background to entire row if cursor is on it
- Use `lipgloss.NewStyle().Background()` to span full width

**Larger checkboxes:**
- Current: `☐` / `☑` (width 3)
- New: `[ · ]` / `[●]` (width 5)
- Checkbox column width increased from 3 to 5

**Checkbox rendering:**
```go
func (m *Model) renderCheckbox(row interface{}) string {
    e := row.(PlatformEntry)
    if e.Selected {
        return styles.CheckboxSelected.Render("[●]")  // Bold, colored
    }
    return styles.CheckboxUnselected.Render("[ · ]") // Dim
}
```

**Visual behavior:**
- Navigation (j/k): Entire row gets highlighted background
- Selection (space): Toggles checkbox on highlighted row
- Multiple rows can have checked boxes, only one row is highlighted at a time

### 3. Style System Updates

New styles in `internal/styles/`:

```go
// Checkbox styles
CheckboxUnselected = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))  // Dim gray
CheckboxSelected   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)  // Green, bold

// Row highlight style
RowHighlight = lipgloss.NewStyle().
    Background(lipgloss.Color("236")).  // Subtle gray background
    Foreground(lipgloss.Color("255"))   // Bright text

// Separator line style
Separator = lipgloss.NewStyle().
    Foreground(lipgloss.Color("240"))   // Dim gray for separator lines

// Top bar styles
TopBar = lipgloss.NewStyle().
    Foreground(lipgloss.Color("39")).   // Blue breadcrumb
    Bold(true)

TopBarContext = lipgloss.NewStyle().
    Foreground(lipgloss.Color("246"))   // Dim context info
```

## Screen Adaptations

Each main screen adopts the `ScreenLayout` with appropriate content:

### Home Screen
- **Breadcrumb**: `["Home"]`
- **Context**: `master  •  3 selected`
- **Bordered boxes**: Repository status (existing Card component)
- **No borders**: Platform table (interactive)
- **Action keys**: `space select`, `b build`, `h hydrate` (dynamic based on selection)
- **Nav keys**: `j/k navigate`, `q quit`

### Build Screen
- **Breadcrumb**: `["Home", "Build"]`
- **Context**: `master  •  building 3 platforms`
- **Bordered boxes**: Build status/progress for each platform
- **No borders**: Build output log (scrollable)
- **Action keys**: `c cancel` (when building), `r retry` (when failed)
- **Nav keys**: `esc back`, `q quit`

### CI Dashboard
- **Breadcrumb**: `["Home", "Pipeline"]`
- **Context**: `master  •  workflow running`
- **Bordered boxes**: Pipeline status, workflow summary
- **No borders**: Jobs table (similar to platform table)
- **Action keys**: `enter view log`, `r refresh`
- **Nav keys**: `j/k navigate`, `esc back`, `q quit`

### Caches / Validations / Hints
- **Breadcrumb**: `["Home", "Caches"]` / `["Home", "Validations"]` / `["Home", "Hints"]`
- **Context**: Relevant summary (counts, states)
- **Bordered boxes**: Summary cards, status info
- **No borders**: Data tables/lists
- **Action keys**: Screen-specific operations
- **Nav keys**: `j/k navigate`, `esc back`, `q quit`

### Hydration Screen
- **Breadcrumb**: `["Home", "Hydrate"]`
- **Context**: `master  •  hydrating 3 platforms`
- **Bordered boxes**: Hydration status/progress
- **No borders**: Output logs
- **Action keys**: `c cancel`, `r retry`
- **Nav keys**: `esc back`, `q quit`

## Strategic Box Usage

**Use bordered boxes for:**
- Repository status
- Build/hydration status cards
- CI pipeline summary
- Metrics and aggregates
- Warning/info messages

**Don't use boxes for:**
- Data tables (platforms, jobs, caches)
- Logs and output streams
- Interactive selection lists
- Navigation menus

**Rationale:** Boxes signal "reference information" while unboxed areas signal "interactive workspace". This creates clear visual hierarchy.

## Benefits

1. **Consistency**: All screens feel like part of the same app
2. **Predictability**: Users know where to find navigation, context, and actions
3. **Maintainability**: Layout logic centralized in `ScreenLayout`
4. **Scalability**: New screens automatically get the right structure
5. **Visual clarity**: Better hierarchy through strategic box usage
6. **Accessibility**: Larger checkboxes, full row highlighting

## Implementation Approach

**Layout Framework First:**
1. Create `ScreenLayout` component
2. Add new styles to `internal/styles/`
3. Enhance `DataTable` with row highlighting and larger checkboxes
4. Migrate all 8 main screens to use `ScreenLayout`
5. Test each screen for consistency

**Migration order:**
1. Home screen (most visible, proves the pattern)
2. Build screen (complex state, tests action key dynamics)
3. CI dashboard (table-heavy, validates DataTable enhancements)
4. Remaining screens (caches, validations, hints, hydration)

## Success Criteria

- [ ] All main screens use `ScreenLayout` component
- [ ] Top bar shows breadcrumb + context on every screen
- [ ] Bottom help panel is persistent and context-aware
- [ ] DataTable supports full row highlighting
- [ ] Checkboxes are `[ · ]` / `[●]` format (width 5)
- [ ] Strategic boxes used consistently (status in boxes, tables without)
- [ ] Visual consistency matches setup wizard aesthetic
- [ ] No inline key help (all in bottom panel)

## Future Enhancements

- Theme customization (color schemes)
- Adaptive layouts for very narrow terminals
- Mouse support for row selection
- Keyboard shortcuts legend (modal on `?`)
