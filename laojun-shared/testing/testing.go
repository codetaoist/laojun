package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestSuite 基础测试套件
type TestSuite struct {
	suite.Suite
	DB           *gorm.DB
	Redis        *redis.Client
	Router       *gin.Engine
	TestData     map[string]interface{}
	CleanupFuncs []func()
}

// SetupSuite 测试套件初始化
func (s *TestSuite) SetupSuite() {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化测试数据存储
	s.TestData = make(map[string]interface{})
	s.CleanupFuncs = make([]func(), 0)
}

// TearDownSuite 测试套件清理
func (s *TestSuite) TearDownSuite() {
	// 执行清理函数
	for i := len(s.CleanupFuncs) - 1; i >= 0; i-- {
		s.CleanupFuncs[i]()
	}

	// 关闭数据库连接
	if s.DB != nil {
		if sqlDB, err := s.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// 关闭 Redis 连接
	if s.Redis != nil {
		s.Redis.Close()
	}
}

// SetupTest 每个测试前的初始化
func (s *TestSuite) SetupTest() {
	// 清理测试数据
	if s.DB != nil {
		s.CleanupDatabase()
	}
	if s.Redis != nil {
		s.CleanupRedis()
	}
}

// TearDownTest 每个测试后的清理
func (s *TestSuite) TearDownTest() {
	// 清理测试数据
	if s.DB != nil {
		s.CleanupDatabase()
	}
	if s.Redis != nil {
		s.CleanupRedis()
	}
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string // sqlite, postgres
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// SetupDatabase 设置测试数据库
func (s *TestSuite) SetupDatabase(config DatabaseConfig) {
	var dialector gorm.Dialector

	switch config.Driver {
	case "sqlite":
		// 使用内存数据库进行测试
		dialector = sqlite.Open(":memory:")
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
		dialector = postgres.Open(dsn)
	default:
		s.T().Fatalf("Unsupported database driver: %s", config.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 静默模式，减少测试输出
	})
	require.NoError(s.T(), err)

	s.DB = db

	// 添加清理函数
	s.AddCleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// SetupRedis 设置测试 Redis
func (s *TestSuite) SetupRedis(config RedisConfig) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	require.NoError(s.T(), err)

	s.Redis = client

	// 添加清理函数
	s.AddCleanup(func() {
		client.Close()
	})
}

// SetupRouter 设置测试路由
func (s *TestSuite) SetupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	s.Router = router
	return router
}

// AddCleanup 添加清理函数
func (s *TestSuite) AddCleanup(cleanup func()) {
	s.CleanupFuncs = append(s.CleanupFuncs, cleanup)
}

// CleanupDatabase 清理测试数据库
func (s *TestSuite) CleanupDatabase() {
	if s.DB == nil {
		return
	}

	// 获取所有表名
	var tables []string
	s.DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)

	// 清空所有表
	for _, table := range tables {
		s.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}
}

// CleanupRedis 清理 Redis
func (s *TestSuite) CleanupRedis() {
	if s.Redis == nil {
		return
	}

	ctx := context.Background()
	s.Redis.FlushDB(ctx)
}

// HTTPTestHelper HTTP 测试辅助工具
type HTTPTestHelper struct {
	Router *gin.Engine
	T      *testing.T
}

// NewHTTPTestHelper 创建 HTTP 测试辅助工具
func NewHTTPTestHelper(router *gin.Engine, t *testing.T) *HTTPTestHelper {
	return &HTTPTestHelper{
		Router: router,
		T:      t,
	}
}

// Request HTTP 请求结构
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
}

// Response HTTP 响应结构
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	JSON       map[string]interface{}
}

// DoRequest 执行 HTTP 请求
func (h *HTTPTestHelper) DoRequest(req Request) *Response {
	var body io.Reader

	// 处理请求体
	if req.Body != nil {
		switch v := req.Body.(type) {
		case string:
			body = strings.NewReader(v)
		case []byte:
			body = bytes.NewReader(v)
		default:
			jsonData, err := json.Marshal(v)
			require.NoError(h.T, err)
			body = bytes.NewReader(jsonData)
		}
	}

	// 创建请求
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	require.NoError(h.T, err)

	// 设置请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 如果是JSON 请求，设置Content-Type
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 执行请求
	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, httpReq)

	// 构建响应
	response := &Response{
		StatusCode: recorder.Code,
		Headers:    make(map[string]string),
		Body:       recorder.Body.String(),
	}

	// 复制响应头
	for key, values := range recorder.Header() {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	// 解析 JSON 响应
	if strings.Contains(response.Headers["Content-Type"], "application/json") {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(response.Body), &jsonData); err == nil {
			response.JSON = jsonData
		}
	}

	return response
}

// GET 执行 GET 请求
func (h *HTTPTestHelper) GET(url string, headers ...map[string]string) *Response {
	req := Request{Method: "GET", URL: url}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return h.DoRequest(req)
}

// POST 执行 POST 请求
func (h *HTTPTestHelper) POST(url string, body interface{}, headers ...map[string]string) *Response {
	req := Request{Method: "POST", URL: url, Body: body}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return h.DoRequest(req)
}

// PUT 执行 PUT 请求
func (h *HTTPTestHelper) PUT(url string, body interface{}, headers ...map[string]string) *Response {
	req := Request{Method: "PUT", URL: url, Body: body}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return h.DoRequest(req)
}

// DELETE 执行 DELETE 请求
func (h *HTTPTestHelper) DELETE(url string, headers ...map[string]string) *Response {
	req := Request{Method: "DELETE", URL: url}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return h.DoRequest(req)
}

// AssertJSON 断言 JSON 响应
func (h *HTTPTestHelper) AssertJSON(response *Response, expected map[string]interface{}) {
	assert.Equal(h.T, expected, response.JSON)
}

// AssertStatus 断言状态码
func (h *HTTPTestHelper) AssertStatus(response *Response, expectedStatus int) {
	assert.Equal(h.T, expectedStatus, response.StatusCode)
}

// AssertContains 断言响应体包含指定内容
func (h *HTTPTestHelper) AssertContains(response *Response, expected string) {
	assert.Contains(h.T, response.Body, expected)
}

// TestDataLoader 测试数据加载器
type TestDataLoader struct {
	DataDir string
}

// NewTestDataLoader 创建测试数据加载器
func NewTestDataLoader(dataDir string) *TestDataLoader {
	return &TestDataLoader{DataDir: dataDir}
}

// LoadJSON 加载 JSON 测试数据
func (l *TestDataLoader) LoadJSON(filename string, target interface{}) error {
	filePath := filepath.Join(l.DataDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// LoadSQL 加载 SQL 测试数据
func (l *TestDataLoader) LoadSQL(filename string) (string, error) {
	filePath := filepath.Join(l.DataDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MockHelper 模拟辅助工具
type MockHelper struct {
	T *testing.T
}

// NewMockHelper 创建模拟辅助工具
func NewMockHelper(t *testing.T) *MockHelper {
	return &MockHelper{T: t}
}

// MockHTTPServer 创建模拟 HTTP 服务器
func (m *MockHelper) MockHTTPServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	return server
}

// MockDatabase 创建模拟数据库
func (m *MockHelper) MockDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(m.T, err)
	return db
}

// MockRedis 创建模拟 Redis 客户端
func (m *MockHelper) MockRedis() *redis.Client {
	// 注意：这里需要引入miniredis 库
	// 为了简化，这里返回一个连接到测试 Redis 实例的客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // 使用专门的测试数据库
	})
	return client
}

// BenchmarkHelper 性能测试辅助工具
type BenchmarkHelper struct {
	B *testing.B
}

// NewBenchmarkHelper 创建性能测试辅助工具
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	return &BenchmarkHelper{B: b}
}

// TimeOperation 测量操作执行时间
func (h *BenchmarkHelper) TimeOperation(operation func()) time.Duration {
	start := time.Now()
	operation()
	return time.Since(start)
}

// BenchmarkHTTP 性能测试 HTTP 请求
func (h *BenchmarkHelper) BenchmarkHTTP(router *gin.Engine, req Request) {
	helper := NewHTTPTestHelper(router, nil)

	h.B.ResetTimer()
	for i := 0; i < h.B.N; i++ {
		helper.DoRequest(req)
	}
}

// TestEnvironment 测试环境配置
type TestEnvironment struct {
	Name        string
	DatabaseURL string
	RedisURL    string
	ConfigPath  string
}

// GetTestEnvironment 获取测试环境配置
func GetTestEnvironment() *TestEnvironment {
	env := os.Getenv("TEST_ENV")
	if env == "" {
		env = "unit"
	}

	switch env {
	case "unit":
		return &TestEnvironment{
			Name:        "unit",
			DatabaseURL: ":memory:",
			RedisURL:    "redis://localhost:6379/15",
			ConfigPath:  "testdata/config/unit.yaml",
		}
	case "integration":
		return &TestEnvironment{
			Name:        "integration",
			DatabaseURL: "postgres://test:test@localhost:5432/test_db?sslmode=disable",
			RedisURL:    "redis://localhost:6379/14",
			ConfigPath:  "testdata/config/integration.yaml",
		}
	case "e2e":
		return &TestEnvironment{
			Name:        "e2e",
			DatabaseURL: "postgres://test:test@localhost:5432/e2e_db?sslmode=disable",
			RedisURL:    "redis://localhost:6379/13",
			ConfigPath:  "testdata/config/e2e.yaml",
		}
	default:
		return &TestEnvironment{
			Name:        "unit",
			DatabaseURL: ":memory:",
			RedisURL:    "redis://localhost:6379/15",
			ConfigPath:  "testdata/config/unit.yaml",
		}
	}
}

// SkipIfShort 在短测试模式下跳过测试
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}

// SkipIfCI 在CI 环境下跳过测试
func SkipIfCI(t *testing.T, reason string) {
	if os.Getenv("CI") != "" {
		t.Skipf("Skipping test in CI environment: %s", reason)
	}
}

// RequireEnv 要求环境变量存在
func RequireEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	require.NotEmpty(t, value, "Environment variable %s is required", key)
	return value
}
