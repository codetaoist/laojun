package services

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemService 系统相关服务
type SystemService struct {
	db *sql.DB
}

func NewSystemService(db *sql.DB) *SystemService {
	return &SystemService{db: db}
}

// GetSettings 获取系统设置
func (s *SystemService) GetSettings(ctx context.Context) ([]models.SystemSetting, error) {
	query := `SELECT id, key, value, data_type, description, is_public, created_at, updated_at FROM sys_settings ORDER BY key`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询系统设置失败: %w", err)
	}
	defer rows.Close()

	var settings []models.SystemSetting
	for rows.Next() {
		var m models.SystemSetting
		if err := rows.Scan(&m.ID, &m.Key, &m.Value, &m.Type, &m.Description, &m.IsPublic, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, m)
	}
	return settings, nil
}

// SaveSettings 批量保存系统设置（UPSERT）
func (s *SystemService) SaveSettings(ctx context.Context, settings []models.SystemSetting) error {
	if len(settings) == 0 {
		return nil
	}

	// 使用事务
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO sys_settings (key, value, data_type, description, is_public, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (key) DO UPDATE SET 
			value = EXCLUDED.value,
			data_type = EXCLUDED.data_type,
			description = EXCLUDED.description,
			is_public = EXCLUDED.is_public,
			updated_at = NOW();
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, m := range settings {
		if _, err := stmt.ExecContext(ctx, m.Key, m.Value, m.Type, m.Description, m.IsPublic); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetAuditLogs 获取审计日志
func (s *SystemService) GetAuditLogs(ctx context.Context, page, pageSize int, level, module, startDate, endDate string) ([]models.AuditLog, int64, error) {
	base := `SELECT id, user_id, target_id, target_type, action, level, description, old_data, new_data, ip_address, user_agent, created_at FROM sys_audit_logs WHERE 1=1`
	countBase := `SELECT COUNT(*) FROM sys_audit_logs WHERE 1=1`
	args := []interface{}{}
	where := ""

	idx := 1
	if level != "" {
		where += fmt.Sprintf(" AND level = $%d", idx)
		args = append(args, level)
		idx++
	}
	if module != "" {
		where += fmt.Sprintf(" AND target_type = $%d", idx)
		args = append(args, module)
		idx++
	}
	if startDate != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", idx)
		args = append(args, startDate)
		idx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", idx)
		args = append(args, endDate)
		idx++
	}

	// 统计总数
	var total int64
	if err := s.db.QueryRowContext(ctx, countBase+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计审计日志失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := base + where + fmt.Sprintf(" ORDER BY created_at DESC OFFSET %d LIMIT %d", offset, pageSize)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询审计日志失败: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var m models.AuditLog
		if err := rows.Scan(&m.ID, &m.UserID, &m.TargetID, &m.TargetType, &m.Action, &m.Level, &m.Description, &m.OldData, &m.NewData, &m.IPAddress, &m.UserAgent, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, m)
	}
	return logs, total, nil
}

// ClearAuditLogs 清理审计日志（按过滤条件）
func (s *SystemService) ClearAuditLogs(ctx context.Context, level, module, before string) (int64, error) {
	base := `DELETE FROM sys_audit_logs WHERE 1=1`
	args := []interface{}{}
	where := ""
	idx := 1
	if level != "" {
		where += fmt.Sprintf(" AND level = $%d", idx)
		args = append(args, level)
		idx++
	}
	if module != "" {
		where += fmt.Sprintf(" AND target_type = $%d", idx)
		args = append(args, module)
		idx++
	}
	if before != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", idx)
		args = append(args, before)
		idx++
	}
	res, err := s.db.ExecContext(ctx, base+where, args...)
	if err != nil {
		return 0, fmt.Errorf("清理审计日志失败: %w", err)
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}

// GetMetrics 获取性能指标
func (s *SystemService) GetMetrics(ctx context.Context) (models.Metrics, error) {
	var result models.Metrics

	// CPU 使用率（取一次采样平均）
	cpuPercents, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	if err == nil && len(cpuPercents) > 0 {
		result.CPUUsage = cpuPercents[0]
	}

	// 内存使用
	if vmStat, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		result.MemoryUsage = vmStat.UsedPercent
	}

	// 磁盘使用率（选择系统盘或根分区）
	if parts, err := disk.PartitionsWithContext(ctx, true); err == nil {
		// 选择第一个有效分
		for _, p := range parts {
			if usage, err := disk.UsageWithContext(ctx, p.Mountpoint); err == nil {
				// 选择第一个有效分
				result.DiskUsage = usage.UsedPercent
				break
			}
		}
	}

	return result, nil
}

// GetSystemInfo 获取系统信息
func (s *SystemService) GetSystemInfo(ctx context.Context) (*models.SystemInfo, error) {
	// 构建系统信息
	systemInfo := &models.SystemInfo{
		// 系统版本信息
		SystemName:  "太上老君管理系统",
		Version:     "v1.0.0",
		BuildTime:   "2025-10-20 10:00:00", // 这里可以在编译时注入
		GitCommit:   "abc123def456",        // 这里可以在编译时注入
		Environment: "生产环境",                // 可以从配置文件读取
		// 运行环境信息
		OS:           runtime.GOOS + " " + runtime.GOARCH,
		Architecture: runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NodeVersion:  "v18.17.0",        // 可以通过执行命令获取
		Database:     "PostgreSQL 14.9", // 可以从数据库查询获取
		Redis:        "6.2.6",           // 可以从Redis查询获取
		Nginx:        "1.20.1",          // 可以通过执行命令获取

		// 许可证信息
		License: models.LicenseInfo{
			Type:        "commercial",
			Description: "本系统采用商业许可证，仅限授权用户使用",
			ExpiryDate:  "2025-01-20",
			IsValid:     true,
		},
	}

	return systemInfo, nil
}
