package sync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun-plugins/shared/models"
)

// ExampleUsage 展示数据同步机制的使用方法
func ExampleUsage() {
	// 创建日志器
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 创建事件总线
	eventBus := NewDefaultEventBus(1000, 4, logger)

	// 创建存储
	storage := NewDefaultSyncStorage()

	// 创建数据同步管理器
	syncManager := NewDataSyncManager(eventBus, storage, logger)

	// 注册数据转换器
	syncManager.RegisterTransformer(&MarketToRuntimeTransformer{})
	syncManager.RegisterTransformer(&RuntimeToMarketTransformer{})

	// 创建示例订阅者
	subscriber := &ExampleSyncSubscriber{logger: logger}
	syncManager.RegisterSubscriber(subscriber)

	// 启动同步管理器
	ctx := context.Background()
	if err := syncManager.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start sync manager")
	}

	// 模拟插件数据同步
	pluginID := uuid.New()
	changes := map[string]interface{}{
		"name":        "Example Plugin",
		"version":     "1.0.0",
		"description": "This is an example plugin",
		"state":       "active",
	}

	// 同步插件元数据更新
	if err := syncManager.SyncPluginData(ctx, pluginID, "metadata_updated", changes, "market"); err != nil {
		logger.WithError(err).Error("Failed to sync plugin data")
	}

	// 同步插件状态变更
	stateChanges := map[string]interface{}{
		"state":      "stopped",
		"error_msg":  "",
		"stopped_at": time.Now(),
	}

	if err := syncManager.SyncPluginData(ctx, pluginID, "state_changed", stateChanges, "runtime"); err != nil {
		logger.WithError(err).Error("Failed to sync plugin state")
	}

	// 等待一段时间让事件处理完成
	time.Sleep(2 * time.Second)

	// 查询同步历史
	history, err := syncManager.GetSyncHistory(ctx, pluginID, time.Now().Add(-1*time.Hour))
	if err != nil {
		logger.WithError(err).Error("Failed to get sync history")
	} else {
		logger.WithField("history_count", len(history)).Info("Retrieved sync history")
	}

	// 更新同步状态
	if err := syncManager.UpdateSyncStatus(ctx, "market", "runtime", "success", "", 1); err != nil {
		logger.WithError(err).Error("Failed to update sync status")
	}

	// 查询同步状态
	status, err := syncManager.GetSyncStatus(ctx, "market", "runtime")
	if err != nil {
		logger.WithError(err).Error("Failed to get sync status")
	} else {
		logger.WithFields(logrus.Fields{
			"source":       status.Source,
			"target":       status.Target,
			"status":       status.Status,
			"last_sync":    status.LastSync,
			"record_count": status.RecordCount,
		}).Info("Retrieved sync status")
	}

	// 停止同步管理器
	if err := syncManager.Stop(ctx); err != nil {
		logger.WithError(err).Error("Failed to stop sync manager")
	}

	logger.Info("Example usage completed")
}

// ExampleSyncSubscriber 示例同步订阅者
type ExampleSyncSubscriber struct {
	logger *logrus.Logger
}

// OnDataSync 处理数据同步事件
func (s *ExampleSyncSubscriber) OnDataSync(ctx context.Context, event *models.PluginSyncEvent) error {
	s.logger.WithFields(logrus.Fields{
		"plugin_id":  event.PluginID,
		"event_type": event.EventType,
		"source":     event.Source,
		"timestamp":  event.Timestamp,
		"version":    event.Version,
	}).Info("Received sync event")

	// 根据事件类型处理不同的同步逻辑
	switch event.EventType {
	case "metadata_updated":
		return s.handleMetadataUpdate(ctx, event)
	case "state_changed":
		return s.handleStateChange(ctx, event)
	case "installed":
		return s.handlePluginInstalled(ctx, event)
	case "uninstalled":
		return s.handlePluginUninstalled(ctx, event)
	default:
		s.logger.WithField("event_type", event.EventType).Warn("Unknown sync event type")
	}

	return nil
}

// GetSubscriptionTypes 返回订阅的事件类型
func (s *ExampleSyncSubscriber) GetSubscriptionTypes() []string {
	return []string{
		"metadata_updated",
		"state_changed",
		"installed",
		"uninstalled",
	}
}

// handleMetadataUpdate 处理元数据更新
func (s *ExampleSyncSubscriber) handleMetadataUpdate(ctx context.Context, event *models.PluginSyncEvent) error {
	s.logger.WithField("plugin_id", event.PluginID).Info("Processing metadata update")

	// 这里可以实现具体的元数据同步逻辑
	// 例如：更新数据库、通知其他服务等

	return nil
}

// handleStateChange 处理状态变更
func (s *ExampleSyncSubscriber) handleStateChange(ctx context.Context, event *models.PluginSyncEvent) error {
	s.logger.WithField("plugin_id", event.PluginID).Info("Processing state change")

	// 这里可以实现具体的状态同步逻辑
	// 例如：更新运行时状态、发送通知等

	return nil
}

// handlePluginInstalled 处理插件安装
func (s *ExampleSyncSubscriber) handlePluginInstalled(ctx context.Context, event *models.PluginSyncEvent) error {
	s.logger.WithField("plugin_id", event.PluginID).Info("Processing plugin installation")

	// 这里可以实现插件安装后的同步逻辑
	// 例如：更新安装统计、发送安装通知等

	return nil
}

// handlePluginUninstalled 处理插件卸载
func (s *ExampleSyncSubscriber) handlePluginUninstalled(ctx context.Context, event *models.PluginSyncEvent) error {
	s.logger.WithField("plugin_id", event.PluginID).Info("Processing plugin uninstallation")

	// 这里可以实现插件卸载后的同步逻辑
	// 例如：清理相关数据、更新统计等

	return nil
}

// ExampleDataTransformation 展示数据转换的使用
func ExampleDataTransformation() {
	logger := logrus.New()

	// 创建市场插件数据
	marketPlugin := &models.UnifiedPluginMetadata{
		ID:          uuid.New(),
		Name:        "Example Plugin",
		Version:     "1.0.0",
		Description: "This is an example plugin for demonstration",
		Author:      "Example Author",
		Category:    "utility",
		Type:        "service",
		Tags:        []string{"example", "demo", "utility"},
		Permissions: []string{"read", "write"},
		Dependencies: []string{
			"golang>=1.21",
		},
		Config: map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 创建转换器
	transformer := &MarketToRuntimeTransformer{}

	// 执行转换
	ctx := context.Background()
	runtimeData, err := transformer.Transform(ctx, marketPlugin)
	if err != nil {
		logger.WithError(err).Error("Failed to transform data")
		return
	}

	// 输出转换结果
	logger.WithField("runtime_data", runtimeData).Info("Data transformation completed")

	// 反向转换示例
	runtimeToMarket := &RuntimeToMarketTransformer{}
	runtimeInput := map[string]interface{}{
		"state":          "running",
		"resource_usage": map[string]interface{}{"cpu": 0.5, "memory": 128},
		"error_msg":      "",
	}

	marketData, err := runtimeToMarket.Transform(ctx, runtimeInput)
	if err != nil {
		logger.WithError(err).Error("Failed to transform runtime data")
		return
	}

	logger.WithField("market_data", marketData).Info("Runtime to market transformation completed")
}

// ExampleEventFiltering 展示事件过滤的使用
func ExampleEventFiltering() {
	logger := logrus.New()

	// 创建事件过滤器
	filter := &EventFilter{
		EventTypes:  []string{"plugin.sync.*", "plugin.state.*"},
		Sources:     []string{"market", "runtime"},
		MinPriority: 1,
		MaxAge:      time.Hour,
	}

	// 创建带过滤器的事件总线
	filteredBus := NewFilteredEventBus(1000, 2, filter, logger)

	// 启动事件总线
	ctx := context.Background()
	if err := filteredBus.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start filtered event bus")
	}

	// 订阅事件
	filteredBus.Subscribe("plugin.sync.*", func(ctx context.Context, event *models.UnifiedPluginEvent) error {
		logger.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.Type,
			"source":     event.Source,
		}).Info("Received filtered event")
		return nil
	})

	// 发布测试事件
	testEvent := &models.UnifiedPluginEvent{
		ID:        uuid.New(),
		Type:      "plugin.sync.test",
		Source:    "market",
		Priority:  2,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"test": "data"},
	}

	if err := filteredBus.Publish(ctx, testEvent); err != nil {
		logger.WithError(err).Error("Failed to publish test event")
	}

	// 等待事件处理
	time.Sleep(1 * time.Second)

	// 停止事件总线
	if err := filteredBus.Stop(ctx); err != nil {
		logger.WithError(err).Error("Failed to stop filtered event bus")
	}

	logger.Info("Event filtering example completed")
}

// RunAllExamples 运行所有示例
func RunAllExamples() {
	fmt.Println("=== Running Data Sync Example ===")
	ExampleUsage()

	fmt.Println("\n=== Running Data Transformation Example ===")
	ExampleDataTransformation()

	fmt.Println("\n=== Running Event Filtering Example ===")
	ExampleEventFiltering()

	fmt.Println("\n=== All Examples Completed ===")
}