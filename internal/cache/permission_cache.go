package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codetaoist/laojun/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	// 缓存键前缀
	UserPermissionPrefix = "user_perm:"
	RolePermissionPrefix = "role_perm:"
	UserRolePrefix       = "user_role:"

	// 缓存过期时间
	DefaultCacheExpiry = 30 * time.Minute
	ShortCacheExpiry   = 5 * time.Minute
	LongCacheExpiry    = 2 * time.Hour
)

// PermissionCache 权限缓存服务
type PermissionCache struct {
	client *redis.Client
}

// NewPermissionCache 创建权限缓存服务
func NewPermissionCache() *PermissionCache {
	return &PermissionCache{
		client: GetRedisClient(),
	}
}

// CachedPermissionResult 缓存的权限检查结果
type CachedPermissionResult struct {
	HasPermission bool      `json:"has_permission"`
	Reason        string    `json:"reason"`
	CachedAt      time.Time `json:"cached_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// UserPermissionKey 生成用户权限缓存键
func (pc *PermissionCache) UserPermissionKey(userID uuid.UUID, deviceType, module, resource, action string) string {
	return fmt.Sprintf("%s%s:%s:%s:%s:%s", UserPermissionPrefix, userID.String(), deviceType, module, resource, action)
}

// UserRoleKey 生成用户角色缓存键
func (pc *PermissionCache) UserRoleKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s", UserRolePrefix, userID.String())
}

// RolePermissionKey 生成角色权限缓存键
func (pc *PermissionCache) RolePermissionKey(roleID uuid.UUID, deviceType, module string) string {
	return fmt.Sprintf("%s%s:%s:%s", RolePermissionPrefix, roleID.String(), deviceType, module)
}

// GetUserPermission 获取缓存的用户权限检查结果
func (pc *PermissionCache) GetUserPermission(userID uuid.UUID, deviceType, module, resource, action string) (*CachedPermissionResult, error) {
	key := pc.UserPermissionKey(userID, deviceType, module, resource, action)

	val, err := pc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to get permission from cache: %w", err)
	}

	var result CachedPermissionResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached permission: %w", err)
	}

	// 检查是否过期
	if time.Now().After(result.ExpiresAt) {
		// 删除过期的缓存项
		pc.client.Del(ctx, key)
		return nil, nil
	}

	return &result, nil
}

// SetUserPermission 设置用户权限检查结果到缓存
func (pc *PermissionCache) SetUserPermission(userID uuid.UUID, deviceType, module, resource, action string, hasPermission bool, reason string, expiry time.Duration) error {
	key := pc.UserPermissionKey(userID, deviceType, module, resource, action)

	result := CachedPermissionResult{
		HasPermission: hasPermission,
		Reason:        reason,
		CachedAt:      time.Now(),
		ExpiresAt:     time.Now().Add(expiry),
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal permission result: %w", err)
	}

	return pc.client.Set(ctx, key, data, expiry).Err()
}

// GetUserRoles 获取缓存的用户角色
func (pc *PermissionCache) GetUserRoles(userID uuid.UUID) ([]string, error) {
	key := pc.UserRoleKey(userID)

	val, err := pc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to get user roles from cache: %w", err)
	}

	var roles []string
	if err := json.Unmarshal([]byte(val), &roles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached roles: %w", err)
	}

	return roles, nil
}

// SetUserRoles 设置用户角色到缓存
func (pc *PermissionCache) SetUserRoles(userID uuid.UUID, roles []string, expiry time.Duration) error {
	key := pc.UserRoleKey(userID)

	data, err := json.Marshal(roles)
	if err != nil {
		return fmt.Errorf("failed to marshal user roles: %w", err)
	}

	return pc.client.Set(ctx, key, data, expiry).Err()
}

// InvalidateUserPermissions 清除用户的所有权限缓存项
func (pc *PermissionCache) InvalidateUserPermissions(userID uuid.UUID) error {
	pattern := fmt.Sprintf("%s%s:*", UserPermissionPrefix, userID.String())

	keys, err := pc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		return pc.client.Del(ctx, keys...).Err()
	}

	return nil
}

// InvalidateUserRoles 清除用户角色缓存
func (pc *PermissionCache) InvalidateUserRoles(userID uuid.UUID) error {
	key := pc.UserRoleKey(userID)
	return pc.client.Del(ctx, key).Err()
}

// InvalidateRolePermissions 清除角色权限缓存
func (pc *PermissionCache) InvalidateRolePermissions(roleID uuid.UUID) error {
	pattern := fmt.Sprintf("%s%s:*", RolePermissionPrefix, roleID.String())

	keys, err := pc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		return pc.client.Del(ctx, keys...).Err()
	}

	return nil
}

// InvalidateAllUserCache 清除用户的所有缓存（权限和角色）
func (pc *PermissionCache) InvalidateAllUserCache(userID uuid.UUID) error {
	// 清除权限缓存
	if err := pc.InvalidateUserPermissions(userID); err != nil {
		return err
	}

	// 清除角色缓存
	return pc.InvalidateUserRoles(userID)
}

// GetCacheStats 获取缓存统计信息
func (pc *PermissionCache) GetCacheStats() (map[string]interface{}, error) {
	info, err := pc.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// 获取权限相关的键数量
	permKeys, err := pc.client.Keys(ctx, UserPermissionPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get permission keys: %w", err)
	}

	roleKeys, err := pc.client.Keys(ctx, UserRolePrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get role keys: %w", err)
	}

	return map[string]interface{}{
		"redis_info":        info,
		"permission_keys":   len(permKeys),
		"role_keys":         len(roleKeys),
		"total_cached_keys": len(permKeys) + len(roleKeys),
	}, nil
}

// WarmupUserCache 预热用户缓存
func (pc *PermissionCache) WarmupUserCache(userID uuid.UUID, permissions []models.UserPermissionCheckRequest) error {
	// 这里可以实现批量预热逻辑
	// 暂时返回nil，后续可以根据需要实现批量预热
	return nil
}

// CleanExpiredCache 清理过期缓存
func (pc *PermissionCache) CleanExpiredCache() error {
	// Redis会自动清理过期键，这里主要是为了提供手动清理接口
	// 可以实现一些自定义的清理逻辑
	return nil
}
