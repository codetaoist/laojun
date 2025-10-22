package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// GoPluginLoader Go插件加载器接口
type GoPluginLoader struct {
	plugins map[string]*LoadedGoPlugin
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

// LoadedGoPlugin 已加载的Go插件
type LoadedGoPlugin struct {
	ID       string
	Plugin   *plugin.Plugin
	Instance BasePlugin
	Config   map[string]interface{}
	LoadTime time.Time
}

// NewGoPluginLoader 创建Go插件加载器
func NewGoPluginLoader(logger *logrus.Logger) *GoPluginLoader {
	return &GoPluginLoader{
		plugins: make(map[string]*LoadedGoPlugin),
		logger:  logger,
	}
}

// LoadPlugin 加载Go插件
func (l *GoPluginLoader) LoadPlugin(pluginPath string, config map[string]interface{}) (BasePlugin, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// 加载插件
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// 查找New函数
	newFunc, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("plugin must export 'New' function: %w", err)
	}

	// 验证New函数签名
	newFuncType := reflect.TypeOf(newFunc)
	if newFuncType.Kind() != reflect.Func {
		return nil, fmt.Errorf("'New' must be a function")
	}

	// 调用New函数创建插件实例
	results := reflect.ValueOf(newFunc).Call(nil)
	if len(results) != 1 {
		return nil, fmt.Errorf("'New' function must return exactly one value")
	}

	// 检查返回值是否实现BasePlugin接口
	pluginInstance := results[0].Interface()
	basePlugin, ok := pluginInstance.(BasePlugin)
	if !ok {
		return nil, fmt.Errorf("plugin must implement BasePlugin interface")
	}

	// 初始化插件
	if err := basePlugin.Initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 生成插件ID
	pluginID := uuid.New().String()

	// 存储已加载的插件
	loadedPlugin := &LoadedGoPlugin{
		ID:       pluginID,
		Plugin:   p,
		Instance: basePlugin,
		Config:   config,
		LoadTime: time.Now(),
	}

	l.plugins[pluginID] = loadedPlugin

	l.logger.Infof("Successfully loaded Go plugin: %s (ID: %s)", pluginPath, pluginID)
	return basePlugin, nil
}

// UnloadPlugin 卸载插件
func (l *GoPluginLoader) UnloadPlugin(pluginID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	loadedPlugin, exists := l.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 清理插件资源
	if err := loadedPlugin.Instance.Cleanup(); err != nil {
		l.logger.Warnf("Failed to cleanup plugin %s: %v", pluginID, err)
	}

	// 从映射中删除
	delete(l.plugins, pluginID)

	l.logger.Infof("Successfully unloaded plugin: %s", pluginID)
	return nil
}

// ReloadPlugin 重新加载插件
func (l *GoPluginLoader) ReloadPlugin(pluginID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	loadedPlugin, exists := l.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 先卸载插件
	if err := loadedPlugin.Instance.Cleanup(); err != nil {
		l.logger.Warnf("Failed to cleanup plugin %s during reload: %v", pluginID, err)
	}

	// 重新加载（这里简化处理，实际需要保存原始路径）
	// 在实际实现中，应该保存插件的原始路径信息
	l.logger.Infof("Plugin %s marked for reload", pluginID)
	return nil
}

// ValidatePlugin 验证插件
func (l *GoPluginLoader) ValidatePlugin(pluginPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// 检查文件扩展名
	if filepath.Ext(pluginPath) != ".so" {
		return fmt.Errorf("plugin file must have .so extension")
	}

	// 尝试打开插件（不实际加载）
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("invalid plugin file: %w", err)
	}

	// 检查必需的导出函数New
	if _, err := p.Lookup("New"); err != nil {
		return fmt.Errorf("plugin must export 'New' function: %w", err)
	}

	return nil
}

// GetLoadedPlugins 获取已加载的插件列表
func (l *GoPluginLoader) GetLoadedPlugins() []string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	var pluginIDs []string
	for id := range l.plugins {
		pluginIDs = append(pluginIDs, id)
	}
	return pluginIDs
}

// GetPlugin 获取插件实例
func (l *GoPluginLoader) GetPlugin(pluginID string) (BasePlugin, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	loadedPlugin, exists := l.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return loadedPlugin.Instance, nil
}

// JSPluginLoader JavaScript插件加载器
type JSPluginLoader struct {
	plugins map[string]*LoadedJSPlugin
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

// LoadedJSPlugin 已加载的JavaScript插件
type LoadedJSPlugin struct {
	ID       string
	Path     string
	Config   map[string]interface{}
	Metadata map[string]interface{}
	LoadTime time.Time
}

// NewJSPluginLoader 创建JavaScript插件加载器
func NewJSPluginLoader(logger *logrus.Logger) *JSPluginLoader {
	return &JSPluginLoader{
		plugins: make(map[string]*LoadedJSPlugin),
		logger:  logger,
	}
}

// LoadPlugin 加载JavaScript插件
func (l *JSPluginLoader) LoadPlugin(pluginPath string, config map[string]interface{}) (BasePlugin, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// 读取插件元数据
	metadataPath := filepath.Join(filepath.Dir(pluginPath), "package.json")
	metadata, err := l.readPluginMetadata(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin metadata: %w", err)
	}

	// 生成插件ID
	pluginID := uuid.New().String()

	// 存储已加载的插件
	loadedPlugin := &LoadedJSPlugin{
		ID:       pluginID,
		Path:     pluginPath,
		Config:   config,
		Metadata: metadata,
		LoadTime: time.Now(),
	}

	l.plugins[pluginID] = loadedPlugin

	l.logger.Infof("Successfully loaded JS plugin: %s (ID: %s)", pluginPath, pluginID)

	// 返回一个包装的BasePlugin实现
	return &JSPluginWrapper{
		id:       pluginID,
		path:     pluginPath,
		config:   config,
		metadata: metadata,
		loader:   l,
	}, nil
}

// UnloadPlugin 卸载JavaScript插件
func (l *JSPluginLoader) UnloadPlugin(pluginID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, exists := l.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 从映射中删除
	delete(l.plugins, pluginID)

	l.logger.Infof("Successfully unloaded JS plugin: %s", pluginID)
	return nil
}

// ReloadPlugin 重新加载JavaScript插件
func (l *JSPluginLoader) ReloadPlugin(pluginID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	loadedPlugin, exists := l.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 重新读取元数据
	metadataPath := filepath.Join(filepath.Dir(loadedPlugin.Path), "package.json")
	metadata, err := l.readPluginMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin metadata during reload: %w", err)
	}

	loadedPlugin.Metadata = metadata
	loadedPlugin.LoadTime = time.Now()

	l.logger.Infof("Successfully reloaded JS plugin: %s", pluginID)
	return nil
}

// ValidatePlugin 验证JavaScript插件
func (l *JSPluginLoader) ValidatePlugin(pluginPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// 检查文件扩展名
	ext := filepath.Ext(pluginPath)
	if ext != ".js" && ext != ".mjs" {
		return fmt.Errorf("plugin file must have .js or .mjs extension")
	}

	// 检查package.json
	metadataPath := filepath.Join(filepath.Dir(pluginPath), "package.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin must have package.json file")
	}

	// 验证元数据
	_, err := l.readPluginMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("invalid plugin metadata: %w", err)
	}

	return nil
}

// readPluginMetadata 读取插件元数据
func (l *JSPluginLoader) readPluginMetadata(metadataPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	// 验证必需字段
	requiredFields := []string{"name", "version", "main"}
	for _, field := range requiredFields {
		if _, exists := metadata[field]; !exists {
			return nil, fmt.Errorf("missing required field: %s", field)
		}
	}

	return metadata, nil
}

// JSPluginWrapper JavaScript插件包装器
type JSPluginWrapper struct {
	id       string
	path     string
	config   map[string]interface{}
	metadata map[string]interface{}
	loader   *JSPluginLoader
}

// GetInfo 获取插件信息
func (w *JSPluginWrapper) GetInfo() PluginInfo {
	return PluginInfo{
		ID:          w.id,
		Name:        w.getMetadataString("name"),
		Version:     w.getMetadataString("version"),
		Description: w.getMetadataString("description"),
		Author:      w.getMetadataString("author"),
		Metadata:    w.convertMetadataToStringMap(),
	}
}

// Initialize 初始化插件
func (w *JSPluginWrapper) Initialize(config map[string]interface{}) error {
	w.config = config
	// JavaScript插件的初始化将在前端处理
	return nil
}

// Cleanup 清理资源
func (w *JSPluginWrapper) Cleanup() error {
	// JavaScript插件的清理将在前端处理
	return nil
}

// HealthCheck 健康检查
func (w *JSPluginWrapper) HealthCheck() error {
	// 检查插件文件是否仍然存在
	if _, err := os.Stat(w.path); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", w.path)
	}
	return nil
}

// getMetadataString 获取元数据字符串值
func (w *JSPluginWrapper) getMetadataString(key string) string {
	if value, exists := w.metadata[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// convertMetadataToStringMap 转换元数据为字符串映射
func (w *JSPluginWrapper) convertMetadataToStringMap() map[string]string {
	result := make(map[string]string)
	for key, value := range w.metadata {
		if str, ok := value.(string); ok {
			result[key] = str
		} else {
			result[key] = fmt.Sprintf("%v", value)
		}
	}
	return result
}

// PluginLoaderManager 插件加载器管理器
type PluginLoaderManager struct {
	goLoader *GoPluginLoader
	jsLoader *JSPluginLoader
	logger   *logrus.Logger
}

// NewPluginLoaderManager 创建插件加载器管理器
func NewPluginLoaderManager(logger *logrus.Logger) *PluginLoaderManager {
	return &PluginLoaderManager{
		goLoader: NewGoPluginLoader(logger),
		jsLoader: NewJSPluginLoader(logger),
		logger:   logger,
	}
}

// LoadPlugin 根据文件类型加载插件
func (m *PluginLoaderManager) LoadPlugin(pluginPath string, config map[string]interface{}) (BasePlugin, error) {
	ext := filepath.Ext(pluginPath)

	switch ext {
	case ".so":
		return m.goLoader.LoadPlugin(pluginPath, config)
	case ".js", ".mjs":
		return m.jsLoader.LoadPlugin(pluginPath, config)
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", ext)
	}
}

// UnloadPlugin 卸载插件
func (m *PluginLoaderManager) UnloadPlugin(pluginID string) error {
	// 尝试从Go加载器卸载插件
	if err := m.goLoader.UnloadPlugin(pluginID); err == nil {
		return nil
	}

	// 尝试从JS加载器卸载插件
	return m.jsLoader.UnloadPlugin(pluginID)
}

// ValidatePlugin 验证插件
func (m *PluginLoaderManager) ValidatePlugin(pluginPath string) error {
	ext := filepath.Ext(pluginPath)

	switch ext {
	case ".so":
		return m.goLoader.ValidatePlugin(pluginPath)
	case ".js", ".mjs":
		return m.jsLoader.ValidatePlugin(pluginPath)
	default:
		return fmt.Errorf("unsupported plugin type: %s", ext)
	}
}

// ReloadPlugin 重新加载插件
func (m *PluginLoaderManager) ReloadPlugin(pluginID string) error {
	// 尝试从Go插件加载器重新加载插件
	if err := m.goLoader.ReloadPlugin(pluginID); err == nil {
		return nil
	}

	// 尝试从JS插件加载器重新加载插件
	return m.jsLoader.ReloadPlugin(pluginID)
}

// GetPlugin 获取插件实例
func (m *PluginLoaderManager) GetPlugin(pluginID string) (BasePlugin, error) {
	// 尝试从Go插件加载器获取插件实例
	if plugin, err := m.goLoader.GetPlugin(pluginID); err == nil {
		return plugin, nil
	}

	// 尝试从JS插件加载器获取插件实例
	// 注意：JS插件加载器需要实现GetPlugin方法
	return nil, fmt.Errorf("plugin not found: %s", pluginID)
}
