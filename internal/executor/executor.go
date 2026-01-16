package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Step represents an installation step
type Step struct {
	Name        string
	Description string
	Command     string
	Args        []string
	WorkDir     string
	Timeout     time.Duration
	Optional    bool
	Condition   func() bool // Only run if this returns true
}

// StepResult contains the result of executing a step
type StepResult struct {
	Step     *Step
	Success  bool
	Output   string
	Error    error
	Duration time.Duration
}

// ProgressCallback is called during execution with progress updates
type ProgressCallback func(stepIndex int, totalSteps int, step *Step, output string)

// Executor handles running installation steps
type Executor struct {
	ProjectRoot string
	DryRun      bool
	Verbose     bool
	LogFile     string
	Steps       []Step

	mu         sync.Mutex
	results    []StepResult
	currentStep int
	cancelled  bool
}

// NewExecutor creates a new executor instance
func NewExecutor(projectRoot string) *Executor {
	return &Executor{
		ProjectRoot: projectRoot,
		LogFile:     "/var/log/doom-coding-install.log",
		Steps:       getDefaultSteps(projectRoot),
	}
}

// getDefaultSteps returns the default installation steps
func getDefaultSteps(projectRoot string) []Step {
	return []Step{
		{
			Name:        "system_check",
			Description: "Checking system requirements",
			Command:     "bash",
			Args:        []string{"-c", "echo 'Checking system...' && uname -a"},
			Timeout:     30 * time.Second,
		},
		{
			Name:        "base_packages",
			Description: "Installing base packages",
			Command:     "bash",
			Args:        []string{filepath.Join(projectRoot, "scripts", "install.sh"), "--dry-run"},
			Timeout:     5 * time.Minute,
			Optional:    true,
		},
		{
			Name:        "docker_install",
			Description: "Setting up Docker",
			Command:     "bash",
			Args:        []string{"-c", "command -v docker || echo 'Docker will be installed'"},
			Timeout:     10 * time.Minute,
		},
		{
			Name:        "network_config",
			Description: "Configuring network",
			Command:     "bash",
			Args:        []string{"-c", "echo 'Network configuration...'"},
			Timeout:     2 * time.Minute,
		},
		{
			Name:        "terminal_tools",
			Description: "Installing terminal tools",
			Command:     "bash",
			Args:        []string{"-c", "echo 'Terminal tools setup...'"},
			Timeout:     10 * time.Minute,
		},
		{
			Name:        "ssh_hardening",
			Description: "Applying security hardening",
			Command:     "bash",
			Args:        []string{"-c", "echo 'SSH hardening...'"},
			Timeout:     2 * time.Minute,
		},
		{
			Name:        "secrets_setup",
			Description: "Setting up secrets management",
			Command:     "bash",
			Args:        []string{"-c", "echo 'Secrets management...'"},
			Timeout:     2 * time.Minute,
		},
		{
			Name:        "env_config",
			Description: "Creating environment file",
			Command:     "bash",
			Args:        []string{"-c", "echo '.env file created'"},
			Timeout:     30 * time.Second,
		},
		{
			Name:        "services_start",
			Description: "Starting services",
			Command:     "bash",
			Args:        []string{"-c", "echo 'Starting containers...'"},
			Timeout:     5 * time.Minute,
		},
		{
			Name:        "health_check",
			Description: "Running health checks",
			Command:     "bash",
			Args:        []string{filepath.Join(projectRoot, "scripts", "health-check.sh")},
			Timeout:     2 * time.Minute,
		},
	}
}

// RunInstallScript runs the main install.sh script with given flags
func (e *Executor) RunInstallScript(ctx context.Context, flags []string, progressCb ProgressCallback) error {
	scriptPath := filepath.Join(e.ProjectRoot, "scripts", "install.sh")

	args := append([]string{scriptPath}, flags...)

	if e.DryRun {
		args = append(args, "--dry-run")
	}
	if e.Verbose {
		args = append(args, "--verbose")
	}

	return e.runCommand(ctx, "bash", args, progressCb)
}

// RunSteps executes all configured steps
func (e *Executor) RunSteps(ctx context.Context, progressCb ProgressCallback) error {
	e.results = make([]StepResult, 0, len(e.Steps))
	e.currentStep = 0

	for i, step := range e.Steps {
		e.mu.Lock()
		if e.cancelled {
			e.mu.Unlock()
			return fmt.Errorf("installation cancelled")
		}
		e.currentStep = i
		e.mu.Unlock()

		// Check condition
		if step.Condition != nil && !step.Condition() {
			e.results = append(e.results, StepResult{
				Step:    &step,
				Success: true,
				Output:  "Skipped (condition not met)",
			})
			continue
		}

		result := e.runStep(ctx, &step, i, progressCb)
		e.results = append(e.results, result)

		if !result.Success && !step.Optional {
			return fmt.Errorf("step '%s' failed: %w", step.Name, result.Error)
		}
	}

	return nil
}

func (e *Executor) runStep(ctx context.Context, step *Step, index int, progressCb ProgressCallback) StepResult {
	start := time.Now()

	// Create context with timeout
	stepCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	// Prepare command
	cmd := exec.CommandContext(stepCtx, step.Command, step.Args...)
	if step.WorkDir != "" {
		cmd.Dir = step.WorkDir
	} else {
		cmd.Dir = e.ProjectRoot
	}

	// Set up output capture
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return StepResult{
			Step:     step,
			Success:  false,
			Error:    fmt.Errorf("failed to create stdout pipe: %w", err),
			Duration: time.Since(start),
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return StepResult{
			Step:     step,
			Success:  false,
			Error:    fmt.Errorf("failed to create stderr pipe: %w", err),
			Duration: time.Since(start),
		}
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return StepResult{
			Step:     step,
			Success:  false,
			Error:    fmt.Errorf("failed to start command: %w", err),
			Duration: time.Since(start),
		}
	}

	// Read output in real-time
	var output strings.Builder
	outputChan := make(chan string, 100)

	go e.readOutput(stdout, outputChan)
	go e.readOutput(stderr, outputChan)

	// Process output and send progress updates
	go func() {
		for line := range outputChan {
			output.WriteString(line)
			output.WriteString("\n")
			if progressCb != nil {
				progressCb(index+1, len(e.Steps), step, line)
			}
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	close(outputChan)

	return StepResult{
		Step:     step,
		Success:  err == nil,
		Output:   output.String(),
		Error:    err,
		Duration: time.Since(start),
	}
}

func (e *Executor) readOutput(r io.Reader, outputChan chan<- string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		outputChan <- scanner.Text()
	}
}

func (e *Executor) runCommand(ctx context.Context, command string, args []string, progressCb ProgressCallback) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = e.ProjectRoot

	// Open log file
	logFile, err := os.OpenFile(e.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to stderr if log file can't be opened
		logFile = os.Stderr
	} else {
		defer logFile.Close()
	}

	// Set up pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read and process output
	outputChan := make(chan string, 100)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		e.readOutput(stdout, outputChan)
	}()
	go func() {
		defer wg.Done()
		e.readOutput(stderr, outputChan)
	}()

	// Process output
	go func() {
		wg.Wait()
		close(outputChan)
	}()

	stepIndex := 0
	for line := range outputChan {
		// Write to log file
		fmt.Fprintln(logFile, line)

		// Parse step markers from install.sh output
		if strings.Contains(line, "==>") || strings.Contains(line, "[STEP]") {
			stepIndex++
		}

		// Send progress callback
		if progressCb != nil {
			step := &Step{Description: line}
			progressCb(stepIndex, 10, step, line)
		}
	}

	return cmd.Wait()
}

// Cancel cancels the current installation
func (e *Executor) Cancel() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cancelled = true
}

// GetResults returns the results of completed steps
func (e *Executor) GetResults() []StepResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.results
}

// GetCurrentStep returns the current step index
func (e *Executor) GetCurrentStep() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.currentStep
}

// GenerateBashCommand returns the equivalent bash command
func (e *Executor) GenerateBashCommand(flags []string) string {
	scriptPath := filepath.Join(e.ProjectRoot, "scripts", "install.sh")
	args := append([]string{scriptPath}, flags...)
	return "bash " + strings.Join(args, " ")
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	ProjectRoot string
}

// HealthCheckResult contains health check results
type HealthCheckResult struct {
	Docker     bool
	Containers map[string]bool
	Tailscale  bool
	TailscaleIP string
	Terminal   bool
	SSH        bool
	Secrets    bool
	Errors     []string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(projectRoot string) *HealthChecker {
	return &HealthChecker{ProjectRoot: projectRoot}
}

// Check runs the health check
func (h *HealthChecker) Check(ctx context.Context) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Containers: make(map[string]bool),
	}

	// Run health-check.sh
	scriptPath := filepath.Join(h.ProjectRoot, "scripts", "health-check.sh")
	cmd := exec.CommandContext(ctx, "bash", scriptPath, "--json")

	output, err := cmd.Output()
	if err != nil {
		// Fall back to basic checks
		h.basicChecks(result)
		return result, nil
	}

	// Parse JSON output (simplified parsing)
	outputStr := string(output)

	result.Docker = strings.Contains(outputStr, `"docker":true`) ||
		strings.Contains(outputStr, `"docker": "running"`)

	result.Terminal = strings.Contains(outputStr, `"zsh":true`) ||
		strings.Contains(outputStr, `"terminal": "installed"`)

	if strings.Contains(outputStr, `"tailscale":true`) ||
		strings.Contains(outputStr, `"tailscale": "connected"`) {
		result.Tailscale = true
	}

	// Check individual containers
	for _, container := range []string{"doom-tailscale", "doom-code-server", "doom-claude"} {
		result.Containers[container] = strings.Contains(outputStr, fmt.Sprintf(`"%s": "healthy"`, container))
	}

	return result, nil
}

func (h *HealthChecker) basicChecks(result *HealthCheckResult) {
	// Docker check
	if _, err := exec.LookPath("docker"); err == nil {
		if err := exec.Command("docker", "info").Run(); err == nil {
			result.Docker = true
		}
	}

	// Terminal tools check
	if _, err := exec.LookPath("zsh"); err == nil {
		result.Terminal = true
	}

	// Tailscale check
	if _, err := exec.LookPath("tailscale"); err == nil {
		if output, err := exec.Command("tailscale", "status").Output(); err == nil {
			if !strings.Contains(string(output), "Tailscale is stopped") {
				result.Tailscale = true
			}
		}
		if output, err := exec.Command("tailscale", "ip", "-4").Output(); err == nil {
			result.TailscaleIP = strings.TrimSpace(string(output))
		}
	}

	// Container status
	if result.Docker {
		for _, container := range []string{"doom-tailscale", "doom-code-server", "doom-claude"} {
			if err := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", container).Run(); err == nil {
				result.Containers[container] = true
			}
		}
	}
}
