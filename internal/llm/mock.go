package llm

import (
	"context"
	"fmt"
	"strings"
)

// MockLLM Mock LLM 实现（用于测试和演示）
type MockLLM struct {
	name string
}

// NewMockLLM 创建 Mock LLM
func NewMockLLM() *MockLLM {
	return &MockLLM{
		name: "MockLLM",
	}
}

// Generate 生成回复
func (m *MockLLM) Generate(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}

	// 提取用户消息内容
	var userContent string
	for _, msg := range messages {
		if msg.Role == "user" {
			userContent = msg.Content
			break
		}
	}

	if userContent == "" {
		// 如果没有用户消息，使用最后一条消息
		userContent = messages[len(messages)-1].Content
	}

	// 简单的模拟回复
	response := fmt.Sprintf("基于提供的信息，我理解您的问题。相关内容已包含在上下文中。这是一个模拟回复。\n\n原始查询: %s", 
		extractQuery(userContent))
	
	return response, nil
}

// GenerateStream 流式生成回复
func (m *MockLLM) GenerateStream(ctx context.Context, messages []Message, callback func(string) error) error {
	response, err := m.Generate(ctx, messages)
	if err != nil {
		return err
	}

	// 模拟流式输出
	words := strings.Fields(response)
	for _, word := range words {
		if err := callback(word + " "); err != nil {
			return err
		}
	}
	return nil
}

// extractQuery 从提示中提取查询
func extractQuery(prompt string) string {
	// 简单提取，查找 "Question:" 或 "查询:" 后的内容
	lines := strings.Split(prompt, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Question:") || strings.Contains(line, "查询:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return prompt
}

