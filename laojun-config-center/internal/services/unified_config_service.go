package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/codetaoist/laojun-shared/config"
)

// UnifiedConfigService 统一配置服务
// 基于laojun-shared/config包提供的统一配置管理接口
type UnifiedConfigService struct {
	manager   config.ConfigManager
	validator config.ConfigValidator
	logger    *zap.Logger
	
	// 服务配置
	authEnabled  bool
	auditEnabled bool
}

// UnifiedConfigServiceOptions 统一配置服务选项
type UnifiedConfigServiceOptions struct {
	Manager      config.ConfigManager
	Validator    config.ConfigValidator
	AuthEnabled  bool
	AuditEnabled bool
}

// NewUnifiedConfigService 创建统一配置服务
func NewUnifiedConfigService(logger *zap.Logger, opts *UnifiedConfigServiceOptions) *UnifiedConfigService {
	if opts == nil {
		opts = &UnifiedConfigServiceOptions{
			AuthEnabled:  false,
			AuditEnabled: true,
		}
	}

	service := &UnifiedConfigService{
		manager:      opts.Manager,
		validator:    opts.Validator,
		logger:       logger,
		authEnabled:  opts.AuthEnabled,
		auditEnabled: opts.AuditEnabled,
	}

	// 如果没有提供配置管理器，创建默认的
	if service.manager == nil {
		factory := config.NewConfigFactory(&config.FactoryConfig{
			DefaultManagerType: config.ConfigManagerTypeDefault,
			DefaultStorageType: config.ConfigStorageTypeMemory,
		})
		
		var err error
		service.manager, err = factory.CreateManagerWithDefaults()
		if err != nil {
			logger.Error("Failed to create default config manager", zap.Error(err))
			return nil
		}
	}

	// 如果没有提供验证器，创建默认的
	if service.validator == nil {
		service.validator = config.NewDefaultConfigValidator()
	}

	return service
}

// GetConfig 获取配置
func (s *UnifiedConfigService) GetConfig(ctx context.Context, key string, operator string) (string, error) {
	// 权限检查
	if err := s.checkPermission(ctx, operator, "read", key); err != nil {
		return "", err
	}

	// 获取配置
	value, err := s.manager.Get(ctx, key)
	if err != nil {
		s.logger.Error("Failed to get config", 
			zap.String("key", key),
			zap.String("operator", operator),
			zap.Error(err))
		return "", err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "get", key, nil, value, operator)
	}

	return value, nil
}

// GetConfigWithType 获取指定类型的配置
func (s *UnifiedConfigService) GetConfigWithType(ctx context.Context, key string, configType config.ConfigType, operator string) (interface{}, error) {
	// 权限检查
	if err := s.checkPermission(ctx, operator, "read", key); err != nil {
		return nil, err
	}

	var result interface{}
	var err error

	switch configType {
	case config.ConfigTypeString:
		result, err = s.manager.Get(ctx, key)
	case config.ConfigTypeInt:
		result, err = s.manager.GetInt(ctx, key)
	case config.ConfigTypeBool:
		result, err = s.manager.GetBool(ctx, key)
	case config.ConfigTypeFloat:
		result, err = s.manager.GetFloat(ctx, key)
	case config.ConfigTypeJSON:
		var jsonValue interface{}
		jsonStr, getErr := s.manager.Get(ctx, key)
		if getErr != nil {
			err = getErr
		} else {
			err = json.Unmarshal([]byte(jsonStr), &jsonValue)
			if err == nil {
				result = jsonValue
			}
		}
	default:
		result, err = s.manager.Get(ctx, key)
	}

	if err != nil {
		s.logger.Error("Failed to get config with type", 
			zap.String("key", key),
			zap.String("type", string(configType)),
			zap.String("operator", operator),
			zap.Error(err))
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "get", key, nil, result, operator)
	}

	return result, nil
}

// SetConfig 设置配置
func (s *UnifiedConfigService) SetConfig(ctx context.Context, key, value string, ttl time.Duration, operator string) error {
	// 权限检查
	if err := s.checkPermission(ctx, operator, "write", key); err != nil {
		return err
	}

	// 验证配置
	if err := s.validator.Validate(ctx, key, value); err != nil {
		s.logger.Warn("Config validation failed", 
			zap.String("key", key),
			zap.String("value", value),
			zap.String("operator", operator),
			zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// 获取旧值用于审计
	var oldValue interface{}
	if s.auditEnabled {
		oldValue, _ = s.manager.Get(ctx, key)
	}

	// 设置配置
	err := s.manager.SetWithTTL(ctx, key, value, ttl)
	if err != nil {
		s.logger.Error("Failed to set config", 
			zap.String("key", key),
			zap.String("value", value),
			zap.Duration("ttl", ttl),
			zap.String("operator", operator),
			zap.Error(err))
		return err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "set", key, oldValue, value, operator)
	}

	s.logger.Info("Config set successfully", 
		zap.String("key", key),
		zap.Duration("ttl", ttl),
		zap.String("operator", operator))

	return nil
}

// DeleteConfig 删除配置
func (s *UnifiedConfigService) DeleteConfig(ctx context.Context, key string, operator string) error {
	// 权限检查
	if err := s.checkPermission(ctx, operator, "delete", key); err != nil {
		return err
	}

	// 获取旧值用于审计
	var oldValue interface{}
	if s.auditEnabled {
		oldValue, _ = s.manager.Get(ctx, key)
	}

	// 删除配置
	err := s.manager.Delete(ctx, key)
	if err != nil {
		s.logger.Error("Failed to delete config", 
			zap.String("key", key),
			zap.String("operator", operator),
			zap.Error(err))
		return err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "delete", key, oldValue, nil, operator)
	}

	s.logger.Info("Config deleted successfully", 
		zap.String("key", key),
		zap.String("operator", operator))

	return nil
}

// ListConfigs 列出配置
func (s *UnifiedConfigService) ListConfigs(ctx context.Context, prefix string, operator string) ([]string, error) {
	// 权限检查
	if err := s.checkPermission(ctx, operator, "list", prefix); err != nil {
		return nil, err
	}

	// 列出配置
	keys, err := s.manager.List(ctx, prefix)
	if err != nil {
		s.logger.Error("Failed to list configs", 
			zap.String("prefix", prefix),
			zap.String("operator", operator),
			zap.Error(err))
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "list", prefix, nil, len(keys), operator)
	}

	return keys, nil
}

// BatchSetConfigs 批量设置配置
func (s *UnifiedConfigService) BatchSetConfigs(ctx context.Context, configs map[string]string, ttl time.Duration, operator string) error {
	// 权限检查
	for key := range configs {
		if err := s.checkPermission(ctx, operator, "write", key); err != nil {
			return err
		}
	}

	// 批量验证配置
	if err := s.validator.ValidateMultiple(ctx, configs); err != nil {
		s.logger.Warn("Batch config validation failed", 
			zap.Int("count", len(configs)),
			zap.String("operator", operator),
			zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// 批量设置配置
	err := s.manager.SetMultiple(ctx, configs, ttl)
	if err != nil {
		s.logger.Error("Failed to batch set configs", 
			zap.Int("count", len(configs)),
			zap.Duration("ttl", ttl),
			zap.String("operator", operator),
			zap.Error(err))
		return err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "batch_set", "multiple", nil, len(configs), operator)
	}

	s.logger.Info("Batch configs set successfully", 
		zap.Int("count", len(configs)),
		zap.Duration("ttl", ttl),
		zap.String("operator", operator))

	return nil
}

// GetHistory 获取配置历史
func (s *UnifiedConfigService) GetHistory(ctx context.Context, key string, limit int, operator string) ([]*config.ConfigHistory, error) {
	// 权限检查
	if err := s.checkPermission(ctx, operator, "read", key); err != nil {
		return nil, err
	}

	// 获取历史记录
	history, err := s.manager.GetHistory(ctx, key, limit)
	if err != nil {
		s.logger.Error("Failed to get config history", 
			zap.String("key", key),
			zap.Int("limit", limit),
			zap.String("operator", operator),
			zap.Error(err))
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "get_history", key, nil, len(history), operator)
	}

	return history, nil
}

// Watch 监听配置变化
func (s *UnifiedConfigService) Watch(ctx context.Context, keys []string, operator string) (<-chan *config.ConfigChangeEvent, error) {
	// 权限检查
	for _, key := range keys {
		if err := s.checkPermission(ctx, operator, "watch", key); err != nil {
			return nil, err
		}
	}

	// 创建监听器
	watcher := config.NewDefaultConfigWatcher()
	
	// 添加监听键
	for _, key := range keys {
		if err := watcher.AddKey(key); err != nil {
			s.logger.Error("Failed to add watch key", 
				zap.String("key", key),
				zap.String("operator", operator),
				zap.Error(err))
			return nil, err
		}
	}

	// 启动监听
	eventChan := make(chan *config.ConfigChangeEvent, 100)
	if err := watcher.Start(ctx, s.manager, func(event *config.ConfigChangeEvent) {
		select {
		case eventChan <- event:
		case <-ctx.Done():
			return
		}
	}); err != nil {
		s.logger.Error("Failed to start config watcher", 
			zap.Strings("keys", keys),
			zap.String("operator", operator),
			zap.Error(err))
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "watch", strings.Join(keys, ","), nil, len(keys), operator)
	}

	s.logger.Info("Config watch started", 
		zap.Strings("keys", keys),
		zap.String("operator", operator))

	return eventChan, nil
}

// HealthCheck 健康检查
func (s *UnifiedConfigService) HealthCheck(ctx context.Context) error {
	health, err := s.manager.Health(ctx)
	if err != nil {
		return err
	}

	if health.Status != config.HealthStatusHealthy {
		return fmt.Errorf("config manager is not healthy: %s", health.Message)
	}

	return nil
}

// AddValidationRule 添加验证规则
func (s *UnifiedConfigService) AddValidationRule(ctx context.Context, key string, rule *config.ValidationRule, operator string) error {
	// 权限检查
	if err := s.checkPermission(ctx, operator, "admin", key); err != nil {
		return err
	}

	// 添加验证规则
	err := s.validator.AddRule(ctx, key, rule)
	if err != nil {
		s.logger.Error("Failed to add validation rule", 
			zap.String("key", key),
			zap.String("rule", rule.Name),
			zap.String("operator", operator),
			zap.Error(err))
		return err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "add_rule", key, nil, rule.Name, operator)
	}

	s.logger.Info("Validation rule added", 
		zap.String("key", key),
		zap.String("rule", rule.Name),
		zap.String("operator", operator))

	return nil
}

// Close 关闭服务
func (s *UnifiedConfigService) Close() error {
	if s.manager != nil {
		return s.manager.Close()
	}
	return nil
}

// checkPermission 检查权限
func (s *UnifiedConfigService) checkPermission(ctx context.Context, operator, action, resource string) error {
	if !s.authEnabled {
		return nil
	}

	// 这里可以实现具体的权限检查逻辑
	// 例如，检查JWT token、RBAC权限等
	
	s.logger.Debug("Permission check", 
		zap.String("operator", operator),
		zap.String("action", action),
		zap.String("resource", resource))

	return nil
}

// auditLog 审计日志
func (s *UnifiedConfigService) auditLog(ctx context.Context, action, key string, oldValue, newValue interface{}, operator string) {
	if !s.auditEnabled {
		return
	}

	s.logger.Info("Config audit log", 
		zap.String("action", action),
		zap.String("key", key),
		zap.Any("old_value", oldValue),
		zap.Any("new_value", newValue),
		zap.String("operator", operator),
		zap.Time("timestamp", time.Now()))
}