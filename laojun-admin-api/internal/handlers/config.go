package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ConfigHandler 配置处理
type ConfigHandler struct {
	storage storage.ConfigStorage
	logger  *logrus.Logger
}

// NewConfigHandler 创建配置处理
func NewConfigHandler(storage storage.ConfigStorage) *ConfigHandler {
	return &ConfigHandler{
		storage: storage,
		logger:  logrus.New(),
	}
}

// GetConfig 获取配置
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")
	key := c.Param("key")

	if service == "" || environment == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service, environment and key are required",
		})
		return
	}

	item, err := h.storage.Get(c.Request.Context(), service, environment, key)
	if err != nil {
		h.logger.Errorf("Failed to get config: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": item,
	})
}

// SetConfig 设置配置
func (h *ConfigHandler) SetConfig(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")
	key := c.Param("key")

	if service == "" || environment == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service, environment and key are required",
		})
		return
	}

	var req struct {
		Value interface{}       `json:"value" binding:"required"`
		Type  string            `json:"type"`
		Tags  map[string]string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 获取用户信息（从认证中间件）
	user := c.GetString("user")
	if user == "" {
		user = "system"
	}

	item := &storage.ConfigItem{
		Key:         key,
		Value:       req.Value,
		Type:        req.Type,
		Service:     service,
		Environment: environment,
		Tags:        req.Tags,
		UpdatedBy:   user,
	}

	// 如果是新配置，设置创建人
	if _, err := h.storage.Get(c.Request.Context(), service, environment, key); err != nil {
		item.CreatedBy = user
	}

	if err := h.storage.Set(c.Request.Context(), item); err != nil {
		h.logger.Errorf("Failed to set config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config updated successfully",
		"data":    item,
	})
}

// DeleteConfig 删除配置
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")
	key := c.Param("key")

	if service == "" || environment == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service, environment and key are required",
		})
		return
	}

	if err := h.storage.Delete(c.Request.Context(), service, environment, key); err != nil {
		h.logger.Errorf("Failed to delete config: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config deleted successfully",
	})
}

// ListConfigs 列出配置
func (h *ConfigHandler) ListConfigs(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")

	if service == "" || environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service and environment are required",
		})
		return
	}

	items, err := h.storage.List(c.Request.Context(), service, environment)
	if err != nil {
		h.logger.Errorf("Failed to list configs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"count": len(items),
	})
}

// SearchConfigs 搜索配置
func (h *ConfigHandler) SearchConfigs(c *gin.Context) {
	var query struct {
		Tags map[string]string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	items, err := h.storage.GetByTags(c.Request.Context(), query.Tags)
	if err != nil {
		h.logger.Errorf("Failed to search configs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"count": len(items),
	})
}

// GetConfigHistory 获取配置历史
func (h *ConfigHandler) GetConfigHistory(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")
	key := c.Param("key")

	if service == "" || environment == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service, environment and key are required",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	items, err := h.storage.GetHistory(c.Request.Context(), service, environment, key, limit)
	if err != nil {
		h.logger.Errorf("Failed to get config history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"count": len(items),
	})
}

// BackupConfigs 备份配置
func (h *ConfigHandler) BackupConfigs(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")

	if service == "" || environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service and environment are required",
		})
		return
	}

	data, err := h.storage.Backup(c.Request.Context(), service, environment)
	if err != nil {
		h.logger.Errorf("Failed to backup configs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	filename := service + "_" + environment + "_" + time.Now().Format("20060102_150405") + ".json"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/json", data)
}

// RestoreConfigs 恢复配置
func (h *ConfigHandler) RestoreConfigs(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")

	if service == "" || environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service and environment are required",
		})
		return
	}

	var req struct {
		Data string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.storage.Restore(c.Request.Context(), service, environment, []byte(req.Data)); err != nil {
		h.logger.Errorf("Failed to restore configs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configs restored successfully",
	})
}

// WatchConfigs 监听配置变化（WebSocket 或 SSE）
func (h *ConfigHandler) WatchConfigs(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")

	if service == "" || environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service and environment are required",
		})
		return
	}

	// 这里应该实现 WebSocket 连接
	// 由于简化，暂时返回 SSE (Server-Sent Events)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	eventChan, err := h.storage.Watch(c.Request.Context(), service, environment)
	if err != nil {
		h.logger.Errorf("Failed to watch configs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 发送事件流
	for {
		select {
		case event := <-eventChan:
			c.SSEvent("config-change", event)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
