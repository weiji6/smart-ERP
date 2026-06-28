package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultAIConfigPath = "config/ai.local.json"

type AIClient struct {
	apiKey             string
	baseURL            string
	chatCompletionsURL string
	model              string
	client             *http.Client
	keyName            string
	configPath         string
	configLoaded       bool
}

type aiFileConfig struct {
	APIKey             string `json:"apiKey"`
	BaseURL            string `json:"baseURL"`
	ChatCompletionsURL string `json:"chatCompletionsURL"`
	Model              string `json:"model"`
}

type aiChatRequest struct {
	Model       string          `json:"model"`
	Messages    []aiChatMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
}

type aiChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiChatResponse struct {
	Choices []struct {
		Message aiChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewAIClientFromEnv() AIClient {
	configPath := os.Getenv("AI_CONFIG_FILE")
	if configPath == "" {
		configPath = defaultAIConfigPath
	}

	fileConfig, configLoaded := loadAIFileConfig(configPath)

	apiKey := strings.TrimSpace(fileConfig.APIKey)
	keyName := ""
	if apiKey != "" {
		keyName = configPath + ":apiKey"
	}
	baseURL := strings.TrimSpace(fileConfig.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	chatCompletionsURL := strings.TrimSpace(fileConfig.ChatCompletionsURL)
	model := strings.TrimSpace(fileConfig.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}

	if v := strings.TrimSpace(os.Getenv("AI_API_KEY")); v != "" {
		apiKey = v
		keyName = "AI_API_KEY"
	} else if v := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); v != "" {
		apiKey = v
		keyName = "OPENAI_API_KEY"
	}
	if v := strings.TrimSpace(os.Getenv("AI_API_BASE_URL")); v != "" {
		baseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("AI_CHAT_COMPLETIONS_URL")); v != "" {
		chatCompletionsURL = v
	}
	if v := strings.TrimSpace(os.Getenv("AI_MODEL")); v != "" {
		model = v
	}

	return AIClient{
		apiKey:             apiKey,
		baseURL:            strings.TrimRight(baseURL, "/"),
		chatCompletionsURL: chatCompletionsURL,
		model:              model,
		client:             &http.Client{Timeout: 25 * time.Second},
		keyName:            keyName,
		configPath:         configPath,
		configLoaded:       configLoaded,
	}
}

func loadAIFileConfig(path string) (aiFileConfig, bool) {
	if strings.TrimSpace(path) == "" {
		return aiFileConfig{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return aiFileConfig{}, false
	}
	var cfg aiFileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return aiFileConfig{}, false
	}
	return cfg, true
}

func (c AIClient) Enabled() bool {
	return c.apiKey != ""
}

type AIStatus struct {
	Enabled            bool   `json:"enabled"`
	KeySource          string `json:"keySource,omitempty"`
	KeyPreview         string `json:"keyPreview,omitempty"`
	ConfigPath         string `json:"configPath"`
	ConfigLoaded       bool   `json:"configLoaded"`
	BaseURL            string `json:"baseURL"`
	ChatCompletionsURL string `json:"chatCompletionsURL"`
	Model              string `json:"model"`
	Warning            string `json:"warning,omitempty"`
}

func (c AIClient) Status() AIStatus {
	endpoint := c.chatCompletionsEndpoint()
	status := AIStatus{
		Enabled:            c.Enabled(),
		ConfigPath:         c.configPath,
		ConfigLoaded:       c.configLoaded,
		BaseURL:            c.baseURL,
		ChatCompletionsURL: endpoint,
		Model:              c.model,
	}
	if c.Enabled() {
		status.KeySource = c.keyName
		status.KeyPreview = maskKey(c.apiKey)
	}
	if strings.HasPrefix(c.apiKey, "sk-or-") && strings.Contains(c.baseURL, "api.openai.com") {
		status.Warning = "当前 Key 看起来像 OpenRouter Key，但 Base URL 是 OpenAI 官方地址，可能会调用失败；OpenRouter 通常应使用 https://openrouter.ai/api/v1。"
	}
	return status
}

func (c AIClient) chatCompletionsEndpoint() string {
	if strings.TrimSpace(c.chatCompletionsURL) != "" {
		return strings.TrimSpace(c.chatCompletionsURL)
	}
	return c.baseURL + "/chat/completions"
}

func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:6] + "..." + key[len(key)-4:]
}

func (c AIClient) Ask(ctx context.Context, prompt string) (string, error) {
	if !c.Enabled() {
		return "", errors.New("未配置 AI_API_KEY 或 OPENAI_API_KEY")
	}
	if c.model == "" {
		return "", errors.New("未配置 AI_MODEL")
	}

	body := aiChatRequest{
		Model: c.model,
		Messages: []aiChatMessage{
			{Role: "system", Content: "你是一名严谨的ERP沙盘经营顾问，必须先给结论，再给风险和修改方案。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("序列化 AI 请求失败: %w", err)
	}

	endpoint := c.chatCompletionsEndpoint()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("创建 AI 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用 AI API 失败: %w", err)
	}
	defer resp.Body.Close()

	var result aiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析 AI 响应失败: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if result.Error != nil && result.Error.Message != "" {
			return "", fmt.Errorf("AI API 返回错误: %s", result.Error.Message)
		}
		return "", fmt.Errorf("AI API 返回状态码 %d", resp.StatusCode)
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", errors.New("AI API 未返回有效建议")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
