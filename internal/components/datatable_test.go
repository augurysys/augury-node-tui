package components

import (
	"fmt"
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
