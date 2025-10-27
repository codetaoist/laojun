package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	serviceManager *services.ServiceManager
	logger         *zap.Logger
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(serviceManager *services.ServiceManager, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		serviceManager: serviceManager,
		logger:         logger,
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceHealth `json:"services"`
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Health 健康检查
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  make(map[string]ServiceHealth),
	}

	// 检查Redis连接
	redisHealth := h.checkRedis(ctx)
	response.Services["redis"] = redisHealth
	if redisHealth.Status != "healthy" {
		response.Status = "unhealthy"
	}

	// 检查服务发现
	discoveryHealth := h.checkDiscovery(ctx)
	response.Services["discovery"] = discoveryHealth
	if discoveryHealth.Status != "healthy" {
		response.Status = "degraded"
	}

	// 根据整体状态设置HTTP状态码
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if response.Status == "degraded" {
		statusCode = http.StatusOK // 降级状态仍返回200
	}

	c.JSON(statusCode, response)
}

// checkRedis 检查Redis连接
func (h *HealthHandler) checkRedis(ctx context.Context) ServiceHealth {
	redis := h.serviceManager.GetRedis()
	if redis == nil {
		return ServiceHealth{
			Status:  "unhealthy",
			Message: "Redis client not initialized",
		}
	}

	err := redis.Ping(ctx).Err()
	if err != nil {
		h.logger.Error("Redis health check failed", zap.Error(err))
		return ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return ServiceHealth{
		Status: "healthy",
	}
}

// checkDiscovery 检查服务发现
func (h *HealthHandler) checkDiscovery(ctx context.Context) ServiceHealth {
	discovery := h.serviceManager.GetDiscovery()
	if discovery == nil {
		return ServiceHealth{
			Status:  "unhealthy",
			Message: "Service discovery not initialized",
		}
	}

	// 尝试发现一个测试服务（这里只是检查服务发现是否可用）
	// 实际实现中可能需要根据具体的服务发现类型进行不同的健康检查
	return ServiceHealth{
		Status: "healthy",
	}
}