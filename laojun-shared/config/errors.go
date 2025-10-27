package config

import "errors"

// 配置管理相关错误
var (
	// ErrConfigNotFound 配置未找到
	ErrConfigNotFound = errors.New("config not found")
	
	// ErrConfigExists 配置已存在
	ErrConfigExists = errors.New("config already exists")
	
	// ErrInvalidConfigKey 无效的配置键
	ErrInvalidConfigKey = errors.New("invalid config key")
	
	// ErrInvalidConfigValue 无效的配置值
	ErrInvalidConfigValue = errors.New("invalid config value")
	
	// ErrInvalidConfigType 无效的配置类型
	ErrInvalidConfigType = errors.New("invalid config type")
	
	// ErrConfigValidationFailed 配置验证失败
	ErrConfigValidationFailed = errors.New("config validation failed")
	
	// ErrConfigLocked 配置被锁定
	ErrConfigLocked = errors.New("config is locked")
	
	// ErrConfigExpired 配置已过期
	ErrConfigExpired = errors.New("config has expired")
	
	// ErrTransactionFailed 事务失败
	ErrTransactionFailed = errors.New("transaction failed")
	
	// ErrTransactionNotStarted 事务未开始
	ErrTransactionNotStarted = errors.New("transaction not started")
	
	// ErrTransactionAlreadyStarted 事务已开始
	ErrTransactionAlreadyStarted = errors.New("transaction already started")
	
	// ErrWatcherNotStarted 监听器未启动
	ErrWatcherNotStarted = errors.New("watcher not started")
	
	// ErrWatcherAlreadyStarted 监听器已启动
	ErrWatcherAlreadyStarted = errors.New("watcher already started")
	
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("connection failed")
	
	// ErrTimeout 超时
	ErrTimeout = errors.New("operation timeout")
	
	// ErrPermissionDenied 权限被拒绝
	ErrPermissionDenied = errors.New("permission denied")
	
	// ErrServiceUnavailable 服务不可用
	ErrServiceUnavailable = errors.New("service unavailable")
	
	// ErrInvalidURL 无效的URL
	ErrInvalidURL = errors.New("invalid URL")
	
	// ErrUnsupportedOperation 不支持的操作
	ErrUnsupportedOperation = errors.New("unsupported operation")
	
	// ErrInvalidConfiguration 无效的配置
	ErrInvalidConfiguration = errors.New("invalid configuration")
	
	// ErrStorageNotAvailable 存储不可用
	ErrStorageNotAvailable = errors.New("storage not available")
	
	// ErrClientNotInitialized 客户端未初始化
	ErrClientNotInitialized = errors.New("client not initialized")
	
	// ErrManagerClosed 管理器已关闭
	ErrManagerClosed = errors.New("manager is closed")
)