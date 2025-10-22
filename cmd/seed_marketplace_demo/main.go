package main

import (
	"fmt"
	"log"

	"github.com/codetaoist/laojun/internal/services"
	"github.com/codetaoist/laojun/pkg/shared/config"
	shareddb "github.com/codetaoist/laojun/pkg/shared/database"
	"github.com/codetaoist/laojun/pkg/shared/models"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := shareddb.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	defer db.Close()

	// 初始化服务
	authService := services.NewAuthService(db)
	communityService := services.NewCommunityService(db)

	// 注册测试用户
	req := &services.RegisterRequest{
		Username:        "testuser1",
		Email:           "testuser1@example.com",
		Password:        "Passw0rd!",
		ConfirmPassword: "Passw0rd!",
		FullName:        "测试用户一",
	}

	user, err := authService.Register(req)
	if err != nil {
		fmt.Printf("Register user: %v\n", err)
	} else {
		fmt.Printf("Registered user: %s (%s)\n", user.User.Username, user.User.ID.String())
	}

	// 获取论坛分类
	categories, err := communityService.GetForumCategories()
	if err != nil {
		log.Fatalf("Failed to get forum categories: %v", err)
	}
	if len(categories) == 0 {
		log.Fatalf("No forum categories found; please ensure migration seeds are applied")
	}
	cat := categories[0]

	// 创建测试帖子
	post := &models.ForumPost{
		CategoryID: cat.ID,
		UserID:     user.User.ID,
		Title:      "Hello World",
		Content:    "这是测试帖子，来自种子脚本。",
	}

	if err := communityService.CreateForumPost(post); err != nil {
		log.Fatalf("Failed to create forum post: %v", err)
	}

	fmt.Printf("Created forum post: %s (category: %s)\n", post.ID.String(), cat.Name)
}