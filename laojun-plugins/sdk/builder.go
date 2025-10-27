package sdk

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// PluginBuilder 插件构建器
type PluginBuilder struct {
	info        *PluginInfo
	config      *ConfigSchema
	routes      []RouteDefinition
	middlewares []Middleware
	handlers    map[string]HandlerFunc
	eventTypes  []string
	schedule    *Schedule
	dataSchema  *DataSchemaDefinition
	hooks       map[string][]HookFunc
	logger      *logrus.Logger
	utils       *Utils
}

// RouteDefinition 路由定义
type RouteDefinition struct {
	Method  string
	Path    string
	Handler HandlerFunc
}

// DataSchemaDefinition 数据模式定义
type DataSchemaDefinition struct {
	Input  *DataSchema
	Output *DataSchema
}

// HookFunc 钩子函数类型
type HookFunc func(ctx context.Context) error

// NewPluginBuilder 创建新的插件构建器
func NewPluginBuilder() *PluginBuilder {
	return &PluginBuilder{
		info: &PluginInfo{
			Version: "1.0.0",
			Tags:    []string{},
		},
		routes:      []RouteDefinition{},
		middlewares: []Middleware{},
		handlers:    make(map[string]HandlerFunc),
		eventTypes:  []string{},
		hooks:       make(map[string][]HookFunc),
		logger:      logrus.New(),
		utils:       NewUtils(),
	}
}

// WithID 设置插件ID
func (b *PluginBuilder) WithID(id string) *PluginBuilder {
	b.info.ID = id
	return b
}

// WithName 设置插件名称
func (b *PluginBuilder) WithName(name string) *PluginBuilder {
	b.info.Name = name
	return b
}

// WithVersion 设置插件版本
func (b *PluginBuilder) WithVersion(version string) *PluginBuilder {
	b.info.Version = version
	return b
}

// WithDescription 设置插件描述
func (b *PluginBuilder) WithDescription(description string) *PluginBuilder {
	b.info.Description = description
	return b
}

// WithAuthor 设置插件作者
func (b *PluginBuilder) WithAuthor(author string) *PluginBuilder {
	b.info.Author = author
	return b
}

// WithCategory 设置插件分类
func (b *PluginBuilder) WithCategory(category string) *PluginBuilder {
	b.info.Category = category
	return b
}

// WithTags 设置插件标签
func (b *PluginBuilder) WithTags(tags ...string) *PluginBuilder {
	b.info.Tags = append(b.info.Tags, tags...)
	return b
}

// WithLogger 设置日志记录器
func (b *PluginBuilder) WithLogger(logger *logrus.Logger) *PluginBuilder {
	b.logger = logger
	return b
}

// WithConfig 设置配置模式
func (b *PluginBuilder) WithConfig(schema *ConfigSchema) *PluginBuilder {
	b.config = schema
	return b
}

// AddRoute 添加HTTP路由
func (b *PluginBuilder) AddRoute(method, path string, handler HandlerFunc) *PluginBuilder {
	b.routes = append(b.routes, RouteDefinition{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
	return b
}

// GET 添加GET路由
func (b *PluginBuilder) GET(path string, handler HandlerFunc) *PluginBuilder {
	return b.AddRoute("GET", path, handler)
}

// POST 添加POST路由
func (b *PluginBuilder) POST(path string, handler HandlerFunc) *PluginBuilder {
	return b.AddRoute("POST", path, handler)
}

// PUT 添加PUT路由
func (b *PluginBuilder) PUT(path string, handler HandlerFunc) *PluginBuilder {
	return b.AddRoute("PUT", path, handler)
}

// DELETE 添加DELETE路由
func (b *PluginBuilder) DELETE(path string, handler HandlerFunc) *PluginBuilder {
	return b.AddRoute("DELETE", path, handler)
}

// AddMiddleware 添加中间件
func (b *PluginBuilder) AddMiddleware(middleware Middleware) *PluginBuilder {
	b.middlewares = append(b.middlewares, middleware)
	return b
}

// WithCORS 添加CORS中间件
func (b *PluginBuilder) WithCORS() *PluginBuilder {
	return b.AddMiddleware(func(next HandlerFunc) HandlerFunc {
		return func(ctx *RequestContext) *Response {
			// 设置CORS头
			ctx.Response.Headers["Access-Control-Allow-Origin"] = "*"
			ctx.Response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
			ctx.Response.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
			
			if ctx.Request.Method == "OPTIONS" {
				return &Response{
					StatusCode: http.StatusOK,
					Headers:    ctx.Response.Headers,
				}
			}
			
			return next(ctx)
		}
	})
}

// WithLogging 添加日志中间件
func (b *PluginBuilder) WithLogging() *PluginBuilder {
	return b.AddMiddleware(func(next HandlerFunc) HandlerFunc {
		return func(ctx *RequestContext) *Response {
			start := time.Now()
			
			b.logger.WithFields(logrus.Fields{
				"method": ctx.Request.Method,
				"path":   ctx.Request.Path,
				"ip":     ctx.Request.Headers["X-Real-IP"],
			}).Info("Request started")
			
			resp := next(ctx)
			
			duration := time.Since(start)
			b.logger.WithFields(logrus.Fields{
				"method":     ctx.Request.Method,
				"path":       ctx.Request.Path,
				"status":     resp.StatusCode,
				"duration":   duration,
			}).Info("Request completed")
			
			return resp
		}
	})
}

// AddEventType 添加事件类型
func (b *PluginBuilder) AddEventType(eventType string) *PluginBuilder {
	b.eventTypes = append(b.eventTypes, eventType)
	return b
}

// WithSchedule 设置定时任务
func (b *PluginBuilder) WithSchedule(schedule *Schedule) *PluginBuilder {
	b.schedule = schedule
	return b
}

// WithCronSchedule 设置Cron定时任务
func (b *PluginBuilder) WithCronSchedule(cron string) *PluginBuilder {
	b.schedule = &Schedule{
		Type:       "cron",
		Expression: cron,
	}
	return b
}

// WithIntervalSchedule 设置间隔定时任务
func (b *PluginBuilder) WithIntervalSchedule(interval time.Duration) *PluginBuilder {
	b.schedule = &Schedule{
		Type:     "interval",
		Interval: interval,
	}
	return b
}

// WithDataSchema 设置数据处理模式
func (b *PluginBuilder) WithDataSchema(input, output *DataSchema) *PluginBuilder {
	b.dataSchema = &DataSchemaDefinition{
		Input:  input,
		Output: output,
	}
	return b
}

// AddHook 添加生命周期钩子
func (b *PluginBuilder) AddHook(event string, hook HookFunc) *PluginBuilder {
	if b.hooks[event] == nil {
		b.hooks[event] = []HookFunc{}
	}
	b.hooks[event] = append(b.hooks[event], hook)
	return b
}

// OnInitialize 添加初始化钩子
func (b *PluginBuilder) OnInitialize(hook HookFunc) *PluginBuilder {
	return b.AddHook("initialize", hook)
}

// OnStart 添加启动钩子
func (b *PluginBuilder) OnStart(hook HookFunc) *PluginBuilder {
	return b.AddHook("start", hook)
}

// OnStop 添加停止钩子
func (b *PluginBuilder) OnStop(hook HookFunc) *PluginBuilder {
	return b.AddHook("stop", hook)
}

// OnCleanup 添加清理钩子
func (b *PluginBuilder) OnCleanup(hook HookFunc) *PluginBuilder {
	return b.AddHook("cleanup", hook)
}

// Build 构建插件
func (b *PluginBuilder) Build() Plugin {
	// 验证必要字段
	if b.info.ID == "" {
		panic("plugin ID is required")
	}
	if b.info.Name == "" {
		panic("plugin name is required")
	}

	return &BuiltPlugin{
		info:        b.info,
		config:      b.config,
		routes:      b.routes,
		middlewares: b.middlewares,
		handlers:    b.handlers,
		eventTypes:  b.eventTypes,
		schedule:    b.schedule,
		dataSchema:  b.dataSchema,
		hooks:       b.hooks,
		logger:      b.logger,
		utils:       b.utils,
		state:       PluginStateStopped,
	}
}

// BuiltPlugin 构建的插件实现
type BuiltPlugin struct {
	info        *PluginInfo
	config      *ConfigSchema
	routes      []RouteDefinition
	middlewares []Middleware
	handlers    map[string]HandlerFunc
	eventTypes  []string
	schedule    *Schedule
	dataSchema  *DataSchemaDefinition
	hooks       map[string][]HookFunc
	logger      *logrus.Logger
	utils       *Utils
	state       PluginState
	context     *PluginContext
}

// GetMetadata 获取插件元数据
func (p *BuiltPlugin) GetMetadata() *PluginInfo {
	return p.info
}

// Initialize 初始化插件
func (p *BuiltPlugin) Initialize(ctx *PluginContext) error {
	p.context = ctx
	p.state = PluginStateInitializing
	
	// 执行初始化钩子
	if hooks, exists := p.hooks["initialize"]; exists {
		for _, hook := range hooks {
			if err := hook(context.Background()); err != nil {
				p.state = PluginStateError
				return fmt.Errorf("initialization hook failed: %w", err)
			}
		}
	}
	
	p.state = PluginStateInitialized
	p.logger.WithField("plugin_id", p.info.ID).Info("Plugin initialized")
	return nil
}

// Start 启动插件
func (p *BuiltPlugin) Start() error {
	if p.state != PluginStateInitialized {
		return fmt.Errorf("plugin must be initialized before starting")
	}
	
	p.state = PluginStateStarting
	
	// 执行启动钩子
	if hooks, exists := p.hooks["start"]; exists {
		for _, hook := range hooks {
			if err := hook(context.Background()); err != nil {
				p.state = PluginStateError
				return fmt.Errorf("start hook failed: %w", err)
			}
		}
	}
	
	p.state = PluginStateRunning
	p.logger.WithField("plugin_id", p.info.ID).Info("Plugin started")
	return nil
}

// Stop 停止插件
func (p *BuiltPlugin) Stop() error {
	if p.state != PluginStateRunning {
		return fmt.Errorf("plugin is not running")
	}
	
	p.state = PluginStateStopping
	
	// 执行停止钩子
	if hooks, exists := p.hooks["stop"]; exists {
		for _, hook := range hooks {
			if err := hook(context.Background()); err != nil {
				p.logger.WithError(err).Warn("Stop hook failed")
			}
		}
	}
	
	p.state = PluginStateStopped
	p.logger.WithField("plugin_id", p.info.ID).Info("Plugin stopped")
	return nil
}

// Cleanup 清理插件
func (p *BuiltPlugin) Cleanup() error {
	// 执行清理钩子
	if hooks, exists := p.hooks["cleanup"]; exists {
		for _, hook := range hooks {
			if err := hook(context.Background()); err != nil {
				p.logger.WithError(err).Warn("Cleanup hook failed")
			}
		}
	}
	
	p.logger.WithField("plugin_id", p.info.ID).Info("Plugin cleaned up")
	return nil
}

// GetStatus 获取插件状态
func (p *BuiltPlugin) GetStatus() PluginState {
	return p.state
}

// HandleEvent 处理事件
func (p *BuiltPlugin) HandleEvent(ctx context.Context, event *Event) error {
	p.logger.WithFields(logrus.Fields{
		"plugin_id":  p.info.ID,
		"event_type": event.Type,
	}).Debug("Handling event")
	
	return nil
}

// 实现HTTPPlugin接口
func (p *BuiltPlugin) RegisterRoutes(router Router) {
	for _, route := range p.routes {
		switch route.Method {
		case "GET":
			router.GET(route.Path, route.Handler)
		case "POST":
			router.POST(route.Path, route.Handler)
		case "PUT":
			router.PUT(route.Path, route.Handler)
		case "DELETE":
			router.DELETE(route.Path, route.Handler)
		}
	}
}

func (p *BuiltPlugin) GetMiddlewares() []Middleware {
	return p.middlewares
}

// 实现EventPlugin接口
func (p *BuiltPlugin) GetEventTypes() []string {
	return p.eventTypes
}

// 实现ScheduledPlugin接口
func (p *BuiltPlugin) GetSchedule() *Schedule {
	return p.schedule
}

func (p *BuiltPlugin) Execute(ctx context.Context) error {
	p.logger.WithField("plugin_id", p.info.ID).Info("Executing scheduled task")
	return nil
}

// 实现DataPlugin接口
func (p *BuiltPlugin) ProcessData(ctx context.Context, input *DataInput) (*DataOutput, error) {
	p.logger.WithField("plugin_id", p.info.ID).Info("Processing data")
	
	// 默认返回空输出
	return &DataOutput{
		Data:      make(map[string]interface{}),
		Metadata:  make(map[string]string),
		Timestamp: time.Now(),
	}, nil
}

func (p *BuiltPlugin) GetInputSchema() *DataSchema {
	if p.dataSchema != nil {
		return p.dataSchema.Input
	}
	return nil
}

func (p *BuiltPlugin) GetOutputSchema() *DataSchema {
	if p.dataSchema != nil {
		return p.dataSchema.Output
	}
	return nil
}

// 实现ConfigurablePlugin接口
func (p *BuiltPlugin) GetConfigSchema() *ConfigSchema {
	return p.config
}

func (p *BuiltPlugin) ValidateConfig(config map[string]interface{}) error {
	if p.config == nil {
		return nil
	}
	
	validator := NewConfigValidator(p.config)
	errors := validator.Validate(config)
	
	if errors.HasErrors() {
		return errors
	}
	
	return nil
}

func (p *BuiltPlugin) UpdateConfig(config map[string]interface{}) error {
	if err := p.ValidateConfig(config); err != nil {
		return err
	}
	
	p.logger.WithField("plugin_id", p.info.ID).Info("Config updated")
	return nil
}