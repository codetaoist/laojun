package models

import (
	"time"

	"github.com/google/uuid"
)

// PluginAuditStatus 插件审核状态
type PluginAuditStatus string

const (
	PluginAuditStatusPending   PluginAuditStatus = "pending"   // 待审核
	PluginAuditStatusReviewing PluginAuditStatus = "reviewing" // 审核中
	PluginAuditStatusApproved  PluginAuditStatus = "approved"  // 已通过
	PluginAuditStatusRejected  PluginAuditStatus = "rejected"  // 已拒绝
	PluginAuditStatusRevision  PluginAuditStatus = "revision"  // 需修改
)

// PluginSubmissionType 插件提交类型
type PluginSubmissionType string

const (
	PluginSubmissionTypeNew    PluginSubmissionType = "new"    // 新插件
	PluginSubmissionTypeUpdate PluginSubmissionType = "update" // 更新版本
)

// PluginAuditPriority 审核优先级
type PluginAuditPriority string

const (
	PluginAuditPriorityLow    PluginAuditPriority = "low"    // 低优先级
	PluginAuditPriorityNormal PluginAuditPriority = "normal" // 普通优先级
	PluginAuditPriorityHigh   PluginAuditPriority = "high"   // 高优先级
	PluginAuditPriorityUrgent PluginAuditPriority = "urgent" // 紧急
)

// PluginAuditRecord 插件审核记录
type PluginAuditRecord struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	PluginID         uuid.UUID              `json:"plugin_id" db:"plugin_id"`
	PluginName       string                 `json:"plugin_name" db:"plugin_name"`
	PluginVersion    string                 `json:"plugin_version" db:"plugin_version"`
	DeveloperID      uuid.UUID              `json:"developer_id" db:"developer_id"`
	DeveloperName    string                 `json:"developer_name" db:"developer_name"`
	SubmissionType   PluginSubmissionType   `json:"submission_type" db:"submission_type"`
	Status           PluginAuditStatus      `json:"status" db:"status"`
	Priority         PluginAuditPriority    `json:"priority" db:"priority"`
	AssignedAuditorID *uuid.UUID            `json:"assigned_auditor_id" db:"assigned_auditor_id"`
	AssignedAuditorName *string             `json:"assigned_auditor_name" db:"assigned_auditor_name"`
	SubmissionData   string                 `json:"submission_data" db:"submission_data"` // JSON格式的插件数据
	AuditNotes       *string                `json:"audit_notes" db:"audit_notes"`
	RejectionReason  *string                `json:"rejection_reason" db:"rejection_reason"`
	SecurityScore    *int                   `json:"security_score" db:"security_score"`    // 安全评分 0-100
	QualityScore     *int                   `json:"quality_score" db:"quality_score"`      // 质量评分 0-100
	PerformanceScore *int                   `json:"performance_score" db:"performance_score"` // 性能评分 0-100
	SubmittedAt      time.Time              `json:"submitted_at" db:"submitted_at"`
	AssignedAt       *time.Time             `json:"assigned_at" db:"assigned_at"`
	ReviewStartedAt  *time.Time             `json:"review_started_at" db:"review_started_at"`
	CompletedAt      *time.Time             `json:"completed_at" db:"completed_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// PluginAuditComment 审核评论
type PluginAuditComment struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AuditRecordID uuid.UUID `json:"audit_record_id" db:"audit_record_id"`
	AuditorID     uuid.UUID `json:"auditor_id" db:"auditor_id"`
	AuditorName   string    `json:"auditor_name" db:"auditor_name"`
	CommentType   string    `json:"comment_type" db:"comment_type"` // general, security, quality, performance
	Content       string    `json:"content" db:"content"`
	IsInternal    bool      `json:"is_internal" db:"is_internal"` // 是否为内部评论（不对开发者可见）
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// PluginAuditChecklist 审核检查清单
type PluginAuditChecklist struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AuditRecordID uuid.UUID `json:"audit_record_id" db:"audit_record_id"`
	CheckCategory string    `json:"check_category" db:"check_category"` // security, quality, performance, compliance
	CheckItem     string    `json:"check_item" db:"check_item"`
	CheckResult   string    `json:"check_result" db:"check_result"` // pass, fail, warning, skip
	CheckNotes    *string   `json:"check_notes" db:"check_notes"`
	CheckedBy     uuid.UUID `json:"checked_by" db:"checked_by"`
	CheckedAt     time.Time `json:"checked_at" db:"checked_at"`
}

// DeveloperProfile 开发者档案
type DeveloperProfile struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	DeveloperName     string     `json:"developer_name" db:"developer_name"`
	CompanyName       *string    `json:"company_name" db:"company_name"`
	Website           *string    `json:"website" db:"website"`
	ContactEmail      string     `json:"contact_email" db:"contact_email"`
	ContactPhone      *string    `json:"contact_phone" db:"contact_phone"`
	Biography         *string    `json:"biography" db:"biography"`
	Avatar            *string    `json:"avatar" db:"avatar"`
	IsVerified        bool       `json:"is_verified" db:"is_verified"`
	VerificationLevel string     `json:"verification_level" db:"verification_level"` // basic, standard, premium
	TrustScore        int        `json:"trust_score" db:"trust_score"`               // 信任评分 0-100
	TotalPlugins      int        `json:"total_plugins" db:"total_plugins"`
	ApprovedPlugins   int        `json:"approved_plugins" db:"approved_plugins"`
	RejectedPlugins   int        `json:"rejected_plugins" db:"rejected_plugins"`
	AverageRating     float64    `json:"average_rating" db:"average_rating"`
	LastSubmissionAt  *time.Time `json:"last_submission_at" db:"last_submission_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// AuditorProfile 审核员档案
type AuditorProfile struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	AuditorLevel        string     `json:"auditor_level" db:"auditor_level"` // junior, senior, expert, lead
	Specializations     string     `json:"specializations" db:"specializations"` // JSON数组: ["security", "performance", "ui"]
	MaxConcurrentAudits int        `json:"max_concurrent_audits" db:"max_concurrent_audits"`
	CurrentAudits       int        `json:"current_audits" db:"current_audits"`
	TotalAudits         int        `json:"total_audits" db:"total_audits"`
	CompletedAudits     int        `json:"completed_audits" db:"completed_audits"`
	AverageAuditTime    *int       `json:"average_audit_time" db:"average_audit_time"` // 平均审核时间（小时）
	QualityRating       float64    `json:"quality_rating" db:"quality_rating"`         // 审核质量评分
	IsActive            bool       `json:"is_active" db:"is_active"`
	LastAuditAt         *time.Time `json:"last_audit_at" db:"last_audit_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// PluginAuditWorkflow 审核工作流配置
type PluginAuditWorkflow struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	WorkflowName          string    `json:"workflow_name" db:"workflow_name"`
	PluginCategory        string    `json:"plugin_category" db:"plugin_category"`
	RequiredAuditorLevel  string    `json:"required_auditor_level" db:"required_auditor_level"`
	RequiredSpecializations string  `json:"required_specializations" db:"required_specializations"` // JSON数组
	AutoAssignmentEnabled bool      `json:"auto_assignment_enabled" db:"auto_assignment_enabled"`
	MaxAuditDays          int       `json:"max_audit_days" db:"max_audit_days"`
	RequiresPeerReview    bool      `json:"requires_peer_review" db:"requires_peer_review"`
	ChecklistTemplate     string    `json:"checklist_template" db:"checklist_template"` // JSON格式的检查清单模板
	IsActive              bool      `json:"is_active" db:"is_active"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// PluginAuditStatistics 审核统计
type PluginAuditStatistics struct {
	TotalSubmissions    int64   `json:"total_submissions"`
	PendingAudits       int64   `json:"pending_audits"`
	InReviewAudits      int64   `json:"in_review_audits"`
	CompletedAudits     int64   `json:"completed_audits"`
	ApprovedRate        float64 `json:"approved_rate"`
	AverageAuditTime    float64 `json:"average_audit_time"` // 小时
	AverageSecurityScore float64 `json:"average_security_score"`
	AverageQualityScore  float64 `json:"average_quality_score"`
	ActiveAuditors      int64   `json:"active_auditors"`
	OverdueAudits       int64   `json:"overdue_audits"`
}

// 请求和响应模型

// PluginAuditSubmissionRequest 插件审核提交请求
type PluginAuditSubmissionRequest struct {
	PluginID         uuid.UUID              `json:"plugin_id" binding:"required"`
	PluginName       string                 `json:"plugin_name" binding:"required"`
	PluginVersion    string                 `json:"plugin_version" binding:"required"`
	SubmissionType   PluginSubmissionType   `json:"submission_type" binding:"required"`
	Priority         PluginAuditPriority    `json:"priority"`
	SubmissionData   map[string]interface{} `json:"submission_data" binding:"required"`
	DeveloperNotes   string                 `json:"developer_notes"`
}

// PluginAuditAssignmentRequest 审核分配请求
type PluginAuditAssignmentRequest struct {
	AuditorID uuid.UUID `json:"auditor_id" binding:"required"`
	Notes     string    `json:"notes"`
}

// PluginAuditReviewRequest 审核评审请求
type PluginAuditReviewRequest struct {
	Status           PluginAuditStatus `json:"status" binding:"required"`
	AuditNotes       string            `json:"audit_notes"`
	RejectionReason  string            `json:"rejection_reason"`
	SecurityScore    *int              `json:"security_score"`
	QualityScore     *int              `json:"quality_score"`
	PerformanceScore *int              `json:"performance_score"`
	Comments         []AuditCommentRequest `json:"comments"`
	ChecklistResults []ChecklistResultRequest `json:"checklist_results"`
}

// AuditCommentRequest 审核评论请求
type AuditCommentRequest struct {
	CommentType string `json:"comment_type" binding:"required"`
	Content     string `json:"content" binding:"required"`
	IsInternal  bool   `json:"is_internal"`
}

// ChecklistResultRequest 检查清单结果请求
type ChecklistResultRequest struct {
	CheckCategory string  `json:"check_category" binding:"required"`
	CheckItem     string  `json:"check_item" binding:"required"`
	CheckResult   string  `json:"check_result" binding:"required"`
	CheckNotes    *string `json:"check_notes"`
}

// PluginAuditListRequest 审核列表查询请求
type PluginAuditListRequest struct {
	Status       []PluginAuditStatus    `json:"status"`
	Priority     []PluginAuditPriority  `json:"priority"`
	AuditorID    *uuid.UUID             `json:"auditor_id"`
	DeveloperID  *uuid.UUID             `json:"developer_id"`
	SubmissionType []PluginSubmissionType `json:"submission_type"`
	DateFrom     *time.Time             `json:"date_from"`
	DateTo       *time.Time             `json:"date_to"`
	Page         int                    `json:"page"`
	Size         int                    `json:"size"`
	SortBy       string                 `json:"sort_by"`
	SortOrder    string                 `json:"sort_order"`
}

// PluginAuditResponse 审核记录响应
type PluginAuditResponse struct {
	PluginAuditRecord
	Comments         []PluginAuditComment    `json:"comments,omitempty"`
	ChecklistResults []PluginAuditChecklist  `json:"checklist_results,omitempty"`
	DeveloperProfile *DeveloperProfile       `json:"developer_profile,omitempty"`
	AuditorProfile   *AuditorProfile         `json:"auditor_profile,omitempty"`
}

// PluginAuditListResponse 审核列表响应
type PluginAuditListResponse struct {
	Records []PluginAuditResponse `json:"records"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	Size    int                   `json:"size"`
}

// DeveloperVerificationRequest 开发者认证请求
type DeveloperVerificationRequest struct {
	DeveloperName string  `json:"developer_name" binding:"required"`
	CompanyName   *string `json:"company_name"`
	Website       *string `json:"website"`
	ContactEmail  string  `json:"contact_email" binding:"required,email"`
	ContactPhone  *string `json:"contact_phone"`
	Biography     *string `json:"biography"`
	Documents     []string `json:"documents"` // 认证文档URL列表
}

// AuditorProfileUpdateRequest 审核员档案更新请求
type AuditorProfileUpdateRequest struct {
	AuditorLevel        string   `json:"auditor_level"`
	Specializations     []string `json:"specializations"`
	MaxConcurrentAudits int      `json:"max_concurrent_audits"`
	IsActive            bool     `json:"is_active"`
}