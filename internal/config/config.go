package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	LLM       LLMConfig       `yaml:"llm"`
	VideoFeed VideoFeedConfig `yaml:"video_feed"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type LLMConfig struct {
	BaseURL   string `yaml:"base_url"`
	Model     string `yaml:"model"`
	APIKeyEnv string `yaml:"api_key_env"`
	APIKey    string `yaml:"-"`
}

type VideoFeedConfig struct {
	BaseURL string `yaml:"base_url"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Address: "127.0.0.1:8090",
		},
		Database: DatabaseConfig{
			DSN: "data/video-ops-agent.db",
		},
		LLM: LLMConfig{
			BaseURL:   "https://api.openai.com/v1",
			Model:     "gpt-4o-mini",
			APIKeyEnv: "OPENAI_API_KEY",
		},
		VideoFeed: VideoFeedConfig{
			BaseURL: "http://127.0.0.1:8080",
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	if path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read config file %q: %w", path, err)
		}
		if err := yaml.Unmarshal(content, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
		}
	}

	if cfg.LLM.APIKeyEnv != "" {
		cfg.LLM.APIKey = os.Getenv(cfg.LLM.APIKeyEnv)
	}

	return cfg, nil
}
