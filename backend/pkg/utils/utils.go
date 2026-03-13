package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// RandomDelay 随机延迟
func RandomDelay(min, max float64) time.Duration {
	if min >= max {
		return time.Duration(min * float64(time.Second))
	}
	delay := min + rand.Float64()*(max-min)
	return time.Duration(delay * float64(time.Second))
}

// SanitizeFilename 清理文件名
func SanitizeFilename(filename string) string {
	// 移除非法字符
	reg := regexp.MustCompile(`[\\/:*?"<>|]`)
	safe := reg.ReplaceAllString(filename, "")
	safe = strings.TrimSpace(safe)

	// 限制长度
	if len(safe) > 100 {
		safe = safe[:100]
	}

	if safe == "" {
		safe = "unknown"
	}

	return safe
}

// GetIDRange 获取ID范围（用于封面URL）
func GetIDRange(id int) string {
	if id < 1000 {
		return "0-999"
	}
	start := (id / 1000) * 1000
	end := start + 999
	return fmt.Sprintf("%d-%d", start, end)
}

// Contains 检查切片是否包含元素
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Unique 去重
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := []T{}
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// ChunkSlice 切片分块
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
