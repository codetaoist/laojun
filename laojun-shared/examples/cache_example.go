package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/cache"
	"github.com/codetaoist/laojun-shared/config"
)

func main() {
	// 缓存使用示例
	fmt.Println("=== 缓存管理器使用示例 ===")

	// 1. 创建内存缓存配置
	cacheConfig := &cache.CacheConfig{
		Type:              cache.CacheTypeMemory,
		EnableMultiLevel:  false,
		MemoryCleanup:     time.Minute * 5,
		DefaultExpiration: time.Minute * 10,
	}

	// 2. 创建缓存管理器
	manager, err := cache.NewManager(cacheConfig, nil)
	if err != nil {
		log.Fatal("创建缓存管理器失败:", err)
	}

	ctx := context.Background()

	// 3. 基本的Set/Get操作
	fmt.Println("\n--- 基本操作 ---")
	
	// 设置字符串值
	err = manager.Set(ctx, "user:1001", "张三", time.Minute*5)
	if err != nil {
		log.Printf("设置缓存失败: %v", err)
	} else {
		fmt.Println("✓ 设置用户信息成功")
	}

	// 获取字符串值
	value, err := manager.Get(ctx, "user:1001")
	if err != nil {
		log.Printf("获取缓存失败: %v", err)
	} else {
		fmt.Printf("✓ 获取用户信息: %s\n", value)
	}

	// 4. JSON对象操作
	fmt.Println("\n--- JSON对象操作 ---")
	
	user := map[string]interface{}{
		"id":    1001,
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   25,
	}

	// 设置JSON对象
	err = manager.SetJSON(ctx, "user:detail:1001", user, time.Minute*10)
	if err != nil {
		log.Printf("设置JSON缓存失败: %v", err)
	} else {
		fmt.Println("✓ 设置用户详情成功")
	}

	// 获取JSON对象
	var retrievedUser map[string]interface{}
	err = manager.GetJSON(ctx, "user:detail:1001", &retrievedUser)
	if err != nil {
		log.Printf("获取JSON缓存失败: %v", err)
	} else {
		fmt.Printf("✓ 获取用户详情: %+v\n", retrievedUser)
	}

	// 5. 检查键是否存在
	fmt.Println("\n--- 存在性检查 ---")
	
	exists, err := manager.Exists(ctx, "user:1001")
	if err != nil {
		log.Printf("检查键存在性失败: %v", err)
	} else {
		fmt.Printf("✓ 键 'user:1001' 存在: %t\n", exists > 0)
	}

	exists, err = manager.Exists(ctx, "user:9999")
	if err != nil {
		log.Printf("检查键存在性失败: %v", err)
	} else {
		fmt.Printf("✓ 键 'user:9999' 存在: %t\n", exists > 0)
	}

	// 6. 删除操作
	fmt.Println("\n--- 删除操作 ---")
	
	err = manager.Del(ctx, "user:1001")
	if err != nil {
		log.Printf("删除缓存失败: %v", err)
	} else {
		fmt.Println("✓ 删除用户信息成功")
	}

	// 验证删除结果
	exists, err = manager.Exists(ctx, "user:1001")
	if err != nil {
		log.Printf("检查键存在性失败: %v", err)
	} else {
		fmt.Printf("✓ 删除后键 'user:1001' 存在: %t\n", exists > 0)
	}

	// 7. 批量操作
	fmt.Println("\n--- 批量操作 ---")
	
	// 批量设置
	keys := []string{"batch:1", "batch:2", "batch:3"}
	values := []interface{}{"值1", "值2", "值3"}
	
	for i, key := range keys {
		err = manager.Set(ctx, key, values[i], time.Minute*5)
		if err != nil {
			log.Printf("批量设置失败 %s: %v", key, err)
		}
	}
	fmt.Println("✓ 批量设置完成")

	// 批量获取
	for _, key := range keys {
		value, err := manager.Get(ctx, key)
		if err != nil {
			log.Printf("批量获取失败 %s: %v", key, err)
		} else {
			fmt.Printf("✓ %s = %s\n", key, value)
		}
	}

	// 8. Redis缓存示例（需要Redis连接）
	fmt.Println("\n--- Redis缓存示例 ---")
	
	redisConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	redisCacheConfig := &cache.CacheConfig{
		Type:              cache.CacheTypeRedis,
		EnableMultiLevel:  false,
		DefaultExpiration: time.Minute * 10,
	}

	redisManager, err := cache.NewManager(redisCacheConfig, redisConfig)
	if err != nil {
		fmt.Printf("⚠ Redis缓存管理器创建失败（可能Redis未启动）: %v\n", err)
	} else {
		// Redis操作示例
		err = redisManager.Set(ctx, "redis:test", "Redis测试值", time.Minute*5)
		if err != nil {
			fmt.Printf("⚠ Redis设置失败: %v\n", err)
		} else {
			value, err := redisManager.Get(ctx, "redis:test")
			if err != nil {
				fmt.Printf("⚠ Redis获取失败: %v\n", err)
			} else {
				fmt.Printf("✓ Redis测试成功: %s\n", value)
			}
		}
	}

	fmt.Println("\n=== 缓存示例完成 ===")
}