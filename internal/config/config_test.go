package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eefret/gator/internal/config"
)

// overrideHome sets HOME to a temporary directory for testing and
// restores the original value when the test is done.
func overrideHome(t *testing.T) string {
	originalHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("Failed to set HOME env variable: %v", err)
	}
	t.Cleanup(func() {
		// Restore the original HOME value
		os.Setenv("HOME", originalHome)
	})
	return tempHome
}

// TestReadValidFile creates a valid config file and ensures that Read()
// returns a Config struct with the expected values.
func TestReadValidFile(t *testing.T) {
	tempHome := overrideHome(t)

	// Prepare a valid config file at $HOME/.gatorconfig.json.
	configFilePath := filepath.Join(tempHome, ".gatorconfig.json")
	validContent := `{"db_url": "postgres://example"}`
	if err := os.WriteFile(configFilePath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write valid config file: %v", err)
	}

	// Read the configuration.
	cfg, err := config.Read()
	if err != nil {
		t.Fatalf("Expected Read to succeed, got error: %v", err)
	}
	if cfg.DbURL != "postgres://example" {
		t.Errorf("Expected DbURL to be 'postgres://example', got %q", cfg.DbURL)
	}
	if cfg.CurrentUserName != "" {
		t.Errorf("Expected CurrentUserName to be empty, got %q", cfg.CurrentUserName)
	}
}

// TestReadInvalidJSON writes a config file with invalid JSON and verifies
// that Read() returns an error.
func TestReadInvalidJSON(t *testing.T) {
	tempHome := overrideHome(t)

	configFilePath := filepath.Join(tempHome, ".gatorconfig.json")
	invalidContent := `{"db_url": "postgres://example",}` // Invalid JSON (trailing comma)
	if err := os.WriteFile(configFilePath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	_, err := config.Read()
	if err == nil {
		t.Fatal("Expected error when reading invalid JSON, got nil")
	}
}

// TestSetUser verifies that calling SetUser updates the configuration file
// with the new user value.
func TestSetUser(t *testing.T) {
	tempHome := overrideHome(t)

	// Create a valid config file.
	configFilePath := filepath.Join(tempHome, ".gatorconfig.json")
	validContent := `{"db_url": "postgres://example"}`
	if err := os.WriteFile(configFilePath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write valid config file: %v", err)
	}

	// Read the config.
	cfg, err := config.Read()
	if err != nil {
		t.Fatalf("Expected Read to succeed, got error: %v", err)
	}

	// Update the user.
	if err := cfg.SetUser("testuser"); err != nil {
		t.Fatalf("Expected SetUser to succeed, got error: %v", err)
	}

	// Read the config again to verify the update.
	updatedCfg, err := config.Read()
	if err != nil {
		t.Fatalf("Expected Read after SetUser to succeed, got error: %v", err)
	}
	if updatedCfg.CurrentUserName != "testuser" {
		t.Errorf("Expected CurrentUserName to be 'testuser', got %q", updatedCfg.CurrentUserName)
	}
}

// TestReadNoFile checks that Read() returns an error when the configuration
// file does not exist.
func TestReadNoFile(t *testing.T) {
	overrideHome(t) // Set HOME to a temporary directory; no config file is created.

	_, err := config.Read()
	if err == nil {
		t.Fatal("Expected error when config file does not exist, got nil")
	}
}
