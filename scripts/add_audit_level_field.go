package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "laojun")
	dbPassword := getEnv("DB_PASSWORD", "laojun123")
	dbName := getEnv("DB_NAME", "laojun")

	// 连接数据库
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("成功连接到数据库")

	// 检查 level 字段是否已存在
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'sys_audit_logs' AND column_name = 'level')").Scan(&exists)
	if err != nil {
		log.Fatal("Failed to check level field:", err)
	}

	if exists {
		fmt.Println("level 字段已存在，无需添加")
	} else {
		fmt.Println("添加 level 字段到 sys_audit_logs 表...")
		
		// 添加 level 字段
		_, err = db.Exec("ALTER TABLE sys_audit_logs ADD COLUMN level VARCHAR(20) DEFAULT 'info'")
		if err != nil {
			log.Fatal("Failed to add level field:", err)
		}
		
		fmt.Println("✅ 成功添加 level 字段")
	}

	// 检查其他可能缺失的字段
	requiredFields := map[string]string{
		"target_id":   "UUID",
		"target_type": "VARCHAR(50)",
		"description": "TEXT",
		"old_data":    "TEXT",
		"new_data":    "TEXT",
	}

	fmt.Println("\n=== 检查其他必需字段 ===")
	for field, dataType := range requiredFields {
		var fieldExists bool
		err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'sys_audit_logs' AND column_name = $1)", field).Scan(&fieldExists)
		if err != nil {
			log.Printf("Failed to check field %s: %v", field, err)
			continue
		}

		if fieldExists {
			fmt.Printf("✅ %s 字段存在\n", field)
		} else {
			fmt.Printf("❌ %s 字段不存在，正在添加...\n", field)
			
			alterSQL := fmt.Sprintf("ALTER TABLE sys_audit_logs ADD COLUMN %s %s", field, dataType)
			_, err = db.Exec(alterSQL)
			if err != nil {
				log.Printf("Failed to add field %s: %v", field, err)
			} else {
				fmt.Printf("✅ 成功添加 %s 字段\n", field)
			}
		}
	}

	// 显示更新后的表结构
	fmt.Println("\n=== sys_audit_logs 表结构 ===")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable, column_default 
		FROM information_schema.columns 
		WHERE table_name = 'sys_audit_logs' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("Failed to query table structure:", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-10s %s\n", "字段名", "数据类型", "可为空", "默认值")
	fmt.Println("------------------------------------------------------------")

	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		defaultValue := "NULL"
		if columnDefault.Valid {
			defaultValue = columnDefault.String
		}

		fmt.Printf("%-20s %-20s %-10s %s\n", columnName, dataType, isNullable, defaultValue)
	}

	fmt.Println("\n✅ sys_audit_logs 表结构更新完成")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}