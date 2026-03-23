package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds the application configuration from environment variables.
type Config struct {
	APIKey           string `env:"OPENAI_API_KEY,required"`
	Model            string `env:"AI_MODEL"   envDefault:"gpt-4o-mini"`
	Memory           bool   `env:"AI_MEMORY"  envDefault:"true"`
	DataDir          string `env:"AI_DATA_DIR" envDefault:".ai-chat"`
	ElevenLabsAPIKey string `env:"ELEVENLABS_API_KEY"`
	FalAPIKey        string `env:"FAL_API_KEY"`
}

// Load reads configuration from .env file (if present) and environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists (silently ignore if missing)
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("missing config: %w", err)
	}

	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return cfg, nil
}

// DBPath returns the path to the SQLite database.
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, "chat.db")
}
