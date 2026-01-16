package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstallStep represents an installation step
type InstallStep struct {
	Name        string
	Description string
	Status      StepStatus
	StartTime   time.Time
	EndTime     time.Time
	Error       error
}

// StepStatus represents the status of a step
type StepStatus int

const (
	StatusPending StepStatus = iota
	StatusRunning
	StatusComplete
	StatusFailed
	StatusSkipped
)

// InstallProgressMsg is sent to update progress
type InstallProgressMsg struct {
	Step   int
	Output string
}

// InstallDoneMsg is sent when installation completes
type InstallDoneMsg struct {
	Success bool
	Error   error
}

// ProgressScreen shows installation progress
type ProgressScreen struct {
	Width       int
	Height      int
	Steps       []InstallStep
	CurrentStep int
	LogLines    []string
	MaxLogLines int
	Installing  bool
	Complete    bool
	Error       error

	spinner spinner.Model
}

// NewProgressScreen creates a new progress screen
func NewProgressScreen() ProgressScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A7C34"))

	steps := []InstallStep{
		{Name: "System Check", Description: "Verifying system requirements"},
		{Name: "Base Packages", Description: "Installing essential packages"},
		{Name: "Docker Setup", Description: "Installing and configuring Docker"},
		{Name: "Network Config", Description: "Configuring network settings"},
		{Name: "Terminal Tools", Description: "Setting up zsh, tmux, nvm, pyenv"},
		{Name: "SSH Hardening", Description: "Applying security configuration"},
		{Name: "Secrets Setup", Description: "Configuring SOPS/age encryption"},
		{Name: "Environment", Description: "Creating environment file"},
		{Name: "Services", Description: "Starting Docker containers"},
		{Name: "Health Check", Description: "Verifying installation"},
	}

	return ProgressScreen{
		Steps:       steps,
		CurrentStep: 0,
		LogLines:    make([]string, 0),
		MaxLogLines: 8,
		Installing:  false,
		spinner:     s,
	}
}

// Init initializes the screen
func (s ProgressScreen) Init() tea.Cmd {
	return s.spinner.Tick
}

// Update handles messages
func (s ProgressScreen) Update(msg tea.Msg) (ProgressScreen, tea.Cmd, ScreenAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Limited interaction during installation
		if s.Complete {
			switch msg.String() {
			case "enter", " ":
				return s, nil, ActionNext
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

	case InstallProgressMsg:
		s.updateProgress(msg.Step, msg.Output)

	case InstallDoneMsg:
		s.Installing = false
		s.Complete = true
		if !msg.Success {
			s.Error = msg.Error
			if s.CurrentStep < len(s.Steps) {
				s.Steps[s.CurrentStep].Status = StatusFailed
				s.Steps[s.CurrentStep].Error = msg.Error
				s.Steps[s.CurrentStep].EndTime = time.Now()
			}
		} else {
			// Mark remaining steps as complete
			for i := s.CurrentStep; i < len(s.Steps); i++ {
				if s.Steps[i].Status == StatusRunning || s.Steps[i].Status == StatusPending {
					s.Steps[i].Status = StatusComplete
					s.Steps[i].EndTime = time.Now()
				}
			}
		}
	}

	return s, nil, ActionNone
}

func (s *ProgressScreen) updateProgress(step int, output string) {
	// Update step status
	if step > 0 && step <= len(s.Steps) {
		// Complete previous step
		if s.CurrentStep > 0 && s.CurrentStep <= len(s.Steps) {
			prev := &s.Steps[s.CurrentStep-1]
			if prev.Status == StatusRunning {
				prev.Status = StatusComplete
				prev.EndTime = time.Now()
			}
		}

		s.CurrentStep = step
		current := &s.Steps[step-1]
		current.Status = StatusRunning
		if current.StartTime.IsZero() {
			current.StartTime = time.Now()
		}
	}

	// Add to log
	if output != "" {
		s.LogLines = append(s.LogLines, output)
		if len(s.LogLines) > s.MaxLogLines {
			s.LogLines = s.LogLines[len(s.LogLines)-s.MaxLogLines:]
		}
	}
}

// View renders the screen
func (s ProgressScreen) View() string {
	forestGreen := lipgloss.Color("#2E521D")
	tanBrown := lipgloss.Color("#7C5E46")
	lightGreen := lipgloss.Color("#4A7C34")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
	darkGray := lipgloss.Color("#666666")
	green := lipgloss.Color("#69DB7C")
	red := lipgloss.Color("#FF6B6B")

	titleStyle := lipgloss.NewStyle().
		Foreground(forestGreen).
		Bold(true)

	stepStyle := lipgloss.NewStyle().
		Foreground(white)

	runningStyle := lipgloss.NewStyle().
		Foreground(lightGreen).
		Bold(true)

	completeStyle := lipgloss.NewStyle().
		Foreground(green)

	failedStyle := lipgloss.NewStyle().
		Foreground(red)

	pendingStyle := lipgloss.NewStyle().
		Foreground(darkGray)

	skippedStyle := lipgloss.NewStyle().
		Foreground(gray)

	logStyle := lipgloss.NewStyle().
		Foreground(gray)

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		MarginTop(1)

	progressStyle := lipgloss.NewStyle().
		Foreground(lightGreen)

	var sb strings.Builder

	sb.WriteString("\n")

	title := "Installing..."
	if s.Complete {
		if s.Error != nil {
			title = "Installation Failed"
		} else {
			title = "Installation Complete!"
		}
	}
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	// Progress bar
	completed := 0
	for _, step := range s.Steps {
		if step.Status == StatusComplete || step.Status == StatusSkipped {
			completed++
		}
	}
	percent := float64(completed) / float64(len(s.Steps))
	barWidth := 50

	filled := int(percent * float64(barWidth))
	empty := barWidth - filled

	bar := progressStyle.Render(strings.Repeat("█", filled))
	bar += pendingStyle.Render(strings.Repeat("░", empty))

	sb.WriteString(fmt.Sprintf("[%s] %3.0f%%\n\n", bar, percent*100))

	// Steps
	for i, step := range s.Steps {
		icon := s.getStepIcon(step.Status, i == s.CurrentStep-1)
		style := s.getStepStyle(step.Status, i == s.CurrentStep-1)

		stepText := step.Name
		if step.Description != "" {
			stepText += " - " + step.Description
		}

		sb.WriteString(fmt.Sprintf("  %s %s", icon, style.Render(stepText)))

		// Duration for completed steps
		if (step.Status == StatusComplete || step.Status == StatusFailed) && !step.EndTime.IsZero() {
			duration := step.EndTime.Sub(step.StartTime)
			sb.WriteString(pendingStyle.Render(fmt.Sprintf(" (%.1fs)", duration.Seconds())))
		}

		// Error message for failed steps
		if step.Status == StatusFailed && step.Error != nil {
			sb.WriteString("\n")
			sb.WriteString(failedStyle.Render(fmt.Sprintf("      Error: %s", step.Error.Error())))
		}

		sb.WriteString("\n")
	}

	// Log output
	if len(s.LogLines) > 0 {
		sb.WriteString("\n")
		sb.WriteString(stepStyle.Render("Output:"))
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("─", 60))
		sb.WriteString("\n")
		for _, line := range s.LogLines {
			// Truncate long lines
			if len(line) > 58 {
				line = line[:55] + "..."
			}
			sb.WriteString(logStyle.Render("  " + line))
			sb.WriteString("\n")
		}
	}

	// Help
	sb.WriteString("\n")
	if s.Complete {
		sb.WriteString(helpStyle.Render("[Enter] Continue  [q] Quit"))
	} else {
		sb.WriteString(helpStyle.Render("Installation in progress... Please wait."))
	}

	return sb.String()
}

func (s ProgressScreen) getStepIcon(status StepStatus, isCurrent bool) string {
	forestGreen := lipgloss.Color("#4A7C34")
	green := lipgloss.Color("#69DB7C")
	red := lipgloss.Color("#FF6B6B")
	gray := lipgloss.Color("#888888")
	darkGray := lipgloss.Color("#666666")

	completeStyle := lipgloss.NewStyle().Foreground(green)
	failedStyle := lipgloss.NewStyle().Foreground(red)
	skippedStyle := lipgloss.NewStyle().Foreground(gray)
	pendingStyle := lipgloss.NewStyle().Foreground(darkGray)
	_ = lipgloss.NewStyle().Foreground(forestGreen)

	switch status {
	case StatusComplete:
		return completeStyle.Render("✓")
	case StatusFailed:
		return failedStyle.Render("✗")
	case StatusSkipped:
		return skippedStyle.Render("○")
	case StatusRunning:
		return s.spinner.View()
	default:
		if isCurrent {
			return s.spinner.View()
		}
		return pendingStyle.Render("○")
	}
}

func (s ProgressScreen) getStepStyle(status StepStatus, isCurrent bool) lipgloss.Style {
	lightGreen := lipgloss.Color("#4A7C34")
	white := lipgloss.Color("#FFFFFF")
	gray := lipgloss.Color("#888888")
	darkGray := lipgloss.Color("#666666")
	green := lipgloss.Color("#69DB7C")
	red := lipgloss.Color("#FF6B6B")

	runningStyle := lipgloss.NewStyle().Foreground(lightGreen).Bold(true)
	completeStyle := lipgloss.NewStyle().Foreground(green)
	failedStyle := lipgloss.NewStyle().Foreground(red)
	skippedStyle := lipgloss.NewStyle().Foreground(gray)
	pendingStyle := lipgloss.NewStyle().Foreground(darkGray)
	normalStyle := lipgloss.NewStyle().Foreground(white)
	_ = normalStyle

	switch status {
	case StatusComplete:
		return completeStyle
	case StatusFailed:
		return failedStyle
	case StatusSkipped:
		return skippedStyle
	case StatusRunning:
		return runningStyle
	default:
		if isCurrent {
			return runningStyle
		}
		return pendingStyle
	}
}

// Start starts the installation
func (s *ProgressScreen) Start() {
	s.Installing = true
	s.CurrentStep = 1
	s.Steps[0].Status = StatusRunning
	s.Steps[0].StartTime = time.Now()
}

// SetStep sets the current step
func (s *ProgressScreen) SetStep(step int) {
	if step > 0 && step <= len(s.Steps) {
		s.CurrentStep = step
		for i := 0; i < step-1; i++ {
			if s.Steps[i].Status == StatusPending || s.Steps[i].Status == StatusRunning {
				s.Steps[i].Status = StatusComplete
				if s.Steps[i].EndTime.IsZero() {
					s.Steps[i].EndTime = time.Now()
				}
			}
		}
		s.Steps[step-1].Status = StatusRunning
		if s.Steps[step-1].StartTime.IsZero() {
			s.Steps[step-1].StartTime = time.Now()
		}
	}
}

// AddLogLine adds a line to the output log
func (s *ProgressScreen) AddLogLine(line string) {
	s.LogLines = append(s.LogLines, line)
	if len(s.LogLines) > s.MaxLogLines {
		s.LogLines = s.LogLines[len(s.LogLines)-s.MaxLogLines:]
	}
}

// Complete marks installation as complete
func (s *ProgressScreen) Complete(success bool, err error) {
	s.Installing = false
	s.Complete = true
	s.Error = err

	if success {
		for i := range s.Steps {
			if s.Steps[i].Status == StatusRunning || s.Steps[i].Status == StatusPending {
				s.Steps[i].Status = StatusComplete
				if s.Steps[i].EndTime.IsZero() {
					s.Steps[i].EndTime = time.Now()
				}
			}
		}
	} else if s.CurrentStep > 0 && s.CurrentStep <= len(s.Steps) {
		s.Steps[s.CurrentStep-1].Status = StatusFailed
		s.Steps[s.CurrentStep-1].Error = err
		s.Steps[s.CurrentStep-1].EndTime = time.Now()
	}
}

// SkipStep marks a step as skipped
func (s *ProgressScreen) SkipStep(step int) {
	if step > 0 && step <= len(s.Steps) {
		s.Steps[step-1].Status = StatusSkipped
	}
}

// UpdateSteps updates which steps to show based on selected components
func (s *ProgressScreen) UpdateSteps(skipDocker, skipTailscale, skipTerminal, skipHardening, skipSecrets bool) {
	// Mark skipped steps
	if skipDocker {
		s.Steps[2].Status = StatusSkipped // Docker Setup
	}
	if skipTailscale {
		s.Steps[3].Status = StatusSkipped // Network Config (Tailscale)
	}
	if skipTerminal {
		s.Steps[4].Status = StatusSkipped // Terminal Tools
	}
	if skipHardening {
		s.Steps[5].Status = StatusSkipped // SSH Hardening
	}
	if skipSecrets {
		s.Steps[6].Status = StatusSkipped // Secrets Setup
	}
}
