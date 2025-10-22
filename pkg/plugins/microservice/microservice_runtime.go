package microservice

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MicroserviceRuntime 微服务插件运行时
type MicroserviceRuntime struct {
	dockerManager *DockerManager
	grpcServer    *PluginServiceServer
	plugins       map[string]*MicroservicePluginInstance
	mutex         sync.RWMutex
	config        *MicroserviceConfig
	logger        *log.Logger
	eventBus      EventBus
	healthChecker *HealthChecker
	loadBalancer  *LoadBalancer
}

// MicroserviceConfig 微服务运行时配置
type MicroserviceConfig struct {
	Docker       *DockerConfig   `json:"docker"`
	GRPC         *GRPCConfig     `json:"grpc"`
	PluginDir    string          `json:"plugin_dir"`
	ImagePrefix  string          `json:"image_prefix"`
	NetworkName  string          `json:"network_name"`
	HealthCheck  *HealthConfig   `json:"health_check"`
	LoadBalancer *LBConfig       `json:"load_balancer"`
	Security     *SecurityConfig `json:"security"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	Retries     int           `json:"retries"`
	StartPeriod time.Duration `json:"start_period"`
}

// LBConfig 负载均衡配置
type LBConfig struct {
	Strategy    string `json:"strategy"` // round_robin, least_connections, weighted
	HealthCheck bool   `json:"health_check"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableTLS        bool   `json:"enable_tls"`
	CertFile         string `json:"cert_file"`
	KeyFile          string `json:"key_file"`
	EnableAuth       bool   `json:"enable_auth"`
	AuthToken        string `json:"auth_token"`
	NetworkIsolation bool   `json:"network_isolation"`
}

// MicroservicePluginInstance 微服务插件实例
type MicroservicePluginInstance struct {
	ID           string                 `json:"id"`
	PluginID     string                 `json:"plugin_id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Image        string                 `json:"image"`
	Status       string                 `json:"status"`
	ContainerID  string                 `json:"container_id"`
	Endpoint     string                 `json:"endpoint"`
	Port         int                    `json:"port"`
	Config       map[string]interface{} `json:"config"`
	Metadata     map[string]string      `json:"metadata"`
	Health       *PluginHealth          `json:"health"`
	Stats        *PluginStats           `json:"stats"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastAccessed time.Time              `json:"last_accessed"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	runtime  *MicroserviceRuntime
	interval time.Duration
	timeout  time.Duration
	retries  int
	stopCh   chan struct{}
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	strategy  string
	instances map[string][]*MicroservicePluginInstance
	mutex     sync.RWMutex
}

// NewMicroserviceRuntime 创建新的微服务运行时
func NewMicroserviceRuntime(config *MicroserviceConfig, eventBus EventBus) (*MicroserviceRuntime, error) {
	// 创建Docker管理器
	dockerManager, err := NewDockerManager(config.Docker)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker manager: %w", err)
	}

	// 创建gRPC服务端
	grpcServer := NewPluginServiceServer(config.GRPC, eventBus)

	// 创建负载均衡器
	loadBalancer := &LoadBalancer{
		strategy:  config.LoadBalancer.Strategy,
		instances: make(map[string][]*MicroservicePluginInstance),
	}

	runtime := &MicroserviceRuntime{
		dockerManager: dockerManager,
		grpcServer:    grpcServer,
		plugins:       make(map[string]*MicroservicePluginInstance),
		config:        config,
		logger:        log.New(log.Writer(), "[MicroserviceRuntime] ", log.LstdFlags),
		eventBus:      eventBus,
		loadBalancer:  loadBalancer,
	}

	// 创建健康检查器
	runtime.healthChecker = &HealthChecker{
		runtime:  runtime,
		interval: config.HealthCheck.Interval,
		timeout:  config.HealthCheck.Timeout,
		retries:  config.HealthCheck.Retries,
		stopCh:   make(chan struct{}),
	}

	return runtime, nil
}

// Start 启动微服务运行时
func (r *MicroserviceRuntime) Start() error {
	// 启动gRPC服务端
	if err := r.grpcServer.Start(); err != nil {
		return fmt.Errorf("failed to start grpc server: %w", err)
	}

	// 启动健康检查器
	go r.healthChecker.Start()

	r.logger.Println("Microservice runtime started")
	return nil
}

// Stop 停止微服务运行时
func (r *MicroserviceRuntime) Stop() error {
	// 停止健康检查器
	close(r.healthChecker.stopCh)

	// 停止所有插件实例
	r.mutex.RLock()
	instances := make([]*MicroservicePluginInstance, 0, len(r.plugins))
	for _, instance := range r.plugins {
		instances = append(instances, instance)
	}
	r.mutex.RUnlock()

	for _, instance := range instances {
		if err := r.StopPlugin(instance.ID); err != nil {
			r.logger.Printf("Failed to stop plugin %s: %v", instance.ID, err)
		}
	}

	// 关闭Docker管理器
	if err := r.dockerManager.Close(); err != nil {
		r.logger.Printf("Failed to close docker manager: %v", err)
	}

	r.logger.Println("Microservice runtime stopped")
	return nil
}

// LoadPlugin 加载插件
func (r *MicroserviceRuntime) LoadPlugin(pluginPath string, config map[string]interface{}) (*MicroservicePluginInstance, error) {
	// 读取插件清单
	manifestPath := filepath.Join(pluginPath, "manifest.json")
	manifest, err := r.loadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	// 验证插件
	if err := r.validatePlugin(manifest); err != nil {
		return nil, fmt.Errorf("plugin validation failed: %w", err)
	}

	// 创建插件实例
	instance := &MicroservicePluginInstance{
		ID:       uuid.New().String(),
		PluginID: manifest.ID,
		Name:     manifest.Name,
		Version:  manifest.Version,
		Image:    r.buildImageName(manifest),
		Status:   "loading",
		Config:   config,
		Metadata: manifest.Metadata,
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
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}

	// 分配端口
	port, err := r.allocatePort()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}
	instance.Port = port
	instance.Endpoint = fmt.Sprintf("localhost:%d", port)

	// 创建容器选项
	containerOptions := &ContainerCreateOptions{
		Image: instance.Image,
		Name:  fmt.Sprintf("plugin-%s-%s", instance.PluginID, instance.ID[:8]),
		Ports: map[string]string{
			"8080": fmt.Sprintf("%d", port),
		},
		Environment: r.buildEnvironment(instance, manifest),
		Labels: map[string]string{
			"plugin.id":      instance.PluginID,
			"instance.id":    instance.ID,
			"plugin.name":    instance.Name,
			"plugin.version": instance.Version,
		},
		Resources: manifest.Resources,
	}

	// 创建容器
	containerInfo, err := r.dockerManager.CreateContainer(instance.PluginID, containerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	instance.ContainerID = containerInfo.ID
	instance.Status = "created"

	// 保存实例
	r.mutex.Lock()
	r.plugins[instance.ID] = instance
	r.mutex.Unlock()

	// 添加到负载均衡器
	r.loadBalancer.AddInstance(instance.PluginID, instance)

	r.logger.Printf("Plugin loaded: %s (%s)", instance.Name, instance.ID)

	// 发布事件
	if r.eventBus != nil {
		event := &Event{
			Type:   "plugin.loaded",
			Source: "microservice-runtime",
			Data: map[string]interface{}{
				"instance_id": instance.ID,
				"plugin_id":   instance.PluginID,
				"name":        instance.Name,
				"version":     instance.Version,
			},
			Timestamp: time.Now(),
		}
		r.eventBus.Publish(event)
	}

	return instance, nil
}

// StartPlugin 启动插件
func (r *MicroserviceRuntime) StartPlugin(instanceID string) error {
	r.mutex.RLock()
	instance, exists := r.plugins[instanceID]
	r.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin instance not found: %s", instanceID)
	}

	// 启动容器
	if err := r.dockerManager.StartContainer(instance.ContainerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// 更新状态为starting
	r.mutex.Lock()
	instance.Status = "starting"
	instance.UpdatedAt = time.Now()
	r.mutex.Unlock()

	// 等待插件就绪
	if err := r.waitForPluginReady(instance); err != nil {
		return fmt.Errorf("plugin failed to start: %w", err)
	}

	// 更新状态为running
	r.mutex.Lock()
	instance.Status = "running"
	instance.UpdatedAt = time.Now()
	r.mutex.Unlock()

	r.logger.Printf("Plugin started: %s (%s)", instance.Name, instance.ID)

	// 发布事件
	if r.eventBus != nil {
		event := &Event{
			Type:   "plugin.started",
			Source: "microservice-runtime",
			Data: map[string]interface{}{
				"instance_id": instance.ID,
				"plugin_id":   instance.PluginID,
				"endpoint":    instance.Endpoint,
			},
			Timestamp: time.Now(),
		}
		r.eventBus.Publish(event)
	}

	return nil
}

// StopPlugin 停止插件
func (r *MicroserviceRuntime) StopPlugin(instanceID string) error {
	r.mutex.RLock()
	instance, exists := r.plugins[instanceID]
	r.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin instance not found: %s", instanceID)
	}

	// 停止容器
	if err := r.dockerManager.StopContainer(instance.ContainerID); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// 更新状态为stopped
	r.mutex.Lock()
	instance.Status = "stopped"
	instance.UpdatedAt = time.Now()
	r.mutex.Unlock()

	r.logger.Printf("Plugin stopped: %s (%s)", instance.Name, instance.ID)

	// 发布事件
	if r.eventBus != nil {
		event := &Event{
			Type:   "plugin.stopped",
			Source: "microservice-runtime",
			Data: map[string]interface{}{
				"instance_id": instance.ID,
				"plugin_id":   instance.PluginID,
			},
			Timestamp: time.Now(),
		}
		r.eventBus.Publish(event)
	}

	return nil
}

// UnloadPlugin 卸载插件
func (r *MicroserviceRuntime) UnloadPlugin(instanceID string) error {
	r.mutex.RLock()
	instance, exists := r.plugins[instanceID]
	r.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin instance not found: %s", instanceID)
	}

	// 先停止插件实例
	if instance.Status == "running" {
		if err := r.StopPlugin(instanceID); err != nil {
			r.logger.Printf("Failed to stop plugin before unload: %v", err)
		}
	}

	// 删除容器
	if err := r.dockerManager.RemoveContainer(instance.ContainerID); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// 从负载均衡器移除
	r.loadBalancer.RemoveInstance(instance.PluginID, instance.ID)

	// 删除实例
	r.mutex.Lock()
	delete(r.plugins, instanceID)
	r.mutex.Unlock()

	r.logger.Printf("Plugin unloaded: %s (%s)", instance.Name, instance.ID)

	// 发布事件
	if r.eventBus != nil {
		event := &Event{
			Type:   "plugin.unloaded",
			Source: "microservice-runtime",
			Data: map[string]interface{}{
				"instance_id": instance.ID,
				"plugin_id":   instance.PluginID,
			},
			Timestamp: time.Now(),
		}
		r.eventBus.Publish(event)
	}

	return nil
}

// GetPlugin 获取插件实例
func (r *MicroserviceRuntime) GetPlugin(instanceID string) (*MicroservicePluginInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.plugins[instanceID]
	if !exists {
		return nil, fmt.Errorf("plugin instance not found: %s", instanceID)
	}

	return instance, nil
}

// ListPlugins 列出所有插件实例
func (r *MicroserviceRuntime) ListPlugins() []*MicroservicePluginInstance {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instances := make([]*MicroservicePluginInstance, 0, len(r.plugins))
	for _, instance := range r.plugins {
		instances = append(instances, instance)
	}

	return instances
}

// GetPluginsByType 根据插件ID获取实例
func (r *MicroserviceRuntime) GetPluginsByType(pluginID string) []*MicroservicePluginInstance {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instances := make([]*MicroservicePluginInstance, 0)
	for _, instance := range r.plugins {
		if instance.PluginID == pluginID {
			instances = append(instances, instance)
		}
	}

	return instances
}

// RouteRequest 路由请求到插件实例
func (r *MicroserviceRuntime) RouteRequest(pluginID string, request interface{}) (interface{}, error) {
	// 获取可用实例
	instance := r.loadBalancer.SelectInstance(pluginID)
	if instance == nil {
		return nil, fmt.Errorf("no available instance for plugin: %s", pluginID)
	}

	// 更新访问时间
	r.mutex.Lock()
	instance.LastAccessed = time.Now()
	instance.Stats.RequestCount++
	r.mutex.Unlock()

	// 这里应该通过gRPC或HTTP调用实际的插件实例
	// 为了演示，我们返回一个模拟响应
	response := map[string]interface{}{
		"instance_id": instance.ID,
		"plugin_id":   instance.PluginID,
		"endpoint":    instance.Endpoint,
		"request":     request,
		"timestamp":   time.Now(),
	}

	return response, nil
}

// 辅助方法

func (r *MicroserviceRuntime) loadManifest(path string) (*PluginManifest, error) {
	// 这里应该读取并解析manifest.json文件
	// 为了演示，返回一个模拟的清单
	return &PluginManifest{
		ID:       "example-plugin",
		Name:     "Example Plugin",
		Version:  "1.0.0",
		Type:     "microservice",
		Runtime:  "docker",
		Metadata: make(map[string]string),
		Resources: &ResourceLimits{
			Memory:    128 * 1024 * 1024, // 128MB
			CPUShares: 512,
		},
	}, nil
}

func (r *MicroserviceRuntime) validatePlugin(manifest *PluginManifest) error {
	if manifest.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.Runtime != "docker" {
		return fmt.Errorf("unsupported runtime: %s", manifest.Runtime)
	}
	return nil
}

func (r *MicroserviceRuntime) buildImageName(manifest *PluginManifest) string {
	return fmt.Sprintf("%s/%s:%s", r.config.ImagePrefix, manifest.ID, manifest.Version)
}

func (r *MicroserviceRuntime) allocatePort() (int, error) {
	// 简单的端口分配策略，从8000开始递增
	// 实际实现应该检查端口是否被占用
	return 8000 + len(r.plugins), nil
}

func (r *MicroserviceRuntime) buildEnvironment(instance *MicroservicePluginInstance, manifest *PluginManifest) map[string]string {
	env := make(map[string]string)
	env["PLUGIN_ID"] = instance.PluginID
	env["INSTANCE_ID"] = instance.ID
	env["PLUGIN_NAME"] = instance.Name
	env["PLUGIN_VERSION"] = instance.Version
	env["GRPC_SERVER"] = fmt.Sprintf("%s:%d", r.config.GRPC.Host, r.config.GRPC.Port)

	// 添加配置
	for key, value := range instance.Config {
		if str, ok := value.(string); ok {
			env[fmt.Sprintf("CONFIG_%s", key)] = str
		}
	}

	return env
}

func (r *MicroserviceRuntime) waitForPluginReady(instance *MicroservicePluginInstance) error {
	// 等待插件就绪的逻辑
	// 这里应该检查插件的健康状态
	// 为了演示，我们简单地等待容器状态变为running
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for plugin to be ready")
		case <-ticker.C:
			// 检查容器状态是否为running
			containerInfo, err := r.dockerManager.GetContainer(instance.ContainerID)
			if err != nil {
				continue
			}
			if containerInfo.Status == "running" {
				return nil
			}
		}
	}
}

// PluginManifest 插件清单
type PluginManifest struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Type      string            `json:"type"`
	Runtime   string            `json:"runtime"`
	Metadata  map[string]string `json:"metadata"`
	Resources *ResourceLimits   `json:"resources"`
}

// 健康检查器方法

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.checkAllPlugins()
		}
	}
}

func (hc *HealthChecker) checkAllPlugins() {
	instances := hc.runtime.ListPlugins()
	for _, instance := range instances {
		if instance.Status == "running" {
			hc.checkPluginHealth(instance)
		}
	}
}

func (hc *HealthChecker) checkPluginHealth(instance *MicroservicePluginInstance) {
	// 实现健康检查逻辑
	// 这里应该调用插件的健康检查端点
	// 为了演示，我们简单地模拟一个健康检查结果
	hc.runtime.mutex.Lock()
	instance.Health.CheckCount++
	instance.Health.LastCheck = time.Now()
	// 模拟健康检查结果
	instance.Health.Status = "healthy"
	hc.runtime.mutex.Unlock()
}

// 负载均衡器方法

func (lb *LoadBalancer) AddInstance(pluginID string, instance *MicroservicePluginInstance) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if lb.instances[pluginID] == nil {
		lb.instances[pluginID] = make([]*MicroservicePluginInstance, 0)
	}
	lb.instances[pluginID] = append(lb.instances[pluginID], instance)
}

func (lb *LoadBalancer) RemoveInstance(pluginID, instanceID string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	instances := lb.instances[pluginID]
	for i, instance := range instances {
		if instance.ID == instanceID {
			lb.instances[pluginID] = append(instances[:i], instances[i+1:]...)
			break
		}
	}
}

func (lb *LoadBalancer) SelectInstance(pluginID string) *MicroservicePluginInstance {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	instances := lb.instances[pluginID]
	if len(instances) == 0 {
		return nil
	}

	// 简单的轮询策略
	// 实际实现应该根据配置的策略选择实例
	return instances[0]
}
