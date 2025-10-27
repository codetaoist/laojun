package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/circuit"
	"github.com/codetaoist/laojun-discovery/internal/config"
	"github.com/codetaoist/laojun-discovery/internal/ratelimit"
	"github.com/codetaoist/laojun-discovery/internal/registry"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnhancedRegistryHandler 增强的服务注册处理器
type EnhancedRegistryHandler struct {
	registry    *registry.ServiceRegistry
	circuitMgr  *circuit.Manager
	rateLimiter *ratelimit.Manager
	config      *config.Config
	logger      *zap.Logger
}

// NewEnhancedRegistryHandler 创建增强的服务注册处理器
func NewEnhancedRegistryHandler(
	registry *registry.ServiceRegistry,
	config *config.Config,
	logger *zap.Logger,
) *EnhancedRegistryHandler {
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

	return &EnhancedRegistryHandler{
		registry:    registry,
		circuitMgr:  circuitMgr,
		rateLimiter: rateLimiter,
		config:      config,
		logger:      logger,
	}
}

// RegisterService 注册服务（增强版）
func (h *EnhancedRegistryHandler) RegisterService(c *gin.Context) {
	// 限流检查
	if h.config.RateLimit.Enabled {
		limiter := h.rateLimiter.GetLimiter("service_registration")
		if !limiter.Allow() {
			h.logger.Warn("Rate limit exceeded for service registration")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded for service registration",
			})
			return
		}
	}

	var req registry.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// 验证请求
	if err := h.validateRegisterRequest(&req); err != nil {
		h.logger.Error("Registration request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Request validation failed",
			"details": err.Error(),
		})
		return
	}

	// 熔断器保护
	var serviceID string
	var err error

	if h.config.Circuit.Enabled {
		breakerName := fmt.Sprintf("registration:%s", req.Name)
		breaker := h.circuitMgr.GetCircuitBreaker(breakerName)
		
		result := breaker.Execute(func() (interface{}, error) {
			return h.registry.Register(c.Request.Context(), &req)
		})

		if result.Error != nil {
			h.logger.Error("Circuit breaker triggered for service registration",
				zap.String("service", req.Name),
				zap.Error(result.Error))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service registration temporarily unavailable",
				"details": result.Error.Error(),
			})
			return
		}
		serviceID = result.Result.(string)
	} else {
		serviceID, err = h.registry.Register(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to register service",
				zap.String("service", req.Name),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to register service",
				"details": err.Error(),
			})
			return
		}
	}

	h.logger.Info("Service registered successfully",
		zap.String("service", req.Name),
		zap.String("service_id", serviceID),
		zap.String("address", req.Address),
		zap.Int("port", req.Port))

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Service registered successfully",
		"service_id": serviceID,
		"service":    req.Name,
		"address":    req.Address,
		"port":       req.Port,
		"tags":       req.Tags,
		"ttl":        req.TTL,
	})
}

// DeregisterService 注销服务（增强版）
func (h *EnhancedRegistryHandler) DeregisterService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	// 限流检查
	if h.config.RateLimit.Enabled {
		limiter := h.rateLimiter.GetLimiter("service_deregistration")
		if !limiter.Allow() {
			h.logger.Warn("Rate limit exceeded for service deregistration")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded for service deregistration",
			})
			return
		}
	}

	// 熔断器保护
	var err error

	if h.config.Circuit.Enabled {
		breakerName := fmt.Sprintf("deregistration:%s", serviceID)
		breaker := h.circuitMgr.GetCircuitBreaker(breakerName)
		
		result := breaker.Execute(func() (interface{}, error) {
			return nil, h.registry.Deregister(c.Request.Context(), serviceID)
		})

		if result.Error != nil {
			h.logger.Error("Circuit breaker triggered for service deregistration",
				zap.String("service_id", serviceID),
				zap.Error(result.Error))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service deregistration temporarily unavailable",
				"details": result.Error.Error(),
			})
			return
		}
	} else {
		err = h.registry.Deregister(c.Request.Context(), serviceID)
		if err != nil {
			h.logger.Error("Failed to deregister service",
				zap.String("service_id", serviceID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to deregister service",
				"details": err.Error(),
			})
			return
		}
	}

	h.logger.Info("Service deregistered successfully",
		zap.String("service_id", serviceID))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Service deregistered successfully",
		"service_id": serviceID,
	})
}

// UpdateServiceHealth 更新服务健康状态（增强版）
func (h *EnhancedRegistryHandler) UpdateServiceHealth(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	var healthUpdate struct {
		Status  string            `json:"status" binding:"required"`
		Output  string            `json:"output,omitempty"`
		Details map[string]string `json:"details,omitempty"`
	}

	if err := c.ShouldBindJSON(&healthUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid health update request",
			"details": err.Error(),
		})
		return
	}

	// 验证健康状态
	if !h.isValidHealthStatus(healthUpdate.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid health status. Must be one of: passing, warning, critical",
		})
		return
	}

	// 限流检查
	if h.config.RateLimit.Enabled {
		limiter := h.rateLimiter.GetLimiter("health_update")
		if !limiter.Allow() {
			h.logger.Warn("Rate limit exceeded for health update")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded for health update",
			})
			return
		}
	}

	// 熔断器保护
	if h.config.Circuit.Enabled {
		breakerName := fmt.Sprintf("health_update:%s", serviceID)
		breaker := h.circuitMgr.GetCircuitBreaker(breakerName)
		
		result := breaker.Execute(func() (interface{}, error) {
			return nil, h.registry.UpdateHealth(c.Request.Context(), serviceID, healthUpdate.Status, healthUpdate.Output)
		})

		if result.Error != nil {
			h.logger.Error("Circuit breaker triggered for health update",
				zap.String("service_id", serviceID),
				zap.Error(result.Error))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Health update temporarily unavailable",
				"details": result.Error.Error(),
			})
			return
		}
	} else {
		err := h.registry.UpdateHealth(c.Request.Context(), serviceID, healthUpdate.Status, healthUpdate.Output)
		if err != nil {
			h.logger.Error("Failed to update service health",
				zap.String("service_id", serviceID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update service health",
				"details": err.Error(),
			})
			return
		}
	}

	h.logger.Info("Service health updated successfully",
		zap.String("service_id", serviceID),
		zap.String("status", healthUpdate.Status))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Service health updated successfully",
		"service_id": serviceID,
		"status":     healthUpdate.Status,
		"timestamp":  time.Now().UTC(),
	})
}

// GetRegistrationStats 获取注册统计信息
func (h *EnhancedRegistryHandler) GetRegistrationStats(c *gin.Context) {
	// 获取限流统计
	regLimiter := h.rateLimiter.GetLimiter("service_registration")
	deregLimiter := h.rateLimiter.GetLimiter("service_deregistration")
	healthLimiter := h.rateLimiter.GetLimiter("health_update")

	regStats := regLimiter.GetStats()
	deregStats := deregLimiter.GetStats()
	healthStats := healthLimiter.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"rate_limiting": gin.H{
			"registration": gin.H{
				"total_requests":   regStats.TotalRequests,
				"allowed_requests": regStats.AllowedRequests,
				"denied_requests":  regStats.DeniedRequests,
			},
			"deregistration": gin.H{
				"total_requests":   deregStats.TotalRequests,
				"allowed_requests": deregStats.AllowedRequests,
				"denied_requests":  deregStats.DeniedRequests,
			},
			"health_update": gin.H{
				"total_requests":   healthStats.TotalRequests,
				"allowed_requests": healthStats.AllowedRequests,
				"denied_requests":  healthStats.DeniedRequests,
			},
		},
		"timestamp": time.Now().UTC(),
	})
}

// validateRegisterRequest 验证注册请求
func (h *EnhancedRegistryHandler) validateRegisterRequest(req *registry.RegisterRequest) error {
	if req.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if req.Address == "" {
		return fmt.Errorf("service address is required")
	}
	if req.Port <= 0 || req.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", req.Port)
	}
	if req.TTL < 0 {
		return fmt.Errorf("TTL cannot be negative")
	}
	return nil
}

// isValidHealthStatus 验证健康状态
func (h *EnhancedRegistryHandler) isValidHealthStatus(status string) bool {
	validStatuses := []string{"passing", "warning", "critical"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}