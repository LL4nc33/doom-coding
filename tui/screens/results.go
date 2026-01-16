package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HealthResult represents a health check result
type HealthResult struct {
	Name    string
	Status  bool
	Message string
}

// ResultsScreen shows installation results
type ResultsScreen struct {
	Width        int
	Height       int
	Success      bool
	Error        error
	HealthChecks []HealthResult
	Mode         DeploymentMode
	TailscaleIP  string
	LocalIPs     []string
}

// NewResultsScreen creates a new results screen
func NewResultsScreen(success bool, err error) ResultsScreen {
	return ResultsScreen{
		Success: success,
		Error:   err,
	}
}

// Init initializes the screen
func (s ResultsScreen) Init() tea.Cmd {
	return nil
}

// Update handles input
func (s ResultsScreen) Update(msg tea.Msg) (ResultsScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "q":
			return s, tea.Quit, ActionQuit
		case "r":
			return s, nil, ActionRefresh
		case "l":
			// View logs
			return s, nil, ActionNone
		}
	case tea.WindowSizeMsg:
		s.Width = msg.Width
		s.Height = msg.Height
	}
	return s, nil, ActionNone
}

// View renders the screen
func (s ResultsScreen) View() string {
	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
	green := lipgloss.Color("#69DB7C")
	red := lipgloss.Color("#FF6B6B")
	yellow := lipgloss.Color("#FFD93D")
	blue := lipgloss.Color("#4DABF7")

	successTitleStyle := lipgloss.NewStyle().
		Foreground(green).
		Bold(true)

	errorTitleStyle := lipgloss.NewStyle().
		Foreground(red).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(white)

	successStyle := lipgloss.NewStyle().
		Foreground(green)

	errorStyle := lipgloss.NewStyle().
		Foreground(red)

	warningStyle := lipgloss.NewStyle().
		Foreground(yellow)

	linkStyle := lipgloss.NewStyle().
		Foreground(blue).
		Underline(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(forestGreen).
		Padding(1, 2)

	var sb strings.Builder

	sb.WriteString("\n")

	if s.Success {
		sb.WriteString(successTitleStyle.Render("🎉 Installation Complete!"))
	} else {
		sb.WriteString(errorTitleStyle.Render("❌ Installation Failed"))
	}
	sb.WriteString("\n\n")

	if !s.Success && s.Error != nil {
		// Error details
		sb.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", s.Error.Error())))
		sb.WriteString("\n\n")

		sb.WriteString(labelStyle.Render("Troubleshooting:"))
		sb.WriteString("\n")
		sb.WriteString(valueStyle.Render("  • Check /var/log/doom-coding-install.log for details"))
		sb.WriteString("\n")
		sb.WriteString(valueStyle.Render("  • Verify network connectivity"))
		sb.WriteString("\n")
		sb.WriteString(valueStyle.Render("  • Ensure sufficient disk space (10GB+ recommended)"))
		sb.WriteString("\n")
		sb.WriteString(valueStyle.Render("  • Run with --verbose flag for more details"))
		sb.WriteString("\n\n")
	} else {
		// Health check results
		sb.WriteString(labelStyle.Render("Health Check Results:"))
		sb.WriteString("\n")

		for _, check := range s.HealthChecks {
			icon := successStyle.Render("✓")
			if !check.Status {
				icon = errorStyle.Render("✗")
			}
			sb.WriteString(fmt.Sprintf("  %s %s", icon, valueStyle.Render(check.Name)))
			if check.Message != "" {
				sb.WriteString(fmt.Sprintf(" - %s", gray.Render(check.Message)))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")

		// Access information
		sb.WriteString(labelStyle.Render("Access Information:"))
		sb.WriteString("\n")

		switch s.Mode {
		case ModeDockerTailscale:
			if s.TailscaleIP != "" {
				sb.WriteString(fmt.Sprintf("  code-server: %s\n", linkStyle.Render(fmt.Sprintf("https://%s:8443", s.TailscaleIP))))
			} else {
				sb.WriteString(valueStyle.Render("  • Run 'tailscale status' to get your Tailscale IP"))
				sb.WriteString("\n")
				sb.WriteString(valueStyle.Render("  • Access code-server at https://<tailscale-ip>:8443"))
				sb.WriteString("\n")
			}
		case ModeDockerLocal:
			sb.WriteString(fmt.Sprintf("  code-server: %s\n", linkStyle.Render("https://localhost:8443")))
			if len(s.LocalIPs) > 0 {
				sb.WriteString(fmt.Sprintf("  Local network: %s\n", linkStyle.Render(fmt.Sprintf("https://%s:8443", s.LocalIPs[0]))))
			}
		case ModeTerminalOnly:
			sb.WriteString(valueStyle.Render("  Terminal tools are ready to use"))
			sb.WriteString("\n")
			sb.WriteString(valueStyle.Render("  • Start a new shell or run 'source ~/.zshrc'"))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")

		// Next steps
		sb.WriteString(labelStyle.Render("Next Steps:"))
		sb.WriteString("\n")

		switch s.Mode {
		case ModeDockerTailscale, ModeDockerLocal:
			sb.WriteString(valueStyle.Render("  1. Open code-server in your browser"))
			sb.WriteString("\n")
			sb.WriteString(valueStyle.Render("  2. Enter your code-server password"))
			sb.WriteString("\n")
			sb.WriteString(valueStyle.Render("  3. Connect to Claude Code: docker exec -it doom-claude zsh"))
			sb.WriteString("\n")
		case ModeTerminalOnly:
			sb.WriteString(valueStyle.Render("  1. Start a new terminal session"))
			sb.WriteString("\n")
			sb.WriteString(valueStyle.Render("  2. Run 'tmux' to start a multiplexer session"))
			sb.WriteString("\n")
			sb.WriteString(valueStyle.Render("  3. Use 'nvm' and 'pyenv' for version management"))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")

		// Documentation
		sb.WriteString(labelStyle.Render("Documentation:"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  %s\n", valueStyle.Render("• See docs/README.md for full documentation")))
		sb.WriteString(fmt.Sprintf("  %s\n", valueStyle.Render("• Run 'doom-tui status' for health checks")))
		sb.WriteString(fmt.Sprintf("  %s\n", valueStyle.Render("• View logs: docker logs doom-code-server")))

		// Warnings
		warnings := s.getWarnings()
		if len(warnings) > 0 {
			sb.WriteString("\n")
			sb.WriteString(warningStyle.Render("Warnings:"))
			sb.WriteString("\n")
			for _, w := range warnings {
				sb.WriteString(warningStyle.Render(fmt.Sprintf("  ⚠ %s", w)))
				sb.WriteString("\n")
			}
		}
	}

	content := sb.String()
	box := boxStyle.Render(content)

	help := helpStyle.Render("[Enter/q] Exit  [r] Re-run Health Check  [l] View Logs")

	return lipgloss.JoinVertical(lipgloss.Left,
		box,
		"",
		help,
	)
}

func (s ResultsScreen) getWarnings() []string {
	var warnings []string

	for _, check := range s.HealthChecks {
		if !check.Status && check.Message != "" {
			warnings = append(warnings, check.Message)
		}
	}

	return warnings
}

// SetHealthChecks sets the health check results
func (s *ResultsScreen) SetHealthChecks(checks []HealthResult) {
	s.HealthChecks = checks
}

// SetAccessInfo sets the access information
func (s *ResultsScreen) SetAccessInfo(mode DeploymentMode, tailscaleIP string, localIPs []string) {
	s.Mode = mode
	s.TailscaleIP = tailscaleIP
	s.LocalIPs = localIPs
}

// AddHealthCheck adds a health check result
func (s *ResultsScreen) AddHealthCheck(name string, status bool, message string) {
	s.HealthChecks = append(s.HealthChecks, HealthResult{
		Name:    name,
		Status:  status,
		Message: message,
	})
}
