package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/laojun-shared/monitoring"
	"github.com/codetaoist/laojun-shared/tracing"
)

// observabilityImpl 实现Observability接口
type observabilityImpl struct {
	config   *Config
	monitor  monitoring.Monitor
	tracer   tracing.Tracer
	mu       sync.RWMutex
	closed   bool
}

// operationImpl 实现Operation接口
type operationImpl struct {
	name        string
	id          string
	startTime   time.Time
	config      *OperationConfig
	span        tracing.Span
	ctx         context.Context
	monitor     monitoring.Monitor
	tracer      tracing.Tracer
	finished    bool
	status      OperationStatus
	attributes  map[string]interface{}
	error       error
	mu          sync.RWMutex
}

// NewObservability 创建新的可观测性实例
func NewObservability(config *Config) (Observability, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	obs := &observabilityImpl{
		config: config,
	}

	// 如果启用了监控，则初始化监控
	if config.Monitoring != nil {
		monitorConfig := config.GetMonitoringConfig()
		monitor, err := monitoring.NewMonitor(*monitorConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create monitor: %w", err)
		}
		obs.monitor = monitor
	}

	// 如果启用了追踪，则初始化追踪
	if config.Tracing != nil {
		tracer, err := tracing.NewTracer(config.GetTracingConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to create tracer: %w", err)
		}
		obs.tracer = tracer
	}

	return obs, nil
}

// Monitor 返回监控实例
func (o *observabilityImpl) Monitor() monitoring.Monitor {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.monitor
}

// Tracer 返回追踪实例
func (o *observabilityImpl) Tracer() tracing.Tracer {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.tracer
}

// StartOperation 开始一个新的统一操作
func (o *observabilityImpl) StartOperation(ctx context.Context, name string, opts ...OperationOption) (Operation, context.Context) {
	// 检查参数有效性
	if ctx == nil {
		ctx = context.Background()
	}
	if name == "" {
		name = "unnamed-operation"
	}

	// 检查实例状态
	o.mu.RLock()
	if o.closed {
		o.mu.RUnlock()
		// 返回一个空操作，避免panic
		return &operationImpl{
			name:       name,
			id:         fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
			startTime:  time.Now(),
			ctx:        ctx,
			finished:   true,
			status:     OperationStatusError,
			attributes: make(map[string]interface{}),
		}, ctx
	}
	o.mu.RUnlock()

	config := &OperationConfig{
		Type:         OperationTypeGeneric,
		Attributes:   make(map[string]interface{}),
		Labels:       make(map[string]string),
		SampleRate:   1.0,
		Detailed:     false,
		CustomFields: make(map[string]interface{}),
	}

	// 应用选项，添加错误处理
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}

	// 生成操作ID
	operationID := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())

	op := &operationImpl{
		name:       name,
		id:         operationID,
		startTime:  time.Now(),
		config:     config,
		ctx:        ctx,
		monitor:    o.monitor,
		tracer:     o.tracer,
		finished:   false,
		status:     OperationStatusStarted,
		attributes: make(map[string]interface{}),
	}

	// 复制初始属性
	for k, v := range config.Attributes {
		op.attributes[k] = v
	}

	// 如果启用了追踪，创建span
	if o.tracer != nil {
		span, spanCtx := o.tracer.StartSpan(ctx, name)
		op.span = span
		op.ctx = spanCtx

		// 设置span属性
		for k, v := range config.Attributes {
			span.SetTag(k, v)
		}
		span.SetTag("operation.type", string(config.Type))
		span.SetTag("operation.id", operationID)
	}

	// 记录操作开始的监控指标
	if o.monitor != nil {
		labels := map[string]string{
			"operation": name,
			"type":      string(config.Type),
		}
		for k, v := range config.Labels {
			labels[k] = v
		}
		
		o.monitor.IncrementCounter(ctx, "operations_started_total", labels)
		o.monitor.SetGauge(ctx, "operations_active", 1, labels)
	}

	return op, op.ctx
}

// HealthCheck 返回健康状态
func (o *observabilityImpl) HealthCheck() HealthStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()

	status := HealthStatus{
		Status:     "healthy",
		Details:    make(map[string]interface{}),
		Timestamp:  time.Now(),
		Components: make(map[string]ComponentHealth),
	}

	// 检查监控组件
	if o.monitor != nil {
		status.Components["monitoring"] = ComponentHealth{
			Status:    "healthy",
			Message:   "监控组件运行正常",
			LastCheck: time.Now(),
		}
	}

	// 检查追踪组件
	if o.tracer != nil {
		status.Components["tracing"] = ComponentHealth{
			Status:    "healthy",
			Message:   "追踪组件运行正常",
			LastCheck: time.Now(),
		}
	}

	// 检查整体状态
	if o.closed {
		status.Status = "unhealthy"
		status.Details["reason"] = "服务已关闭"
	}

	return status
}

// Export 导出数据
func (o *observabilityImpl) Export(ctx context.Context, format string) error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.closed {
		return fmt.Errorf("observability instance is closed")
	}

	// 导出监控数据
	if o.monitor != nil {
		var exportFormat monitoring.ExportFormat
		switch format {
		case "prometheus":
			exportFormat = monitoring.ExportFormatPrometheus
		case "json":
			exportFormat = monitoring.ExportFormatJSON
		case "influxdb":
			exportFormat = monitoring.ExportFormatInfluxDB
		default:
			exportFormat = monitoring.ExportFormatPrometheus
		}
		_, err := o.monitor.Export(ctx, exportFormat)
		if err != nil {
			return fmt.Errorf("failed to export monitoring data: %w", err)
		}
	}

	// 导出追踪数据
	if o.tracer != nil {
		var exportFormat tracing.ExportFormat
		switch format {
		case "jaeger":
			exportFormat = tracing.ExportFormatJaeger
		case "zipkin":
			exportFormat = tracing.ExportFormatZipkin
		case "otlp":
			exportFormat = tracing.ExportFormatOTLP
		case "json":
			exportFormat = tracing.ExportFormatJSON
		default:
			exportFormat = tracing.ExportFormatJaeger
		}
		_, err := o.tracer.Export(ctx, exportFormat)
		if err != nil {
			return fmt.Errorf("failed to export tracing data: %w", err)
		}
	}

	return nil
}

// Close 关闭资源
func (o *observabilityImpl) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.closed {
		return nil
	}

	var errs []error

	// 关闭监控
	if o.monitor != nil {
		if err := o.monitor.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close monitor: %w", err))
		}
	}

	// 关闭追踪
	if o.tracer != nil {
		if err := o.tracer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close tracer: %w", err))
		}
	}

	o.closed = true

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// Name 返回操作名称
func (op *operationImpl) Name() string {
	return op.name
}

// ID 返回操作ID
func (op *operationImpl) ID() string {
	return op.id
}

// StartTime 返回操作开始时间
func (op *operationImpl) StartTime() time.Time {
	return op.startTime
}

// Duration 返回操作持续时间
func (op *operationImpl) Duration() time.Duration {
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	if op.finished && op.span != nil {
		return op.span.Duration()
	}
	return time.Since(op.startTime)
}

// SetStatus 设置操作状态
func (op *operationImpl) SetStatus(status OperationStatus) {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	// 验证状态转换的合理性
	if op.status == OperationStatusSuccess || op.status == OperationStatusError || 
	   op.status == OperationStatusCancelled || op.status == OperationStatusTimeout {
		// 已经是终态，不允许再次修改
		return
	}
	
	op.status = status
	
	if op.span != nil {
		op.span.SetTag("operation.status", string(status))
	}
}

// GetStatus 获取操作状态
func (op *operationImpl) GetStatus() OperationStatus {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.status
}

// SetAttribute 设置属性
func (op *operationImpl) SetAttribute(key string, value interface{}) {
	// 检查参数有效性
	if key == "" {
		return
	}
	
	op.mu.Lock()
	defer op.mu.Unlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	// 确保attributes map已初始化
	if op.attributes == nil {
		op.attributes = make(map[string]interface{})
	}
	
	op.attributes[key] = value
	
	if op.span != nil {
		op.span.SetTag(key, value)
	}
}

// GetAttribute 获取属性
func (op *operationImpl) GetAttribute(key string) interface{} {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.attributes[key]
}

// SetAttributes 设置多个属性
func (op *operationImpl) SetAttributes(attrs map[string]interface{}) {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	for k, v := range attrs {
		op.attributes[k] = v
		if op.span != nil {
			op.span.SetTag(k, v)
		}
	}
}

// GetAttributes 获取所有属性
func (op *operationImpl) GetAttributes() map[string]interface{} {
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range op.attributes {
		result[k] = v
	}
	return result
}

// AddEvent 添加事件
func (op *operationImpl) AddEvent(name string, attrs ...EventAttribute) {
	// 检查参数有效性
	if name == "" {
		return
	}
	
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	if op.span != nil {
		tracingAttrs := make([]tracing.EventAttribute, 0, len(attrs))
		for _, attr := range attrs {
			if attr.Key != "" {
				tracingAttrs = append(tracingAttrs, tracing.EventAttribute{
					Key:   attr.Key,
					Value: attr.Value,
				})
			}
		}
		op.span.AddEvent(name, tracingAttrs...)
	}
}

// SetError 设置错误
func (op *operationImpl) SetError(err error) {
	// 检查参数有效性
	if err == nil {
		return
	}
	
	op.mu.Lock()
	defer op.mu.Unlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	op.error = err
	op.status = OperationStatusError
	
	if op.span != nil {
		op.span.RecordError(err)
		op.span.SetTag("error", true)
		op.span.SetTag("error.message", err.Error())
	}
}

// GetError 获取错误
func (op *operationImpl) GetError() error {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.error
}

// IncrementCounter 增加计数器指标
func (op *operationImpl) IncrementCounter(name string, value float64, labels ...map[string]string) {
	// 检查参数有效性
	if name == "" {
		return
	}
	
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	if op.monitor != nil {
		var mergedLabels map[string]string
		if len(labels) > 0 && labels[0] != nil {
			mergedLabels = labels[0]
		} else {
			mergedLabels = make(map[string]string)
		}
		op.monitor.IncrementCounter(op.ctx, name, mergedLabels)
	}
}

// SetGauge 设置仪表盘指标
func (op *operationImpl) SetGauge(name string, value float64, labels ...map[string]string) {
	// 检查参数有效性
	if name == "" {
		return
	}
	
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	if op.monitor != nil {
		var mergedLabels map[string]string
		if len(labels) > 0 && labels[0] != nil {
			mergedLabels = labels[0]
		} else {
			mergedLabels = make(map[string]string)
		}
		op.monitor.SetGauge(op.ctx, name, value, mergedLabels)
	}
}

// AddCounter 添加计数器指标
func (op *operationImpl) AddCounter(name string, value float64, labels ...map[string]string) {
	// 检查参数有效性
	if name == "" {
		return
	}
	
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	if op.monitor != nil {
		var mergedLabels map[string]string
		if len(labels) > 0 && labels[0] != nil {
			mergedLabels = labels[0]
		} else {
			mergedLabels = make(map[string]string)
		}
		op.monitor.AddCounter(op.ctx, name, value, mergedLabels)
	}
}

// RecordHistogram 记录直方图指标
func (op *operationImpl) RecordHistogram(name string, value float64, labels ...map[string]string) {
	// 检查参数有效性
	if name == "" {
		return
	}
	
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	if op.monitor != nil {
		var mergedLabels map[string]string
		if len(labels) > 0 && labels[0] != nil {
			mergedLabels = labels[0]
		} else {
			mergedLabels = make(map[string]string)
		}
		op.monitor.RecordHistogram(op.ctx, name, value, mergedLabels)
	}
}

// RecordSummary 记录摘要指标
func (op *operationImpl) RecordSummary(name string, value float64, labels ...map[string]string) {
	// 检查参数有效性
	if name == "" {
		return
	}
	
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	// 检查操作是否已完成
	if op.finished {
		return
	}
	
	if op.monitor != nil {
		var mergedLabels map[string]string
		if len(labels) > 0 && labels[0] != nil {
			mergedLabels = labels[0]
		} else {
			mergedLabels = make(map[string]string)
		}
		op.monitor.RecordSummary(op.ctx, name, value, mergedLabels)
	}
}

// StartChild 开始子操作
func (op *operationImpl) StartChild(name string, opts ...OperationOption) Operation {
	// 检查参数有效性
	if name == "" {
		name = "unnamed-child-operation"
	}
	
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	// 检查父操作是否已完成
	if op.finished {
		// 返回一个已完成的空操作，避免panic
		return &operationImpl{
			name:       name,
			id:         fmt.Sprintf("%s-child-%d", op.id, time.Now().UnixNano()),
			startTime:  time.Now(),
			ctx:        op.ctx,
			finished:   true,
			status:     OperationStatusError,
			attributes: make(map[string]interface{}),
		}
	}
	
	// 检查父操作配置是否有效
	if op.config == nil {
		op.config = &OperationConfig{
			Type:         OperationTypeGeneric,
			Attributes:   make(map[string]interface{}),
			Labels:       make(map[string]string),
			SampleRate:   1.0,
			Detailed:     false,
			CustomFields: make(map[string]interface{}),
		}
	}
	
	// 创建子操作配置
	config := &OperationConfig{
		Type:         op.config.Type,
		Attributes:   make(map[string]interface{}),
		Labels:       make(map[string]string),
		SampleRate:   op.config.SampleRate,
		Detailed:     op.config.Detailed,
		CustomFields: make(map[string]interface{}),
	}

	// 复制父操作的属性和标签
	if op.config.Attributes != nil {
		for k, v := range op.config.Attributes {
			config.Attributes[k] = v
		}
	}
	if op.config.Labels != nil {
		for k, v := range op.config.Labels {
			config.Labels[k] = v
		}
	}

	// 应用选项，添加错误处理
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}

	// 生成子操作ID
	childID := fmt.Sprintf("%s-child-%d", op.id, time.Now().UnixNano())

	child := &operationImpl{
		name:       name,
		id:         childID,
		startTime:  time.Now(),
		config:     config,
		ctx:        op.ctx,
		monitor:    op.monitor,
		tracer:     op.tracer,
		finished:   false,
		status:     OperationStatusStarted,
		attributes: make(map[string]interface{}),
	}

	// 复制初始属性
	for k, v := range config.Attributes {
		child.attributes[k] = v
	}

	// 如果父操作有span，创建子span
	if op.span != nil && op.tracer != nil {
		span, spanCtx := op.tracer.StartSpan(op.ctx, name)
		child.span = span
		child.ctx = spanCtx

		// 设置span属性
		for k, v := range config.Attributes {
			span.SetTag(k, v)
		}
		span.SetTag("operation.type", string(config.Type))
		span.SetTag("operation.id", childID)
		span.SetTag("parent.id", op.id)
	}

	return child
}

// Finish 完成操作
func (op *operationImpl) Finish() {
	op.FinishWithOptions(FinishOptions{})
}

// FinishWithOptions 使用选项完成操作
func (op *operationImpl) FinishWithOptions(opts FinishOptions) {
	op.mu.Lock()
	defer op.mu.Unlock()

	if op.finished {
		return
	}

	// 确保attributes map已初始化
	if op.attributes == nil {
		op.attributes = make(map[string]interface{})
	}

	// 设置最终状态
	if opts.FinalStatus != "" {
		op.status = opts.FinalStatus
	} else if op.status == OperationStatusStarted || op.status == OperationStatusRunning {
		// 如果状态还是初始状态，设置为成功
		op.status = OperationStatusSuccess
	}

	// 设置最终错误
	if opts.FinalError != nil {
		op.error = opts.FinalError
		op.status = OperationStatusError
	}

	// 设置额外属性
	if opts.ExtraAttributes != nil {
		for k, v := range opts.ExtraAttributes {
			if k != "" {
				op.attributes[k] = v
			}
		}
	}

	// 记录操作完成的监控指标
	if op.monitor != nil {
		labels := map[string]string{
			"operation": op.name,
			"status":    string(op.status),
		}
		
		// 安全地复制配置标签
		if op.config != nil && op.config.Labels != nil {
			for k, v := range op.config.Labels {
				if k != "" {
					labels[k] = v
				}
			}
		}
		
		// 使用defer确保即使出现panic也能记录指标
		defer func() {
			if r := recover(); r != nil {
				// 记录panic但不重新抛出，避免影响操作完成
			}
		}()
		
		op.monitor.IncrementCounter(op.ctx, "operations_completed_total", labels)
		op.monitor.SetGauge(op.ctx, "operations_active", -1, labels)
		
		// 计算持续时间，避免调用Duration()方法造成死锁
		var duration time.Duration
		if opts.FinishTime != nil {
			duration = opts.FinishTime.Sub(op.startTime)
		} else {
			duration = time.Since(op.startTime)
		}
		
		// 确保持续时间为正值
		if duration < 0 {
			duration = 0
		}
		
		op.monitor.RecordHistogram(op.ctx, "operation_duration_seconds", duration.Seconds(), labels)
	}

	// 完成span
	if op.span != nil {
		// 添加最终属性
		for k, v := range opts.ExtraAttributes {
			op.span.SetTag(k, v)
		}
		
		// 设置最终状态
		op.span.SetTag("operation.status", string(op.status))
		
		// 如果有错误，记录错误
		if op.error != nil {
			op.span.RecordError(op.error)
		}
		
		if opts.FinishTime != nil {
			op.span.FinishWithOptions(tracing.FinishOptions{
				FinishTime: *opts.FinishTime,
			})
		} else {
			op.span.Finish()
		}
	}

	op.finished = true
}

// IsFinished returns whether the operation is finished
func (op *operationImpl) IsFinished() bool {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.finished
}

// Context returns the operation context
func (op *operationImpl) Context() context.Context {
	return op.ctx
}

// Span returns the operation span
func (op *operationImpl) Span() tracing.Span {
	return op.span
}

// mergeLabels merges operation labels with provided labels
func (op *operationImpl) mergeLabels(labels map[string]string) map[string]string {
	merged := make(map[string]string)
	
	// Add operation labels
	for k, v := range op.config.Labels {
		merged[k] = v
	}
	
	// Add provided labels (override operation labels)
	for k, v := range labels {
		merged[k] = v
	}
	
	return merged
}