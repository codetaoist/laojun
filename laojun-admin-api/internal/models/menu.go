package models

import (
	"time"

	"github.com/google/uuid"
)

// Menu 菜单模型
type Menu struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title" binding:"required,max=100"`
	Path        *string    `json:"path" db:"path"`
	Icon        *string    `json:"icon" db:"icon"`
	Component   *string    `json:"component" db:"component"`
	ParentID    *uuid.UUID `json:"parent_id" db:"parent_id"`
	SortOrder   int        `json:"sort_order" db:"sort_order"`
	IsHidden    bool       `json:"is_hidden" db:"is_hidden"`
	IsFavorite  bool       `json:"is_favorite" db:"is_favorite"`
	DeviceTypes *string    `json:"device_types" db:"device_types"` // JSON字符串，存储适配的设备类型
	Permissions *string    `json:"permissions" db:"permissions"`   // JSON字符串，存储权限要求
	CustomIcon  *string    `json:"custom_icon" db:"custom_icon"`   // 自定义图标URL
	Description *string    `json:"description" db:"description"`   // 菜单描述
	Keywords    *string    `json:"keywords" db:"keywords"`         // 搜索关键词
	Level       int        `json:"level" db:"-"`                   // 菜单层级（计算字段）
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`

	// 关联字段
	Children []*Menu `json:"children,omitempty" db:"-"`
	Parent   *Menu   `json:"parent,omitempty" db:"-"`
}

// MenuCreateRequest 创建菜单请求
type MenuCreateRequest struct {
	Title       string     `json:"title" binding:"required,max=100"`
	Path        *string    `json:"path"`
	Icon        *string    `json:"icon"`
	Component   *string    `json:"component"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   int        `json:"sort_order"`
	IsHidden    bool       `json:"is_hidden"`
	IsFavorite  bool       `json:"is_favorite"`
	DeviceTypes *string    `json:"device_types"`
	Permissions *string    `json:"permissions"`
	CustomIcon  *string    `json:"custom_icon"`
	Description *string    `json:"description"`
	Keywords    *string    `json:"keywords"`
}

// MenuUpdateRequest 更新菜单请求
type MenuUpdateRequest struct {
	Title       *string    `json:"title" binding:"omitempty,max=100"`
	Path        *string    `json:"path"`
	Icon        *string    `json:"icon"`
	Component   *string    `json:"component"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   *int       `json:"sort_order"`
	IsHidden    *bool      `json:"is_hidden"`
	IsFavorite  *bool      `json:"is_favorite"`
	DeviceTypes *string    `json:"device_types"`
	Permissions *string    `json:"permissions"`
	CustomIcon  *string    `json:"custom_icon"`
	Description *string    `json:"description"`
	Keywords    *string    `json:"keywords"`
}

// MenuSearchParams 菜单搜索参数
type MenuSearchParams struct {
	Title      *string `form:"title"`
	Path       *string `form:"path"`
	ParentID   *string `form:"parent_id"`
	IsHidden   *bool   `form:"is_hidden"`
	IsFavorite *bool   `form:"is_favorite"`
	DeviceType *string `form:"device_type"`
	Keywords   *string `form:"keywords"`
	Page       int     `form:"page"`
	PageSize   int     `form:"page_size"`
	Search     string  `form:"search"`
	TreeMode   bool    `form:"tree_mode"` // 是否返回树形结构
}

// DeviceType 设备类型枚举
type DeviceType string

const (
	DeviceTypePC     DeviceType = "pc"     // 电脑端
	DeviceTypeWeb    DeviceType = "web"    // WEB端
	DeviceTypeMobile DeviceType = "mobile" // 手机端
	DeviceTypeWatch  DeviceType = "watch"  // 手表端
	DeviceTypeIoT    DeviceType = "iot"    // 物联网端
	DeviceTypeRobot  DeviceType = "robot"  // 机器人端
)

// MenuConfig 菜单配置
type MenuConfig struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description *string                `json:"description" db:"description"`
	DeviceType  DeviceType             `json:"device_type" db:"device_type"`
	Config      map[string]interface{} `json:"config" db:"config"` // JSON配置
	IsActive    bool                   `json:"is_active" db:"is_active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// MenuBatchOperationRequest 批量操作请求
type MenuBatchOperationRequest struct {
	MenuIDs   []uuid.UUID `json:"menu_ids" binding:"required"`
	Operation string      `json:"operation" binding:"required,oneof=delete hide show favorite unfavorite"`
}

// MenuDragSortRequest 拖拽排序请求
type MenuDragSortRequest struct {
	MenuID       uuid.UUID  `json:"menu_id" binding:"required"`
	NewParentID  *uuid.UUID `json:"new_parent_id"`
	NewSortOrder int        `json:"new_sort_order"`
	TargetMenuID *uuid.UUID `json:"target_menu_id"`                               // 目标菜单ID，用于相对位置排序
	Position     string     `json:"position" binding:"oneof=before after inside"` // 相对位置
}

// MenuTreeNode 菜单树节点
type MenuTreeNode struct {
	*Menu
	Children []*MenuTreeNode `json:"children,omitempty"`
}

// MenuMoveRequest 移动菜单请求
type MenuMoveRequest struct {
	TargetParentID *uuid.UUID `json:"target_parent_id"`
	TargetIndex    int        `json:"target_index"`
}

// MenuBatchUpdateRequest 批量更新菜单请求
type MenuBatchUpdateRequest struct {
	MenuIDs []uuid.UUID `json:"menu_ids" binding:"required"`
	Updates struct {
		IsHidden *bool      `json:"is_hidden"`
		ParentID *uuid.UUID `json:"parent_id"`
	} `json:"updates"`
}

// MenuStats 菜单统计信息
type MenuStats struct {
	TotalMenus   int `json:"total_menus"`
	VisibleMenus int `json:"visible_menus"`
	HiddenMenus  int `json:"hidden_menus"`
	MaxDepth     int `json:"max_depth"`
}
