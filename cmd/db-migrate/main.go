package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/codetaoist/laojun/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	var (
		action   = flag.String("action", "up", "迁移操作: up, down, version, reset, rollback")
		version  = flag.String("version", "", "目标版本 (仅用于 rollback)")
		dbURL    = flag.String("db", "", "数据库连接字符串")
		help     = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// 获取数据库连接字符串
	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
		if *dbURL == "" {
			log.Fatal("请提供数据库连接字符串 (-db 参数或 DATABASE_URL 环境变量)")
		}
	}

	// 连接数据库
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// 创建迁移器
	migrator := database.NewMigrator(db)

	// 执行操作
	switch *action {
	case "up":
		if err := migrator.RunMigrations(); err != nil {
			log.Fatalf("执行迁移失败: %v", err)
		}
		fmt.Println("迁移执行成功")

	case "down":
		if err := migrator.Reset(); err != nil {
			log.Fatalf("重置数据库失败: %v", err)
		}
		fmt.Println("数据库重置成功")

	case "version":
		version, dirty, err := migrator.GetVersion()
		if err != nil {
			log.Fatalf("获取版本失败: %v", err)
		}
		if version == 0 {
			fmt.Println("当前版本: 无 (数据库未初始化)")
		} else {
			status := "正常"
			if dirty {
				status = "脏状态"
			}
			fmt.Printf("当前版本: %d (%s)\n", version, status)
		}

	case "reset":
		fmt.Print("确认要重置数据库吗？这将删除所有数据 (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			if err := migrator.Reset(); err != nil {
				log.Fatalf("重置数据库失败: %v", err)
			}
			fmt.Println("数据库重置成功")
		} else {
			fmt.Println("操作已取消")
		}

	case "rollback":
		if *version == "" {
			log.Fatal("回滚操作需要指定目标版本 (-version 参数)")
		}
		targetVersion, err := strconv.ParseUint(*version, 10, 32)
		if err != nil {
			log.Fatalf("无效的版本号: %v", err)
		}
		if err := migrator.Rollback(uint(targetVersion)); err != nil {
			log.Fatalf("回滚失败: %v", err)
		}
		fmt.Printf("回滚到版本 %d 成功\n", targetVersion)

	default:
		fmt.Printf("未知操作: %s\n", *action)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("数据库迁移管理工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  db-migrate [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -action string")
	fmt.Println("        迁移操作 (默认: up)")
	fmt.Println("        up      - 执行所有待执行的迁移")
	fmt.Println("        down    - 回滚所有迁移")
	fmt.Println("        version - 显示当前数据库版本")
	fmt.Println("        reset   - 重置数据库 (交互式确认)")
	fmt.Println("        rollback- 回滚到指定版本")
	fmt.Println("  -version string")
	fmt.Println("        目标版本 (仅用于 rollback 操作)")
	fmt.Println("  -db string")
	fmt.Println("        数据库连接字符串 (或使用 DATABASE_URL 环境变量)")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  db-migrate -action=up -db=\"postgres://user:pass@localhost/dbname?sslmode=disable\"")
	fmt.Println("  db-migrate -action=version")
	fmt.Println("  db-migrate -action=rollback -version=3")
	fmt.Println("  db-migrate -action=reset")
}