package microservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	typesimage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerManager 管理插件容器的生命周期
type DockerManager struct {
	client     *client.Client
	containers map[string]*ContainerInfo
	mutex      sync.RWMutex
	config     *DockerConfig
	network    string
	logger     *log.Logger
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"`
	Ports       map[string]string `json:"ports"`
	Environment map[string]string `json:"environment"`
	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   time.Time         `json:"started_at"`
	StoppedAt   time.Time         `json:"stopped_at"`
	PluginID    string            `json:"plugin_id"`
	Health      *HealthStatus     `json:"health"`
	Resources   *ResourceUsage    `json:"resources"`
}

// DockerConfig Docker管理器配置
type DockerConfig struct {
	Host           string            `json:"host"`
	Version        string            `json:"version"`
	NetworkName    string            `json:"network_name"`
	ImageRegistry  string            `json:"image_registry"`
	PullTimeout    time.Duration     `json:"pull_timeout"`
	StartTimeout   time.Duration     `json:"start_timeout"`
	StopTimeout    time.Duration     `json:"stop_timeout"`
	HealthTimeout  time.Duration     `json:"health_timeout"`
	DefaultEnv     map[string]string `json:"default_env"`
	ResourceLimits *ResourceLimits   `json:"resource_limits"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	Memory     int64   `json:"memory"`      // 内存限制（字节）
	CPUShares  int64   `json:"cpu_shares"`  // CPU份额
	CPUQuota   int64   `json:"cpu_quota"`   // CPU配额
	CPUPeriod  int64   `json:"cpu_period"`  // CPU周期
	PidsLimit  *int64  `json:"pids_limit"`  // 进程数限制
	DiskLimit  int64   `json:"disk_limit"`  // 磁盘限制（字节）
	DiskQuota  int64   `json:"disk_quota"`  // 磁盘配额
	NetworkBPS int64   `json:"network_bps"` // 网络带宽限制
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status      string    `json:"status"`
	LastCheck   time.Time `json:"last_check"`
	FailureCount int      `json:"failure_count"`
	Output      string    `json:"output"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   int64   `json:"memory_usage"`
	MemoryLimit   int64   `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     int64   `json:"network_rx"`
	NetworkTx     int64   `json:"network_tx"`
	BlockRead     int64   `json:"block_read"`
	BlockWrite    int64   `json:"block_write"`
	PidsCount     int     `json:"pids_count"`
}

// ContainerCreateOptions 容器创建选项
type ContainerCreateOptions struct {
	Image       string            `json:"image"`
	Name        string            `json:"name"`
	Ports       map[string]string `json:"ports"`
	Environment map[string]string `json:"environment"`
	Volumes     map[string]string `json:"volumes"`
	Command     []string          `json:"command"`
	WorkingDir  string            `json:"working_dir"`
	User        string            `json:"user"`
	Labels      map[string]string `json:"labels"`
	HealthCheck *HealthCheckConfig `json:"health_check"`
	Resources   *ResourceLimits   `json:"resources"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Test        []string      `json:"test"`
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	StartPeriod time.Duration `json:"start_period"`
	Retries     int           `json:"retries"`
}

// NewDockerManager 创建新的Docker管理器
func NewDockerManager(config *DockerConfig) (*DockerManager, error) {
	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(
		client.WithHost(config.Host),
		client.WithVersion(config.Version),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping docker daemon: %w", err)
	}

	dm := &DockerManager{
		client:     cli,
		containers: make(map[string]*ContainerInfo),
		config:     config,
		network:    config.NetworkName,
		logger:     log.New(log.Writer(), "[DockerManager] ", log.LstdFlags),
	}

	// 确保网络存在
	if err := dm.ensureNetwork(); err != nil {
		return nil, fmt.Errorf("failed to ensure network: %w", err)
	}

	// 启动资源监控
	go dm.startResourceMonitoring()

	return dm, nil
}

// CreateContainer 创建容器
func (dm *DockerManager) CreateContainer(pluginID string, options *ContainerCreateOptions) (*ContainerInfo, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	ctx := context.Background()

	// 检查镜像是否存在，不存在则拉取
	if err := dm.ensureImage(ctx, options.Image); err != nil {
		return nil, fmt.Errorf("failed to ensure image: %w", err)
	}

	// 准备端口映射
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	
	for containerPort, hostPort := range options.Ports {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return nil, fmt.Errorf("invalid port %s: %w", containerPort, err)
		}
		
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	// 准备环境变量
	env := make([]string, 0)
	for key, value := range dm.config.DefaultEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	for key, value := range options.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// 准备卷挂载点
	binds := make([]string, 0)
	for hostPath, containerPath := range options.Volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	// 准备标签
	labels := make(map[string]string)
	labels["plugin.id"] = pluginID
	labels["managed.by"] = "taishanglaojun"
	for key, value := range options.Labels {
		labels[key] = value
	}

	// 准备资源限制
	resources := container.Resources{}
	if options.Resources != nil {
		resources.Memory = options.Resources.Memory
		resources.CPUShares = options.Resources.CPUShares
		resources.CPUQuota = options.Resources.CPUQuota
		resources.CPUPeriod = options.Resources.CPUPeriod
		if options.Resources.PidsLimit != nil {
			resources.PidsLimit = options.Resources.PidsLimit
		}
	} else if dm.config.ResourceLimits != nil {
		resources.Memory = dm.config.ResourceLimits.Memory
		resources.CPUShares = dm.config.ResourceLimits.CPUShares
		resources.CPUQuota = dm.config.ResourceLimits.CPUQuota
		resources.CPUPeriod = dm.config.ResourceLimits.CPUPeriod
		if dm.config.ResourceLimits.PidsLimit != nil {
			resources.PidsLimit = dm.config.ResourceLimits.PidsLimit
		}
	}

	// 准备健康检查配置
	var healthcheck *container.HealthConfig
	if options.HealthCheck != nil {
		healthcheck = &container.HealthConfig{
			Test:        options.HealthCheck.Test,
			Interval:    options.HealthCheck.Interval,
			Timeout:     options.HealthCheck.Timeout,
			StartPeriod: options.HealthCheck.StartPeriod,
			Retries:     options.HealthCheck.Retries,
		}
	}

	// 创建容器配置
	config := &container.Config{
		Image:        options.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
		Labels:       labels,
		Cmd:          options.Command,
		WorkingDir:   options.WorkingDir,
		User:         options.User,
		Healthcheck:  healthcheck,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		Resources:    resources,
		NetworkMode:  container.NetworkMode(dm.network),
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			dm.network: {},
		},
	}

	// 创建容器
	resp, err := dm.client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkConfig,
		nil,
		options.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// 创建容器信息
	containerInfo := &ContainerInfo{
		ID:          resp.ID,
		Name:        options.Name,
		Image:       options.Image,
		Status:      "created",
		Ports:       options.Ports,
		Environment: options.Environment,
		CreatedAt:   time.Now(),
		PluginID:    pluginID,
		Health: &HealthStatus{
			Status:      "unknown",
			LastCheck:   time.Now(),
			FailureCount: 0,
		},
		Resources: &ResourceUsage{},
	}

	dm.containers[resp.ID] = containerInfo
	dm.logger.Printf("Container created: %s (%s) for plugin %s", options.Name, resp.ID[:12], pluginID)

	return containerInfo, nil
}

// StartContainer 启动容器
func (dm *DockerManager) StartContainer(containerID string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), dm.config.StartTimeout)
	defer cancel()

	if err := dm.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	if info, exists := dm.containers[containerID]; exists {
		info.Status = "running"
		info.StartedAt = time.Now()
		dm.logger.Printf("Container started: %s (%s)", info.Name, containerID[:12])
	}

	return nil
}

// StopContainer 停止容器
func (dm *DockerManager) StopContainer(containerID string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), dm.config.StopTimeout)
	defer cancel()

	if err := dm.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	if info, exists := dm.containers[containerID]; exists {
		info.Status = "stopped"
		info.StoppedAt = time.Now()
		dm.logger.Printf("Container stopped: %s (%s)", info.Name, containerID[:12])
	}

	return nil
}

// RemoveContainer 删除容器
func (dm *DockerManager) RemoveContainer(containerID string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	ctx := context.Background()

	// 先停止容器
	if err := dm.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		dm.logger.Printf("Warning: failed to stop container before removal: %v", err)
	}

	// 删除容器
	if err := dm.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	if info, exists := dm.containers[containerID]; exists {
		dm.logger.Printf("Container removed: %s (%s)", info.Name, containerID[:12])
		delete(dm.containers, containerID)
	}

	return nil
}

// GetContainer 获取容器信息
func (dm *DockerManager) GetContainer(containerID string) (*ContainerInfo, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	info, exists := dm.containers[containerID]
	if !exists {
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	return info, nil
}

// ListContainers 列出所有容器
func (dm *DockerManager) ListContainers() []*ContainerInfo {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	containers := make([]*ContainerInfo, 0, len(dm.containers))
	for _, info := range dm.containers {
		containers = append(containers, info)
	}

	return containers
}

// GetContainersByPlugin 根据插件ID获取容器
func (dm *DockerManager) GetContainersByPlugin(pluginID string) []*ContainerInfo {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	containers := make([]*ContainerInfo, 0)
	for _, info := range dm.containers {
		if info.PluginID == pluginID {
			containers = append(containers, info)
		}
	}

	return containers
}

// GetContainerLogs 获取容器日志
func (dm *DockerManager) GetContainerLogs(containerID string, tail int) (string, error) {
	ctx := context.Background()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
		Timestamps: true,
	}

	reader, err := dm.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read container logs: %w", err)
	}

	return string(logs), nil
}

// ensureImage 确保镜像存在
func (dm *DockerManager) ensureImage(ctx context.Context, image string) error {
	// 检查镜像是否存在
	_, _, err := dm.client.ImageInspectWithRaw(ctx, image)
	if err == nil {
		return nil // 镜像已存在
	}

	// 拉取镜像
	dm.logger.Printf("Pulling image: %s", image)
	
	pullCtx, cancel := context.WithTimeout(ctx, dm.config.PullTimeout)
	defer cancel()

	reader, err := dm.client.ImagePull(pullCtx, image, typesimage.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	// 等待拉取完成
	_, err = io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read pull response: %w", err)
	}

	dm.logger.Printf("Image pulled successfully: %s", image)
	return nil
}

// ensureNetwork 确保网络存在
func (dm *DockerManager) ensureNetwork() error {
	ctx := context.Background()

	// 检查网络是否存在
	networks, err := dm.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	for _, net := range networks {
		if net.Name == dm.network {
			return nil // 网络已存在
		}
	}

	// 创建网络
	dm.logger.Printf("Creating network: %s", dm.network)
	
	_, err = dm.client.NetworkCreate(ctx, dm.network, network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			"managed.by": "taishanglaojun",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	dm.logger.Printf("Network created successfully: %s", dm.network)
	return nil
}

// startResourceMonitoring 启动资源监控
func (dm *DockerManager) startResourceMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dm.updateResourceUsage()
	}
}

// updateResourceUsage 更新资源使用情况
func (dm *DockerManager) updateResourceUsage() {
	dm.mutex.RLock()
	containerIDs := make([]string, 0, len(dm.containers))
	for id := range dm.containers {
		containerIDs = append(containerIDs, id)
	}
	dm.mutex.RUnlock()

	for _, containerID := range containerIDs {
		if err := dm.updateContainerStats(containerID); err != nil {
			dm.logger.Printf("Failed to update stats for container %s: %v", containerID[:12], err)
		}
	}
}

// updateContainerStats 更新单个容器的统计信息
func (dm *DockerManager) updateContainerStats(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := dm.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return err
	}
	defer stats.Body.Close()

	var statsData container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsData); err != nil {
		return err
	}

	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if info, exists := dm.containers[containerID]; exists {
		// 计算CPU使用百分比
		cpuPercent := calculateCPUPercent(&statsData)
		
		// 计算内存使用情况
		memoryUsage := int64(statsData.MemoryStats.Usage)
		memoryLimit := int64(statsData.MemoryStats.Limit)
		memoryPercent := float64(memoryUsage) / float64(memoryLimit) * 100

		// 网络统计
		var networkRx, networkTx int64
		for _, network := range statsData.Networks {
			networkRx += int64(network.RxBytes)
			networkTx += int64(network.TxBytes)
		}

		// 磁盘IO统计
		var blockRead, blockWrite int64
		for _, blkio := range statsData.BlkioStats.IoServiceBytesRecursive {
			if strings.ToLower(blkio.Op) == "read" {
				blockRead += int64(blkio.Value)
			} else if strings.ToLower(blkio.Op) == "write" {
				blockWrite += int64(blkio.Value)
			}
		}

		info.Resources = &ResourceUsage{
			CPUPercent:    cpuPercent,
			MemoryUsage:   memoryUsage,
			MemoryLimit:   memoryLimit,
			MemoryPercent: memoryPercent,
			NetworkRx:     networkRx,
			NetworkTx:     networkTx,
			BlockRead:     blockRead,
			BlockWrite:    blockWrite,
			PidsCount:     int(statsData.PidsStats.Current),
		}
	}

	return nil
}

// calculateCPUPercent 计算CPU使用百分比
func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	
	if systemDelta > 0 && cpuDelta > 0 {
		return (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	
	return 0.0
}

// Close 关闭Docker管理客户端
func (dm *DockerManager) Close() error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// 停止所有容器
	for containerID, info := range dm.containers {
		if info.Status == "running" {
			if err := dm.client.ContainerStop(context.Background(), containerID, container.StopOptions{}); err != nil {
				dm.logger.Printf("Failed to stop container %s: %v", containerID[:12], err)
			}
		}
	}

	// 关闭Docker客户端
	return dm.client.Close()
}
