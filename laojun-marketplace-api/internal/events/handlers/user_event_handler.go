package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codetaoist/laojun-marketplace-api/internal/events"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/sirupsen/logrus"
)

// UserEventHandler 用户事件处理器
type UserEventHandler struct {
	userService      *services.UserService
	developerService *services.DeveloperService
	logger           *logrus.Logger
}

// NewUserEventHandler 创建用户事件处理器
func NewUserEventHandler(
	userService *services.UserService,
	developerService *services.DeveloperService,
	logger *logrus.Logger,
) *UserEventHandler {
	return &UserEventHandler{
		userService:      userService,
		developerService: developerService,
		logger:           logger,
	}
}

// Handle 处理事件
func (h *UserEventHandler) Handle(ctx context.Context, event *events.Event) error {
	h.logger.Infof("Handling user event: %s (ID: %s)", event.Type, event.ID)

	switch event.Type {
	case events.UserRegistered:
		return h.handleUserRegistered(ctx, event)
	case events.UserUpdated:
		return h.handleUserUpdated(ctx, event)
	case events.UserDeleted:
		return h.handleUserDeleted(ctx, event)
	case events.UserActivated:
		return h.handleUserActivated(ctx, event)
	case events.UserDeactivated:
		return h.handleUserDeactivated(ctx, event)
	case events.DeveloperRegistered:
		return h.handleDeveloperRegistered(ctx, event)
	default:
		h.logger.Warnf("Unknown user event type: %s", event.Type)
		return nil
	}
}

// GetEventTypes 获取处理的事件类型
func (h *UserEventHandler) GetEventTypes() []events.EventType {
	return []events.EventType{
		events.UserRegistered,
		events.UserUpdated,
		events.UserDeleted,
		events.UserActivated,
		events.UserDeactivated,
		events.DeveloperRegistered,
	}
}

// handleUserRegistered 处理用户注册事件
func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event *events.Event) error {
	var eventData events.UserRegisteredEvent
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal user registered event: %w", err)
	}

	h.logger.Infof("Processing user registered: %s (%s)", eventData.UserID, eventData.Email)

	// 发送欢迎邮件
	if err := h.sendWelcomeEmail(ctx, eventData); err != nil {
		h.logger.Errorf("Failed to send welcome email: %v", err)
		// 不返回错误，因为这不是关键操作
	}

	// 初始化用户偏好设置
	if err := h.initializeUserPreferences(ctx, eventData.UserID); err != nil {
		h.logger.Errorf("Failed to initialize user preferences: %v", err)
	}

	// 更新用户统计
	if err := h.updateUserStats(ctx, "register", 1); err != nil {
		h.logger.Errorf("Failed to update user stats: %v", err)
	}

	// 这里可以添加其他业务逻辑
	// 例如：创建用户钱包、分配新用户优惠券等

	return nil
}

// handleUserUpdated 处理用户更新事件
func (h *UserEventHandler) handleUserUpdated(ctx context.Context, event *events.Event) error {
	var eventData events.UserRegisteredEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal user updated event: %w", err)
	}

	h.logger.Infof("Processing user updated: %s (%s)", eventData.UserID, eventData.Email)

	// 清除用户相关缓存
	if err := h.clearUserCache(ctx, eventData.UserID); err != nil {
		h.logger.Errorf("Failed to clear user cache: %v", err)
	}

	// 同步用户信息到其他系统
	if err := h.syncUserToExternalSystems(ctx, eventData); err != nil {
		h.logger.Errorf("Failed to sync user to external systems: %v", err)
	}

	return nil
}

// handleUserDeleted 处理用户删除事件
func (h *UserEventHandler) handleUserDeleted(ctx context.Context, event *events.Event) error {
	var eventData events.UserRegisteredEvent // 使用相同的数据结构
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal user deleted event: %w", err)
	}

	h.logger.Infof("Processing user deleted: %s (%s)", eventData.UserID, eventData.Email)

	// 清理用户相关数据
	if err := h.cleanupUserData(ctx, eventData.UserID); err != nil {
		h.logger.Errorf("Failed to cleanup user data: %v", err)
	}

	// 更新用户统计
	if err := h.updateUserStats(ctx, "delete", 1); err != nil {
		h.logger.Errorf("Failed to update user stats: %v", err)
	}

	// 从外部系统删除用户
	if err := h.removeUserFromExternalSystems(ctx, eventData); err != nil {
		h.logger.Errorf("Failed to remove user from external systems: %v", err)
	}

	return nil
}

// handleUserActivated 处理用户激活事件
func (h *UserEventHandler) handleUserActivated(ctx context.Context, event *events.Event) error {
	h.logger.Infof("Processing user activated event: %s", event.ID)

	// 这里可以处理用户激活后的逻辑
	// 例如：发送激活成功通知、解锁功能等

	return nil
}

// handleUserDeactivated 处理用户停用事件
func (h *UserEventHandler) handleUserDeactivated(ctx context.Context, event *events.Event) error {
	h.logger.Infof("Processing user deactivated event: %s", event.ID)

	// 这里可以处理用户停用后的逻辑
	// 例如：撤销权限、清理会话等

	return nil
}

// handleDeveloperRegistered 处理开发者注册事件
func (h *UserEventHandler) handleDeveloperRegistered(ctx context.Context, event *events.Event) error {
	var eventData events.DeveloperRegisteredEvent
	if err := h.unmarshalEventData(event.Data, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal developer registered event: %w", err)
	}

	h.logger.Infof("Processing developer registered: %s for user %s", 
		eventData.DeveloperID, eventData.UserID)

	// 发送开发者欢迎邮件
	if err := h.sendDeveloperWelcomeEmail(ctx, eventData); err != nil {
		h.logger.Errorf("Failed to send developer welcome email: %v", err)
	}

	// 初始化开发者工作区
	if err := h.initializeDeveloperWorkspace(ctx, eventData.DeveloperID); err != nil {
		h.logger.Errorf("Failed to initialize developer workspace: %v", err)
	}

	// 更新开发者统计
	if err := h.updateDeveloperStats(ctx, "register", 1); err != nil {
		h.logger.Errorf("Failed to update developer stats: %v", err)
	}

	return nil
}

// sendWelcomeEmail 发送欢迎邮件
func (h *UserEventHandler) sendWelcomeEmail(ctx context.Context, eventData events.UserRegisteredEvent) error {
	h.logger.Infof("Should send welcome email to %s", eventData.Email)
	
	// TODO: 实现邮件发送逻辑
	// 这里可以调用邮件服务或消息队列
	
	return nil
}

// sendDeveloperWelcomeEmail 发送开发者欢迎邮件
func (h *UserEventHandler) sendDeveloperWelcomeEmail(ctx context.Context, eventData events.DeveloperRegisteredEvent) error {
	h.logger.Infof("Should send developer welcome email for developer %s", eventData.DeveloperID)
	
	// TODO: 实现开发者欢迎邮件发送逻辑
	
	return nil
}

// initializeUserPreferences 初始化用户偏好设置
func (h *UserEventHandler) initializeUserPreferences(ctx context.Context, userID string) error {
	h.logger.Infof("Should initialize preferences for user %s", userID)
	
	// TODO: 实现用户偏好设置初始化
	// 例如：语言偏好、通知设置等
	
	return nil
}

// initializeDeveloperWorkspace 初始化开发者工作区
func (h *UserEventHandler) initializeDeveloperWorkspace(ctx context.Context, developerID string) error {
	h.logger.Infof("Should initialize workspace for developer %s", developerID)
	
	// TODO: 实现开发者工作区初始化
	// 例如：创建默认项目、设置开发环境等
	
	return nil
}

// clearUserCache 清除用户缓存
func (h *UserEventHandler) clearUserCache(ctx context.Context, userID string) error {
	h.logger.Infof("Should clear cache for user %s", userID)
	
	// TODO: 实现缓存清理逻辑
	
	return nil
}

// cleanupUserData 清理用户数据
func (h *UserEventHandler) cleanupUserData(ctx context.Context, userID string) error {
	h.logger.Infof("Should cleanup data for user %s", userID)
	
	// TODO: 实现用户数据清理逻辑
	// 例如：删除用户文件、清理关联数据等
	
	return nil
}

// syncUserToExternalSystems 同步用户到外部系统
func (h *UserEventHandler) syncUserToExternalSystems(ctx context.Context, eventData events.UserRegisteredEvent) error {
	h.logger.Infof("Should sync user %s to external systems", eventData.UserID)
	
	// TODO: 实现外部系统同步逻辑
	// 例如：同步到CRM、分析系统等
	
	return nil
}

// removeUserFromExternalSystems 从外部系统删除用户
func (h *UserEventHandler) removeUserFromExternalSystems(ctx context.Context, eventData events.UserRegisteredEvent) error {
	h.logger.Infof("Should remove user %s from external systems", eventData.UserID)
	
	// TODO: 实现外部系统删除逻辑
	
	return nil
}

// updateUserStats 更新用户统计
func (h *UserEventHandler) updateUserStats(ctx context.Context, action string, delta int) error {
	h.logger.Infof("Should update user stats: action=%s, delta=%d", action, delta)
	
	// TODO: 实现用户统计更新逻辑
	
	return nil
}

// updateDeveloperStats 更新开发者统计
func (h *UserEventHandler) updateDeveloperStats(ctx context.Context, action string, delta int) error {
	h.logger.Infof("Should update developer stats: action=%s, delta=%d", action, delta)
	
	// TODO: 实现开发者统计更新逻辑
	
	return nil
}

// unmarshalEventData 解析事件数据
func (h *UserEventHandler) unmarshalEventData(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}