package middleware

import (
	"net/http"
	"strings"

	"github.com/codetaoist/laojun-marketplace-api/internal/models"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionMiddleware 权限中间件结构体
type PermissionMiddleware struct {
	permissionService *services.PermissionService
}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware(permissionService *services.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionService: permissionService,
	}
}

// RequireExtendedPermission 要求扩展权限的中间件
func (pm *PermissionMiddleware) RequireExtendedPermission(module, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 获取设备类型，默认为web
		deviceType := c.GetHeader("X-Device-Type")
		if deviceType == "" {
			deviceType = "web"
		}

		// 检查权限
		req := models.UserPermissionCheckRequest{
			UserID:     userID.(uuid.UUID),
			DeviceType: deviceType,
			Module:     module,
			Resource:   resource,
			Action:     action,
		}

		result, err := pm.permissionService.CheckUserPermission(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "权限检查失败",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !result.HasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "您没有执行此操作的权限",
				"required_permission": map[string]string{
					"module":      module,
					"resource":    resource,
					"action":      action,
					"device_type": deviceType,
				},
			})
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中，供后续使用
		c.Set("permission_check_result", result)
		c.Next()
	}
}

// RequireAnyExtendedPermission 要求任意一个扩展权限的中间件
func (pm *PermissionMiddleware) RequireAnyExtendedPermission(permissions []PermissionSpec) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 获取设备类型，默认为web
		deviceType := c.GetHeader("X-Device-Type")
		if deviceType == "" {
			deviceType = "web"
		}

		// 检查是否有任意一个权限
		hasAnyPermission := false
		var lastResult *models.UserPermissionCheckResponse

		for _, perm := range permissions {
			req := models.UserPermissionCheckRequest{
				UserID:     userID.(uuid.UUID),
				DeviceType: deviceType,
				Module:     perm.Module,
				Resource:   perm.Resource,
				Action:     perm.Action,
			}

			result, err := pm.permissionService.CheckUserPermission(req)
			if err != nil {
				continue // 忽略错误，继续检查下一个权限
			}

			lastResult = result
			if result.HasPermission {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":                "权限不足",
				"message":              "您没有执行此操作的权限",
				"required_permissions": permissions,
			})
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		c.Set("permission_check_result", lastResult)
		c.Next()
	}
}

// RequireAllExtendedPermissions 要求所有扩展权限的中间件
func (pm *PermissionMiddleware) RequireAllExtendedPermissions(permissions []PermissionSpec) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 获取设备类型，默认为web
		deviceType := c.GetHeader("X-Device-Type")
		if deviceType == "" {
			deviceType = "web"
		}

		// 检查是否拥有所有权限
		var missingPermissions []PermissionSpec

		for _, perm := range permissions {
			req := models.UserPermissionCheckRequest{
				UserID:     userID.(uuid.UUID),
				DeviceType: deviceType,
				Module:     perm.Module,
				Resource:   perm.Resource,
				Action:     perm.Action,
			}

			result, err := pm.permissionService.CheckUserPermission(req)
			if err != nil || !result.HasPermission {
				missingPermissions = append(missingPermissions, perm)
			}
		}

		if len(missingPermissions) > 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "权限不足",
				"message":             "您缺少以下权限",
				"missing_permissions": missingPermissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireDeviceTypePermission 要求特定设备类型权限的中间件
func (pm *PermissionMiddleware) RequireDeviceTypePermission(allowedDeviceTypes []string, module, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 获取设备类型
		deviceType := c.GetHeader("X-Device-Type")
		if deviceType == "" {
			deviceType = "web"
		}

		// 检查设备类型是否被允许
		deviceTypeAllowed := false
		for _, allowedType := range allowedDeviceTypes {
			if deviceType == allowedType {
				deviceTypeAllowed = true
				break
			}
		}

		if !deviceTypeAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":                "设备类型不被允许",
				"message":              "当前设备类型无法访问此功能",
				"current_device_type":  deviceType,
				"allowed_device_types": allowedDeviceTypes,
			})
			c.Abort()
			return
		}

		// 检查权限
		req := models.UserPermissionCheckRequest{
			UserID:     userID.(uuid.UUID),
			DeviceType: deviceType,
			Module:     module,
			Resource:   resource,
			Action:     action,
		}

		result, err := pm.permissionService.CheckUserPermission(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "权限检查失败",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !result.HasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "您没有在当前设备上执行此操作的权限",
				"required_permission": map[string]string{
					"module":      module,
					"resource":    resource,
					"action":      action,
					"device_type": deviceType,
				},
			})
			c.Abort()
			return
		}

		c.Set("permission_check_result", result)
		c.Next()
	}
}

// DynamicPermissionCheck 动态权限检查中间件
func (pm *PermissionMiddleware) DynamicPermissionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求路径和方法推断权限要求
		path := c.Request.URL.Path
		method := c.Request.Method

		module, resource, action := pm.inferPermissionFromRequest(path, method)
		if module == "" || resource == "" || action == "" {
			// 如果无法推断权限，则跳过检查
			c.Next()
			return
		}

		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 获取设备类型
		deviceType := c.GetHeader("X-Device-Type")
		if deviceType == "" {
			deviceType = "web"
		}

		// 检查权限
		req := models.UserPermissionCheckRequest{
			UserID:     userID.(uuid.UUID),
			DeviceType: deviceType,
			Module:     module,
			Resource:   resource,
			Action:     action,
		}

		result, err := pm.permissionService.CheckUserPermission(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "权限检查失败",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !result.HasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "您没有执行此操作的权限",
				"inferred_permission": map[string]string{
					"module":      module,
					"resource":    resource,
					"action":      action,
					"device_type": deviceType,
				},
			})
			c.Abort()
			return
		}

		c.Set("permission_check_result", result)
		c.Next()
	}
}

// inferPermissionFromRequest 从请求推断权限要求
func (pm *PermissionMiddleware) inferPermissionFromRequest(path, method string) (module, resource, action string) {
	// 移除API前缀
	path = strings.TrimPrefix(path, "/api/v1")

	// 分割路径
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return "", "", ""
	}

	// 根据路径推断模块和资源
	switch parts[0] {
	case "users":
		module = "user"
		resource = "user"
	case "roles":
		module = "permission"
		resource = "role"
	case "permissions":
		module = "permission"
		resource = "permission"
	case "user-groups":
		module = "permission"
		resource = "user_group"
	case "device-types":
		module = "system"
		resource = "device_type"
	case "modules":
		module = "system"
		resource = "module"
	default:
		return "", "", ""
	}

	// 根据HTTP方法推断动作
	switch method {
	case "GET":
		if len(parts) > 1 {
			action = "view"
		} else {
			action = "list"
		}
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "edit"
	case "DELETE":
		action = "delete"
	default:
		action = "unknown"
	}

	return module, resource, action
}

// PermissionSpec 权限规格
type PermissionSpec struct {
	Module   string `json:"module"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// LogPermissionCheck 记录权限检查的中间件
func (pm *PermissionMiddleware) LogPermissionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 获取权限检查结果
		if result, exists := c.Get("permission_check_result"); exists {
			if permResult, ok := result.(*models.UserPermissionCheckResponse); ok {
				// 这里可以记录权限检查日志
				// 例如：记录到数据库、发送到日志系统等
				c.Header("X-Permission-Check", permResult.Reason)
			}
		}
	}
}
