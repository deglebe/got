package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitHubToken string `yaml:"github_token"`
}

// load configuration from the config file
func loadConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "got", "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil // return empty config if file doesn't exist
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func saveConfig(config *Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "got", "config.yaml")

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getGitHubToken() (string, error) {
	config, err := loadConfig()
	if err != nil {
		return "", err
	}

	if config.GitHubToken != "" && config.GitHubToken != "your_github_token_here" {
		return config.GitHubToken, nil
	}

	// token not set, prompt user
	token, err := showGitHubTokenForm()
	if err != nil {
		return "", err
	}

	if err := validateGitHubToken(token); err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	config.GitHubToken = token
	if err := saveConfig(config); err != nil {
		fmt.Printf("warning: could not save token to config: %v\n", err)
	}

	return token, nil
}
