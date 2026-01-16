package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RadioItem represents a single radio option
type RadioItem struct {
	Label       string
	Description string
	Hint        string
	Icon        string
	Enabled     bool
}

// RadioGroup is a component for selecting a single option
type RadioGroup struct {
	Title    string
	Items    []RadioItem
	Selected int
	Cursor   int
	Focused  bool

	// Styles
	TitleStyle       lipgloss.Style
	ItemStyle        lipgloss.Style
	SelectedStyle    lipgloss.Style
	DisabledStyle    lipgloss.Style
	DescriptionStyle lipgloss.Style
	HintStyle        lipgloss.Style
	CursorStyle      lipgloss.Style
}

// NewRadioGroup creates a new radio group
func NewRadioGroup(title string, items []RadioItem) RadioGroup {
	return RadioGroup{
		Title:    title,
		Items:    items,
		Selected: 0,
		Cursor:   0,
		Focused:  true,
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
			PaddingLeft(5),
		HintStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C5E46")).
			Italic(true).
			PaddingLeft(5),
		CursorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A7C34")).
			Bold(true),
	}
}

// Init initializes the component
func (r RadioGroup) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (r RadioGroup) Update(msg tea.Msg) (RadioGroup, tea.Cmd) {
	if !r.Focused {
		return r, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			r.MoveCursor(-1)
		case "down", "j":
			r.MoveCursor(1)
		case " ", "enter":
			r.Select()
		}
	}

	return r, nil
}

// View renders the component
func (r RadioGroup) View() string {
	var sb strings.Builder

	if r.Title != "" {
		sb.WriteString(r.TitleStyle.Render(r.Title))
		sb.WriteString("\n\n")
	}

	for i, item := range r.Items {
		cursor := "  "
		if i == r.Cursor && r.Focused {
			cursor = r.CursorStyle.Render("▸ ")
		}

		var radio string
		var style lipgloss.Style

		if !item.Enabled {
			radio = "( )"
			style = r.DisabledStyle
		} else if i == r.Selected {
			radio = r.SelectedStyle.Render("(●)")
			style = r.SelectedStyle
		} else {
			radio = "( )"
			style = r.ItemStyle
		}

		if i == r.Cursor && r.Focused && item.Enabled {
			style = r.SelectedStyle
		}

		icon := ""
		if item.Icon != "" {
			icon = item.Icon + " "
		}

		sb.WriteString(fmt.Sprintf("%s%s %s%s\n", cursor, radio, icon, style.Render(item.Label)))

		if item.Description != "" {
			sb.WriteString(r.DescriptionStyle.Render(item.Description))
			sb.WriteString("\n")
		}

		if item.Hint != "" {
			sb.WriteString(r.HintStyle.Render(item.Hint))
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// MoveCursor moves the cursor by delta
func (r *RadioGroup) MoveCursor(delta int) {
	r.Cursor += delta
	if r.Cursor < 0 {
		r.Cursor = len(r.Items) - 1
	}
	if r.Cursor >= len(r.Items) {
		r.Cursor = 0
	}

	// Skip disabled items
	attempts := 0
	for !r.Items[r.Cursor].Enabled && attempts < len(r.Items) {
		r.Cursor += delta
		if r.Cursor < 0 {
			r.Cursor = len(r.Items) - 1
		}
		if r.Cursor >= len(r.Items) {
			r.Cursor = 0
		}
		attempts++
	}
}

// Select selects the current cursor position
func (r *RadioGroup) Select() {
	if r.Cursor >= 0 && r.Cursor < len(r.Items) {
		if r.Items[r.Cursor].Enabled {
			r.Selected = r.Cursor
		}
	}
}

// GetSelected returns the selected index
func (r *RadioGroup) GetSelected() int {
	return r.Selected
}

// GetSelectedItem returns the selected item
func (r *RadioGroup) GetSelectedItem() *RadioItem {
	if r.Selected >= 0 && r.Selected < len(r.Items) {
		return &r.Items[r.Selected]
	}
	return nil
}

// SetEnabled enables or disables an item
func (r *RadioGroup) SetEnabled(index int, enabled bool) {
	if index >= 0 && index < len(r.Items) {
		r.Items[index].Enabled = enabled
	}
}

// SetSelected sets the selected index
func (r *RadioGroup) SetSelected(index int) {
	if index >= 0 && index < len(r.Items) && r.Items[index].Enabled {
		r.Selected = index
	}
}
