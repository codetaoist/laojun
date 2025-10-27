package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codetaoist/laojun-marketplace-api/internal/events"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/sirupsen/logrus"
)

// ReviewEventHandler 评价事件处理器
type ReviewEventHandler struct {
	reviewService           *services.ReviewService
	reviewModerationService *services.ReviewModerationService
	pluginService           *services.PluginService
	logger                  *logrus.Logger
}

// NewReviewEventHandler 创建评价事件处理器
func NewReviewEventHandler(
	reviewService *services.ReviewService,
	reviewModerationService *services.ReviewModerationService,
	pluginService *services.PluginService,
	logger *logrus.Logger,
) *ReviewEventHandler {
	return &ReviewEventHandler{
		reviewService:           reviewService,
		reviewModerationService: reviewModerationService,
		pluginService:           pluginService,
		logger:                  logger,
	}
}

// Handle 处理事件
func (h *ReviewEventHandler) Handle(ctx context.Context, event *events.Event) error {
	h.logger.Infof("Handling review event: %s (ID: %s)", event.Type, event.ID)

	switch event.Type {
	case events.ReviewCreated:
		return h.handleReviewCreated(ctx, event)
	case events.ReviewUpdated:
		return h.handleReviewUpdated(ctx, event)
	case events.ReviewDeleted:
		return h.handleReviewDeleted(ctx, event)
	case events.ReviewModerated:
		return h.handleReviewModerated(ctx, event)
	case events.ReviewFlagged:
		return h.handleReviewFlagged(ctx, event)
	default:
		h.logger.Warnf("Unknown review event type: %s", event.Type)
		return nil
	}
}

// GetEventTypes 获取处理的事件类型
func (h *ReviewEventHandler) GetEventTypes() []events.EventType {
	return []events.EventType{
		events.ReviewCreated,
		events.ReviewUpdated,
		events.ReviewDeleted,
		events.ReviewModerated,
		events.ReviewFlagged,
	}
}

// handleReviewCreated 处理评价创建事件
func (h *ReviewEventHandler) handleReviewCreated(ctx context.Context, event *events.Event) error {
	var eventData events.ReviewCreatedEvent
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal review created event: %w", err)
	}

	h.logger.Infof("Processing review created: %s for plugin %s by user %s", 
		eventData.ReviewID, eventData.PluginID, eventData.UserID)

	// 更新插件评分统计
	if err := h.updatePluginRatingStats(ctx, eventData.PluginID, eventData.Rating, 1); err != nil {
		h.logger.Errorf("Failed to update plugin rating stats: %v", err)
	}

	// 检查是否需要自动审核
	if err := h.checkAutoModeration(ctx, eventData); err != nil {
		h.logger.Errorf("Failed to check auto moderation: %v", err)
	}

	// 这里可以添加其他业务逻辑
	// 例如：发送通知、更新搜索索引等

	return nil
}

// handleReviewUpdated 处理评价更新事件
func (h *ReviewEventHandler) handleReviewUpdated(ctx context.Context, event *events.Event) error {
	var eventData events.ReviewCreatedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal review updated event: %w", err)
	}

	h.logger.Infof("Processing review updated: %s for plugin %s by user %s", 
		eventData.ReviewID, eventData.PluginID, eventData.UserID)

	// 这里可能需要重新计算插件评分统计
	// 由于需要知道原来的评分，这里先记录日志
	h.logger.Infof("Should recalculate plugin %s rating stats after review update", eventData.PluginID)

	// 这里可以添加其他业务逻辑
	// 例如：清除缓存、重新审核等

	return nil
}

// handleReviewDeleted 处理评价删除事件
func (h *ReviewEventHandler) handleReviewDeleted(ctx context.Context, event *events.Event) error {
	var eventData events.ReviewCreatedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal review deleted event: %w", err)
	}

	h.logger.Infof("Processing review deleted: %s for plugin %s by user %s", 
		eventData.ReviewID, eventData.PluginID, eventData.UserID)

	// 更新插件评分统计（减少计数）
	if err := h.updatePluginRatingStats(ctx, eventData.PluginID, eventData.Rating, -1); err != nil {
		h.logger.Errorf("Failed to update plugin rating stats: %v", err)
	}

	// 这里可以添加其他业务逻辑
	// 例如：清理相关数据、发送通知等

	return nil
}

// handleReviewModerated 处理评价审核事件
func (h *ReviewEventHandler) handleReviewModerated(ctx context.Context, event *events.Event) error {
	// 这里可以处理评价审核后的逻辑
	h.logger.Infof("Processing review moderated event: %s", event.ID)

	// 根据审核结果执行不同的操作
	// 例如：通过审核的评价可以显示，拒绝的评价需要隐藏等

	return nil
}

// handleReviewFlagged 处理评价举报事件
func (h *ReviewEventHandler) handleReviewFlagged(ctx context.Context, event *events.Event) error {
	h.logger.Infof("Processing review flagged event: %s", event.ID)

	// 这里可以处理评价被举报后的逻辑
	// 例如：自动隐藏、发送给审核员等

	return nil
}

// updatePluginRatingStats 更新插件评分统计
func (h *ReviewEventHandler) updatePluginRatingStats(ctx context.Context, pluginID string, rating int, delta int) error {
	// 这里应该调用插件服务来更新评分统计
	// 由于当前插件服务可能没有这个方法，我们先记录日志
	h.logger.Infof("Should update plugin %s rating stats: rating=%d, delta=%d", pluginID, rating, delta)
	
	// TODO: 实现插件评分统计更新逻辑
	// return h.pluginService.UpdateRatingStats(ctx, pluginID, rating, delta)
	
	return nil
}

// checkAutoModeration 检查自动审核
func (h *ReviewEventHandler) checkAutoModeration(ctx context.Context, eventData events.ReviewCreatedEvent) error {
	// 这里可以实现自动审核逻辑
	// 例如：检查敏感词、垃圾内容等
	h.logger.Infof("Should check auto moderation for review %s", eventData.ReviewID)
	
	// 简单的示例：如果评分过低，可能需要人工审核
	if eventData.Rating <= 2 {
		h.logger.Infof("Low rating review %s may need manual moderation", eventData.ReviewID)
		// TODO: 标记为需要审核
	}
	
	// 检查内容长度
	if len(eventData.Content) < 10 {
		h.logger.Infof("Short review %s may need manual moderation", eventData.ReviewID)
		// TODO: 标记为需要审核
	}
	
	return nil
}

// unmarshalEventData 解析事件数据
func (h *ReviewEventHandler) unmarshalEventData(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}