package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/codetaoist/laojun-config-center/internal/storage"
	"github.com/codetaoist/laojun-shared/config"
)

// ConfigService 配置服务
type ConfigService struct {
	storage storage.ConfigStorage
	logger  *zap.Logger
	
	// 统一配置管理器
	configManager config.ConfigManager
	
	// 缓存配置
	cacheEnabled bool
	cacheTTL     time.Duration
	
	// 权限配置
	authEnabled bool
	
	// 审计配置
	auditEnabled bool
}

// ConfigServiceOptions 配置服务选项
type ConfigServiceOptions struct {
	CacheEnabled  bool
	CacheTTL      time.Duration
	AuthEnabled   bool
	AuditEnabled  bool
	ConfigManager config.ConfigManager
}

// NewConfigService 创建配置服务
func NewConfigService(storage storage.ConfigStorage, logger *zap.Logger, opts *ConfigServiceOptions) *ConfigService {
	if opts == nil {
		opts = &ConfigServiceOptions{
			CacheEnabled: false,
			CacheTTL:     5 * time.Minute,
			AuthEnabled:  false,
			AuditEnabled: true,
		}
	}

	service := &ConfigService{
		storage:       storage,
		logger:        logger,
		configManager: opts.ConfigManager,
		cacheEnabled:  opts.CacheEnabled,
		cacheTTL:      opts.CacheTTL,
		authEnabled:   opts.AuthEnabled,
		auditEnabled:  opts.AuditEnabled,
	}
	
	// 如果没有提供统一配置管理器，创建一个默认的
	if service.configManager == nil {
		// 使用内存存储创建默认配置管理器
		memoryStorage := config.NewMemoryConfigStorage(&config.ConfigOptions{
			MaxHistorySize: 100,
		})
		service.configManager = config.NewDefaultConfigManager(memoryStorage, nil, nil, &config.ConfigOptions{
			EnableCache:    opts.CacheEnabled,
			CacheTTL:       opts.CacheTTL,
			MaxHistorySize: 100,
		})
	}

	return service
}

// GetConfig 获取配置
func (s *ConfigService) GetConfig(ctx context.Context, service, environment, key string, operator string) (*storage.ConfigItem, error) {
	// 参数验证
	if err := s.validateConfigKey(service, environment, key); err != nil {
		return nil, err
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "read", service, environment, key); err != nil {
			return nil, err
		}
	}

	// 获取配置
	item, err := s.storage.Get(ctx, service, environment, key)
	if err != nil {
		s.logger.Error("failed to get config",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "get", service, environment, key, nil, item.Value, operator)
	}

	return item, nil
}

// SetConfig 设置配置
func (s *ConfigService) SetConfig(ctx context.Context, item *storage.ConfigItem, operator string) error {
	// 参数验证
	if err := s.validateConfigItem(item); err != nil {
		return err
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "write", item.Service, item.Environment, item.Key); err != nil {
			return err
		}
	}

	// 获取旧值用于审计
	var oldValue interface{}
	if s.auditEnabled {
		if oldItem, err := s.storage.Get(ctx, item.Service, item.Environment, item.Key); err == nil {
			oldValue = oldItem.Value
		}
	}

	// 设置操作者
	item.UpdatedBy = operator

	// 保存配置
	if err := s.storage.Set(ctx, item); err != nil {
		s.logger.Error("failed to set config",
			zap.String("service", item.Service),
			zap.String("environment", item.Environment),
			zap.String("key", item.Key),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "set", item.Service, item.Environment, item.Key, oldValue, item.Value, operator)
	}

	s.logger.Info("config set successfully",
		zap.String("service", item.Service),
		zap.String("environment", item.Environment),
		zap.String("key", item.Key),
		zap.String("operator", operator),
		zap.Int64("version", item.Version),
	)

	return nil
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(ctx context.Context, service, environment, key string, operator string) error {
	// 参数验证
	if err := s.validateConfigKey(service, environment, key); err != nil {
		return err
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "delete", service, environment, key); err != nil {
			return err
		}
	}

	// 获取旧值用于审计
	var oldValue interface{}
	if s.auditEnabled {
		if oldItem, err := s.storage.Get(ctx, service, environment, key); err == nil {
			oldValue = oldItem.Value
		}
	}

	// 删除配置
	if err := s.storage.Delete(ctx, service, environment, key); err != nil {
		s.logger.Error("failed to delete config",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "delete", service, environment, key, oldValue, nil, operator)
	}

	s.logger.Info("config deleted successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("key", key),
		zap.String("operator", operator),
	)

	return nil
}

// ListConfigs 列出配置
func (s *ConfigService) ListConfigs(ctx context.Context, service, environment string, operator string) ([]*storage.ConfigItem, error) {
	// 参数验证
	if service == "" || environment == "" {
		return nil, &storage.ValidationError{
			Field:   "service/environment",
			Message: "service and environment are required",
		}
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "list", service, environment, "*"); err != nil {
			return nil, err
		}
	}

	// 获取配置列表
	items, err := s.storage.List(ctx, service, environment)
	if err != nil {
		s.logger.Error("failed to list configs",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "list", service, environment, "*", nil, len(items), operator)
	}

	return items, nil
}

// SearchConfigs 搜索配置
func (s *ConfigService) SearchConfigs(ctx context.Context, query *storage.SearchQuery, operator string) ([]*storage.ConfigItem, error) {
	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "search", query.Service, query.Environment, query.Key); err != nil {
			return nil, err
		}
	}

	// 搜索配置
	items, err := s.storage.Search(ctx, query)
	if err != nil {
		s.logger.Error("failed to search configs",
			zap.String("operator", operator),
			zap.Any("query", query),
			zap.Error(err),
		)
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "search", query.Service, query.Environment, query.Key, query, len(items), operator)
	}

	return items, nil
}

// GetHistory 获取配置历史
func (s *ConfigService) GetHistory(ctx context.Context, service, environment, key string, limit int, operator string) ([]*storage.ConfigHistory, error) {
	// 参数验证
	if err := s.validateConfigKey(service, environment, key); err != nil {
		return nil, err
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "history", service, environment, key); err != nil {
			return nil, err
		}
	}

	// 获取历史记录
	history, err := s.storage.GetHistory(ctx, service, environment, key, limit)
	if err != nil {
		s.logger.Error("failed to get config history",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "history", service, environment, key, nil, len(history), operator)
	}

	return history, nil
}

// GetVersion 获取指定版本的配置
func (s *ConfigService) GetVersion(ctx context.Context, service, environment, key string, version int64, operator string) (*storage.ConfigItem, error) {
	// 参数验证
	if err := s.validateConfigKey(service, environment, key); err != nil {
		return nil, err
	}

	if version <= 0 {
		return nil, &storage.ValidationError{
			Field:   "version",
			Message: "version must be greater than 0",
		}
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "read", service, environment, key); err != nil {
			return nil, err
		}
	}

	// 获取指定版本
	item, err := s.storage.GetVersion(ctx, service, environment, key, version)
	if err != nil {
		s.logger.Error("failed to get config version",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.Int64("version", version),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "get_version", service, environment, key, version, item.Value, operator)
	}

	return item, nil
}

// Rollback 回滚配置
func (s *ConfigService) Rollback(ctx context.Context, service, environment, key string, version int64, operator string) error {
	// 参数验证
	if err := s.validateConfigKey(service, environment, key); err != nil {
		return err
	}

	if version <= 0 {
		return &storage.ValidationError{
			Field:   "version",
			Message: "version must be greater than 0",
		}
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "rollback", service, environment, key); err != nil {
			return err
		}
	}

	// 获取当前值用于审计
	var oldValue interface{}
	if s.auditEnabled {
		if currentItem, err := s.storage.Get(ctx, service, environment, key); err == nil {
			oldValue = currentItem.Value
		}
	}

	// 执行回滚
	if err := s.storage.Rollback(ctx, service, environment, key, version, operator); err != nil {
		s.logger.Error("failed to rollback config",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("key", key),
			zap.Int64("version", version),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return err
	}

	// 获取回滚后的值用于审计
	var newValue interface{}
	if s.auditEnabled {
		if newItem, err := s.storage.Get(ctx, service, environment, key); err == nil {
			newValue = newItem.Value
		}
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "rollback", service, environment, key, oldValue, newValue, operator)
	}

	s.logger.Info("config rolled back successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("key", key),
		zap.Int64("version", version),
		zap.String("operator", operator),
	)

	return nil
}

// BatchSetConfigs 批量设置配置
func (s *ConfigService) BatchSetConfigs(ctx context.Context, items []*storage.ConfigItem, operator string) error {
	if len(items) == 0 {
		return &storage.ValidationError{
			Field:   "items",
			Message: "items cannot be empty",
		}
	}

	// 验证所有配置项
	for i, item := range items {
		if err := s.validateConfigItem(item); err != nil {
			return fmt.Errorf("item[%d]: %w", i, err)
		}
		
		// 权限检查
		if s.authEnabled {
			if err := s.checkPermission(ctx, operator, "write", item.Service, item.Environment, item.Key); err != nil {
				return fmt.Errorf("item[%d]: %w", i, err)
			}
		}
		
		// 设置操作者
		item.UpdatedBy = operator
	}

	// 批量设置
	if err := s.storage.SetMultiple(ctx, items); err != nil {
		s.logger.Error("failed to batch set configs",
			zap.String("operator", operator),
			zap.Int("count", len(items)),
			zap.Error(err),
		)
		return err
	}

	// 审计日志
	if s.auditEnabled {
		for _, item := range items {
			s.auditLog(ctx, "batch_set", item.Service, item.Environment, item.Key, nil, item.Value, operator)
		}
	}

	s.logger.Info("configs batch set successfully",
		zap.String("operator", operator),
		zap.Int("count", len(items)),
	)

	return nil
}

// BatchDeleteConfigs 批量删除配置
func (s *ConfigService) BatchDeleteConfigs(ctx context.Context, keys []storage.ConfigKey, operator string) error {
	if len(keys) == 0 {
		return &storage.ValidationError{
			Field:   "keys",
			Message: "keys cannot be empty",
		}
	}

	// 验证和权限检查
	for i, key := range keys {
		if err := s.validateConfigKey(key.Service, key.Environment, key.Key); err != nil {
			return fmt.Errorf("key[%d]: %w", i, err)
		}
		
		if s.authEnabled {
			if err := s.checkPermission(ctx, operator, "delete", key.Service, key.Environment, key.Key); err != nil {
				return fmt.Errorf("key[%d]: %w", i, err)
			}
		}
	}

	// 批量删除
	if err := s.storage.DeleteMultiple(ctx, keys); err != nil {
		s.logger.Error("failed to batch delete configs",
			zap.String("operator", operator),
			zap.Int("count", len(keys)),
			zap.Error(err),
		)
		return err
	}

	// 审计日志
	if s.auditEnabled {
		for _, key := range keys {
			s.auditLog(ctx, "batch_delete", key.Service, key.Environment, key.Key, nil, nil, operator)
		}
	}

	s.logger.Info("configs batch deleted successfully",
		zap.String("operator", operator),
		zap.Int("count", len(keys)),
	)

	return nil
}

// Watch 监听配置变更
func (s *ConfigService) Watch(ctx context.Context, service, environment string, operator string) (<-chan *storage.WatchEvent, error) {
	// 参数验证
	if service == "" || environment == "" {
		return nil, &storage.ValidationError{
			Field:   "service/environment",
			Message: "service and environment are required",
		}
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "watch", service, environment, "*"); err != nil {
			return nil, err
		}
	}

	// 开始监听
	eventChan, err := s.storage.Watch(ctx, service, environment)
	if err != nil {
		s.logger.Error("failed to watch configs",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "watch", service, environment, "*", nil, nil, operator)
	}

	s.logger.Info("started watching configs",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("operator", operator),
	)

	return eventChan, nil
}

// Backup 备份配置
func (s *ConfigService) Backup(ctx context.Context, service, environment string, operator string) ([]byte, error) {
	// 参数验证
	if service == "" || environment == "" {
		return nil, &storage.ValidationError{
			Field:   "service/environment",
			Message: "service and environment are required",
		}
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "backup", service, environment, "*"); err != nil {
			return nil, err
		}
	}

	// 执行备份
	data, err := s.storage.Backup(ctx, service, environment)
	if err != nil {
		s.logger.Error("failed to backup configs",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return nil, err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "backup", service, environment, "*", nil, len(data), operator)
	}

	s.logger.Info("configs backed up successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("operator", operator),
		zap.Int("size", len(data)),
	)

	return data, nil
}

// Restore 恢复配置
func (s *ConfigService) Restore(ctx context.Context, service, environment string, data []byte, operator string) error {
	// 参数验证
	if service == "" || environment == "" {
		return &storage.ValidationError{
			Field:   "service/environment",
			Message: "service and environment are required",
		}
	}

	if len(data) == 0 {
		return &storage.ValidationError{
			Field:   "data",
			Message: "backup data cannot be empty",
		}
	}

	// 权限检查
	if s.authEnabled {
		if err := s.checkPermission(ctx, operator, "restore", service, environment, "*"); err != nil {
			return err
		}
	}

	// 执行恢复
	if err := s.storage.Restore(ctx, service, environment, data, operator); err != nil {
		s.logger.Error("failed to restore configs",
			zap.String("service", service),
			zap.String("environment", environment),
			zap.String("operator", operator),
			zap.Error(err),
		)
		return err
	}

	// 审计日志
	if s.auditEnabled {
		s.auditLog(ctx, "restore", service, environment, "*", nil, len(data), operator)
	}

	s.logger.Info("configs restored successfully",
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("operator", operator),
		zap.Int("size", len(data)),
	)

	return nil
}

// HealthCheck 健康检查
func (s *ConfigService) HealthCheck(ctx context.Context) error {
	return s.storage.HealthCheck(ctx)
}

// 验证方法

// validateConfigKey 验证配置键
func (s *ConfigService) validateConfigKey(service, environment, key string) error {
	if service == "" {
		return &storage.ValidationError{
			Field:   "service",
			Message: "service is required",
		}
	}
	if environment == "" {
		return &storage.ValidationError{
			Field:   "environment",
			Message: "environment is required",
		}
	}
	if key == "" {
		return &storage.ValidationError{
			Field:   "key",
			Message: "key is required",
		}
	}

	// 检查字符限制
	if strings.Contains(service, ":") || strings.Contains(environment, ":") || strings.Contains(key, ":") {
		return &storage.ValidationError{
			Field:   "service/environment/key",
			Message: "service, environment, and key cannot contain ':' character",
		}
	}

	return nil
}

// validateConfigItem 验证配置项
func (s *ConfigService) validateConfigItem(item *storage.ConfigItem) error {
	if item == nil {
		return &storage.ValidationError{
			Field:   "item",
			Message: "config item is required",
		}
	}

	if err := s.validateConfigKey(item.Service, item.Environment, item.Key); err != nil {
		return err
	}

	if item.Value == nil {
		return &storage.ValidationError{
			Field:   "value",
			Message: "value is required",
		}
	}

	return nil
}

// checkPermission 检查权限（占位符实现）
func (s *ConfigService) checkPermission(ctx context.Context, operator, action, service, environment, key string) error {
	// TODO: 实现实际的权限检查逻辑
	// 这里可以集成RBAC或其他权限系统
	return nil
}

// auditLog 审计日志
func (s *ConfigService) auditLog(ctx context.Context, action, service, environment, key string, oldValue, newValue interface{}, operator string) {
	s.logger.Info("config audit",
		zap.String("action", action),
		zap.String("service", service),
		zap.String("environment", environment),
		zap.String("key", key),
		zap.Any("old_value", oldValue),
		zap.Any("new_value", newValue),
		zap.String("operator", operator),
		zap.Time("timestamp", time.Now()),
	)
}