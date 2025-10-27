package sdk

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LifecycleManager 插件生命周期管理器
type LifecycleManager interface {
	// RegisterPlugin 注册插件
	RegisterPlugin(plugin Plugin) error
	
	// InitializePlugin 初始化插件
	InitializePlugin(ctx context.Context, pluginID string, config map[string]interface{}) error
	
	// StartPlugin 启动插件
	StartPlugin(ctx context.Context, pluginID string) error
	
	// StopPlugin 停止插件
	StopPlugin(ctx context.Context, pluginID string) error
	
	// RestartPlugin 重启插件
	RestartPlugin(ctx context.Context, pluginID string) error
	
	// UnregisterPlugin 注销插件
	UnregisterPlugin(ctx context.Context, pluginID string) error
	
	// GetPluginState 获取插件状态
	GetPluginState(pluginID string) (PluginState, error)
	
	// GetPluginHealth 获取插件健康状态
	GetPluginHealth(ctx context.Context, pluginID string) (*HealthStatus, error)
	
	// ListPlugins 列出所有插件
	ListPlugins() map[string]PluginState
}

// DefaultLifecycleManager 默认生命周期管理器实现
type DefaultLifecycleManager struct {
	plugins     map[string]Plugin
	states      map[string]PluginState
	contexts    map[string]*PluginContext
	healthCheck map[string]*HealthChecker
	logger      *logrus.Logger
	mu          sync.RWMutex
}

// NewDefaultLifecycleManager 创建默认生命周期管理器
func NewDefaultLifecycleManager(logger *logrus.Logger) *DefaultLifecycleManager {
	return &DefaultLifecycleManager{
		plugins:     make(map[string]Plugin),
		states:      make(map[string]PluginState),
		contexts:    make(map[string]*PluginContext),
		healthCheck: make(map[string]*HealthChecker),
		logger:      logger,
	}
}

// RegisterPlugin 注册插件
func (m *DefaultLifecycleManager) RegisterPlugin(plugin Plugin) error {
	info := plugin.GetInfo()
	if info == nil {
		return &PluginError{
			Code:    ErrCodeInvalidManifest,
			Message: "plugin info is nil",
		}
	}

	pluginID := info.ID
	if pluginID == "" {
		return &PluginError{
			Code:    ErrCodeInvalidManifest,
			Message: "plugin ID is empty",
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查插件是否已存在
	if _, exists := m.plugins[pluginID]; exists {
		return &PluginError{
			Code:    ErrCodePluginAlreadyExists,
			Message: fmt.Sprintf("plugin %s already registered", pluginID),
		}
	}

	// 注册插件
	m.plugins[pluginID] = plugin
	m.states[pluginID] = StateLoaded

	m.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"name":      info.Name,
		"version":   info.Version,
	}).Info("Plugin registered successfully")

	return nil
}

// InitializePlugin 初始化插件
func (m *DefaultLifecycleManager) InitializePlugin(ctx context.Context, pluginID string, config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return &PluginError{
			Code:    ErrCodePluginNotFound,
			Message: fmt.Sprintf("plugin %s not found", pluginID),
		}
	}

	// 检查当前状态
	currentState := m.states[pluginID]
	if currentState != StateLoaded {
		return &PluginError{
			Code:    "INVALID_STATE",
			Message: fmt.Sprintf("plugin %s is in %s state, expected %s", pluginID, currentState, StateLoaded),
		}
	}

	// 创建插件上下文
	pluginCtx := &PluginContext{
		PluginID:   pluginID,
		Config:     config,
		Logger:     m.logger.WithField("plugin_id", pluginID),
		DataDir:    fmt.Sprintf("/tmp/plugins/%s/data", pluginID),
		TempDir:    fmt.Sprintf("/tmp/plugins/%s/temp", pluginID),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Metadata:   make(map[string]string),
	}

	// 初始化插件
	if err := plugin.Initialize(ctx, pluginCtx); err != nil {
		m.states[pluginID] = StateError
		return &PluginError{
			Code:    ErrCodeInitializationFailed,
			Message: fmt.Sprintf("failed to initialize plugin %s", pluginID),
			Cause:   err,
		}
	}

	// 保存上下文
	m.contexts[pluginID] = pluginCtx
	m.states[pluginID] = StateInitialized

	// 创建健康检查器
	if hc, ok := plugin.(HealthCheckable); ok {
		m.healthCheck[pluginID] = NewHealthChecker(pluginID, hc, m.logger)
	}

	m.logger.WithField("plugin_id", pluginID).Info("Plugin initialized successfully")
	return nil
}

// StartPlugin 启动插件
func (m *DefaultLifecycleManager) StartPlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return &PluginError{
			Code:    ErrCodePluginNotFound,
			Message: fmt.Sprintf("plugin %s not found", pluginID),
		}
	}

	// 检查当前状态
	currentState := m.states[pluginID]
	if currentState != StateInitialized && currentState != StateStopped {
		return &PluginError{
			Code:    "INVALID_STATE",
			Message: fmt.Sprintf("plugin %s is in %s state, cannot start", pluginID, currentState),
		}
	}

	// 启动插件
	if err := plugin.Start(ctx); err != nil {
		m.states[pluginID] = StateError
		return &PluginError{
			Code:    "START_FAILED",
			Message: fmt.Sprintf("failed to start plugin %s", pluginID),
			Cause:   err,
		}
	}

	m.states[pluginID] = StateStarted

	// 启动健康检查
	if checker, exists := m.healthCheck[pluginID]; exists {
		checker.Start(ctx)
	}

	m.logger.WithField("plugin_id", pluginID).Info("Plugin started successfully")
	return nil
}

// StopPlugin 停止插件
func (m *DefaultLifecycleManager) StopPlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return &PluginError{
			Code:    ErrCodePluginNotFound,
			Message: fmt.Sprintf("plugin %s not found", pluginID),
		}
	}

	// 检查当前状态
	currentState := m.states[pluginID]
	if currentState != StateStarted {
		return &PluginError{
			Code:    "INVALID_STATE",
			Message: fmt.Sprintf("plugin %s is in %s state, cannot stop", pluginID, currentState),
		}
	}

	// 停止健康检查
	if checker, exists := m.healthCheck[pluginID]; exists {
		checker.Stop()
	}

	// 停止插件
	if err := plugin.Stop(ctx); err != nil {
		m.states[pluginID] = StateError
		return &PluginError{
			Code:    "STOP_FAILED",
			Message: fmt.Sprintf("failed to stop plugin %s", pluginID),
			Cause:   err,
		}
	}

	m.states[pluginID] = StateStopped

	m.logger.WithField("plugin_id", pluginID).Info("Plugin stopped successfully")
	return nil
}

// RestartPlugin 重启插件
func (m *DefaultLifecycleManager) RestartPlugin(ctx context.Context, pluginID string) error {
	// 先停止插件
	if err := m.StopPlugin(ctx, pluginID); err != nil {
		return err
	}

	// 等待一段时间
	time.Sleep(1 * time.Second)

	// 再启动插件
	return m.StartPlugin(ctx, pluginID)
}

// UnregisterPlugin 注销插件
func (m *DefaultLifecycleManager) UnregisterPlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return &PluginError{
			Code:    ErrCodePluginNotFound,
			Message: fmt.Sprintf("plugin %s not found", pluginID),
		}
	}

	// 如果插件正在运行，先停止
	if m.states[pluginID] == StateStarted {
		if err := plugin.Stop(ctx); err != nil {
			m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to stop plugin during unregistration")
		}
	}

	// 清理资源
	if err := plugin.Cleanup(ctx); err != nil {
		m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to cleanup plugin during unregistration")
	}

	// 停止健康检查
	if checker, exists := m.healthCheck[pluginID]; exists {
		checker.Stop()
		delete(m.healthCheck, pluginID)
	}

	// 删除插件
	delete(m.plugins, pluginID)
	delete(m.states, pluginID)
	delete(m.contexts, pluginID)

	m.logger.WithField("plugin_id", pluginID).Info("Plugin unregistered successfully")
	return nil
}

// GetPluginState 获取插件状态
func (m *DefaultLifecycleManager) GetPluginState(pluginID string) (PluginState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[pluginID]
	if !exists {
		return "", &PluginError{
			Code:    ErrCodePluginNotFound,
			Message: fmt.Sprintf("plugin %s not found", pluginID),
		}
	}

	return state, nil
}

// GetPluginHealth 获取插件健康状态
func (m *DefaultLifecycleManager) GetPluginHealth(ctx context.Context, pluginID string) (*HealthStatus, error) {
	m.mu.RLock()
	plugin, exists := m.plugins[pluginID]
	m.mu.RUnlock()

	if !exists {
		return nil, &PluginError{
			Code:    ErrCodePluginNotFound,
			Message: fmt.Sprintf("plugin %s not found", pluginID),
		}
	}

	return plugin.GetHealth(ctx)
}

// ListPlugins 列出所有插件
func (m *DefaultLifecycleManager) ListPlugins() map[string]PluginState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]PluginState)
	for pluginID, state := range m.states {
		result[pluginID] = state
	}

	return result
}

// HealthCheckable 健康检查接口
type HealthCheckable interface {
	GetHealth(ctx context.Context) (*HealthStatus, error)
}

// HealthChecker 健康检查器
type HealthChecker struct {
	pluginID string
	plugin   HealthCheckable
	logger   *logrus.Logger
	interval time.Duration
	timeout  time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(pluginID string, plugin HealthCheckable, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		pluginID: pluginID,
		plugin:   plugin,
		logger:   logger,
		interval: 30 * time.Second,
		timeout:  5 * time.Second,
	}
}

// Start 启动健康检查
func (h *HealthChecker) Start(ctx context.Context) {
	h.ctx, h.cancel = context.WithCancel(ctx)
	
	h.wg.Add(1)
	go h.run()
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	if h.cancel != nil {
		h.cancel()
	}
	h.wg.Wait()
}

// run 运行健康检查
func (h *HealthChecker) run() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.checkHealth()
		}
	}
}

// checkHealth 执行健康检查
func (h *HealthChecker) checkHealth() {
	ctx, cancel := context.WithTimeout(h.ctx, h.timeout)
	defer cancel()

	health, err := h.plugin.GetHealth(ctx)
	if err != nil {
		h.logger.WithError(err).WithField("plugin_id", h.pluginID).Warn("Health check failed")
		return
	}

	if health.Status != "healthy" {
		h.logger.WithFields(logrus.Fields{
			"plugin_id": h.pluginID,
			"status":    health.Status,
			"message":   health.Message,
		}).Warn("Plugin health check returned unhealthy status")
	}
}

// StateTransition 状态转换
type StateTransition struct {
	From   PluginState
	To     PluginState
	Action string
}

// ValidTransitions 有效的状态转换
var ValidTransitions = []StateTransition{
	{StateUnloaded, StateLoaded, "load"},
	{StateLoaded, StateInitialized, "initialize"},
	{StateInitialized, StateStarted, "start"},
	{StateStarted, StateStopped, "stop"},
	{StateStopped, StateStarted, "start"},
	{StateStopped, StateUnloaded, "unload"},
	{StateError, StateUnloaded, "unload"},
}

// IsValidTransition 检查状态转换是否有效
func IsValidTransition(from, to PluginState) bool {
	for _, transition := range ValidTransitions {
		if transition.From == from && transition.To == to {
			return true
		}
	}
	return false
}