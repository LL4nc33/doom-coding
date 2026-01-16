package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Component represents an installable component
type Component struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Selected    bool
	Enabled     bool
	Required    bool
}

// ComponentScreen allows selecting components to install
type ComponentScreen struct {
	Width      int
	Height     int
	Components []Component
	Cursor     int
	Mode       DeploymentMode
}

// NewComponentScreen creates a new component selection screen
func NewComponentScreen(mode DeploymentMode) ComponentScreen {
	components := []Component{
		{
			ID:          "docker",
			Name:        "Docker",
			Description: "Container runtime for running services",
			Icon:        "🐳",
			Selected:    mode != ModeTerminalOnly,
			Enabled:     mode != ModeTerminalOnly,
			Required:    mode != ModeTerminalOnly,
		},
		{
			ID:          "tailscale",
			Name:        "Tailscale",
			Description: "Secure VPN for remote access",
			Icon:        "🔐",
			Selected:    mode == ModeDockerTailscale,
			Enabled:     mode == ModeDockerTailscale,
			Required:    mode == ModeDockerTailscale,
		},
		{
			ID:          "terminal",
			Name:        "Terminal Tools",
			Description: "zsh, tmux, Oh My Zsh, nvm, pyenv",
			Icon:        "💻",
			Selected:    true,
			Enabled:     true,
			Required:    false,
		},
		{
			ID:          "hardening",
			Name:        "SSH Hardening",
			Description: "Security configuration and fail2ban",
			Icon:        "🔒",
			Selected:    true,
			Enabled:     true,
			Required:    false,
		},
		{
			ID:          "secrets",
			Name:        "Secrets Management",
			Description: "SOPS/age encryption for credentials",
			Icon:        "🔑",
			Selected:    true,
			Enabled:     true,
			Required:    false,
		},
	}

	return ComponentScreen{
		Components: components,
		Cursor:     0,
		Mode:       mode,
	}
}

// Init initializes the screen
func (s ComponentScreen) Init() tea.Cmd {
	return nil
}

// Update handles input
func (s ComponentScreen) Update(msg tea.Msg) (ComponentScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			s.moveCursor(-1)
		case "down", "j":
			s.moveCursor(1)
		case " ", "x":
			s.toggle()
		case "a":
			s.selectAll()
		case "n":
			s.selectNone()
		case "enter":
			return s, nil, ActionNext
		case "esc":
			return s, nil, ActionBack
		case "q", "ctrl+c":
			return s, tea.Quit, ActionQuit
		}
	case tea.WindowSizeMsg:
		s.Width = msg.Width
		s.Height = msg.Height
	}
	return s, nil, ActionNone
}

func (s *ComponentScreen) moveCursor(delta int) {
	s.Cursor += delta
	if s.Cursor < 0 {
		s.Cursor = len(s.Components) - 1
	}
	if s.Cursor >= len(s.Components) {
		s.Cursor = 0
	}
}

func (s *ComponentScreen) toggle() {
	c := &s.Components[s.Cursor]
	if c.Enabled && !c.Required {
		c.Selected = !c.Selected
	}
}

func (s *ComponentScreen) selectAll() {
	for i := range s.Components {
		if s.Components[i].Enabled {
			s.Components[i].Selected = true
		}
	}
}

func (s *ComponentScreen) selectNone() {
	for i := range s.Components {
		if s.Components[i].Enabled && !s.Components[i].Required {
			s.Components[i].Selected = false
		}
	}
}

// View renders the screen
func (s ComponentScreen) View() string {
	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	lightGreen := lipgloss.Color("#4A7C34")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
	darkGray := lipgloss.Color("#666666")
	green := lipgloss.Color("#69DB7C")

	titleStyle := lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lightGreen).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(white)

	disabledStyle := lipgloss.NewStyle().
		Foreground(darkGray)

	descStyle := lipgloss.NewStyle().
		Foreground(gray).
		PaddingLeft(7)

	requiredStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		Italic(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(1)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lightGreen).
		Bold(true)

	checkStyle := lipgloss.NewStyle().
		Foreground(green)

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("Component Selection"))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render("Choose which components to install:"))
	sb.WriteString("\n\n")

	for i, comp := range s.Components {
		cursor := "  "
		if i == s.Cursor {
			cursor = cursorStyle.Render("▸ ")
		}

		var checkbox string
		var nameStyle lipgloss.Style

		if !comp.Enabled {
			checkbox = disabledStyle.Render("[-]")
			nameStyle = disabledStyle
		} else if comp.Selected {
			checkbox = checkStyle.Render("[✓]")
			if i == s.Cursor {
				nameStyle = selectedStyle
			} else {
				nameStyle = normalStyle
			}
		} else {
			checkbox = normalStyle.Render("[ ]")
			if i == s.Cursor {
				nameStyle = selectedStyle
			} else {
				nameStyle = normalStyle
			}
		}

		name := comp.Name
		if comp.Required {
			name += " " + requiredStyle.Render("(required)")
		}

		sb.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, checkbox, comp.Icon, nameStyle.Render(name)))
		sb.WriteString(descStyle.Render(comp.Description))
		sb.WriteString("\n\n")
	}

	sb.WriteString(helpStyle.Render("[↑/↓] Navigate  [Space] Toggle  [a] All  [n] None  [Enter] Continue  [Esc] Back"))

	return sb.String()
}

// GetSelectedComponents returns the list of selected component IDs
func (s ComponentScreen) GetSelectedComponents() []string {
	var selected []string
	for _, comp := range s.Components {
		if comp.Selected {
			selected = append(selected, comp.ID)
		}
	}
	return selected
}

// IsSelected returns whether a component is selected
func (s ComponentScreen) IsSelected(id string) bool {
	for _, comp := range s.Components {
		if comp.ID == id {
			return comp.Selected
		}
	}
	return false
}

// UpdateForMode updates component availability based on deployment mode
func (s *ComponentScreen) UpdateForMode(mode DeploymentMode) {
	s.Mode = mode

	for i := range s.Components {
		switch s.Components[i].ID {
		case "docker":
			s.Components[i].Enabled = mode != ModeTerminalOnly
			s.Components[i].Required = mode != ModeTerminalOnly
			s.Components[i].Selected = mode != ModeTerminalOnly
		case "tailscale":
			s.Components[i].Enabled = mode == ModeDockerTailscale
			s.Components[i].Required = mode == ModeDockerTailscale
			s.Components[i].Selected = mode == ModeDockerTailscale
		}
	}
}
