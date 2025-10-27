package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestOptimizedBatchProcessor_Basic(t *testing.T) {
	config := &Config{
		BufferSize:  10,
		FlushPeriod: 100 * time.Millisecond,
		Export: &ExportConfig{
			BatchSize: 5,
			Timeout:   5 * time.Second,
			Endpoints: map[string]string{
				"test": "console://",
			},
		},
	}

	bp := NewOptimizedBatchProcessor(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bp.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start optimized batch processor: %v", err)
	}
	defer bp.Stop()

	// 添加一些测试项目
	for i := 0; i < 3; i++ {
		item := BatchItem{
			Type:      "metric",
			Name:      "test_metric",
			Value:     float64(i),
			Timestamp: time.Now(),
			Labels: map[string]string{
				"test": "value",
			},
		}
		
		err := bp.AddItemOptimized(item)
		if err != nil {
			t.Errorf("Failed to add item %d: %v", i, err)
		}
	}

	// 等待处理
	time.Sleep(200 * time.Millisecond)

	stats := bp.GetOptimizedStats()
	if stats.TotalItems != 3 {
		t.Errorf("Expected 3 total items, got %d", stats.TotalItems)
	}
}

func TestOptimizedBatchProcessor_Cache(t *testing.T) {
	// 创建测试HTTP服务器
	var requestCount int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &Config{
		BufferSize:  10,
		FlushPeriod: 50 * time.Millisecond,
		Export: &ExportConfig{
			BatchSize: 2,
			Timeout:   5 * time.Second,
			Endpoints: map[string]string{
				"test": server.URL,
			},
		},
	}

	bp := NewOptimizedBatchProcessor(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bp.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start optimized batch processor: %v", err)
	}
	defer bp.Stop()

	// 添加相同的项目多次（应该被缓存）
	item := BatchItem{
		Type:      "metric",
		Name:      "cached_metric",
		Value:     100.0,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"cache": "test",
		},
	}

	// 第一次添加
	for i := 0; i < 2; i++ {
		err := bp.AddItemOptimized(item)
		if err != nil {
			t.Errorf("Failed to add item %d: %v", i, err)
		}
	}

	// 等待第一批处理
	time.Sleep(100 * time.Millisecond)

	// 再次添加相同的项目
	for i := 0; i < 2; i++ {
		err := bp.AddItemOptimized(item)
		if err != nil {
			t.Errorf("Failed to add cached item %d: %v", i, err)
		}
	}

	// 等待第二批处理
	time.Sleep(100 * time.Millisecond)

	// 检查缓存统计
	cacheStats := bp.GetCacheStats()
	t.Logf("Cache stats: %+v", cacheStats)

	// 检查请求计数
	mu.Lock()
	finalRequestCount := requestCount
	mu.Unlock()

	t.Logf("Total HTTP requests: %d", finalRequestCount)
}

func TestOptimizedBatchProcessor_Performance(t *testing.T) {
	// 创建测试HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &Config{
		BufferSize:  1000,
		FlushPeriod: 100 * time.Millisecond,
		Export: &ExportConfig{
			BatchSize: 100,
			Timeout:   5 * time.Second,
			Endpoints: map[string]string{
				"test": server.URL,
			},
		},
	}

	bp := NewOptimizedBatchProcessor(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := bp.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start optimized batch processor: %v", err)
	}
	defer bp.Stop()

	// 性能测试：添加大量项目
	startTime := time.Now()
	itemCount := 1000

	for i := 0; i < itemCount; i++ {
		item := BatchItem{
			Type:      "metric",
			Name:      "perf_metric",
			Value:     float64(i),
			Timestamp: time.Now(),
			Labels: map[string]string{
				"index": string(rune(i)),
			},
		}
		
		err := bp.AddItemOptimized(item)
		if err != nil {
			t.Errorf("Failed to add item %d: %v", i, err)
		}
	}

	addDuration := time.Since(startTime)
	t.Logf("Added %d items in %v (%.2f items/sec)", 
		itemCount, addDuration, float64(itemCount)/addDuration.Seconds())

	// 等待所有项目处理完成
	time.Sleep(2 * time.Second)

	stats := bp.GetOptimizedStats()
	t.Logf("Final stats: %+v", stats)

	if stats.TotalItems != int64(itemCount) {
		t.Errorf("Expected %d total items, got %d", itemCount, stats.TotalItems)
	}
}

func TestOptimizedBatchProcessor_ConnectionPool(t *testing.T) {
	// 创建测试HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟一些处理时间
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &Config{
		BufferSize:  50,
		FlushPeriod: 50 * time.Millisecond,
		Export: &ExportConfig{
			BatchSize: 10,
			Timeout:   5 * time.Second,
			Endpoints: map[string]string{
				"test": server.URL,
			},
		},
	}

	bp := NewOptimizedBatchProcessor(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := bp.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start optimized batch processor: %v", err)
	}
	defer bp.Stop()

	// 添加项目以触发多个并发请求
	for i := 0; i < 50; i++ {
		item := BatchItem{
			Type:      "metric",
			Name:      "conn_test_metric",
			Value:     float64(i),
			Timestamp: time.Now(),
		}
		
		err := bp.AddItemOptimized(item)
		if err != nil {
			t.Errorf("Failed to add item %d: %v", i, err)
		}
	}

	// 等待处理
	time.Sleep(1 * time.Second)

	// 检查连接池统计
	poolStats := bp.GetConnectionPoolStats()
	t.Logf("Connection pool stats: %+v", poolStats)

	for name, stats := range poolStats {
		t.Logf("Exporter %s: %+v", name, stats)
	}
}

func TestOptimizedBatchProcessor_BufferUtilization(t *testing.T) {
	config := &Config{
		BufferSize:  10,
		FlushPeriod: 1 * time.Second, // 长刷新周期以测试缓冲区利用率
		Export: &ExportConfig{
			BatchSize: 20, // 大于缓冲区大小
			Timeout:   5 * time.Second,
			Endpoints: map[string]string{
				"test": "console://",
			},
		},
	}

	bp := NewOptimizedBatchProcessor(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bp.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start optimized batch processor: %v", err)
	}
	defer bp.Stop()

	// 添加项目到缓冲区
	for i := 0; i < 5; i++ {
		item := BatchItem{
			Type:  "metric",
			Name:  "buffer_test",
			Value: float64(i),
		}
		
		err := bp.AddItemOptimized(item)
		if err != nil {
			t.Errorf("Failed to add item %d: %v", i, err)
		}
		
		utilization := bp.GetBufferUtilization()
		expectedUtilization := float64(i+1) / 10.0
		
		if utilization != expectedUtilization {
			t.Errorf("Expected buffer utilization %.2f, got %.2f", 
				expectedUtilization, utilization)
		}
	}
}

func TestOptimizedBatchProcessor_StatsReset(t *testing.T) {
	config := &Config{
		BufferSize:  10,
		FlushPeriod: 100 * time.Millisecond,
		Export: &ExportConfig{
			BatchSize: 5,
			Timeout:   5 * time.Second,
			Endpoints: map[string]string{
				"test": "console://",
			},
		},
	}

	bp := NewOptimizedBatchProcessor(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bp.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start optimized batch processor: %v", err)
	}
	defer bp.Stop()

	// 添加一些项目
	for i := 0; i < 3; i++ {
		item := BatchItem{
			Type:  "metric",
			Name:  "reset_test",
			Value: float64(i),
		}
		
		err := bp.AddItemOptimized(item)
		if err != nil {
			t.Errorf("Failed to add item %d: %v", i, err)
		}
	}

	// 等待处理
	time.Sleep(200 * time.Millisecond)

	// 检查统计信息
	stats := bp.GetOptimizedStats()
	if stats.TotalItems == 0 {
		t.Error("Expected non-zero total items before reset")
	}

	// 重置统计信息
	bp.ResetStats()

	// 检查重置后的统计信息
	resetStats := bp.GetOptimizedStats()
	if resetStats.TotalItems != 0 {
		t.Errorf("Expected zero total items after reset, got %d", resetStats.TotalItems)
	}
}

// 基准测试
func BenchmarkOptimizedBatchProcessor_AddItem(b *testing.B) {
	config := &Config{
		BufferSize:  10000,
		FlushPeriod: 1 * time.Second,
		Export: &ExportConfig{
			BatchSize: 1000,
			Timeout:   5 * time.Second,
			Endpoints: map[string]string{
				"test": "console://",
			},
		},
	}

	bp := NewOptimizedBatchProcessor(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := bp.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start optimized batch processor: %v", err)
	}
	defer bp.Stop()

	item := BatchItem{
		Type:      "metric",
		Name:      "benchmark_metric",
		Value:     100.0,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"benchmark": "test",
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := bp.AddItemOptimized(item)
			if err != nil {
				b.Errorf("Failed to add item: %v", err)
			}
		}
	})
}

func BenchmarkOptimizedBatchProcessor_Serialization(b *testing.B) {
	items := make([]BatchItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = BatchItem{
			Type:      "metric",
			Name:      "benchmark_metric",
			Value:     float64(i),
			Timestamp: time.Now(),
			Labels: map[string]string{
				"index": string(rune(i)),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(items)
		if err != nil {
			b.Errorf("Failed to marshal items: %v", err)
		}
	}
}