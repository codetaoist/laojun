package models

import (
	"time"

	"github.com/google/uuid"
)

// Role 角色模型
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description *string   `json:"description" db:"description"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description *string   `json:"description" db:"description"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name"`
	Description *string      `json:"description"`
	IsSystem    bool         `json:"is_system"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Permissions []Permission `json:"permissions,omitempty"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID  uuid.UUID   `json:"user_id" binding:"required"`
	RoleIDs []uuid.UUID `json:"role_ids" binding:"required"`
}

// AssignPermissionsRequest 为角色分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" binding:"required"`
}

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description"`
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	DisplayName string `json:"display_name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description"`
}
