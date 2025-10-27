package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PermissionHandler struct {
	permissionService *services.PermissionService
}

func NewPermissionHandler(permissionService *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// DeviceType 相关处理

// GetDeviceTypes 获取设备类型列表
func (h *PermissionHandler) GetDeviceTypes(c *gin.Context) {
	deviceTypes, err := h.permissionService.GetDeviceTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取设备类型列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    deviceTypes,
		"message": "获取设备类型列表成功",
	})
}

// CreateDeviceType 创建设备类型
func (h *PermissionHandler) CreateDeviceType(c *gin.Context) {
	var req models.DeviceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	deviceType, err := h.permissionService.CreateDeviceType(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建设备类型失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    deviceType,
		"message": "创建设备类型成功",
	})
}

// Module 相关处理

// GetModules 获取模块列表
func (h *PermissionHandler) GetModules(c *gin.Context) {
	modules, err := h.permissionService.GetModules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取模块列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    modules,
		"message": "获取模块列表成功",
	})
}

// CreateModule 创建模块
func (h *PermissionHandler) CreateModule(c *gin.Context) {
	var req models.ModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	module, err := h.permissionService.CreateModule(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建模块失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    module,
		"message": "创建模块成功",
	})
}

// UserGroup 相关处理

// GetUserGroups 获取用户组列表
func (h *PermissionHandler) GetUserGroups(c *gin.Context) {
	userGroups, err := h.permissionService.GetUserGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户组列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    userGroups,
		"message": "获取用户组列表成功",
	})
}

// CreateUserGroup 创建用户组
func (h *PermissionHandler) CreateUserGroup(c *gin.Context) {
	var req models.UserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	userGroup, err := h.permissionService.CreateUserGroup(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建用户组失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    userGroup,
		"message": "创建用户组成功",
	})
}

// AddUsersToGroup 添加用户到用户组
func (h *PermissionHandler) AddUsersToGroup(c *gin.Context) {
	var req models.UserGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	err := h.permissionService.AddUsersToGroup(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "添加用户到用户组失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "添加用户到用户组成功",
	})
}

// ExtendedPermission 相关处理

// GetExtendedPermissions 获取扩展权限列表
func (h *PermissionHandler) GetExtendedPermissions(c *gin.Context) {
	moduleIDStr := c.Query("module_id")
	deviceTypeIDStr := c.Query("device_type_id")

	var moduleID, deviceTypeID uuid.UUID
	var err error

	if moduleIDStr != "" {
		moduleID, err = uuid.Parse(moduleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的模块ID",
			})
			return
		}
	}

	if deviceTypeIDStr != "" {
		deviceTypeID, err = uuid.Parse(deviceTypeIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的设备类型ID",
			})
			return
		}
	}

	permissions, err := h.permissionService.GetExtendedPermissions(moduleID, deviceTypeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取扩展权限列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    permissions,
		"message": "获取扩展权限列表成功",
	})
}

// CreateExtendedPermission 创建扩展权限
func (h *PermissionHandler) CreateExtendedPermission(c *gin.Context) {
	var req models.ExtendedPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	permission, err := h.permissionService.CreateExtendedPermission(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建扩展权限失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    permission,
		"message": "创建扩展权限成功",
	})
}

// 权限检查相关处理器

// CheckUserPermission 检查用户权限
func (h *PermissionHandler) CheckUserPermission(c *gin.Context) {
	var req models.UserPermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	result, err := h.permissionService.CheckUserPermission(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "权限检查失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "权限检查成功",
	})
}

// CheckCurrentUserPermission 检查当前用户权限
func (h *PermissionHandler) CheckCurrentUserPermission(c *gin.Context) {
	// 从上下文获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	deviceType := c.Query("device_type")
	module := c.Query("module")
	resource := c.Query("resource")
	action := c.Query("action")

	if deviceType == "" || module == "" || resource == "" || action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少必要参数: device_type, module, resource, action",
		})
		return
	}

	req := models.UserPermissionCheckRequest{
		UserID:     userID.(uuid.UUID),
		DeviceType: deviceType,
		Module:     module,
		Resource:   resource,
		Action:     action,
	}

	result, err := h.permissionService.CheckUserPermission(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "权限检查失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "权限检查成功",
	})
}

// 权限同步相关处理

// SyncUserPermissions 同步用户权限
func (h *PermissionHandler) SyncUserPermissions(c *gin.Context) {
	var req models.PermissionSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	result, err := h.permissionService.SyncUserPermissions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "权限同步失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "权限同步成功",
	})
}

// SyncCurrentUserPermissions 同步当前用户权限
func (h *PermissionHandler) SyncCurrentUserPermissions(c *gin.Context) {
	// 从上下文获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	deviceType := c.Query("device_type")
	if deviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少必要参数: device_type",
		})
		return
	}

	sessionID := c.Query("session_id")

	req := models.PermissionSyncRequest{
		UserID:     userID.(uuid.UUID),
		DeviceType: deviceType,
		SessionID:  sessionID,
	}

	result, err := h.permissionService.SyncUserPermissions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "权限同步失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "权限同步成功",
	})
}

// PermissionTemplate 相关处理

// GetPermissionTemplates 获取权限模板列表
func (h *PermissionHandler) GetPermissionTemplates(c *gin.Context) {
	templates, err := h.permissionService.GetPermissionTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取权限模板列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    templates,
		"message": "获取权限模板列表成功",
	})
}

// CreatePermissionTemplate 创建权限模板
func (h *PermissionHandler) CreatePermissionTemplate(c *gin.Context) {
	var req models.PermissionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	template, err := h.permissionService.CreatePermissionTemplate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建权限模板失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    template,
		"message": "创建权限模板成功",
	})
}

// ApplyPermissionTemplate 应用权限模板
func (h *PermissionHandler) ApplyPermissionTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的模板ID",
		})
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	err = h.permissionService.ApplyPermissionTemplate(templateID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "应用权限模板失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "权限模板应用成功",
	})
}

// 缓存管理相关处理

// InvalidateUserCache 清除用户缓存
func (h *PermissionHandler) InvalidateUserCache(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的用户ID",
			"details": err.Error(),
		})
		return
	}

	if err := h.permissionService.InvalidateUserCache(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "清除用户缓存失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户缓存清除成功",
	})
}

// InvalidateCurrentUserCache 清除当前用户缓存
func (h *PermissionHandler) InvalidateCurrentUserCache(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	if err := h.permissionService.InvalidateUserCache(userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "清除当前用户缓存失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "当前用户缓存清除成功",
	})
}

// GetCacheStats 获取缓存统计信息
func (h *PermissionHandler) GetCacheStats(c *gin.Context) {
	stats, err := h.permissionService.GetCacheStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取缓存统计信息失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    stats,
		"message": "获取缓存统计信息成功",
	})
}

// WarmupUserCache 预热用户缓存
func (h *PermissionHandler) WarmupUserCache(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的用户ID",
			"details": err.Error(),
		})
		return
	}

	if err := h.permissionService.WarmupUserCache(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "预热用户缓存失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户缓存预热成功",
	})
}

// GetPermissions 返回基础权限列表供前端抽屉使用（支持查询参数：search, resource, action, isSystem, page, pageSize）
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	// 解析查询参数
	search := c.Query("search")
	resource := c.Query("resource")
	action := c.Query("action")
	isSystemStr := c.Query("isSystem")
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")

	var isSystem *bool
	if isSystemStr != "" {
		v := strings.ToLower(isSystemStr)
		b := v == "true" || v == "1"
		isSystem = &b
	}

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	pageSize := 10
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	params := models.PermissionQueryParams{
		Search:   search,
		Resource: resource,
		Action:   action,
		IsSystem: isSystem,
		Page:     page,
		Limit:    pageSize,
	}

	perms, total, err := h.permissionService.GetPermissionsPaginated(params)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 计算总页数
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "获取权限列表成功",
		"data": gin.H{
			"data":       perms,
			"items":      perms,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		},
	})
}

// GetPermission 根据ID获取权限
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	h.GetPermissionByID(c)
}

// 获取权限详情
func (h *PermissionHandler) GetPermissionByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.permissionService.GetBasicPermissionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p, "message": "获取权限详情成功"})
}

// 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req models.BasicPermissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}
	p, err := h.permissionService.CreateBasicPermission(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建权限失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p, "message": "创建权限成功"})
}

// 更新权限
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	var req models.BasicPermissionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}
	p, err := h.permissionService.UpdateBasicPermission(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新权限失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p, "message": "更新权限成功"})
}

// 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if err := h.permissionService.DeleteBasicPermission(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除权限失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除权限成功"})
}

// 批量删除权限
func (h *PermissionHandler) BatchDeletePermissions(c *gin.Context) {
	var req models.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}
	if err := h.permissionService.BatchDeletePermissions(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除权限失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "批量删除权限成功"})
}

// 权限统计
func (h *PermissionHandler) GetPermissionStats(c *gin.Context) {
	stats, err := h.permissionService.GetPermissionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取权限统计失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats, "message": "获取权限统计成功"})
}

// 资源列表
func (h *PermissionHandler) GetResources(c *gin.Context) {
	list, err := h.permissionService.GetDistinctResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取资源列表失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "message": "获取资源列表成功"})
}

// 动作列表
func (h *PermissionHandler) GetActions(c *gin.Context) {
	list, err := h.permissionService.GetDistinctActions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取动作列表失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "message": "获取动作列表成功"})
}

// 权限使用情况
func (h *PermissionHandler) CheckPermissionUsage(c *gin.Context) {
	id := c.Param("id")
	res, err := h.permissionService.CheckPermissionUsage(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限使用失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res, "message": "检查权限使用成功"})
}

// 导出权限配置（返回二进制流）
func (h *PermissionHandler) ExportPermissions(c *gin.Context) {
	format := c.Query("format")
	payload, contentType, filename, err := h.permissionService.ExportPermissions(format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出权限失败", "details": err.Error()})
		return
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, payload)
}

// 导入权限配置（multipart/form-data 格式）
func (h *PermissionHandler) ImportPermissions(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件", "details": err.Error()})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败", "details": err.Error()})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败", "details": err.Error()})
		return
	}
	result, err := h.permissionService.ImportPermissions(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入权限失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result, "message": "导入权限完成"})
}

// 同步系统权限
func (h *PermissionHandler) SyncSystemPermissions(c *gin.Context) {
	res, err := h.permissionService.SyncSystemPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "同步系统权限失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res, "message": "同步系统权限完成"})
}
