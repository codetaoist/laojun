package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-discovery/internal/registry"
	"github.com/codetaoist/laojun-discovery/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DiscoveryHandler 服务发现处理器
type DiscoveryHandler struct {
	registry *registry.ServiceRegistry
	logger   *zap.Logger
}

// NewDiscoveryHandler 创建服务发现处理器
func NewDiscoveryHandler(registry *registry.ServiceRegistry, logger *zap.Logger) *DiscoveryHandler {
	return &DiscoveryHandler{
		registry: registry,
		logger:   logger,
	}
}

// DiscoverServices 发现所有服务
func (h *DiscoveryHandler) DiscoverServices(c *gin.Context) {
	ctx := c.Request.Context()
	
	// 获取查询参数
	tags := c.QueryArray("tag")
	
	services, err := h.registry.ListAllServices(ctx)
	if err != nil {
		h.logger.Error("Failed to discover services", zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to discover services",
			"details": err.Error(),
		})
		return
	}

	// 按标签过滤
	if len(tags) > 0 {
		filteredServices := make(map[string][]*storage.ServiceInstance)
		for serviceName, instances := range services {
			filteredInstances := make([]*storage.ServiceInstance, 0)
			for _, instance := range instances {
				if h.hasAllTags(instance.Tags, tags) {
					filteredInstances = append(filteredInstances, instance)
				}
			}
			if len(filteredInstances) > 0 {
				filteredServices[serviceName] = filteredInstances
			}
		}
		services = filteredServices
	}

	// 统计信息
	serviceCount := len(services)
	totalInstances := 0
	healthyInstances := 0

	for _, instances := range services {
		totalInstances += len(instances)
		for _, instance := range instances {
			if instance.Health.Status == "passing" {
				healthyInstances++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"statistics": gin.H{
			"service_count":     serviceCount,
			"total_instances":   totalInstances,
			"healthy_instances": healthyInstances,
		},
	})
}

// DiscoverService 发现指定服务
func (h *DiscoveryHandler) DiscoverService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	// 获取查询参数
	tags := c.QueryArray("tag")
	limitStr := c.Query("limit")
	
	ctx := c.Request.Context()
	
	instances, err := h.registry.ListServices(ctx, serviceName)
	if err != nil {
		h.logger.Error("Failed to discover service",
			zap.String("service", serviceName),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to discover service",
			"details": err.Error(),
		})
		return
	}

	// 按标签过滤
	if len(tags) > 0 {
		filteredInstances := make([]*storage.ServiceInstance, 0)
		for _, instance := range instances {
			if h.hasAllTags(instance.Tags, tags) {
				filteredInstances = append(filteredInstances, instance)
			}
		}
		instances = filteredInstances
	}

	// 应用限制
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if len(instances) > limit {
				instances = instances[:limit]
			}
		}
	}

	// 统计健康实例
	healthyCount := 0
	for _, instance := range instances {
		if instance.Health.Status == "passing" {
			healthyCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"instances": instances,
		"statistics": gin.H{
			"total_instances":   len(instances),
			"healthy_instances": healthyCount,
		},
		"filters": gin.H{
			"tags":  tags,
			"limit": limitStr,
		},
	})
}

// GetHealthyInstances 获取健康的服务实例
func (h *DiscoveryHandler) GetHealthyInstances(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	// 获取查询参数
	tags := c.QueryArray("tag")
	limitStr := c.Query("limit")
	
	ctx := c.Request.Context()
	
	instances, err := h.registry.GetHealthyServices(ctx, serviceName)
	if err != nil {
		h.logger.Error("Failed to get healthy instances",
			zap.String("service", serviceName),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get healthy instances",
			"details": err.Error(),
		})
		return
	}

	// 按标签过滤
	if len(tags) > 0 {
		filteredInstances := make([]*storage.ServiceInstance, 0)
		for _, instance := range instances {
			if h.hasAllTags(instance.Tags, tags) {
				filteredInstances = append(filteredInstances, instance)
			}
		}
		instances = filteredInstances
	}

	// 应用限制
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if len(instances) > limit {
				instances = instances[:limit]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"instances": instances,
		"count": len(instances),
		"filters": gin.H{
			"tags":        tags,
			"limit":       limitStr,
			"healthy_only": true,
		},
	})
}

// hasAllTags 检查实例是否包含所有指定标签
func (h *DiscoveryHandler) hasAllTags(instanceTags, requiredTags []string) bool {
	if len(requiredTags) == 0 {
		return true
	}

	tagMap := make(map[string]bool)
	for _, tag := range instanceTags {
		tagMap[tag] = true
	}

	for _, requiredTag := range requiredTags {
		if !tagMap[requiredTag] {
			return false
		}
	}

	return true
}