package services

import (
	"fmt"
	"time"

	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(username, email, password string) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	err := s.db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error
	if err == nil {
		if existingUser.Username == username {
			return nil, fmt.Errorf("username '%s' already exists", username)
		}
		if existingUser.Email == email {
			return nil, fmt.Errorf("email '%s' already exists", email)
		}
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.db.Create(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 为新用户分配默认角色
	err = s.assignDefaultRole(user.ID)
	if err != nil {
		// 记录错误但不影响用户创建
		fmt.Printf("Warning: failed to assign default role to user %s: %v\n", user.ID, err)
	}

	return user, nil
}

// GetUser 获取用户
func (s *UserService) GetUser(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(userID uuid.UUID, username, email *string) (*models.User, error) {
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if username != nil {
		// 检查新用户名是否已存在
		var existingUser models.User
		err = s.db.Where("username = ? AND id != ?", *username, userID).First(&existingUser).Error
		if err == nil {
			return nil, fmt.Errorf("username '%s' already exists", *username)
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing username: %w", err)
		}
		updates["username"] = *username
	}

	if email != nil {
		// 检查新邮箱是否已存在
		var existingUser models.User
		err = s.db.Where("email = ? AND id != ?", *email, userID).First(&existingUser).Error
		if err == nil {
			return nil, fmt.Errorf("email '%s' already exists", *email)
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		updates["email"] = *email
	}

	err = s.db.Model(&user).Updates(updates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新密码
	err = s.db.Model(&user).Updates(map[string]interface{}{
		"password_hash": string(hashedPassword),
		"updated_at":    time.Now(),
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// VerifyPassword 验证用户密码
func (s *UserService) VerifyPassword(userID uuid.UUID, password string) error {
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

// AuthenticateUser 用户认证
func (s *UserService) AuthenticateUser(username, password string) (*models.User, error) {
	var user models.User
	err := s.db.Where("username = ? OR email = ?", username, username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	return &user, nil
}

// DeactivateUser 停用用户
func (s *UserService) DeactivateUser(userID uuid.UUID) error {
	err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// ActivateUser 激活用户
func (s *UserService) ActivateUser(userID uuid.UUID) error {
	err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_active":  true,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	// 删除用户角色关联
	err := s.db.Table("az_user_roles").Where("user_id = ?", userID).Delete(nil).Error
	if err != nil {
		return fmt.Errorf("failed to delete user roles: %w", err)
	}

	// 删除用户
	err = s.db.Where("id = ?", userID).Delete(&models.User{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetUsers 分页获取用户列表
func (s *UserService) GetUsers(page, limit int, isActive *bool) ([]models.User, *models.PaginationMeta, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取用户列表
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get users: %w", err)
	}

	// 计算分页信息
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	
	meta := &models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return users, meta, nil
}

// assignDefaultRole 为用户分配默认角色
func (s *UserService) assignDefaultRole(userID uuid.UUID) error {
	// 查找默认用户角色
	var role models.Role
	err := s.db.Where("name = ?", "user").First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有找到默认角色，创建一个
			role = models.Role{
				ID:          uuid.New(),
				Name:        "user",
				Description: stringPtr("Default user role"),
				IsSystem:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			err = s.db.Create(&role).Error
			if err != nil {
				return fmt.Errorf("failed to create default role: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get default role: %w", err)
		}
	}

	// 分配角色给用户
	userRole := map[string]interface{}{
		"id":      uuid.New(),
		"user_id": userID,
		"role_id": role.ID,
	}

	err = s.db.Table("az_user_roles").Create(userRole).Error
	if err != nil {
		return fmt.Errorf("failed to assign default role: %w", err)
	}

	return nil
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}