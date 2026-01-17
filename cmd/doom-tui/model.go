package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen identifiers
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenDetection
	ScreenDeploymentMode
	ScreenComponents
	ScreenConfiguration
	ScreenPreview
	ScreenProgress
	ScreenResults
)

// DeploymentMode options
type DeploymentMode int

const (
	ModeDockerTailscale DeploymentMode = iota
	ModeDockerLocal
	ModeNativeTailscale
	ModeTerminalOnly
)

// Component selection
type Component struct {
	Name        string
	Description string
	Selected    bool
	Enabled     bool // Can be disabled based on dependencies
}

// SystemInfo holds detected system information
type SystemInfo struct {
	OS           string
	Arch         string
	IsLXC        bool
	HasTUN       bool
	DockerExists bool
	TailscaleUp  bool
	Hostname     string
	Username     string
}

// Configuration holds user input values
type Configuration struct {
	TailscaleKey   string
	CodePassword   string
	SudoPassword   string
	AnthropicKey   string
	Timezone       string
	WorkspacePath  string
	PUID           string
	PGID           string
}

// Model is the main application state
type Model struct {
	projectRoot   string
	screen        Screen
	width         int
	height        int

	// System detection
	systemInfo    SystemInfo
	detecting     bool

	// User selections
	deploymentMode DeploymentMode
	components     []Component
	config         Configuration

	// UI components
	spinner        spinner.Model
	progress       progress.Model
	inputs         []textinput.Model
	focusIndex     int
	cursor         int

	// Installation state
	installing     bool
	installStep    int
	installSteps   []string
	installOutput  []string
	installErr     error

	// Results
	healthResults  map[string]bool
	accessInfo     []string
}

// Brand colors
var (
	forestGreen = lipgloss.Color("#2E521D")
	tanBrown    = lipgloss.Color("#7C5E46")
	lightGreen  = lipgloss.Color("#4A7C34")
	cream       = lipgloss.Color("#F5F5DC")

	titleStyle = lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true).
		MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(tanBrown).
		MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
		Foreground(lightGreen).
		Bold(true)

	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	disabledStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginTop(1)

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(forestGreen).
		Padding(1, 2)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000"))

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA00"))
)

// NewModel creates a new TUI model
func NewModel(projectRoot string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(forestGreen)

	p := progress.New(progress.WithDefaultGradient())

	// Initialize text inputs for configuration
	inputs := make([]textinput.Model, 6)

	// Tailscale auth key
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "tskey-auth-xxxxxxxxxxxxx"
	inputs[0].CharLimit = 100
	inputs[0].Width = 50

	// code-server password
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "your-secure-password"
	inputs[1].CharLimit = 50
	inputs[1].Width = 50
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '*'

	// Sudo password
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "container-sudo-password"
	inputs[2].CharLimit = 50
	inputs[2].Width = 50
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '*'

	// Anthropic API key
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "sk-ant-api03-xxxxxxxxxxxxx"
	inputs[3].CharLimit = 150
	inputs[3].Width = 50
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '*'

	// Timezone
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Europe/Berlin"
	inputs[4].SetValue("Europe/Berlin")
	inputs[4].CharLimit = 50
	inputs[4].Width = 50

	// Workspace path
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "./workspace"
	inputs[5].SetValue("./workspace")
	inputs[5].CharLimit = 100
	inputs[5].Width = 50

	inputs[0].Focus()

	return Model{
		projectRoot:    projectRoot,
		screen:         ScreenWelcome,
		spinner:        s,
		progress:       p,
		inputs:         inputs,
		deploymentMode: ModeDockerTailscale,
		components: []Component{
			{Name: "Docker", Description: "Container runtime for services", Selected: true, Enabled: true},
			{Name: "Tailscale", Description: "Secure VPN access", Selected: true, Enabled: true},
			{Name: "Terminal Tools", Description: "zsh, tmux, nvm, pyenv", Selected: true, Enabled: true},
			{Name: "SSH Hardening", Description: "Security configuration", Selected: true, Enabled: true},
			{Name: "Secrets Management", Description: "SOPS/age encryption", Selected: true, Enabled: true},
		},
		config: Configuration{
			Timezone:      "Europe/Berlin",
			WorkspacePath: "./workspace",
			PUID:          "1000",
			PGID:          "1000",
		},
		installSteps: []string{
			"Checking system requirements",
			"Installing base packages",
			"Setting up Docker",
			"Configuring network",
			"Installing terminal tools",
			"Applying security hardening",
			"Setting up secrets management",
			"Creating environment file",
			"Starting services",
			"Running health checks",
		},
		healthResults: make(map[string]bool),
	}
}

// Messages
type (
	detectionDoneMsg struct{ info SystemInfo }
	tickMsg          struct{}
	installStepMsg   struct{ step int; output string }
	installDoneMsg   struct{ err error }
	healthCheckMsg   struct{ results map[string]bool }
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.EnterAltScreen,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case detectionDoneMsg:
		m.systemInfo = msg.info
		m.detecting = false
		return m, nil

	case installStepMsg:
		m.installStep = msg.step
		m.installOutput = append(m.installOutput, msg.output)
		// Keep only last 10 lines
		if len(m.installOutput) > 10 {
			m.installOutput = m.installOutput[len(m.installOutput)-10:]
		}
		return m, nil

	case installDoneMsg:
		m.installing = false
		m.installErr = msg.err
		if msg.err == nil {
			m.screen = ScreenResults
			return m, m.runHealthCheck()
		}
		return m, nil

	case healthCheckMsg:
		m.healthResults = msg.results
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	// Update text inputs if on configuration screen
	if m.screen == ScreenConfiguration {
		return m.updateInputs(msg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.installing {
			return m, nil // Don't quit during installation
		}
		return m, tea.Quit

	case "esc":
		if m.screen > ScreenWelcome && !m.installing {
			m.screen--
			return m, nil
		}
	}

	switch m.screen {
	case ScreenWelcome:
		return m.handleWelcomeKeys(msg)
	case ScreenDetection:
		return m.handleDetectionKeys(msg)
	case ScreenDeploymentMode:
		return m.handleDeploymentKeys(msg)
	case ScreenComponents:
		return m.handleComponentKeys(msg)
	case ScreenConfiguration:
		return m.handleConfigKeys(msg)
	case ScreenPreview:
		return m.handlePreviewKeys(msg)
	case ScreenProgress:
		return m.handleProgressKeys(msg)
	case ScreenResults:
		return m.handleResultsKeys(msg)
	}

	return m, nil
}

func (m Model) handleWelcomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		m.screen = ScreenDetection
		m.detecting = true
		return m, tea.Batch(m.spinner.Tick, m.detectSystem())
	case "h":
		// Show help
		return m, nil
	}
	return m, nil
}

func (m Model) handleDetectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.detecting {
		return m, nil
	}
	switch msg.String() {
	case "enter", " ":
		m.screen = ScreenDeploymentMode
		m.updateComponentsForMode()
		return m, nil
	case "r":
		m.detecting = true
		return m, tea.Batch(m.spinner.Tick, m.detectSystem())
	}
	return m, nil
}

func (m Model) handleDeploymentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 3 {
			m.cursor++
		}
	case "enter", " ":
		m.deploymentMode = DeploymentMode(m.cursor)
		m.updateComponentsForMode()
		m.screen = ScreenComponents
		m.cursor = 0
	}
	return m, nil
}

func (m Model) handleComponentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.components)-1 {
			m.cursor++
		}
	case " ":
		if m.components[m.cursor].Enabled {
			m.components[m.cursor].Selected = !m.components[m.cursor].Selected
		}
	case "enter":
		m.screen = ScreenConfiguration
		m.cursor = 0
		m.focusIndex = 0
		m.inputs[0].Focus()
	}
	return m, nil
}

func (m Model) handleConfigKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		m.focusIndex++
		if m.focusIndex >= len(m.inputs) {
			m.focusIndex = 0
		}
		return m.focusInput()
	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.inputs) - 1
		}
		return m.focusInput()
	case "enter":
		if m.focusIndex == len(m.inputs)-1 {
			m.saveInputs()
			m.screen = ScreenPreview
			return m, nil
		}
		m.focusIndex++
		if m.focusIndex >= len(m.inputs) {
			m.focusIndex = 0
		}
		return m.focusInput()
	}
	return m, nil
}

func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "i":
		m.screen = ScreenProgress
		m.installing = true
		return m, tea.Batch(m.spinner.Tick, m.runInstallation())
	case "e":
		// Export config
		return m, nil
	case "s":
		// Show bash command
		return m, nil
	}
	return m, nil
}

func (m Model) handleProgressKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Limited interaction during installation
	return m, nil
}

func (m Model) handleResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "q":
		return m, tea.Quit
	case "r":
		return m, m.runHealthCheck()
	case "l":
		// View logs
		return m, nil
	}
	return m, nil
}

func (m *Model) updateComponentsForMode() {
	switch m.deploymentMode {
	case ModeDockerTailscale:
		m.components[0].Selected = true // Docker
		m.components[0].Enabled = true
		m.components[1].Selected = true // Tailscale
		m.components[1].Enabled = true
	case ModeNativeTailscale:
		m.components[0].Selected = true  // Docker
		m.components[0].Enabled = true
		m.components[1].Selected = false // Tailscale (uses host Tailscale)
		m.components[1].Enabled = false
	case ModeDockerLocal:
		m.components[0].Selected = true // Docker
		m.components[0].Enabled = true
		m.components[1].Selected = false // Tailscale
		m.components[1].Enabled = false
	case ModeTerminalOnly:
		m.components[0].Selected = false // Docker
		m.components[0].Enabled = false
		m.components[1].Selected = false // Tailscale
		m.components[1].Enabled = false
	}
}

func (m Model) focusInput() (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) updateInputs(msg tea.Msg) (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) saveInputs() {
	m.config.TailscaleKey = m.inputs[0].Value()
	m.config.CodePassword = m.inputs[1].Value()
	m.config.SudoPassword = m.inputs[2].Value()
	m.config.AnthropicKey = m.inputs[3].Value()
	m.config.Timezone = m.inputs[4].Value()
	m.config.WorkspacePath = m.inputs[5].Value()
}

// Commands
func (m Model) detectSystem() tea.Cmd {
	return func() tea.Msg {
		info := SystemInfo{
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
			Username: os.Getenv("USER"),
		}

		// Get hostname
		if hostname, err := os.Hostname(); err == nil {
			info.Hostname = hostname
		}

		// Check if running in LXC
		if _, err := os.Stat("/dev/lxc"); err == nil {
			info.IsLXC = true
		}
		// Also check cgroup for LXC
		if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
			if strings.Contains(string(data), "lxc") {
				info.IsLXC = true
			}
		}

		// Check TUN device
		if _, err := os.Stat("/dev/net/tun"); err == nil {
			info.HasTUN = true
		}

		// Check Docker
		if _, err := exec.LookPath("docker"); err == nil {
			info.DockerExists = true
		}

		// Check Tailscale
		if cmd := exec.Command("tailscale", "status", "--json"); cmd.Run() == nil {
			info.TailscaleUp = true
		}

		return detectionDoneMsg{info: info}
	}
}

func (m Model) runInstallation() tea.Cmd {
	return func() tea.Msg {
		// Build command arguments
		args := []string{filepath.Join(m.projectRoot, "scripts", "install.sh"), "--unattended"}

		// Add component flags
		if !m.components[0].Selected {
			args = append(args, "--skip-docker")
		}
		if !m.components[1].Selected || m.deploymentMode == ModeDockerLocal {
			args = append(args, "--skip-tailscale")
		}
		if m.deploymentMode == ModeNativeTailscale {
			args = append(args, "--native-tailscale")
		}
		if !m.components[2].Selected {
			args = append(args, "--skip-terminal")
		}
		if !m.components[3].Selected {
			args = append(args, "--skip-hardening")
		}
		if !m.components[4].Selected {
			args = append(args, "--skip-secrets")
		}

		// Add credentials
		if m.config.TailscaleKey != "" {
			args = append(args, fmt.Sprintf("--tailscale-key=%s", m.config.TailscaleKey))
		}
		if m.config.CodePassword != "" {
			args = append(args, fmt.Sprintf("--code-password=%s", m.config.CodePassword))
		}
		if m.config.AnthropicKey != "" {
			args = append(args, fmt.Sprintf("--anthropic-key=%s", m.config.AnthropicKey))
		}

		// Write .env file first
		envContent := m.generateEnvFile()
		envPath := filepath.Join(m.projectRoot, ".env")
		if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
			return installDoneMsg{err: err}
		}

		// Run installation
		cmd := exec.Command("bash", args...)
		cmd.Dir = m.projectRoot

		if err := cmd.Run(); err != nil {
			return installDoneMsg{err: err}
		}

		return installDoneMsg{err: nil}
	}
}

func (m Model) runHealthCheck() tea.Cmd {
	return func() tea.Msg {
		results := make(map[string]bool)

		healthScript := filepath.Join(m.projectRoot, "scripts", "health-check.sh")
		cmd := exec.Command("bash", healthScript, "--json")
		output, err := cmd.Output()

		if err == nil {
			// Parse JSON output (simplified)
			results["Docker"] = strings.Contains(string(output), `"docker": "running"`) ||
				strings.Contains(string(output), `"docker":true`)
			results["Containers"] = strings.Contains(string(output), `"healthy"`)
			results["Terminal"] = strings.Contains(string(output), `"zsh":true`) ||
				strings.Contains(string(output), `"zsh": "installed"`)
		} else {
			// Fallback to basic checks
			if _, err := exec.LookPath("docker"); err == nil {
				results["Docker"] = true
			}
			if _, err := exec.LookPath("zsh"); err == nil {
				results["Terminal"] = true
			}
		}

		return healthCheckMsg{results: results}
	}
}

func (m Model) generateEnvFile() string {
	var sb strings.Builder

	sb.WriteString("# Doom Coding Environment Configuration\n")
	sb.WriteString("# Generated by doom-tui\n\n")

	sb.WriteString("# Tailscale Authentication\n")
	if m.config.TailscaleKey != "" {
		sb.WriteString(fmt.Sprintf("TS_AUTHKEY=%s\n", m.config.TailscaleKey))
	} else {
		sb.WriteString("# TS_AUTHKEY=tskey-auth-xxxxx\n")
	}
	sb.WriteString("\n")

	sb.WriteString("# code-server Configuration\n")
	sb.WriteString(fmt.Sprintf("CODE_SERVER_PASSWORD=%s\n", m.config.CodePassword))
	sb.WriteString(fmt.Sprintf("SUDO_PASSWORD=%s\n", m.config.SudoPassword))
	sb.WriteString("\n")

	sb.WriteString("# User Settings\n")
	sb.WriteString(fmt.Sprintf("PUID=%s\n", m.config.PUID))
	sb.WriteString(fmt.Sprintf("PGID=%s\n", m.config.PGID))
	sb.WriteString(fmt.Sprintf("TZ=%s\n", m.config.Timezone))
	sb.WriteString("\n")

	sb.WriteString("# Paths\n")
	sb.WriteString(fmt.Sprintf("WORKSPACE_PATH=%s\n", m.config.WorkspacePath))
	sb.WriteString("\n")

	sb.WriteString("# Architecture (auto-detected)\n")
	sb.WriteString(fmt.Sprintf("TARGETARCH=%s\n", runtime.GOARCH))
	sb.WriteString("\n")

	sb.WriteString("# Tailscale Options\n")
	sb.WriteString("TS_ACCEPT_DNS=false\n")
	sb.WriteString("# TS_EXTRA_ARGS=--advertise-tags=tag:doom-coding\n")
	sb.WriteString("\n")

	sb.WriteString("# Advanced\n")
	sb.WriteString("CLAUDE_AUTOMATION=--dangerously-skip-permissions\n")
	sb.WriteString("CODE_SERVER_PORT=8443\n")

	return sb.String()
}
