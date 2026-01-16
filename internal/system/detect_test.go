package system

import (
	"os"
	"strings"
	"testing"
)

func TestDetectSystem(t *testing.T) {
	info, err := DetectSystem()
	if err != nil {
		t.Fatalf("DetectSystem returned error: %v", err)
	}

	if info == nil {
		t.Fatal("DetectSystem returned nil")
	}

	// Basic info should not be empty
	if info.OS == "" {
		t.Error("OS should not be empty")
	}

	if info.Arch == "" {
		t.Error("Arch should not be empty")
	}

	// Should be one of the known operating systems
	validOS := []string{"linux", "darwin", "windows"}
	found := false
	for _, os := range validOS {
		if info.OS == os {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("OS should be one of %v, got %s", validOS, info.OS)
	}

	// Should be one of the known architectures
	validArch := []string{"amd64", "arm64", "386", "arm"}
	found = false
	for _, arch := range validArch {
		if info.Arch == arch {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Arch should be one of %v, got %s", validArch, info.Arch)
	}
}

func TestDetectSystemHostname(t *testing.T) {
	info, err := DetectSystem()
	if err != nil {
		t.Fatalf("DetectSystem returned error: %v", err)
	}

	// Hostname should match os.Hostname() if available
	osHostname, err := os.Hostname()
	if err == nil && info.Hostname != osHostname {
		t.Errorf("Hostname mismatch: got %q, expected %q", info.Hostname, osHostname)
	}
}

func TestDetectSystemUserInfo(t *testing.T) {
	info, err := DetectSystem()
	if err != nil {
		t.Fatalf("DetectSystem returned error: %v", err)
	}

	// Username should not be empty if we're running in a normal environment
	if info.Username == "" {
		t.Log("Warning: Username is empty (may be expected in some environments)")
	}

	// HomeDir should exist if set
	if info.HomeDir != "" {
		if _, err := os.Stat(info.HomeDir); os.IsNotExist(err) {
			t.Errorf("HomeDir %q does not exist", info.HomeDir)
		}
	}
}

func TestDetectContainerType(t *testing.T) {
	info, err := DetectSystem()
	if err != nil {
		t.Fatalf("DetectSystem returned error: %v", err)
	}

	// If IsContainer is true, at least one container type should be set
	if info.IsContainer {
		if !info.IsLXC && !info.IsDocker && !info.IsWSL {
			t.Error("IsContainer is true but no container type is detected")
		}
	}

	// Container types should be mutually exclusive in most cases
	containerTypes := 0
	if info.IsLXC {
		containerTypes++
	}
	if info.IsDocker {
		containerTypes++
	}
	// WSL can coexist with others, so don't count it

	if containerTypes > 1 {
		t.Log("Warning: Multiple container types detected (unusual but possible)")
	}
}

func TestSystemInfoGetRecommendedMode(t *testing.T) {
	tests := []struct {
		name     string
		info     *SystemInfo
		wantMode string
	}{
		{
			name: "LXC without TUN",
			info: &SystemInfo{
				IsLXC:   true,
				HasTUN:  false,
			},
			wantMode: "local",
		},
		{
			name: "Tailscale already running",
			info: &SystemInfo{
				TailscaleInstalled: true,
				TailscaleRunning:   true,
				HasTUN:             true,
			},
			wantMode: "tailscale",
		},
		{
			name: "TUN available",
			info: &SystemInfo{
				HasTUN: true,
			},
			wantMode: "tailscale",
		},
		{
			name: "No TUN, not LXC",
			info: &SystemInfo{
				HasTUN: false,
				IsLXC:  false,
			},
			wantMode: "local",
		},
		{
			name: "Default case",
			info: &SystemInfo{},
			wantMode: "local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.GetRecommendedMode()
			if got != tt.wantMode {
				t.Errorf("GetRecommendedMode() = %q, want %q", got, tt.wantMode)
			}
		})
	}
}

func TestSystemInfoGetWarnings(t *testing.T) {
	tests := []struct {
		name         string
		info         *SystemInfo
		wantWarnings []string
	}{
		{
			name: "LXC without TUN",
			info: &SystemInfo{
				IsLXC:  true,
				HasTUN: false,
			},
			wantWarnings: []string{"LXC container without TUN device"},
		},
		{
			name: "Low disk space",
			info: &SystemInfo{
				DiskFreeGB: 5.0,
			},
			wantWarnings: []string{"Low disk space"},
		},
		{
			name: "Low memory",
			info: &SystemInfo{
				MemoryTotalGB: 1.5,
			},
			wantWarnings: []string{"Low memory"},
		},
		{
			name: "Docker installed but not running",
			info: &SystemInfo{
				DockerInstalled: true,
				DockerRunning:   false,
			},
			wantWarnings: []string{"Docker installed but not running"},
		},
		{
			name: "Multiple warnings",
			info: &SystemInfo{
				IsLXC:           true,
				HasTUN:          false,
				DiskFreeGB:      5.0,
				MemoryTotalGB:   1.0,
				DockerInstalled: true,
				DockerRunning:   false,
			},
			wantWarnings: []string{
				"LXC container without TUN device",
				"Low disk space",
				"Low memory",
				"Docker installed but not running",
			},
		},
		{
			name: "No warnings",
			info: &SystemInfo{
				HasTUN:          true,
				DiskFreeGB:      50.0,
				MemoryTotalGB:   8.0,
				DockerInstalled: true,
				DockerRunning:   true,
			},
			wantWarnings: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := tt.info.GetWarnings()

			if len(warnings) != len(tt.wantWarnings) {
				t.Errorf("GetWarnings() returned %d warnings, want %d: %v", len(warnings), len(tt.wantWarnings), warnings)
			}

			for _, wantWarning := range tt.wantWarnings {
				found := false
				for _, warning := range warnings {
					if strings.Contains(warning, wantWarning) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected warning containing %q not found in %v", wantWarning, warnings)
				}
			}
		})
	}
}

func TestSystemInfoString(t *testing.T) {
	info := &SystemInfo{
		Hostname:     "testhost",
		Username:     "testuser",
		OS:           "linux",
		Arch:         "amd64",
		Distribution: "debian",
		Version:      "12",
		IsContainer:  true,
		IsLXC:        true,
		HasTUN:       true,
		TUNPath:      "/dev/net/tun",
		LocalIPs:     []string{"192.168.1.100", "10.0.0.1"},
	}

	str := info.String()

	// Check required content
	expectedContent := []string{
		"testhost",
		"testuser",
		"linux",
		"amd64",
		"debian",
		"12",
		"LXC",
		"Available",
		"192.168.1.100",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(str, expected) {
			t.Errorf("String() should contain %q, got: %s", expected, str)
		}
	}
}

func TestSystemInfoStringMinimal(t *testing.T) {
	info := &SystemInfo{
		Hostname: "minimal",
		Username: "user",
		OS:       "linux",
		Arch:     "amd64",
	}

	str := info.String()

	// Should not panic with minimal info
	if str == "" {
		t.Error("String() should not return empty string")
	}

	// Should contain basic info
	if !strings.Contains(str, "minimal") {
		t.Error("String() should contain hostname")
	}
	if !strings.Contains(str, "linux/amd64") {
		t.Error("String() should contain OS/Arch")
	}
}

func TestParseHexIP(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"0100007F", "127.0.0.1"},
		{"C0A80101", "192.168.1.1"},
		{"0A000001", "10.0.0.1"},
		{"FFFFFFFF", "255.255.255.255"},
		{"00000000", "0.0.0.0"},
		// Invalid cases
		{"invalid", ""},
		{"1234", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			got := parseHexIP(tt.hex)
			if got != tt.want {
				t.Errorf("parseHexIP(%q) = %q, want %q", tt.hex, got, tt.want)
			}
		})
	}
}

func TestDetectDistribution(t *testing.T) {
	info := &SystemInfo{}
	detectDistribution(info)

	// On Linux systems, we should get some distribution info
	if info.OS == "linux" {
		// Distribution might be empty on some minimal systems
		if info.Distribution != "" {
			t.Logf("Detected distribution: %s %s", info.Distribution, info.Version)
		}
	}
}

func TestDetectDevices(t *testing.T) {
	info := &SystemInfo{}
	detectDevices(info)

	// Just verify it doesn't panic
	t.Logf("HasTUN: %v, TUNPath: %s", info.HasTUN, info.TUNPath)
}

func TestDetectSoftware(t *testing.T) {
	info := &SystemInfo{}
	detectSoftware(info)

	// Just verify it doesn't panic and returns reasonable values
	t.Logf("Docker: installed=%v, running=%v, version=%s",
		info.DockerInstalled, info.DockerRunning, info.DockerVersion)
	t.Logf("Tailscale: installed=%v, running=%v, ip=%s",
		info.TailscaleInstalled, info.TailscaleRunning, info.TailscaleIP)
	t.Logf("Shell tools: zsh=%v, tmux=%v",
		info.ZshInstalled, info.TmuxInstalled)
}

func TestDetectNetwork(t *testing.T) {
	info := &SystemInfo{}
	detectNetwork(info)

	// Should find at least one IP on most systems
	t.Logf("LocalIPs: %v", info.LocalIPs)
	t.Logf("DefaultGateway: %s", info.DefaultGateway)

	// Verify IPs are valid format if present
	for _, ip := range info.LocalIPs {
		if ip == "" {
			t.Error("Empty IP in LocalIPs list")
		}
		// Basic format check
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			t.Errorf("Invalid IP format: %s", ip)
		}
	}
}

func TestDetectResources(t *testing.T) {
	info := &SystemInfo{}
	detectResources(info)

	// Should have positive values on most systems
	t.Logf("DiskFreeGB: %.2f", info.DiskFreeGB)
	t.Logf("MemoryTotalGB: %.2f", info.MemoryTotalGB)

	// These should generally be positive on a real system
	if info.DiskFreeGB < 0 {
		t.Error("DiskFreeGB should not be negative")
	}
	if info.MemoryTotalGB < 0 {
		t.Error("MemoryTotalGB should not be negative")
	}
}

func TestStatfs(t *testing.T) {
	var stat syscallStatfs
	err := statfs(".", &stat)

	// On most systems this should succeed for the current directory
	if err != nil {
		t.Logf("statfs failed (may be expected in some environments): %v", err)
		return
	}

	if stat.Bavail == 0 {
		t.Log("Warning: Bavail is 0 (unusual but possible on full disk)")
	}
}

func TestDetectContainerMethods(t *testing.T) {
	info := &SystemInfo{}
	detectContainer(info)

	// Just verify the detection doesn't panic
	t.Logf("Container detection: IsContainer=%v, IsLXC=%v, IsDocker=%v, IsWSL=%v",
		info.IsContainer, info.IsLXC, info.IsDocker, info.IsWSL)
}
