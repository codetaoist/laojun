package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codetaoist/laojun-marketplace-api/internal/events"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/sirupsen/logrus"
)

// PluginEventHandler 插件事件处理器
type PluginEventHandler struct {
	pluginService    *services.PluginService
	categoryService  *services.CategoryService
	developerService *services.DeveloperService
	logger           *logrus.Logger
}

// NewPluginEventHandler 创建插件事件处理器
func NewPluginEventHandler(
	pluginService *services.PluginService,
	categoryService *services.CategoryService,
	developerService *services.DeveloperService,
	logger *logrus.Logger,
) *PluginEventHandler {
	return &PluginEventHandler{
		pluginService:    pluginService,
		categoryService:  categoryService,
		developerService: developerService,
		logger:           logger,
	}
}

// Handle 处理事件
func (h *PluginEventHandler) Handle(ctx context.Context, event *events.Event) error {
	h.logger.Infof("Handling plugin event: %s (ID: %s)", event.Type, event.ID)

	switch event.Type {
	case events.PluginCreated:
		return h.handlePluginCreated(ctx, event)
	case events.PluginUpdated:
		return h.handlePluginUpdated(ctx, event)
	case events.PluginDeleted:
		return h.handlePluginDeleted(ctx, event)
	case events.PluginPublished:
		return h.handlePluginPublished(ctx, event)
	case events.PluginUnpublished:
		return h.handlePluginUnpublished(ctx, event)
	default:
		h.logger.Warnf("Unknown plugin event type: %s", event.Type)
		return nil
	}
}

// GetEventTypes 获取处理的事件类型
func (h *PluginEventHandler) GetEventTypes() []events.EventType {
	return []events.EventType{
		events.PluginCreated,
		events.PluginUpdated,
		events.PluginDeleted,
		events.PluginPublished,
		events.PluginUnpublished,
	}
}

// handlePluginCreated 处理插件创建事件
func (h *PluginEventHandler) handlePluginCreated(ctx context.Context, event *events.Event) error {
	var eventData events.PluginCreatedEvent
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal plugin created event: %w", err)
	}

	h.logger.Infof("Processing plugin created: %s by developer %s", 
		eventData.Name, eventData.DeveloperID)

	// 更新分类统计
	if err := h.updateCategoryStats(ctx, eventData.CategoryID, 1); err != nil {
		h.logger.Errorf("Failed to update category stats: %v", err)
	}

	// 更新开发者统计
	if err := h.updateDeveloperStats(ctx, eventData.DeveloperID, "plugin_created"); err != nil {
		h.logger.Errorf("Failed to update developer stats: %v", err)
	}

	// 这里可以添加其他业务逻辑
	// 例如：发送通知、索引更新等

	return nil
}

// handlePluginUpdated 处理插件更新事件
func (h *PluginEventHandler) handlePluginUpdated(ctx context.Context, event *events.Event) error {
	var eventData events.PluginCreatedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal plugin updated event: %w", err)
	}

	h.logger.Infof("Processing plugin updated: %s by developer %s", 
		eventData.Name, eventData.DeveloperID)

	// 这里可以添加插件更新后的业务逻辑
	// 例如：清除缓存、更新搜索索引等

	return nil
}

// handlePluginDeleted 处理插件删除事件
func (h *PluginEventHandler) handlePluginDeleted(ctx context.Context, event *events.Event) error {
	var eventData events.PluginCreatedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal plugin deleted event: %w", err)
	}

	h.logger.Infof("Processing plugin deleted: %s by developer %s", 
		eventData.Name, eventData.DeveloperID)

	// 更新分类统计（减少计数）
	if err := h.updateCategoryStats(ctx, eventData.CategoryID, -1); err != nil {
		h.logger.Errorf("Failed to update category stats: %v", err)
	}

	// 更新开发者统计
	if err := h.updateDeveloperStats(ctx, eventData.DeveloperID, "plugin_deleted"); err != nil {
		h.logger.Errorf("Failed to update developer stats: %v", err)
	}

	// 这里可以添加其他业务逻辑
	// 例如：清理相关数据、发送通知等

	return nil
}

// handlePluginPublished 处理插件发布事件
func (h *PluginEventHandler) handlePluginPublished(ctx context.Context, event *events.Event) error {
	var eventData events.PluginCreatedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal plugin published event: %w", err)
	}

	h.logger.Infof("Processing plugin published: %s by developer %s", 
		eventData.Name, eventData.DeveloperID)

	// 更新开发者统计
	if err := h.updateDeveloperStats(ctx, eventData.DeveloperID, "plugin_published"); err != nil {
		h.logger.Errorf("Failed to update developer stats: %v", err)
	}

	// 这里可以添加其他业务逻辑
	// 例如：发送发布通知、更新搜索索引等

	return nil
}

// handlePluginUnpublished 处理插件下架事件
func (h *PluginEventHandler) handlePluginUnpublished(ctx context.Context, event *events.Event) error {
	var eventData events.PluginCreatedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal plugin unpublished event: %w", err)
	}

	h.logger.Infof("Processing plugin unpublished: %s by developer %s", 
		eventData.Name, eventData.DeveloperID)

	// 更新开发者统计
	if err := h.updateDeveloperStats(ctx, eventData.DeveloperID, "plugin_unpublished"); err != nil {
		h.logger.Errorf("Failed to update developer stats: %v", err)
	}

	// 这里可以添加其他业务逻辑
	// 例如：发送下架通知、更新搜索索引等

	return nil
}

// updateCategoryStats 更新分类统计
func (h *PluginEventHandler) updateCategoryStats(ctx context.Context, categoryID string, delta int) error {
	// 这里应该调用分类服务来更新统计
	// 由于当前分类服务可能没有这个方法，我们先记录日志
	h.logger.Infof("Should update category %s stats by %d", categoryID, delta)
	
	// TODO: 实现分类统计更新逻辑
	// return h.categoryService.UpdateStats(ctx, categoryID, delta)
	
	return nil
}

// updateDeveloperStats 更新开发者统计
func (h *PluginEventHandler) updateDeveloperStats(ctx context.Context, developerID string, action string) error {
	// 这里应该调用开发者服务来更新统计
	// 由于当前开发者服务可能没有这个方法，我们先记录日志
	h.logger.Infof("Should update developer %s stats for action: %s", developerID, action)
	
	// TODO: 实现开发者统计更新逻辑
	// return h.developerService.UpdateStats(ctx, developerID, action)
	
	return nil
}

// unmarshalEventData 解析事件数据
func (h *PluginEventHandler) unmarshalEventData(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}