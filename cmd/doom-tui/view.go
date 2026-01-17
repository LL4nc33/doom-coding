package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current screen
func (m Model) View() string {
	switch m.screen {
	case ScreenWelcome:
		return m.viewWelcome()
	case ScreenDetection:
		return m.viewDetection()
	case ScreenDeploymentMode:
		return m.viewDeploymentMode()
	case ScreenComponents:
		return m.viewComponents()
	case ScreenConfiguration:
		return m.viewConfiguration()
	case ScreenPreview:
		return m.viewPreview()
	case ScreenProgress:
		return m.viewProgress()
	case ScreenResults:
		return m.viewResults()
	default:
		return "Unknown screen"
	}
}

func (m Model) viewWelcome() string {
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

	bannerStyle := lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true)

	title := titleStyle.Render("Interactive Setup Wizard")
	version := subtitleStyle.Render(fmt.Sprintf("Version %s", Version))

	description := normalStyle.Render(`
Welcome to the Doom Coding setup wizard!

This interactive tool will guide you through setting up your
development environment with:

  • Docker containers for code-server and Claude Code
  • Tailscale VPN for secure remote access
  • Terminal tools (zsh, tmux, nvm, pyenv)
  • SSH hardening and security configuration
  • Secrets management with SOPS/age encryption
`)

	help := helpStyle.Render("[Enter] Continue  [h] Help  [q] Quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		bannerStyle.Render(banner),
		"",
		title,
		version,
		description,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) viewDetection() string {
	title := titleStyle.Render("System Detection")

	var content string
	if m.detecting {
		content = fmt.Sprintf("%s Detecting system configuration...", m.spinner.View())
	} else {
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("  %-20s %s\n", "Hostname:", m.systemInfo.Hostname))
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", "Username:", m.systemInfo.Username))
		sb.WriteString(fmt.Sprintf("  %-20s %s/%s\n", "Platform:", m.systemInfo.OS, m.systemInfo.Arch))
		sb.WriteString("\n")

		// Container detection
		containerType := "Bare Metal / VM"
		if m.systemInfo.IsLXC {
			containerType = "LXC Container"
		}
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", "Environment:", containerType))

		// TUN device
		tunStatus := successStyle.Render("✓ Available")
		tunNote := ""
		if !m.systemInfo.HasTUN {
			tunStatus = warningStyle.Render("✗ Not Available")
			tunNote = "\n  └─ Tailscale VPN requires TUN device. Local network mode recommended."
		}
		sb.WriteString(fmt.Sprintf("  %-20s %s%s\n", "TUN Device:", tunStatus, tunNote))

		// Docker
		dockerStatus := successStyle.Render("✓ Installed")
		if !m.systemInfo.DockerExists {
			dockerStatus = normalStyle.Render("○ Will be installed")
		}
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", "Docker:", dockerStatus))

		// Tailscale
		if m.systemInfo.TailscaleUp {
			sb.WriteString(fmt.Sprintf("  %-20s %s\n", "Tailscale:", successStyle.Render("✓ Connected")))
		}

		content = sb.String()
	}

	box := boxStyle.Render(content)

	var help string
	if m.detecting {
		help = helpStyle.Render("Detecting...")
	} else {
		help = helpStyle.Render("[Enter] Continue  [r] Re-detect  [Esc] Back")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		box,
		"",
		help,
	)
}

func (m Model) viewDeploymentMode() string {
	title := titleStyle.Render("Deployment Mode")
	subtitle := subtitleStyle.Render("Select how you want to deploy Doom Coding:")

	modes := []struct {
		icon        string
		name        string
		description string
		note        string
	}{
		{
			icon:        "🌐",
			name:        "Docker + Tailscale",
			description: "Full deployment with Tailscale container",
			note:        "Recommended for new Tailscale setups",
		},
		{
			icon:        "🏠",
			name:        "Docker + Local Network",
			description: "Containers accessible on local network only",
			note:        "Best for LXC without TUN device",
		},
		{
			icon:        "🔗",
			name:        "Docker + Host Tailscale",
			description: "Use existing Tailscale on host",
			note:        "Best when Tailscale is already running",
		},
		{
			icon:        "⚡",
			name:        "Terminal Tools Only",
			description: "Minimal installation (~200MB RAM)",
			note:        "No containers, just CLI tools",
		},
	}

	var options strings.Builder
	for i, mode := range modes {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = selectedStyle.Render("▸ ")
			style = selectedStyle
		}

		options.WriteString(fmt.Sprintf("%s%s %s\n", cursor, mode.icon, style.Render(mode.name)))
		options.WriteString(fmt.Sprintf("     %s\n", disabledStyle.Render(mode.description)))
		options.WriteString(fmt.Sprintf("     %s\n\n", helpStyle.Render(mode.note)))
	}

	// Recommendation based on system detection
	var recommendation string
	if !m.systemInfo.HasTUN && m.systemInfo.IsLXC {
		recommendation = warningStyle.Render("\n⚠ TUN device not available - Local Network mode recommended")
	}

	help := helpStyle.Render("[↑/↓] Navigate  [Enter] Select  [Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		subtitle,
		"",
		options.String(),
		recommendation,
		"",
		help,
	)
}

func (m Model) viewComponents() string {
	title := titleStyle.Render("Component Selection")
	subtitle := subtitleStyle.Render("Choose which components to install:")

	var options strings.Builder
	for i, comp := range m.components {
		cursor := "  "
		checkbox := "[ ]"
		style := normalStyle

		if !comp.Enabled {
			style = disabledStyle
			checkbox = "[-]"
		} else if comp.Selected {
			checkbox = selectedStyle.Render("[✓]")
		}

		if i == m.cursor {
			cursor = selectedStyle.Render("▸ ")
		}

		options.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, style.Render(comp.Name)))
		options.WriteString(fmt.Sprintf("       %s\n\n", disabledStyle.Render(comp.Description)))
	}

	help := helpStyle.Render("[↑/↓] Navigate  [Space] Toggle  [Enter] Continue  [Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		subtitle,
		"",
		options.String(),
		"",
		help,
	)
}

func (m Model) viewConfiguration() string {
	title := titleStyle.Render("Configuration")
	subtitle := subtitleStyle.Render("Enter your configuration values:")

	labels := []string{
		"Tailscale Auth Key:",
		"code-server Password:",
		"Sudo Password:",
		"Anthropic API Key:",
		"Timezone:",
		"Workspace Path:",
	}

	hints := []string{
		"Get from https://login.tailscale.com/admin/settings/keys",
		"Password for web IDE access",
		"Password for sudo in containers",
		"Get from https://console.anthropic.com",
		"e.g., Europe/Berlin, America/New_York",
		"Path to shared workspace directory",
	}

	var form strings.Builder
	for i, label := range labels {
		style := normalStyle
		if i == m.focusIndex {
			style = selectedStyle
		}

		form.WriteString(fmt.Sprintf("  %s\n", style.Render(label)))
		form.WriteString(fmt.Sprintf("  %s\n", m.inputs[i].View()))
		form.WriteString(fmt.Sprintf("  %s\n\n", disabledStyle.Render(hints[i])))
	}

	// Skip Tailscale key if not needed
	var note string
	if m.deploymentMode == ModeDockerLocal || !m.components[1].Selected {
		note = helpStyle.Render("Note: Tailscale auth key not required for local network mode")
	}

	help := helpStyle.Render("[Tab/↓] Next field  [Shift+Tab/↑] Previous  [Enter] Continue  [Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		subtitle,
		"",
		form.String(),
		note,
		"",
		help,
	)
}

func (m Model) viewPreview() string {
	title := titleStyle.Render("Installation Preview")
	subtitle := subtitleStyle.Render("Review your configuration before installing:")

	// Deployment summary
	var deployMode string
	switch m.deploymentMode {
	case ModeDockerTailscale:
		deployMode = "Docker + Tailscale (VPN access)"
	case ModeNativeTailscale:
		deployMode = "Docker + Host Tailscale"
	case ModeDockerLocal:
		deployMode = "Docker + Local Network"
	case ModeTerminalOnly:
		deployMode = "Terminal Tools Only"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("  %-20s %s\n", "Deployment:", selectedStyle.Render(deployMode)))
	summary.WriteString("\n  Components:\n")

	for _, comp := range m.components {
		status := disabledStyle.Render("○ Skip")
		if comp.Selected {
			status = successStyle.Render("✓ Install")
		}
		if !comp.Enabled {
			status = disabledStyle.Render("- N/A")
		}
		summary.WriteString(fmt.Sprintf("    %s %s\n", status, comp.Name))
	}

	summary.WriteString("\n  Configuration:\n")
	summary.WriteString(fmt.Sprintf("    %-18s %s\n", "Timezone:", m.config.Timezone))
	summary.WriteString(fmt.Sprintf("    %-18s %s\n", "Workspace:", m.config.WorkspacePath))

	if m.config.TailscaleKey != "" {
		summary.WriteString(fmt.Sprintf("    %-18s %s\n", "Tailscale Key:", "***configured***"))
	}
	if m.config.CodePassword != "" {
		summary.WriteString(fmt.Sprintf("    %-18s %s\n", "code-server:", "***configured***"))
	}
	if m.config.AnthropicKey != "" {
		summary.WriteString(fmt.Sprintf("    %-18s %s\n", "Anthropic API:", "***configured***"))
	}

	box := boxStyle.Render(summary.String())

	// Show equivalent command
	cmdPreview := helpStyle.Render(fmt.Sprintf("Equivalent command: scripts/install.sh --unattended %s",
		m.getBashFlags()))

	help := helpStyle.Render("[Enter/i] Install  [e] Export Config  [s] Show Command  [Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		subtitle,
		"",
		box,
		"",
		cmdPreview,
		"",
		help,
	)
}

func (m Model) viewProgress() string {
	title := titleStyle.Render("Installing...")

	// Progress bar
	progressPercent := float64(m.installStep) / float64(len(m.installSteps))
	progressBar := m.progress.ViewAs(progressPercent)

	// Current step
	currentStep := "Preparing..."
	if m.installStep > 0 && m.installStep <= len(m.installSteps) {
		currentStep = m.installSteps[m.installStep-1]
	}

	stepInfo := fmt.Sprintf("%s Step %d/%d: %s",
		m.spinner.View(),
		m.installStep,
		len(m.installSteps),
		currentStep,
	)

	// Output log
	var outputLog strings.Builder
	outputLog.WriteString("  Output:\n")
	outputLog.WriteString("  " + strings.Repeat("─", 50) + "\n")
	for _, line := range m.installOutput {
		outputLog.WriteString(fmt.Sprintf("  %s\n", line))
	}
	if len(m.installOutput) == 0 {
		outputLog.WriteString("  Waiting for output...\n")
	}

	help := helpStyle.Render("Installation in progress... Please wait.")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		progressBar,
		"",
		stepInfo,
		"",
		outputLog.String(),
		"",
		help,
	)
}

func (m Model) viewResults() string {
	var title string
	if m.installErr != nil {
		title = errorStyle.Render("Installation Failed")
	} else {
		title = successStyle.Render("Installation Complete!")
	}

	var content strings.Builder

	if m.installErr != nil {
		content.WriteString(fmt.Sprintf("  Error: %s\n\n", m.installErr.Error()))
		content.WriteString("  Troubleshooting:\n")
		content.WriteString("    • Check /var/log/doom-coding-install.log\n")
		content.WriteString("    • Verify network connectivity\n")
		content.WriteString("    • Ensure sufficient disk space\n")
	} else {
		content.WriteString("  Health Check Results:\n")
		for name, healthy := range m.healthResults {
			status := successStyle.Render("✓")
			if !healthy {
				status = errorStyle.Render("✗")
			}
			content.WriteString(fmt.Sprintf("    %s %s\n", status, name))
		}

		content.WriteString("\n  Access Information:\n")
		if m.deploymentMode == ModeDockerTailscale {
			content.WriteString("    • code-server: https://<tailscale-ip>:8443\n")
			content.WriteString("    • Run 'tailscale status' to get your IP\n")
		} else if m.deploymentMode == ModeNativeTailscale {
			content.WriteString("    • code-server: https://<host-tailscale-ip>:8443\n")
			content.WriteString("    • Run 'tailscale ip' to get your host's Tailscale IP\n")
		} else if m.deploymentMode == ModeDockerLocal {
			content.WriteString("    • code-server: https://localhost:8443\n")
			content.WriteString("    • Or use your machine's local IP\n")
		}

		content.WriteString("\n  Next Steps:\n")
		content.WriteString("    • Open code-server in your browser\n")
		content.WriteString("    • Connect to Claude Code container\n")
		content.WriteString("    • Check documentation at docs/README.md\n")
	}

	box := boxStyle.Render(content.String())

	help := helpStyle.Render("[Enter/q] Exit  [r] Re-run Health Check  [l] View Logs")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		box,
		"",
		help,
	)
}

func (m Model) getBashFlags() string {
	var flags []string

	if !m.components[0].Selected {
		flags = append(flags, "--skip-docker")
	}
	if !m.components[1].Selected || m.deploymentMode == ModeDockerLocal {
		flags = append(flags, "--skip-tailscale")
	}
	if m.deploymentMode == ModeNativeTailscale {
		flags = append(flags, "--native-tailscale")
	}
	if !m.components[2].Selected {
		flags = append(flags, "--skip-terminal")
	}
	if !m.components[3].Selected {
		flags = append(flags, "--skip-hardening")
	}
	if !m.components[4].Selected {
		flags = append(flags, "--skip-secrets")
	}

	return strings.Join(flags, " ")
}
