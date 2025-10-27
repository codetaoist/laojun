package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuditorMiddleware 审核员权限中间件
func AuditorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// TODO: 从数据库或缓存中检查用户是否具有审核员权限
		// 这里需要查询用户角色表，检查用户是否具有 "auditor" 角色
		// 或者检查用户是否在审核员表中
		
		// 临时实现：假设所有认证用户都可以是审核员
		// 在实际实现中，需要根据用户角色或权限进行验证
		isAuditor := checkAuditorPermission(uid)
		if !isAuditor {
			c.JSON(http.StatusForbidden, gin.H{"error": "Auditor permission required"})
			c.Abort()
			return
		}

		// 将审核员信息添加到上下文
		c.Set("auditor_id", uid)
		c.Set("is_auditor", true)

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// TODO: 从数据库或缓存中检查用户是否具有管理员权限
		// 这里需要查询用户角色表，检查用户是否具有 "admin" 或 "super_admin" 角色
		
		// 临时实现：假设所有认证用户都可以是管理员
		// 在实际实现中，需要根据用户角色或权限进行验证
		isAdmin := checkAdminPermission(uid)
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Administrator permission required"})
			c.Abort()
			return
		}

		// 将管理员信息添加到上下文
		c.Set("admin_id", uid)
		c.Set("is_admin", true)

		c.Next()
	}
}

// WebhookSignatureMiddleware Webhook签名验证中间件
func WebhookSignatureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取签名头
		signature := c.GetHeader("X-Webhook-Signature")
		if signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing webhook signature"})
			c.Abort()
			return
		}

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		// 验证签名
		if !verifyWebhookSignature(body, signature) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
			c.Abort()
			return
		}

		// 将请求体重新设置到上下文中，供后续处理器使用
		c.Set("webhook_body", body)
		c.Set("webhook_verified", true)

		c.Next()
	}
}

// DeveloperMiddleware 开发者权限中间件
func DeveloperMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// TODO: 检查用户是否为已认证的开发者
		// 这里需要查询开发者档案表，检查用户是否已完成开发者认证
		
		// 临时实现：假设所有认证用户都可以是开发者
		isDeveloper := checkDeveloperStatus(uid)
		if !isDeveloper {
			c.JSON(http.StatusForbidden, gin.H{"error": "Developer verification required"})
			c.Abort()
			return
		}

		// 将开发者信息添加到上下文
		c.Set("developer_id", uid)
		c.Set("is_developer", true)

		c.Next()
	}
}

// PluginOwnershipMiddleware 插件所有权验证中间件
func PluginOwnershipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取插件ID
		pluginIDStr := c.Param("plugin_id")
		if pluginIDStr == "" {
			pluginIDStr = c.Query("plugin_id")
		}

		if pluginIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID required"})
			c.Abort()
			return
		}

		pluginID, err := uuid.Parse(pluginIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plugin ID"})
			c.Abort()
			return
		}

		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// TODO: 检查用户是否为插件的所有者
		// 这里需要查询插件表，检查插件的开发者ID是否与当前用户ID匹配
		
		// 临时实现：假设用户拥有所有插件
		isOwner := checkPluginOwnership(uid, pluginID)
		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "Plugin ownership required"})
			c.Abort()
			return
		}

		// 将插件信息添加到上下文
		c.Set("plugin_id", pluginID)
		c.Set("is_plugin_owner", true)

		c.Next()
	}
}

// 辅助函数

// checkAuditorPermission 检查审核员权限
func checkAuditorPermission(userID uuid.UUID) bool {
	// TODO: 实现实际的权限检查逻辑
	// 1. 查询用户角色表
	// 2. 检查用户是否具有 "auditor" 角色
	// 3. 或者检查用户是否在审核员表中
	
	// 临时返回true，实际实现时需要连接数据库查询
	logrus.WithField("user_id", userID).Debug("Checking auditor permission (placeholder)")
	return true
}

// checkAdminPermission 检查管理员权限
func checkAdminPermission(userID uuid.UUID) bool {
	// TODO: 实现实际的权限检查逻辑
	// 1. 查询用户角色表
	// 2. 检查用户是否具有 "admin" 或 "super_admin" 角色
	
	// 临时返回true，实际实现时需要连接数据库查询
	logrus.WithField("user_id", userID).Debug("Checking admin permission (placeholder)")
	return true
}

// checkDeveloperStatus 检查开发者状态
func checkDeveloperStatus(userID uuid.UUID) bool {
	// TODO: 实现实际的开发者状态检查逻辑
	// 1. 查询开发者档案表
	// 2. 检查用户是否已完成开发者认证
	// 3. 检查认证状态是否有效
	
	// 临时返回true，实际实现时需要连接数据库查询
	logrus.WithField("user_id", userID).Debug("Checking developer status (placeholder)")
	return true
}

// checkPluginOwnership 检查插件所有权
func checkPluginOwnership(userID, pluginID uuid.UUID) bool {
	// TODO: 实现实际的插件所有权检查逻辑
	// 1. 查询插件表
	// 2. 检查插件的开发者ID是否与当前用户ID匹配
	
	// 临时返回true，实际实现时需要连接数据库查询
	logrus.WithFields(logrus.Fields{
		"user_id":   userID,
		"plugin_id": pluginID,
	}).Debug("Checking plugin ownership (placeholder)")
	return true
}

// verifyWebhookSignature 验证Webhook签名
func verifyWebhookSignature(body []byte, signature string) bool {
	// TODO: 从配置中获取Webhook密钥
	webhookSecret := getWebhookSecret()
	if webhookSecret == "" {
		logrus.Warn("Webhook secret not configured")
		return false
	}

	// 计算HMAC-SHA256签名
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// 比较签名（支持sha256=前缀格式）
	if strings.HasPrefix(signature, "sha256=") {
		signature = strings.TrimPrefix(signature, "sha256=")
	}

	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// getWebhookSecret 获取Webhook密钥
func getWebhookSecret() string {
	// TODO: 从配置文件或环境变量中获取Webhook密钥
	// 这里应该从配置管理系统中获取密钥
	
	// 临时返回固定值，实际实现时需要从安全的配置源获取
	return "your-webhook-secret-key"
}