package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-gateway/internal/routes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteHandler 路由管理处理器
type RouteHandler struct {
	routeManager *routes.DynamicRouteManager
	logger       *zap.Logger
}

// NewRouteHandler 创建路由处理器
func NewRouteHandler(routeManager *routes.DynamicRouteManager, logger *zap.Logger) *RouteHandler {
	return &RouteHandler{
		routeManager: routeManager,
		logger:       logger,
	}
}

// CreateRoute 创建路由
func (rh *RouteHandler) CreateRoute(c *gin.Context) {
	var route routes.RouteInfo
	if err := c.ShouldBindJSON(&route); err != nil {
		rh.logger.Error("Failed to bind route data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid route data",
			"details": err.Error(),
		})
		return
	}

	if err := rh.routeManager.AddRoute(&route); err != nil {
		rh.logger.Error("Failed to create route", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create route",
			"details": err.Error(),
		})
		return
	}

	rh.logger.Info("Route created successfully", zap.String("id", route.ID))
	c.JSON(http.StatusCreated, gin.H{
		"message": "Route created successfully",
		"route":   route,
	})
}

// GetRoute 获取路由信息
func (rh *RouteHandler) GetRoute(c *gin.Context) {
	routeID := c.Param("id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Route ID is required",
		})
		return
	}

	route, err := rh.routeManager.GetRoute(routeID)
	if err != nil {
		rh.logger.Error("Failed to get route", zap.String("id", routeID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"route": route,
	})
}

// UpdateRoute 更新路由
func (rh *RouteHandler) UpdateRoute(c *gin.Context) {
	routeID := c.Param("id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Route ID is required",
		})
		return
	}

	var updatedRoute routes.RouteInfo
	if err := c.ShouldBindJSON(&updatedRoute); err != nil {
		rh.logger.Error("Failed to bind route data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid route data",
			"details": err.Error(),
		})
		return
	}

	if err := rh.routeManager.UpdateRoute(routeID, &updatedRoute); err != nil {
		rh.logger.Error("Failed to update route", zap.String("id", routeID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update route",
			"details": err.Error(),
		})
		return
	}

	rh.logger.Info("Route updated successfully", zap.String("id", routeID))
	c.JSON(http.StatusOK, gin.H{
		"message": "Route updated successfully",
		"route":   updatedRoute,
	})
}

// DeleteRoute 删除路由
func (rh *RouteHandler) DeleteRoute(c *gin.Context) {
	routeID := c.Param("id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Route ID is required",
		})
		return
	}

	if err := rh.routeManager.RemoveRoute(routeID); err != nil {
		rh.logger.Error("Failed to delete route", zap.String("id", routeID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete route",
			"details": err.Error(),
		})
		return
	}

	rh.logger.Info("Route deleted successfully", zap.String("id", routeID))
	c.JSON(http.StatusOK, gin.H{
		"message": "Route deleted successfully",
	})
}

// ListRoutes 列出所有路由
func (rh *RouteHandler) ListRoutes(c *gin.Context) {
	// 获取查询参数
	service := c.Query("service")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	routes := rh.routeManager.ListRoutes()

	// 过滤路由
	var filteredRoutes []*routes.RouteInfo
	for _, route := range routes {
		// 按服务过滤
		if service != "" && route.Service != service {
			continue
		}
		// 按状态过滤
		if status != "" && route.Status != status {
			continue
		}
		filteredRoutes = append(filteredRoutes, route)
	}

	// 分页
	total := len(filteredRoutes)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedRoutes := filteredRoutes[start:end]

	c.JSON(http.StatusOK, gin.H{
		"routes": paginatedRoutes,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetRoutesByService 根据服务获取路由
func (rh *RouteHandler) GetRoutesByService(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	routes := rh.routeManager.GetRoutesByService(serviceName)

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"routes":  routes,
		"count":   len(routes),
	})
}

// ToggleRoute 启用/禁用路由
func (rh *RouteHandler) ToggleRoute(c *gin.Context) {
	routeID := c.Param("id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Route ID is required",
		})
		return
	}

	var request struct {
		Active bool `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		rh.logger.Error("Failed to bind toggle data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid toggle data",
			"details": err.Error(),
		})
		return
	}

	if err := rh.routeManager.ToggleRoute(routeID, request.Active); err != nil {
		rh.logger.Error("Failed to toggle route", zap.String("id", routeID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to toggle route",
			"details": err.Error(),
		})
		return
	}

	status := "inactive"
	if request.Active {
		status = "active"
	}

	rh.logger.Info("Route toggled successfully", zap.String("id", routeID), zap.String("status", status))
	c.JSON(http.StatusOK, gin.H{
		"message": "Route status updated successfully",
		"status":  status,
	})
}

// GetRouteStats 获取路由统计信息
func (rh *RouteHandler) GetRouteStats(c *gin.Context) {
	routes := rh.routeManager.ListRoutes()

	stats := map[string]interface{}{
		"total_routes": len(routes),
		"active_routes": 0,
		"inactive_routes": 0,
		"deprecated_routes": 0,
		"services": make(map[string]int),
		"methods": make(map[string]int),
	}

	services := make(map[string]int)
	methods := make(map[string]int)

	for _, route := range routes {
		// 统计状态
		switch route.Status {
		case "active":
			stats["active_routes"] = stats["active_routes"].(int) + 1
		case "inactive":
			stats["inactive_routes"] = stats["inactive_routes"].(int) + 1
		case "deprecated":
			stats["deprecated_routes"] = stats["deprecated_routes"].(int) + 1
		}

		// 统计服务
		services[route.Service]++

		// 统计方法
		methods[route.Method]++
	}

	stats["services"] = services
	stats["methods"] = methods

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// ValidateRoute 验证路由配置
func (rh *RouteHandler) ValidateRoute(c *gin.Context) {
	var route routes.RouteInfo
	if err := c.ShouldBindJSON(&route); err != nil {
		rh.logger.Error("Failed to bind route data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid route data",
			"details": err.Error(),
		})
		return
	}

	// 这里可以添加更详细的验证逻辑
	validationErrors := []string{}

	if route.Path == "" {
		validationErrors = append(validationErrors, "Path is required")
	}
	if route.Method == "" {
		validationErrors = append(validationErrors, "Method is required")
	}
	if route.Service == "" && route.Target == "" {
		validationErrors = append(validationErrors, "Either service or target is required")
	}

	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":  false,
			"errors": validationErrors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Route configuration is valid",
	})
}

// ExportRoutes 导出路由配置
func (rh *RouteHandler) ExportRoutes(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	routes := rh.routeManager.ListRoutes()

	switch format {
	case "json":
		c.Header("Content-Disposition", "attachment; filename=routes.json")
		c.JSON(http.StatusOK, routes)
	case "yaml":
		// 这里可以添加YAML导出逻辑
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "YAML export not implemented yet",
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported format. Supported formats: json, yaml",
		})
	}
}

// ImportRoutes 导入路由配置
func (rh *RouteHandler) ImportRoutes(c *gin.Context) {
	var importRequest struct {
		Routes []routes.RouteInfo `json:"routes"`
		Mode   string             `json:"mode"` // replace, merge
	}

	if err := c.ShouldBindJSON(&importRequest); err != nil {
		rh.logger.Error("Failed to bind import data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid import data",
			"details": err.Error(),
		})
		return
	}

	successCount := 0
	errorCount := 0
	errors := []string{}

	for _, route := range importRequest.Routes {
		if err := rh.routeManager.AddRoute(&route); err != nil {
			errorCount++
			errors = append(errors, err.Error())
		} else {
			successCount++
		}
	}

	rh.logger.Info("Routes import completed",
		zap.Int("success", successCount),
		zap.Int("errors", errorCount))

	c.JSON(http.StatusOK, gin.H{
		"message":      "Import completed",
		"success_count": successCount,
		"error_count":   errorCount,
		"errors":       errors,
	})
}