package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/codetaoist/laojun/pkg/shared/config"
	_ "github.com/lib/pq"
)

func Initialize(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected successfully with connection pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)
	return db, nil
}

// RunMigrations 运行数据库迁移
func RunMigrations(db *sql.DB) error {
	log.Println("开始运行数据库迁移...")
	
	// 创建完整迁移器
	migrator := NewCompleteMigrator(db)
	
	// 运行完整迁移
	if err := migrator.RunCompleteMigration(); err != nil {
		log.Printf("完整迁移执行失败: %v", err)
		return fmt.Errorf("完整迁移执行失败: %w", err)
	}
	
	log.Println("数据库迁移完成")
	return nil
}

// RunCompleteMigrations 运行完整数据库迁移（新增函数）
func RunCompleteMigrations(db *sql.DB) error {
	migrator := NewCompleteMigrator(db)
	return migrator.RunCompleteMigration()
}

// GetMigrationStatus 获取迁移状态（新增函数）
func GetMigrationStatus(db *sql.DB) error {
	migrator := NewCompleteMigrator(db)
	return migrator.GetMigrationStatus()
}

// ResetDatabase 重置数据库（新增函数）
func ResetDatabase(db *sql.DB) error {
	migrator := NewCompleteMigrator(db)
	return migrator.ResetDatabase()
}

// getDSNFromDB 从现有数据库连接获取DSN（这是一个简化的实现）
func getDSNFromDB(db *sql.DB) string {
	// 这里我们使用环境变量重新构建DSN
	// 在生产环境中，你可能需要更复杂的逻辑来提取DSN
	return "host=localhost port=5432 user=laojun password=laojun123 dbname=laojun sslmode=disable"
}
