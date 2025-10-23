package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/codetaoist/laojun/internal/services"
	sharedconfig "github.com/codetaoist/laojun/pkg/shared/config"
	shareddb "github.com/codetaoist/laojun/pkg/shared/database"
)

func main() {
	// 定义命令行参数
	var (
		username = flag.String("username", "admin", "管理员用户名")
		email    = flag.String("email", "admin@laojun.local", "管理员邮箱")
		password = flag.String("password", "admin123", "管理员密码")
		force    = flag.Bool("force", false, "强制创建（如果用户已存在则跳过）")
	)
	flag.Parse()

	// 验证参数
	if *username == "" || *email == "" || *password == "" {
		fmt.Println("错误: 用户名、邮箱和密码不能为空")
		flag.Usage()
		os.Exit(1)
	}

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

	// 创建AdminAuthService
	adminAuthService := services.NewAdminAuthService(db)

	// 尝试创建管理员用户
	fmt.Printf("正在创建管理员用户...\n")
	fmt.Printf("用户名: %s\n", *username)
	fmt.Printf("邮箱: %s\n", *email)
	fmt.Printf("密码: %s\n", *password)

	user, err := adminAuthService.CreateAdminUser(*username, *email, *password)
	if err != nil {
		if !*force && (err.Error() == "用户名或邮箱已存在") {
			fmt.Printf("用户已存在: %v\n", err)
			fmt.Println("如果要强制创建，请使用 -force 参数")
			os.Exit(1)
		} else if *force && (err.Error() == "用户名或邮箱已存在") {
			fmt.Println("用户已存在，跳过创建")
			os.Exit(0)
		} else {
			log.Fatalf("创建管理员用户失败: %v", err)
		}
	}

	fmt.Printf("✅ 管理员用户创建成功!\n")
	fmt.Printf("用户ID: %s\n", user.ID)
	fmt.Printf("用户名: %s\n", user.Username)
	fmt.Printf("邮箱: %s\n", user.Email)
	fmt.Printf("创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("\n现在您可以使用这些凭据登录后台管理系统了！")
}