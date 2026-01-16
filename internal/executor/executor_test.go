package executor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewExecutor(t *testing.T) {
	projectRoot := "/test/project"
	exec := NewExecutor(projectRoot)

	if exec.ProjectRoot != projectRoot {
		t.Errorf("Expected ProjectRoot=%q, got %q", projectRoot, exec.ProjectRoot)
	}

	if exec.DryRun {
		t.Error("DryRun should be false by default")
	}

	if exec.Verbose {
		t.Error("Verbose should be false by default")
	}

	if exec.LogFile != "/var/log/doom-coding-install.log" {
		t.Errorf("Unexpected LogFile: %q", exec.LogFile)
	}

	if len(exec.Steps) == 0 {
		t.Error("Steps should be initialized with default steps")
	}
}

func TestGetDefaultSteps(t *testing.T) {
	projectRoot := "/test/project"
	steps := getDefaultSteps(projectRoot)

	if len(steps) == 0 {
		t.Fatal("Should return default steps")
	}

	// Check that required steps exist
	stepNames := make(map[string]bool)
	for _, step := range steps {
		stepNames[step.Name] = true
	}

	requiredSteps := []string{
		"system_check",
		"docker_install",
		"health_check",
	}

	for _, required := range requiredSteps {
		if !stepNames[required] {
			t.Errorf("Missing required step: %s", required)
		}
	}
}

func TestStepStructure(t *testing.T) {
	step := Step{
		Name:        "test_step",
		Description: "Test step description",
		Command:     "echo",
		Args:        []string{"hello"},
		WorkDir:     "/tmp",
		Timeout:     30 * time.Second,
		Optional:    true,
		Condition:   func() bool { return true },
	}

	if step.Name != "test_step" {
		t.Error("Name mismatch")
	}
	if step.Description != "Test step description" {
		t.Error("Description mismatch")
	}
	if step.Command != "echo" {
		t.Error("Command mismatch")
	}
	if len(step.Args) != 1 || step.Args[0] != "hello" {
		t.Error("Args mismatch")
	}
	if step.WorkDir != "/tmp" {
		t.Error("WorkDir mismatch")
	}
	if step.Timeout != 30*time.Second {
		t.Error("Timeout mismatch")
	}
	if !step.Optional {
		t.Error("Optional should be true")
	}
	if step.Condition == nil {
		t.Error("Condition should be set")
	}
}

func TestStepResultStructure(t *testing.T) {
	step := &Step{Name: "test"}
	result := StepResult{
		Step:     step,
		Success:  true,
		Output:   "test output",
		Error:    nil,
		Duration: 5 * time.Second,
	}

	if result.Step != step {
		t.Error("Step mismatch")
	}
	if !result.Success {
		t.Error("Success should be true")
	}
	if result.Output != "test output" {
		t.Error("Output mismatch")
	}
	if result.Error != nil {
		t.Error("Error should be nil")
	}
	if result.Duration != 5*time.Second {
		t.Error("Duration mismatch")
	}
}

func TestExecutorCancel(t *testing.T) {
	exec := NewExecutor("/test")

	exec.Cancel()

	exec.mu.Lock()
	cancelled := exec.cancelled
	exec.mu.Unlock()

	if !cancelled {
		t.Error("Executor should be cancelled")
	}
}

func TestExecutorGetResults(t *testing.T) {
	exec := NewExecutor("/test")

	// Initially empty
	results := exec.GetResults()
	if len(results) != 0 {
		t.Error("Results should be empty initially")
	}

	// Add some results
	exec.mu.Lock()
	exec.results = []StepResult{
		{Success: true},
		{Success: false},
	}
	exec.mu.Unlock()

	results = exec.GetResults()
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestExecutorGetCurrentStep(t *testing.T) {
	exec := NewExecutor("/test")

	// Initially 0
	if exec.GetCurrentStep() != 0 {
		t.Errorf("Expected currentStep=0, got %d", exec.GetCurrentStep())
	}

	// Update current step
	exec.mu.Lock()
	exec.currentStep = 5
	exec.mu.Unlock()

	if exec.GetCurrentStep() != 5 {
		t.Errorf("Expected currentStep=5, got %d", exec.GetCurrentStep())
	}
}

func TestExecutorGenerateBashCommand(t *testing.T) {
	exec := NewExecutor("/test/project")

	flags := []string{"--unattended", "--skip-docker", "--code-password=secret"}
	cmd := exec.GenerateBashCommand(flags)

	expected := "bash /test/project/scripts/install.sh --unattended --skip-docker --code-password=secret"
	if cmd != expected {
		t.Errorf("Expected command:\n%s\nGot:\n%s", expected, cmd)
	}
}

func TestExecutorGenerateBashCommandNoFlags(t *testing.T) {
	exec := NewExecutor("/test/project")

	cmd := exec.GenerateBashCommand([]string{})

	expected := "bash /test/project/scripts/install.sh"
	if cmd != expected {
		t.Errorf("Expected command:\n%s\nGot:\n%s", expected, cmd)
	}
}

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker("/test/project")

	if hc.ProjectRoot != "/test/project" {
		t.Errorf("Expected ProjectRoot=/test/project, got %q", hc.ProjectRoot)
	}
}

func TestHealthCheckResultStructure(t *testing.T) {
	result := HealthCheckResult{
		Docker:      true,
		Containers:  map[string]bool{"doom-code-server": true},
		Tailscale:   true,
		TailscaleIP: "100.64.1.1",
		Terminal:    true,
		SSH:         true,
		Secrets:     true,
		Errors:      []string{"warning message"},
	}

	if !result.Docker {
		t.Error("Docker should be true")
	}
	if !result.Containers["doom-code-server"] {
		t.Error("Container should be true")
	}
	if !result.Tailscale {
		t.Error("Tailscale should be true")
	}
	if result.TailscaleIP != "100.64.1.1" {
		t.Error("TailscaleIP mismatch")
	}
	if !result.Terminal {
		t.Error("Terminal should be true")
	}
	if !result.SSH {
		t.Error("SSH should be true")
	}
	if !result.Secrets {
		t.Error("Secrets should be true")
	}
	if len(result.Errors) != 1 {
		t.Error("Errors should have 1 item")
	}
}

func TestExecutorRunStepsSimple(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a simple executor with basic steps
	exec := &Executor{
		ProjectRoot: tmpDir,
		DryRun:      false,
		Verbose:     false,
		Steps: []Step{
			{
				Name:        "simple_echo",
				Description: "Simple echo test",
				Command:     "echo",
				Args:        []string{"hello world"},
				Timeout:     5 * time.Second,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var progressCalls int
	progressCb := func(stepIndex int, totalSteps int, step *Step, output string) {
		progressCalls++
	}

	err := exec.RunSteps(ctx, progressCb)
	if err != nil {
		t.Fatalf("RunSteps failed: %v", err)
	}

	results := exec.GetResults()
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if !results[0].Success {
		t.Errorf("Step should have succeeded: %v", results[0].Error)
	}
}

func TestExecutorRunStepsWithCondition(t *testing.T) {
	tmpDir := t.TempDir()

	conditionCalled := false
	exec := &Executor{
		ProjectRoot: tmpDir,
		Steps: []Step{
			{
				Name:    "conditional_step",
				Command: "echo",
				Args:    []string{"should not run"},
				Timeout: 5 * time.Second,
				Condition: func() bool {
					conditionCalled = true
					return false // Don't run this step
				},
			},
		},
	}

	ctx := context.Background()
	err := exec.RunSteps(ctx, nil)
	if err != nil {
		t.Fatalf("RunSteps failed: %v", err)
	}

	if !conditionCalled {
		t.Error("Condition should have been called")
	}

	results := exec.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if !results[0].Success {
		t.Error("Skipped step should be marked as success")
	}
	if !strings.Contains(results[0].Output, "Skipped") {
		t.Errorf("Output should indicate skipped: %q", results[0].Output)
	}
}

func TestExecutorRunStepsCancelled(t *testing.T) {
	tmpDir := t.TempDir()

	exec := &Executor{
		ProjectRoot: tmpDir,
		Steps: []Step{
			{
				Name:    "step1",
				Command: "sleep",
				Args:    []string{"10"},
				Timeout: 30 * time.Second,
			},
		},
	}

	// Cancel before running
	exec.Cancel()

	ctx := context.Background()
	err := exec.RunSteps(ctx, nil)

	if err == nil {
		t.Error("Expected error for cancelled execution")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("Expected 'cancelled' in error, got: %v", err)
	}
}

func TestExecutorRunStepsOptionalFailure(t *testing.T) {
	tmpDir := t.TempDir()

	exec := &Executor{
		ProjectRoot: tmpDir,
		Steps: []Step{
			{
				Name:     "optional_fail",
				Command:  "false", // This command fails
				Timeout:  5 * time.Second,
				Optional: true,
			},
			{
				Name:    "after_optional",
				Command: "echo",
				Args:    []string{"success"},
				Timeout: 5 * time.Second,
			},
		},
	}

	ctx := context.Background()
	err := exec.RunSteps(ctx, nil)

	// Should not fail because first step is optional
	if err != nil {
		t.Fatalf("RunSteps should not fail for optional step failure: %v", err)
	}

	results := exec.GetResults()
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0].Success {
		t.Error("First optional step should have failed")
	}
	if !results[1].Success {
		t.Error("Second step should have succeeded")
	}
}

func TestExecutorRunStepsRequiredFailure(t *testing.T) {
	tmpDir := t.TempDir()

	exec := &Executor{
		ProjectRoot: tmpDir,
		Steps: []Step{
			{
				Name:     "required_fail",
				Command:  "false",
				Timeout:  5 * time.Second,
				Optional: false,
			},
			{
				Name:    "should_not_run",
				Command: "echo",
				Args:    []string{"should not execute"},
				Timeout: 5 * time.Second,
			},
		},
	}

	ctx := context.Background()
	err := exec.RunSteps(ctx, nil)

	// Should fail because first step is not optional
	if err == nil {
		t.Fatal("RunSteps should fail for required step failure")
	}

	if !strings.Contains(err.Error(), "required_fail") {
		t.Errorf("Error should mention failed step name: %v", err)
	}

	// Second step should not have been executed
	results := exec.GetResults()
	if len(results) != 1 {
		t.Errorf("Expected 1 result (second step should not run), got %d", len(results))
	}
}

func TestExecutorRunStepsTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	exec := &Executor{
		ProjectRoot: tmpDir,
		Steps: []Step{
			{
				Name:     "timeout_step",
				Command:  "sleep",
				Args:     []string{"30"}, // Sleep for 30 seconds
				Timeout:  100 * time.Millisecond, // But timeout after 100ms
				Optional: true,
			},
		},
	}

	ctx := context.Background()
	start := time.Now()
	exec.RunSteps(ctx, nil)
	duration := time.Since(start)

	// Should complete within reasonable time (much less than 30 seconds)
	if duration > 5*time.Second {
		t.Errorf("Step should have timed out quickly, took %v", duration)
	}

	results := exec.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Success {
		t.Error("Timed out step should not be marked as success")
	}
}

func TestHealthCheckerBasicChecks(t *testing.T) {
	hc := NewHealthChecker(t.TempDir())
	result := &HealthCheckResult{
		Containers: make(map[string]bool),
	}

	// Run basic checks
	hc.basicChecks(result)

	// Results depend on what's installed on the system
	// We just verify it doesn't panic
	t.Logf("Docker: %v", result.Docker)
	t.Logf("Terminal: %v", result.Terminal)
	t.Logf("Tailscale: %v", result.Tailscale)
}

func TestHealthCheckerCheck(t *testing.T) {
	// Create a temp directory with a mock health-check script
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}

	// Create a simple health-check.sh that outputs JSON
	healthScript := `#!/bin/bash
echo '{"docker":true,"terminal":"installed"}'
`
	scriptPath := filepath.Join(scriptsDir, "health-check.sh")
	if err := os.WriteFile(scriptPath, []byte(healthScript), 0755); err != nil {
		t.Fatalf("Failed to write health-check script: %v", err)
	}

	hc := NewHealthChecker(tmpDir)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := hc.Check(ctx)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// The result depends on the mock script output parsing
	t.Logf("Health check result: Docker=%v, Terminal=%v", result.Docker, result.Terminal)
}

func TestProgressCallbackType(t *testing.T) {
	var cb ProgressCallback = func(stepIndex int, totalSteps int, step *Step, output string) {
		// Verify callback parameters
		if stepIndex < 0 {
			t.Error("stepIndex should not be negative")
		}
		if totalSteps < 0 {
			t.Error("totalSteps should not be negative")
		}
	}

	// Test that callback can be called
	step := &Step{Name: "test"}
	cb(1, 10, step, "output")
}

func TestExecutorDryRunFlags(t *testing.T) {
	exec := NewExecutor("/test/project")
	exec.DryRun = true
	exec.Verbose = true

	// These flags would be added when running install script
	// We test the structure is correct
	if !exec.DryRun {
		t.Error("DryRun should be true")
	}
	if !exec.Verbose {
		t.Error("Verbose should be true")
	}
}

func TestExecutorWithWorkDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a subdirectory
	workDir := filepath.Join(tmpDir, "workdir")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create workdir: %v", err)
	}

	exec := &Executor{
		ProjectRoot: tmpDir,
		Steps: []Step{
			{
				Name:    "pwd_test",
				Command: "pwd",
				WorkDir: workDir,
				Timeout: 5 * time.Second,
			},
		},
	}

	ctx := context.Background()
	err := exec.RunSteps(ctx, nil)
	if err != nil {
		t.Fatalf("RunSteps failed: %v", err)
	}

	results := exec.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Output should contain the work directory
	if !strings.Contains(results[0].Output, "workdir") {
		t.Errorf("Output should contain workdir path: %q", results[0].Output)
	}
}
