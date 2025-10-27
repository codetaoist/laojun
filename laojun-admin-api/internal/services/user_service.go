package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(page, size int, search string) (*models.UserListResponse, error) {
	offset := (page - 1) * size

	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClause += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex+1)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ua_admin %s", whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 查询用户列表
	query := fmt.Sprintf(`
		SELECT id, username, email, avatar, is_active, created_at, updated_at, last_login_at
		FROM ua_admin %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, size, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.UserResponse
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Avatar,
			&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return &models.UserListResponse{
		Users: users,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uuid.UUID) (*models.UserResponse, error) {
	query := `
		SELECT id, username, email, avatar, is_active, created_at, updated_at, last_login_at
		FROM ua_admin WHERE id = $1
	`

	var user models.UserResponse
	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Avatar,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取用户角色
	roles, err := s.getUserRoles(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	user.Roles = roles

	return &user, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *models.UserCreateRequest) (*models.UserResponse, error) {
	// 检查用户名和邮箱是否已存在
	exists, err := s.checkUserExists(req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username or email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 插入用户
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO ua_admin (id, username, email, password_hash, avatar, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var avatar *string
	if req.Avatar != "" {
		avatar = &req.Avatar
	}

	_, err = s.db.Exec(query, id, req.Username, req.Email, string(hashedPassword), avatar, true, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.GetUserByID(id)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uuid.UUID, req *models.UserUpdateRequest) (*models.UserResponse, error) {
	// 检查用户是否存在
	_, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	// 构建更新查询
	setParts := []string{"updated_at = $1"}
	args := []interface{}{time.Now()}
	argIndex := 2

	if req.Username != "" {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, req.Username)
		argIndex++
	}

	if req.Email != "" {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, req.Email)
		argIndex++
	}

	if req.Avatar != "" {
		setParts = append(setParts, fmt.Sprintf("avatar = $%d", argIndex))
		args = append(args, req.Avatar)
		argIndex++
	}

	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(setParts) == 1 {
		return s.GetUserByID(id)
	}

	// 正确构建SQL查询
	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf("UPDATE ua_admin SET %s WHERE id = $%d", setClause, argIndex)
	args = append(args, id)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUserByID(id)
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(id uuid.UUID, isActive bool) error {
	query := "UPDATE ua_admin SET is_active = $1, updated_at = $2 WHERE id = $3"
	_, err := s.db.Exec(query, isActive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uuid.UUID) error {
	query := "DELETE FROM ua_admin WHERE id = $1"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(id uuid.UUID, req *models.ChangePasswordRequest) error {
	// 获取当前密码哈希
	var currentHash string
	query := "SELECT password_hash FROM ua_admin WHERE id = $1"
	err := s.db.QueryRow(query, id).Scan(&currentHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user password: %w", err)
	}

	// 验证旧密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPassword))
	if err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	// 加密新密码
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新密码
	updateQuery := "UPDATE ua_admin SET password_hash = $1, updated_at = $2 WHERE id = $3"
	_, err = s.db.Exec(updateQuery, string(newHash), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ResetPassword 管理员重置用户密码（无需验证旧密码）
func (s *UserService) ResetPassword(id uuid.UUID, req *models.ResetPasswordRequest) error {
	// 检查用户是否存在
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM ua_admin WHERE id = $1)"
	err := s.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// 加密新密码
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新密码
	updateQuery := "UPDATE ua_admin SET password_hash = $1, updated_at = $2 WHERE id = $3"
	_, err = s.db.Exec(updateQuery, string(newHash), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// checkUserExists 检查用户名或邮箱是否已存在
func (s *UserService) checkUserExists(username, email string) (bool, error) {
	query := "SELECT COUNT(*) FROM ua_admin WHERE username = $1 OR email = $2"
	var count int
	err := s.db.QueryRow(query, username, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}

// getUserRoles 获取用户角色（增强版本）
func (s *UserService) getUserRoles(userID uuid.UUID) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM az_roles r
		INNER JOIN az_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.id IS NOT NULL
		ORDER BY r.is_system DESC, r.name ASC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		// 验证角色数据完整性
		if role.ID == uuid.Nil || role.Name == "" {
			continue // 跳过无效的角色记录
		}

		roles = append(roles, role)
	}

	// 检查是否有扫描错误
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during role scanning: %w", err)
	}

	return roles, nil
}

// ValidateUserRoleConsistency 验证用户角色数据一致性
func (s *UserService) ValidateUserRoleConsistency(userID uuid.UUID) error {
	// 检查是否存在孤立的用户角色关联
	query := `
		SELECT COUNT(*) FROM az_user_roles ur
		LEFT JOIN ua_admin u ON ur.user_id = u.id
		LEFT JOIN az_roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND (u.id IS NULL OR r.id IS NULL)
	`

	var orphanCount int
	err := s.db.QueryRow(query, userID).Scan(&orphanCount)
	if err != nil {
		return fmt.Errorf("failed to check role consistency: %w", err)
	}

	if orphanCount > 0 {
		// 清理孤立的关联记录
		cleanupQuery := `
			DELETE FROM az_user_roles 
			WHERE user_id = $1 AND (
				NOT EXISTS (SELECT 1 FROM ua_admin WHERE id = user_id) OR
				NOT EXISTS (SELECT 1 FROM az_roles WHERE id = role_id)
			)
		`
		_, err = s.db.Exec(cleanupQuery, userID)
		if err != nil {
			return fmt.Errorf("failed to cleanup orphan role associations: %w", err)
		}
	}

	return nil
}

// GetUserRoles 对外暴露的获取用户角色方法（包装内部实现）
func (s *UserService) GetUserRoles(userID uuid.UUID) ([]models.Role, error) {
	return s.getUserRoles(userID)
}
