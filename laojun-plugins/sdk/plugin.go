package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Homepage    string            `json:"homepage,omitempty"`
	Repository  string            `json:"repository,omitempty"`
	License     string            `json:"license,omitempty"`
	Config      map[string]string `json:"config,omitempty"`
	Permissions []string          `json:"permissions"`
}

// PluginContext 插件上下文
type PluginContext struct {
	PluginID    string                 `json:"plugin_id"`
	Config      map[string]interface{} `json:"config"`
	Logger      *logrus.Logger         `json:"-"`
	DataDir     string                 `json:"data_dir"`
	TempDir     string                 `json:"temp_dir"`
	HTTPClient  *http.Client           `json:"-"`
	Registry    RegistryClient         `json:"-"`
	EventBus    EventBusClient         `json:"-"`
	Metadata    map[string]string      `json:"metadata"`
}

// Plugin 插件接口
type Plugin interface {
	// GetInfo 获取插件信息
	GetInfo() *PluginInfo

	// Initialize 初始化插件
	Initialize(ctx context.Context, pluginCtx *PluginContext) error

	// Start 启动插件
	Start(ctx context.Context) error

	// Stop 停止插件
	Stop(ctx context.Context) error

	// Cleanup 清理资源
	Cleanup(ctx context.Context) error

	// HandleRequest 处理请求
	HandleRequest(ctx context.Context, request *Request) (*Response, error)

	// GetHealth 获取健康状态
	GetHealth(ctx context.Context) (*HealthStatus, error)

	// GetMetrics 获取指标
	GetMetrics(ctx context.Context) (*Metrics, error)
}

// Request 请求结构
type Request struct {
	ID       string                 `json:"id"`
	Method   string                 `json:"method"`
	Path     string                 `json:"path"`
	Headers  map[string]string      `json:"headers"`
	Query    map[string]string      `json:"query"`
	Body     []byte                 `json:"body,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// Response 响应结构
type Response struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       []byte                 `json:"body,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]string      `json:"metadata,omitempty"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status      string            `json:"status"` // healthy, unhealthy, unknown
	Message     string            `json:"message,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	LastChecked time.Time         `json:"last_checked"`
}

// Metrics 指标
type Metrics struct {
	RequestCount    int64             `json:"request_count"`
	ErrorCount      int64             `json:"error_count"`
	ResponseTime    time.Duration     `json:"response_time"`
	MemoryUsage     int64             `json:"memory_usage"`
	CPUUsage        float64           `json:"cpu_usage"`
	GoroutineCount  int               `json:"goroutine_count"`
	CustomMetrics   map[string]int64  `json:"custom_metrics,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// BasePlugin 基础插件实现
type BasePlugin struct {
	info    *PluginInfo
	context *PluginContext
	logger  *logrus.Logger
	started bool
	metrics *Metrics
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(info *PluginInfo) *BasePlugin {
	return &BasePlugin{
		info: info,
		metrics: &Metrics{
			CustomMetrics: make(map[string]int64),
			Labels:        make(map[string]string),
			Timestamp:     time.Now(),
		},
	}
}

// GetInfo 获取插件信息
func (p *BasePlugin) GetInfo() *PluginInfo {
	return p.info
}

// Initialize 初始化插件
func (p *BasePlugin) Initialize(ctx context.Context, pluginCtx *PluginContext) error {
	p.context = pluginCtx
	p.logger = pluginCtx.Logger.WithField("plugin_id", pluginCtx.PluginID).Logger

	p.logger.WithField("plugin_info", p.info).Info("Initializing plugin")
	return nil
}

// Start 启动插件
func (p *BasePlugin) Start(ctx context.Context) error {
	if p.started {
		return fmt.Errorf("plugin already started")
	}

	p.logger.Info("Starting plugin")
	p.started = true
	return nil
}

// Stop 停止插件
func (p *BasePlugin) Stop(ctx context.Context) error {
	if !p.started {
		return fmt.Errorf("plugin not started")
	}

	p.logger.Info("Stopping plugin")
	p.started = false
	return nil
}

// Cleanup 清理资源
func (p *BasePlugin) Cleanup(ctx context.Context) error {
	p.logger.Info("Cleaning up plugin")
	return nil
}

// HandleRequest 处理请求
func (p *BasePlugin) HandleRequest(ctx context.Context, request *Request) (*Response, error) {
	p.metrics.RequestCount++
	startTime := time.Now()

	defer func() {
		p.metrics.ResponseTime = time.Since(startTime)
		p.metrics.Timestamp = time.Now()
	}()

	// 默认实现返回404
	return &Response{
		StatusCode: http.StatusNotFound,
		Error:      "handler not implemented",
	}, nil
}

// GetHealth 获取健康状态
func (p *BasePlugin) GetHealth(ctx context.Context) (*HealthStatus, error) {
	status := "healthy"
	message := "Plugin is running normally"

	if !p.started {
		status = "unhealthy"
		message = "Plugin is not started"
	}

	return &HealthStatus{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
	}, nil
}

// GetMetrics 获取指标
func (p *BasePlugin) GetMetrics(ctx context.Context) (*Metrics, error) {
	// 更新时间戳
	p.metrics.Timestamp = time.Now()
	
	// 返回指标副本
	metricsCopy := *p.metrics
	return &metricsCopy, nil
}

// GetContext 获取插件上下文
func (p *BasePlugin) GetContext() *PluginContext {
	return p.context
}

// GetLogger 获取日志记录器
func (p *BasePlugin) GetLogger() *logrus.Logger {
	return p.logger
}

// IsStarted 检查插件是否已启动
func (p *BasePlugin) IsStarted() bool {
	return p.started
}

// IncrementMetric 增加自定义指标
func (p *BasePlugin) IncrementMetric(name string, value int64) {
	if p.metrics.CustomMetrics == nil {
		p.metrics.CustomMetrics = make(map[string]int64)
	}
	p.metrics.CustomMetrics[name] += value
}

// SetMetricLabel 设置指标标签
func (p *BasePlugin) SetMetricLabel(key, value string) {
	if p.metrics.Labels == nil {
		p.metrics.Labels = make(map[string]string)
	}
	p.metrics.Labels[key] = value
}

// IncrementErrorCount 增加错误计数
func (p *BasePlugin) IncrementErrorCount() {
	p.metrics.ErrorCount++
}

// HTTPPlugin HTTP插件接口
type HTTPPlugin interface {
	Plugin

	// RegisterRoutes 注册路由
	RegisterRoutes(router Router) error

	// GetMiddlewares 获取中间件
	GetMiddlewares() []Middleware
}

// Router 路由器接口
type Router interface {
	// GET 注册GET路由
	GET(path string, handler HandlerFunc)

	// POST 注册POST路由
	POST(path string, handler HandlerFunc)

	// PUT 注册PUT路由
	PUT(path string, handler HandlerFunc)

	// DELETE 注册DELETE路由
	DELETE(path string, handler HandlerFunc)

	// PATCH 注册PATCH路由
	PATCH(path string, handler HandlerFunc)

	// Any 注册任意方法路由
	Any(path string, handler HandlerFunc)

	// Group 创建路由组
	Group(prefix string) Router

	// Use 使用中间件
	Use(middleware Middleware)
}

// HandlerFunc 处理函数
type HandlerFunc func(ctx *RequestContext) error

// Middleware 中间件
type Middleware func(next HandlerFunc) HandlerFunc

// RequestContext 请求上下文
type RequestContext struct {
	Request     *Request
	Response    *Response
	PluginCtx   *PluginContext
	Values      map[string]interface{}
	Logger      *logrus.Logger
}

// NewRequestContext 创建请求上下文
func NewRequestContext(req *Request, pluginCtx *PluginContext) *RequestContext {
	return &RequestContext{
		Request:   req,
		Response:  &Response{Headers: make(map[string]string)},
		PluginCtx: pluginCtx,
		Values:    make(map[string]interface{}),
		Logger:    pluginCtx.Logger,
	}
}

// JSON 返回JSON响应
func (c *RequestContext) JSON(statusCode int, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	c.Response.StatusCode = statusCode
	c.Response.Headers["Content-Type"] = "application/json"
	c.Response.Body = jsonData
	return nil
}

// String 返回字符串响应
func (c *RequestContext) String(statusCode int, text string) error {
	c.Response.StatusCode = statusCode
	c.Response.Headers["Content-Type"] = "text/plain"
	c.Response.Body = []byte(text)
	return nil
}

// Error 返回错误响应
func (c *RequestContext) Error(statusCode int, message string) error {
	c.Response.StatusCode = statusCode
	c.Response.Error = message
	return nil
}

// GetParam 获取路径参数
func (c *RequestContext) GetParam(key string) string {
	if value, exists := c.Request.Params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetQuery 获取查询参数
func (c *RequestContext) GetQuery(key string) string {
	return c.Request.Query[key]
}

// GetHeader 获取请求头
func (c *RequestContext) GetHeader(key string) string {
	return c.Request.Headers[key]
}

// SetHeader 设置响应头
func (c *RequestContext) SetHeader(key, value string) {
	c.Response.Headers[key] = value
}

// GetValue 获取上下文值
func (c *RequestContext) GetValue(key string) interface{} {
	return c.Values[key]
}

// SetValue 设置上下文值
func (c *RequestContext) SetValue(key string, value interface{}) {
	c.Values[key] = value
}

// BindJSON 绑定JSON数据
func (c *RequestContext) BindJSON(obj interface{}) error {
	if len(c.Request.Body) == 0 {
		return fmt.Errorf("empty request body")
	}
	return json.Unmarshal(c.Request.Body, obj)
}

// EventPlugin 事件插件接口
type EventPlugin interface {
	Plugin

	// HandleEvent 处理事件
	HandleEvent(ctx context.Context, event *Event) error

	// GetEventTypes 获取关注的事件类型
	GetEventTypes() []string
}

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata"`
}

// ScheduledPlugin 定时任务插件接口
type ScheduledPlugin interface {
	Plugin

	// GetSchedule 获取调度配置
	GetSchedule() *Schedule

	// Execute 执行定时任务
	Execute(ctx context.Context) error
}

// Schedule 调度配置
type Schedule struct {
	Cron        string        `json:"cron,omitempty"`        // Cron表达式
	Interval    time.Duration `json:"interval,omitempty"`    // 间隔时间
	Immediate   bool          `json:"immediate"`             // 是否立即执行
	MaxRetries  int           `json:"max_retries"`           // 最大重试次数
	Timeout     time.Duration `json:"timeout"`               // 执行超时时间
}

// DataPlugin 数据处理插件接口
type DataPlugin interface {
	Plugin

	// ProcessData 处理数据
	ProcessData(ctx context.Context, input *DataInput) (*DataOutput, error)

	// GetInputSchema 获取输入数据模式
	GetInputSchema() *DataSchema

	// GetOutputSchema 获取输出数据模式
	GetOutputSchema() *DataSchema
}

// DataInput 数据输入
type DataInput struct {
	Type     string                 `json:"type"`
	Data     interface{}            `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}

// DataOutput 数据输出
type DataOutput struct {
	Type     string                 `json:"type"`
	Data     interface{}            `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
	Error    string                 `json:"error,omitempty"`
}

// DataSchema 数据模式
type DataSchema struct {
	Type        string                    `json:"type"`
	Properties  map[string]*PropertySpec  `json:"properties"`
	Required    []string                  `json:"required"`
	Description string                    `json:"description"`
}

// PropertySpec 属性规范
type PropertySpec struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Format      string      `json:"format,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
}

// ConfigurablePlugin 可配置插件接口
type ConfigurablePlugin interface {
	Plugin

	// GetConfigSchema 获取配置模式
	GetConfigSchema() *ConfigSchema

	// ValidateConfig 验证配置
	ValidateConfig(config map[string]interface{}) error

	// UpdateConfig 更新配置
	UpdateConfig(ctx context.Context, config map[string]interface{}) error
}

// ConfigSchema 配置模式
type ConfigSchema struct {
	Properties map[string]*ConfigProperty `json:"properties"`
	Required   []string                   `json:"required"`
}

// ConfigProperty 配置属性
type ConfigProperty struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
}

// PluginBuilder 插件构建器
type PluginBuilder struct {
	info        *PluginInfo
	initFunc    func(ctx context.Context, pluginCtx *PluginContext) error
	startFunc   func(ctx context.Context) error
	stopFunc    func(ctx context.Context) error
	cleanupFunc func(ctx context.Context) error
	handlers    map[string]HandlerFunc
	middlewares []Middleware
}

// NewPluginBuilder 创建插件构建器
func NewPluginBuilder(info *PluginInfo) *PluginBuilder {
	return &PluginBuilder{
		info:     info,
		handlers: make(map[string]HandlerFunc),
	}
}

// OnInit 设置初始化函数
func (b *PluginBuilder) OnInit(fn func(ctx context.Context, pluginCtx *PluginContext) error) *PluginBuilder {
	b.initFunc = fn
	return b
}

// OnStart 设置启动函数
func (b *PluginBuilder) OnStart(fn func(ctx context.Context) error) *PluginBuilder {
	b.startFunc = fn
	return b
}

// OnStop 设置停止函数
func (b *PluginBuilder) OnStop(fn func(ctx context.Context) error) *PluginBuilder {
	b.stopFunc = fn
	return b
}

// OnCleanup 设置清理函数
func (b *PluginBuilder) OnCleanup(fn func(ctx context.Context) error) *PluginBuilder {
	b.cleanupFunc = fn
	return b
}

// AddHandler 添加处理器
func (b *PluginBuilder) AddHandler(path string, handler HandlerFunc) *PluginBuilder {
	b.handlers[path] = handler
	return b
}

// AddMiddleware 添加中间件
func (b *PluginBuilder) AddMiddleware(middleware Middleware) *PluginBuilder {
	b.middlewares = append(b.middlewares, middleware)
	return b
}

// Build 构建插件
func (b *PluginBuilder) Build() Plugin {
	return &builtPlugin{
		BasePlugin:  NewBasePlugin(b.info),
		initFunc:    b.initFunc,
		startFunc:   b.startFunc,
		stopFunc:    b.stopFunc,
		cleanupFunc: b.cleanupFunc,
		handlers:    b.handlers,
		middlewares: b.middlewares,
	}
}

// builtPlugin 构建的插件
type builtPlugin struct {
	*BasePlugin
	initFunc    func(ctx context.Context, pluginCtx *PluginContext) error
	startFunc   func(ctx context.Context) error
	stopFunc    func(ctx context.Context) error
	cleanupFunc func(ctx context.Context) error
	handlers    map[string]HandlerFunc
	middlewares []Middleware
}

// Initialize 初始化插件
func (p *builtPlugin) Initialize(ctx context.Context, pluginCtx *PluginContext) error {
	if err := p.BasePlugin.Initialize(ctx, pluginCtx); err != nil {
		return err
	}

	if p.initFunc != nil {
		return p.initFunc(ctx, pluginCtx)
	}

	return nil
}

// Start 启动插件
func (p *builtPlugin) Start(ctx context.Context) error {
	if err := p.BasePlugin.Start(ctx); err != nil {
		return err
	}

	if p.startFunc != nil {
		return p.startFunc(ctx)
	}

	return nil
}

// Stop 停止插件
func (p *builtPlugin) Stop(ctx context.Context) error {
	if p.stopFunc != nil {
		if err := p.stopFunc(ctx); err != nil {
			return err
		}
	}

	return p.BasePlugin.Stop(ctx)
}

// Cleanup 清理资源
func (p *builtPlugin) Cleanup(ctx context.Context) error {
	if p.cleanupFunc != nil {
		if err := p.cleanupFunc(ctx); err != nil {
			return err
		}
	}

	return p.BasePlugin.Cleanup(ctx)
}

// HandleRequest 处理请求
func (p *builtPlugin) HandleRequest(ctx context.Context, request *Request) (*Response, error) {
	handler, exists := p.handlers[request.Path]
	if !exists {
		return p.BasePlugin.HandleRequest(ctx, request)
	}

	reqCtx := NewRequestContext(request, p.GetContext())
	
	// 应用中间件
	finalHandler := handler
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		finalHandler = p.middlewares[i](finalHandler)
	}

	if err := finalHandler(reqCtx); err != nil {
		p.IncrementErrorCount()
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Error:      err.Error(),
		}, err
	}

	return reqCtx.Response, nil
}