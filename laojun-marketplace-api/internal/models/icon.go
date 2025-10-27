package models

import (
	"time"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Icon 图标模型
type Icon struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	IconType  string         `json:"icon_type" db:"icon_type"`
	IconData  string         `json:"icon_data" db:"icon_data"`
	Category  string         `json:"category" db:"category"`
	Tags      pq.StringArray `json:"tags" db:"tags"`
	IsActive  bool           `json:"is_active" db:"is_active"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// IconSearchParams 图标搜索参数
type IconSearchParams struct {
	Page     int    `form:"page" json:"page"`
	Limit    int    `form:"limit" json:"limit"`
	Category string `form:"category" json:"category"`
	IconType string `form:"icon_type" json:"icon_type"`
	Keyword  string `form:"keyword" json:"keyword"`
	IsActive *bool  `form:"is_active" json:"is_active"`
}

// CreateIconRequest 创建图标请求
type CreateIconRequest struct {
	Name     string   `json:"name" binding:"required"`
	IconType string   `json:"icon_type" binding:"required,oneof=antd custom svg font"`
	IconData string   `json:"icon_data" binding:"required"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

// UpdateIconRequest 更新图标请求
type UpdateIconRequest struct {
	Name     string   `json:"name"`
	IconType string   `json:"icon_type" binding:"omitempty,oneof=antd custom svg font"`
	IconData string   `json:"icon_data"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	IsActive *bool    `json:"is_active"`
}

// IconCategory 图标分类
type IconCategory struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// IconStats 图标统计
type IconStats struct {
	TotalIcons    int            `json:"total_icons"`
	ActiveIcons   int            `json:"active_icons"`
	InactiveIcons int            `json:"inactive_icons"`
	Categories    []IconCategory `json:"categories"`
	IconTypes     map[string]int `json:"icon_types"`
}
