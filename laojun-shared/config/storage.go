package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MemoryConfigStorage 内存配置存储
type MemoryConfigStorage struct {
	data    map[string]*ConfigItem
	history map[string][]*ConfigHistory
	mu      sync.RWMutex
	logger  *logrus.Logger
	options *ConfigOptions
}

// NewMemoryConfigStorage 创建内存配置存储
func NewMemoryConfigStorage(options *ConfigOptions) *MemoryConfigStorage {
	return &MemoryConfigStorage{
		data:    make(map[string]*ConfigItem),
		history: make(map[string][]*ConfigHistory),
		logger:  logrus.New(),
		options: options,
	}
}

// Get 获取配置
func (s *MemoryConfigStorage) Get(ctx context.Context, key string) (*ConfigItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	item, exists := s.data[key]
	if !exists {
		return nil, ErrConfigNotFound
	}
	
	// 检查TTL
	if item.TTL > 0 && time.Since(item.UpdatedAt) > item.TTL {
		// 配置已过期，异步删除
		go func() {
			s.mu.Lock()
			delete(s.data, key)
			s.mu.Unlock()
		}()
		return nil, ErrConfigNotFound
	}
	
	// 返回副本以避免并发修改
	return &ConfigItem{
		Key:       item.Key,
		Value:     item.Value,
		Type:      item.Type,
		Version:   item.Version,
		TTL:       item.TTL,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		Metadata:  copyMetadata(item.Metadata),
	}, nil
}

// Set 设置配置
func (s *MemoryConfigStorage) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	var version int64 = 1
	var createdAt time.Time = now
	
	// 检查是否已存在
	if existing, exists := s.data[key]; exists {
		version = existing.Version + 1
		createdAt = existing.CreatedAt
		
		// 保存历史记录
		s.saveHistory(key, existing)
	}
	
	// 创建新的配置项
	item := &ConfigItem{
		Key:       key,
		Value:     value,
		Type:      ConfigTypeString,
		Version:   version,
		TTL:       ttl,
		CreatedAt: createdAt,
		UpdatedAt: now,
		Metadata:  make(map[string]string),
	}
	
	s.data[key] = item
	
	s.logger.WithFields(logrus.Fields{
		"key":     key,
		"version": version,
		"ttl":     ttl,
	}).Debug("Config set in memory storage")
	
	return nil
}

// Delete 删除配置
func (s *MemoryConfigStorage) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	item, exists := s.data[key]
	if !exists {
		return ErrConfigNotFound
	}
	
	// 保存历史记录
	s.saveHistory(key, item)
	
	delete(s.data, key)
	
	s.logger.WithField("key", key).Debug("Config deleted from memory storage")
	return nil
}

// Exists 检查配置是否存在
func (s *MemoryConfigStorage) Exists(ctx context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	item, exists := s.data[key]
	if !exists {
		return false, nil
	}
	
	// 检查TTL
	if item.TTL > 0 && time.Since(item.UpdatedAt) > item.TTL {
		return false, nil
	}
	
	return true, nil
}

// List 列出配置键
func (s *MemoryConfigStorage) List(ctx context.Context, prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var keys []string
	now := time.Now()
	
	for key, item := range s.data {
		// 检查TTL
		if item.TTL > 0 && now.Sub(item.UpdatedAt) > item.TTL {
			continue
		}
		
		// 检查前缀
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	
	sort.Strings(keys)
	return keys, nil
}

// GetMultiple 批量获取配置
func (s *MemoryConfigStorage) GetMultiple(ctx context.Context, keys []string) (map[string]*ConfigItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make(map[string]*ConfigItem)
	now := time.Now()
	
	for _, key := range keys {
		if item, exists := s.data[key]; exists {
			// 检查TTL
			if item.TTL > 0 && now.Sub(item.UpdatedAt) > item.TTL {
				continue
			}
			
			// 返回副本
			result[key] = &ConfigItem{
				Key:       item.Key,
				Value:     item.Value,
				Type:      item.Type,
				Version:   item.Version,
				TTL:       item.TTL,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
				Metadata:  copyMetadata(item.Metadata),
			}
		}
	}
	
	return result, nil
}

// SetMultiple 批量设置配置
func (s *MemoryConfigStorage) SetMultiple(ctx context.Context, configs map[string]string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	
	for key, value := range configs {
		var version int64 = 1
		var createdAt time.Time = now
		
		// 检查是否已存在
		if existing, exists := s.data[key]; exists {
			version = existing.Version + 1
			createdAt = existing.CreatedAt
			
			// 保存历史记录
			s.saveHistory(key, existing)
		}
		
		// 创建新的配置项
		item := &ConfigItem{
			Key:       key,
			Value:     value,
			Type:      ConfigTypeString,
			Version:   version,
			TTL:       ttl,
			CreatedAt: createdAt,
			UpdatedAt: now,
			Metadata:  make(map[string]string),
		}
		
		s.data[key] = item
	}
	
	s.logger.WithField("count", len(configs)).Debug("Multiple configs set in memory storage")
	return nil
}

// GetHistory 获取配置历史
func (s *MemoryConfigStorage) GetHistory(ctx context.Context, key string, limit int) ([]*ConfigHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	history, exists := s.history[key]
	if !exists {
		return []*ConfigHistory{}, nil
	}
	
	// 复制历史记录
	result := make([]*ConfigHistory, len(history))
	copy(result, history)
	
	// 按时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	
	// 限制数量
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	
	return result, nil
}

// GetVersion 获取指定版本的配置
func (s *MemoryConfigStorage) GetVersion(ctx context.Context, key string, version int64) (*ConfigItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 检查当前版本
	if item, exists := s.data[key]; exists && item.Version == version {
		return &ConfigItem{
			Key:       item.Key,
			Value:     item.Value,
			Type:      item.Type,
			Version:   item.Version,
			TTL:       item.TTL,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Metadata:  copyMetadata(item.Metadata),
		}, nil
	}
	
	// 在历史记录中查找
	if history, exists := s.history[key]; exists {
		for _, h := range history {
			if h.Version == version {
				return &ConfigItem{
					Key:       key,
					Value:     h.Value,
					Type:      h.Type,
					Version:   h.Version,
					TTL:       0, // 历史版本不设置TTL
					CreatedAt: h.Timestamp,
					UpdatedAt: h.Timestamp,
					Metadata:  copyMetadata(h.Metadata),
				}, nil
			}
		}
	}
	
	return nil, ErrConfigNotFound
}

// Close 关闭存储
func (s *MemoryConfigStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.data = make(map[string]*ConfigItem)
	s.history = make(map[string][]*ConfigHistory)
	
	s.logger.Info("Memory config storage closed")
	return nil
}

// saveHistory 保存历史记录
func (s *MemoryConfigStorage) saveHistory(key string, item *ConfigItem) {
	history := &ConfigHistory{
		Key:       key,
		Value:     item.Value,
		Type:      item.Type,
		Version:   item.Version,
		Timestamp: item.UpdatedAt,
		Metadata:  copyMetadata(item.Metadata),
	}
	
	if s.history[key] == nil {
		s.history[key] = make([]*ConfigHistory, 0)
	}
	
	s.history[key] = append(s.history[key], history)
	
	// 限制历史记录数量
	maxHistory := 100
	if s.options != nil && s.options.MaxHistorySize > 0 {
		maxHistory = s.options.MaxHistorySize
	}
	
	if len(s.history[key]) > maxHistory {
		s.history[key] = s.history[key][len(s.history[key])-maxHistory:]
	}
}

// FileConfigStorage 文件配置存储
type FileConfigStorage struct {
	filePath string
	data     map[string]*ConfigItem
	history  map[string][]*ConfigHistory
	mu       sync.RWMutex
	logger   *logrus.Logger
	options  *ConfigOptions
}

// NewFileConfigStorage 创建文件配置存储
func NewFileConfigStorage(filePath string, options *ConfigOptions) *FileConfigStorage {
	storage := &FileConfigStorage{
		filePath: filePath,
		data:     make(map[string]*ConfigItem),
		history:  make(map[string][]*ConfigHistory),
		logger:   logrus.New(),
		options:  options,
	}
	
	// 加载现有数据
	if err := storage.load(); err != nil {
		storage.logger.WithError(err).Warn("Failed to load config from file")
	}
	
	return storage
}

// Get 获取配置
func (s *FileConfigStorage) Get(ctx context.Context, key string) (*ConfigItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	item, exists := s.data[key]
	if !exists {
		return nil, ErrConfigNotFound
	}
	
	// 检查TTL
	if item.TTL > 0 && time.Since(item.UpdatedAt) > item.TTL {
		return nil, ErrConfigNotFound
	}
	
	// 返回副本
	return &ConfigItem{
		Key:       item.Key,
		Value:     item.Value,
		Type:      item.Type,
		Version:   item.Version,
		TTL:       item.TTL,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		Metadata:  copyMetadata(item.Metadata),
	}, nil
}

// Set 设置配置
func (s *FileConfigStorage) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	var version int64 = 1
	var createdAt time.Time = now
	
	// 检查是否已存在
	if existing, exists := s.data[key]; exists {
		version = existing.Version + 1
		createdAt = existing.CreatedAt
		
		// 保存历史记录
		s.saveHistory(key, existing)
	}
	
	// 创建新的配置项
	item := &ConfigItem{
		Key:       key,
		Value:     value,
		Type:      ConfigTypeString,
		Version:   version,
		TTL:       ttl,
		CreatedAt: createdAt,
		UpdatedAt: now,
		Metadata:  make(map[string]string),
	}
	
	s.data[key] = item
	
	// 保存到文件
	if err := s.save(); err != nil {
		return fmt.Errorf("failed to save config to file: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"key":     key,
		"version": version,
		"ttl":     ttl,
	}).Debug("Config set in file storage")
	
	return nil
}

// Delete 删除配置
func (s *FileConfigStorage) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	item, exists := s.data[key]
	if !exists {
		return ErrConfigNotFound
	}
	
	// 保存历史记录
	s.saveHistory(key, item)
	
	delete(s.data, key)
	
	// 保存到文件
	if err := s.save(); err != nil {
		return fmt.Errorf("failed to save config to file: %w", err)
	}
	
	s.logger.WithField("key", key).Debug("Config deleted from file storage")
	return nil
}

// Exists 检查配置是否存在
func (s *FileConfigStorage) Exists(ctx context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	item, exists := s.data[key]
	if !exists {
		return false, nil
	}
	
	// 检查TTL
	if item.TTL > 0 && time.Since(item.UpdatedAt) > item.TTL {
		return false, nil
	}
	
	return true, nil
}

// List 列出配置键
func (s *FileConfigStorage) List(ctx context.Context, prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var keys []string
	now := time.Now()
	
	for key, item := range s.data {
		// 检查TTL
		if item.TTL > 0 && now.Sub(item.UpdatedAt) > item.TTL {
			continue
		}
		
		// 检查前缀
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	
	sort.Strings(keys)
	return keys, nil
}

// GetMultiple 批量获取配置
func (s *FileConfigStorage) GetMultiple(ctx context.Context, keys []string) (map[string]*ConfigItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make(map[string]*ConfigItem)
	now := time.Now()
	
	for _, key := range keys {
		if item, exists := s.data[key]; exists {
			// 检查TTL
			if item.TTL > 0 && now.Sub(item.UpdatedAt) > item.TTL {
				continue
			}
			
			// 返回副本
			result[key] = &ConfigItem{
				Key:       item.Key,
				Value:     item.Value,
				Type:      item.Type,
				Version:   item.Version,
				TTL:       item.TTL,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
				Metadata:  copyMetadata(item.Metadata),
			}
		}
	}
	
	return result, nil
}

// SetMultiple 批量设置配置
func (s *FileConfigStorage) SetMultiple(ctx context.Context, configs map[string]string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	
	for key, value := range configs {
		var version int64 = 1
		var createdAt time.Time = now
		
		// 检查是否已存在
		if existing, exists := s.data[key]; exists {
			version = existing.Version + 1
			createdAt = existing.CreatedAt
			
			// 保存历史记录
			s.saveHistory(key, existing)
		}
		
		// 创建新的配置项
		item := &ConfigItem{
			Key:       key,
			Value:     value,
			Type:      ConfigTypeString,
			Version:   version,
			TTL:       ttl,
			CreatedAt: createdAt,
			UpdatedAt: now,
			Metadata:  make(map[string]string),
		}
		
		s.data[key] = item
	}
	
	// 保存到文件
	if err := s.save(); err != nil {
		return fmt.Errorf("failed to save config to file: %w", err)
	}
	
	s.logger.WithField("count", len(configs)).Debug("Multiple configs set in file storage")
	return nil
}

// GetHistory 获取配置历史
func (s *FileConfigStorage) GetHistory(ctx context.Context, key string, limit int) ([]*ConfigHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	history, exists := s.history[key]
	if !exists {
		return []*ConfigHistory{}, nil
	}
	
	// 复制历史记录
	result := make([]*ConfigHistory, len(history))
	copy(result, history)
	
	// 按时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	
	// 限制数量
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	
	return result, nil
}

// GetVersion 获取指定版本的配置
func (s *FileConfigStorage) GetVersion(ctx context.Context, key string, version int64) (*ConfigItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 检查当前版本
	if item, exists := s.data[key]; exists && item.Version == version {
		return &ConfigItem{
			Key:       item.Key,
			Value:     item.Value,
			Type:      item.Type,
			Version:   item.Version,
			TTL:       item.TTL,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Metadata:  copyMetadata(item.Metadata),
		}, nil
	}
	
	// 在历史记录中查找
	if history, exists := s.history[key]; exists {
		for _, h := range history {
			if h.Version == version {
				return &ConfigItem{
					Key:       key,
					Value:     h.Value,
					Type:      h.Type,
					Version:   h.Version,
					TTL:       0, // 历史版本不设置TTL
					CreatedAt: h.Timestamp,
					UpdatedAt: h.Timestamp,
					Metadata:  copyMetadata(h.Metadata),
				}, nil
			}
		}
	}
	
	return nil, ErrConfigNotFound
}

// Close 关闭存储
func (s *FileConfigStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 最后一次保存
	if err := s.save(); err != nil {
		s.logger.WithError(err).Error("Failed to save config on close")
	}
	
	s.data = make(map[string]*ConfigItem)
	s.history = make(map[string][]*ConfigHistory)
	
	s.logger.Info("File config storage closed")
	return nil
}

// load 从文件加载配置
func (s *FileConfigStorage) load() error {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil // 文件不存在，不是错误
	}
	
	data, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	var fileData struct {
		Data    map[string]*ConfigItem            `json:"data"`
		History map[string][]*ConfigHistory `json:"history"`
	}
	
	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	
	s.data = fileData.Data
	if s.data == nil {
		s.data = make(map[string]*ConfigItem)
	}
	
	s.history = fileData.History
	if s.history == nil {
		s.history = make(map[string][]*ConfigHistory)
	}
	
	s.logger.WithField("count", len(s.data)).Debug("Config loaded from file")
	return nil
}

// save 保存配置到文件
func (s *FileConfigStorage) save() error {
	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	fileData := struct {
		Data    map[string]*ConfigItem            `json:"data"`
		History map[string][]*ConfigHistory `json:"history"`
	}{
		Data:    s.data,
		History: s.history,
	}
	
	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}
	
	if err := ioutil.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// saveHistory 保存历史记录
func (s *FileConfigStorage) saveHistory(key string, item *ConfigItem) {
	history := &ConfigHistory{
		Key:       key,
		Value:     item.Value,
		Type:      item.Type,
		Version:   item.Version,
		Timestamp: item.UpdatedAt,
		Metadata:  copyMetadata(item.Metadata),
	}
	
	if s.history[key] == nil {
		s.history[key] = make([]*ConfigHistory, 0)
	}
	
	s.history[key] = append(s.history[key], history)
	
	// 限制历史记录数量
	maxHistory := 100
	if s.options != nil && s.options.MaxHistorySize > 0 {
		maxHistory = s.options.MaxHistorySize
	}
	
	if len(s.history[key]) > maxHistory {
		s.history[key] = s.history[key][len(s.history[key])-maxHistory:]
	}
}

// copyMetadata 复制元数据
func copyMetadata(metadata map[string]string) map[string]string {
	if metadata == nil {
		return make(map[string]string)
	}
	
	result := make(map[string]string)
	for k, v := range metadata {
		result[k] = v
	}
	
	return result
}