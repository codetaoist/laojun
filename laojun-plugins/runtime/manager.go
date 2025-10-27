package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PluginManager 插件管理器接口
type PluginManager interface {
	// Start 启动管理器
	Start(ctx context.Context) error

	// Stop 停止管理器
	Stop(ctx context.Context) error

	// LoadPlugin 加载插件
	LoadPlugin(ctx context.Context, pluginPath string) error

	// UnloadPlugin 卸载插件
	UnloadPlugin(ctx context.Context, pluginID string) error

	// StartPlugin 启动插件
	StartPlugin(ctx context.Context, pluginID string) error

	// StopPlugin 停止插件
	StopPlugin(ctx context.Context, pluginID string) error

	// GetPlugin 获取插件实例
	GetPlugin(pluginID string) (Plugin, error)

	// GetPluginInfo 获取插件信息
	GetPluginInfo(pluginID string) (*PluginInfo, error)

	// ListPlugins 列出所有插件
	ListPlugins() map[string]*PluginInfo

	// SendEvent 发送事件到插件
	SendEvent(ctx context.Context, pluginID string, event Event) error

	// BroadcastEvent 广播事件到所有插件
	BroadcastEvent(ctx context.Context, event Event) error

	// GetResourceUsage 获取资源使用情况
	GetResourceUsage(pluginID string) (*ResourceUsage, error)
}

// DefaultPluginManager 默认插件管理器实现
type DefaultPluginManager struct {
	plugins     map[string]Plugin
	pluginInfos map[string]*PluginInfo
	loader      PluginLoader
	sandbox     Sandbox
	logger      *logrus.Logger
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
}

// NewDefaultPluginManager 创建默认插件管理器
func NewDefaultPluginManager(loader PluginLoader, sandbox Sandbox, logger *logrus.Logger) *DefaultPluginManager {
	manager := &DefaultPluginManager{
		plugins:     make(map[string]Plugin),
		pluginInfos: make(map[string]*PluginInfo),
		loader:      loader,
		sandbox:     sandbox,
		logger:      logger,
		running:     false,
	}

	return manager
}

// Start 启动管理器
func (m *DefaultPluginManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("plugin manager already started")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true

	// 启动资源监控
	m.startResourceMonitoring()

	m.logger.Info("Plugin manager started")
	return nil
}

// Stop 停止管理器
func (m *DefaultPluginManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.logger.Info("Stopping plugin manager")

	// 停止所有插件
	for pluginID := range m.plugins {
		if err := m.StopPlugin(ctx, pluginID); err != nil {
			m.logger.WithError(err).WithField("plugin_id", pluginID).Error("Failed to stop plugin")
		}
	}

	// 取消上下文
	if m.cancel != nil {
		m.cancel()
	}

	// 等待所有goroutine结束
	m.wg.Wait()

	m.running = false
	m.logger.Info("Plugin manager stopped")
	return nil
}

// LoadPlugin 加载插件
func (m *DefaultPluginManager) LoadPlugin(ctx context.Context, pluginPath string) error {
	m.logger.WithField("path", pluginPath).Info("Loading plugin")

	// 使用加载器加载插件
	plugin, err := m.loader.LoadPlugin(ctx, pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	metadata := plugin.GetMetadata()
	pluginID := metadata.ID

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查插件是否已存在
	if _, exists := m.plugins[pluginID]; exists {
		return fmt.Errorf("plugin %s already loaded", pluginID)
	}

	// 在沙箱中初始化插件
	if err := m.sandbox.InitializePlugin(ctx, plugin); err != nil {
		return fmt.Errorf("failed to initialize plugin in sandbox: %w", err)
	}

	// 存储插件信息
	now := time.Now()
	m.plugins[pluginID] = plugin
	m.pluginInfos[pluginID] = &PluginInfo{
		Metadata: metadata,
		State:    StateLoaded,
		LoadedAt: &now,
		ResourceUsage: ResourceUsage{
			LastUpdated: now,
		},
	}

	m.logger.WithField("plugin_id", pluginID).Info("Plugin loaded successfully")
	return nil
}

// UnloadPlugin 卸载插件
func (m *DefaultPluginManager) UnloadPlugin(ctx context.Context, pluginID string) error {
	m.logger.WithField("plugin_id", pluginID).Info("Unloading plugin")

	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	info := m.pluginInfos[pluginID]

	// 如果插件正在运行，先停止它
	if info.State == StateRunning {
		if err := plugin.Stop(ctx); err != nil {
			m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to stop plugin during unload")
		}
	}

	// 清理插件资源
	if err := plugin.Cleanup(ctx); err != nil {
		m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to cleanup plugin")
	}

	// 从沙箱中移除插件
	if err := m.sandbox.RemovePlugin(ctx, pluginID); err != nil {
		m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to remove plugin from sandbox")
	}

	// 移除插件
	delete(m.plugins, pluginID)
	delete(m.pluginInfos, pluginID)

	m.logger.WithField("plugin_id", pluginID).Info("Plugin unloaded successfully")
	return nil
}

// StartPlugin 启动插件
func (m *DefaultPluginManager) StartPlugin(ctx context.Context, pluginID string) error {
	m.logger.WithField("plugin_id", pluginID).Info("Starting plugin")

	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	info := m.pluginInfos[pluginID]
	if info.State == StateRunning {
		return fmt.Errorf("plugin %s is already running", pluginID)
	}

	// 启动插件
	if err := plugin.Start(ctx); err != nil {
		info.State = StateError
		info.ErrorMsg = err.Error()
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// 更新状态
	now := time.Now()
	info.State = StateRunning
	info.StartedAt = &now
	info.ErrorMsg = ""

	m.logger.WithField("plugin_id", pluginID).Info("Plugin started successfully")
	return nil
}

// StopPlugin 停止插件
func (m *DefaultPluginManager) StopPlugin(ctx context.Context, pluginID string) error {
	m.logger.WithField("plugin_id", pluginID).Info("Stopping plugin")

	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	info := m.pluginInfos[pluginID]
	if info.State != StateRunning {
		return fmt.Errorf("plugin %s is not running", pluginID)
	}

	// 停止插件
	if err := plugin.Stop(ctx); err != nil {
		info.State = StateError
		info.ErrorMsg = err.Error()
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	// 更新状态
	now := time.Now()
	info.State = StateStopped
	info.StoppedAt = &now
	info.ErrorMsg = ""

	m.logger.WithField("plugin_id", pluginID).Info("Plugin stopped successfully")
	return nil
}

// GetPlugin 获取插件实例
func (m *DefaultPluginManager) GetPlugin(pluginID string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return plugin, nil
}

// GetPluginInfo 获取插件信息
func (m *DefaultPluginManager) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.pluginInfos[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	// 返回副本以避免并发修改
	infoCopy := *info
	return &infoCopy, nil
}

// ListPlugins 列出所有插件
func (m *DefaultPluginManager) ListPlugins() map[string]*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*PluginInfo)
	for id, info := range m.pluginInfos {
		infoCopy := *info
		result[id] = &infoCopy
	}

	return result
}

// SendEvent 发送事件到插件
func (m *DefaultPluginManager) SendEvent(ctx context.Context, pluginID string, event Event) error {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return err
	}

	return plugin.HandleEvent(ctx, event)
}

// BroadcastEvent 广播事件到所有插件
func (m *DefaultPluginManager) BroadcastEvent(ctx context.Context, event Event) error {
	m.mu.RLock()
	plugins := make(map[string]Plugin)
	for id, plugin := range m.plugins {
		plugins[id] = plugin
	}
	m.mu.RUnlock()

	var errors []error
	for pluginID, plugin := range plugins {
		if err := plugin.HandleEvent(ctx, event); err != nil {
			m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to send event to plugin")
			errors = append(errors, fmt.Errorf("plugin %s: %w", pluginID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send event to %d plugins", len(errors))
	}

	return nil
}

// GetResourceUsage 获取资源使用情况
func (m *DefaultPluginManager) GetResourceUsage(pluginID string) (*ResourceUsage, error) {
	info, err := m.GetPluginInfo(pluginID)
	if err != nil {
		return nil, err
	}

	return &info.ResourceUsage, nil
}

// startResourceMonitoring 启动资源监控
func (m *DefaultPluginManager) startResourceMonitoring() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.updateResourceUsage()
			}
		}
	}()
}

// updateResourceUsage 更新资源使用情况
func (m *DefaultPluginManager) updateResourceUsage() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, info := range m.pluginInfos {
		if info.State == StateRunning {
			// 这里应该实际测量资源使用情况
			// 为了简化，我们使用模拟数据
			info.ResourceUsage = ResourceUsage{
				CPUPercent:     0.0, // 实际实现中应该测量真实的CPU使用率
				MemoryBytes:    0,   // 实际实现中应该测量真实的内存使用量
				GoroutineCount: 0,   // 实际实现中应该测量真实的goroutine数量
				LastUpdated:    time.Now(),
			}
		}
	}
}

// Shutdown 关闭管理器
func (m *DefaultPluginManager) Shutdown(ctx context.Context) error {
	m.logger.Info("Shutting down plugin manager")

	// 停止所有插件
	m.mu.RLock()
	pluginIDs := make([]string, 0, len(m.plugins))
	for id := range m.plugins {
		pluginIDs = append(pluginIDs, id)
	}
	m.mu.RUnlock()

	for _, pluginID := range pluginIDs {
		if err := m.StopPlugin(ctx, pluginID); err != nil {
			m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to stop plugin during shutdown")
		}
		if err := m.UnloadPlugin(ctx, pluginID); err != nil {
			m.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to unload plugin during shutdown")
		}
	}

	// 停止监控
	m.cancel()
	m.wg.Wait()

	m.logger.Info("Plugin manager shutdown completed")
	return nil
}
