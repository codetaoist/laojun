package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Executor 插件执行器接口
type Executor interface {
	// Execute 执行插件任务
	Execute(ctx context.Context, task *ExecutionTask) (*ExecutionResult, error)

	// ExecuteAsync 异步执行插件任务
	ExecuteAsync(task *ExecutionTask) (string, error)

	// GetExecutionResult 获取执行结果
	GetExecutionResult(taskID string) (*ExecutionResult, error)

	// CancelExecution 取消执行
	CancelExecution(taskID string) error

	// GetRunningTasks 获取正在运行的任务
	GetRunningTasks() []*ExecutionTask

	// GetExecutionHistory 获取执行历史
	GetExecutionHistory(limit int) []*ExecutionResult

	// SetConcurrencyLimit 设置并发限制
	SetConcurrencyLimit(limit int) error

	// Start 启动执行器
	Start(ctx context.Context) error

	// Stop 停止执行器
	Stop(ctx context.Context) error
}

// ExecutionTask 执行任务
type ExecutionTask struct {
	ID          string                 `json:"id"`
	PluginID    string                 `json:"plugin_id"`
	Type        TaskType               `json:"type"`
	Method      string                 `json:"method"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"` // 数字越小优先级越高
	Timeout     time.Duration          `json:"timeout"`
	Retry       *RetryConfig           `json:"retry,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Status      TaskStatus             `json:"status"`
	Context     context.Context        `json:"-"`
	Cancel      context.CancelFunc     `json:"-"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	TaskID      string                 `json:"task_id"`
	PluginID    string                 `json:"plugin_id"`
	Status      TaskStatus             `json:"status"`
	Result      interface{}            `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	RetryCount  int                    `json:"retry_count"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskType 任务类型
type TaskType string

const (
	TaskTypeHTTP      TaskType = "http"
	TaskTypeEvent     TaskType = "event"
	TaskTypeScheduled TaskType = "scheduled"
	TaskTypeData      TaskType = "data"
	TaskTypeCustom    TaskType = "custom"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusTimeout   TaskStatus = "timeout"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
	BackoffType BackoffType   `json:"backoff_type"`
	MaxDelay    time.Duration `json:"max_delay"`
}

// BackoffType 退避类型
type BackoffType string

const (
	BackoffTypeFixed       BackoffType = "fixed"
	BackoffTypeLinear      BackoffType = "linear"
	BackoffTypeExponential BackoffType = "exponential"
)

// DefaultExecutor 默认执行器实现
type DefaultExecutor struct {
	pluginManager    PluginManager
	taskQueue        chan *ExecutionTask
	results          map[string]*ExecutionResult
	runningTasks     map[string]*ExecutionTask
	concurrencyLimit int
	workerCount      int
	workers          []*ExecutorWorker
	logger           *logrus.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.RWMutex
	running          bool
}

// ExecutorWorker 执行器工作者
type ExecutorWorker struct {
	ID       int
	executor *DefaultExecutor
	logger   *logrus.Logger
}

// NewDefaultExecutor 创建默认执行器
func NewDefaultExecutor(
	pluginManager PluginManager,
	concurrencyLimit int,
	queueSize int,
	logger *logrus.Logger,
) *DefaultExecutor {
	if concurrencyLimit <= 0 {
		concurrencyLimit = 10
	}
	if queueSize <= 0 {
		queueSize = 1000
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &DefaultExecutor{
		pluginManager:    pluginManager,
		taskQueue:        make(chan *ExecutionTask, queueSize),
		results:          make(map[string]*ExecutionResult),
		runningTasks:     make(map[string]*ExecutionTask),
		concurrencyLimit: concurrencyLimit,
		workerCount:      concurrencyLimit,
		logger:           logger,
	}
}

// Start 启动执行器
func (e *DefaultExecutor) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("executor already running")
	}

	e.ctx, e.cancel = context.WithCancel(ctx)
	e.running = true

	// 启动工作者
	e.workers = make([]*ExecutorWorker, e.workerCount)
	for i := 0; i < e.workerCount; i++ {
		worker := &ExecutorWorker{
			ID:       i,
			executor: e,
			logger:   e.logger.WithField("worker", i).Logger,
		}
		e.workers[i] = worker

		e.wg.Add(1)
		go worker.run()
	}

	e.logger.Infof("Executor started with %d workers", e.workerCount)
	return nil
}

// Stop 停止执行器
func (e *DefaultExecutor) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return fmt.Errorf("executor not running")
	}

	e.cancel()
	e.running = false

	// 关闭任务队列
	close(e.taskQueue)

	// 等待工作者结束
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.logger.Info("Executor stopped")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for executor to stop")
	}
}

// Execute 执行插件任务
func (e *DefaultExecutor) Execute(ctx context.Context, task *ExecutionTask) (*ExecutionResult, error) {
	if task.ID == "" {
		task.ID = e.generateTaskID()
	}

	task.CreatedAt = time.Now()
	task.Status = TaskStatusPending
	task.Context = ctx

	// 同步执行
	result, err := e.executeTask(task)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.results[task.ID] = result
	e.mu.Unlock()

	return result, nil
}

// ExecuteAsync 异步执行插件任务
func (e *DefaultExecutor) ExecuteAsync(task *ExecutionTask) (string, error) {
	if task.ID == "" {
		task.ID = e.generateTaskID()
	}

	task.CreatedAt = time.Now()
	task.Status = TaskStatusPending

	select {
	case e.taskQueue <- task:
		e.logger.Infof("Task %s queued for async execution", task.ID)
		return task.ID, nil
	default:
		return "", fmt.Errorf("task queue is full")
	}
}

// GetExecutionResult 获取执行结果
func (e *DefaultExecutor) GetExecutionResult(taskID string) (*ExecutionResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result, exists := e.results[taskID]
	if !exists {
		return nil, fmt.Errorf("execution result not found for task %s", taskID)
	}

	return result, nil
}

// CancelExecution 取消执行
func (e *DefaultExecutor) CancelExecution(taskID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	task, exists := e.runningTasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found or not running", taskID)
	}

	if task.Cancel != nil {
		task.Cancel()
	}

	task.Status = TaskStatusCancelled
	e.logger.Infof("Task %s cancelled", taskID)
	return nil
}

// GetRunningTasks 获取正在运行的任务
func (e *DefaultExecutor) GetRunningTasks() []*ExecutionTask {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var tasks []*ExecutionTask
	for _, task := range e.runningTasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetExecutionHistory 获取执行历史
func (e *DefaultExecutor) GetExecutionHistory(limit int) []*ExecutionResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var results []*ExecutionResult
	for _, result := range e.results {
		results = append(results, result)
	}

	// 按完成时间排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].CompletedAt.Before(results[j].CompletedAt) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// SetConcurrencyLimit 设置并发限制
func (e *DefaultExecutor) SetConcurrencyLimit(limit int) error {
	if limit <= 0 {
		return fmt.Errorf("concurrency limit must be positive")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.concurrencyLimit = limit
	e.logger.Infof("Concurrency limit set to %d", limit)
	return nil
}

// executeTask 执行任务
func (e *DefaultExecutor) executeTask(task *ExecutionTask) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// 设置超时上下文
	ctx := task.Context
	if task.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, task.Timeout)
		defer cancel()
	}

	// 获取插件
	plugin, err := e.pluginManager.GetPlugin(task.PluginID)
	if err != nil {
		return e.createErrorResult(task, startTime, fmt.Sprintf("plugin not found: %v", err)), nil
	}

	// 执行任务
	var result interface{}
	var execErr error

	switch task.Type {
	case TaskTypeHTTP:
		result, execErr = e.executeHTTPTask(ctx, plugin, task)
	case TaskTypeEvent:
		result, execErr = e.executeEventTask(ctx, plugin, task)
	case TaskTypeScheduled:
		result, execErr = e.executeScheduledTask(ctx, plugin, task)
	case TaskTypeData:
		result, execErr = e.executeDataTask(ctx, plugin, task)
	case TaskTypeCustom:
		result, execErr = e.executeCustomTask(ctx, plugin, task)
	default:
		execErr = fmt.Errorf("unsupported task type: %s", task.Type)
	}

	duration := time.Since(startTime)

	if execErr != nil {
		return e.createErrorResult(task, startTime, execErr.Error()), nil
	}

	return &ExecutionResult{
		TaskID:      task.ID,
		PluginID:    task.PluginID,
		Status:      TaskStatusCompleted,
		Result:      result,
		Duration:    duration,
		StartedAt:   startTime,
		CompletedAt: time.Now(),
	}, nil
}

// executeHTTPTask 执行HTTP任务
func (e *DefaultExecutor) executeHTTPTask(ctx context.Context, plugin Plugin, task *ExecutionTask) (interface{}, error) {
	// 检查插件是否支持HTTP处理
	metadata := plugin.GetMetadata()
	if metadata.Type != "http" {
		return nil, fmt.Errorf("plugin does not support HTTP tasks")
	}

	// 构造HTTP请求数据
	requestData := map[string]interface{}{
		"id":         task.ID,
		"method":     task.Method,
		"path":       "/",
		"headers":    make(map[string]string),
		"body":       task.Parameters,
		"remoteAddr": "127.0.0.1",
		"timestamp":  time.Now(),
	}

	// 使用插件的通用处理方法
	return plugin.ProcessData(ctx, requestData)
}

// executeEventTask 执行事件任务
func (e *DefaultExecutor) executeEventTask(ctx context.Context, plugin Plugin, task *ExecutionTask) (interface{}, error) {
	// 检查插件是否支持事件处理
	metadata := plugin.GetMetadata()
	if metadata.Type != "event" {
		return nil, fmt.Errorf("plugin does not support event tasks")
	}

	// 构造事件数据
	eventData := map[string]interface{}{
		"id":        task.ID,
		"type":      task.Method,
		"source":    "executor",
		"data":      task.Parameters,
		"timestamp": time.Now(),
	}

	// 使用插件的事件处理方法
	err := plugin.HandleEvent(ctx, eventData)
	return nil, err
}

// executeScheduledTask 执行定时任务
func (e *DefaultExecutor) executeScheduledTask(ctx context.Context, plugin Plugin, task *ExecutionTask) (interface{}, error) {
	// 检查插件是否支持定时任务
	metadata := plugin.GetMetadata()
	if metadata.Type != "scheduled" {
		return nil, fmt.Errorf("plugin does not support scheduled tasks")
	}

	// 构造定时任务数据
	taskData := map[string]interface{}{
		"method":     task.Method,
		"parameters": task.Parameters,
		"scheduled":  true,
		"timestamp":  time.Now(),
	}

	return plugin.ProcessData(ctx, taskData)
}

// executeDataTask 执行数据任务
func (e *DefaultExecutor) executeDataTask(ctx context.Context, plugin Plugin, task *ExecutionTask) (interface{}, error) {
	// 检查插件是否支持数据处理
	metadata := plugin.GetMetadata()
	if metadata.Type != "data" {
		return nil, fmt.Errorf("plugin does not support data tasks")
	}

	return plugin.ProcessData(ctx, task.Parameters)
}

// executeCustomTask 执行自定义任务
func (e *DefaultExecutor) executeCustomTask(ctx context.Context, plugin Plugin, task *ExecutionTask) (interface{}, error) {
	// 构造自定义请求数据
	requestData := map[string]interface{}{
		"id":         task.ID,
		"method":     task.Method,
		"path":       "/custom",
		"headers":    make(map[string]string),
		"body":       task.Parameters,
		"remoteAddr": "127.0.0.1",
		"timestamp":  time.Now(),
	}

	return plugin.ProcessData(ctx, requestData)
}

// createErrorResult 创建错误结果
func (e *DefaultExecutor) createErrorResult(task *ExecutionTask, startTime time.Time, errorMsg string) *ExecutionResult {
	return &ExecutionResult{
		TaskID:      task.ID,
		PluginID:    task.PluginID,
		Status:      TaskStatusFailed,
		Error:       errorMsg,
		Duration:    time.Since(startTime),
		StartedAt:   startTime,
		CompletedAt: time.Now(),
	}
}

// generateTaskID 生成任务ID
func (e *DefaultExecutor) generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// run 工作者运行循环
func (w *ExecutorWorker) run() {
	defer w.executor.wg.Done()

	w.logger.Info("Worker started")
	defer w.logger.Info("Worker stopped")

	for {
		select {
		case task, ok := <-w.executor.taskQueue:
			if !ok {
				return // 队列已关闭
			}
			w.processTask(task)
		case <-w.executor.ctx.Done():
			return
		}
	}
}

// processTask 处理任务
func (w *ExecutorWorker) processTask(task *ExecutionTask) {
	w.logger.Infof("Processing task %s", task.ID)

	// 设置任务状态
	task.Status = TaskStatusRunning
	startTime := time.Now()
	task.StartedAt = &startTime

	// 添加到运行任务列表
	w.executor.mu.Lock()
	w.executor.runningTasks[task.ID] = task
	w.executor.mu.Unlock()

	// 设置取消上下文
	ctx, cancel := context.WithCancel(w.executor.ctx)
	task.Context = ctx
	task.Cancel = cancel

	// 执行任务
	result, err := w.executor.executeTask(task)
	if err != nil {
		w.logger.Errorf("Task %s execution failed: %v", task.ID, err)
		result = w.executor.createErrorResult(task, startTime, err.Error())
	}

	// 处理重试
	if result.Status == TaskStatusFailed && task.Retry != nil && result.RetryCount < task.Retry.MaxRetries {
		w.retryTask(task, result)
		return
	}

	// 完成任务
	completedTime := time.Now()
	task.CompletedAt = &completedTime
	task.Status = result.Status

	// 保存结果
	w.executor.mu.Lock()
	w.executor.results[task.ID] = result
	delete(w.executor.runningTasks, task.ID)
	w.executor.mu.Unlock()

	w.logger.Infof("Task %s completed with status %s", task.ID, result.Status)
}

// retryTask 重试任务
func (w *ExecutorWorker) retryTask(task *ExecutionTask, lastResult *ExecutionResult) {
	retryCount := lastResult.RetryCount + 1
	delay := w.calculateRetryDelay(task.Retry, retryCount)

	w.logger.Infof("Retrying task %s (attempt %d) after %v", task.ID, retryCount, delay)

	// 等待重试延迟
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		// 重新排队任务
		task.Status = TaskStatusPending
		select {
		case w.executor.taskQueue <- task:
			// 任务重新排队成功
		default:
			// 队列满，标记为失败
			w.logger.Errorf("Failed to requeue task %s: queue full", task.ID)
			lastResult.Status = TaskStatusFailed
			w.executor.mu.Lock()
			w.executor.results[task.ID] = lastResult
			delete(w.executor.runningTasks, task.ID)
			w.executor.mu.Unlock()
		}
	case <-w.executor.ctx.Done():
		// 执行器已停止
		return
	}
}

// calculateRetryDelay 计算重试延迟
func (w *ExecutorWorker) calculateRetryDelay(retry *RetryConfig, retryCount int) time.Duration {
	baseDelay := retry.RetryDelay
	if baseDelay <= 0 {
		baseDelay = time.Second
	}

	var delay time.Duration
	switch retry.BackoffType {
	case BackoffTypeFixed:
		delay = baseDelay
	case BackoffTypeLinear:
		delay = baseDelay * time.Duration(retryCount)
	case BackoffTypeExponential:
		delay = baseDelay * time.Duration(1<<uint(retryCount-1))
	default:
		delay = baseDelay
	}

	if retry.MaxDelay > 0 && delay > retry.MaxDelay {
		delay = retry.MaxDelay
	}

	return delay
}