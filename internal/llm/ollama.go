package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// OllamaConfig Ollama 配置
type OllamaConfig struct {
	BaseURL string        // Ollama 服务地址，默认 http://localhost:11434
	Model   string        // 模型名称，如 "llama2", "qwen2.5:3b" 等
	Timeout time.Duration // 请求超时时间
}

// NewOllamaConfigFromEnv 从环境变量创建配置
func NewOllamaConfigFromEnv() *OllamaConfig {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "qwen2.5:3b-instruct" // 默认模型
	}

	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("OLLAMA_TIMEOUT"); timeoutStr != "" {
		if parsed, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsed
		}
	}

	return &OllamaConfig{
		BaseURL: baseURL,
		Model:   model,
		Timeout: timeout,
	}
}

// Ollama LLM 实现
type Ollama struct {
	config  *OllamaConfig
	client  *http.Client
	baseURL string
	model   string
}

// NewOllama 创建 Ollama LLM
func NewOllama(config *OllamaConfig) (*Ollama, error) {
	if config == nil {
		config = NewOllamaConfigFromEnv()
	}

	if config.Model == "" {
		return nil, fmt.Errorf("Ollama model is required")
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &Ollama{
		config:  config,
		client:  client,
		baseURL: config.BaseURL,
		model:   config.Model,
	}, nil
}

// ollamaChatRequest Ollama API 请求结构
type ollamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ollamaMessage Ollama 消息结构
type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatResponse Ollama API 响应结构
type ollamaChatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// Generate 生成回复
func (o *Ollama) Generate(ctx context.Context, messages []Message) (string, error) {
	if o == nil {
		return "", fmt.Errorf("Ollama client is nil")
	}
	if o.config == nil {
		return "", fmt.Errorf("Ollama config is nil")
	}
	if o.client == nil {
		return "", fmt.Errorf("Ollama HTTP client is not initialized")
	}

	// 转换消息格式
	ollamaMessages := make([]ollamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 构建请求
	reqBody := ollamaChatRequest{
		Model:    o.model,
		Messages: ollamaMessages,
		Stream:   false,
		Options: map[string]interface{}{
			"temperature": 0.7,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var chatResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != "" {
		return "", fmt.Errorf("Ollama API error: %s", chatResp.Error)
	}

	return chatResp.Message.Content, nil
}

// GenerateStream 流式生成回复
func (o *Ollama) GenerateStream(ctx context.Context, messages []Message, callback func(string) error) error {
	if o == nil {
		return fmt.Errorf("Ollama client is nil")
	}
	if o.config == nil {
		return fmt.Errorf("Ollama config is nil")
	}
	if o.client == nil {
		return fmt.Errorf("Ollama HTTP client is not initialized")
	}

	// 转换消息格式
	ollamaMessages := make([]ollamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 构建请求（流式）
	reqBody := ollamaChatRequest{
		Model:    o.model,
		Messages: ollamaMessages,
		Stream:   true,
		Options: map[string]interface{}{
			"temperature": 0.7,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 流式读取响应
	decoder := json.NewDecoder(resp.Body)
	for {
		var chatResp ollamaChatResponse
		if err := decoder.Decode(&chatResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode stream response: %w", err)
		}

		if chatResp.Error != "" {
			return fmt.Errorf("Ollama API error: %s", chatResp.Error)
		}

		// 调用回调函数处理每个片段
		if chatResp.Message.Content != "" {
			if err := callback(chatResp.Message.Content); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}

		// 如果完成，退出循环
		if chatResp.Done {
			break
		}
	}

	return nil
}
