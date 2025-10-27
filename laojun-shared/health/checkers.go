package health

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// BaseChecker 基础检查器
type BaseChecker struct {
	name        string
	checkerType CheckerType
	priority    Priority
	config      CheckerConfig
}

// NewBaseChecker 创建基础检查器
func NewBaseChecker(name string, checkerType CheckerType, priority Priority) *BaseChecker {
	return &BaseChecker{
		name:        name,
		checkerType: checkerType,
		priority:    priority,
		config: CheckerConfig{
			Name:     name,
			Type:     checkerType,
			Priority: priority,
			Enabled:  true,
			Timeout:  30 * time.Second,
			Retries:  3,
			Metadata: make(map[string]string),
			Tags:     make([]string, 0),
		},
	}
}

func (c *BaseChecker) Name() string {
	return c.name
}

func (c *BaseChecker) Type() CheckerType {
	return c.checkerType
}

func (c *BaseChecker) Priority() Priority {
	return c.priority
}

func (c *BaseChecker) Config() CheckerConfig {
	return c.config
}

func (c *BaseChecker) SetConfig(config CheckerConfig) {
	c.config = config
}

// EnhancedDatabaseChecker 增强的数据库检查器
type EnhancedDatabaseChecker struct {
	*BaseChecker
	db             *sql.DB
	queryTimeout   time.Duration
	testQuery      string
	maxConnections int
}

// NewEnhancedDatabaseChecker 创建增强的数据库检查器
func NewEnhancedDatabaseChecker(name string, db *sql.DB) *EnhancedDatabaseChecker {
	return &EnhancedDatabaseChecker{
		BaseChecker:    NewBaseChecker(name, CheckerTypeDatabase, PriorityHigh),
		db:             db,
		queryTimeout:   5 * time.Second,
		testQuery:      "SELECT 1",
		maxConnections: 100,
	}
}

func (c *EnhancedDatabaseChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
	}

	// 设置查询超时
	queryCtx, cancel := context.WithTimeout(ctx, c.queryTimeout)
	defer cancel()

	// 检查数据库连接
	if err := c.db.PingContext(queryCtx); err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Database ping failed"
		result.Duration = time.Since(start)
		return result
	}

	// 执行测试查询
	var testResult int
	if err := c.db.QueryRowContext(queryCtx, c.testQuery).Scan(&testResult); err != nil {
		result.Status = StatusDegraded
		result.Error = err.Error()
		result.Message = "Database test query failed"
	} else {
		result.Status = StatusHealthy
		result.Message = "Database connection successful"
	}

	// 获取连接统计
	stats := c.db.Stats()
	result.Metadata["open_connections"] = fmt.Sprintf("%d", stats.OpenConnections)
	result.Metadata["in_use"] = fmt.Sprintf("%d", stats.InUse)
	result.Metadata["idle"] = fmt.Sprintf("%d", stats.Idle)
	result.Metadata["max_open_connections"] = fmt.Sprintf("%d", stats.MaxOpenConnections)
	result.Metadata["wait_count"] = fmt.Sprintf("%d", stats.WaitCount)
	result.Metadata["wait_duration"] = stats.WaitDuration.String()

	// 检查连接池状态
	if stats.OpenConnections > c.maxConnections {
		result.Status = StatusDegraded
		result.Message = "Database connection pool near limit"
	}

	result.Duration = time.Since(start)
	return result
}

// EnhancedRedisChecker 增强的Redis检查器
type EnhancedRedisChecker struct {
	*BaseChecker
	client      *redis.Client
	testKey     string
	testValue   string
	pingTimeout time.Duration
}

// NewEnhancedRedisChecker 创建增强的Redis检查器
func NewEnhancedRedisChecker(name string, client *redis.Client) *EnhancedRedisChecker {
	return &EnhancedRedisChecker{
		BaseChecker: NewBaseChecker(name, CheckerTypeCache, PriorityMedium),
		client:      client,
		testKey:     "health_check_test",
		testValue:   "ok",
		pingTimeout: 5 * time.Second,
	}
}

func (c *EnhancedRedisChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
	}

	// 设置超时
	pingCtx, cancel := context.WithTimeout(ctx, c.pingTimeout)
	defer cancel()

	// 检查Redis连接
	if err := c.client.Ping(pingCtx).Err(); err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis ping failed"
		result.Duration = time.Since(start)
		return result
	}

	// 测试读写操作
	if err := c.client.Set(pingCtx, c.testKey, c.testValue, time.Minute).Err(); err != nil {
		result.Status = StatusDegraded
		result.Error = err.Error()
		result.Message = "Redis write test failed"
	} else if val, err := c.client.Get(pingCtx, c.testKey).Result(); err != nil {
		result.Status = StatusDegraded
		result.Error = err.Error()
		result.Message = "Redis read test failed"
	} else if val != c.testValue {
		result.Status = StatusDegraded
		result.Message = "Redis read/write test value mismatch"
	} else {
		result.Status = StatusHealthy
		result.Message = "Redis connection successful"
		
		// 清理测试键
		c.client.Del(pingCtx, c.testKey)
	}

	// 获取Redis信息
	if info, err := c.client.Info(pingCtx, "memory", "clients", "stats").Result(); err == nil {
		result.Metadata["redis_info"] = info
	}

	// 获取连接池统计
	poolStats := c.client.PoolStats()
	result.Metadata["pool_hits"] = fmt.Sprintf("%d", poolStats.Hits)
	result.Metadata["pool_misses"] = fmt.Sprintf("%d", poolStats.Misses)
	result.Metadata["pool_timeouts"] = fmt.Sprintf("%d", poolStats.Timeouts)
	result.Metadata["pool_total_conns"] = fmt.Sprintf("%d", poolStats.TotalConns)
	result.Metadata["pool_idle_conns"] = fmt.Sprintf("%d", poolStats.IdleConns)

	result.Duration = time.Since(start)
	return result
}

// EnhancedHTTPChecker 增强的HTTP检查器
type EnhancedHTTPChecker struct {
	*BaseChecker
	url            string
	method         string
	expectedStatus int
	expectedBody   string
	headers        map[string]string
	client         *http.Client
}

// NewEnhancedHTTPChecker 创建增强的HTTP检查器
func NewEnhancedHTTPChecker(name, url string) *EnhancedHTTPChecker {
	return &EnhancedHTTPChecker{
		BaseChecker:    NewBaseChecker(name, CheckerTypeHTTP, PriorityMedium),
		url:            url,
		method:         "GET",
		expectedStatus: 200,
		headers:        make(map[string]string),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *EnhancedHTTPChecker) SetMethod(method string) *EnhancedHTTPChecker {
	c.method = method
	return c
}

func (c *EnhancedHTTPChecker) SetExpectedStatus(status int) *EnhancedHTTPChecker {
	c.expectedStatus = status
	return c
}

func (c *EnhancedHTTPChecker) SetExpectedBody(body string) *EnhancedHTTPChecker {
	c.expectedBody = body
	return c
}

func (c *EnhancedHTTPChecker) SetHeader(key, value string) *EnhancedHTTPChecker {
	c.headers[key] = value
	return c
}

func (c *EnhancedHTTPChecker) SetTimeout(timeout time.Duration) *EnhancedHTTPChecker {
	c.client.Timeout = timeout
	return c
}

func (c *EnhancedHTTPChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, c.method, c.url, nil)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Failed to create HTTP request"
		result.Duration = time.Since(start)
		return result
	}

	// 设置请求头
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "HTTP request failed"
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 记录响应信息
	result.Metadata["status_code"] = fmt.Sprintf("%d", resp.StatusCode)
	result.Metadata["content_length"] = fmt.Sprintf("%d", resp.ContentLength)
	result.Metadata["content_type"] = resp.Header.Get("Content-Type")

	// 检查状态码
	if resp.StatusCode != c.expectedStatus {
		if resp.StatusCode >= 500 {
			result.Status = StatusUnhealthy
			result.Message = fmt.Sprintf("HTTP service returned server error: %d", resp.StatusCode)
		} else {
			result.Status = StatusDegraded
			result.Message = fmt.Sprintf("HTTP service returned unexpected status: %d", resp.StatusCode)
		}
	} else {
		result.Status = StatusHealthy
		result.Message = "HTTP service is healthy"
	}

	// 检查响应体（如果指定了期望的响应体）
	if c.expectedBody != "" {
		body := make([]byte, 1024) // 限制读取大小
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])
		
		if bodyStr != c.expectedBody {
			result.Status = StatusDegraded
			result.Message = "HTTP response body mismatch"
			result.Metadata["expected_body"] = c.expectedBody
			result.Metadata["actual_body"] = bodyStr
		}
	}

	result.Duration = time.Since(start)
	return result
}

// SystemChecker 系统资源检查器
type SystemChecker struct {
	*BaseChecker
	cpuThreshold    float64
	memoryThreshold float64
	diskThreshold   float64
}

// NewSystemChecker 创建系统检查器
func NewSystemChecker(name string) *SystemChecker {
	return &SystemChecker{
		BaseChecker:     NewBaseChecker(name, CheckerTypeSystem, PriorityLow),
		cpuThreshold:    80.0,
		memoryThreshold: 85.0,
		diskThreshold:   90.0,
	}
}

func (c *SystemChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
		Status:    StatusHealthy,
		Message:   "System resources are healthy",
	}

	// 这里可以实现具体的系统资源检查逻辑
	// 由于涉及到系统调用，这里只提供框架
	
	result.Duration = time.Since(start)
	return result
}

// ApplicationChecker 应用程序检查器
type ApplicationChecker struct {
	*BaseChecker
	checkFunc func(ctx context.Context) (Status, string, map[string]string, error)
}

// NewApplicationChecker 创建应用程序检查器
func NewApplicationChecker(name string, checkFunc func(ctx context.Context) (Status, string, map[string]string, error)) *ApplicationChecker {
	return &ApplicationChecker{
		BaseChecker: NewBaseChecker(name, CheckerTypeApplication, PriorityMedium),
		checkFunc:   checkFunc,
	}
}

func (c *ApplicationChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
	}

	if c.checkFunc != nil {
		status, message, metadata, err := c.checkFunc(ctx)
		result.Status = status
		result.Message = message
		if metadata != nil {
			result.Metadata = metadata
		}
		if err != nil {
			result.Error = err.Error()
		}
	} else {
		result.Status = StatusUnknown
		result.Message = "No check function provided"
	}

	result.Duration = time.Since(start)
	return result
}