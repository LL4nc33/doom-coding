package service

import (
	"context"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/test-project")

	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	if m.projectRoot != "/tmp/test-project" {
		t.Errorf("Expected projectRoot /tmp/test-project, got %s", m.projectRoot)
	}

	if len(m.doomContainers) != 3 {
		t.Errorf("Expected 3 doom containers, got %d", len(m.doomContainers))
	}

	expectedContainers := []string{"doom-tailscale", "doom-code-server", "doom-claude"}
	for i, expected := range expectedContainers {
		if m.doomContainers[i] != expected {
			t.Errorf("Expected container %s at index %d, got %s", expected, i, m.doomContainers[i])
		}
	}
}

func TestIsPortFree(t *testing.T) {
	m := NewManager("/tmp/test")

	// Test that a random high port is free
	// Note: This could fail in unusual environments
	freePort := m.findFreePort(context.Background(), 19999)
	if freePort == 0 {
		t.Skip("Could not find a free port in test range")
	}

	if !m.isPortFree(freePort) {
		t.Errorf("Port %d should be free after findFreePort returned it", freePort)
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{
			input:    "key1=value1,key2=value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			input:    "com.doom-coding.service=tailscale",
			expected: map[string]string{"com.doom-coding.service": "tailscale"},
		},
		{
			input:    "",
			expected: map[string]string{},
		},
	}

	for _, tc := range tests {
		result := parseLabels(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("For input %q, expected %d labels, got %d", tc.input, len(tc.expected), len(result))
			continue
		}
		for k, v := range tc.expected {
			if result[k] != v {
				t.Errorf("For input %q, expected %s=%s, got %s=%s", tc.input, k, v, k, result[k])
			}
		}
	}
}

func TestParseFirstPort(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			input:    "0.0.0.0:8443->8443/tcp",
			expected: 8443,
		},
		{
			input:    "0.0.0.0:8443->8443/tcp, 0.0.0.0:7681->7681/tcp",
			expected: 8443,
		},
		{
			input:    "",
			expected: 0,
		},
		{
			input:    "invalid",
			expected: 0,
		},
	}

	for _, tc := range tests {
		result := parseFirstPort(tc.input)
		if result != tc.expected {
			t.Errorf("For input %q, expected %d, got %d", tc.input, tc.expected, result)
		}
	}
}

func TestServiceState(t *testing.T) {
	states := []struct {
		state    ServiceState
		expected string
	}{
		{StateUnknown, "unknown"},
		{StateStopped, "stopped"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StateHealthy, "healthy"},
		{StateUnhealthy, "unhealthy"},
		{StateStopping, "stopping"},
	}

	for _, s := range states {
		if string(s.state) != s.expected {
			t.Errorf("Expected state %s, got %s", s.expected, string(s.state))
		}
	}
}

func TestPortRange(t *testing.T) {
	if DefaultPortRange.Start != 8000 {
		t.Errorf("Expected default port range start 8000, got %d", DefaultPortRange.Start)
	}
	if DefaultPortRange.End != 9000 {
		t.Errorf("Expected default port range end 9000, got %d", DefaultPortRange.End)
	}
}

func TestDeduplicateServices(t *testing.T) {
	m := NewManager("/tmp/test")

	services := []ServiceInfo{
		{
			Name:          "doom-code-server",
			Type:          TypeCodeServer,
			ContainerName: "doom-code-server",
			Port:          8443,
		},
		{
			Name:          "doom-code-server",
			Type:          TypeCodeServer,
			ContainerName: "doom-code-server",
			Port:          8443,
		},
		{
			Name:          "external-service",
			Type:          TypeExternal,
			Port:          9000,
			PID:           1234,
		},
	}

	result := m.deduplicateServices(services)
	if len(result) != 2 {
		t.Errorf("Expected 2 deduplicated services, got %d", len(result))
	}
}

func TestMigrationStrategy(t *testing.T) {
	strategies := []struct {
		strategy MigrationStrategy
		expected int
	}{
		{StrategyFresh, 0},
		{StrategyUpgrade, 1},
		{StrategyMigrate, 2},
		{StrategyParallel, 3},
	}

	for _, s := range strategies {
		if int(s.strategy) != s.expected {
			t.Errorf("Expected strategy value %d, got %d", s.expected, int(s.strategy))
		}
	}
}

func TestMigrationPlanSummary(t *testing.T) {
	plan := &MigrationPlan{
		Strategy: StrategyFresh,
		PortMappings: map[string]int{
			"code-server": 8443,
		},
	}

	summary := plan.GetMigrationSummary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	if !containsString(summary, "Fresh Installation") {
		t.Error("Summary should contain strategy name")
	}
}

func TestLogLevel(t *testing.T) {
	levels := []struct {
		level    LogLevel
		expected string
	}{
		{LogDebug, "DEBUG"},
		{LogInfo, "INFO"},
		{LogWarning, "WARNING"},
		{LogError, "ERROR"},
		{LogProgress, "PROGRESS"},
	}

	for _, l := range levels {
		if l.level.String() != l.expected {
			t.Errorf("Expected level string %s, got %s", l.expected, l.level.String())
		}
	}
}

func TestServiceType(t *testing.T) {
	types := []struct {
		stype    ServiceType
		expected string
	}{
		{TypeDoomCoding, "doom-coding"},
		{TypeExternal, "external"},
		{TypeCodeServer, "code-server"},
		{TypeTailscale, "tailscale"},
	}

	for _, st := range types {
		if string(st.stype) != st.expected {
			t.Errorf("Expected type %s, got %s", st.expected, string(st.stype))
		}
	}
}

func TestPortConflict(t *testing.T) {
	conflict := PortConflict{
		Port:           8443,
		Protocol:       "tcp",
		RequestedBy:    "doom-code-server",
		CanResolve:     true,
		ResolutionHint: "Use alternative port",
	}

	if conflict.Port != 8443 {
		t.Errorf("Expected port 8443, got %d", conflict.Port)
	}
	if !conflict.CanResolve {
		t.Error("Expected CanResolve to be true")
	}
}

func TestLifecycleManagerTimeout(t *testing.T) {
	m := NewManager("/tmp/test")
	lm := NewLifecycleManager(m, "/tmp/test", "docker-compose.yml")

	// Default timeout
	if lm.timeout != 2*time.Minute {
		t.Errorf("Expected default timeout 2m, got %v", lm.timeout)
	}

	// Set custom timeout
	lm.SetTimeout(5 * time.Minute)
	if lm.timeout != 5*time.Minute {
		t.Errorf("Expected timeout 5m, got %v", lm.timeout)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
