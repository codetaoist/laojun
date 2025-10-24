package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/codetaoist/laojun/internal/database"
	"github.com/codetaoist/laojun/pkg/shared/config"
)

func main() {
	var (
		action     = flag.String("action", "up", "迁移操作: up, down, status, reset")
		configPath = flag.String("config", "./configs/database.yaml", "数据库配置文件路径")
		help       = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 加载数据库配置
	cfg, err := config.LoadDatabaseConfig(*configPath)
	if err != nil {
		log.Fatalf("加载数据库配置失败: %v", err)
	}

	// 连接数据库
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 执行相应的操作
	switch *action {
	case "up":
		if err := database.RunCompleteMigrations(db); err != nil {
			log.Fatalf("执行迁移失败: %v", err)
		}
		fmt.Println("✅ 数据库迁移执行成功")

	case "status":
		if err := database.GetMigrationStatus(db); err != nil {
			log.Fatalf("获取迁移状态失败: %v", err)
		}

	case "reset":
		fmt.Print("⚠️  确定要重置数据库吗？这将删除所有数据！(y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			if err := database.ResetDatabase(db); err != nil {
				log.Fatalf("重置数据库失败: %v", err)
			}
			fmt.Println("✅ 数据库重置成功")
		} else {
			fmt.Println("❌ 操作已取消")
		}

	case "down":
		fmt.Println("⚠️  完整迁移不支持回滚操作，请使用 reset 重置数据库后重新迁移")

	default:
		fmt.Printf("❌ 不支持的操作: %s\n", *action)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("完整数据库迁移工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  go run cmd/db-complete-migrate/main.go [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -action string")
	fmt.Println("        迁移操作 (默认: up)")
	fmt.Println("        up     - 执行数据库迁移")
	fmt.Println("        status - 查看迁移状态")
	fmt.Println("        reset  - 重置数据库（删除所有表）")
	fmt.Println("  -config string")
	fmt.Println("        数据库配置文件路径 (默认: ./configs/database.yaml)")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 执行迁移")
	fmt.Println("  go run cmd/db-complete-migrate/main.go -action=up")
	fmt.Println()
	fmt.Println("  # 查看迁移状态")
	fmt.Println("  go run cmd/db-complete-migrate/main.go -action=status")
	fmt.Println()
	fmt.Println("  # 重置数据库")
	fmt.Println("  go run cmd/db-complete-migrate/main.go -action=reset")
	fmt.Println()
	fmt.Println("  # 使用自定义配置文件")
	fmt.Println("  go run cmd/db-complete-migrate/main.go -config=./custom-config.yaml")
}