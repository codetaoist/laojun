package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/laojun-shared/middleware"
	"github.com/codetaoist/laojun-shared/observability"
)

func main() {
	fmt.Println("=== HTTP中间件可观测性示例 ===")

	// 创建可观测性配置
	config := observability.DefaultConfig()
	config.ServiceName = "http-middleware-example"
	config.ServiceVersion = "1.0.0"
	config.Environment = "development"
	config.EnableMonitoring()
	config.EnableTracing()

	// 初始化可观测性
	obs, err := observability.NewObservability(config)
	if err != nil {
		log.Fatalf("初始化可观测性失败: %v", err)
	}
	defer obs.Close()

	// 创建Gin路由器
	router := gin.New()

	// 配置可观测性中间件
	obsConfig := middleware.DefaultObservabilityConfig()
	obsConfig.ServiceName = "web-api"
	obsConfig.RecordUserAgent = true
	obsConfig.LabelExtractor = func(c *gin.Context) map[string]string {
		labels := make(map[string]string)
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			labels["user_id"] = userID
		}
		if apiVersion := c.GetHeader("X-API-Version"); apiVersion != "" {
			labels["api_version"] = apiVersion
		}
		return labels
	}
	obsConfig.AttributeExtractor = func(c *gin.Context) map[string]interface{} {
		attrs := make(map[string]interface{})
		if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
			attrs["session_id"] = sessionID
		}
		return attrs
	}

	// 添加中间件
	router.Use(middleware.RequestID()) // 请求ID中间件
	router.Use(middleware.ObservabilityMiddleware(obs, obsConfig)) // 可观测性中间件
	router.Use(gin.Recovery()) // 恢复中间件

	// 定义路由
	setupRoutes(router)

	// 启动服务器（非阻塞）
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		fmt.Println("HTTP服务器启动在 :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("服务器启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 模拟一些HTTP请求
	simulateRequests()

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
	}

	fmt.Println("=== HTTP中间件示例完成! ===")
}

// setupRoutes 设置路由
func setupRoutes(router *gin.Engine) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		// 添加自定义事件
		middleware.AddEventToRequest(c, "health_check", map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// 用户API
	userGroup := router.Group("/api/v1/users")
	{
		userGroup.GET("/:id", getUserHandler)
		userGroup.POST("/", createUserHandler)
		userGroup.PUT("/:id", updateUserHandler)
		userGroup.DELETE("/:id", deleteUserHandler)
	}

	// 产品API
	productGroup := router.Group("/api/v1/products")
	{
		productGroup.GET("/", getProductsHandler)
		productGroup.GET("/:id", getProductHandler)
		productGroup.POST("/", createProductHandler)
	}

	// 模拟错误的端点
	router.GET("/error", func(c *gin.Context) {
		// 记录错误事件
		middleware.AddEventToRequest(c, "error_occurred", map[string]interface{}{
			"error_type": "simulated_error",
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "This is a simulated error",
		})
	})

	// 模拟慢请求的端点
	router.GET("/slow", func(c *gin.Context) {
		// 启动子操作来追踪慢操作
		if childOp := middleware.StartChildOperation(c, "slow_processing"); childOp != nil {
			childOp.SetAttribute("processing_type", "slow_simulation")
			
			// 模拟慢处理
			time.Sleep(2 * time.Second)
			
			childOp.SetStatus(observability.OperationStatusSuccess)
			childOp.Finish()
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Slow operation completed",
		})
	})
}

// 用户处理器
func getUserHandler(c *gin.Context) {
	userID := c.Param("id")
	
	// 设置用户相关属性
	middleware.SetRequestAttribute(c, "user.id", userID)
	middleware.SetRequestAttribute(c, "operation.type", "user_retrieval")

	// 增加用户查询计数器
	middleware.IncrementRequestCounter(c, "user_queries_total", map[string]string{
		"operation": "get_user",
	})

	c.JSON(http.StatusOK, gin.H{
		"id":   userID,
		"name": "John Doe",
		"email": "john@example.com",
	})
}

func createUserHandler(c *gin.Context) {
	// 记录用户创建事件
	middleware.AddEventToRequest(c, "user_creation_started", nil)

	// 模拟用户创建逻辑
	time.Sleep(100 * time.Millisecond)

	middleware.IncrementRequestCounter(c, "user_operations_total", map[string]string{
		"operation": "create_user",
	})

	c.JSON(http.StatusCreated, gin.H{
		"id":      "123",
		"message": "User created successfully",
	})
}

func updateUserHandler(c *gin.Context) {
	userID := c.Param("id")
	middleware.SetRequestAttribute(c, "user.id", userID)

	c.JSON(http.StatusOK, gin.H{
		"id":      userID,
		"message": "User updated successfully",
	})
}

func deleteUserHandler(c *gin.Context) {
	userID := c.Param("id")
	middleware.SetRequestAttribute(c, "user.id", userID)

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// 产品处理器
func getProductsHandler(c *gin.Context) {
	// 记录查询参数
	if limit := c.Query("limit"); limit != "" {
		middleware.SetRequestAttribute(c, "query.limit", limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"products": []gin.H{
			{"id": "1", "name": "Product 1"},
			{"id": "2", "name": "Product 2"},
		},
	})
}

func getProductHandler(c *gin.Context) {
	productID := c.Param("id")
	middleware.SetRequestAttribute(c, "product.id", productID)

	c.JSON(http.StatusOK, gin.H{
		"id":   productID,
		"name": "Sample Product",
	})
}

func createProductHandler(c *gin.Context) {
	middleware.AddEventToRequest(c, "product_creation", map[string]interface{}{
		"timestamp": time.Now().Unix(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"id":      "456",
		"message": "Product created successfully",
	})
}

// simulateRequests 模拟HTTP请求
func simulateRequests() {
	fmt.Println("\n--- 模拟HTTP请求 ---")

	client := &http.Client{Timeout: 10 * time.Second}
	baseURL := "http://localhost:8080"

	requests := []struct {
		method string
		path   string
		headers map[string]string
	}{
		{"GET", "/health", nil},
		{"GET", "/api/v1/users/123", map[string]string{"X-User-ID": "current_user", "X-API-Version": "v1"}},
		{"POST", "/api/v1/users", map[string]string{"X-Session-ID": "session_123"}},
		{"GET", "/api/v1/products", nil},
		{"GET", "/api/v1/products/456", nil},
		{"POST", "/api/v1/products", nil},
		{"PUT", "/api/v1/users/123", nil},
		{"DELETE", "/api/v1/users/123", nil},
		{"GET", "/error", nil},
		{"GET", "/slow", nil},
	}

	for i, req := range requests {
		fmt.Printf("发送请求 %d: %s %s\n", i+1, req.method, req.path)

		httpReq, err := http.NewRequest(req.method, baseURL+req.path, nil)
		if err != nil {
			fmt.Printf("创建请求失败: %v\n", err)
			continue
		}

		// 添加自定义头
		for key, value := range req.headers {
			httpReq.Header.Set(key, value)
		}

		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}
		resp.Body.Close()

		fmt.Printf("响应状态: %d\n", resp.StatusCode)
		time.Sleep(100 * time.Millisecond) // 短暂延迟
	}

	fmt.Println("--- 请求模拟完成 ---")
}