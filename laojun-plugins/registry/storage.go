package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PluginStorage 插件存储接口
type PluginStorage interface {
	// Save 保存插件注册信息
	Save(ctx context.Context, plugin *PluginRegistration) error

	// Get 获取插件注册信息
	Get(ctx context.Context, pluginID string) (*PluginRegistration, error)

	// Delete 删除插件注册信息
	Delete(ctx context.Context, pluginID string) error

	// List 列出所有插件
	List(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error)

	// Update 更新插件信息
	Update(ctx context.Context, pluginID string, updates map[string]interface{}) error

	// UpdateStatus 更新插件状态
	UpdateStatus(ctx context.Context, pluginID string, status PluginStatus) error

	// UpdateMetrics 更新插件指标
	UpdateMetrics(ctx context.Context, pluginID string, metrics *PluginMetrics) error

	// GetByEndpoint 根据端点获取插件
	GetByEndpoint(ctx context.Context, endpoint string) (*PluginRegistration, error)

	// Search 搜索插件
	Search(ctx context.Context, query string, filters *PluginFilter) ([]*PluginRegistration, error)

	// Close 关闭存储
	Close() error
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	plugins map[string]*PluginRegistration
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage(logger *logrus.Logger) *MemoryStorage {
	return &MemoryStorage{
		plugins: make(map[string]*PluginRegistration),
		logger:  logger,
	}
}

// Save 保存插件注册信息
func (s *MemoryStorage) Save(ctx context.Context, plugin *PluginRegistration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 深拷贝插件信息
	pluginCopy := *plugin
	s.plugins[plugin.ID] = &pluginCopy

	s.logger.WithField("plugin_id", plugin.ID).Debug("Plugin saved to memory storage")
	return nil
}

// Get 获取插件注册信息
func (s *MemoryStorage) Get(ctx context.Context, pluginID string) (*PluginRegistration, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	plugin, exists := s.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 返回拷贝以避免并发修改
	pluginCopy := *plugin
	return &pluginCopy, nil
}

// Delete 删除插件注册信息
func (s *MemoryStorage) Delete(ctx context.Context, pluginID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.plugins[pluginID]; !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	delete(s.plugins, pluginID)
	s.logger.WithField("plugin_id", pluginID).Debug("Plugin deleted from memory storage")
	return nil
}

// List 列出所有插件
func (s *MemoryStorage) List(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result []*PluginRegistration
	for _, plugin := range s.plugins {
		if s.matchesFilter(plugin, filter) {
			pluginCopy := *plugin
			result = append(result, &pluginCopy)
		}
	}

	return result, nil
}

// Update 更新插件信息
func (s *MemoryStorage) Update(ctx context.Context, pluginID string, updates map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	plugin, exists := s.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		plugin.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		plugin.Description = description
	}
	if version, ok := updates["version"].(string); ok {
		plugin.Version = version
	}
	if category, ok := updates["category"].(string); ok {
		plugin.Category = category
	}
	if tags, ok := updates["tags"].([]string); ok {
		plugin.Tags = tags
	}

	plugin.UpdatedAt = time.Now()

	s.logger.WithField("plugin_id", pluginID).Debug("Plugin updated in memory storage")
	return nil
}

// UpdateStatus 更新插件状态
func (s *MemoryStorage) UpdateStatus(ctx context.Context, pluginID string, status PluginStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	plugin, exists := s.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	plugin.Status = status
	plugin.UpdatedAt = time.Now()

	s.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"status":    status,
	}).Debug("Plugin status updated in memory storage")

	return nil
}

// UpdateMetrics 更新插件指标
func (s *MemoryStorage) UpdateMetrics(ctx context.Context, pluginID string, metrics *PluginMetrics) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	plugin, exists := s.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	plugin.Metrics = metrics
	plugin.UpdatedAt = time.Now()

	return nil
}

// GetByEndpoint 根据端点获取插件
func (s *MemoryStorage) GetByEndpoint(ctx context.Context, endpoint string) (*PluginRegistration, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, plugin := range s.plugins {
		for _, ep := range plugin.Endpoints {
			if ep.URL == endpoint {
				pluginCopy := *plugin
				return &pluginCopy, nil
			}
		}
	}

	return nil, fmt.Errorf("plugin not found for endpoint: %s", endpoint)
}

// Search 搜索插件
func (s *MemoryStorage) Search(ctx context.Context, query string, filters *PluginFilter) ([]*PluginRegistration, error) {
	// 简化实现，直接调用List并在应用层过滤
	return s.List(ctx, filters)
}

// Close 关闭存储
func (s *MemoryStorage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.plugins = make(map[string]*PluginRegistration)
	return nil
}

// matchesFilter 检查插件是否匹配过滤条件
func (s *MemoryStorage) matchesFilter(plugin *PluginRegistration, filter *PluginFilter) bool {
	if filter == nil {
		return true
	}

	// 状态过滤
	if filter.Status != "" && plugin.Status != filter.Status {
		return false
	}

	// 分类过滤
	if filter.Category != "" && plugin.Category != filter.Category {
		return false
	}

	// 作者过滤
	if filter.Author != "" && plugin.Author != filter.Author {
		return false
	}

	// 标签过滤
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
			for _, pluginTag := range plugin.Tags {
				if pluginTag == filterTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// FileStorage 文件存储实现
type FileStorage struct {
	dataDir string
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

// NewFileStorage 创建文件存储
func NewFileStorage(dataDir string, logger *logrus.Logger) (*FileStorage, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
		logger:  logger,
	}, nil
}

// Save 保存插件注册信息
func (s *FileStorage) Save(ctx context.Context, plugin *PluginRegistration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filePath := s.getPluginFilePath(plugin.ID)
	
	data, err := json.MarshalIndent(plugin, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugin: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write plugin file: %w", err)
	}

	s.logger.WithField("plugin_id", plugin.ID).Debug("Plugin saved to file storage")
	return nil
}

// Get 获取插件注册信息
func (s *FileStorage) Get(ctx context.Context, pluginID string) (*PluginRegistration, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	filePath := s.getPluginFilePath(pluginID)
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plugin not found: %s", pluginID)
		}
		return nil, fmt.Errorf("failed to read plugin file: %w", err)
	}

	var plugin PluginRegistration
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin: %w", err)
	}

	return &plugin, nil
}

// Delete 删除插件注册信息
func (s *FileStorage) Delete(ctx context.Context, pluginID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filePath := s.getPluginFilePath(pluginID)
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plugin not found: %s", pluginID)
		}
		return fmt.Errorf("failed to delete plugin file: %w", err)
	}

	s.logger.WithField("plugin_id", pluginID).Debug("Plugin deleted from file storage")
	return nil
}

// List 列出所有插件
func (s *FileStorage) List(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	files, err := ioutil.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var result []*PluginRegistration
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			pluginID := file.Name()[:len(file.Name())-5] // 移除.json后缀
			
			plugin, err := s.Get(ctx, pluginID)
			if err != nil {
				s.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to load plugin")
				continue
			}

			if s.matchesFilter(plugin, filter) {
				result = append(result, plugin)
			}
		}
	}

	return result, nil
}

// Update 更新插件信息
func (s *FileStorage) Update(ctx context.Context, pluginID string, updates map[string]interface{}) error {
	plugin, err := s.Get(ctx, pluginID)
	if err != nil {
		return err
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		plugin.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		plugin.Description = description
	}
	if version, ok := updates["version"].(string); ok {
		plugin.Version = version
	}
	if category, ok := updates["category"].(string); ok {
		plugin.Category = category
	}
	if tags, ok := updates["tags"].([]string); ok {
		plugin.Tags = tags
	}

	plugin.UpdatedAt = time.Now()

	return s.Save(ctx, plugin)
}

// UpdateStatus 更新插件状态
func (s *FileStorage) UpdateStatus(ctx context.Context, pluginID string, status PluginStatus) error {
	plugin, err := s.Get(ctx, pluginID)
	if err != nil {
		return err
	}

	plugin.Status = status
	plugin.UpdatedAt = time.Now()

	return s.Save(ctx, plugin)
}

// UpdateMetrics 更新插件指标
func (s *FileStorage) UpdateMetrics(ctx context.Context, pluginID string, metrics *PluginMetrics) error {
	plugin, err := s.Get(ctx, pluginID)
	if err != nil {
		return err
	}

	plugin.Metrics = metrics
	plugin.UpdatedAt = time.Now()

	return s.Save(ctx, plugin)
}

// GetByEndpoint 根据端点获取插件
func (s *FileStorage) GetByEndpoint(ctx context.Context, endpoint string) (*PluginRegistration, error) {
	plugins, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, plugin := range plugins {
		for _, ep := range plugin.Endpoints {
			if ep.URL == endpoint {
				return plugin, nil
			}
		}
	}

	return nil, fmt.Errorf("plugin not found for endpoint: %s", endpoint)
}

// Search 搜索插件
func (s *FileStorage) Search(ctx context.Context, query string, filters *PluginFilter) ([]*PluginRegistration, error) {
	// 简化实现，直接调用List并在应用层过滤
	return s.List(ctx, filters)
}

// Close 关闭存储
func (s *FileStorage) Close() error {
	// 文件存储无需特殊关闭操作
	return nil
}

// getPluginFilePath 获取插件文件路径
func (s *FileStorage) getPluginFilePath(pluginID string) string {
	return filepath.Join(s.dataDir, pluginID+".json")
}

// matchesFilter 检查插件是否匹配过滤条件
func (s *FileStorage) matchesFilter(plugin *PluginRegistration, filter *PluginFilter) bool {
	if filter == nil {
		return true
	}

	// 状态过滤
	if filter.Status != "" && plugin.Status != filter.Status {
		return false
	}

	// 分类过滤
	if filter.Category != "" && plugin.Category != filter.Category {
		return false
	}

	// 作者过滤
	if filter.Author != "" && plugin.Author != filter.Author {
		return false
	}

	// 标签过滤
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
			for _, pluginTag := range plugin.Tags {
				if pluginTag == filterTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// CachedStorage 带缓存的存储实现
type CachedStorage struct {
	backend PluginStorage
	cache   *MemoryStorage
	logger  *logrus.Logger
}

// NewCachedStorage 创建带缓存的存储
func NewCachedStorage(backend PluginStorage, logger *logrus.Logger) *CachedStorage {
	return &CachedStorage{
		backend: backend,
		cache:   NewMemoryStorage(logger),
		logger:  logger,
	}
}

// Save 保存插件注册信息
func (s *CachedStorage) Save(ctx context.Context, plugin *PluginRegistration) error {
	// 先保存到后端存储
	if err := s.backend.Save(ctx, plugin); err != nil {
		return err
	}

	// 然后更新缓存
	return s.cache.Save(ctx, plugin)
}

// Get 获取插件注册信息
func (s *CachedStorage) Get(ctx context.Context, pluginID string) (*PluginRegistration, error) {
	// 先尝试从缓存获取
	plugin, err := s.cache.Get(ctx, pluginID)
	if err == nil {
		return plugin, nil
	}

	// 缓存未命中，从后端存储获取
	plugin, err = s.backend.Get(ctx, pluginID)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	s.cache.Save(ctx, plugin)
	return plugin, nil
}

// Delete 删除插件注册信息
func (s *CachedStorage) Delete(ctx context.Context, pluginID string) error {
	// 先从后端存储删除
	if err := s.backend.Delete(ctx, pluginID); err != nil {
		return err
	}

	// 然后从缓存删除
	s.cache.Delete(ctx, pluginID)
	return nil
}

// List 列出所有插件
func (s *CachedStorage) List(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error) {
	// 直接从后端存储获取最新数据
	return s.backend.List(ctx, filter)
}

// Update 更新插件信息
func (s *CachedStorage) Update(ctx context.Context, pluginID string, updates map[string]interface{}) error {
	// 先更新后端存储
	if err := s.backend.Update(ctx, pluginID, updates); err != nil {
		return err
	}

	// 然后更新缓存
	return s.cache.Update(ctx, pluginID, updates)
}

// UpdateStatus 更新插件状态
func (s *CachedStorage) UpdateStatus(ctx context.Context, pluginID string, status PluginStatus) error {
	// 先更新后端存储
	if err := s.backend.UpdateStatus(ctx, pluginID, status); err != nil {
		return err
	}

	// 然后更新缓存
	return s.cache.UpdateStatus(ctx, pluginID, status)
}

// UpdateMetrics 更新插件指标
func (s *CachedStorage) UpdateMetrics(ctx context.Context, pluginID string, metrics *PluginMetrics) error {
	// 先更新后端存储
	if err := s.backend.UpdateMetrics(ctx, pluginID, metrics); err != nil {
		return err
	}

	// 然后更新缓存
	return s.cache.UpdateMetrics(ctx, pluginID, metrics)
}

// GetByEndpoint 根据端点获取插件
func (s *CachedStorage) GetByEndpoint(ctx context.Context, endpoint string) (*PluginRegistration, error) {
	// 先尝试从缓存获取
	plugin, err := s.cache.GetByEndpoint(ctx, endpoint)
	if err == nil {
		return plugin, nil
	}

	// 缓存未命中，从后端存储获取
	return s.backend.GetByEndpoint(ctx, endpoint)
}

// Search 搜索插件
func (s *CachedStorage) Search(ctx context.Context, query string, filters *PluginFilter) ([]*PluginRegistration, error) {
	// 直接从后端存储搜索
	return s.backend.Search(ctx, query, filters)
}

// Close 关闭存储
func (s *CachedStorage) Close() error {
	if err := s.cache.Close(); err != nil {
		s.logger.WithError(err).Warn("Failed to close cache storage")
	}
	
	return s.backend.Close()
}