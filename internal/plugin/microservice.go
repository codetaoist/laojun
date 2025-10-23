package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	typesimage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MicroservicePlugin 微服务插件模型
type MicroservicePlugin struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	DockerImage     string                 `json:"docker_image"`
	ContainerID     string                 `json:"container_id"`
	ServicePort     int                    `json:"service_port"`
	HealthCheckPath string                 `json:"health_check_path"`
	GRPCProtoFile   string                 `json:"grpc_proto_file"`
	ServiceEndpoint string                 `json:"service_endpoint"`
	Status          string                 `json:"status"`
	Config          map[string]interface{} `json:"config"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// MicroservicePluginManager 微服务插件管理器
type MicroservicePluginManager struct {
	dockerClient *client.Client
	plugins      map[string]*MicroservicePlugin
	mutex        sync.RWMutex
	logger       *logrus.Logger
	networkName  string
}

// NewMicroservicePluginManager 创建微服务插件管理器
func NewMicroservicePluginManager(logger *logrus.Logger) (*MicroservicePluginManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	manager := &MicroservicePluginManager{
		dockerClient: dockerClient,
		plugins:      make(map[string]*MicroservicePlugin),
		logger:       logger,
		networkName:  "plugin-network",
	}

	// 创建插件网络
	if err := manager.createPluginNetwork(); err != nil {
		logger.Warnf("Failed to create plugin network: %v", err)
	}

	return manager, nil
}

// createPluginNetwork 创建插件网络
func (m *MicroservicePluginManager) createPluginNetwork() error {
	ctx := context.Background()

	// 检查网络是否已存在
	networks, err := m.dockerClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return err
	}

	for _, net := range networks {
		if net.Name == m.networkName {
			m.logger.Infof("Plugin network %s already exists", m.networkName)
			return nil
		}
	}

	// 创建网络
	_, err = m.dockerClient.NetworkCreate(ctx, m.networkName, network.CreateOptions{
		Driver: "bridge",
		Options: map[string]string{
			"com.docker.network.bridge.name": m.networkName,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create plugin network: %w", err)
	}

	m.logger.Infof("Created plugin network: %s", m.networkName)
	return nil
}

// DeployPlugin 部署微服务插件
func (m *MicroservicePluginManager) DeployPlugin(pluginConfig *MicroservicePlugin) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ctx := context.Background()

	// 拉取Docker镜像
	if err := m.pullDockerImage(ctx, pluginConfig.DockerImage); err != nil {
		return fmt.Errorf("failed to pull Docker image: %w", err)
	}

	// 创建容器配置
	containerConfig := &container.Config{
		Image: pluginConfig.DockerImage,
		Env:   m.buildEnvironmentVariables(pluginConfig.Config),
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", pluginConfig.ServicePort)): struct{}{},
		},
		Labels: map[string]string{
			"plugin.id":      pluginConfig.ID,
			"plugin.name":    pluginConfig.Name,
			"plugin.version": pluginConfig.Version,
			"plugin.type":    "microservice",
		},
	}

	// 创建主机配置
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(m.networkName),
		Resources: container.Resources{
			Memory:   512 * 1024 * 1024, // 512MB
			CPUQuota: 50000,             // 50% CPU
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// 创建网络配置
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			m.networkName: {
				Aliases: []string{pluginConfig.Name},
			},
		},
	}

	// 创建容器
	resp, err := m.dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		fmt.Sprintf("plugin-%s", pluginConfig.ID),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	pluginConfig.ContainerID = resp.ID

	// 启动容器
	if err := m.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// 等待容器启动并进行健康检查
	if err := m.waitForPluginReady(pluginConfig); err != nil {
		// 如果启动失败，清理容器
		if err := m.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); err != nil {
			m.logger.Warnf("Failed to remove container %s after startup failure: %v", resp.ID, err)
		}
		return fmt.Errorf("plugin failed to start: %w", err)
	}

	// 获取容器信息更新服务端点
	containerInfo, err := m.dockerClient.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	// 更新服务端点
	if networkSettings, ok := containerInfo.NetworkSettings.Networks[m.networkName]; ok {
		pluginConfig.ServiceEndpoint = fmt.Sprintf("%s:%d", networkSettings.IPAddress, pluginConfig.ServicePort)
	}

	pluginConfig.Status = "running"
	pluginConfig.CreatedAt = time.Now()
	pluginConfig.UpdatedAt = time.Now()

	// 存储插件信息
	m.plugins[pluginConfig.ID] = pluginConfig

	m.logger.Infof("Successfully deployed microservice plugin: %s (Container: %s)", pluginConfig.Name, resp.ID)
	return nil
}

// pullDockerImage 拉取Docker镜像
func (m *MicroservicePluginManager) pullDockerImage(ctx context.Context, image string) error {
	reader, err := m.dockerClient.ImagePull(ctx, image, typesimage.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// 读取拉取进度（简化处理）
	_, err = io.Copy(io.Discard, reader)
	return err
}

// buildEnvironmentVariables 构建环境变量
func (m *MicroservicePluginManager) buildEnvironmentVariables(config map[string]interface{}) []string {
	var envVars []string

	for key, value := range config {
		envVars = append(envVars, fmt.Sprintf("%s=%v", strings.ToUpper(key), value))
	}

	return envVars
}

// waitForPluginReady 等待插件就绪
func (m *MicroservicePluginManager) waitForPluginReady(plugin *MicroservicePlugin) error {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("plugin startup timeout")
		case <-ticker.C:
			if err := m.checkPluginHealth(plugin); err == nil {
				return nil
			}
		}
	}
}

// checkPluginHealth 检查插件健康状态
func (m *MicroservicePluginManager) checkPluginHealth(plugin *MicroservicePlugin) error {
	if plugin.ServiceEndpoint == "" {
		return fmt.Errorf("service endpoint not available")
	}

	// HTTP健康检查
	if plugin.HealthCheckPath != "" {
		url := fmt.Sprintf("http://%s%s", plugin.ServiceEndpoint, plugin.HealthCheckPath)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
		}
	}

	return nil
}

// StartPlugin 启动插件
func (m *MicroservicePluginManager) StartPlugin(pluginID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	ctx := context.Background()

	// 启动容器
	if err := m.dockerClient.ContainerStart(ctx, plugin.ContainerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// 等待插件就绪
	if err := m.waitForPluginReady(plugin); err != nil {
		return fmt.Errorf("plugin failed to start: %w", err)
	}

	plugin.Status = "running"
	plugin.UpdatedAt = time.Now()

	m.logger.Infof("Successfully started microservice plugin: %s", plugin.Name)
	return nil
}

// StopPlugin 停止插件
func (m *MicroservicePluginManager) StopPlugin(pluginID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	ctx := context.Background()

	// 停止容器
	timeoutSeconds := 30
	if err := m.dockerClient.ContainerStop(ctx, plugin.ContainerID, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	plugin.Status = "stopped"
	plugin.UpdatedAt = time.Now()

	m.logger.Infof("Successfully stopped microservice plugin: %s", plugin.Name)
	return nil
}

// RemovePlugin 移除插件
func (m *MicroservicePluginManager) RemovePlugin(pluginID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	ctx := context.Background()

	// 移除容器
	if err := m.dockerClient.ContainerRemove(ctx, plugin.ContainerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// 从映射中删除
	delete(m.plugins, pluginID)

	m.logger.Infof("Successfully removed microservice plugin: %s", plugin.Name)
	return nil
}

// RestartPlugin 重启插件
func (m *MicroservicePluginManager) RestartPlugin(pluginID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	ctx := context.Background()

	// 重启容器
	timeoutSeconds := 30
	stopOptions := container.StopOptions{
		Timeout: &timeoutSeconds,
	}
	if err := m.dockerClient.ContainerRestart(ctx, plugin.ContainerID, stopOptions); err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}

	// 等待插件就绪
	if err := m.waitForPluginReady(plugin); err != nil {
		return fmt.Errorf("plugin failed to restart: %w", err)
	}

	plugin.Status = "running"
	plugin.UpdatedAt = time.Now()

	m.logger.Infof("Successfully restarted microservice plugin: %s", plugin.Name)
	return nil
}

// GetPluginStatus 获取插件状态
func (m *MicroservicePluginManager) GetPluginStatus(pluginID string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return "", fmt.Errorf("plugin not found: %s", pluginID)
	}

	ctx := context.Background()

	// 检查容器状态
	containerInfo, err := m.dockerClient.ContainerInspect(ctx, plugin.ContainerID)
	if err != nil {
		return "error", err
	}

	if containerInfo.State.Running {
		plugin.Status = "running"
	} else {
		plugin.Status = "stopped"
	}

	plugin.UpdatedAt = time.Now()
	return plugin.Status, nil
}

// CallPlugin 调用插件服务
func (m *MicroservicePluginManager) CallPlugin(pluginID string, method string, params map[string]interface{}) (*PluginResult, error) {
	m.mutex.RLock()
	plugin, exists := m.plugins[pluginID]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	if plugin.Status != "running" {
		return nil, fmt.Errorf("plugin is not running: %s", plugin.Status)
	}

	startTime := time.Now()

	// 根据插件类型选择调用方式
	var result *PluginResult
	var err error

	if strings.Contains(plugin.GRPCProtoFile, ".proto") {
		// gRPC调用
		result, err = m.callPluginGRPC(plugin, method, params)
	} else {
		// HTTP REST调用
		result, err = m.callPluginHTTP(plugin, method, params)
	}

	if result != nil {
		result.Duration = time.Since(startTime)
	}

	return result, err
}

// callPluginGRPC 通过gRPC调用插件
func (m *MicroservicePluginManager) callPluginGRPC(plugin *MicroservicePlugin, method string, params map[string]interface{}) (*PluginResult, error) {
	// 建立gRPC连接
	conn, err := grpc.Dial(plugin.ServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC service: %w", err)
	}
	defer conn.Close()

	// 这里需要根据具体的proto文件生成的代码来实现
	// 简化示例，实际需要动态调用生成的gRPC客户端代码
	result := &PluginResult{
		Success: true,
		Data:    map[string]interface{}{"message": "gRPC call successful"},
	}

	return result, nil
}

// callPluginHTTP 通过HTTP调用插件
func (m *MicroservicePluginManager) callPluginHTTP(plugin *MicroservicePlugin, method string, params map[string]interface{}) (*PluginResult, error) {
	url := fmt.Sprintf("http://%s/api/%s", plugin.ServiceEndpoint, method)

	// 构建请求体
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送HTTP请求
	resp, err := http.Post(url, "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to call plugin HTTP API: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应
	var responseData map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result := &PluginResult{
		Success: resp.StatusCode == http.StatusOK,
		Data:    responseData,
	}

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	return result, nil
}

// GetPluginMetrics 获取插件指标
func (m *MicroservicePluginManager) GetPluginMetrics(pluginID string) (map[string]interface{}, error) {
	m.mutex.RLock()
	plugin, exists := m.plugins[pluginID]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	ctx := context.Background()

	// 获取容器统计信息
	stats, err := m.dockerClient.ContainerStats(ctx, plugin.ContainerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	// 读取统计数据
	var containerStats container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&containerStats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	// 构建指标数据
	metrics := map[string]interface{}{
		"plugin_id":    plugin.ID,
		"plugin_name":  plugin.Name,
		"status":       plugin.Status,
		"uptime":       time.Since(plugin.CreatedAt).Seconds(),
		"memory_usage": containerStats.MemoryStats.Usage,
		"memory_limit": containerStats.MemoryStats.Limit,
		"cpu_usage":    containerStats.CPUStats.CPUUsage.TotalUsage,
		"network_rx":   containerStats.Networks["eth0"].RxBytes,
		"network_tx":   containerStats.Networks["eth0"].TxBytes,
		"last_updated": plugin.UpdatedAt,
	}

	return metrics, nil
}

// ListPlugins 列出所有插件
func (m *MicroservicePluginManager) ListPlugins() []*MicroservicePlugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var plugins []*MicroservicePlugin
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// GetPlugin 获取插件信息
func (m *MicroservicePluginManager) GetPlugin(pluginID string) (*MicroservicePlugin, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return plugin, nil
}

// Close 关闭管理插件
func (m *MicroservicePluginManager) Close() error {
	// 停止所有插件
	for pluginID := range m.plugins {
		if err := m.StopPlugin(pluginID); err != nil {
			m.logger.Warnf("Failed to stop plugin %s: %v", pluginID, err)
		}
	}

	// 关闭Docker客户端连接
	return m.dockerClient.Close()
}
