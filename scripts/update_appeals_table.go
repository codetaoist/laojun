package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接信息
	connStr := "host=localhost port=5432 user=laojun password=laojun123 dbname=laojun sslmode=disable"
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("Updating mp_developer_appeals table structure...")

	// 添加缺失的字段
	alterStatements := []string{
		"ALTER TABLE mp_developer_appeals ADD COLUMN IF NOT EXISTS review_id UUID REFERENCES mp_plugin_reviews(id)",
		"ALTER TABLE mp_developer_appeals ADD COLUMN IF NOT EXISTS appeal_type VARCHAR(50) DEFAULT 'review_rejection'",
		"ALTER TABLE mp_developer_appeals ADD COLUMN IF NOT EXISTS admin_id UUID REFERENCES ua_admin(id)",
		"ALTER TABLE mp_developer_appeals ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		
		// 重命名字段以匹配模型
		"ALTER TABLE mp_developer_appeals RENAME COLUMN admin_response TO admin_reply",
		"ALTER TABLE mp_developer_appeals RENAME COLUMN resolved_at TO processed_at",
	}

	for i, stmt := range alterStatements {
		fmt.Printf("Executing statement %d: %s\n", i+1, stmt)
		_, err = db.Exec(stmt)
		if err != nil {
			// 某些语句可能会失败（如字段已存在或重命名失败），这是正常的
			fmt.Printf("Warning: %v\n", err)
		} else {
			fmt.Printf("Success!\n")
		}
	}

	// 检查最终的表结构
	fmt.Println("\nFinal table structure:")
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'mp_developer_appeals' 
		AND table_schema = 'public'
		ORDER BY ordinal_position;
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query table structure:", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-12s %s\n", "Column Name", "Data Type", "Nullable", "Default")
	fmt.Println("--------------------------------------------------------------------------------")
	
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString
		
		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault); err != nil {
			log.Fatal("Failed to scan column info:", err)
		}
		
		defaultValue := "NULL"
		if columnDefault.Valid {
			defaultValue = columnDefault.String
		}
		
		fmt.Printf("%-20s %-20s %-12s %s\n", columnName, dataType, isNullable, defaultValue)
	}

	fmt.Println("\nTable update completed!")
}