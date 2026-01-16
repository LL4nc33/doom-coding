package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WelcomeScreen is the initial welcome screen
type WelcomeScreen struct {
	Width   int
	Height  int
	Version string
}

// NewWelcomeScreen creates a new welcome screen
func NewWelcomeScreen(version string) WelcomeScreen {
	return WelcomeScreen{
		Version: version,
	}
}

// Init initializes the screen
func (s WelcomeScreen) Init() tea.Cmd {
	return nil
}

// Update handles input
func (s WelcomeScreen) Update(msg tea.Msg) (WelcomeScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return s, nil, ActionNext
		case "h", "?":
			return s, nil, ActionHelp
		case "q", "ctrl+c":
			return s, tea.Quit, ActionQuit
		}
	case tea.WindowSizeMsg:
		s.Width = msg.Width
		s.Height = msg.Height
	}
	return s, nil, ActionNone
}

// View renders the screen
func (s WelcomeScreen) View() string {
	banner := `
    ██████╗  ██████╗  ██████╗ ███╗   ███╗
    ██╔══██╗██╔═══██╗██╔═══██╗████╗ ████║
    ██║  ██║██║   ██║██║   ██║██╔████╔██║
    ██║  ██║██║   ██║██║   ██║██║╚██╔╝██║
    ██████╔╝╚██████╔╝╚██████╔╝██║ ╚═╝ ██║
    ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝     ╚═╝
     ██████╗ ██████╗ ██████╗ ██╗███╗   ██╗ ██████╗
    ██╔════╝██╔═══██╗██╔══██╗██║████╗  ██║██╔════╝
    ██║     ██║   ██║██║  ██║██║██╔██╗ ██║██║  ███╗
    ██║     ██║   ██║██║  ██║██║██║╚██╗██║██║   ██║
    ╚██████╗╚██████╔╝██████╔╝██║██║ ╚████║╚██████╔╝
     ╚═════╝ ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝
`

	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")

	bannerStyle := lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true)

	titleStyle := lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(tanBrown)

	descStyle := lipgloss.NewStyle().
		Foreground(white)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(2)

	description := `Welcome to the Doom Coding setup wizard!

This interactive tool will guide you through setting up your
development environment with:

  • Docker containers for code-server and Claude Code
  • Tailscale VPN for secure remote access
  • Terminal tools (zsh, tmux, nvm, pyenv)
  • SSH hardening and security configuration
  • Secrets management with SOPS/age encryption`

	content := lipgloss.JoinVertical(lipgloss.Center,
		bannerStyle.Render(banner),
		"",
		titleStyle.Render("Interactive Setup Wizard"),
		subtitleStyle.Render(fmt.Sprintf("Version %s", s.Version)),
		"",
		descStyle.Render(description),
		helpStyle.Render("[Enter] Continue  [h] Help  [q] Quit"),
	)

	return lipgloss.Place(s.Width, s.Height, lipgloss.Center, lipgloss.Center, content)
}

// ScreenAction represents navigation actions
type ScreenAction int

const (
	ActionNone ScreenAction = iota
	ActionNext
	ActionBack
	ActionQuit
	ActionHelp
	ActionSubmit
	ActionRefresh
)
