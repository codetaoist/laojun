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
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "laojun_admin")

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// 检查是否已经存在唯一约束
	checkConstraintQuery := `
		SELECT constraint_name 
		FROM information_schema.table_constraints 
		WHERE table_name = 'lj_menus' 
		AND constraint_type = 'UNIQUE'
		AND constraint_name LIKE '%title%path%component%'
	`

	var constraintName string
	err = db.QueryRow(checkConstraintQuery).Scan(&constraintName)
	if err == nil {
		fmt.Printf("Unique constraint already exists: %s\n", constraintName)
		return
	} else if err != sql.ErrNoRows {
		log.Fatalf("Failed to check existing constraints: %v", err)
	}

	// 添加唯一约束
	fmt.Println("Adding unique constraint to lj_menus table...")

	addConstraintQuery := `
		ALTER TABLE lj_menus 
		ADD CONSTRAINT uk_lj_menus_title_path_component 
		UNIQUE (title, path, component)
	`

	_, err = db.Exec(addConstraintQuery)
	if err != nil {
		log.Fatalf("Failed to add unique constraint: %v", err)
	}

	fmt.Println("Successfully added unique constraint to lj_menus table!")

	// 验证约束是否添加成功
	err = db.QueryRow(checkConstraintQuery).Scan(&constraintName)
	if err != nil {
		log.Printf("Warning: Could not verify constraint creation: %v", err)
	} else {
		fmt.Printf("Verified: Unique constraint created successfully: %s\n", constraintName)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
