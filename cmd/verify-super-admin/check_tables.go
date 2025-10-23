package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=laojun password=laojun123 dbname=laojun sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}
	fmt.Println("数据库连接成功!")

	// 查询所有表
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("查询表失败:", err)
	}
	defer rows.Close()

	fmt.Println("\n数据库中的所有表:")
	fmt.Println(strings.Repeat("-", 40))
	
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		fmt.Println(tableName)
	}
}