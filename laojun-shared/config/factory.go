package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ConfigManagerType 配置管理器类型
type ConfigManagerType string

const (
	// ConfigManagerTypeMemory 内存配置管理器
	ConfigManagerTypeMemory ConfigManagerType = "memory"
	// ConfigManagerTypeFile 文件配置管理器
	ConfigManagerTypeFile ConfigManagerType = "file"
	// ConfigManagerTypeHTTP HTTP配置管理器
	ConfigManagerTypeHTTP ConfigManagerType = "http"
	// ConfigManagerTypeGRPC gRPC配置管理器
	ConfigManagerTypeGRPC ConfigManagerType = "grpc"
	// ConfigManagerTypeEtcd etcd配置管理器
	ConfigManagerTypeEtcd ConfigManagerType = "etcd"
	// ConfigManagerTypeConsul Consul配置管理器
	ConfigManagerTypeConsul ConfigManagerType = "consul"
)

// ConfigWatcherType 配置监听器类型
type ConfigWatcherType string

const (
	// ConfigWatcherTypePolling 轮询监听器
	ConfigWatcherTypePolling ConfigWatcherType = "polling"
	// ConfigWatcherTypeEventDriven 事件驱动监听器
	ConfigWatcherTypeEventDriven ConfigWatcherType = "event-driven"
)

// ConfigStorageType 配置存储类型
type ConfigStorageType string

const (
	// ConfigStorageTypeMemory 内存存储
	ConfigStorageTypeMemory ConfigStorageType = "memory"
	// ConfigStorageTypeFile 文件存储
	ConfigStorageTypeFile ConfigStorageType = "file"
	// ConfigStorageTypeDatabase 数据库存储
	ConfigStorageTypeDatabase ConfigStorageType = "database"
	// ConfigStorageTypeRedis Redis存储
	ConfigStorageTypeRedis ConfigStorageType = "redis"
)

// ConfigClientType 配置客户端类型
type ConfigClientType string

const (
	// ConfigClientTypeHTTP HTTP客户端
	ConfigClientTypeHTTP ConfigClientType = "http"
	// ConfigClientTypeGRPC gRPC客户端
	ConfigClientTypeGRPC ConfigClientType = "grpc"
)

// FactoryConfig 工厂配置
type FactoryConfig struct {
	// 管理器类型
	ManagerType ConfigManagerType `json:"manager_type" yaml:"manager_type"`
	
	// 存储配置
	StorageType   ConfigStorageType `json:"storage_type" yaml:"storage_type"`
	StorageConfig map[string]interface{} `json:"storage_config" yaml:"storage_config"`
	
	// 客户端配置
	ClientType   ConfigClientType `json:"client_type" yaml:"client_type"`
	ClientConfig map[string]interface{} `json:"client_config" yaml:"client_config"`
	
	// 监听器配置
	WatcherType   ConfigWatcherType `json:"watcher_type" yaml:"watcher_type"`
	WatcherConfig map[string]interface{} `json:"watcher_config" yaml:"watcher_config"`
	
	// 通用配置选项
	Options *ConfigOptions `json:"options" yaml:"options"`
}

// ConfigFactory 配置管理工厂
type ConfigFactory struct {
	logger *logrus.Logger
}

// NewConfigFactory 创建配置管理工厂
func NewConfigFactory() *ConfigFactory {
	return &ConfigFactory{
		logger: logrus.New(),
	}
}

// CreateManager 创建配置管理器
func (f *ConfigFactory) CreateManager(config *FactoryConfig) (ConfigManager, error) {
	if config == nil {
		return nil, fmt.Errorf("factory config is required")
	}
	
	// 创建存储
	storage, err := f.CreateStorage(config.StorageType, config.StorageConfig, config.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	
	// 创建客户端（如果需要）
	var client ConfigClient
	if config.ClientType != "" {
		client, err = f.CreateClient(config.ClientType, config.ClientConfig, config.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}
	}
	
	// 创建监听器（如果需要）
	var watcher ConfigWatcher
	if config.WatcherType != "" {
		watcher, err = f.CreateWatcher(config.WatcherType, storage, config.WatcherConfig, config.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to create watcher: %w", err)
		}
	}
	
	// 根据管理器类型创建管理器
	switch config.ManagerType {
	case ConfigManagerTypeMemory, ConfigManagerTypeFile:
		return NewDefaultConfigManager(storage, client, watcher, config.Options), nil
	case ConfigManagerTypeHTTP, ConfigManagerTypeGRPC:
		if client == nil {
			return nil, fmt.Errorf("client is required for %s manager", config.ManagerType)
		}
		return NewDefaultConfigManager(storage, client, watcher, config.Options), nil
	default:
		return nil, fmt.Errorf("unsupported manager type: %s", config.ManagerType)
	}
}

// CreateStorage 创建配置存储
func (f *ConfigFactory) CreateStorage(storageType ConfigStorageType, config map[string]interface{}, options *ConfigOptions) (ConfigStorage, error) {
	switch storageType {
	case ConfigStorageTypeMemory:
		return NewMemoryConfigStorage(options), nil
	case ConfigStorageTypeFile:
		filePath, ok := config["file_path"].(string)
		if !ok {
			return nil, fmt.Errorf("file_path is required for file storage")
		}
		return NewFileConfigStorage(filePath, options), nil
	case ConfigStorageTypeDatabase:
		// TODO: 实现数据库存储
		return nil, fmt.Errorf("database storage not implemented yet")
	case ConfigStorageTypeRedis:
		// TODO: 实现Redis存储
		return nil, fmt.Errorf("redis storage not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// CreateClient 创建配置客户端
func (f *ConfigFactory) CreateClient(clientType ConfigClientType, config map[string]interface{}, options *ConfigOptions) (ConfigClient, error) {
	switch clientType {
	case ConfigClientTypeHTTP:
		baseURL, ok := config["base_url"].(string)
		if !ok {
			return nil, fmt.Errorf("base_url is required for HTTP client")
		}
		
		client := NewHTTPConfigClient(baseURL, options)
		
		// 设置认证信息（如果提供）
		if token, ok := config["token"].(string); ok {
			client.SetAuth(token, "", "")
		} else if username, ok := config["username"].(string); ok {
			if password, ok := config["password"].(string); ok {
				client.SetAuth("", username, password)
			}
		}
		
		return client, nil
	case ConfigClientTypeGRPC:
		address, ok := config["address"].(string)
		if !ok {
			return nil, fmt.Errorf("address is required for gRPC client")
		}
		return NewGRPCConfigClient(address, options), nil
	default:
		return nil, fmt.Errorf("unsupported client type: %s", clientType)
	}
}

// CreateWatcher 创建配置监听器
func (f *ConfigFactory) CreateWatcher(watcherType ConfigWatcherType, storage ConfigStorage, config map[string]interface{}, options *ConfigOptions) (ConfigWatcher, error) {
	switch watcherType {
	case ConfigWatcherTypePolling:
		return NewPollingConfigWatcher(storage, options), nil
	case ConfigWatcherTypeEventDriven:
		return NewEventDrivenConfigWatcher(storage, options), nil
	default:
		return nil, fmt.Errorf("unsupported watcher type: %s", watcherType)
	}
}

// CreateManagerFromURL 从URL创建配置管理器
func (f *ConfigFactory) CreateManagerFromURL(url string, options *ConfigOptions) (ConfigManager, error) {
	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}
	
	// 解析URL确定类型
	var config *FactoryConfig
	
	if strings.HasPrefix(url, "memory://") {
		config = &FactoryConfig{
			ManagerType: ConfigManagerTypeMemory,
			StorageType: ConfigStorageTypeMemory,
			Options:     options,
		}
	} else if strings.HasPrefix(url, "file://") {
		filePath := strings.TrimPrefix(url, "file://")
		config = &FactoryConfig{
			ManagerType: ConfigManagerTypeFile,
			StorageType: ConfigStorageTypeFile,
			StorageConfig: map[string]interface{}{
				"file_path": filePath,
			},
			Options: options,
		}
	} else if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		config = &FactoryConfig{
			ManagerType: ConfigManagerTypeHTTP,
			StorageType: ConfigStorageTypeMemory, // 使用内存作为缓存
			ClientType:  ConfigClientTypeHTTP,
			ClientConfig: map[string]interface{}{
				"base_url": url,
			},
			WatcherType: ConfigWatcherTypePolling,
			Options:     options,
		}
	} else if strings.HasPrefix(url, "grpc://") {
		address := strings.TrimPrefix(url, "grpc://")
		config = &FactoryConfig{
			ManagerType: ConfigManagerTypeGRPC,
			StorageType: ConfigStorageTypeMemory, // 使用内存作为缓存
			ClientType:  ConfigClientTypeGRPC,
			ClientConfig: map[string]interface{}{
				"address": address,
			},
			WatcherType: ConfigWatcherTypePolling,
			Options:     options,
		}
	} else {
		return nil, fmt.Errorf("unsupported URL scheme: %s", url)
	}
	
	return f.CreateManager(config)
}

// CreateManagerWithDefaults 使用默认配置创建配置管理器
func (f *ConfigFactory) CreateManagerWithDefaults() ConfigManager {
	config := &FactoryConfig{
		ManagerType: ConfigManagerTypeMemory,
		StorageType: ConfigStorageTypeMemory,
		Options: &ConfigOptions{
			CacheEnabled:    true,
			CacheSize:       1000,
			CacheTTL:        5 * time.Minute,
			Timeout:         30 * time.Second,
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			WatchInterval:   30 * time.Second,
			WatchBufferSize: 100,
		},
	}
	
	manager, err := f.CreateManager(config)
	if err != nil {
		f.logger.WithError(err).Error("Failed to create default config manager")
		// 返回一个基本的内存管理器作为后备
		return NewDefaultConfigManager(
			NewMemoryConfigStorage(config.Options),
			nil,
			nil,
			config.Options,
		)
	}
	
	return manager
}

// ValidateConfig 验证工厂配置
func (f *ConfigFactory) ValidateConfig(config *FactoryConfig) error {
	if config == nil {
		return fmt.Errorf("config is required")
	}
	
	// 验证管理器类型
	switch config.ManagerType {
	case ConfigManagerTypeMemory, ConfigManagerTypeFile, ConfigManagerTypeHTTP, ConfigManagerTypeGRPC:
		// 支持的类型
	default:
		return fmt.Errorf("unsupported manager type: %s", config.ManagerType)
	}
	
	// 验证存储类型
	switch config.StorageType {
	case ConfigStorageTypeMemory, ConfigStorageTypeFile:
		// 支持的类型
	case ConfigStorageTypeDatabase, ConfigStorageTypeRedis:
		return fmt.Errorf("storage type %s is not implemented yet", config.StorageType)
	default:
		return fmt.Errorf("unsupported storage type: %s", config.StorageType)
	}
	
	// 验证客户端配置
	if config.ClientType != "" {
		switch config.ClientType {
		case ConfigClientTypeHTTP:
			if config.ClientConfig["base_url"] == nil {
				return fmt.Errorf("base_url is required for HTTP client")
			}
		case ConfigClientTypeGRPC:
			if config.ClientConfig["address"] == nil {
				return fmt.Errorf("address is required for gRPC client")
			}
		default:
			return fmt.Errorf("unsupported client type: %s", config.ClientType)
		}
	}
	
	// 验证监听器配置
	if config.WatcherType != "" {
		switch config.WatcherType {
		case ConfigWatcherTypePolling, ConfigWatcherTypeEventDriven:
			// 支持的类型
		default:
			return fmt.Errorf("unsupported watcher type: %s", config.WatcherType)
		}
	}
	
	return nil
}

// GetSupportedTypes 获取支持的类型列表
func (f *ConfigFactory) GetSupportedTypes() map[string][]string {
	return map[string][]string{
		"managers": {
			string(ConfigManagerTypeMemory),
			string(ConfigManagerTypeFile),
			string(ConfigManagerTypeHTTP),
			string(ConfigManagerTypeGRPC),
		},
		"storages": {
			string(ConfigStorageTypeMemory),
			string(ConfigStorageTypeFile),
		},
		"clients": {
			string(ConfigClientTypeHTTP),
			string(ConfigClientTypeGRPC),
		},
		"watchers": {
			string(ConfigWatcherTypePolling),
			string(ConfigWatcherTypeEventDriven),
		},
	}
}

// 全局工厂实例
var defaultFactory = NewConfigFactory()

// CreateManager 使用默认工厂创建配置管理器
func CreateManager(config *FactoryConfig) (ConfigManager, error) {
	return defaultFactory.CreateManager(config)
}

// CreateManagerFromURL 使用默认工厂从URL创建配置管理器
func CreateManagerFromURL(url string, options *ConfigOptions) (ConfigManager, error) {
	return defaultFactory.CreateManagerFromURL(url, options)
}

// CreateManagerWithDefaults 使用默认工厂创建默认配置管理器
func CreateManagerWithDefaults() ConfigManager {
	return defaultFactory.CreateManagerWithDefaults()
}