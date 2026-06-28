package httpapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewAIClientFromConfigFile(t *testing.T) {
	configPath := writeAIConfig(t, `{
		"apiKey": "sk-or-v1-config-key-123456",
		"baseURL": "https://openrouter.ai/api/v1",
		"model": "openai/gpt-4o-mini"
	}`)
	t.Setenv("AI_CONFIG_FILE", configPath)
	t.Setenv("AI_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("AI_API_BASE_URL", "")
	t.Setenv("AI_CHAT_COMPLETIONS_URL", "")
	t.Setenv("AI_MODEL", "")

	client := NewAIClientFromEnv()
	status := client.Status()

	if !status.Enabled {
		t.Fatal("期望配置文件中的 apiKey 可以启用 AI")
	}
	if !status.ConfigLoaded {
		t.Fatal("期望标记配置文件已加载")
	}
	if status.KeySource != configPath+":apiKey" {
		t.Fatalf("KeySource = %q, want %q", status.KeySource, configPath+":apiKey")
	}
	if status.BaseURL != "https://openrouter.ai/api/v1" {
		t.Fatalf("BaseURL = %q", status.BaseURL)
	}
	if status.Model != "openai/gpt-4o-mini" {
		t.Fatalf("Model = %q", status.Model)
	}
	if status.TimeoutSeconds != 180 {
		t.Fatalf("TimeoutSeconds = %d, want 180", status.TimeoutSeconds)
	}
	if strings.Contains(status.KeyPreview, "config-key") {
		t.Fatalf("KeyPreview 泄露了完整 Key: %q", status.KeyPreview)
	}
}

func TestAIEnvOverridesConfigFile(t *testing.T) {
	configPath := writeAIConfig(t, `{
		"apiKey": "sk-or-v1-config-key-123456",
		"baseURL": "https://openrouter.ai/api/v1",
		"model": "openai/gpt-4o-mini",
		"timeoutSeconds": 60
	}`)
	t.Setenv("AI_CONFIG_FILE", configPath)
	t.Setenv("AI_API_KEY", "sk-env-key-123456")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("AI_API_BASE_URL", "https://api.openai.com/v1")
	t.Setenv("AI_CHAT_COMPLETIONS_URL", "")
	t.Setenv("AI_MODEL", "gpt-4o-mini")
	t.Setenv("AI_TIMEOUT_SECONDS", "240")

	client := NewAIClientFromEnv()
	status := client.Status()

	if status.KeySource != "AI_API_KEY" {
		t.Fatalf("KeySource = %q, want AI_API_KEY", status.KeySource)
	}
	if status.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("BaseURL = %q", status.BaseURL)
	}
	if status.Model != "gpt-4o-mini" {
		t.Fatalf("Model = %q", status.Model)
	}
	if status.TimeoutSeconds != 240 {
		t.Fatalf("TimeoutSeconds = %d, want 240", status.TimeoutSeconds)
	}
}

func writeAIConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ai.local.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	return path
}
