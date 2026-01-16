package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CheckboxItem represents a single checkbox option
type CheckboxItem struct {
	Label       string
	Description string
	Checked     bool
	Enabled     bool
}

// CheckboxGroup is a component for selecting multiple options
type CheckboxGroup struct {
	Title    string
	Items    []CheckboxItem
	Cursor   int
	Focused  bool

	// Styles
	TitleStyle       lipgloss.Style
	ItemStyle        lipgloss.Style
	SelectedStyle    lipgloss.Style
	DisabledStyle    lipgloss.Style
	DescriptionStyle lipgloss.Style
	CursorStyle      lipgloss.Style
}

// NewCheckboxGroup creates a new checkbox group
func NewCheckboxGroup(title string, items []CheckboxItem) CheckboxGroup {
	return CheckboxGroup{
		Title:   title,
		Items:   items,
		Cursor:  0,
		Focused: true,
		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2E521D")).
			Bold(true).
			MarginBottom(1),
		ItemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")),
		SelectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A7C34")).
			Bold(true),
		DisabledStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		DescriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			PaddingLeft(7),
		CursorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A7C34")).
			Bold(true),
	}
}

// Init initializes the component
func (c CheckboxGroup) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (c CheckboxGroup) Update(msg tea.Msg) (CheckboxGroup, tea.Cmd) {
	if !c.Focused {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			c.MoveCursor(-1)
		case "down", "j":
			c.MoveCursor(1)
		case " ", "x":
			c.Toggle()
		case "a":
			c.SelectAll()
		case "n":
			c.SelectNone()
		}
	}

	return c, nil
}

// View renders the component
func (c CheckboxGroup) View() string {
	var sb strings.Builder

	if c.Title != "" {
		sb.WriteString(c.TitleStyle.Render(c.Title))
		sb.WriteString("\n\n")
	}

	for i, item := range c.Items {
		cursor := "  "
		if i == c.Cursor && c.Focused {
			cursor = c.CursorStyle.Render("▸ ")
		}

		var checkbox string
		var style lipgloss.Style

		if !item.Enabled {
			checkbox = "[-]"
			style = c.DisabledStyle
		} else if item.Checked {
			checkbox = c.SelectedStyle.Render("[✓]")
			style = c.SelectedStyle
		} else {
			checkbox = "[ ]"
			style = c.ItemStyle
		}

		if i == c.Cursor && c.Focused {
			style = c.SelectedStyle
		}

		sb.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, style.Render(item.Label)))

		if item.Description != "" {
			sb.WriteString(c.DescriptionStyle.Render(item.Description))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// MoveCursor moves the cursor by delta
func (c *CheckboxGroup) MoveCursor(delta int) {
	c.Cursor += delta
	if c.Cursor < 0 {
		c.Cursor = len(c.Items) - 1
	}
	if c.Cursor >= len(c.Items) {
		c.Cursor = 0
	}
}

// Toggle toggles the current item
func (c *CheckboxGroup) Toggle() {
	if c.Cursor >= 0 && c.Cursor < len(c.Items) {
		if c.Items[c.Cursor].Enabled {
			c.Items[c.Cursor].Checked = !c.Items[c.Cursor].Checked
		}
	}
}

// SelectAll selects all enabled items
func (c *CheckboxGroup) SelectAll() {
	for i := range c.Items {
		if c.Items[i].Enabled {
			c.Items[i].Checked = true
		}
	}
}

// SelectNone deselects all items
func (c *CheckboxGroup) SelectNone() {
	for i := range c.Items {
		c.Items[i].Checked = false
	}
}

// GetSelected returns indices of selected items
func (c *CheckboxGroup) GetSelected() []int {
	var selected []int
	for i, item := range c.Items {
		if item.Checked {
			selected = append(selected, i)
		}
	}
	return selected
}

// SetEnabled enables or disables an item
func (c *CheckboxGroup) SetEnabled(index int, enabled bool) {
	if index >= 0 && index < len(c.Items) {
		c.Items[index].Enabled = enabled
		if !enabled {
			c.Items[index].Checked = false
		}
	}
}

// SetChecked sets the checked state of an item
func (c *CheckboxGroup) SetChecked(index int, checked bool) {
	if index >= 0 && index < len(c.Items) && c.Items[index].Enabled {
		c.Items[index].Checked = checked
	}
}
