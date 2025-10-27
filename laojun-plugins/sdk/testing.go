package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// TestSuite 测试套件
type TestSuite struct {
	plugin   Plugin
	context  *PluginContext
	server   *httptest.Server
	client   *http.Client
	logger   *logrus.Logger
	t        *testing.T
}

// NewTestSuite 创建测试套件
func NewTestSuite(t *testing.T, plugin Plugin) *TestSuite {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	
	return &TestSuite{
		plugin: plugin,
		context: &PluginContext{
			ID:     plugin.GetMetadata().ID,
			Config: make(map[string]interface{}),
			Logger: logger,
		},
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
		t:      t,
	}
}

// Setup 设置测试环境
func (ts *TestSuite) Setup() error {
	// 初始化插件
	if err := ts.plugin.Initialize(ts.context); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}
	
	// 启动插件
	if err := ts.plugin.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}
	
	// 如果是HTTP插件，创建测试服务器
	if httpPlugin, ok := ts.plugin.(HTTPPlugin); ok {
		mux := http.NewServeMux()
		
		// 创建路由器适配器
		router := &TestRouter{mux: mux}
		httpPlugin.RegisterRoutes(router)
		
		// 应用中间件
		handler := ts.applyMiddlewares(mux, httpPlugin.GetMiddlewares())
		ts.server = httptest.NewServer(handler)
	}
	
	return nil
}

// Teardown 清理测试环境
func (ts *TestSuite) Teardown() error {
	if ts.server != nil {
		ts.server.Close()
	}
	
	if err := ts.plugin.Stop(); err != nil {
		ts.logger.WithError(err).Warn("Failed to stop plugin")
	}
	
	if err := ts.plugin.Cleanup(); err != nil {
		ts.logger.WithError(err).Warn("Failed to cleanup plugin")
	}
	
	return nil
}

// TestRouter 测试路由器
type TestRouter struct {
	mux *http.ServeMux
}

func (r *TestRouter) GET(path string, handler HandlerFunc) {
	r.mux.HandleFunc(path, r.wrapHandler("GET", handler))
}

func (r *TestRouter) POST(path string, handler HandlerFunc) {
	r.mux.HandleFunc(path, r.wrapHandler("POST", handler))
}

func (r *TestRouter) PUT(path string, handler HandlerFunc) {
	r.mux.HandleFunc(path, r.wrapHandler("PUT", handler))
}

func (r *TestRouter) DELETE(path string, handler HandlerFunc) {
	r.mux.HandleFunc(path, r.wrapHandler("DELETE", handler))
}

func (r *TestRouter) wrapHandler(method string, handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// 读取请求体
		body, _ := io.ReadAll(req.Body)
		req.Body.Close()
		
		// 创建请求上下文
		ctx := &RequestContext{
			Request: &Request{
				Method:  req.Method,
				Path:    req.URL.Path,
				Query:   req.URL.Query(),
				Headers: make(map[string]string),
				Body:    body,
			},
			Response: &Response{
				Headers: make(map[string]string),
			},
		}
		
		// 复制请求头
		for key, values := range req.Header {
			if len(values) > 0 {
				ctx.Request.Headers[key] = values[0]
			}
		}
		
		// 调用处理器
		resp := handler(ctx)
		
		// 写入响应
		for key, value := range resp.Headers {
			w.Header().Set(key, value)
		}
		
		w.WriteHeader(resp.StatusCode)
		
		if resp.Body != nil {
			w.Write(resp.Body)
		}
	}
}

// applyMiddlewares 应用中间件
func (ts *TestSuite) applyMiddlewares(handler http.Handler, middlewares []Middleware) http.Handler {
	// 这里简化处理，实际应该正确应用中间件链
	return handler
}

// HTTPTestCase HTTP测试用例
type HTTPTestCase struct {
	Name           string
	Method         string
	Path           string
	Headers        map[string]string
	Body           interface{}
	ExpectedStatus int
	ExpectedBody   interface{}
	ExpectedHeaders map[string]string
	Validator      func(*http.Response, []byte) error
}

// RunHTTPTest 运行HTTP测试
func (ts *TestSuite) RunHTTPTest(testCase HTTPTestCase) {
	ts.t.Run(testCase.Name, func(t *testing.T) {
		if ts.server == nil {
			t.Fatal("HTTP server not started")
		}
		
		// 准备请求体
		var bodyReader io.Reader
		if testCase.Body != nil {
			if str, ok := testCase.Body.(string); ok {
				bodyReader = strings.NewReader(str)
			} else {
				bodyBytes, err := json.Marshal(testCase.Body)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
				bodyReader = bytes.NewReader(bodyBytes)
			}
		}
		
		// 创建请求
		req, err := http.NewRequest(testCase.Method, ts.server.URL+testCase.Path, bodyReader)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		// 设置请求头
		for key, value := range testCase.Headers {
			req.Header.Set(key, value)
		}
		
		// 发送请求
		resp, err := ts.client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()
		
		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		
		// 检查状态码
		if testCase.ExpectedStatus != 0 && resp.StatusCode != testCase.ExpectedStatus {
			t.Errorf("Expected status %d, got %d", testCase.ExpectedStatus, resp.StatusCode)
		}
		
		// 检查响应头
		for key, expectedValue := range testCase.ExpectedHeaders {
			actualValue := resp.Header.Get(key)
			if actualValue != expectedValue {
				t.Errorf("Expected header %s: %s, got: %s", key, expectedValue, actualValue)
			}
		}
		
		// 检查响应体
		if testCase.ExpectedBody != nil {
			if err := ts.compareResponse(testCase.ExpectedBody, respBody); err != nil {
				t.Errorf("Response body mismatch: %v", err)
			}
		}
		
		// 自定义验证器
		if testCase.Validator != nil {
			if err := testCase.Validator(resp, respBody); err != nil {
				t.Errorf("Custom validation failed: %v", err)
			}
		}
	})
}

// compareResponse 比较响应
func (ts *TestSuite) compareResponse(expected interface{}, actual []byte) error {
	switch exp := expected.(type) {
	case string:
		if string(actual) != exp {
			return fmt.Errorf("expected %q, got %q", exp, string(actual))
		}
	case []byte:
		if !bytes.Equal(actual, exp) {
			return fmt.Errorf("expected %q, got %q", string(exp), string(actual))
		}
	default:
		// JSON比较
		expectedBytes, err := json.Marshal(expected)
		if err != nil {
			return fmt.Errorf("failed to marshal expected response: %w", err)
		}
		
		var expectedJSON, actualJSON interface{}
		if err := json.Unmarshal(expectedBytes, &expectedJSON); err != nil {
			return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
		}
		
		if err := json.Unmarshal(actual, &actualJSON); err != nil {
			return fmt.Errorf("failed to unmarshal actual JSON: %w", err)
		}
		
		if !reflect.DeepEqual(expectedJSON, actualJSON) {
			return fmt.Errorf("JSON mismatch:\nexpected: %s\nactual: %s", 
				string(expectedBytes), string(actual))
		}
	}
	
	return nil
}

// EventTestCase 事件测试用例
type EventTestCase struct {
	Name      string
	Event     *Event
	Validator func(error) error
}

// RunEventTest 运行事件测试
func (ts *TestSuite) RunEventTest(testCase EventTestCase) {
	ts.t.Run(testCase.Name, func(t *testing.T) {
		eventPlugin, ok := ts.plugin.(EventPlugin)
		if !ok {
			t.Skip("Plugin does not implement EventPlugin")
		}
		
		err := eventPlugin.HandleEvent(context.Background(), testCase.Event)
		
		if testCase.Validator != nil {
			if validationErr := testCase.Validator(err); validationErr != nil {
				t.Errorf("Event validation failed: %v", validationErr)
			}
		}
	})
}

// ScheduleTestCase 定时任务测试用例
type ScheduleTestCase struct {
	Name      string
	Validator func(error) error
}

// RunScheduleTest 运行定时任务测试
func (ts *TestSuite) RunScheduleTest(testCase ScheduleTestCase) {
	ts.t.Run(testCase.Name, func(t *testing.T) {
		scheduledPlugin, ok := ts.plugin.(ScheduledPlugin)
		if !ok {
			t.Skip("Plugin does not implement ScheduledPlugin")
		}
		
		err := scheduledPlugin.Execute(context.Background())
		
		if testCase.Validator != nil {
			if validationErr := testCase.Validator(err); validationErr != nil {
				t.Errorf("Schedule validation failed: %v", validationErr)
			}
		}
	})
}

// DataTestCase 数据处理测试用例
type DataTestCase struct {
	Name           string
	Input          *DataInput
	ExpectedOutput *DataOutput
	Validator      func(*DataOutput, error) error
}

// RunDataTest 运行数据处理测试
func (ts *TestSuite) RunDataTest(testCase DataTestCase) {
	ts.t.Run(testCase.Name, func(t *testing.T) {
		dataPlugin, ok := ts.plugin.(DataPlugin)
		if !ok {
			t.Skip("Plugin does not implement DataPlugin")
		}
		
		output, err := dataPlugin.ProcessData(context.Background(), testCase.Input)
		
		if testCase.ExpectedOutput != nil && err == nil {
			if !reflect.DeepEqual(output.Data, testCase.ExpectedOutput.Data) {
				t.Errorf("Data mismatch:\nexpected: %+v\nactual: %+v", 
					testCase.ExpectedOutput.Data, output.Data)
			}
		}
		
		if testCase.Validator != nil {
			if validationErr := testCase.Validator(output, err); validationErr != nil {
				t.Errorf("Data validation failed: %v", validationErr)
			}
		}
	})
}

// ConfigTestCase 配置测试用例
type ConfigTestCase struct {
	Name          string
	Config        map[string]interface{}
	ExpectError   bool
	ErrorContains string
}

// RunConfigTest 运行配置测试
func (ts *TestSuite) RunConfigTest(testCase ConfigTestCase) {
	ts.t.Run(testCase.Name, func(t *testing.T) {
		configurablePlugin, ok := ts.plugin.(ConfigurablePlugin)
		if !ok {
			t.Skip("Plugin does not implement ConfigurablePlugin")
		}
		
		err := configurablePlugin.ValidateConfig(testCase.Config)
		
		if testCase.ExpectError {
			if err == nil {
				t.Error("Expected error but got none")
			} else if testCase.ErrorContains != "" && !strings.Contains(err.Error(), testCase.ErrorContains) {
				t.Errorf("Expected error to contain %q, got: %v", testCase.ErrorContains, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

// MockPlugin 模拟插件
type MockPlugin struct {
	*BasePlugin
	initFunc    func(*PluginContext) error
	startFunc   func() error
	stopFunc    func() error
	cleanupFunc func() error
	eventFunc   func(context.Context, *Event) error
}

// NewMockPlugin 创建模拟插件
func NewMockPlugin(id, name string) *MockPlugin {
	return &MockPlugin{
		BasePlugin: &BasePlugin{
			info: &PluginInfo{
				ID:      id,
				Name:    name,
				Version: "1.0.0",
			},
			logger: logrus.New(),
		},
	}
}

// OnInitialize 设置初始化函数
func (m *MockPlugin) OnInitialize(fn func(*PluginContext) error) *MockPlugin {
	m.initFunc = fn
	return m
}

// OnStart 设置启动函数
func (m *MockPlugin) OnStart(fn func() error) *MockPlugin {
	m.startFunc = fn
	return m
}

// OnStop 设置停止函数
func (m *MockPlugin) OnStop(fn func() error) *MockPlugin {
	m.stopFunc = fn
	return m
}

// OnCleanup 设置清理函数
func (m *MockPlugin) OnCleanup(fn func() error) *MockPlugin {
	m.cleanupFunc = fn
	return m
}

// OnEvent 设置事件处理函数
func (m *MockPlugin) OnEvent(fn func(context.Context, *Event) error) *MockPlugin {
	m.eventFunc = fn
	return m
}

// Initialize 初始化
func (m *MockPlugin) Initialize(ctx *PluginContext) error {
	if m.initFunc != nil {
		return m.initFunc(ctx)
	}
	return m.BasePlugin.Initialize(ctx)
}

// Start 启动
func (m *MockPlugin) Start() error {
	if m.startFunc != nil {
		return m.startFunc()
	}
	return m.BasePlugin.Start()
}

// Stop 停止
func (m *MockPlugin) Stop() error {
	if m.stopFunc != nil {
		return m.stopFunc()
	}
	return m.BasePlugin.Stop()
}

// Cleanup 清理
func (m *MockPlugin) Cleanup() error {
	if m.cleanupFunc != nil {
		return m.cleanupFunc()
	}
	return m.BasePlugin.Cleanup()
}

// HandleEvent 处理事件
func (m *MockPlugin) HandleEvent(ctx context.Context, event *Event) error {
	if m.eventFunc != nil {
		return m.eventFunc(ctx, event)
	}
	return nil
}

// TestHelper 测试辅助工具
type TestHelper struct {
	t *testing.T
}

// NewTestHelper 创建测试辅助工具
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// AssertNoError 断言无错误
func (h *TestHelper) AssertNoError(err error) {
	if err != nil {
		h.t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError 断言有错误
func (h *TestHelper) AssertError(err error) {
	if err == nil {
		h.t.Fatal("Expected error, got none")
	}
}

// AssertEqual 断言相等
func (h *TestHelper) AssertEqual(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		h.t.Fatalf("Expected %+v, got %+v", expected, actual)
	}
}

// AssertNotEqual 断言不相等
func (h *TestHelper) AssertNotEqual(expected, actual interface{}) {
	if reflect.DeepEqual(expected, actual) {
		h.t.Fatalf("Expected values to be different, but both were %+v", expected)
	}
}

// AssertContains 断言包含
func (h *TestHelper) AssertContains(haystack, needle string) {
	if !strings.Contains(haystack, needle) {
		h.t.Fatalf("Expected %q to contain %q", haystack, needle)
	}
}

// AssertTrue 断言为真
func (h *TestHelper) AssertTrue(condition bool) {
	if !condition {
		h.t.Fatal("Expected condition to be true")
	}
}

// AssertFalse 断言为假
func (h *TestHelper) AssertFalse(condition bool) {
	if condition {
		h.t.Fatal("Expected condition to be false")
	}
}