// Package service provides service lifecycle management for doom-coding
// including detection of existing services, conflict resolution, and graceful startup/shutdown
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ServiceState represents the current state of a service
type ServiceState string

const (
	StateUnknown  ServiceState = "unknown"
	StateStopped  ServiceState = "stopped"
	StateStarting ServiceState = "starting"
	StateRunning  ServiceState = "running"
	StateHealthy  ServiceState = "healthy"
	StateUnhealthy ServiceState = "unhealthy"
	StateStopping ServiceState = "stopping"
)

// ServiceType identifies what kind of service this is
type ServiceType string

const (
	TypeDoomCoding   ServiceType = "doom-coding"     // Our managed service
	TypeExternal     ServiceType = "external"        // Unknown external service
	TypeCodeServer   ServiceType = "code-server"     // Any code-server instance
	TypeTailscale    ServiceType = "tailscale"       // Tailscale daemon
)

// ServiceInfo contains information about a detected service
type ServiceInfo struct {
	Name          string       `json:"name"`
	Type          ServiceType  `json:"type"`
	State         ServiceState `json:"state"`
	ContainerID   string       `json:"container_id,omitempty"`
	ContainerName string       `json:"container_name,omitempty"`
	Port          int          `json:"port,omitempty"`
	Protocol      string       `json:"protocol,omitempty"` // tcp, udp
	PID           int          `json:"pid,omitempty"`
	ProcessName   string       `json:"process_name,omitempty"`
	Version       string       `json:"version,omitempty"`
	IsDoomManaged bool         `json:"is_doom_managed"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// PortConflict represents a port that is already in use
type PortConflict struct {
	Port            int          `json:"port"`
	Protocol        string       `json:"protocol"`
	RequestedBy     string       `json:"requested_by"`      // Service that wants this port
	OccupiedBy      *ServiceInfo `json:"occupied_by"`       // Service currently using port
	CanResolve      bool         `json:"can_resolve"`       // Whether we can auto-resolve
	ResolutionHint  string       `json:"resolution_hint"`   // User-friendly suggestion
}

// ConflictResolution represents how to resolve a conflict
type ConflictResolution int

const (
	ResolutionNone      ConflictResolution = iota // No resolution needed
	ResolutionRelocate                            // Move doom-coding to different port
	ResolutionMigrate                             // Migrate existing service to doom-coding
	ResolutionStop                                // Stop conflicting service
	ResolutionSkip                                // Skip this service
	ResolutionManual                              // Requires manual intervention
)

// Manager handles service detection and lifecycle
type Manager struct {
	projectRoot      string
	defaultPorts     map[string]int
	portRange        PortRange
	doomContainers   []string
	verbose          bool
}

// PortRange defines the range for dynamic port allocation
type PortRange struct {
	Start int
	End   int
}

// DefaultPortRange is the range used for automatic port allocation
var DefaultPortRange = PortRange{Start: 8000, End: 9000}

// NewManager creates a new service manager
func NewManager(projectRoot string) *Manager {
	return &Manager{
		projectRoot: projectRoot,
		defaultPorts: map[string]int{
			"code-server": 8443,
			"ttyd":        7681,
			"tailscale":   0, // No port exposed directly
		},
		portRange: DefaultPortRange,
		doomContainers: []string{
			"doom-tailscale",
			"doom-code-server",
			"doom-claude",
		},
	}
}

// SetVerbose enables verbose logging
func (m *Manager) SetVerbose(verbose bool) {
	m.verbose = verbose
}

// DetectExistingServices scans for all doom-coding related services
func (m *Manager) DetectExistingServices(ctx context.Context) ([]ServiceInfo, error) {
	var services []ServiceInfo

	// 1. Check Docker containers
	dockerServices, err := m.detectDockerServices(ctx)
	if err == nil {
		services = append(services, dockerServices...)
	}

	// 2. Check host processes on relevant ports
	portServices, err := m.detectPortServices(ctx)
	if err == nil {
		services = append(services, portServices...)
	}

	// 3. Check host Tailscale
	if ts := m.detectHostTailscale(ctx); ts != nil {
		services = append(services, *ts)
	}

	return m.deduplicateServices(services), nil
}

// detectDockerServices finds all Docker containers with doom-coding labels
func (m *Manager) detectDockerServices(ctx context.Context) ([]ServiceInfo, error) {
	// Check if Docker is available
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err != nil {
		return nil, fmt.Errorf("docker not available: %w", err)
	}

	// List containers with doom-coding labels or names
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var services []ServiceInfo
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}

		var container struct {
			ID      string `json:"ID"`
			Names   string `json:"Names"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Ports   string `json:"Ports"`
			Labels  string `json:"Labels"`
			Image   string `json:"Image"`
		}
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue
		}

		// Check if this is a doom-coding container
		isDoom := false
		for _, name := range m.doomContainers {
			if strings.Contains(container.Names, name) {
				isDoom = true
				break
			}
		}
		if strings.Contains(container.Labels, "com.doom-coding") {
			isDoom = true
		}

		// Also check for any code-server container
		isCodeServer := strings.Contains(container.Image, "code-server") ||
			strings.Contains(container.Names, "code-server")

		if isDoom || isCodeServer {
			svc := ServiceInfo{
				Name:          container.Names,
				ContainerID:   container.ID,
				ContainerName: container.Names,
				IsDoomManaged: isDoom,
				Labels:        parseLabels(container.Labels),
			}

			// Determine service type
			if strings.Contains(container.Names, "tailscale") {
				svc.Type = TypeTailscale
			} else if isCodeServer {
				svc.Type = TypeCodeServer
			} else {
				svc.Type = TypeDoomCoding
			}

			// Parse state
			switch strings.ToLower(container.State) {
			case "running":
				svc.State = StateRunning
			case "exited", "dead":
				svc.State = StateStopped
			case "created":
				svc.State = StateStopped
			default:
				svc.State = StateUnknown
			}

			// Parse ports
			if port := parseFirstPort(container.Ports); port > 0 {
				svc.Port = port
				svc.Protocol = "tcp"
			}

			services = append(services, svc)
		}
	}

	return services, nil
}

// detectPortServices checks which processes are using our target ports
func (m *Manager) detectPortServices(ctx context.Context) ([]ServiceInfo, error) {
	var services []ServiceInfo

	portsToCheck := []int{8443, 7681}
	for _, port := range portsToCheck {
		svc := m.checkPort(ctx, port)
		if svc != nil {
			services = append(services, *svc)
		}
	}

	return services, nil
}

// checkPort checks if a specific port is in use
func (m *Manager) checkPort(ctx context.Context, port int) *ServiceInfo {
	// Try to bind to the port to check availability
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		listener.Close()
		return nil // Port is free
	}

	// Port is in use, try to identify the process
	svc := &ServiceInfo{
		Port:     port,
		Protocol: "tcp",
		State:    StateRunning,
	}

	// Use lsof to identify the process (Linux)
	cmd := exec.CommandContext(ctx, "lsof", "-i", fmt.Sprintf(":%d", port), "-P", "-n", "-t")
	if output, err := cmd.Output(); err == nil {
		pid, _ := strconv.Atoi(strings.TrimSpace(string(output)))
		if pid > 0 {
			svc.PID = pid
			// Get process name
			if name, err := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "comm=").Output(); err == nil {
				svc.ProcessName = strings.TrimSpace(string(name))
			}
		}
	}

	// Try ss as fallback
	if svc.PID == 0 {
		cmd = exec.CommandContext(ctx, "ss", "-tlpn", fmt.Sprintf("sport = :%d", port))
		if output, err := cmd.Output(); err == nil {
			svc.Name = fmt.Sprintf("Process on port %d", port)
			if strings.Contains(string(output), "code-server") {
				svc.Type = TypeCodeServer
				svc.Name = "code-server (external)"
			}
		}
	}

	// Determine if this is a doom-managed service
	if svc.ProcessName != "" {
		svc.Name = svc.ProcessName
	}
	if svc.Name == "" {
		svc.Name = fmt.Sprintf("Unknown service on port %d", port)
	}
	svc.Type = TypeExternal

	return svc
}

// detectHostTailscale checks for Tailscale running on the host
func (m *Manager) detectHostTailscale(ctx context.Context) *ServiceInfo {
	// Check if tailscale command exists
	if _, err := exec.LookPath("tailscale"); err != nil {
		return nil
	}

	svc := &ServiceInfo{
		Name: "Host Tailscale",
		Type: TypeTailscale,
	}

	// Check status
	cmd := exec.CommandContext(ctx, "tailscale", "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		svc.State = StateStopped
		return svc
	}

	var status struct {
		BackendState string `json:"BackendState"`
		Version      string `json:"Version"`
		Self         struct {
			TailscaleIPs []string `json:"TailscaleIPs"`
		} `json:"Self"`
	}
	if err := json.Unmarshal(output, &status); err == nil {
		if status.BackendState == "Running" {
			svc.State = StateRunning
		} else {
			svc.State = StateStopped
		}
		svc.Version = status.Version
	}

	return svc
}

// CheckPortConflicts checks for port conflicts with our target configuration
func (m *Manager) CheckPortConflicts(ctx context.Context, targetPorts map[string]int) ([]PortConflict, error) {
	var conflicts []PortConflict

	existingServices, err := m.DetectExistingServices(ctx)
	if err != nil {
		return nil, err
	}

	// Build a map of occupied ports
	occupiedPorts := make(map[int]*ServiceInfo)
	for i := range existingServices {
		svc := &existingServices[i]
		if svc.Port > 0 && svc.State == StateRunning {
			occupiedPorts[svc.Port] = svc
		}
	}

	// Check each target port
	for serviceName, port := range targetPorts {
		if occupier, exists := occupiedPorts[port]; exists {
			conflict := PortConflict{
				Port:        port,
				Protocol:    "tcp",
				RequestedBy: serviceName,
				OccupiedBy:  occupier,
			}

			// Determine resolution strategy
			if occupier.IsDoomManaged {
				// This is our own service, probably from a previous installation
				conflict.CanResolve = true
				conflict.ResolutionHint = "Previous doom-coding installation detected. Will upgrade/restart."
			} else if occupier.Type == TypeCodeServer {
				// External code-server
				conflict.CanResolve = true
				conflict.ResolutionHint = fmt.Sprintf("Existing code-server found. Suggest relocating to port %d or migrating.", m.findFreePort(ctx, port))
			} else {
				// Unknown service
				conflict.CanResolve = false
				conflict.ResolutionHint = fmt.Sprintf("Port %d in use by %s. Consider using --port=%d or stop the conflicting service.", port, occupier.Name, m.findFreePort(ctx, port))
			}

			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, nil
}

// findFreePort finds the next available port starting from the preferred port
func (m *Manager) findFreePort(ctx context.Context, preferred int) int {
	// Try preferred port first
	if m.isPortFree(preferred) {
		return preferred
	}

	// Search in the configured range
	for port := m.portRange.Start; port <= m.portRange.End; port++ {
		if m.isPortFree(port) {
			return port
		}
	}

	// Fallback: let the OS assign
	return 0
}

// isPortFree checks if a port is available
func (m *Manager) isPortFree(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// FindAvailablePorts returns available ports for all services
func (m *Manager) FindAvailablePorts(ctx context.Context) (map[string]int, error) {
	result := make(map[string]int)

	for service, defaultPort := range m.defaultPorts {
		if defaultPort == 0 {
			result[service] = 0
			continue
		}
		result[service] = m.findFreePort(ctx, defaultPort)
	}

	return result, nil
}

// StopDoomServices gracefully stops all doom-coding containers
func (m *Manager) StopDoomServices(ctx context.Context, timeout time.Duration) error {
	for _, containerName := range m.doomContainers {
		// Check if container exists
		if err := exec.CommandContext(ctx, "docker", "inspect", containerName).Run(); err != nil {
			continue // Container doesn't exist
		}

		// Stop with timeout
		stopCtx, cancel := context.WithTimeout(ctx, timeout)
		cmd := exec.CommandContext(stopCtx, "docker", "stop", "-t", strconv.Itoa(int(timeout.Seconds())), containerName)
		err := cmd.Run()
		cancel()

		if err != nil {
			// Force kill if graceful stop failed
			exec.CommandContext(ctx, "docker", "kill", containerName).Run()
		}
	}

	return nil
}

// RemoveDoomContainers removes all doom-coding containers
func (m *Manager) RemoveDoomContainers(ctx context.Context) error {
	for _, containerName := range m.doomContainers {
		exec.CommandContext(ctx, "docker", "rm", "-f", containerName).Run()
	}
	return nil
}

// deduplicateServices removes duplicate service entries
func (m *Manager) deduplicateServices(services []ServiceInfo) []ServiceInfo {
	seen := make(map[string]bool)
	var result []ServiceInfo

	for _, svc := range services {
		key := fmt.Sprintf("%s-%s-%d", svc.Type, svc.ContainerName, svc.Port)
		if svc.ContainerName == "" {
			key = fmt.Sprintf("%s-%d-%d", svc.Type, svc.Port, svc.PID)
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, svc)
		}
	}

	return result
}

// parseLabels parses Docker label string into a map
func parseLabels(labelStr string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(labelStr, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// parseFirstPort extracts the first port number from Docker port string
func parseFirstPort(portStr string) int {
	// Format: "0.0.0.0:8443->8443/tcp, ..."
	if portStr == "" {
		return 0
	}
	parts := strings.Split(portStr, "->")
	if len(parts) < 2 {
		return 0
	}
	hostPart := parts[0]
	portParts := strings.Split(hostPart, ":")
	if len(portParts) < 2 {
		return 0
	}
	port, _ := strconv.Atoi(portParts[len(portParts)-1])
	return port
}
