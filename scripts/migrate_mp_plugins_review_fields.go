package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// 加载环境变量
	err := godotenv.Load(".env.local")
	if err != nil {
		err = godotenv.Load(".env.development")
		if err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// 构建数据库连接字符串
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "laojun")
	dbPassword := getEnv("DB_PASSWORD", "laojun123")
	dbName := getEnv("DB_NAME", "laojun")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 测试连接
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== 开始执行 mp_plugins 表审核字段迁移 ===")

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}
	defer tx.Rollback()

	// 执行迁移SQL
	migrationSQL := []string{
		// 添加审核状态字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS review_status VARCHAR(50) DEFAULT 'pending'",
		
		// 添加审核优先级字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS review_priority VARCHAR(20) DEFAULT 'normal'",
		
		// 添加自动审核评分字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS auto_review_score DECIMAL(3,2)",
		
		// 添加自动审核结果字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS auto_review_result VARCHAR(50)",
		
		// 添加审核备注字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS review_notes TEXT",
		
		// 添加审核完成时间字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP WITH TIME ZONE",
		
		// 添加审核员ID字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS reviewer_id UUID",
		
		// 添加提交审核时间字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS submitted_for_review_at TIMESTAMP WITH TIME ZONE",
		
		// 添加拒绝原因字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS rejection_reason TEXT",
		
		// 添加申诉次数字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS appeal_count INTEGER DEFAULT 0",
		
		// 添加最后申诉时间字段
		"ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS last_appeal_at TIMESTAMP WITH TIME ZONE",
	}

	// 执行字段添加
	for i, sql := range migrationSQL {
		fmt.Printf("执行步骤 %d: %s\n", i+1, sql)
		_, err = tx.Exec(sql)
		if err != nil {
			log.Printf("Failed to execute SQL: %s, Error: %v", sql, err)
			return
		}
		fmt.Printf("✅ 步骤 %d 完成\n", i+1)
	}

	// 添加约束 (先检查是否存在，避免重复添加)
	constraintSQL := []string{
		`DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_review_status') THEN
				ALTER TABLE mp_plugins ADD CONSTRAINT chk_review_status 
				CHECK (review_status IN ('pending', 'in_review', 'approved', 'rejected', 'appealing', 'suspended'));
			END IF;
		END $$`,
		
		`DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_review_priority') THEN
				ALTER TABLE mp_plugins ADD CONSTRAINT chk_review_priority 
				CHECK (review_priority IN ('low', 'normal', 'high', 'urgent'));
			END IF;
		END $$`,
		
		`DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_auto_review_result') THEN
				ALTER TABLE mp_plugins ADD CONSTRAINT chk_auto_review_result 
				CHECK (auto_review_result IN ('pass', 'fail', 'manual_review', 'pending'));
			END IF;
		END $$`,
	}

	fmt.Println("\n=== 添加约束 ===")
	for i, sql := range constraintSQL {
		fmt.Printf("执行约束 %d\n", i+1)
		_, err = tx.Exec(sql)
		if err != nil {
			log.Printf("Failed to add constraint: %s, Error: %v", sql, err)
			// 约束失败不是致命错误，继续执行
		} else {
			fmt.Printf("✅ 约束 %d 添加成功\n", i+1)
		}
	}

	// 创建索引
	indexSQL := []string{
		"CREATE INDEX IF NOT EXISTS idx_mp_plugins_review_status ON mp_plugins(review_status)",
		"CREATE INDEX IF NOT EXISTS idx_mp_plugins_review_priority ON mp_plugins(review_priority)",
		"CREATE INDEX IF NOT EXISTS idx_mp_plugins_reviewer_id ON mp_plugins(reviewer_id)",
		"CREATE INDEX IF NOT EXISTS idx_mp_plugins_submitted_for_review_at ON mp_plugins(submitted_for_review_at)",
		"CREATE INDEX IF NOT EXISTS idx_mp_plugins_reviewed_at ON mp_plugins(reviewed_at)",
	}

	fmt.Println("\n=== 创建索引 ===")
	for i, sql := range indexSQL {
		fmt.Printf("创建索引 %d\n", i+1)
		_, err = tx.Exec(sql)
		if err != nil {
			log.Printf("Failed to create index: %s, Error: %v", sql, err)
		} else {
			fmt.Printf("✅ 索引 %d 创建成功\n", i+1)
		}
	}

	// 更新现有数据
	updateSQL := `
		UPDATE mp_plugins SET 
			review_status = CASE 
				WHEN status = 'approved' THEN 'approved'
				WHEN status = 'rejected' THEN 'rejected'
				WHEN status = 'suspended' THEN 'suspended'
				WHEN status = 'pending' THEN 'pending'
				ELSE 'pending'
			END,
			submitted_for_review_at = CASE 
				WHEN status IN ('pending', 'approved', 'rejected') THEN created_at
				ELSE NULL
			END,
			reviewed_at = CASE 
				WHEN status IN ('approved', 'rejected', 'suspended') THEN updated_at
				ELSE NULL
			END
		WHERE review_status = 'pending' OR review_status IS NULL`

	fmt.Println("\n=== 更新现有数据 ===")
	result, err := tx.Exec(updateSQL)
	if err != nil {
		log.Printf("Failed to update existing data: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✅ 更新了 %d 条记录\n", rowsAffected)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("\n=== 迁移完成，验证结果 ===")

	// 验证迁移结果
	var totalPlugins, pendingReview, approved, rejected int
	err = db.QueryRow(`
		SELECT 
			COUNT(*) as total_plugins,
			COUNT(CASE WHEN review_status = 'pending' THEN 1 END) as pending_review,
			COUNT(CASE WHEN review_status = 'approved' THEN 1 END) as approved,
			COUNT(CASE WHEN review_status = 'rejected' THEN 1 END) as rejected
		FROM mp_plugins`).Scan(&totalPlugins, &pendingReview, &approved, &rejected)

	if err != nil {
		log.Printf("Failed to verify migration: %v", err)
	} else {
		fmt.Printf("总插件数: %d\n", totalPlugins)
		fmt.Printf("待审核: %d\n", pendingReview)
		fmt.Printf("已批准: %d\n", approved)
		fmt.Printf("已拒绝: %d\n", rejected)
	}

	fmt.Println("\n🎉 mp_plugins 表审核字段迁移成功完成！")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}