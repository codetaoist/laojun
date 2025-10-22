package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/taishanglaojun/plugins/core"
)

// InProcessRuntime 进程内插件运行时管理
type InProcessRuntime struct {
	goRuntime   *GoPluginRuntime
	jsRuntime   *JSRuntime
	config      *InProcessConfig
	securityMgr core.SecurityManager
	eventBus    core.EventBus
	registry    core.PluginRegistry
	mutex       sync.RWMutex
	stats       *RuntimeStats
}

// InProcessConfig 进程内运行时配置
type InProcessConfig struct {
	GoConfig      *GoRuntimeConfig `json:"go_config"`
	JSConfig      *JSRuntimeConfig `json:"js_config"`
	EnableStats   bool             `json:"enable_stats"`
	StatsInterval time.Duration    `json:"stats_interval"`
	MaxPlugins    int              `json:"max_plugins"`
	PluginDir     string           `json:"plugin_dir"`
}

// RuntimeStats 运行时统计信息
type RuntimeStats struct {
	TotalPlugins   int                    `json:"total_plugins"`
	GoPlugins      int                    `json:"go_plugins"`
	JSPlugins      int                    `json:"js_plugins"`
	RunningPlugins int                    `json:"running_plugins"`
	LoadedPlugins  int                    `json:"loaded_plugins"`
	FailedPlugins  int                    `json:"failed_plugins"`
	TotalRequests  int64                  `json:"total_requests"`
	TotalErrors    int64                  `json:"total_errors"`
	AverageLatency time.Duration          `json:"average_latency"`
	MemoryUsage    int64                  `json:"memory_usage"`
	LastUpdate     time.Time              `json:"last_update"`
	PluginStats    map[string]interface{} `json:"plugin_stats"`
	mutex          sync.RWMutex
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Version    string             `json:"version"`
	Type       core.PluginType    `json:"type"`
	Runtime    core.PluginRuntime `json:"runtime"`
	Status     core.PluginStatus  `json:"status"`
	LoadTime   time.Time          `json:"load_time"`
	LastAccess time.Time          `json:"last_access"`
	Config     map[string]any     `json:"config"`
	Stats      interface{}        `json:"stats"`
}

// NewInProcessRuntime 创建进程内运行时管理
func NewInProcessRuntime(
	config *InProcessConfig,
	securityMgr core.SecurityManager,
	eventBus core.EventBus,
	registry core.PluginRegistry,
) *InProcessRuntime {
	runtime := &InProcessRuntime{
		config:      config,
		securityMgr: securityMgr,
		eventBus:    eventBus,
		registry:    registry,
		stats: &RuntimeStats{
			PluginStats: make(map[string]interface{}),
			LastUpdate:  time.Now(),
		},
	}

	// 初始化Go运行时环境
	if config.GoConfig != nil {
		runtime.goRuntime = NewGoPluginRuntime(config.GoConfig, securityMgr)
	}

	// 初始化JavaScript运行时环境
	if config.JSConfig != nil {
		runtime.jsRuntime = NewJSRuntime(config.JSConfig, securityMgr)
	}

	// 启动统计收集
	if config.EnableStats {
		go runtime.collectStats()
	}

	return runtime
}

// LoadPlugin 加载插件
func (r *InProcessRuntime) LoadPlugin(ctx context.Context, pluginPath string, config map[string]any) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查插件总数限制
	if r.getTotalPluginCount() >= r.config.MaxPlugins {
		return fmt.Errorf("maximum plugin limit reached: %d", r.config.MaxPlugins)
	}

	// 根据文件扩展名确定插件类型
	ext := strings.ToLower(filepath.Ext(pluginPath))

	switch ext {
	case ".so":
		return r.loadGoPlugin(ctx, pluginPath, config)
	case ".js":
		return r.loadJSPlugin(ctx, pluginPath, config)
	default:
		return fmt.Errorf("unsupported plugin type: %s", ext)
	}
}

// loadGoPlugin 加载Go插件
func (r *InProcessRuntime) loadGoPlugin(ctx context.Context, pluginPath string, config map[string]any) error {
	if r.goRuntime == nil {
		return fmt.Errorf("Go runtime not initialized")
	}

	instance, err := r.goRuntime.LoadPlugin(ctx, pluginPath, config)
	if err != nil {
		r.stats.FailedPlugins++
		r.publishEvent(ctx, "plugin.load.failed", map[string]any{
			"path":  pluginPath,
			"error": err.Error(),
			"type":  "go",
		})
		return fmt.Errorf("failed to load Go plugin: %w", err)
	}

	r.stats.GoPlugins++
	r.stats.LoadedPlugins++
	r.stats.TotalPlugins++

	// 注册到插件注册表
	if err := r.registry.Register(ctx, instance.Metadata); err != nil {
		// 记录警告但不失败
	}

	// 发布加载成功事件
	r.publishEvent(ctx, "plugin.loaded", map[string]any{
		"id":      instance.ID,
		"name":    instance.Metadata.Name,
		"version": instance.Metadata.Version,
		"type":    "go",
	})

	return nil
}

// loadJSPlugin 加载JavaScript插件
func (r *InProcessRuntime) loadJSPlugin(ctx context.Context, pluginPath string, config map[string]any) error {
	if r.jsRuntime == nil {
		return fmt.Errorf("JavaScript runtime not initialized")
	}

	instance, err := r.jsRuntime.LoadPlugin(ctx, pluginPath, config)
	if err != nil {
		r.stats.FailedPlugins++
		r.publishEvent(ctx, "plugin.load.failed", map[string]any{
			"path":  pluginPath,
			"error": err.Error(),
			"type":  "js",
		})
		return fmt.Errorf("failed to load JavaScript plugin: %w", err)
	}

	r.stats.JSPlugins++
	r.stats.LoadedPlugins++
	r.stats.TotalPlugins++

	// 注册到插件注册表
	if err := r.registry.Register(ctx, instance.Metadata); err != nil {
		// 记录警告但不失败
	}

	// 发布加载成功事件
	r.publishEvent(ctx, "plugin.loaded", map[string]any{
		"id":      instance.ID,
		"name":    instance.Metadata.Name,
		"version": instance.Metadata.Version,
		"type":    "js",
	})

	return nil
}

// UnloadPlugin 卸载插件
func (r *InProcessRuntime) UnloadPlugin(ctx context.Context, pluginID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 尝试从Go运行时卸载插件
	if r.goRuntime != nil {
		if _, err := r.goRuntime.GetPlugin(pluginID); err == nil {
			if err := r.goRuntime.UnloadPlugin(ctx, pluginID); err != nil {
				return fmt.Errorf("failed to unload Go plugin: %w", err)
			}
			r.stats.GoPlugins--
			r.stats.TotalPlugins--
			r.publishUnloadEvent(ctx, pluginID, "go")
			return nil
		}
	}

	// 尝试从JavaScript运行时卸载插件
	if r.jsRuntime != nil {
		if _, err := r.jsRuntime.GetPlugin(pluginID); err == nil {
			if err := r.jsRuntime.UnloadPlugin(ctx, pluginID); err != nil {
				return fmt.Errorf("failed to unload JavaScript plugin: %w", err)
			}
			r.stats.JSPlugins--
			r.stats.TotalPlugins--
			r.publishUnloadEvent(ctx, pluginID, "js")
			return nil
		}
	}

	return fmt.Errorf("plugin not found: %s", pluginID)
}

// ExecutePlugin 执行插件请求
func (r *InProcessRuntime) ExecutePlugin(ctx context.Context, pluginID string, req *core.PluginRequest) (*core.PluginResponse, error) {
	startTime := time.Now()

	// 尝试从Go运行时执行插件
	if r.goRuntime != nil {
		if _, err := r.goRuntime.GetPlugin(pluginID); err == nil {
			response, err := r.goRuntime.ExecutePlugin(ctx, pluginID, req)
			r.updateExecutionStats(time.Since(startTime), err)
			return response, err
		}
	}

	// 尝试从JavaScript运行时执行插件
	if r.jsRuntime != nil {
		if _, err := r.jsRuntime.GetPlugin(pluginID); err == nil {
			response, err := r.jsRuntime.ExecutePlugin(ctx, pluginID, req)
			r.updateExecutionStats(time.Since(startTime), err)
			return response, err
		}
	}

	return nil, fmt.Errorf("plugin not found: %s", pluginID)
}

// StartPlugin 启动插件
func (r *InProcessRuntime) StartPlugin(ctx context.Context, pluginID string) error {
	// 尝试启动Go插件
	if r.goRuntime != nil {
		if instance, err := r.goRuntime.GetPlugin(pluginID); err == nil {
			if err := instance.Start(ctx); err != nil {
				return fmt.Errorf("failed to start Go plugin: %w", err)
			}
			r.stats.RunningPlugins++
			r.publishEvent(ctx, "plugin.started", map[string]any{
				"id":   pluginID,
				"type": "go",
			})
			return nil
		}
	}

	// 尝试启动JavaScript插件
	if r.jsRuntime != nil {
		if instance, err := r.jsRuntime.GetPlugin(pluginID); err == nil {
			instance.mutex.Lock()
			instance.Status = core.StatusRunning
			instance.mutex.Unlock()
			r.stats.RunningPlugins++
			r.publishEvent(ctx, "plugin.started", map[string]any{
				"id":   pluginID,
				"type": "js",
			})
			return nil
		}
	}

	return fmt.Errorf("plugin not found: %s", pluginID)
}

// StopPlugin 停止插件
func (r *InProcessRuntime) StopPlugin(ctx context.Context, pluginID string) error {
	// 尝试停止Go插件
	if r.goRuntime != nil {
		if instance, err := r.goRuntime.GetPlugin(pluginID); err == nil {
			if err := instance.Stop(ctx); err != nil {
				return fmt.Errorf("failed to stop Go plugin: %w", err)
			}
			r.stats.RunningPlugins--
			r.publishEvent(ctx, "plugin.stopped", map[string]any{
				"id":   pluginID,
				"type": "go",
			})
			return nil
		}
	}

	// 尝试停止JavaScript插件
	if r.jsRuntime != nil {
		if instance, err := r.jsRuntime.GetPlugin(pluginID); err == nil {
			instance.mutex.Lock()
			instance.Status = core.StatusStopped
			instance.mutex.Unlock()
			r.stats.RunningPlugins--
			r.publishEvent(ctx, "plugin.stopped", map[string]any{
				"id":   pluginID,
				"type": "js",
			})
			return nil
		}
	}

	return fmt.Errorf("plugin not found: %s", pluginID)
}

// ListPlugins 列出所有插件
func (r *InProcessRuntime) ListPlugins() []*PluginInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var plugins []*PluginInfo

	// 获取Go插件
	if r.goRuntime != nil {
		goPlugins := r.goRuntime.ListPlugins()
		for _, plugin := range goPlugins {
			plugins = append(plugins, &PluginInfo{
				ID:         plugin.ID,
				Name:       plugin.Metadata.Name,
				Version:    plugin.Metadata.Version,
				Type:       plugin.Metadata.Type,
				Runtime:    core.RuntimeGo,
				Status:     plugin.Status,
				LoadTime:   plugin.LoadTime,
				LastAccess: plugin.LastAccess,
				Config:     plugin.Config,
				Stats:      plugin.GetStats(),
			})
		}
	}

	// 获取JavaScript插件
	if r.jsRuntime != nil {
		for _, instance := range r.jsRuntime.instances {
			plugins = append(plugins, &PluginInfo{
				ID:         instance.ID,
				Name:       instance.Metadata.Name,
				Version:    instance.Metadata.Version,
				Type:       instance.Metadata.Type,
				Runtime:    core.RuntimeJS,
				Status:     instance.Status,
				LoadTime:   instance.LoadTime,
				LastAccess: instance.LastAccess,
				Config:     instance.Config,
				Stats:      instance.GetStats(),
			})
		}
	}

	return plugins
}

// GetPlugin 获取插件信息
func (r *InProcessRuntime) GetPlugin(pluginID string) (*PluginInfo, error) {
	plugins := r.ListPlugins()
	for _, plugin := range plugins {
		if plugin.ID == pluginID {
			return plugin, nil
		}
	}
	return nil, fmt.Errorf("plugin not found: %s", pluginID)
}

// GetStats 获取运行时统计信息
func (r *InProcessRuntime) GetStats() *RuntimeStats {
	r.stats.mutex.RLock()
	defer r.stats.mutex.RUnlock()

	// 创建副本以避免并发访问问题
	stats := &RuntimeStats{
		TotalPlugins:   r.stats.TotalPlugins,
		GoPlugins:      r.stats.GoPlugins,
		JSPlugins:      r.stats.JSPlugins,
		RunningPlugins: r.stats.RunningPlugins,
		LoadedPlugins:  r.stats.LoadedPlugins,
		FailedPlugins:  r.stats.FailedPlugins,
		TotalRequests:  r.stats.TotalRequests,
		TotalErrors:    r.stats.TotalErrors,
		AverageLatency: r.stats.AverageLatency,
		MemoryUsage:    r.stats.MemoryUsage,
		LastUpdate:     r.stats.LastUpdate,
		PluginStats:    make(map[string]interface{}),
	}

	// 复制插件统计信息
	for k, v := range r.stats.PluginStats {
		stats.PluginStats[k] = v
	}

	return stats
}

// 辅助方法

// getTotalPluginCount 获取插件总数
func (r *InProcessRuntime) getTotalPluginCount() int {
	count := 0
	if r.goRuntime != nil {
		count += len(r.goRuntime.plugins)
	}
	if r.jsRuntime != nil {
		count += len(r.jsRuntime.instances)
	}
	return count
}

// updateExecutionStats 更新执行统计信息
func (r *InProcessRuntime) updateExecutionStats(duration time.Duration, err error) {
	r.stats.mutex.Lock()
	defer r.stats.mutex.Unlock()

	r.stats.TotalRequests++
	if err != nil {
		r.stats.TotalErrors++
	}

	// 更新平均延迟
	if r.stats.TotalRequests > 0 {
		totalDuration := r.stats.AverageLatency * time.Duration(r.stats.TotalRequests-1)
		r.stats.AverageLatency = (totalDuration + duration) / time.Duration(r.stats.TotalRequests)
	}
}

// publishEvent 发布事件
func (r *InProcessRuntime) publishEvent(ctx context.Context, eventType string, data map[string]any) {
	if r.eventBus != nil {
		event := &core.PluginEvent{
			Type:      eventType,
			Source:    "inprocess-runtime",
			Data:      data,
			Timestamp: time.Now(),
		}
		r.eventBus.Publish(ctx, event)
	}
}

// publishUnloadEvent 发布卸载事件
func (r *InProcessRuntime) publishUnloadEvent(ctx context.Context, pluginID, pluginType string) {
	r.publishEvent(ctx, "plugin.unloaded", map[string]any{
		"id":   pluginID,
		"type": pluginType,
	})
}

// collectStats 收集统计信息
func (r *InProcessRuntime) collectStats() {
	ticker := time.NewTicker(r.config.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.updateStats()
		}
	}
}

// updateStats 更新统计信息
func (r *InProcessRuntime) updateStats() {
	r.stats.mutex.Lock()
	defer r.stats.mutex.Unlock()

	// 更新插件计数
	r.stats.TotalPlugins = r.getTotalPluginCount()
	r.stats.LastUpdate = time.Now()

	// 收集各插件的详细统计信息
	plugins := r.ListPlugins()
	for _, plugin := range plugins {
		r.stats.PluginStats[plugin.ID] = plugin.Stats
	}
}

// Stop 停止运行时环境
func (r *InProcessRuntime) Stop(ctx context.Context) error {
	var errors []error

	// 停止Go运行时环境
	if r.goRuntime != nil {
		if err := r.goRuntime.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop Go runtime: %w", err))
		}
	}

	// 停止JavaScript运行时环境
	if r.jsRuntime != nil {
		if err := r.jsRuntime.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop JavaScript runtime: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping runtime: %v", errors)
	}

	return nil
}
