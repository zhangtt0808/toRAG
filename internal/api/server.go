package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"goRag/internal/rag"
	"goRag/internal/retriever"
)

// Server API 服务器
type Server struct {
	ragService *rag.RAGService
	router     *gin.Engine
	httpServer *http.Server
}

// NewServer 创建新的 API 服务器
func NewServer(ragService *rag.RAGService) *Server {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	server := &Server{
		ragService: ragService,
		router:     router,
		httpServer: &http.Server{
			Addr:    ":8080",
			Handler: router,
		},
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// registerRoutes 注册所有路由
func (s *Server) registerRoutes() {
	api := s.router.Group("/api/v1")
	{
		api.POST("/query", s.handleQuery)
		api.POST("/documents", s.handleAddDocuments)
		api.DELETE("/documents", s.handleDeleteDocument)
		api.GET("/health", s.handleHealth)
	}
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	log.Printf("Starting server on %s", s.httpServer.Addr)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(shutdownCtx)
	}()

	return s.httpServer.ListenAndServe()
}

// QueryRequest 查询请求
type QueryRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k,omitempty"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Answer string `json:"answer"`
}

// DocumentRequest 文档请求
type DocumentRequest struct {
	Documents []DocumentItem `json:"documents" binding:"required"`
}

// DocumentItem 文档项
type DocumentItem struct {
	ID       string                 `json:"id" binding:"required"`
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// handleQuery 处理查询请求
func (s *Server) handleQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	answer, err := s.ragService.Query(c.Request.Context(), req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, QueryResponse{Answer: answer})
}

// handleAddDocuments 处理添加文档请求
func (s *Server) handleAddDocuments(c *gin.Context) {
	var req DocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	documents := make([]retriever.Document, len(req.Documents))
	for i, doc := range req.Documents {
		documents[i] = retriever.Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Metadata: doc.Metadata,
		}
	}

	if err := s.ragService.AddDocuments(c.Request.Context(), documents); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success"})
}

// handleDeleteDocument 处理删除文档请求
func (s *Server) handleDeleteDocument(c *gin.Context) {
	documentID := c.Query("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Document ID is required"})
		return
	}

	if err := s.ragService.DeleteDocument(c.Request.Context(), documentID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleHealth 处理健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
