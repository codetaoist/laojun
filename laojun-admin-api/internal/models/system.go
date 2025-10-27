package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemSetting 系统设置模型
type SystemSetting struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`    // JSON字符串，兼容多类型存储
	Type        string    `json:"type" db:"data_type"` // string, number, boolean, json
	Category    string    `json:"category" db:"category"`
	Description string    `json:"description" db:"description"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	IsSensitive bool      `json:"is_sensitive" db:"is_sensitive"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog 审计日志模型（查询返回）
type AuditLog struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	TargetID    *uuid.UUID `json:"target_id" db:"target_id"`
	TargetType  string     `json:"target_type" db:"target_type"`
	Action      string     `json:"action" db:"action"`
	Level       string     `json:"level" db:"level"`
	Description string     `json:"description" db:"description"`
	OldData     string     `json:"old_data" db:"old_data"`
	NewData     string     `json:"new_data" db:"new_data"`
	IPAddress   string     `json:"ip_address" db:"ip_address"`
	UserAgent   string     `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Metrics 性能指标
type Metrics struct {
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	DiskUsage   float64 `json:"diskUsage"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	// 系统版本信息
	SystemName  string `json:"systemName"`
	Version     string `json:"version"`
	BuildTime   string `json:"buildTime"`
	GitCommit   string `json:"gitCommit"`
	Environment string `json:"environment"`

	// 运行环境信息
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	GoVersion    string `json:"goVersion"`
	NodeVersion  string `json:"nodeVersion"`
	Database     string `json:"database"`
	Redis        string `json:"redis"`
	Nginx        string `json:"nginx"`

	// 许可证信息
	License LicenseInfo `json:"license"`
}

// LicenseInfo 许可证信息
type LicenseInfo struct {
	Type        string `json:"type"`        // 许可证类型：commercial, open_source, trial
	Description string `json:"description"` // 许可证描述
	ExpiryDate  string `json:"expiryDate"`  // 有效期至
	IsValid     bool   `json:"isValid"`     // 是否有效
}
