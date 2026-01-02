package ranker

import (
	"context"
	"sort"
)

// RankedItem 排序项
type RankedItem struct {
	ID    string
	Score float64
}

// Ranker 排序器接口
type Ranker interface {
	// Rank 对结果进行排序
	Rank(ctx context.Context, items []RankedItem) ([]RankedItem, error)
}

// Service 排序服务
type Service struct {
	ranker Ranker
}

// NewService 创建新的排序服务
func NewService(ranker Ranker) *Service {
	return &Service{
		ranker: ranker,
	}
}

// Rank 排序结果
func (s *Service) Rank(ctx context.Context, items []RankedItem) ([]RankedItem, error) {
	return s.ranker.Rank(ctx, items)
}

// SimpleRanker 简单排序器（按分数降序）
type SimpleRanker struct{}

// NewSimpleRanker 创建简单排序器
func NewSimpleRanker() *SimpleRanker {
	return &SimpleRanker{}
}

// Rank 对结果进行排序
func (r *SimpleRanker) Rank(ctx context.Context, items []RankedItem) ([]RankedItem, error) {
	result := make([]RankedItem, len(items))
	copy(result, items)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result, nil
}
