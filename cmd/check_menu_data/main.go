package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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

	fmt.Println("数据库连接成功!")
	fmt.Println("=== 菜单表数据 ===")

	// 查询菜单数据
	rows, err := db.Query(`
		SELECT id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at 
		FROM lj_menus 
		ORDER BY sort_order, created_at
	`)
	if err != nil {
		log.Fatal("查询菜单数据失败:", err)
	}
	defer rows.Close()

	fmt.Printf("%-5s %-15s %-20s %-15s %-20s %-10s %-5s %-8s %-20s\n",
		"ID", "Title", "Path", "Icon", "Component", "ParentID", "Sort", "Hidden", "CreatedAt")
	fmt.Println(strings.Repeat("-", 120))

	for rows.Next() {
		var id, title, sortOrder string
		var path, icon, component, parentID sql.NullString
		var isHidden bool
		var createdAt string

		err := rows.Scan(&id, &title, &path, &icon, &component, &parentID, &sortOrder, &isHidden, &createdAt)
		if err != nil {
			log.Fatal("扫描行数据失败:", err)
		}

		// 处理NULL值
		pathStr := "NULL"
		if path.Valid {
			pathStr = path.String
		}
		iconStr := "NULL"
		if icon.Valid {
			iconStr = icon.String
		}
		componentStr := "NULL"
		if component.Valid {
			componentStr = component.String
		}
		parentIDStr := "NULL"
		if parentID.Valid {
			parentIDStr = parentID.String
		}

		fmt.Printf("%-5s %-15s %-20s %-15s %-20s %-10s %-5s %-8t %-20s\n",
			id, title, pathStr, iconStr, componentStr, parentIDStr, sortOrder, isHidden, createdAt[:19])
	}

	// 统计重复数据
	fmt.Println("\n=== 重复数据统计 ===")

	// 检查重复的title
	rows, err = db.Query(`
		SELECT title, COUNT(*) as count 
		FROM lj_menus 
		GROUP BY title 
		HAVING COUNT(*) > 1
		ORDER BY count DESC
	`)
	if err != nil {
		log.Fatal("查询重复title失败:", err)
	}
	defer rows.Close()

	fmt.Println("重复的菜单标题:")
	for rows.Next() {
		var title string
		var count int
		err := rows.Scan(&title, &count)
		if err != nil {
			log.Fatal("扫描重复数据失败:", err)
		}
		fmt.Printf("  %s: %d 条记录\n", title, count)
	}

	// 检查重复的path
	rows, err = db.Query(`
		SELECT path, COUNT(*) as count 
		FROM lj_menus 
		WHERE path IS NOT NULL AND path != ''
		GROUP BY path 
		HAVING COUNT(*) > 1
		ORDER BY count DESC
	`)
	if err != nil {
		log.Fatal("查询重复path失败:", err)
	}
	defer rows.Close()

	fmt.Println("\n重复的菜单路径:")
	for rows.Next() {
		var path string
		var count int
		err := rows.Scan(&path, &count)
		if err != nil {
			log.Fatal("扫描重复路径失败:", err)
		}
		fmt.Printf("  %s: %d 条记录\n", path, count)
	}

	// 总记录数
	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM lj_menus").Scan(&totalCount)
	if err != nil {
		log.Fatal("查询总记录数失败:", err)
	}
	fmt.Printf("\n总菜单记录数: %d\n", totalCount)
}
