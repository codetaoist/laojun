package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codetaoist/laojun-shared/config"
	_ "github.com/lib/pq"
)

// DB 数据库连接包装器
type DB struct {
	*sql.DB
	config *config.DatabaseConfig
}

// NewDB 创建新的数据库连接包装器
func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		DB:     db,
		config: cfg,
	}, nil
}

// Close 关闭数据库连接包装器
func (db *DB) Close() error {
	return db.DB.Close()
}

// Health 检查数据库健康状态
func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// GetStats 获取数据库连接统计信息
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}

// Transaction 执行事务
func (db *DB) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}
