package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/google/uuid"
)

type RoleService struct {
	db           *sql.DB
	auditService *AuditService // 添加审计服务依赖
}

func NewRoleService(db *sql.DB, auditService *AuditService) *RoleService {
	return &RoleService{
		db:           db,
		auditService: auditService,
	}
}

// AssignRolesToUserWithAudit 为用户分配角色（带审计日志）
func (s *RoleService) AssignRolesToUserWithAudit(userID uuid.UUID, roleIDs []uuid.UUID, operatorID uuid.UUID, ipAddress, userAgent string) error {
	// 获取用户信息
	var username string
	err := s.db.QueryRow("SELECT username FROM ua_admin WHERE id = $1", userID).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 获取操作前的角色
	oldRoles, err := s.getUserRolesByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get current user roles: %w", err)
	}

	// 执行角色分配
	if err := s.AssignRolesToUser(userID, roleIDs); err != nil {
		return err
	}

	// 获取操作后的角色
	newRoles, err := s.getUserRolesByID(userID)
	if err != nil {
		// 角色分配成功但获取新角色失败，记录警告但不影响主要操作
		fmt.Printf("Warning: failed to get new user roles for audit: %v\n", err)
		newRoles = []models.Role{} // 使用空切片而不是nil
	}

	// 记录审计日志
	if s.auditService != nil {
		if auditErr := s.auditService.LogRoleAssignment(
			operatorID,
			userID,
			username,
			oldRoles,
			newRoles,
			ipAddress,
			userAgent,
		); auditErr != nil {
			// 审计日志失败不影响主要操作，但应该记录错误
			fmt.Printf("Warning: failed to log role assignment audit: %v\n", auditErr)
		}
	}

	return nil
}

// getUserRolesByID 获取用户的角色列�?
func (s *RoleService) getUserRolesByID(userID uuid.UUID) ([]models.Role, error) {
	query := `
        SELECT r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
        FROM az_roles r
        INNER JOIN az_user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1
        ORDER BY r.name
    `

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRoles 获取角色列表（支持分页与搜索）
func (s *RoleService) GetRoles(page, size int, search string) (*models.PaginatedResponse[models.Role], error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	// 统计总数
	var total int
	countQuery := "SELECT COUNT(*) FROM az_roles"
	var where string
	var args []interface{}
	argIndex := 1
	if search != "" {
		where = fmt.Sprintf(" WHERE name ILIKE $%d OR display_name ILIKE $%d", argIndex, argIndex+1)
		like := "%" + search + "%"
		args = append(args, like, like)
		argIndex += 2
	}

	if err := s.db.QueryRow(countQuery+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count roles: %w", err)
	}

	// 查询列表
	listQuery := fmt.Sprintf(`
        SELECT id, name, display_name, description, is_system, created_at, updated_at
        FROM az_roles%s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, argIndex, argIndex+1)
	args = append(args, size, offset)

	rows, err := s.db.Query(listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var r models.Role
		if err := rows.Scan(&r.ID, &r.Name, &r.DisplayName, &r.Description, &r.IsSystem, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, r)
	}

	totalPages := 0
	if size > 0 {
		totalPages = (total + size - 1) / size
	}

	return &models.PaginatedResponse[models.Role]{
		Data:       roles,
		Total:      int64(total),
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}

// GetRoleByID 获取角色详情
func (s *RoleService) GetRoleByID(id uuid.UUID) (*models.Role, error) {
	var r models.Role
	err := s.db.QueryRow(`
        SELECT id, name, display_name, description, is_system, created_at, updated_at
        FROM az_roles WHERE id = $1
    `, id).Scan(&r.ID, &r.Name, &r.DisplayName, &r.Description, &r.IsSystem, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &r, nil
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(req models.RoleCreateRequest) (*models.Role, error) {
	id := uuid.New()
	now := time.Now()
	_, err := s.db.Exec(`
        INSERT INTO az_roles (id, name, display_name, description, is_system, created_at, updated_at)
        VALUES ($1, $2, $3, $4, false, $5, $6)
    `, id, req.Name, req.DisplayName, nullOrString(req.Description), now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return s.GetRoleByID(id)
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(id uuid.UUID, req models.RoleUpdateRequest) (*models.Role, error) {
	now := time.Now()
	_, err := s.db.Exec(`
        UPDATE az_roles SET display_name = $1, description = $2, updated_at = $3 WHERE id = $4
    `, req.DisplayName, nullOrString(req.Description), now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	return s.GetRoleByID(id)
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM az_roles WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// AssignRolesToUser 为用户分配角色（覆盖式，增强事务处理）
func (s *RoleService) AssignRolesToUser(userID uuid.UUID, roleIDs []uuid.UUID) error {
	// 验证用户是否存在
	var userExists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM ua_admin WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return fmt.Errorf("user not found")
	}

	// 验证所有角色是否存在
	if len(roleIDs) > 0 {
		placeholders := make([]string, len(roleIDs))
		args := make([]interface{}, len(roleIDs))
		for i, roleID := range roleIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = roleID
		}

		var roleCount int
		query := fmt.Sprintf("SELECT COUNT(*) FROM az_roles WHERE id IN (%s)", strings.Join(placeholders, ","))
		err = s.db.QueryRow(query, args...).Scan(&roleCount)
		if err != nil {
			return fmt.Errorf("failed to validate roles: %w", err)
		}
		if roleCount != len(roleIDs) {
			return fmt.Errorf("one or more roles not found")
		}
	}

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 确保事务正确处理
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("transaction failed: %v, rollback failed: %v", err, rollbackErr)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	// 记录操作前的角色状态（用于审计日志）
	var oldRoleIDs []uuid.UUID
	rows, queryErr := tx.Query("SELECT role_id FROM az_user_roles WHERE user_id = $1", userID)
	if queryErr != nil {
		err = fmt.Errorf("failed to query existing roles: %w", queryErr)
		return err
	}
	for rows.Next() {
		var roleID uuid.UUID
		if scanErr := rows.Scan(&roleID); scanErr != nil {
			rows.Close()
			err = fmt.Errorf("failed to scan role ID: %w", scanErr)
			return err
		}
		oldRoleIDs = append(oldRoleIDs, roleID)
	}
	rows.Close()

	// 清空用户现有角色
	if _, execErr := tx.Exec("DELETE FROM az_user_roles WHERE user_id = $1", userID); execErr != nil {
		err = fmt.Errorf("failed to clear user roles: %w", execErr)
		return err
	}

	// 批量插入新角色关联记录
	for _, roleID := range roleIDs {
		if _, execErr := tx.Exec("INSERT INTO az_user_roles (user_id, role_id) VALUES ($1, $2)", userID, roleID); execErr != nil {
			err = fmt.Errorf("failed to assign role %s: %w", roleID, execErr)
			return err
		}
	}

	// 记录审计日志（如果有审计服务）
	// TODO: 添加审计日志记录

	return nil
}

func nullOrString(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// GetRolePermissions 获取角色的权限列表
func (s *RoleService) GetRolePermissions(roleID uuid.UUID) ([]models.Permission, error) {
	query := `
        SELECT p.id, p.name, p.code, p.description, p.resource, p.action, p.created_at
        FROM az_permissions p
        INNER JOIN az_role_permissions rp ON p.id = rp.permission_id
        WHERE rp.role_id = $1
        ORDER BY p.resource, p.action
    `
	rows, err := s.db.Query(query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}
	defer rows.Close()

	var perms []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.Description, &p.Resource, &p.Action, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// AssignPermissionsToRole 为角色分配权限（覆盖式）
func (s *RoleService) AssignPermissionsToRole(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 清空角色现有权限
	if _, err = tx.Exec("DELETE FROM az_role_permissions WHERE role_id = $1", roleID); err != nil {
		return fmt.Errorf("failed to clear role permissions: %w", err)
	}

	// 批量插入新权限关联
	for _, pid := range permissionIDs {
		if _, err = tx.Exec("INSERT INTO az_role_permissions (role_id, permission_id) VALUES ($1, $2)", roleID, pid); err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}
	}

	return nil
}

func (s *RoleService) GetUserByID(id uuid.UUID) (*models.UserResponse, error) {
	var user models.UserResponse
	err := s.db.QueryRow(`
        SELECT id, username, email, avatar, is_active, created_at, updated_at, last_login_at
        FROM ua_admin WHERE id = $1
    `, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Avatar,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	roles, err := s.getUserRolesByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	user.Roles = roles

	return &user, nil
}
