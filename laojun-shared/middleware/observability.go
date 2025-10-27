package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/laojun-shared/observability"
)

// ObservabilityConfig 可观测性中间件配置
type ObservabilityConfig struct {
	// 是否启用监控
	EnableMonitoring bool
	// 是否启用追踪
	EnableTracing bool
	// 服务名称
	ServiceName string
	// 操作名称前缀
	OperationPrefix string
	// 是否记录请求体大小
	RecordRequestSize bool
	// 是否记录响应体大小
	RecordResponseSize bool
	// 是否记录用户代理
	RecordUserAgent bool
	// 是否记录客户端IP
	RecordClientIP bool
	// 自定义标签提取器
	LabelExtractor func(*gin.Context) map[string]string
	// 自定义属性提取器
	AttributeExtractor func(*gin.Context) map[string]interface{}
}

// DefaultObservabilityConfig 返回默认配置
func DefaultObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		EnableMonitoring:   true,
		EnableTracing:      true,
		ServiceName:        "http-service",
		OperationPrefix:    "http",
		RecordRequestSize:  true,
		RecordResponseSize: true,
		RecordUserAgent:    false,
		RecordClientIP:     true,
	}
}

// ObservabilityMiddleware 可观测性中间件
func ObservabilityMiddleware(obs observability.Observability, config ObservabilityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 构建操作名称
		operationName := buildOperationName(config.OperationPrefix, c.Request.Method, c.FullPath())
		
		// 启动操作
	operation, ctx := obs.StartOperation(c.Request.Context(), operationName)
	if operation == nil {
		// 如果启动操作失败，继续处理请求但不记录可观测性数据
		c.Next()
		return
	}
	
	// 更新请求上下文
	c.Request = c.Request.WithContext(ctx)

		// 记录开始时间
		startTime := time.Now()

		// 设置基本属性
		setBasicAttributes(operation, c, config)

		// 设置自定义属性
		if config.AttributeExtractor != nil {
			if attrs := config.AttributeExtractor(c); attrs != nil {
				for key, value := range attrs {
					operation.SetAttribute(key, value)
				}
			}
		}

		// 将操作存储到上下文中，供后续中间件或处理器使用
		c.Set("observability_operation", operation)

		// 处理请求
		c.Next()

		// 请求处理完成后的处理
		finishOperation(operation, c, config, startTime)
	}
}

// buildOperationName 构建操作名称
func buildOperationName(prefix, method, path string) string {
	if path == "" {
		path = "unknown"
	}
	return fmt.Sprintf("%s.%s.%s", prefix, method, sanitizePath(path))
}

// sanitizePath 清理路径，移除参数并替换特殊字符
func sanitizePath(path string) string {
	// 移除查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	
	// 替换特殊字符
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, ":", "_")
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")
	
	// 移除开头的下划线
	if strings.HasPrefix(path, "_") {
		path = path[1:]
	}
	
	if path == "" {
		path = "root"
	}
	
	return path
}

// setBasicAttributes 设置基本属性
func setBasicAttributes(operation observability.Operation, c *gin.Context, config ObservabilityConfig) {
	// HTTP相关属性
	operation.SetAttribute("http.method", c.Request.Method)
	operation.SetAttribute("http.url", c.Request.URL.String())
	operation.SetAttribute("http.scheme", c.Request.URL.Scheme)
	operation.SetAttribute("http.host", c.Request.Host)
	operation.SetAttribute("http.path", c.Request.URL.Path)
	
	// 请求头
	if userAgent := c.Request.UserAgent(); userAgent != "" && config.RecordUserAgent {
		operation.SetAttribute("http.user_agent", userAgent)
	}
	
	// 客户端IP
	if config.RecordClientIP {
		operation.SetAttribute("http.client_ip", c.ClientIP())
	}
	
	// 请求大小
	if config.RecordRequestSize && c.Request.ContentLength > 0 {
		operation.SetAttribute("http.request_size", c.Request.ContentLength)
	}
	
	// 请求ID（如果存在）
	if requestID := c.GetString("request_id"); requestID != "" {
		operation.SetAttribute("http.request_id", requestID)
	}
	
	// 服务信息
	operation.SetAttribute("service.name", config.ServiceName)
}

// finishOperation 完成操作并记录指标
func finishOperation(operation observability.Operation, c *gin.Context, config ObservabilityConfig, startTime time.Time) {
	// 计算持续时间
	duration := time.Since(startTime)
	
	// 获取响应状态码
	statusCode := c.Writer.Status()
	statusStr := strconv.Itoa(statusCode)
	
	// 设置响应相关属性
	operation.SetAttribute("http.status_code", statusCode)
	operation.SetAttribute("http.status_text", http.StatusText(statusCode))
	
	// 响应大小
	if config.RecordResponseSize {
		operation.SetAttribute("http.response_size", c.Writer.Size())
	}
	
	// 设置操作状态
	if statusCode >= 400 {
		if statusCode >= 500 {
			operation.SetStatus(observability.OperationStatusError)
		} else {
			operation.SetStatus(observability.OperationStatusError)
		}
		
		// 记录错误信息
		if len(c.Errors) > 0 {
			operation.SetError(c.Errors.Last())
		}
	} else {
		operation.SetStatus(observability.OperationStatusSuccess)
	}
	
	// 记录监控指标
	recordMetrics(operation, c, config, duration, statusStr)
	
	// 完成操作
	operation.Finish()
}

// recordMetrics 记录监控指标
func recordMetrics(operation observability.Operation, c *gin.Context, config ObservabilityConfig, duration time.Duration, statusCode string) {
	// 构建标签
	labels := map[string]string{
		"method":      c.Request.Method,
		"path":        sanitizePath(c.FullPath()),
		"status_code": statusCode,
		"service":     config.ServiceName,
	}
	
	// 添加自定义标签
	if config.LabelExtractor != nil {
		if customLabels := config.LabelExtractor(c); customLabels != nil {
			for key, value := range customLabels {
				labels[key] = value
			}
		}
	}
	
	// 记录请求计数
	operation.IncrementCounter("http_requests_total", 1, labels)
	
	// 记录请求持续时间
	operation.RecordHistogram("http_request_duration_seconds", duration.Seconds(), labels)
	
	// 记录请求大小
	if config.RecordRequestSize && c.Request.ContentLength > 0 {
		operation.RecordHistogram("http_request_size_bytes", float64(c.Request.ContentLength), labels)
	}
	
	// 记录响应大小
	if config.RecordResponseSize {
		operation.RecordHistogram("http_response_size_bytes", float64(c.Writer.Size()), labels)
	}
	
	// 记录活跃连接数（使用gauge）
	operation.SetGauge("http_active_requests", 1, labels)
}

// GetOperationFromContext 从Gin上下文中获取可观测性操作
func GetOperationFromContext(c *gin.Context) observability.Operation {
	if op, exists := c.Get("observability_operation"); exists {
		if operation, ok := op.(observability.Operation); ok {
			return operation
		}
	}
	return nil
}

// AddEventToRequest 为当前请求添加事件
func AddEventToRequest(c *gin.Context, name string, attributes map[string]interface{}) {
	if operation := GetOperationFromContext(c); operation != nil {
		var attrs []observability.EventAttribute
		for key, value := range attributes {
			attrs = append(attrs, observability.NewEventAttribute(key, value))
		}
		operation.AddEvent(name, attrs...)
	}
}

// SetRequestAttribute 为当前请求设置属性
func SetRequestAttribute(c *gin.Context, key string, value interface{}) {
	if operation := GetOperationFromContext(c); operation != nil {
		operation.SetAttribute(key, value)
	}
}

// IncrementRequestCounter 为当前请求增加计数器
func IncrementRequestCounter(c *gin.Context, name string, labels map[string]string) {
	if operation := GetOperationFromContext(c); operation != nil {
		operation.IncrementCounter(name, 1, labels)
	}
}

// RecordRequestMetric 为当前请求记录指标
func RecordRequestMetric(c *gin.Context, name string, value float64, labels map[string]string) {
	if operation := GetOperationFromContext(c); operation != nil {
		operation.RecordHistogram(name, value, labels)
	}
}

// StartChildOperation 为当前请求启动子操作
func StartChildOperation(c *gin.Context, name string, opts ...observability.OperationOption) observability.Operation {
	if parentOp := GetOperationFromContext(c); parentOp != nil {
		return parentOp.StartChild(name, opts...)
	}
	return nil
}