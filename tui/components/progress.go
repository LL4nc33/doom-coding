package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressStep represents a single step in the progress
type ProgressStep struct {
	Name        string
	Description string
	Status      StepStatus
	Output      string
	StartTime   time.Time
	EndTime     time.Time
	Error       error
}

// StepStatus represents the status of a step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepComplete
	StepFailed
	StepSkipped
)

// Progress is a component for showing multi-step progress
type Progress struct {
	Title       string
	Steps       []ProgressStep
	CurrentStep int
	LogLines    []string
	MaxLogLines int
	Width       int

	spinner spinner.Model

	// Styles
	TitleStyle       lipgloss.Style
	StepStyle        lipgloss.Style
	RunningStyle     lipgloss.Style
	CompleteStyle    lipgloss.Style
	FailedStyle      lipgloss.Style
	SkippedStyle     lipgloss.Style
	PendingStyle     lipgloss.Style
	LogStyle         lipgloss.Style
	ProgressBarStyle lipgloss.Style
}

// NewProgress creates a new progress component
func NewProgress(title string, steps []ProgressStep) Progress {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A7C34"))

	return Progress{
		Title:       title,
		Steps:       steps,
		CurrentStep: 0,
		LogLines:    make([]string, 0),
		MaxLogLines: 8,
		Width:       60,
		spinner:     s,
		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2E521D")).
			Bold(true).
			MarginBottom(1),
		StepStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")),
		RunningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A7C34")).
			Bold(true),
		CompleteStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#69DB7C")),
		FailedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")),
		SkippedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		PendingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		LogStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		ProgressBarStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A7C34")),
	}
}

// Init initializes the component
func (p Progress) Init() tea.Cmd {
	return p.spinner.Tick
}

// Update handles messages
func (p Progress) Update(msg tea.Msg) (Progress, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		return p, cmd
	}
	return p, nil
}

// View renders the progress component
func (p Progress) View() string {
	var sb strings.Builder

	// Title
	sb.WriteString(p.TitleStyle.Render(p.Title))
	sb.WriteString("\n\n")

	// Progress bar
	sb.WriteString(p.renderProgressBar())
	sb.WriteString("\n\n")

	// Steps
	for i, step := range p.Steps {
		icon := p.getStepIcon(step.Status, i == p.CurrentStep)
		style := p.getStepStyle(step.Status, i == p.CurrentStep)

		// Step line
		stepText := fmt.Sprintf("%s %s", icon, step.Name)
		if step.Description != "" {
			stepText += " - " + step.Description
		}
		sb.WriteString(style.Render(stepText))

		// Duration for completed steps
		if step.Status == StepComplete || step.Status == StepFailed {
			duration := step.EndTime.Sub(step.StartTime)
			sb.WriteString(p.PendingStyle.Render(fmt.Sprintf(" (%.1fs)", duration.Seconds())))
		}

		sb.WriteString("\n")
	}

	// Log output
	if len(p.LogLines) > 0 {
		sb.WriteString("\n")
		sb.WriteString(p.StepStyle.Render("Output:"))
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("─", p.Width))
		sb.WriteString("\n")
		for _, line := range p.LogLines {
			// Truncate long lines
			if len(line) > p.Width-2 {
				line = line[:p.Width-5] + "..."
			}
			sb.WriteString(p.LogStyle.Render(line))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (p Progress) renderProgressBar() string {
	// Count completed steps
	completed := 0
	for _, step := range p.Steps {
		if step.Status == StepComplete || step.Status == StepSkipped {
			completed++
		}
	}

	percent := float64(completed) / float64(len(p.Steps))
	barWidth := p.Width - 10

	filled := int(percent * float64(barWidth))
	empty := barWidth - filled

	bar := p.ProgressBarStyle.Render(strings.Repeat("█", filled))
	bar += p.PendingStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("[%s] %3.0f%%", bar, percent*100)
}

func (p Progress) getStepIcon(status StepStatus, isCurrent bool) string {
	switch status {
	case StepComplete:
		return p.CompleteStyle.Render("✓")
	case StepFailed:
		return p.FailedStyle.Render("✗")
	case StepSkipped:
		return p.SkippedStyle.Render("○")
	case StepRunning:
		return p.spinner.View()
	default:
		if isCurrent {
			return p.spinner.View()
		}
		return p.PendingStyle.Render("○")
	}
}

func (p Progress) getStepStyle(status StepStatus, isCurrent bool) lipgloss.Style {
	switch status {
	case StepComplete:
		return p.CompleteStyle
	case StepFailed:
		return p.FailedStyle
	case StepSkipped:
		return p.SkippedStyle
	case StepRunning:
		return p.RunningStyle
	default:
		if isCurrent {
			return p.RunningStyle
		}
		return p.PendingStyle
	}
}

// SetCurrentStep sets the current step
func (p *Progress) SetCurrentStep(index int) {
	if index >= 0 && index < len(p.Steps) {
		// Mark previous step as complete if it was running
		if p.CurrentStep < len(p.Steps) && p.Steps[p.CurrentStep].Status == StepRunning {
			p.Steps[p.CurrentStep].Status = StepComplete
			p.Steps[p.CurrentStep].EndTime = time.Now()
		}

		p.CurrentStep = index
		p.Steps[index].Status = StepRunning
		p.Steps[index].StartTime = time.Now()
	}
}

// CompleteStep marks a step as complete
func (p *Progress) CompleteStep(index int) {
	if index >= 0 && index < len(p.Steps) {
		p.Steps[index].Status = StepComplete
		p.Steps[index].EndTime = time.Now()
	}
}

// FailStep marks a step as failed
func (p *Progress) FailStep(index int, err error) {
	if index >= 0 && index < len(p.Steps) {
		p.Steps[index].Status = StepFailed
		p.Steps[index].Error = err
		p.Steps[index].EndTime = time.Now()
	}
}

// SkipStep marks a step as skipped
func (p *Progress) SkipStep(index int) {
	if index >= 0 && index < len(p.Steps) {
		p.Steps[index].Status = StepSkipped
	}
}

// AddLogLine adds a line to the log output
func (p *Progress) AddLogLine(line string) {
	p.LogLines = append(p.LogLines, line)
	// Keep only the last MaxLogLines
	if len(p.LogLines) > p.MaxLogLines {
		p.LogLines = p.LogLines[len(p.LogLines)-p.MaxLogLines:]
	}
}

// SetOutput sets the output for a step
func (p *Progress) SetOutput(index int, output string) {
	if index >= 0 && index < len(p.Steps) {
		p.Steps[index].Output = output
	}
}

// GetCompletedCount returns the number of completed steps
func (p Progress) GetCompletedCount() int {
	count := 0
	for _, step := range p.Steps {
		if step.Status == StepComplete {
			count++
		}
	}
	return count
}

// GetFailedCount returns the number of failed steps
func (p Progress) GetFailedCount() int {
	count := 0
	for _, step := range p.Steps {
		if step.Status == StepFailed {
			count++
		}
	}
	return count
}

// IsComplete returns true if all steps are done
func (p Progress) IsComplete() bool {
	for _, step := range p.Steps {
		if step.Status != StepComplete && step.Status != StepFailed && step.Status != StepSkipped {
			return false
		}
	}
	return true
}

// HasFailed returns true if any step has failed
func (p Progress) HasFailed() bool {
	return p.GetFailedCount() > 0
}
