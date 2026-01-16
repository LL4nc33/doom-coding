package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCheckboxGroup(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Description: "First option", Checked: false, Enabled: true},
		{Label: "Option 2", Description: "Second option", Checked: true, Enabled: true},
		{Label: "Option 3", Description: "Disabled", Checked: false, Enabled: false},
	}

	cg := NewCheckboxGroup("Test Group", items)

	if cg.Title != "Test Group" {
		t.Errorf("Expected title 'Test Group', got %q", cg.Title)
	}

	if len(cg.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(cg.Items))
	}

	if cg.Cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", cg.Cursor)
	}

	if !cg.Focused {
		t.Error("Expected Focused to be true by default")
	}
}

func TestCheckboxGroupInit(t *testing.T) {
	cg := NewCheckboxGroup("Test", []CheckboxItem{})
	cmd := cg.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestCheckboxGroupMoveCursor(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
		{Label: "Option 3", Enabled: true},
	}
	cg := NewCheckboxGroup("Test", items)

	// Move down
	cg.MoveCursor(1)
	if cg.Cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", cg.Cursor)
	}

	// Move down again
	cg.MoveCursor(1)
	if cg.Cursor != 2 {
		t.Errorf("Expected cursor at 2, got %d", cg.Cursor)
	}

	// Move down - should wrap to 0
	cg.MoveCursor(1)
	if cg.Cursor != 0 {
		t.Errorf("Expected cursor to wrap to 0, got %d", cg.Cursor)
	}

	// Move up - should wrap to end
	cg.MoveCursor(-1)
	if cg.Cursor != 2 {
		t.Errorf("Expected cursor to wrap to 2, got %d", cg.Cursor)
	}
}

func TestCheckboxGroupToggle(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: false, Enabled: true},
		{Label: "Option 2", Checked: true, Enabled: true},
		{Label: "Disabled", Checked: false, Enabled: false},
	}
	cg := NewCheckboxGroup("Test", items)

	// Toggle first item (unchecked -> checked)
	cg.Toggle()
	if !cg.Items[0].Checked {
		t.Error("Expected first item to be checked after toggle")
	}

	// Toggle again (checked -> unchecked)
	cg.Toggle()
	if cg.Items[0].Checked {
		t.Error("Expected first item to be unchecked after second toggle")
	}

	// Move to second item and toggle
	cg.MoveCursor(1)
	cg.Toggle()
	if cg.Items[1].Checked {
		t.Error("Expected second item to be unchecked after toggle")
	}

	// Move to disabled item and try to toggle - should not change
	cg.MoveCursor(1)
	cg.Toggle()
	if cg.Items[2].Checked {
		t.Error("Disabled item should not be toggled")
	}
}

func TestCheckboxGroupSelectAll(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: false, Enabled: true},
		{Label: "Option 2", Checked: false, Enabled: true},
		{Label: "Disabled", Checked: false, Enabled: false},
	}
	cg := NewCheckboxGroup("Test", items)

	cg.SelectAll()

	if !cg.Items[0].Checked {
		t.Error("First item should be checked")
	}
	if !cg.Items[1].Checked {
		t.Error("Second item should be checked")
	}
	if cg.Items[2].Checked {
		t.Error("Disabled item should not be checked")
	}
}

func TestCheckboxGroupSelectNone(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: true, Enabled: true},
		{Label: "Option 2", Checked: true, Enabled: true},
		{Label: "Disabled", Checked: false, Enabled: false},
	}
	cg := NewCheckboxGroup("Test", items)

	cg.SelectNone()

	if cg.Items[0].Checked {
		t.Error("First item should be unchecked")
	}
	if cg.Items[1].Checked {
		t.Error("Second item should be unchecked")
	}
}

func TestCheckboxGroupGetSelected(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: true, Enabled: true},
		{Label: "Option 2", Checked: false, Enabled: true},
		{Label: "Option 3", Checked: true, Enabled: true},
	}
	cg := NewCheckboxGroup("Test", items)

	selected := cg.GetSelected()

	if len(selected) != 2 {
		t.Errorf("Expected 2 selected items, got %d", len(selected))
	}

	if selected[0] != 0 || selected[1] != 2 {
		t.Errorf("Expected indices [0, 2], got %v", selected)
	}
}

func TestCheckboxGroupSetEnabled(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: true, Enabled: true},
	}
	cg := NewCheckboxGroup("Test", items)

	// Disable item
	cg.SetEnabled(0, false)
	if cg.Items[0].Enabled {
		t.Error("Item should be disabled")
	}
	if cg.Items[0].Checked {
		t.Error("Disabled item should be unchecked automatically")
	}

	// Re-enable item
	cg.SetEnabled(0, true)
	if !cg.Items[0].Enabled {
		t.Error("Item should be enabled")
	}

	// Test with invalid index
	cg.SetEnabled(-1, false) // Should not panic
	cg.SetEnabled(100, false) // Should not panic
}

func TestCheckboxGroupSetChecked(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: false, Enabled: true},
		{Label: "Disabled", Checked: false, Enabled: false},
	}
	cg := NewCheckboxGroup("Test", items)

	// Set checked
	cg.SetChecked(0, true)
	if !cg.Items[0].Checked {
		t.Error("Item should be checked")
	}

	// Try to set disabled item
	cg.SetChecked(1, true)
	if cg.Items[1].Checked {
		t.Error("Disabled item should not be checked")
	}

	// Test with invalid index
	cg.SetChecked(-1, true) // Should not panic
	cg.SetChecked(100, true) // Should not panic
}

func TestCheckboxGroupUpdate(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: false, Enabled: true},
		{Label: "Option 2", Checked: false, Enabled: true},
	}
	cg := NewCheckboxGroup("Test", items)

	// Test down key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	cg, _ = cg.Update(downMsg)
	if cg.Cursor != 1 {
		t.Errorf("Down key should move cursor to 1, got %d", cg.Cursor)
	}

	// Test up key
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	cg, _ = cg.Update(upMsg)
	if cg.Cursor != 0 {
		t.Errorf("Up key should move cursor to 0, got %d", cg.Cursor)
	}

	// Test j key (vim-style down)
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	cg, _ = cg.Update(jMsg)
	if cg.Cursor != 1 {
		t.Errorf("'j' key should move cursor to 1, got %d", cg.Cursor)
	}

	// Test k key (vim-style up)
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	cg, _ = cg.Update(kMsg)
	if cg.Cursor != 0 {
		t.Errorf("'k' key should move cursor to 0, got %d", cg.Cursor)
	}

	// Test space to toggle
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	cg, _ = cg.Update(spaceMsg)
	if !cg.Items[0].Checked {
		t.Error("Space should toggle item to checked")
	}

	// Test 'x' to toggle
	xMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	cg, _ = cg.Update(xMsg)
	if cg.Items[0].Checked {
		t.Error("'x' should toggle item to unchecked")
	}

	// Test 'a' to select all
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	cg, _ = cg.Update(aMsg)
	if !cg.Items[0].Checked || !cg.Items[1].Checked {
		t.Error("'a' should select all items")
	}

	// Test 'n' to select none
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	cg, _ = cg.Update(nMsg)
	if cg.Items[0].Checked || cg.Items[1].Checked {
		t.Error("'n' should deselect all items")
	}
}

func TestCheckboxGroupUpdateNotFocused(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: false, Enabled: true},
	}
	cg := NewCheckboxGroup("Test", items)
	cg.Focused = false

	// Keys should not work when not focused
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	cg, _ = cg.Update(spaceMsg)
	if cg.Items[0].Checked {
		t.Error("Should not toggle when not focused")
	}
}

func TestCheckboxGroupView(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Description: "First option", Checked: true, Enabled: true},
		{Label: "Option 2", Description: "", Checked: false, Enabled: true},
		{Label: "Disabled", Description: "Cannot select", Checked: false, Enabled: false},
	}
	cg := NewCheckboxGroup("Test Group", items)

	view := cg.View()

	// Check title is present
	if !strings.Contains(view, "Test Group") {
		t.Error("View should contain title")
	}

	// Check all labels are present
	if !strings.Contains(view, "Option 1") {
		t.Error("View should contain Option 1")
	}
	if !strings.Contains(view, "Option 2") {
		t.Error("View should contain Option 2")
	}
	if !strings.Contains(view, "Disabled") {
		t.Error("View should contain Disabled")
	}

	// Check descriptions are present
	if !strings.Contains(view, "First option") {
		t.Error("View should contain description for Option 1")
	}
	if !strings.Contains(view, "Cannot select") {
		t.Error("View should contain description for Disabled")
	}
}

func TestCheckboxGroupViewNoTitle(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: false, Enabled: true},
	}
	cg := NewCheckboxGroup("", items)

	view := cg.View()

	// Should still render without error
	if !strings.Contains(view, "Option 1") {
		t.Error("View should contain Option 1 even without title")
	}
}

func TestCheckboxGroupViewCursor(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Option 1", Checked: false, Enabled: true},
		{Label: "Option 2", Checked: false, Enabled: true},
	}
	cg := NewCheckboxGroup("Test", items)
	cg.Focused = true
	cg.Cursor = 1

	view := cg.View()

	// The cursor indicator should be present (it's styled, so just check the view is non-empty)
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestCheckboxGroupEmptyItems(t *testing.T) {
	cg := NewCheckboxGroup("Empty", []CheckboxItem{})

	// Should handle empty items gracefully
	cg.MoveCursor(1)  // Should not panic
	cg.Toggle()       // Should not panic
	cg.SelectAll()    // Should not panic
	cg.SelectNone()   // Should not panic

	selected := cg.GetSelected()
	if len(selected) != 0 {
		t.Errorf("Expected empty selection, got %v", selected)
	}

	view := cg.View()
	if !strings.Contains(view, "Empty") {
		t.Error("View should still contain title")
	}
}

func TestCheckboxItem(t *testing.T) {
	item := CheckboxItem{
		Label:       "Test Label",
		Description: "Test Description",
		Checked:     true,
		Enabled:     true,
	}

	if item.Label != "Test Label" {
		t.Error("Label mismatch")
	}
	if item.Description != "Test Description" {
		t.Error("Description mismatch")
	}
	if !item.Checked {
		t.Error("Checked should be true")
	}
	if !item.Enabled {
		t.Error("Enabled should be true")
	}
}
