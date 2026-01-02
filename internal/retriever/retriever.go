package retriever

import (
	"context"
)

// Document 文档结构
type Document struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
}

// RetrievalResult 检索结果
type RetrievalResult struct {
	Document Document
	Score    float64
}

// Retriever 检索器接口
type Retriever interface {
	// Retrieve 根据查询检索相关文档
	Retrieve(ctx context.Context, query string, topK int) ([]RetrievalResult, error)

	// AddDocuments 添加文档到检索器
	AddDocuments(ctx context.Context, documents []Document) error

	// DeleteDocument 删除文档
	DeleteDocument(ctx context.Context, documentID string) error
}

// Service 检索服务
type Service struct {
	retriever Retriever
}

// NewService 创建新的检索服务
func NewService(retriever Retriever) *Service {
	return &Service{
		retriever: retriever,
	}
}

// Retrieve 检索文档
func (s *Service) Retrieve(ctx context.Context, query string, topK int) ([]RetrievalResult, error) {
	return s.retriever.Retrieve(ctx, query, topK)
}

// AddDocuments 添加文档
func (s *Service) AddDocuments(ctx context.Context, documents []Document) error {
	return s.retriever.AddDocuments(ctx, documents)
}

// DeleteDocument 删除文档
func (s *Service) DeleteDocument(ctx context.Context, documentID string) error {
	return s.retriever.DeleteDocument(ctx, documentID)
}
