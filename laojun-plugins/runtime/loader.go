package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/sirupsen/logrus"
)

// PluginLoader 插件加载器接口
type PluginLoader interface {
	// LoadPlugin 从指定路径加载插件
	LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error)

	// ValidatePlugin 验证插件
	ValidatePlugin(ctx context.Context, pluginPath string) error

	// GetPluginMetadata 获取插件元数据
	GetPluginMetadata(pluginPath string) (*PluginMetadata, error)
}

// DefaultPluginLoader 默认插件加载器实现
type DefaultPluginLoader struct {
	logger *logrus.Logger
}

// NewDefaultPluginLoader 创建默认插件加载器
func NewDefaultPluginLoader(logger *logrus.Logger) *DefaultPluginLoader {
	return &DefaultPluginLoader{
		logger: logger,
	}
}

// LoadPlugin 从指定路径加载插件
func (l *DefaultPluginLoader) LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error) {
	l.logger.WithField("path", pluginPath).Info("Loading plugin from path")

	// 验证插件路径
	if err := l.ValidatePlugin(ctx, pluginPath); err != nil {
		return nil, fmt.Errorf("plugin validation failed: %w", err)
	}

	// 获取插件元数据
	metadata, err := l.GetPluginMetadata(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin metadata: %w", err)
	}

	// 根据插件类型加载
	if strings.HasSuffix(pluginPath, ".so") {
		return l.loadNativePlugin(pluginPath, metadata)
	}

	// 默认作为目录插件处理
	return l.loadDirectoryPlugin(pluginPath, metadata)
}

// loadNativePlugin 加载原生插件(.so文件)
func (l *DefaultPluginLoader) loadNativePlugin(pluginPath string, metadata *PluginMetadata) (Plugin, error) {
	l.logger.WithField("path", pluginPath).Info("Loading native plugin")

	// 打开插件文件
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// 查找插件工厂函数
	factorySymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin factory function 'NewPlugin' not found: %w", err)
	}

	// 类型断言
	factory, ok := factorySymbol.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("invalid plugin factory function signature")
	}

	// 创建插件实例
	pluginInstance := factory()
	if pluginInstance == nil {
		return nil, fmt.Errorf("plugin factory returned nil")
	}

	return pluginInstance, nil
}

// loadDirectoryPlugin 加载目录插件
func (l *DefaultPluginLoader) loadDirectoryPlugin(pluginPath string, metadata *PluginMetadata) (Plugin, error) {
	l.logger.WithField("path", pluginPath).Info("Loading directory plugin")

	// 创建目录插件实例
	dirPlugin := &DirectoryPlugin{
		metadata: *metadata,
		path:     pluginPath,
		logger:   l.logger,
		state:    StateUnloaded,
	}

	return dirPlugin, nil
}

// ValidatePlugin 验证插件
func (l *DefaultPluginLoader) ValidatePlugin(ctx context.Context, pluginPath string) error {
	// 检查路径是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin path does not exist: %s", pluginPath)
	}

	// 检查是否为目录或.so文件
	info, err := os.Stat(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to get plugin path info: %w", err)
	}

	if info.IsDir() {
		// 验证目录插件
		return l.validateDirectoryPlugin(pluginPath)
	} else if strings.HasSuffix(pluginPath, ".so") {
		// 验证原生插件
		return l.validateNativePlugin(pluginPath)
	}

	return fmt.Errorf("unsupported plugin type: %s", pluginPath)
}

// validateDirectoryPlugin 验证目录插件
func (l *DefaultPluginLoader) validateDirectoryPlugin(pluginPath string) error {
	// 检查必需的文件
	manifestPath := filepath.Join(pluginPath, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest.json not found in plugin directory")
	}

	return nil
}

// validateNativePlugin 验证原生插件
func (l *DefaultPluginLoader) validateNativePlugin(pluginPath string) error {
	// 检查文件是否可读
	file, err := os.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("cannot open plugin file: %w", err)
	}
	defer file.Close()

	return nil
}

// GetPluginMetadata 获取插件元数据
func (l *DefaultPluginLoader) GetPluginMetadata(pluginPath string) (*PluginMetadata, error) {
	var manifestPath string

	info, err := os.Stat(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin path info: %w", err)
	}

	if info.IsDir() {
		manifestPath = filepath.Join(pluginPath, "manifest.json")
	} else {
		// 对于.so文件，查找同名的.json文件
		manifestPath = strings.TrimSuffix(pluginPath, ".so") + ".json"
	}

	// 读取manifest文件
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// 解析JSON
	var metadata PluginMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	// 验证必需字段
	if metadata.ID == "" {
		return nil, fmt.Errorf("plugin ID is required")
	}
	if metadata.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	if metadata.Version == "" {
		return nil, fmt.Errorf("plugin version is required")
	}

	return &metadata, nil
}

// DirectoryPlugin 目录插件实现
type DirectoryPlugin struct {
	metadata PluginMetadata
	path     string
	logger   *logrus.Logger
	state    PluginState
	config   map[string]interface{}
}

// GetMetadata 获取插件元数据
func (p *DirectoryPlugin) GetMetadata() PluginMetadata {
	return p.metadata
}

// Initialize 初始化插件
func (p *DirectoryPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	p.logger.WithField("plugin_id", p.metadata.ID).Info("Initializing directory plugin")
	
	p.config = config
	p.state = StateLoaded
	
	return nil
}

// Start 启动插件
func (p *DirectoryPlugin) Start(ctx context.Context) error {
	p.logger.WithField("plugin_id", p.metadata.ID).Info("Starting directory plugin")
	
	if p.state != StateLoaded && p.state != StateStopped {
		return fmt.Errorf("plugin is not in a startable state: %s", p.state)
	}
	
	p.state = StateRunning
	return nil
}

// Stop 停止插件
func (p *DirectoryPlugin) Stop(ctx context.Context) error {
	p.logger.WithField("plugin_id", p.metadata.ID).Info("Stopping directory plugin")
	
	if p.state != StateRunning {
		return fmt.Errorf("plugin is not running")
	}
	
	p.state = StateStopped
	return nil
}

// Cleanup 清理插件资源
func (p *DirectoryPlugin) Cleanup(ctx context.Context) error {
	p.logger.WithField("plugin_id", p.metadata.ID).Info("Cleaning up directory plugin")
	
	p.state = StateUnloaded
	p.config = nil
	
	return nil
}

// GetStatus 获取插件状态
func (p *DirectoryPlugin) GetStatus() PluginState {
	return p.state
}

// HandleEvent 处理事件
func (p *DirectoryPlugin) HandleEvent(ctx context.Context, event interface{}) error {
	p.logger.WithFields(logrus.Fields{
		"plugin_id": p.metadata.ID,
		"event":     event,
	}).Info("Handling event")

	// 目录插件的事件处理逻辑
	// 这里可以根据事件类型进行不同的处理
	return nil
}

// ProcessData 处理数据
func (p *DirectoryPlugin) ProcessData(ctx context.Context, data interface{}) (interface{}, error) {
	p.logger.WithFields(logrus.Fields{
		"plugin_id": p.metadata.ID,
		"data":      data,
	}).Info("Processing data")

	// 目录插件的数据处理逻辑
	// 这里可以根据数据类型进行不同的处理
	return data, nil
}