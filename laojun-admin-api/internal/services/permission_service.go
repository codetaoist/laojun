package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/cache"
	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type PermissionService struct {
	db    *sql.DB
	cache *cache.PermissionCache
}

func NewPermissionService(db *sql.DB, redisClient *redis.Client) *PermissionService {
	var permCache *cache.PermissionCache
	if redisClient != nil {
		permCache = cache.NewPermissionCache(redisClient)
	}
	return &PermissionService{
		db:    db,
		cache: permCache,
	}
}

// DeviceType 相关方法

func (s *PermissionService) GetDeviceTypes() ([]models.DeviceTypeModel, error) {
	query := `
		SELECT id, code, name, description, icon, sort_order, is_active, created_at, updated_at
		FROM sm_device_types
		WHERE is_active = true
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deviceTypes []models.DeviceTypeModel
	for rows.Next() {
		var dt models.DeviceTypeModel
		err := rows.Scan(&dt.ID, &dt.Code, &dt.Name, &dt.Description, &dt.Icon,
			&dt.SortOrder, &dt.IsActive, &dt.CreatedAt, &dt.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deviceTypes = append(deviceTypes, dt)
	}

	return deviceTypes, nil
}

func (s *PermissionService) CreateDeviceType(req models.DeviceTypeRequest) (*models.DeviceTypeModel, error) {
	query := `
		INSERT INTO sm_device_types (code, name, description, icon, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, name, description, icon, sort_order, is_active, created_at, updated_at
	`

	var dt models.DeviceTypeModel
	err := s.db.QueryRow(query, req.Code, req.Name, req.Description, req.Icon, req.SortOrder).
		Scan(&dt.ID, &dt.Code, &dt.Name, &dt.Description, &dt.Icon,
			&dt.SortOrder, &dt.IsActive, &dt.CreatedAt, &dt.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &dt, nil
}

// Module 相关方法

func (s *PermissionService) GetModules() ([]models.Module, error) {
	query := `
		SELECT id, code, name, description, icon, sort_order, is_active, created_at, updated_at
		FROM sm_modules
		WHERE is_active = true
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []models.Module
	for rows.Next() {
		var m models.Module
		err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.Description, &m.Icon,
			&m.SortOrder, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}

	return modules, nil
}

func (s *PermissionService) CreateModule(req models.ModuleRequest) (*models.Module, error) {
	query := `
		INSERT INTO sm_modules (code, name, description, icon, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, name, description, icon, sort_order, is_active, created_at, updated_at
	`

	var m models.Module
	err := s.db.QueryRow(query, req.Code, req.Name, req.Description, req.Icon, req.SortOrder).
		Scan(&m.ID, &m.Code, &m.Name, &m.Description, &m.Icon,
			&m.SortOrder, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

// UserGroup 相关方法

func (s *PermissionService) GetUserGroups() ([]models.UserGroupWithMembers, error) {
	query := `
		SELECT ug.id, ug.name, ug.description, ug.is_active, ug.created_at, ug.updated_at
		FROM ug_user_groups ug
		WHERE ug.is_active = true
		ORDER BY ug.created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userGroups []models.UserGroupWithMembers
	for rows.Next() {
		var ug models.UserGroupWithMembers
		err := rows.Scan(&ug.ID, &ug.Name, &ug.Description, &ug.IsActive,
			&ug.CreatedAt, &ug.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// 获取用户组成员
		members, err := s.getUserGroupMembers(ug.ID)
		if err != nil {
			return nil, err
		}
		ug.Members = members

		userGroups = append(userGroups, ug)
	}

	return userGroups, nil
}

func (s *PermissionService) getUserGroupMembers(userGroupID uuid.UUID) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.is_active, u.created_at, u.updated_at
		FROM ua_admin u
		INNER JOIN ug_user_group_members ugm ON u.id = ugm.user_id
		WHERE ugm.user_group_id = $1 AND u.is_active = true
		ORDER BY u.username ASC
	`

	rows, err := s.db.Query(query, userGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (s *PermissionService) CreateUserGroup(req models.UserGroupRequest) (*models.UserGroup, error) {
	query := `
		INSERT INTO ug_user_groups (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, is_active, created_at, updated_at
	`

	var ug models.UserGroup
	err := s.db.QueryRow(query, req.Name, req.Description).
		Scan(&ug.ID, &ug.Name, &ug.Description, &ug.IsActive,
			&ug.CreatedAt, &ug.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &ug, nil
}

func (s *PermissionService) AddUsersToGroup(req models.UserGroupMemberRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 先删除现有组成员
	_, err = tx.Exec("DELETE FROM ug_user_group_members WHERE user_group_id = $1", req.UserGroupID)
	if err != nil {
		return err
	}

	// 添加新组成员
	for _, userID := range req.UserIDs {
		_, err = tx.Exec(`
			INSERT INTO ug_user_group_members (user_group_id, user_id)
			VALUES ($1, $2)
		`, req.UserGroupID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ExtendedPermission 相关方法

func (s *PermissionService) GetExtendedPermissions(moduleID, deviceTypeID uuid.UUID) ([]models.ExtendedPermissionWithDetails, error) {
	query := `
		SELECT ep.id, ep.module_id, ep.device_type_id, ep.resource, ep.action, ep.description,
			   ep.element_type, ep.element_code, ep.created_at,
			   m.name as module_name, dt.name as device_type_name
		FROM pe_extended_permissions ep
		INNER JOIN sm_modules m ON ep.module_id = m.id
		INNER JOIN sm_device_types dt ON ep.device_type_id = dt.id
		WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	if moduleID != uuid.Nil {
		query += fmt.Sprintf(" AND ep.module_id = $%d", argIndex)
		args = append(args, moduleID)
		argIndex++
	}

	if deviceTypeID != uuid.Nil {
		query += fmt.Sprintf(" AND ep.device_type_id = $%d", argIndex)
		args = append(args, deviceTypeID)
		argIndex++
	}

	query += " ORDER BY m.sort_order, dt.sort_order, ep.resource, ep.action"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.ExtendedPermissionWithDetails
	for rows.Next() {
		var ep models.ExtendedPermissionWithDetails
		err := rows.Scan(&ep.ID, &ep.ModuleID, &ep.DeviceTypeID, &ep.Resource, &ep.Action,
			&ep.Description, &ep.ElementType, &ep.ElementCode, &ep.CreatedAt,
			&ep.ModuleName, &ep.DeviceTypeName)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, ep)
	}

	return permissions, nil
}

func (s *PermissionService) CreateExtendedPermission(req models.ExtendedPermissionRequest) (*models.ExtendedPermission, error) {
	query := `
		INSERT INTO pe_extended_permissions (module_id, device_type_id, resource, action, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, module_id, device_type_id, resource, action, description, element_type, element_code, created_at
	`

	var ep models.ExtendedPermission
	err := s.db.QueryRow(query, req.ModuleID, req.DeviceTypeID, req.Resource, req.Action, req.Description).
		Scan(&ep.ID, &ep.ModuleID, &ep.DeviceTypeID, &ep.Resource, &ep.Action,
			&ep.Description, &ep.ElementType, &ep.ElementCode, &ep.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &ep, nil
}

// 权限检查相关方法

// CheckUserPermission 检查用户权限
func (s *PermissionService) CheckUserPermission(req models.UserPermissionCheckRequest) (*models.UserPermissionCheckResponse, error) {
	// 首先尝试从缓存获取结果
	if s.cache != nil {
		hasPermission, err := s.cache.GetUserPermission(req.UserID, req.Resource, req.Action)
		if err == nil {
			return &models.UserPermissionCheckResponse{
				HasPermission: hasPermission,
				Reason:        "从缓存获取",
			}, nil
		}
	}

	// 缓存未命中，执行实际权限检查
	result, err := s.performPermissionCheck(req)
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if s.cache != nil {
		expiry := cache.DefaultCacheExpiry
		// 对于超级管理员权限，使用更长的缓存时间
		if result.Reason == "超级管理员权限" {
			expiry = cache.LongCacheExpiry
		}

		s.cache.SetUserPermission(req.UserID, req.Resource, req.Action, result.HasPermission, expiry)
	}

	return result, nil
}

// performPermissionCheck 执行实际的权限检查逻辑
func (s *PermissionService) performPermissionCheck(req models.UserPermissionCheckRequest) (*models.UserPermissionCheckResponse, error) {
	// 0. 首先检查用户是否是超级管理员
	isSuperAdmin, err := s.checkSuperAdminRole(req.UserID)
	if err != nil {
		return nil, err
	}

	if isSuperAdmin {
		return &models.UserPermissionCheckResponse{
			HasPermission: true,
			Reason:        "超级管理员权限",
		}, nil
	}

	// 1. 检查用户是否有直接权限
	hasDirectPermission, err := s.checkDirectPermission(req)
	if err != nil {
		return nil, err
	}

	if hasDirectPermission {
		return &models.UserPermissionCheckResponse{
			HasPermission: true,
			Reason:        "直接权限",
		}, nil
	}

	// 2. 检查用户组权限
	hasGroupPermission, err := s.checkGroupPermission(req)
	if err != nil {
		return nil, err
	}

	if hasGroupPermission {
		return &models.UserPermissionCheckResponse{
			HasPermission: true,
			Reason:        "用户组权限",
		}, nil
	}

	// 3. 检查角色权限
	hasRolePermission, err := s.checkRolePermission(req)
	if err != nil {
		return nil, err
	}

	if hasRolePermission {
		return &models.UserPermissionCheckResponse{
			HasPermission: true,
			Reason:        "角色权限",
		}, nil
	}

	return &models.UserPermissionCheckResponse{
		HasPermission: false,
		Reason:        "无权访问",
	}, nil
}

func (s *PermissionService) checkDirectPermission(req models.UserPermissionCheckRequest) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM pe_user_device_permissions udp
		INNER JOIN sm_device_types dt ON udp.device_type_id = dt.id
		WHERE udp.user_id = $1 
		  AND dt.code = $2 
		  AND udp.permissions ? $3
		  AND udp.permissions->$3 ? $4
		  AND udp.is_active = true
		  AND (udp.expires_at IS NULL OR udp.expires_at > NOW())
	`

	// 构建权限键，格式为 "module.resource"
	permissionKey := req.Module + "." + req.Resource

	var count int
	err := s.db.QueryRow(query, req.UserID, req.DeviceType, permissionKey, req.Action).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *PermissionService) checkGroupPermission(req models.UserPermissionCheckRequest) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM ug_user_group_members ugm
		INNER JOIN ug_user_group_permissions ugp ON ugm.group_id = ugp.group_id
		INNER JOIN az_permissions p ON ugp.permission_id = p.id
		INNER JOIN sm_device_types dt ON ugp.device_type_id = dt.id
		WHERE ugm.user_id = $1 
		  AND dt.code = $2 
		  AND p.resource = $3 
		  AND p.action = $4
		  AND ugp.granted = true
	`

	var count int
	err := s.db.QueryRow(query, req.UserID, req.DeviceType, req.Resource, req.Action).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *PermissionService) checkRolePermission(req models.UserPermissionCheckRequest) (bool, error) {
	// 这里可以实现基于传统角色权限的检查
	// 将传统权限映射到扩展权限
	permissionCode := fmt.Sprintf("%s:%s", strings.ToLower(req.Resource), strings.ToLower(req.Action))

	query := `
		SELECT COUNT(*)
		FROM az_user_roles ur
		INNER JOIN az_role_permissions rp ON ur.role_id = rp.role_id
		INNER JOIN az_permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1 AND p.code = $2
	`

	var count int
	err := s.db.QueryRow(query, req.UserID, permissionCode).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// checkSuperAdminRole 检查用户是否是超级管理员角色
func (s *PermissionService) checkSuperAdminRole(userID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM az_user_roles ur
		INNER JOIN az_roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.name = 'super_admin'
	`

	var count int
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 权限同步相关方法

func (s *PermissionService) SyncUserPermissions(req models.PermissionSyncRequest) (*models.PermissionSyncResponse, error) {
	query := `
		SELECT DISTINCT ep.id, ep.module_id, ep.device_type_id, ep.resource, ep.action, 
			   ep.description, ep.element_type, ep.element_code, ep.created_at
		FROM pe_extended_permissions ep
		INNER JOIN sm_device_types dt ON ep.device_type_id = dt.id
		WHERE dt.code = $1
		  AND (
			-- 直接权限
			EXISTS (
				SELECT 1 FROM pe_user_device_permissions udp 
				WHERE udp.user_id = $2 AND udp.extended_permission_id = ep.id AND udp.grant_type = 'allow'
			)
			OR
			-- 用户组权限
			EXISTS (
				SELECT 1 FROM ug_user_group_members ugm
				INNER JOIN ug_user_group_permissions ugp ON ugm.user_group_id = ugp.user_group_id
				WHERE ugm.user_id = $2 AND ugp.extended_permission_id = ep.id
			)
		  )
		ORDER BY ep.module_id, ep.resource, ep.action
	`

	rows, err := s.db.Query(query, req.DeviceType, req.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.ExtendedPermission
	for rows.Next() {
		var ep models.ExtendedPermission
		err := rows.Scan(&ep.ID, &ep.ModuleID, &ep.DeviceTypeID, &ep.Resource, &ep.Action,
			&ep.Description, &ep.ElementType, &ep.ElementCode, &ep.CreatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, ep)
	}

	return &models.PermissionSyncResponse{
		Permissions: permissions,
		Timestamp:   time.Now(),
		Version:     fmt.Sprintf("v%d", time.Now().Unix()),
	}, nil
}

// PermissionTemplate 相关方法

func (s *PermissionService) GetPermissionTemplates() ([]models.PermissionTemplate, error) {
	query := `
		SELECT id, name, description, template_data, is_system, created_at, updated_at
		FROM ug_permission_templates
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.PermissionTemplate
	for rows.Next() {
		var pt models.PermissionTemplate
		err := rows.Scan(&pt.ID, &pt.Name, &pt.Description, &pt.TemplateData,
			&pt.IsSystem, &pt.CreatedAt, &pt.UpdatedAt)
		if err != nil {
			return nil, err
		}
		templates = append(templates, pt)
	}

	return templates, nil
}

func (s *PermissionService) CreatePermissionTemplate(req models.PermissionTemplateRequest) (*models.PermissionTemplate, error) {
	// 验证 JSON 格式
	var templateData interface{}
	if err := json.Unmarshal([]byte(req.TemplateData), &templateData); err != nil {
		return nil, fmt.Errorf("无效的模板数据格式: %v", err)
	}

	query := `
		INSERT INTO ug_permission_templates (name, description, template_data)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, template_data, is_system, created_at, updated_at
	`

	var pt models.PermissionTemplate
	err := s.db.QueryRow(query, req.Name, req.Description, req.TemplateData).
		Scan(&pt.ID, &pt.Name, &pt.Description, &pt.TemplateData,
			&pt.IsSystem, &pt.CreatedAt, &pt.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &pt, nil
}

func (s *PermissionService) ApplyPermissionTemplate(templateID, userID uuid.UUID) error {
	// 获取模板数据
	var templateData string
	err := s.db.QueryRow("SELECT template_data FROM ug_permission_templates WHERE id = $1 AND is_system = true", templateID).Scan(&templateData)
	if err != nil {
		return err
	}

	// 解析模板数据并应用权限
	var template map[string]interface{}
	if err := json.Unmarshal([]byte(templateData), &template); err != nil {
		return err
	}

	// 这里可以根据模板数据的结构来实现具体的权限应用逻辑
	// 例如：根据模板中的模块、动作等信息来分配相应的扩展权限

	// 应用权限模板后，清除用户缓存
	if s.cache != nil {
		s.cache.InvalidateAllUserCache(userID)
	}

	return nil
}

// InvalidateUserCache 清除用户缓存
func (s *PermissionService) InvalidateUserCache(userID uuid.UUID) error {
	if s.cache != nil {
		return s.cache.InvalidateAllUserCache(userID)
	}
	return nil
}

// InvalidateRoleCache 清除角色缓存
func (s *PermissionService) InvalidateRoleCache(roleID uuid.UUID) error {
	if s.cache != nil {
		return s.cache.InvalidateRolePermissions(roleID)
	}
	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *PermissionService) GetCacheStats() (map[string]interface{}, error) {
	if s.cache != nil {
		return s.cache.GetCacheStats()
	}
	return map[string]interface{}{
		"cache_enabled": false,
		"message":       "缓存未启用",
	}, nil
}

// WarmupUserCache 预热用户缓存
func (s *PermissionService) WarmupUserCache(userID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}

	// 获取用户的常用权限并预热缓存
	// 这里可以根据实际业务需求实现预热逻辑
	return nil
}

// GetBasicPermissions 获取基础权限列表，供前端抽屉展示
func (s *PermissionService) GetBasicPermissions() ([]models.BasicPermission, error) {
	query := `
		SELECT id, name, resource, action, description, false AS is_system, created_at
		FROM az_permissions
		ORDER BY resource, action
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var perms []models.BasicPermission
	for rows.Next() {
		var p models.BasicPermission
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description, &p.IsSystem, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// GetBasicPermissionsFiltered 获取基础权限列表（支持搜索与资源/动作过滤）
func (s *PermissionService) GetBasicPermissionsFiltered(keyword, resource, action string) ([]models.BasicPermission, error) {
	query := `
		SELECT id, name, resource, action, description, false AS is_system, created_at
		FROM az_permissions
		WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	if keyword != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1)
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
		argIndex += 2
	}
	if resource != "" {
		query += fmt.Sprintf(" AND resource = $%d", argIndex)
		args = append(args, resource)
		argIndex++
	}
	if action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	query += " ORDER BY resource, action"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var perms []models.BasicPermission
	for rows.Next() {
		var p models.BasicPermission
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description, &p.IsSystem, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// removed duplicate GetPermissionsPaginated (deprecated copy)
func (s *PermissionService) GetPermissionsPaginated(params models.PermissionQueryParams) ([]models.BasicPermission, int, error) {
	// 构建过滤条件
	whereClauses := []string{"1=1"}
	var args []interface{}
	argIndex := 1

	if params.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d OR resource ILIKE $%d OR action ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		// 使用一个参数匹配四个字段的 ILIKE
		args = append(args, "%"+params.Search+"%")
		argIndex++
	}
	if params.Resource != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, params.Resource)
		argIndex++
	}
	if params.Action != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, params.Action)
		argIndex++
	}
	// params.IsSystem 被忽略：az_permissions 表中没有 is_system 字段

	base := "FROM az_permissions WHERE " + strings.Join(whereClauses, " AND ")
	// 统计总数
	countQuery := "SELECT COUNT(*) " + base
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// 分页参数
	page := params.Page
	if page <= 0 {
		page = 1
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	// 查询数据
	dataQuery := "SELECT id, name, resource, action, description, false AS is_system, created_at " + base + fmt.Sprintf(" ORDER BY resource, action LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	rows, err := s.db.Query(dataQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var perms []models.BasicPermission
	for rows.Next() {
		var p models.BasicPermission
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description, &p.IsSystem, &p.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}

	return perms, total, nil
}

// 创建基础权限
func (s *PermissionService) CreateBasicPermission(req models.BasicPermissionCreateRequest) (*models.BasicPermission, error) {
	// 检查重复（resource + action 唯一）
	var exists int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM az_permissions WHERE resource = $1 AND action = $2", req.Resource, req.Action).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, fmt.Errorf("权限已存在: %s:%s", req.Resource, req.Action)
	}

	code := fmt.Sprintf("%s:%s", strings.ToLower(req.Resource), strings.ToLower(req.Action))
	query := `INSERT INTO az_permissions (name, resource, action, description, code)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, name, resource, action, description, false AS is_system, created_at`
	var p models.BasicPermission
	if err := s.db.QueryRow(query, req.Name, req.Resource, req.Action, req.Description, code).
		Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description, &p.IsSystem, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

// 获取基础权限详情
func (s *PermissionService) GetBasicPermissionByID(id string) (*models.BasicPermission, error) {
	query := `SELECT id, name, resource, action, description, false AS is_system, created_at FROM az_permissions WHERE id = $1`
	var p models.BasicPermission
	if err := s.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description, &p.IsSystem, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

// 更新基础权限
func (s *PermissionService) UpdateBasicPermission(id string, req models.BasicPermissionUpdateRequest) (*models.BasicPermission, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *req.Name)
		idx++
	}
	if req.Resource != nil {
		sets = append(sets, fmt.Sprintf("resource = $%d", idx))
		args = append(args, *req.Resource)
		idx++
	}
	if req.Action != nil {
		sets = append(sets, fmt.Sprintf("action = $%d", idx))
		args = append(args, *req.Action)
		idx++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", idx))
		args = append(args, *req.Description)
		idx++
	}

	// 如果资源或动作变化，更新 code
	if req.Resource != nil || req.Action != nil {
		var curRes, curAct string
		if err := s.db.QueryRow("SELECT resource, action FROM az_permissions WHERE id = $1", id).Scan(&curRes, &curAct); err != nil {
			return nil, err
		}
		if req.Resource != nil {
			curRes = *req.Resource
		}
		if req.Action != nil {
			curAct = *req.Action
		}
		code := fmt.Sprintf("%s:%s", strings.ToLower(curRes), strings.ToLower(curAct))
		sets = append(sets, fmt.Sprintf("code = $%d", idx))
		args = append(args, code)
		idx++
	}

	if len(sets) == 0 {
		// 无需更新，直接返回当前详情
		return s.GetBasicPermissionByID(id)
	}

	query := "UPDATE az_permissions SET " + strings.Join(sets, ", ") + fmt.Sprintf(" WHERE id = $%d", idx)
	args = append(args, id)
	if _, err := s.db.Exec(query, args...); err != nil {
		return nil, err
	}
	return s.GetBasicPermissionByID(id)
}

// 删除基础权限
func (s *PermissionService) DeleteBasicPermission(id string) error {
	// 删除角色-权限关联，避免外键约束
	_, _ = s.db.Exec("DELETE FROM az_role_permissions WHERE permission_id = $1", id)
	_, err := s.db.Exec("DELETE FROM az_permissions WHERE id = $1", id)
	return err
}

// 批量删除基础权限
func (s *PermissionService) BatchDeletePermissions(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	placeholders := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))
	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}

	// 删除角色-权限关联
	if _, err := tx.Exec("DELETE FROM az_role_permissions WHERE permission_id IN ("+strings.Join(placeholders, ",")+")", args...); err != nil {
		return err
	}
	// 删除权限
	if _, err := tx.Exec("DELETE FROM az_permissions WHERE id IN ("+strings.Join(placeholders, ",")+")", args...); err != nil {
		return err
	}

	return tx.Commit()
}

// 资源列表
func (s *PermissionService) GetDistinctResources() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT resource FROM az_permissions ORDER BY resource")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

// 动作列表
func (s *PermissionService) GetDistinctActions() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT action FROM az_permissions ORDER BY action")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

// 权限使用情况
func (s *PermissionService) CheckPermissionUsage(id string) (*models.PermissionUsageResponse, error) {
	// 角色使用
	roleRows, err := s.db.Query(`
		SELECT r.id, r.name, r.display_name
		FROM az_roles r
		INNER JOIN az_role_permissions rp ON rp.role_id = r.id
		WHERE rp.permission_id = $1
		ORDER BY r.name ASC`, id)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()
	roles := []models.RoleSummary{}
	for roleRows.Next() {
		var rs models.RoleSummary
		if err := roleRows.Scan(&rs.ID, &rs.Name, &rs.DisplayName); err != nil {
			return nil, err
		}
		roles = append(roles, rs)
	}

	// 用户使用（通过角色关联）
	userRows, err := s.db.Query(`
		SELECT u.id, u.username, COALESCE(u.name, u.username) as name
		FROM ua_admin u
		INNER JOIN az_user_roles ur ON ur.user_id = u.id
		INNER JOIN az_role_permissions rp ON rp.role_id = ur.role_id
		WHERE rp.permission_id = $1
		ORDER BY u.username ASC`, id)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()
	users := []models.UserSummary{}
	for userRows.Next() {
		var us models.UserSummary
		if err := userRows.Scan(&us.ID, &us.Username, &us.Name); err != nil {
			return nil, err
		}
		users = append(users, us)
	}

	return &models.PermissionUsageResponse{
		IsUsed:      len(roles) > 0 || len(users) > 0,
		UsedByRoles: roles,
		UsedByUsers: users,
	}, nil
}

// 权限统计
func (s *PermissionService) GetPermissionStats() (*models.PermissionStats, error) {
	stats := &models.PermissionStats{
		ByResource: map[string]int{},
		ByAction:   map[string]int{},
	}
	// 总数
	if err := s.db.QueryRow("SELECT COUNT(*) FROM az_permissions").Scan(&stats.TotalPermissions); err != nil {
		return nil, err
	}
	stats.SystemPermissions = 0
	stats.CustomPermissions = stats.TotalPermissions - stats.SystemPermissions
	// 按资源统计
	resRows, err := s.db.Query("SELECT resource, COUNT(*) FROM az_permissions GROUP BY resource")
	if err != nil {
		return nil, err
	}
	defer resRows.Close()
	for resRows.Next() {
		var k string
		var v int
		if err := resRows.Scan(&k, &v); err != nil {
			return nil, err
		}
		stats.ByResource[k] = v
	}
	// 按动作统计
	actionRows, err := s.db.Query("SELECT action, COUNT(*) FROM az_permissions GROUP BY action")
	if err != nil {
		return nil, err
	}
	defer actionRows.Close()
	for actionRows.Next() {
		var k string
		var v int
		if err := actionRows.Scan(&k, &v); err != nil {
			return nil, err
		}
		stats.ByAction[k] = v
	}
	// 最近创建的权限（不包括系统权限）
	recentRows, err := s.db.Query(`SELECT id, name, resource, action, description, false AS is_system, created_at FROM az_permissions ORDER BY created_at DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var p models.BasicPermission
		if err := recentRows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description, &p.IsSystem, &p.CreatedAt); err != nil {
			return nil, err
		}
		stats.RecentlyCreated = append(stats.RecentlyCreated, p)
	}
	// 最常用（被角色引用次数最多）
	usedRows, err := s.db.Query(`
		SELECT p.id, p.name, p.resource, p.action, p.description, false AS is_system, p.created_at, COUNT(rp.role_id) as cnt
		FROM az_permissions p
		LEFT JOIN az_role_permissions rp ON rp.permission_id = p.id
		GROUP BY p.id, p.name, p.resource, p.action, p.description, p.created_at
		ORDER BY cnt DESC NULLS LAST, p.resource ASC, p.action ASC
		LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer usedRows.Close()
	for usedRows.Next() {
		var p models.BasicPermission
		var cnt int
		if err := usedRows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description, &p.IsSystem, &p.CreatedAt, &cnt); err != nil {
			return nil, err
		}
		stats.MostUsed = append(stats.MostUsed, p)
	}
	return stats, nil
}

// 导出权限
func (s *PermissionService) ExportPermissions(format string) ([]byte, string, string, error) {
	perms, err := s.GetBasicPermissions()
	if err != nil {
		return nil, "", "", err
	}
	if strings.ToLower(format) == "csv" {
		var b strings.Builder
		b.WriteString("id,name,resource,action,description,is_system,created_at\n")
		for _, p := range perms {
			// 简单CSV（未严格转义逗号与引号）
			b.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%t,%s\n",
				p.ID, p.Name, p.Resource, p.Action, strings.ReplaceAll(p.Description, ",", " "), p.IsSystem, p.CreatedAt.Format(time.RFC3339)))
		}
		return []byte(b.String()), "text/csv", "permissions.csv", nil
	}
	// 默认 JSON
	data, err := json.Marshal(perms)
	if err != nil {
		return nil, "", "", err
	}
	return data, "application/json", "permissions.json", nil
}

// 导入权限（仅支持 JSON 格式）
func (s *PermissionService) ImportPermissions(data []byte) (*models.ImportResult, error) {
	result := &models.ImportResult{Errors: []string{}}
	var items []models.BasicPermission
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("无效的JSON格式: %w", err)
	}
	for _, it := range items {
		var count int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM az_permissions WHERE resource = $1 AND action = $2", it.Resource, it.Action).Scan(&count); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("查询失败 %s:%s: %v", it.Resource, it.Action, err))
			continue
		}
		code := fmt.Sprintf("%s:%s", strings.ToLower(it.Resource), strings.ToLower(it.Action))
		if count == 0 {
			// 插入
			_, err := s.db.Exec(`INSERT INTO az_permissions (name, resource, action, description, code) VALUES ($1, $2, $3, $4, $5)`,
				it.Name, it.Resource, it.Action, it.Description, code)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("插入失败 %s:%s: %v", it.Resource, it.Action, err))
				continue
			}
			result.Success++
		} else {
			// 更新名称与描述
			_, err := s.db.Exec(`UPDATE az_permissions SET name = $1, description = $2, code = $3 WHERE resource = $4 AND action = $5`,
				it.Name, it.Description, code, it.Resource, it.Action)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("更新失败 %s:%s: %v", it.Resource, it.Action, err))
				continue
			}
			result.Success++
		}
	}
	return result, nil
}

// 同步系统权限（占位实现）
func (s *PermissionService) SyncSystemPermissions() (*models.SyncResult, error) {
	// 实际逻辑可根据系统内置权限来源进行比对补全
	return &models.SyncResult{Added: 0, Updated: 0, Removed: 0}, nil
}
