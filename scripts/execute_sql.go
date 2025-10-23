package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 命令行参数
	var sqlFile string
	flag.StringVar(&sqlFile, "file", "create_missing_tables.sql", "SQL文件路径")
	flag.Parse()

	// 数据库连接信息
	connStr := "host=localhost port=5432 user=laojun password=laojun123 dbname=laojun sslmode=disable"
	
	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	err = db.Ping()
	if err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}
	fmt.Println("数据库连接成功!")

	// 读取SQL文件
	sqlContent, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("读取SQL文件失败:", err)
	}

	// 执行SQL
	fmt.Println("开始执行SQL脚本...")
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		log.Fatal("执行SQL脚本失败:", err)
	}

	fmt.Println("SQL脚本执行成功!")
	
	// 验证表是否创建成功
	tables := []string{
		"sm_device_types", "sm_modules", "ug_user_group_members", 
		"ug_user_group_permissions", "ug_permission_templates", "az_permissions",
		"sm_menus", "ua_jwt_keys", "sys_icons", "sys_settings",
		"sys_audit_logs", "pe_extended_permissions", "pe_user_device_permissions",
		"ua_admin", "ua_user_sessions", "az_roles", "az_user_roles", "az_role_permissions",
		"ug_user_groups", "pe_permission_inheritance", "mp_developers", "mp_categories", "mp_plugins",
	}

	fmt.Println("\n验证表创建结果:")
	for _, table := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`
		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			fmt.Printf("  %s: 检查失败 - %v\n", table, err)
		} else if exists {
			fmt.Printf("  %s: ✓ 创建成功\n", table)
		} else {
			fmt.Printf("  %s: ✗ 创建失败\n", table)
		}
	}
}