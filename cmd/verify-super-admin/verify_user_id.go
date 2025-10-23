package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接配置
	connStr := "host=localhost port=5432 user=laojun password=laojun123 dbname=laojun sslmode=disable"
	
	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}
	fmt.Println("数据库连接成功!")

	// 读取SQL文件
	sqlContent, err := ioutil.ReadFile("../../verify_user_id.sql")
	if err != nil {
		log.Fatal("读取SQL文件失败:", err)
	}

	fmt.Printf("SQL文件读取成功，内容长度: %d\n", len(sqlContent))

	// 分割SQL语句
	queries := strings.Split(string(sqlContent), ";")
	
	fmt.Printf("分割后的查询数量: %d\n", len(queries))

	for i, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" || !strings.Contains(strings.ToUpper(query), "SELECT") {
			fmt.Printf("跳过查询 %d (空或非SELECT)\n", i)
			continue
		}

		fmt.Printf("\n执行查询 %d:\n%s\n", i, query)

		rows, err := db.Query(query)
		if err != nil {
			fmt.Printf("查询 %d 执行失败: %v\n", i, err)
			continue
		}

		// 获取列名
		columns, err := rows.Columns()
		if err != nil {
			fmt.Printf("获取列名失败: %v\n", err)
			rows.Close()
			continue
		}

		fmt.Printf("列名: %v\n", columns)

		// 读取数据
		for rows.Next() {
			// 创建接收数据的切片
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			// 扫描行数据
			if err := rows.Scan(valuePtrs...); err != nil {
				fmt.Printf("扫描行数据失败: %v\n", err)
				continue
			}

			// 打印数据
			for i, col := range columns {
				var v interface{}
				val := values[i]
				if b, ok := val.([]byte); ok {
					v = string(b)
				} else {
					v = val
				}
				fmt.Printf("%s: %v\n", col, v)
			}
			fmt.Println("---")
		}

		rows.Close()
	}
}