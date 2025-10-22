package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接字符串
	dbURL := "postgres://laojun:change-me@localhost:5432/laojun?sslmode=disable"

	// 连接数据库
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("数据库连接成功")

	// 读取SQL文件
	sqlContent, err := ioutil.ReadFile("migration.sql")
	if err != nil {
		log.Fatal("读取migration.sql文件失败:", err)
	}

	// 执行SQL
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		log.Fatal("执行SQL失败:", err)
	}

	fmt.Println("数据库迁移完成！")

	// 验证数据
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM lj_users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		log.Fatal("查询admin用户失败:", err)
	}

	if count > 0 {
		fmt.Println("admin用户创建成功")
	} else {
		fmt.Println("admin用户创建失败")
	}

	// 检查角色数
	err = db.QueryRow("SELECT COUNT(*) FROM lj_roles").Scan(&count)
	if err != nil {
		log.Fatal("查询角色数量失败:", err)
	}
	fmt.Printf("创建 %d 个角色\n", count)

	// 检查权限数
	err = db.QueryRow("SELECT COUNT(*) FROM lj_permissions").Scan(&count)
	if err != nil {
		log.Fatal("查询权限数量失败:", err)
	}
	fmt.Printf("创建 %d 个权限\n", count)

	// 检查扩展权限数
	err = db.QueryRow("SELECT COUNT(*) FROM lj_extended_permissions").Scan(&count)
	if err != nil {
		log.Fatal("查询扩展权限数量失败:", err)
	}
	fmt.Printf("创建 %d 个扩展权限\n", count)
}
