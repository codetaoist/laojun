package models

import (
	"time"

	"github.com/google/uuid"
)

// PluginReviewStatus 插件审核状态
type PluginReviewStatus string

const (
	ReviewStatusPending   PluginReviewStatus = "pending"   // 待审核
	ReviewStatusApproved  PluginReviewStatus = "approved"  // 已通过
	ReviewStatusRejected  PluginReviewStatus = "rejected"  // 已拒绝
	ReviewStatusSuspended PluginReviewStatus = "suspended" // 已暂停
)

// ReviewType 审核类型
type ReviewType string

const (
	ReviewTypeAuto   ReviewType = "auto"   // 自动审核
	ReviewTypeManual ReviewType = "manual" // 人工审核
	ReviewTypeAppeal ReviewType = "appeal" // 申诉审核
)

// ReviewResult 审核结果
type ReviewResult string

const (
	ReviewResultPass ReviewResult = "pass" // 通过
	ReviewResultFail ReviewResult = "fail" // 不通过
	ReviewResultWarn ReviewResult = "warn" // 警告
)

// ReviewPriority 审核优先级
type ReviewPriority string

const (
	ReviewPriorityLow    ReviewPriority = "low"    // 低优先级
	ReviewPriorityNormal ReviewPriority = "normal" // 普通优先级
	ReviewPriorityHigh   ReviewPriority = "high"   // 高优先级
	ReviewPriorityUrgent ReviewPriority = "urgent" // 紧急优先级
)

// PluginReview 插件审核记录
type PluginReview struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	PluginID        uuid.UUID              `json:"plugin_id" db:"plugin_id"`
	ReviewerID      *uuid.UUID             `json:"reviewer_id" db:"reviewer_id"`
	PreviousStatus  string                 `json:"previous_status" db:"previous_status"`
	NewStatus       string                 `json:"new_status" db:"new_status"`
	ReviewType      ReviewType             `json:"review_type" db:"review_type"`
	ReviewResult    ReviewResult           `json:"review_result" db:"review_result"`
	Priority        ReviewPriority         `json:"priority" db:"priority"`
	ReviewNotes     string                 `json:"review_notes" db:"review_notes"`
	ReviewChecklist map[string]interface{} `json:"review_checklist" db:"review_checklist"` // JSON
	AutoReviewData  map[string]interface{} `json:"auto_review_data" db:"auto_review_data"` // JSON
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`

	// 关联数据
	Plugin   *Plugin `json:"plugin,omitempty"`
	Reviewer *User   `json:"reviewer,omitempty"`
}

// DeveloperAppeal 开发者申诉
type DeveloperAppeal struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	PluginID     uuid.UUID  `json:"plugin_id" db:"plugin_id"`
	DeveloperID  uuid.UUID  `json:"developer_id" db:"developer_id"`
	ReviewID     *uuid.UUID `json:"review_id" db:"review_id"`
	AppealType   string     `json:"appeal_type" db:"appeal_type"`
	AppealReason string     `json:"appeal_reason" db:"appeal_reason"`
	AppealStatus string     `json:"appeal_status" db:"appeal_status"`
	AdminID      *uuid.UUID `json:"admin_id" db:"admin_id"`
	AdminReply   string     `json:"admin_reply" db:"admin_reply"`
	ProcessedAt  *time.Time `json:"processed_at" db:"processed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// 关联数据
	Plugin    *Plugin       `json:"plugin,omitempty"`
	Developer *User         `json:"developer,omitempty"`
	Review    *PluginReview `json:"review,omitempty"`
	Admin     *User         `json:"admin,omitempty"`
}

// ReviewerWorkload 审核员工作負載
type ReviewerWorkload struct {
	ID                uuid.UUID `json:"id" db:"id"`
	ReviewerID        uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	Date              time.Time `json:"date" db:"date"`
	ReviewsAssigned   int       `json:"reviews_assigned" db:"reviews_assigned"`
	ReviewsCompleted  int       `json:"reviews_completed" db:"reviews_completed"`
	ReviewsPending    int       `json:"reviews_pending" db:"reviews_pending"`
	AverageReviewTime int       `json:"average_review_time" db:"average_review_time"` // 分钟
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Reviewer *User `json:"reviewer,omitempty"`
}

// ReviewConfig 审核配置
type ReviewConfig struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ConfigKey   string    `json:"config_key" db:"config_key"`
	ConfigValue string    `json:"config_value" db:"config_value"`
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ReviewTemplate 审核模板
type ReviewTemplate struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	TemplateName string                 `json:"template_name" db:"template_name"`
	TemplateType string                 `json:"template_type" db:"template_type"`
	CategoryID   *uuid.UUID             `json:"category_id" db:"category_id"`
	Checklist    map[string]interface{} `json:"checklist" db:"checklist"` // JSON
	IsActive     bool                   `json:"is_active" db:"is_active"`
	CreatedBy    uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`

	// 关联数据
	Category *Category `json:"category,omitempty"`
	Creator  *User     `json:"creator,omitempty"`
}

// AutoReviewLog 自动审核日志
type AutoReviewLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	PluginID     uuid.UUID              `json:"plugin_id" db:"plugin_id"`
	ReviewType   string                 `json:"review_type" db:"review_type"`
	CheckResult  ReviewResult           `json:"check_result" db:"check_result"`
	Score        float64                `json:"score" db:"score"`
	Details      map[string]interface{} `json:"details" db:"details"` // JSON
	ErrorMessage string                 `json:"error_message" db:"error_message"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`

	// 关联数据
	Plugin *Plugin `json:"plugin,omitempty"`
}

// PluginVersionReview 插件版本审核关联
type PluginVersionReview struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PluginID  uuid.UUID `json:"plugin_id" db:"plugin_id"`
	ReviewID  uuid.UUID `json:"review_id" db:"review_id"`
	Version   string    `json:"version" db:"version"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// 关联数据
	Plugin *Plugin       `json:"plugin,omitempty"`
	Review *PluginReview `json:"review,omitempty"`
}

// 审核相关的请求和响应模型

// ReviewQueueParams 审核队列查询参数
type ReviewQueueParams struct {
	Status     []PluginReviewStatus `json:"status" form:"status"`
	Priority   []ReviewPriority     `json:"priority" form:"priority"`
	ReviewType []ReviewType         `json:"review_type" form:"review_type"`
	CategoryID *uuid.UUID           `json:"category_id" form:"category_id"`
	ReviewerID *uuid.UUID           `json:"reviewer_id" form:"reviewer_id"`
	DateFrom   *time.Time           `json:"date_from" form:"date_from"`
	DateTo     *time.Time           `json:"date_to" form:"date_to"`
	Page       int                  `json:"page" form:"page"`
	Limit      int                  `json:"limit" form:"limit"`
	SortBy     string               `json:"sort_by" form:"sort_by"`
	SortOrder  string               `json:"sort_order" form:"sort_order"`
}

// ReviewRequest 审核请求
type ReviewRequest struct {
	PluginID        uuid.UUID              `json:"plugin_id" binding:"required"`
	ReviewResult    ReviewResult           `json:"review_result" binding:"required"`
	ReviewNotes     string                 `json:"review_notes"`
	ReviewChecklist map[string]interface{} `json:"review_checklist"`
}

// BatchReviewRequest 批量审核请求
type BatchReviewRequest struct {
	PluginIDs       []uuid.UUID            `json:"plugin_ids" binding:"required"`
	ReviewResult    ReviewResult           `json:"review_result" binding:"required"`
	ReviewNotes     string                 `json:"review_notes"`
	ReviewChecklist map[string]interface{} `json:"review_checklist"`
}

// AppealRequest 申诉请求
type AppealRequest struct {
	PluginID     uuid.UUID  `json:"plugin_id" binding:"required"`
	ReviewID     *uuid.UUID `json:"review_id"`
	AppealType   string     `json:"appeal_type" binding:"required"`
	AppealReason string     `json:"appeal_reason" binding:"required"`
}

// AppealProcessRequest 申诉处理请求
type AppealProcessRequest struct {
	AppealID   uuid.UUID `json:"appeal_id" binding:"required"`
	AdminReply string    `json:"admin_reply" binding:"required"`
	Approved   bool      `json:"approved"`
}

// AppealListParams 申诉列表查询参数
type AppealListParams struct {
	Status      []string   `json:"status" form:"status"`
	PluginID    *uuid.UUID `json:"plugin_id" form:"plugin_id"`
	DeveloperID *uuid.UUID `json:"developer_id" form:"developer_id"`
	DateFrom    *time.Time `json:"date_from" form:"date_from"`
	DateTo      *time.Time `json:"date_to" form:"date_to"`
	Page        int        `json:"page" form:"page"`
	Limit       int        `json:"limit" form:"limit"`
	SortBy      string     `json:"sort_by" form:"sort_by"`
	SortOrder   string     `json:"sort_order" form:"sort_order"`
}

// ReviewStats 审核统计
type ReviewStats struct {
	TotalPlugins      int                         `json:"total_plugins"`
	PendingReviews    int                         `json:"pending_reviews"`
	ApprovedPlugins   int                         `json:"approved_plugins"`
	RejectedPlugins   int                         `json:"rejected_plugins"`
	SuspendedPlugins  int                         `json:"suspended_plugins"`
	StatusBreakdown   map[PluginReviewStatus]int  `json:"status_breakdown"`
	PriorityBreakdown map[ReviewPriority]int      `json:"priority_breakdown"`
	CategoryBreakdown map[string]int              `json:"category_breakdown"`
	ReviewerWorkload  map[string]ReviewerWorkload `json:"reviewer_workload"`
	AverageReviewTime float64                     `json:"average_review_time"`
	ReviewTrends      []ReviewTrendData           `json:"review_trends"`
}

// ReviewTrendData 审核趋势数据
type ReviewTrendData struct {
	Date      time.Time `json:"date"`
	Submitted int       `json:"submitted"`
	Approved  int       `json:"approved"`
	Rejected  int       `json:"rejected"`
}
