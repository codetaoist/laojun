package models

import (
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// 使用shared模块的权限相关模型
type Permission = models.Permission
type Role = models.Role
type PaginationMeta = models.PaginationMeta

// UserPermissionCheckRequest 用户权限检查请求
type UserPermissionCheckRequest struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	DeviceType string    `json:"device_type"`
	Module     string    `json:"module"`
	Resource   string    `json:"resource" binding:"required"`
	Action     string    `json:"action" binding:"required"`
}

// UserPermissionCheckResult 用户权限检查结果
type UserPermissionCheckResult struct {
	UserID        uuid.UUID   `json:"user_id"`
	DeviceType    string      `json:"device_type"`
	Module        string      `json:"module"`
	Resource      string      `json:"resource"`
	Action        string      `json:"action"`
	HasPermission bool        `json:"has_permission"`
	Permission    *Permission `json:"permission,omitempty"`
}

// RoleCreateRequest 角色创建请求
type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// RoleUpdateRequest 角色更新请求
type RoleUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
}

// PasswordUpdateRequest 密码更新请求
type PasswordUpdateRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// RoleAssignRequest 角色分配请求
type RoleAssignRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

// PermissionAssignRequest 权限分配请求
type PermissionAssignRequest struct {
	RoleID       uuid.UUID `json:"role_id" binding:"required"`
	PermissionID uuid.UUID `json:"permission_id" binding:"required"`
}