package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRadioGroup(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Description: "First option", Enabled: true},
		{Label: "Option 2", Description: "Second option", Hint: "Recommended", Enabled: true},
		{Label: "Option 3", Description: "Disabled", Icon: "!", Enabled: false},
	}

	rg := NewRadioGroup("Test Group", items)

	if rg.Title != "Test Group" {
		t.Errorf("Expected title 'Test Group', got %q", rg.Title)
	}

	if len(rg.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(rg.Items))
	}

	if rg.Selected != 0 {
		t.Errorf("Expected selected at 0, got %d", rg.Selected)
	}

	if rg.Cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", rg.Cursor)
	}

	if !rg.Focused {
		t.Error("Expected Focused to be true by default")
	}
}

func TestRadioGroupInit(t *testing.T) {
	rg := NewRadioGroup("Test", []RadioItem{})
	cmd := rg.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestRadioGroupMoveCursor(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
		{Label: "Option 3", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)

	// Move down
	rg.MoveCursor(1)
	if rg.Cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", rg.Cursor)
	}

	// Move down again
	rg.MoveCursor(1)
	if rg.Cursor != 2 {
		t.Errorf("Expected cursor at 2, got %d", rg.Cursor)
	}

	// Move down - should wrap to 0
	rg.MoveCursor(1)
	if rg.Cursor != 0 {
		t.Errorf("Expected cursor to wrap to 0, got %d", rg.Cursor)
	}

	// Move up - should wrap to end
	rg.MoveCursor(-1)
	if rg.Cursor != 2 {
		t.Errorf("Expected cursor to wrap to 2, got %d", rg.Cursor)
	}
}

func TestRadioGroupMoveCursorSkipsDisabled(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Disabled 1", Enabled: false},
		{Label: "Disabled 2", Enabled: false},
		{Label: "Option 4", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)

	// Move down - should skip disabled items
	rg.MoveCursor(1)
	if rg.Cursor != 3 {
		t.Errorf("Expected cursor to skip to 3, got %d", rg.Cursor)
	}

	// Move up - should skip disabled items
	rg.MoveCursor(-1)
	if rg.Cursor != 0 {
		t.Errorf("Expected cursor to skip back to 0, got %d", rg.Cursor)
	}
}

func TestRadioGroupSelect(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
		{Label: "Disabled", Enabled: false},
	}
	rg := NewRadioGroup("Test", items)

	// Select first item (default)
	rg.Select()
	if rg.Selected != 0 {
		t.Errorf("Expected selected at 0, got %d", rg.Selected)
	}

	// Move to second and select
	rg.MoveCursor(1)
	rg.Select()
	if rg.Selected != 1 {
		t.Errorf("Expected selected at 1, got %d", rg.Selected)
	}

	// Move to disabled and try to select - should not change
	rg.Cursor = 2 // Manually set cursor to disabled item
	rg.Select()
	if rg.Selected != 1 {
		t.Errorf("Expected selected to remain at 1, got %d", rg.Selected)
	}
}

func TestRadioGroupGetSelected(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)

	if rg.GetSelected() != 0 {
		t.Errorf("Expected selected to be 0, got %d", rg.GetSelected())
	}

	rg.Selected = 1
	if rg.GetSelected() != 1 {
		t.Errorf("Expected selected to be 1, got %d", rg.GetSelected())
	}
}

func TestRadioGroupGetSelectedItem(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Description: "First", Enabled: true},
		{Label: "Option 2", Description: "Second", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)

	item := rg.GetSelectedItem()
	if item == nil {
		t.Fatal("GetSelectedItem returned nil")
	}
	if item.Label != "Option 1" {
		t.Errorf("Expected label 'Option 1', got %q", item.Label)
	}

	rg.Selected = 1
	item = rg.GetSelectedItem()
	if item.Label != "Option 2" {
		t.Errorf("Expected label 'Option 2', got %q", item.Label)
	}
}

func TestRadioGroupGetSelectedItemOutOfBounds(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)

	rg.Selected = -1
	if rg.GetSelectedItem() != nil {
		t.Error("GetSelectedItem should return nil for invalid index")
	}

	rg.Selected = 100
	if rg.GetSelectedItem() != nil {
		t.Error("GetSelectedItem should return nil for out of bounds index")
	}
}

func TestRadioGroupSetEnabled(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)

	// Disable item
	rg.SetEnabled(0, false)
	if rg.Items[0].Enabled {
		t.Error("Item should be disabled")
	}

	// Re-enable item
	rg.SetEnabled(0, true)
	if !rg.Items[0].Enabled {
		t.Error("Item should be enabled")
	}

	// Test with invalid index
	rg.SetEnabled(-1, false) // Should not panic
	rg.SetEnabled(100, false) // Should not panic
}

func TestRadioGroupSetSelected(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
		{Label: "Disabled", Enabled: false},
	}
	rg := NewRadioGroup("Test", items)

	// Set selected to second item
	rg.SetSelected(1)
	if rg.Selected != 1 {
		t.Errorf("Expected selected at 1, got %d", rg.Selected)
	}

	// Try to set selected to disabled item - should not change
	rg.SetSelected(2)
	if rg.Selected != 1 {
		t.Errorf("Expected selected to remain at 1, got %d", rg.Selected)
	}

	// Test with invalid index
	rg.SetSelected(-1)   // Should not change
	rg.SetSelected(100)  // Should not change
	if rg.Selected != 1 {
		t.Errorf("Expected selected to remain at 1 after invalid operations, got %d", rg.Selected)
	}
}

func TestRadioGroupUpdate(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)

	// Test down key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	rg, _ = rg.Update(downMsg)
	if rg.Cursor != 1 {
		t.Errorf("Down key should move cursor to 1, got %d", rg.Cursor)
	}

	// Test up key
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	rg, _ = rg.Update(upMsg)
	if rg.Cursor != 0 {
		t.Errorf("Up key should move cursor to 0, got %d", rg.Cursor)
	}

	// Test j key (vim-style down)
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	rg, _ = rg.Update(jMsg)
	if rg.Cursor != 1 {
		t.Errorf("'j' key should move cursor to 1, got %d", rg.Cursor)
	}

	// Test k key (vim-style up)
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	rg, _ = rg.Update(kMsg)
	if rg.Cursor != 0 {
		t.Errorf("'k' key should move cursor to 0, got %d", rg.Cursor)
	}

	// Test space to select
	rg.MoveCursor(1)
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	rg, _ = rg.Update(spaceMsg)
	if rg.Selected != 1 {
		t.Errorf("Space should select current item, got selected=%d", rg.Selected)
	}

	// Test enter to select
	rg.MoveCursor(-1)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	rg, _ = rg.Update(enterMsg)
	if rg.Selected != 0 {
		t.Errorf("Enter should select current item, got selected=%d", rg.Selected)
	}
}

func TestRadioGroupUpdateNotFocused(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)
	rg.Focused = false

	// Keys should not work when not focused
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	rg, _ = rg.Update(downMsg)
	if rg.Cursor != 0 {
		t.Error("Should not move cursor when not focused")
	}
}

func TestRadioGroupView(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Description: "First option", Hint: "", Icon: "", Enabled: true},
		{Label: "Option 2", Description: "", Hint: "Recommended", Icon: "*", Enabled: true},
		{Label: "Disabled", Description: "Cannot select", Hint: "", Icon: "", Enabled: false},
	}
	rg := NewRadioGroup("Test Group", items)

	view := rg.View()

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

	// Check hint is present
	if !strings.Contains(view, "Recommended") {
		t.Error("View should contain hint for Option 2")
	}

	// Check icon is present
	if !strings.Contains(view, "*") {
		t.Error("View should contain icon for Option 2")
	}
}

func TestRadioGroupViewNoTitle(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
	}
	rg := NewRadioGroup("", items)

	view := rg.View()

	// Should still render without error
	if !strings.Contains(view, "Option 1") {
		t.Error("View should contain Option 1 even without title")
	}
}

func TestRadioGroupViewSelectedIndicator(t *testing.T) {
	items := []RadioItem{
		{Label: "Option 1", Enabled: true},
		{Label: "Option 2", Enabled: true},
	}
	rg := NewRadioGroup("Test", items)
	rg.Selected = 1

	view := rg.View()

	// View should render (the selected indicator is styled)
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestRadioGroupEmptyItems(t *testing.T) {
	rg := NewRadioGroup("Empty", []RadioItem{})

	// Should handle empty items gracefully
	rg.MoveCursor(1)  // Should not panic
	rg.Select()       // Should not panic

	selected := rg.GetSelected()
	if selected != 0 {
		t.Errorf("Expected selected=0 for empty group, got %d", selected)
	}

	item := rg.GetSelectedItem()
	if item != nil {
		t.Error("GetSelectedItem should return nil for empty group")
	}

	view := rg.View()
	if !strings.Contains(view, "Empty") {
		t.Error("View should still contain title")
	}
}

func TestRadioItem(t *testing.T) {
	item := RadioItem{
		Label:       "Test Label",
		Description: "Test Description",
		Hint:        "Test Hint",
		Icon:        "!",
		Enabled:     true,
	}

	if item.Label != "Test Label" {
		t.Error("Label mismatch")
	}
	if item.Description != "Test Description" {
		t.Error("Description mismatch")
	}
	if item.Hint != "Test Hint" {
		t.Error("Hint mismatch")
	}
	if item.Icon != "!" {
		t.Error("Icon mismatch")
	}
	if !item.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestRadioGroupAllDisabled(t *testing.T) {
	items := []RadioItem{
		{Label: "Disabled 1", Enabled: false},
		{Label: "Disabled 2", Enabled: false},
		{Label: "Disabled 3", Enabled: false},
	}
	rg := NewRadioGroup("Test", items)

	// MoveCursor should handle all disabled items
	rg.MoveCursor(1)
	// The cursor should still move (wrapping occurs), but at some point it returns
	// This tests that it doesn't infinite loop

	// Select should not change selection on disabled items
	rg.Select()
	// No panic is the success here
}
