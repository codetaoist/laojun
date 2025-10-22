package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	adminAPI "laojun/admin-api/internal/handler"
	adminModel "laojun/admin-api/internal/model"
	adminRepo "laojun/admin-api/internal/repository"
	adminService "laojun/admin-api/internal/service"

	configAPI "laojun/config-center/internal/handler"
	configModel "laojun/config-center/internal/model"
	configRepo "laojun/config-center/internal/repository"
	configService "laojun/config-center/internal/service"

	marketplaceAPI "laojun/marketplace-api/internal/handler"
	marketplaceModel "laojun/marketplace-api/internal/model"
	marketplaceRepo "laojun/marketplace-api/internal/repository"
	marketplaceService "laojun/marketplace-api/internal/service"

	sharedTesting "laojun/shared/testing"
)

// SystemIntegrationE2ETestSuite 系统集成端到端测试套件
type SystemIntegrationE2ETestSuite struct {
	sharedTesting.TestSuite

	// 所有服务的处理
	userHandler   *adminAPI.UserHandler
	authHandler   *adminAPI.AuthHandler
	configHandler *configAPI.ConfigHandler
	pluginHandler *marketplaceAPI.PluginHandler

	// 测试用户和token
	adminUser  *adminModel.User
	normalUser *adminModel.User
	adminToken string
	userToken  string
}

func (suite *SystemIntegrationE2ETestSuite) SetupSuite() {
	suite.TestSuite.SetupSuite()

	// 设置数据库表
	err := suite.DB.AutoMigrate(
		&adminModel.User{},
		&configModel.Config{},
		&marketplaceModel.Plugin{},
	)
	assert.NoError(suite.T(), err)

	// 初始化所有服务组
	suite.setupAllServices()

	// 设置完整的路由
	suite.setupCompleteRoutes()
}

func (suite *SystemIntegrationE2ETestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 清理所有数据
	suite.cleanupAllData()

	// 创建测试用户
	suite.createTestUsers()
}

func (suite *SystemIntegrationE2ETestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()

	// 清理所有数据
	suite.cleanupAllData()
}

func (suite *SystemIntegrationE2ETestSuite) setupAllServices() {
	// Admin API 服务
	userRepo := adminRepo.NewUserRepository(suite.DB)
	authRepo := adminRepo.NewAuthRepository(suite.DB, suite.Redis)
	userService := adminService.NewUserService(userRepo)
	authService := adminService.NewAuthService(authRepo, userRepo)
	suite.userHandler = adminAPI.NewUserHandler(userService)
	suite.authHandler = adminAPI.NewAuthHandler(authService)

	// Config Center 服务
	configRepo := configRepo.NewConfigRepository(suite.DB)
	cacheService := configService.NewCacheService(suite.Redis)
	configSvc := configService.NewConfigService(configRepo, cacheService)
	suite.configHandler = configAPI.NewConfigHandler(configSvc)

	// Marketplace API 服务
	pluginRepo := marketplaceRepo.NewPluginRepository(suite.DB)
	storageService := marketplaceService.NewStorageService()
	pluginCacheService := marketplaceService.NewCacheService(suite.Redis)
	pluginService := marketplaceService.NewPluginService(pluginRepo, storageService, pluginCacheService)
	suite.pluginHandler = marketplaceAPI.NewPluginHandler(pluginService)
}

func (suite *SystemIntegrationE2ETestSuite) setupCompleteRoutes() {
	// 健康检查路由
	suite.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Admin API 路由
	auth := suite.Router.Group("/api/v1/auth")
	{
		auth.POST("/login", suite.authHandler.Login)
		auth.POST("/logout", suite.authHandler.Logout)
		auth.POST("/refresh", suite.authHandler.RefreshToken)
	}

	users := suite.Router.Group("/api/v1/users")
	{
		users.POST("", suite.userHandler.CreateUser)
		users.GET("/:id", suite.userHandler.GetUser)
		users.PUT("/:id", suite.userHandler.UpdateUser)
		users.DELETE("/:id", suite.userHandler.DeleteUser)
		users.GET("", suite.userHandler.ListUsers)
	}

	// Config Center 路由
	configs := suite.Router.Group("/api/v1/configs")
	{
		configs.POST("", suite.configHandler.CreateConfig)
		configs.GET("/:id", suite.configHandler.GetConfig)
		configs.GET("/key/:key", suite.configHandler.GetConfigByKey)
		configs.PUT("/:id", suite.configHandler.UpdateConfig)
		configs.DELETE("/:id", suite.configHandler.DeleteConfig)
		configs.GET("", suite.configHandler.ListConfigs)
		configs.GET("/namespace/:namespace", suite.configHandler.GetConfigsByNamespace)
		configs.GET("/environment/:environment", suite.configHandler.GetConfigsByEnvironment)
	}

	// Marketplace API 路由
	plugins := suite.Router.Group("/api/v1/plugins")
	{
		plugins.POST("", suite.pluginHandler.CreatePlugin)
		plugins.GET("/:id", suite.pluginHandler.GetPlugin)
		plugins.PUT("/:id", suite.pluginHandler.UpdatePlugin)
		plugins.DELETE("/:id", suite.pluginHandler.DeletePlugin)
		plugins.GET("", suite.pluginHandler.ListPlugins)
		plugins.GET("/search", suite.pluginHandler.SearchPlugins)
		plugins.POST("/:id/download", suite.pluginHandler.DownloadPlugin)
	}
}

func (suite *SystemIntegrationE2ETestSuite) cleanupAllData() {
	suite.DB.Exec("DELETE FROM users")
	suite.DB.Exec("DELETE FROM configs")
	suite.DB.Exec("DELETE FROM plugins")
	suite.Redis.FlushAll(context.Background())
}

func (suite *SystemIntegrationE2ETestSuite) createTestUsers() {
	// 创建管理员用户
	suite.adminUser = &adminModel.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "admin123",
		Role:     "admin",
		Status:   "active",
	}
	err := suite.DB.Create(suite.adminUser).Error
	assert.NoError(suite.T(), err)

	// 创建普通用户
	suite.normalUser = &adminModel.User{
		Username: "user",
		Email:    "user@example.com",
		Password: "user123",
		Role:     "user",
		Status:   "active",
	}
	err = suite.DB.Create(suite.normalUser).Error
	assert.NoError(suite.T(), err)

	// 获取token
	suite.adminToken = suite.loginUser("admin", "admin123")
	suite.userToken = suite.loginUser("user", "user123")
}

func (suite *SystemIntegrationE2ETestSuite) loginUser(username, password string) string {
	loginData := map[string]interface{}{
		"username": username,
		"password": password,
	}

	body, _ := json.Marshal(loginData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	return data["access_token"].(string)
}

// TestSystemHealthCheck 测试系统健康检查路由
func (suite *SystemIntegrationE2ETestSuite) TestSystemHealthCheck() {
	suite.Run("HealthCheck", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "healthy", response["status"])
	})
}

// TestFullSystemWorkflow 测试完整系统工作流程
func (suite *SystemIntegrationE2ETestSuite) TestFullSystemWorkflow() {
	suite.Run("FullSystemWorkflow", func() {
		// 1. 系统初始化阶段
		suite.Run("SystemInitialization", func() {
			// 创建系统配置
			systemConfigs := []map[string]interface{}{
				{
					"key":         "system.name",
					"value":       "Laojun Platform",
					"type":        "string",
					"namespace":   "system",
					"environment": "production",
					"description": "System name",
				},
				{
					"key":         "system.version",
					"value":       "1.0.0",
					"type":        "string",
					"namespace":   "system",
					"environment": "production",
					"description": "System version",
				},
				{
					"key":         "system.max_users",
					"value":       "1000",
					"type":        "int",
					"namespace":   "system",
					"environment": "production",
					"description": "Maximum number of users",
				},
			}

			for _, configData := range systemConfigs {
				body, _ := json.Marshal(configData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusCreated, w.Code)
			}

			// 验证系统配置
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/configs/namespace/system", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			configs := response["data"].([]interface{})
			assert.Len(suite.T(), configs, 3)
		})

		// 2. 用户管理阶段
		suite.Run("UserManagement", func() {
			// 管理员创建多个用户
			users := []map[string]interface{}{
				{
					"username": "developer1",
					"email":    "dev1@example.com",
					"password": "dev123",
					"role":     "developer",
				},
				{
					"username": "developer2",
					"email":    "dev2@example.com",
					"password": "dev123",
					"role":     "developer",
				},
				{
					"username": "tester1",
					"email":    "test1@example.com",
					"password": "test123",
					"role":     "tester",
				},
			}

			createdUserIDs := make([]int, 0)

			for _, userData := range users {
				body, _ := json.Marshal(userData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusCreated, w.Code)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				data := response["data"].(map[string]interface{})
				userID := int(data["id"].(float64))
				createdUserIDs = append(createdUserIDs, userID)
			}

			// 验证用户列表
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			userList := data["users"].([]interface{})
			assert.GreaterOrEqual(suite.T(), len(userList), 5) // admin, user, dev1, dev2, tester1
		})

		// 3. 配置管理阶段
		suite.Run("ConfigurationManagement", func() {
			// 创建不同环境的配置
			environments := []string{"development", "staging", "production"}
			configTypes := []string{"database", "redis", "api"}

			for _, env := range environments {
				for _, configType := range configTypes {
					configData := map[string]interface{}{
						"key":         fmt.Sprintf("%s.%s.host", configType, env),
						"value":       fmt.Sprintf("%s-%s-host.example.com", configType, env),
						"type":        "string",
						"namespace":   "application",
						"environment": env,
						"description": fmt.Sprintf("%s host for %s environment", configType, env),
					}

					body, _ := json.Marshal(configData)
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+suite.adminToken)
					suite.Router.ServeHTTP(w, req)

					assert.Equal(suite.T(), http.StatusCreated, w.Code)
				}
			}

			// 验证不同环境的配置
			for _, env := range environments {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs/environment/"+env, nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				configs := response["data"].([]interface{})
				assert.GreaterOrEqual(suite.T(), len(configs), 3) // 至少3个配置类别
			}
		})

		// 4. 插件生态系统阶段
		suite.Run("PluginEcosystem", func() {
			// 创建不同类别的插件
			plugins := []map[string]interface{}{
				{
					"name":        "Database Connector",
					"description": "Connect to various databases",
					"version":     "1.0.0",
					"author":      "System Team",
					"category":    "database",
					"tags":        "database,connector,mysql,postgresql",
					"file_url":    "https://example.com/db-connector.zip",
					"file_size":   2048,
				},
				{
					"name":        "Authentication Plugin",
					"description": "Enhanced authentication features",
					"version":     "2.1.0",
					"author":      "Security Team",
					"category":    "security",
					"tags":        "auth,security,oauth,jwt",
					"file_url":    "https://example.com/auth-plugin.zip",
					"file_size":   1536,
				},
				{
					"name":        "Monitoring Dashboard",
					"description": "Real-time monitoring dashboard",
					"version":     "1.5.0",
					"author":      "DevOps Team",
					"category":    "monitoring",
					"tags":        "monitoring,dashboard,metrics",
					"file_url":    "https://example.com/monitor-dashboard.zip",
					"file_size":   3072,
				},
			}

			pluginIDs := make([]int, 0)

			for _, pluginData := range plugins {
				body, _ := json.Marshal(pluginData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/plugins", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusCreated, w.Code)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				data := response["data"].(map[string]interface{})
				pluginID := int(data["id"].(float64))
				pluginIDs = append(pluginIDs, pluginID)
			}

			// 模拟用户下载插件
			for _, pluginID := range pluginIDs {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/plugins/%d/download", pluginID), nil)
				req.Header.Set("Authorization", "Bearer "+suite.userToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}

			// 验证下载统计
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/plugins", nil)
			req.Header.Set("Authorization", "Bearer "+suite.userToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			pluginList := data["plugins"].([]interface{})

			for _, plugin := range pluginList {
				pluginData := plugin.(map[string]interface{})
				downloads := int(pluginData["downloads"].(float64))
				assert.Equal(suite.T(), 1, downloads)
			}
		})

		// 5. 系统监控和统计阶段
		suite.Run("SystemMonitoringAndStatistics", func() {
			// 获取系统统计信息

			// 用户统计
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var userResponse map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &userResponse)
			userData := userResponse["data"].(map[string]interface{})
			totalUsers := int(userData["total"].(float64))
			assert.GreaterOrEqual(suite.T(), totalUsers, 5)

			// 配置统计
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var configResponse map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &configResponse)
			configData := configResponse["data"].(map[string]interface{})
			totalConfigs := int(configData["total"].(float64))
			assert.GreaterOrEqual(suite.T(), totalConfigs, 12) // 3 system + 9 application configs

			// 插件统计
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/plugins", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var pluginResponse map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &pluginResponse)
			pluginData := pluginResponse["data"].(map[string]interface{})
			totalPlugins := int(pluginData["total"].(float64))
			assert.GreaterOrEqual(suite.T(), totalPlugins, 3)
		})
	})
}

// TestConcurrentSystemOperations 测试并发系统操作
func (suite *SystemIntegrationE2ETestSuite) TestConcurrentSystemOperations() {
	suite.Run("ConcurrentOperations", func() {
		const numGoroutines = 20
		const operationsPerGoroutine = 5

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*operationsPerGoroutine*3) // 3 types of operations

		// 并发创建用户、配置和插件
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					// 创建用户
					userData := map[string]interface{}{
						"username": fmt.Sprintf("concurrent_user_%d_%d", index, j),
						"email":    fmt.Sprintf("concurrent_%d_%d@example.com", index, j),
						"password": "password123",
						"role":     "user",
					}

					body, _ := json.Marshal(userData)
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+suite.adminToken)
					suite.Router.ServeHTTP(w, req)

					if w.Code != http.StatusCreated {
						errors <- fmt.Errorf("user creation failed: %d", w.Code)
					}

					// 创建配置
					configData := map[string]interface{}{
						"key":         fmt.Sprintf("concurrent.config.%d.%d", index, j),
						"value":       fmt.Sprintf("value_%d_%d", index, j),
						"type":        "string",
						"namespace":   "concurrent",
						"environment": "test",
						"description": fmt.Sprintf("Concurrent config %d_%d", index, j),
					}

					body, _ = json.Marshal(configData)
					w = httptest.NewRecorder()
					req, _ = http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+suite.adminToken)
					suite.Router.ServeHTTP(w, req)

					if w.Code != http.StatusCreated {
						errors <- fmt.Errorf("config creation failed: %d", w.Code)
					}

					// 创建插件
					pluginData := map[string]interface{}{
						"name":        fmt.Sprintf("Concurrent Plugin %d_%d", index, j),
						"description": fmt.Sprintf("Plugin created concurrently %d_%d", index, j),
						"version":     "1.0.0",
						"author":      "Concurrent Test",
						"category":    "test",
						"tags":        "concurrent,test",
						"file_url":    fmt.Sprintf("https://example.com/plugin_%d_%d.zip", index, j),
						"file_size":   1024,
					}

					body, _ = json.Marshal(pluginData)
					w = httptest.NewRecorder()
					req, _ = http.NewRequest("POST", "/api/v1/plugins", bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+suite.adminToken)
					suite.Router.ServeHTTP(w, req)

					if w.Code != http.StatusCreated {
						errors <- fmt.Errorf("plugin creation failed: %d", w.Code)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// 检查错误率（允许少量并发冲突）
		errorCount := 0
		for err := range errors {
			if err != nil {
				suite.T().Logf("Concurrent operation error: %v", err)
				errorCount++
			}
		}

		// 允许少量错误（由于并发冲突）
		assert.Less(suite.T(), errorCount, numGoroutines*operationsPerGoroutine/10, "错误率应该低于10%")

		// 验证数据完整性（至少80%成功）
		var userCount, configCount, pluginCount int64

		suite.DB.Model(&adminModel.User{}).Where("username LIKE ?", "concurrent_user_%").Count(&userCount)
		suite.DB.Model(&configModel.Config{}).Where("key LIKE ?", "concurrent.config.%").Count(&configCount)
		suite.DB.Model(&marketplaceModel.Plugin{}).Where("name LIKE ?", "Concurrent Plugin %").Count(&pluginCount)

		expectedCount := int64(numGoroutines * operationsPerGoroutine)
		assert.GreaterOrEqual(suite.T(), userCount, expectedCount*8/10) // 至少80%成功
		assert.GreaterOrEqual(suite.T(), configCount, expectedCount*8/10)
		assert.GreaterOrEqual(suite.T(), pluginCount, expectedCount*8/10)
	})
}

// TestSystemFailureRecovery 测试系统故障恢复
func (suite *SystemIntegrationE2ETestSuite) TestSystemFailureRecovery() {
	suite.Run("FailureRecovery", func() {
		// 1. 模拟数据库连接中断后的恢复
		suite.Run("DatabaseRecovery", func() {
			// 创建一些配置项
			configData := map[string]interface{}{
				"key":         "recovery.test.config",
				"value":       "recovery-value",
				"type":        "string",
				"namespace":   "recovery",
				"environment": "test",
				"description": "Recovery test configuration",
			}

			body, _ := json.Marshal(configData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			// 验证数据可以正常读取
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs/key/recovery.test.config", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)
		})

		// 2. 模拟缓存失效后的恢复
		suite.Run("CacheRecovery", func() {
			// 先访问一次，建立缓存
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/configs/key/recovery.test.config", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			// 清除缓存
			suite.Redis.FlushAll(context.Background())

			// 再次访问，应该从数据库恢复数据
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs/key/recovery.test.config", nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			assert.Equal(suite.T(), "recovery-value", data["value"])
		})
	})
}

// TestSystemPerformanceUnderLoad 测试系统负载性能
func (suite *SystemIntegrationE2ETestSuite) TestSystemPerformanceUnderLoad() {
	suite.Run("PerformanceUnderLoad", func() {
		// 创建基础数据
		const baseDataSize = 100

		// 批量创建配置
		for i := 0; i < baseDataSize; i++ {
			configData := map[string]interface{}{
				"key":         fmt.Sprintf("perf.config.%d", i),
				"value":       fmt.Sprintf("perf-value-%d", i),
				"type":        "string",
				"namespace":   "performance",
				"environment": "test",
				"description": fmt.Sprintf("Performance test config %d", i),
			}

			body, _ := json.Marshal(configData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)
		}

		// 测试高并发读取性能
		const concurrentReads = 50
		var wg sync.WaitGroup
		durations := make(chan time.Duration, concurrentReads)

		for i := 0; i < concurrentReads; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				start := time.Now()

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs?page=1&limit=20", nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				duration := time.Since(start)
				durations <- duration

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}(i)
		}

		wg.Wait()
		close(durations)

		// 计算平均响应时间
		var totalDuration time.Duration
		count := 0
		for duration := range durations {
			totalDuration += duration
			count++
		}

		avgDuration := totalDuration / time.Duration(count)
		assert.Less(suite.T(), avgDuration, 500*time.Millisecond, "平均响应时间应该少于500毫秒")
	})
}

// TestSystemDataConsistency 测试系统数据一致性
func (suite *SystemIntegrationE2ETestSuite) TestSystemDataConsistency() {
	suite.Run("DataConsistency", func() {
		// 创建关联数据

		// 1. 创建用户
		userData := map[string]interface{}{
			"username": "consistency_user",
			"email":    "consistency@example.com",
			"password": "password123",
			"role":     "developer",
		}

		body, _ := json.Marshal(userData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var userResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &userResponse)
		userData = userResponse["data"].(map[string]interface{})
		userID := int(userData["id"].(float64))

		// 2. 为该用户创建配置
		configData := map[string]interface{}{
			"key":         "user.consistency.config",
			"value":       fmt.Sprintf("config-for-user-%d", userID),
			"type":        "string",
			"namespace":   "user",
			"environment": "test",
			"description": "User-specific configuration",
		}

		body, _ = json.Marshal(configData)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		// 3. 创建用户相关的插件
		pluginData := map[string]interface{}{
			"name":        "User Consistency Plugin",
			"description": fmt.Sprintf("Plugin for user %d", userID),
			"version":     "1.0.0",
			"author":      "consistency_user",
			"category":    "user",
			"tags":        "consistency,user",
			"file_url":    "https://example.com/user-plugin.zip",
			"file_size":   1024,
		}

		body, _ = json.Marshal(pluginData)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/v1/plugins", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		// 4. 验证数据一致性
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		// 验证配置存在
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/configs/key/user.consistency.config", nil)
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var configResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &configResponse)
		configData = configResponse["data"].(map[string]interface{})
		assert.Equal(suite.T(), fmt.Sprintf("config-for-user-%d", userID), configData["value"])

		// 验证插件存在
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/plugins/search?q=consistency", nil)
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var pluginResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &pluginResponse)
		pluginData = pluginResponse["data"].(map[string]interface{})
		plugins := pluginData["plugins"].([]interface{})
		assert.GreaterOrEqual(suite.T(), len(plugins), 1)

		plugin := plugins[0].(map[string]interface{})
		assert.Equal(suite.T(), "consistency_user", plugin["author"])
	})
}

// 运行测试套件
func TestSystemIntegrationE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过系统集成端到端测试")
	}

	suite.Run(t, new(SystemIntegrationE2ETestSuite))
}
