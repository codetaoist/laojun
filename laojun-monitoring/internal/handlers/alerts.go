package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AlertHandler 告警处理器
type AlertHandler struct {
	monitoringService *services.MonitoringService
	logger           *zap.Logger
}

// NewAlertHandler 创建告警处理器
func NewAlertHandler(monitoringService *services.MonitoringService, logger *zap.Logger) *AlertHandler {
	return &AlertHandler{
		monitoringService: monitoringService,
		logger:           logger,
	}
}

// Alert 告警信息
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	FiredAt     *time.Time        `json:"fired_at,omitempty"`
	ResolvedAt  *time.Time        `json:"resolved_at,omitempty"`
}

// CreateAlertRequest 创建告警请求
type CreateAlertRequest struct {
	Name        string            `json:"name" binding:"required"`
	Query       string            `json:"query" binding:"required"`
	Duration    string            `json:"duration" binding:"required"`
	Severity    string            `json:"severity" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// UpdateAlertRequest 更新告警请求
type UpdateAlertRequest struct {
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Duration    string            `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// SilenceRequest 静默告警请求
type SilenceRequest struct {
	Duration string `json:"duration" binding:"required"`
	Comment  string `json:"comment"`
}

// AlertResponse 告警响应
type AlertResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error,omitempty"`
}

// ListAlerts 获取告警列表
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	// 获取查询参数
	status := c.Query("status")
	severity := c.Query("severity")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	
	limit := 50 // 默认限制
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	offset := 0 // 默认偏移
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// 模拟告警数据
	alerts := h.getMockAlerts(status, severity, limit, offset)
	
	response := AlertResponse{
		Status: "success",
		Data: gin.H{
			"alerts": alerts,
			"total":  len(alerts),
			"limit":  limit,
			"offset": offset,
			"filters": gin.H{
				"status":   status,
				"severity": severity,
			},
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// CreateAlert 创建告警
func (h *AlertHandler) CreateAlert(c *gin.Context) {
	var req CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create alert request", zap.Error(err))
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	
	// 验证持续时间格式
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  "Invalid duration format",
		})
		return
	}
	
	// 验证严重级别
	validSeverities := map[string]bool{
		"critical": true,
		"warning":  true,
		"info":     true,
	}
	
	if !validSeverities[req.Severity] {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  "Invalid severity. Must be one of: critical, warning, info",
		})
		return
	}
	
	// 创建告警
	alert := Alert{
		ID:          generateAlertID(),
		Name:        req.Name,
		Query:       req.Query,
		Duration:    duration,
		Severity:    req.Severity,
		Status:      "pending",
		Labels:      req.Labels,
		Annotations: req.Annotations,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	h.logger.Info("Creating alert",
		zap.String("id", alert.ID),
		zap.String("name", alert.Name),
		zap.String("severity", alert.Severity))
	
	// 这里应该将告警保存到存储中
	// 由于这是模拟，我们只是返回创建的告警
	
	response := AlertResponse{
		Status: "success",
		Data:   alert,
	}
	
	c.JSON(http.StatusCreated, response)
}

// GetAlert 获取单个告警
func (h *AlertHandler) GetAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  "Alert ID is required",
		})
		return
	}
	
	// 模拟获取告警
	alert := h.getMockAlert(alertID)
	if alert == nil {
		c.JSON(http.StatusNotFound, AlertResponse{
			Status: "error",
			Error:  "Alert not found",
		})
		return
	}
	
	response := AlertResponse{
		Status: "success",
		Data:   alert,
	}
	
	c.JSON(http.StatusOK, response)
}

// UpdateAlert 更新告警
func (h *AlertHandler) UpdateAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  "Alert ID is required",
		})
		return
	}
	
	var req UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update alert request", zap.Error(err))
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	
	// 获取现有告警
	alert := h.getMockAlert(alertID)
	if alert == nil {
		c.JSON(http.StatusNotFound, AlertResponse{
			Status: "error",
			Error:  "Alert not found",
		})
		return
	}
	
	// 更新告警字段
	if req.Name != "" {
		alert.Name = req.Name
	}
	if req.Query != "" {
		alert.Query = req.Query
	}
	if req.Duration != "" {
		if duration, err := time.ParseDuration(req.Duration); err == nil {
			alert.Duration = duration
		}
	}
	if req.Severity != "" {
		alert.Severity = req.Severity
	}
	if req.Labels != nil {
		alert.Labels = req.Labels
	}
	if req.Annotations != nil {
		alert.Annotations = req.Annotations
	}
	alert.UpdatedAt = time.Now()
	
	h.logger.Info("Updating alert",
		zap.String("id", alertID),
		zap.String("name", alert.Name))
	
	response := AlertResponse{
		Status: "success",
		Data:   alert,
	}
	
	c.JSON(http.StatusOK, response)
}

// DeleteAlert 删除告警
func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  "Alert ID is required",
		})
		return
	}
	
	// 检查告警是否存在
	alert := h.getMockAlert(alertID)
	if alert == nil {
		c.JSON(http.StatusNotFound, AlertResponse{
			Status: "error",
			Error:  "Alert not found",
		})
		return
	}
	
	h.logger.Info("Deleting alert", zap.String("id", alertID))
	
	// 这里应该从存储中删除告警
	
	response := AlertResponse{
		Status: "success",
		Data: gin.H{
			"message": "Alert deleted successfully",
			"id":      alertID,
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// SilenceAlert 静默告警
func (h *AlertHandler) SilenceAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  "Alert ID is required",
		})
		return
	}
	
	var req SilenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid silence request", zap.Error(err))
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	
	// 验证持续时间
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Status: "error",
			Error:  "Invalid duration format",
		})
		return
	}
	
	// 检查告警是否存在
	alert := h.getMockAlert(alertID)
	if alert == nil {
		c.JSON(http.StatusNotFound, AlertResponse{
			Status: "error",
			Error:  "Alert not found",
		})
		return
	}
	
	h.logger.Info("Silencing alert",
		zap.String("id", alertID),
		zap.Duration("duration", duration),
		zap.String("comment", req.Comment))
	
	// 这里应该实际静默告警
	
	response := AlertResponse{
		Status: "success",
		Data: gin.H{
			"message":   "Alert silenced successfully",
			"id":        alertID,
			"duration":  duration.String(),
			"comment":   req.Comment,
			"until":     time.Now().Add(duration),
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// 辅助函数

// generateAlertID 生成告警ID
func generateAlertID() string {
	return "alert-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// getMockAlerts 获取模拟告警列表
func (h *AlertHandler) getMockAlerts(status, severity string, limit, offset int) []Alert {
	alerts := []Alert{
		{
			ID:       "alert-001",
			Name:     "High CPU Usage",
			Query:    "cpu_usage > 80",
			Duration: 5 * time.Minute,
			Severity: "warning",
			Status:   "firing",
			Labels: map[string]string{
				"service": "api-server",
				"env":     "production",
			},
			Annotations: map[string]string{
				"description": "CPU usage is above 80%",
				"runbook":     "https://runbook.example.com/cpu-high",
			},
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-30 * time.Minute),
			FiredAt:   timePtr(time.Now().Add(-30 * time.Minute)),
		},
		{
			ID:       "alert-002",
			Name:     "Memory Usage Critical",
			Query:    "memory_usage > 95",
			Duration: 2 * time.Minute,
			Severity: "critical",
			Status:   "firing",
			Labels: map[string]string{
				"service": "database",
				"env":     "production",
			},
			Annotations: map[string]string{
				"description": "Memory usage is critically high",
				"runbook":     "https://runbook.example.com/memory-critical",
			},
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-10 * time.Minute),
			FiredAt:   timePtr(time.Now().Add(-10 * time.Minute)),
		},
		{
			ID:       "alert-003",
			Name:     "Disk Space Low",
			Query:    "disk_free < 10",
			Duration: 10 * time.Minute,
			Severity: "info",
			Status:   "resolved",
			Labels: map[string]string{
				"service": "storage",
				"env":     "staging",
			},
			Annotations: map[string]string{
				"description": "Disk space is running low",
				"runbook":     "https://runbook.example.com/disk-space",
			},
			CreatedAt:  time.Now().Add(-3 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
			FiredAt:    timePtr(time.Now().Add(-2 * time.Hour)),
			ResolvedAt: timePtr(time.Now().Add(-1 * time.Hour)),
		},
	}
	
	// 应用过滤器
	filteredAlerts := make([]Alert, 0)
	for _, alert := range alerts {
		if status != "" && alert.Status != status {
			continue
		}
		if severity != "" && alert.Severity != severity {
			continue
		}
		filteredAlerts = append(filteredAlerts, alert)
	}
	
	// 应用分页
	start := offset
	end := offset + limit
	if start >= len(filteredAlerts) {
		return []Alert{}
	}
	if end > len(filteredAlerts) {
		end = len(filteredAlerts)
	}
	
	return filteredAlerts[start:end]
}

// getMockAlert 获取模拟单个告警
func (h *AlertHandler) getMockAlert(id string) *Alert {
	alerts := h.getMockAlerts("", "", 100, 0)
	for _, alert := range alerts {
		if alert.ID == id {
			return &alert
		}
	}
	return nil
}

// timePtr 返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}