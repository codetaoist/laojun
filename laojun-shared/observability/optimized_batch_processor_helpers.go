package observability

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"runtime"
	"sync/atomic"
	"time"
)

// generateCacheKey 生成缓存键
func (bp *OptimizedBatchProcessor) generateCacheKey(items []BatchItem) string {
	hasher := sha256.New()
	
	// 对items进行哈希
	for _, item := range items {
		data, _ := json.Marshal(item)
		hasher.Write(data)
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// getCachedData 获取缓存数据
func (bp *OptimizedBatchProcessor) getCachedData(key string) []byte {
	bp.metricsCache.mu.RLock()
	defer bp.metricsCache.mu.RUnlock()
	
	cached, exists := bp.metricsCache.cache[key]
	if !exists {
		bp.metricsCache.misses++
		return nil
	}
	
	// 检查是否过期
	if time.Since(cached.Timestamp) > bp.metricsCache.ttl {
		delete(bp.metricsCache.cache, key)
		bp.metricsCache.misses++
		return nil
	}
	
	// 更新访问计数
	cached.AccessCount++
	bp.metricsCache.hits++
	
	return cached.Data
}

// setCachedData 设置缓存数据
func (bp *OptimizedBatchProcessor) setCachedData(key string, data []byte) {
	bp.metricsCache.mu.Lock()
	defer bp.metricsCache.mu.Unlock()
	
	// 检查缓存大小限制
	if len(bp.metricsCache.cache) >= bp.metricsCache.maxSize {
		bp.evictLeastUsedCache()
	}
	
	// 计算数据哈希
	hasher := fnv.New64a()
	hasher.Write(data)
	hash := hasher.Sum64()
	
	bp.metricsCache.cache[key] = &CachedMetric{
		Data:        data,
		Hash:        hash,
		Timestamp:   time.Now(),
		AccessCount: 1,
	}
}

// evictLeastUsedCache 驱逐最少使用的缓存
func (bp *OptimizedBatchProcessor) evictLeastUsedCache() {
	var oldestKey string
	var oldestTime time.Time
	var leastUsedKey string
	var leastUsedCount int64 = -1
	
	for key, cached := range bp.metricsCache.cache {
		// 找到最旧的缓存
		if oldestTime.IsZero() || cached.Timestamp.Before(oldestTime) {
			oldestTime = cached.Timestamp
			oldestKey = key
		}
		
		// 找到最少使用的缓存
		if leastUsedCount == -1 || cached.AccessCount < leastUsedCount {
			leastUsedCount = cached.AccessCount
			leastUsedKey = key
		}
	}
	
	// 优先删除最少使用的，如果访问次数相同则删除最旧的
	keyToDelete := leastUsedKey
	if leastUsedCount > 10 { // 如果最少使用的也被访问了很多次，则删除最旧的
		keyToDelete = oldestKey
	}
	
	delete(bp.metricsCache.cache, keyToDelete)
}

// cleanupExpiredCache 清理过期缓存
func (bp *OptimizedBatchProcessor) cleanupExpiredCache() {
	bp.metricsCache.mu.Lock()
	defer bp.metricsCache.mu.Unlock()
	
	now := time.Now()
	for key, cached := range bp.metricsCache.cache {
		if now.Sub(cached.Timestamp) > bp.metricsCache.ttl {
			delete(bp.metricsCache.cache, key)
		}
	}
}

// exportCachedToExporter 导出缓存数据到导出器
func (bp *OptimizedBatchProcessor) exportCachedToExporter(ctx context.Context, exporter OptimizedExporter, cachedData []byte) {
	retryCount := 0
	maxRetries := bp.exportConfig.MaxRetries
	
	for retryCount <= maxRetries {
		err := exporter.ExportCached(ctx, cachedData)
		if err == nil {
			atomic.AddInt64(&bp.stats.ProcessedItems, int64(len(cachedData)))
			return
		}
		
		if retryCount < maxRetries {
			retryCount++
			backoffDuration := bp.calculateBackoff(retryCount)
			
			select {
			case <-ctx.Done():
				atomic.AddInt64(&bp.stats.FailedExports, 1)
				return
			case <-time.After(backoffDuration):
				continue
			}
		}
	}
	
	atomic.AddInt64(&bp.stats.FailedExports, 1)
}

// exportToOptimizedExporter 导出到优化导出器
func (bp *OptimizedBatchProcessor) exportToOptimizedExporter(ctx context.Context, exporter OptimizedExporter, items []BatchItem, cacheKey string) {
	retryCount := 0
	maxRetries := bp.exportConfig.MaxRetries
	
	for retryCount <= maxRetries {
		err := exporter.Export(ctx, items)
		if err == nil {
			atomic.AddInt64(&bp.stats.ProcessedItems, int64(len(items)))
			
			// 缓存成功导出的数据
			if data, marshalErr := json.Marshal(items); marshalErr == nil {
				bp.setCachedData(cacheKey, data)
			}
			return
		}
		
		if retryCount < maxRetries {
			retryCount++
			backoffDuration := bp.calculateBackoff(retryCount)
			
			select {
			case <-ctx.Done():
				atomic.AddInt64(&bp.stats.FailedExports, 1)
				return
			case <-time.After(backoffDuration):
				continue
			}
		}
	}
	
	atomic.AddInt64(&bp.stats.FailedExports, 1)
}

// calculateBackoff 计算退避时间
func (bp *OptimizedBatchProcessor) calculateBackoff(retryCount int) time.Duration {
	baseDelay := bp.exportConfig.RetryDelay
	maxDelay := 30 * time.Second // 最大退避时间
	
	// 指数退避
	delay := baseDelay
	for i := 1; i < retryCount; i++ {
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
			break
		}
	}
	
	return delay
}

// updatePerformanceStats 更新性能统计
func (bp *OptimizedBatchProcessor) updatePerformanceStats() {
	bp.perfMonitor.mu.Lock()
	defer bp.perfMonitor.mu.Unlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	bp.perfMonitor.memoryUsage = int64(m.Alloc)
	bp.perfMonitor.gcCount = int64(m.NumGC)
	bp.perfMonitor.goroutineCount = runtime.NumGoroutine()
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime: time.Now(),
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.startTime = time.Now()
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// 可以在这里添加清理逻辑
}

// GetStats 获取性能统计
func (pm *PerformanceMonitor) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return map[string]interface{}{
		"uptime":           time.Since(pm.startTime),
		"memory_usage":     pm.memoryUsage,
		"gc_count":         pm.gcCount,
		"goroutine_count":  pm.goroutineCount,
		"cpu_usage":        pm.cpuUsage,
	}
}

// IsRunning 检查批量处理器是否正在运行
func (bp *OptimizedBatchProcessor) IsRunning() bool {
	return atomic.LoadInt32(&bp.running) == 1
}

// GetCacheStats 获取缓存统计信息
func (bp *OptimizedBatchProcessor) GetCacheStats() map[string]interface{} {
	bp.metricsCache.mu.RLock()
	defer bp.metricsCache.mu.RUnlock()
	
	total := bp.metricsCache.hits + bp.metricsCache.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(bp.metricsCache.hits) / float64(total)
	}
	
	return map[string]interface{}{
		"cache_size":    len(bp.metricsCache.cache),
		"max_size":      bp.metricsCache.maxSize,
		"hits":          bp.metricsCache.hits,
		"misses":        bp.metricsCache.misses,
		"hit_rate":      hitRate,
		"ttl_seconds":   bp.metricsCache.ttl.Seconds(),
	}
}

// GetConnectionPoolStats 获取所有导出器的连接池统计
func (bp *OptimizedBatchProcessor) GetConnectionPoolStats() map[string]ConnectionPoolStats {
	stats := make(map[string]ConnectionPoolStats)
	
	for name, exporter := range bp.exporters {
		stats[name] = exporter.GetConnectionPoolStats()
	}
	
	return stats
}

// FlushNow 立即刷新缓冲区
func (bp *OptimizedBatchProcessor) FlushNow(ctx context.Context) error {
	if !bp.IsRunning() {
		return fmt.Errorf("optimized batch processor is not running")
	}
	
	bp.flushBufferOptimized(ctx)
	return nil
}

// GetBufferUtilization 获取缓冲区利用率
func (bp *OptimizedBatchProcessor) GetBufferUtilization() float64 {
	bp.bufferMu.RLock()
	defer bp.bufferMu.RUnlock()
	
	if bp.config.BufferSize == 0 {
		return 0.0
	}
	
	return float64(len(bp.buffer)) / float64(bp.config.BufferSize)
}

// ResetStats 重置统计信息
func (bp *OptimizedBatchProcessor) ResetStats() {
	atomic.StoreInt64(&bp.stats.TotalItems, 0)
	atomic.StoreInt64(&bp.stats.ProcessedItems, 0)
	atomic.StoreInt64(&bp.stats.DroppedItems, 0)
	atomic.StoreInt64(&bp.stats.ExportedBatches, 0)
	atomic.StoreInt64(&bp.stats.FailedExports, 0)
	
	bp.stats.LastFlushTime = time.Time{}
	bp.stats.LastExportTime = time.Time{}
	bp.stats.AverageFlushTime = 0
	
	// 重置缓存统计
	bp.metricsCache.mu.Lock()
	bp.metricsCache.hits = 0
	bp.metricsCache.misses = 0
	bp.metricsCache.mu.Unlock()
}

// ClearCache 清空缓存
func (bp *OptimizedBatchProcessor) ClearCache() {
	bp.metricsCache.mu.Lock()
	defer bp.metricsCache.mu.Unlock()
	
	bp.metricsCache.cache = make(map[string]*CachedMetric)
	bp.metricsCache.hits = 0
	bp.metricsCache.misses = 0
}