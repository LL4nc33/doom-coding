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

// MigrationStrategy defines how to handle an existing installation
type MigrationStrategy int

const (
	StrategyFresh    MigrationStrategy = iota // Fresh install, no existing data
	StrategyUpgrade                           // Upgrade existing doom-coding
	StrategyMigrate                           // Migrate from external code-server
	StrategyParallel                          // Run alongside existing (different ports)
)

// MigrationPlan describes how to handle existing installations
type MigrationPlan struct {
	Strategy         MigrationStrategy      `json:"strategy"`
	ExistingServices []ServiceInfo          `json:"existing_services"`
	Actions          []MigrationAction      `json:"actions"`
	PortMappings     map[string]int         `json:"port_mappings"`
	Warnings         []string               `json:"warnings"`
	RequiresConfirm  bool                   `json:"requires_confirm"`
}

// MigrationAction represents a single action in the migration
type MigrationAction struct {
	Order       int    `json:"order"`
	Type        string `json:"type"` // "backup", "stop", "remove", "migrate_data", "start"
	Target      string `json:"target"`
	Description string `json:"description"`
	Reversible  bool   `json:"reversible"`
}

// MigrationResult contains the outcome of a migration
type MigrationResult struct {
	Success       bool              `json:"success"`
	CompletedAt   time.Time         `json:"completed_at"`
	Actions       []ActionResult    `json:"actions"`
	BackupPath    string            `json:"backup_path,omitempty"`
	Error         error             `json:"error,omitempty"`
}

// ActionResult is the result of a single migration action
type ActionResult struct {
	Action    MigrationAction `json:"action"`
	Success   bool            `json:"success"`
	Output    string          `json:"output,omitempty"`
	Error     string          `json:"error,omitempty"`
	Duration  time.Duration   `json:"duration"`
}

// Migrator handles migration from existing installations
type Migrator struct {
	manager     *Manager
	projectRoot string
	backupDir   string
	dryRun      bool
}

// NewMigrator creates a new migrator
func NewMigrator(manager *Manager, projectRoot string) *Migrator {
	return &Migrator{
		manager:     manager,
		projectRoot: projectRoot,
		backupDir:   filepath.Join(projectRoot, ".migration-backup"),
	}
}

// SetDryRun enables dry-run mode
func (m *Migrator) SetDryRun(dryRun bool) {
	m.dryRun = dryRun
}

// AnalyzeExisting analyzes existing services and creates a migration plan
func (m *Migrator) AnalyzeExisting(ctx context.Context, targetPorts map[string]int) (*MigrationPlan, error) {
	plan := &MigrationPlan{
		PortMappings: targetPorts,
	}

	// Detect existing services
	services, err := m.manager.DetectExistingServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect services: %w", err)
	}
	plan.ExistingServices = services

	// Categorize services
	var doomServices, externalCodeServers, otherServices []ServiceInfo
	for _, svc := range services {
		if svc.IsDoomManaged {
			doomServices = append(doomServices, svc)
		} else if svc.Type == TypeCodeServer {
			externalCodeServers = append(externalCodeServers, svc)
		} else if svc.Port > 0 {
			otherServices = append(otherServices, svc)
		}
	}

	// Determine strategy
	if len(doomServices) > 0 {
		plan.Strategy = StrategyUpgrade
		plan.Actions = m.createUpgradeActions(doomServices)
	} else if len(externalCodeServers) > 0 {
		plan.Strategy = StrategyMigrate
		plan.Actions = m.createMigrateActions(externalCodeServers, targetPorts)
		plan.RequiresConfirm = true
		plan.Warnings = append(plan.Warnings,
			"External code-server detected. Migration will preserve your extensions and settings.")
	} else if len(otherServices) > 0 {
		// Check for port conflicts
		hasConflicts := false
		for _, svc := range otherServices {
			for _, port := range targetPorts {
				if svc.Port == port {
					hasConflicts = true
					plan.Warnings = append(plan.Warnings,
						fmt.Sprintf("Port %d is in use by %s. Doom-coding will use port %d instead.",
							port, svc.Name, m.manager.findFreePort(ctx, port)))
				}
			}
		}
		if hasConflicts {
			plan.Strategy = StrategyParallel
			plan.PortMappings = m.resolvePortConflicts(ctx, targetPorts, otherServices)
		} else {
			plan.Strategy = StrategyFresh
		}
	} else {
		plan.Strategy = StrategyFresh
	}

	return plan, nil
}

// createUpgradeActions creates actions for upgrading existing doom-coding
func (m *Migrator) createUpgradeActions(services []ServiceInfo) []MigrationAction {
	var actions []MigrationAction
	order := 1

	// Backup current state
	actions = append(actions, MigrationAction{
		Order:       order,
		Type:        "backup",
		Target:      "doom-coding-config",
		Description: "Backup current configuration and data",
		Reversible:  true,
	})
	order++

	// Stop existing containers
	for _, svc := range services {
		if svc.ContainerName != "" && svc.State == StateRunning {
			actions = append(actions, MigrationAction{
				Order:       order,
				Type:        "stop",
				Target:      svc.ContainerName,
				Description: fmt.Sprintf("Stop container %s", svc.ContainerName),
				Reversible:  true,
			})
			order++
		}
	}

	// Pull new images
	actions = append(actions, MigrationAction{
		Order:       order,
		Type:        "pull",
		Target:      "doom-coding-images",
		Description: "Pull latest container images",
		Reversible:  false,
	})
	order++

	// Start new containers
	actions = append(actions, MigrationAction{
		Order:       order,
		Type:        "start",
		Target:      "doom-coding",
		Description: "Start updated containers",
		Reversible:  true,
	})

	return actions
}

// createMigrateActions creates actions for migrating from external code-server
func (m *Migrator) createMigrateActions(services []ServiceInfo, targetPorts map[string]int) []MigrationAction {
	var actions []MigrationAction
	order := 1

	// Backup
	actions = append(actions, MigrationAction{
		Order:       order,
		Type:        "backup",
		Target:      "code-server-config",
		Description: "Backup code-server extensions and settings",
		Reversible:  true,
	})
	order++

	// Stop external code-server
	for _, svc := range services {
		if svc.State == StateRunning {
			actions = append(actions, MigrationAction{
				Order:       order,
				Type:        "stop",
				Target:      svc.ContainerName,
				Description: fmt.Sprintf("Stop external code-server (%s)", svc.Name),
				Reversible:  true,
			})
			order++
		}
	}

	// Migrate data
	actions = append(actions, MigrationAction{
		Order:       order,
		Type:        "migrate_data",
		Target:      "extensions",
		Description: "Migrate VS Code extensions to doom-coding",
		Reversible:  true,
	})
	order++

	actions = append(actions, MigrationAction{
		Order:       order,
		Type:        "migrate_data",
		Target:      "settings",
		Description: "Migrate VS Code settings to doom-coding",
		Reversible:  true,
	})
	order++

	// Start doom-coding
	actions = append(actions, MigrationAction{
		Order:       order,
		Type:        "start",
		Target:      "doom-coding",
		Description: "Start doom-coding containers",
		Reversible:  true,
	})

	return actions
}

// resolvePortConflicts adjusts port mappings to avoid conflicts
func (m *Migrator) resolvePortConflicts(ctx context.Context, targetPorts map[string]int, occupiedServices []ServiceInfo) map[string]int {
	result := make(map[string]int)

	// Build set of occupied ports
	occupied := make(map[int]bool)
	for _, svc := range occupiedServices {
		if svc.Port > 0 {
			occupied[svc.Port] = true
		}
	}

	for service, port := range targetPorts {
		if occupied[port] {
			// Find alternative
			newPort := m.manager.findFreePort(ctx, port)
			result[service] = newPort
		} else {
			result[service] = port
		}
	}

	return result
}

// Execute runs the migration plan
func (m *Migrator) Execute(ctx context.Context, plan *MigrationPlan) (*MigrationResult, error) {
	result := &MigrationResult{
		CompletedAt: time.Now(),
	}

	for _, action := range plan.Actions {
		if m.dryRun {
			result.Actions = append(result.Actions, ActionResult{
				Action:  action,
				Success: true,
				Output:  "[DRY RUN] Would execute",
			})
			continue
		}

		start := time.Now()
		actionResult := m.executeAction(ctx, action)
		actionResult.Duration = time.Since(start)
		result.Actions = append(result.Actions, actionResult)

		if !actionResult.Success {
			result.Error = fmt.Errorf("action '%s' failed: %s", action.Description, actionResult.Error)
			return result, result.Error
		}
	}

	result.Success = true
	return result, nil
}

// executeAction executes a single migration action
func (m *Migrator) executeAction(ctx context.Context, action MigrationAction) ActionResult {
	result := ActionResult{Action: action}

	switch action.Type {
	case "backup":
		err := m.backup(ctx, action.Target)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Output = fmt.Sprintf("Backup created at %s", m.backupDir)
		}

	case "stop":
		output, err := exec.CommandContext(ctx, "docker", "stop", "-t", "30", action.Target).CombinedOutput()
		if err != nil {
			result.Error = fmt.Sprintf("%v: %s", err, string(output))
		} else {
			result.Success = true
			result.Output = "Container stopped"
		}

	case "remove":
		output, err := exec.CommandContext(ctx, "docker", "rm", "-f", action.Target).CombinedOutput()
		if err != nil {
			result.Error = fmt.Sprintf("%v: %s", err, string(output))
		} else {
			result.Success = true
			result.Output = "Container removed"
		}

	case "pull":
		// Pull images via docker compose
		composeFile := filepath.Join(m.projectRoot, "docker-compose.yml")
		output, err := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "pull").CombinedOutput()
		if err != nil {
			result.Error = fmt.Sprintf("%v: %s", err, string(output))
		} else {
			result.Success = true
			result.Output = "Images pulled"
		}

	case "migrate_data":
		err := m.migrateData(ctx, action.Target)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Output = fmt.Sprintf("Migrated %s", action.Target)
		}

	case "start":
		composeFile := filepath.Join(m.projectRoot, "docker-compose.yml")
		output, err := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "up", "-d").CombinedOutput()
		if err != nil {
			result.Error = fmt.Sprintf("%v: %s", err, string(output))
		} else {
			result.Success = true
			result.Output = "Containers started"
		}

	default:
		result.Error = fmt.Sprintf("unknown action type: %s", action.Type)
	}

	return result
}

// backup creates a backup of the specified target
func (m *Migrator) backup(ctx context.Context, target string) error {
	// Create backup directory
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(m.backupDir, timestamp)
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	switch target {
	case "doom-coding-config":
		// Backup .env file
		if err := copyFile(
			filepath.Join(m.projectRoot, ".env"),
			filepath.Join(backupPath, ".env"),
		); err != nil && !os.IsNotExist(err) {
			return err
		}

		// Backup Docker volumes
		for _, volume := range []string{"doom-code-server-config", "doom-claude-config"} {
			m.backupVolume(ctx, volume, backupPath)
		}

	case "code-server-config":
		// Try to find code-server config in common locations
		configPaths := []string{
			"/config/.local/share/code-server",
			filepath.Join(os.Getenv("HOME"), ".local/share/code-server"),
			"/home/coder/.local/share/code-server",
		}
		for _, path := range configPaths {
			if _, err := os.Stat(path); err == nil {
				// Copy recursively
				exec.CommandContext(ctx, "cp", "-r", path, filepath.Join(backupPath, "code-server")).Run()
				break
			}
		}
	}

	return nil
}

// backupVolume backs up a Docker volume
func (m *Migrator) backupVolume(ctx context.Context, volumeName, backupPath string) error {
	tarPath := filepath.Join(backupPath, volumeName+".tar")

	// Use a temporary container to access the volume
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", volumeName+":/data",
		"-v", backupPath+":/backup",
		"alpine",
		"tar", "cvf", "/backup/"+volumeName+".tar", "-C", "/data", ".")

	return cmd.Run()
}

// migrateData migrates data from old installation
func (m *Migrator) migrateData(ctx context.Context, target string) error {
	switch target {
	case "extensions":
		// Find source extensions
		sourcePaths := []string{
			"/config/.local/share/code-server/extensions",
			filepath.Join(os.Getenv("HOME"), ".local/share/code-server/extensions"),
		}

		for _, source := range sourcePaths {
			if _, err := os.Stat(source); err == nil {
				// Copy to doom-coding volume via docker cp
				cmd := exec.CommandContext(ctx, "docker", "cp",
					source,
					"doom-code-server:/config/.local/share/code-server/")
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to migrate extensions: %w", err)
				}
				return nil
			}
		}

	case "settings":
		// Find source settings
		sourcePaths := []string{
			"/config/.local/share/code-server/User/settings.json",
			filepath.Join(os.Getenv("HOME"), ".local/share/code-server/User/settings.json"),
		}

		for _, source := range sourcePaths {
			if _, err := os.Stat(source); err == nil {
				cmd := exec.CommandContext(ctx, "docker", "cp",
					source,
					"doom-code-server:/config/.local/share/code-server/User/")
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to migrate settings: %w", err)
				}
				return nil
			}
		}
	}

	return nil
}

// Rollback attempts to rollback a failed migration
func (m *Migrator) Rollback(ctx context.Context, result *MigrationResult) error {
	// Execute actions in reverse order
	for i := len(result.Actions) - 1; i >= 0; i-- {
		action := result.Actions[i]
		if !action.Action.Reversible {
			continue
		}

		switch action.Action.Type {
		case "stop":
			// Restart the container
			exec.CommandContext(ctx, "docker", "start", action.Action.Target).Run()
		case "remove":
			// Cannot easily restore a removed container
			// Would need to restore from backup
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// GetMigrationSummary returns a human-readable summary of the migration plan
func (plan *MigrationPlan) GetMigrationSummary() string {
	var sb strings.Builder

	sb.WriteString("Migration Plan Summary\n")
	sb.WriteString("======================\n\n")

	// Strategy
	sb.WriteString(fmt.Sprintf("Strategy: %s\n\n", plan.getStrategyName()))

	// Existing services
	if len(plan.ExistingServices) > 0 {
		sb.WriteString("Detected Services:\n")
		for _, svc := range plan.ExistingServices {
			status := string(svc.State)
			if svc.IsDoomManaged {
				status += " (doom-managed)"
			}
			sb.WriteString(fmt.Sprintf("  - %s [%s]\n", svc.Name, status))
		}
		sb.WriteString("\n")
	}

	// Actions
	if len(plan.Actions) > 0 {
		sb.WriteString("Planned Actions:\n")
		for _, action := range plan.Actions {
			reversible := ""
			if action.Reversible {
				reversible = " [reversible]"
			}
			sb.WriteString(fmt.Sprintf("  %d. %s%s\n", action.Order, action.Description, reversible))
		}
		sb.WriteString("\n")
	}

	// Port mappings
	sb.WriteString("Port Configuration:\n")
	for service, port := range plan.PortMappings {
		sb.WriteString(fmt.Sprintf("  - %s: %d\n", service, port))
	}
	sb.WriteString("\n")

	// Warnings
	if len(plan.Warnings) > 0 {
		sb.WriteString("Warnings:\n")
		for _, warning := range plan.Warnings {
			sb.WriteString(fmt.Sprintf("  ! %s\n", warning))
		}
	}

	return sb.String()
}

func (plan *MigrationPlan) getStrategyName() string {
	switch plan.Strategy {
	case StrategyFresh:
		return "Fresh Installation"
	case StrategyUpgrade:
		return "Upgrade Existing"
	case StrategyMigrate:
		return "Migrate from External"
	case StrategyParallel:
		return "Parallel Installation"
	default:
		return "Unknown"
	}
}
