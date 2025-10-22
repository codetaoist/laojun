package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量获取数据库连接信息
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "laojun"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "laojun123"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "laojun"
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("数据库连接成功")

	// 检查mp_plugins表是否存在
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'mp_plugins')").Scan(&exists)
	if err != nil {
		log.Fatal("检查表存在性失败:", err)
	}

	if !exists {
		fmt.Println("mp_plugins表不存在")
		
		// 检查plugins表是否存在
		err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'plugins')").Scan(&exists)
		if err != nil {
			log.Fatal("检查plugins表存在性失败:", err)
		}
		
		if exists {
			fmt.Println("plugins表存在，需要重命名为mp_plugins")
		} else {
			fmt.Println("plugins表也不存在，需要创建表")
		}
		return
	}

	fmt.Println("mp_plugins表存在，检查字段结构...")

	// 查询表结构
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'mp_plugins' 
		ORDER BY ordinal_position`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("查询表结构失败:", err)
	}
	defer rows.Close()

	fmt.Println("\nmp_plugins表字段结构:")
	fmt.Println("字段名\t\t数据类型\t\t可空\t\t默认值")
	fmt.Println("------------------------------------------------------------")

	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			log.Fatal("扫描行失败:", err)
		}

		defaultVal := "NULL"
		if columnDefault.Valid {
			defaultVal = columnDefault.String
		}

		fmt.Printf("%-20s %-20s %-10s %s\n", columnName, dataType, isNullable, defaultVal)
	}

	// 检查关键字段
	fmt.Println("\n检查关键字段:")
	
	// 检查category_id字段
	var categoryIdExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'mp_plugins' AND column_name = 'category_id')").Scan(&categoryIdExists)
	if err != nil {
		log.Fatal("检查category_id字段失败:", err)
	}
	
	if categoryIdExists {
		fmt.Println("✓ category_id字段存在")
	} else {
		fmt.Println("✗ category_id字段不存在")
	}

	// 检查developer_id字段
	var developerIdExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'mp_plugins' AND column_name = 'developer_id')").Scan(&developerIdExists)
	if err != nil {
		log.Fatal("检查developer_id字段失败:", err)
	}
	
	if developerIdExists {
		fmt.Println("✓ developer_id字段存在")
	} else {
		fmt.Println("✗ developer_id字段不存在")
	}

	// 检查short_description字段
	var shortDescExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'mp_plugins' AND column_name = 'short_description')").Scan(&shortDescExists)
	if err != nil {
		log.Fatal("检查short_description字段失败:", err)
	}
	
	if shortDescExists {
		fmt.Println("✓ short_description字段存在")
	} else {
		fmt.Println("✗ short_description字段不存在")
	}

	// 检查mp_categories表
	var categoriesExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'mp_categories')").Scan(&categoriesExists)
	if err != nil {
		log.Fatal("检查mp_categories表存在性失败:", err)
	}
	
	if categoriesExists {
		fmt.Println("✓ mp_categories表存在")
	} else {
		fmt.Println("✗ mp_categories表不存在")
	}
}