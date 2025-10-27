package models

import (
	"time"

	"github.com/google/uuid"
)

// Plugin 插件模型
type Plugin struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Author      string    `json:"author" db:"author"`
	CategoryID  uuid.UUID `json:"category_id" db:"category_id"`
	Icon        string    `json:"icon" db:"icon"`
	Screenshots []string  `json:"screenshots" db:"screenshots"` // JSON array
	Tags        []string  `json:"tags" db:"tags"`               // JSON array
	Price       float64   `json:"price" db:"price"`
	IsFeatured  bool      `json:"is_featured" db:"is_featured"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Downloads   int       `json:"downloads" db:"downloads"`
	Rating      float64   `json:"rating" db:"rating"`
	ReviewCount int       `json:"review_count" db:"review_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// 审核相关字段
	ReviewStatus           string     `json:"review_status" db:"review_status"`
	ReviewPriority         string     `json:"review_priority" db:"review_priority"`
	AutoReviewScore        *float64   `json:"auto_review_score" db:"auto_review_score"`
	AutoReviewResult       *string    `json:"auto_review_result" db:"auto_review_result"`
	ReviewNotes            *string    `json:"review_notes" db:"review_notes"`
	ReviewedAt             *time.Time `json:"reviewed_at" db:"reviewed_at"`
	ReviewerID             *uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	SubmittedForReviewAt   *time.Time `json:"submitted_for_review_at" db:"submitted_for_review_at"`
	RejectionReason        *string    `json:"rejection_reason" db:"rejection_reason"`
	AppealCount            int        `json:"appeal_count" db:"appeal_count"`
	LastAppealAt           *time.Time `json:"last_appeal_at" db:"last_appeal_at"`

	// 关联数据
	Category      *Category       `json:"category,omitempty"`
	LatestVersion *PluginVersion  `json:"latest_version,omitempty"`
	Versions      []PluginVersion `json:"versions,omitempty"`
	Reviews       []Review        `json:"reviews,omitempty"`
	Reviewer      *User           `json:"reviewer,omitempty"`
}

// PluginVersion 插件版本模型
type PluginVersion struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PluginID    uuid.UUID `json:"plugin_id" db:"plugin_id"`
	Version     string    `json:"version" db:"version"`
	Changelog   string    `json:"changelog" db:"changelog"`
	DownloadURL string    `json:"download_url" db:"download_url"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	MinVersion  string    `json:"min_version" db:"min_version"` // 最低支持版本
	MaxVersion  string    `json:"max_version" db:"max_version"` // 最高支持版本
	IsStable    bool      `json:"is_stable" db:"is_stable"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Downloads   int       `json:"downloads" db:"downloads"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Category 插件分类模型
type Category struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`
	Color       string    `json:"color" db:"color"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	PluginCount int `json:"plugin_count,omitempty"`
}

// Review 插件评价模型
type Review struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PluginID  uuid.UUID `json:"plugin_id" db:"plugin_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"` // 1-5星评分
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	IsHelpful bool      `json:"is_helpful" db:"is_helpful"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	User   *User   `json:"user,omitempty"`
	Plugin *Plugin `json:"plugin,omitempty"`
}

// Purchase 插件购买记录模型
type Purchase struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	PluginID  uuid.UUID `json:"plugin_id" db:"plugin_id"`
	Price     float64   `json:"price" db:"price"`
	Status    string    `json:"status" db:"status"` // pending, completed, failed, refunded
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	User   *User   `json:"user,omitempty"`
	Plugin *Plugin `json:"plugin,omitempty"`
}



// PaginationParams 分页参数
type PaginationParams struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
}

// PaginatedResponse 分页响应
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// ApiResponse 通用API响应
type ApiResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
