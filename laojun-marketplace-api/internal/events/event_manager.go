package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// EventManager 事件管理器
type EventManager struct {
	eventBus EventBus
	handlers map[EventType][]EventHandler
	logger   *logrus.Logger
	mu       sync.RWMutex
	started  bool
}

// NewEventManager 创建事件管理器
func NewEventManager(eventBus EventBus, logger *logrus.Logger) *EventManager {
	return &EventManager{
		eventBus: eventBus,
		handlers: make(map[EventType][]EventHandler),
		logger:   logger,
	}
}

// RegisterHandler 注册事件处理器
func (em *EventManager) RegisterHandler(handler EventHandler) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	eventTypes := handler.GetEventTypes()
	for _, eventType := range eventTypes {
		em.handlers[eventType] = append(em.handlers[eventType], handler)
		em.logger.Infof("Registered handler for event type: %s", eventType)
	}

	return nil
}

// RegisterHandlers 批量注册事件处理器
func (em *EventManager) RegisterHandlers(handlers ...EventHandler) error {
	for _, handler := range handlers {
		if err := em.RegisterHandler(handler); err != nil {
			return fmt.Errorf("failed to register handler: %w", err)
		}
	}
	return nil
}

// Start 启动事件管理器
func (em *EventManager) Start(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.started {
		return fmt.Errorf("event manager already started")
	}

	// 为每种事件类型订阅处理器
	for eventType, handlers := range em.handlers {
		for _, handler := range handlers {
			if err := em.eventBus.Subscribe(eventType, handler); err != nil {
				return fmt.Errorf("failed to subscribe handler for event type %s: %w", eventType, err)
			}
		}
	}

	// 启动事件总线
	if err := em.eventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	em.started = true
	em.logger.Info("Event manager started successfully")

	return nil
}

// Stop 停止事件管理器
func (em *EventManager) Stop(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if !em.started {
		return nil
	}

	// 停止事件总线
	if err := em.eventBus.Stop(ctx); err != nil {
		em.logger.Errorf("Failed to stop event bus: %v", err)
		return err
	}

	em.started = false
	em.logger.Info("Event manager stopped")

	return nil
}

// PublishEvent 发布事件
func (em *EventManager) PublishEvent(ctx context.Context, event *Event) error {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if !em.started {
		return fmt.Errorf("event manager not started")
	}

	em.logger.Infof("Publishing event: %s (ID: %s)", event.Type, event.ID)

	if err := em.eventBus.Publish(ctx, event); err != nil {
		em.logger.Errorf("Failed to publish event %s: %v", event.ID, err)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// PublishUserRegistered 发布用户注册事件
func (em *EventManager) PublishUserRegistered(ctx context.Context, userID, email, username string) error {
	event := NewEvent(UserRegistered, map[string]interface{}{
		"user_id":  userID,
		"email":    email,
		"username": username,
	})
	return em.PublishEvent(ctx, event)
}

// PublishPluginCreated 发布插件创建事件
func (em *EventManager) PublishPluginCreated(ctx context.Context, pluginID, name, developerID, categoryID string, price float64) error {
	event := NewEvent(PluginCreated, map[string]interface{}{
		"plugin_id":    pluginID,
		"name":         name,
		"developer_id": developerID,
		"category_id":  categoryID,
		"price":        price,
	})
	return em.PublishEvent(ctx, event)
}

// PublishPaymentOrderCreated 发布支付订单创建事件
func (em *EventManager) PublishPaymentOrderCreated(ctx context.Context, orderID, userID, pluginID string, amount float64, currency string) error {
	event := NewEvent(PaymentOrderCreated, map[string]interface{}{
		"order_id":  orderID,
		"user_id":   userID,
		"plugin_id": pluginID,
		"amount":    amount,
		"currency":  currency,
	})
	return em.PublishEvent(ctx, event)
}

// PublishPaymentOrderCompleted 发布支付订单完成事件
func (em *EventManager) PublishPaymentOrderCompleted(ctx context.Context, orderID, userID, pluginID string, amount float64, currency string) error {
	event := NewEvent(PaymentOrderCompleted, map[string]interface{}{
		"order_id":  orderID,
		"user_id":   userID,
		"plugin_id": pluginID,
		"amount":    amount,
		"currency":  currency,
	})
	return em.PublishEvent(ctx, event)
}

// PublishReviewCreated 发布评价创建事件
func (em *EventManager) PublishReviewCreated(ctx context.Context, reviewID, userID, pluginID, content string, rating int) error {
	event := NewEvent(ReviewCreated, map[string]interface{}{
		"review_id": reviewID,
		"user_id":   userID,
		"plugin_id": pluginID,
		"content":   content,
		"rating":    rating,
	})
	return em.PublishEvent(ctx, event)
}

// PublishDeveloperRegistered 发布开发者注册事件
func (em *EventManager) PublishDeveloperRegistered(ctx context.Context, developerID, userID, companyName string) error {
	event := NewEvent(DeveloperRegistered, map[string]interface{}{
		"developer_id":   developerID,
		"user_id":        userID,
		"company_name":   companyName,
	})
	return em.PublishEvent(ctx, event)
}

// GetHandlerCount 获取处理器数量
func (em *EventManager) GetHandlerCount() map[EventType]int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	counts := make(map[EventType]int)
	for eventType, handlers := range em.handlers {
		counts[eventType] = len(handlers)
	}
	return counts
}

// IsStarted 检查是否已启动
func (em *EventManager) IsStarted() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.started
}

// GetEventBus 获取事件总线
func (em *EventManager) GetEventBus() EventBus {
	return em.eventBus
}