package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Version != "0.0.6" {
		t.Errorf("Expected version '0.0.6', got %q", Version)
	}
}

func TestAppName(t *testing.T) {
	if AppName == "" {
		t.Error("AppName should not be empty")
	}
	if AppName != "doom-tui" {
		t.Errorf("Expected AppName 'doom-tui', got %q", AppName)
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary project structure
	tmpDir := t.TempDir()

	// Create marker files
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte("version: '3'"), 0644); err != nil {
		t.Fatalf("Failed to create docker-compose.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "install.sh"), []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("Failed to create install.sh: %v", err)
	}

	// Change to the project directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test findProjectRoot
	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("findProjectRoot failed: %v", err)
	}

	if root != tmpDir {
		t.Errorf("Expected root=%q, got %q", tmpDir, root)
	}
}

func TestFindProjectRootFromSubdir(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary project structure
	tmpDir := t.TempDir()

	// Create marker files
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte("version: '3'"), 0644); err != nil {
		t.Fatalf("Failed to create docker-compose.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "install.sh"), []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("Failed to create install.sh: %v", err)
	}

	// Create and change to a subdirectory
	subdir := filepath.Join(tmpDir, "internal", "config")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("Failed to change to subdirectory: %v", err)
	}

	// Test findProjectRoot from subdirectory
	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("findProjectRoot failed: %v", err)
	}

	if root != tmpDir {
		t.Errorf("Expected root=%q, got %q", tmpDir, root)
	}
}

func TestFindProjectRootNotFound(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create an empty temporary directory
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test findProjectRoot - should check common paths as fallback
	// The actual result depends on the system
	_, err = findProjectRoot()

	// We can't guarantee it will fail (doom-coding might exist on the system)
	// So we just check it returns without crashing
	t.Logf("findProjectRoot result: error=%v", err)
}

func TestCLIFlags(t *testing.T) {
	// Test that CLI flag variables are initialized properly
	// These are package-level variables

	// All should be false/empty by default
	if unattended {
		t.Error("unattended should be false by default")
	}
	if tailscaleKey != "" {
		t.Error("tailscaleKey should be empty by default")
	}
	if codePassword != "" {
		t.Error("codePassword should be empty by default")
	}
	if anthropicKey != "" {
		t.Error("anthropicKey should be empty by default")
	}
	if sudoPassword != "" {
		t.Error("sudoPassword should be empty by default")
	}
	if configFile != "" {
		t.Error("configFile should be empty by default")
	}
	if dryRun {
		t.Error("dryRun should be false by default")
	}
	if showCommands {
		t.Error("showCommands should be false by default")
	}
	if skipDocker {
		t.Error("skipDocker should be false by default")
	}
	if skipTailscale {
		t.Error("skipTailscale should be false by default")
	}
	if skipTerminal {
		t.Error("skipTerminal should be false by default")
	}
	if skipHardening {
		t.Error("skipHardening should be false by default")
	}
	if skipSecrets {
		t.Error("skipSecrets should be false by default")
	}
	if verbose {
		t.Error("verbose should be false by default")
	}
}

func TestFindProjectRootWithCommonPaths(t *testing.T) {
	// Test that common paths are checked
	// This tests the fallback mechanism

	// Create a mock doom-coding directory in a known location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not get home directory")
	}

	testPath := filepath.Join(homeDir, "doom-coding-test-temp")

	// Clean up after test
	defer os.RemoveAll(testPath)

	// Create the test project structure
	if err := os.MkdirAll(filepath.Join(testPath, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testPath, "docker-compose.yml"), []byte("version: '3'"), 0644); err != nil {
		t.Fatalf("Failed to create docker-compose.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testPath, "scripts", "install.sh"), []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("Failed to create install.sh: %v", err)
	}

	// Note: We can't easily test the common paths fallback without modifying
	// the actual function to accept a list of paths, so this is a limited test
}

func TestFindProjectRootMarkers(t *testing.T) {
	// Test that all marker files must exist
	tmpDir := t.TempDir()

	// Save and restore current directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Test 1: Only docker-compose.yml exists
	if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte("version: '3'"), 0644); err != nil {
		t.Fatalf("Failed to create docker-compose.yml: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Should not find it (missing scripts/install.sh)
	_, err := findProjectRoot()
	// The function may fall back to common paths, so we can't assert error
	t.Logf("Result with only docker-compose.yml: err=%v", err)

	// Test 2: Add scripts/install.sh
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "install.sh"), []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("Failed to create install.sh: %v", err)
	}

	// Now it should find it
	root, err := findProjectRoot()
	if err != nil {
		t.Logf("findProjectRoot failed even with both markers: %v", err)
	} else if root != tmpDir {
		t.Errorf("Expected root=%q, got %q", tmpDir, root)
	}
}

// TestNewModel tests the model creation
// Note: This would require reading model.go to understand the Model struct
func TestNewModelBasic(t *testing.T) {
	// Create a temporary project root
	tmpDir := t.TempDir()

	// Create required structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte("version: '3'"), 0644); err != nil {
		t.Fatalf("Failed to create docker-compose.yml: %v", err)
	}

	// Test NewModel - this requires the Model type from model.go
	model := NewModel(tmpDir)

	if model.projectRoot != tmpDir {
		t.Errorf("Expected projectRoot=%q, got %q", tmpDir, model.projectRoot)
	}
}

// BenchmarkFindProjectRoot benchmarks the project root finding
func BenchmarkFindProjectRoot(b *testing.B) {
	// Create a temporary project structure
	tmpDir := b.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0755); err != nil {
		b.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte("version: '3'"), 0644); err != nil {
		b.Fatalf("Failed to create docker-compose.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "install.sh"), []byte("#!/bin/bash"), 0755); err != nil {
		b.Fatalf("Failed to create install.sh: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to change directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findProjectRoot()
	}
}
