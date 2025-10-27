package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "laojun_marketplace_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// HTTP请求持续时间
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "laojun_marketplace_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 当前活跃连接数
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "laojun_marketplace_active_connections",
			Help: "Number of active connections",
		},
	)

	// 业务指标
	pluginOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "laojun_marketplace_plugin_operations_total",
			Help: "Total number of plugin operations",
		},
		[]string{"operation", "status"},
	)

	userOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "laojun_marketplace_user_operations_total",
			Help: "Total number of user operations",
		},
		[]string{"operation", "status"},
	)

	paymentOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "laojun_marketplace_payment_operations_total",
			Help: "Total number of payment operations",
		},
		[]string{"operation", "status"},
	)

	// 数据库操作指标
	dbOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "laojun_marketplace_db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)

	dbOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "laojun_marketplace_db_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// 缓存操作指标
	cacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "laojun_marketplace_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"},
	)

	cacheHitRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "laojun_marketplace_cache_hit_ratio",
			Help: "Cache hit ratio",
		},
		[]string{"cache_type"},
	)
)

// PrometheusMiddleware 创建Prometheus监控中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 增加活跃连接数
		activeConnections.Inc()
		defer activeConnections.Dec()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// RecordPluginOperation 记录插件操作指标
func RecordPluginOperation(operation, status string) {
	pluginOperations.WithLabelValues(operation, status).Inc()
}

// RecordUserOperation 记录用户操作指标
func RecordUserOperation(operation, status string) {
	userOperations.WithLabelValues(operation, status).Inc()
}

// RecordPaymentOperation 记录支付操作指标
func RecordPaymentOperation(operation, status string) {
	paymentOperations.WithLabelValues(operation, status).Inc()
}

// RecordDBOperation 记录数据库操作指标
func RecordDBOperation(operation, table, status string, duration time.Duration) {
	dbOperations.WithLabelValues(operation, table, status).Inc()
	dbOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCacheOperation 记录缓存操作指标
func RecordCacheOperation(operation, status string) {
	cacheOperations.WithLabelValues(operation, status).Inc()
}

// UpdateCacheHitRatio 更新缓存命中率
func UpdateCacheHitRatio(cacheType string, ratio float64) {
	cacheHitRatio.WithLabelValues(cacheType).Set(ratio)
}

// BusinessMetrics 业务指标收集器
type BusinessMetrics struct {
	// 插件相关指标
	PluginDownloads   prometheus.Counter
	PluginUploads     prometheus.Counter
	PluginViews       prometheus.Counter
	
	// 用户相关指标
	UserRegistrations prometheus.Counter
	UserLogins        prometheus.Counter
	
	// 支付相关指标
	PaymentSuccess    prometheus.Counter
	PaymentFailure    prometheus.Counter
	PaymentAmount     prometheus.Histogram
	
	// 评价相关指标
	ReviewsCreated    prometheus.Counter
	ReviewsApproved   prometheus.Counter
	ReviewsRejected   prometheus.Counter
}

// NewBusinessMetrics 创建业务指标收集器
func NewBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{
		PluginDownloads: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_plugin_downloads_total",
			Help: "Total number of plugin downloads",
		}),
		PluginUploads: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_plugin_uploads_total",
			Help: "Total number of plugin uploads",
		}),
		PluginViews: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_plugin_views_total",
			Help: "Total number of plugin views",
		}),
		UserRegistrations: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_user_registrations_total",
			Help: "Total number of user registrations",
		}),
		UserLogins: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_user_logins_total",
			Help: "Total number of user logins",
		}),
		PaymentSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_payment_success_total",
			Help: "Total number of successful payments",
		}),
		PaymentFailure: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_payment_failure_total",
			Help: "Total number of failed payments",
		}),
		PaymentAmount: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "laojun_marketplace_payment_amount",
			Help:    "Payment amount distribution",
			Buckets: []float64{1, 5, 10, 50, 100, 500, 1000, 5000},
		}),
		ReviewsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_reviews_created_total",
			Help: "Total number of reviews created",
		}),
		ReviewsApproved: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_reviews_approved_total",
			Help: "Total number of reviews approved",
		}),
		ReviewsRejected: promauto.NewCounter(prometheus.CounterOpts{
			Name: "laojun_marketplace_reviews_rejected_total",
			Help: "Total number of reviews rejected",
		}),
	}
}