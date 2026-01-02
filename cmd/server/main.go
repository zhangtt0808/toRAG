package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goRag/internal/api"
	"goRag/internal/embedding"
	"goRag/internal/llm"
	"goRag/internal/rag"
	"goRag/internal/retriever"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Initializing RAG system...")

	// 1. 初始化嵌入服务

	var embedder embedding.Embedder
	embedder, err := embedding.NewOllamaEmbedder(embedding.NewOllamaEmbedderConfigFromEnv())
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}
	embeddingService := embedding.NewService(embedder)
	log.Println("✓ Embedding service initialized")

	// 2. 初始化检索服务
	memoryRetriever, err := retriever.NewMemoryRetriever(embedder)
	if err != nil {
		log.Fatalf("Failed to create memory retriever: %v", err)
	}
	retrieverService := retriever.NewService(memoryRetriever)
	log.Println("✓ Retriever service initialized")

	// 3. 初始化 LLM 服务
	// 优先尝试使用 OpenAI，如果没有配置或连接失败，则使用 Mock LLM
	var llmImpl llm.LLM
	ollamaConfig := llm.NewOllamaConfigFromEnv()
	if ollamaConfig != nil {
		ollamaLLM, err := llm.NewOllama(ollamaConfig)
		if err == nil && ollamaLLM != nil {
			llmImpl = ollamaLLM
			log.Println("✓ LLM service initialized (using Ollama)")
		}
	}
	if llmImpl == nil {
		llmImpl = llm.NewMockLLM()
		log.Println("✓ LLM service initialized (using MockLLM)")
	}
	llmService := llm.NewService(llmImpl)

	// 4. 初始化 RAG 服务
	ragService := rag.NewRAGService(
		embeddingService,
		retrieverService,
		llmService,
	)
	log.Println("✓ RAG service initialized")

	// 5. 初始化 API 服务器
	apiServer := api.NewServer(ragService)
	log.Println("✓ API server initialized")

	// 启动服务器
	go func() {
		if err := apiServer.Start(ctx); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("Server is running on http://localhost:8080")
	log.Println("API endpoints:")
	log.Println("  POST   /api/v1/query      - Query documents")
	log.Println("  POST   /api/v1/documents   - Add documents")
	log.Println("  DELETE /api/v1/documents  - Delete document")
	log.Println("  GET    /api/v1/health     - Health check")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	cancel()
}
