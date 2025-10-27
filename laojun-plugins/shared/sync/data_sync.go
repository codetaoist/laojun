package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun-plugins/shared/models"
)

// DataSyncManager 数据同步管理器
type DataSyncManager struct {
	eventBus     EventBus
	storage      SyncStorage
	transformers map[string]DataTransformer
	subscribers  map[string][]SyncSubscriber
	logger       *logrus.Logger
	mu           sync.RWMutex
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(ctx context.Context, event *models.UnifiedPluginEvent) error
	Subscribe(eventType string, handler func(ctx context.Context, event *models.UnifiedPluginEvent) error) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// SyncStorage 同步存储接口
type SyncStorage interface {
	SaveSyncEvent(ctx context.Context, event *models.PluginSyncEvent) error
	GetSyncEvents(ctx context.Context, pluginID uuid.UUID, since time.Time) ([]*models.PluginSyncEvent, error)
	UpdateSyncStatus(ctx context.Context, status *models.DataSyncStatus) error
	GetSyncStatus(ctx context.Context, source, target string) (*models.DataSyncStatus, error)
}

// DataTransformer 数据转换器接口
type DataTransformer interface {
	Transform(ctx context.Context, data interface{}) (interface{}, error)
	GetSourceType() string
	GetTargetType() string
}

// SyncSubscriber 同步订阅者接口
type SyncSubscriber interface {
	OnDataSync(ctx context.Context, event *models.PluginSyncEvent) error
	GetSubscriptionTypes() []string
}

// NewDataSyncManager 创建数据同步管理器
func NewDataSyncManager(eventBus EventBus, storage SyncStorage, logger *logrus.Logger) *DataSyncManager {
	if logger == nil {
		logger = logrus.New()
	}

	return &DataSyncManager{
		eventBus:     eventBus,
		storage:      storage,
		transformers: make(map[string]DataTransformer),
		subscribers:  make(map[string][]SyncSubscriber),
		logger:       logger,
	}
}

// Start 启动数据同步管理器
func (dsm *DataSyncManager) Start(ctx context.Context) error {
	dsm.mu.Lock()
	defer dsm.mu.Unlock()

	if dsm.running {
		return fmt.Errorf("data sync manager is already running")
	}

	dsm.ctx, dsm.cancel = context.WithCancel(ctx)
	dsm.running = true

	// 启动事件总线
	if err := dsm.eventBus.Start(dsm.ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// 订阅同步事件
	if err := dsm.subscribeToSyncEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to sync events: %w", err)
	}

	dsm.logger.Info("Data sync manager started")
	return nil
}

// Stop 停止数据同步管理器
func (dsm *DataSyncManager) Stop(ctx context.Context) error {
	dsm.mu.Lock()
	defer dsm.mu.Unlock()

	if !dsm.running {
		return fmt.Errorf("data sync manager is not running")
	}

	dsm.cancel()

	// 停止事件总线
	if err := dsm.eventBus.Stop(ctx); err != nil {
		dsm.logger.WithError(err).Error("Failed to stop event bus")
	}

	dsm.running = false
	dsm.logger.Info("Data sync manager stopped")
	return nil
}

// RegisterTransformer 注册数据转换器
func (dsm *DataSyncManager) RegisterTransformer(transformer DataTransformer) {
	dsm.mu.Lock()
	defer dsm.mu.Unlock()

	key := fmt.Sprintf("%s->%s", transformer.GetSourceType(), transformer.GetTargetType())
	dsm.transformers[key] = transformer

	dsm.logger.WithFields(logrus.Fields{
		"source": transformer.GetSourceType(),
		"target": transformer.GetTargetType(),
	}).Info("Data transformer registered")
}

// RegisterSubscriber 注册同步订阅者
func (dsm *DataSyncManager) RegisterSubscriber(subscriber SyncSubscriber) {
	dsm.mu.Lock()
	defer dsm.mu.Unlock()

	for _, eventType := range subscriber.GetSubscriptionTypes() {
		dsm.subscribers[eventType] = append(dsm.subscribers[eventType], subscriber)
	}

	dsm.logger.WithField("types", subscriber.GetSubscriptionTypes()).Info("Sync subscriber registered")
}

// SyncPluginData 同步插件数据
func (dsm *DataSyncManager) SyncPluginData(ctx context.Context, pluginID uuid.UUID, eventType string, changes map[string]interface{}, source string) error {
	// 创建同步事件
	syncEvent := &models.PluginSyncEvent{
		EventType: eventType,
		PluginID:  pluginID,
		Changes:   changes,
		Source:    source,
		Timestamp: time.Now(),
		Version:   time.Now().UnixNano(), // 简单的版本控制
	}

	// 保存同步事件
	if err := dsm.storage.SaveSyncEvent(ctx, syncEvent); err != nil {
		dsm.logger.WithError(err).Error("Failed to save sync event")
		return err
	}

	// 发布事件
	event := &models.UnifiedPluginEvent{
		ID:        uuid.New(),
		Type:      "plugin.sync." + eventType,
		Source:    source,
		Data:      syncEvent,
		Priority:  1,
		Timestamp: time.Now(),
	}

	if err := dsm.eventBus.Publish(ctx, event); err != nil {
		dsm.logger.WithError(err).Error("Failed to publish sync event")
		return err
	}

	dsm.logger.WithFields(logrus.Fields{
		"plugin_id":  pluginID,
		"event_type": eventType,
		"source":     source,
	}).Info("Plugin data sync event published")

	return nil
}

// TransformData 转换数据
func (dsm *DataSyncManager) TransformData(ctx context.Context, data interface{}, sourceType, targetType string) (interface{}, error) {
	dsm.mu.RLock()
	key := fmt.Sprintf("%s->%s", sourceType, targetType)
	transformer, exists := dsm.transformers[key]
	dsm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no transformer found for %s -> %s", sourceType, targetType)
	}

	return transformer.Transform(ctx, data)
}

// subscribeToSyncEvents 订阅同步事件
func (dsm *DataSyncManager) subscribeToSyncEvents() error {
	return dsm.eventBus.Subscribe("plugin.sync.*", dsm.handleSyncEvent)
}

// handleSyncEvent 处理同步事件
func (dsm *DataSyncManager) handleSyncEvent(ctx context.Context, event *models.UnifiedPluginEvent) error {
	syncEvent, ok := event.Data.(*models.PluginSyncEvent)
	if !ok {
		dsm.logger.Error("Invalid sync event data type")
		return fmt.Errorf("invalid sync event data type")
	}

	// 通知订阅者
	dsm.mu.RLock()
	subscribers := dsm.subscribers[syncEvent.EventType]
	dsm.mu.RUnlock()

	for _, subscriber := range subscribers {
		if err := subscriber.OnDataSync(ctx, syncEvent); err != nil {
			dsm.logger.WithError(err).WithField("subscriber", fmt.Sprintf("%T", subscriber)).Error("Subscriber failed to handle sync event")
		}
	}

	return nil
}

// GetSyncHistory 获取同步历史
func (dsm *DataSyncManager) GetSyncHistory(ctx context.Context, pluginID uuid.UUID, since time.Time) ([]*models.PluginSyncEvent, error) {
	return dsm.storage.GetSyncEvents(ctx, pluginID, since)
}

// UpdateSyncStatus 更新同步状态
func (dsm *DataSyncManager) UpdateSyncStatus(ctx context.Context, source, target string, status string, errorMsg string, recordCount int) error {
	syncStatus := &models.DataSyncStatus{
		Source:      source,
		Target:      target,
		LastSync:    time.Now(),
		Status:      status,
		ErrorMsg:    errorMsg,
		RecordCount: recordCount,
	}

	return dsm.storage.UpdateSyncStatus(ctx, syncStatus)
}

// GetSyncStatus 获取同步状态
func (dsm *DataSyncManager) GetSyncStatus(ctx context.Context, source, target string) (*models.DataSyncStatus, error) {
	return dsm.storage.GetSyncStatus(ctx, source, target)
}

// MarketToRuntimeTransformer 市场到运行时数据转换器
type MarketToRuntimeTransformer struct{}

func (t *MarketToRuntimeTransformer) Transform(ctx context.Context, data interface{}) (interface{}, error) {
	marketPlugin, ok := data.(*models.UnifiedPluginMetadata)
	if !ok {
		return nil, fmt.Errorf("invalid data type for market to runtime transformation")
	}

	// 转换为运行时插件元数据
	runtimeMetadata := map[string]interface{}{
		"id":           marketPlugin.ID.String(),
		"name":         marketPlugin.Name,
		"version":      marketPlugin.Version,
		"description":  marketPlugin.Description,
		"author":       marketPlugin.Author,
		"category":     marketPlugin.Category,
		"type":         marketPlugin.Type,
		"tags":         marketPlugin.Tags,
		"permissions":  marketPlugin.Permissions,
		"dependencies": marketPlugin.Dependencies,
		"config":       marketPlugin.Config,
		"created_at":   marketPlugin.CreatedAt,
		"updated_at":   marketPlugin.UpdatedAt,
	}

	return runtimeMetadata, nil
}

func (t *MarketToRuntimeTransformer) GetSourceType() string {
	return "market"
}

func (t *MarketToRuntimeTransformer) GetTargetType() string {
	return "runtime"
}

// RuntimeToMarketTransformer 运行时到市场数据转换器
type RuntimeToMarketTransformer struct{}

func (t *RuntimeToMarketTransformer) Transform(ctx context.Context, data interface{}) (interface{}, error) {
	runtimeData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data type for runtime to market transformation")
	}

	// 转换为市场插件数据
	marketData := map[string]interface{}{
		"runtime_state":     runtimeData["state"],
		"resource_usage":    runtimeData["resource_usage"],
		"last_health_check": time.Now(),
		"error_msg":         runtimeData["error_msg"],
	}

	return marketData, nil
}

func (t *RuntimeToMarketTransformer) GetSourceType() string {
	return "runtime"
}

func (t *RuntimeToMarketTransformer) GetTargetType() string {
	return "market"
}

// DefaultSyncStorage 默认同步存储实现
type DefaultSyncStorage struct {
	events     map[uuid.UUID][]*models.PluginSyncEvent
	statuses   map[string]*models.DataSyncStatus
	mu         sync.RWMutex
}

func NewDefaultSyncStorage() *DefaultSyncStorage {
	return &DefaultSyncStorage{
		events:   make(map[uuid.UUID][]*models.PluginSyncEvent),
		statuses: make(map[string]*models.DataSyncStatus),
	}
}

func (s *DefaultSyncStorage) SaveSyncEvent(ctx context.Context, event *models.PluginSyncEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events[event.PluginID] = append(s.events[event.PluginID], event)
	return nil
}

func (s *DefaultSyncStorage) GetSyncEvents(ctx context.Context, pluginID uuid.UUID, since time.Time) ([]*models.PluginSyncEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[pluginID]
	var result []*models.PluginSyncEvent

	for _, event := range events {
		if event.Timestamp.After(since) {
			result = append(result, event)
		}
	}

	return result, nil
}

func (s *DefaultSyncStorage) UpdateSyncStatus(ctx context.Context, status *models.DataSyncStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s->%s", status.Source, status.Target)
	s.statuses[key] = status
	return nil
}

func (s *DefaultSyncStorage) GetSyncStatus(ctx context.Context, source, target string) (*models.DataSyncStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s->%s", source, target)
	status, exists := s.statuses[key]
	if !exists {
		return nil, fmt.Errorf("sync status not found for %s -> %s", source, target)
	}

	return status, nil
}