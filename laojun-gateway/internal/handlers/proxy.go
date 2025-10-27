package handlers

import (
	"net/http"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/codetaoist/laojun-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	proxyService *proxy.Service
	logger       *zap.Logger
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(proxyService *proxy.Service, logger *zap.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		logger:       logger,
	}
}

// ProxyRequest 代理请求
func (h *ProxyHandler) ProxyRequest(c *gin.Context, route config.RouteConfig) {
	// 记录请求开始
	h.logger.Debug("Proxying request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("service", route.Service),
		zap.String("target", route.Target))

	// 执行代理请求
	err := h.proxyService.ProxyRequest(c, route)
	if err != nil {
		h.logger.Error("Proxy request failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("service", route.Service),
			zap.String("target", route.Target),
			zap.Error(err))

		// 如果响应还没有写入，返回错误响应
		if !c.Writer.Written() {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Proxy request failed",
				"code":  "PROXY_ERROR",
				"details": err.Error(),
			})
		}
		return
	}

	h.logger.Debug("Proxy request completed",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("service", route.Service),
		zap.Int("status", c.Writer.Status()))
}