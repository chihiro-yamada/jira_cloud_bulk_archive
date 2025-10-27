package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	JiraBaseURL    string
	JiraEmail      string
	JiraAPIToken   string
	JiraProjectKey string
	ArchiveLabel   string
	MaxWorkers     int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		JiraBaseURL:    os.Getenv("JIRA_BASE_URL"),
		JiraEmail:      os.Getenv("JIRA_EMAIL"),
		JiraAPIToken:   os.Getenv("JIRA_API_TOKEN"),
		JiraProjectKey: os.Getenv("JIRA_PROJECT_KEY"),
		ArchiveLabel:   getEnvOrDefault("ARCHIVE_LABEL", "archive"),
		MaxWorkers:     getIntEnvOrDefault("MAX_WORKERS", 5),
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if all required configuration values are present
func (c *Config) Validate() error {
	if c.JiraBaseURL == "" {
		return fmt.Errorf("JIRA_BASE_URL is required")
	}
	if c.JiraEmail == "" {
		return fmt.Errorf("JIRA_EMAIL is required")
	}
	if c.JiraAPIToken == "" {
		return fmt.Errorf("JIRA_API_TOKEN is required")
	}
	if c.JiraProjectKey == "" {
		return fmt.Errorf("JIRA_PROJECT_KEY is required")
	}
	if c.MaxWorkers < 1 {
		return fmt.Errorf("MAX_WORKERS must be at least 1")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
