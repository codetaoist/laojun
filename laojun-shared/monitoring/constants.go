package monitoring

import "time"

// MetricNamespace 指标命名空间
const (
	// 系统级别命名空间
	NamespaceSystem     = "laojun_system"
	NamespaceService    = "laojun_service"
	NamespaceGateway    = "laojun_gateway"
	NamespaceDiscovery  = "laojun_discovery"
	NamespaceMonitoring = "laojun_monitoring"
	NamespacePlugins    = "laojun_plugins"
	NamespaceMarketplace = "laojun_marketplace"
	
	// 功能级别命名空间
	NamespaceHTTP       = "laojun_http"
	NamespaceGRPC       = "laojun_grpc"
	NamespaceDatabase   = "laojun_database"
	NamespaceCache      = "laojun_cache"
	NamespaceQueue      = "laojun_queue"
	NamespaceAuth       = "laojun_auth"
	NamespaceHealth     = "laojun_health"
)

// MetricSubsystem 指标子系统
const (
	// HTTP相关子系统
	SubsystemHTTPServer   = "http_server"
	SubsystemHTTPClient   = "http_client"
	SubsystemHTTPProxy    = "http_proxy"
	SubsystemHTTPGateway  = "http_gateway"
	
	// 数据库相关子系统
	SubsystemDBConnection = "db_connection"
	SubsystemDBQuery      = "db_query"
	SubsystemDBTransaction = "db_transaction"
	SubsystemDBMigration  = "db_migration"
	
	// 缓存相关子系统
	SubsystemCacheRedis   = "cache_redis"
	SubsystemCacheMemory  = "cache_memory"
	SubsystemCacheHit     = "cache_hit"
	SubsystemCacheMiss    = "cache_miss"
	
	// 服务发现相关子系统
	SubsystemServiceRegistry   = "service_registry"
	SubsystemServiceDiscovery  = "service_discovery"
	SubsystemServiceHealth     = "service_health"
	SubsystemLoadBalancer      = "load_balancer"
	
	// 认证相关子系统
	SubsystemAuthJWT      = "auth_jwt"
	SubsystemAuthOAuth    = "auth_oauth"
	SubsystemAuthSession  = "auth_session"
	SubsystemAuthRBAC     = "auth_rbac"
	
	// 插件相关子系统
	SubsystemPluginEngine   = "plugin_engine"
	SubsystemPluginRegistry = "plugin_registry"
	SubsystemPluginRuntime  = "plugin_runtime"
	SubsystemPluginSandbox  = "plugin_sandbox"
	
	// 市场相关子系统
	SubsystemMarketplaceAPI = "marketplace_api"
	SubsystemMarketplaceWeb = "marketplace_web"
	SubsystemMarketplacePlugin = "marketplace_plugin"
)

// MetricName 标准指标名称
const (
	// 通用指标
	MetricRequestsTotal        = "requests_total"
	MetricRequestDuration      = "request_duration_seconds"
	MetricRequestSize          = "request_size_bytes"
	MetricResponseSize         = "response_size_bytes"
	MetricErrorsTotal          = "errors_total"
	MetricActiveConnections    = "active_connections"
	MetricConnectionsTotal     = "connections_total"
	
	// HTTP指标
	MetricHTTPRequestsTotal           = "http_requests_total"
	MetricHTTPRequestDuration         = "http_request_duration_seconds"
	MetricHTTPRequestSize             = "http_request_size_bytes"
	MetricHTTPResponseSize            = "http_response_size_bytes"
	MetricHTTPActiveRequests          = "http_active_requests"
	MetricHTTPConcurrentConnections   = "http_concurrent_connections"
	
	// 数据库指标
	MetricDBConnectionsOpen           = "db_connections_open"
	MetricDBConnectionsIdle           = "db_connections_idle"
	MetricDBConnectionsInUse          = "db_connections_in_use"
	MetricDBConnectionsWaitCount      = "db_connections_wait_count"
	MetricDBConnectionsWaitDuration   = "db_connections_wait_duration_seconds"
	MetricDBQueriesTotal              = "db_queries_total"
	MetricDBQueryDuration             = "db_query_duration_seconds"
	MetricDBTransactionsTotal         = "db_transactions_total"
	MetricDBTransactionDuration       = "db_transaction_duration_seconds"
	
	// 缓存指标
	MetricCacheHitsTotal              = "cache_hits_total"
	MetricCacheMissesTotal            = "cache_misses_total"
	MetricCacheOperationsTotal        = "cache_operations_total"
	MetricCacheOperationDuration      = "cache_operation_duration_seconds"
	MetricCacheSize                   = "cache_size_bytes"
	MetricCacheEntries                = "cache_entries"
	MetricCacheEvictionsTotal         = "cache_evictions_total"
	
	// 服务发现指标
	MetricServiceRegistrationsTotal   = "service_registrations_total"
	MetricServiceDeregistrationsTotal = "service_deregistrations_total"
	MetricServiceDiscoveryTotal       = "service_discovery_total"
	MetricServiceHealthChecksTotal    = "service_health_checks_total"
	MetricServiceHealthCheckDuration  = "service_health_check_duration_seconds"
	MetricActiveServices              = "active_services"
	MetricServiceInstances            = "service_instances"
	
	// 认证指标
	MetricAuthAttemptsTotal           = "auth_attempts_total"
	MetricAuthSuccessTotal            = "auth_success_total"
	MetricAuthFailuresTotal           = "auth_failures_total"
	MetricAuthTokensActive            = "auth_tokens_active"
	MetricAuthTokensIssued            = "auth_tokens_issued"
	MetricAuthTokensExpired           = "auth_tokens_expired"
	MetricAuthSessionsActive          = "auth_sessions_active"
	
	// 插件指标
	MetricPluginsLoaded               = "plugins_loaded"
	MetricPluginsActive               = "plugins_active"
	MetricPluginExecutionsTotal       = "plugin_executions_total"
	MetricPluginExecutionDuration     = "plugin_execution_duration_seconds"
	MetricPluginErrorsTotal           = "plugin_errors_total"
	MetricPluginMemoryUsage           = "plugin_memory_usage_bytes"
	MetricPluginCPUUsage              = "plugin_cpu_usage_seconds"
	
	// 系统指标
	MetricCPUUsage                    = "cpu_usage_percent"
	MetricMemoryUsage                 = "memory_usage_bytes"
	MetricMemoryUsagePercent          = "memory_usage_percent"
	MetricDiskUsage                   = "disk_usage_bytes"
	MetricDiskUsagePercent            = "disk_usage_percent"
	MetricNetworkBytesReceived        = "network_bytes_received"
	MetricNetworkBytesSent            = "network_bytes_sent"
	MetricGoroutines                  = "goroutines"
	MetricGCDuration                  = "gc_duration_seconds"
	
	// 健康检查指标
	MetricHealthChecksTotal           = "health_checks_total"
	MetricHealthCheckDuration         = "health_check_duration_seconds"
	MetricHealthCheckStatus           = "health_check_status"
	MetricHealthyServices             = "healthy_services"
	MetricUnhealthyServices           = "unhealthy_services"
)

// MetricLabel 标准标签名称
const (
	// 通用标签
	LabelService      = "service"
	LabelVersion      = "version"
	LabelEnvironment  = "environment"
	LabelRegion       = "region"
	LabelZone         = "zone"
	LabelInstance     = "instance"
	LabelMethod       = "method"
	LabelStatus       = "status"
	LabelStatusCode   = "status_code"
	LabelError        = "error"
	LabelErrorType    = "error_type"
	LabelEndpoint     = "endpoint"
	LabelHandler      = "handler"
	LabelRoute        = "route"
	LabelProtocol     = "protocol"
	
	// HTTP标签
	LabelHTTPMethod   = "http_method"
	LabelHTTPStatus   = "http_status"
	LabelHTTPRoute    = "http_route"
	LabelHTTPHandler  = "http_handler"
	LabelUserAgent    = "user_agent"
	LabelRemoteAddr   = "remote_addr"
	
	// 数据库标签
	LabelDBName       = "db_name"
	LabelDBTable      = "db_table"
	LabelDBOperation  = "db_operation"
	LabelDBDriver     = "db_driver"
	LabelQueryType    = "query_type"
	
	// 缓存标签
	LabelCacheType    = "cache_type"
	LabelCacheKey     = "cache_key"
	LabelCacheOperation = "cache_operation"
	
	// 服务发现标签
	LabelServiceName  = "service_name"
	LabelServiceType  = "service_type"
	LabelHealthStatus = "health_status"
	LabelCheckType    = "check_type"
	LabelCheckName    = "check_name"
	
	// 认证标签
	LabelAuthType     = "auth_type"
	LabelAuthProvider = "auth_provider"
	LabelUserID       = "user_id"
	LabelRole         = "role"
	LabelPermission   = "permission"
	
	// 插件标签
	LabelPluginName   = "plugin_name"
	LabelPluginType   = "plugin_type"
	LabelPluginVersion = "plugin_version"
	LabelPluginStatus = "plugin_status"
	
	// 系统标签
	LabelCPUCore      = "cpu_core"
	LabelDiskDevice   = "disk_device"
	LabelNetworkInterface = "network_interface"
)

// MetricHelp 指标帮助信息
var MetricHelp = map[string]string{
	MetricRequestsTotal:        "Total number of requests processed",
	MetricRequestDuration:      "Duration of requests in seconds",
	MetricRequestSize:          "Size of requests in bytes",
	MetricResponseSize:         "Size of responses in bytes",
	MetricErrorsTotal:          "Total number of errors",
	MetricActiveConnections:    "Number of active connections",
	MetricConnectionsTotal:     "Total number of connections",
	
	MetricHTTPRequestsTotal:           "Total number of HTTP requests",
	MetricHTTPRequestDuration:         "Duration of HTTP requests in seconds",
	MetricHTTPRequestSize:             "Size of HTTP requests in bytes",
	MetricHTTPResponseSize:            "Size of HTTP responses in bytes",
	MetricHTTPActiveRequests:          "Number of active HTTP requests",
	MetricHTTPConcurrentConnections:   "Number of concurrent HTTP connections",
	
	MetricDBConnectionsOpen:           "Number of open database connections",
	MetricDBConnectionsIdle:           "Number of idle database connections",
	MetricDBConnectionsInUse:          "Number of database connections in use",
	MetricDBConnectionsWaitCount:      "Number of database connection waits",
	MetricDBConnectionsWaitDuration:   "Duration of database connection waits in seconds",
	MetricDBQueriesTotal:              "Total number of database queries",
	MetricDBQueryDuration:             "Duration of database queries in seconds",
	MetricDBTransactionsTotal:         "Total number of database transactions",
	MetricDBTransactionDuration:       "Duration of database transactions in seconds",
	
	MetricCacheHitsTotal:              "Total number of cache hits",
	MetricCacheMissesTotal:            "Total number of cache misses",
	MetricCacheOperationsTotal:        "Total number of cache operations",
	MetricCacheOperationDuration:      "Duration of cache operations in seconds",
	MetricCacheSize:                   "Size of cache in bytes",
	MetricCacheEntries:                "Number of cache entries",
	MetricCacheEvictionsTotal:         "Total number of cache evictions",
	
	MetricServiceRegistrationsTotal:   "Total number of service registrations",
	MetricServiceDeregistrationsTotal: "Total number of service deregistrations",
	MetricServiceDiscoveryTotal:       "Total number of service discoveries",
	MetricServiceHealthChecksTotal:    "Total number of service health checks",
	MetricServiceHealthCheckDuration:  "Duration of service health checks in seconds",
	MetricActiveServices:              "Number of active services",
	MetricServiceInstances:            "Number of service instances",
	
	MetricAuthAttemptsTotal:           "Total number of authentication attempts",
	MetricAuthSuccessTotal:            "Total number of successful authentications",
	MetricAuthFailuresTotal:           "Total number of failed authentications",
	MetricAuthTokensActive:            "Number of active authentication tokens",
	MetricAuthTokensIssued:            "Number of issued authentication tokens",
	MetricAuthTokensExpired:           "Number of expired authentication tokens",
	MetricAuthSessionsActive:          "Number of active authentication sessions",
	
	MetricPluginsLoaded:               "Number of loaded plugins",
	MetricPluginsActive:               "Number of active plugins",
	MetricPluginExecutionsTotal:       "Total number of plugin executions",
	MetricPluginExecutionDuration:     "Duration of plugin executions in seconds",
	MetricPluginErrorsTotal:           "Total number of plugin errors",
	MetricPluginMemoryUsage:           "Memory usage of plugins in bytes",
	MetricPluginCPUUsage:              "CPU usage of plugins in seconds",
	
	MetricCPUUsage:                    "CPU usage percentage",
	MetricMemoryUsage:                 "Memory usage in bytes",
	MetricMemoryUsagePercent:          "Memory usage percentage",
	MetricDiskUsage:                   "Disk usage in bytes",
	MetricDiskUsagePercent:            "Disk usage percentage",
	MetricNetworkBytesReceived:        "Network bytes received",
	MetricNetworkBytesSent:            "Network bytes sent",
	MetricGoroutines:                  "Number of goroutines",
	MetricGCDuration:                  "Garbage collection duration in seconds",
	
	MetricHealthChecksTotal:           "Total number of health checks",
	MetricHealthCheckDuration:         "Duration of health checks in seconds",
	MetricHealthCheckStatus:           "Status of health checks",
	MetricHealthyServices:             "Number of healthy services",
	MetricUnhealthyServices:           "Number of unhealthy services",
}

// DefaultMetricBuckets 默认指标桶配置
var DefaultMetricBuckets = map[string][]float64{
	MetricRequestDuration:             {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	MetricHTTPRequestDuration:         {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	MetricDBQueryDuration:             {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	MetricDBTransactionDuration:       {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	MetricCacheOperationDuration:      {0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	MetricServiceHealthCheckDuration:  {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	MetricPluginExecutionDuration:     {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	MetricHealthCheckDuration:         {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	MetricGCDuration:                  {0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
}

// DefaultMetricSizeBuckets 默认大小指标桶配置
var DefaultMetricSizeBuckets = map[string][]float64{
	MetricRequestSize:     {100, 1000, 10000, 100000, 1000000, 10000000},
	MetricResponseSize:    {100, 1000, 10000, 100000, 1000000, 10000000},
	MetricHTTPRequestSize: {100, 1000, 10000, 100000, 1000000, 10000000},
	MetricHTTPResponseSize: {100, 1000, 10000, 100000, 1000000, 10000000},
	MetricCacheSize:       {1024, 10240, 102400, 1048576, 10485760, 104857600},
	MetricPluginMemoryUsage: {1048576, 10485760, 104857600, 1073741824, 10737418240},
	MetricMemoryUsage:     {1048576, 10485760, 104857600, 1073741824, 10737418240},
	MetricDiskUsage:       {1073741824, 10737418240, 107374182400, 1099511627776},
}

// AlertThresholds 告警阈值配置
type AlertThresholds struct {
	// HTTP相关阈值
	HTTPErrorRateWarning    float64       `yaml:"http_error_rate_warning" default:"0.05"`
	HTTPErrorRateCritical   float64       `yaml:"http_error_rate_critical" default:"0.1"`
	HTTPLatencyWarning      time.Duration `yaml:"http_latency_warning" default:"1s"`
	HTTPLatencyCritical     time.Duration `yaml:"http_latency_critical" default:"5s"`
	
	// 数据库相关阈值
	DBConnectionsWarning    int           `yaml:"db_connections_warning" default:"80"`
	DBConnectionsCritical   int           `yaml:"db_connections_critical" default:"95"`
	DBQueryLatencyWarning   time.Duration `yaml:"db_query_latency_warning" default:"1s"`
	DBQueryLatencyCritical  time.Duration `yaml:"db_query_latency_critical" default:"5s"`
	
	// 缓存相关阈值
	CacheHitRateWarning     float64       `yaml:"cache_hit_rate_warning" default:"0.8"`
	CacheHitRateCritical    float64       `yaml:"cache_hit_rate_critical" default:"0.6"`
	
	// 系统相关阈值
	CPUUsageWarning         float64       `yaml:"cpu_usage_warning" default:"80"`
	CPUUsageCritical        float64       `yaml:"cpu_usage_critical" default:"90"`
	MemoryUsageWarning      float64       `yaml:"memory_usage_warning" default:"80"`
	MemoryUsageCritical     float64       `yaml:"memory_usage_critical" default:"90"`
	DiskUsageWarning        float64       `yaml:"disk_usage_warning" default:"80"`
	DiskUsageCritical       float64       `yaml:"disk_usage_critical" default:"90"`
	
	// 服务相关阈值
	ServiceHealthWarning    float64       `yaml:"service_health_warning" default:"0.9"`
	ServiceHealthCritical   float64       `yaml:"service_health_critical" default:"0.8"`
}

// DefaultAlertThresholds 默认告警阈值
var DefaultAlertThresholds = AlertThresholds{
	HTTPErrorRateWarning:    0.05,
	HTTPErrorRateCritical:   0.1,
	HTTPLatencyWarning:      1 * time.Second,
	HTTPLatencyCritical:     5 * time.Second,
	DBConnectionsWarning:    80,
	DBConnectionsCritical:   95,
	DBQueryLatencyWarning:   1 * time.Second,
	DBQueryLatencyCritical:  5 * time.Second,
	CacheHitRateWarning:     0.8,
	CacheHitRateCritical:    0.6,
	CPUUsageWarning:         80,
	CPUUsageCritical:        90,
	MemoryUsageWarning:      80,
	MemoryUsageCritical:     90,
	DiskUsageWarning:        80,
	DiskUsageCritical:       90,
	ServiceHealthWarning:    0.9,
	ServiceHealthCritical:   0.8,
}