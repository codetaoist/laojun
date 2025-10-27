package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisStorage Redis存储实现
type RedisStorage struct {
	client *redis.Client
	logger *zap.Logger
	
	// 监听器管理
	watchers map[string]chan *WatchEvent
	pubsub   *redis.PubSub
}

// NewRedisStorage 创建Redis存储
func NewRedisStorage(addr, password string, db int, logger *zap.Logger) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	storage := &RedisStorage{
		client:   client,
		logger:   logger,
		watchers: make(map[string]chan *WatchEvent),
	}

	// 启动发布订阅监听
	storage.startPubSubListener()

	return storage, nil
}

// Get 获取配置项
func (r *RedisStorage) Get(ctx context.Context, service, environment, key string) (*ConfigItem, error) {
	configKey := r.buildKey(service, environment, key)
	
	data, err := r.client.Get(ctx, configKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, &ConfigNotFoundError{
				Service:     service,
				Environment: environment,
				Key:         key,
			}
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	var item ConfigItem
	if err := json.Unmarshal([]byte(data), &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &item, nil
}

// Set 设置配置项
func (r *RedisStorage) Set(ctx context.Context, item *ConfigItem) error {
	if err := r.Validate(ctx, item); err != nil {
		return err
	}

	configKey := r.buildKey(item.Service, item.Environment, item.Key)
	
	// 检查是否存在旧值
	var oldItem *ConfigItem
	if existingData, err := r.client.Get(ctx, configKey).Result(); err == nil {
		oldItem = &ConfigItem{}
		json.Unmarshal([]byte(existingData), oldItem)
	}

	// 更新版本和时间戳
	if oldItem != nil {
		item.Version = oldItem.Version + 1
		item.CreatedAt = oldItem.CreatedAt
	} else {
		item.Version = 1
		item.CreatedAt = time.Now()
	}
	item.UpdatedAt = time.Now()

	// 序列化配置项
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 使用事务保存配置和历史记录
	pipe := r.client.TxPipeline()
	
	// 保存配置
	pipe.Set(ctx, configKey, data, 0)
	
	// 保存历史记录
	historyKey := r.buildHistoryKey(item.Service, item.Environment, item.Key)
	history := &ConfigHistory{
		ID:          time.Now().UnixNano(),
		Service:     item.Service,
		Environment: item.Environment,
		Key:         item.Key,
		NewValue:    item.Value,
		Version:     item.Version,
		CreatedAt:   time.Now(),
		CreatedBy:   item.UpdatedBy,
	}
	
	if oldItem != nil {
		history.Operation = "update"
		history.OldValue = oldItem.Value
	} else {
		history.Operation = "create"
	}

	historyData, _ := json.Marshal(history)
	pipe.LPush(ctx, historyKey, historyData)
	pipe.LTrim(ctx, historyKey, 0, 99) // 保留最近100条历史记录

	// 执行事务
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// 发布变更事件
	r.publishEvent(&WatchEvent{
		Type:        history.Operation,
		Service:     item.Service,
		Environment: item.Environment,
		Key:         item.Key,
		OldValue:    history.OldValue,
		NewValue:    item.Value,
		Version:     item.Version,
		Timestamp:   time.Now(),
	})

	return nil
}

// Delete 删除配置项
func (r *RedisStorage) Delete(ctx context.Context, service, environment, key string) error {
	configKey := r.buildKey(service, environment, key)
	
	// 获取旧值用于历史记录
	oldItem, err := r.Get(ctx, service, environment, key)
	if err != nil {
		return err
	}

	// 使用事务删除配置和添加历史记录
	pipe := r.client.TxPipeline()
	
	// 删除配置
	pipe.Del(ctx, configKey)
	
	// 添加删除历史记录
	historyKey := r.buildHistoryKey(service, environment, key)
	history := &ConfigHistory{
		ID:          time.Now().UnixNano(),
		Service:     service,
		Environment: environment,
		Key:         key,
		OldValue:    oldItem.Value,
		Version:     oldItem.Version + 1,
		Operation:   "delete",
		CreatedAt:   time.Now(),
	}
	
	historyData, _ := json.Marshal(history)
	pipe.LPush(ctx, historyKey, historyData)

	// 执行事务
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	// 发布删除事件
	r.publishEvent(&WatchEvent{
		Type:        "delete",
		Service:     service,
		Environment: environment,
		Key:         key,
		OldValue:    oldItem.Value,
		Version:     history.Version,
		Timestamp:   time.Now(),
	})

	return nil
}

// List 列出配置项
func (r *RedisStorage) List(ctx context.Context, service, environment string) ([]*ConfigItem, error) {
	pattern := r.buildKey(service, environment, "*")
	
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list config keys: %w", err)
	}

	if len(keys) == 0 {
		return []*ConfigItem{}, nil
	}

	// 批量获取配置
	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to get configs: %w", err)
	}

	var items []*ConfigItem
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			continue
		}

		var item ConfigItem
		if err := json.Unmarshal([]byte(data), &item); err != nil {
			continue
		}
		items = append(items, &item)
	}

	return items, nil
}

// Exists 检查配置是否存在
func (r *RedisStorage) Exists(ctx context.Context, service, environment, key string) (bool, error) {
	configKey := r.buildKey(service, environment, key)
	
	count, err := r.client.Exists(ctx, configKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check config existence: %w", err)
	}

	return count > 0, nil
}

// GetMultiple 批量获取配置
func (r *RedisStorage) GetMultiple(ctx context.Context, keys []ConfigKey) ([]*ConfigItem, error) {
	if len(keys) == 0 {
		return []*ConfigItem{}, nil
	}

	// 构建Redis键
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = r.buildKey(key.Service, key.Environment, key.Key)
	}

	// 批量获取
	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(redisKeys))
	
	for i, key := range redisKeys {
		cmds[i] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to get multiple configs: %w", err)
	}

	var items []*ConfigItem
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			continue
		}

		var item ConfigItem
		if err := json.Unmarshal([]byte(data), &item); err != nil {
			continue
		}
		items = append(items, &item)
	}

	return items, nil
}

// SetMultiple 批量设置配置
func (r *RedisStorage) SetMultiple(ctx context.Context, items []*ConfigItem) error {
	if len(items) == 0 {
		return nil
	}

	// 验证所有配置项
	for _, item := range items {
		if err := r.Validate(ctx, item); err != nil {
			return err
		}
	}

	// 使用事务批量设置
	pipe := r.client.TxPipeline()
	
	for _, item := range items {
		configKey := r.buildKey(item.Service, item.Environment, item.Key)
		
		// 更新时间戳和版本
		item.UpdatedAt = time.Now()
		if item.Version == 0 {
			item.Version = 1
			item.CreatedAt = time.Now()
		}

		data, _ := json.Marshal(item)
		pipe.Set(ctx, configKey, data, 0)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to set multiple configs: %w", err)
	}

	return nil
}

// DeleteMultiple 批量删除配置
func (r *RedisStorage) DeleteMultiple(ctx context.Context, keys []ConfigKey) error {
	if len(keys) == 0 {
		return nil
	}

	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = r.buildKey(key.Service, key.Environment, key.Key)
	}

	if err := r.client.Del(ctx, redisKeys...).Err(); err != nil {
		return fmt.Errorf("failed to delete multiple configs: %w", err)
	}

	return nil
}

// Search 搜索配置
func (r *RedisStorage) Search(ctx context.Context, query *SearchQuery) ([]*ConfigItem, error) {
	// 构建搜索模式
	pattern := r.buildKey(
		r.getSearchPattern(query.Service),
		r.getSearchPattern(query.Environment),
		r.getSearchPattern(query.Key),
	)

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to search configs: %w", err)
	}

	if len(keys) == 0 {
		return []*ConfigItem{}, nil
	}

	// 获取配置项
	var items []*ConfigItem
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var item ConfigItem
		if err := json.Unmarshal([]byte(data), &item); err != nil {
			continue
		}

		// 应用过滤条件
		if r.matchesQuery(&item, query) {
			items = append(items, &item)
		}
	}

	// 应用分页
	if query.Limit > 0 {
		start := query.Offset
		end := start + query.Limit
		if start >= len(items) {
			return []*ConfigItem{}, nil
		}
		if end > len(items) {
			end = len(items)
		}
		items = items[start:end]
	}

	return items, nil
}

// GetHistory 获取配置历史
func (r *RedisStorage) GetHistory(ctx context.Context, service, environment, key string, limit int) ([]*ConfigHistory, error) {
	historyKey := r.buildHistoryKey(service, environment, key)
	
	if limit <= 0 {
		limit = 10
	}

	data, err := r.client.LRange(ctx, historyKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get config history: %w", err)
	}

	var history []*ConfigHistory
	for _, item := range data {
		var h ConfigHistory
		if err := json.Unmarshal([]byte(item), &h); err != nil {
			continue
		}
		history = append(history, &h)
	}

	return history, nil
}

// GetVersion 获取指定版本的配置
func (r *RedisStorage) GetVersion(ctx context.Context, service, environment, key string, version int64) (*ConfigItem, error) {
	history, err := r.GetHistory(ctx, service, environment, key, 100)
	if err != nil {
		return nil, err
	}

	for _, h := range history {
		if h.Version == version && h.Operation != "delete" {
			return &ConfigItem{
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

	return nil, &ConfigNotFoundError{
		Service:     service,
		Environment: environment,
		Key:         key,
	}
}

// Rollback 回滚配置
func (r *RedisStorage) Rollback(ctx context.Context, service, environment, key string, version int64, operator string) error {
	// 获取指定版本的配置
	item, err := r.GetVersion(ctx, service, environment, key, version)
	if err != nil {
		return err
	}

	// 更新操作者
	item.UpdatedBy = operator
	
	// 重新设置配置
	return r.Set(ctx, item)
}

// Watch 监听配置变更
func (r *RedisStorage) Watch(ctx context.Context, service, environment string) (<-chan *WatchEvent, error) {
	watchKey := fmt.Sprintf("%s:%s", service, environment)
	
	// 创建事件通道
	eventChan := make(chan *WatchEvent, 100)
	r.watchers[watchKey] = eventChan

	return eventChan, nil
}

// StopWatch 停止监听
func (r *RedisStorage) StopWatch(service, environment string) {
	watchKey := fmt.Sprintf("%s:%s", service, environment)
	
	if ch, exists := r.watchers[watchKey]; exists {
		close(ch)
		delete(r.watchers, watchKey)
	}
}

// Backup 备份配置
func (r *RedisStorage) Backup(ctx context.Context, service, environment string) ([]byte, error) {
	items, err := r.List(ctx, service, environment)
	if err != nil {
		return nil, err
	}

	return json.Marshal(items)
}

// Restore 恢复配置
func (r *RedisStorage) Restore(ctx context.Context, service, environment string, data []byte, operator string) error {
	var items []*ConfigItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}

	// 更新操作者
	for _, item := range items {
		item.UpdatedBy = operator
	}

	return r.SetMultiple(ctx, items)
}

// Validate 验证配置项
func (r *RedisStorage) Validate(ctx context.Context, item *ConfigItem) error {
	if item.Service == "" {
		return &ValidationError{Field: "service", Message: "service is required"}
	}
	if item.Environment == "" {
		return &ValidationError{Field: "environment", Message: "environment is required"}
	}
	if item.Key == "" {
		return &ValidationError{Field: "key", Message: "key is required"}
	}
	if item.Value == nil {
		return &ValidationError{Field: "value", Message: "value is required"}
	}

	return nil
}

// HealthCheck 健康检查
func (r *RedisStorage) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭存储
func (r *RedisStorage) Close() error {
	// 关闭所有监听器
	for _, ch := range r.watchers {
		close(ch)
	}
	
	// 关闭发布订阅
	if r.pubsub != nil {
		r.pubsub.Close()
	}

	return r.client.Close()
}

// 辅助方法

// buildKey 构建Redis键
func (r *RedisStorage) buildKey(service, environment, key string) string {
	return fmt.Sprintf("config:%s:%s:%s", service, environment, key)
}

// buildHistoryKey 构建历史记录键
func (r *RedisStorage) buildHistoryKey(service, environment, key string) string {
	return fmt.Sprintf("config_history:%s:%s:%s", service, environment, key)
}

// getSearchPattern 获取搜索模式
func (r *RedisStorage) getSearchPattern(value string) string {
	if value == "" {
		return "*"
	}
	return value
}

// matchesQuery 检查配置项是否匹配查询条件
func (r *RedisStorage) matchesQuery(item *ConfigItem, query *SearchQuery) bool {
	// 检查值匹配
	if query.Value != "" {
		valueStr := fmt.Sprintf("%v", item.Value)
		if !strings.Contains(strings.ToLower(valueStr), strings.ToLower(query.Value)) {
			return false
		}
	}

	// 检查标签匹配
	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			found := false
			for _, itemTag := range item.Tags {
				if itemTag == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// startPubSubListener 启动发布订阅监听器
func (r *RedisStorage) startPubSubListener() {
	r.pubsub = r.client.Subscribe(context.Background(), "config_events")
	
	go func() {
		for msg := range r.pubsub.Channel() {
			var event WatchEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}

			// 分发事件到相应的监听器
			watchKey := fmt.Sprintf("%s:%s", event.Service, event.Environment)
			if ch, exists := r.watchers[watchKey]; exists {
				select {
				case ch <- &event:
				default:
					// 通道满了，跳过这个事件
				}
			}
		}
	}()
}

// publishEvent 发布事件
func (r *RedisStorage) publishEvent(event *WatchEvent) {
	data, _ := json.Marshal(event)
	r.client.Publish(context.Background(), "config_events", data)
}