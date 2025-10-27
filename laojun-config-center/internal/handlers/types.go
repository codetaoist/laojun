package handlers

import (
	"net/http"
	"time"
	
	"github.com/gorilla/websocket"
	"github.com/codetaoist/laojun-config-center/internal/storage"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该进行适当的来源检查
	},
}

// SetConfigRequest 设置配置请求
type SetConfigRequest struct {
	Value       interface{}            `json:"value" binding:"required"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	UpdatedBy   string                 `json:"updated_by"`
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	Operation string                  `json:"operation" binding:"required"` // set, delete, get
	Configs   []*storage.ConfigItem   `json:"configs,omitempty"`
	Keys      []storage.ConfigKey     `json:"keys,omitempty"`
}

// BatchResult 批量操作结果
type BatchResult struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Key         string `json:"key"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Service     string   `json:"service"`
	Environment string   `json:"environment"`
	Key         string   `json:"key"`
	Value       string   `json:"value"`
	Tags        []string `json:"tags"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
}

// RollbackRequest 回滚请求
type RollbackRequest struct {
	Version  int64  `json:"version" binding:"required"`
	Operator string `json:"operator" binding:"required"`
}

// RestoreRequest 恢复请求
type RestoreRequest struct {
	Configs  []*storage.ConfigItem `json:"configs" binding:"required"`
	Operator string                `json:"operator" binding:"required"`
}

// WatchEvent WebSocket事件
type WatchEvent struct {
	Type        string      `json:"type"`
	Service     string      `json:"service"`
	Environment string      `json:"environment"`
	Key         string      `json:"key"`
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty"`
	Version     int64       `json:"version"`
	Timestamp   time.Time   `json:"timestamp"`
	Operator    string      `json:"operator,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// ListResponse 列表响应
type ListResponse struct {
	Configs []*storage.ConfigItem `json:"configs"`
	Count   int                   `json:"count"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Configs []*storage.ConfigItem `json:"configs"`
	Count   int                   `json:"count"`
	Query   *storage.SearchQuery  `json:"query"`
}

// HistoryResponse 历史响应
type HistoryResponse struct {
	History []*storage.ConfigHistory `json:"history"`
	Count   int                      `json:"count"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}