// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the Ark configuration
type Config struct {
	// Current active profile
	CurrentProfile string `yaml:"current_profile"`

	// Agent configuration
	Agent AgentConfig `yaml:"agent"`

	// Backend configuration
	Backend BackendConfig `yaml:"backend"`

	// User profiles for different AWS accounts/roles
	Profiles map[string]Profile `yaml:"profiles"`

	// Training preferences
	Training TrainingConfig `yaml:"training"`
}

// AgentConfig holds agent-specific settings
type AgentConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// BackendConfig holds backend-specific settings
type BackendConfig struct {
	URL string `yaml:"url"`
}

// Profile represents an AWS profile configuration
type Profile struct {
	Name        string `yaml:"name"`
	Region      string `yaml:"region"`
	AWSProfile  string `yaml:"aws_profile"` // AWS CLI profile name
	Description string `yaml:"description"`
}

// TrainingConfig holds training preferences
type TrainingConfig struct {
	Enabled      bool     `yaml:"enabled"`
	SkipModules  []string `yaml:"skip_modules"`
	AutoComplete bool     `yaml:"auto_complete"`
}

// DefaultConfig returns a new config with default values
func DefaultConfig() *Config {
	return &Config{
		CurrentProfile: "default",
		Agent: AgentConfig{
			Host: "127.0.0.1",
			Port: 8737,
		},
		Backend: BackendConfig{
			URL: "http://localhost:8080",
		},
		Profiles: map[string]Profile{
			"default": {
				Name:        "default",
				Region:      "us-east-1",
				AWSProfile:  "default",
				Description: "Default AWS profile",
			},
		},
		Training: TrainingConfig{
			Enabled:      true,
			SkipModules:  []string{},
			AutoComplete: false,
		},
	}
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// Save saves configuration to a file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() (string, error) {
	// Check environment variable
	if path := os.Getenv("ARK_CONFIG"); path != "" {
		return path, nil
	}

	// Use ~/.ark/config.yml
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	return filepath.Join(home, ".ark", "config.yml"), nil
}

// Set sets a configuration value by key path (e.g., "agent.port")
func (c *Config) Set(key, value string) error {
	// Simple implementation for common keys
	switch key {
	case "current_profile":
		c.CurrentProfile = value
	case "agent.host":
		c.Agent.Host = value
	case "agent.port":
		var port int
		if _, err := fmt.Sscanf(value, "%d", &port); err != nil {
			return fmt.Errorf("invalid port number: %s", value)
		}
		c.Agent.Port = port
	case "backend.url":
		c.Backend.URL = value
	case "training.enabled":
		c.Training.Enabled = value == "true"
	case "training.auto_complete":
		c.Training.AutoComplete = value == "true"
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// Get gets a configuration value by key path
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "current_profile":
		return c.CurrentProfile, nil
	case "agent.host":
		return c.Agent.Host, nil
	case "agent.port":
		return fmt.Sprintf("%d", c.Agent.Port), nil
	case "backend.url":
		return c.Backend.URL, nil
	case "training.enabled":
		return fmt.Sprintf("%t", c.Training.Enabled), nil
	case "training.auto_complete":
		return fmt.Sprintf("%t", c.Training.AutoComplete), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}
