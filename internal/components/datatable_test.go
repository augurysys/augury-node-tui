package components

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/augurysys/augury-node-tui/internal/ansi"
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

func TestDataTable_EmptyTablePressG(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 20, Renderer: func(r interface{}) string {
			return r.(testRow).Name
		}},
	}

	table := NewDataTable(columns)
	// No rows set

	// Should not panic on G
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	table.Update(msg)

	// View should not panic
	view := table.View()
	if view == "" {
		t.Error("View should render even with empty table")
	}
}

func TestDataTable_NilRenderer(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 20}, // No Renderer
	}

	table := NewDataTable(columns)
	table.SetRows([]interface{}{testRow{Name: "Test"}})

	// Should not panic with nil renderer
	view := table.View()
	if view == "" {
		t.Error("View should render even with nil renderer")
	}
}

func TestDataTable_Sorting(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 20, Sortable: true, Renderer: func(r interface{}) string {
			return r.(testRow).Name
		}},
		{Header: "Count", Width: 10, Sortable: true, Renderer: func(r interface{}) string {
			return fmt.Sprintf("%d", r.(testRow).Count)
		}},
	}

	table := NewDataTable(columns)
	rows := []interface{}{
		testRow{Name: "Charlie", Count: 30},
		testRow{Name: "Alice", Count: 10},
		testRow{Name: "Bob", Count: 20},
	}
	table.SetRows(rows)

	// Press 1 to sort by column 0 (Name) ascending
	table.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	view := table.View()
	if !strings.Contains(view, " ▲") {
		t.Error("Header should show sort indicator ▲ when sorted ascending")
	}
	first := table.SelectedRow()
	if first == nil {
		t.Fatal("SelectedRow returned nil")
	}
	if first.(testRow).Name != "Alice" {
		t.Errorf("After sort by Name asc, first row should be Alice, got %s", first.(testRow).Name)
	}

	// Press 1 again to toggle descending
	table.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	view = table.View()
	if !strings.Contains(view, " ▼") {
		t.Error("Header should show sort indicator ▼ when sorted descending")
	}
	first = table.SelectedRow()
	if first.(testRow).Name != "Charlie" {
		t.Errorf("After sort by Name desc, first row should be Charlie, got %s", first.(testRow).Name)
	}

	// Press 2 to sort by Count
	table.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	first = table.SelectedRow()
	if first.(testRow).Count != 10 {
		t.Errorf("After sort by Count asc, first row should have Count 10, got %d", first.(testRow).Count)
	}
}

func TestDataTable_RightAlign(t *testing.T) {
	columns := []Column{
		{Header: "Name", Width: 10, Align: AlignLeft, Renderer: func(r interface{}) string {
			return r.(testRow).Name
		}},
		{Header: "Count", Width: 10, Align: AlignRight, Renderer: func(r interface{}) string {
			return fmt.Sprintf("%d", r.(testRow).Count)
		}},
	}

	table := NewDataTable(columns)
	table.SetRows([]interface{}{
		testRow{Name: "x", Count: 123},
	})

	view := table.View()
	clean := ansi.StripAnsi(view)
	// Right-aligned "123" in width 10 should have spaces on the left
	// Count column width is 10, "123" is 3 chars, so 7 spaces before
	if !strings.Contains(clean, "       123") {
		t.Errorf("Right-aligned count should have leading spaces; view: %q", clean)
	}
}
