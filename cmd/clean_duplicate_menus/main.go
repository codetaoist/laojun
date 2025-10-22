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

	// 查询重复的菜单记录
	duplicateQuery := `
		SELECT title, path, icon, component, sort_order, COUNT(*) as count
		FROM lj_menus 
		GROUP BY title, path, icon, component, sort_order 
		HAVING COUNT(*) > 1
		ORDER BY title, sort_order
	`

	rows, err := db.Query(duplicateQuery)
	if err != nil {
		log.Fatalf("Failed to query duplicate menus: %v", err)
	}
	defer rows.Close()

	fmt.Println("Found duplicate menu records:")
	fmt.Println("Title\t\tPath\t\tIcon\t\tComponent\t\tSort Order\tCount")
	fmt.Println("-------------------------------------------------------------------")

	hasDuplicates := false
	for rows.Next() {
		var title, path, icon, component string
		var sortOrder, count int

		err := rows.Scan(&title, &path, &icon, &component, &sortOrder, &count)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		hasDuplicates = true
		fmt.Printf("%s\t\t%s\t\t%s\t\t%s\t\t%d\t\t%d\n",
			title, path, icon, component, sortOrder, count)
	}

	if !hasDuplicates {
		fmt.Println("No duplicate menu records found.")
		return
	}

	// 询问用户是否要清理重复记录
	fmt.Print("\nDo you want to clean up duplicate records? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("Cleanup cancelled.")
		return
	}

	// 开始清理重复记录
	fmt.Println("\nStarting cleanup process...")

	// 删除重复记录，保留最早创建的记录
	cleanupQuery := `
		DELETE FROM lj_menus 
		WHERE id NOT IN (
			SELECT DISTINCT ON (title, path, icon, component, sort_order) id
			FROM lj_menus 
			ORDER BY title, path, icon, component, sort_order, created_at ASC
		)
	`

	result, err := db.Exec(cleanupQuery)
	if err != nil {
		log.Fatalf("Failed to cleanup duplicate menus: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Warning: Could not get rows affected count: %v", err)
	} else {
		fmt.Printf("Successfully deleted %d duplicate menu records.\n", rowsAffected)
	}

	// 验证清理结果
	var remainingCount int
	countQuery := "SELECT COUNT(*) FROM lj_menus"
	err = db.QueryRow(countQuery).Scan(&remainingCount)
	if err != nil {
		log.Printf("Warning: Could not get remaining menu count: %v", err)
	} else {
		fmt.Printf("Remaining menu records: %d\n", remainingCount)
	}

	fmt.Println("Cleanup completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
