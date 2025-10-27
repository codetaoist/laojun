package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/circuit"
	"github.com/codetaoist/laojun-discovery/internal/config"
	"github.com/codetaoist/laojun-discovery/internal/loadbalancer"
	"github.com/codetaoist/laojun-discovery/internal/ratelimit"
	"github.com/codetaoist/laojun-discovery/internal/registry"
	"github.com/codetaoist/laojun-discovery/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnhancedDiscoveryHandler 增强的服务发现处理器
type EnhancedDiscoveryHandler struct {
	registry      *registry.ServiceRegistry
	lbManager     *loadbalancer.Manager
	circuitMgr    *circuit.Manager
	rateLimiter   *ratelimit.Manager
	config        *config.Config
	logger        *zap.Logger
}

// NewEnhancedDiscoveryHandler 创建增强的服务发现处理器
func NewEnhancedDiscoveryHandler(
	registry *registry.ServiceRegistry,
	config *config.Config,
	logger *zap.Logger,
) *EnhancedDiscoveryHandler {
	// 初始化负载均衡管理器
	lbManager := loadbalancer.NewManager(loadbalancer.Config{
		DefaultAlgorithm:   config.LoadBalance.Algorithm,
		HealthCheckEnabled: config.LoadBalance.HealthCheckEnabled,
		StatsEnabled:       config.LoadBalance.StatsEnabled,
	}, logger)

	// 初始化熔断器管理器
	circuitMgr := circuit.NewManager(circuit.Config{
		FailureThreshold: config.Circuit.FailureThreshold,
		SuccessThreshold: config.Circuit.SuccessThreshold,
		Timeout:          time.Duration(config.Circuit.Timeout) * time.Second,
		MaxRequests:      config.Circuit.MaxRequests,
		Interval:         time.Duration(config.Circuit.Interval) * time.Second,
		MinRequests:      config.Circuit.MinRequests,
		FailureRatio:     config.Circuit.FailureRatio,
	}, logger)

	// 初始化限流管理器
	rateLimiter := ratelimit.NewManager(ratelimit.Config{
		Algorithm: config.RateLimit.Algorithm,
		Rate:      config.RateLimit.Rate,
		Burst:     config.RateLimit.Burst,
		Window:    time.Duration(config.RateLimit.Window) * time.Second,
		Capacity:  config.RateLimit.Capacity,
	}, logger)

	return &EnhancedDiscoveryHandler{
		registry:    registry,
		lbManager:   lbManager,
		circuitMgr:  circuitMgr,
		rateLimiter: rateLimiter,
		config:      config,
		logger:      logger,
	}
}

// DiscoverWithLoadBalancing 带负载均衡的服务发现
func (h *EnhancedDiscoveryHandler) DiscoverWithLoadBalancing(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	// 限流检查
	if h.config.RateLimit.Enabled {
		limiter := h.rateLimiter.GetLimiter(fmt.Sprintf("discovery:%s", serviceName))
		if !limiter.Allow() {
			h.logger.Warn("Rate limit exceeded for service discovery",
				zap.String("service", serviceName))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
	}

	// 熔断器检查
	var instances []*storage.ServiceInstance
	var err error

	if h.config.Circuit.Enabled {
		breaker := h.circuitMgr.GetCircuitBreaker(fmt.Sprintf("discovery:%s", serviceName))
		result := breaker.Execute(func() (interface{}, error) {
			return h.getServiceInstances(c.Request.Context(), serviceName, c)
		})

		if result.Error != nil {
			h.logger.Error("Circuit breaker triggered for service discovery",
				zap.String("service", serviceName),
				zap.Error(result.Error))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service temporarily unavailable",
				"details": result.Error.Error(),
			})
			return
		}
		instances = result.Result.([]*storage.ServiceInstance)
	} else {
		instances, err = h.getServiceInstances(c.Request.Context(), serviceName, c)
		if err != nil {
			h.logger.Error("Failed to get service instances",
				zap.String("service", serviceName),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get service instances",
				"details": err.Error(),
			})
			return
		}
	}

	// 负载均衡选择
	algorithm := c.Query("algorithm")
	if algorithm == "" {
		algorithm = h.config.LoadBalance.Algorithm
	}

	lb := h.lbManager.GetLoadBalancer(serviceName, algorithm)
	selectedInstance, err := lb.Select(instances)
	if err != nil {
		h.logger.Error("Load balancer failed to select instance",
			zap.String("service", serviceName),
			zap.String("algorithm", algorithm),
			zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "No available instances",
			"details": err.Error(),
		})
		return
	}

	// 获取负载均衡统计信息
	stats := lb.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"service":           serviceName,
		"selected_instance": selectedInstance,
		"algorithm":         algorithm,
		"total_instances":   len(instances),
		"load_balancer_stats": gin.H{
			"total_requests":    stats.TotalRequests,
			"successful_requests": stats.SuccessfulRequests,
			"failed_requests":   stats.FailedRequests,
			"average_response_time": stats.AverageResponseTime.Milliseconds(),
		},
	})
}

// DiscoverMultipleInstances 发现多个服务实例（带负载均衡）
func (h *EnhancedDiscoveryHandler) DiscoverMultipleInstances(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	countStr := c.Query("count")
	count := 1
	if countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	// 限流检查
	if h.config.RateLimit.Enabled {
		limiter := h.rateLimiter.GetLimiter(fmt.Sprintf("discovery:%s", serviceName))
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
	}

	// 获取服务实例
	instances, err := h.getServiceInstances(c.Request.Context(), serviceName, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get service instances",
			"details": err.Error(),
		})
		return
	}

	// 负载均衡选择多个实例
	algorithm := c.Query("algorithm")
	if algorithm == "" {
		algorithm = h.config.LoadBalance.Algorithm
	}

	lb := h.lbManager.GetLoadBalancer(serviceName, algorithm)
	selectedInstances := make([]*storage.ServiceInstance, 0, count)

	for i := 0; i < count && len(instances) > 0; i++ {
		instance, err := lb.Select(instances)
		if err != nil {
			break
		}
		selectedInstances = append(selectedInstances, instance)
	}

	if len(selectedInstances) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "No available instances",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service":            serviceName,
		"selected_instances": selectedInstances,
		"algorithm":          algorithm,
		"requested_count":    count,
		"actual_count":       len(selectedInstances),
		"total_instances":    len(instances),
	})
}

// GetCircuitBreakerStatus 获取熔断器状态
func (h *EnhancedDiscoveryHandler) GetCircuitBreakerStatus(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	breakerName := fmt.Sprintf("discovery:%s", serviceName)
	breaker := h.circuitMgr.GetCircuitBreaker(breakerName)
	stats := breaker.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"service":         serviceName,
		"circuit_breaker": gin.H{
			"name":              breakerName,
			"state":             stats.State.String(),
			"total_requests":    stats.TotalRequests,
			"successful_requests": stats.SuccessfulRequests,
			"failed_requests":   stats.FailedRequests,
			"failure_rate":      stats.FailureRate,
			"next_attempt":      stats.NextAttempt,
		},
	})
}

// GetLoadBalancerStats 获取负载均衡器统计信息
func (h *EnhancedDiscoveryHandler) GetLoadBalancerStats(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	algorithm := c.Query("algorithm")
	if algorithm == "" {
		algorithm = h.config.LoadBalance.Algorithm
	}

	lb := h.lbManager.GetLoadBalancer(serviceName, algorithm)
	stats := lb.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"service":   serviceName,
		"algorithm": algorithm,
		"stats": gin.H{
			"total_requests":       stats.TotalRequests,
			"successful_requests":  stats.SuccessfulRequests,
			"failed_requests":      stats.FailedRequests,
			"average_response_time": stats.AverageResponseTime.Milliseconds(),
			"last_request_time":    stats.LastRequestTime,
		},
	})
}

// GetRateLimitStatus 获取限流状态
func (h *EnhancedDiscoveryHandler) GetRateLimitStatus(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	limiterName := fmt.Sprintf("discovery:%s", serviceName)
	limiter := h.rateLimiter.GetLimiter(limiterName)
	stats := limiter.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"service":     serviceName,
		"rate_limiter": gin.H{
			"name":            limiterName,
			"algorithm":       stats.Algorithm,
			"rate":            stats.Rate,
			"burst":           stats.Burst,
			"tokens":          stats.Tokens,
			"last_refill":     stats.LastRefill,
			"total_requests":  stats.TotalRequests,
			"allowed_requests": stats.AllowedRequests,
			"denied_requests": stats.DeniedRequests,
		},
	})
}

// getServiceInstances 获取服务实例的辅助方法
func (h *EnhancedDiscoveryHandler) getServiceInstances(ctx context.Context, serviceName string, c *gin.Context) ([]*storage.ServiceInstance, error) {
	// 获取查询参数
	tags := c.QueryArray("tag")
	healthyOnly := c.Query("healthy") == "true"

	var instances []*storage.ServiceInstance
	var err error

	if healthyOnly {
		instances, err = h.registry.GetHealthyServices(ctx, serviceName)
	} else {
		instances, err = h.registry.ListServices(ctx, serviceName)
	}

	if err != nil {
		return nil, err
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

	return instances, nil
}

// hasAllTags 检查实例是否包含所有指定标签
func (h *EnhancedDiscoveryHandler) hasAllTags(instanceTags, requiredTags []string) bool {
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