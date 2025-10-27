package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/registry"
	"github.com/codetaoist/laojun-discovery/internal/storage"
	sharedRegistry "github.com/codetaoist/laojun-shared/registry"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DiscoveryRegistryHandler 发现服务注册处理器
type DiscoveryRegistryHandler struct {
	registry *registry.DiscoveryServiceRegistry
	logger   *zap.Logger
}

// NewDiscoveryRegistryHandler 创建发现服务注册处理器
func NewDiscoveryRegistryHandler(registry *registry.DiscoveryServiceRegistry, logger *zap.Logger) *DiscoveryRegistryHandler {
	return &DiscoveryRegistryHandler{
		registry: registry,
		logger:   logger,
	}
}

// RegisterService 注册服务
func (h *DiscoveryRegistryHandler) RegisterService(c *gin.Context) {
	var req sharedRegistry.ServiceInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 验证请求参数
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	if req.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service address is required",
		})
		return
	}

	if req.Port <= 0 || req.Port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid port number",
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.registry.RegisterService(ctx, &req); err != nil {
		h.logger.Error("Failed to register service",
			zap.String("service", req.Name),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Service registered successfully",
		"service": req,
	})
}

// DeregisterService 注销服务
func (h *DiscoveryRegistryHandler) DeregisterService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.registry.DeregisterService(ctx, serviceID); err != nil {
		h.logger.Error("Failed to deregister service",
			zap.String("service_id", serviceID),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to deregister service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service deregistered successfully",
	})
}

// UpdateHealth 更新健康状态
func (h *DiscoveryRegistryHandler) UpdateHealth(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	var req sharedRegistry.HealthCheck
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.registry.UpdateServiceHealth(ctx, serviceID, &req); err != nil {
		h.logger.Error("Failed to update health",
			zap.String("service_id", serviceID),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update health",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Health status updated successfully",
	})
}

// ListServices 列出服务
func (h *DiscoveryRegistryHandler) ListServices(c *gin.Context) {
	serviceName := c.Query("name")
	
	ctx := c.Request.Context()
	
	if serviceName != "" {
		// 列出指定服务的实例
		instances, err := h.registry.DiscoverServices(ctx, serviceName)
		if err != nil {
			h.logger.Error("Failed to list services",
				zap.String("service", serviceName),
				zap.Error(err))
			
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list services",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"instances": instances,
			"count": len(instances),
		})
	} else {
		// 获取注册中心健康状态（包含所有服务信息）
		health, err := h.registry.GetRegistryHealth(ctx)
		if err != nil {
			h.logger.Error("Failed to get registry health", zap.Error(err))
			
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list services",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"services": health.Services,
			"service_count": len(health.Services),
			"total_instances": health.TotalServices,
		})
	}
}

// GetService 获取服务详情
func (h *DiscoveryRegistryHandler) GetService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}
	
	ctx := c.Request.Context()
	
	service, err := h.registry.GetServiceByID(ctx, serviceID)
	if err != nil {
		h.logger.Error("Failed to get service",
			zap.String("service_id", serviceID),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": service,
	})
}

// RegistryHandler 服务注册处理器 (已弃用，建议使用 DiscoveryRegistryHandler)
type RegistryHandler struct {
	registry *registry.ServiceRegistry
	logger   *zap.Logger
}

// NewRegistryHandler 创建服务注册处理器 (已弃用，建议使用 NewDiscoveryRegistryHandler)
func NewRegistryHandler(registry *registry.ServiceRegistry, logger *zap.Logger) *RegistryHandler {
	return &RegistryHandler{
		registry: registry,
		logger:   logger,
	}
}

// RegisterService 注册服务
func (h *RegistryHandler) RegisterService(c *gin.Context) {
	var req registry.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 验证请求参数
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	if req.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service address is required",
		})
		return
	}

	if req.Port <= 0 || req.Port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid port number",
		})
		return
	}

	ctx := c.Request.Context()
	instance, err := h.registry.Register(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to register service",
			zap.String("service", req.Name),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Service registered successfully",
		"service": instance,
	})
}

// DeregisterService 注销服务
func (h *RegistryHandler) DeregisterService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.registry.Deregister(ctx, serviceID); err != nil {
		h.logger.Error("Failed to deregister service",
			zap.String("service_id", serviceID),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to deregister service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service deregistered successfully",
	})
}

// UpdateHealth 更新健康状态
func (h *RegistryHandler) UpdateHealth(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Output string `json:"output"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 验证状态值
	if req.Status != "passing" && req.Status != "warning" && req.Status != "critical" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status value. Must be one of: passing, warning, critical",
		})
		return
	}

	ctx := c.Request.Context()
	health := storage.HealthStatus{
		Status:      req.Status,
		Output:      req.Output,
		LastChecked: time.Now(),
	}
	if err := h.registry.UpdateHealth(ctx, serviceID, health); err != nil {
		h.logger.Error("Failed to update health",
			zap.String("service_id", serviceID),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update health",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Health status updated successfully",
	})
}

// ListServices 列出服务
func (h *RegistryHandler) ListServices(c *gin.Context) {
	serviceName := c.Query("name")
	
	ctx := c.Request.Context()
	
	if serviceName != "" {
		// 列出指定服务的实例
		instances, err := h.registry.ListServices(ctx, serviceName)
		if err != nil {
			h.logger.Error("Failed to list services",
				zap.String("service", serviceName),
				zap.Error(err))
			
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list services",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"instances": instances,
			"count": len(instances),
		})
	} else {
		// 列出所有服务
		services, err := h.registry.ListAllServices(ctx)
		if err != nil {
			h.logger.Error("Failed to list all services", zap.Error(err))
			
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list services",
				"details": err.Error(),
			})
			return
		}

		totalInstances := 0
		for _, instances := range services {
			totalInstances += len(instances)
		}

		c.JSON(http.StatusOK, gin.H{
			"services": services,
			"service_count": len(services),
			"instance_count": totalInstances,
		})
	}
}

// GetService 获取服务详情
func (h *RegistryHandler) GetService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	// 获取查询参数
	healthyOnly := c.Query("healthy") == "true"
	limitStr := c.Query("limit")
	
	ctx := c.Request.Context()
	
	var instances interface{}
	var err error
	
	if healthyOnly {
		instances, err = h.registry.GetHealthyServices(ctx, serviceName)
	} else {
		instances, err = h.registry.ListServices(ctx, serviceName)
	}
	
	if err != nil {
		h.logger.Error("Failed to get service",
			zap.String("service", serviceName),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get service",
			"details": err.Error(),
		})
		return
	}

	// 应用限制
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if instanceList, ok := instances.([]*storage.ServiceInstance); ok {
				if len(instanceList) > limit {
					instances = instanceList[:limit]
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"instances": instances,
		"healthy_only": healthyOnly,
	})
}