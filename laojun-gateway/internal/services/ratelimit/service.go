package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Service 限流服务接口
type Service interface {
	// 检查是否允许请求
	Allow(key string, limit int) (bool, error)
	// 检查全局限流
	AllowGlobal() (bool, error)
	// 检查用户限流
	AllowUser(userID string) (bool, error)
	// 检查IP限流
	AllowIP(ip string) (bool, error)
	// 检查路径限流
	AllowPath(path, method string) (bool, error)
}

// RedisRateLimitService Redis限流服务实现
type RedisRateLimitService struct {
	config      config.RateLimitConfig
	redis       *redis.Client
	logger      *zap.Logger
	globalLimiter *rate.Limiter
}

// NewService 创建限流服务
func NewService(cfg config.RateLimitConfig, redisClient *redis.Client, logger *zap.Logger) Service {
	globalLimiter := rate.NewLimiter(rate.Limit(cfg.GlobalRate), cfg.GlobalRate)
	
	return &RedisRateLimitService{
		config:        cfg,
		redis:         redisClient,
		logger:        logger,
		globalLimiter: globalLimiter,
	}
}

// Allow 检查是否允许请求
func (r *RedisRateLimitService) Allow(key string, limit int) (bool, error) {
	if !r.config.Enabled {
		return true, nil
	}

	ctx := context.Background()
	now := time.Now()
	window := now.Truncate(time.Minute)
	
	// 使用滑动窗口算法
	pipe := r.redis.Pipeline()
	
	// 清理过期的计数
	expiredWindow := window.Add(-time.Minute)
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", expiredWindow.Unix()))
	
	// 添加当前请求
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.Unix()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// 获取当前窗口内的请求数
	pipe.ZCount(ctx, key, fmt.Sprintf("%d", window.Unix()), "+inf")
	
	// 设置过期时间
	pipe.Expire(ctx, key, time.Minute*2)
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("Rate limit check failed", zap.Error(err))
		return false, err
	}
	
	// 获取当前请求数
	count := results[2].(*redis.IntCmd).Val()
	
	allowed := count <= int64(limit)
	
	if !allowed {
		r.logger.Warn("Rate limit exceeded", 
			zap.String("key", key),
			zap.Int64("count", count),
			zap.Int("limit", limit))
	}
	
	return allowed, nil
}

// AllowGlobal 检查全局限流
func (r *RedisRateLimitService) AllowGlobal() (bool, error) {
	if !r.config.Enabled {
		return true, nil
	}
	
	return r.globalLimiter.Allow(), nil
}

// AllowUser 检查用户限流
func (r *RedisRateLimitService) AllowUser(userID string) (bool, error) {
	if !r.config.Enabled {
		return true, nil
	}
	
	key := fmt.Sprintf("rate_limit:user:%s", userID)
	return r.Allow(key, r.config.UserRate)
}

// AllowIP 检查IP限流
func (r *RedisRateLimitService) AllowIP(ip string) (bool, error) {
	if !r.config.Enabled {
		return true, nil
	}
	
	key := fmt.Sprintf("rate_limit:ip:%s", ip)
	return r.Allow(key, r.config.IPRate)
}

// AllowPath 检查路径限流
func (r *RedisRateLimitService) AllowPath(path, method string) (bool, error) {
	if !r.config.Enabled {
		return true, nil
	}
	
	// 检查是否有特定路径的限流规则
	for _, rule := range r.config.Rules {
		if rule.Path == path && (rule.Method == "" || rule.Method == method) {
			key := fmt.Sprintf("rate_limit:path:%s:%s", method, path)
			return r.Allow(key, rule.Rate)
		}
	}
	
	return true, nil
}