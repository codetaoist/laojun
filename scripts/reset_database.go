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
	// 数据库连接信息
	dbHost := "localhost"
	dbPort := "5432"
	dbUser := "laojun"
	dbPassword := "laojun123"
	dbName := "laojun"

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 1. 清空数据库中的所有表
	fmt.Println("🧹 开始清空数据库...")
	if err := dropAllTables(db); err != nil {
		log.Fatalf("清空数据库失败: %v", err)
	}
	fmt.Println("✅ 数据库清空完成")

	// 2. 执行新的迁移脚本
	fmt.Println("📦 开始执行数据库迁移...")
	if err := executeMigration(db, "create_missing_tables.sql"); err != nil {
		log.Fatalf("执行迁移失败: %v", err)
	}
	fmt.Println("✅ 数据库迁移完成")

	// 3. 验证表创建
	fmt.Println("🔍 验证表创建...")
	if err := verifyTables(db); err != nil {
		log.Fatalf("表验证失败: %v", err)
	}
	fmt.Println("✅ 所有表创建成功")

	fmt.Println("🎉 数据库重置和迁移完成！")
}

// 清空数据库中的所有表
func dropAllTables(db *sql.DB) error {
	// 获取所有表名
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("查询表名失败: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("扫描表名失败: %v", err)
		}
		tables = append(tables, tableName)
	}

	// 删除所有表
	if len(tables) > 0 {
		fmt.Printf("发现 %d 个表，开始删除...\n", len(tables))
		
		// 先禁用外键约束检查
		_, err = db.Exec("SET session_replication_role = replica;")
		if err != nil {
			return fmt.Errorf("禁用外键约束失败: %v", err)
		}

		for _, table := range tables {
			dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
			_, err := db.Exec(dropSQL)
			if err != nil {
				return fmt.Errorf("删除表 %s 失败: %v", table, err)
			}
			fmt.Printf("  ✓ 删除表: %s\n", table)
		}

		// 重新启用外键约束检查
		_, err = db.Exec("SET session_replication_role = DEFAULT;")
		if err != nil {
			return fmt.Errorf("启用外键约束失败: %v", err)
		}
	} else {
		fmt.Println("数据库中没有表需要删除")
	}

	return nil
}

// 执行迁移脚本
func executeMigration(db *sql.DB, filename string) error {
	// 读取SQL文件
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取文件 %s 失败: %v", filename, err)
	}

	sqlContent := string(content)
	
	// 分割SQL语句（按分号分割，但忽略注释中的分号）
	statements := splitSQL(sqlContent)
	
	fmt.Printf("执行 %d 条SQL语句...\n", len(statements))

	// 执行每条SQL语句
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		_, err := db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("执行第 %d 条SQL语句失败: %v\nSQL: %s", i+1, err, stmt)
		}
		
		// 显示进度
		if i%10 == 0 || strings.Contains(strings.ToUpper(stmt), "CREATE TABLE") {
			fmt.Printf("  ✓ 执行进度: %d/%d\n", i+1, len(statements))
		}
	}

	return nil
}

// 分割SQL语句
func splitSQL(content string) []string {
	var statements []string
	var current strings.Builder
	
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 跳过空行
		if trimmed == "" {
			continue
		}
		
		// 跳过注释行
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		
		current.WriteString(line)
		current.WriteString("\n")
		
		// 如果行以分号结尾，则认为是一条完整的语句
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}
	
	// 处理最后一条语句（如果没有以分号结尾）
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	
	return statements
}

// 验证表创建
func verifyTables(db *sql.DB) error {
	expectedTables := []string{
		"ua_admin", "ua_user_sessions", "ua_jwt_keys",
		"az_roles", "az_permissions", "az_modules", "az_user_roles", "az_role_permissions",
		"sm_menus", "sm_device_types",
		"ug_user_groups", "ug_user_group_members", "ug_permission_templates", "ug_user_group_permissions",
		"pe_extended_permissions", "pe_permission_inheritance", "pe_user_device_permissions",
		"sys_settings", "sys_icons", "sys_audit_logs",
		"mp_developers", "mp_categories", "mp_plugins", "mp_plugin_versions", 
		"mp_user_favorites", "mp_user_purchases", "mp_plugin_reviews",
	}

	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		ORDER BY tablename
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("查询表名失败: %v", err)
	}
	defer rows.Close()

	var actualTables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("扫描表名失败: %v", err)
		}
		actualTables = append(actualTables, tableName)
	}

	fmt.Printf("期望创建 %d 个表，实际创建 %d 个表\n", len(expectedTables), len(actualTables))
	
	// 检查每个期望的表是否存在
	tableMap := make(map[string]bool)
	for _, table := range actualTables {
		tableMap[table] = true
	}

	var missingTables []string
	for _, expected := range expectedTables {
		if !tableMap[expected] {
			missingTables = append(missingTables, expected)
		} else {
			fmt.Printf("  ✓ %s\n", expected)
		}
	}

	if len(missingTables) > 0 {
		return fmt.Errorf("缺失的表: %v", missingTables)
	}

	return nil
}