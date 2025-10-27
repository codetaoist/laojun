package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/health"
)

func main() {
	fmt.Println("=== 健康检查使用示例 ===")

	// 1. 创建健康检查器
	fmt.Println("\n--- 创建健康检查器 ---")
	
	config := health.Config{
		Timeout: time.Second * 5,
	}
	
	h := health.New(config)
	fmt.Println("✓ 健康检查器创建成功")

	// 2. 添加数据库健康检查
	fmt.Println("\n--- 数据库健康检查 ---")
	
	dbChecker := health.NewCustomChecker("database", func(ctx context.Context) (health.Status, string, error) {
		// 模拟数据库连接检查
		fmt.Println("  检查数据库连接...")
		time.Sleep(time.Millisecond * 100) // 模拟检查耗时
		
		// 这里可以添加真实的数据库连接检查逻辑
		// 例如: return db.PingContext(ctx)
		
		return health.StatusHealthy, "Database connection is healthy", nil // 模拟检查成功
	})
	
	h.AddChecker(dbChecker)
	fmt.Println("✓ 数据库健康检查器已添加")

	// 3. 添加Redis健康检查
	fmt.Println("\n--- Redis健康检查 ---")
	
	redisChecker := health.NewCustomChecker("redis", func(ctx context.Context) (health.Status, string, error) {
		// 模拟Redis连接检查
		fmt.Println("  检查Redis连接...")
		time.Sleep(time.Millisecond * 50) // 模拟检查耗时
		
		// 这里可以添加真实的Redis连接检查逻辑
		// 例如: return redisClient.Ping(ctx).Err()
		
		return health.StatusHealthy, "Redis connection is healthy", nil // 模拟检查成功
	})
	
	h.AddChecker(redisChecker)
	fmt.Println("✓ Redis健康检查器已添加")

	// 4. 添加外部API健康检查
	fmt.Println("\n--- 外部API健康检查 ---")
	
	apiChecker := health.NewCustomChecker("external-api", func(ctx context.Context) (health.Status, string, error) {
		// 模拟外部API检查
		fmt.Println("  检查外部API...")
		time.Sleep(time.Millisecond * 200) // 模拟检查耗时
		
		// 这里可以添加真实的API健康检查逻辑
		// 例如: 
		// resp, err := http.Get("https://api.example.com/health")
		// if err != nil || resp.StatusCode != 200 {
		//     return health.StatusUnhealthy, "API不可用", fmt.Errorf("API不可用")
		// }
		
		return health.StatusHealthy, "External API is healthy", nil // 模拟检查成功
	})
	
	h.AddChecker(apiChecker)
	fmt.Println("✓ 外部API健康检查器已添加")

	// 5. 添加一个会失败的检查器（演示失败情况）
	fmt.Println("\n--- 失败检查器示例 ---")
	
	failingChecker := health.NewCustomChecker("failing-service", func(ctx context.Context) (health.Status, string, error) {
		// 模拟失败的服务检查
		fmt.Println("  检查失败服务...")
		return health.StatusUnhealthy, "Service unavailable", fmt.Errorf("服务不可用")
	})
	
	h.AddChecker(failingChecker)
	fmt.Println("✓ 失败检查器已添加（用于演示）")

	// 6. 执行健康检查
	fmt.Println("\n--- 执行健康检查 ---")
	
	ctx := context.Background()
	report := h.Check(ctx)
	
	// 7. 显示检查结果
	fmt.Println("\n--- 健康检查报告 ---")
	
	fmt.Printf("整体状态: %s\n", report.Status)
	fmt.Printf("检查时间: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("总耗时: %v\n", report.Duration)
	fmt.Printf("检查项数量: %d\n", len(report.Checks))
	
	fmt.Println("\n详细检查结果:")
	for name, check := range report.Checks {
		status := "✓ 正常"
		if check.Status != "pass" {
			status = "✗ 异常"
		}
		
		fmt.Printf("  %s [%s]: %s", name, status, check.Status)
		if check.Error != "" {
			fmt.Printf(" - 错误: %s", check.Error)
		}
		fmt.Printf(" (耗时: %v)\n", check.Duration)
	}

	// 8. JSON格式输出
	fmt.Println("\n--- JSON格式报告 ---")
	
	jsonReport, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("JSON序列化失败: %v", err)
	} else {
		fmt.Println(string(jsonReport))
	}

	// 9. 演示超时情况
	fmt.Println("\n--- 超时检查示例 ---")
	
	timeoutConfig := health.Config{
		Timeout: time.Millisecond * 50, // 设置很短的超时时间
	}
	
	timeoutHealth := health.New(timeoutConfig)
	
	slowChecker := health.NewCustomChecker("slow-service", func(ctx context.Context) (health.Status, string, error) {
		// 模拟慢服务
		fmt.Println("  检查慢服务...")
		time.Sleep(time.Millisecond * 100) // 超过超时时间
		return health.StatusHealthy, "Slow service is healthy", nil
	})
	
	timeoutHealth.AddChecker(slowChecker)
	
	timeoutReport := timeoutHealth.Check(ctx)
	fmt.Printf("超时检查状态: %s\n", timeoutReport.Status)
	
	for name, check := range timeoutReport.Checks {
		fmt.Printf("  %s: %s", name, check.Status)
		if check.Error != "" {
			fmt.Printf(" - 错误: %s", check.Error)
		}
		fmt.Println()
	}

	// 10. 实际应用场景示例
	fmt.Println("\n--- 实际应用场景 ---")
	
	// 创建生产环境健康检查器
	prodHealth := health.New(health.Config{
		Timeout: time.Second * 10,
	})
	
	// 添加关键服务检查
	prodHealth.AddChecker(health.NewCustomChecker("mysql", func(ctx context.Context) (health.Status, string, error) {
		// 实际的MySQL检查
		fmt.Println("  检查MySQL数据库...")
		// return db.PingContext(ctx)
		return health.StatusHealthy, "MySQL database is healthy", nil
	}))
	
	prodHealth.AddChecker(health.NewCustomChecker("redis-cache", func(ctx context.Context) (health.Status, string, error) {
		// 实际的Redis检查
		fmt.Println("  检查Redis缓存...")
		// return redisClient.Ping(ctx).Err()
		return health.StatusHealthy, "Redis cache is healthy", nil
	}))
	
	prodHealth.AddChecker(health.NewCustomChecker("message-queue", func(ctx context.Context) (health.Status, string, error) {
		// 实际的消息队列检查
		fmt.Println("  检查消息队列...")
		// return mqClient.HealthCheck()
		return health.StatusHealthy, "Message queue is healthy", nil
	}))
	
	// 执行生产环境检查
	prodReport := prodHealth.Check(ctx)
	
	fmt.Printf("生产环境健康状态: %s\n", prodReport.Status)
	fmt.Printf("所有服务正常: %t\n", prodReport.Status == "pass")
	
	// 可以基于健康检查结果决定是否继续提供服务
	if prodReport.Status != "pass" {
		fmt.Println("⚠ 警告: 部分服务异常，建议检查系统状态")
	} else {
		fmt.Println("✓ 所有服务运行正常")
	}

	fmt.Println("\n=== 健康检查示例完成 ===")
}