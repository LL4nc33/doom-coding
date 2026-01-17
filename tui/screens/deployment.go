package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DeploymentMode represents the deployment type
type DeploymentMode int

const (
	ModeDockerTailscale DeploymentMode = iota
	ModeDockerLocal
	ModeNativeTailscale
	ModeTerminalOnly
)

// String returns the string representation of the mode
func (m DeploymentMode) String() string {
	switch m {
	case ModeDockerTailscale:
		return "tailscale"
	case ModeDockerLocal:
		return "local"
	case ModeNativeTailscale:
		return "native-tailscale"
	case ModeTerminalOnly:
		return "terminal-only"
	default:
		return "unknown"
	}
}

// DeploymentOption represents a deployment choice
type DeploymentOption struct {
	Mode        DeploymentMode
	Icon        string
	Name        string
	Description string
	Hint        string
	Recommended bool
	Enabled     bool
}

// DeploymentScreen allows selecting the deployment mode
type DeploymentScreen struct {
	Width       int
	Height      int
	Options     []DeploymentOption
	Cursor      int
	Selected    DeploymentMode
	Recommended DeploymentMode
}

// NewDeploymentScreen creates a new deployment screen
func NewDeploymentScreen(hasTUN bool, isLXC bool, hostTailscaleRunning bool) DeploymentScreen {
	recommended := ModeDockerTailscale
	if hostTailscaleRunning {
		recommended = ModeNativeTailscale
	} else if !hasTUN && isLXC {
		recommended = ModeDockerLocal
	}

	options := []DeploymentOption{
		{
			Mode:        ModeDockerTailscale,
			Icon:        "🌐",
			Name:        "Docker + Tailscale",
			Description: "Full deployment with Tailscale container for VPN access",
			Hint:        "Recommended for new Tailscale setups",
			Recommended: recommended == ModeDockerTailscale,
			Enabled:     hasTUN || !isLXC,
		},
		{
			Mode:        ModeNativeTailscale,
			Icon:        "🔗",
			Name:        "Docker + Host Tailscale",
			Description: "Use existing Tailscale on host, no TUN device needed",
			Hint:        "Best when Tailscale is already running on host",
			Recommended: recommended == ModeNativeTailscale,
			Enabled:     hostTailscaleRunning,
		},
		{
			Mode:        ModeDockerLocal,
			Icon:        "🏠",
			Name:        "Docker + Local Network",
			Description: "Containers accessible on local network only",
			Hint:        "Best for LXC without TUN device or home lab setups",
			Recommended: recommended == ModeDockerLocal,
			Enabled:     true,
		},
		{
			Mode:        ModeTerminalOnly,
			Icon:        "⚡",
			Name:        "Terminal Tools Only",
			Description: "Minimal installation with CLI tools (~200MB RAM)",
			Hint:        "No containers, just zsh, tmux, nvm, pyenv",
			Recommended: false,
			Enabled:     true,
		},
	}

	// Set cursor to recommended option
	cursor := 0
	for i, opt := range options {
		if opt.Recommended {
			cursor = i
			break
		}
	}

	return DeploymentScreen{
		Options:     options,
		Cursor:      cursor,
		Selected:    recommended,
		Recommended: recommended,
	}
}

// Init initializes the screen
func (s DeploymentScreen) Init() tea.Cmd {
	return nil
}

// Update handles input
func (s DeploymentScreen) Update(msg tea.Msg) (DeploymentScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			s.moveCursor(-1)
		case "down", "j":
			s.moveCursor(1)
		case "enter", " ":
			if s.Options[s.Cursor].Enabled {
				s.Selected = s.Options[s.Cursor].Mode
				return s, nil, ActionNext
			}
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

func (s *DeploymentScreen) moveCursor(delta int) {
	s.Cursor += delta
	if s.Cursor < 0 {
		s.Cursor = len(s.Options) - 1
	}
	if s.Cursor >= len(s.Options) {
		s.Cursor = 0
	}

	// Skip disabled options
	attempts := 0
	for !s.Options[s.Cursor].Enabled && attempts < len(s.Options) {
		s.Cursor += delta
		if s.Cursor < 0 {
			s.Cursor = len(s.Options) - 1
		}
		if s.Cursor >= len(s.Options) {
			s.Cursor = 0
		}
		attempts++
	}
}

// View renders the screen
func (s DeploymentScreen) View() string {
	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	lightGreen := lipgloss.Color("#4A7C34")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
	darkGray := lipgloss.Color("#666666")
	yellow := lipgloss.Color("#FFD93D")

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
		PaddingLeft(5)

	hintStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		Italic(true).
		PaddingLeft(5)

	recommendedStyle := lipgloss.NewStyle().
		Foreground(yellow).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(1)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lightGreen).
		Bold(true)

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("Deployment Mode"))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render("Select how you want to deploy Doom Coding:"))
	sb.WriteString("\n\n")

	for i, opt := range s.Options {
		cursor := "  "
		if i == s.Cursor {
			cursor = cursorStyle.Render("▸ ")
		}

		var radio string
		var nameStyle lipgloss.Style

		if !opt.Enabled {
			radio = disabledStyle.Render("( )")
			nameStyle = disabledStyle
		} else if i == s.Cursor {
			radio = selectedStyle.Render("(●)")
			nameStyle = selectedStyle
		} else {
			radio = normalStyle.Render("( )")
			nameStyle = normalStyle
		}

		name := opt.Name
		if opt.Recommended {
			name += " " + recommendedStyle.Render("(Recommended)")
		}

		sb.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, radio, opt.Icon, nameStyle.Render(name)))
		sb.WriteString(descStyle.Render(opt.Description))
		sb.WriteString("\n")
		sb.WriteString(hintStyle.Render(opt.Hint))
		sb.WriteString("\n\n")
	}

	// Warning for disabled options
	if !s.Options[0].Enabled {
		sb.WriteString(lipgloss.NewStyle().Foreground(yellow).Render("⚠ Tailscale mode unavailable - TUN device not detected"))
		sb.WriteString("\n\n")
	}

	sb.WriteString(helpStyle.Render("[↑/↓] Navigate  [Enter] Select  [Esc] Back  [q] Quit"))

	return sb.String()
}

// GetSelectedMode returns the selected deployment mode
func (s DeploymentScreen) GetSelectedMode() DeploymentMode {
	return s.Selected
}
