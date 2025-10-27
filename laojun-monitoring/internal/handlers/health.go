package handlers

import (
	"net/http"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	monitoringService *services.MonitoringService
	logger           *zap.Logger
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(monitoringService *services.MonitoringService, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		monitoringService: monitoringService,
		logger:           logger,
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Service   string                 `json:"service"`
	Uptime    string                 `json:"uptime"`
	Details   map[string]interface{} `json:"details"`
}

// Health 健康检查端点
func (h *HealthHandler) Health(c *gin.Context) {
	health := h.monitoringService.Health()
	
	response := HealthResponse{
		Status:    health["status"].(string),
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Service:   "laojun-monitoring",
		Uptime:    health["uptime"].(string),
		Details:   health,
	}
	
	// 根据健康状态设置 HTTP 状态码
	statusCode := http.StatusOK
	if response.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	} else if response.Status == "unhealthy" {
		statusCode = http.StatusInternalServerError
	}
	
	c.JSON(statusCode, response)
}

// Ready 就绪检查端点
func (h *HealthHandler) Ready(c *gin.Context) {
	ready := h.monitoringService.Ready()
	
	response := gin.H{
		"ready":     ready,
		"timestamp": time.Now(),
		"service":   "laojun-monitoring",
	}
	
	if ready {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}