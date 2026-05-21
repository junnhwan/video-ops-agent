package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesDefaultsWhenPathIsEmpty(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Address != "127.0.0.1:8090" {
		t.Fatalf("server address = %q, want %q", cfg.Server.Address, "127.0.0.1:8090")
	}
	if cfg.Database.DSN != "data/video-ops-agent.db" {
		t.Fatalf("database dsn = %q, want %q", cfg.Database.DSN, "data/video-ops-agent.db")
	}
	if cfg.LLM.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("llm base url = %q, want %q", cfg.LLM.BaseURL, "https://api.openai.com/v1")
	}
	if cfg.LLM.Model != "gpt-4o-mini" {
		t.Fatalf("llm model = %q, want %q", cfg.LLM.Model, "gpt-4o-mini")
	}
	if cfg.LLM.APIKeyEnv != "OPENAI_API_KEY" {
		t.Fatalf("llm api key env = %q, want %q", cfg.LLM.APIKeyEnv, "OPENAI_API_KEY")
	}
	if cfg.LLM.APIKey != "" {
		t.Fatalf("llm api key should be empty when environment variable is empty")
	}
	if cfg.VideoFeed.BaseURL != "http://127.0.0.1:8080" {
		t.Fatalf("video-feed base url = %q, want %q", cfg.VideoFeed.BaseURL, "http://127.0.0.1:8080")
	}
}

func TestLoadMergesYAMLAndResolvesAPIKeyFromEnvironment(t *testing.T) {
	const apiKeyEnv = "VIDEO_OPS_AGENT_TEST_API_KEY"
	t.Setenv(apiKeyEnv, "test-secret")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  address: "127.0.0.1:9099"
database:
  dsn: "file:test.db?cache=shared"
llm:
  base_url: "http://llm.local/v1"
  model: "ops-test-model"
  api_key_env: "VIDEO_OPS_AGENT_TEST_API_KEY"
video_feed:
  base_url: "http://video-feed.local"
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Address != "127.0.0.1:9099" {
		t.Fatalf("server address = %q, want %q", cfg.Server.Address, "127.0.0.1:9099")
	}
	if cfg.Database.DSN != "file:test.db?cache=shared" {
		t.Fatalf("database dsn = %q, want %q", cfg.Database.DSN, "file:test.db?cache=shared")
	}
	if cfg.LLM.BaseURL != "http://llm.local/v1" {
		t.Fatalf("llm base url = %q, want %q", cfg.LLM.BaseURL, "http://llm.local/v1")
	}
	if cfg.LLM.Model != "ops-test-model" {
		t.Fatalf("llm model = %q, want %q", cfg.LLM.Model, "ops-test-model")
	}
	if cfg.LLM.APIKeyEnv != apiKeyEnv {
		t.Fatalf("llm api key env = %q, want %q", cfg.LLM.APIKeyEnv, apiKeyEnv)
	}
	if cfg.LLM.APIKey != "test-secret" {
		t.Fatalf("llm api key = %q, want environment secret", cfg.LLM.APIKey)
	}
	if cfg.VideoFeed.BaseURL != "http://video-feed.local" {
		t.Fatalf("video-feed base url = %q, want %q", cfg.VideoFeed.BaseURL, "http://video-feed.local")
	}
}
