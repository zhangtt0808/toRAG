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

func test() {
	ctx := context.Background()

	// 初始化各个服务
	embedder := embedding.NewSimpleEmbedder(128)
	embeddingService := embedding.NewService(embedder)

	memoryRetriever, err := retriever.NewMemoryRetriever(embedder)
	if err != nil {
		log.Fatalf("Failed to create retriever: %v", err)
	}
	retrieverService := retriever.NewService(memoryRetriever)

	llmImpl := llm.NewMockLLM()
	llmService := llm.NewService(llmImpl)

	// 创建 RAG 服务
	ragService := rag.NewRAGService(
		embeddingService,
		retrieverService,
		llmService,
	)

	// 添加示例文档
	documents := []retriever.Document{
		{
			ID:      "doc1",
			Content: "Go 是一种开源编程语言，由 Google 开发。它专为构建简单、可靠和高效的软件而设计。",
			Metadata: map[string]interface{}{
				"source": "example",
			},
		},
		{
			ID:      "doc2",
			Content: "RAG (Retrieval-Augmented Generation) 是一种结合检索和生成的技术，用于提高大语言模型的准确性。",
			Metadata: map[string]interface{}{
				"source": "example",
			},
		},
		{
			ID:      "doc3",
			Content: "向量数据库用于存储和检索高维向量，常用于相似度搜索和推荐系统。",
			Metadata: map[string]interface{}{
				"source": "example",
			},
		},
	}

	fmt.Println("Adding documents...")
	if err := ragService.AddDocuments(ctx, documents); err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}
	fmt.Printf("✓ Added %d documents\n\n", len(documents))

	// 执行查询
	queries := []string{
		"什么是 Go 语言？",
		"RAG 是什么？",
		"向量数据库的用途",
	}

	for _, query := range queries {
		fmt.Printf("Query: %s\n", query)
		answer, err := ragService.Query(ctx, query, 2)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("Answer: %s\n\n", answer)
	}
}
