package services

import (
	"fmt"
	"time"

	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleService 角色服务
type RoleService struct {
	db *gorm.DB
}

// NewRoleService 创建角色服务
func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(name, description string, isSystem bool) (*models.Role, error) {
	// 检查角色名是否已存在
	var existingRole models.Role
	err := s.db.Where("name = ?", name).First(&existingRole).Error
	if err == nil {
		return nil, fmt.Errorf("role with name '%s' already exists", name)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing role: %w", err)
	}

	role := &models.Role{
		ID:          uuid.New(),
		Name:        name,
		Description: &description,
		IsSystem:    isSystem,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.db.Create(role).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRole 获取角色
func (s *RoleService) GetRole(roleID uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := s.db.Where("id = ?", roleID).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// GetRoleByName 根据名称获取角色
func (s *RoleService) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := s.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(roleID uuid.UUID, name, description *string) (*models.Role, error) {
	var role models.Role
	err := s.db.Where("id = ?", roleID).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// 检查系统角色是否可以修改
	if role.IsSystem {
		return nil, fmt.Errorf("system role cannot be modified")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if name != nil {
		// 检查新名称是否已存在
		var existingRole models.Role
		err = s.db.Where("name = ? AND id != ?", *name, roleID).First(&existingRole).Error
		if err == nil {
			return nil, fmt.Errorf("role with name '%s' already exists", *name)
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing role: %w", err)
		}
		updates["name"] = *name
	}

	if description != nil {
		updates["description"] = *description
	}

	err = s.db.Model(&role).Updates(updates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return &role, nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(roleID uuid.UUID) error {
	var role models.Role
	err := s.db.Where("id = ?", roleID).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查系统角色是否可以删除
	if role.IsSystem {
		return fmt.Errorf("system role cannot be deleted")
	}

	// 检查是否有用户使用此角色
	var userCount int64
	err = s.db.Table("az_user_roles").Where("role_id = ?", roleID).Count(&userCount).Error
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("role is in use by %d users, cannot delete", userCount)
	}

	// 删除角色权限关联
	err = s.db.Table("az_role_permissions").Where("role_id = ?", roleID).Delete(nil).Error
	if err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	// 删除角色
	err = s.db.Delete(&role).Error
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// GetAllRoles 获取所有角色
func (s *RoleService) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Order("is_system DESC, created_at ASC").Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return roles, nil
}

// GetRoles 分页获取角色
func (s *RoleService) GetRoles(page, limit int) ([]models.Role, *models.PaginationMeta, error) {
	var roles []models.Role
	var total int64

	// 计算总数
	err := s.db.Model(&models.Role{}).Count(&total).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count roles: %w", err)
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取角色列表
	err = s.db.Order("is_system DESC, created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&roles).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get roles: %w", err)
	}

	// 计算分页信息
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	
	meta := models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return roles, meta, nil
}

// AssignPermissionToRole 为角色分配权限
func (s *RoleService) AssignPermissionToRole(roleID, permissionID uuid.UUID) error {
	// 检查角色是否存在
	var role models.Role
	err := s.db.Where("id = ?", roleID).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查权限是否存在
	var permission models.Permission
	err = s.db.Where("id = ?", permissionID).First(&permission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("permission not found")
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// 检查是否已经分配
	var count int64
	err = s.db.Table("az_role_permissions").Where("role_id = ? AND permission_id = ?", roleID, permissionID).Count(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check existing permission assignment: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("permission already assigned to role")
	}

	// 分配权限
	rolePermission := map[string]interface{}{
		"id":            uuid.New(),
		"role_id":       roleID,
		"permission_id": permissionID,
	}

	err = s.db.Table("az_role_permissions").Create(rolePermission).Error
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole 移除角色权限
func (s *RoleService) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	err := s.db.Table("az_role_permissions").Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(nil).Error
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	return nil
}

// GetRoleUsers 获取角色下的用户
func (s *RoleService) GetRoleUsers(roleID uuid.UUID) ([]models.User, error) {
	var users []models.User
	query := `
		SELECT u.*
		FROM az_users u
		INNER JOIN az_user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1
		ORDER BY u.created_at DESC
	`

	err := s.db.Raw(query, roleID).Scan(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}

	return users, nil
}