package handlers

import (
	"net/http"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	serviceManager *services.ServiceManager
	logger         *zap.Logger
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]interface{} `json:"services"`
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(serviceManager *services.ServiceManager, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		serviceManager: serviceManager,
		logger:         logger,
	}
}

// Health 健康检查端点
func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	
	// 检查服务管理器健康状态
	isHealthy := h.serviceManager.IsHealthy(ctx)
	
	status := "healthy"
	httpStatus := http.StatusOK
	if !isHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	// 获取统计信息
	stats, err := h.serviceManager.GetStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get stats", zap.Error(err))
		stats = map[string]interface{}{
			"error": "Failed to get stats",
		}
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services:  stats,
	}

	c.JSON(httpStatus, response)
}

// Ready 就绪检查端点
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	
	// 检查服务管理器就绪状态
	isReady := h.serviceManager.IsReady(ctx)
	
	status := "ready"
	httpStatus := http.StatusOK
	if !isReady {
		status = "not ready"
		httpStatus = http.StatusServiceUnavailable
	}

	response := gin.H{
		"status":    status,
		"timestamp": time.Now(),
	}

	c.JSON(httpStatus, response)
}