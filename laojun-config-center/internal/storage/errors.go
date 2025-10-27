package storage

import "fmt"

// ConfigNotFoundError 配置未找到错误
type ConfigNotFoundError struct {
	Service     string
	Environment string
	Key         string
}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("config not found: service=%s, environment=%s, key=%s", 
		e.Service, e.Environment, e.Key)
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// StorageError 存储错误
type StorageError struct {
	Operation string
	Cause     error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error during %s: %v", e.Operation, e.Cause)
}

func (e *StorageError) Unwrap() error {
	return e.Cause
}

// DuplicateConfigError 重复配置错误
type DuplicateConfigError struct {
	Service     string
	Environment string
	Key         string
}

func (e *DuplicateConfigError) Error() string {
	return fmt.Sprintf("duplicate config: service=%s, environment=%s, key=%s", 
		e.Service, e.Environment, e.Key)
}

// VersionConflictError 版本冲突错误
type VersionConflictError struct {
	Service        string
	Environment    string
	Key            string
	CurrentVersion int64
	RequestVersion int64
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("version conflict: service=%s, environment=%s, key=%s, current=%d, request=%d", 
		e.Service, e.Environment, e.Key, e.CurrentVersion, e.RequestVersion)
}

// PermissionDeniedError 权限拒绝错误
type PermissionDeniedError struct {
	Operation string
	Resource  string
	User      string
}

func (e *PermissionDeniedError) Error() string {
	return fmt.Sprintf("permission denied: user=%s, operation=%s, resource=%s", 
		e.User, e.Operation, e.Resource)
}

// ConfigKey 配置键结构
type ConfigKey struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Key         string `json:"key"`
}

func (ck ConfigKey) String() string {
	return fmt.Sprintf("%s:%s:%s", ck.Service, ck.Environment, ck.Key)
}