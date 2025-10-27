package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/codetaoist/laojun-gateway/internal/services/discovery"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Service 代理服务
type Service struct {
	config    config.ProxyConfig
	discovery discovery.Service
	logger    *zap.Logger
	client    *http.Client
	balancer  LoadBalancer
}

// NewService 创建代理服务
func NewService(cfg config.ProxyConfig, discoveryService discovery.Service, logger *zap.Logger) *Service {
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	var balancer LoadBalancer
	switch cfg.LoadBalancer {
	case "round_robin":
		balancer = NewRoundRobinBalancer()
	case "random":
		balancer = NewRandomBalancer()
	case "weighted":
		balancer = NewWeightedBalancer()
	default:
		balancer = NewRoundRobinBalancer()
	}

	return &Service{
		config:    cfg,
		discovery: discoveryService,
		logger:    logger,
		client:    client,
		balancer:  balancer,
	}
}

// ProxyRequest 代理请求
func (s *Service) ProxyRequest(c *gin.Context, route config.RouteConfig) error {
	// 获取目标服务实例
	target, err := s.getTarget(route)
	if err != nil {
		return fmt.Errorf("failed to get target: %w", err)
	}

	// 构建目标URL
	targetURL, err := s.buildTargetURL(c, route, target)
	if err != nil {
		return fmt.Errorf("failed to build target URL: %w", err)
	}

	// 创建代理请求
	proxyReq, err := s.createProxyRequest(c, targetURL)
	if err != nil {
		return fmt.Errorf("failed to create proxy request: %w", err)
	}

	// 添加自定义头部
	s.addCustomHeaders(proxyReq, route.Headers)

	// 执行请求
	resp, err := s.executeRequest(proxyReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 复制响应
	s.copyResponse(c, resp)

	return nil
}

// getTarget 获取目标服务实例
func (s *Service) getTarget(route config.RouteConfig) (string, error) {
	if route.Target != "" {
		// 使用静态目标
		return route.Target, nil
	}

	if route.Service != "" {
		// 使用服务发现
		instances, err := s.discovery.GetHealthyInstances(route.Service)
		if err != nil {
			return "", err
		}

		if len(instances) == 0 {
			return "", fmt.Errorf("no healthy instances found for service: %s", route.Service)
		}

		// 使用负载均衡选择实例
		instance := s.balancer.Select(instances)
		return fmt.Sprintf("http://%s:%d", instance.Address, instance.Port), nil
	}

	return "", fmt.Errorf("no target or service specified")
}

// buildTargetURL 构建目标URL
func (s *Service) buildTargetURL(c *gin.Context, route config.RouteConfig, target string) (*url.URL, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	// 处理路径
	path := c.Request.URL.Path
	if route.StripPrefix {
		// 移除路由前缀
		routePath := strings.TrimSuffix(route.Path, "*")
		if strings.HasPrefix(path, routePath) {
			path = strings.TrimPrefix(path, routePath)
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
		}
	}

	targetURL.Path = strings.TrimSuffix(targetURL.Path, "/") + path
	targetURL.RawQuery = c.Request.URL.RawQuery

	return targetURL, nil
}

// createProxyRequest 创建代理请求
func (s *Service) createProxyRequest(c *gin.Context, targetURL *url.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		c.Request.Method,
		targetURL.String(),
		c.Request.Body,
	)
	if err != nil {
		return nil, err
	}

	// 复制头部
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 设置代理头部
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", "http")
	if c.Request.TLS != nil {
		req.Header.Set("X-Forwarded-Proto", "https")
	}
	req.Header.Set("X-Forwarded-Host", c.Request.Host)

	return req, nil
}

// addCustomHeaders 添加自定义头部
func (s *Service) addCustomHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// executeRequest 执行请求（带重试）
func (s *Service) executeRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= s.config.RetryCount; i++ {
		resp, err = s.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		if i < s.config.RetryCount {
			s.logger.Warn("Request failed, retrying",
				zap.Int("attempt", i+1),
				zap.String("url", req.URL.String()),
				zap.Error(err))
			
			// 等待一段时间后重试
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", s.config.RetryCount, err)
}

// copyResponse 复制响应
func (s *Service) copyResponse(c *gin.Context, resp *http.Response) {
	// 复制状态码
	c.Status(resp.StatusCode)

	// 复制头部
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 复制响应体
	_, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		s.logger.Error("Failed to copy response body", zap.Error(err))
	}
}