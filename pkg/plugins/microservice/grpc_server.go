package microservice

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PluginServiceServer gRPC服务端实现
type PluginServiceServer struct {
	UnimplementedPluginServiceServer
	plugins    map[string]*MicroservicePlugin
	mutex      sync.RWMutex
	config     *GRPCConfig
	logger     *log.Logger
	healthSrv  *health.Server
	eventBus   EventBus
	middleware []Middleware
}

// GRPCConfig gRPC服务器配置
type GRPCConfig struct {
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	MaxRecvMsgSize    int           `json:"max_recv_msg_size"`
	MaxSendMsgSize    int           `json:"max_send_msg_size"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	KeepaliveTime     time.Duration `json:"keepalive_time"`
	KeepaliveTimeout  time.Duration `json:"keepalive_timeout"`
	MaxConnectionIdle time.Duration `json:"max_connection_idle"`
	MaxConnectionAge  time.Duration `json:"max_connection_age"`
	EnableReflection  bool          `json:"enable_reflection"`
	EnableHealthCheck bool          `json:"enable_health_check"`
	TLSCertFile       string        `json:"tls_cert_file"`
	TLSKeyFile        string        `json:"tls_key_file"`
	EnableTLS         bool          `json:"enable_tls"`
	AuthToken         string        `json:"auth_token"`
	EnableAuth        bool          `json:"enable_auth"`
}

// MicroservicePlugin 微服务插件信息
type MicroservicePlugin struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Status      string                 `json:"status"`
	ContainerID string                 `json:"container_id"`
	Endpoint    string                 `json:"endpoint"`
	Metadata    map[string]string      `json:"metadata"`
	Config      map[string]interface{} `json:"config"`
	Health      *PluginHealth          `json:"health"`
	Stats       *PluginStats           `json:"stats"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PluginHealth 插件健康状态
type PluginHealth struct {
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	LastCheck    time.Time `json:"last_check"`
	CheckCount   int64     `json:"check_count"`
	FailureCount int64     `json:"failure_count"`
}

// PluginStats 插件统计信息
type PluginStats struct {
	RequestCount    int64     `json:"request_count"`
	ErrorCount      int64     `json:"error_count"`
	LastRequestTime time.Time `json:"last_request_time"`
	AverageLatency  float64   `json:"average_latency"`
	TotalLatency    float64   `json:"total_latency"`
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(event *Event) error
	Subscribe(eventType string, handler EventHandler) error
}

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventHandler 事件处理函数
type EventHandler func(event *Event) error

// Middleware 中间件接口
type Middleware interface {
	Process(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	token string
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	logger *log.Logger
}

// MetricsMiddleware 指标中间件
type MetricsMiddleware struct {
	requestCount map[string]int64
	mutex        sync.RWMutex
}

// NewPluginServiceServer 创建新的gRPC服务端实现
func NewPluginServiceServer(config *GRPCConfig, eventBus EventBus) *PluginServiceServer {
	server := &PluginServiceServer{
		plugins:    make(map[string]*MicroservicePlugin),
		config:     config,
		logger:     log.New(log.Writer(), "[gRPC] ", log.LstdFlags),
		eventBus:   eventBus,
		middleware: make([]Middleware, 0),
	}

	// 添加默认中间件
	if config.EnableAuth {
		server.AddMiddleware(&AuthMiddleware{token: config.AuthToken})
	}
	server.AddMiddleware(&LoggingMiddleware{logger: server.logger})
	server.AddMiddleware(&MetricsMiddleware{requestCount: make(map[string]int64)})

	return server
}

// AddMiddleware 添加中间件
func (s *PluginServiceServer) AddMiddleware(middleware Middleware) {
	s.middleware = append(s.middleware, middleware)
}

// Start 启动gRPC服务端
func (s *PluginServiceServer) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	// 配置gRPC服务器选项
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(s.config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(s.config.MaxSendMsgSize),
		grpc.ConnectionTimeout(s.config.ConnectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    s.config.KeepaliveTime,
			Timeout: s.config.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             s.config.KeepaliveTime,
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(s.unaryInterceptor),
	}

	// 创建gRPC服务端实例
	grpcServer := grpc.NewServer(opts...)

	// 注册服务
	RegisterPluginServiceServer(grpcServer, s)

	// 启用健康检查
	if s.config.EnableHealthCheck {
		s.healthSrv = health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, s.healthSrv)
		s.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	// 启用反射
	if s.config.EnableReflection {
		reflection.Register(grpcServer)
	}

	s.logger.Printf("gRPC server starting on %s", address)

	// 启动服务端
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			s.logger.Printf("gRPC server error: %v", err)
		}
	}()

	return nil
}

// RegisterPlugin 注册插件
func (s *PluginServiceServer) RegisterPlugin(ctx context.Context, req *RegisterPluginRequest) (*RegisterPluginResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	plugin := &MicroservicePlugin{
		ID:          req.PluginId,
		Name:        req.Name,
		Version:     req.Version,
		Status:      "registered",
		ContainerID: req.ContainerId,
		Endpoint:    req.Endpoint,
		Metadata:    req.Metadata,
		Config:      make(map[string]interface{}),
		Health: &PluginHealth{
			Status:       "unknown",
			LastCheck:    time.Now(),
			CheckCount:   0,
			FailureCount: 0,
		},
		Stats: &PluginStats{
			RequestCount:    0,
			ErrorCount:      0,
			LastRequestTime: time.Now(),
			AverageLatency:  0,
			TotalLatency:    0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.plugins[req.PluginId] = plugin
	s.logger.Printf("Plugin registered: %s (%s)", req.Name, req.PluginId)

	// 发布插件注册事件
	if s.eventBus != nil {
		event := &Event{
			Type:   "plugin.registered",
			Source: "grpc-server",
			Data: map[string]interface{}{
				"plugin_id": req.PluginId,
				"name":      req.Name,
				"version":   req.Version,
			},
			Timestamp: time.Now(),
		}
		s.eventBus.Publish(event)
	}

	return &RegisterPluginResponse{
		Success: true,
		Message: "Plugin registered successfully",
	}, nil
}

// UnregisterPlugin 注销插件
func (s *PluginServiceServer) UnregisterPlugin(ctx context.Context, req *UnregisterPluginRequest) (*UnregisterPluginResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	plugin, exists := s.plugins[req.PluginId]
	if !exists {
		return &UnregisterPluginResponse{
			Success: false,
			Message: "Plugin not found",
		}, status.Errorf(codes.NotFound, "plugin not found: %s", req.PluginId)
	}

	delete(s.plugins, req.PluginId)
	s.logger.Printf("Plugin unregistered: %s (%s)", plugin.Name, req.PluginId)

	// 发布插件注销事件
	if s.eventBus != nil {
		event := &Event{
			Type:   "plugin.unregistered",
			Source: "grpc-server",
			Data: map[string]interface{}{
				"plugin_id": req.PluginId,
				"name":      plugin.Name,
			},
			Timestamp: time.Now(),
		}
		s.eventBus.Publish(event)
	}

	return &UnregisterPluginResponse{
		Success: true,
		Message: "Plugin unregistered successfully",
	}, nil
}

// HandleRequest 处理请求
func (s *PluginServiceServer) HandleRequest(ctx context.Context, req *PluginRequest) (*PluginResponse, error) {
	s.mutex.RLock()
	plugin, exists := s.plugins[req.PluginId]
	s.mutex.RUnlock()

	if !exists {
		return &PluginResponse{
			Success: false,
			Message: "Plugin not found",
		}, status.Errorf(codes.NotFound, "plugin not found: %s", req.PluginId)
	}

	startTime := time.Now()

	// 更新插件统计
	s.mutex.Lock()
	plugin.Stats.RequestCount++
	plugin.Stats.LastRequestTime = startTime
	s.mutex.Unlock()

	// 这里应该转发请求到实际的插件容器
	// 为了演示，我们返回一个模拟响应
	response := &PluginResponse{
		RequestId: req.RequestId,
		Success:   true,
		Message:   "Request processed successfully",
		Data:      req.Data, // 回显数据
		Timestamp: timestamppb.New(time.Now()),
	}

	// 计算延迟
	latency := time.Since(startTime).Seconds()
	s.mutex.Lock()
	plugin.Stats.TotalLatency += latency
	plugin.Stats.AverageLatency = plugin.Stats.TotalLatency / float64(plugin.Stats.RequestCount)
	s.mutex.Unlock()

	return response, nil
}

// GetPluginHealth 获取插件健康状态
func (s *PluginServiceServer) GetPluginHealth(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	s.mutex.RLock()
	plugin, exists := s.plugins[req.PluginId]
	s.mutex.RUnlock()

	if !exists {
		return &HealthCheckResponse{
			Status:  "NOT_FOUND",
			Message: "Plugin not found",
		}, status.Errorf(codes.NotFound, "plugin not found: %s", req.PluginId)
	}

	// 更新健康检查统计
	s.mutex.Lock()
	plugin.Health.CheckCount++
	plugin.Health.LastCheck = time.Now()
	s.mutex.Unlock()

	return &HealthCheckResponse{
		Status:    plugin.Health.Status,
		Message:   plugin.Health.Message,
		Timestamp: timestamppb.New(plugin.Health.LastCheck),
	}, nil
}

// ListPlugins 列出所有插件
func (s *PluginServiceServer) ListPlugins(ctx context.Context, req *ListPluginsRequest) (*ListPluginsResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	plugins := make([]*PluginInfo, 0, len(s.plugins))
	for _, plugin := range s.plugins {
		pluginInfo := &PluginInfo{
			Id:          plugin.ID,
			Name:        plugin.Name,
			Version:     plugin.Version,
			Status:      plugin.Status,
			ContainerId: plugin.ContainerID,
			Endpoint:    plugin.Endpoint,
			Metadata:    plugin.Metadata,
			CreatedAt:   timestamppb.New(plugin.CreatedAt),
			UpdatedAt:   timestamppb.New(plugin.UpdatedAt),
		}
		plugins = append(plugins, pluginInfo)
	}

	return &ListPluginsResponse{
		Plugins: plugins,
		Total:   int32(len(plugins)),
	}, nil
}

// GetPluginStats 获取插件统计信息
func (s *PluginServiceServer) GetPluginStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	s.mutex.RLock()
	plugin, exists := s.plugins[req.PluginId]
	s.mutex.RUnlock()

	if !exists {
		return &StatsResponse{}, status.Errorf(codes.NotFound, "plugin not found: %s", req.PluginId)
	}

	return &StatsResponse{
		PluginId:        plugin.ID,
		RequestCount:    plugin.Stats.RequestCount,
		ErrorCount:      plugin.Stats.ErrorCount,
		AverageLatency:  plugin.Stats.AverageLatency,
		LastRequestTime: timestamppb.New(plugin.Stats.LastRequestTime),
	}, nil
}

// PublishEvent 发布事件
func (s *PluginServiceServer) PublishEvent(ctx context.Context, req *EventRequest) (*EventResponse, error) {
	if s.eventBus == nil {
		return &EventResponse{
			Success: false,
			Message: "Event bus not available",
		}, status.Errorf(codes.Unavailable, "event bus not available")
	}

	event := &Event{
		ID:        req.EventId,
		Type:      req.EventType,
		Source:    req.Source,
		Data:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// 转换事件数据
	if req.Data != nil {
		var structData structpb.Struct
		if err := req.Data.UnmarshalTo(&structData); err == nil {
			for key, value := range structData.Fields {
				event.Data[key] = value.AsInterface()
			}
		}
	}

	if err := s.eventBus.Publish(event); err != nil {
		return &EventResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to publish event: %v", err),
		}, status.Errorf(codes.Internal, "failed to publish event: %v", err)
	}

	return &EventResponse{
		Success: true,
		Message: "Event published successfully",
	}, nil
}

// unaryInterceptor 统一拦截器
func (s *PluginServiceServer) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 应用中间件
	for _, middleware := range s.middleware {
		resp, err := middleware.Process(ctx, req, info, handler)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			return resp, nil
		}
	}

	// 执行实际的处理器
	return handler(ctx, req)
}

// Process 认证中间件处理
func (m *AuthMiddleware) Process(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 跳过健康检查和反射服务的认证
	if info.FullMethod == "/grpc.health.v1.Health/Check" ||
		info.FullMethod == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo" {
		return handler(ctx, req)
	}

	// 检查认证令牌
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing authorization token")
	}

	token := tokens[0]
	if token != "Bearer "+m.token {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token")
	}

	return handler(ctx, req)
}

// Process 日志中间件处理
func (m *LoggingMiddleware) Process(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	status := "OK"
	if err != nil {
		status = "ERROR"
	}

	m.logger.Printf("Method: %s, Duration: %v, Status: %s", info.FullMethod, duration, status)

	return resp, err
}

// Process 指标中间件处理
func (m *MetricsMiddleware) Process(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	m.mutex.Lock()
	m.requestCount[info.FullMethod]++
	m.mutex.Unlock()

	return handler(ctx, req)
}

// GetMetrics 获取指标
func (m *MetricsMiddleware) GetMetrics() map[string]int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metrics := make(map[string]int64)
	for method, count := range m.requestCount {
		metrics[method] = count
	}

	return metrics
}

// 这些是需要根据实际的protobuf定义来实现的接口
// 以下是模拟的结构体定义
type UnimplementedPluginServiceServer struct{}

func (UnimplementedPluginServiceServer) RegisterPlugin(context.Context, *RegisterPluginRequest) (*RegisterPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterPlugin not implemented")
}

func (UnimplementedPluginServiceServer) UnregisterPlugin(context.Context, *UnregisterPluginRequest) (*UnregisterPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterPlugin not implemented")
}

func (UnimplementedPluginServiceServer) HandleRequest(context.Context, *PluginRequest) (*PluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleRequest not implemented")
}

func (UnimplementedPluginServiceServer) GetPluginHealth(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPluginHealth not implemented")
}

func (UnimplementedPluginServiceServer) ListPlugins(context.Context, *ListPluginsRequest) (*ListPluginsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPlugins not implemented")
}

func (UnimplementedPluginServiceServer) GetPluginStats(context.Context, *StatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPluginStats not implemented")
}

func (UnimplementedPluginServiceServer) PublishEvent(context.Context, *EventRequest) (*EventResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PublishEvent not implemented")
}

// 模拟的请求和响应结构
type RegisterPluginRequest struct {
	PluginId    string            `json:"plugin_id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	ContainerId string            `json:"container_id"`
	Endpoint    string            `json:"endpoint"`
	Metadata    map[string]string `json:"metadata"`
}

type RegisterPluginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UnregisterPluginRequest struct {
	PluginId string `json:"plugin_id"`
}

type UnregisterPluginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type PluginRequest struct {
	RequestId string     `json:"request_id"`
	PluginId  string     `json:"plugin_id"`
	Method    string     `json:"method"`
	Data      *anypb.Any `json:"data"`
}

type PluginResponse struct {
	RequestId string                 `json:"request_id"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      *anypb.Any             `json:"data"`
	Timestamp *timestamppb.Timestamp `json:"timestamp"`
}

type HealthCheckRequest struct {
	PluginId string `json:"plugin_id"`
}

type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Timestamp *timestamppb.Timestamp `json:"timestamp"`
}

type ListPluginsRequest struct {
	Filter string `json:"filter"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

type ListPluginsResponse struct {
	Plugins []*PluginInfo `json:"plugins"`
	Total   int32         `json:"total"`
}

type PluginInfo struct {
	Id          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Status      string                 `json:"status"`
	ContainerId string                 `json:"container_id"`
	Endpoint    string                 `json:"endpoint"`
	Metadata    map[string]string      `json:"metadata"`
	CreatedAt   *timestamppb.Timestamp `json:"created_at"`
	UpdatedAt   *timestamppb.Timestamp `json:"updated_at"`
}

type StatsRequest struct {
	PluginId string `json:"plugin_id"`
}

type StatsResponse struct {
	PluginId        string                 `json:"plugin_id"`
	RequestCount    int64                  `json:"request_count"`
	ErrorCount      int64                  `json:"error_count"`
	AverageLatency  float64                `json:"average_latency"`
	LastRequestTime *timestamppb.Timestamp `json:"last_request_time"`
}

type EventRequest struct {
	EventId   string     `json:"event_id"`
	EventType string     `json:"event_type"`
	Source    string     `json:"source"`
	Data      *anypb.Any `json:"data"`
}

type EventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RegisterPluginServiceServer 注册服务
func RegisterPluginServiceServer(s *grpc.Server, srv *PluginServiceServer) {
	// 这里应该使用实际生成的protobuf代码
	// s.RegisterService(&PluginService_ServiceDesc, srv)
}
