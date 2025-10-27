package test

import (
	"testing"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/config"
	"github.com/codetaoist/laojun-monitoring/internal/exporters"
	"github.com/codetaoist/laojun-monitoring/internal/services"
	"go.uber.org/zap"
)

func TestBatchProcessorIntegration(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		BatchProcessor: config.BatchProcessorConfig{
			Enabled:       true,
			Name:          "test-batch-processor",
			BatchSize:     10,
			FlushInterval: 2 * time.Second,
			BufferSize:    100,
			MaxRetries:    3,
			RetryDelay:    1 * time.Second,
			Exporters: []config.BatchExporterConfig{
				{
					Type:    "console",
					Name:    "test-console",
					Enabled: true,
					Config: map[string]interface{}{
						"format":    "json",
						"timestamp": true,
					},
				},
			},
		},
		Query: config.QueryConfig{
			Type: "memory",
		},
		Exporters: config.ExportersConfig{
			Prometheus: config.PrometheusExporterConfig{
				Enabled: false,
			},
		},
		Alerting: config.AlertingConfig{
			Enabled: false,
		},
	}

	// 创建日志器
	logger, _ := zap.NewDevelopment()

	// 创建监控服务
	service, err := services.NewMonitoringService(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create monitoring service: %v", err)
	}

	// 启动服务

	if err := service.Start(); err != nil {
		t.Fatalf("Failed to start monitoring service: %v", err)
	}
	defer service.Stop()

	// 验证批量处理器已启动
	batchProcessor := service.GetBatchProcessor()
	if batchProcessor == nil {
		t.Fatal("Batch processor is nil")
	}

	if !batchProcessor.IsRunning() {
		t.Fatal("Batch processor is not running")
	}

	if !batchProcessor.IsHealthy() {
		t.Fatal("Batch processor is not healthy")
	}

	// 测试数据处理
	testData := []exporters.MetricData{
		{
			Name:      "test_metric_1",
			Value:     100.0,
			Timestamp: time.Now(),
			Labels: map[string]string{
				"service": "test",
				"env":     "development",
			},
		},
		{
			Name:      "test_metric_2",
			Value:     200.0,
			Timestamp: time.Now(),
			Labels: map[string]string{
				"service": "test",
				"env":     "development",
			},
		},
	}

	// 发送测试数据
	for _, data := range testData {
		if err := batchProcessor.Export(data); err != nil {
			t.Errorf("Failed to export data: %v", err)
		}
	}

	// 等待数据处理
	time.Sleep(3 * time.Second)

	// 检查统计信息
	stats := batchProcessor.GetStats()
	if stats.ProcessedItems == 0 {
		t.Error("No items were processed")
	}

	t.Logf("Batch processor stats: %+v", stats)

	// 检查健康状态
	health := service.Health()
	if health["status"] != "healthy" {
		t.Errorf("Service is not healthy: %v", health)
	}

	batchProcessorHealth, exists := health["batch_processor"]
	if !exists {
		t.Error("Batch processor health not found in service health")
	}

	t.Logf("Service health: %+v", health)
	t.Logf("Batch processor health: %+v", batchProcessorHealth)
}

func TestBatchProcessorConfiguration(t *testing.T) {
	// 测试不同配置的批量处理器
	testCases := []struct {
		name   string
		config config.BatchProcessorConfig
	}{
		{
			name: "Small batch size",
			config: config.BatchProcessorConfig{
				Enabled:       true,
				Name:          "small-batch",
				BatchSize:     5,
				FlushInterval: 1 * time.Second,
				BufferSize:    50,
				MaxRetries:    2,
				RetryDelay:    500 * time.Millisecond,
				Exporters: []config.BatchExporterConfig{
					{
						Type:    "console",
						Name:    "console-small",
						Enabled: true,
						Config: map[string]interface{}{
							"format": "text",
						},
					},
				},
			},
		},
		{
			name: "Large batch size",
			config: config.BatchProcessorConfig{
				Enabled:       true,
				Name:          "large-batch",
				BatchSize:     100,
				FlushInterval: 5 * time.Second,
				BufferSize:    1000,
				MaxRetries:    5,
				RetryDelay:    2 * time.Second,
				Exporters: []config.BatchExporterConfig{
					{
						Type:    "console",
						Name:    "console-large",
						Enabled: true,
						Config: map[string]interface{}{
							"format":    "json",
							"timestamp": true,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				BatchProcessor: tc.config,
				Query: config.QueryConfig{
					Type: "memory",
				},
				Exporters: config.ExportersConfig{
					Prometheus: config.PrometheusExporterConfig{
						Enabled: false,
					},
				},
				Alerting: config.AlertingConfig{
					Enabled: false,
				},
			}

			logger, _ := zap.NewDevelopment()
			service, err := services.NewMonitoringService(cfg, logger)
			if err != nil {
				t.Fatalf("Failed to create monitoring service: %v", err)
			}

			if err := service.Start(); err != nil {
				t.Fatalf("Failed to start monitoring service: %v", err)
			}
			defer service.Stop()

			batchProcessor := service.GetBatchProcessor()
			if batchProcessor == nil {
				t.Fatal("Batch processor is nil")
			}

			if !batchProcessor.IsRunning() {
				t.Fatal("Batch processor is not running")
			}

			// 验证配置
			if batchProcessor.GetBatchSize() != tc.config.BatchSize {
				t.Errorf("Expected batch size %d, got %d", tc.config.BatchSize, batchProcessor.GetBatchSize())
			}

			if batchProcessor.GetFlushInterval() != tc.config.FlushInterval {
				t.Errorf("Expected flush interval %v, got %v", tc.config.FlushInterval, batchProcessor.GetFlushInterval())
			}

			t.Logf("Configuration test passed for %s", tc.name)
		})
	}
}