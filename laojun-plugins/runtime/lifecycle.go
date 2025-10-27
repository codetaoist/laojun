package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LifecycleManager 插件生命周期管理器接口
type LifecycleManager interface {
	// InitializePlugin 初始化插件
	InitializePlugin(ctx context.Context, pluginID string, config map[string]interface{}) error

	// StartPlugin 启动插件
	StartPlugin(ctx context.Context, pluginID string) error

	// StopPlugin 停止插件
	StopPlugin(ctx context.Context, pluginID string) error

	// RestartPlugin 重启插件
	RestartPlugin(ctx context.Context, pluginID string) error

	// CleanupPlugin 清理插件
	CleanupPlugin(ctx context.Context, pluginID string) error

	// GetPluginLifecycleState 获取插件生命周期状态
	GetPluginLifecycleState(pluginID string) (*LifecycleState, error)

	// SetLifecycleHook 设置生命周期钩子
	SetLifecycleHook(hook LifecycleHook) error

	// RemoveLifecycleHook 移除生命周期钩子
	RemoveLifecycleHook(hookID string) error

	// GetLifecycleHistory 获取生命周期历史
	GetLifecycleHistory(pluginID string) ([]*LifecycleEvent, error)
}

// LifecycleState 插件生命周期状态
type LifecycleState struct {
	PluginID        string                 `json:"plugin_id"`
	CurrentState    PluginState           `json:"current_state"`
	PreviousState   PluginState           `json:"previous_state"`
	StateChangedAt  time.Time             `json:"state_changed_at"`
	InitializedAt   *time.Time            `json:"initialized_at,omitempty"`
	StartedAt       *time.Time            `json:"started_at,omitempty"`
	StoppedAt       *time.Time            `json:"stopped_at,omitempty"`
	LastError       string                `json:"last_error,omitempty"`
	RestartCount    int                   `json:"restart_count"`
	Configuration   map[string]interface{} `json:"configuration,omitempty"`
}

// LifecycleEvent 生命周期事件
type LifecycleEvent struct {
	PluginID    string      `json:"plugin_id"`
	EventType   string      `json:"event_type"` // initialize, start, stop, restart, cleanup, error
	FromState   PluginState `json:"from_state"`
	ToState     PluginState `json:"to_state"`
	Timestamp   time.Time   `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
	Success     bool        `json:"success"`
	Error       string      `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LifecycleHook 生命周期钩子
type LifecycleHook struct {
	ID          string                    `json:"id"`
	EventType   string                    `json:"event_type"` // before_initialize, after_initialize, etc.
	PluginID    string                    `json:"plugin_id,omitempty"` // 空表示全局钩子
	Handler     LifecycleHookHandler      `json:"-"`
	Priority    int                       `json:"priority"` // 数字越小优先级越高
	Enabled     bool                      `json:"enabled"`
	CreatedAt   time.Time                 `json:"created_at"`
}

// LifecycleHookHandler 生命周期钩子处理函数
type LifecycleHookHandler func(ctx context.Context, event *LifecycleEvent) error

// DefaultLifecycleManager 默认生命周期管理器实现
type DefaultLifecycleManager struct {
	pluginManager PluginManager
	registry      PluginRegistry
	states        map[string]*LifecycleState
	history       map[string][]*LifecycleEvent
	hooks         map[string][]*LifecycleHook // key: event_type
	logger        *logrus.Logger
	mu            sync.RWMutex
}

// NewDefaultLifecycleManager 创建默认生命周期管理器
func NewDefaultLifecycleManager(
	pluginManager PluginManager,
	registry PluginRegistry,
	logger *logrus.Logger,
) *DefaultLifecycleManager {
	return &DefaultLifecycleManager{
		pluginManager: pluginManager,
		registry:      registry,
		states:        make(map[string]*LifecycleState),
		history:       make(map[string][]*LifecycleEvent),
		hooks:         make(map[string][]*LifecycleHook),
		logger:        logger,
	}
}

// InitializePlugin 初始化插件
func (lm *DefaultLifecycleManager) InitializePlugin(
	ctx context.Context,
	pluginID string,
	config map[string]interface{},
) error {
	startTime := time.Now()
	
	// 执行前置钩子
	if err := lm.executeHooks(ctx, "before_initialize", pluginID, nil); err != nil {
		return fmt.Errorf("before_initialize hook failed: %w", err)
	}

	// 获取当前状态
	currentState := lm.getCurrentState(pluginID)
	
	// 创建生命周期事件
	event := &LifecycleEvent{
		PluginID:  pluginID,
		EventType: "initialize",
		FromState: currentState,
		ToState:   StateInitializing,
		Timestamp: startTime,
		Metadata:  map[string]interface{}{"config": config},
	}

	// 更新状态
	lm.updateState(pluginID, StateInitializing, currentState, config)

	var err error
	defer func() {
		event.Duration = time.Since(startTime)
		event.Success = err == nil
		if err != nil {
			event.Error = err.Error()
			event.ToState = StateError
			lm.updateStateWithError(pluginID, StateError, err.Error())
		} else {
			event.ToState = StateInitialized
			lm.updateState(pluginID, StateInitialized, StateInitializing, config)
		}
		lm.addHistoryEvent(pluginID, event)
	}()

	// 执行初始化
	plugin, pluginErr := lm.pluginManager.GetPlugin(pluginID)
	if pluginErr != nil {
		err = fmt.Errorf("failed to get plugin: %w", pluginErr)
		return err
	}

	if plugin == nil {
		err = fmt.Errorf("plugin %s not found", pluginID)
		return err
	}

	// 调用插件的初始化方法
	if initErr := plugin.Initialize(ctx, config); initErr != nil {
		err = fmt.Errorf("plugin initialization failed: %w", initErr)
		return err
	}

	// 更新注册中心状态
	if regErr := lm.registry.UpdatePluginStatus(pluginID, StateInitialized); regErr != nil {
		lm.logger.WithError(regErr).Warn("Failed to update plugin status in registry")
	}

	// 执行后置钩子
	if hookErr := lm.executeHooks(ctx, "after_initialize", pluginID, event); hookErr != nil {
		lm.logger.WithError(hookErr).Warn("after_initialize hook failed")
	}

	lm.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"duration":  event.Duration,
	}).Info("Plugin initialized successfully")

	return nil
}

// StartPlugin 启动插件
func (lm *DefaultLifecycleManager) StartPlugin(ctx context.Context, pluginID string) error {
	startTime := time.Now()
	
	// 执行前置钩子
	if err := lm.executeHooks(ctx, "before_start", pluginID, nil); err != nil {
		return fmt.Errorf("before_start hook failed: %w", err)
	}

	// 获取当前状态
	currentState := lm.getCurrentState(pluginID)
	
	// 检查状态是否允许启动
	if currentState != StateInitialized && currentState != StateStopped {
		return fmt.Errorf("plugin %s cannot be started from state %s", pluginID, currentState.String())
	}

	// 创建生命周期事件
	event := &LifecycleEvent{
		PluginID:  pluginID,
		EventType: "start",
		FromState: currentState,
		ToState:   StateStarting,
		Timestamp: startTime,
	}

	// 更新状态
	lm.updateState(pluginID, StateStarting, currentState, nil)

	var err error
	defer func() {
		event.Duration = time.Since(startTime)
		event.Success = err == nil
		if err != nil {
			event.Error = err.Error()
			event.ToState = StateError
			lm.updateStateWithError(pluginID, StateError, err.Error())
		} else {
			event.ToState = StateRunning
			lm.updateState(pluginID, StateRunning, StateStarting, nil)
		}
		lm.addHistoryEvent(pluginID, event)
	}()

	// 执行启动
	plugin, pluginErr := lm.pluginManager.GetPlugin(pluginID)
	if pluginErr != nil {
		err = fmt.Errorf("failed to get plugin: %w", pluginErr)
		return err
	}

	if plugin == nil {
		err = fmt.Errorf("plugin %s not found", pluginID)
		return err
	}

	// 调用插件的启动方法
	if startErr := plugin.Start(ctx); startErr != nil {
		err = fmt.Errorf("plugin start failed: %w", startErr)
		return err
	}

	// 更新注册中心状态
	if regErr := lm.registry.UpdatePluginStatus(pluginID, StateRunning); regErr != nil {
		lm.logger.WithError(regErr).Warn("Failed to update plugin status in registry")
	}

	// 执行后置钩子
	if hookErr := lm.executeHooks(ctx, "after_start", pluginID, event); hookErr != nil {
		lm.logger.WithError(hookErr).Warn("after_start hook failed")
	}

	lm.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"duration":  event.Duration,
	}).Info("Plugin started successfully")

	return nil
}

// StopPlugin 停止插件
func (lm *DefaultLifecycleManager) StopPlugin(ctx context.Context, pluginID string) error {
	startTime := time.Now()
	
	// 执行前置钩子
	if err := lm.executeHooks(ctx, "before_stop", pluginID, nil); err != nil {
		return fmt.Errorf("before_stop hook failed: %w", err)
	}

	// 获取当前状态
	currentState := lm.getCurrentState(pluginID)
	
	// 检查状态是否允许停止
	if currentState != StateRunning {
		return fmt.Errorf("plugin %s cannot be stopped from state %s", pluginID, currentState.String())
	}

	// 创建生命周期事件
	event := &LifecycleEvent{
		PluginID:  pluginID,
		EventType: "stop",
		FromState: currentState,
		ToState:   StateStopping,
		Timestamp: startTime,
	}

	// 更新状态
	lm.updateState(pluginID, StateStopping, currentState, nil)

	var err error
	defer func() {
		event.Duration = time.Since(startTime)
		event.Success = err == nil
		if err != nil {
			event.Error = err.Error()
			event.ToState = StateError
			lm.updateStateWithError(pluginID, StateError, err.Error())
		} else {
			event.ToState = StateStopped
			lm.updateState(pluginID, StateStopped, StateStopping, nil)
		}
		lm.addHistoryEvent(pluginID, event)
	}()

	// 执行停止
	plugin, pluginErr := lm.pluginManager.GetPlugin(pluginID)
	if pluginErr != nil {
		err = fmt.Errorf("failed to get plugin: %w", pluginErr)
		return err
	}

	if plugin == nil {
		err = fmt.Errorf("plugin %s not found", pluginID)
		return err
	}

	// 调用插件的停止方法
	if stopErr := plugin.Stop(ctx); stopErr != nil {
		err = fmt.Errorf("plugin stop failed: %w", stopErr)
		return err
	}

	// 更新注册中心状态
	if regErr := lm.registry.UpdatePluginStatus(pluginID, StateStopped); regErr != nil {
		lm.logger.WithError(regErr).Warn("Failed to update plugin status in registry")
	}

	// 执行后置钩子
	if hookErr := lm.executeHooks(ctx, "after_stop", pluginID, event); hookErr != nil {
		lm.logger.WithError(hookErr).Warn("after_stop hook failed")
	}

	lm.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"duration":  event.Duration,
	}).Info("Plugin stopped successfully")

	return nil
}

// RestartPlugin 重启插件
func (lm *DefaultLifecycleManager) RestartPlugin(ctx context.Context, pluginID string) error {
	// 先停止插件
	if err := lm.StopPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to stop plugin for restart: %w", err)
	}

	// 等待一小段时间确保完全停止
	time.Sleep(100 * time.Millisecond)

	// 再启动插件
	if err := lm.StartPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to start plugin after restart: %w", err)
	}

	// 增加重启计数
	lm.mu.Lock()
	if state, exists := lm.states[pluginID]; exists {
		state.RestartCount++
	}
	lm.mu.Unlock()

	lm.logger.WithField("plugin_id", pluginID).Info("Plugin restarted successfully")

	return nil
}

// CleanupPlugin 清理插件
func (lm *DefaultLifecycleManager) CleanupPlugin(ctx context.Context, pluginID string) error {
	startTime := time.Now()
	
	// 执行前置钩子
	if err := lm.executeHooks(ctx, "before_cleanup", pluginID, nil); err != nil {
		return fmt.Errorf("before_cleanup hook failed: %w", err)
	}

	// 获取当前状态
	currentState := lm.getCurrentState(pluginID)
	
	// 创建生命周期事件
	event := &LifecycleEvent{
		PluginID:  pluginID,
		EventType: "cleanup",
		FromState: currentState,
		ToState:   StateUnloaded,
		Timestamp: startTime,
	}

	var err error
	defer func() {
		event.Duration = time.Since(startTime)
		event.Success = err == nil
		if err != nil {
			event.Error = err.Error()
			event.ToState = StateError
			lm.updateStateWithError(pluginID, StateError, err.Error())
		} else {
			event.ToState = StateUnloaded
			lm.removeState(pluginID)
		}
		lm.addHistoryEvent(pluginID, event)
	}()

	// 如果插件正在运行，先停止它
	if currentState == StateRunning {
		if stopErr := lm.StopPlugin(ctx, pluginID); stopErr != nil {
			lm.logger.WithError(stopErr).Warn("Failed to stop plugin before cleanup")
		}
	}

	// 执行清理
	plugin, pluginErr := lm.pluginManager.GetPlugin(pluginID)
	if pluginErr != nil {
		err = fmt.Errorf("failed to get plugin: %w", pluginErr)
		return err
	}

	if plugin != nil {
		// 调用插件的清理方法
		if cleanupErr := plugin.Cleanup(ctx); cleanupErr != nil {
			err = fmt.Errorf("plugin cleanup failed: %w", cleanupErr)
			return err
		}
	}

	// 从注册中心注销
	if regErr := lm.registry.UnregisterPlugin(pluginID); regErr != nil {
		lm.logger.WithError(regErr).Warn("Failed to unregister plugin from registry")
	}

	// 执行后置钩子
	if hookErr := lm.executeHooks(ctx, "after_cleanup", pluginID, event); hookErr != nil {
		lm.logger.WithError(hookErr).Warn("after_cleanup hook failed")
	}

	lm.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"duration":  event.Duration,
	}).Info("Plugin cleaned up successfully")

	return nil
}

// GetPluginLifecycleState 获取插件生命周期状态
func (lm *DefaultLifecycleManager) GetPluginLifecycleState(pluginID string) (*LifecycleState, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	state, exists := lm.states[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s lifecycle state not found", pluginID)
	}

	// 返回状态的副本
	stateCopy := *state
	return &stateCopy, nil
}

// SetLifecycleHook 设置生命周期钩子
func (lm *DefaultLifecycleManager) SetLifecycleHook(hook LifecycleHook) error {
	if hook.Handler == nil {
		return fmt.Errorf("hook handler cannot be nil")
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.hooks[hook.EventType] == nil {
		lm.hooks[hook.EventType] = make([]*LifecycleHook, 0)
	}

	hook.CreatedAt = time.Now()
	lm.hooks[hook.EventType] = append(lm.hooks[hook.EventType], &hook)

	lm.logger.WithFields(logrus.Fields{
		"hook_id":    hook.ID,
		"event_type": hook.EventType,
		"plugin_id":  hook.PluginID,
	}).Debug("Lifecycle hook registered")

	return nil
}

// RemoveLifecycleHook 移除生命周期钩子
func (lm *DefaultLifecycleManager) RemoveLifecycleHook(hookID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	for eventType, hooks := range lm.hooks {
		for i, hook := range hooks {
			if hook.ID == hookID {
				lm.hooks[eventType] = append(hooks[:i], hooks[i+1:]...)
				lm.logger.WithField("hook_id", hookID).Debug("Lifecycle hook removed")
				return nil
			}
		}
	}

	return fmt.Errorf("hook %s not found", hookID)
}

// GetLifecycleHistory 获取生命周期历史
func (lm *DefaultLifecycleManager) GetLifecycleHistory(pluginID string) ([]*LifecycleEvent, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	history, exists := lm.history[pluginID]
	if !exists {
		return []*LifecycleEvent{}, nil
	}

	// 返回历史的副本
	historyCopy := make([]*LifecycleEvent, len(history))
	copy(historyCopy, history)

	return historyCopy, nil
}

// 辅助方法

// getCurrentState 获取当前状态
func (lm *DefaultLifecycleManager) getCurrentState(pluginID string) PluginState {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if state, exists := lm.states[pluginID]; exists {
		return state.CurrentState
	}

	return StateUnloaded
}

// updateState 更新状态
func (lm *DefaultLifecycleManager) updateState(
	pluginID string,
	newState PluginState,
	previousState PluginState,
	config map[string]interface{},
) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	state, exists := lm.states[pluginID]
	if !exists {
		state = &LifecycleState{
			PluginID: pluginID,
		}
		lm.states[pluginID] = state
	}

	state.PreviousState = previousState
	state.CurrentState = newState
	state.StateChangedAt = now
	state.LastError = ""

	if config != nil {
		state.Configuration = config
	}

	// 更新特定状态的时间戳
	switch newState {
	case StateInitialized:
		state.InitializedAt = &now
	case StateRunning:
		state.StartedAt = &now
	case StateStopped:
		state.StoppedAt = &now
	}
}

// updateStateWithError 更新状态并设置错误
func (lm *DefaultLifecycleManager) updateStateWithError(pluginID string, newState PluginState, errorMsg string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	state, exists := lm.states[pluginID]
	if !exists {
		state = &LifecycleState{
			PluginID: pluginID,
		}
		lm.states[pluginID] = state
	}

	state.PreviousState = state.CurrentState
	state.CurrentState = newState
	state.StateChangedAt = now
	state.LastError = errorMsg
}

// removeState 移除状态
func (lm *DefaultLifecycleManager) removeState(pluginID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	delete(lm.states, pluginID)
}

// addHistoryEvent 添加历史事件
func (lm *DefaultLifecycleManager) addHistoryEvent(pluginID string, event *LifecycleEvent) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.history[pluginID] == nil {
		lm.history[pluginID] = make([]*LifecycleEvent, 0)
	}

	lm.history[pluginID] = append(lm.history[pluginID], event)

	// 限制历史记录数量，保留最近的100条
	if len(lm.history[pluginID]) > 100 {
		lm.history[pluginID] = lm.history[pluginID][1:]
	}
}

// executeHooks 执行钩子
func (lm *DefaultLifecycleManager) executeHooks(
	ctx context.Context,
	eventType string,
	pluginID string,
	event *LifecycleEvent,
) error {
	lm.mu.RLock()
	hooks := lm.hooks[eventType]
	lm.mu.RUnlock()

	if len(hooks) == 0 {
		return nil
	}

	// 按优先级排序执行钩子
	for _, hook := range hooks {
		if !hook.Enabled {
			continue
		}

		// 检查钩子是否适用于当前插件
		if hook.PluginID != "" && hook.PluginID != pluginID {
			continue
		}

		// 执行钩子
		if err := hook.Handler(ctx, event); err != nil {
			lm.logger.WithFields(logrus.Fields{
				"hook_id":    hook.ID,
				"event_type": eventType,
				"plugin_id":  pluginID,
				"error":      err,
			}).Error("Lifecycle hook execution failed")
			return err
		}
	}

	return nil
}