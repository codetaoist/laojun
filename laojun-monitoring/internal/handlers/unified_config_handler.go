package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
)

// UnifiedConfigHandler 统一配置管理处理器
type UnifiedConfigHandler struct {
	configManager sharedconfig.ConfigManager
	logger        *zap.Logger
}

// NewUnifiedConfigHandler 创建新的统一配置管理处理器
func NewUnifiedConfigHandler(configManager sharedconfig.ConfigManager, logger *zap.Logger) *UnifiedConfigHandler {
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

// GetConfigWithType 获取配置及其类型
func (h *UnifiedConfigHandler) GetConfigWithType(c *gin.Context) {
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
		"type":  sharedconfig.GetValueType(value),
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
		"key":   key,
		"value": req.Value,
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

	c.JSON(http.StatusOK, gin.H{"message": "config deleted successfully"})
}

// ListConfigs 列出所有配置
func (h *UnifiedConfigHandler) ListConfigs(c *gin.Context) {
	configs := h.configManager.List()
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// BatchSetConfigs 批量设置配置
func (h *UnifiedConfigHandler) BatchSetConfigs(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results := make(map[string]interface{})
	for key, value := range req {
		if err := h.configManager.Set(key, value); err != nil {
			h.logger.Error("Failed to set config in batch", zap.String("key", key), zap.Error(err))
			results[key] = gin.H{"error": err.Error()}
		} else {
			results[key] = gin.H{"success": true}
		}
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
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
	})
}

// WatchConfigs 监听配置变化
func (h *UnifiedConfigHandler) WatchConfigs(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建监听通道
	ch := h.configManager.Watch(key)
	defer h.configManager.Unwatch(key, ch)

	// 发送当前值
	if value, exists := h.configManager.Get(key); exists {
		c.SSEvent("config", gin.H{
			"key":   key,
			"value": value,
			"type":  "current",
		})
		c.Writer.Flush()
	}

	// 监听变化
	for {
		select {
		case event := <-ch:
			c.SSEvent("config", gin.H{
				"key":   key,
				"value": event.Value,
				"type":  "update",
			})
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

	_, exists := h.configManager.Get(key)
	c.JSON(http.StatusOK, gin.H{
		"key":    key,
		"exists": exists,
	})
}

// HealthCheck 健康检查
func (h *UnifiedConfigHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "unified-config-handler",
	})
}