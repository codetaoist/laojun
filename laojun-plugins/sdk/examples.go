package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// ExampleHTTPPlugin 示例HTTP插件
type ExampleHTTPPlugin struct {
	*BasePlugin
	server *http.Server
}

// NewExampleHTTPPlugin 创建示例HTTP插件
func NewExampleHTTPPlugin() *ExampleHTTPPlugin {
	base := &BasePlugin{
		info: &PluginInfo{
			ID:          "example-http-plugin",
			Name:        "Example HTTP Plugin",
			Version:     "1.0.0",
			Description: "An example HTTP plugin demonstrating basic functionality",
			Author:      "Laojun Team",
			Category:    "web",
			Tags:        []string{"example", "http", "demo"},
		},
		logger: logrus.New(),
	}

	return &ExampleHTTPPlugin{
		BasePlugin: base,
	}
}

// Initialize 初始化插件
func (p *ExampleHTTPPlugin) Initialize(ctx *PluginContext) error {
	if err := p.BasePlugin.Initialize(ctx); err != nil {
		return err
	}

	p.logger.Info("Example HTTP plugin initialized")
	return nil
}

// Start 启动插件
func (p *ExampleHTTPPlugin) Start() error {
	if err := p.BasePlugin.Start(); err != nil {
		return err
	}

	// 启动HTTP服务器
	mux := http.NewServeMux()
	p.registerRoutes(mux)

	p.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.WithError(err).Error("HTTP server error")
		}
	}()

	p.logger.Info("Example HTTP plugin started on :8080")
	return nil
}

// Stop 停止插件
func (p *ExampleHTTPPlugin) Stop() error {
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := p.server.Shutdown(ctx); err != nil {
			p.logger.WithError(err).Error("Failed to shutdown HTTP server")
		}
	}

	return p.BasePlugin.Stop()
}

// RegisterRoutes 注册路由
func (p *ExampleHTTPPlugin) RegisterRoutes(router Router) {
	router.GET("/hello", p.handleHello)
	router.POST("/echo", p.handleEcho)
	router.GET("/status", p.handleStatus)
}

// GetMiddlewares 获取中间件
func (p *ExampleHTTPPlugin) GetMiddlewares() []Middleware {
	return []Middleware{
		p.loggingMiddleware,
		p.corsMiddleware,
	}
}

// registerRoutes 注册内部路由
func (p *ExampleHTTPPlugin) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/hello", p.handleHelloHTTP)
	mux.HandleFunc("/echo", p.handleEchoHTTP)
	mux.HandleFunc("/status", p.handleStatusHTTP)
}

// handleHello 处理hello请求
func (p *ExampleHTTPPlugin) handleHello(ctx *RequestContext) *Response {
	name := ctx.Query("name")
	if name == "" {
		name = "World"
	}

	return &Response{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       fmt.Sprintf(`{"message": "Hello, %s!"}`, name),
	}
}

// handleEcho 处理echo请求
func (p *ExampleHTTPPlugin) handleEcho(ctx *RequestContext) *Response {
	return &Response{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       ctx.Body,
	}
}

// handleStatus 处理状态请求
func (p *ExampleHTTPPlugin) handleStatus(ctx *RequestContext) *Response {
	status := map[string]interface{}{
		"plugin_id": p.info.ID,
		"status":    "running",
		"uptime":    time.Since(p.startTime).String(),
		"version":   p.info.Version,
	}

	data, _ := json.Marshal(status)
	return &Response{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(data),
	}
}

// HTTP处理器适配
func (p *ExampleHTTPPlugin) handleHelloHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &RequestContext{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	// 转换头部
	for key, values := range r.Header {
		if len(values) > 0 {
			ctx.Headers[key] = values[0]
		}
	}

	// 转换查询参数
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			ctx.Query[key] = values[0]
		}
	}

	resp := p.handleHello(ctx)
	p.writeHTTPResponse(w, resp)
}

func (p *ExampleHTTPPlugin) handleEchoHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &RequestContext{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: make(map[string]string),
	}

	// 读取请求体
	if r.Body != nil {
		defer r.Body.Close()
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		ctx.Body = string(body)
	}

	resp := p.handleEcho(ctx)
	p.writeHTTPResponse(w, resp)
}

func (p *ExampleHTTPPlugin) handleStatusHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &RequestContext{
		Method: r.Method,
		Path:   r.URL.Path,
	}

	resp := p.handleStatus(ctx)
	p.writeHTTPResponse(w, resp)
}

// writeHTTPResponse 写入HTTP响应
func (p *ExampleHTTPPlugin) writeHTTPResponse(w http.ResponseWriter, resp *Response) {
	// 设置头部
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 写入响应体
	if resp.Body != "" {
		w.Write([]byte(resp.Body))
	}
}

// 中间件
func (p *ExampleHTTPPlugin) loggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *RequestContext) *Response {
		start := time.Now()
		resp := next(ctx)
		duration := time.Since(start)

		p.logger.WithFields(logrus.Fields{
			"method":     ctx.Method,
			"path":       ctx.Path,
			"status":     resp.StatusCode,
			"duration":   duration.String(),
			"user_agent": ctx.Headers["User-Agent"],
		}).Info("HTTP request processed")

		return resp
	}
}

func (p *ExampleHTTPPlugin) corsMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *RequestContext) *Response {
		resp := next(ctx)

		if resp.Headers == nil {
			resp.Headers = make(map[string]string)
		}

		resp.Headers["Access-Control-Allow-Origin"] = "*"
		resp.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
		resp.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"

		return resp
	}
}

// ExampleEventPlugin 示例事件插件
type ExampleEventPlugin struct {
	*BasePlugin
	eventCount int
}

// NewExampleEventPlugin 创建示例事件插件
func NewExampleEventPlugin() *ExampleEventPlugin {
	base := &BasePlugin{
		info: &PluginInfo{
			ID:          "example-event-plugin",
			Name:        "Example Event Plugin",
			Version:     "1.0.0",
			Description: "An example event plugin demonstrating event handling",
			Author:      "Laojun Team",
			Category:    "event",
			Tags:        []string{"example", "event", "demo"},
		},
		logger: logrus.New(),
	}

	return &ExampleEventPlugin{
		BasePlugin: base,
	}
}

// HandleEvent 处理事件
func (p *ExampleEventPlugin) HandleEvent(ctx context.Context, event *Event) error {
	p.eventCount++

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"source":     event.Source,
		"count":      p.eventCount,
	}).Info("Event received")

	// 根据事件类型处理
	switch event.Type {
	case "plugin.started":
		p.handlePluginStarted(event)
	case "plugin.stopped":
		p.handlePluginStopped(event)
	case "system.health_check":
		p.handleHealthCheck(event)
	default:
		p.logger.WithField("event_type", event.Type).Debug("Unknown event type")
	}

	return nil
}

// GetEventTypes 获取支持的事件类型
func (p *ExampleEventPlugin) GetEventTypes() []string {
	return []string{
		"plugin.started",
		"plugin.stopped",
		"system.health_check",
		"user.action",
	}
}

// handlePluginStarted 处理插件启动事件
func (p *ExampleEventPlugin) handlePluginStarted(event *Event) {
	p.logger.WithField("plugin_id", event.Data["plugin_id"]).Info("Plugin started event handled")
}

// handlePluginStopped 处理插件停止事件
func (p *ExampleEventPlugin) handlePluginStopped(event *Event) {
	p.logger.WithField("plugin_id", event.Data["plugin_id"]).Info("Plugin stopped event handled")
}

// handleHealthCheck 处理健康检查事件
func (p *ExampleEventPlugin) handleHealthCheck(event *Event) {
	p.logger.Debug("Health check event handled")
}

// ExampleScheduledPlugin 示例定时任务插件
type ExampleScheduledPlugin struct {
	*BasePlugin
	executionCount int
}

// NewExampleScheduledPlugin 创建示例定时任务插件
func NewExampleScheduledPlugin() *ExampleScheduledPlugin {
	base := &BasePlugin{
		info: &PluginInfo{
			ID:          "example-scheduled-plugin",
			Name:        "Example Scheduled Plugin",
			Version:     "1.0.0",
			Description: "An example scheduled plugin demonstrating cron jobs",
			Author:      "Laojun Team",
			Category:    "scheduler",
			Tags:        []string{"example", "scheduler", "cron", "demo"},
		},
		logger: logrus.New(),
	}

	return &ExampleScheduledPlugin{
		BasePlugin: base,
	}
}

// GetSchedule 获取调度配置
func (p *ExampleScheduledPlugin) GetSchedule() *Schedule {
	return &Schedule{
		Cron:        "*/30 * * * * *", // 每30秒执行一次
		Timezone:    "UTC",
		MaxRetries:  3,
		RetryDelay:  time.Second * 5,
		Timeout:     time.Minute * 2,
		Description: "Example scheduled task that runs every 30 seconds",
	}
}

// Execute 执行定时任务
func (p *ExampleScheduledPlugin) Execute(ctx context.Context) error {
	p.executionCount++

	p.logger.WithFields(logrus.Fields{
		"execution_count": p.executionCount,
		"timestamp":       time.Now().Format(time.RFC3339),
	}).Info("Scheduled task executed")

	// 模拟一些工作
	select {
	case <-time.After(time.Second * 2):
		p.logger.Debug("Scheduled task work completed")
	case <-ctx.Done():
		p.logger.Warn("Scheduled task cancelled")
		return ctx.Err()
	}

	return nil
}

// ExampleDataPlugin 示例数据处理插件
type ExampleDataPlugin struct {
	*BasePlugin
	processedCount int
}

// NewExampleDataPlugin 创建示例数据处理插件
func NewExampleDataPlugin() *ExampleDataPlugin {
	base := &BasePlugin{
		info: &PluginInfo{
			ID:          "example-data-plugin",
			Name:        "Example Data Plugin",
			Version:     "1.0.0",
			Description: "An example data plugin demonstrating data processing",
			Author:      "Laojun Team",
			Category:    "data",
			Tags:        []string{"example", "data", "processing", "demo"},
		},
		logger: logrus.New(),
	}

	return &ExampleDataPlugin{
		BasePlugin: base,
	}
}

// ProcessData 处理数据
func (p *ExampleDataPlugin) ProcessData(ctx context.Context, input *DataInput) (*DataOutput, error) {
	p.processedCount++

	p.logger.WithFields(logrus.Fields{
		"input_type":      input.Type,
		"input_size":      len(input.Data),
		"processed_count": p.processedCount,
	}).Info("Processing data")

	// 根据输入类型处理数据
	var result map[string]interface{}
	var outputType string

	switch input.Type {
	case "json":
		var data map[string]interface{}
		if err := json.Unmarshal(input.Data, &data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		// 添加处理时间戳
		data["processed_at"] = time.Now().Format(time.RFC3339)
		data["processed_by"] = p.info.ID

		result = data
		outputType = "json"

	case "text":
		result = map[string]interface{}{
			"original_text": string(input.Data),
			"word_count":    len(strings.Fields(string(input.Data))),
			"char_count":    len(input.Data),
			"processed_at":  time.Now().Format(time.RFC3339),
			"processed_by":  p.info.ID,
		}
		outputType = "json"

	default:
		return nil, fmt.Errorf("unsupported input type: %s", input.Type)
	}

	// 序列化结果
	outputData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &DataOutput{
		Type: outputType,
		Data: outputData,
		Metadata: map[string]interface{}{
			"processing_time": time.Since(time.Now()).String(),
			"plugin_version":  p.info.Version,
		},
	}, nil
}

// GetInputSchema 获取输入数据模式
func (p *ExampleDataPlugin) GetInputSchema() *DataSchema {
	return &DataSchema{
		Type:        "object",
		Description: "Input data for processing",
		Properties: map[string]*PropertySpec{
			"type": {
				Type:        "string",
				Description: "Data type (json, text)",
				Required:    true,
				Enum:        []interface{}{"json", "text"},
			},
			"data": {
				Type:        "string",
				Description: "Raw data to process",
				Required:    true,
			},
		},
	}
}

// GetOutputSchema 获取输出数据模式
func (p *ExampleDataPlugin) GetOutputSchema() *DataSchema {
	return &DataSchema{
		Type:        "object",
		Description: "Processed data output",
		Properties: map[string]*PropertySpec{
			"type": {
				Type:        "string",
				Description: "Output data type",
			},
			"data": {
				Type:        "string",
				Description: "Processed data",
			},
			"metadata": {
				Type:        "object",
				Description: "Processing metadata",
			},
		},
	}
}

// ExampleConfigurablePlugin 示例可配置插件
type ExampleConfigurablePlugin struct {
	*BasePlugin
	config *ExampleConfig
}

// ExampleConfig 示例配置
type ExampleConfig struct {
	APIKey      string        `json:"api_key"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	EnableDebug bool          `json:"enable_debug"`
	Endpoints   []string      `json:"endpoints"`
}

// NewExampleConfigurablePlugin 创建示例可配置插件
func NewExampleConfigurablePlugin() *ExampleConfigurablePlugin {
	base := &BasePlugin{
		info: &PluginInfo{
			ID:          "example-configurable-plugin",
			Name:        "Example Configurable Plugin",
			Version:     "1.0.0",
			Description: "An example configurable plugin demonstrating configuration management",
			Author:      "Laojun Team",
			Category:    "utility",
			Tags:        []string{"example", "configurable", "demo"},
		},
		logger: logrus.New(),
	}

	return &ExampleConfigurablePlugin{
		BasePlugin: base,
		config: &ExampleConfig{
			Timeout:     time.Second * 30,
			MaxRetries:  3,
			EnableDebug: false,
		},
	}
}

// GetConfigSchema 获取配置模式
func (p *ExampleConfigurablePlugin) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Type:        "object",
		Description: "Plugin configuration schema",
		Properties: map[string]*ConfigProperty{
			"api_key": {
				Type:        "string",
				Description: "API key for external service",
				Required:    true,
				Sensitive:   true,
			},
			"timeout": {
				Type:        "string",
				Description: "Request timeout duration (e.g., '30s', '1m')",
				Default:     "30s",
				Pattern:     `^\d+[smh]$`,
			},
			"max_retries": {
				Type:        "integer",
				Description: "Maximum number of retries",
				Default:     3,
				Minimum:     &[]float64{0}[0],
				Maximum:     &[]float64{10}[0],
			},
			"enable_debug": {
				Type:        "boolean",
				Description: "Enable debug logging",
				Default:     false,
			},
			"endpoints": {
				Type:        "array",
				Description: "List of API endpoints",
				Items: &ConfigProperty{
					Type: "string",
				},
				MinItems: &[]int{1}[0],
				MaxItems: &[]int{10}[0],
			},
		},
	}
}

// ValidateConfig 验证配置
func (p *ExampleConfigurablePlugin) ValidateConfig(config map[string]interface{}) error {
	validator := NewConfigValidator(p.GetConfigSchema())
	errors := validator.Validate(config)

	if errors.HasErrors() {
		return errors
	}

	// 自定义验证逻辑
	if apiKey, ok := config["api_key"].(string); ok && len(apiKey) < 10 {
		return fmt.Errorf("api_key must be at least 10 characters long")
	}

	return nil
}

// UpdateConfig 更新配置
func (p *ExampleConfigurablePlugin) UpdateConfig(config map[string]interface{}) error {
	if err := p.ValidateConfig(config); err != nil {
		return err
	}

	// 更新配置
	if apiKey, ok := config["api_key"].(string); ok {
		p.config.APIKey = apiKey
	}

	if timeoutStr, ok := config["timeout"].(string); ok {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			p.config.Timeout = timeout
		}
	}

	if maxRetries, ok := config["max_retries"].(float64); ok {
		p.config.MaxRetries = int(maxRetries)
	}

	if enableDebug, ok := config["enable_debug"].(bool); ok {
		p.config.EnableDebug = enableDebug
		if enableDebug {
			p.logger.SetLevel(logrus.DebugLevel)
		} else {
			p.logger.SetLevel(logrus.InfoLevel)
		}
	}

	if endpoints, ok := config["endpoints"].([]interface{}); ok {
		p.config.Endpoints = make([]string, len(endpoints))
		for i, endpoint := range endpoints {
			if str, ok := endpoint.(string); ok {
				p.config.Endpoints[i] = str
			}
		}
	}

	p.logger.WithField("config", p.config).Info("Configuration updated")
	return nil
}

// ExamplePluginBuilder 示例插件构建器使用
func ExamplePluginBuilder() Plugin {
	return NewPluginBuilder("example-builder-plugin").
		WithName("Example Builder Plugin").
		WithVersion("1.0.0").
		WithDescription("Plugin created using builder pattern").
		WithAuthor("Laojun Team").
		WithCategory("example").
		WithTags("builder", "pattern", "demo").
		WithHTTPHandler("/hello", func(ctx *RequestContext) *Response {
			return &Response{
				StatusCode: 200,
				Headers:    map[string]string{"Content-Type": "text/plain"},
				Body:       "Hello from builder plugin!",
			}
		}).
		WithEventHandler("test.event", func(ctx context.Context, event *Event) error {
			fmt.Printf("Received event: %s\n", event.Type)
			return nil
		}).
		WithSchedule("0 */5 * * * *", func(ctx context.Context) error {
			fmt.Println("Scheduled task executed")
			return nil
		}).
		Build()
}