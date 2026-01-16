package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a single form field
type FormField struct {
	Label       string
	Placeholder string
	Help        string
	Value       string
	Required    bool
	Secret      bool
	Validator   func(string) error

	input textinput.Model
}

// Form is a component for collecting multiple text inputs
type Form struct {
	Title       string
	Fields      []FormField
	FocusIndex  int
	Focused     bool
	Submitted   bool
	Errors      map[int]string

	// Styles
	TitleStyle       lipgloss.Style
	LabelStyle       lipgloss.Style
	FocusedStyle     lipgloss.Style
	HelpStyle        lipgloss.Style
	ErrorStyle       lipgloss.Style
	RequiredStyle    lipgloss.Style
}

// NewForm creates a new form with the given fields
func NewForm(title string, fields []FormField) Form {
	f := Form{
		Title:      title,
		Fields:     fields,
		FocusIndex: 0,
		Focused:    true,
		Errors:     make(map[int]string),
		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2E521D")).
			Bold(true).
			MarginBottom(1),
		LabelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C5E46")).
			Bold(true),
		FocusedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A7C34")),
		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")),
		RequiredStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true),
	}

	// Initialize text inputs for each field
	for i := range f.Fields {
		input := textinput.New()
		input.Placeholder = f.Fields[i].Placeholder
		input.CharLimit = 150
		input.Width = 50

		if f.Fields[i].Secret {
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '*'
		}

		if f.Fields[i].Value != "" {
			input.SetValue(f.Fields[i].Value)
		}

		if i == 0 {
			input.Focus()
		}

		f.Fields[i].input = input
	}

	return f
}

// Init initializes the component
func (f Form) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	if !f.Focused {
		return f, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down", "enter":
			if msg.String() == "enter" && f.FocusIndex == len(f.Fields)-1 {
				// On last field, try to submit
				if f.Validate() {
					f.Submitted = true
					return f, nil
				}
			}
			f.nextField()
			return f, nil
		case "shift+tab", "up":
			f.prevField()
			return f, nil
		}
	}

	// Update the focused input
	if f.FocusIndex >= 0 && f.FocusIndex < len(f.Fields) {
		var cmd tea.Cmd
		f.Fields[f.FocusIndex].input, cmd = f.Fields[f.FocusIndex].input.Update(msg)
		// Sync the value
		f.Fields[f.FocusIndex].Value = f.Fields[f.FocusIndex].input.Value()
		return f, cmd
	}

	return f, nil
}

// View renders the form
func (f Form) View() string {
	var sb strings.Builder

	if f.Title != "" {
		sb.WriteString(f.TitleStyle.Render(f.Title))
		sb.WriteString("\n\n")
	}

	for i, field := range f.Fields {
		// Label with required marker
		label := field.Label
		if field.Required {
			label = label + f.RequiredStyle.Render(" *")
		}

		style := f.LabelStyle
		if i == f.FocusIndex && f.Focused {
			style = f.FocusedStyle
		}
		sb.WriteString(style.Render(label))
		sb.WriteString("\n")

		// Input field
		sb.WriteString(field.input.View())
		sb.WriteString("\n")

		// Error message
		if err, exists := f.Errors[i]; exists && err != "" {
			sb.WriteString(f.ErrorStyle.Render("  ⚠ " + err))
			sb.WriteString("\n")
		}

		// Help text
		if field.Help != "" {
			sb.WriteString(f.HelpStyle.Render("  " + field.Help))
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (f *Form) nextField() {
	f.Fields[f.FocusIndex].input.Blur()
	f.FocusIndex++
	if f.FocusIndex >= len(f.Fields) {
		f.FocusIndex = 0
	}
	f.Fields[f.FocusIndex].input.Focus()
}

func (f *Form) prevField() {
	f.Fields[f.FocusIndex].input.Blur()
	f.FocusIndex--
	if f.FocusIndex < 0 {
		f.FocusIndex = len(f.Fields) - 1
	}
	f.Fields[f.FocusIndex].input.Focus()
}

// Validate validates all fields and returns true if all pass
func (f *Form) Validate() bool {
	f.Errors = make(map[int]string)
	valid := true

	for i, field := range f.Fields {
		value := field.input.Value()

		// Check required
		if field.Required && value == "" {
			f.Errors[i] = "This field is required"
			valid = false
			continue
		}

		// Run custom validator
		if field.Validator != nil && value != "" {
			if err := field.Validator(value); err != nil {
				f.Errors[i] = err.Error()
				valid = false
			}
		}
	}

	return valid
}

// GetValues returns all field values as a map
func (f *Form) GetValues() map[string]string {
	values := make(map[string]string)
	for _, field := range f.Fields {
		values[field.Label] = field.input.Value()
	}
	return values
}

// GetValue returns the value of a field by label
func (f *Form) GetValue(label string) string {
	for _, field := range f.Fields {
		if field.Label == label {
			return field.input.Value()
		}
	}
	return ""
}

// SetValue sets the value of a field by index
func (f *Form) SetValue(index int, value string) {
	if index >= 0 && index < len(f.Fields) {
		f.Fields[index].input.SetValue(value)
		f.Fields[index].Value = value
	}
}

// Focus focuses the form
func (f *Form) Focus() {
	f.Focused = true
	if f.FocusIndex >= 0 && f.FocusIndex < len(f.Fields) {
		f.Fields[f.FocusIndex].input.Focus()
	}
}

// Blur removes focus from the form
func (f *Form) Blur() {
	f.Focused = false
	for i := range f.Fields {
		f.Fields[i].input.Blur()
	}
}

// Reset resets the form to initial state
func (f *Form) Reset() {
	f.FocusIndex = 0
	f.Submitted = false
	f.Errors = make(map[int]string)
	for i := range f.Fields {
		f.Fields[i].input.Reset()
	}
	if len(f.Fields) > 0 {
		f.Fields[0].input.Focus()
	}
}

// Common validators
func ValidateNotEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("cannot be empty")
	}
	return nil
}

func ValidateMinLength(min int) func(string) error {
	return func(value string) error {
		if len(value) < min {
			return fmt.Errorf("must be at least %d characters", min)
		}
		return nil
	}
}

func ValidateTailscaleKey(value string) error {
	if !strings.HasPrefix(value, "tskey-") {
		return fmt.Errorf("should start with 'tskey-'")
	}
	return nil
}

func ValidateAnthropicKey(value string) error {
	if !strings.HasPrefix(value, "sk-ant-") {
		return fmt.Errorf("should start with 'sk-ant-'")
	}
	return nil
}
