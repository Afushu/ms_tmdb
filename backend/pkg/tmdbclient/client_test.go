package tmdbclient

import (
	"context"
	"testing"
	"time"
)

func TestClientLimiter(t *testing.T) {
	client := NewClient("dummy_key", "https://api.tmdb.org/3", "zh-CN", 10, "")
	if client.limiter == nil {
		t.Fatal("limiter 应该被初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 消费令牌测试
	if err := client.limiter.Wait(ctx); err != nil {
		t.Fatalf("正常获取限流令牌失败: %v", err)
	}
}
