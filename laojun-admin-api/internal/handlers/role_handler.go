package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// GetRoles 获取角色列表
func (h *RoleHandler) GetRoles(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	result, err := h.roleService.GetRoles(page, size, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取角色列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "获取角色列表成功",
	})
}

// GetRole 获取角色详情
func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	role, err := h.roleService.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    role,
		"message": "获取角色成功",
	})
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req models.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	role, err := h.roleService.CreateRole(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建角色失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    role,
		"message": "创建角色成功",
	})
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	var req models.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	role, err := h.roleService.UpdateRole(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新角色失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    role,
		"message": "更新角色成功",
	})
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	if err := h.roleService.DeleteRole(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除角色失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除角色成功",
	})
}

// AssignRolesToUser 为用户分配角色（增强版本）
func (h *RoleHandler) AssignRolesToUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req models.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 确保UserID正确设置：如果请求体中没有UserID或与路径ID不一致，以路径ID为准
	if req.UserID == uuid.Nil || req.UserID != userID {
		req.UserID = userID
	}

	// 业务逻辑验证
	if err := h.validateRoleAssignment(userID, req.RoleIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "角色分配验证失败",
			"details": err.Error(),
		})
		return
	}

	// 获取操作者信息：如果请求头中没有操作者ID，返回未授权错误
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到操作者信息"})
		return
	}

	// 获取客户端信息用于审核日志
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 执行角色分配（带审计日志）
	if err := h.roleService.AssignRolesToUserWithAudit(
		req.UserID,
		req.RoleIDs,
		operatorID.(uuid.UUID),
		ipAddress,
		userAgent,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "分配用户角色失败",
			"details": err.Error(),
		})
		return
	}

	// 清理相关缓存（如果有缓存服务）
	// TODO: 清理用户权限缓存

	c.JSON(http.StatusOK, gin.H{
		"message":    "分配用户角色成功",
		"user_id":    userID,
		"role_count": len(req.RoleIDs),
	})
}

// validateRoleAssignment 验证角色分配的业务逻辑
func (h *RoleHandler) validateRoleAssignment(userID uuid.UUID, roleIDs []uuid.UUID) error {
	// 检查是否尝试分配重复角色ID
	roleMap := make(map[uuid.UUID]bool)
	for _, roleID := range roleIDs {
		if roleMap[roleID] {
			return fmt.Errorf("duplicate role ID: %s", roleID)
		}
		roleMap[roleID] = true
	}

	// 验证角色是否存在且可分配
	for _, roleID := range roleIDs {
		role, err := h.roleService.GetRoleByID(roleID)
		if err != nil {
			return fmt.Errorf("role %s not found: %w", roleID, err)
		}

		// 检查系统角色的分配限制（可根据业务需求调整）
		if role.IsSystem && role.Name == "admin" {
			// 可以添加额外的管理员角色分配限制
			// 例如：只有超级管理员才能分配管理员角色
			return fmt.Errorf("only superadmin can assign admin role")
		}
	}

	// 检查用户是否存在
	// 这个验证也会在 roleService.AssignRolesToUser 中进行，但在这里提前验证可以提供更好的错误信息
	_, err := h.roleService.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user %s not found: %w", userID, err)
	}

	return nil
}

// GetRolePermissions 获取角色的权限
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	perms, err := h.roleService.GetRolePermissions(roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取角色权限失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    perms,
		"message": "获取角色权限成功",
	})
}

// AssignPermissions 为角色分配权限
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	h.AssignPermissionsToRole(c)
}

// AssignPermissionsToRole 为角色分配权限
func (h *RoleHandler) AssignPermissionsToRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	var req models.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.roleService.AssignPermissionsToRole(roleID, req.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "分配角色权限失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "分配角色权限成功",
	})
}
