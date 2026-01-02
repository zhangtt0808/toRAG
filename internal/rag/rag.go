package rag

import (
	"context"
	"fmt"
	"log"
	"strings"

	"goRag/internal/embedding"
	"goRag/internal/llm"
	"goRag/internal/prompt"
	"goRag/internal/ranker"
	"goRag/internal/retriever"
)

// RAGService RAG 服务
type RAGService struct {
	embeddingService *embedding.Service
	retrieverService *retriever.Service
	rankerService    *ranker.Service
	promptService    *prompt.Service
	llmService       *llm.Service
}

// NewRAGService 创建新的 RAG 服务（使用依赖注入）
func NewRAGService(
	embeddingService *embedding.Service,
	retrieverService *retriever.Service,
	llmService *llm.Service,
) *RAGService {
	promptTemplate := prompt.DefaultTemplate()

	return &RAGService{
		embeddingService: embeddingService,
		retrieverService: retrieverService,
		rankerService:    ranker.NewService(ranker.NewSimpleRanker()),
		promptService:    prompt.NewService(promptTemplate),
		llmService:       llmService,
	}
}

// Query 查询并生成回答
// 这是 RAG 系统的核心方法，实现了完整的 RAG 流程
//
// 流程说明：
// 1. 检索：根据用户问题，在文档库中找到最相关的文档
// 2. 构建上下文：把检索到的文档内容组合起来
// 3. 构建提示词：把问题和上下文组合成 LLM 能理解的格式
// 4. 生成回答：调用 LLM 基于上下文生成回答
//
// 参数：
//   - ctx: 上下文（用于超时控制等）
//   - query: 用户的问题
//   - topK: 返回最相关的 K 个文档（比如 topK=5 表示找 5 个最相关的）
//
// 返回：
//   - string: LLM 生成的回答
//   - error: 错误信息
func (r *RAGService) Query(ctx context.Context, query string, topK int) (string, error) {
	// 检查服务是否初始化
	if r.retrieverService == nil {
		return "", fmt.Errorf("retriever service is not initialized")
	}
	if r.llmService == nil {
		return "", fmt.Errorf("llm service is not initialized")
	}

	// ========== 步骤 1: 检索相关文档 ==========
	// 这里会：
	// 1. 把用户问题转换成向量（在 Retriever 内部调用 Embedding）
	// 2. 计算问题向量和所有文档向量的相似度
	// 3. 返回相似度最高的 topK 个文档
	results, err := r.retrieverService.Retrieve(ctx, query, topK)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve documents: %w", err)
	}

	// 如果没有找到相关文档，直接返回
	if len(results) == 0 {
		return "No relevant documents found.", nil
	}
	log.Println("retrieved documents: ", results)

	// ========== 步骤 2: 构建上下文 ==========
	// 把检索到的文档内容提取出来，组合成一个长文本
	// 这个长文本就是 LLM 的"参考资料"
	contextParts := make([]string, 0, len(results))
	for _, result := range results {
		contextParts = append(contextParts, result.Document.Content)
	}
	context := strings.Join(contextParts, "\n\n") // 用两个换行符分隔不同文档

	// ========== 步骤 3: 构建提示词 ==========
	// 把用户问题和检索到的文档内容组合成提示词
	// 提示词格式类似：
	//   Context: [文档内容1]\n\n[文档内容2]
	//   Question: [用户问题]
	//   Answer:
	promptText := r.promptService.BuildPrompt(context, query)

	// ========== 步骤 4: 生成回答 ==========
	// 调用 LLM，让它基于提示词生成回答
	// LLM 会看到：
	// - 用户的问题
	// - 相关的文档内容
	// 然后基于这些信息生成回答
	messages := []llm.Message{
		{
			Role:    "user",
			Content: promptText,
		},
	}
	log.Println("prompt messages: ", messages)

	answer, err := r.llmService.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	return answer, nil
}

// AddDocuments 添加文档
func (r *RAGService) AddDocuments(ctx context.Context, documents []retriever.Document) error {
	if r.retrieverService == nil {
		return fmt.Errorf("retriever service is not initialized")
	}
	return r.retrieverService.AddDocuments(ctx, documents)
}

// DeleteDocument 删除文档
func (r *RAGService) DeleteDocument(ctx context.Context, documentID string) error {
	if r.retrieverService == nil {
		return fmt.Errorf("retriever service is not initialized")
	}
	return r.retrieverService.DeleteDocument(ctx, documentID)
}
