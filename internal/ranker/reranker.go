package ranker

import (
	"context"
	"math"
	"sort"
)

// Reranker 重排序器（基于交叉编码器思想）
// 使用更复杂的排序算法，考虑多个因素
type Reranker struct {
	// 可以添加权重配置等
	scoreThreshold float64
}

// NewReranker 创建重排序器
func NewReranker(scoreThreshold float64) *Reranker {
	return &Reranker{
		scoreThreshold: scoreThreshold,
	}
}

// Rank 对结果进行重排序
// 使用更复杂的排序策略：分数、长度、多样性等
func (r *Reranker) Rank(ctx context.Context, items []RankedItem) ([]RankedItem, error) {
	if len(items) == 0 {
		return items, nil
	}

	result := make([]RankedItem, 0, len(items))
	
	// 过滤低分项
	for _, item := range items {
		if item.Score >= r.scoreThreshold {
			result = append(result, item)
		}
	}

	// 按分数降序排序
	sort.Slice(result, func(i, j int) bool {
		// 可以添加其他排序因素
		// 例如：多样性、长度等
		return result[i].Score > result[j].Score
	})

	// 应用多样性过滤（简单实现：去重相似结果）
	result = r.diversityFilter(result)

	return result, nil
}

// diversityFilter 多样性过滤（简单实现）
func (r *Reranker) diversityFilter(items []RankedItem) []RankedItem {
	if len(items) <= 1 {
		return items
	}

	filtered := make([]RankedItem, 0, len(items))
	seen := make(map[string]bool)

	for _, item := range items {
		// 简单的去重（实际可以使用更复杂的相似度计算）
		if !seen[item.ID] {
			seen[item.ID] = true
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// BM25Ranker BM25 风格的排序器
type BM25Ranker struct {
	k1 float64 // 词频饱和度参数
	b  float64 // 长度归一化参数
}

// NewBM25Ranker 创建 BM25 排序器
func NewBM25Ranker(k1, b float64) *BM25Ranker {
	if k1 <= 0 {
		k1 = 1.2
	}
	if b < 0 || b > 1 {
		b = 0.75
	}
	return &BM25Ranker{
		k1: k1,
		b:  b,
	}
}

// Rank 使用 BM25 算法排序
func (b *BM25Ranker) Rank(ctx context.Context, items []RankedItem) ([]RankedItem, error) {
	if len(items) == 0 {
		return items, nil
	}

	result := make([]RankedItem, len(items))
	copy(result, items)

	// 计算平均分数（用于长度归一化）
	avgScore := 0.0
	for _, item := range result {
		avgScore += item.Score
	}
	avgScore /= float64(len(result))

	// 应用 BM25 公式调整分数
	for i := range result {
		// 简化的 BM25 计算
		// 实际 BM25 需要词频和文档长度信息
		// 这里使用分数作为基础
		normalizedScore := result[i].Score / math.Max(avgScore, 0.001)
		result[i].Score = normalizedScore * result[i].Score
	}

	// 按调整后的分数排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result, nil
}

