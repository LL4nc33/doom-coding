package components

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

func TestNewProgress(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1", Description: "First step"},
		{Name: "Step 2", Description: "Second step"},
		{Name: "Step 3", Description: "Third step"},
	}

	p := NewProgress("Installation Progress", steps)

	if p.Title != "Installation Progress" {
		t.Errorf("Expected title 'Installation Progress', got %q", p.Title)
	}

	if len(p.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(p.Steps))
	}

	if p.CurrentStep != 0 {
		t.Errorf("Expected CurrentStep at 0, got %d", p.CurrentStep)
	}

	if p.MaxLogLines != 8 {
		t.Errorf("Expected MaxLogLines=8, got %d", p.MaxLogLines)
	}

	if p.Width != 60 {
		t.Errorf("Expected Width=60, got %d", p.Width)
	}

	if len(p.LogLines) != 0 {
		t.Error("LogLines should be empty initially")
	}
}

func TestProgressInit(t *testing.T) {
	p := NewProgress("Test", []ProgressStep{})
	cmd := p.Init()

	// Init should return spinner tick command
	if cmd == nil {
		t.Error("Init should return spinner tick command")
	}
}

func TestProgressUpdate(t *testing.T) {
	p := NewProgress("Test", []ProgressStep{
		{Name: "Step 1"},
	})

	// Send spinner tick message
	tickMsg := spinner.TickMsg{}
	p, cmd := p.Update(tickMsg)

	// Should return another tick command for animation
	if cmd == nil {
		t.Error("Update should return command for spinner animation")
	}
}

func TestProgressStepStatus(t *testing.T) {
	// Test step status constants
	if StepPending != 0 {
		t.Error("StepPending should be 0")
	}
	if StepRunning != 1 {
		t.Error("StepRunning should be 1")
	}
	if StepComplete != 2 {
		t.Error("StepComplete should be 2")
	}
	if StepFailed != 3 {
		t.Error("StepFailed should be 3")
	}
	if StepSkipped != 4 {
		t.Error("StepSkipped should be 4")
	}
}

func TestProgressSetCurrentStep(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1"},
		{Name: "Step 2"},
		{Name: "Step 3"},
	}
	p := NewProgress("Test", steps)

	// Set first step to running
	p.SetCurrentStep(0)
	if p.Steps[0].Status != StepRunning {
		t.Error("First step should be running")
	}
	if p.Steps[0].StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	// Set second step - first should be completed
	p.SetCurrentStep(1)
	if p.Steps[0].Status != StepComplete {
		t.Error("First step should be completed")
	}
	if p.Steps[0].EndTime.IsZero() {
		t.Error("EndTime should be set for completed step")
	}
	if p.Steps[1].Status != StepRunning {
		t.Error("Second step should be running")
	}

	// Test with invalid index
	p.SetCurrentStep(-1)  // Should not panic
	p.SetCurrentStep(100) // Should not panic
}

func TestProgressCompleteStep(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1", Status: StepRunning, StartTime: time.Now()},
	}
	p := NewProgress("Test", steps)

	p.CompleteStep(0)
	if p.Steps[0].Status != StepComplete {
		t.Error("Step should be completed")
	}
	if p.Steps[0].EndTime.IsZero() {
		t.Error("EndTime should be set")
	}

	// Test with invalid index
	p.CompleteStep(-1)  // Should not panic
	p.CompleteStep(100) // Should not panic
}

func TestProgressFailStep(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1", Status: StepRunning, StartTime: time.Now()},
	}
	p := NewProgress("Test", steps)

	testErr := errors.New("test error")
	p.FailStep(0, testErr)

	if p.Steps[0].Status != StepFailed {
		t.Error("Step should be failed")
	}
	if p.Steps[0].Error != testErr {
		t.Error("Error should be set")
	}
	if p.Steps[0].EndTime.IsZero() {
		t.Error("EndTime should be set")
	}

	// Test with invalid index
	p.FailStep(-1, testErr)  // Should not panic
	p.FailStep(100, testErr) // Should not panic
}

func TestProgressSkipStep(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1", Status: StepPending},
	}
	p := NewProgress("Test", steps)

	p.SkipStep(0)
	if p.Steps[0].Status != StepSkipped {
		t.Error("Step should be skipped")
	}

	// Test with invalid index
	p.SkipStep(-1)  // Should not panic
	p.SkipStep(100) // Should not panic
}

func TestProgressAddLogLine(t *testing.T) {
	p := NewProgress("Test", []ProgressStep{})
	p.MaxLogLines = 3

	// Add lines
	p.AddLogLine("Line 1")
	p.AddLogLine("Line 2")
	p.AddLogLine("Line 3")

	if len(p.LogLines) != 3 {
		t.Errorf("Expected 3 log lines, got %d", len(p.LogLines))
	}

	// Add more lines - should rotate
	p.AddLogLine("Line 4")
	p.AddLogLine("Line 5")

	if len(p.LogLines) != 3 {
		t.Errorf("Log lines should be limited to MaxLogLines=%d, got %d", p.MaxLogLines, len(p.LogLines))
	}

	// Should have the most recent lines
	if p.LogLines[0] != "Line 3" {
		t.Errorf("Expected 'Line 3', got %q", p.LogLines[0])
	}
	if p.LogLines[2] != "Line 5" {
		t.Errorf("Expected 'Line 5', got %q", p.LogLines[2])
	}
}

func TestProgressSetOutput(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1"},
	}
	p := NewProgress("Test", steps)

	p.SetOutput(0, "test output")
	if p.Steps[0].Output != "test output" {
		t.Errorf("Expected output 'test output', got %q", p.Steps[0].Output)
	}

	// Test with invalid index
	p.SetOutput(-1, "invalid")  // Should not panic
	p.SetOutput(100, "invalid") // Should not panic
}

func TestProgressGetCompletedCount(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1", Status: StepComplete},
		{Name: "Step 2", Status: StepRunning},
		{Name: "Step 3", Status: StepComplete},
		{Name: "Step 4", Status: StepFailed},
		{Name: "Step 5", Status: StepSkipped},
	}
	p := NewProgress("Test", steps)

	count := p.GetCompletedCount()
	if count != 2 {
		t.Errorf("Expected 2 completed steps, got %d", count)
	}
}

func TestProgressGetFailedCount(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Step 1", Status: StepComplete},
		{Name: "Step 2", Status: StepFailed},
		{Name: "Step 3", Status: StepFailed},
		{Name: "Step 4", Status: StepRunning},
	}
	p := NewProgress("Test", steps)

	count := p.GetFailedCount()
	if count != 2 {
		t.Errorf("Expected 2 failed steps, got %d", count)
	}
}

func TestProgressIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		steps    []ProgressStep
		expected bool
	}{
		{
			name: "all complete",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepComplete},
			},
			expected: true,
		},
		{
			name: "all skipped",
			steps: []ProgressStep{
				{Status: StepSkipped},
				{Status: StepSkipped},
			},
			expected: true,
		},
		{
			name: "mixed complete and skipped",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepSkipped},
			},
			expected: true,
		},
		{
			name: "with failed",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepFailed},
			},
			expected: true,
		},
		{
			name: "still running",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepRunning},
			},
			expected: false,
		},
		{
			name: "pending",
			steps: []ProgressStep{
				{Status: StepPending},
			},
			expected: false,
		},
		{
			name: "empty",
			steps: []ProgressStep{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgress("Test", tt.steps)
			if p.IsComplete() != tt.expected {
				t.Errorf("IsComplete() = %v, expected %v", p.IsComplete(), tt.expected)
			}
		})
	}
}

func TestProgressHasFailed(t *testing.T) {
	tests := []struct {
		name     string
		steps    []ProgressStep
		expected bool
	}{
		{
			name: "no failures",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepComplete},
			},
			expected: false,
		},
		{
			name: "with failure",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepFailed},
			},
			expected: true,
		},
		{
			name: "multiple failures",
			steps: []ProgressStep{
				{Status: StepFailed},
				{Status: StepFailed},
			},
			expected: true,
		},
		{
			name: "empty",
			steps: []ProgressStep{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgress("Test", tt.steps)
			if p.HasFailed() != tt.expected {
				t.Errorf("HasFailed() = %v, expected %v", p.HasFailed(), tt.expected)
			}
		})
	}
}

func TestProgressView(t *testing.T) {
	steps := []ProgressStep{
		{Name: "Install Docker", Description: "Installing Docker engine", Status: StepComplete},
		{Name: "Configure Network", Description: "Setting up networking", Status: StepRunning},
		{Name: "Start Services", Description: "Starting containers", Status: StepPending},
	}
	p := NewProgress("Installation Progress", steps)
	p.CurrentStep = 1

	view := p.View()

	// Check title
	if !strings.Contains(view, "Installation Progress") {
		t.Error("View should contain title")
	}

	// Check step names
	if !strings.Contains(view, "Install Docker") {
		t.Error("View should contain step name 'Install Docker'")
	}
	if !strings.Contains(view, "Configure Network") {
		t.Error("View should contain step name 'Configure Network'")
	}
	if !strings.Contains(view, "Start Services") {
		t.Error("View should contain step name 'Start Services'")
	}

	// Progress bar should be present
	if !strings.Contains(view, "[") || !strings.Contains(view, "]") {
		t.Error("View should contain progress bar")
	}

	// Percentage should be shown
	if !strings.Contains(view, "%") {
		t.Error("View should contain percentage")
	}
}

func TestProgressViewWithLogLines(t *testing.T) {
	p := NewProgress("Test", []ProgressStep{
		{Name: "Step 1", Status: StepRunning},
	})
	p.AddLogLine("Log entry 1")
	p.AddLogLine("Log entry 2")

	view := p.View()

	// Log output should be visible
	if !strings.Contains(view, "Output") {
		t.Error("View should contain 'Output' section when logs are present")
	}
	if !strings.Contains(view, "Log entry 1") {
		t.Error("View should contain log line 1")
	}
	if !strings.Contains(view, "Log entry 2") {
		t.Error("View should contain log line 2")
	}
}

func TestProgressViewTruncatesLongLogLines(t *testing.T) {
	p := NewProgress("Test", []ProgressStep{
		{Name: "Step 1", Status: StepRunning},
	})
	p.Width = 20 // Small width to test truncation

	longLine := "This is a very long log line that should be truncated"
	p.AddLogLine(longLine)

	view := p.View()

	// The full long line should not appear
	if strings.Contains(view, longLine) {
		t.Error("Long log line should be truncated")
	}
	// Truncated lines end with "..."
	if !strings.Contains(view, "...") {
		t.Error("Truncated line should end with '...'")
	}
}

func TestProgressViewStepDuration(t *testing.T) {
	startTime := time.Now().Add(-5 * time.Second)
	endTime := time.Now()

	steps := []ProgressStep{
		{
			Name:      "Completed Step",
			Status:    StepComplete,
			StartTime: startTime,
			EndTime:   endTime,
		},
	}
	p := NewProgress("Test", steps)

	view := p.View()

	// Should show duration for completed steps
	// Duration format is "(X.Xs)"
	if !strings.Contains(view, "s)") {
		t.Error("View should show duration for completed step")
	}
}

func TestProgressRenderProgressBar(t *testing.T) {
	tests := []struct {
		name           string
		steps          []ProgressStep
		expectPercent  string
	}{
		{
			name: "0% complete",
			steps: []ProgressStep{
				{Status: StepPending},
				{Status: StepPending},
			},
			expectPercent: "0%",
		},
		{
			name: "50% complete",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepPending},
			},
			expectPercent: "50%",
		},
		{
			name: "100% complete",
			steps: []ProgressStep{
				{Status: StepComplete},
				{Status: StepComplete},
			},
			expectPercent: "100%",
		},
		{
			name: "skipped counts as complete",
			steps: []ProgressStep{
				{Status: StepSkipped},
				{Status: StepSkipped},
			},
			expectPercent: "100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgress("Test", tt.steps)
			view := p.View()

			if !strings.Contains(view, tt.expectPercent) {
				t.Errorf("Expected %s in view, got: %s", tt.expectPercent, view)
			}
		})
	}
}

func TestProgressStepIcons(t *testing.T) {
	p := NewProgress("Test", []ProgressStep{
		{Name: "Complete", Status: StepComplete},
		{Name: "Failed", Status: StepFailed},
		{Name: "Skipped", Status: StepSkipped},
		{Name: "Running", Status: StepRunning},
		{Name: "Pending", Status: StepPending},
	})
	p.CurrentStep = 3 // Running step

	view := p.View()

	// The view should contain all step names
	if !strings.Contains(view, "Complete") {
		t.Error("View should contain 'Complete' step")
	}
	if !strings.Contains(view, "Failed") {
		t.Error("View should contain 'Failed' step")
	}
	if !strings.Contains(view, "Skipped") {
		t.Error("View should contain 'Skipped' step")
	}
	if !strings.Contains(view, "Running") {
		t.Error("View should contain 'Running' step")
	}
	if !strings.Contains(view, "Pending") {
		t.Error("View should contain 'Pending' step")
	}
}

func TestProgressEmptySteps(t *testing.T) {
	p := NewProgress("Empty Progress", []ProgressStep{})

	// Should handle empty steps gracefully
	p.SetCurrentStep(0)   // Should not panic
	p.CompleteStep(0)     // Should not panic
	p.FailStep(0, nil)    // Should not panic
	p.SkipStep(0)         // Should not panic
	p.SetOutput(0, "out") // Should not panic

	if !p.IsComplete() {
		t.Error("Empty progress should be considered complete")
	}
	if p.HasFailed() {
		t.Error("Empty progress should not have failures")
	}
	if p.GetCompletedCount() != 0 {
		t.Error("Empty progress should have 0 completed")
	}
	if p.GetFailedCount() != 0 {
		t.Error("Empty progress should have 0 failed")
	}

	view := p.View()
	if !strings.Contains(view, "Empty Progress") {
		t.Error("View should contain title even with no steps")
	}
}

func TestProgressStep(t *testing.T) {
	step := ProgressStep{
		Name:        "Test Step",
		Description: "Test Description",
		Status:      StepRunning,
		Output:      "Output text",
		StartTime:   time.Now(),
		Error:       errors.New("test error"),
	}

	if step.Name != "Test Step" {
		t.Error("Name mismatch")
	}
	if step.Description != "Test Description" {
		t.Error("Description mismatch")
	}
	if step.Status != StepRunning {
		t.Error("Status mismatch")
	}
	if step.Output != "Output text" {
		t.Error("Output mismatch")
	}
	if step.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
	if step.Error == nil {
		t.Error("Error should be set")
	}
}
