package main

import (
	"database/sql"
	"fmt"
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

	fmt.Println("开始清理菜单表数据...")

	// 1. 清空菜单表	_, err = db.Exec("DELETE FROM lj_menus")
	if err != nil {
		log.Fatal("清空菜单表失败:", err)
	}
	fmt.Println("菜单表已清空")

	// 2. 重置序列（如果有的话）	_, err = db.Exec("ALTER SEQUENCE IF EXISTS lj_menus_id_seq RESTART WITH 1")
	if err != nil {
		// 忽略序列不存在的错误
		fmt.Println("注意: 序列重置可能失败（正常情况，如果使用UUID主键）")
	}

	// 3. 插入正确的菜单数据	fmt.Println("开始插入正确的菜单数据...")

	menuData := []struct {
		id        string
		title     string
		path      *string
		icon      *string
		component *string
		parentID  *string
		sortOrder int
		isHidden  bool
	}{
		// 主菜单项
		{"550e8400-e29b-41d4-a716-446655440001", "仪表板", stringPtr("/dashboard"), stringPtr("DashboardOutlined"), stringPtr("Dashboard"), nil, 1, false},
		{"550e8400-e29b-41d4-a716-446655440002", "用户管理", stringPtr("/users"), stringPtr("UserOutlined"), stringPtr("UserManagement"), nil, 2, false},
		{"550e8400-e29b-41d4-a716-446655440003", "角色管理", stringPtr("/roles"), stringPtr("TeamOutlined"), stringPtr("RoleManagement"), nil, 3, false},
		{"550e8400-e29b-41d4-a716-446655440004", "权限管理", stringPtr("/permissions"), stringPtr("SafetyOutlined"), stringPtr("PermissionManagement"), nil, 4, false},
		{"550e8400-e29b-41d4-a716-446655440005", "用户组管理", stringPtr("/user-groups"), stringPtr("UsergroupAddOutlined"), stringPtr("UserGroupManagement"), nil, 5, false},
		{"550e8400-e29b-41d4-a716-446655440006", "菜单管理", stringPtr("/menus"), stringPtr("MenuOutlined"), stringPtr("MenuManagement"), nil, 6, false},
		{"550e8400-e29b-41d4-a716-446655440007", "系统设置", stringPtr("/settings"), stringPtr("SettingOutlined"), stringPtr("SystemSettings"), nil, 7, false},
	}

	for _, menu := range menuData {
		_, err := db.Exec(`
			INSERT INTO lj_menus (id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, menu.id, menu.title, menu.path, menu.icon, menu.component, menu.parentID, menu.sortOrder, menu.isHidden)

		if err != nil {
			log.Fatalf("插入菜单 '%s' 失败: %v", menu.title, err)
		}
		fmt.Printf("菜单 '%s' 已插入\n", menu.title)
	}

	// 4. 验证插入结果
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM lj_menus").Scan(&count)
	if err != nil {
		log.Fatal("查询菜单数量失败:", err)
	}

	fmt.Printf("\n清理和恢复完成！\n")
	fmt.Printf("当前菜单记录数: %d\n", count)

	// 5. 显示当前菜单列表
	fmt.Println("\n当前菜单列表:")
	rows, err := db.Query("SELECT id, title, path, sort_order FROM lj_menus ORDER BY sort_order")
	if err != nil {
		log.Fatal("查询菜单列表失败:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, title string
		var path sql.NullString
		var sortOrder int

		err := rows.Scan(&id, &title, &path, &sortOrder)
		if err != nil {
			log.Fatal("扫描菜单数据失败:", err)
		}

		pathStr := "NULL"
		if path.Valid {
			pathStr = path.String
		}

		fmt.Printf("  %d. %s (%s)\n", sortOrder, title, pathStr)
	}
}

func stringPtr(s string) *string {
	return &s
}
