package embedding

import (
	"context"
)

// Embedder 定义嵌入接口
type Embedder interface {
	// EmbedText 将文本转换为向量嵌入
	EmbedText(ctx context.Context, text string) ([]float32, error)

	// EmbedTexts 批量将文本转换为向量嵌入
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)

	// GetDimension 返回嵌入向量的维度
	GetDimension() int
}

// Service 嵌入服务
type Service struct {
	embedder Embedder
}

// NewService 创建新的嵌入服务
func NewService(embedder Embedder) *Service {
	return &Service{
		embedder: embedder,
	}
}

// Embed 嵌入文本
func (s *Service) Embed(ctx context.Context, text string) ([]float32, error) {
	return s.embedder.EmbedText(ctx, text)
}

// EmbedBatch 批量嵌入文本
func (s *Service) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return s.embedder.EmbedTexts(ctx, texts)
}

// GetDimension 获取嵌入维度
func (s *Service) GetDimension() int {
	return s.embedder.GetDimension()
}
