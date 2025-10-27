package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-shared/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UnifiedConfigHandler 统一配置管理处理器
type UnifiedConfigHandler struct {
	configManager config.ConfigManager
	logger        *zap.Logger
}

// NewUnifiedConfigHandler 创建统一配置管理处理器
func NewUnifiedConfigHandler(configManager config.ConfigManager, logger *zap.Logger) *UnifiedConfigHandler {
	return &UnifiedConfigHandler{
		configManager: configManager,
		logger:        logger,
	}
}

// GetConfig 获取配置
func (h *UnifiedConfigHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	value, exists := h.configManager.Get(key)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   key,
		"value": value,
	})
}

// GetConfigWithType 获取配置及其类型信息
func (h *UnifiedConfigHandler) GetConfigWithType(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	value, valueType, exists := h.configManager.GetWithType(key)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   key,
		"value": value,
		"type":  valueType,
	})
}

// SetConfig 设置配置
func (h *UnifiedConfigHandler) SetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	var req struct {
		Value interface{} `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configManager.Set(key, req.Value); err != nil {
		h.logger.Error("Failed to set config", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "config set successfully",
		"key":     key,
		"value":   req.Value,
	})
}

// DeleteConfig 删除配置
func (h *UnifiedConfigHandler) DeleteConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	if err := h.configManager.Delete(key); err != nil {
		h.logger.Error("Failed to delete config", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "config deleted successfully",
		"key":     key,
	})
}

// ListConfigs 列出所有配置
func (h *UnifiedConfigHandler) ListConfigs(c *gin.Context) {
	configs := h.configManager.List()
	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"count":   len(configs),
	})
}

// BatchSetConfigs 批量设置配置
func (h *UnifiedConfigHandler) BatchSetConfigs(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configManager.BatchSet(req); err != nil {
		h.logger.Error("Failed to batch set configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to batch set configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "configs set successfully",
		"count":   len(req),
	})
}

// GetConfigHistory 获取配置历史
func (h *UnifiedConfigHandler) GetConfigHistory(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	history := h.configManager.GetHistory(key, limit)
	c.JSON(http.StatusOK, gin.H{
		"key":     key,
		"history": history,
		"count":   len(history),
	})
}

// WatchConfigs 监听配置变化
func (h *UnifiedConfigHandler) WatchConfigs(c *gin.Context) {
	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建监听通道
	watchCh := make(chan config.ConfigChange, 100)
	h.configManager.Watch(watchCh)

	// 发送配置变化事件
	for {
		select {
		case change := <-watchCh:
			c.SSEvent("config-change", change)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

// ExistsConfig 检查配置是否存在
func (h *UnifiedConfigHandler) ExistsConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	exists := h.configManager.Exists(key)
	c.JSON(http.StatusOK, gin.H{
		"key":    key,
		"exists": exists,
	})
}

// HealthCheck 健康检查
func (h *UnifiedConfigHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "unified-config",
		"module":  "gateway",
	})
}