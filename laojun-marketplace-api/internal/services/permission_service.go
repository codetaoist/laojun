package services

import (
	"fmt"

	"github.com/codetaoist/laojun-marketplace-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionService 权限服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限服务
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// CheckUserPermission 检查用户权限
func (s *PermissionService) CheckUserPermission(req models.UserPermissionCheckRequest) (*models.UserPermissionCheckResult, error) {
	// 构建权限检查查询
	var count int64
	query := `
		SELECT COUNT(*)
		FROM az_user_roles ur
		INNER JOIN az_role_permissions rp ON ur.role_id = rp.role_id
		INNER JOIN az_permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1 
		AND p.resource = $2 
		AND p.action = $3
	`
	
	err := s.db.Raw(query, req.UserID, req.Resource, req.Action).Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check user permission: %w", err)
	}

	result := &models.UserPermissionCheckResult{
		UserID:        req.UserID,
		DeviceType:    req.DeviceType,
		Module:        req.Module,
		Resource:      req.Resource,
		Action:        req.Action,
		HasPermission: count > 0,
	}

	// 如果有权限，获取权限详情
	if result.HasPermission {
		var permission models.Permission
		permQuery := `
			SELECT p.*
			FROM az_user_roles ur
			INNER JOIN az_role_permissions rp ON ur.role_id = rp.role_id
			INNER JOIN az_permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1 
			AND p.resource = $2 
			AND p.action = $3
			LIMIT 1
		`
		
		err = s.db.Raw(permQuery, req.UserID, req.Resource, req.Action).Scan(&permission).Error
		if err == nil {
			result.Permission = &permission
		}
	}

	return result, nil
}

// GetUserPermissions 获取用户所有权限
func (s *PermissionService) GetUserPermissions(userID uuid.UUID) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT DISTINCT p.*
		FROM az_permissions p
		INNER JOIN az_role_permissions rp ON p.id = rp.permission_id
		INNER JOIN az_user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.resource, p.action
	`
	
	err := s.db.Raw(query, userID).Scan(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// GetUserRoles 获取用户角色
func (s *PermissionService) GetUserRoles(userID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	query := `
		SELECT r.*
		FROM az_roles r
		INNER JOIN az_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.is_system DESC, r.created_at ASC
	`
	
	err := s.db.Raw(query, userID).Scan(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// HasRole 检查用户是否有指定角色
func (s *PermissionService) HasRole(userID uuid.UUID, roleName string) (bool, error) {
	var count int64
	query := `
		SELECT COUNT(*)
		FROM az_user_roles ur
		INNER JOIN az_roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.name = $2
	`
	
	err := s.db.Raw(query, userID, roleName).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user role: %w", err)
	}

	return count > 0, nil
}

// HasPermission 检查用户是否有指定权限
func (s *PermissionService) HasPermission(userID uuid.UUID, resource, action string) (bool, error) {
	req := models.UserPermissionCheckRequest{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}
	
	result, err := s.CheckUserPermission(req)
	if err != nil {
		return false, err
	}
	
	return result.HasPermission, nil
}

// GetAllPermissions 获取所有权限
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := s.db.Find(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions: %w", err)
	}

	return permissions, nil
}

// GetRolePermissions 获取角色权限
func (s *PermissionService) GetRolePermissions(roleID uuid.UUID) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT p.*
		FROM az_permissions p
		INNER JOIN az_role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	
	err := s.db.Raw(query, roleID).Scan(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// AssignRoleToUser 为用户分配角色
func (s *PermissionService) AssignRoleToUser(userID, roleID uuid.UUID) error {
	// 检查是否已经分配
	var count int64
	err := s.db.Table("az_user_roles").Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check existing role assignment: %w", err)
	}
	
	if count > 0 {
		return fmt.Errorf("role already assigned to user")
	}

	// 分配角色
	userRole := map[string]interface{}{
		"id":      uuid.New(),
		"user_id": userID,
		"role_id": roleID,
	}
	
	err = s.db.Table("az_user_roles").Create(userRole).Error
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	return nil
}

// RemoveRoleFromUser 移除用户角色
func (s *PermissionService) RemoveRoleFromUser(userID, roleID uuid.UUID) error {
	err := s.db.Table("az_user_roles").Where("user_id = ? AND role_id = ?", userID, roleID).Delete(nil).Error
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	return nil
}