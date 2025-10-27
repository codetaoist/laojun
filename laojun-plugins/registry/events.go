package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EventType 事件类型
type EventType string

const (
	// 插件生命周期事件
	EventPluginRegistered   EventType = "plugin.registered"
	EventPluginUnregistered EventType = "plugin.unregistered"
	EventPluginStarted      EventType = "plugin.started"
	EventPluginStopped      EventType = "plugin.stopped"
	EventPluginUpdated      EventType = "plugin.updated"
	EventPluginFailed       EventType = "plugin.failed"

	// 插件状态事件
	EventPluginHealthy   EventType = "plugin.healthy"
	EventPluginUnhealthy EventType = "plugin.unhealthy"
	EventPluginTimeout   EventType = "plugin.timeout"

	// 插件指标事件
	EventPluginMetricsUpdated EventType = "plugin.metrics.updated"
	EventPluginOverloaded     EventType = "plugin.overloaded"
	EventPluginUnderloaded    EventType = "plugin.underloaded"

	// 系统事件
	EventRegistryStarted EventType = "registry.started"
	EventRegistryStopped EventType = "registry.stopped"
	EventRegistryError   EventType = "registry.error"
)

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	PluginID  string                 `json:"plugin_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error

// EventSubscription 事件订阅
type EventSubscription struct {
	ID       string       `json:"id"`
	Types    []EventType  `json:"types"`
	Handler  EventHandler `json:"-"`
	Filter   EventFilter  `json:"filter,omitempty"`
	Created  time.Time    `json:"created"`
	Active   bool         `json:"active"`
}

// EventFilter 事件过滤器
type EventFilter struct {
	PluginIDs []string          `json:"plugin_ids,omitempty"`
	Sources   []string          `json:"sources,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// EventBus 事件总线接口
type EventBus interface {
	// Publish 发布事件
	Publish(ctx context.Context, event *Event) error

	// Subscribe 订阅事件
	Subscribe(ctx context.Context, types []EventType, handler EventHandler) (*EventSubscription, error)

	// SubscribeWithFilter 带过滤器订阅事件
	SubscribeWithFilter(ctx context.Context, types []EventType, filter EventFilter, handler EventHandler) (*EventSubscription, error)

	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, subscriptionID string) error

	// GetSubscriptions 获取所有订阅
	GetSubscriptions(ctx context.Context) ([]*EventSubscription, error)

	// Close 关闭事件总线
	Close() error
}

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	subscriptions map[string]*EventSubscription
	mutex         sync.RWMutex
	logger        *logrus.Logger
	bufferSize    int
	workers       int
	eventChan     chan *Event
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewDefaultEventBus 创建默认事件总线
func NewDefaultEventBus(logger *logrus.Logger, bufferSize, workers int) *DefaultEventBus {
	bus := &DefaultEventBus{
		subscriptions: make(map[string]*EventSubscription),
		logger:        logger,
		bufferSize:    bufferSize,
		workers:       workers,
		eventChan:     make(chan *Event, bufferSize),
		stopChan:      make(chan struct{}),
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		bus.wg.Add(1)
		go bus.worker()
	}

	return bus
}

// Publish 发布事件
func (b *DefaultEventBus) Publish(ctx context.Context, event *Event) error {
	if event.ID == "" {
		event.ID = b.generateEventID()
	}
	
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	b.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"plugin_id":  event.PluginID,
	}).Debug("Publishing event")

	select {
	case b.eventChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event buffer full, dropping event")
	}
}

// Subscribe 订阅事件
func (b *DefaultEventBus) Subscribe(ctx context.Context, types []EventType, handler EventHandler) (*EventSubscription, error) {
	return b.SubscribeWithFilter(ctx, types, EventFilter{}, handler)
}

// SubscribeWithFilter 带过滤器订阅事件
func (b *DefaultEventBus) SubscribeWithFilter(ctx context.Context, types []EventType, filter EventFilter, handler EventHandler) (*EventSubscription, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscription := &EventSubscription{
		ID:      b.generateSubscriptionID(),
		Types:   types,
		Handler: handler,
		Filter:  filter,
		Created: time.Now(),
		Active:  true,
	}

	b.subscriptions[subscription.ID] = subscription

	b.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"event_types":     types,
	}).Debug("Event subscription created")

	return subscription, nil
}

// Unsubscribe 取消订阅
func (b *DefaultEventBus) Unsubscribe(ctx context.Context, subscriptionID string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscription, exists := b.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	subscription.Active = false
	delete(b.subscriptions, subscriptionID)

	b.logger.WithField("subscription_id", subscriptionID).Debug("Event subscription removed")
	return nil
}

// GetSubscriptions 获取所有订阅
func (b *DefaultEventBus) GetSubscriptions(ctx context.Context) ([]*EventSubscription, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var subscriptions []*EventSubscription
	for _, sub := range b.subscriptions {
		if sub.Active {
			// 创建副本，不包含Handler
			subCopy := &EventSubscription{
				ID:      sub.ID,
				Types:   sub.Types,
				Filter:  sub.Filter,
				Created: sub.Created,
				Active:  sub.Active,
			}
			subscriptions = append(subscriptions, subCopy)
		}
	}

	return subscriptions, nil
}

// Close 关闭事件总线
func (b *DefaultEventBus) Close() error {
	close(b.stopChan)
	b.wg.Wait()
	close(b.eventChan)

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 清理所有订阅
	for id, sub := range b.subscriptions {
		sub.Active = false
		delete(b.subscriptions, id)
	}

	b.logger.Info("Event bus closed")
	return nil
}

// worker 事件处理工作协程
func (b *DefaultEventBus) worker() {
	defer b.wg.Done()

	for {
		select {
		case event := <-b.eventChan:
			b.processEvent(event)
		case <-b.stopChan:
			return
		}
	}
}

// processEvent 处理事件
func (b *DefaultEventBus) processEvent(event *Event) {
	b.mutex.RLock()
	subscriptions := make([]*EventSubscription, 0, len(b.subscriptions))
	for _, sub := range b.subscriptions {
		if sub.Active && b.matchesSubscription(event, sub) {
			subscriptions = append(subscriptions, sub)
		}
	}
	b.mutex.RUnlock()

	// 并发处理所有匹配的订阅
	var wg sync.WaitGroup
	for _, sub := range subscriptions {
		wg.Add(1)
		go func(subscription *EventSubscription) {
			defer wg.Done()
			b.handleEvent(event, subscription)
		}(sub)
	}
	wg.Wait()
}

// handleEvent 处理单个订阅的事件
func (b *DefaultEventBus) handleEvent(event *Event, subscription *EventSubscription) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			b.logger.WithFields(logrus.Fields{
				"event_id":        event.ID,
				"subscription_id": subscription.ID,
				"panic":           r,
			}).Error("Event handler panicked")
		}
	}()

	if err := subscription.Handler(ctx, event); err != nil {
		b.logger.WithFields(logrus.Fields{
			"event_id":        event.ID,
			"subscription_id": subscription.ID,
			"error":           err,
		}).Error("Event handler failed")
	}
}

// matchesSubscription 检查事件是否匹配订阅
func (b *DefaultEventBus) matchesSubscription(event *Event, subscription *EventSubscription) bool {
	// 检查事件类型
	typeMatches := false
	for _, eventType := range subscription.Types {
		if event.Type == eventType {
			typeMatches = true
			break
		}
	}
	if !typeMatches {
		return false
	}

	// 检查过滤器
	filter := subscription.Filter

	// 插件ID过滤
	if len(filter.PluginIDs) > 0 {
		found := false
		for _, pluginID := range filter.PluginIDs {
			if event.PluginID == pluginID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 来源过滤
	if len(filter.Sources) > 0 {
		found := false
		for _, source := range filter.Sources {
			if event.Source == source {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 元数据过滤
	if len(filter.Metadata) > 0 {
		for key, value := range filter.Metadata {
			if eventValue, exists := event.Metadata[key]; !exists || eventValue != value {
				return false
			}
		}
	}

	return true
}

// generateEventID 生成事件ID
func (b *DefaultEventBus) generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// generateSubscriptionID 生成订阅ID
func (b *DefaultEventBus) generateSubscriptionID() string {
	return fmt.Sprintf("sub_%d", time.Now().UnixNano())
}

// EventPublisher 事件发布器
type EventPublisher struct {
	bus    EventBus
	source string
	logger *logrus.Logger
}

// NewEventPublisher 创建事件发布器
func NewEventPublisher(bus EventBus, source string, logger *logrus.Logger) *EventPublisher {
	return &EventPublisher{
		bus:    bus,
		source: source,
		logger: logger,
	}
}

// PublishPluginRegistered 发布插件注册事件
func (p *EventPublisher) PublishPluginRegistered(ctx context.Context, pluginID string, plugin *PluginRegistration) error {
	event := &Event{
		Type:     EventPluginRegistered,
		Source:   p.source,
		PluginID: pluginID,
		Data: map[string]interface{}{
			"plugin": plugin,
		},
	}
	return p.bus.Publish(ctx, event)
}

// PublishPluginUnregistered 发布插件注销事件
func (p *EventPublisher) PublishPluginUnregistered(ctx context.Context, pluginID string) error {
	event := &Event{
		Type:     EventPluginUnregistered,
		Source:   p.source,
		PluginID: pluginID,
	}
	return p.bus.Publish(ctx, event)
}

// PublishPluginStatusChanged 发布插件状态变更事件
func (p *EventPublisher) PublishPluginStatusChanged(ctx context.Context, pluginID string, oldStatus, newStatus PluginStatus) error {
	var eventType EventType
	switch newStatus {
	case StatusActive:
		eventType = EventPluginStarted
	case StatusInactive:
		eventType = EventPluginStopped
	case StatusFailed:
		eventType = EventPluginFailed
	default:
		eventType = EventPluginUpdated
	}

	event := &Event{
		Type:     eventType,
		Source:   p.source,
		PluginID: pluginID,
		Data: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": newStatus,
		},
	}
	return p.bus.Publish(ctx, event)
}

// PublishPluginHealthChanged 发布插件健康状态变更事件
func (p *EventPublisher) PublishPluginHealthChanged(ctx context.Context, pluginID string, healthy bool, healthInfo *HealthInfo) error {
	var eventType EventType
	if healthy {
		eventType = EventPluginHealthy
	} else {
		eventType = EventPluginUnhealthy
	}

	event := &Event{
		Type:     eventType,
		Source:   p.source,
		PluginID: pluginID,
		Data: map[string]interface{}{
			"healthy":     healthy,
			"health_info": healthInfo,
		},
	}
	return p.bus.Publish(ctx, event)
}

// PublishPluginMetricsUpdated 发布插件指标更新事件
func (p *EventPublisher) PublishPluginMetricsUpdated(ctx context.Context, pluginID string, metrics *PluginMetrics) error {
	event := &Event{
		Type:     EventPluginMetricsUpdated,
		Source:   p.source,
		PluginID: pluginID,
		Data: map[string]interface{}{
			"metrics": metrics,
		},
	}
	return p.bus.Publish(ctx, event)
}

// PublishPluginOverloaded 发布插件过载事件
func (p *EventPublisher) PublishPluginOverloaded(ctx context.Context, pluginID string, metrics *PluginMetrics) error {
	event := &Event{
		Type:     EventPluginOverloaded,
		Source:   p.source,
		PluginID: pluginID,
		Data: map[string]interface{}{
			"metrics": metrics,
		},
	}
	return p.bus.Publish(ctx, event)
}

// PublishRegistryStarted 发布注册中心启动事件
func (p *EventPublisher) PublishRegistryStarted(ctx context.Context) error {
	event := &Event{
		Type:   EventRegistryStarted,
		Source: p.source,
	}
	return p.bus.Publish(ctx, event)
}

// PublishRegistryStopped 发布注册中心停止事件
func (p *EventPublisher) PublishRegistryStopped(ctx context.Context) error {
	event := &Event{
		Type:   EventRegistryStopped,
		Source: p.source,
	}
	return p.bus.Publish(ctx, event)
}

// PublishRegistryError 发布注册中心错误事件
func (p *EventPublisher) PublishRegistryError(ctx context.Context, err error) error {
	event := &Event{
		Type:   EventRegistryError,
		Source: p.source,
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	}
	return p.bus.Publish(ctx, event)
}

// EventLogger 事件日志记录器
type EventLogger struct {
	logger *logrus.Logger
}

// NewEventLogger 创建事件日志记录器
func NewEventLogger(logger *logrus.Logger) *EventLogger {
	return &EventLogger{
		logger: logger,
	}
}

// HandleEvent 处理事件日志记录
func (l *EventLogger) HandleEvent(ctx context.Context, event *Event) error {
	fields := logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"source":     event.Source,
		"timestamp":  event.Timestamp,
	}

	if event.PluginID != "" {
		fields["plugin_id"] = event.PluginID
	}

	if len(event.Data) > 0 {
		fields["data"] = event.Data
	}

	if len(event.Metadata) > 0 {
		fields["metadata"] = event.Metadata
	}

	switch event.Type {
	case EventPluginFailed, EventPluginUnhealthy, EventRegistryError:
		l.logger.WithFields(fields).Error("Plugin event")
	case EventPluginTimeout, EventPluginOverloaded:
		l.logger.WithFields(fields).Warn("Plugin event")
	default:
		l.logger.WithFields(fields).Info("Plugin event")
	}

	return nil
}