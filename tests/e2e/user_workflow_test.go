package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// UserWorkflowE2ETestSuite 用户工作流端到端测试套件
type UserWorkflowE2ETestSuite struct {
	sharedTesting.TestSuite

	// Admin API 组件
	userHandler *adminAPI.UserHandler
	authHandler *adminAPI.AuthHandler

	// Config Center 组件
	configHandler *configAPI.ConfigHandler

	// Marketplace API 组件
	pluginHandler *marketplaceAPI.PluginHandler

	// 测试数据
	testUser   *adminModel.User
	authToken  string
	testConfig *configModel.Config
	testPlugin *marketplaceModel.Plugin
}

func (suite *UserWorkflowE2ETestSuite) SetupSuite() {
	suite.TestSuite.SetupSuite()

	// 设置数据库表
	err := suite.DB.AutoMigrate(
		&adminModel.User{},
		&configModel.Config{},
		&marketplaceModel.Plugin{},
	)
	assert.NoError(suite.T(), err)

	// 初始化 Admin API 组件
	suite.setupAdminAPI()

	// 初始化 Config Center 组件
	suite.setupConfigCenter()

	// 初始化 Marketplace API 组件
	suite.setupMarketplaceAPI()

	// 设置路由
	suite.setupRoutes()
}

func (suite *UserWorkflowE2ETestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 清理数据库和缓存
	suite.DB.Exec("DELETE FROM users")
	suite.DB.Exec("DELETE FROM configs")
	suite.DB.Exec("DELETE FROM plugins")
	suite.Redis.FlushAll(context.Background())

	// 创建测试数据
	suite.createTestData()
}

func (suite *UserWorkflowE2ETestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()

	// 清理数据库和缓存
	suite.DB.Exec("DELETE FROM users")
	suite.DB.Exec("DELETE FROM configs")
	suite.DB.Exec("DELETE FROM plugins")
	suite.Redis.FlushAll(context.Background())
}

func (suite *UserWorkflowE2ETestSuite) setupAdminAPI() {
	userRepo := adminRepo.NewUserRepository(suite.DB)
	authRepo := adminRepo.NewAuthRepository(suite.DB, suite.Redis)

	userService := adminService.NewUserService(userRepo)
	authService := adminService.NewAuthService(authRepo, userRepo)

	suite.userHandler = adminAPI.NewUserHandler(userService)
	suite.authHandler = adminAPI.NewAuthHandler(authService)
}

func (suite *UserWorkflowE2ETestSuite) setupConfigCenter() {
	configRepo := configRepo.NewConfigRepository(suite.DB)
	cacheService := configService.NewCacheService(suite.Redis)
	configSvc := configService.NewConfigService(configRepo, cacheService)

	suite.configHandler = configAPI.NewConfigHandler(configSvc)
}

func (suite *UserWorkflowE2ETestSuite) setupMarketplaceAPI() {
	pluginRepo := marketplaceRepo.NewPluginRepository(suite.DB)
	storageService := marketplaceService.NewStorageService()
	cacheService := marketplaceService.NewCacheService(suite.Redis)
	pluginService := marketplaceService.NewPluginService(pluginRepo, storageService, cacheService)

	suite.pluginHandler = marketplaceAPI.NewPluginHandler(pluginService)
}

func (suite *UserWorkflowE2ETestSuite) setupRoutes() {
	// Admin API 路由
	auth := suite.Router.Group("/api/v1/auth")
	{
		auth.POST("/login", suite.authHandler.Login)
		auth.POST("/logout", suite.authHandler.Logout)
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

func (suite *UserWorkflowE2ETestSuite) createTestData() {
	// 创建测试用户
	suite.testUser = &adminModel.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "admin",
		Status:   "active",
	}
	err := suite.DB.Create(suite.testUser).Error
	assert.NoError(suite.T(), err)

	// 登录获取token
	suite.loginAndGetToken()

	// 创建测试配置
	suite.testConfig = &configModel.Config{
		Key:         "test.app.name",
		Value:       "Laojun Platform",
		Type:        "string",
		Namespace:   "default",
		Environment: "development",
		Description: "Application name configuration",
		Version:     1,
		Status:      "active",
	}
	err = suite.DB.Create(suite.testConfig).Error
	assert.NoError(suite.T(), err)

	// 创建测试插件
	suite.testPlugin = &marketplaceModel.Plugin{
		Name:        "Test Plugin",
		Description: "A test plugin for e2e testing",
		Version:     "1.0.0",
		Author:      "Test Author",
		Category:    "utility",
		Tags:        "test,utility",
		Status:      "published",
		Downloads:   0,
		Rating:      0.0,
		FileSize:    1024,
		FileURL:     "https://example.com/test-plugin.zip",
	}
	err = suite.DB.Create(suite.testPlugin).Error
	assert.NoError(suite.T(), err)
}

func (suite *UserWorkflowE2ETestSuite) loginAndGetToken() {
	loginData := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
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
	suite.authToken = data["access_token"].(string)
}

// TestCompleteUserWorkflow 测试完整的用户工作流
func (suite *UserWorkflowE2ETestSuite) TestCompleteUserWorkflow() {
	suite.Run("CompleteWorkflow", func() {
		// 1. 用户注册和登录
		suite.Run("UserRegistrationAndLogin", func() {
			// 创建新用户
			userData := map[string]interface{}{
				"username": "newuser",
				"email":    "newuser@example.com",
				"password": "password123",
				"role":     "user",
			}

			body, _ := json.Marshal(userData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(suite.T(), "success", response["status"])

			// 新用户登录
			loginData := map[string]interface{}{
				"username": "newuser",
				"password": "password123",
			}

			body, _ = json.Marshal(loginData)
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(suite.T(), "success", response["status"])

			data := response["data"].(map[string]interface{})
			newUserToken := data["access_token"].(string)
			assert.NotEmpty(suite.T(), newUserToken)
		})

		// 2. 配置管理工作流
		suite.Run("ConfigurationManagement", func() {
			// 创建配置
			configData := map[string]interface{}{
				"key":         "user.workflow.test",
				"value":       "test-value",
				"type":        "string",
				"namespace":   "default",
				"environment": "development",
				"description": "User workflow test configuration",
			}

			body, _ := json.Marshal(configData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			configID := int(data["id"].(float64))

			// 读取配置
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs/key/user.workflow.test", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			json.Unmarshal(w.Body.Bytes(), &response)
			data = response["data"].(map[string]interface{})
			assert.Equal(suite.T(), "test-value", data["value"])

			// 更新配置
			updateData := map[string]interface{}{
				"value": "updated-test-value",
			}

			body, _ = json.Marshal(updateData)
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/v1/configs/%d", configID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			// 验证更新
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs/key/user.workflow.test", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			json.Unmarshal(w.Body.Bytes(), &response)
			data = response["data"].(map[string]interface{})
			assert.Equal(suite.T(), "updated-test-value", data["value"])
		})

		// 3. 插件管理工作流
		suite.Run("PluginManagement", func() {
			// 创建插件
			pluginData := map[string]interface{}{
				"name":        "User Workflow Plugin",
				"description": "A plugin created during user workflow test",
				"version":     "1.0.0",
				"author":      "Test User",
				"category":    "utility",
				"tags":        "test,workflow",
				"file_url":    "https://example.com/workflow-plugin.zip",
				"file_size":   2048,
			}

			body, _ := json.Marshal(pluginData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/plugins", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			pluginID := int(data["id"].(float64))

			// 搜索插件
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/plugins/search?q=workflow", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			json.Unmarshal(w.Body.Bytes(), &response)
			data = response["data"].(map[string]interface{})
			plugins := data["plugins"].([]interface{})
			assert.GreaterOrEqual(suite.T(), len(plugins), 1)

			// 下载插件
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/plugins/%d/download", pluginID), nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			// 验证下载次数增加
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/plugins/%d", pluginID), nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			json.Unmarshal(w.Body.Bytes(), &response)
			data = response["data"].(map[string]interface{})
			downloads := int(data["downloads"].(float64))
			assert.Equal(suite.T(), 1, downloads)
		})

		// 4. 跨服务数据一致性验证
		suite.Run("CrossServiceDataConsistency", func() {
			// 验证用户数据在所有服务中的一致性
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			users := data["users"].([]interface{})
			assert.GreaterOrEqual(suite.T(), len(users), 2) // 至少有testuser和newuser

			// 验证配置数据
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			json.Unmarshal(w.Body.Bytes(), &response)
			data = response["data"].(map[string]interface{})
			configs := data["configs"].([]interface{})
			assert.GreaterOrEqual(suite.T(), len(configs), 2) // 至少有testConfig和新创建的配置

			// 验证插件数据
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/plugins", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			json.Unmarshal(w.Body.Bytes(), &response)
			data = response["data"].(map[string]interface{})
			plugins := data["plugins"].([]interface{})
			assert.GreaterOrEqual(suite.T(), len(plugins), 2) // 至少有testPlugin和新创建的插件
		})

		// 5. 用户登出
		suite.Run("UserLogout", func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(suite.T(), "success", response["status"])

			// 验证token已失效
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
		})
	})
}

// TestErrorRecoveryWorkflow 测试错误恢复工作流
func (suite *UserWorkflowE2ETestSuite) TestErrorRecoveryWorkflow() {
	suite.Run("ErrorRecovery", func() {
		// 1. 模拟网络错误后的重试
		suite.Run("NetworkErrorRetry", func() {
			// 尝试访问不存在的资源
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/configs/99999", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusNotFound, w.Code)

			// 创建该资源以触发错误恢复
			configData := map[string]interface{}{
				"key":         "error.recovery.test",
				"value":       "recovery-value",
				"type":        "string",
				"namespace":   "default",
				"environment": "development",
				"description": "Error recovery test configuration",
			}

			body, _ := json.Marshal(configData)
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			// 重试访问，现在应该成功
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs/key/error.recovery.test", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)
		})

		// 2. 模拟数据冲突后的处理
		suite.Run("DataConflictHandling", func() {
			// 尝试创建重复的配置
			w := httptest.NewRecorder()
			configData := map[string]interface{}{
				"key":         "test.app.name", // 已存在的key
				"value":       "duplicate-value",
				"type":        "string",
				"namespace":   "default",
				"environment": "development",
			}

			body, _ := json.Marshal(configData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusConflict, w.Code)

			// 使用不同的key重新创建
			configData["key"] = "test.app.name.duplicate"
			body, _ = json.Marshal(configData)
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)
		})
	})
}

// TestPerformanceWorkflow 测试性能相关工作流
func (suite *UserWorkflowE2ETestSuite) TestPerformanceWorkflow() {
	suite.Run("PerformanceWorkflow", func() {
		// 1. 批量操作性能测试
		suite.Run("BatchOperations", func() {
			const batchSize = 50

			// 批量创建配置
			start := time.Now()
			for i := 0; i < batchSize; i++ {
				configData := map[string]interface{}{
					"key":         fmt.Sprintf("batch.config.%d", i),
					"value":       fmt.Sprintf("batch-value-%d", i),
					"type":        "string",
					"namespace":   "batch",
					"environment": "test",
					"description": fmt.Sprintf("Batch config %d", i),
				}

				body, _ := json.Marshal(configData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.authToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusCreated, w.Code)
			}
			duration := time.Since(start)

			avgTime := duration / batchSize
			assert.Less(suite.T(), avgTime, 100*time.Millisecond, "平均创建时间应该少于100毫秒")

			// 批量查询性能测试
			start = time.Now()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/configs?page=1&limit=100", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			queryDuration := time.Since(start)
			assert.Equal(suite.T(), http.StatusOK, w.Code)
			assert.Less(suite.T(), queryDuration, 1*time.Second, "批量查询时间应该少于1秒")
		})

		// 2. 缓存性能测试
		suite.Run("CachePerformance", func() {
			// 第一次访问（缓存未命中）
			start := time.Now()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/configs/key/test.app.name", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)
			firstAccessTime := time.Since(start)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			// 第二次访问（缓存命中）
			start = time.Now()
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/configs/key/test.app.name", nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)
			secondAccessTime := time.Since(start)

			assert.Equal(suite.T(), http.StatusOK, w.Code)
			assert.Less(suite.T(), secondAccessTime, firstAccessTime, "缓存命中应该更快")
		})
	})
}

// TestSecurityWorkflow 测试安全相关工作流
func (suite *UserWorkflowE2ETestSuite) TestSecurityWorkflow() {
	suite.Run("SecurityWorkflow", func() {
		// 1. 未授权访问测试
		suite.Run("UnauthorizedAccess", func() {
			// 不提供token
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/users", nil)
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

			// 提供无效token
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer invalid_token")
			suite.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
		})

		// 2. 输入验证测试
		suite.Run("InputValidation", func() {
			// SQL注入尝试
			maliciousData := map[string]interface{}{
				"key":   "'; DROP TABLE configs; --",
				"value": "malicious-value",
				"type":  "string",
			}

			body, _ := json.Marshal(maliciousData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			suite.Router.ServeHTTP(w, req)

			// 应该被验证拦截或安全处理
			assert.NotEqual(suite.T(), http.StatusOK, w.Code)

			// 验证数据库表仍然存在
			var count int64
			err := suite.DB.Model(&configModel.Config{}).Count(&count).Error
			assert.NoError(suite.T(), err)
			assert.Greater(suite.T(), count, int64(0))
		})
	})
}

// 运行测试套件
func TestUserWorkflowE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端测试")
	}

	suite.Run(t, new(UserWorkflowE2ETestSuite))
}
