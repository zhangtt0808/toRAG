package embedding

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

// OllamaEmbedderConfig Ollama Embedding 配置
type OllamaEmbedderConfig struct {
	BaseURL   string        // Ollama 服务地址，默认 http://localhost:11434
	Model     string        // 嵌入模型名称，如 "nomic-embed-text", "mxbai-embed-large" 等
	Timeout   time.Duration // 请求超时时间
	Dimension int           // 向量维度（如果已知，避免每次调用 API）
}

// NewOllamaEmbedderConfigFromEnv 从环境变量创建配置
func NewOllamaEmbedderConfigFromEnv() *OllamaEmbedderConfig {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_EMBED_MODEL")
	if model == "" {
		model = "qwen3-embedding:0.6b" // 默认嵌入模型
	}

	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("OLLAMA_TIMEOUT"); timeoutStr != "" {
		if parsed, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsed
		}
	}

	// 尝试从环境变量获取维度，如果没有则设为 0（需要动态获取）
	dimension := 0
	if dimStr := os.Getenv("OLLAMA_EMBED_DIMENSION"); dimStr != "" {
		fmt.Sscanf(dimStr, "%d", &dimension)
	} else {
		dimension = 768
	}

	return &OllamaEmbedderConfig{
		BaseURL:   baseURL,
		Model:     model,
		Timeout:   timeout,
		Dimension: dimension,
	}
}

// OllamaEmbedder Ollama 嵌入器实现
type OllamaEmbedder struct {
	config    *OllamaEmbedderConfig
	client    *http.Client
	baseURL   string
	model     string
	dimension int // 缓存的维度
}

// NewOllamaEmbedder 创建 Ollama 嵌入器
func NewOllamaEmbedder(config *OllamaEmbedderConfig) (*OllamaEmbedder, error) {
	if config == nil {
		config = NewOllamaEmbedderConfigFromEnv()
	}

	if config.Model == "" {
		return nil, fmt.Errorf("Ollama embedding model is required")
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	embedder := &OllamaEmbedder{
		config:    config,
		client:    client,
		baseURL:   config.BaseURL,
		model:     config.Model,
		dimension: config.Dimension,
	}

	// 如果维度未知，尝试获取一次
	if embedder.dimension == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if dim, err := embedder.fetchDimension(ctx); err == nil {
			embedder.dimension = dim
		} else {
			// 如果获取失败，使用常见模型的默认维度
			// nomic-embed-text 是 768 维
			embedder.dimension = 768
		}
	}

	return embedder, nil
}

// fetchDimension 通过调用 API 获取向量维度
func (o *OllamaEmbedder) fetchDimension(ctx context.Context) (int, error) {
	// 使用一个测试文本获取维度
	testText := "test"
	vectors, err := o.embedBatch(ctx, []string{testText})
	if err != nil {
		return 0, err
	}
	if len(vectors) > 0 && len(vectors[0]) > 0 {
		return len(vectors[0]), nil
	}
	return 0, fmt.Errorf("failed to get dimension from API")
}

// ollamaEmbedRequest Ollama Embedding API 请求结构
type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// ollamaEmbedResponse Ollama Embedding API 响应结构
type ollamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"` // Ollama 返回的是 float64
}

// embedBatch 批量嵌入（内部方法）
func (o *OllamaEmbedder) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// 构建请求
	reqBody := ollamaEmbedRequest{
		Model: o.model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/api/embed", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var embedResp ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 转换 float64 到 float32
	if len(embedResp.Embeddings) != len(texts) {
		return nil, fmt.Errorf("mismatched number of embeddings: expected %d, got %d", len(texts), len(embedResp.Embeddings))
	}

	results := make([][]float32, len(embedResp.Embeddings))
	for i, emb := range embedResp.Embeddings {
		results[i] = make([]float32, len(emb))
		for j, v := range emb {
			results[i][j] = float32(v)
		}
	}

	return results, nil
}

// EmbedText 将文本转换为向量嵌入
func (o *OllamaEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if o == nil {
		return nil, fmt.Errorf("OllamaEmbedder is nil")
	}
	if o.client == nil {
		return nil, fmt.Errorf("Ollama HTTP client is not initialized")
	}

	vectors, err := o.embedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(vectors) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return vectors[0], nil
}

// EmbedTexts 批量将文本转换为向量嵌入
func (o *OllamaEmbedder) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if o == nil {
		return nil, fmt.Errorf("OllamaEmbedder is nil")
	}
	if o.client == nil {
		return nil, fmt.Errorf("Ollama HTTP client is not initialized")
	}

	return o.embedBatch(ctx, texts)
}

// GetDimension 返回嵌入向量的维度
func (o *OllamaEmbedder) GetDimension() int {
	if o == nil {
		return 0
	}
	return o.dimension
}
