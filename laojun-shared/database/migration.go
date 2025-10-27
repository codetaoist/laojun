package database

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Migration 迁移结构
type Migration struct {
	ID            uint   `gorm:"primaryKey"`
	Version       string `gorm:"uniqueIndex;not null"`
	Name          string `gorm:"not null"`
	UpSQL         string `gorm:"type:text"`
	DownSQL       string `gorm:"type:text"`
	ExecutedAt    time.Time
	ExecutionTime int64 // 执行时间（毫秒）
}

// MigrationManager 迁移管理工具
type MigrationManager struct {
	db            *gorm.DB
	migrationsDir string
}

// NewMigrationManager 创建迁移管理工具
func NewMigrationManager(db *gorm.DB, migrationsDir string) *MigrationManager {
	return &MigrationManager{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// Initialize 初始化迁移表
func (mm *MigrationManager) Initialize() error {
	return mm.db.AutoMigrate(&Migration{})
}

// Migrate 执行迁移
func (mm *MigrationManager) Migrate() error {
	// 获取所有迁移文件
	files, err := mm.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// 获取已执行的迁移
	executed, err := mm.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// 执行未执行的迁移
	for _, file := range files {
		version := mm.extractVersion(file)
		if _, exists := executed[version]; !exists {
			if err := mm.executeMigration(file); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}
		}
	}

	return nil
}

// Rollback 回滚迁移
func (mm *MigrationManager) Rollback(steps int) error {
	// 获取已执行的迁移（按执行时间倒序）
	var migrations []Migration
	if err := mm.db.Order("executed_at DESC").Limit(steps).Find(&migrations).Error; err != nil {
		return fmt.Errorf("failed to get migrations for rollback: %w", err)
	}

	// 执行回滚
	for _, migration := range migrations {
		if err := mm.executeRollback(migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// RollbackTo 回滚到指定版本
func (mm *MigrationManager) RollbackTo(version string) error {
	// 获取需要回滚的迁移
	var migrations []Migration
	if err := mm.db.Where("version > ?", version).Order("executed_at DESC").Find(&migrations).Error; err != nil {
		return fmt.Errorf("failed to get migrations for rollback: %w", err)
	}

	// 执行回滚
	for _, migration := range migrations {
		if err := mm.executeRollback(migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// Status 获取迁移状态
func (mm *MigrationManager) Status() ([]MigrationStatus, error) {
	// 获取所有迁移文件
	files, err := mm.getMigrationFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}

	// 获取已执行的迁移
	executed, err := mm.getExecutedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	var status []MigrationStatus
	for _, file := range files {
		version := mm.extractVersion(file)
		migration, exists := executed[version]

		s := MigrationStatus{
			Version:  version,
			Name:     mm.extractName(file),
			Executed: exists,
		}

		if exists {
			s.ExecutedAt = migration.ExecutedAt
			s.ExecutionTime = migration.ExecutionTime
		}

		status = append(status, s)
	}

	return status, nil
}

// CreateMigration 创建新的迁移文件
func (mm *MigrationManager) CreateMigration(name string) (string, error) {
	timestamp := time.Now().Format("20060102150405")
	version := fmt.Sprintf("%s_%s", timestamp, strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	filename := fmt.Sprintf("%s.sql", version)
	filepath := filepath.Join(mm.migrationsDir, filename)

	template := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied


-- +migrate Down
-- SQL in section 'Down' is executed when this migration is rolled back

`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := ioutil.WriteFile(filepath, []byte(template), 0644); err != nil {
		return "", fmt.Errorf("failed to create migration file: %w", err)
	}

	return filepath, nil
}

// ValidateMigrations 验证迁移文件
func (mm *MigrationManager) ValidateMigrations() error {
	files, err := mm.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	for _, file := range files {
		if err := mm.validateMigrationFile(file); err != nil {
			return fmt.Errorf("invalid migration file %s: %w", file, err)
		}
	}

	return nil
}

// RepairMigrations 修复迁移状态
func (mm *MigrationManager) RepairMigrations() error {
	// 获取所有迁移文件
	files, err := mm.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// 获取已执行的迁移
	executed, err := mm.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// 检查并修复不一致的状态
	for _, file := range files {
		version := mm.extractVersion(file)
		if _, exists := executed[version]; !exists {
			// 检查数据库中是否存在该迁移创建的表/字段
			if mm.isMigrationApplied(file) {
				// 标记为已执行
				migration := Migration{
					Version:       version,
					Name:          mm.extractName(file),
					ExecutedAt:    time.Now(),
					ExecutionTime: 0,
				}
				if err := mm.db.Create(&migration).Error; err != nil {
					return fmt.Errorf("failed to repair migration %s: %w", version, err)
				}
			}
		}
	}

	return nil
}

// 私有方法

// getMigrationFiles 获取迁移文件列表
func (mm *MigrationManager) getMigrationFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(mm.migrationsDir, "*.sql"))
	if err != nil {
		return nil, err
	}

	// 按文件名排序
	sort.Strings(files)
	return files, nil
}

// getExecutedMigrations 获取已执行的迁移
func (mm *MigrationManager) getExecutedMigrations() (map[string]Migration, error) {
	var migrations []Migration
	if err := mm.db.Find(&migrations).Error; err != nil {
		return nil, err
	}

	executed := make(map[string]Migration)
	for _, migration := range migrations {
		executed[migration.Version] = migration
	}

	return executed, nil
}

// executeMigration 执行迁移
func (mm *MigrationManager) executeMigration(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	upSQL, downSQL, err := mm.parseMigrationContent(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse migration content: %w", err)
	}

	start := time.Now()

	// 在事务中执行迁移
	err = mm.db.Transaction(func(tx *gorm.DB) error {
		// 执行 UP SQL
		if upSQL != "" {
			if err := tx.Exec(upSQL).Error; err != nil {
				return fmt.Errorf("failed to execute up SQL: %w", err)
			}
		}

		// 记录迁移
		migration := Migration{
			Version:       mm.extractVersion(file),
			Name:          mm.extractName(file),
			UpSQL:         upSQL,
			DownSQL:       downSQL,
			ExecutedAt:    time.Now(),
			ExecutionTime: time.Since(start).Milliseconds(),
		}

		return tx.Create(&migration).Error
	})

	if err != nil {
		return fmt.Errorf("migration transaction failed: %w", err)
	}

	return nil
}

// executeRollback 执行回滚
func (mm *MigrationManager) executeRollback(migration Migration) error {
	if migration.DownSQL == "" {
		return fmt.Errorf("no down SQL available for migration %s", migration.Version)
	}

	// 在事务中执行回滚
	return mm.db.Transaction(func(tx *gorm.DB) error {
		// 执行 DOWN SQL
		if err := tx.Exec(migration.DownSQL).Error; err != nil {
			return fmt.Errorf("failed to execute down SQL: %w", err)
		}

		// 删除迁移记录
		return tx.Delete(&migration).Error
	})
}

// parseMigrationContent 解析迁移文件内容
func (mm *MigrationManager) parseMigrationContent(content string) (upSQL, downSQL string, err error) {
	lines := strings.Split(content, "\n")
	var currentSection string
	var upLines, downLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "-- +migrate Up") {
			currentSection = "up"
			continue
		} else if strings.HasPrefix(line, "-- +migrate Down") {
			currentSection = "down"
			continue
		}

		// 跳过注释行和空行
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		switch currentSection {
		case "up":
			upLines = append(upLines, line)
		case "down":
			downLines = append(downLines, line)
		}
	}

	upSQL = strings.Join(upLines, "\n")
	downSQL = strings.Join(downLines, "\n")

	return upSQL, downSQL, nil
}

// extractVersion 从文件名提取版本号
func (mm *MigrationManager) extractVersion(file string) string {
	filename := filepath.Base(file)
	return strings.TrimSuffix(filename, ".sql")
}

// extractName 从文件名提取迁移名称
func (mm *MigrationManager) extractName(file string) string {
	version := mm.extractVersion(file)
	parts := strings.SplitN(version, "_", 2)
	if len(parts) > 1 {
		return strings.ReplaceAll(parts[1], "_", " ")
	}
	return version
}

// validateMigrationFile 验证迁移文件
func (mm *MigrationManager) validateMigrationFile(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// 检查是否包含必要的标记
	if !strings.Contains(contentStr, "-- +migrate Up") {
		return fmt.Errorf("missing '-- +migrate Up' marker")
	}

	if !strings.Contains(contentStr, "-- +migrate Down") {
		return fmt.Errorf("missing '-- +migrate Down' marker")
	}

	// 检查版本号格式
	version := mm.extractVersion(file)
	if len(version) < 14 {
		return fmt.Errorf("invalid version format: %s", version)
	}

	// 检查时间戳部分
	timestampPart := version[:14]
	if _, err := strconv.ParseInt(timestampPart, 10, 64); err != nil {
		return fmt.Errorf("invalid timestamp in version: %s", timestampPart)
	}

	return nil
}

// isMigrationApplied 检查迁移是否已应用（通过检查数据库结构）
func (mm *MigrationManager) isMigrationApplied(file string) bool {
	// 这里可以实现更复杂的逻辑来检查迁移是否已应用
	// 例如检查特定的表或字段是否存在
	// 目前简单返回false
	return false
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Version       string    `json:"version"`
	Name          string    `json:"name"`
	Executed      bool      `json:"executed"`
	ExecutedAt    time.Time `json:"executed_at,omitempty"`
	ExecutionTime int64     `json:"execution_time,omitempty"`
}

// MigrationInfo 迁移信息
type MigrationInfo struct {
	TotalMigrations    int `json:"total_migrations"`
	ExecutedMigrations int `json:"executed_migrations"`
	PendingMigrations  int `json:"pending_migrations"`
}

// GetMigrationInfo 获取迁移信息
func (mm *MigrationManager) GetMigrationInfo() (*MigrationInfo, error) {
	status, err := mm.Status()
	if err != nil {
		return nil, err
	}

	info := &MigrationInfo{
		TotalMigrations: len(status),
	}

	for _, s := range status {
		if s.Executed {
			info.ExecutedMigrations++
		} else {
			info.PendingMigrations++
		}
	}

	return info, nil
}
