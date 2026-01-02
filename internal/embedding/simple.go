package embedding

import (
	"context"
	"crypto/md5"
	"math"
	"strings"
	"unicode"
)

// SimpleEmbedder 简单的内存嵌入器实现
// 使用基于词频的简单向量化方法（适用于演示和测试）
type SimpleEmbedder struct {
	dimension int
	vocab     map[string]int
	vocabSize int
}

// NewSimpleEmbedder 创建简单嵌入器
// dimension: 向量维度
func NewSimpleEmbedder(dimension int) *SimpleEmbedder {
	return &SimpleEmbedder{
		dimension: dimension,
		vocab:     make(map[string]int),
		vocabSize: 0,
	}
}

// tokenize 分词（简单实现）
func (e *SimpleEmbedder) tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
	return words
}

// hashWord 将单词哈希到向量维度
func (e *SimpleEmbedder) hashWord(word string) int {
	hash := md5.Sum([]byte(word))
	hashInt := int(hash[0])<<24 | int(hash[1])<<16 | int(hash[2])<<8 | int(hash[3])
	return int(math.Abs(float64(hashInt))) % e.dimension
}

// EmbedText 将文本转换为向量嵌入
func (e *SimpleEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	vector := make([]float32, e.dimension)
	words := e.tokenize(text)
	
	if len(words) == 0 {
		return vector, nil
	}

	// 计算词频
	wordFreq := make(map[string]float32)
	for _, word := range words {
		wordFreq[word]++
	}

	// 归一化
	maxFreq := float32(0)
	for _, freq := range wordFreq {
		if freq > maxFreq {
			maxFreq = freq
		}
	}

	// 构建向量
	for word, freq := range wordFreq {
		index := e.hashWord(word)
		if maxFreq > 0 {
			vector[index] += freq / maxFreq
		}
	}

	// L2 归一化
	norm := float32(0)
	for _, v := range vector {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector, nil
}

// EmbedTexts 批量将文本转换为向量嵌入
func (e *SimpleEmbedder) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := e.EmbedText(ctx, text)
		if err != nil {
			return nil, err
		}
		results[i] = vec
	}
	return results, nil
}

// GetDimension 返回嵌入向量的维度
func (e *SimpleEmbedder) GetDimension() int {
	return e.dimension
}

