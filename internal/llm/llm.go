package llm

import (
	"context"
)

// Message 消息结构
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// LLM 大语言模型接口
type LLM interface {
	// Generate 生成回复
	Generate(ctx context.Context, messages []Message) (string, error)

	// GenerateStream 流式生成回复
	GenerateStream(ctx context.Context, messages []Message, callback func(string) error) error
}

// Service LLM 服务
type Service struct {
	llm LLM
}

// NewService 创建新的 LLM 服务
func NewService(llm LLM) *Service {
	return &Service{
		llm: llm,
	}
}

// Generate 生成回复
func (s *Service) Generate(ctx context.Context, messages []Message) (string, error) {
	return s.llm.Generate(ctx, messages)
}

// GenerateStream 流式生成回复
func (s *Service) GenerateStream(ctx context.Context, messages []Message, callback func(string) error) error {
	return s.llm.GenerateStream(ctx, messages, callback)
}
