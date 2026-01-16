package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SystemInfo contains detected system information
type SystemInfo struct {
	Hostname     string
	Username     string
	OS           string
	Arch         string
	Distribution string
	Version      string
	IsLXC        bool
	IsDocker     bool
	IsWSL        bool
	HasTUN       bool
	DockerExists bool
	DockerRunning bool
	TailscaleUp  bool
	TailscaleIP  string
	LocalIPs     []string
	DiskFreeGB   float64
	MemoryGB     float64
}

// DetectionMsg is sent when detection completes
type DetectionMsg struct {
	Info SystemInfo
	Err  error
}

// DetectionScreen shows system detection results
type DetectionScreen struct {
	Width     int
	Height    int
	Detecting bool
	Info      SystemInfo
	Error     error
	spinner   spinner.Model
}

// NewDetectionScreen creates a new detection screen
func NewDetectionScreen() DetectionScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A7C34"))

	return DetectionScreen{
		Detecting: true,
		spinner:   s,
	}
}

// Init initializes the screen
func (s DetectionScreen) Init() tea.Cmd {
	return s.spinner.Tick
}

// Update handles input and messages
func (s DetectionScreen) Update(msg tea.Msg) (DetectionScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !s.Detecting {
			switch msg.String() {
			case "enter", " ":
				return s, nil, ActionNext
			case "r":
				s.Detecting = true
				return s, s.spinner.Tick, ActionRefresh
			case "esc":
				return s, nil, ActionBack
			case "q", "ctrl+c":
				return s, tea.Quit, ActionQuit
			}
		}

	case tea.WindowSizeMsg:
		s.Width = msg.Width
		s.Height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd, ActionNone

	case DetectionMsg:
		s.Detecting = false
		s.Info = msg.Info
		s.Error = msg.Err
	}

	return s, nil, ActionNone
}

// View renders the screen
func (s DetectionScreen) View() string {
	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
	green := lipgloss.Color("#69DB7C")
	yellow := lipgloss.Color("#FFD93D")
	red := lipgloss.Color("#FF6B6B")

	titleStyle := lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(tanBrown).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Foreground(white)

	successStyle := lipgloss.NewStyle().
		Foreground(green)

	warningStyle := lipgloss.NewStyle().
		Foreground(yellow)

	errorStyle := lipgloss.NewStyle().
		Foreground(red)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(forestGreen).
		Padding(1, 2)

	title := titleStyle.Render("System Detection")

	var content string
	if s.Detecting {
		content = fmt.Sprintf("%s Detecting system configuration...", s.spinner.View())
	} else if s.Error != nil {
		content = errorStyle.Render(fmt.Sprintf("Detection failed: %v", s.Error))
	} else {
		var lines string

		// Basic info
		lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Hostname:"), valueStyle.Render(s.Info.Hostname))
		lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Username:"), valueStyle.Render(s.Info.Username))
		lines += fmt.Sprintf("%s%s/%s", labelStyle.Render("Platform:"), valueStyle.Render(s.Info.OS), valueStyle.Render(s.Info.Arch))
		if s.Info.Distribution != "" {
			lines += fmt.Sprintf(" (%s", s.Info.Distribution)
			if s.Info.Version != "" {
				lines += " " + s.Info.Version
			}
			lines += ")"
		}
		lines += "\n\n"

		// Environment
		envType := "Bare Metal / VM"
		if s.Info.IsLXC {
			envType = "LXC Container"
		} else if s.Info.IsDocker {
			envType = "Docker Container"
		} else if s.Info.IsWSL {
			envType = "Windows Subsystem for Linux"
		}
		lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Environment:"), valueStyle.Render(envType))

		// TUN device
		tunStatus := successStyle.Render("✓ Available")
		if !s.Info.HasTUN {
			tunStatus = warningStyle.Render("✗ Not Available")
		}
		lines += fmt.Sprintf("%s%s\n", labelStyle.Render("TUN Device:"), tunStatus)
		if !s.Info.HasTUN && s.Info.IsLXC {
			lines += "                    " + warningStyle.Render("└─ Tailscale requires TUN. Local network mode recommended.") + "\n"
		}

		// Docker
		var dockerStatus string
		if s.Info.DockerExists {
			if s.Info.DockerRunning {
				dockerStatus = successStyle.Render("✓ Running")
			} else {
				dockerStatus = warningStyle.Render("✓ Installed (not running)")
			}
		} else {
			dockerStatus = valueStyle.Render("○ Will be installed")
		}
		lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Docker:"), dockerStatus)

		// Tailscale
		if s.Info.TailscaleUp {
			lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Tailscale:"), successStyle.Render("✓ Connected"))
			if s.Info.TailscaleIP != "" {
				lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Tailscale IP:"), valueStyle.Render(s.Info.TailscaleIP))
			}
		}

		// Network IPs
		if len(s.Info.LocalIPs) > 0 {
			lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Local IPs:"), valueStyle.Render(s.Info.LocalIPs[0]))
			for i := 1; i < len(s.Info.LocalIPs) && i < 3; i++ {
				lines += fmt.Sprintf("                    %s\n", valueStyle.Render(s.Info.LocalIPs[i]))
			}
		}

		// Resources
		if s.Info.DiskFreeGB > 0 {
			diskStyle := successStyle
			if s.Info.DiskFreeGB < 10 {
				diskStyle = warningStyle
			} else if s.Info.DiskFreeGB < 5 {
				diskStyle = errorStyle
			}
			lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Disk Free:"), diskStyle.Render(fmt.Sprintf("%.1f GB", s.Info.DiskFreeGB)))
		}

		if s.Info.MemoryGB > 0 {
			memStyle := successStyle
			if s.Info.MemoryGB < 2 {
				memStyle = warningStyle
			} else if s.Info.MemoryGB < 1 {
				memStyle = errorStyle
			}
			lines += fmt.Sprintf("%s%s\n", labelStyle.Render("Total Memory:"), memStyle.Render(fmt.Sprintf("%.1f GB", s.Info.MemoryGB)))
		}

		content = lines
	}

	box := boxStyle.Render(content)

	var help string
	if s.Detecting {
		help = helpStyle.Render("Detecting...")
	} else {
		help = helpStyle.Render("[Enter] Continue  [r] Re-detect  [Esc] Back  [q] Quit")
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

// SetInfo sets the system info
func (s *DetectionScreen) SetInfo(info SystemInfo) {
	s.Info = info
	s.Detecting = false
}

// StartDetecting starts the detection spinner
func (s *DetectionScreen) StartDetecting() {
	s.Detecting = true
}
