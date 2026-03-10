package components

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/augurysys/augury-node-tui/internal/ansi"
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
	columns      []Column
	rows         []interface{}
	selectedIdx  int
	sortColumnIdx int // -1 = no sort, 0..n = column index
	sortAsc      bool
	width        int
	height       int
	scrollOff    int // Scroll offset for virtualization
}

// NewDataTable creates a new table with given columns
func NewDataTable(columns []Column) *DataTable {
	return &DataTable{
		columns:       columns,
		rows:          []interface{}{},
		selectedIdx:  0,
		sortColumnIdx: -1,
		sortAsc:       true,
		width:         80,
		height:        20,
		scrollOff:     0,
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
	if t.sortColumnIdx >= 0 && t.sortColumnIdx < len(t.columns) {
		t.sortRows()
	}
}

// SetHeight sets visible row count
func (t *DataTable) SetHeight(height int) {
	t.height = height
}

// SetWidth sets total table width
func (t *DataTable) SetWidth(width int) {
	t.width = width
}

// Update handles key messages for navigation and sorting
func (t *DataTable) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		switch k {
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
			if len(t.rows) > 0 {
				t.selectedIdx = len(t.rows) - 1
				t.adjustScroll()
			}
		case "1", "2", "3", "4", "5", "6":
			colIdx := int(k[0] - '1')
			if colIdx < len(t.columns) && t.columns[colIdx].Sortable {
				if t.sortColumnIdx == colIdx {
					t.sortAsc = !t.sortAsc
				} else {
					t.sortColumnIdx = colIdx
					t.sortAsc = true
				}
				t.sortRows()
			}
		case "s":
			t.cycleSortColumn()
		}
	}
	return nil
}

// cycleSortColumn advances to the next sortable column; at the last, wraps to the first and toggles direction.
func (t *DataTable) cycleSortColumn() {
	var sortable []int
	for i, c := range t.columns {
		if c.Sortable {
			sortable = append(sortable, i)
		}
	}
	if len(sortable) == 0 {
		return
	}
	next := -1
	for _, i := range sortable {
		if i > t.sortColumnIdx {
			next = i
			break
		}
	}
	if next >= 0 {
		t.sortColumnIdx = next
		t.sortAsc = true
	} else {
		t.sortColumnIdx = sortable[0]
		t.sortAsc = !t.sortAsc
	}
	t.sortRows()
}

func (t *DataTable) sortRows() {
	if t.sortColumnIdx < 0 || t.sortColumnIdx >= len(t.columns) || len(t.rows) == 0 {
		return
	}
	col := t.columns[t.sortColumnIdx]
	if col.Renderer == nil {
		return
	}
	sort.SliceStable(t.rows, func(i, j int) bool {
		vi := ansi.StripAnsi(col.Renderer(t.rows[i]))
		vj := ansi.StripAnsi(col.Renderer(t.rows[j]))
		if t.sortAsc {
			return vi < vj
		}
		return vi > vj
	})
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
	for i, col := range t.columns {
		headerStyle := typo.Section
		hdr := col.Header
		if t.sortColumnIdx == i {
			if t.sortAsc {
				hdr += " ▲"
			} else {
				hdr += " ▼"
			}
		}
		cell := headerStyle.Render(truncate(hdr, col.Width))
		headerCells = append(headerCells, cell)
	}
	result.WriteString(strings.Join(headerCells, " │ ") + "\n")
	result.WriteString(strings.Repeat("─", t.width) + "\n")

	if len(t.rows) == 0 {
		return result.String()
	}

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
			var content string
			if col.Renderer != nil {
				content = col.Renderer(row)
			}
			cell := alignCell(content, col.Width, col.Align)

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

// truncate limits string to display width, adding ellipsis if needed
func truncate(s string, width int) string {
	if width <= 0 {
		return s
	}

	displayWidth := runewidth.StringWidth(s)
	if displayWidth <= width {
		return s + strings.Repeat(" ", width-displayWidth)
	}

	return runewidth.Truncate(s, width-1, "…")
}

// alignCell formats cell content according to column alignment.
// Note: if content exceeds width, it is truncated and ANSI styling is removed.
func alignCell(content string, width int, align Alignment) string {
	if width <= 0 {
		return content
	}
	plain := ansi.StripAnsi(content)
	displayWidth := runewidth.StringWidth(plain)
	if displayWidth > width {
		content = runewidth.Truncate(plain, width-1, "…")
		displayWidth = runewidth.StringWidth(content)
	}
	padding := width - displayWidth
	if padding <= 0 {
		return content
	}
	switch align {
	case AlignRight:
		return strings.Repeat(" ", padding) + content
	case AlignCenter:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + content + strings.Repeat(" ", right)
	default:
		return content + strings.Repeat(" ", padding)
	}
}
