package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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

	// 创建备份文件
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("menu_backup_%s.sql", timestamp)

	file, err := os.Create(backupFile)
	if err != nil {
		log.Fatal("创建备份文件失败:", err)
	}
	defer file.Close()

	// 写入备份文件头部
	file.WriteString("-- 菜单表备份文件\n")
	file.WriteString(fmt.Sprintf("-- 备份时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	file.WriteString("-- 总记录数: 266\n\n")
	file.WriteString("-- 删除现有表（如果存在）\n")
	file.WriteString("DROP TABLE IF EXISTS lj_menus_backup;\n\n")
	file.WriteString("-- 创建备份表结构\n")
	file.WriteString(`CREATE TABLE lj_menus_backup (
    id UUID PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    path VARCHAR(200),
    icon VARCHAR(100),
    component VARCHAR(200),
    parent_id UUID,
    sort_order INTEGER DEFAULT 0,
    is_hidden BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

`)

	// 查询所有菜单数据
	rows, err := db.Query(`
		SELECT id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at 
		FROM lj_menus 
		ORDER BY created_at
	`)
	if err != nil {
		log.Fatal("查询菜单数据失败:", err)
	}
	defer rows.Close()

	file.WriteString("-- 插入备份数据\n")
	count := 0
	for rows.Next() {
		var id, title string
		var path, icon, component, parentID sql.NullString
		var sortOrder int
		var isHidden bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &title, &path, &icon, &component, &parentID, &sortOrder, &isHidden, &createdAt, &updatedAt)
		if err != nil {
			log.Fatal("扫描行数据失败:", err)
		}

		// 构建INSERT语句
		insertSQL := fmt.Sprintf("INSERT INTO lj_menus_backup (id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at) VALUES ('%s', '%s', ", id, title)

		if path.Valid {
			insertSQL += fmt.Sprintf("'%s', ", path.String)
		} else {
			insertSQL += "NULL, "
		}

		if icon.Valid {
			insertSQL += fmt.Sprintf("'%s', ", icon.String)
		} else {
			insertSQL += "NULL, "
		}

		if component.Valid {
			insertSQL += fmt.Sprintf("'%s', ", component.String)
		} else {
			insertSQL += "NULL, "
		}

		if parentID.Valid {
			insertSQL += fmt.Sprintf("'%s', ", parentID.String)
		} else {
			insertSQL += "NULL, "
		}

		insertSQL += fmt.Sprintf("%d, %t, '%s', '%s');\n",
			sortOrder, isHidden,
			createdAt.Format("2006-01-02 15:04:05"),
			updatedAt.Format("2006-01-02 15:04:05"))

		file.WriteString(insertSQL)
		count++
	}

	file.WriteString(fmt.Sprintf("\n-- 备份完成，共备份 %d 条记录\n", count))

	fmt.Printf("菜单数据备份完成！\n")
	fmt.Printf("备份文件: %s\n", backupFile)
	fmt.Printf("备份记录数: %d\n", count)
}
