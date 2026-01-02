package retriever

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"

	"goRag/internal/embedding"
)

// MemoryRetriever 内存向量检索器
// 使用内存存储文档和向量，支持向量相似度检索
type MemoryRetriever struct {
	mu        sync.RWMutex
	documents map[string]Document
	vectors   map[string][]float32
	embedder  embedding.Embedder
	dimension int
}

// NewMemoryRetriever 创建内存检索器
func NewMemoryRetriever(embedder embedding.Embedder) (*MemoryRetriever, error) {
	if embedder == nil {
		return nil, fmt.Errorf("embedder cannot be nil")
	}

	return &MemoryRetriever{
		documents: make(map[string]Document),
		vectors:   make(map[string][]float32),
		embedder:  embedder,
		dimension: embedder.GetDimension(),
	}, nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// AddDocuments 添加文档到检索器
func (m *MemoryRetriever) AddDocuments(ctx context.Context, documents []Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 批量嵌入文档内容
	texts := make([]string, len(documents))
	for i, doc := range documents {
		texts[i] = doc.Content
	}

	vectors, err := m.embedder.EmbedTexts(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to embed documents: %w", err)
	}

	// 存储文档和向量
	for i, doc := range documents {
		m.documents[doc.ID] = doc
		m.vectors[doc.ID] = vectors[i]
	}

	return nil
}

// DeleteDocument 删除文档
func (m *MemoryRetriever) DeleteDocument(ctx context.Context, documentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.documents, documentID)
	delete(m.vectors, documentID)
	return nil
}

// Retrieve 根据查询检索相关文档
func (m *MemoryRetriever) Retrieve(ctx context.Context, query string, topK int) ([]RetrievalResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 嵌入查询
	queryVector, err := m.embedder.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	log.Println("query vector: ", queryVector)

	// 计算所有文档的相似度
	type scoreDoc struct {
		doc   Document
		score float64
	}

	scores := make([]scoreDoc, 0, len(m.documents))
	for id, doc := range m.documents {
		vector := m.vectors[id]
		score := cosineSimilarity(queryVector, vector)
		scores = append(scores, scoreDoc{
			doc:   doc,
			score: score,
		})
		// log the scores and the document id
		log.Println("score: ", score, "document id: ", id)
	}

	// 按分数排序（降序）
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].score < scores[j].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// 取 topK
	if topK > len(scores) {
		topK = len(scores)
	}

	results := make([]RetrievalResult, topK)
	for i := 0; i < topK; i++ {
		results[i] = RetrievalResult{
			Document: scores[i].doc,
			Score:    scores[i].score,
		}
	}

	return results, nil
}
