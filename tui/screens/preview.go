package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PreviewScreen shows a summary before installation
type PreviewScreen struct {
	Width        int
	Height       int
	Mode         DeploymentMode
	Components   []string
	Config       map[string]string
	EnvPreview   string
	BashCommand  string
}

// NewPreviewScreen creates a new preview screen
func NewPreviewScreen(mode DeploymentMode, components []string, config map[string]string) PreviewScreen {
	return PreviewScreen{
		Mode:       mode,
		Components: components,
		Config:     config,
	}
}

// Init initializes the screen
func (s PreviewScreen) Init() tea.Cmd {
	return nil
}

// Update handles input
func (s PreviewScreen) Update(msg tea.Msg) (PreviewScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "i":
			return s, nil, ActionSubmit
		case "e":
			// Export config
			return s, nil, ActionNone
		case "s":
			// Show bash command
			return s, nil, ActionNone
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

// View renders the screen
func (s PreviewScreen) View() string {
	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	lightGreen := lipgloss.Color("#4A7C34")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
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

	labelStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		Width(18)

	valueStyle := lipgloss.NewStyle().
		Foreground(white)

	checkStyle := lipgloss.NewStyle().
		Foreground(green)

	skipStyle := lipgloss.NewStyle().
		Foreground(gray)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(forestGreen).
		Padding(1, 2)

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("Installation Preview"))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render("Review your configuration before installing:"))
	sb.WriteString("\n\n")

	// Deployment mode
	modeStr := s.getModeString()
	sb.WriteString(fmt.Sprintf("  %s%s\n\n", labelStyle.Render("Deployment:"), selectedStyle.Render(modeStr)))

	// Components
	sb.WriteString("  Components:\n")
	allComponents := []string{"docker", "tailscale", "terminal", "hardening", "secrets"}
	componentNames := map[string]string{
		"docker":    "Docker",
		"tailscale": "Tailscale",
		"terminal":  "Terminal Tools",
		"hardening": "SSH Hardening",
		"secrets":   "Secrets Management",
	}

	for _, comp := range allComponents {
		name := componentNames[comp]
		if s.hasComponent(comp) {
			sb.WriteString(fmt.Sprintf("    %s %s\n", checkStyle.Render("✓ Install"), valueStyle.Render(name)))
		} else {
			sb.WriteString(fmt.Sprintf("    %s %s\n", skipStyle.Render("○ Skip"), skipStyle.Render(name)))
		}
	}

	sb.WriteString("\n  Configuration:\n")

	// Show non-sensitive config
	if tz, ok := s.Config["timezone"]; ok && tz != "" {
		sb.WriteString(fmt.Sprintf("    %s%s\n", labelStyle.Render("Timezone:"), valueStyle.Render(tz)))
	}
	if ws, ok := s.Config["workspace"]; ok && ws != "" {
		sb.WriteString(fmt.Sprintf("    %s%s\n", labelStyle.Render("Workspace:"), valueStyle.Render(ws)))
	}

	// Show configured credentials (masked)
	if _, ok := s.Config["tailscale_key"]; ok && s.Config["tailscale_key"] != "" {
		sb.WriteString(fmt.Sprintf("    %s%s\n", labelStyle.Render("Tailscale Key:"), checkStyle.Render("✓ configured")))
	}
	if _, ok := s.Config["code_password"]; ok && s.Config["code_password"] != "" {
		sb.WriteString(fmt.Sprintf("    %s%s\n", labelStyle.Render("code-server:"), checkStyle.Render("✓ configured")))
	}
	if _, ok := s.Config["anthropic_key"]; ok && s.Config["anthropic_key"] != "" {
		sb.WriteString(fmt.Sprintf("    %s%s\n", labelStyle.Render("Anthropic API:"), checkStyle.Render("✓ configured")))
	}

	content := sb.String()
	box := boxStyle.Render(content)

	// Command preview
	cmdPreview := helpStyle.Render(fmt.Sprintf("Command: scripts/install.sh --unattended %s", s.getBashFlags()))

	help := helpStyle.Render("[Enter/i] Install  [e] Export Config  [s] Show Command  [Esc] Back  [q] Quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		box,
		"",
		cmdPreview,
		"",
		help,
	)
}

func (s PreviewScreen) getModeString() string {
	switch s.Mode {
	case ModeDockerTailscale:
		return "Docker + Tailscale (VPN access)"
	case ModeDockerLocal:
		return "Docker + Local Network"
	case ModeTerminalOnly:
		return "Terminal Tools Only"
	default:
		return "Unknown"
	}
}

func (s PreviewScreen) hasComponent(id string) bool {
	for _, comp := range s.Components {
		if comp == id {
			return true
		}
	}
	return false
}

func (s PreviewScreen) getBashFlags() string {
	var flags []string

	if !s.hasComponent("docker") {
		flags = append(flags, "--skip-docker")
	}
	if !s.hasComponent("tailscale") || s.Mode == ModeDockerLocal {
		flags = append(flags, "--skip-tailscale")
	}
	if !s.hasComponent("terminal") {
		flags = append(flags, "--skip-terminal")
	}
	if !s.hasComponent("hardening") {
		flags = append(flags, "--skip-hardening")
	}
	if !s.hasComponent("secrets") {
		flags = append(flags, "--skip-secrets")
	}

	return strings.Join(flags, " ")
}

// SetEnvPreview sets the .env file preview
func (s *PreviewScreen) SetEnvPreview(preview string) {
	s.EnvPreview = preview
}

// SetBashCommand sets the equivalent bash command
func (s *PreviewScreen) SetBashCommand(cmd string) {
	s.BashCommand = cmd
}

// GetBashFlags returns the bash flags for install.sh
func (s PreviewScreen) GetBashFlags() []string {
	var flags []string

	flags = append(flags, "--unattended")

	if !s.hasComponent("docker") {
		flags = append(flags, "--skip-docker")
	}
	if !s.hasComponent("tailscale") || s.Mode == ModeDockerLocal {
		flags = append(flags, "--skip-tailscale")
	}
	if !s.hasComponent("terminal") {
		flags = append(flags, "--skip-terminal")
	}
	if !s.hasComponent("hardening") {
		flags = append(flags, "--skip-hardening")
	}
	if !s.hasComponent("secrets") {
		flags = append(flags, "--skip-secrets")
	}

	if key := s.Config["tailscale_key"]; key != "" {
		flags = append(flags, fmt.Sprintf("--tailscale-key=%s", key))
	}
	if pwd := s.Config["code_password"]; pwd != "" {
		flags = append(flags, fmt.Sprintf("--code-password=%s", pwd))
	}
	if key := s.Config["anthropic_key"]; key != "" {
		flags = append(flags, fmt.Sprintf("--anthropic-key=%s", key))
	}

	return flags
}
