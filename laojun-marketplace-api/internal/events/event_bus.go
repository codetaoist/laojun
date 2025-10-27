package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	GetEventTypes() []EventType
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(ctx context.Context, event *Event) error
	Subscribe(handler EventHandler) error
	Unsubscribe(handler EventHandler) error
	Start(ctx context.Context) error
	Stop() error
}

// InMemoryEventBus 内存事件总线实现
type InMemoryEventBus struct {
	handlers map[EventType][]EventHandler
	mutex    sync.RWMutex
	logger   *logrus.Logger
	running  bool
	stopCh   chan struct{}
}

// NewInMemoryEventBus 创建内存事件总线
func NewInMemoryEventBus(logger *logrus.Logger) *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[EventType][]EventHandler),
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Publish 发布事件
func (bus *InMemoryEventBus) Publish(ctx context.Context, event *Event) error {
	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.Type]
	bus.mutex.RUnlock()

	if !exists {
		bus.logger.Debugf("No handlers found for event type: %s", event.Type)
		return nil
	}

	// 异步处理事件
	go func() {
		for _, handler := range handlers {
			go func(h EventHandler) {
				if err := h.Handle(ctx, event); err != nil {
					bus.logger.Errorf("Error handling event %s: %v", event.ID, err)
				}
			}(handler)
		}
	}()

	bus.logger.Infof("Published event: %s (ID: %s)", event.Type, event.ID)
	return nil
}

// Subscribe 订阅事件
func (bus *InMemoryEventBus) Subscribe(handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, eventType := range handler.GetEventTypes() {
		bus.handlers[eventType] = append(bus.handlers[eventType], handler)
		bus.logger.Infof("Subscribed handler for event type: %s", eventType)
	}

	return nil
}

// Unsubscribe 取消订阅
func (bus *InMemoryEventBus) Unsubscribe(handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, eventType := range handler.GetEventTypes() {
		handlers := bus.handlers[eventType]
		for i, h := range handlers {
			if h == handler {
				bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
		bus.logger.Infof("Unsubscribed handler for event type: %s", eventType)
	}

	return nil
}

// Start 启动事件总线
func (bus *InMemoryEventBus) Start(ctx context.Context) error {
	bus.running = true
	bus.logger.Info("InMemory event bus started")
	return nil
}

// Stop 停止事件总线
func (bus *InMemoryEventBus) Stop() error {
	bus.running = false
	close(bus.stopCh)
	bus.logger.Info("InMemory event bus stopped")
	return nil
}

// RedisEventBus Redis事件总线实现
type RedisEventBus struct {
	client   *redis.Client
	handlers map[EventType][]EventHandler
	mutex    sync.RWMutex
	logger   *logrus.Logger
	running  bool
	stopCh   chan struct{}
	prefix   string
}

// NewRedisEventBus 创建Redis事件总线
func NewRedisEventBus(client *redis.Client, logger *logrus.Logger, prefix string) *RedisEventBus {
	if prefix == "" {
		prefix = "events"
	}

	return &RedisEventBus{
		client:   client,
		handlers: make(map[EventType][]EventHandler),
		logger:   logger,
		stopCh:   make(chan struct{}),
		prefix:   prefix,
	}
}

// Publish 发布事件到Redis
func (bus *RedisEventBus) Publish(ctx context.Context, event *Event) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := fmt.Sprintf("%s:%s", bus.prefix, event.Type)
	if err := bus.client.Publish(ctx, channel, eventData).Err(); err != nil {
		return fmt.Errorf("failed to publish event to Redis: %w", err)
	}

	bus.logger.Infof("Published event to Redis: %s (ID: %s)", event.Type, event.ID)
	return nil
}

// Subscribe 订阅事件
func (bus *RedisEventBus) Subscribe(handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, eventType := range handler.GetEventTypes() {
		bus.handlers[eventType] = append(bus.handlers[eventType], handler)
		bus.logger.Infof("Subscribed handler for event type: %s", eventType)
	}

	return nil
}

// Unsubscribe 取消订阅
func (bus *RedisEventBus) Unsubscribe(handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, eventType := range handler.GetEventTypes() {
		handlers := bus.handlers[eventType]
		for i, h := range handlers {
			if h == handler {
				bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
		bus.logger.Infof("Unsubscribed handler for event type: %s", eventType)
	}

	return nil
}

// Start 启动Redis事件总线
func (bus *RedisEventBus) Start(ctx context.Context) error {
	bus.running = true

	// 获取所有需要订阅的事件类型
	bus.mutex.RLock()
	var channels []string
	for eventType := range bus.handlers {
		channels = append(channels, fmt.Sprintf("%s:%s", bus.prefix, eventType))
	}
	bus.mutex.RUnlock()

	if len(channels) == 0 {
		bus.logger.Info("No event handlers registered, Redis event bus started without subscriptions")
		return nil
	}

	// 订阅Redis频道
	pubsub := bus.client.Subscribe(ctx, channels...)

	go func() {
		defer pubsub.Close()

		for {
			select {
			case <-bus.stopCh:
				return
			default:
				msg, err := pubsub.ReceiveTimeout(ctx, time.Second)
				if err != nil {
					if err == redis.Nil {
						continue
					}
					bus.logger.Errorf("Error receiving message from Redis: %v", err)
					continue
				}

				switch m := msg.(type) {
				case *redis.Message:
					bus.handleRedisMessage(ctx, m)
				}
			}
		}
	}()

	bus.logger.Info("Redis event bus started")
	return nil
}

// Stop 停止Redis事件总线
func (bus *RedisEventBus) Stop() error {
	bus.running = false
	close(bus.stopCh)
	bus.logger.Info("Redis event bus stopped")
	return nil
}

// handleRedisMessage 处理Redis消息
func (bus *RedisEventBus) handleRedisMessage(ctx context.Context, msg *redis.Message) {
	var event Event
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		bus.logger.Errorf("Failed to unmarshal event: %v", err)
		return
	}

	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.Type]
	bus.mutex.RUnlock()

	if !exists {
		bus.logger.Debugf("No handlers found for event type: %s", event.Type)
		return
	}

	// 异步处理事件
	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h.Handle(ctx, &event); err != nil {
				bus.logger.Errorf("Error handling event %s: %v", event.ID, err)
			}
		}(handler)
	}
}

// EventBusManager 事件总线管理器
type EventBusManager struct {
	buses  map[string]EventBus
	logger *logrus.Logger
}

// NewEventBusManager 创建事件总线管理器
func NewEventBusManager(logger *logrus.Logger) *EventBusManager {
	return &EventBusManager{
		buses:  make(map[string]EventBus),
		logger: logger,
	}
}

// RegisterBus 注册事件总线
func (manager *EventBusManager) RegisterBus(name string, bus EventBus) {
	manager.buses[name] = bus
	manager.logger.Infof("Registered event bus: %s", name)
}

// GetBus 获取事件总线
func (manager *EventBusManager) GetBus(name string) (EventBus, bool) {
	bus, exists := manager.buses[name]
	return bus, exists
}

// StartAll 启动所有事件总线
func (manager *EventBusManager) StartAll(ctx context.Context) error {
	for name, bus := range manager.buses {
		if err := bus.Start(ctx); err != nil {
			return fmt.Errorf("failed to start event bus %s: %w", name, err)
		}
	}
	return nil
}

// StopAll 停止所有事件总线
func (manager *EventBusManager) StopAll() error {
	for name, bus := range manager.buses {
		if err := bus.Stop(); err != nil {
			manager.logger.Errorf("Failed to stop event bus %s: %v", name, err)
		}
	}
	return nil
}