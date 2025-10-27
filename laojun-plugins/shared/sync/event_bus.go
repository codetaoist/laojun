package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/codetaoist/laojun-plugins/shared/models"
)

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	subscribers map[string][]EventHandler
	buffer      chan *models.UnifiedPluginEvent
	workers     []*EventWorker
	logger      *logrus.Logger
	mu          sync.RWMutex
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// EventHandler 事件处理器函数类型
type EventHandler func(ctx context.Context, event *models.UnifiedPluginEvent) error

// EventWorker 事件工作器
type EventWorker struct {
	id       int
	eventBus *DefaultEventBus
	logger   *logrus.Logger
}

// NewDefaultEventBus 创建默认事件总线
func NewDefaultEventBus(bufferSize int, workerCount int, logger *logrus.Logger) *DefaultEventBus {
	if logger == nil {
		logger = logrus.New()
	}

	bus := &DefaultEventBus{
		subscribers: make(map[string][]EventHandler),
		buffer:      make(chan *models.UnifiedPluginEvent, bufferSize),
		logger:      logger,
	}

	// 创建工作器
	for i := 0; i < workerCount; i++ {
		worker := &EventWorker{
			id:       i,
			eventBus: bus,
			logger:   logger,
		}
		bus.workers = append(bus.workers, worker)
	}

	return bus
}

// Start 启动事件总线
func (eb *DefaultEventBus) Start(ctx context.Context) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.running {
		return fmt.Errorf("event bus is already running")
	}

	eb.ctx, eb.cancel = context.WithCancel(ctx)
	eb.running = true

	// 启动工作器
	for _, worker := range eb.workers {
		eb.wg.Add(1)
		go worker.run(eb.ctx, &eb.wg)
	}

	eb.logger.WithField("worker_count", len(eb.workers)).Info("Event bus started")
	return nil
}

// Stop 停止事件总线
func (eb *DefaultEventBus) Stop(ctx context.Context) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.running {
		return fmt.Errorf("event bus is not running")
	}

	eb.cancel()
	close(eb.buffer)

	// 等待所有工作器完成
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		eb.logger.Info("Event bus stopped gracefully")
	case <-time.After(5 * time.Second):
		eb.logger.Warn("Event bus stop timeout")
	}

	eb.running = false
	return nil
}

// Publish 发布事件
func (eb *DefaultEventBus) Publish(ctx context.Context, event *models.UnifiedPluginEvent) error {
	eb.mu.RLock()
	running := eb.running
	eb.mu.RUnlock()

	if !running {
		return fmt.Errorf("event bus is not running")
	}

	// 设置事件ID（如果没有）
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	// 设置时间戳（如果没有）
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case eb.buffer <- event:
		eb.logger.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.Type,
			"source":     event.Source,
		}).Debug("Event published")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("event publish timeout")
	}
}

// Subscribe 订阅事件
func (eb *DefaultEventBus) Subscribe(eventType string, handler func(ctx context.Context, event *models.UnifiedPluginEvent) error) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], EventHandler(handler))

	eb.logger.WithField("event_type", eventType).Info("Event handler subscribed")
	return nil
}

// Unsubscribe 取消订阅事件
func (eb *DefaultEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.subscribers[eventType]
	for i, h := range handlers {
		// 比较函数指针（简单实现）
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			eb.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			eb.logger.WithField("event_type", eventType).Info("Event handler unsubscribed")
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// GetSubscriberCount 获取订阅者数量
func (eb *DefaultEventBus) GetSubscriberCount(eventType string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return len(eb.subscribers[eventType])
}

// GetTotalSubscriberCount 获取总订阅者数量
func (eb *DefaultEventBus) GetTotalSubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	total := 0
	for _, handlers := range eb.subscribers {
		total += len(handlers)
	}
	return total
}

// run 工作器运行方法
func (w *EventWorker) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	w.logger.WithField("worker_id", w.id).Info("Event worker started")

	for {
		select {
		case event, ok := <-w.eventBus.buffer:
			if !ok {
				w.logger.WithField("worker_id", w.id).Info("Event worker stopped - channel closed")
				return
			}
			w.handleEvent(ctx, event)
		case <-ctx.Done():
			w.logger.WithField("worker_id", w.id).Info("Event worker stopped - context cancelled")
			return
		}
	}
}

// handleEvent 处理事件
func (w *EventWorker) handleEvent(ctx context.Context, event *models.UnifiedPluginEvent) {
	w.eventBus.mu.RLock()
	subscribers := make(map[string][]EventHandler)
	
	// 查找匹配的订阅者
	for eventType, handlers := range w.eventBus.subscribers {
		if w.matchEventType(event.Type, eventType) {
			subscribers[eventType] = make([]EventHandler, len(handlers))
			copy(subscribers[eventType], handlers)
		}
	}
	w.eventBus.mu.RUnlock()

	// 执行处理器
	for _, handlers := range subscribers {
		for _, handler := range handlers {
			func() {
				defer func() {
					if r := recover(); r != nil {
						w.logger.WithFields(logrus.Fields{
							"event_id":   event.ID,
							"event_type": event.Type,
							"panic":      r,
						}).Error("Event handler panicked")
					}
				}()

				if err := handler(ctx, event); err != nil {
					w.logger.WithFields(logrus.Fields{
						"event_id":   event.ID,
						"event_type": event.Type,
						"error":      err,
					}).Error("Event handler failed")
				}
			}()
		}

		w.logger.WithFields(logrus.Fields{
			"event_id":      event.ID,
			"event_type":    event.Type,
			"handler_count": len(handlers),
		}).Debug("Event processed")
	}
}

// matchEventType 匹配事件类型（支持通配符）
func (w *EventWorker) matchEventType(eventType, pattern string) bool {
	// 精确匹配
	if eventType == pattern {
		return true
	}

	// 通配符匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(eventType, prefix)
	}

	return false
}

// EventMetrics 事件指标
type EventMetrics struct {
	PublishedCount  int64
	ProcessedCount  int64
	FailedCount     int64
	SubscriberCount int
	BufferSize      int
	BufferUsage     int
}

// GetMetrics 获取事件总线指标
func (eb *DefaultEventBus) GetMetrics() *EventMetrics {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return &EventMetrics{
		SubscriberCount: eb.GetTotalSubscriberCount(),
		BufferSize:      cap(eb.buffer),
		BufferUsage:     len(eb.buffer),
	}
}

// EventFilter 事件过滤器
type EventFilter struct {
	EventTypes []string
	Sources    []string
	MinPriority int
	MaxAge     time.Duration
}

// Match 检查事件是否匹配过滤器
func (f *EventFilter) Match(event *models.UnifiedPluginEvent) bool {
	// 检查事件类型
	if len(f.EventTypes) > 0 {
		matched := false
		for _, eventType := range f.EventTypes {
			if strings.HasSuffix(eventType, "*") {
				prefix := strings.TrimSuffix(eventType, "*")
				if strings.HasPrefix(event.Type, prefix) {
					matched = true
					break
				}
			} else if event.Type == eventType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查来源
	if len(f.Sources) > 0 {
		matched := false
		for _, source := range f.Sources {
			if event.Source == source {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查优先级
	if event.Priority < f.MinPriority {
		return false
	}

	// 检查年龄
	if f.MaxAge > 0 && time.Since(event.Timestamp) > f.MaxAge {
		return false
	}

	return true
}

// FilteredEventBus 带过滤器的事件总线
type FilteredEventBus struct {
	*DefaultEventBus
	filter *EventFilter
}

// NewFilteredEventBus 创建带过滤器的事件总线
func NewFilteredEventBus(bufferSize int, workerCount int, filter *EventFilter, logger *logrus.Logger) *FilteredEventBus {
	return &FilteredEventBus{
		DefaultEventBus: NewDefaultEventBus(bufferSize, workerCount, logger),
		filter:          filter,
	}
}

// Publish 发布事件（带过滤）
func (feb *FilteredEventBus) Publish(ctx context.Context, event *models.UnifiedPluginEvent) error {
	if feb.filter != nil && !feb.filter.Match(event) {
		feb.logger.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.Type,
			"source":     event.Source,
		}).Debug("Event filtered out")
		return nil
	}

	return feb.DefaultEventBus.Publish(ctx, event)
}