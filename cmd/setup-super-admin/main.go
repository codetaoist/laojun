package main

import (
	"fmt"
	"io/ioutil"
	"log"

	sharedconfig "github.com/codetaoist/laojun/pkg/shared/config"
	shareddb "github.com/codetaoist/laojun/pkg/shared/database"
)

func main() {
	// 加载 .env 文件
	sharedconfig.LoadDotenv()
	
	// 加载配置
	cfg, err := sharedconfig.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库连接
	db, err := shareddb.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 读取 SQL 脚本
	sqlContent, err := ioutil.ReadFile("../../setup_super_admin.sql")
	if err != nil {
		log.Fatalf("读取 SQL 脚本失败: %v", err)
	}

	// 执行 SQL 脚本
	fmt.Println("正在设置 super_admin 角色...")
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		log.Fatalf("执行 SQL 脚本失败: %v", err)
	}

	fmt.Println("✅ super_admin 角色设置成功!")
	fmt.Println("admin 用户现在拥有 super_admin 角色")
}