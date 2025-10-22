package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RateLimitHandler 频率限制处理
type RateLimitHandler struct {
	redisClient *redis.Client
}

// NewRateLimitHandler 创建频率限制处理
func NewRateLimitHandler(redisClient *redis.Client) *RateLimitHandler {
	return &RateLimitHandler{
		redisClient: redisClient,
	}
}

// RateLimitStatus 频率限制状态
type RateLimitStatus struct {
	Key       string `json:"key"`
	Current   int64  `json:"current"`
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	ResetTime int64  `json:"reset_time"`
}

// RateLimitStats 频率限制统计
type RateLimitStats struct {
	TotalKeys       int64             `json:"total_keys"`
	ActiveLimits    []RateLimitStatus `json:"active_limits"`
	TopLimitedIPs   []string          `json:"top_limited_ips"`
	TopLimitedUsers []string          `json:"top_limited_users"`
}

// GetStats 获取频率限制统计信息
func (h *RateLimitHandler) GetStats(c *gin.Context) {
	ctx := h.redisClient.Context()

	// 获取所有频率限制相关的键
	keys, err := h.redisClient.Keys(ctx, "rate_limit:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取频率限制统计失败",
			"details": err.Error(),
		})
		return
	}

	stats := RateLimitStats{
		TotalKeys:    int64(len(keys)),
		ActiveLimits: make([]RateLimitStatus, 0),
	}

	// 获取活跃的限制信息
	for _, key := range keys {
		count, err := h.redisClient.ZCard(ctx, key).Result()
		if err != nil {
			continue
		}

		if count > 0 {
			ttl, _ := h.redisClient.TTL(ctx, key).Result()
			resetTime := time.Now().Add(ttl).Unix()

			status := RateLimitStatus{
				Key:       key,
				Current:   count,
				ResetTime: resetTime,
			}
			stats.ActiveLimits = append(stats.ActiveLimits, status)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "频率限制统计获取成功",
		"data":    stats,
	})
}

// ClearUserRateLimit 清除用户的频率限制
func (h *RateLimitHandler) ClearUserRateLimit(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	ctx := h.redisClient.Context()

	// 查找该用户的所有频率限制键
	pattern := "rate_limit:*user:" + userID.String()
	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查找用户频率限制失败",
			"details": err.Error(),
		})
		return
	}

	// 删除所有相关的键
	if len(keys) > 0 {
		err = h.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "清除用户频率限制失败",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "用户频率限制清除成功",
		"cleared_keys": len(keys),
	})
}

// ClearIPRateLimit 清除IP的频率限制
func (h *RateLimitHandler) ClearIPRateLimit(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "IP地址不能为空",
		})
		return
	}

	ctx := h.redisClient.Context()

	// 查找该IP的所有频率限制键
	pattern := "rate_limit:*ip:" + ip
	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查找IP频率限制失败",
			"details": err.Error(),
		})
		return
	}

	// 删除所有相关的键
	if len(keys) > 0 {
		err = h.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "清除IP频率限制失败",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "IP频率限制清除成功",
		"cleared_keys": len(keys),
	})
}

// GetCurrentUserRateLimit 获取当前用户的频率限制状态
func (h *RateLimitHandler) GetCurrentUserRateLimit(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户ID格式错误",
		})
		return
	}

	ctx := h.redisClient.Context()

	// 查找当前用户的频率限制键
	pattern := "rate_limit:*user:" + uid.String()
	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户频率限制状态失败",
			"details": err.Error(),
		})
		return
	}

	var limits []RateLimitStatus
	for _, key := range keys {
		count, err := h.redisClient.ZCard(ctx, key).Result()
		if err != nil {
			continue
		}

		ttl, _ := h.redisClient.TTL(ctx, key).Result()
		resetTime := time.Now().Add(ttl).Unix()

		status := RateLimitStatus{
			Key:       key,
			Current:   count,
			ResetTime: resetTime,
		}
		limits = append(limits, status)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户频率限制状态获取成功",
		"data":    limits,
	})
}

// ClearCurrentUserRateLimit 清除当前用户的频率限制
func (h *RateLimitHandler) ClearCurrentUserRateLimit(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户ID格式错误",
		})
		return
	}

	ctx := h.redisClient.Context()

	// 查找当前用户的所有频率限制键
	pattern := "rate_limit:*user:" + uid.String()
	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查找用户频率限制失败",
			"details": err.Error(),
		})
		return
	}

	// 删除所有相关的键
	if len(keys) > 0 {
		err = h.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "清除用户频率限制失败",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "当前用户频率限制清除成功",
		"cleared_keys": len(keys),
	})
}

// SetCustomRateLimit 设置自定义频率限制
func (h *RateLimitHandler) SetCustomRateLimit(c *gin.Context) {
	var req struct {
		Key     string `json:"key" binding:"required"`
		Limit   int    `json:"limit" binding:"required,min=1"`
		Window  int    `json:"window" binding:"required,min=1"` // 秒数
		Current int    `json:"current"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	ctx := h.redisClient.Context()
	now := time.Now()

	// 清除现有的限制
	h.redisClient.Del(ctx, req.Key)

	// 设置新的限制
	for i := 0; i < req.Current; i++ {
		h.redisClient.ZAdd(ctx, req.Key, &redis.Z{
			Score:  float64(now.Add(-time.Duration(i) * time.Second).UnixNano()),
			Member: now.Add(-time.Duration(i) * time.Second).UnixNano(),
		})
	}

	// 设置过期时间
	h.redisClient.Expire(ctx, req.Key, time.Duration(req.Window)*time.Second)

	c.JSON(http.StatusOK, gin.H{
		"message": "自定义频率限制设置成功",
		"key":     req.Key,
		"limit":   req.Limit,
		"window":  req.Window,
		"current": req.Current,
	})
}

// ResetLimit 重置指定键的频率限制
func (h *RateLimitHandler) ResetLimit(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "键名不能为空",
		})
		return
	}

	ctx := h.redisClient.Context()

	// 删除指定的频率限制键
	err := h.redisClient.Del(ctx, key).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "重置频率限制失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "频率限制重置成功",
		"key":     key,
	})
}

// GetBlockedIPs 获取被阻止的IP列表
func (h *RateLimitHandler) GetBlockedIPs(c *gin.Context) {
	ctx := h.redisClient.Context()

	// 查找所有被阻止的IP键
	keys, err := h.redisClient.Keys(ctx, "rate_limit:blocked:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取被阻止IP列表失败",
			"details": err.Error(),
		})
		return
	}

	var blockedIPs []map[string]interface{}
	for _, key := range keys {
		// 提取IP地址
		ip := key[len("rate_limit:blocked:"):]

		// 获取阻止时间
		ttl, _ := h.redisClient.TTL(ctx, key).Result()

		blockedIPs = append(blockedIPs, map[string]interface{}{
			"ip":         ip,
			"key":        key,
			"expires_in": int64(ttl.Seconds()),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "被阻止IP列表获取成功",
		"data":    blockedIPs,
		"total":   len(blockedIPs),
	})
}

// UnblockIP 解除IP阻止
func (h *RateLimitHandler) UnblockIP(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "IP地址不能为空",
		})
		return
	}

	ctx := h.redisClient.Context()

	// 查找该IP的所有阻止键
	pattern := "rate_limit:blocked:" + ip
	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查找IP阻止记录失败",
			"details": err.Error(),
		})
		return
	}

	// 删除所有相关的键
	if len(keys) > 0 {
		err = h.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "解除IP阻止失败",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "IP阻止解除成功",
		"ip":           ip,
		"cleared_keys": len(keys),
	})
}
