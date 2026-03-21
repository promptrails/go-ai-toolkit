package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

// Config holds the application configuration from environment variables.
type Config struct {
	APIKey  string `env:"OPENAI_API_KEY,required"`
	Model   string `env:"AI_MODEL"   envDefault:"gpt-4o-mini"`
	Memory  bool   `env:"AI_MEMORY"  envDefault:"true"`
	DataDir string `env:"AI_DATA_DIR"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("missing config: %w", err)
	}

	if cfg.DataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.DataDir = filepath.Join(home, ".ai-chat")
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return cfg, nil
}

// DBPath returns the path to the SQLite database.
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, "chat.db")
}
