package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// "github.com/codetaoist/laojun/tools"
)

func main() {
	var (
		configDir   = flag.String("config-dir", "d:/taishanglaojun/etc/laojun", "配置文件目录")
		centerURL   = flag.String("center-url", "http://localhost:8090", "配置中心URL")
		dryRun      = flag.Bool("dry-run", false, "干运行模式，不实际迁移配置")
		environment = flag.String("env", "development", "环境名称")
		help        = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	fmt.Printf("配置迁移工具\n")
	fmt.Printf("配置目录: %s\n", *configDir)
	fmt.Printf("配置中心: %s\n", *centerURL)
	fmt.Printf("环境: %s\n", *environment)
	fmt.Printf("干运行模式: %t\n", *dryRun)
	fmt.Println()

	// 检查配置目录是否存在
	if _, err := os.Stat(*configDir); os.IsNotExist(err) {
		log.Fatalf("配置目录不存在 %s", *configDir)
	}

	// 创建迁移工具
	// migrator := tools.NewConfigMigrator(*centerURL, *dryRun)
	fmt.Println("迁移工具暂未实现")

	// 获取所有配置文件
	configFiles, err := findConfigFiles(*configDir)
	if err != nil {
		log.Fatalf("查找配置文件失败: %v", err)
	}

	if len(configFiles) == 0 {
		fmt.Println("未找到配置文件")
		return
	}

	fmt.Printf("找到 %d 个配置文件\n", len(configFiles))
	for _, file := range configFiles {
		fmt.Printf("  - %s\n", file)
	}
	fmt.Println()

	// 执行迁移
	// ctx := context.Background()
	// totalMigrated := 0

	for _, configFile := range configFiles {
		fmt.Printf("发现配置文件: %s\n", configFile)

		// count, err := migrator.MigrateFile(ctx, configFile, *environment)
		// if err != nil {
		// 	log.Printf("迁移文件 %s 失败: %v", configFile, err)
		// 	continue
		// }

		// totalMigrated += count
		// fmt.Printf("  迁移 %d 个配置项\n", count)
	}

	fmt.Printf("\n发现 %d 个配置文件（迁移功能暂未实现）\n", len(configFiles))

	if *dryRun {
		fmt.Println("\n注意: 这是干运行模式，配置未实际写入配置中心")
		fmt.Println("要执行实际迁移，请移除 --dry-run 参数")
	}
}

// findConfigFiles 查找配置文件
func findConfigFiles(configDir string) ([]string, error) {
	var configFiles []string

	err := filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" || ext == ".json" {
			configFiles = append(configFiles, path)
		}

		return nil
	})

	return configFiles, err
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("配置迁移工具 - 将现有配置文件迁移到配置中心")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  migrate [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --config-dir string    配置文件目录 (默认: d:/taishanglaojun/etc/laojun)")
	fmt.Println("  --center-url string    配置中心URL (默认: http://localhost:8090)")
	fmt.Println("  --env string           环境名称 (默认: development)")
	fmt.Println("  --dry-run              干运行模式，不实际迁移配置")
	fmt.Println("  --help                 显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 干运行模式查看将要迁移的配置")
	fmt.Println("  migrate --dry-run")
	fmt.Println()
	fmt.Println("  # 迁移到生产环境")
	fmt.Println("  migrate --env production")
	fmt.Println()
	fmt.Println("  # 指定配置目录和配置中心URL")
	fmt.Println("  migrate --config-dir /path/to/configs --center-url http://config-center:8090")
}
