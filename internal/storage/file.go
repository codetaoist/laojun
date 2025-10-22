package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// FileStorage 基于文件的配置存储实现
type FileStorage struct {
	basePath string
	watcher  *fsnotify.Watcher
	watchers map[string]chan *ConfigChangeEvent
	mutex    sync.RWMutex
	logger   *logrus.Logger
}

// NewFileStorage 创建文件存储实例
func NewFileStorage(basePath string, enableWatch bool) (*FileStorage, error) {
	fs := &FileStorage{
		basePath: basePath,
		watchers: make(map[string]chan *ConfigChangeEvent),
		logger:   logrus.New(),
	}

	// 确保基础目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	// 启用文件监听
	if enableWatch {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("failed to create file watcher: %w", err)
		}
		fs.watcher = watcher

		// 监听基础目录
		if err := watcher.Add(basePath); err != nil {
			return nil, fmt.Errorf("failed to watch base path: %w", err)
		}

		go fs.handleFileEvents()
	}

	return fs, nil
}

// Get 获取配置项
func (fs *FileStorage) Get(ctx context.Context, service, environment, key string) (*ConfigItem, error) {
	filePath := fs.getFilePath(service, environment, key)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found: %s/%s/%s", service, environment, key)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var item ConfigItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &item, nil
}

// Set 设置配置项
func (fs *FileStorage) Set(ctx context.Context, item *ConfigItem) error {
	filePath := fs.getFilePath(item.Service, item.Environment, item.Key)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 设置时间戳
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 触发变化事件
	fs.notifyChange(&ConfigChangeEvent{
		Type:        "update",
		Service:     item.Service,
		Environment: item.Environment,
		Key:         item.Key,
		NewValue:    item.Value,
		Timestamp:   now,
		User:        item.UpdatedBy,
	})

	return nil
}

// Delete 删除配置项
func (fs *FileStorage) Delete(ctx context.Context, service, environment, key string) error {
	filePath := fs.getFilePath(service, environment, key)

	// 先获取旧值用于事件通知
	oldItem, _ := fs.Get(ctx, service, environment, key)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config not found: %s/%s/%s", service, environment, key)
		}
		return fmt.Errorf("failed to delete config file: %w", err)
	}

	// 触发变化事件
	var oldValue interface{}
	if oldItem != nil {
		oldValue = oldItem.Value
	}

	fs.notifyChange(&ConfigChangeEvent{
		Type:        "delete",
		Service:     service,
		Environment: environment,
		Key:         key,
		OldValue:    oldValue,
		Timestamp:   time.Now(),
	})

	return nil
}

// List 列出配置项
func (storage *FileStorage) List(ctx context.Context, service, environment string) ([]*ConfigItem, error) {
	dirPath := storage.getDirPath(service, environment)

	var items []*ConfigItem

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			storage.logger.Warnf("Failed to read config file %s: %v", path, err)
			return nil
		}

		var item ConfigItem
		if err := json.Unmarshal(data, &item); err != nil {
			storage.logger.Warnf("Failed to unmarshal config file %s: %v", path, err)
			return nil
		}

		items = append(items, &item)
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	return items, nil
}

// GetByTags 根据标签获取配置项
func (storage *FileStorage) GetByTags(ctx context.Context, tags map[string]string) ([]*ConfigItem, error) {
	var items []*ConfigItem

	err := filepath.WalkDir(storage.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var item ConfigItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil
		}

		// 检查标签匹配
		if storage.matchTags(item.Tags, tags) {
			items = append(items, &item)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search configs by tags: %w", err)
	}

	return items, nil
}

// Watch 监听配置变化
func (fs *FileStorage) Watch(ctx context.Context, service, environment string) (<-chan *ConfigChangeEvent, error) {
	if fs.watcher == nil {
		return nil, fmt.Errorf("file watching is not enabled")
	}

	key := fmt.Sprintf("%s/%s", service, environment)

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if ch, exists := fs.watchers[key]; exists {
		return ch, nil
	}

	ch := make(chan *ConfigChangeEvent, 100)
	fs.watchers[key] = ch

	// 添加目录监听
	dirPath := fs.getDirPath(service, environment)
	if err := os.MkdirAll(dirPath, 0755); err == nil {
		fs.watcher.Add(dirPath)
	}

	return ch, nil
}

// GetHistory 获取配置历史（文件存储暂不支持）
func (fs *FileStorage) GetHistory(ctx context.Context, service, environment, key string, limit int) ([]*ConfigItem, error) {
	return nil, fmt.Errorf("history not supported in file storage")
}

// Backup 备份配置
func (fs *FileStorage) Backup(ctx context.Context, service, environment string) ([]byte, error) {
	items, err := fs.List(ctx, service, environment)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(items, "", "  ")
}

// Restore 恢复配置
func (fs *FileStorage) Restore(ctx context.Context, service, environment string, data []byte) error {
	var items []*ConfigItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}

	for _, item := range items {
		if err := fs.Set(ctx, item); err != nil {
			return fmt.Errorf("failed to restore config %s: %w", item.Key, err)
		}
	}

	return nil
}

// Close 关闭存储
func (fs *FileStorage) Close() error {
	if fs.watcher != nil {
		return fs.watcher.Close()
	}
	return nil
}

// 辅助方法
func (fs *FileStorage) getFilePath(service, environment, key string) string {
	return filepath.Join(fs.basePath, service, environment, key+".json")
}

func (fs *FileStorage) getDirPath(service, environment string) string {
	return filepath.Join(fs.basePath, service, environment)
}

func (fs *FileStorage) matchTags(itemTags, queryTags map[string]string) bool {
	for key, value := range queryTags {
		if itemTags[key] != value {
			return false
		}
	}
	return true
}

func (fs *FileStorage) handleFileEvents() {
	for {
		select {
		case event, ok := <-fs.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				fs.handleFileChange(event.Name, "update")
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				fs.handleFileChange(event.Name, "delete")
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				fs.handleFileChange(event.Name, "create")
			}

		case err, ok := <-fs.watcher.Errors:
			if !ok {
				return
			}
			fs.logger.Errorf("File watcher error: %v", err)
		}
	}
}

func (fs *FileStorage) handleFileChange(filePath, eventType string) {
	if !strings.HasSuffix(filePath, ".json") {
		return
	}

	// 解析文件路径获取服务和环境信息
	relPath, err := filepath.Rel(fs.basePath, filePath)
	if err != nil {
		return
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 3 {
		return
	}

	service := parts[0]
	environment := parts[1]
	key := strings.TrimSuffix(parts[2], ".json")

	event := &ConfigChangeEvent{
		Type:        eventType,
		Service:     service,
		Environment: environment,
		Key:         key,
		Timestamp:   time.Now(),
	}

	fs.notifyChange(event)
}

func (fs *FileStorage) notifyChange(event *ConfigChangeEvent) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	key := fmt.Sprintf("%s/%s", event.Service, event.Environment)
	if ch, exists := fs.watchers[key]; exists {
		select {
		case ch <- event:
		default:
			// 通道满了，丢弃事件
			fs.logger.Warnf("Config change event dropped for %s", key)
		}
	}
}
