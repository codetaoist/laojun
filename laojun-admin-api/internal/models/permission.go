package models

import (
	"time"

	"github.com/google/uuid"
)

// DeviceTypeModel 设备类型模型
type DeviceTypeModel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Module 模块模型
type Module struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserGroup 用户组模型
type UserGroup struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserGroupMember 用户组成员模型
type UserGroupMember struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserGroupID uuid.UUID `json:"user_group_id" db:"user_group_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PermissionTemplate 权限模板模型
type PermissionTemplate struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	TemplateData string    `json:"template_data" db:"template_data"` // JSON格式的模板数�?
	IsSystem     bool      `json:"is_system" db:"is_system"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ExtendedPermission 扩展权限模型
type ExtendedPermission struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ModuleID     uuid.UUID `json:"module_id" db:"module_id"`
	DeviceTypeID uuid.UUID `json:"device_type_id" db:"device_type_id"`
	Resource     string    `json:"resource" db:"resource"`
	Action       string    `json:"action" db:"action"`
	Description  string    `json:"description" db:"description"`
	ElementType  *string   `json:"element_type" db:"element_type"`
	ElementCode  *string   `json:"element_code" db:"element_code"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// 关联数据
	Module     *Module          `json:"module,omitempty"`
	DeviceType *DeviceTypeModel `json:"device_type,omitempty"`
}

// PermissionInheritance 权限继承模型
type PermissionInheritance struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	ParentExtendedPermID uuid.UUID `json:"parent_extended_perm_id" db:"parent_extended_perm_id"`
	ChildExtendedPermID  uuid.UUID `json:"child_extended_perm_id" db:"child_extended_perm_id"`
	InheritanceType      string    `json:"inheritance_type" db:"inheritance_type"` // 'include', 'exclude'
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// UserGroupPermission 用户组权限模�?
type UserGroupPermission struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	UserGroupID          uuid.UUID `json:"user_group_id" db:"user_group_id"`
	ExtendedPermissionID uuid.UUID `json:"extended_permission_id" db:"extended_permission_id"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// UserDevicePermission 用户设备权限模型
type UserDevicePermission struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	UserID               uuid.UUID `json:"user_id" db:"user_id"`
	DeviceTypeID         uuid.UUID `json:"device_type_id" db:"device_type_id"`
	ExtendedPermissionID uuid.UUID `json:"extended_permission_id" db:"extended_permission_id"`
	GrantType            string    `json:"grant_type" db:"grant_type"` // 'allow', 'deny'
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// 请求和响应结构体

// DeviceTypeRequest 设备类型请求
type DeviceTypeRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

// ModuleRequest 模块请求
type ModuleRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

// UserGroupRequest 用户组请�?
type UserGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UserGroupMemberRequest 用户组成员请�?
type UserGroupMemberRequest struct {
	UserGroupID uuid.UUID   `json:"user_group_id" binding:"required"`
	UserIDs     []uuid.UUID `json:"user_ids" binding:"required"`
}

// PermissionTemplateRequest 权限模板请求
type PermissionTemplateRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	TemplateData string `json:"template_data" binding:"required"`
}

// ExtendedPermissionRequest 扩展权限请求
type ExtendedPermissionRequest struct {
	ModuleID     uuid.UUID `json:"module_id" binding:"required"`
	DeviceTypeID uuid.UUID `json:"device_type_id" binding:"required"`
	Resource     string    `json:"resource" binding:"required"`
	Action       string    `json:"action" binding:"required"`
	Description  string    `json:"description"`
}

// UserPermissionCheckRequest 用户权限检查请�?
type UserPermissionCheckRequest struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	DeviceType string    `json:"device_type" binding:"required"`
	Module     string    `json:"module" binding:"required"`
	Resource   string    `json:"resource" binding:"required"`
	Action     string    `json:"action" binding:"required"`
}

// UserPermissionCheckResponse 用户权限检查响�?
type UserPermissionCheckResponse struct {
	HasPermission bool   `json:"has_permission"`
	Reason        string `json:"reason,omitempty"`
}

// PermissionSyncRequest 权限同步请求
type PermissionSyncRequest struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	DeviceType string    `json:"device_type" binding:"required"`
	SessionID  string    `json:"session_id"`
}

// PermissionSyncResponse 权限同步响应
type PermissionSyncResponse struct {
	Permissions []ExtendedPermission `json:"permissions"`
	Timestamp   time.Time            `json:"timestamp"`
	Version     string               `json:"version"`
}

// UserGroupWithMembers 带成员的用户组模�?
type UserGroupWithMembers struct {
	UserGroup
	Members []User `json:"members"`
}

// ExtendedPermissionWithDetails 带详细信息的扩展权限
type ExtendedPermissionWithDetails struct {
	ExtendedPermission
	ModuleName     string `json:"module_name"`
	DeviceTypeName string `json:"device_type_name"`
}

// BasicPermission 基础权限模型（供前端抽屉展示�?
type BasicPermission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"isSystem"`
	CreatedAt   time.Time `json:"createdAt"`
}

// 权限查询参数
// 注意：与前端 PermissionQueryParams 对齐
type PermissionQueryParams struct {
	Search   string `json:"search"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	IsSystem *bool  `json:"isSystem"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

// 新增：基础权限创建请求
type BasicPermissionCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

// 新增：基础权限更新请求（可选字段）
type BasicPermissionUpdateRequest struct {
	Name        *string `json:"name"`
	Resource    *string `json:"resource"`
	Action      *string `json:"action"`
	Description *string `json:"description"`
}

// 新增：批量删除请�?
type BatchDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// 新增：权限统计结�?
type PermissionStats struct {
	TotalPermissions  int               `json:"totalPermissions"`
	SystemPermissions int               `json:"systemPermissions"`
	CustomPermissions int               `json:"customPermissions"`
	ByResource        map[string]int    `json:"byResource"`
	ByAction          map[string]int    `json:"byAction"`
	RecentlyCreated   []BasicPermission `json:"recentlyCreated"`
	MostUsed          []BasicPermission `json:"mostUsed"`
}

// 新增：角色与用户概要
type RoleSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName *string   `json:"displayName,omitempty"`
}

type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Name     string    `json:"name"`
}

// 新增：权限使用情况响�?
type PermissionUsageResponse struct {
	IsUsed      bool          `json:"isUsed"`
	UsedByRoles []RoleSummary `json:"usedByRoles"`
	UsedByUsers []UserSummary `json:"usedByUsers"`
}

// 新增：导入结�?
type ImportResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// 新增：同步结果（系统权限同步�?
type SyncResult struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Removed int `json:"removed"`
}
