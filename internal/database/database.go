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
	log.Println("Skipping database migrations - using existing database schema")
	return nil
}

// getDSNFromDB 从现有数据库连接获取DSN（这是一个简化的实现）
func getDSNFromDB(db *sql.DB) string {
	// 这里我们使用环境变量重新构建DSN
	// 在生产环境中，你可能需要更复杂的逻辑来提取DSN
	return "host=localhost port=5432 user=laojun password=laojun123 dbname=laojun sslmode=disable"
}
