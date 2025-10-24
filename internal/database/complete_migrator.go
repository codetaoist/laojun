package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// CompleteMigrator 完整数据库迁移器
type CompleteMigrator struct {
	db             *sql.DB
	migrationsPath string
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	Version     string
	Name        string
	ExecutedAt  time.Time
	Checksum    string
	Success     bool
}

// NewCompleteMigrator 创建新的完整迁移器
func NewCompleteMigrator(db *sql.DB) *CompleteMigrator {
	mp := os.Getenv("COMPLETE_MIGRATIONS_DIR")
	if mp == "" {
		mp = "./db/migrations/final"
	}
	return &CompleteMigrator{
		db:             db,
		migrationsPath: mp,
	}
}

// InitializeMigrationTable 初始化迁移记录表
func (cm *CompleteMigrator) InitializeMigrationTable() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS complete_schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(64),
			success BOOLEAN DEFAULT true,
			execution_time_ms INTEGER DEFAULT 0
		);
		
		CREATE INDEX IF NOT EXISTS idx_complete_migrations_version ON complete_schema_migrations(version);
		CREATE INDEX IF NOT EXISTS idx_complete_migrations_executed_at ON complete_schema_migrations(executed_at);
	`
	
	_, err := cm.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %v", err)
	}
	
	log.Println("Migration table initialized successfully")
	return nil
}

// RunCompleteMigration 运行完整数据库迁移
func (cm *CompleteMigrator) RunCompleteMigration() error {
	log.Println("开始执行完整数据库迁移...")

	// 初始化迁移表
	if err := cm.InitializeMigrationTable(); err != nil {
		return err
	}

	// 获取迁移文件
	migrationFiles, err := cm.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("获取迁移文件失败: %v", err)
	}

	if len(migrationFiles) == 0 {
		log.Println("没有找到迁移文件")
		return nil
	}

	// 检查已执行的迁移
	executedMigrations, err := cm.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("获取已执行迁移失败: %v", err)
	}

	// 执行未执行的迁移
	for _, file := range migrationFiles {
		version := cm.extractVersionFromFile(file)
		if cm.isMigrationExecuted(version, executedMigrations) {
			log.Printf("迁移 %s 已执行，跳过", version)
			continue
		}

		log.Printf("执行迁移: %s", file)
		if err := cm.executeMigrationFile(file); err != nil {
			return fmt.Errorf("执行迁移文件 %s 失败: %v", file, err)
		}
	}

	log.Println("完整数据库迁移执行完成")
	return nil
}

// getMigrationFiles 获取迁移文件列表
func (cm *CompleteMigrator) getMigrationFiles() ([]string, error) {
	files, err := ioutil.ReadDir(cm.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录失败: %v", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		// 只处理 .up.sql 文件
		if strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, filepath.Join(cm.migrationsPath, file.Name()))
		}
	}

	// 按文件名排序
	sort.Strings(migrationFiles)
	return migrationFiles, nil
}

// getExecutedMigrations 获取已执行的迁移
func (cm *CompleteMigrator) getExecutedMigrations() (map[string]MigrationRecord, error) {
	query := `
		SELECT version, name, executed_at, checksum, success 
		FROM complete_schema_migrations 
		WHERE success = true
		ORDER BY executed_at
	`
	
	rows, err := cm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[string]MigrationRecord)
	for rows.Next() {
		var record MigrationRecord
		err := rows.Scan(&record.Version, &record.Name, &record.ExecutedAt, &record.Checksum, &record.Success)
		if err != nil {
			return nil, err
		}
		executed[record.Version] = record
	}

	return executed, nil
}

// isMigrationExecuted 检查迁移是否已执行
func (cm *CompleteMigrator) isMigrationExecuted(version string, executed map[string]MigrationRecord) bool {
	_, exists := executed[version]
	return exists
}

// extractVersionFromFile 从文件名提取版本号
func (cm *CompleteMigrator) extractVersionFromFile(filePath string) string {
	fileName := filepath.Base(filePath)
	// 提取版本号，例如从 "20251024024016_complete_schema.up.sql" 提取 "20251024024016"
	parts := strings.Split(fileName, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return fileName
}

// extractNameFromFile 从文件名提取迁移名称
func (cm *CompleteMigrator) extractNameFromFile(filePath string) string {
	fileName := filepath.Base(filePath)
	// 移除扩展名和方向
	name := strings.TrimSuffix(fileName, ".up.sql")
	name = strings.TrimSuffix(name, ".down.sql")
	
	// 移除版本号前缀
	parts := strings.Split(name, "_")
	if len(parts) > 1 {
		return strings.Join(parts[1:], "_")
	}
	return name
}

// executeMigrationFile 执行迁移文件
func (cm *CompleteMigrator) executeMigrationFile(filePath string) error {
	start := time.Now()
	
	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %v", err)
	}

	sqlContent := string(content)
	version := cm.extractVersionFromFile(filePath)
	name := cm.extractNameFromFile(filePath)

	// 在事务中执行迁移
	tx, err := cm.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 分割并执行SQL语句
	statements := cm.splitSQL(sqlContent)
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		_, err := tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("执行第 %d 条SQL语句失败: %v\nSQL: %s", i+1, err, stmt)
		}
	}

	// 记录迁移执行
	executionTime := time.Since(start).Milliseconds()
	checksum := cm.calculateChecksum(sqlContent)
	
	insertSQL := `
		INSERT INTO complete_schema_migrations (version, name, executed_at, checksum, success, execution_time_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err = tx.Exec(insertSQL, version, name, time.Now(), checksum, true, executionTime)
	if err != nil {
		return fmt.Errorf("记录迁移执行失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("迁移 %s 执行成功，耗时 %d ms", version, executionTime)
	return nil
}

// splitSQL 分割SQL语句
func (cm *CompleteMigrator) splitSQL(content string) []string {
	// 简单的SQL分割，按分号分割但忽略注释
	lines := strings.Split(content, "\n")
	var statements []string
	var currentStatement strings.Builder
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		
		currentStatement.WriteString(line)
		currentStatement.WriteString(" ")
		
		// 如果行以分号结尾，则认为是一个完整的语句
		if strings.HasSuffix(line, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			currentStatement.Reset()
		}
	}
	
	// 处理最后一个语句（如果没有以分号结尾）
	if currentStatement.Len() > 0 {
		stmt := strings.TrimSpace(currentStatement.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	
	return statements
}

// calculateChecksum 计算文件内容的校验和
func (cm *CompleteMigrator) calculateChecksum(content string) string {
	// 简单的校验和计算，实际应用中可以使用更复杂的算法
	hash := 0
	for _, char := range content {
		hash = hash*31 + int(char)
	}
	return fmt.Sprintf("%x", hash)
}

// GetMigrationStatus 获取迁移状态
func (cm *CompleteMigrator) GetMigrationStatus() error {
	query := `
		SELECT version, name, executed_at, execution_time_ms, success
		FROM complete_schema_migrations
		ORDER BY executed_at DESC
		LIMIT 10
	`
	
	rows, err := cm.db.Query(query)
	if err != nil {
		return fmt.Errorf("查询迁移状态失败: %v", err)
	}
	defer rows.Close()

	log.Println("最近的迁移记录:")
	log.Println("版本\t\t名称\t\t\t执行时间\t\t耗时(ms)\t状态")
	log.Println("--------------------------------------------------------------------")
	
	for rows.Next() {
		var version, name string
		var executedAt time.Time
		var executionTime int
		var success bool
		
		err := rows.Scan(&version, &name, &executedAt, &executionTime, &success)
		if err != nil {
			return err
		}
		
		status := "成功"
		if !success {
			status = "失败"
		}
		
		log.Printf("%s\t%s\t%s\t%d\t\t%s", 
			version, name, executedAt.Format("2006-01-02 15:04:05"), executionTime, status)
	}
	
	return nil
}

// ResetDatabase 重置数据库（删除所有表）
func (cm *CompleteMigrator) ResetDatabase() error {
	log.Println("开始重置数据库...")
	
	// 获取所有用户表
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT LIKE 'pg_%'
		ORDER BY tablename
	`
	
	rows, err := cm.db.Query(query)
	if err != nil {
		return fmt.Errorf("获取表列表失败: %v", err)
	}
	defer rows.Close()
	
	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}
	
	// 删除所有表
	for _, table := range tables {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		_, err := cm.db.Exec(dropSQL)
		if err != nil {
			log.Printf("删除表 %s 失败: %v", table, err)
		} else {
			log.Printf("删除表: %s", table)
		}
	}
	
	log.Println("数据库重置完成")
	return nil
}