# 太上老君插件系统 - 共享模块

本目录包含插件系统的共享组件，提供统一的数据模型和同步机制，确保市场端和运行时端的数据一致性。

## 目录结构

```
shared/
├── models/          # 统一数据模型
│   ├── unified_plugin.go  # 统一插件数据结构
│   └── go.mod            # 模块定义
└── sync/            # 数据同步机制
    ├── data_sync.go      # 数据同步管理器
    ├── event_bus.go      # 事件总线实现
    ├── example_usage.go  # 使用示例
    └── go.mod           # 模块定义
```

## 核心功能

### 1. 统一数据模型 (models)

提供标准化的插件数据结构，确保市场端和运行时端使用一致的数据格式：

- **UnifiedPluginMetadata**: 统一的插件元数据结构
- **UnifiedPluginState**: 标准化的插件状态枚举
- **UnifiedPluginEvent**: 统一的插件事件结构
- **PluginSyncEvent**: 数据同步事件结构
- **DataSyncStatus**: 同步状态跟踪

### 2. 数据同步机制 (sync)

实现事件驱动的数据同步，支持市场端和运行时端的实时数据同步：

#### 核心组件

- **DataSyncManager**: 数据同步管理器，协调整个同步流程
- **EventBus**: 事件总线，支持发布/订阅模式的事件传递
- **DataTransformer**: 数据转换器，处理不同格式间的数据转换
- **SyncStorage**: 同步存储，记录同步历史和状态

#### 主要特性

- **事件驱动**: 基于事件的异步数据同步
- **双向同步**: 支持市场到运行时和运行时到市场的双向数据同步
- **数据转换**: 自动处理不同数据格式间的转换
- **状态跟踪**: 完整的同步历史和状态记录
- **错误处理**: 完善的错误处理和重试机制
- **事件过滤**: 支持基于类型、来源、优先级等条件的事件过滤

## 使用示例

### 数据同步

```go
import "github.com/codetaoist/laojun-plugins/shared/sync"

func main() {
    // 创建日志器
    logger := logrus.New()
    
    // 创建事件总线
    eventBus := sync.NewDefaultEventBus(1000, 4, logger)
    
    // 创建存储
    storage := sync.NewDefaultSyncStorage()
    
    // 创建数据同步管理器
    syncManager := sync.NewDataSyncManager(eventBus, storage, logger)
    
    // 注册数据转换器
    syncManager.RegisterTransformer(&sync.MarketToRuntimeTransformer{})
    syncManager.RegisterTransformer(&sync.RuntimeToMarketTransformer{})
    
    // 启动同步管理器
    ctx := context.Background()
    if err := syncManager.Start(ctx); err != nil {
        logger.Fatal("Failed to start sync manager:", err)
    }
    
    // 同步插件数据
    pluginID := uuid.New()
    changes := map[string]interface{}{
        "name": "Example Plugin",
        "version": "1.0.0",
        "state": "active",
    }
    
    err := syncManager.SyncPluginData(ctx, pluginID, "metadata_updated", changes, "market")
    if err != nil {
        logger.Error("Failed to sync plugin data:", err)
    }
}
```

### 自定义同步订阅者

```go
type CustomSyncSubscriber struct {
    logger *logrus.Logger
}

func (s *CustomSyncSubscriber) OnDataSync(ctx context.Context, event *models.PluginSyncEvent) error {
    // 处理同步事件
    s.logger.WithFields(logrus.Fields{
        "plugin_id": event.PluginID,
        "event_type": event.EventType,
        "source": event.Source,
    }).Info("Processing sync event")
    
    // 实现具体的同步逻辑
    return nil
}

func (s *CustomSyncSubscriber) GetSubscriptionTypes() []string {
    return []string{"metadata_updated", "state_changed"}
}

// 注册订阅者
syncManager.RegisterSubscriber(&CustomSyncSubscriber{logger: logger})
```

### 事件过滤

```go
// 创建事件过滤器
filter := &sync.EventFilter{
    EventTypes: []string{"plugin.sync.*"},
    Sources: []string{"market", "runtime"},
    MinPriority: 1,
    MaxAge: time.Hour,
}

// 创建带过滤器的事件总线
filteredBus := sync.NewFilteredEventBus(1000, 2, filter, logger)
```

## 同步事件类型

系统支持以下同步事件类型：

- `metadata_updated`: 插件元数据更新
- `state_changed`: 插件状态变更
- `installed`: 插件安装
- `uninstalled`: 插件卸载
- `config_updated`: 配置更新
- `health_check`: 健康检查

## 数据转换

系统提供内置的数据转换器：

- **MarketToRuntimeTransformer**: 将市场端数据转换为运行时格式
- **RuntimeToMarketTransformer**: 将运行时数据转换为市场端格式

也可以实现自定义转换器：

```go
type CustomTransformer struct{}

func (t *CustomTransformer) Transform(ctx context.Context, data interface{}) (interface{}, error) {
    // 实现数据转换逻辑
    return transformedData, nil
}

func (t *CustomTransformer) GetSourceType() string {
    return "custom_source"
}

func (t *CustomTransformer) GetTargetType() string {
    return "custom_target"
}

// 注册转换器
syncManager.RegisterTransformer(&CustomTransformer{})
```

## 监控和指标

事件总线提供丰富的监控指标：

```go
metrics := eventBus.GetMetrics()
logger.WithFields(logrus.Fields{
    "subscriber_count": metrics.SubscriberCount,
    "buffer_size": metrics.BufferSize,
    "buffer_usage": metrics.BufferUsage,
}).Info("Event bus metrics")
```

## 错误处理

系统提供完善的错误处理机制：

- 事件处理失败时会记录错误日志
- 支持同步状态跟踪和错误信息记录
- 提供重试机制和故障恢复

## 性能优化

- 使用缓冲通道避免阻塞
- 多工作器并发处理事件
- 支持事件过滤减少不必要的处理
- 内存中存储提供快速访问

## 扩展性

系统设计具有良好的扩展性：

- 接口化设计，易于替换实现
- 插件化的转换器和订阅者
- 支持自定义存储后端
- 灵活的事件过滤机制

## 注意事项

1. 确保在使用前启动同步管理器
2. 合理设置事件总线的缓冲区大小和工作器数量
3. 实现订阅者时注意错误处理
4. 定期检查同步状态和历史记录
5. 在生产环境中使用持久化存储替代内存存储

## 示例代码

完整的使用示例请参考 `example_usage.go` 文件，包含：

- 基本数据同步使用
- 数据转换示例
- 事件过滤示例
- 自定义订阅者实现

运行示例：

```go
import "github.com/codetaoist/laojun-plugins/shared/sync"

func main() {
    sync.RunAllExamples()
}
```