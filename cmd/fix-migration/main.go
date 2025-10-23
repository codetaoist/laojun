package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 构建数据库连接字符串
	dbURL := "postgres://laojun:laojun123@localhost:5432/laojun?sslmode=disable"
	
	// 连接数据库
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// 修复迁移状态
	_, err = db.Exec("UPDATE schema_migrations SET dirty = false WHERE version = 8")
	if err != nil {
		log.Fatalf("修复迁移状态失败: %v", err)
	}

	fmt.Println("数据库迁移状态修复成功")
}