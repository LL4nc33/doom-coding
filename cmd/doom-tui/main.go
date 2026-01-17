package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Version information
const (
	Version = "0.0.6a"
	AppName = "doom-tui"
)

// CLI flags
var (
	unattended     bool
	tailscaleKey   string
	codePassword   string
	anthropicKey   string
	sudoPassword   string
	configFile     string
	dryRun         bool
	showCommands   bool
	skipDocker     bool
	skipTailscale  bool
	skipTerminal   bool
	skipHardening  bool
	skipSecrets    bool
	verbose        bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     AppName,
		Short:   "Doom Coding - Interactive TUI Setup",
		Long:    `An interactive TUI for setting up the Doom Coding development environment.`,
		Version: Version,
		RunE:    runTUI,
	}

	// CLI flags for automation
	rootCmd.Flags().BoolVar(&unattended, "unattended", false, "Run in unattended mode (no prompts)")
	rootCmd.Flags().StringVar(&tailscaleKey, "tailscale-key", "", "Tailscale auth key")
	rootCmd.Flags().StringVar(&codePassword, "code-password", "", "code-server password")
	rootCmd.Flags().StringVar(&anthropicKey, "anthropic-key", "", "Anthropic API key")
	rootCmd.Flags().StringVar(&sudoPassword, "sudo-password", "", "Container sudo password")
	rootCmd.Flags().StringVar(&configFile, "config", "", "Load configuration from JSON file")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
	rootCmd.Flags().BoolVar(&showCommands, "show-commands", false, "Show equivalent bash commands")
	rootCmd.Flags().BoolVar(&skipDocker, "skip-docker", false, "Skip Docker installation")
	rootCmd.Flags().BoolVar(&skipTailscale, "skip-tailscale", false, "Skip Tailscale (use local network)")
	rootCmd.Flags().BoolVar(&skipTerminal, "skip-terminal", false, "Skip terminal tools setup")
	rootCmd.Flags().BoolVar(&skipHardening, "skip-hardening", false, "Skip SSH hardening")
	rootCmd.Flags().BoolVar(&skipSecrets, "skip-secrets", false, "Skip secrets management setup")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	// CLI-only mode subcommand
	cliCmd := &cobra.Command{
		Use:   "cli",
		Short: "Run in CLI mode (no TUI)",
		RunE:  runCLI,
	}
	rootCmd.AddCommand(cliCmd)

	// Status subcommand
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check installation status",
		RunE:  runStatus,
	}
	rootCmd.AddCommand(statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	// If unattended mode, run CLI instead
	if unattended {
		return runCLI(cmd, args)
	}

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %w", err)
	}

	// Initialize the TUI model
	model := NewModel(projectRoot)

	// Run the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func runCLI(cmd *cobra.Command, args []string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %w", err)
	}

	// Build install.sh command with flags
	installArgs := []string{filepath.Join(projectRoot, "scripts", "install.sh")}

	if unattended {
		installArgs = append(installArgs, "--unattended")
	}
	if tailscaleKey != "" {
		installArgs = append(installArgs, fmt.Sprintf("--tailscale-key=%s", tailscaleKey))
	}
	if codePassword != "" {
		installArgs = append(installArgs, fmt.Sprintf("--code-password=%s", codePassword))
	}
	if anthropicKey != "" {
		installArgs = append(installArgs, fmt.Sprintf("--anthropic-key=%s", anthropicKey))
	}
	if skipDocker {
		installArgs = append(installArgs, "--skip-docker")
	}
	if skipTailscale {
		installArgs = append(installArgs, "--skip-tailscale")
	}
	if skipTerminal {
		installArgs = append(installArgs, "--skip-terminal")
	}
	if skipHardening {
		installArgs = append(installArgs, "--skip-hardening")
	}
	if skipSecrets {
		installArgs = append(installArgs, "--skip-secrets")
	}
	if dryRun {
		installArgs = append(installArgs, "--dry-run")
	}
	if verbose {
		installArgs = append(installArgs, "--verbose")
	}

	if showCommands {
		fmt.Println("Equivalent bash command:")
		fmt.Println(strings.Join(installArgs, " "))
		return nil
	}

	// Execute install script
	execCmd := exec.Command("bash", installArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	return execCmd.Run()
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %w", err)
	}

	healthCheck := filepath.Join(projectRoot, "scripts", "health-check.sh")
	execCmd := exec.Command("bash", healthCheck)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func findProjectRoot() (string, error) {
	// Check if we're in the doom-coding directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for docker-compose.yml as marker
	markers := []string{"docker-compose.yml", "scripts/install.sh"}

	dir := cwd
	for {
		found := true
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(dir, marker)); os.IsNotExist(err) {
				found = false
				break
			}
		}
		if found {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: check common locations
	homeDir, _ := os.UserHomeDir()
	commonPaths := []string{
		"/config/repos/doom-coding",
		filepath.Join(homeDir, "doom-coding"),
		"/opt/doom-coding",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(filepath.Join(path, "docker-compose.yml")); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("doom-coding project not found")
}
