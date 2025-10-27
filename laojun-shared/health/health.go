package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Status 健康状态
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
	StatusUnknown   Status = "unknown"
)

// CheckResult 检查结�?
type CheckResult struct {
	Name      string            `json:"name"`
	Status    Status            `json:"status"`
	Message   string            `json:"message,omitempty"`
	Error     string            `json:"error,omitempty"`
	Duration  time.Duration     `json:"duration"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// HealthReport 健康报告
type HealthReport struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Service   ServiceInfo            `json:"service"`
	Checks    map[string]CheckResult `json:"checks"`
	Summary   Summary                `json:"summary"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Environment string        `json:"environment"`
	StartTime   time.Time     `json:"start_time"`
	Uptime      time.Duration `json:"uptime"`
}

// Summary 摘要信息
type Summary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Degraded  int `json:"degraded"`
	Unknown   int `json:"unknown"`
}

// Checker 健康检查器接口
type Checker interface {
	Check(ctx context.Context) CheckResult
	Name() string
}

// Config 健康检查配�?
type Config struct {
	Enabled     bool          `yaml:"enabled" env:"HEALTH_ENABLED" config:"health.enabled" default:"true"`
	Path        string        `yaml:"path" env:"HEALTH_PATH" config:"health.path" default:"/health"`
	Timeout     time.Duration `yaml:"timeout" env:"HEALTH_TIMEOUT" config:"health.timeout" default:"30s"`
	Service     string        `yaml:"service" env:"SERVICE_NAME" config:"health.service" default:"unknown"`
	Version     string        `yaml:"version" env:"SERVICE_VERSION" config:"health.version" default:"unknown"`
	Environment string        `yaml:"environment" env:"ENVIRONMENT" config:"health.environment" default:"development"`
}

// Health 健康检查管理器
type Health struct {
	config    Config
	checkers  map[string]Checker
	startTime time.Time
	mu        sync.RWMutex
}

// New 创建健康检查管理器
func New(config Config) *Health {
	return &Health{
		config:    config,
		checkers:  make(map[string]Checker),
		startTime: time.Now(),
	}
}

// AddChecker 添加检查器
func (h *Health) AddChecker(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[checker.Name()] = checker
}

// RemoveChecker 移除检查器
func (h *Health) RemoveChecker(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checkers, name)
}

// Check 执行健康检�?
func (h *Health) Check(ctx context.Context) HealthReport {
	start := time.Now()

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, h.config.Timeout)
	defer cancel()

	h.mu.RLock()
	checkers := make(map[string]Checker, len(h.checkers))
	for name, checker := range h.checkers {
		checkers[name] = checker
	}
	h.mu.RUnlock()

	// 并发执行检�?
	results := make(map[string]CheckResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker Checker) {
			defer wg.Done()

			result := checker.Check(ctx)

			mu.Lock()
			results[name] = result
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()

	// 计算总体状�?
	overallStatus := h.calculateOverallStatus(results)

	// 计算摘要
	summary := h.calculateSummary(results)

	return HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Service: ServiceInfo{
			Name:        h.config.Service,
			Version:     h.config.Version,
			Environment: h.config.Environment,
			StartTime:   h.startTime,
			Uptime:      time.Since(h.startTime),
		},
		Checks:  results,
		Summary: summary,
	}
}

// calculateOverallStatus 计算总体状�?
func (h *Health) calculateOverallStatus(results map[string]CheckResult) Status {
	if len(results) == 0 {
		return StatusHealthy
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, result := range results {
		switch result.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		case StatusUnknown:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}

	return StatusHealthy
}

// calculateSummary 计算摘要
func (h *Health) calculateSummary(results map[string]CheckResult) Summary {
	summary := Summary{
		Total: len(results),
	}

	for _, result := range results {
		switch result.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusUnhealthy:
			summary.Unhealthy++
		case StatusDegraded:
			summary.Degraded++
		case StatusUnknown:
			summary.Unknown++
		}
	}

	return summary
}

// Handler 获取 HTTP 处理函数
func (h *Health) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := h.Check(r.Context())

		w.Header().Set("Content-Type", "application/json")

		// 根据状态设�?HTTP 状态码
		switch report.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // 降级但仍可用
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(report)
	}
}

// GinHandler 获取 Gin 处理函数
func (h *Health) GinHandler() gin.HandlerFunc {
	return gin.WrapH(h.Handler())
}

// 内置检查器

// DatabaseChecker 数据库检查器
type DatabaseChecker struct {
	name string
	db   *sql.DB
}

// NewDatabaseChecker 创建数据库检查器
func NewDatabaseChecker(name string, db *sql.DB) *DatabaseChecker {
	return &DatabaseChecker{
		name: name,
		db:   db,
	}
}

func (c *DatabaseChecker) Name() string {
	return c.name
}

func (c *DatabaseChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
	}

	// 检查数据库连接
	if err := c.db.PingContext(ctx); err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Database connection failed"
	} else {
		result.Status = StatusHealthy
		result.Message = "Database connection successful"

		// 获取连接统计
		stats := c.db.Stats()
		result.Metadata["open_connections"] = fmt.Sprintf("%d", stats.OpenConnections)
		result.Metadata["in_use"] = fmt.Sprintf("%d", stats.InUse)
		result.Metadata["idle"] = fmt.Sprintf("%d", stats.Idle)
	}

	result.Duration = time.Since(start)
	return result
}

// RedisChecker Redis 检查器
type RedisChecker struct {
	name   string
	client *redis.Client
}

// NewRedisChecker 创建 Redis 检查器
func NewRedisChecker(name string, client *redis.Client) *RedisChecker {
	return &RedisChecker{
		name:   name,
		client: client,
	}
}

func (c *RedisChecker) Name() string {
	return c.name
}

func (c *RedisChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
	}

	// 检�?Redis 连接
	if err := c.client.Ping(ctx).Err(); err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis connection failed"
	} else {
		result.Status = StatusHealthy
		result.Message = "Redis connection successful"

		// 获取 Redis 信息
		if info, err := c.client.Info(ctx, "memory").Result(); err == nil {
			result.Metadata["info"] = info
		}
	}

	result.Duration = time.Since(start)
	return result
}

// HTTPChecker HTTP 服务检查器
type HTTPChecker struct {
	name   string
	url    string
	client *http.Client
}

// NewHTTPChecker 创建 HTTP 检查器
func NewHTTPChecker(name, url string) *HTTPChecker {
	return &HTTPChecker{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *HTTPChecker) Name() string {
	return c.name
}

func (c *HTTPChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  make(map[string]string),
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Failed to create request"
		result.Duration = time.Since(start)
		return result
	}

	// 发送请�?
	resp, err := c.client.Do(req)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "HTTP request failed"
	} else {
		defer resp.Body.Close()

		result.Metadata["status_code"] = fmt.Sprintf("%d", resp.StatusCode)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result.Status = StatusHealthy
			result.Message = "HTTP service is healthy"
		} else if resp.StatusCode >= 500 {
			result.Status = StatusUnhealthy
			result.Message = "HTTP service returned server error"
		} else {
			result.Status = StatusDegraded
			result.Message = "HTTP service returned client error"
		}
	}

	result.Duration = time.Since(start)
	return result
}

// CustomChecker 自定义检查器
type CustomChecker struct {
	name     string
	checkFn  func(ctx context.Context) (Status, string, error)
	metadata map[string]string
}

// NewCustomChecker 创建自定义检查器
func NewCustomChecker(name string, checkFn func(ctx context.Context) (Status, string, error)) *CustomChecker {
	return &CustomChecker{
		name:     name,
		checkFn:  checkFn,
		metadata: make(map[string]string),
	}
}

func (c *CustomChecker) Name() string {
	return c.name
}

func (c *CustomChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	result := CheckResult{
		Name:      c.name,
		Timestamp: start,
		Metadata:  c.metadata,
	}

	status, message, err := c.checkFn(ctx)
	result.Status = status
	result.Message = message
	if err != nil {
		result.Error = err.Error()
	}

	result.Duration = time.Since(start)
	return result
}

// SetMetadata 设置元数�?
func (c *CustomChecker) SetMetadata(key, value string) {
	c.metadata[key] = value
}

// DefaultHealth 默认健康检查实�?
var DefaultHealth *Health

// init 初始化默认健康检�?
func init() {
	DefaultHealth = New(Config{
		Enabled:     true,
		Path:        "/health",
		Timeout:     30 * time.Second,
		Service:     "unknown",
		Version:     "unknown",
		Environment: "development",
	})
}
