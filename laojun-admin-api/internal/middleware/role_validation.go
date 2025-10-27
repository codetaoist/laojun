package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RoleValidationMiddleware 角色验证中间件
type RoleValidationMiddleware struct {
	roleService *services.RoleService
	userService *services.UserService
}

// NewRoleValidationMiddleware 创建角色验证中间件
func NewRoleValidationMiddleware(roleService *services.RoleService, userService *services.UserService) *RoleValidationMiddleware {
	return &RoleValidationMiddleware{
		roleService: roleService,
		userService: userService,
	}
}

// ValidateRoleAssignment 验证角色分配的中间件
func (m *RoleValidationMiddleware) ValidateRoleAssignment() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对角色分配相关的请求进行验证
		if !m.isRoleAssignmentRequest(c) {
			c.Next()
			return
		}

		// 获取当前操作用户信息
		currentUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到当前用户信息"})
			c.Abort()
			return
		}

		// 解析请求参数
		targetUserID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			c.Abort()
			return
		}

		var req models.AssignRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
			c.Abort()
			return
		}

		// 执行业务逻辑验证
		if err := m.validateBusinessRules(currentUserID.(uuid.UUID), targetUserID, req.RoleIDs); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "角色分配验证失败",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// 将验证后的请求数据存储到上下文中
		c.Set("validated_role_request", req)
		c.Next()
	}
}

// isRoleAssignmentRequest 检查是否为角色分配请求
func (m *RoleValidationMiddleware) isRoleAssignmentRequest(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	// 匹配 POST /users/:id/roles 路径
	return method == "POST" && strings.Contains(path, "/users/") && strings.HasSuffix(path, "/roles")
}

// validateBusinessRules 验证业务规则
func (m *RoleValidationMiddleware) validateBusinessRules(currentUserID, targetUserID uuid.UUID, roleIDs []uuid.UUID) error {
	// 1. 获取当前用户的角色信息
	currentUserRoles, err := m.userService.GetUserRoles(currentUserID)
	if err != nil {
		return fmt.Errorf("获取当前用户角色失败: %w", err)
	}

	// 2. 获取目标用户的当前角色信息
	targetUserRoles, err := m.userService.GetUserRoles(targetUserID)
	if err != nil {
		return fmt.Errorf("获取目标用户角色失败: %w", err)
	}

	// 3. 检查是否尝试给自己分配角色
	if currentUserID == targetUserID {
		return fmt.Errorf("不能为自己分配角色")
	}

	// 4. 验证要分配的角色是否存在
	for _, roleID := range roleIDs {
		role, err := m.roleService.GetRoleByID(roleID)
		if err != nil {
			return fmt.Errorf("角色 %s 不存在", roleID)
		}

		// 5. 检查系统角色的分配限制
		if role.IsSystem {
			if err := m.validateSystemRoleAssignment(currentUserRoles, *role); err != nil {
				return err
			}
		}
	}

	// 6. 检查是否尝试移除目标用户的关键角色
	if err := m.validateCriticalRoleRemoval(targetUserRoles, roleIDs); err != nil {
		return err
	}

	return nil
}

// validateSystemRoleAssignment 验证系统角色分配
func (m *RoleValidationMiddleware) validateSystemRoleAssignment(currentUserRoles []models.Role, targetRole models.Role) error {
	// 检查当前用户是否有权限分配系统角色
	hasAdminRole := false
	for _, role := range currentUserRoles {
		if role.IsSystem && (role.Name == "admin" || role.Name == "super_admin") {
			hasAdminRole = true
			break
		}
	}

	if !hasAdminRole {
		return fmt.Errorf("只有管理员才能分配系统角色 %s", targetRole.Name)
	}

	// 特殊检查：超级管理员角色只能由超级管理员分配
	if targetRole.Name == "super_admin" {
		hasSuperAdminRole := false
		for _, role := range currentUserRoles {
			if role.Name == "super_admin" {
				hasSuperAdminRole = true
				break
			}
		}
		if !hasSuperAdminRole {
			return fmt.Errorf("只有超级管理员才能分配超级管理员角色")
		}
	}

	return nil
}

// validateCriticalRoleRemoval 验证关键角色移除
func (m *RoleValidationMiddleware) validateCriticalRoleRemoval(currentRoles []models.Role, newRoleIDs []uuid.UUID) error {
	// 创建新角色ID的映射
	newRoleMap := make(map[uuid.UUID]bool)
	for _, roleID := range newRoleIDs {
		newRoleMap[roleID] = true
	}

	// 检查是否移除了关键的系统角色
	for _, role := range currentRoles {
		if role.IsSystem && !newRoleMap[role.ID] {
			// 如果这是最后一个具有该系统角色的用户，则不允许移除
			if role.Name == "admin" || role.Name == "super_admin" {
				// 这里可以添加检查是否还有其他用户具有该角色的逻辑
				// 为了简化，暂时允许移除，但可以根据业务需求调整
				return fmt.Errorf("不能移除最后一个 %s 角色", role.Name)
			}
		}
	}

	return nil
}

// RequireRoleManagementPermission 要求角色管理权限的中间件
func (m *RoleValidationMiddleware) RequireRoleManagementPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		// 检查用户是否有角色管理权限
		userRoles, err := m.userService.GetUserRoles(userID.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户权限失败"})
			c.Abort()
			return
		}

		hasPermission := false
		for _, role := range userRoles {
			// 检查是否有管理员角色或特定的角色管理权
			if role.IsSystem && (role.Name == "admin" || role.Name == "super_admin") {
				hasPermission = true
				break
			}
			// 这里可以添加更细粒度的权限检查，例如检查是否有特定模块的角色管理权
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，无法管理用户角色"})
			c.Abort()
			return
		}

		c.Next()
	}
}
