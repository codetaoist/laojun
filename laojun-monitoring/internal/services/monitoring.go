package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/config"
	"github.com/codetaoist/laojun-monitoring/internal/collectors"
	"github.com/codetaoist/laojun-monitoring/internal/exporters"
	"github.com/codetaoist/laojun-monitoring/internal/alerting"
	"github.com/codetaoist/laojun-monitoring/internal/query"
	"github.com/codetaoist/laojun-monitoring/internal/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

// MonitoringService 监控服务
type MonitoringService struct {
	config     *config.Config
	logger     *zap.Logger
	collectors []collectors.Collector
	exporters  []exporters.Exporter
	alertManager *alerting.AlertManager
	registry   *prometheus.Registry
	batchProcessor *exporters.BatchProcessorExporter
	querier    query.MetricQuerier
	
	// 错误处理
	errorMonitor    *errors.ErrorMonitor
	recoveryManager *errors.RecoveryManager
	errorReporter   errors.ErrorReporter
	
	// 服务状态
	running    bool
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	
	// 统计信息
	stats      *ServiceStats
}

// ServiceStats 服务统计信息
type ServiceStats struct {
	StartTime        time.Time `json:"start_time"`
	MetricsCollected int64     `json:"metrics_collected"`
	AlertsTriggered  int64     `json:"alerts_triggered"`
	ExportsCompleted int64     `json:"exports_completed"`
	Errors           int64     `json:"errors"`
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(cfg *config.Config, logger *zap.Logger) (*MonitoringService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建 Prometheus 注册表
	registry := prometheus.NewRegistry()
	
	service := &MonitoringService{
		config:   cfg,
		logger:   logger,
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,
		stats: &ServiceStats{
			StartTime: time.Now(),
		},
	}
	
	// 初始化收集器
	if err := service.initCollectors(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init collectors: %w", err)
	}
	
	// 初始化导出器
	if err := service.initExporters(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init exporters: %w", err)
	}
	
	// 初始化批量处理器
	if err := service.initBatchProcessor(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize batch processor: %w", err)
	}
	
	// 初始化查询器
	if err := service.initQuerier(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init querier: %w", err)
	}
	
	// 初始化告警管理器
	if err := service.initAlertManager(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init alert manager: %w", err)
	}
	
	// 初始化错误处理组件
	if err := service.initErrorHandling(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init error handling: %w", err)
	}
	
	return service, nil
}

// Start 启动监控服务
func (s *MonitoringService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return fmt.Errorf("monitoring service is already running")
	}
	
	s.logger.Info("Starting monitoring service")
	
	// 启动收集器
	for _, collector := range s.collectors {
		if err := collector.Start(); err != nil {
			s.logger.Error("Failed to start collector",
				zap.String("collector", collector.Name()),
				zap.Error(err))
			return fmt.Errorf("failed to start collector %s: %w", collector.Name(), err)
		}
		s.logger.Info("Started collector", zap.String("name", collector.Name()))
	}
	
	// 启动导出器
	for _, exporter := range s.exporters {
		if err := exporter.Start(s.ctx); err != nil {
			s.logger.Error("Failed to start exporter",
				zap.String("exporter", exporter.Name()),
				zap.Error(err))
			return fmt.Errorf("failed to start exporter %s: %w", exporter.Name(), err)
		}
		s.logger.Info("Started exporter", zap.String("name", exporter.Name()))
	}
	
	// 启动批量处理器
	if s.batchProcessor != nil {
		if err := s.batchProcessor.Start(s.ctx); err != nil {
			s.logger.Error("Failed to start batch processor", zap.Error(err))
			return fmt.Errorf("failed to start batch processor: %w", err)
		}
		s.logger.Info("Started batch processor")
	}
	
	// 启动告警管理器
	if s.alertManager != nil {
		if err := s.alertManager.Start(); err != nil {
			s.logger.Error("Failed to start alert manager", zap.Error(err))
			return fmt.Errorf("failed to start alert manager: %w", err)
		}
		s.logger.Info("Started alert manager")
	}
	
	// 启动错误监控器
	if s.errorMonitor != nil {
		if err := s.errorMonitor.Start(); err != nil {
			s.logger.Error("Failed to start error monitor", zap.Error(err))
			return fmt.Errorf("failed to start error monitor: %w", err)
		}
		s.logger.Info("Started error monitor")
	}
	
	// 启动指标收集和存储到查询器的goroutine
	go s.collectAndStoreMetrics()
	
	s.running = true
	s.logger.Info("Monitoring service started successfully")
	
	return nil
}

// Stop 停止监控服务
func (s *MonitoringService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return nil
	}
	
	s.logger.Info("Stopping monitoring service")
	
	// 取消上下文
	s.cancel()
	
	// 停止错误监控器
	if s.errorMonitor != nil {
		if err := s.errorMonitor.Close(); err != nil {
			s.logger.Error("Failed to stop error monitor", zap.Error(err))
		}
	}
	
	// 停止告警管理器
	if s.alertManager != nil {
		if err := s.alertManager.Stop(); err != nil {
			s.logger.Error("Failed to stop alert manager", zap.Error(err))
		}
	}
	
	// 停止批量处理器
	if s.batchProcessor != nil {
		if err := s.batchProcessor.Stop(); err != nil {
			s.logger.Error("Failed to stop batch processor", zap.Error(err))
		}
	}
	
	// 停止导出器
	for _, exporter := range s.exporters {
		if err := exporter.Stop(); err != nil {
			s.logger.Error("Failed to stop exporter",
				zap.String("exporter", exporter.Name()),
				zap.Error(err))
		}
	}
	
	// 停止收集器
	for _, collector := range s.collectors {
		if err := collector.Stop(); err != nil {
			s.logger.Error("Failed to stop collector",
				zap.String("collector", collector.Name()),
				zap.Error(err))
		}
	}
	
	s.running = false
	s.logger.Info("Monitoring service stopped")
	
	return nil
}

// IsRunning 检查服务是否运行中
func (s *MonitoringService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetStats 获取服务统计信息
func (s *MonitoringService) GetStats() *ServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 复制统计信息
	stats := *s.stats
	return &stats
}

// GetRegistry 获取 Prometheus 注册表
func (s *MonitoringService) GetRegistry() *prometheus.Registry {
	return s.registry
}

// GetCollectors 获取收集器列表
func (s *MonitoringService) GetCollectors() []collectors.Collector {
	return s.collectors
}

// GetExporters 获取导出器列表
func (s *MonitoringService) GetExporters() []exporters.Exporter {
	return s.exporters
}

// GetAlertManager 获取告警管理器
func (s *MonitoringService) GetAlertManager() *alerting.AlertManager {
	return s.alertManager
}

// GetBatchProcessor 获取批量处理器
func (s *MonitoringService) GetBatchProcessor() *exporters.BatchProcessorExporter {
	return s.batchProcessor
}

func (s *MonitoringService) GetQuerier() query.MetricQuerier {
	return s.querier
}

// Health 检查服务健康状态
func (s *MonitoringService) Health() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	health := map[string]interface{}{
		"status":  "healthy",
		"running": s.running,
		"uptime":  time.Since(s.stats.StartTime).String(),
		"stats":   s.stats,
	}
	
	// 检查收集器状态
	collectorsHealth := make(map[string]string)
	for _, collector := range s.collectors {
		if collector.IsHealthy() {
			collectorsHealth[collector.Name()] = "healthy"
		} else {
			collectorsHealth[collector.Name()] = "unhealthy"
			health["status"] = "degraded"
		}
	}
	health["collectors"] = collectorsHealth
	
	// 检查导出器状态
	exportersHealth := make(map[string]string)
	for _, exporter := range s.exporters {
		if exporter.IsHealthy() {
			exportersHealth[exporter.Name()] = "healthy"
		} else {
			exportersHealth[exporter.Name()] = "unhealthy"
			health["status"] = "degraded"
		}
	}
	health["exporters"] = exportersHealth
	
	// 检查批量处理器状态
	if s.batchProcessor != nil {
		if s.batchProcessor.IsHealthy() {
			health["batch_processor"] = "healthy"
		} else {
			health["batch_processor"] = "unhealthy"
			health["status"] = "degraded"
		}
	}
	
	// 检查告警管理器状态
	if s.alertManager != nil {
		if s.alertManager.IsHealthy() {
			health["alert_manager"] = "healthy"
		} else {
			health["alert_manager"] = "unhealthy"
			health["status"] = "degraded"
		}
	}
	
	return health
}

// Ready 检查服务就绪状态
func (s *MonitoringService) Ready() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if !s.running {
		return false
	}
	
	// 检查所有组件是否就绪
	for _, collector := range s.collectors {
		if !collector.IsReady() {
			return false
		}
	}
	
	for _, exporter := range s.exporters {
		if !exporter.IsReady() {
			return false
		}
	}
	
	if s.alertManager != nil && !s.alertManager.IsReady() {
		return false
	}
	
	return true
}

// initCollectors 初始化收集器
func (s *MonitoringService) initCollectors() error {
	s.collectors = make([]collectors.Collector, 0)
	
	// 创建logrus logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	
	// 系统收集器
	if s.config.Collectors.System.Enabled {
		systemCollector := collectors.NewSystemCollector(
			s.config.Collectors.System.Interval,
			logrusLogger,
		)
		s.collectors = append(s.collectors, systemCollector)
	}
	
	// 应用收集器
	if s.config.Collectors.Application.Enabled {
		appCollector := collectors.NewApplicationCollector(
			s.config.Collectors.Application.Interval,
			logrusLogger,
		)
		s.collectors = append(s.collectors, appCollector)
	}
	
	// 网络收集器
	if s.config.Collectors.Network.Enabled {
		networkCollector := collectors.NewNetworkCollector(
			s.config.Collectors.Network.Interval,
			logrusLogger,
		)
		s.collectors = append(s.collectors, networkCollector)
	}
	
	return nil
}

// initExporters 初始化导出器
func (s *MonitoringService) initExporters() error {
	s.exporters = make([]exporters.Exporter, 0)
	
	// 创建logrus logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	
	// Prometheus 导出器
	if s.config.Exporters.Prometheus.Enabled {
		// 转换配置类型
		prometheusConfig := exporters.PrometheusConfig{
			Enabled: s.config.Exporters.Prometheus.Enabled,
			Host:    "0.0.0.0", // 使用默认host
			Port:    s.config.Exporters.Prometheus.Port,
			Path:    s.config.Exporters.Prometheus.Path,
		}
		
		prometheusExporter := exporters.NewPrometheusExporter(
			prometheusConfig,
			logrusLogger,
		)
		s.exporters = append(s.exporters, prometheusExporter)
	}
	
	return nil
}

// initBatchProcessor 初始化批量处理器
func (s *MonitoringService) initBatchProcessor() error {
	if !s.config.BatchProcessor.Enabled {
		return nil
	}
	
	processor, err := exporters.NewBatchProcessorExporter(
		&s.config.BatchProcessor,
		s.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create batch processor: %w", err)
	}
	
	s.batchProcessor = processor
	return nil
}

// initQuerier 初始化查询器
func (s *MonitoringService) initQuerier() error {
	s.logger.Info("Initializing metric querier")
	
	// 创建logrus logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	
	// 创建查询器工厂
	factory := query.NewQuerierFactory(logrusLogger)
	
	// 验证查询配置
	if err := query.ValidateQueryConfig(s.config.Query); err != nil {
		s.logger.Error("Invalid query configuration", zap.Error(err))
		return fmt.Errorf("invalid query configuration: %w", err)
	}
	
	// 根据配置创建查询器
	querier, err := factory.CreateQuerier(s.config.Query)
	if err != nil {
		s.logger.Error("Failed to create querier", zap.Error(err))
		return fmt.Errorf("failed to create querier: %w", err)
	}
	
	s.querier = querier
	s.logger.Info("Metric querier initialized successfully", zap.String("type", s.config.Query.Type))
	
	return nil
}

// initAlertManager 初始化告警管理器
func (s *MonitoringService) initAlertManager() error {
	if !s.config.Alerting.Enabled {
		return nil
	}

	// 创建logrus logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)

	// 转换配置类型
	alertingConfig := s.convertAlertingConfig(s.config.Alerting)

	// 创建告警管理器
	s.alertManager = alerting.NewAlertManager(alertingConfig, s.querier, logrusLogger)

	return nil
}

// convertAlertingConfig 转换告警配置类型
// collectAndStoreMetrics 定期收集指标并存储到查询器
func (s *MonitoringService) collectAndStoreMetrics() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒收集一次指标
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.collectCurrentMetrics()
		}
	}
}

// collectCurrentMetrics 收集当前指标并存储到查询器
func (s *MonitoringService) collectCurrentMetrics() {
	if s.querier == nil {
		return
	}
	
	memQuerier, ok := s.querier.(*query.MemoryQuerier)
	if !ok {
		return
	}
	
	now := time.Now()
	
	// 收集系统指标
	for _, collector := range s.collectors {
		switch collector.(type) {
		case *collectors.SystemCollector:
			s.collectSystemMetrics(memQuerier, now)
		case *collectors.ApplicationCollector:
			s.collectApplicationMetrics(memQuerier, now)
		case *collectors.NetworkCollector:
			s.collectNetworkMetrics(memQuerier, now)
		}
	}
}

// collectSystemMetrics 收集系统指标
func (s *MonitoringService) collectSystemMetrics(querier *query.MemoryQuerier, timestamp time.Time) {
	// 模拟收集CPU使用率
	cpuUsage := float64(30 + (timestamp.Unix() % 40)) // 30-70%的CPU使用率
	querier.AddMetric("cpu_usage", cpuUsage, map[string]string{})
	
	// 模拟收集内存使用率
	memUsage := float64(50 + (timestamp.Unix() % 30)) // 50-80%的内存使用率
	querier.AddMetric("memory_usage", memUsage, map[string]string{})
	
	// 模拟收集磁盘使用率
	diskUsage := float64(20 + (timestamp.Unix() % 60)) // 20-80%的磁盘使用率
	querier.AddMetric("disk_usage", diskUsage, map[string]string{})
}

// collectApplicationMetrics 收集应用指标
func (s *MonitoringService) collectApplicationMetrics(querier *query.MemoryQuerier, timestamp time.Time) {
	// 模拟收集Goroutine数量
	goroutineCount := float64(100 + (timestamp.Unix() % 200)) // 100-300个goroutine
	querier.AddMetric("goroutine_count", goroutineCount, map[string]string{})
	
	// 模拟收集堆内存使用
	heapUsage := float64(1024*1024 + (timestamp.Unix() % (10*1024*1024))) // 1MB-11MB
	querier.AddMetric("heap_usage", heapUsage, map[string]string{})
}

// collectNetworkMetrics 收集网络指标
func (s *MonitoringService) collectNetworkMetrics(querier *query.MemoryQuerier, timestamp time.Time) {
	// 模拟收集网络流量
	networkIn := float64(1024 + (timestamp.Unix() % 10240)) // 1KB-11KB/s
	networkOut := float64(512 + (timestamp.Unix() % 5120))  // 512B-5.5KB/s
	
	querier.AddMetric("network_in", networkIn, map[string]string{})
	querier.AddMetric("network_out", networkOut, map[string]string{})
}

func (s *MonitoringService) convertAlertingConfig(cfg config.AlertingConfig) alerting.AlertingConfig {
	// 转换规则
	var rules []alerting.AlertRule
	for i, rule := range cfg.Rules {
		alertRule := alerting.AlertRule{
			ID:          fmt.Sprintf("rule_%d", i),
			Name:        rule.Name,
			Query:       rule.Query,
			Duration:    rule.Duration,
			Severity:    rule.Severity,
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
			Enabled:     true,
		}
		rules = append(rules, alertRule)
	}

	// 转换接收器
	var receivers []alerting.Receiver
	for name, receiver := range cfg.Receivers {
		// 转换Config从map[string]string到map[string]interface{}
		config := make(map[string]interface{})
		for k, v := range receiver.Config {
			config[k] = v
		}
		
		alertReceiver := alerting.Receiver{
			Name:    name,
			Type:    receiver.Type,
			Config:  config,
			Enabled: true,
		}
		receivers = append(receivers, alertReceiver)
	}

	// 转换路由
	var routes []alerting.Route
	for _, route := range cfg.Routes {
		alertRoute := alerting.Route{
			Matchers: route.Match,
			Receiver: route.Receiver,
		}
		routes = append(routes, alertRoute)
	}

	return alerting.AlertingConfig{
		Enabled:            cfg.Enabled,
		EvaluationInterval: cfg.EvaluationInterval,
		Rules:              rules,
		Receivers:          receivers,
		Routes:             routes,
	}
}

// initErrorHandling 初始化错误处理组件
func (s *MonitoringService) initErrorHandling() error {
	// 创建错误报告器
	s.errorReporter = errors.NewInMemoryErrorReporter(s.logger, 1000)
	
	// 创建恢复管理器
	s.recoveryManager = errors.NewRecoveryManager(s.logger)
	
	// 注册默认恢复策略
	s.recoveryManager.RegisterStrategy(errors.ErrorTypeSystem, errors.NewDefaultRecoveryStrategy(
		errors.NewExponentialBackoff(time.Second, 30*time.Second, 2.0),
		3, // 最大重试次数
	))
	
	s.recoveryManager.RegisterStrategy(errors.ErrorTypeNetwork, errors.NewDefaultRecoveryStrategy(
		errors.NewLinearBackoff(time.Second, 5*time.Second),
		5, // 网络错误重试更多次
	))
	
	s.recoveryManager.RegisterStrategy(errors.ErrorTypeValidation, errors.NewDefaultRecoveryStrategy(
		errors.NewFixedBackoff(500*time.Millisecond),
		1, // 验证错误通常不需要重试
	))
	
	// 创建错误监控器
	s.errorMonitor = errors.NewErrorMonitor(s.errorReporter, s.recoveryManager, s.logger)
	
	// 设置错误阈值
	s.errorMonitor.SetThreshold(errors.ErrorTypeSystem, errors.ErrorThreshold{
		MaxErrorsPerMinute: 10,
		MaxErrorsPerHour:   100,
		AlertSeverity:      errors.SeverityHigh,
	})
	
	s.errorMonitor.SetThreshold(errors.ErrorTypeNetwork, errors.ErrorThreshold{
		MaxErrorsPerMinute: 20,
		MaxErrorsPerHour:   200,
		AlertSeverity:      errors.SeverityMedium,
	})
	
	return nil
}

// GetErrorMonitor 获取错误监控器
func (s *MonitoringService) GetErrorMonitor() *errors.ErrorMonitor {
	return s.errorMonitor
}

// GetRecoveryManager 获取恢复管理器
func (s *MonitoringService) GetRecoveryManager() *errors.RecoveryManager {
	return s.recoveryManager
}

// GetErrorReporter 获取错误报告器
func (s *MonitoringService) GetErrorReporter() errors.ErrorReporter {
	return s.errorReporter
}

// HandleError 处理错误的便捷方法
func (s *MonitoringService) HandleError(ctx context.Context, err error) error {
	// 如果不是MonitoringError，则包装它
	var monitoringErr *errors.MonitoringError
	if monErr, ok := errors.AsMonitoringError(err); ok {
		monitoringErr = monErr
	} else {
		monitoringErr = errors.NewError(errors.ErrorTypeSystem, "UNKNOWN_001", "Unknown error occurred").
			WithCause(err).
			WithComponent("MonitoringService").
			Build()
	}
	
	// 增加错误计数
	s.mu.Lock()
	s.stats.Errors++
	s.mu.Unlock()
	
	// 使用错误监控器处理错误
	if s.errorMonitor != nil {
		return s.errorMonitor.HandleError(ctx, monitoringErr)
	}
	
	// 如果没有错误监控器，直接记录日志
	s.logger.Error("Error occurred",
		zap.String("error_type", string(monitoringErr.Type)),
		zap.String("error_code", monitoringErr.Code),
		zap.String("component", monitoringErr.Component),
		zap.Error(err))
	
	return err
}