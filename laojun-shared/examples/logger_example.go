package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/codetaoist/laojun-shared/logger"
)

func main() {
	fmt.Println("=== 日志记录使用示例 ===")

	// 1. 基本日志配置
	fmt.Println("\n--- 基本日志配置 ---")
	
	config := logger.Config{
		Level:  "info",
		Format: "json",
		Output: "console",
	}
	
	log := logger.New(config)
	fmt.Println("✓ 日志记录器创建成功")

	// 2. 不同级别的日志记录
	fmt.Println("\n--- 不同级别日志 ---")
	
	log.Debug("这是调试信息", map[string]interface{}{
		"module": "example",
		"action": "debug_test",
	})
	
	log.Info("应用启动成功", map[string]interface{}{
		"version": "1.0.0",
		"port":    8080,
		"env":     "development",
	})
	
	log.Warn("配置文件使用默认值", map[string]interface{}{
		"config_file": "/etc/app/config.yaml",
		"reason":      "文件不存在",
	})
	
	log.Error("数据库连接失败", map[string]interface{}{
		"database": "mysql",
		"host":     "localhost:3306",
		"error":    "connection refused",
	})

	// 3. 带错误对象的日志
	fmt.Println("\n--- 错误日志示例 ---")
	
	err := errors.New("网络连接超时")
	log.Error("API调用失败", map[string]interface{}{
		"api_url":     "https://api.example.com/users",
		"method":      "GET",
		"timeout":     "30s",
		"retry_count": 3,
		"error":       err.Error(),
	})

	// 4. 结构化日志记录
	fmt.Println("\n--- 结构化日志 ---")
	
	// 用户操作日志
	log.Info("用户登录", map[string]interface{}{
		"user_id":    1001,
		"username":   "zhangsan",
		"ip_address": "192.168.1.100",
		"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"login_time": time.Now().Format("2006-01-02 15:04:05"),
	})
	
	// 业务操作日志
	log.Info("订单创建", map[string]interface{}{
		"order_id":     "ORD-20240115-001",
		"user_id":      1001,
		"product_ids":  []int{101, 102, 103},
		"total_amount": 299.99,
		"currency":     "CNY",
		"payment_method": "alipay",
	})
	
	// 系统性能日志
	log.Info("API响应时间", map[string]interface{}{
		"endpoint":      "/api/v1/users",
		"method":        "GET",
		"response_time": "125ms",
		"status_code":   200,
		"request_size":  "1.2KB",
		"response_size": "15.8KB",
	})

	// 5. 文件日志配置
	fmt.Println("\n--- 文件日志配置 ---")
	
	fileConfig := logger.Config{
		Level:  "info",
		Format: "json",
		Output: "file",
		File: logger.FileConfig{
			Filename:   "./logs/app.log",
			MaxSize:    100, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		},
	}
	
	fileLogger := logger.New(fileConfig)
	fmt.Println("✓ 文件日志记录器创建成功")
	
	fileLogger.Info("文件日志测试", map[string]interface{}{
		"message": "这条日志将写入文件",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})

	// 6. 不同格式的日志
	fmt.Println("\n--- 不同格式日志 ---")
	
	// JSON格式
	jsonLogger := logger.New(logger.Config{
		Level:  "info",
		Format: "json",
		Output: "console",
	})
	
	jsonLogger.Info("JSON格式日志", map[string]interface{}{
		"format": "json",
		"structured": true,
	})
	
	// 文本格式
	textLogger := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
		Output: "console",
	})
	
	textLogger.Info("文本格式日志", map[string]interface{}{
		"format": "text",
		"readable": true,
	})

	// 7. 实际应用场景示例
	fmt.Println("\n--- 实际应用场景 ---")
	
	// 模拟Web服务器日志
	simulateWebServerLogs(log)
	
	// 模拟数据库操作日志
	simulateDatabaseLogs(log)
	
	// 模拟业务逻辑日志
	simulateBusinessLogs(log)

	fmt.Println("\n=== 日志记录示例完成 ===")
}

// 模拟Web服务器日志
func simulateWebServerLogs(log logger.Logger) {
	fmt.Println("\n--- Web服务器日志 ---")
	
	// 请求开始
	log.Info("HTTP请求开始", map[string]interface{}{
		"request_id": "req-123456",
		"method":     "POST",
		"path":       "/api/v1/users",
		"ip":         "192.168.1.100",
		"user_agent": "curl/7.68.0",
	})
	
	// 请求处理
	log.Debug("处理用户创建请求", map[string]interface{}{
		"request_id": "req-123456",
		"payload":    map[string]interface{}{
			"name":  "张三",
			"email": "zhangsan@example.com",
		},
	})
	
	// 请求完成
	log.Info("HTTP请求完成", map[string]interface{}{
		"request_id":    "req-123456",
		"status_code":   201,
		"response_time": "245ms",
		"response_size": "156 bytes",
	})
}

// 模拟数据库操作日志
func simulateDatabaseLogs(log logger.Logger) {
	fmt.Println("\n--- 数据库操作日志 ---")
	
	// 数据库连接
	log.Info("数据库连接建立", map[string]interface{}{
		"database": "mysql",
		"host":     "localhost:3306",
		"schema":   "app_db",
		"pool_size": 10,
	})
	
	// SQL查询
	log.Debug("执行SQL查询", map[string]interface{}{
		"sql":           "SELECT * FROM users WHERE email = ?",
		"params":        []string{"zhangsan@example.com"},
		"execution_time": "15ms",
	})
	
	// 慢查询警告
	log.Warn("慢查询检测", map[string]interface{}{
		"sql":           "SELECT * FROM orders o JOIN users u ON o.user_id = u.id WHERE o.created_at > ?",
		"execution_time": "2.5s",
		"threshold":     "1s",
		"affected_rows": 15000,
	})
}

// 模拟业务逻辑日志
func simulateBusinessLogs(log logger.Logger) {
	fmt.Println("\n--- 业务逻辑日志 ---")
	
	// 订单处理流程
	orderID := "ORD-20240115-001"
	
	log.Info("订单处理开始", map[string]interface{}{
		"order_id": orderID,
		"user_id":  1001,
		"step":     "validation",
	})
	
	log.Info("库存检查", map[string]interface{}{
		"order_id":   orderID,
		"product_id": 101,
		"requested":  2,
		"available":  10,
		"step":       "inventory_check",
	})
	
	log.Info("支付处理", map[string]interface{}{
		"order_id":       orderID,
		"payment_method": "alipay",
		"amount":         299.99,
		"currency":       "CNY",
		"step":           "payment",
	})
	
	log.Info("订单处理完成", map[string]interface{}{
		"order_id":     orderID,
		"status":       "completed",
		"total_time":   "1.2s",
		"step":         "completion",
	})
	
	// 异常情况处理
	log.Error("支付失败", map[string]interface{}{
		"order_id":       "ORD-20240115-002",
		"payment_method": "wechat_pay",
		"amount":         199.99,
		"error_code":     "INSUFFICIENT_BALANCE",
		"error_message":  "余额不足",
		"step":           "payment_failed",
	})
}