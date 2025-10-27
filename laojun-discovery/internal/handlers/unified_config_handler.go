package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/laojun-shared/config"
)

// UnifiedConfigHandler 统一配置处理器
type UnifiedConfigHandler struct {
	manager config.ConfigManager
	logger  *zap.Logger
}

// NewUnifiedConfigHandler 创建统一配置处理器
func NewUnifiedConfigHandler(manager config.ConfigManager, logger *zap.Logger) *UnifiedConfigHandler {
	return &UnifiedConfigHandler{
		manager: manager,
		logger:  logger,
	}
}

// GetConfig 获取配置
func (h *UnifiedConfigHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	value, err := h.manager.Get(ctx, key)
	if err != nil {
		if err == config.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "config not found",
			})
			return
		}
		h.logger.Error("Failed to get config", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   key,
		"value": value,
	})
}

// GetConfigWithType 获取指定类型的配置
func (h *UnifiedConfigHandler) GetConfigWithType(c *gin.Context) {
	key := c.Param("key")
	configType := c.Query("type")
	
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var result interface{}
	var err error

	switch strings.ToLower(configType) {
	case "string":
		result, err = h.manager.GetString(ctx, key)
	case "int":
		result, err = h.manager.GetInt(ctx, key)
	case "bool":
		result, err = h.manager.GetBool(ctx, key)
	case "float":
		result, err = h.manager.GetFloat(ctx, key)
	case "duration":
		result, err = h.manager.GetDuration(ctx, key)
	default:
		result, err = h.manager.Get(ctx, key)
	}

	if err != nil {
		if err == config.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "config not found",
			})
			return
		}
		h.logger.Error("Failed to get config", zap.String("key", key), zap.String("type", configType), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   key,
		"value": result,
		"type":  configType,
	})
}

// SetConfig 设置配置
func (h *UnifiedConfigHandler) SetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key is required",
		})
		return
	}

	var req struct {
		Value   interface{}            `json:"value" binding:"required"`
		Options *config.ConfigOptions `json:"options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.manager.Set(ctx, key, req.Value, req.Options)
	if err != nil {
		h.logger.Error("Failed to set config", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "config set successfully",
		"key":     key,
	})
}

// DeleteConfig 删除配置
func (h *UnifiedConfigHandler) DeleteConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.manager.Delete(ctx, key)
	if err != nil {
		if err == config.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "config not found",
			})
			return
		}
		h.logger.Error("Failed to delete config", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "config deleted successfully",
		"key":     key,
	})
}

// ListConfigs 列出配置
func (h *UnifiedConfigHandler) ListConfigs(c *gin.Context) {
	prefix := c.Query("prefix")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 100 // 默认限制
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	keys, err := h.manager.List(ctx, prefix)
	if err != nil {
		h.logger.Error("Failed to list configs", zap.String("prefix", prefix), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// 应用分页
	total := len(keys)
	start := offset
	end := offset + limit

	if start >= total {
		keys = []string{}
	} else {
		if end > total {
			end = total
		}
		keys = keys[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":   keys,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// BatchSetConfigs 批量设置配置
func (h *UnifiedConfigHandler) BatchSetConfigs(c *gin.Context) {
	var req struct {
		Configs map[string]interface{} `json:"configs" binding:"required"`
		Options *config.ConfigOptions  `json:"options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	err := h.manager.SetMultiple(ctx, req.Configs, req.Options)
	if err != nil {
		h.logger.Error("Failed to batch set configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "configs set successfully",
		"count":   len(req.Configs),
	})
}

// GetConfigHistory 获取配置历史
func (h *UnifiedConfigHandler) GetConfigHistory(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key is required",
		})
		return
	}

	limitStr := c.Query("limit")
	limit := 10 // 默认限制
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	history, err := h.manager.GetHistory(ctx, key, limit)
	if err != nil {
		h.logger.Error("Failed to get config history", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":     key,
		"history": history,
	})
}

// WatchConfigs 监听配置变化
func (h *UnifiedConfigHandler) WatchConfigs(c *gin.Context) {
	keys := c.QueryArray("key")
	if len(keys) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one key is required",
		})
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	watcher, err := h.manager.Watch(ctx, keys...)
	if err != nil {
		h.logger.Error("Failed to create watcher", zap.Strings("keys", keys), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	defer watcher.Close()

	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 发送初始连接确认
	c.SSEvent("connected", gin.H{
		"message": "watching configs",
		"keys":    keys,
	})
	c.Writer.Flush()

	// 监听配置变化
	for {
		select {
		case event := <-watcher.Events():
			c.SSEvent("config_change", gin.H{
				"key":       event.Key,
				"value":     event.Value,
				"operation": event.Operation,
				"timestamp": event.Timestamp,
			})
			c.Writer.Flush()
		case err := <-watcher.Errors():
			h.logger.Error("Watcher error", zap.Error(err))
			c.SSEvent("error", gin.H{
				"error": err.Error(),
			})
			c.Writer.Flush()
			return
		case <-ctx.Done():
			return
		}
	}
}

// HealthCheck 健康检查
func (h *UnifiedConfigHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.manager.Health(ctx)
	if err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// ExistsConfig 检查配置是否存在
func (h *UnifiedConfigHandler) ExistsConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	exists, err := h.manager.Exists(ctx, key)
	if err != nil {
		h.logger.Error("Failed to check config existence", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":    key,
		"exists": exists,
	})
}