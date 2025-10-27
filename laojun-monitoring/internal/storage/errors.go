package storage

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	ErrUnsupportedStorageType = errors.New("unsupported storage type")
	ErrStorageNotInitialized  = errors.New("storage not initialized")
	ErrStorageConnectionFailed = errors.New("storage connection failed")
	ErrInvalidQuery           = errors.New("invalid query")
	ErrQueryTimeout           = errors.New("query timeout")
	ErrWriteFailed            = errors.New("write operation failed")
	ErrReadFailed             = errors.New("read operation failed")
	ErrStorageUnavailable     = errors.New("storage unavailable")
	ErrInvalidTimeRange       = errors.New("invalid time range")
	ErrTooManyResults         = errors.New("too many results")
)

// StorageError 存储错误
type StorageError struct {
	Op      string // 操作名称
	Storage string // 存储类型
	Err     error  // 原始错误
}

func (e *StorageError) Error() string {
	if e.Storage != "" {
		return fmt.Sprintf("storage %s: %s: %v", e.Storage, e.Op, e.Err)
	}
	return fmt.Sprintf("storage operation %s: %v", e.Op, e.Err)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError 创建存储错误
func NewStorageError(op, storage string, err error) *StorageError {
	return &StorageError{
		Op:      op,
		Storage: storage,
		Err:     err,
	}
}

// QueryError 查询错误
type QueryError struct {
	Query   string
	Message string
	Err     error
}

func (e *QueryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("query error: %s (query: %s): %v", e.Message, e.Query, e.Err)
	}
	return fmt.Sprintf("query error: %s (query: %s)", e.Message, e.Query)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}

// NewQueryError 创建查询错误
func NewQueryError(query, message string, err error) *QueryError {
	return &QueryError{
		Query:   query,
		Message: message,
		Err:     err,
	}
}

// ConnectionError 连接错误
type ConnectionError struct {
	Address string
	Err     error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection failed to %s: %v", e.Address, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// NewConnectionError 创建连接错误
func NewConnectionError(address string, err error) *ConnectionError {
	return &ConnectionError{
		Address: address,
		Err:     err,
	}
}