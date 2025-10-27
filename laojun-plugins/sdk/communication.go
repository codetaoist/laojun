package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EventBus 事件总线接口
type EventBus interface {
	// Subscribe 订阅事件
	Subscribe(eventType string, handler EventHandler) error
	
	// Unsubscribe 取消订阅
	Unsubscribe(eventType string, handler EventHandler) error
	
	// Publish 发布事件
	Publish(ctx context.Context, event *Event) error
	
	// PublishAsync 异步发布事件
	PublishAsync(event *Event) error
	
	// Close 关闭事件总线
	Close() error
}

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Priority  EventPriority          `json:"priority"`
	TTL       time.Duration          `json:"ttl,omitempty"`
}

// EventPriority 事件优先级
type EventPriority int

const (
	PriorityLow EventPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	subscribers map[string][]EventHandler
	asyncQueue  chan *Event
	logger      *logrus.Logger
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewDefaultEventBus 创建默认事件总线
func NewDefaultEventBus(logger *logrus.Logger) *DefaultEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	
	bus := &DefaultEventBus{
		subscribers: make(map[string][]EventHandler),
		asyncQueue:  make(chan *Event, 1000),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// 启动异步事件处理器
	bus.wg.Add(1)
	go bus.processAsyncEvents()
	
	return bus
}

// Subscribe 订阅事件
func (bus *DefaultEventBus) Subscribe(eventType string, handler EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("event handler cannot be nil")
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)
	
	bus.logger.WithFields(logrus.Fields{
		"event_type": eventType,
		"handlers":   len(bus.subscribers[eventType]),
	}).Debug("Event handler subscribed")

	return nil
}

// Unsubscribe 取消订阅
func (bus *DefaultEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	handlers, exists := bus.subscribers[eventType]
	if !exists {
		return fmt.Errorf("no subscribers for event type: %s", eventType)
	}

	// 移除处理器（这里简化实现，实际应该比较函数指针）
	for i, h := range handlers {
		if &h == &handler {
			bus.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	bus.logger.WithFields(logrus.Fields{
		"event_type": eventType,
		"handlers":   len(bus.subscribers[eventType]),
	}).Debug("Event handler unsubscribed")

	return nil
}

// Publish 发布事件
func (bus *DefaultEventBus) Publish(ctx context.Context, event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// 设置事件时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 生成事件ID
	if event.ID == "" {
		event.ID = generateEventID()
	}

	bus.mu.RLock()
	handlers, exists := bus.subscribers[event.Type]
	bus.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		bus.logger.WithField("event_type", event.Type).Debug("No subscribers for event")
		return nil
	}

	// 同步处理事件
	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			
			// 创建超时上下文
			handlerCtx := ctx
			if event.TTL > 0 {
				var cancel context.CancelFunc
				handlerCtx, cancel = context.WithTimeout(ctx, event.TTL)
				defer cancel()
			}

			if err := h(handlerCtx, event); err != nil {
				errChan <- err
				bus.logger.WithError(err).WithFields(logrus.Fields{
					"event_id":   event.ID,
					"event_type": event.Type,
				}).Error("Event handler failed")
			}
		}(handler)
	}

	wg.Wait()
	close(errChan)

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("event handling failed with %d errors", len(errors))
	}

	bus.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"handlers":   len(handlers),
	}).Debug("Event published successfully")

	return nil
}

// PublishAsync 异步发布事件
func (bus *DefaultEventBus) PublishAsync(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// 设置事件时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 生成事件ID
	if event.ID == "" {
		event.ID = generateEventID()
	}

	select {
	case bus.asyncQueue <- event:
		return nil
	default:
		return fmt.Errorf("async event queue is full")
	}
}

// processAsyncEvents 处理异步事件
func (bus *DefaultEventBus) processAsyncEvents() {
	defer bus.wg.Done()

	for {
		select {
		case <-bus.ctx.Done():
			return
		case event := <-bus.asyncQueue:
			if err := bus.Publish(bus.ctx, event); err != nil {
				bus.logger.WithError(err).WithFields(logrus.Fields{
					"event_id":   event.ID,
					"event_type": event.Type,
				}).Error("Failed to process async event")
			}
		}
	}
}

// Close 关闭事件总线
func (bus *DefaultEventBus) Close() error {
	bus.cancel()
	bus.wg.Wait()
	close(bus.asyncQueue)
	return nil
}

// generateEventID 生成事件ID
func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// MessageBroker 消息代理接口
type MessageBroker interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, message *Message) error
	
	// SendRequest 发送请求并等待响应
	SendRequest(ctx context.Context, request *Message) (*Message, error)
	
	// RegisterHandler 注册消息处理器
	RegisterHandler(messageType string, handler MessageHandler) error
	
	// UnregisterHandler 注销消息处理器
	UnregisterHandler(messageType string) error
	
	// Close 关闭消息代理
	Close() error
}

// Message 消息结构
type Message struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Data        map[string]interface{} `json:"data"`
	Headers     map[string]string      `json:"headers"`
	Timestamp   time.Time              `json:"timestamp"`
	ReplyTo     string                 `json:"replyTo,omitempty"`
	CorrelationID string               `json:"correlationId,omitempty"`
}

// MessageHandler 消息处理器
type MessageHandler func(ctx context.Context, message *Message) (*Message, error)

// DefaultMessageBroker 默认消息代理实现
type DefaultMessageBroker struct {
	handlers      map[string]MessageHandler
	pendingReplies map[string]chan *Message
	logger        *logrus.Logger
	mu            sync.RWMutex
}

// NewDefaultMessageBroker 创建默认消息代理
func NewDefaultMessageBroker(logger *logrus.Logger) *DefaultMessageBroker {
	return &DefaultMessageBroker{
		handlers:       make(map[string]MessageHandler),
		pendingReplies: make(map[string]chan *Message),
		logger:         logger,
	}
}

// SendMessage 发送消息
func (broker *DefaultMessageBroker) SendMessage(ctx context.Context, message *Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// 设置消息时间戳
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// 生成消息ID
	if message.ID == "" {
		message.ID = generateMessageID()
	}

	// 检查是否是回复消息
	if message.CorrelationID != "" {
		broker.mu.RLock()
		replyChan, exists := broker.pendingReplies[message.CorrelationID]
		broker.mu.RUnlock()

		if exists {
			select {
			case replyChan <- message:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// 查找消息处理器
	broker.mu.RLock()
	handler, exists := broker.handlers[message.Type]
	broker.mu.RUnlock()

	if !exists {
		broker.logger.WithField("message_type", message.Type).Debug("No handler for message type")
		return nil
	}

	// 处理消息
	reply, err := handler(ctx, message)
	if err != nil {
		broker.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":   message.ID,
			"message_type": message.Type,
		}).Error("Message handler failed")
		return err
	}

	// 发送回复
	if reply != nil && message.ReplyTo != "" {
		reply.To = message.ReplyTo
		reply.CorrelationID = message.ID
		return broker.SendMessage(ctx, reply)
	}

	return nil
}

// SendRequest 发送请求并等待响应
func (broker *DefaultMessageBroker) SendRequest(ctx context.Context, request *Message) (*Message, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// 生成请求ID
	if request.ID == "" {
		request.ID = generateMessageID()
	}

	// 设置回复地址
	request.ReplyTo = "reply_" + request.ID

	// 创建回复通道
	replyChan := make(chan *Message, 1)
	broker.mu.Lock()
	broker.pendingReplies[request.ID] = replyChan
	broker.mu.Unlock()

	// 清理回复通道
	defer func() {
		broker.mu.Lock()
		delete(broker.pendingReplies, request.ID)
		broker.mu.Unlock()
		close(replyChan)
	}()

	// 发送请求
	if err := broker.SendMessage(ctx, request); err != nil {
		return nil, err
	}

	// 等待回复
	select {
	case reply := <-replyChan:
		return reply, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// RegisterHandler 注册消息处理器
func (broker *DefaultMessageBroker) RegisterHandler(messageType string, handler MessageHandler) error {
	if messageType == "" {
		return fmt.Errorf("message type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("message handler cannot be nil")
	}

	broker.mu.Lock()
	defer broker.mu.Unlock()

	broker.handlers[messageType] = handler

	broker.logger.WithField("message_type", messageType).Debug("Message handler registered")
	return nil
}

// UnregisterHandler 注销消息处理器
func (broker *DefaultMessageBroker) UnregisterHandler(messageType string) error {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	delete(broker.handlers, messageType)

	broker.logger.WithField("message_type", messageType).Debug("Message handler unregistered")
	return nil
}

// Close 关闭消息代理
func (broker *DefaultMessageBroker) Close() error {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	// 关闭所有待处理的回复通道
	for id, ch := range broker.pendingReplies {
		close(ch)
		delete(broker.pendingReplies, id)
	}

	return nil
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// CommunicationManager 通信管理器
type CommunicationManager struct {
	eventBus      EventBus
	messageBroker MessageBroker
	logger        *logrus.Logger
}

// NewCommunicationManager 创建通信管理器
func NewCommunicationManager(logger *logrus.Logger) *CommunicationManager {
	return &CommunicationManager{
		eventBus:      NewDefaultEventBus(logger),
		messageBroker: NewDefaultMessageBroker(logger),
		logger:        logger,
	}
}

// GetEventBus 获取事件总线
func (cm *CommunicationManager) GetEventBus() EventBus {
	return cm.eventBus
}

// GetMessageBroker 获取消息代理
func (cm *CommunicationManager) GetMessageBroker() MessageBroker {
	return cm.messageBroker
}

// Close 关闭通信管理器
func (cm *CommunicationManager) Close() error {
	if err := cm.eventBus.Close(); err != nil {
		cm.logger.WithError(err).Error("Failed to close event bus")
	}
	
	if err := cm.messageBroker.Close(); err != nil {
		cm.logger.WithError(err).Error("Failed to close message broker")
	}
	
	return nil
}

// PluginCommunicator 插件通信器
type PluginCommunicator struct {
	pluginID string
	eventBus EventBus
	broker   MessageBroker
	logger   *logrus.Logger
}

// NewPluginCommunicator 创建插件通信器
func NewPluginCommunicator(pluginID string, cm *CommunicationManager) *PluginCommunicator {
	return &PluginCommunicator{
		pluginID: pluginID,
		eventBus: cm.GetEventBus(),
		broker:   cm.GetMessageBroker(),
		logger:   cm.logger.WithField("plugin_id", pluginID),
	}
}

// EmitEvent 发送事件
func (pc *PluginCommunicator) EmitEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	event := &Event{
		Type:      eventType,
		Source:    pc.pluginID,
		Data:      data,
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
	}

	return pc.eventBus.Publish(ctx, event)
}

// OnEvent 监听事件
func (pc *PluginCommunicator) OnEvent(eventType string, handler EventHandler) error {
	return pc.eventBus.Subscribe(eventType, handler)
}

// SendMessage 发送消息
func (pc *PluginCommunicator) SendMessage(ctx context.Context, to, messageType string, data map[string]interface{}) error {
	message := &Message{
		Type:      messageType,
		From:      pc.pluginID,
		To:        to,
		Data:      data,
		Timestamp: time.Now(),
	}

	return pc.broker.SendMessage(ctx, message)
}

// SendRequest 发送请求
func (pc *PluginCommunicator) SendRequest(ctx context.Context, to, messageType string, data map[string]interface{}) (*Message, error) {
	request := &Message{
		Type:      messageType,
		From:      pc.pluginID,
		To:        to,
		Data:      data,
		Timestamp: time.Now(),
	}

	return pc.broker.SendRequest(ctx, request)
}

// OnMessage 监听消息
func (pc *PluginCommunicator) OnMessage(messageType string, handler MessageHandler) error {
	return pc.broker.RegisterHandler(messageType, handler)
}

// 预定义事件类型
const (
	EventPluginLoaded     = "plugin.loaded"
	EventPluginStarted    = "plugin.started"
	EventPluginStopped    = "plugin.stopped"
	EventPluginError      = "plugin.error"
	EventPluginHealthy    = "plugin.healthy"
	EventPluginUnhealthy  = "plugin.unhealthy"
	EventDataChanged      = "data.changed"
	EventConfigChanged    = "config.changed"
	EventResourceExceeded = "resource.exceeded"
)

// 预定义消息类型
const (
	MessageTypeRequest  = "request"
	MessageTypeResponse = "response"
	MessageTypeCommand  = "command"
	MessageTypeQuery    = "query"
	MessageTypeNotify   = "notify"
)