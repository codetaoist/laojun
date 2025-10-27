package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EventBus 事件总线接口
type EventBus interface {
	// Subscribe 订阅事件
	Subscribe(eventType string, handler EventHandler) (string, error)

	// Unsubscribe 取消订阅
	Unsubscribe(subscriptionID string) error

	// Publish 发布事件
	Publish(ctx context.Context, event *Event) error

	// PublishAsync 异步发布事件
	PublishAsync(event *Event) error

	// GetSubscriptions 获取订阅信息
	GetSubscriptions() map[string][]*Subscription

	// GetEventHistory 获取事件历史
	GetEventHistory(limit int) []*Event

	// Start 启动事件总线
	Start(ctx context.Context) error

	// Stop 停止事件总线
	Stop(ctx context.Context) error
}

// EventHandler 事件处理函数
type EventHandler func(ctx context.Context, event *Event) error

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target,omitempty"` // 空表示广播
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Priority  int                    `json:"priority"` // 数字越小优先级越高
	TTL       time.Duration          `json:"ttl,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// Subscription 订阅信息
type Subscription struct {
	ID        string       `json:"id"`
	EventType string       `json:"event_type"`
	Handler   EventHandler `json:"-"`
	CreatedAt time.Time    `json:"created_at"`
	Active    bool         `json:"active"`
	Filter    *EventFilter `json:"filter,omitempty"`
}

// EventFilter 事件过滤器
type EventFilter struct {
	Source   string            `json:"source,omitempty"`
	Target   string            `json:"target,omitempty"`
	Priority *int              `json:"priority,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	subscriptions map[string][]*Subscription // key: event_type
	history       []*Event
	eventChan     chan *Event
	workerCount   int
	bufferSize    int
	logger        *logrus.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
	running       bool
}

// NewDefaultEventBus 创建默认事件总线
func NewDefaultEventBus(bufferSize int, workerCount int, logger *logrus.Logger) *DefaultEventBus {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	if workerCount <= 0 {
		workerCount = 5
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &DefaultEventBus{
		subscriptions: make(map[string][]*Subscription),
		history:       make([]*Event, 0),
		eventChan:     make(chan *Event, bufferSize),
		workerCount:   workerCount,
		bufferSize:    bufferSize,
		logger:        logger,
		running:       false,
	}
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

	// 启动工作协程
	for i := 0; i < eb.workerCount; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}

	eb.logger.WithField("worker_count", eb.workerCount).Info("Event bus started")
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
	close(eb.eventChan)

	// 等待所有工作协程结束
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		eb.logger.Info("Event bus stopped gracefully")
	case <-ctx.Done():
		eb.logger.Warn("Event bus stop timeout")
		return ctx.Err()
	}

	eb.running = false
	return nil
}

// Subscribe 订阅事件
func (eb *DefaultEventBus) Subscribe(eventType string, handler EventHandler) (string, error) {
	if eventType == "" {
		return "", fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return "", fmt.Errorf("handler cannot be nil")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	subscriptionID := eb.generateSubscriptionID()
	subscription := &Subscription{
		ID:        subscriptionID,
		EventType: eventType,
		Handler:   handler,
		CreatedAt: time.Now(),
		Active:    true,
	}

	if eb.subscriptions[eventType] == nil {
		eb.subscriptions[eventType] = make([]*Subscription, 0)
	}

	eb.subscriptions[eventType] = append(eb.subscriptions[eventType], subscription)

	eb.logger.WithFields(logrus.Fields{
		"subscription_id": subscriptionID,
		"event_type":      eventType,
	}).Debug("Event subscription created")

	return subscriptionID, nil
}

// Unsubscribe 取消订阅
func (eb *DefaultEventBus) Unsubscribe(subscriptionID string) error {
	if subscriptionID == "" {
		return fmt.Errorf("subscription ID cannot be empty")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	for eventType, subscriptions := range eb.subscriptions {
		for i, subscription := range subscriptions {
			if subscription.ID == subscriptionID {
				// 标记为非活跃
				subscription.Active = false
				// 从切片中移除
				eb.subscriptions[eventType] = append(subscriptions[:i], subscriptions[i+1:]...)
				
				eb.logger.WithField("subscription_id", subscriptionID).Debug("Event subscription removed")
				return nil
			}
		}
	}

	return fmt.Errorf("subscription %s not found", subscriptionID)
}

// Publish 发布事件
func (eb *DefaultEventBus) Publish(ctx context.Context, event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eb.mu.RLock()
	running := eb.running
	eb.mu.RUnlock()

	if !running {
		return fmt.Errorf("event bus is not running")
	}

	// 设置事件ID和时间戳
	if event.ID == "" {
		event.ID = eb.generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 检查TTL
	if event.TTL > 0 && time.Since(event.Timestamp) > event.TTL {
		eb.logger.WithField("event_id", event.ID).Debug("Event expired, skipping")
		return nil
	}

	// 同步处理事件
	return eb.processEvent(ctx, event)
}

// PublishAsync 异步发布事件
func (eb *DefaultEventBus) PublishAsync(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eb.mu.RLock()
	running := eb.running
	eb.mu.RUnlock()

	if !running {
		return fmt.Errorf("event bus is not running")
	}

	// 设置事件ID和时间戳
	if event.ID == "" {
		event.ID = eb.generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 检查TTL
	if event.TTL > 0 && time.Since(event.Timestamp) > event.TTL {
		eb.logger.WithField("event_id", event.ID).Debug("Event expired, skipping")
		return nil
	}

	// 异步发送到事件通道
	select {
	case eb.eventChan <- event:
		return nil
	default:
		return fmt.Errorf("event channel is full")
	}
}

// GetSubscriptions 获取订阅信息
func (eb *DefaultEventBus) GetSubscriptions() map[string][]*Subscription {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	result := make(map[string][]*Subscription)
	for eventType, subscriptions := range eb.subscriptions {
		result[eventType] = make([]*Subscription, len(subscriptions))
		copy(result[eventType], subscriptions)
	}

	return result
}

// GetEventHistory 获取事件历史
func (eb *DefaultEventBus) GetEventHistory(limit int) []*Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if limit <= 0 || limit > len(eb.history) {
		limit = len(eb.history)
	}

	// 返回最近的事件
	start := len(eb.history) - limit
	result := make([]*Event, limit)
	copy(result, eb.history[start:])

	return result
}

// worker 事件处理工作协程
func (eb *DefaultEventBus) worker(workerID int) {
	defer eb.wg.Done()

	eb.logger.WithField("worker_id", workerID).Debug("Event bus worker started")

	for {
		select {
		case <-eb.ctx.Done():
			eb.logger.WithField("worker_id", workerID).Debug("Event bus worker stopped")
			return
		case event, ok := <-eb.eventChan:
			if !ok {
				eb.logger.WithField("worker_id", workerID).Debug("Event channel closed, worker stopping")
				return
			}

			if err := eb.processEvent(eb.ctx, event); err != nil {
				eb.logger.WithError(err).WithField("event_id", event.ID).Error("Failed to process event")
			}
		}
	}
}

// processEvent 处理事件
func (eb *DefaultEventBus) processEvent(ctx context.Context, event *Event) error {
	eb.mu.RLock()
	subscriptions := eb.subscriptions[event.Type]
	eb.mu.RUnlock()

	if len(subscriptions) == 0 {
		eb.logger.WithField("event_type", event.Type).Debug("No subscribers for event type")
		return nil
	}

	// 添加到历史记录
	eb.addToHistory(event)

	// 处理订阅
	var wg sync.WaitGroup
	for _, subscription := range subscriptions {
		if !subscription.Active {
			continue
		}

		// 检查过滤器
		if subscription.Filter != nil && !eb.matchesFilter(event, subscription.Filter) {
			continue
		}

		wg.Add(1)
		go func(sub *Subscription) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					eb.logger.WithFields(logrus.Fields{
						"subscription_id": sub.ID,
						"event_id":        event.ID,
						"panic":           r,
					}).Error("Event handler panic")
				}
			}()

			if err := sub.Handler(ctx, event); err != nil {
				eb.logger.WithFields(logrus.Fields{
					"subscription_id": sub.ID,
					"event_id":        event.ID,
					"error":           err,
				}).Error("Event handler error")
			}
		}(subscription)
	}

	wg.Wait()

	eb.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"handlers":   len(subscriptions),
	}).Debug("Event processed")

	return nil
}

// matchesFilter 检查事件是否匹配过滤器
func (eb *DefaultEventBus) matchesFilter(event *Event, filter *EventFilter) bool {
	if filter.Source != "" && event.Source != filter.Source {
		return false
	}

	if filter.Target != "" && event.Target != filter.Target {
		return false
	}

	if filter.Priority != nil && event.Priority != *filter.Priority {
		return false
	}

	if filter.Metadata != nil {
		for key, value := range filter.Metadata {
			if event.Metadata[key] != value {
				return false
			}
		}
	}

	return true
}

// addToHistory 添加到历史记录
func (eb *DefaultEventBus) addToHistory(event *Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.history = append(eb.history, event)

	// 限制历史记录数量
	maxHistory := 1000
	if len(eb.history) > maxHistory {
		eb.history = eb.history[1:]
	}
}

// generateSubscriptionID 生成订阅ID
func (eb *DefaultEventBus) generateSubscriptionID() string {
	return fmt.Sprintf("sub_%d", time.Now().UnixNano())
}

// generateEventID 生成事件ID
func (eb *DefaultEventBus) generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// GetEventBusStats 获取事件总线统计信息
func (eb *DefaultEventBus) GetEventBusStats() *EventBusStats {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	stats := &EventBusStats{
		Running:           eb.running,
		WorkerCount:       eb.workerCount,
		BufferSize:        eb.bufferSize,
		EventChannelSize:  len(eb.eventChan),
		HistorySize:       len(eb.history),
		SubscriptionCount: 0,
		EventTypeCount:    len(eb.subscriptions),
	}

	for _, subscriptions := range eb.subscriptions {
		stats.SubscriptionCount += len(subscriptions)
	}

	return stats
}

// EventBusStats 事件总线统计信息
type EventBusStats struct {
	Running           bool `json:"running"`
	WorkerCount       int  `json:"worker_count"`
	BufferSize        int  `json:"buffer_size"`
	EventChannelSize  int  `json:"event_channel_size"`
	HistorySize       int  `json:"history_size"`
	SubscriptionCount int  `json:"subscription_count"`
	EventTypeCount    int  `json:"event_type_count"`
}