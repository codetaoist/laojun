// Package events 实现插件事件总线系统
package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"../core"
)

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	handlers    map[string][]core.EventHandler
	subscribers map[string]map[string]core.EventHandler // eventType -> pluginID -> handler
	mutex       sync.RWMutex
	eventQueue  chan *core.PluginEvent
	workers     int
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewDefaultEventBus 创建新的事件总线
func NewDefaultEventBus(workers int) *DefaultEventBus {
	if workers <= 0 {
		workers = 5 // 默认5个工作协程
	}

	bus := &DefaultEventBus{
		handlers:    make(map[string][]core.EventHandler),
		subscribers: make(map[string]map[string]core.EventHandler),
		eventQueue:  make(chan *core.PluginEvent, 1000), // 缓冲1000个事件
		workers:     workers,
		stopChan:    make(chan struct{}),
	}

	// 启动工作协程
	bus.startWorkers()

	return bus
}

// Subscribe 订阅事件
func (bus *DefaultEventBus) Subscribe(eventType string, handler core.EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	// 添加到处理器列表
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)

	return nil
}

// SubscribePlugin 插件订阅事件
func (bus *DefaultEventBus) SubscribePlugin(pluginID, eventType string, handler core.EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	// 初始化事件类型的订阅者映射
	if bus.subscribers[eventType] == nil {
		bus.subscribers[eventType] = make(map[string]core.EventHandler)
	}

	// 添加插件订阅
	bus.subscribers[eventType][pluginID] = handler

	// 同时添加到全局处理器列表
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)

	return nil
}

// Unsubscribe 取消订阅事件
func (bus *DefaultEventBus) Unsubscribe(eventType string, handler core.EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	handlers := bus.handlers[eventType]
	for i, h := range handlers {
		// 这里简化比较，实际可能需要更复杂的处理器比较逻辑
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			// 移除全局处理器列表中的对应项
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// UnsubscribePlugin 插件取消订阅事件
func (bus *DefaultEventBus) UnsubscribePlugin(pluginID, eventType string) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	// 从插件订阅映射中移除
	if subscribers := bus.subscribers[eventType]; subscribers != nil {
		if handler, exists := subscribers[pluginID]; exists {
			delete(subscribers, pluginID)

			// 同时从全局处理器列表中移除
			handlers := bus.handlers[eventType]
			for i, h := range handlers {
				if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
					bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
					break
				}
			}
		}
	}

	return nil
}

// Publish 发布事件（异步）
func (bus *DefaultEventBus) Publish(ctx context.Context, event *core.PluginEvent) error {
	// 设置事件时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 将事件放入队列
	select {
	case bus.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event queue is full")
	}
}

// PublishSync 发布事件（同步）
func (bus *DefaultEventBus) PublishSync(ctx context.Context, event *core.PluginEvent) error {
	// 设置事件时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 直接处理事件
	return bus.processEvent(ctx, event)
}

// processEvent 处理事件
func (bus *DefaultEventBus) processEvent(ctx context.Context, event *core.PluginEvent) error {
	bus.mutex.RLock()
	handlers := make([]core.EventHandler, len(bus.handlers[event.Type]))
	copy(handlers, bus.handlers[event.Type])
	bus.mutex.RUnlock()

	// 并发处理所有处理器
	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h core.EventHandler) {
			defer wg.Done()

			// 创建带超时的上下文，30秒超时
			handlerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := h.HandleEvent(handlerCtx, event); err != nil {
				errChan <- fmt.Errorf("handler error: %w", err)
			}
		}(handler)
	}

	// 等待所有处理器完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("event processing errors: %v", errors)
	}

	return nil
}

// startWorkers 启动工作协程
func (bus *DefaultEventBus) startWorkers() {
	for i := 0; i < bus.workers; i++ {
		bus.wg.Add(1)
		go bus.worker()
	}
}

// worker 工作协程
func (bus *DefaultEventBus) worker() {
	defer bus.wg.Done()

	for {
		select {
		case event := <-bus.eventQueue:
			// 处理事件
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			if err := bus.processEvent(ctx, event); err != nil {
				// 记录错误，实际应该使用日志系统
				fmt.Printf("Event processing error: %v\n", err)
			}
			cancel()

		case <-bus.stopChan:
			return
		}
	}
}

// Stop 停止事件总线
func (bus *DefaultEventBus) Stop() error {
	close(bus.stopChan)
	bus.wg.Wait()
	close(bus.eventQueue)
	return nil
}

// GetSubscribers 获取事件订阅者插件ID列表
func (bus *DefaultEventBus) GetSubscribers(eventType string) []string {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	var subscribers []string
	if pluginMap := bus.subscribers[eventType]; pluginMap != nil {
		for pluginID := range pluginMap {
			subscribers = append(subscribers, pluginID)
		}
	}

	return subscribers
}

// GetEventTypes 获取所有事件类型
func (bus *DefaultEventBus) GetEventTypes() []string {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	var eventTypes []string
	for eventType := range bus.handlers {
		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes
}

// EventStats 事件统计信息
type EventStats struct {
	TotalEvents     int64
	EventsByType    map[string]int64
	SubscriberCount map[string]int
	QueueSize       int
	WorkerCount     int
}

// GetStats 获取事件总线统计信息
func (bus *DefaultEventBus) GetStats() *EventStats {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	stats := &EventStats{
		EventsByType:    make(map[string]int64),
		SubscriberCount: make(map[string]int),
		QueueSize:       len(bus.eventQueue),
		WorkerCount:     bus.workers,
	}

	// 统计订阅者数
	for eventType, pluginMap := range bus.subscribers {
		stats.SubscriberCount[eventType] = len(pluginMap)
	}

	// 统计事件类型数
	for eventType := range bus.handlers {
		stats.EventsByType[eventType] = 0
	}

	return stats
}

// SystemEventHandler 系统事件处理
type SystemEventHandler struct {
	handlerFunc func(ctx context.Context, event *core.PluginEvent) error
}

// NewSystemEventHandler 创建系统事件处理
func NewSystemEventHandler(handlerFunc func(ctx context.Context, event *core.PluginEvent) error) *SystemEventHandler {
	return &SystemEventHandler{
		handlerFunc: handlerFunc,
	}
}

// HandleEvent 处理事件
func (h *SystemEventHandler) HandleEvent(ctx context.Context, event *core.PluginEvent) error {
	return h.handlerFunc(ctx, event)
}

// PluginEventHandler 插件事件处理器
type PluginEventHandler struct {
	pluginID string
	plugin   core.Plugin
}

// NewPluginEventHandler 创建插件事件处理
func NewPluginEventHandler(pluginID string, plugin core.Plugin) *PluginEventHandler {
	return &PluginEventHandler{
		pluginID: pluginID,
		plugin:   plugin,
	}
}

// HandleEvent 处理事件
func (h *PluginEventHandler) HandleEvent(ctx context.Context, event *core.PluginEvent) error {
	return h.plugin.HandleEvent(ctx, event)
}

// EventFilter 事件过滤
type EventFilter struct {
	EventTypes []string
	SourceIDs  []string
	TargetIDs  []string
	FilterFunc func(*core.PluginEvent) bool
}

// Match 检查事件是否匹配过滤器
func (f *EventFilter) Match(event *core.PluginEvent) bool {
	// 检查事件类型
	if len(f.EventTypes) > 0 {
		found := false
		for _, eventType := range f.EventTypes {
			if event.Type == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查源ID
	if len(f.SourceIDs) > 0 {
		found := false
		for _, sourceID := range f.SourceIDs {
			if event.SourceID == sourceID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查目标ID
	if len(f.TargetIDs) > 0 {
		found := false
		for _, targetID := range f.TargetIDs {
			if event.TargetID == targetID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 自定义过滤函数
	if f.FilterFunc != nil {
		return f.FilterFunc(event)
	}

	return true
}

// FilteredEventBus 带过滤器的事件总线
type FilteredEventBus struct {
	*DefaultEventBus
	filters map[string]*EventFilter // handlerID -> filter
}

// NewFilteredEventBus 创建带过滤器的事件总线
func NewFilteredEventBus(workers int) *FilteredEventBus {
	return &FilteredEventBus{
		DefaultEventBus: NewDefaultEventBus(workers),
		filters:         make(map[string]*EventFilter),
	}
}

// SubscribeWithFilter 带过滤器的事件订阅
func (bus *FilteredEventBus) SubscribeWithFilter(eventType string, handler core.EventHandler, filter *EventFilter) error {
	handlerID := fmt.Sprintf("%s_%p", eventType, handler)

	bus.mutex.Lock()
	bus.filters[handlerID] = filter
	bus.mutex.Unlock()

	return bus.Subscribe(eventType, handler)
}
