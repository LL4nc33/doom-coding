package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg == nil {
		t.Fatal("NewDefaultConfig returned nil")
	}

	// Check default deployment mode
	if cfg.DeploymentMode != "tailscale" {
		t.Errorf("Expected DeploymentMode=tailscale, got %s", cfg.DeploymentMode)
	}

	// Check component defaults
	if !cfg.Components.Docker {
		t.Error("Expected Docker=true by default")
	}
	if !cfg.Components.Tailscale {
		t.Error("Expected Tailscale=true by default")
	}
	if !cfg.Components.TerminalTools {
		t.Error("Expected TerminalTools=true by default")
	}
	if !cfg.Components.SSHHardening {
		t.Error("Expected SSHHardening=true by default")
	}
	if !cfg.Components.SecretsManager {
		t.Error("Expected SecretsManager=true by default")
	}

	// Check environment defaults
	if cfg.Environment.PUID != "1000" {
		t.Errorf("Expected PUID=1000, got %s", cfg.Environment.PUID)
	}
	if cfg.Environment.PGID != "1000" {
		t.Errorf("Expected PGID=1000, got %s", cfg.Environment.PGID)
	}
	if cfg.Environment.Timezone != "Europe/Berlin" {
		t.Errorf("Expected Timezone=Europe/Berlin, got %s", cfg.Environment.Timezone)
	}
	if cfg.Environment.WorkspacePath != "./workspace" {
		t.Errorf("Expected WorkspacePath=./workspace, got %s", cfg.Environment.WorkspacePath)
	}

	// Check advanced defaults
	if cfg.Advanced.CodeServerPort != 8443 {
		t.Errorf("Expected CodeServerPort=8443, got %d", cfg.Advanced.CodeServerPort)
	}
	if cfg.Advanced.TargetArch != runtime.GOARCH {
		t.Errorf("Expected TargetArch=%s, got %s", runtime.GOARCH, cfg.Advanced.TargetArch)
	}
	if cfg.Advanced.TSAcceptDNS != false {
		t.Error("Expected TSAcceptDNS=false by default")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		wantErrs   int
		wantSubstr string
	}{
		{
			name: "valid config with docker and tailscale",
			config: &Config{
				DeploymentMode: "tailscale",
				Components: ComponentSelection{
					Docker:    true,
					Tailscale: true,
				},
				Credentials: Credentials{
					CodePassword:  "securepassword123",
					SudoPassword:  "sudopass123",
					TailscaleKey:  "tskey-auth-test",
				},
				Environment: Environment{
					WorkspacePath: "./workspace",
				},
			},
			wantErrs: 0,
		},
		{
			name: "missing code password",
			config: &Config{
				DeploymentMode: "tailscale",
				Components:     ComponentSelection{Docker: true},
				Credentials:    Credentials{SudoPassword: "sudopass"},
				Environment:    Environment{WorkspacePath: "./workspace"},
			},
			wantErrs:   1,
			wantSubstr: "code-server password",
		},
		{
			name: "missing sudo password",
			config: &Config{
				DeploymentMode: "local",
				Components:     ComponentSelection{Docker: true},
				Credentials:    Credentials{CodePassword: "codepass"},
				Environment:    Environment{WorkspacePath: "./workspace"},
			},
			wantErrs:   1,
			wantSubstr: "sudo password",
		},
		{
			name: "missing both passwords",
			config: &Config{
				DeploymentMode: "local",
				Components:     ComponentSelection{Docker: true},
				Credentials:    Credentials{},
				Environment:    Environment{WorkspacePath: "./workspace"},
			},
			wantErrs: 2,
		},
		{
			name: "tailscale mode without key",
			config: &Config{
				DeploymentMode: "tailscale",
				Components: ComponentSelection{
					Docker:    true,
					Tailscale: true,
				},
				Credentials: Credentials{
					CodePassword: "codepass",
					SudoPassword: "sudopass",
				},
				Environment: Environment{WorkspacePath: "./workspace"},
			},
			wantErrs:   1,
			wantSubstr: "Tailscale auth key",
		},
		{
			name: "empty workspace path",
			config: &Config{
				DeploymentMode: "local",
				Components:     ComponentSelection{Docker: false},
				Credentials:    Credentials{},
				Environment:    Environment{WorkspacePath: ""},
			},
			wantErrs:   1,
			wantSubstr: "workspace path",
		},
		{
			name: "local mode without tailscale - valid",
			config: &Config{
				DeploymentMode: "local",
				Components: ComponentSelection{
					Docker:    true,
					Tailscale: false,
				},
				Credentials: Credentials{
					CodePassword: "codepass",
					SudoPassword: "sudopass",
				},
				Environment: Environment{WorkspacePath: "./workspace"},
			},
			wantErrs: 0,
		},
		{
			name: "no docker - no password validation",
			config: &Config{
				DeploymentMode: "terminal-only",
				Components:     ComponentSelection{Docker: false},
				Credentials:    Credentials{},
				Environment:    Environment{WorkspacePath: "./workspace"},
			},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Validate()
			if len(errs) != tt.wantErrs {
				t.Errorf("Validate() returned %d errors, want %d: %v", len(errs), tt.wantErrs, errs)
			}
			if tt.wantSubstr != "" && len(errs) > 0 {
				found := false
				for _, err := range errs {
					if strings.Contains(err, tt.wantSubstr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, got %v", tt.wantSubstr, errs)
				}
			}
		})
	}
}

func TestConfigGenerateEnvFile(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Credentials.CodePassword = "testcodepass"
	cfg.Credentials.SudoPassword = "testsudopass"
	cfg.Credentials.TailscaleKey = "tskey-auth-test123"
	cfg.Environment.PUID = "1001"
	cfg.Environment.PGID = "1001"
	cfg.Environment.Timezone = "UTC"
	cfg.Environment.WorkspacePath = "/custom/workspace"
	cfg.Advanced.CodeServerPort = 9000
	cfg.Advanced.TSAcceptDNS = true
	cfg.Advanced.TSExtraArgs = "--advertise-tags=tag:test"

	env := cfg.GenerateEnvFile()

	// Check required entries
	expectedEntries := []string{
		"CODE_SERVER_PASSWORD=testcodepass",
		"SUDO_PASSWORD=testsudopass",
		"TS_AUTHKEY=tskey-auth-test123",
		"PUID=1001",
		"PGID=1001",
		"TZ=UTC",
		"WORKSPACE_PATH=/custom/workspace",
		"CODE_SERVER_PORT=9000",
		"TS_ACCEPT_DNS=true",
		"TS_EXTRA_ARGS=--advertise-tags=tag:test",
		"TARGETARCH=",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(env, entry) {
			t.Errorf("Generated env should contain %q", entry)
		}
	}

	// Check header comments
	if !strings.Contains(env, "# Doom Coding Environment Configuration") {
		t.Error("Generated env should contain header comment")
	}
	if !strings.Contains(env, "# Generated by doom-tui") {
		t.Error("Generated env should contain generator comment")
	}
}

func TestConfigGenerateEnvFileWithoutTailscale(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Credentials.CodePassword = "testpass"
	cfg.Credentials.SudoPassword = "sudopass"
	cfg.Credentials.TailscaleKey = "" // No tailscale key

	env := cfg.GenerateEnvFile()

	// Should have commented out TS_AUTHKEY
	if !strings.Contains(env, "# TS_AUTHKEY=") {
		t.Error("Generated env should contain commented TS_AUTHKEY when not set")
	}
	// Should not have actual TS_AUTHKEY
	if strings.Contains(env, "TS_AUTHKEY=tskey-") {
		t.Error("Generated env should not contain actual TS_AUTHKEY when not set")
	}
}

func TestConfigWriteEnvFile(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Credentials.CodePassword = "testpass"
	cfg.Credentials.SudoPassword = "sudopass"

	tmpDir := t.TempDir()

	err := cfg.WriteEnvFile(tmpDir)
	if err != nil {
		t.Fatalf("WriteEnvFile failed: %v", err)
	}

	envPath := filepath.Join(tmpDir, ".env")
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read .env: %v", err)
	}

	if !strings.Contains(string(content), "CODE_SERVER_PASSWORD=testpass") {
		t.Error(".env file should contain CODE_SERVER_PASSWORD")
	}
	if !strings.Contains(string(content), "SUDO_PASSWORD=sudopass") {
		t.Error(".env file should contain SUDO_PASSWORD")
	}

	// Check file permissions (should be 0600)
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Failed to stat .env: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected .env permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestConfigWriteSecretsFile(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Credentials.AnthropicKey = "sk-ant-api03-test123"

	tmpDir := t.TempDir()

	err := cfg.WriteSecretsFile(tmpDir)
	if err != nil {
		t.Fatalf("WriteSecretsFile failed: %v", err)
	}

	secretPath := filepath.Join(tmpDir, "secrets", "anthropic_api_key.txt")
	content, err := os.ReadFile(secretPath)
	if err != nil {
		t.Fatalf("Failed to read secret file: %v", err)
	}

	if string(content) != "sk-ant-api03-test123" {
		t.Errorf("Secret file should contain API key, got %s", string(content))
	}

	// Check directory permissions
	secretsDir := filepath.Join(tmpDir, "secrets")
	info, err := os.Stat(secretsDir)
	if err != nil {
		t.Fatalf("Failed to stat secrets dir: %v", err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("Expected secrets dir permissions 0700, got %o", info.Mode().Perm())
	}
}

func TestConfigWriteSecretsFileEmpty(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Credentials.AnthropicKey = "" // No API key

	tmpDir := t.TempDir()

	err := cfg.WriteSecretsFile(tmpDir)
	if err != nil {
		t.Fatalf("WriteSecretsFile failed: %v", err)
	}

	// Should not create secrets directory
	secretsDir := filepath.Join(tmpDir, "secrets")
	if _, err := os.Stat(secretsDir); !os.IsNotExist(err) {
		t.Error("Should not create secrets directory when no API key")
	}
}

func TestConfigSaveAndLoadFromFile(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.DeploymentMode = "local"
	cfg.Components.Docker = true
	cfg.Components.Tailscale = false
	cfg.Credentials.CodePassword = "testpass"
	cfg.Environment.PUID = "2000"

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Save
	err := cfg.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Load
	loaded, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Compare
	if loaded.DeploymentMode != cfg.DeploymentMode {
		t.Errorf("DeploymentMode mismatch: got %s, want %s", loaded.DeploymentMode, cfg.DeploymentMode)
	}
	if loaded.Components.Docker != cfg.Components.Docker {
		t.Errorf("Docker mismatch: got %v, want %v", loaded.Components.Docker, cfg.Components.Docker)
	}
	if loaded.Components.Tailscale != cfg.Components.Tailscale {
		t.Errorf("Tailscale mismatch: got %v, want %v", loaded.Components.Tailscale, cfg.Components.Tailscale)
	}
	if loaded.Credentials.CodePassword != cfg.Credentials.CodePassword {
		t.Errorf("CodePassword mismatch: got %s, want %s", loaded.Credentials.CodePassword, cfg.Credentials.CodePassword)
	}
	if loaded.Environment.PUID != cfg.Environment.PUID {
		t.Errorf("PUID mismatch: got %s, want %s", loaded.Environment.PUID, cfg.Environment.PUID)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadFromFile should fail for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("Expected 'failed to read config file' error, got: %v", err)
	}
}

func TestLoadFromFileInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(configPath, []byte("{invalid json}"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadFromFile(configPath)
	if err == nil {
		t.Error("LoadFromFile should fail for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse config file") {
		t.Errorf("Expected 'failed to parse config file' error, got: %v", err)
	}
}

func TestConfigGenerateBashFlags(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		wantFlags     []string
		wantNotFlags  []string
	}{
		{
			name: "full installation",
			config: &Config{
				DeploymentMode: "tailscale",
				Components: ComponentSelection{
					Docker:         true,
					Tailscale:      true,
					TerminalTools:  true,
					SSHHardening:   true,
					SecretsManager: true,
				},
				Credentials: Credentials{
					TailscaleKey:  "tskey-auth-test",
					CodePassword:  "codepass",
					AnthropicKey:  "sk-ant-test",
				},
			},
			wantFlags: []string{
				"--unattended",
				"--tailscale-key=tskey-auth-test",
				"--code-password=codepass",
				"--anthropic-key=sk-ant-test",
			},
			wantNotFlags: []string{
				"--skip-docker",
				"--skip-tailscale",
				"--skip-terminal",
				"--skip-hardening",
				"--skip-secrets",
			},
		},
		{
			name: "skip all components",
			config: &Config{
				DeploymentMode: "local",
				Components: ComponentSelection{
					Docker:         false,
					Tailscale:      false,
					TerminalTools:  false,
					SSHHardening:   false,
					SecretsManager: false,
				},
				Credentials: Credentials{},
			},
			wantFlags: []string{
				"--unattended",
				"--skip-docker",
				"--skip-tailscale",
				"--skip-terminal",
				"--skip-hardening",
				"--skip-secrets",
			},
		},
		{
			name: "local mode skips tailscale",
			config: &Config{
				DeploymentMode: "local",
				Components: ComponentSelection{
					Docker:    true,
					Tailscale: true, // Even if true, local mode skips it
				},
				Credentials: Credentials{},
			},
			wantFlags: []string{
				"--unattended",
				"--skip-tailscale",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := tt.config.GenerateBashFlags()

			// Check wanted flags are present
			for _, wantFlag := range tt.wantFlags {
				found := false
				for _, flag := range flags {
					if flag == wantFlag || strings.HasPrefix(flag, wantFlag) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected flag %q not found in %v", wantFlag, flags)
				}
			}

			// Check unwanted flags are absent
			for _, notWantFlag := range tt.wantNotFlags {
				for _, flag := range flags {
					if flag == notWantFlag {
						t.Errorf("Unexpected flag %q found in %v", notWantFlag, flags)
					}
				}
			}
		})
	}
}

func TestGetComposeFile(t *testing.T) {
	tests := []struct {
		deploymentMode string
		wantFile       string
	}{
		{"tailscale", "docker-compose.yml"},
		{"local", "docker-compose.lxc.yml"},
		{"terminal-only", ""},
	}

	for _, tt := range tests {
		t.Run(tt.deploymentMode, func(t *testing.T) {
			cfg := &Config{DeploymentMode: tt.deploymentMode}
			got := cfg.GetComposeFile()
			if got != tt.wantFile {
				t.Errorf("GetComposeFile() = %q, want %q", got, tt.wantFile)
			}
		})
	}
}
