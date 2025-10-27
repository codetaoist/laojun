package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/laojun-config-center/internal/storage"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	storage storage.ConfigStorage
	logger  *zap.Logger
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(storage storage.ConfigStorage, logger *zap.Logger) *ConfigHandler {
	return &ConfigHandler{
		storage: storage,
		logger:  logger,
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	item, err := h.storage.Get(ctx, service, environment, key)
	if err != nil {
		if _, ok := err.(*storage.ConfigNotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "config not found",
			})
			return
		}

		h.logger.Error("Failed to get config",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, item)
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
		Value       interface{}            `json:"value" binding:"required"`
		Type        string                 `json:"type"`
		Description string                 `json:"description"`
		Tags        []string               `json:"tags"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// 获取操作者信息
	operator := h.getOperator(c)

	item := &storage.ConfigItem{
		Service:     service,
		Environment: environment,
		Key:         key,
		Value:       req.Value,
		Type:        req.Type,
		Description: req.Description,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		CreatedBy:   operator,
		UpdatedBy:   operator,
	}

	// 设置默认类型
	if item.Type == "" {
		item.Type = "string"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.storage.Set(ctx, item); err != nil {
		if validationErr, ok := err.(*storage.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "validation failed: " + validationErr.Error(),
			})
			return
		}

		h.logger.Error("Failed to set config",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	h.logger.Info("Config set successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("key", key),
		zap.String("operator", operator),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "config set successfully",
		"version": item.Version,
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.storage.Delete(ctx, service, environment, key); err != nil {
		if _, ok := err.(*storage.ConfigNotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "config not found",
			})
			return
		}

		h.logger.Error("Failed to delete config",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	operator := h.getOperator(c)
	h.logger.Info("Config deleted successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("key", key),
		zap.String("operator", operator),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "config deleted successfully",
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	items, err := h.storage.List(ctx, service, environment)
	if err != nil {
		h.logger.Error("Failed to list configs",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": items,
		"count":   len(items),
	})
}

// SearchConfigs 搜索配置
func (h *ConfigHandler) SearchConfigs(c *gin.Context) {
	var query storage.SearchQuery

	// 从查询参数获取搜索条件
	query.Service = c.Query("service")
	query.Environment = c.Query("environment")
	query.Key = c.Query("key")
	query.Value = c.Query("value")

	// 解析标签
	if tagsStr := c.Query("tags"); tagsStr != "" {
		query.Tags = strings.Split(tagsStr, ",")
	}

	// 解析元数据
	query.Metadata = make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "metadata.") {
			metaKey := strings.TrimPrefix(key, "metadata.")
			if len(values) > 0 {
				query.Metadata[metaKey] = values[0]
			}
		}
	}

	// 解析分页参数
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	// 设置默认限制
	if query.Limit == 0 {
		query.Limit = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	items, err := h.storage.Search(ctx, &query)
	if err != nil {
		h.logger.Error("Failed to search configs", zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": items,
		"count":   len(items),
		"query":   query,
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

	// 解析限制参数
	limit := 50 // 默认限制
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	history, err := h.storage.GetHistory(ctx, service, environment, key, limit)
	if err != nil {
		h.logger.Error("Failed to get config history",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// RollbackConfig 回滚配置
func (h *ConfigHandler) RollbackConfig(c *gin.Context) {
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
		Version int64  `json:"version" binding:"required"`
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	operator := h.getOperator(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.storage.Rollback(ctx, service, environment, key, req.Version, operator); err != nil {
		if _, ok := err.(*storage.ConfigNotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "config or version not found",
			})
			return
		}

		h.logger.Error("Failed to rollback config",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.Int64("version", req.Version),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	h.logger.Info("Config rolled back successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("key", key),
		zap.Int64("version", req.Version),
		zap.String("operator", operator),
		zap.String("reason", req.Reason),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "config rolled back successfully",
		"version": req.Version,
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	data, err := h.storage.Backup(ctx, service, environment)
	if err != nil {
		h.logger.Error("Failed to backup configs",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	operator := h.getOperator(c)
	h.logger.Info("Configs backed up successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("operator", operator),
	)

	// 设置下载头
	filename := fmt.Sprintf("%s-%s-backup-%s.yaml", service, environment, time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/x-yaml")

	c.Data(http.StatusOK, "application/x-yaml", data)
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

	// 获取上传的文件
	file, err := c.FormFile("backup")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "backup file is required",
		})
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to open backup file",
		})
		return
	}
	defer src.Close()

	data := make([]byte, file.Size)
	if _, err := src.Read(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read backup file",
		})
		return
	}

	operator := h.getOperator(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	if err := h.storage.Restore(ctx, service, environment, data, operator); err != nil {
		h.logger.Error("Failed to restore configs",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to restore configs: " + err.Error(),
		})
		return
	}

	h.logger.Info("Configs restored successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("operator", operator),
		zap.String("filename", file.Filename),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "configs restored successfully",
	})
}



// WatchConfigs 监听配置变化
func (h *ConfigHandler) WatchConfigs(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")

	if service == "" || environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service and environment are required",
		})
		return
	}

	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// 开始监听
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	eventCh, err := h.storage.Watch(ctx, service, environment)
	if err != nil {
		h.logger.Error("Failed to start watching",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("Started watching configs",
		zap.String("service", service),
		zap.String("environment", environment),
	)

	// 发送事件到WebSocket
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				return
			}

			if err := conn.WriteJSON(event); err != nil {
				h.logger.Error("Failed to send watch event", zap.Error(err))
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// BatchSetConfigs 批量设置配置
func (h *ConfigHandler) BatchSetConfigs(c *gin.Context) {
	var req struct {
		Configs []struct {
			Service     string                 `json:"service" binding:"required"`
			Environment string                 `json:"environment" binding:"required"`
			Key         string                 `json:"key" binding:"required"`
			Value       interface{}            `json:"value" binding:"required"`
			Type        string                 `json:"type"`
			Description string                 `json:"description"`
			Tags        []string               `json:"tags"`
			Metadata    map[string]interface{} `json:"metadata"`
		} `json:"configs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	operator := h.getOperator(c)

	// 转换为存储格式
	var items []*storage.ConfigItem
	for _, config := range req.Configs {
		item := &storage.ConfigItem{
			Service:     config.Service,
			Environment: config.Environment,
			Key:         config.Key,
			Value:       config.Value,
			Type:        config.Type,
			Description: config.Description,
			Tags:        config.Tags,
			Metadata:    config.Metadata,
			CreatedBy:   operator,
			UpdatedBy:   operator,
		}

		if item.Type == "" {
			item.Type = "string"
		}

		items = append(items, item)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.storage.SetMultiple(ctx, items); err != nil {
		h.logger.Error("Failed to batch set configs",
			zap.String("operator", operator),
			zap.Int("count", len(items)),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	h.logger.Info("Batch set configs successfully",
		zap.String("operator", operator),
		zap.Int("count", len(items)),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "configs set successfully",
		"count":   len(items),
	})
}

// BatchDeleteConfigs 批量删除配置
func (h *ConfigHandler) BatchDeleteConfigs(c *gin.Context) {
	var req struct {
		Keys []struct {
			Service     string `json:"service" binding:"required"`
			Environment string `json:"environment" binding:"required"`
			Key         string `json:"key" binding:"required"`
		} `json:"keys" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// 转换为存储格式
	var keys []storage.ConfigKey
	for _, key := range req.Keys {
		keys = append(keys, storage.ConfigKey{
			Service:     key.Service,
			Environment: key.Environment,
			Key:         key.Key,
		})
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.storage.DeleteMultiple(ctx, keys); err != nil {
		operator := h.getOperator(c)
		h.logger.Error("Failed to batch delete configs",
			zap.String("operator", operator),
			zap.Int("count", len(keys)),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	operator := h.getOperator(c)
	h.logger.Info("Batch delete configs successfully",
		zap.String("operator", operator),
		zap.Int("count", len(keys)),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "configs deleted successfully",
		"count":   len(keys),
	})
}

// 辅助方法

func (h *ConfigHandler) getOperator(c *gin.Context) string {
	// 从请求头获取操作者信息
	if operator := c.GetHeader("X-Operator"); operator != "" {
		return operator
	}

	// 从认证信息获取
	if user, exists := c.Get("user"); exists {
		if userStr, ok := user.(string); ok {
			return userStr
		}
	}

	// 默认操作者
	return "anonymous"
}