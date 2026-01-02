# goRag

一个基于 Go 语言的 RAG (Retrieval-Augmented Generation) 系统。

## 项目结构

```
├── cmd/
│   └── server/
│       └── main.go          # 应用程序入口
├── internal/
│   ├── embedding/           # 文本嵌入服务
│   ├── retriever/           # 文档检索服务
│   ├── ranker/              # 结果排序服务
│   ├── prompt/              # 提示词构建服务
│   ├── llm/                 # 大语言模型服务
│   ├── rag/                 # RAG 核心服务
│   └── api/                 # HTTP API 服务器
└── README.md
```

## 功能特性

- **文本嵌入**: 将文本转换为向量表示
- **文档检索**: 基于向量相似度的文档检索
- **结果排序**: 对检索结果进行排序和重排
- **提示词构建**: 动态构建 LLM 提示词
- **LLM 集成**: 支持多种大语言模型
- **RESTful API**: 提供 HTTP API 接口

## 快速开始

### 安装依赖

```bash
go mod download
```

### 运行服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API 接口

所有 API 接口都在 `/api/v1` 路径下。

### 健康检查

```bash
GET /api/v1/health
```

响应：
```json
{
  "status": "healthy"
}
```

### 查询

```bash
POST /api/v1/query
Content-Type: application/json

{
  "query": "你的问题",
  "top_k": 5
}
```

响应：
```json
{
  "answer": "基于检索到的文档生成的回答"
}
```

### 添加文档

```bash
POST /api/v1/documents
Content-Type: application/json

{
  "documents": [
    {
      "id": "doc1",
      "content": "文档内容",
      "metadata": {}
    }
  ]
}
```

响应：
```json
{
  "status": "success"
}
```

### 删除文档

```bash
DELETE /api/v1/documents?id=doc1
```

响应：
```json
{
  "status": "success"
}
```

## 架构说明

本项目采用模块化设计，参考 WeKnora 架构，各个组件通过接口解耦：

### 核心模块

1. **Embedding (嵌入模块)**
   - 接口: `embedding.Embedder`
   - 实现: `embedding.SimpleEmbedder` - 基于词频的简单向量化
   - 功能: 将文本转换为向量表示

2. **Retriever (检索模块)**
   - 接口: `retriever.Retriever`
   - 实现: `retriever.MemoryRetriever` - 内存向量检索器
   - 功能: 基于向量相似度的文档检索

3. **Ranker (排序模块)**
   - 接口: `ranker.Ranker`
   - 实现: 
     - `ranker.SimpleRanker` - 简单分数排序
     - `ranker.Reranker` - 重排序器（支持多样性过滤）
     - `ranker.BM25Ranker` - BM25 风格排序
   - 功能: 对检索结果进行排序和重排

4. **Prompt (提示模块)**
   - 功能: 动态构建 LLM 提示词模板
   - 支持自定义模板

5. **LLM (大语言模型模块)**
   - 接口: `llm.LLM`
   - 实现:
     - `llm.MockLLM` - Mock 实现（用于测试）
     - `llm.OpenAI` - OpenAI 接口（需要实现）
   - 功能: 生成回答

### 实现细节

- **简单嵌入器**: 使用词频和哈希函数生成固定维度向量
- **内存检索器**: 使用余弦相似度进行向量检索
- **Mock LLM**: 提供基本的模拟回复功能

### 扩展指南

你可以根据需要实现这些接口的具体实现：

- 替换 `SimpleEmbedder` 为 OpenAI Embeddings 或其他嵌入服务
- 替换 `MemoryRetriever` 为向量数据库（如 Qdrant、Pinecone、pgvector）
- 替换 `MockLLM` 为真实的 LLM API（OpenAI、Claude、本地模型等）

## 许可证

MIT License

