package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigField represents a configuration input field
type ConfigField struct {
	Key         string
	Label       string
	Placeholder string
	Help        string
	Required    bool
	Secret      bool
	Value       string
	Enabled     bool
	input       textinput.Model
}

// ConfigScreen handles configuration input
type ConfigScreen struct {
	Width      int
	Height     int
	Fields     []ConfigField
	FocusIndex int
	Errors     map[int]string
	Mode       DeploymentMode
}

// NewConfigScreen creates a new configuration screen
func NewConfigScreen(mode DeploymentMode, needsTailscale bool) ConfigScreen {
	fields := []ConfigField{
		{
			Key:         "tailscale_key",
			Label:       "Tailscale Auth Key",
			Placeholder: "tskey-auth-xxxxxxxxxxxxxxxxxxxxx",
			Help:        "Get from https://login.tailscale.com/admin/settings/keys",
			Required:    mode == ModeDockerTailscale,
			Secret:      false,
			Enabled:     mode == ModeDockerTailscale,
		},
		{
			Key:         "code_password",
			Label:       "code-server Password",
			Placeholder: "your-secure-password",
			Help:        "Password for web IDE access (min 8 characters)",
			Required:    mode != ModeTerminalOnly,
			Secret:      true,
			Enabled:     mode != ModeTerminalOnly,
		},
		{
			Key:         "sudo_password",
			Label:       "Container Sudo Password",
			Placeholder: "container-sudo-password",
			Help:        "Password for sudo in containers",
			Required:    mode != ModeTerminalOnly,
			Secret:      true,
			Enabled:     mode != ModeTerminalOnly,
		},
		{
			Key:         "anthropic_key",
			Label:       "Anthropic API Key",
			Placeholder: "sk-ant-api03-xxxxxxxxxxxxxxxxxxxxx",
			Help:        "Get from https://console.anthropic.com (optional)",
			Required:    false,
			Secret:      true,
			Enabled:     mode != ModeTerminalOnly,
		},
		{
			Key:         "timezone",
			Label:       "Timezone",
			Placeholder: "Europe/Berlin",
			Help:        "e.g., America/New_York, Asia/Tokyo",
			Required:    false,
			Secret:      false,
			Value:       "Europe/Berlin",
			Enabled:     true,
		},
		{
			Key:         "workspace",
			Label:       "Workspace Path",
			Placeholder: "./workspace",
			Help:        "Path for shared workspace directory",
			Required:    false,
			Secret:      false,
			Value:       "./workspace",
			Enabled:     true,
		},
	}

	// Initialize text inputs
	for i := range fields {
		input := textinput.New()
		input.Placeholder = fields[i].Placeholder
		input.CharLimit = 150
		input.Width = 50

		if fields[i].Secret {
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '*'
		}

		if fields[i].Value != "" {
			input.SetValue(fields[i].Value)
		}

		fields[i].input = input
	}

	// Find first enabled field to focus
	focusIndex := 0
	for i, f := range fields {
		if f.Enabled {
			focusIndex = i
			fields[i].input.Focus()
			break
		}
	}

	return ConfigScreen{
		Fields:     fields,
		FocusIndex: focusIndex,
		Errors:     make(map[int]string),
		Mode:       mode,
	}
}

// Init initializes the screen
func (s ConfigScreen) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles input
func (s ConfigScreen) Update(msg tea.Msg) (ConfigScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			s.nextField()
			return s, nil, ActionNone
		case "shift+tab", "up":
			s.prevField()
			return s, nil, ActionNone
		case "enter":
			// Check if last enabled field
			lastEnabled := s.FocusIndex
			for i := s.FocusIndex + 1; i < len(s.Fields); i++ {
				if s.Fields[i].Enabled {
					lastEnabled = i
				}
			}
			if s.FocusIndex >= lastEnabled {
				if s.validate() {
					return s, nil, ActionNext
				}
			} else {
				s.nextField()
			}
			return s, nil, ActionNone
		case "esc":
			return s, nil, ActionBack
		case "q", "ctrl+c":
			return s, tea.Quit, ActionQuit
		case "ctrl+v":
			// Toggle password visibility
			if s.Fields[s.FocusIndex].Secret {
				if s.Fields[s.FocusIndex].input.EchoMode == textinput.EchoPassword {
					s.Fields[s.FocusIndex].input.EchoMode = textinput.EchoNormal
				} else {
					s.Fields[s.FocusIndex].input.EchoMode = textinput.EchoPassword
				}
			}
			return s, nil, ActionNone
		}
	case tea.WindowSizeMsg:
		s.Width = msg.Width
		s.Height = msg.Height
	}

	// Update current input
	if s.FocusIndex >= 0 && s.FocusIndex < len(s.Fields) && s.Fields[s.FocusIndex].Enabled {
		var cmd tea.Cmd
		s.Fields[s.FocusIndex].input, cmd = s.Fields[s.FocusIndex].input.Update(msg)
		s.Fields[s.FocusIndex].Value = s.Fields[s.FocusIndex].input.Value()
		return s, cmd, ActionNone
	}

	return s, nil, ActionNone
}

func (s *ConfigScreen) nextField() {
	s.Fields[s.FocusIndex].input.Blur()

	for {
		s.FocusIndex++
		if s.FocusIndex >= len(s.Fields) {
			s.FocusIndex = 0
		}
		if s.Fields[s.FocusIndex].Enabled {
			break
		}
	}

	s.Fields[s.FocusIndex].input.Focus()
}

func (s *ConfigScreen) prevField() {
	s.Fields[s.FocusIndex].input.Blur()

	for {
		s.FocusIndex--
		if s.FocusIndex < 0 {
			s.FocusIndex = len(s.Fields) - 1
		}
		if s.Fields[s.FocusIndex].Enabled {
			break
		}
	}

	s.Fields[s.FocusIndex].input.Focus()
}

func (s *ConfigScreen) validate() bool {
	s.Errors = make(map[int]string)
	valid := true

	for i, field := range s.Fields {
		if !field.Enabled {
			continue
		}

		value := field.input.Value()

		if field.Required && value == "" {
			s.Errors[i] = "This field is required"
			valid = false
			continue
		}

		// Field-specific validation
		switch field.Key {
		case "tailscale_key":
			if value != "" && !strings.HasPrefix(value, "tskey-") {
				s.Errors[i] = "Should start with 'tskey-'"
				valid = false
			}
		case "code_password", "sudo_password":
			if field.Required && len(value) < 8 {
				s.Errors[i] = "Must be at least 8 characters"
				valid = false
			}
		case "anthropic_key":
			if value != "" && !strings.HasPrefix(value, "sk-ant-") {
				s.Errors[i] = "Should start with 'sk-ant-'"
				valid = false
			}
		}
	}

	return valid
}

// View renders the screen
func (s ConfigScreen) View() string {
	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	lightGreen := lipgloss.Color("#4A7C34")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
	darkGray := lipgloss.Color("#666666")
	red := lipgloss.Color("#FF6B6B")

	titleStyle := lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		Bold(true)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(lightGreen).
		Bold(true)

	disabledStyle := lipgloss.NewStyle().
		Foreground(darkGray)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		Italic(true).
		PaddingLeft(2)

	errorStyle := lipgloss.NewStyle().
		Foreground(red)

	requiredStyle := lipgloss.NewStyle().
		Foreground(red).
		Bold(true)

	noteStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(1)

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("Configuration"))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render("Enter your configuration values:"))
	sb.WriteString("\n\n")

	for i, field := range s.Fields {
		if !field.Enabled {
			continue
		}

		// Label
		label := field.Label
		if field.Required {
			label += requiredStyle.Render(" *")
		}

		style := labelStyle
		if i == s.FocusIndex {
			style = focusedLabelStyle
		}
		if !field.Enabled {
			style = disabledStyle
		}

		sb.WriteString("  " + style.Render(label))
		sb.WriteString("\n")

		// Input
		sb.WriteString("  " + field.input.View())
		sb.WriteString("\n")

		// Error
		if err, exists := s.Errors[i]; exists && err != "" {
			sb.WriteString("  " + errorStyle.Render("⚠ "+err))
			sb.WriteString("\n")
		}

		// Help
		sb.WriteString(helpStyle.Render(field.Help))
		sb.WriteString("\n\n")
	}

	// Notes
	if s.Mode == ModeDockerTailscale {
		sb.WriteString(noteStyle.Render("💡 Tip: Press Ctrl+V to toggle password visibility"))
	}

	sb.WriteString("\n")
	sb.WriteString(noteStyle.Render("[Tab/↓] Next field  [Shift+Tab/↑] Previous  [Enter] Continue  [Esc] Back"))

	return sb.String()
}

// GetValue returns the value of a field by key
func (s ConfigScreen) GetValue(key string) string {
	for _, field := range s.Fields {
		if field.Key == key {
			return field.input.Value()
		}
	}
	return ""
}

// GetAllValues returns all field values
func (s ConfigScreen) GetAllValues() map[string]string {
	values := make(map[string]string)
	for _, field := range s.Fields {
		if field.Enabled {
			values[field.Key] = field.input.Value()
		}
	}
	return values
}

// UpdateForMode updates field availability based on deployment mode
func (s *ConfigScreen) UpdateForMode(mode DeploymentMode) {
	s.Mode = mode

	for i := range s.Fields {
		switch s.Fields[i].Key {
		case "tailscale_key":
			s.Fields[i].Enabled = mode == ModeDockerTailscale
			s.Fields[i].Required = mode == ModeDockerTailscale
		case "code_password", "sudo_password", "anthropic_key":
			s.Fields[i].Enabled = mode != ModeTerminalOnly
			if s.Fields[i].Key != "anthropic_key" {
				s.Fields[i].Required = mode != ModeTerminalOnly
			}
		}
	}

	// Re-focus first enabled field
	for i, f := range s.Fields {
		if f.Enabled {
			s.FocusIndex = i
			s.Fields[i].input.Focus()
			break
		}
	}
}
