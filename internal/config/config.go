package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

// Config represents the structure of the JSON configuration file.
// The struct tags map the Go fields to the corresponding JSON keys.
type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name,omitempty"`
}

// Read reads the JSON configuration file located in the user's HOME directory,
// unmarshals its content into a Config struct, and returns a pointer to it.
func Read() (*Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SetUser sets the current user name in the configuration, and writes the updated
// configuration back to the JSON file.
func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	return write(c)
}

// getConfigFilePath constructs the full path to the configuration file
// by reading the user's home directory and joining it with the config file name.
func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configFileName), nil
}

// write marshals the Config struct to JSON and writes it to the configuration file.
func write(cfg *Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
