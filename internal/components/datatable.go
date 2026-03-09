package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/augurysys/augury-node-tui/internal/styles"
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
	Width    int // -1 for flex
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
	if visibleRows < 1 {
		visibleRows = 1
	}
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
	if visibleRows < 1 {
		visibleRows = 1
	}
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
	if width <= 0 {
		return s
	}
	if len(s) <= width {
		return s + strings.Repeat(" ", width-len(s))
	}
	return s[:width-1] + "…"
}
