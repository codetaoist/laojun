package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/codetaoist/laojun-shared/observability"
)

func main() {
	fmt.Println("=== 性能优化和批量处理示例 ===")

	// 测试1: 基本批量处理
	testBasicBatchProcessing()

	// 测试2: 高并发数据写入
	testHighConcurrencyWrites()

	// 测试3: 批量大小和刷新策略
	testBatchSizeAndFlushStrategy()

	// 测试4: 导出器性能测试
	testExporterPerformance()

	// 测试5: 内存使用优化
	testMemoryOptimization()

	// 测试6: 错误处理和重试机制
	testErrorHandlingAndRetry()

	fmt.Println("=== 性能优化示例完成! ===")
}

// testBasicBatchProcessing 测试基本批量处理
func testBasicBatchProcessing() {
	fmt.Println("\n--- 测试1: 基本批量处理 ---")

	// 创建配置
	config := &observability.Config{
		ServiceName:    "performance-test",
		ServiceVersion: "1.0.0",
		Environment:    "testing",
		SampleRate:     1.0,
		Timeout:        30 * time.Second,
		BufferSize:     1000,
		FlushPeriod:    2 * time.Second,
		Export: &observability.ExportConfig{
			Endpoints: map[string]string{
				"console": "console://localhost",
			},
			Formats:      []string{"json"},
			Timeout:      5 * time.Second,
			MaxRetries:   2,
			RetryDelay:   500 * time.Millisecond,
			BatchSize:    10,
			BatchTimeout: 1 * time.Second,
			MaxQueueSize: 1000,
		},
	}

	// 创建批量处理器
	processor := observability.NewBatchProcessor(config)

	// 启动处理器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := processor.Start(ctx); err != nil {
		log.Fatalf("Failed to start batch processor: %v", err)
	}
	defer processor.Stop()

	fmt.Printf("批量处理器已启动，缓冲区大小: %d\n", config.BufferSize)

	// 添加一些测试数据
	for i := 0; i < 25; i++ {
		// 添加指标
		err := processor.AddMetric(
			fmt.Sprintf("test_metric_%d", i),
			float64(i*10),
			map[string]string{
				"service": "test",
				"method":  fmt.Sprintf("method_%d", i%3),
			},
		)
		if err != nil {
			fmt.Printf("添加指标失败: %v\n", err)
		}

		// 添加事件
		err = processor.AddEvent(
			fmt.Sprintf("test_event_%d", i),
			map[string]interface{}{
				"user_id":   fmt.Sprintf("user_%d", i%5),
				"action":    "click",
				"timestamp": time.Now().Unix(),
			},
		)
		if err != nil {
			fmt.Printf("添加事件失败: %v\n", err)
		}

		// 模拟一些延迟
		if i%5 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 等待一段时间让数据被处理
	time.Sleep(3 * time.Second)

	// 强制刷新剩余数据
	processor.ForceFlush()
	time.Sleep(1 * time.Second)

	// 显示统计信息
	stats := processor.GetStats()
	fmt.Printf("处理统计:\n")
	fmt.Printf("  总项目数: %d\n", stats.TotalItems)
	fmt.Printf("  已处理项目数: %d\n", stats.ProcessedItems)
	fmt.Printf("  丢弃项目数: %d\n", stats.DroppedItems)
	fmt.Printf("  导出批次数: %d\n", stats.ExportedBatches)
	fmt.Printf("  失败导出数: %d\n", stats.FailedExports)
	fmt.Printf("  当前缓冲区大小: %d\n", stats.BufferSize)
	fmt.Printf("  平均刷新时间: %v\n", stats.AverageFlushTime)
}

// testHighConcurrencyWrites 测试高并发数据写入
func testHighConcurrencyWrites() {
	fmt.Println("\n--- 测试2: 高并发数据写入 ---")

	config := &observability.Config{
		ServiceName:    "concurrency-test",
		ServiceVersion: "1.0.0",
		Environment:    "testing",
		SampleRate:     1.0,
		Timeout:        30 * time.Second,
		BufferSize:     5000,
		FlushPeriod:    1 * time.Second,
		Export: &observability.ExportConfig{
			Endpoints: map[string]string{
				"console": "console://localhost",
			},
			Formats:      []string{"json"},
			Timeout:      5 * time.Second,
			MaxRetries:   1,
			RetryDelay:   100 * time.Millisecond,
			BatchSize:    100,
			BatchTimeout: 500 * time.Millisecond,
			MaxQueueSize: 5000,
		},
	}

	processor := observability.NewBatchProcessor(config)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := processor.Start(ctx); err != nil {
		log.Fatalf("Failed to start batch processor: %v", err)
	}
	defer processor.Stop()

	// 并发写入测试
	const numGoroutines = 10
	const itemsPerGoroutine = 100

	var wg sync.WaitGroup
	startTime := time.Now()

	fmt.Printf("启动 %d 个goroutine，每个写入 %d 个项目...\n", numGoroutines, itemsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < itemsPerGoroutine; j++ {
				// 随机选择数据类型
				switch rand.Intn(3) {
				case 0:
					processor.AddMetric(
						fmt.Sprintf("concurrent_metric_%d_%d", goroutineID, j),
						rand.Float64()*100,
						map[string]string{
							"goroutine": fmt.Sprintf("%d", goroutineID),
							"type":      "metric",
						},
					)
				case 1:
					processor.AddEvent(
						fmt.Sprintf("concurrent_event_%d_%d", goroutineID, j),
						map[string]interface{}{
							"goroutine": goroutineID,
							"sequence":  j,
							"random":    rand.Intn(1000),
						},
					)
				case 2:
					processor.AddTrace(
						fmt.Sprintf("concurrent_trace_%d_%d", goroutineID, j),
						map[string]interface{}{
							"goroutine":  goroutineID,
							"operation": "test_operation",
							"duration":  rand.Intn(100),
						},
					)
				}
			}
		}(i)
	}

	wg.Wait()
	writeTime := time.Since(startTime)

	fmt.Printf("并发写入完成，耗时: %v\n", writeTime)
	fmt.Printf("写入速率: %.2f items/second\n", float64(numGoroutines*itemsPerGoroutine)/writeTime.Seconds())

	// 等待处理完成
	time.Sleep(3 * time.Second)
	processor.ForceFlush()
	time.Sleep(1 * time.Second)

	stats := processor.GetStats()
	fmt.Printf("并发处理统计:\n")
	fmt.Printf("  总项目数: %d\n", stats.TotalItems)
	fmt.Printf("  已处理项目数: %d\n", stats.ProcessedItems)
	fmt.Printf("  丢弃项目数: %d\n", stats.DroppedItems)
	fmt.Printf("  导出批次数: %d\n", stats.ExportedBatches)
	fmt.Printf("  处理效率: %.2f%%\n", float64(stats.ProcessedItems)/float64(stats.TotalItems)*100)
}

// testBatchSizeAndFlushStrategy 测试批量大小和刷新策略
func testBatchSizeAndFlushStrategy() {
	fmt.Println("\n--- 测试3: 批量大小和刷新策略 ---")

	testConfigs := []struct {
		name      string
		batchSize int
		flushPeriod time.Duration
	}{
		{"小批量快刷新", 5, 500 * time.Millisecond},
		{"中批量中刷新", 20, 1 * time.Second},
		{"大批量慢刷新", 50, 2 * time.Second},
	}

	for _, tc := range testConfigs {
		fmt.Printf("\n测试配置: %s (批量大小: %d, 刷新周期: %v)\n", tc.name, tc.batchSize, tc.flushPeriod)

		config := &observability.Config{
			ServiceName:    "batch-strategy-test",
			ServiceVersion: "1.0.0",
			Environment:    "testing",
			SampleRate:     1.0,
			Timeout:        30 * time.Second,
			BufferSize:     1000,
			FlushPeriod:    tc.flushPeriod,
			Export: &observability.ExportConfig{
				Endpoints: map[string]string{
					"console": "console://localhost",
				},
				Formats:      []string{"json"},
				Timeout:      5 * time.Second,
				MaxRetries:   1,
				RetryDelay:   100 * time.Millisecond,
				BatchSize:    tc.batchSize,
				BatchTimeout: tc.flushPeriod / 2,
				MaxQueueSize: 1000,
			},
		}

		processor := observability.NewBatchProcessor(config)

		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		
		if err := processor.Start(ctx); err != nil {
			log.Printf("Failed to start batch processor: %v", err)
			cancel()
			continue
		}

		startTime := time.Now()

		// 添加固定数量的数据
		for i := 0; i < 30; i++ {
			processor.AddMetric(
				fmt.Sprintf("strategy_test_metric_%d", i),
				float64(i),
				map[string]string{"strategy": tc.name},
			)
			time.Sleep(50 * time.Millisecond) // 模拟真实数据流
		}

		// 等待处理完成
		time.Sleep(tc.flushPeriod * 2)
		processor.ForceFlush()
		time.Sleep(500 * time.Millisecond)

		totalTime := time.Since(startTime)
		stats := processor.GetStats()

		fmt.Printf("  处理时间: %v\n", totalTime)
		fmt.Printf("  导出批次数: %d\n", stats.ExportedBatches)
		fmt.Printf("  平均批次大小: %.1f\n", float64(stats.ProcessedItems)/float64(stats.ExportedBatches))
		fmt.Printf("  平均刷新时间: %v\n", stats.AverageFlushTime)

		processor.Stop()
		cancel()
	}
}

// testExporterPerformance 测试导出器性能
func testExporterPerformance() {
	fmt.Println("\n--- 测试4: 导出器性能测试 ---")

	config := &observability.Config{
		ServiceName:    "exporter-perf-test",
		ServiceVersion: "1.0.0",
		Environment:    "testing",
		SampleRate:     1.0,
		Timeout:        30 * time.Second,
		BufferSize:     2000,
		FlushPeriod:    1 * time.Second,
		Export: &observability.ExportConfig{
			Endpoints: map[string]string{
				"console": "console://localhost",
				"file":    "file:///tmp/test.log",
			},
			Formats:      []string{"json"},
			Timeout:      3 * time.Second,
			MaxRetries:   2,
			RetryDelay:   200 * time.Millisecond,
			BatchSize:    50,
			BatchTimeout: 500 * time.Millisecond,
			MaxQueueSize: 2000,
		},
	}

	processor := observability.NewBatchProcessor(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := processor.Start(ctx); err != nil {
		log.Fatalf("Failed to start batch processor: %v", err)
	}
	defer processor.Stop()

	fmt.Printf("导出器列表: %v\n", processor.GetExporters())

	// 生成大量数据测试导出性能
	const totalItems = 200
	startTime := time.Now()

	for i := 0; i < totalItems; i++ {
		processor.AddMetric(
			fmt.Sprintf("perf_metric_%d", i),
			rand.Float64()*1000,
			map[string]string{
				"category": fmt.Sprintf("cat_%d", i%10),
				"priority": fmt.Sprintf("p%d", i%3),
			},
		)
	}

	dataGenTime := time.Since(startTime)
	fmt.Printf("数据生成时间: %v (%.2f items/sec)\n", 
		dataGenTime, float64(totalItems)/dataGenTime.Seconds())

	// 等待所有数据被处理
	time.Sleep(3 * time.Second)
	processor.ForceFlush()
	time.Sleep(1 * time.Second)

	totalTime := time.Since(startTime)
	stats := processor.GetStats()

	fmt.Printf("总处理时间: %v\n", totalTime)
	fmt.Printf("处理吞吐量: %.2f items/sec\n", float64(stats.ProcessedItems)/totalTime.Seconds())
	fmt.Printf("导出效率: %.2f batches/sec\n", float64(stats.ExportedBatches)/totalTime.Seconds())
}

// testMemoryOptimization 测试内存使用优化
func testMemoryOptimization() {
	fmt.Println("\n--- 测试5: 内存使用优化 ---")

	config := &observability.Config{
		ServiceName:    "memory-opt-test",
		ServiceVersion: "1.0.0",
		Environment:    "testing",
		SampleRate:     1.0,
		Timeout:        30 * time.Second,
		BufferSize:     100, // 较小的缓冲区
		FlushPeriod:    200 * time.Millisecond, // 频繁刷新
		Export: &observability.ExportConfig{
			Endpoints: map[string]string{
				"console": "console://localhost",
			},
			Formats:      []string{"json"},
			Timeout:      2 * time.Second,
			MaxRetries:   1,
			RetryDelay:   50 * time.Millisecond,
			BatchSize:    20, // 小批量
			BatchTimeout: 100 * time.Millisecond,
			MaxQueueSize: 100,
		},
	}

	processor := observability.NewBatchProcessor(config)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := processor.Start(ctx); err != nil {
		log.Fatalf("Failed to start batch processor: %v", err)
	}
	defer processor.Stop()

	fmt.Printf("内存优化配置 - 缓冲区大小: %d, 批量大小: %d\n", 
		config.BufferSize, config.Export.BatchSize)

	// 持续添加数据，测试内存使用
	const duration = 5 * time.Second
	const itemsPerSecond = 50

	startTime := time.Now()
	itemCount := 0

	for time.Since(startTime) < duration {
		processor.AddMetric(
			fmt.Sprintf("memory_test_metric_%d", itemCount),
			rand.Float64()*100,
			map[string]string{
				"test": "memory_optimization",
			},
		)
		itemCount++

		// 监控缓冲区大小
		if itemCount%20 == 0 {
			bufferSize := processor.GetBufferSize()
			fmt.Printf("项目 %d: 缓冲区大小 = %d\n", itemCount, bufferSize)
		}

		time.Sleep(time.Second / itemsPerSecond)
	}

	// 最终统计
	time.Sleep(1 * time.Second)
	stats := processor.GetStats()

	fmt.Printf("内存优化测试结果:\n")
	fmt.Printf("  生成项目数: %d\n", itemCount)
	fmt.Printf("  处理项目数: %d\n", stats.ProcessedItems)
	fmt.Printf("  最终缓冲区大小: %d\n", stats.BufferSize)
	fmt.Printf("  内存使用效率: %.2f%%\n", 
		float64(stats.ProcessedItems)/float64(itemCount)*100)
}

// testErrorHandlingAndRetry 测试错误处理和重试机制
func testErrorHandlingAndRetry() {
	fmt.Println("\n--- 测试6: 错误处理和重试机制 ---")

	config := &observability.Config{
		ServiceName:    "error-handling-test",
		ServiceVersion: "1.0.0",
		Environment:    "testing",
		SampleRate:     1.0,
		Timeout:        30 * time.Second,
		BufferSize:     500,
		FlushPeriod:    1 * time.Second,
		Export: &observability.ExportConfig{
			Endpoints: map[string]string{
				"console":     "console://localhost",
				"invalid-url": "http://invalid-endpoint-that-will-fail:9999/api/data",
			},
			Formats:       []string{"json"},
			Timeout:       1 * time.Second, // 短超时以触发错误
			MaxRetries:    3,
			RetryDelay:    200 * time.Millisecond,
			BatchSize:     10,
			BatchTimeout:  500 * time.Millisecond,
			MaxQueueSize:  500,
			DropOnFailure: false, // 不丢弃失败的数据
		},
	}

	processor := observability.NewBatchProcessor(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := processor.Start(ctx); err != nil {
		log.Fatalf("Failed to start batch processor: %v", err)
	}
	defer processor.Stop()

	fmt.Printf("错误处理配置 - 最大重试: %d, 重试延迟: %v\n", 
		config.Export.MaxRetries, config.Export.RetryDelay)

	// 添加一些数据，其中一些会导出失败
	for i := 0; i < 25; i++ {
		processor.AddMetric(
			fmt.Sprintf("error_test_metric_%d", i),
			float64(i),
			map[string]string{
				"test":  "error_handling",
				"batch": fmt.Sprintf("%d", i/10),
			},
		)
	}

	// 等待处理和重试完成
	time.Sleep(5 * time.Second)
	processor.ForceFlush()
	time.Sleep(2 * time.Second)

	stats := processor.GetStats()
	fmt.Printf("错误处理测试结果:\n")
	fmt.Printf("  总项目数: %d\n", stats.TotalItems)
	fmt.Printf("  已处理项目数: %d\n", stats.ProcessedItems)
	fmt.Printf("  丢弃项目数: %d\n", stats.DroppedItems)
	fmt.Printf("  导出批次数: %d\n", stats.ExportedBatches)
	fmt.Printf("  失败导出数: %d\n", stats.FailedExports)
	fmt.Printf("  成功率: %.2f%%\n", 
		float64(stats.ExportedBatches-stats.FailedExports)/float64(stats.ExportedBatches)*100)
}