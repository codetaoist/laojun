package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

// Migrator 数据库迁移器
type Migrator struct {
	db *sql.DB
}

// NewMigrator 创建新的迁移器
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// RunMigrations 执行数据库迁移
func (m *Migrator) RunMigrations() error {
	log.Println("开始执行数据库迁移...")

	// 创建迁移记录表
	if err := m.createMigrationTable(); err != nil {
		return fmt.Errorf("创建迁移记录表失败: %v", err)
	}

	// 检查是否已经执行过迁移
	if migrated, err := m.isMigrated(); err != nil {
		return fmt.Errorf("检查迁移状态失败: %v", err)
	} else if migrated {
		// 进一步校验关键社区表是否存在
		exists, err := m.essentialTablesExist()
		if err != nil {
			return fmt.Errorf("检查关键表失败: %v", err)
		}
		if exists {
			log.Println("数据库迁移已经执行过，关键表存在，跳过迁移")
			return nil
		}
		log.Println("检测到关键社区表缺失，仍将执行迁移SQL以补全")
	}

	// 读取迁移文件
	migrationFile := filepath.Join("sql", "migration.sql")
	sqlContent, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %v", err)
	}

	// 执行迁移SQL
	if _, err := m.db.Exec(string(sqlContent)); err != nil {
		return fmt.Errorf("执行迁移SQL失败: %v", err)
	}

	// 记录迁移完成
	if err := m.markMigrated(); err != nil {
		return fmt.Errorf("记录迁移状态失败: %v", err)
	}

	log.Println("数据库迁移执行完成")
	return nil
}

// createMigrationTable 创建迁移记录表
func (m *Migrator) createMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(version)
		)
	`
	_, err := m.db.Exec(query)
	return err
}

// isMigrated 检查是否已经执行过迁移
func (m *Migrator) isMigrated() (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM schema_migrations WHERE version = 'marketplace_v1'"
	err := m.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// markMigrated 标记迁移已完成
func (m *Migrator) markMigrated() error {
	query := "INSERT INTO schema_migrations (version) VALUES ('marketplace_v1') ON CONFLICT (version) DO NOTHING"
	_, err := m.db.Exec(query)
	return err
}

// essentialTablesExist 关键社区表是否存在（若缺失则需要再次执行迁移）
func (m *Migrator) essentialTablesExist() (bool, error) {
	names := []string{
		"mp_users",
		"mp_forum_categories",
		"mp_forum_posts",
		"mp_forum_replies",
	}
	for _, name := range names {
		var ok bool
		if err := m.db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", name).Scan(&ok); err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
