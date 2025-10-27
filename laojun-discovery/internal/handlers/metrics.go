package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MetricsHandler 监控指标处理器
type MetricsHandler struct {
	logger *zap.Logger
}

// NewMetricsHandler 创建监控指标处理器
func NewMetricsHandler(logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

// GetMetricsHandler 返回Prometheus格式的监控指标
func GetMetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		
		// 基础指标
		metrics := `# HELP laojun_discovery_up Service discovery is up
# TYPE laojun_discovery_up gauge
laojun_discovery_up 1

# HELP laojun_discovery_services_total Total number of registered services
# TYPE laojun_discovery_services_total gauge
laojun_discovery_services_total 0

# HELP laojun_discovery_instances_total Total number of service instances
# TYPE laojun_discovery_instances_total gauge
laojun_discovery_instances_total 0

# HELP laojun_discovery_healthy_instances_total Total number of healthy service instances
# TYPE laojun_discovery_healthy_instances_total gauge
laojun_discovery_healthy_instances_total 0

# HELP laojun_discovery_requests_total Total number of requests
# TYPE laojun_discovery_requests_total counter
laojun_discovery_requests_total 0

# HELP laojun_discovery_request_duration_seconds Request duration in seconds
# TYPE laojun_discovery_request_duration_seconds histogram
laojun_discovery_request_duration_seconds_bucket{le="0.1"} 0
laojun_discovery_request_duration_seconds_bucket{le="0.5"} 0
laojun_discovery_request_duration_seconds_bucket{le="1.0"} 0
laojun_discovery_request_duration_seconds_bucket{le="2.5"} 0
laojun_discovery_request_duration_seconds_bucket{le="5.0"} 0
laojun_discovery_request_duration_seconds_bucket{le="10.0"} 0
laojun_discovery_request_duration_seconds_bucket{le="+Inf"} 0
laojun_discovery_request_duration_seconds_sum 0
laojun_discovery_request_duration_seconds_count 0
`
		w.Write([]byte(metrics))
	})
}

// GetMetrics 获取监控指标（Gin处理器版本）
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	
	// 基础指标
	metrics := `# HELP laojun_discovery_up Service discovery is up
# TYPE laojun_discovery_up gauge
laojun_discovery_up 1

# HELP laojun_discovery_services_total Total number of registered services
# TYPE laojun_discovery_services_total gauge
laojun_discovery_services_total 0

# HELP laojun_discovery_instances_total Total number of service instances
# TYPE laojun_discovery_instances_total gauge
laojun_discovery_instances_total 0

# HELP laojun_discovery_healthy_instances_total Total number of healthy service instances
# TYPE laojun_discovery_healthy_instances_total gauge
laojun_discovery_healthy_instances_total 0

# HELP laojun_discovery_requests_total Total number of requests
# TYPE laojun_discovery_requests_total counter
laojun_discovery_requests_total 0
`
	
	c.String(http.StatusOK, metrics)
}

// GetHealth 健康检查端点
func (h *MetricsHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "laojun-discovery",
		"timestamp": strconv.FormatInt(c.Request.Context().Value("timestamp").(int64), 10),
	})
}

// GetReady 就绪检查端点
func (h *MetricsHandler) GetReady(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"service": "laojun-discovery",
	})
}