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
	fmt.Println("=== 高级中间件集成示例 ===")

	// 创建可观测性配置
	config := observability.DefaultConfig()
	config.ServiceName = "advanced-middleware-demo"
	config.ServiceVersion = "2.0.0"
	config.Environment = "production"
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

	// 配置中间件链
	setupMiddlewareChain(router, obs)

	// 定义路由
	setupAdvancedRoutes(router)

	// 启动服务器（非阻塞）
	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	go func() {
		fmt.Println("高级中间件服务器启动在 :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("服务器启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 模拟复杂的HTTP请求场景
	simulateAdvancedRequests()

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
	}

	fmt.Println("=== 高级中间件示例完成! ===")
}

// setupMiddlewareChain 设置中间件链
func setupMiddlewareChain(router *gin.Engine, obs observability.Observability) {
	// 1. 请求ID中间件（最先执行）
	router.Use(middleware.RequestID())

	// 2. 安全头中间件
	router.Use(middleware.SecurityHeaders())

	// 3. CORS中间件
	corsConfig := middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-User-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	router.Use(middleware.CORS(corsConfig))

	// 4. 限流中间件
	rateLimitConfig := middleware.RateLimitConfig{
		Enabled: true,
		RPS:     10,
		Burst:   20,
	}
	router.Use(middleware.RateLimit(rateLimitConfig))

	// 5. 可观测性中间件（核心）
	obsConfig := middleware.DefaultObservabilityConfig()
	obsConfig.ServiceName = "advanced-api"
	obsConfig.RecordUserAgent = true
	obsConfig.RecordClientIP = true
	obsConfig.LabelExtractor = func(c *gin.Context) map[string]string {
		labels := make(map[string]string)
		
		// 用户信息
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			labels["user_id"] = userID
		}
		
		// API版本
		if apiVersion := c.GetHeader("X-API-Version"); apiVersion != "" {
			labels["api_version"] = apiVersion
		}
		
		// 客户端类型
		if clientType := c.GetHeader("X-Client-Type"); clientType != "" {
			labels["client_type"] = clientType
		}
		
		// 地理位置（模拟）
		if region := c.GetHeader("X-Region"); region != "" {
			labels["region"] = region
		}
		
		return labels
	}
	obsConfig.AttributeExtractor = func(c *gin.Context) map[string]interface{} {
		attrs := make(map[string]interface{})
		
		// 会话信息
		if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
			attrs["session_id"] = sessionID
		}
		
		// 设备信息
		if deviceID := c.GetHeader("X-Device-ID"); deviceID != "" {
			attrs["device_id"] = deviceID
		}
		
		// 实验标识
		if experiment := c.GetHeader("X-Experiment"); experiment != "" {
			attrs["experiment"] = experiment
		}
		
		return attrs
	}
	router.Use(middleware.ObservabilityMiddleware(obs, obsConfig))

	// 6. 请求大小限制中间件
	router.Use(middleware.RequestSizeLimit(1024 * 1024)) // 1MB

	// 7. 内容类型验证中间件（仅对POST/PUT请求）
	router.Use(middleware.ContentType("application/json", "application/x-www-form-urlencoded"))

	// 8. 恢复中间件（最后执行，捕获panic）
	router.Use(middleware.Recovery())
}

// setupAdvancedRoutes 设置高级路由
func setupAdvancedRoutes(router *gin.Engine) {
	// 健康检查和指标端点
	router.GET("/health", healthCheckHandler)
	router.GET("/metrics", metricsHandler)
	router.GET("/ready", readinessHandler)

	// API版本1
	v1 := router.Group("/api/v1")
	{
		// 用户管理
		users := v1.Group("/users")
		{
			users.GET("/", listUsersHandler)
			users.GET("/:id", getUserDetailHandler)
			users.POST("/", createUserAdvancedHandler)
			users.PUT("/:id", updateUserAdvancedHandler)
			users.DELETE("/:id", deleteUserAdvancedHandler)
			users.GET("/:id/profile", getUserProfileHandler)
			users.POST("/:id/avatar", uploadAvatarHandler)
		}

		// 订单管理
		orders := v1.Group("/orders")
		{
			orders.GET("/", listOrdersHandler)
			orders.GET("/:id", getOrderDetailHandler)
			orders.POST("/", createOrderHandler)
			orders.PUT("/:id/status", updateOrderStatusHandler)
			orders.POST("/:id/payment", processPaymentHandler)
		}

		// 分析端点
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/dashboard", dashboardHandler)
			analytics.GET("/reports/:type", reportHandler)
			analytics.POST("/events", trackEventHandler)
		}
	}

	// API版本2（新功能）
	v2 := router.Group("/api/v2")
	{
		v2.GET("/features", listFeaturesHandler)
		v2.POST("/batch", batchOperationHandler)
		v2.GET("/stream", streamDataHandler)
	}

	// 管理端点
	admin := router.Group("/admin")
	{
		admin.GET("/stats", adminStatsHandler)
		admin.POST("/cache/clear", clearCacheHandler)
		admin.GET("/config", getConfigHandler)
	}
}

// 处理器实现
func healthCheckHandler(c *gin.Context) {
	middleware.AddEventToRequest(c, "health_check_performed", map[string]interface{}{
		"check_type": "basic",
		"timestamp": time.Now().Unix(),
	})

	middleware.IncrementRequestCounter(c, "health_checks_total", map[string]string{
		"check_type": "basic",
	})

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "2.0.0",
	})
}

func metricsHandler(c *gin.Context) {
	middleware.SetRequestAttribute(c, "endpoint_type", "metrics")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics endpoint - would return Prometheus metrics",
	})
}

func readinessHandler(c *gin.Context) {
	// 模拟依赖检查
	if childOp := middleware.StartChildOperation(c, "dependency_check"); childOp != nil {
		childOp.SetAttribute("check_type", "database")
		time.Sleep(50 * time.Millisecond) // 模拟数据库检查
		childOp.SetStatus(observability.OperationStatusSuccess)
		childOp.Finish()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"dependencies": gin.H{
			"database": "healthy",
			"cache":    "healthy",
		},
	})
}

func listUsersHandler(c *gin.Context) {
	// 记录查询参数
	if page := c.Query("page"); page != "" {
		middleware.SetRequestAttribute(c, "query.page", page)
	}
	if limit := c.Query("limit"); limit != "" {
		middleware.SetRequestAttribute(c, "query.limit", limit)
	}

	middleware.IncrementRequestCounter(c, "user_list_requests", map[string]string{
		"operation": "list",
	})

	c.JSON(http.StatusOK, gin.H{
		"users": []gin.H{
			{"id": "1", "name": "Alice", "email": "alice@example.com"},
			{"id": "2", "name": "Bob", "email": "bob@example.com"},
		},
		"total": 2,
		"page":  1,
	})
}

func getUserDetailHandler(c *gin.Context) {
	userID := c.Param("id")
	middleware.SetRequestAttribute(c, "user.id", userID)

	// 模拟数据库查询
	if childOp := middleware.StartChildOperation(c, "database_query"); childOp != nil {
		childOp.SetAttribute("query_type", "user_detail")
		childOp.SetAttribute("user_id", userID)
		time.Sleep(30 * time.Millisecond)
		childOp.SetStatus(observability.OperationStatusSuccess)
		childOp.Finish()
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    userID,
		"name":  "User " + userID,
		"email": "user" + userID + "@example.com",
		"profile": gin.H{
			"created_at": time.Now().Add(-30 * 24 * time.Hour).Unix(),
			"last_login": time.Now().Add(-2 * time.Hour).Unix(),
		},
	})
}

func createUserAdvancedHandler(c *gin.Context) {
	middleware.AddEventToRequest(c, "user_creation_started", map[string]interface{}{
		"source": "api",
	})

	// 模拟验证
	if childOp := middleware.StartChildOperation(c, "input_validation"); childOp != nil {
		time.Sleep(20 * time.Millisecond)
		childOp.SetStatus(observability.OperationStatusSuccess)
		childOp.Finish()
	}

	// 模拟数据库插入
	if childOp := middleware.StartChildOperation(c, "database_insert"); childOp != nil {
		childOp.SetAttribute("table", "users")
		time.Sleep(100 * time.Millisecond)
		childOp.SetStatus(observability.OperationStatusSuccess)
		childOp.Finish()
	}

	middleware.IncrementRequestCounter(c, "user_operations_total", map[string]string{
		"operation": "create",
		"status":    "success",
	})

	c.JSON(http.StatusCreated, gin.H{
		"id":      "new_user_123",
		"message": "User created successfully",
	})
}

func updateUserAdvancedHandler(c *gin.Context) {
	userID := c.Param("id")
	middleware.SetRequestAttribute(c, "user.id", userID)

	c.JSON(http.StatusOK, gin.H{
		"id":      userID,
		"message": "User updated successfully",
	})
}

func deleteUserAdvancedHandler(c *gin.Context) {
	userID := c.Param("id")
	middleware.SetRequestAttribute(c, "user.id", userID)

	middleware.AddEventToRequest(c, "user_deletion", map[string]interface{}{
		"user_id": userID,
		"reason":  "user_request",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

func getUserProfileHandler(c *gin.Context) {
	userID := c.Param("id")
	middleware.SetRequestAttribute(c, "user.id", userID)

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"profile": gin.H{
			"bio":      "User biography",
			"location": "City, Country",
		},
	})
}

func uploadAvatarHandler(c *gin.Context) {
	userID := c.Param("id")
	middleware.SetRequestAttribute(c, "user.id", userID)
	middleware.SetRequestAttribute(c, "operation.type", "file_upload")

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar uploaded successfully",
	})
}

func listOrdersHandler(c *gin.Context) {
	middleware.IncrementRequestCounter(c, "order_list_requests", nil)

	c.JSON(http.StatusOK, gin.H{
		"orders": []gin.H{
			{"id": "order_1", "status": "completed", "total": 99.99},
			{"id": "order_2", "status": "pending", "total": 149.99},
		},
	})
}

func getOrderDetailHandler(c *gin.Context) {
	orderID := c.Param("id")
	middleware.SetRequestAttribute(c, "order.id", orderID)

	c.JSON(http.StatusOK, gin.H{
		"id":     orderID,
		"status": "completed",
		"items":  []gin.H{{"name": "Product A", "quantity": 2}},
	})
}

func createOrderHandler(c *gin.Context) {
	middleware.AddEventToRequest(c, "order_creation", map[string]interface{}{
		"source": "web",
	})

	// 模拟订单处理
	if childOp := middleware.StartChildOperation(c, "order_processing"); childOp != nil {
		childOp.SetAttribute("processing_type", "standard")
		time.Sleep(200 * time.Millisecond)
		childOp.SetStatus(observability.OperationStatusSuccess)
		childOp.Finish()
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      "order_new_456",
		"message": "Order created successfully",
	})
}

func updateOrderStatusHandler(c *gin.Context) {
	orderID := c.Param("id")
	middleware.SetRequestAttribute(c, "order.id", orderID)

	c.JSON(http.StatusOK, gin.H{
		"id":      orderID,
		"message": "Order status updated",
	})
}

func processPaymentHandler(c *gin.Context) {
	orderID := c.Param("id")
	middleware.SetRequestAttribute(c, "order.id", orderID)
	middleware.SetRequestAttribute(c, "operation.type", "payment")

	// 模拟支付处理
	if childOp := middleware.StartChildOperation(c, "payment_processing"); childOp != nil {
		childOp.SetAttribute("payment_method", "credit_card")
		time.Sleep(500 * time.Millisecond) // 支付处理较慢
		childOp.SetStatus(observability.OperationStatusSuccess)
		childOp.Finish()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment processed successfully",
	})
}

func dashboardHandler(c *gin.Context) {
	middleware.SetRequestAttribute(c, "dashboard.type", "main")

	c.JSON(http.StatusOK, gin.H{
		"dashboard": "analytics data",
	})
}

func reportHandler(c *gin.Context) {
	reportType := c.Param("type")
	middleware.SetRequestAttribute(c, "report.type", reportType)

	c.JSON(http.StatusOK, gin.H{
		"report_type": reportType,
		"data":        "report data",
	})
}

func trackEventHandler(c *gin.Context) {
	middleware.AddEventToRequest(c, "analytics_event_tracked", nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Event tracked",
	})
}

func listFeaturesHandler(c *gin.Context) {
	middleware.SetRequestAttribute(c, "api.version", "v2")

	c.JSON(http.StatusOK, gin.H{
		"features": []string{"feature_a", "feature_b"},
	})
}

func batchOperationHandler(c *gin.Context) {
	middleware.SetRequestAttribute(c, "operation.type", "batch")

	// 模拟批处理
	if childOp := middleware.StartChildOperation(c, "batch_processing"); childOp != nil {
		childOp.SetAttribute("batch_size", 100)
		time.Sleep(1 * time.Second) // 批处理较慢
		childOp.SetStatus(observability.OperationStatusSuccess)
		childOp.Finish()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch operation completed",
	})
}

func streamDataHandler(c *gin.Context) {
	middleware.SetRequestAttribute(c, "operation.type", "stream")

	c.JSON(http.StatusOK, gin.H{
		"message": "Stream data endpoint",
	})
}

func adminStatsHandler(c *gin.Context) {
	middleware.SetRequestAttribute(c, "admin.operation", "stats")

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"active_users":    1000,
			"total_requests":  50000,
			"error_rate":      0.01,
		},
	})
}

func clearCacheHandler(c *gin.Context) {
	middleware.AddEventToRequest(c, "cache_cleared", map[string]interface{}{
		"admin_user": c.GetHeader("X-User-ID"),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}

func getConfigHandler(c *gin.Context) {
	middleware.SetRequestAttribute(c, "admin.operation", "config_view")

	c.JSON(http.StatusOK, gin.H{
		"config": gin.H{
			"version":     "2.0.0",
			"environment": "production",
		},
	})
}

// simulateAdvancedRequests 模拟复杂的HTTP请求场景
func simulateAdvancedRequests() {
	fmt.Println("\n--- 模拟高级HTTP请求场景 ---")

	client := &http.Client{Timeout: 15 * time.Second}
	baseURL := "http://localhost:8081"

	// 定义复杂的请求场景
	scenarios := []struct {
		name    string
		method  string
		path    string
		headers map[string]string
	}{
		{
			name:   "健康检查",
			method: "GET",
			path:   "/health",
			headers: map[string]string{
				"X-Client-Type": "monitoring",
			},
		},
		{
			name:   "用户列表查询",
			method: "GET",
			path:   "/api/v1/users?page=1&limit=10",
			headers: map[string]string{
				"X-User-ID":     "admin_123",
				"X-API-Version": "v1",
				"X-Region":      "us-east-1",
			},
		},
		{
			name:   "用户详情查询",
			method: "GET",
			path:   "/api/v1/users/456",
			headers: map[string]string{
				"X-User-ID":    "user_789",
				"X-Session-ID": "session_abc123",
				"X-Device-ID":  "device_xyz789",
			},
		},
		{
			name:   "创建新用户",
			method: "POST",
			path:   "/api/v1/users",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"X-User-ID":     "admin_456",
				"X-Experiment":  "new_user_flow_v2",
			},
		},
		{
			name:   "订单列表",
			method: "GET",
			path:   "/api/v1/orders",
			headers: map[string]string{
				"X-User-ID": "user_123",
			},
		},
		{
			name:   "创建订单",
			method: "POST",
			path:   "/api/v1/orders",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-User-ID":    "user_456",
			},
		},
		{
			name:   "支付处理",
			method: "POST",
			path:   "/api/v1/orders/order_123/payment",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-User-ID":    "user_789",
			},
		},
		{
			name:   "分析仪表板",
			method: "GET",
			path:   "/api/v1/analytics/dashboard",
			headers: map[string]string{
				"X-User-ID": "analyst_123",
			},
		},
		{
			name:   "批处理操作",
			method: "POST",
			path:   "/api/v2/batch",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"X-API-Version": "v2",
			},
		},
		{
			name:   "管理统计",
			method: "GET",
			path:   "/admin/stats",
			headers: map[string]string{
				"X-User-ID": "admin_root",
			},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("执行场景 %d: %s (%s %s)\n", i+1, scenario.name, scenario.method, scenario.path)

		req, err := http.NewRequest(scenario.method, baseURL+scenario.path, nil)
		if err != nil {
			fmt.Printf("创建请求失败: %v\n", err)
			continue
		}

		// 添加头部
		for key, value := range scenario.headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}
		resp.Body.Close()

		fmt.Printf("响应状态: %d\n", resp.StatusCode)
		time.Sleep(200 * time.Millisecond) // 模拟真实间隔
	}

	fmt.Println("--- 高级请求场景完成 ---")
}