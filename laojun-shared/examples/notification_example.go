package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/notification"
)

func main() {
	fmt.Println("=== NOTIFICATION 示例 ===")
	
	// 1. 创建配置
	config := notification.DefaultConfig()
	config.Debug = true
	
	// 2. 创建实例
	impl := notification.New(config)
	defer func() {
		if closer, ok := impl.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				log.Printf("关闭失败: %v", err)
			}
		}
	}()
	
	// 3. 使用示例
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// TODO: Add your usage examples here
	fmt.Println("TODO: 添加使用示例")
	
	fmt.Println("=== 示例完成 ===")
}
