# Component Design System for augury-node-tui

**Date:** 2026-03-08  
**Status:** Approved  
**Research Basis:** "The Architecture of Modern Terminal User Interfaces" (2026)

## Problem

`augury-node-tui` currently has ~8k lines with ad-hoc styling and manual layout logic scattered across screens. Adding Phase 2/3 features (cache management, log slicing, dependency graphs, build metrics) will increase complexity significantly. The research on modern TUI best practices shows that successful tools like k9s and lazygit succeed by establishing reusable component systems early.

Without a component design system:
- Visual inconsistency across screens
- Duplicated rendering logic (tables, status indicators, key hints)
- Difficult to add advanced features (graphs, metrics dashboards)
- Hard to maintain as feature set grows
- Violates research's "composition over creation" principle

## Goal

Establish a three-tier component hierarchy following research-backed TUI architecture patterns, then refactor existing screens to use it, enabling rapid development of Phase 2/3/4 features with consistent UX.

## Philosophy & Architecture

### Three-Tier Component Hierarchy

Following the research's component design principles:

**Tier 1 - Low-Level Atoms** (`internal/components/primitives/`):
- Single-purpose, stateless UI elements
- Examples: Card, StatusBadge, KeyHint, ProgressBar
- No business logic, pure rendering functions
- Highly reusable across all screens

**Tier 2 - Mid-Level Molecules** (`internal/components/`):
- Stateful aggregates that combine atoms
- Examples: DataTable, LogViewer, GraphPanel, CommandDisplay, MetricsBar
- Contain interaction logic (scrolling, selection, filtering)
- Emit standardized messages for parent screens

**Tier 3 - Screen-Level Organisms** (existing `internal/home`, `internal/build`, etc.):
- Compose molecules into full screens
- Own domain state (selected platforms, filter text)
- Handle business logic and engine communication

### Design Token System

**Current state:** Ad-hoc colors defined in `internal/styles/theme.go` with Tokyo Night pastels.

**New approach:** Centralized design tokens in `internal/styles/tokens.go`:

```go
type Palette struct {
    // Base colors (Catppuccin Mocha - research-backed for readability)
    Base      string  // #1E1E2E (background)
    Surface0  string  // #313244 (elevated surfaces)
    Overlay0  string  // #6C7086 (dimmed text)
    Text      string  // #CDD6F4 (primary text)
    
    // Semantic status colors (research's status-based mapping)
    Success   string  // #A6E3A1 (green - running/success)
    Warning   string  // #F9E2AF (yellow - pending/warning)
    Error     string  // #F38BA8 (red - error/failure)
    Info      string  // #89B4FA (blue - info/neutral)
    
    // Accent colors (categorical mapping for entity types)
    AccentPink   string  // #F5C2E7 (platforms)
    AccentMauve  string  // #CBA6F7 (builds)
    AccentPeach  string  // #FAB387 (caches)
    AccentTeal   string  // #94E2D5 (validations)
}

type Typography struct {
    Title     lipgloss.Style  // Bold, AccentPink
    Section   lipgloss.Style  // Bold, Text
    Body      lipgloss.Style  // Normal, Text
    Dim       lipgloss.Style  // Overlay0
    Highlight lipgloss.Style  // AccentMauve (selected items)
}

type Borders struct {
    Thick  lipgloss.Border
    Thin   lipgloss.Border
    None   lipgloss.Border
}
```

**Color Mapping Strategies** (from research):
- **Categorical:** Different colors for different entity types (platforms=pink, builds=mauve)
- **Sequential:** Gradient for numerical magnitude (CPU 50%=yellow, 90%=red)
- **Status-based:** Universal vocabulary (green=success, red=error, yellow=warning)

### Refactor Strategy

**Incremental Conversion** (not big-bang rewrite):

**Wave 1 - Simple Screens (validate component patterns):**
1. Success screen → `Card` + `StatusBadge`
2. Hints screen → `Card` + `KeyHint` list  
3. Splash screen → `Card` for centered content

**Wave 2 - Medium Complexity:**
4. Validations screen → `DataTable` + `StatusBadge`
5. Hydration screen → `DataTable` + `ProgressBar`

**Wave 3 - Complex Screens:**
6. Caches screen → `DataTable` with multi-column sort + `MetricsBar`
7. Build screen → `ParallelTracker` + `LogViewer` + `MetricsBar`
8. Home screen → `Card` (repo status) + `DataTable` (platforms) + `GraphPanel` preview

Each wave is tested before moving to next. Existing functionality preserved during refactor.

## Component Catalog

### Tier 1: Atoms (Primitives)

#### Card
**Purpose:** Bordered container with optional title (research's "grid-based layout" constraint)

```go
type Card struct {
    Title   string
    Content string
    Style   CardStyle  // Compact, Normal, Emphasized
}

func (c Card) Render(width int) string
```

**Visual:** Uses `tokens.Borders.Thin` and word-wraps content to fit width.

#### StatusBadge
**Purpose:** Colored status indicator (research's status-based mapping)

```go
type Status int
const (
    StatusSuccess Status = iota
    StatusError
    StatusWarning
    StatusRunning
    StatusBlocked
    StatusUnavailable
)

type StatusBadge struct {
    Label  string
    Status Status
}

func (b StatusBadge) Render() string
```

**Visual:** `✓ Success` (green), `✗ Error` (red), `⚠ Warning` (yellow), `● Running` (blue), `⊘ Blocked` (dim)

#### KeyHint
**Purpose:** Consistent key binding display (research's "in-situ documentation")

```go
type KeyHint struct {
    Key         string
    Description string
    Enabled     bool
}

func (h KeyHint) Render() string
```

**Visual:** `[b] build` (enabled), `[b] build (blocked: Nix)` (disabled, dimmed)

#### ProgressBar
**Purpose:** Percentage or fraction display (research's sequential mapping)

```go
type ProgressBar struct {
    Current int
    Total   int
    Width   int
    Label   string
}

func (p ProgressBar) Render() string
```

**Visual:** `Building: ████████░░ 82% (820/1000)`

### Tier 2: Molecules (Smart Components)

#### DataTable
**Purpose:** Sortable, navigable table (replaces manual table logic across screens)

```go
type Column struct {
    Header    string
    Width     int
    Sortable  bool
    Renderer  func(row interface{}) string
}

type DataTable struct {
    Columns     []Column
    Rows        []interface{}
    SelectedIdx int
    SortColumn  string
    SortAsc     bool
    Width       int
    Height      int
}

func NewDataTable(columns []Column) *DataTable
func (t *DataTable) SetRows(rows []interface{})
func (t *DataTable) Update(msg tea.Msg) tea.Cmd
func (t *DataTable) View() string
func (t *DataTable) SelectedRow() interface{}
```

**Behavior:**
- `j/k` or arrow keys for navigation
- `enter` to emit `TableRowSelectedMsg`
- Click column header to sort (future enhancement)
- Virtualization: only renders visible rows + 10 buffer

**Visual:** Uses `tokens.AccentMauve` for selected row, `tokens.Borders.Thin` for table borders.

#### LogViewer
**Purpose:** Scrollable log with error navigation (research's "Error Zoom" from k9s)

```go
type ErrorLocation struct {
    LineNumber int
    Context    string  // 5 lines before + error + 5 lines after
    Level      ErrorLevel
}

type LogViewer struct {
    Content     string
    Errors      []ErrorLocation
    CurrentErr  int
    Viewport    viewport.Model
}

func NewLogViewer(content string) *LogViewer
func (v *LogViewer) JumpToFirstError() tea.Cmd
func (v *LogViewer) NextError() tea.Cmd
func (v *LogViewer) PrevError() tea.Cmd
func (v *LogViewer) Update(msg tea.Msg) tea.Cmd
func (v *LogViewer) View() string
```

**Behavior:**
- `e` to jump to first error (sets viewport position)
- `n`/`N` to cycle through errors
- Error lines highlighted in red
- Context window (5 lines before/after) shown with dimmed style
- Status bar shows `[Error 2/5]`

**Parsing:** Uses `internal/logs/parser.go` with regex patterns for Nix, Bitbake, GCC errors.

#### GraphPanel
**Purpose:** ASCII-art dependency graph rendering (research's DAG visualization)

```go
type GraphNode struct {
    ID          string
    Label       string
    Status      Status
    Children    []string
    Metadata    map[string]string  // rebuild-reason, cache-state, etc.
}

type GraphPanel struct {
    Nodes       map[string]GraphNode
    RootID      string
    SelectedID  string
    Collapsed   map[string]bool  // Which nodes are collapsed
    Width       int
    Height      int
}

func NewGraphPanel(nodes map[string]GraphNode, rootID string) *GraphPanel
func (g *GraphPanel) Update(msg tea.Msg) tea.Cmd
func (g *GraphPanel) View() string
```

**Behavior:**
- `j/k` to navigate tree
- `enter` to expand/collapse node
- `w` to show "why rebuild" detail for selected node
- Unicode box-drawing characters: `├─●`, `└─●`

**Visual Example:**
```
● augury-node-tui [selected]
├─● nix-flake-env
│ ├─● rustc-1.75.0
│ └─● gcc-wrapper-13.2.0 [rebuilding]
│   └─● glibc-2.38 [cache-miss]
└─● libhalo-deps
```

#### MetricsBar
**Purpose:** Real-time resource usage (research's "Build Dashboard TUI")

```go
type MetricsBar struct {
    CPU        float64  // 0.0 to 1.0
    Memory     float64
    Disk       float64
    HotProcess string   // Most active process
    Width      int
}

func (m MetricsBar) Render() string
```

**Visual:** `CPU: ████░ 82%  MEM: ███░░ 65%  DISK: ██░░░ 48%  Hot: gcc (3 threads)`

Uses sequential color mapping (research): <50% dim, 50-80% yellow, >80% red.

#### ParallelTracker
**Purpose:** Show concurrent build lanes (research's "parallel progress tracks")

```go
type BuildLane struct {
    Platform string
    Progress float64  // 0.0 to 1.0
    Status   Status
    Current  string   // Current package/step
}

type ParallelTracker struct {
    Lanes  []BuildLane
    Width  int
    Height int
}

func (p ParallelTracker) Render() string
```

**Visual:**
```
▶ node2       ████████░░ 82%  gcc-wrapper-13.2.0
▶ moxa-uc3100 ███░░░░░░░ 32%  glibc-2.38
□ cassia-x2000 queued
```

## Visual Design System

### Color Palette (Catppuccin Mocha)

Migrating from current Tokyo Night pastels to Catppuccin Mocha (research shows warm pastels reduce eye strain):

**Current (Tokyo Night):**
- Primary: `#7AA2F7` (blue)
- Success: `#9ECE6A` (green)
- Error: `#F7768E` (red)

**New (Catppuccin Mocha):**
- All colors defined in tokens above
- Warmer, more cohesive palette
- Better contrast ratios for terminal readability

### Information Aesthetics

Following research's "data aesthetics transcends decoration":

**Categorical Color Mapping:**
- Platform entities: Pink badges (`#F5C2E7`)
- Build units: Mauve badges (`#CBA6F7`)
- Cache types: Peach badges (`#FAB387`)
- Validations: Teal badges (`#94E2D5`)

**Sequential Color Mapping:**
- Resource usage: Light (0-50%) → Yellow (50-80%) → Dark Red (80-100%)
- Build progress: Blue gradient for progress bars

**Status-Based Mapping:**
- Success/Running: Green (`#A6E3A1`)
- Error/Failed: Red (`#F38BA8`)
- Warning/Pending: Yellow (`#F9E2AF`)
- Info/Neutral: Blue (`#89B4FA`)
- Blocked: Dimmed (`#6C7086`)

### Typography & Spacing

**Hierarchy:**
- `TitleStyle` - Bold, 16pt equivalent, AccentPink (screen headers)
- `SectionStyle` - Bold, 14pt, Text (subsections)
- `BodyStyle` - Normal, 12pt, Text (content)
- `DimStyle` - Normal, 12pt, Overlay0 (secondary info)
- `HighlightStyle` - Bold, 12pt, AccentMauve (selected items)

**Spacing Tokens:**
- `SpacingCompact` - 0 lines between elements
- `SpacingNormal` - 1 line
- `SpacingSpaciousm` - 2 lines

## Interaction Patterns & Ergonomics

### Standardized Keybindings

**Global Navigation** (consistent across all screens):
- `q` - Quit/back (universal standard)
- `?` - Help overlay (context-specific keys)
- `tab` - Cycle between panes
- `h,j,k,l` - Vim navigation

**Screen Actions** (mnemonic pattern from research):
- `a` - Add/Create
- `d` - Delete (with confirmation)
- `e` - Error jump (first error in logs)
- `r` - Refresh/Reload
- `/` - Filter/Search

**Existing Action Keys** (preserved):
- `b` - Build
- `h` - Hydrate
- `c` - Caches
- `v` - Validations
- `g` - Graph (new: dependency graph screen)
- `o` - Hints (existing)

### Safety & Intentional Friction

Following research's principle that "friction should exist for destructive actions":

**Destructive Actions Require Confirmation:**
- Cache deletion → Modal: "Delete 2.3 GB of cached artifacts? Type 'yes' to confirm"
- Force rebuild → Warning: "This will invalidate sstate cache"
- Config reset → Two-step: press `d`, then press `confirm`

**Non-Destructive Actions Are Frictionless:**
- Navigation between screens: single keypress
- Viewing logs/graphs: immediate
- Refreshing data: single `r` keypress

**Confirmation Modal Pattern:**
```
┌─ Confirm Deletion ─────────────────────┐
│                                        │
│  Delete local caches for moxa-uc3100? │
│                                        │
│  This will remove:                     │
│   - buildroot dl/  (1.2 GB)           │
│   - buildroot ccache/  (450 MB)       │
│                                        │
│  Type 'yes' to confirm: ___            │
│                                        │
│  [esc] Cancel                          │
└────────────────────────────────────────┘
```

### In-Situ Documentation

Research's "manual as a failure" principle - interface provides its own documentation:

**Footer Key Hints:** Every screen shows relevant hotkeys:
```
b build  •  h hydrate  •  e first error  •  / filter  •  ? help
```

**Disabled Action Feedback:** When action unavailable, show blocking reason:
```
[b] build (blocked: Nix not ready - run setup)
```

**Confirmation Explanations:** Modals explain consequences before action.

## Component Specifications

### Primitives (Tier 1)

#### Card
```go
package primitives

type CardStyle int
const (
    CardCompact CardStyle = iota    // No padding
    CardNormal                       // Normal padding
    CardEmphasized                   // Thick border, accent color
)

type Card struct {
    Title   string
    Content string
    Style   CardStyle
}

func (c Card) Render(width int) string {
    // Word-wrap content to fit width
    // Apply border based on style
    // Return rendered string
}
```

**Usage:** Repo status box, validation results, help overlays.

#### StatusBadge
```go
package primitives

type Status int
const (
    StatusSuccess Status = iota
    StatusError
    StatusWarning
    StatusRunning
    StatusBlocked
    StatusUnavailable
)

type StatusBadge struct {
    Label  string
    Status Status
}

func (b StatusBadge) Render() string {
    // Map Status to color + icon
    // Return styled string like "✓ Success"
}
```

**Usage:** Build status, cache state, Nix readiness, validation pass/fail.

#### KeyHint
```go
package primitives

type KeyHint struct {
    Key         string
    Description string
    Enabled     bool
}

func (h KeyHint) Render() string {
    // Enabled: "[b] build" (normal)
    // Disabled: "[b] build (blocked: reason)" (dimmed)
}
```

**Usage:** Footer key lists, help overlays, confirmation modals.

#### ProgressBar
```go
package primitives

type ProgressBar struct {
    Current int
    Total   int
    Width   int
    Label   string
}

func (p ProgressBar) Render() string {
    // Calculate percentage
    // Render filled/unfilled blocks
    // Return: "Building: ████████░░ 82% (820/1000)"
}
```

**Usage:** Build progress, download progress, cache sync progress.

### Molecules (Tier 2)

#### DataTable
```go
package components

type Column struct {
    Header    string
    Width     int          // Characters, or -1 for flex
    Sortable  bool
    Align     Alignment    // Left, Right, Center
    Renderer  func(row interface{}) string
}

type DataTable struct {
    columns     []Column
    rows        []interface{}
    selectedIdx int
    sortColumn  string
    sortAsc     bool
    width       int
    height      int
    viewport    viewport.Model  // From bubbles
}

func NewDataTable(columns []Column) *DataTable
func (t *DataTable) SetRows(rows []interface{})
func (t *DataTable) Update(msg tea.Msg) tea.Cmd
func (t *DataTable) View() string
func (t *DataTable) SelectedRow() interface{}
```

**Update Behavior:**
- `j`/`k` or `↑`/`↓` - Navigate rows
- `enter` - Emit `TableRowSelectedMsg{RowIndex, RowData}`
- `1`-`9` - Sort by column N (if sortable)
- `g`/`G` - Jump to top/bottom

**Virtualization:** Only renders visible rows (height-2 for header) + 10 row buffer for smooth scrolling.

**Usage:** Platform lists, cache tables, validation results, build-unit tables.

#### LogViewer
```go
package components

type ErrorLevel int
const (
    ErrorLevelCritical ErrorLevel = iota
    ErrorLevelError
    ErrorLevelWarning
)

type ErrorLocation struct {
    LineNumber int
    LineText   string
    Level      ErrorLevel
    Context    []string  // Lines before/after
}

type LogViewer struct {
    content     string
    errors      []ErrorLocation
    currentErr  int
    viewport    viewport.Model
    filterLevel ErrorLevel  // Show only this level and above
}

func NewLogViewer(content string) *LogViewer
func (v *LogViewer) JumpToFirstError() tea.Cmd
func (v *LogViewer) NextError() tea.Cmd
func (v *LogViewer) PrevError() tea.Cmd
func (v *LogViewer) SetFilter(level ErrorLevel)
func (v *LogViewer) Update(msg tea.Msg) tea.Cmd
func (v *LogViewer) View() string
```

**Update Behavior:**
- `e` - Jump to first error
- `n`/`N` - Next/previous error
- `j`/`k` or `↑`/`↓` - Scroll normally
- `PgUp`/`PgDn` - Page up/down
- `/` - Filter by text (opens filter input)
- `1`/`2`/`3` - Filter by level (Critical/Error/Warning)

**Error Parsing:** Uses `internal/logs/parser.go` with patterns:
```go
var ErrorPatterns = []struct {
    Regex      *regexp.Regexp
    Level      ErrorLevel
    Suggestion string
}{
    {regexp.MustCompile(`error: experimental Nix feature.*disabled`), ErrorLevelError, "Enable nix-command and flakes"},
    {regexp.MustCompile(`ERROR: Task.*failed`), ErrorLevelError, "Check tmp/work/<package>/temp/log.do_*"},
    {regexp.MustCompile(`undefined reference to`), ErrorLevelError, "Missing library in DEPENDS"},
    // ... more patterns
}
```

**Visual:** Error lines in red, context dimmed, current error has cursor: `→ ERROR: ...`

**Usage:** Build logs, validation logs, any long-form output.

#### GraphPanel
```go
package components

type GraphNode struct {
    ID          string
    Label       string
    Status      Status
    Children    []string
    Metadata    map[string]string  // rebuild-reason, hash, cache-state
    Collapsed   bool
}

type GraphPanel struct {
    nodes       map[string]GraphNode
    rootID      string
    selectedID  string
    width       int
    height      int
    viewport    viewport.Model
}

func NewGraphPanel(nodes map[string]GraphNode, rootID string) *GraphPanel
func (g *GraphPanel) Update(msg tea.Msg) tea.Cmd
func (g *GraphPanel) View() string
func (g *GraphPanel) ShowWhyRebuild(nodeID string) string
```

**Update Behavior:**
- `j`/`k` - Navigate tree (depth-first traversal)
- `enter` - Expand/collapse node
- `w` - Show "why rebuild" modal for selected node
- `tab` - Switch between tree view and metadata pane

**Rendering:** Uses Unicode box-drawing:
- `├─` branch
- `└─` last branch
- `│` vertical line
- `●` node (colored by status)

**Data Source:** 
- Nix: Parse `nix show-derivation .#dev-env` JSON output
- Yocto: Parse `bitbake -g <target>` DOT output (future)

**Usage:** Dependency graph screen (`g` key), build detail view (collapsed preview).

#### CommandDisplay
```go
package components

type CommandDisplay struct {
    Command     string
    Description string
    Executing   bool
    ExitCode    *int
}

func (c CommandDisplay) Render() string
```

**Visual:** Following lazygit's "command log as pedagogical tool":
```
Running: nix develop .#dev-env --command scripts/devices/node2-build.sh
[●] Building...
```

After completion:
```
✓ nix develop .#dev-env --command scripts/devices/node2-build.sh (exit 0, 2m34s)
```

**Usage:** Shows actual commands being executed (transparency), helps users learn CLI equivalents.

#### MetricsBar
```go
package components

type Metric struct {
    Label      string
    Value      float64  // 0.0 to 1.0
    Unit       string   // "%", "GB", etc.
    Threshold  float64  // Warn threshold
}

type MetricsBar struct {
    Metrics []Metric
    Width   int
}

func (m MetricsBar) Render() string
```

**Data Collection:** Uses `gopsutil` to poll every 500ms:
- CPU: `cpu.Percent()`
- Memory: `mem.VirtualMemory()`
- Disk: `disk.Usage("/")`

**Visual:** Sequential color mapping - gradient from blue (normal) to yellow (high) to red (critical).

## Screen Refactor Examples

### Home Screen (Before → After)

**Before:** Manual layout with `lipgloss.JoinVertical`:
```go
func (m *Model) View() string {
    var sections []string
    sections = append(sections, styles.Title.Render("📁 Repository"))
    sections = append(sections, fmt.Sprintf("  Root: %s", m.Status.Root))
    // ... manual styling for each element
    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
```

**After:** Composed from components:
```go
func (m *Model) View() string {
    // Repo status card
    repoCard := primitives.Card{
        Title: "📁 Repository",
        Content: m.renderRepoStatus(),
        Style: primitives.CardNormal,
    }
    
    // Platform table
    platformTable := m.platformTable.View()
    
    // Metrics bar (if enabled)
    var metricsBar string
    if m.showMetrics {
        metricsBar = m.metrics.Render()
    }
    
    // Compose with proper spacing
    return lipgloss.JoinVertical(
        lipgloss.Left,
        repoCard.Render(m.Width),
        "",
        metricsBar,
        platformTable,
        m.renderKeyHints(),
    )
}

func (m *Model) renderRepoStatus() string {
    return fmt.Sprintf(
        "%s\n%s\n%s",
        fmt.Sprintf("Root: %s", m.Status.Root),
        primitives.StatusBadge{Label: "Nix", Status: m.nixStatus}.Render(),
        primitives.StatusBadge{Label: "Paths", Status: m.pathStatus}.Render(),
    )
}
```

**Benefits:**
- Consistent styling via tokens
- Reusable `Card` and `StatusBadge`
- Easier to add metrics bar
- Clear separation: layout logic vs rendering logic

### Build Screen (After Refactor)

**Layout:** Three-pane split (research's "Build Dashboard"):
```go
func (m *Model) View() string {
    // Calculate pane widths using exact fractions (research technique)
    leftWidth := (m.Width * 3) / 10      // 30%
    middleWidth := (m.Width * 3) / 10    // 30%
    rightWidth := m.Width - leftWidth - middleWidth - 2  // 40% (minus borders)
    
    // Left pane: Parallel tracker
    leftPane := m.parallelTracker.Render()
    
    // Middle pane: Metrics
    middlePane := lipgloss.JoinVertical(
        lipgloss.Left,
        m.metricsBar.Render(),
        "",
        m.commandDisplay.Render(),
    )
    
    // Right pane: Log viewer with error navigation
    rightPane := m.logViewer.View()
    
    return lipgloss.JoinHorizontal(
        lipgloss.Top,
        styles.Border.Width(leftWidth).Render(leftPane),
        styles.Border.Width(middleWidth).Render(middlePane),
        styles.Border.Width(rightWidth).Render(rightPane),
    )
}
```

**Features:**
- Real-time parallel build progress (research's key need)
- Live resource metrics (identify bottlenecks)
- Error zoom in log pane (`e` key)
- Command transparency (lazygit pattern)

### Caches Screen (After Refactor)

**Uses DataTable component:**
```go
func (m *Model) initCacheTable() {
    columns := []components.Column{
        {Header: "Platform", Width: 20, Sortable: true, Renderer: m.renderPlatform},
        {Header: "Type", Width: 15, Sortable: true, Renderer: m.renderCacheType},
        {Header: "Local", Width: 10, Sortable: true, Renderer: m.renderLocalState},
        {Header: "Remote", Width: 10, Sortable: true, Renderer: m.renderRemoteState},
        {Header: "Size", Width: 12, Sortable: true, Align: AlignRight, Renderer: m.renderSize},
        {Header: "Actions", Width: -1, Sortable: false, Renderer: m.renderActions},
    }
    m.cacheTable = components.NewDataTable(columns)
    m.cacheTable.SetRows(m.fetchCacheData())
}

func (m *Model) renderLocalState(row interface{}) string {
    cache := row.(CacheEntry)
    status := StatusSuccess
    if !cache.LocalPresent {
        status = StatusUnavailable
    }
    return primitives.StatusBadge{
        Label: cache.LocalState,
        Status: status,
    }.Render()
}
```

**Benefits:**
- Sorting by any column (click header or press `1`-`6`)
- Consistent status colors (categorical mapping)
- Reusable across similar table screens

## Advanced Features

### Dependency Graph Visualization

**New Screen:** `internal/graph/model.go` (accessible via `g` key from home)

**Data Flow:**
1. User selects platforms on home screen
2. Press `g` to open graph screen
3. Graph screen runs `nix show-derivation .#dev-env` for selected platforms
4. Parser (`internal/graph/nix_parser.go`) converts JSON to `GraphNode` map
5. `GraphPanel` component renders ASCII tree

**Graph Parser:**
```go
package graph

func ParseNixDerivation(jsonOutput string) (map[string]GraphNode, string, error) {
    // Parse nix show-derivation JSON
    // Build adjacency list
    // Extract rebuild reasons from inputDrvs
    // Return nodes map + root ID
}
```

**Why-Rebuild Detail:**
```go
func (g *GraphPanel) ShowWhyRebuild(nodeID string) string {
    node := g.nodes[nodeID]
    return fmt.Sprintf(
        "Rebuild Reason: %s\n\nDependency Changes:\n%s",
        node.Metadata["rebuild-reason"],
        node.Metadata["dep-changes"],
    )
}
```

For Yocto, this would call `bitbake-diffsigs` (research identifies as critical for embedded devs).

### Build Metrics Dashboard

**Real-Time Metrics Collection:**
```go
package metrics

func StartCollector(interval time.Duration) chan SystemMetrics {
    ch := make(chan SystemMetrics)
    go func() {
        ticker := time.NewTicker(interval)
        for range ticker.C {
            ch <- SystemMetrics{
                CPU:    cpu.Percent(0, false)[0],
                Memory: mem.VirtualMemory().UsedPercent,
                Disk:   disk.Usage("/").UsedPercent,
                Hot:    getHottestProcess(),
            }
        }
    }()
    return ch
}
```

**Integration with Build Screen:**
- `MetricsBar` component updates every 500ms via message
- Uses research's "overwrite-don't-clear" to avoid flicker
- Shows "Hot: gcc (3 threads)" to identify bottlenecks

**Parallel Build Tracks:**
- `ParallelTracker` component shows lanes for concurrent builds
- Each lane has progress bar (research's "parallel progress tracks")
- Color-coded by status (running=blue, success=green, failed=red)

## Performance Optimizations

### Differential Rendering

Research shows "overwrite-don't-clear" prevents flicker:

```go
package components

type FrameCache struct {
    lastFrame string
}

func (f *FrameCache) Render(newView string) string {
    if f.lastFrame == "" {
        f.lastFrame = newView
        return newView
    }
    
    // Calculate diff between frames
    // Emit ANSI cursor movements + overwrites for changed regions only
    // Much faster than clearing screen and redrawing
    
    diff := calculateDiff(f.lastFrame, newView)
    f.lastFrame = newView
    return diff
}
```

**Usage:** Applied to high-frequency updates (metrics bar, progress bars, log streaming).

### Layout with Exact Geometry

Research's "fractions module" technique to avoid off-by-one gaps:

```go
func SplitHorizontal(totalWidth, numPanes int) []int {
    widths := make([]int, numPanes)
    for i := 0; i < numPanes; i++ {
        // Exact integer division
        start := (totalWidth * i) / numPanes
        end := (totalWidth * (i + 1)) / numPanes
        widths[i] = end - start
    }
    return widths
}
```

**Usage:** Multi-pane splits (build screen's 3-pane layout).

### Unicode Width Calculation

Research requirement for proper emoji/CJK handling:

```go
import "github.com/mattn/go-runewidth"

func TruncateToWidth(s string, maxWidth int) string {
    width := 0
    for i, r := range s {
        w := runewidth.RuneWidth(r)
        if width + w > maxWidth {
            return s[:i] + "…"
        }
        width += w
    }
    return s
}
```

**Usage:** All table cells, card content, any truncation logic.

### Log Sampling

Research's "ingestion filtering" for high-volume builds:

```go
package logs

type Sampler struct {
    buffer     *ring.Ring  // Last 10MB
    sampleRate time.Duration  // 100ms
    lastSample time.Time
}

func (s *Sampler) Append(line string) bool {
    now := time.Now()
    if now.Sub(s.lastSample) < s.sampleRate {
        return false  // Skip this line
    }
    s.lastSample = now
    s.buffer.Value = line
    s.buffer = s.buffer.Next()
    return true  // Line added
}
```

**Usage:** During Yocto builds with 5000+ package compilations, sample to keep TUI responsive. Full logs always written to disk.

## Phase 2/3 Feature Integration

### Phase 2: Cache Management

**Build-Unit Cache Table** (uses `DataTable` component):
- Columns: unit, fingerprint, local, remote, last action, availability
- Actions: `B` build, `R` pull, `D` delete (with confirmation modal using `Card`)
- Status badges: cached (green), missing (red), syncing (blue)

**Platform Cache Table** (uses `DataTable` component):
- Rows: buildroot dl/ccache, yocto downloads/sstate, go/cargo caches
- Actions: `P` pull, `U` push, `X` clean (confirmation required)
- Disk usage shown with `MetricsBar`

### Phase 3: Advanced Features

**Log Slicing** (uses `LogViewer` component):
- Tabs: "Full Log" / "First Error"
- `tab` to switch, `e` to jump, `j/k` to scroll
- Error count in status bar

**Developer-Downloads Awareness** (uses `StatusBadge`):
- Per-platform badges: built (green), hydrated (blue), missing (red), unavailable (dim)
- Shown in home screen platform list and post-build summary

**Mandatory Nix Gating** (uses disabled action states):
- Home screen shows `Nix: ready` or `Nix: not ready` with `StatusBadge`
- Build/hydrate keys disabled when Nix blocked
- Key hints show: `[b] build (blocked: Nix not ready)`

### Phase 4: Dependency Graph

**New Screen** (`internal/graph/model.go`):
- Accessible via `g` key from home (when platforms selected)
- Uses `GraphPanel` component for rendering
- Shows Nix derivation tree for selected platforms
- `w` key shows why-rebuild details (bitbake-diffsigs for Yocto)

## Error Handling & Recovery

### Error Display Strategy

**Structured Error Types:**
```go
type BuildError struct {
    Raw         string
    Level       ErrorLevel
    Source      string          // "gcc", "nix", "bitbake", "script"
    Location    *ErrorLocation  // File:line if parseable
    Suggestion  string          // Auto-generated
}
```

**Error Parsing:** Regex patterns for common build system errors:
- Nix: `error: experimental.*disabled`, `error: getting status of.*Permission denied`
- Bitbake: `ERROR: Task.*failed`, `ERROR: ExpansionError`
- GCC: `undefined reference to`, `error: .*undeclared`

**Error Zoom Pattern** (k9s-style):
1. Build fails, log shows errors
2. Press `e` from any screen showing that build
3. `LogViewer` jumps to first error with context highlighted
4. Suggestion shown if pattern matched: `Suggestion: Add libfoo to DEPENDS`

### Graceful Degradation

Research principle: "blocked capability states are explicit and non-fatal"

**Scenarios:**
- Nix unavailable → Actions disabled, show setup instructions, navigation still works
- Graph parsing fails → Show error message, offer text fallback (list view)
- Metrics collection fails → Hide metrics bar, show "Metrics unavailable", build continues
- Developer-downloads missing → Show "unavailable" badges, don't block operations

**Recovery:**
- `r` key refreshes capabilities and re-checks blockers
- Explicit error messages with remediation steps
- No silent failures

## Testing Strategy

### Unit Tests

**Atom Tests:**
```go
func TestCard_RendersWithinWidth(t *testing.T)
func TestStatusBadge_ColorMapping(t *testing.T)
func TestKeyHint_DisabledState(t *testing.T)
func TestProgressBar_FractionDisplay(t *testing.T)
```

**Molecule Tests:**
```go
func TestDataTable_Sorting(t *testing.T)
func TestDataTable_VirtualizationWith1000Rows(t *testing.T)
func TestLogViewer_ErrorNavigation(t *testing.T)
func TestLogViewer_FilterByLevel(t *testing.T)
func TestGraphPanel_TreeRendering(t *testing.T)
func TestGraphPanel_WhyRebuildDetail(t *testing.T)
```

### Integration Tests

**Screen Composition:**
```go
func TestHomeScreen_UsesComponents(t *testing.T)
func TestBuildScreen_ThreePaneLayout(t *testing.T)
func TestCachesScreen_TableAndMetrics(t *testing.T)
```

**Feature Flows:**
```go
func TestErrorZoom_FromHomeToLog(t *testing.T)
func TestGraphScreen_FromPlatformSelection(t *testing.T)
func TestCacheDeletion_ConfirmationFlow(t *testing.T)
```

### Visual Regression Tests

```go
// docs/visual_test.go
func TestVisualContracts_CardRendering(t *testing.T) {
    card := primitives.Card{Title: "Test", Content: "Content", Style: primitives.CardNormal}
    got := card.Render(40)
    want := readGoldenFile(t, "card-normal-40w.txt")
    if got != want {
        t.Errorf("Card rendering changed:\n%s", diff(want, got))
    }
}
```

**Golden files:** Store expected visual output for components, detect unintended changes.

### Performance Tests

```go
func BenchmarkDiffRender_LargeView(b *testing.B)
func BenchmarkDataTable_1000Rows(b *testing.B)
func BenchmarkLogParser_10MBLog(b *testing.B)
func BenchmarkGraphPanel_100Nodes(b *testing.B)
```

**Targets:**
- Diff render: >100 FPS (16ms per frame)
- DataTable scroll: <5ms latency
- Log parsing: <100ms for 10MB file
- Graph rendering: <200ms for 100-node tree

## Dependencies

**New Required:**
- `github.com/mattn/go-runewidth` - Unicode width calculation (research requirement)
- `github.com/shirou/gopsutil/v3` - System metrics (CPU/mem/disk)
- `github.com/charmbracelet/bubbles/table` - Foundation for DataTable
- `github.com/charmbracelet/bubbles/viewport` - Foundation for LogViewer

**Optional (Phase 4+):**
- `github.com/goccy/go-graphviz` - DOT format parsing for Yocto graphs

**Already Available:**
- `github.com/charmbracelet/bubbletea` - Framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/atotto/clipboard` - Clipboard support
- `github.com/pelletier/go-toml/v2` - Config

**Build Constraints:**
- Pure Go, no C dependencies (cross-compilation)
- Binary size target: <15MB (k9s is ~20MB, aim lower)
- Startup time target: <100ms (critical for flow state per research)

## Implementation Approach

### Refactor Sequence

**Wave 1 - Foundation (Week 1):**
1. Create `internal/styles/tokens.go` with Catppuccin Mocha palette
2. Create `internal/components/primitives/` with Card, StatusBadge, KeyHint, ProgressBar
3. Add unit tests for all atoms
4. Refactor success + hints screens to validate patterns

**Wave 2 - Molecules (Week 2):**
5. Create `DataTable` component (with bubbles/table foundation)
6. Create `LogViewer` component (with error parsing)
7. Add unit tests for both molecules
8. Refactor validations + hydration screens

**Wave 3 - Complex Screens (Week 3):**
9. Create `ParallelTracker`, `MetricsBar`, `CommandDisplay`
10. Refactor caches screen (DataTable + metrics)
11. Refactor build screen (3-pane layout)
12. Refactor home screen (Card + DataTable + metrics)

**Wave 4 - Advanced Features (Week 4-5):**
13. Create `GraphPanel` component
14. Implement graph screen with Nix parser
15. Add integration tests for full flows

**Wave 5 - Phase 2/3 Features (Week 6-7):**
16. Implement cache management actions (now trivial with DataTable)
17. Implement log slicing (now trivial with LogViewer)
18. Add developer-downloads integration (StatusBadge updates)

### Validation Checkpoints

After each wave:
- All tests pass (unit + integration)
- No visual regressions (golden file tests)
- Existing functionality preserved
- Performance benchmarks meet targets

## Risk Mitigation

**Risk:** Large refactor introduces bugs.  
**Mitigation:** Incremental waves, test each wave before advancing.

**Risk:** Over-engineering components that aren't reused.  
**Mitigation:** Start with most reusable (Card, StatusBadge) and only add components as screens need them.

**Risk:** Performance degradation from abstraction overhead.  
**Mitigation:** Benchmark after each wave, use differential rendering for high-frequency updates.

**Risk:** Breaking existing Phase 2/3 plans.  
**Mitigation:** Component system enables Phase 2/3, doesn't replace it. Features are still required, just easier to build.

## Success Criteria

**After completion, augury-node-tui will have:**
- ✅ Consistent visual language across all screens (Catppuccin Mocha)
- ✅ Reusable component library (10-15 components covering all patterns)
- ✅ Research-backed performance (differential rendering, exact geometry)
- ✅ Advanced features ready (dependency graph, build metrics, error zoom)
- ✅ Phase 2/3 implementation trivially easy (reuse DataTable, LogViewer)
- ✅ "Joy of constraints" achieved (every visual element serves function)
- ✅ Matches quality of k9s and lazygit (research's gold standards)

## Future Phases

**Phase 5 (Post-Component-System):**
- AI error translation (LLM-powered explanations)
- Build command suggestions (learn from history)
- Automated remediation (ctrl+f to apply suggested fixes)

**Phase 6 (Advanced Observability):**
- Yocto dependency graph (bitbake -g parsing)
- Real-time sstate-cache topology map
- Build timeline visualization (Gantt chart of parallel builds)

## References

- Research: "The Architecture of Modern Terminal User Interfaces" (2026)
- Existing plans: `docs/plans/2026-03-08-augury-node-tui-phase2-phase3-design.md`
- Codebase: `augury-node-tui` (~8k lines, Bubble Tea + Lip Gloss)
