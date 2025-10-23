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

	// 读取SQL文件
	sqlContent, err := ioutil.ReadFile("../../verify_super_admin.sql")
	if err != nil {
		log.Fatal("读取SQL文件失败:", err)
	}
	fmt.Printf("SQL文件读取成功，内容长度: %d\n", len(sqlContent))

	// 分割SQL语句
	queries := strings.Split(string(sqlContent), ";")
	fmt.Printf("分割后的查询数量: %d\n", len(queries))
	
	for i, query := range queries {
		query = strings.TrimSpace(query)
		fmt.Printf("查询 %d: [%s]\n", i, query)
		if query == "" {
			fmt.Printf("跳过查询 %d (空)\n", i)
			continue
		}
		
		// 检查是否包含SELECT语句
		if !strings.Contains(strings.ToUpper(query), "SELECT") {
			fmt.Printf("跳过查询 %d (不包含SELECT)\n", i)
			continue
		}

		fmt.Printf("\n执行查询:\n%s\n", query)
		fmt.Println(strings.Repeat("-", 80))

		rows, err := db.Query(query)
		if err != nil {
			log.Printf("查询执行失败: %v", err)
			continue
		}

		// 获取列名
		columns, err := rows.Columns()
		if err != nil {
			log.Printf("获取列名失败: %v", err)
			rows.Close()
			continue
		}

		// 打印列名
		for i, col := range columns {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print(col)
		}
		fmt.Println()

		// 创建接收数据的slice
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// 读取数据
		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				log.Printf("扫描行失败: %v", err)
				continue
			}

			for i, val := range values {
				if i > 0 {
					fmt.Print("\t")
				}
				if val != nil {
					fmt.Print(val)
				} else {
					fmt.Print("NULL")
				}
			}
			fmt.Println()
		}

		rows.Close()
		fmt.Println()
	}
}