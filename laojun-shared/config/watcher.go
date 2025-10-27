package config

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DefaultConfigWatcher 默认配置监听器实现
type DefaultConfigWatcher struct {
	storage   ConfigStorage
	logger    *logrus.Logger
	options   *ConfigOptions
	
	// 监听状态
	running bool
	mu      sync.RWMutex
	
	// 监听的键和回调
	watchedKeys map[string][]ConfigChangeCallback
	keysMux     sync.RWMutex
	
	// 控制通道
	stopChan chan struct{}
	doneChan chan struct{}
	
	// 轮询间隔
	interval time.Duration
	
	// 上次检查的版本
	lastVersions map[string]int64
	versionsMux  sync.RWMutex
}

// NewDefaultConfigWatcher 创建默认配置监听器
func NewDefaultConfigWatcher(storage ConfigStorage, options *ConfigOptions) *DefaultConfigWatcher {
	interval := 30 * time.Second
	if options != nil && options.WatchInterval > 0 {
		interval = options.WatchInterval
	}
	
	return &DefaultConfigWatcher{
		storage:      storage,
		logger:       logrus.New(),
		options:      options,
		watchedKeys:  make(map[string][]ConfigChangeCallback),
		lastVersions: make(map[string]int64),
		stopChan:     make(chan struct{}),
		doneChan:     make(chan struct{}),
		interval:     interval,
	}
}

// Start 开始监听
func (w *DefaultConfigWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.running {
		return nil
	}
	
	w.running = true
	
	go w.watchLoop(ctx)
	
	w.logger.Info("Config watcher started")
	return nil
}

// Stop 停止监听
func (w *DefaultConfigWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if !w.running {
		return nil
	}
	
	w.running = false
	close(w.stopChan)
	
	// 等待监听循环结束
	<-w.doneChan
	
	w.logger.Info("Config watcher stopped")
	return nil
}

// AddKey 添加监听键
func (w *DefaultConfigWatcher) AddKey(key string, callback ConfigChangeCallback) error {
	w.keysMux.Lock()
	defer w.keysMux.Unlock()
	
	if w.watchedKeys[key] == nil {
		w.watchedKeys[key] = make([]ConfigChangeCallback, 0)
	}
	w.watchedKeys[key] = append(w.watchedKeys[key], callback)
	
	// 初始化版本信息
	w.versionsMux.Lock()
	if _, exists := w.lastVersions[key]; !exists {
		// 获取当前版本
		if item, err := w.storage.Get(context.Background(), key); err == nil {
			w.lastVersions[key] = item.Version
		}
	}
	w.versionsMux.Unlock()
	
	w.logger.WithField("key", key).Debug("Added key to watcher")
	return nil
}

// RemoveKey 移除监听键
func (w *DefaultConfigWatcher) RemoveKey(key string) error {
	w.keysMux.Lock()
	defer w.keysMux.Unlock()
	
	delete(w.watchedKeys, key)
	
	w.versionsMux.Lock()
	delete(w.lastVersions, key)
	w.versionsMux.Unlock()
	
	w.logger.WithField("key", key).Debug("Removed key from watcher")
	return nil
}

// IsWatching 检查是否正在监听指定键
func (w *DefaultConfigWatcher) IsWatching(key string) bool {
	w.keysMux.RLock()
	defer w.keysMux.RUnlock()
	
	_, exists := w.watchedKeys[key]
	return exists
}

// GetWatchedKeys 获取所有监听的键
func (w *DefaultConfigWatcher) GetWatchedKeys() []string {
	w.keysMux.RLock()
	defer w.keysMux.RUnlock()
	
	keys := make([]string, 0, len(w.watchedKeys))
	for key := range w.watchedKeys {
		keys = append(keys, key)
	}
	
	return keys
}

// watchLoop 监听循环
func (w *DefaultConfigWatcher) watchLoop(ctx context.Context) {
	defer close(w.doneChan)
	
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkForChanges(ctx)
		}
	}
}

// checkForChanges 检查配置变化
func (w *DefaultConfigWatcher) checkForChanges(ctx context.Context) {
	w.keysMux.RLock()
	keys := make([]string, 0, len(w.watchedKeys))
	for key := range w.watchedKeys {
		keys = append(keys, key)
	}
	w.keysMux.RUnlock()
	
	if len(keys) == 0 {
		return
	}
	
	// 批量获取当前配置
	items, err := w.storage.GetMultiple(ctx, keys)
	if err != nil {
		w.logger.WithError(err).Error("Failed to get configs for watching")
		return
	}
	
	// 检查每个键的变化
	for key, item := range items {
		w.versionsMux.RLock()
		lastVersion := w.lastVersions[key]
		w.versionsMux.RUnlock()
		
		if item.Version != lastVersion {
			// 配置发生变化
			w.versionsMux.Lock()
			oldVersion := w.lastVersions[key]
			w.lastVersions[key] = item.Version
			w.versionsMux.Unlock()
			
			// 创建变化事件
			event := &ConfigChangeEvent{
				Type:      EventTypeUpdate,
				Key:       key,
				NewValue:  item.Value,
				Version:   item.Version,
				Timestamp: time.Now(),
			}
			
			// 如果是新增的配置
			if oldVersion == 0 {
				event.Type = EventTypeCreate
			}
			
			// 触发回调
			w.triggerCallbacks(key, event)
		}
	}
	
	// 检查已删除的配置
	w.checkDeletedConfigs(ctx, keys, items)
}

// checkDeletedConfigs 检查已删除的配置
func (w *DefaultConfigWatcher) checkDeletedConfigs(ctx context.Context, watchedKeys []string, currentItems map[string]*ConfigItem) {
	w.versionsMux.RLock()
	defer w.versionsMux.RUnlock()
	
	for _, key := range watchedKeys {
		if _, exists := currentItems[key]; !exists {
			// 配置已被删除
			if lastVersion, hadVersion := w.lastVersions[key]; hadVersion && lastVersion > 0 {
				// 创建删除事件
				event := &ConfigChangeEvent{
					Type:      EventTypeDelete,
					Key:       key,
					Version:   time.Now().UnixNano(),
					Timestamp: time.Now(),
				}
				
				// 触发回调
				w.triggerCallbacks(key, event)
				
				// 更新版本为0表示已删除
				w.versionsMux.RUnlock()
				w.versionsMux.Lock()
				w.lastVersions[key] = 0
				w.versionsMux.Unlock()
				w.versionsMux.RLock()
			}
		}
	}
}

// triggerCallbacks 触发回调函数
func (w *DefaultConfigWatcher) triggerCallbacks(key string, event *ConfigChangeEvent) {
	w.keysMux.RLock()
	callbacks := w.watchedKeys[key]
	w.keysMux.RUnlock()
	
	for _, callback := range callbacks {
		go func(cb ConfigChangeCallback) {
			if err := cb(event); err != nil {
				w.logger.WithError(err).WithField("key", key).Error("Config change callback failed")
			}
		}(callback)
	}
	
	w.logger.WithFields(logrus.Fields{
		"key":     key,
		"type":    event.Type,
		"version": event.Version,
	}).Debug("Config change detected")
}

// PollingConfigWatcher 基于轮询的配置监听器
type PollingConfigWatcher struct {
	*DefaultConfigWatcher
}

// NewPollingConfigWatcher 创建基于轮询的配置监听器
func NewPollingConfigWatcher(storage ConfigStorage, options *ConfigOptions) *PollingConfigWatcher {
	return &PollingConfigWatcher{
		DefaultConfigWatcher: NewDefaultConfigWatcher(storage, options),
	}
}

// EventDrivenConfigWatcher 基于事件驱动的配置监听器
type EventDrivenConfigWatcher struct {
	storage   ConfigStorage
	logger    *logrus.Logger
	options   *ConfigOptions
	
	// 监听状态
	running bool
	mu      sync.RWMutex
	
	// 监听的键和回调
	watchedKeys map[string][]ConfigChangeCallback
	keysMux     sync.RWMutex
	
	// 事件通道
	eventChan chan *ConfigChangeEvent
	stopChan  chan struct{}
	doneChan  chan struct{}
}

// NewEventDrivenConfigWatcher 创建基于事件驱动的配置监听器
func NewEventDrivenConfigWatcher(storage ConfigStorage, options *ConfigOptions) *EventDrivenConfigWatcher {
	bufferSize := 100
	if options != nil && options.WatchBufferSize > 0 {
		bufferSize = options.WatchBufferSize
	}
	
	return &EventDrivenConfigWatcher{
		storage:     storage,
		logger:      logrus.New(),
		options:     options,
		watchedKeys: make(map[string][]ConfigChangeCallback),
		eventChan:   make(chan *ConfigChangeEvent, bufferSize),
		stopChan:    make(chan struct{}),
		doneChan:    make(chan struct{}),
	}
}

// Start 开始监听
func (w *EventDrivenConfigWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.running {
		return nil
	}
	
	w.running = true
	
	go w.eventLoop(ctx)
	
	w.logger.Info("Event-driven config watcher started")
	return nil
}

// Stop 停止监听
func (w *EventDrivenConfigWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if !w.running {
		return nil
	}
	
	w.running = false
	close(w.stopChan)
	
	// 等待事件循环结束
	<-w.doneChan
	
	w.logger.Info("Event-driven config watcher stopped")
	return nil
}

// AddKey 添加监听键
func (w *EventDrivenConfigWatcher) AddKey(key string, callback ConfigChangeCallback) error {
	w.keysMux.Lock()
	defer w.keysMux.Unlock()
	
	if w.watchedKeys[key] == nil {
		w.watchedKeys[key] = make([]ConfigChangeCallback, 0)
	}
	w.watchedKeys[key] = append(w.watchedKeys[key], callback)
	
	w.logger.WithField("key", key).Debug("Added key to event-driven watcher")
	return nil
}

// RemoveKey 移除监听键
func (w *EventDrivenConfigWatcher) RemoveKey(key string) error {
	w.keysMux.Lock()
	defer w.keysMux.Unlock()
	
	delete(w.watchedKeys, key)
	
	w.logger.WithField("key", key).Debug("Removed key from event-driven watcher")
	return nil
}

// IsWatching 检查是否正在监听指定键
func (w *EventDrivenConfigWatcher) IsWatching(key string) bool {
	w.keysMux.RLock()
	defer w.keysMux.RUnlock()
	
	_, exists := w.watchedKeys[key]
	return exists
}

// GetWatchedKeys 获取所有监听的键
func (w *EventDrivenConfigWatcher) GetWatchedKeys() []string {
	w.keysMux.RLock()
	defer w.keysMux.RUnlock()
	
	keys := make([]string, 0, len(w.watchedKeys))
	for key := range w.watchedKeys {
		keys = append(keys, key)
	}
	
	return keys
}

// PublishEvent 发布配置变化事件
func (w *EventDrivenConfigWatcher) PublishEvent(event *ConfigChangeEvent) {
	select {
	case w.eventChan <- event:
	default:
		w.logger.Warn("Event channel is full, dropping event")
	}
}

// eventLoop 事件循环
func (w *EventDrivenConfigWatcher) eventLoop(ctx context.Context) {
	defer close(w.doneChan)
	
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case event := <-w.eventChan:
			w.handleEvent(event)
		}
	}
}

// handleEvent 处理配置变化事件
func (w *EventDrivenConfigWatcher) handleEvent(event *ConfigChangeEvent) {
	w.keysMux.RLock()
	callbacks := w.watchedKeys[event.Key]
	w.keysMux.RUnlock()
	
	if len(callbacks) == 0 {
		return
	}
	
	for _, callback := range callbacks {
		go func(cb ConfigChangeCallback) {
			if err := cb(event); err != nil {
				w.logger.WithError(err).WithField("key", event.Key).Error("Config change callback failed")
			}
		}(callback)
	}
	
	w.logger.WithFields(logrus.Fields{
		"key":     event.Key,
		"type":    event.Type,
		"version": event.Version,
	}).Debug("Config change event handled")
}