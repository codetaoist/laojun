package errors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestInMemoryErrorReporter(t *testing.T) {
	logger := zap.NewNop()
	reporter := NewInMemoryErrorReporter(logger, 10)

	t.Run("报告错误", func(t *testing.T) {
		err := NewError(ErrorTypeNetwork, "NET_001", "Network error").
			WithComponent("TestComponent").
			Build()

		report := &ErrorReport{
			ID:        generateErrorID(),
			Error:     err,
			Timestamp: time.Now(),
			Context:   map[string]interface{}{"test": "value"},
		}

		reportErr := reporter.ReportError(context.Background(), report)
		if reportErr != nil {
			t.Errorf("Expected no error when reporting, got %v", reportErr)
		}
	})

	t.Run("获取错误统计", func(t *testing.T) {
		// 清空之前的报告
		reporter = NewInMemoryErrorReporter(logger, 10)

		// 报告几个错误
		errors := []*MonitoringError{
			NewError(ErrorTypeNetwork, "NET_001", "Network error 1").Build(),
			NewError(ErrorTypeNetwork, "NET_002", "Network error 2").Build(),
			NewError(ErrorTypeSystem, "SYS_001", "System error").Build(),
		}

		for _, err := range errors {
			report := &ErrorReport{
				ID:        generateErrorID(),
				Error:     err,
				Timestamp: time.Now(),
			}
			reporter.ReportError(context.Background(), report)
		}

		stats, err := reporter.GetStatistics(context.Background(), time.Hour)
		if err != nil {
			t.Fatalf("Failed to get statistics: %v", err)
		}
		if stats.TotalErrors != 3 {
			t.Errorf("Expected total errors 3, got %d", stats.TotalErrors)
		}
		if stats.ErrorsByType[ErrorTypeNetwork] != 2 {
			t.Errorf("Expected 2 network errors, got %d", stats.ErrorsByType[ErrorTypeNetwork])
		}
		if stats.ErrorsByType[ErrorTypeSystem] != 1 {
			t.Errorf("Expected 1 system error, got %d", stats.ErrorsByType[ErrorTypeSystem])
		}
	})

	t.Run("获取最近错误", func(t *testing.T) {
		reporter = NewInMemoryErrorReporter(logger, 10)

		authErr := NewError(ErrorTypeAuth, "AUTH_001", "Auth error").Build()
		report := &ErrorReport{
			ID:        generateErrorID(),
			Error:     authErr,
			Timestamp: time.Now(),
		}
		reporter.ReportError(context.Background(), report)

		recentErrors, err := reporter.GetReports(context.Background(), time.Hour)
		if err != nil {
			t.Fatalf("Failed to get reports: %v", err)
		}
		if len(recentErrors) != 1 {
			t.Errorf("Expected 1 recent error, got %d", len(recentErrors))
		}
		if recentErrors[0].Error.Code != "AUTH_001" {
			t.Errorf("Expected error code 'AUTH_001', got '%s'", recentErrors[0].Error.Code)
		}
	})

	t.Run("限制最近错误数量", func(t *testing.T) {
		reporter = NewInMemoryErrorReporter(logger, 3)

		// 报告超过限制的错误数量
		for i := 0; i < 15; i++ {
			err := NewError(ErrorTypeSystem, fmt.Sprintf("SYS_%03d", i), "System error").Build()
			report := &ErrorReport{
				ID:        generateErrorID(),
				Error:     err,
				Timestamp: time.Now(),
			}
			reporter.ReportError(context.Background(), report)
		}

		recentErrors, err := reporter.GetReports(context.Background(), time.Hour)
		if err != nil {
			t.Fatalf("Failed to get reports: %v", err)
		}
		// 内存报告器默认最多保存10个错误
		if len(recentErrors) > 10 {
			t.Errorf("Expected at most 10 recent errors, got %d", len(recentErrors))
		}
	})
}

func TestFileErrorReporter(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "error_reporter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "errors.log")
	logger := zap.NewNop()
	reporter := NewFileErrorReporter(logFile, logger)

	t.Run("报告错误到文件", func(t *testing.T) {
		err := NewError(ErrorTypeConfig, "CONFIG_001", "Config error").
			WithComponent("ConfigLoader").
			Build()

		report := &ErrorReport{
			ID:        generateErrorID(),
			Error:     err,
			Timestamp: time.Now(),
			Context:   map[string]interface{}{"file": "config.yaml"},
		}

		reportErr := reporter.ReportError(context.Background(), report)
		if reportErr != nil {
			t.Errorf("Expected no error when reporting to file, got %v", reportErr)
		}

		// 验证文件是否存在
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("Expected log file to be created")
		}
	})

	t.Run("文件权限错误", func(t *testing.T) {
		// 尝试写入不存在的目录
		invalidPath := filepath.Join(tempDir, "nonexistent", "errors.log")
		logger := zap.NewNop()
		invalidReporter := NewFileErrorReporter(invalidPath, logger)

		err := NewError(ErrorTypeSystem, "SYS_001", "System error").Build()
		report := &ErrorReport{
			ID:        generateErrorID(),
			Error:     err,
			Timestamp: time.Now(),
		}

		reportErr := invalidReporter.ReportError(context.Background(), report)
		if reportErr == nil {
			t.Error("Expected error when writing to invalid path")
		}
	})
}

func TestErrorMonitor(t *testing.T) {
	logger := zap.NewNop()
	reporter := NewInMemoryErrorReporter(logger, 10)
	recoveryManager := NewRecoveryManager(logger)
	
	monitor := NewErrorMonitor(reporter, recoveryManager, logger)

	t.Run("设置错误阈值", func(t *testing.T) {
		threshold := ErrorThreshold{
			MaxErrorsPerMinute: 5,
			MaxErrorsPerHour:   50,
			AlertSeverity:      SeverityHigh,
		}

		monitor.SetThreshold(ErrorTypeNetwork, threshold)

		// 验证阈值设置（通过触发错误来间接验证）
		err := NewError(ErrorTypeNetwork, "NET_001", "Network error").
			WithRecoverable(true).
			Build()
		
		ctx := context.Background()
		handleErr := monitor.HandleError(ctx, err)
		if handleErr != nil {
			t.Errorf("Expected no error when handling, got %v", handleErr)
		}
	})

	t.Run("处理错误", func(t *testing.T) {
		sysErr := NewError(ErrorTypeSystem, "SYS_001", "System error").
			WithComponent("TestComponent").
			WithRecoverable(true).
			Build()

		ctx := context.Background()
		handleErr := monitor.HandleError(ctx, sysErr)
		if handleErr != nil {
			t.Errorf("Expected no error when handling, got %v", handleErr)
		}

		// 验证错误已被报告
		stats, err := reporter.GetStatistics(context.Background(), time.Hour)
		if err != nil {
			t.Fatalf("Failed to get statistics: %v", err)
		}
		if stats.TotalErrors == 0 {
			t.Error("Expected error to be reported")
		}
	})

	t.Run("启动和停止监控", func(t *testing.T) {
		// 启动监控
		startErr := monitor.Start()
		if startErr != nil {
			t.Errorf("Expected no error when starting monitor, got %v", startErr)
		}

		// 停止监控
		stopErr := monitor.Stop()
		if stopErr != nil {
			t.Errorf("Expected no error when stopping monitor, got %v", stopErr)
		}
	})

	t.Run("错误阈值检查", func(t *testing.T) {
		// 设置较低的阈值进行测试
		threshold := ErrorThreshold{
			MaxErrorsPerMinute: 2,
			MaxErrorsPerHour:   10,
			AlertSeverity:      SeverityMedium,
		}
		monitor.SetThreshold(ErrorTypeValidation, threshold)

		// 快速报告多个错误以触发阈值
		ctx := context.Background()
		for i := 0; i < 3; i++ {
			err := NewError(ErrorTypeValidation, fmt.Sprintf("VAL_%03d", i), "Validation error").Build()
			monitor.HandleError(ctx, err)
		}

		// 验证统计信息
		stats, err := reporter.GetStatistics(context.Background(), time.Hour)
		if err != nil {
			t.Fatalf("Failed to get statistics: %v", err)
		}
		if stats.ErrorsByType[ErrorTypeValidation] != 3 {
			t.Errorf("Expected 3 validation errors, got %d", stats.ErrorsByType[ErrorTypeValidation])
		}
	})
}

func TestErrorThreshold(t *testing.T) {
	t.Run("阈值结构", func(t *testing.T) {
		threshold := ErrorThreshold{
			MaxErrorsPerMinute: 10,
			MaxErrorsPerHour:   100,
			AlertSeverity:      SeverityHigh,
		}

		if threshold.MaxErrorsPerMinute != 10 {
			t.Errorf("Expected MaxErrorsPerMinute 10, got %d", threshold.MaxErrorsPerMinute)
		}
		if threshold.MaxErrorsPerHour != 100 {
			t.Errorf("Expected MaxErrorsPerHour 100, got %d", threshold.MaxErrorsPerHour)
		}
		if threshold.AlertSeverity != SeverityHigh {
			t.Errorf("Expected AlertSeverity %s, got %s", SeverityHigh, threshold.AlertSeverity)
		}
	})
}

func TestErrorReport(t *testing.T) {
	t.Run("错误报告结构", func(t *testing.T) {
		err := NewError(ErrorTypeAuth, "AUTH_001", "Auth error").Build()
		timestamp := time.Now()
		context := map[string]interface{}{
			"user_id": "12345",
			"action":  "login",
		}

		report := &ErrorReport{
			ID:        generateErrorID(),
			Error:     err,
			Timestamp: timestamp,
			Context:   context,
		}

		if report.Error != err {
			t.Error("Expected error to match")
		}
		if report.Timestamp != timestamp {
			t.Error("Expected timestamp to match")
		}
		if report.Context["user_id"] != "12345" {
			t.Errorf("Expected user_id '12345', got %v", report.Context["user_id"])
		}
		if report.Context["action"] != "login" {
			t.Errorf("Expected action 'login', got %v", report.Context["action"])
		}
	})
}

func TestErrorStatistics(t *testing.T) {
	t.Run("统计信息结构", func(t *testing.T) {
		stats := &ErrorStatistics{
			TotalErrors:      100,
			ErrorsByType:     make(map[ErrorType]int64),
			ErrorsBySeverity: make(map[ErrorSeverity]int64),
			ErrorsByComponent: make(map[string]int64),
			ErrorsByCode:     make(map[string]int64),
			LastUpdated:      time.Now(),
			TimeWindow:       time.Hour,
		}

		stats.ErrorsByType[ErrorTypeNetwork] = 50
		stats.ErrorsByType[ErrorTypeSystem] = 30
		stats.ErrorsBySeverity[SeverityHigh] = 20
		stats.ErrorsBySeverity[SeverityMedium] = 80

		if stats.TotalErrors != 100 {
			t.Errorf("Expected TotalErrors 100, got %d", stats.TotalErrors)
		}
		if stats.ErrorsByType[ErrorTypeNetwork] != 50 {
			t.Errorf("Expected 50 network errors, got %d", stats.ErrorsByType[ErrorTypeNetwork])
		}
		if stats.ErrorsBySeverity[SeverityHigh] != 20 {
			t.Errorf("Expected 20 high severity errors, got %d", stats.ErrorsBySeverity[SeverityHigh])
		}
	})
}