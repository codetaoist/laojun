package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codetaoist/laojun-marketplace-api/internal/events"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/sirupsen/logrus"
)

// PaymentEventHandler 支付事件处理器
type PaymentEventHandler struct {
	paymentService *services.PaymentService
	pluginService  *services.PluginService
	logger         *logrus.Logger
}

// NewPaymentEventHandler 创建支付事件处理器
func NewPaymentEventHandler(
	paymentService *services.PaymentService,
	pluginService *services.PluginService,
	logger *logrus.Logger,
) *PaymentEventHandler {
	return &PaymentEventHandler{
		paymentService: paymentService,
		pluginService:  pluginService,
		logger:         logger,
	}
}

// Handle 处理事件
func (h *PaymentEventHandler) Handle(ctx context.Context, event *events.Event) error {
	h.logger.Infof("Handling payment event: %s (ID: %s)", event.Type, event.ID)

	switch event.Type {
	case events.PaymentOrderCreated:
		return h.handlePaymentOrderCreated(ctx, event)
	case events.PaymentOrderCompleted:
		return h.handlePaymentOrderCompleted(ctx, event)
	case events.PaymentOrderCancelled:
		return h.handlePaymentOrderCancelled(ctx, event)
	case events.PaymentOrderRefunded:
		return h.handlePaymentOrderRefunded(ctx, event)
	default:
		h.logger.Warnf("Unknown payment event type: %s", event.Type)
		return nil
	}
}

// GetEventTypes 获取处理的事件类型
func (h *PaymentEventHandler) GetEventTypes() []events.EventType {
	return []events.EventType{
		events.PaymentOrderCreated,
		events.PaymentOrderCompleted,
		events.PaymentOrderCancelled,
		events.PaymentOrderRefunded,
	}
}

// handlePaymentOrderCreated 处理支付订单创建事件
func (h *PaymentEventHandler) handlePaymentOrderCreated(ctx context.Context, event *events.Event) error {
	var eventData events.PaymentOrderCreatedEvent
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal payment order created event: %w", err)
	}

	h.logger.Infof("Processing payment order created: %s for user %s", 
		eventData.OrderNumber, eventData.UserID)

	// 这里可以添加订单创建后的业务逻辑
	// 例如：发送通知、更新统计数据等

	return nil
}

// handlePaymentOrderCompleted 处理支付订单完成事件
func (h *PaymentEventHandler) handlePaymentOrderCompleted(ctx context.Context, event *events.Event) error {
	var eventData events.PaymentOrderCompletedEvent
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal payment order completed event: %w", err)
	}

	h.logger.Infof("Processing payment order completed: %s for user %s", 
		eventData.OrderNumber, eventData.UserID)

	// 更新插件销售统计
	if err := h.updatePluginSalesStats(ctx, eventData.PluginID, eventData.Amount); err != nil {
		h.logger.Errorf("Failed to update plugin sales stats: %v", err)
		// 不返回错误，避免影响其他处理逻辑
	}

	// 这里可以添加其他业务逻辑
	// 例如：发送购买成功通知、更新用户权限等

	return nil
}

// handlePaymentOrderCancelled 处理支付订单取消事件
func (h *PaymentEventHandler) handlePaymentOrderCancelled(ctx context.Context, event *events.Event) error {
	var eventData events.PaymentOrderCreatedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal payment order cancelled event: %w", err)
	}

	h.logger.Infof("Processing payment order cancelled: %s for user %s", 
		eventData.OrderNumber, eventData.UserID)

	// 这里可以添加订单取消后的业务逻辑
	// 例如：发送取消通知、释放库存等

	return nil
}

// handlePaymentOrderRefunded 处理支付订单退款事件
func (h *PaymentEventHandler) handlePaymentOrderRefunded(ctx context.Context, event *events.Event) error {
	var eventData events.PaymentOrderCompletedEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal payment order refunded event: %w", err)
	}

	h.logger.Infof("Processing payment order refunded: %s for user %s", 
		eventData.OrderNumber, eventData.UserID)

	// 更新插件销售统计（减少销售额）
	if err := h.updatePluginSalesStats(ctx, eventData.PluginID, -eventData.Amount); err != nil {
		h.logger.Errorf("Failed to update plugin sales stats for refund: %v", err)
		// 不返回错误，避免影响其他处理逻辑
	}

	// 这里可以添加其他业务逻辑
	// 例如：发送退款通知、撤销用户权限等

	return nil
}

// updatePluginSalesStats 更新插件销售统计
func (h *PaymentEventHandler) updatePluginSalesStats(ctx context.Context, pluginID string, amount float64) error {
	// 这里应该调用插件服务来更新销售统计
	// 由于当前插件服务可能没有这个方法，我们先记录日志
	h.logger.Infof("Should update plugin %s sales stats by %f", pluginID, amount)
	
	// TODO: 实现插件销售统计更新逻辑
	// return h.pluginService.UpdateSalesStats(ctx, pluginID, amount)
	
	return nil
}

// unmarshalEventData 解析事件数据
func (h *PaymentEventHandler) unmarshalEventData(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}