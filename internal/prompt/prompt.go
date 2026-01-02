package prompt

import (
	"fmt"
	"strings"
)

// Template 提示模板
type Template struct {
	SystemPrompt string
	UserPrompt   string
}

// Builder 提示构建器
type Builder struct {
	template Template
}

// NewBuilder 创建新的提示构建器
func NewBuilder(template Template) *Builder {
	return &Builder{
		template: template,
	}
}

// Build 构建提示
func (b *Builder) Build(context string, query string) string {
	userPrompt := strings.ReplaceAll(b.template.UserPrompt, "{{context}}", context)
	userPrompt = strings.ReplaceAll(userPrompt, "{{query}}", query)

	return fmt.Sprintf("System: %s\n\nUser: %s", b.template.SystemPrompt, userPrompt)
}

// DefaultTemplate 默认模板
func DefaultTemplate() Template {
	return Template{
		SystemPrompt: "You are a helpful assistant that answers questions based on the provided context.",
		UserPrompt:   "Context: {{context}}\n\nQuestion: {{query}}\n\nAnswer:",
	}
}

// Service 提示服务
type Service struct {
	builder *Builder
}

// NewService 创建新的提示服务
func NewService(template Template) *Service {
	return &Service{
		builder: NewBuilder(template),
	}
}

// BuildPrompt 构建提示
func (s *Service) BuildPrompt(context string, query string) string {
	return s.builder.Build(context, query)
}
