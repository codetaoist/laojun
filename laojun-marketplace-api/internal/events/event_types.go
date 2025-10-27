package events

import (
	"time"

	"github.com/google/uuid"
)

// EventType 事件类型
type EventType string

const (
	// 用户相关事件
	UserRegistered   EventType = "user.registered"
	UserUpdated      EventType = "user.updated"
	UserDeleted      EventType = "user.deleted"
	UserActivated    EventType = "user.activated"
	UserDeactivated  EventType = "user.deactivated"
	UserLoggedIn     EventType = "user.logged_in"
	UserLoggedOut    EventType = "user.logged_out"

	// 插件相关事件
	PluginCreated     EventType = "plugin.created"
	PluginUpdated     EventType = "plugin.updated"
	PluginDeleted     EventType = "plugin.deleted"
	PluginPublished   EventType = "plugin.published"
	PluginUnpublished EventType = "plugin.unpublished"
	PluginDownloaded  EventType = "plugin.downloaded"
	PluginFavorited   EventType = "plugin.favorited"
	PluginUnfavorited EventType = "plugin.unfavorited"

	// 支付相关事件
	PaymentOrderCreated   EventType = "payment.order.created"
	PaymentOrderCompleted EventType = "payment.order.completed"
	PaymentOrderCancelled EventType = "payment.order.cancelled"
	PaymentOrderRefunded  EventType = "payment.order.refunded"
	PaymentOrderExpired   EventType = "payment.order.expired"

	// 评价相关事件
	ReviewCreated   EventType = "review.created"
	ReviewUpdated   EventType = "review.updated"
	ReviewDeleted   EventType = "review.deleted"
	ReviewModerated EventType = "review.moderated"
	ReviewFlagged   EventType = "review.flagged"
	ReviewApproved  EventType = "review.approved"
	ReviewRejected  EventType = "review.rejected"

	// 开发者相关事件
	DeveloperRegistered EventType = "developer.registered"
	DeveloperUpdated    EventType = "developer.updated"
	DeveloperVerified   EventType = "developer.verified"
	DeveloperSuspended  EventType = "developer.suspended"
	DeveloperReactivated EventType = "developer.reactivated"

	// 分类相关事件
	CategoryCreated EventType = "category.created"
	CategoryUpdated EventType = "category.updated"
	CategoryDeleted EventType = "category.deleted"

	// 社区相关事件
	ForumPostCreated EventType = "forum.post.created"
	ForumPostUpdated EventType = "forum.post.updated"
	ForumPostDeleted EventType = "forum.post.deleted"
	BlogPostCreated  EventType = "blog.post.created"
	BlogPostUpdated  EventType = "blog.post.updated"
	BlogPostDeleted  EventType = "blog.post.deleted"

	// 系统相关事件
	SystemMaintenance EventType = "system.maintenance"
	SystemAlert       EventType = "system.alert"
	SystemBackup      EventType = "system.backup"
	SystemHealthCheck EventType = "system.health_check"
)

// Event 基础事件结构
type Event struct {
	ID          uuid.UUID              `json:"id"`
	Type        EventType              `json:"type"`
	Source      string                 `json:"source"`
	Subject     string                 `json:"subject"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	TraceID     string                 `json:"trace_id,omitempty"`
	CorrelationID string               `json:"correlation_id,omitempty"`
}

// NewEvent 创建新事件
func NewEvent(eventType EventType, source, subject string, data map[string]interface{}) *Event {
	return &Event{
		ID:        uuid.New(),
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now(),
		Version:   "1.0",
	}
}

// WithTraceID 设置追踪ID
func (e *Event) WithTraceID(traceID string) *Event {
	e.TraceID = traceID
	return e
}

// WithCorrelationID 设置关联ID
func (e *Event) WithCorrelationID(correlationID string) *Event {
	e.CorrelationID = correlationID
	return e
}

// UserRegisteredEvent 用户注册事件数据
type UserRegisteredEvent struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
}

// PluginCreatedEvent 插件创建事件数据
type PluginCreatedEvent struct {
	PluginID    uuid.UUID `json:"plugin_id"`
	Name        string    `json:"name"`
	DeveloperID uuid.UUID `json:"developer_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`
}

// PaymentOrderCreatedEvent 支付订单创建事件数据
type PaymentOrderCreatedEvent struct {
	OrderID     uuid.UUID `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	UserID      uuid.UUID `json:"user_id"`
	PluginID    uuid.UUID `json:"plugin_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Method      string    `json:"method"`
}

// PaymentOrderCompletedEvent 支付订单完成事件数据
type PaymentOrderCompletedEvent struct {
	OrderID       uuid.UUID `json:"order_id"`
	OrderNumber   string    `json:"order_number"`
	UserID        uuid.UUID `json:"user_id"`
	PluginID      uuid.UUID `json:"plugin_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Method        string    `json:"method"`
	TransactionID string    `json:"transaction_id"`
}

// ReviewCreatedEvent 评价创建事件数据
type ReviewCreatedEvent struct {
	ReviewID uuid.UUID `json:"review_id"`
	UserID   uuid.UUID `json:"user_id"`
	PluginID uuid.UUID `json:"plugin_id"`
	Rating   int       `json:"rating"`
	Content  string    `json:"content"`
}

// DeveloperRegisteredEvent 开发者注册事件数据
type DeveloperRegisteredEvent struct {
	DeveloperID uuid.UUID `json:"developer_id"`
	UserID      uuid.UUID `json:"user_id"`
	CompanyName string    `json:"company_name"`
	Website     string    `json:"website"`
	Status      string    `json:"status"`
}