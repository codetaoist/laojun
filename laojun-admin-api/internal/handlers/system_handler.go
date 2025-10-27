package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/gin-gonic/gin"
)

// SystemHandler 系统相关处理
type SystemHandler struct {
	systemService *services.SystemService
}

func NewSystemHandler(systemService *services.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

// GetConfigs 获取系统配置
func (h *SystemHandler) GetConfigs(c *gin.Context) {
	settings, err := h.systemService.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统配置失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// SaveConfigs 保存系统配置
func (h *SystemHandler) SaveConfigs(c *gin.Context) {
	var req []models.SystemSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}
	if err := h.systemService.SaveSettings(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存系统配置失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
}

// GetLogs 获取审计日志
func (h *SystemHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	level := c.Query("level")
	module := c.Query("module")
	start := c.Query("startDate")
	end := c.Query("endDate")

	logs, total, err := h.systemService.GetAuditLogs(c.Request.Context(), page, pageSize, level, module, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审计日志失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs, "total": total})
}

// ClearLogs 清理审计日志
func (h *SystemHandler) ClearLogs(c *gin.Context) {
	level := c.Query("level")
	module := c.Query("module")
	before := c.Query("before")
	affected, err := h.systemService.ClearAuditLogs(c.Request.Context(), level, module, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清理审计日志失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "清理成功", "affected": affected})
}

// GetMetrics 获取性能指标
func (h *SystemHandler) GetMetrics(c *gin.Context) {
	m, err := h.systemService.GetMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取性能指标失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	systemInfo, err := h.systemService.GetSystemInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统信息失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": systemInfo})
}

// GetSystemConfig 获取系统配置
func (h *SystemHandler) GetSystemConfig(c *gin.Context) {
	h.GetConfigs(c)
}

// UpdateSystemConfig 更新系统配置
func (h *SystemHandler) UpdateSystemConfig(c *gin.Context) {
	h.SaveConfigs(c)
}

// GetSystemStatus 获取系统状态
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	h.GetSystemInfo(c)
}
