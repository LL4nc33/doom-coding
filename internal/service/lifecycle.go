package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LifecycleManager handles clean service startup and shutdown
type LifecycleManager struct {
	manager      *Manager
	projectRoot  string
	composeFile  string
	logger       *Logger
	timeout      time.Duration
	healthChecks bool
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(manager *Manager, projectRoot, composeFile string) *LifecycleManager {
	return &LifecycleManager{
		manager:      manager,
		projectRoot:  projectRoot,
		composeFile:  composeFile,
		timeout:      2 * time.Minute,
		healthChecks: true,
	}
}

// SetLogger sets the logger for output
func (lm *LifecycleManager) SetLogger(logger *Logger) {
	lm.logger = logger
}

// SetTimeout sets the operation timeout
func (lm *LifecycleManager) SetTimeout(timeout time.Duration) {
	lm.timeout = timeout
}

// SetHealthChecks enables/disables health check waiting
func (lm *LifecycleManager) SetHealthChecks(enabled bool) {
	lm.healthChecks = enabled
}

// StartupResult contains the result of starting services
type StartupResult struct {
	Success        bool
	StartedAt      time.Time
	Duration       time.Duration
	Services       []ServiceStatus
	AccessURLs     map[string]string
	Errors         []string
	Warnings       []string
}

// ServiceStatus represents the status of a single service after startup
type ServiceStatus struct {
	Name       string
	Container  string
	State      ServiceState
	Port       int
	HealthURL  string
	Error      string
}

// ShutdownResult contains the result of stopping services
type ShutdownResult struct {
	Success     bool
	StoppedAt   time.Time
	Duration    time.Duration
	Services    []ServiceStatus
	Errors      []string
}

// PreStartCheck performs pre-startup validation
func (lm *LifecycleManager) PreStartCheck(ctx context.Context) (*MigrationPlan, error) {
	lm.log(LogInfo, "startup", "Running pre-start checks...")

	// Check Docker
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err != nil {
		return nil, fmt.Errorf("docker is not running or not accessible: %w", err)
	}
	lm.log(LogDebug, "startup", "Docker is available")

	// Check docker-compose file exists
	composePath := filepath.Join(lm.projectRoot, lm.composeFile)
	if _, err := os.Stat(composePath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s", composePath)
	}
	lm.log(LogDebug, "startup", fmt.Sprintf("Using compose file: %s", lm.composeFile))

	// Detect existing services and plan migration
	migrator := NewMigrator(lm.manager, lm.projectRoot)

	targetPorts := map[string]int{
		"code-server": 8443,
		"ttyd":        7681,
	}

	plan, err := migrator.AnalyzeExisting(ctx, targetPorts)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze existing services: %w", err)
	}

	// Log what we found
	if len(plan.ExistingServices) > 0 {
		lm.log(LogInfo, "startup", fmt.Sprintf("Detected %d existing services", len(plan.ExistingServices)))
		for _, svc := range plan.ExistingServices {
			lm.log(LogDebug, "startup", fmt.Sprintf("  - %s (%s) [%s]", svc.Name, svc.Type, svc.State))
		}
	}

	return plan, nil
}

// Start starts the doom-coding services
func (lm *LifecycleManager) Start(ctx context.Context, plan *MigrationPlan) (*StartupResult, error) {
	result := &StartupResult{
		StartedAt:  time.Now(),
		AccessURLs: make(map[string]string),
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, lm.timeout)
	defer cancel()

	// Execute migration plan if needed
	if plan != nil && len(plan.Actions) > 0 {
		lm.log(LogInfo, "startup", "Executing migration plan...")
		migrator := NewMigrator(lm.manager, lm.projectRoot)
		migrator.SetDryRun(false)
		_, err := migrator.Execute(ctx, plan)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Migration failed: %v", err))
			// Continue with startup attempt anyway
		}
	}

	// Pull images (with filtered output)
	lm.log(LogInfo, "startup", "Pulling container images...")
	if err := lm.pullImages(ctx); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Pull warning: %v", err))
	}

	// Start services
	lm.log(LogInfo, "startup", "Starting services...")
	if err := lm.startServices(ctx); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Start failed: %v", err))
		return result, err
	}

	// Wait for health checks
	if lm.healthChecks {
		lm.log(LogInfo, "startup", "Waiting for services to be healthy...")
		statuses := lm.waitForHealth(ctx)
		result.Services = statuses

		// Check if all healthy
		allHealthy := true
		for _, svc := range statuses {
			if svc.State != StateHealthy && svc.State != StateRunning {
				allHealthy = false
				if svc.Error != "" {
					result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %s", svc.Name, svc.Error))
				}
			}
		}

		if !allHealthy {
			result.Warnings = append(result.Warnings, "Some services are not yet healthy. They may still be starting.")
		}
	}

	// Determine access URLs
	result.AccessURLs = lm.getAccessURLs(ctx)

	result.Success = len(result.Errors) == 0
	result.Duration = time.Since(result.StartedAt)

	return result, nil
}

// pullImages pulls the container images with filtered output
func (lm *LifecycleManager) pullImages(ctx context.Context) error {
	composePath := filepath.Join(lm.projectRoot, lm.composeFile)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "pull")
	cmd.Dir = lm.projectRoot

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Process output through filter
	if lm.logger != nil {
		go lm.logger.NewStreamFilter("docker-pull", stdout).Process()
		go lm.logger.NewStreamFilter("docker-pull", stderr).Process()
	}

	return cmd.Wait()
}

// startServices starts the containers
func (lm *LifecycleManager) startServices(ctx context.Context) error {
	composePath := filepath.Join(lm.projectRoot, lm.composeFile)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "up", "-d")
	cmd.Dir = lm.projectRoot

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Process output through filter
	if lm.logger != nil {
		go lm.logger.NewStreamFilter("docker-up", stdout).Process()
		go lm.logger.NewStreamFilter("docker-up", stderr).Process()
	}

	return cmd.Wait()
}

// waitForHealth waits for services to become healthy
func (lm *LifecycleManager) waitForHealth(ctx context.Context) []ServiceStatus {
	containers := []struct {
		name      string
		container string
		port      int
	}{
		{"Tailscale", "doom-tailscale", 0},
		{"code-server", "doom-code-server", 8443},
		{"Claude", "doom-claude", 7681},
	}

	var statuses []ServiceStatus

	for _, c := range containers {
		status := ServiceStatus{
			Name:      c.name,
			Container: c.container,
			Port:      c.port,
		}

		// Check if container exists
		if err := exec.CommandContext(ctx, "docker", "inspect", c.container).Run(); err != nil {
			status.State = StateStopped
			status.Error = "Container not found"
			statuses = append(statuses, status)
			continue
		}

		// Wait for health with timeout
		healthCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		status.State = lm.waitForContainerHealth(healthCtx, c.container)
		cancel()

		if status.State == StateHealthy {
			lm.log(LogInfo, "health", fmt.Sprintf("%s is healthy", c.name))
		} else if status.State == StateRunning {
			lm.log(LogDebug, "health", fmt.Sprintf("%s is running (no healthcheck)", c.name))
		} else {
			lm.log(LogWarning, "health", fmt.Sprintf("%s is %s", c.name, status.State))
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// waitForContainerHealth waits for a specific container to be healthy
func (lm *LifecycleManager) waitForContainerHealth(ctx context.Context, container string) ServiceState {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return StateUnhealthy
		case <-ticker.C:
			// Check container state
			output, err := exec.CommandContext(ctx, "docker", "inspect",
				"--format", "{{.State.Status}},{{.State.Health.Status}}",
				container).Output()
			if err != nil {
				continue
			}

			parts := strings.Split(strings.TrimSpace(string(output)), ",")
			if len(parts) < 1 {
				continue
			}

			runState := parts[0]
			healthState := ""
			if len(parts) > 1 {
				healthState = parts[1]
			}

			if runState != "running" {
				return StateStopped
			}

			if healthState == "" || healthState == "<no value>" {
				// No healthcheck defined, consider running as success
				return StateRunning
			}

			if healthState == "healthy" {
				return StateHealthy
			}
		}
	}
}

// getAccessURLs determines the URLs to access services
func (lm *LifecycleManager) getAccessURLs(ctx context.Context) map[string]string {
	urls := make(map[string]string)

	// Try Tailscale IP first
	if output, err := exec.CommandContext(ctx, "tailscale", "ip", "-4").Output(); err == nil {
		tsIP := strings.TrimSpace(string(output))
		if tsIP != "" {
			urls["code-server"] = fmt.Sprintf("https://%s:8443", tsIP)
			urls["ttyd"] = fmt.Sprintf("http://%s:7681", tsIP)
			return urls
		}
	}

	// Check container Tailscale
	if output, err := exec.CommandContext(ctx, "docker", "exec", "doom-tailscale",
		"tailscale", "ip", "-4").Output(); err == nil {
		tsIP := strings.TrimSpace(string(output))
		if tsIP != "" {
			urls["code-server"] = fmt.Sprintf("https://%s:8443", tsIP)
			urls["ttyd"] = fmt.Sprintf("http://%s:7681", tsIP)
			return urls
		}
	}

	// Fallback to local IPs
	if output, err := exec.Command("hostname", "-I").Output(); err == nil {
		ips := strings.Fields(string(output))
		if len(ips) > 0 {
			localIP := ips[0]
			urls["code-server"] = fmt.Sprintf("https://%s:8443", localIP)
			urls["ttyd"] = fmt.Sprintf("http://%s:7681", localIP)
		}
	}

	// Ultimate fallback
	if len(urls) == 0 {
		urls["code-server"] = "https://localhost:8443"
		urls["ttyd"] = "http://localhost:7681"
	}

	return urls
}

// Stop gracefully stops the doom-coding services
func (lm *LifecycleManager) Stop(ctx context.Context) (*ShutdownResult, error) {
	result := &ShutdownResult{
		StoppedAt: time.Now(),
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, lm.timeout)
	defer cancel()

	lm.log(LogInfo, "shutdown", "Stopping services...")

	composePath := filepath.Join(lm.projectRoot, lm.composeFile)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "down")
	cmd.Dir = lm.projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Stop failed: %v\n%s", err, string(output)))
	}

	// Verify containers stopped
	for _, containerName := range lm.manager.doomContainers {
		status := ServiceStatus{
			Container: containerName,
		}

		output, err := exec.CommandContext(ctx, "docker", "inspect",
			"--format", "{{.State.Status}}", containerName).Output()
		if err != nil {
			// Container doesn't exist, which is fine
			status.State = StateStopped
		} else {
			state := strings.TrimSpace(string(output))
			if state == "running" {
				status.State = StateRunning
				// Force stop
				exec.CommandContext(ctx, "docker", "stop", "-t", "5", containerName).Run()
			} else {
				status.State = StateStopped
			}
		}

		result.Services = append(result.Services, status)
	}

	result.Success = len(result.Errors) == 0
	result.Duration = time.Since(result.StoppedAt)

	if result.Success {
		lm.log(LogInfo, "shutdown", "All services stopped")
	}

	return result, nil
}

// Restart restarts the doom-coding services
func (lm *LifecycleManager) Restart(ctx context.Context) (*StartupResult, error) {
	// Stop first
	_, err := lm.Stop(ctx)
	if err != nil {
		lm.log(LogWarning, "restart", fmt.Sprintf("Stop had issues: %v", err))
	}

	// Small delay
	time.Sleep(2 * time.Second)

	// Start again
	return lm.Start(ctx, nil)
}

// Status returns the current status of all services
func (lm *LifecycleManager) Status(ctx context.Context) []ServiceStatus {
	var statuses []ServiceStatus

	containers := []struct {
		name      string
		container string
		port      int
	}{
		{"Tailscale", "doom-tailscale", 0},
		{"code-server", "doom-code-server", 8443},
		{"Claude", "doom-claude", 7681},
	}

	for _, c := range containers {
		status := ServiceStatus{
			Name:      c.name,
			Container: c.container,
			Port:      c.port,
		}

		output, err := exec.CommandContext(ctx, "docker", "inspect",
			"--format", "{{.State.Status}},{{.State.Health.Status}}",
			c.container).Output()
		if err != nil {
			status.State = StateStopped
			statuses = append(statuses, status)
			continue
		}

		parts := strings.Split(strings.TrimSpace(string(output)), ",")
		runState := parts[0]
		healthState := ""
		if len(parts) > 1 {
			healthState = parts[1]
		}

		switch runState {
		case "running":
			if healthState == "healthy" {
				status.State = StateHealthy
			} else if healthState == "unhealthy" {
				status.State = StateUnhealthy
			} else {
				status.State = StateRunning
			}
		case "exited", "dead":
			status.State = StateStopped
		default:
			status.State = StateUnknown
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// log logs a message if logger is available
func (lm *LifecycleManager) log(level LogLevel, source, message string) {
	if lm.logger != nil {
		lm.logger.Log(level, source, message)
	}
}
