package main

import (
	"context"
	"fmt"
	"log"

	"goRag/internal/embedding"
	"goRag/internal/llm"
	"goRag/internal/rag"
	"goRag/internal/retriever"
)

// 这是一个超级简化的示例，帮助你理解 RAG 的核心流程
// 建议先看这个，再看完整的 example.go

func main() {
	fmt.Println("=== RAG 系统简化示例 ===\n")

	ctx := context.Background()

	// ============================================
	// 第一步：初始化各个组件
	// ============================================
	fmt.Println("【步骤 1】初始化组件...")

	// 1.1 创建嵌入器（把文本变成向量）
	embedder := embedding.NewSimpleEmbedder(128) // 128 维向量
	embeddingService := embedding.NewService(embedder)
	fmt.Println("  ✓ 嵌入器已创建（文本 → 向量）")

	// 1.2 创建检索器（存储和查找文档）
	memoryRetriever, err := retriever.NewMemoryRetriever(embedder)
	if err != nil {
		log.Fatalf("创建检索器失败: %v", err)
	}
	retrieverService := retriever.NewService(memoryRetriever)
	fmt.Println("  ✓ 检索器已创建（向量存储和查找）")

	// 1.3 创建 LLM（生成回答）
	llmImpl := llm.NewMockLLM()
	llmService := llm.NewService(llmImpl)
	fmt.Println("  ✓ LLM 已创建（生成回答）")

	// 1.4 组装成 RAG 服务
	ragService := rag.NewRAGService(
		embeddingService,
		retrieverService,
		llmService,
	)
	fmt.Println("  ✓ RAG 服务已组装完成\n")

	// ============================================
	// 第二步：添加一些文档到系统
	// ============================================
	fmt.Println("【步骤 2】添加文档到系统...")

	documents := []retriever.Document{
		{
			ID:      "doc1",
			Content: "Go 语言是 Google 开发的开源编程语言，特点是简洁、高效、并发友好。",
		},
		{
			ID:      "doc2",
			Content: "RAG 是检索增强生成技术，通过检索相关文档来增强 LLM 的回答准确性。",
		},
		{
			ID:      "doc3",
			Content: "向量数据库用于存储和检索高维向量，常用于相似度搜索。",
		},
	}

	if err := ragService.AddDocuments(ctx, documents); err != nil {
		log.Fatalf("添加文档失败: %v", err)
	}
	fmt.Printf("  ✓ 已添加 %d 个文档\n\n", len(documents))

	// ============================================
	// 第三步：用户提问，系统回答
	// ============================================
	fmt.Println("【步骤 3】用户提问，系统处理...\n")

	query := "什么是 Go 语言？"
	fmt.Printf("用户问题: %s\n\n", query)

	// RAG 的核心流程（在 rag.Query 方法中）：
	// 1. 把问题向量化
	// 2. 在文档库中找最相似的文档（Top-K）
	// 3. 把问题和文档组合成提示词
	// 4. 调用 LLM 生成回答

	fmt.Println("系统内部处理流程:")
	fmt.Println("  1. 向量化问题...")
	fmt.Println("  2. 检索相似文档...")
	fmt.Println("  3. 构建提示词...")
	fmt.Println("  4. 调用 LLM 生成回答...\n")

	answer, err := ragService.Query(ctx, query, 2) // topK=2，找最相似的 2 个文档
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}

	fmt.Printf("系统回答: %s\n\n", answer)

	// ============================================
	// 总结：RAG 的核心思想
	// ============================================
	fmt.Println("=== RAG 核心思想总结 ===")
	fmt.Println("传统 LLM: 用户问题 → LLM → 回答（可能不准确）")
	fmt.Println("RAG 系统: 用户问题 → 检索相关文档 → LLM（基于文档）→ 回答（更准确）")
	fmt.Println("\n关键点:")
	fmt.Println("  • 向量化：把文本变成数字，方便计算相似度")
	fmt.Println("  • 检索：找到最相关的文档")
	fmt.Println("  • 增强：把文档作为上下文给 LLM")
	fmt.Println("  • 生成：LLM 基于真实文档生成回答")
}

