package system

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// SystemInfo contains detected system information
type SystemInfo struct {
	// Basic info
	Hostname     string
	Username     string
	HomeDir      string
	OS           string
	Arch         string
	Distribution string
	Version      string

	// Container detection
	IsContainer bool
	IsLXC       bool
	IsDocker    bool
	IsWSL       bool

	// Device availability
	HasTUN      bool
	TUNPath     string

	// Software detection
	DockerInstalled    bool
	DockerRunning      bool
	DockerVersion      string
	TailscaleInstalled bool
	TailscaleRunning   bool
	TailscaleIP        string
	ZshInstalled       bool
	TmuxInstalled      bool

	// Network info
	LocalIPs      []string
	DefaultGateway string

	// Resources
	DiskFreeGB   float64
	MemoryTotalGB float64
}

// DetectSystem performs comprehensive system detection
func DetectSystem() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Basic info
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	if currentUser, err := user.Current(); err == nil {
		info.Username = currentUser.Username
		info.HomeDir = currentUser.HomeDir
	}

	// Detect distribution
	detectDistribution(info)

	// Container detection
	detectContainer(info)

	// Device detection
	detectDevices(info)

	// Software detection
	detectSoftware(info)

	// Network detection
	detectNetwork(info)

	// Resource detection
	detectResources(info)

	return info, nil
}

func detectDistribution(info *SystemInfo) {
	// Try /etc/os-release first (most common)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ID=") {
				info.Distribution = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
			} else if strings.HasPrefix(line, "VERSION_ID=") {
				info.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
			}
		}
	}

	// Fallback to lsb_release
	if info.Distribution == "" {
		if output, err := exec.Command("lsb_release", "-si").Output(); err == nil {
			info.Distribution = strings.TrimSpace(string(output))
		}
	}
}

func detectContainer(info *SystemInfo) {
	// Check for LXC
	if _, err := os.Stat("/dev/lxc"); err == nil {
		info.IsLXC = true
		info.IsContainer = true
	}

	// Check cgroup for container hints
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "lxc") {
			info.IsLXC = true
			info.IsContainer = true
		}
		if strings.Contains(content, "docker") {
			info.IsDocker = true
			info.IsContainer = true
		}
	}

	// Check for /.dockerenv
	if _, err := os.Stat("/.dockerenv"); err == nil {
		info.IsDocker = true
		info.IsContainer = true
	}

	// Check for WSL
	if data, err := os.ReadFile("/proc/version"); err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "microsoft") || strings.Contains(content, "wsl") {
			info.IsWSL = true
		}
	}
}

func detectDevices(info *SystemInfo) {
	// Check TUN device
	tunPaths := []string{"/dev/net/tun", "/dev/tun"}
	for _, path := range tunPaths {
		if _, err := os.Stat(path); err == nil {
			info.HasTUN = true
			info.TUNPath = path
			break
		}
	}

	// Additional TUN check - try to open the device
	if !info.HasTUN {
		// In LXC, the device might exist but not be accessible
		// We can check if the module is loaded
		if data, err := os.ReadFile("/proc/modules"); err == nil {
			if strings.Contains(string(data), "tun") {
				info.HasTUN = true
				info.TUNPath = "/dev/net/tun"
			}
		}
	}
}

func detectSoftware(info *SystemInfo) {
	// Docker
	if path, err := exec.LookPath("docker"); err == nil {
		info.DockerInstalled = true
		// Get version
		if output, err := exec.Command(path, "--version").Output(); err == nil {
			parts := strings.Split(string(output), ",")
			if len(parts) > 0 {
				info.DockerVersion = strings.TrimSpace(parts[0])
			}
		}
		// Check if running
		if err := exec.Command(path, "info").Run(); err == nil {
			info.DockerRunning = true
		}
	}

	// Tailscale
	if _, err := exec.LookPath("tailscale"); err == nil {
		info.TailscaleInstalled = true
		// Check if connected
		if output, err := exec.Command("tailscale", "status", "--json").Output(); err == nil {
			if strings.Contains(string(output), `"BackendState":"Running"`) {
				info.TailscaleRunning = true
			}
		}
		// Get IP
		if output, err := exec.Command("tailscale", "ip", "-4").Output(); err == nil {
			info.TailscaleIP = strings.TrimSpace(string(output))
		}
	}

	// Zsh
	if _, err := exec.LookPath("zsh"); err == nil {
		info.ZshInstalled = true
	}

	// Tmux
	if _, err := exec.LookPath("tmux"); err == nil {
		info.TmuxInstalled = true
	}
}

func detectNetwork(info *SystemInfo) {
	// Get local IPs
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					info.LocalIPs = append(info.LocalIPs, ipnet.IP.String())
				}
			}
		}
	}

	// Get default gateway (Linux-specific)
	if data, err := os.ReadFile("/proc/net/route"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 3 && fields[1] == "00000000" {
				// Found default route, parse gateway
				gateway := parseHexIP(fields[2])
				if gateway != "" {
					info.DefaultGateway = gateway
				}
				break
			}
		}
	}
}

func detectResources(info *SystemInfo) {
	// Disk space
	var stat syscallStatfs
	if err := statfs(".", &stat); err == nil {
		info.DiskFreeGB = float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024 * 1024)
	}

	// Memory
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					var kb int64
					fmt.Sscanf(fields[1], "%d", &kb)
					info.MemoryTotalGB = float64(kb) / (1024 * 1024)
				}
				break
			}
		}
	}
}

func parseHexIP(hex string) string {
	if len(hex) != 8 {
		return ""
	}
	var ip [4]byte
	for i := 0; i < 4; i++ {
		var b byte
		fmt.Sscanf(hex[i*2:i*2+2], "%x", &b)
		ip[3-i] = b
	}
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

// GetRecommendedMode returns the recommended deployment mode based on system detection
func (s *SystemInfo) GetRecommendedMode() string {
	if !s.HasTUN && s.IsLXC {
		return "local" // LXC without TUN can't use Tailscale in kernel mode
	}
	if s.TailscaleInstalled && s.TailscaleRunning {
		return "tailscale" // Already have Tailscale
	}
	if s.HasTUN {
		return "tailscale" // TUN available, Tailscale is recommended
	}
	return "local" // Default to local for safety
}

// GetWarnings returns any warnings based on system state
func (s *SystemInfo) GetWarnings() []string {
	var warnings []string

	if s.IsLXC && !s.HasTUN {
		warnings = append(warnings, "LXC container without TUN device - Tailscale VPN not available")
	}

	if s.DiskFreeGB < 10 {
		warnings = append(warnings, fmt.Sprintf("Low disk space: %.1f GB free", s.DiskFreeGB))
	}

	if s.MemoryTotalGB < 2 {
		warnings = append(warnings, fmt.Sprintf("Low memory: %.1f GB total", s.MemoryTotalGB))
	}

	if s.DockerInstalled && !s.DockerRunning {
		warnings = append(warnings, "Docker installed but not running")
	}

	return warnings
}

// String returns a human-readable summary
func (s *SystemInfo) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Host: %s (%s)\n", s.Hostname, s.Username))
	sb.WriteString(fmt.Sprintf("OS: %s/%s", s.OS, s.Arch))
	if s.Distribution != "" {
		sb.WriteString(fmt.Sprintf(" (%s", s.Distribution))
		if s.Version != "" {
			sb.WriteString(fmt.Sprintf(" %s", s.Version))
		}
		sb.WriteString(")")
	}
	sb.WriteString("\n")

	if s.IsContainer {
		containerType := "Unknown container"
		if s.IsLXC {
			containerType = "LXC"
		} else if s.IsDocker {
			containerType = "Docker"
		}
		sb.WriteString(fmt.Sprintf("Environment: %s\n", containerType))
	}

	tunStatus := "Available"
	if !s.HasTUN {
		tunStatus = "Not available"
	}
	sb.WriteString(fmt.Sprintf("TUN device: %s\n", tunStatus))

	if len(s.LocalIPs) > 0 {
		sb.WriteString(fmt.Sprintf("Local IPs: %s\n", strings.Join(s.LocalIPs, ", ")))
	}

	return sb.String()
}

// Syscall types for disk space (platform-specific)
type syscallStatfs struct {
	Bsize   int64
	Bavail  uint64
}

func statfs(path string, stat *syscallStatfs) error {
	// Use df command as a portable fallback
	output, err := exec.Command("df", "-B1", path).Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return fmt.Errorf("unexpected df fields")
	}

	var avail uint64
	fmt.Sscanf(fields[3], "%d", &avail)
	stat.Bsize = 1
	stat.Bavail = avail

	return nil
}
