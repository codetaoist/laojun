package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"github.com/codetaoist/laojun-config-center/internal/storage"
)

// FileStorage 文件存储实现
type FileStorage struct {
	basePath string
	mu       sync.RWMutex
	watchers map[string]*fsnotify.Watcher
	channels map[string]chan *storage.WatchEvent
}

// NewFileStorage 创建文件存储
func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &FileStorage{
		basePath: basePath,
		watchers: make(map[string]*fsnotify.Watcher),
		channels: make(map[string]chan *storage.WatchEvent),
	}, nil
}

// Get 获取配置
func (fs *FileStorage) Get(ctx context.Context, service, environment, key string) (*storage.ConfigItem, error) {
	filePath := fs.getConfigPath(service, environment, key)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &storage.ConfigNotFoundError{
				Service:     service,
				Environment: environment,
				Key:         key,
			}
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var item storage.ConfigItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &item, nil
}

// Set 设置配置
func (fs *FileStorage) Set(ctx context.Context, item *storage.ConfigItem) error {
	if err := fs.Validate(ctx, item); err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 检查是否存在旧配置
	oldItem, _ := fs.Get(ctx, item.Service, item.Environment, item.Key)

	// 设置版本和时间戳
	if oldItem != nil {
		item.Version = oldItem.Version + 1
		item.CreatedAt = oldItem.CreatedAt
	} else {
		item.Version = 1
		item.CreatedAt = time.Now()
	}
	item.UpdatedAt = time.Now()

	// 保存配置
	filePath := fs.getConfigPath(item.Service, item.Environment, item.Key)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 保存历史记录
	if err := fs.saveHistory(item, oldItem); err != nil {
		// 历史记录保存失败不影响主操作
		fmt.Printf("Warning: failed to save history: %v\n", err)
	}

	// 发送监听事件
	fs.sendWatchEvent(item, oldItem)

	return nil
}

// Delete 删除配置
func (fs *FileStorage) Delete(ctx context.Context, service, environment, key string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 获取旧配置用于历史记录
	oldItem, err := fs.Get(ctx, service, environment, key)
	if err != nil {
		return err
	}

	filePath := fs.getConfigPath(service, environment, key)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete config file: %w", err)
	}

	// 保存历史记录
	if err := fs.saveDeleteHistory(oldItem); err != nil {
		fmt.Printf("Warning: failed to save delete history: %v\n", err)
	}

	// 发送监听事件
	fs.sendDeleteEvent(oldItem)

	return nil
}

// List 列出配置
func (fs *FileStorage) List(ctx context.Context, service, environment string) ([]*storage.ConfigItem, error) {
	dirPath := fs.getServiceEnvPath(service, environment)
	
	var items []*storage.ConfigItem
	err := fs.walkDir(dirPath, func(path string, info os.FileInfo) error {
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // 忽略读取错误
		}

		var item storage.ConfigItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil // 忽略解析错误
		}

		items = append(items, &item)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// 按键名排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})

	return items, nil
}

// walkDir 递归遍历目录
func (fs *FileStorage) walkDir(dir string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}
		return fn(path, info)
	})
}

// Exists 检查配置是否存在
func (fs *FileStorage) Exists(ctx context.Context, service, environment, key string) (bool, error) {
	filePath := fs.getConfigPath(service, environment, key)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetMultiple 批量获取配置
func (fs *FileStorage) GetMultiple(ctx context.Context, keys []storage.ConfigKey) ([]*storage.ConfigItem, error) {
	var items []*storage.ConfigItem
	for _, key := range keys {
		item, err := fs.Get(ctx, key.Service, key.Environment, key.Key)
		if err != nil {
			if _, ok := err.(*storage.ConfigNotFoundError); ok {
				continue // 跳过不存在的配置
			}
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// SetMultiple 批量设置配置
func (fs *FileStorage) SetMultiple(ctx context.Context, items []*storage.ConfigItem) error {
	for _, item := range items {
		if err := fs.Set(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMultiple 批量删除配置
func (fs *FileStorage) DeleteMultiple(ctx context.Context, keys []storage.ConfigKey) error {
	for _, key := range keys {
		if err := fs.Delete(ctx, key.Service, key.Environment, key.Key); err != nil {
			if _, ok := err.(*storage.ConfigNotFoundError); ok {
				continue // 跳过不存在的配置
			}
			return err
		}
	}
	return nil
}

// Search 搜索配置
func (fs *FileStorage) Search(ctx context.Context, query *storage.SearchQuery) ([]*storage.ConfigItem, error) {
	var allItems []*storage.ConfigItem
	
	// 如果指定了服务和环境，只搜索该范围
	if query.Service != "" && query.Environment != "" {
		items, err := fs.List(ctx, query.Service, query.Environment)
		if err != nil {
			return nil, err
		}
		allItems = items
	} else {
		// 否则搜索所有配置
		err := fs.walkDir(fs.basePath, func(path string, info os.FileInfo) error {
			if info.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			var item storage.ConfigItem
			if err := json.Unmarshal(data, &item); err != nil {
				return nil
			}

			allItems = append(allItems, &item)
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to search configs: %w", err)
		}
	}

	// 过滤结果
	var results []*storage.ConfigItem
	for _, item := range allItems {
		if fs.matchesQuery(item, query) {
			results = append(results, item)
		}
	}

	// 分页
	if query.Offset > 0 {
		if query.Offset >= len(results) {
			return []*storage.ConfigItem{}, nil
		}
		results = results[query.Offset:]
	}

	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

// GetHistory 获取配置历史
func (fs *FileStorage) GetHistory(ctx context.Context, service, environment, key string, limit int) ([]*storage.ConfigHistory, error) {
	historyPath := fs.getHistoryPath(service, environment, key)
	
	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*storage.ConfigHistory{}, nil
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history []*storage.ConfigHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	// 按时间倒序排序
	sort.Slice(history, func(i, j int) bool {
		return history[i].CreatedAt.After(history[j].CreatedAt)
	})

	// 限制数量
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	return history, nil
}

// GetVersion 获取指定版本的配置
func (fs *FileStorage) GetVersion(ctx context.Context, service, environment, key string, version int64) (*storage.ConfigItem, error) {
	history, err := fs.GetHistory(ctx, service, environment, key, 0)
	if err != nil {
		return nil, err
	}

	for _, h := range history {
		if h.Version == version {
			return &storage.ConfigItem{
				Service:     h.Service,
				Environment: h.Environment,
				Key:         h.Key,
				Value:       h.NewValue,
				Version:     h.Version,
				CreatedAt:   h.CreatedAt,
				UpdatedAt:   h.CreatedAt,
				CreatedBy:   h.CreatedBy,
				UpdatedBy:   h.CreatedBy,
			}, nil
		}
	}

	return nil, &storage.ConfigNotFoundError{
		Service:     service,
		Environment: environment,
		Key:         key,
	}
}

// Rollback 回滚到指定版本
func (fs *FileStorage) Rollback(ctx context.Context, service, environment, key string, version int64, operator string) error {
	// 获取指定版本的配置
	versionItem, err := fs.GetVersion(ctx, service, environment, key, version)
	if err != nil {
		return err
	}

	// 设置操作者
	versionItem.UpdatedBy = operator
	versionItem.UpdatedAt = time.Now()

	// 保存配置
	return fs.Set(ctx, versionItem)
}

// Watch 监听配置变化
func (fs *FileStorage) Watch(ctx context.Context, service, environment string) (<-chan *storage.WatchEvent, error) {
	watchKey := service + "/" + environment
	
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 如果已经在监听，返回现有通道
	if ch, exists := fs.channels[watchKey]; exists {
		return ch, nil
	}

	// 创建文件监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	// 添加监听目录
	watchPath := fs.getServiceEnvPath(service, environment)
	if err := os.MkdirAll(watchPath, 0755); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to create watch directory: %w", err)
	}

	if err := watcher.Add(watchPath); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to add watch path: %w", err)
	}

	// 创建事件通道
	ch := make(chan *storage.WatchEvent, 100)
	fs.watchers[watchKey] = watcher
	fs.channels[watchKey] = ch

	// 启动监听协程
	go fs.watchLoop(watcher, ch, service, environment)

	return ch, nil
}

// StopWatch 停止监听
func (fs *FileStorage) StopWatch(service, environment string) {
	watchKey := service + "/" + environment
	
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if watcher, exists := fs.watchers[watchKey]; exists {
		watcher.Close()
		delete(fs.watchers, watchKey)
	}

	if ch, exists := fs.channels[watchKey]; exists {
		close(ch)
		delete(fs.channels, watchKey)
	}
}

// Backup 备份配置
func (fs *FileStorage) Backup(ctx context.Context, service, environment string) ([]byte, error) {
	items, err := fs.List(ctx, service, environment)
	if err != nil {
		return nil, err
	}

	backup := map[string]interface{}{
		"service":     service,
		"environment": environment,
		"timestamp":   time.Now(),
		"configs":     items,
	}

	return yaml.Marshal(backup)
}

// Restore 恢复配置
func (fs *FileStorage) Restore(ctx context.Context, service, environment string, data []byte, operator string) error {
	var backup map[string]interface{}
	if err := yaml.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}

	configsData, ok := backup["configs"]
	if !ok {
		return fmt.Errorf("invalid backup data: missing configs")
	}

	// 转换为配置项
	configsBytes, err := json.Marshal(configsData)
	if err != nil {
		return fmt.Errorf("failed to marshal configs: %w", err)
	}

	var items []*storage.ConfigItem
	if err := json.Unmarshal(configsBytes, &items); err != nil {
		return fmt.Errorf("failed to unmarshal configs: %w", err)
	}

	// 设置操作者
	for _, item := range items {
		item.UpdatedBy = operator
		item.UpdatedAt = time.Now()
	}

	// 批量设置配置
	return fs.SetMultiple(ctx, items)
}

// Validate 验证配置
func (fs *FileStorage) Validate(ctx context.Context, item *storage.ConfigItem) error {
	if item.Service == "" {
		return &storage.ValidationError{Field: "service", Message: "service is required"}
	}
	if item.Environment == "" {
		return &storage.ValidationError{Field: "environment", Message: "environment is required"}
	}
	if item.Key == "" {
		return &storage.ValidationError{Field: "key", Message: "key is required"}
	}
	if item.Value == nil {
		return &storage.ValidationError{Field: "value", Message: "value is required"}
	}
	return nil
}

// HealthCheck 健康检查
func (fs *FileStorage) HealthCheck(ctx context.Context) error {
	// 检查基础路径是否可访问
	if _, err := os.Stat(fs.basePath); err != nil {
		return fmt.Errorf("base path not accessible: %w", err)
	}

	// 尝试创建临时文件
	tempFile := filepath.Join(fs.basePath, ".health_check")
	if err := os.WriteFile(tempFile, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("cannot write to storage: %w", err)
	}

	// 清理临时文件
	os.Remove(tempFile)
	return nil
}

// Close 关闭存储
func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 关闭所有监听器
	for _, watcher := range fs.watchers {
		watcher.Close()
	}

	// 关闭所有通道
	for _, ch := range fs.channels {
		close(ch)
	}

	fs.watchers = make(map[string]*fsnotify.Watcher)
	fs.channels = make(map[string]chan *storage.WatchEvent)

	return nil
}

// 辅助方法

func (fs *FileStorage) getConfigPath(service, environment, key string) string {
	return filepath.Join(fs.basePath, service, environment, key+".json")
}

func (fs *FileStorage) getServiceEnvPath(service, environment string) string {
	return filepath.Join(fs.basePath, service, environment)
}

func (fs *FileStorage) getHistoryPath(service, environment, key string) string {
	return filepath.Join(fs.basePath, service, environment, ".history", key+".json")
}

func (fs *FileStorage) saveHistory(item *storage.ConfigItem, oldItem *storage.ConfigItem) error {
	historyPath := fs.getHistoryPath(item.Service, item.Environment, item.Key)
	
	// 创建历史目录
	if err := os.MkdirAll(filepath.Dir(historyPath), 0755); err != nil {
		return err
	}

	// 读取现有历史
	var history []*storage.ConfigHistory
	if data, err := os.ReadFile(historyPath); err == nil {
		json.Unmarshal(data, &history)
	}

	// 添加新历史记录
	operation := "create"
	var oldValue interface{}
	if oldItem != nil {
		operation = "update"
		oldValue = oldItem.Value
	}

	historyItem := &storage.ConfigHistory{
		ID:          time.Now().UnixNano(),
		Service:     item.Service,
		Environment: item.Environment,
		Key:         item.Key,
		OldValue:    oldValue,
		NewValue:    item.Value,
		Version:     item.Version,
		Operation:   operation,
		CreatedAt:   item.UpdatedAt,
		CreatedBy:   item.UpdatedBy,
	}

	history = append(history, historyItem)

	// 保存历史
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func (fs *FileStorage) saveDeleteHistory(item *storage.ConfigItem) error {
	historyPath := fs.getHistoryPath(item.Service, item.Environment, item.Key)
	
	// 读取现有历史
	var history []*storage.ConfigHistory
	if data, err := os.ReadFile(historyPath); err == nil {
		json.Unmarshal(data, &history)
	}

	// 添加删除记录
	historyItem := &storage.ConfigHistory{
		ID:          time.Now().UnixNano(),
		Service:     item.Service,
		Environment: item.Environment,
		Key:         item.Key,
		OldValue:    item.Value,
		NewValue:    nil,
		Version:     item.Version + 1,
		Operation:   "delete",
		CreatedAt:   time.Now(),
		CreatedBy:   "system",
	}

	history = append(history, historyItem)

	// 保存历史
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func (fs *FileStorage) sendWatchEvent(item *storage.ConfigItem, oldItem *storage.ConfigItem) {
	watchKey := item.Service + "/" + item.Environment
	
	if ch, exists := fs.channels[watchKey]; exists {
		eventType := "create"
		var oldValue interface{}
		if oldItem != nil {
			eventType = "update"
			oldValue = oldItem.Value
		}

		event := &storage.WatchEvent{
			Type:        eventType,
			Service:     item.Service,
			Environment: item.Environment,
			Key:         item.Key,
			OldValue:    oldValue,
			NewValue:    item.Value,
			Version:     item.Version,
			Timestamp:   time.Now(),
		}

		select {
		case ch <- event:
		default:
			// 通道满了，丢弃事件
		}
	}
}

func (fs *FileStorage) sendDeleteEvent(item *storage.ConfigItem) {
	watchKey := item.Service + "/" + item.Environment
	
	if ch, exists := fs.channels[watchKey]; exists {
		event := &storage.WatchEvent{
			Type:        "delete",
			Service:     item.Service,
			Environment: item.Environment,
			Key:         item.Key,
			OldValue:    item.Value,
			NewValue:    nil,
			Version:     item.Version,
			Timestamp:   time.Now(),
		}

		select {
		case ch <- event:
		default:
			// 通道满了，丢弃事件
		}
	}
}

func (fs *FileStorage) matchesQuery(item *storage.ConfigItem, query *storage.SearchQuery) bool {
	// 服务匹配
	if query.Service != "" && !strings.Contains(strings.ToLower(item.Service), strings.ToLower(query.Service)) {
		return false
	}

	// 环境匹配
	if query.Environment != "" && !strings.Contains(strings.ToLower(item.Environment), strings.ToLower(query.Environment)) {
		return false
	}

	// 键匹配
	if query.Key != "" && !strings.Contains(strings.ToLower(item.Key), strings.ToLower(query.Key)) {
		return false
	}

	// 值匹配
	if query.Value != "" {
		valueStr := fmt.Sprintf("%v", item.Value)
		if !strings.Contains(strings.ToLower(valueStr), strings.ToLower(query.Value)) {
			return false
		}
	}

	// 标签匹配
	if len(query.Tags) > 0 {
		for _, queryTag := range query.Tags {
			found := false
			for _, itemTag := range item.Tags {
				if strings.EqualFold(itemTag, queryTag) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// 元数据匹配
	if len(query.Metadata) > 0 {
		for key, value := range query.Metadata {
			if itemValue, exists := item.Metadata[key]; !exists {
				return false
			} else if itemValueStr := fmt.Sprintf("%v", itemValue); !strings.Contains(strings.ToLower(itemValueStr), strings.ToLower(value)) {
				return false
			}
		}
	}

	return true
}

func (fs *FileStorage) watchLoop(watcher *fsnotify.Watcher, ch chan *storage.WatchEvent, service, environment string) {
	defer watcher.Close()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// 只处理配置文件的变化
			if !strings.HasSuffix(event.Name, ".json") || strings.Contains(event.Name, ".history") {
				continue
			}

			// 解析文件名获取配置键
			key := strings.TrimSuffix(filepath.Base(event.Name), ".json")

			var watchEvent *storage.WatchEvent
			if event.Op&fsnotify.Write == fsnotify.Write {
				// 文件修改
				if item, err := fs.Get(context.Background(), service, environment, key); err == nil {
					watchEvent = &storage.WatchEvent{
						Type:        "update",
						Service:     service,
						Environment: environment,
						Key:         key,
						NewValue:    item.Value,
						Version:     item.Version,
						Timestamp:   time.Now(),
					}
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				// 文件删除
				watchEvent = &storage.WatchEvent{
					Type:        "delete",
					Service:     service,
					Environment: environment,
					Key:         key,
					Timestamp:   time.Now(),
				}
			}

			if watchEvent != nil {
				select {
				case ch <- watchEvent:
				default:
					// 通道满了，丢弃事件
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watch error: %v\n", err)
		}
	}
}