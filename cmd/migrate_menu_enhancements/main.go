package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接配置
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "123456"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "laojun_admin"
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("数据库连接成功")

	// 读取迁移文件
	migrationFile := "d:\\taishanglaojun\\usr\\src\\laojun\\admin-api\\migrations\\add_menu_enhancements.sql"
	sqlContent, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("读取迁移文件失败: %v", err)
	}

	// 执行迁移
	fmt.Println("开始执行数据库迁移...")
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		log.Fatalf("执行迁移失败: %v", err)
	}

	fmt.Println("数据库迁移执行成功！")

	// 验证迁移结果
	fmt.Println("验证迁移结果...")

	// 检查新字段是否添加成功
	var columnExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'lj_menus' AND column_name = 'is_favorite'
		)
	`).Scan(&columnExists)
	if err != nil {
		log.Printf("验证字段失败: %v", err)
	} else if columnExists {
		fmt.Println("菜单表新字段添加成功")
	}

	// 检查新表是否创建成功
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'lj_menu_configs'
		)
	`).Scan(&tableExists)
	if err != nil {
		log.Printf("验证表失败: %v", err)
	} else if tableExists {
		fmt.Println("菜单配置表创建成功")
	}

	// 检查图标库表是否创建成功
	var iconTableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'lj_icon_library'
		)
	`).Scan(&iconTableExists)
	if err != nil {
		log.Printf("验证图标库表失败: %v", err)
	} else if iconTableExists {
		fmt.Println("图标库表创建成功")
	}

	// 检查用户偏好表
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'lj_user_menu_preferences'
		)
	`).Scan(&tableExists)
	if err != nil {
		log.Printf("验证用户偏好表失败: %v", err)
	} else if tableExists {
		fmt.Println("用户菜单偏好表创建成功")
	}

	// 检查图标库数据
	var iconCount int
	err = db.QueryRow("SELECT COUNT(*) FROM lj_icon_library").Scan(&iconCount)
	if err != nil {
		log.Printf("检查图标库数据失败: %v", err)
	} else {
		fmt.Printf("图标库数据插入成功，共 %d 个图标\n", iconCount)
	}

	// 检查配置数据
	var configCount int
	err = db.QueryRow("SELECT COUNT(*) FROM lj_menu_configs").Scan(&configCount)
	if err != nil {
		log.Printf("检查配置数据失败: %v", err)
	} else {
		fmt.Printf("菜单配置数据插入成功，共 %d 个配置\n", configCount)
	}

	fmt.Println("数据库迁移验证完成！")
}
