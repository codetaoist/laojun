package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator 数据库迁移器
type Migrator struct {
	db          *sql.DB
	migrationsPath string
}

// NewMigrator 创建新的迁移器
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:             db,
		migrationsPath: "d:/taishanglaojun/db/migrations",
	}
}

// RunMigrations 执行数据库迁移
func (m *Migrator) RunMigrations() error {
	log.Println("开始执行数据库迁移...")

	// 创建 migrate 实例
	migrator, err := m.createMigrator()
	if err != nil {
		return fmt.Errorf("创建迁移器失败: %v", err)
	}
	// 注意：不调用 migrator.Close() 以避免关闭底层数据库连接

	// 获取当前版本
	currentVersion, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取当前迁移版本失败: %v", err)
	}

	if dirty {
		return fmt.Errorf("数据库处于脏状态，请手动修复")
	}

	// 执行迁移到最新版本
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行迁移失败: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Printf("数据库已是最新版本 (v%d)，无需迁移", currentVersion)
	} else {
		newVersion, _, _ := migrator.Version()
		log.Printf("数据库迁移完成，从版本 %d 升级到版本 %d", currentVersion, newVersion)
	}

	return nil
}

// createMigrator 创建 migrate 实例
func (m *Migrator) createMigrator() (*migrate.Migrate, error) {
	// 创建数据库驱动
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("创建数据库驱动失败: %v", err)
	}

	// 获取迁移文件路径
	migrationsPath, err := filepath.Abs(m.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("获取迁移文件路径失败: %v", err)
	}

	// 创建文件源，使用正确的 file:// URL 格式
	fileURL := "file://" + filepath.ToSlash(migrationsPath)
	fileSource, err := (&file.File{}).Open(fileURL)
	if err != nil {
		return nil, fmt.Errorf("打开迁移文件源失败: %v", err)
	}

	// 创建 migrate 实例
	migrator, err := migrate.NewWithInstance("file", fileSource, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("创建 migrate 实例失败: %v", err)
	}

	return migrator, nil
}

// Rollback 回滚到指定版本
func (m *Migrator) Rollback(version uint) error {
	log.Printf("开始回滚数据库到版本 %d...", version)

	migrator, err := m.createMigrator()
	if err != nil {
		return fmt.Errorf("创建迁移器失败: %v", err)
	}
	// 注意：不调用 migrator.Close() 以避免关闭底层数据库连接

	err = migrator.Migrate(version)
	if err != nil {
		return fmt.Errorf("回滚失败: %v", err)
	}

	log.Printf("数据库回滚到版本 %d 完成", version)
	return nil
}

// GetVersion 获取当前数据库版本
func (m *Migrator) GetVersion() (uint, bool, error) {
	migrator, err := m.createMigrator()
	if err != nil {
		return 0, false, fmt.Errorf("创建迁移器失败: %v", err)
	}
	// 注意：不调用 migrator.Close() 以避免关闭底层数据库连接

	version, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("获取版本失败: %v", err)
	}

	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}

	return version, dirty, nil
}

// Reset 重置数据库（删除所有表）
func (m *Migrator) Reset() error {
	log.Println("开始重置数据库...")

	migrator, err := m.createMigrator()
	if err != nil {
		return fmt.Errorf("创建迁移器失败: %v", err)
	}
	// 注意：不调用 migrator.Close() 以避免关闭底层数据库连接

	err = migrator.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("重置数据库失败: %v", err)
	}

	log.Println("数据库重置完成")
	return nil
}
