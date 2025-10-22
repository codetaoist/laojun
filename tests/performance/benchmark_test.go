package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
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

// PerformanceBenchmarkSuite 性能基准测试套件
type PerformanceBenchmarkSuite struct {
	sharedTesting.TestSuite

	// 服务处理�?
	userHandler   *adminAPI.UserHandler
	authHandler   *adminAPI.AuthHandler
	configHandler *configAPI.ConfigHandler
	pluginHandler *marketplaceAPI.PluginHandler

	// 测试数据
	adminToken  string
	testUsers   []*adminModel.User
	testConfigs []*configModel.Config
	testPlugins []*marketplaceModel.Plugin
}

func (suite *PerformanceBenchmarkSuite) SetupSuite() {
	suite.TestSuite.SetupSuite()

	// 设置数据库表
	err := suite.DB.AutoMigrate(
		&adminModel.User{},
		&configModel.Config{},
		&marketplaceModel.Plugin{},
	)
	assert.NoError(suite.T(), err)

	// 初始化服务
	suite.setupServices()

	// 设置路由
	suite.setupRoutes()

	// 创建测试数据
	suite.createTestData()
}

func (suite *PerformanceBenchmarkSuite) setupServices() {
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

func (suite *PerformanceBenchmarkSuite) setupRoutes() {
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
	}
}

func (suite *PerformanceBenchmarkSuite) createTestData() {
	// 创建管理员用户
	adminUser := &adminModel.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "admin123",
		Role:     "admin",
		Status:   "active",
	}
	err := suite.DB.Create(adminUser).Error
	assert.NoError(suite.T(), err)

	// 获取管理员token
	suite.adminToken = suite.loginUser("admin", "admin123")

	// 批量创建测试用户
	const userCount = 1000
	suite.testUsers = make([]*adminModel.User, 0, userCount)

	for i := 0; i < userCount; i++ {
		user := &adminModel.User{
			Username: fmt.Sprintf("user_%d", i),
			Email:    fmt.Sprintf("user_%d@example.com", i),
			Password: "password123",
			Role:     "user",
			Status:   "active",
		}
		err := suite.DB.Create(user).Error
		assert.NoError(suite.T(), err)
		suite.testUsers = append(suite.testUsers, user)
	}

	// 批量创建测试配置
	const configCount = 2000
	suite.testConfigs = make([]*configModel.Config, 0, configCount)

	for i := 0; i < configCount; i++ {
		config := &configModel.Config{
			Key:         fmt.Sprintf("test.config.%d", i),
			Value:       fmt.Sprintf("test-value-%d", i),
			Type:        "string",
			Namespace:   "test",
			Environment: "benchmark",
			Description: fmt.Sprintf("Benchmark test config %d", i),
		}
		err := suite.DB.Create(config).Error
		assert.NoError(suite.T(), err)
		suite.testConfigs = append(suite.testConfigs, config)
	}

	// 批量创建测试插件
	const pluginCount = 500
	suite.testPlugins = make([]*marketplaceModel.Plugin, 0, pluginCount)

	for i := 0; i < pluginCount; i++ {
		plugin := &marketplaceModel.Plugin{
			Name:        fmt.Sprintf("Benchmark Plugin %d", i),
			Description: fmt.Sprintf("Plugin for benchmark testing %d", i),
			Version:     "1.0.0",
			Author:      "Benchmark Team",
			Category:    "benchmark",
			Tags:        "benchmark,test,performance",
			FileURL:     fmt.Sprintf("https://example.com/plugin_%d.zip", i),
			FileSize:    1024 + int64(i),
		}
		err := suite.DB.Create(plugin).Error
		assert.NoError(suite.T(), err)
		suite.testPlugins = append(suite.testPlugins, plugin)
	}
}

func (suite *PerformanceBenchmarkSuite) loginUser(username, password string) string {
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

// BenchmarkUserOperations 用户操作基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkUserOperations() {
	suite.Run("UserOperations", func() {
		// 测试用户列表查询性能
		suite.Run("ListUsers", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/users?page=1&limit=20", nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试单个用户查询性能
		suite.Run("GetUser", func() {
			b := suite.T().(*testing.B)
			userID := suite.testUsers[0].ID
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试用户创建性能
		suite.Run("CreateUser", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				userData := map[string]interface{}{
					"username": fmt.Sprintf("benchmark_user_%d_%d", b.N, i),
					"email":    fmt.Sprintf("benchmark_%d_%d@example.com", b.N, i),
					"password": "password123",
					"role":     "user",
				}

				body, _ := json.Marshal(userData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusCreated, w.Code)
			}
		})
	})
}

// BenchmarkConfigOperations 配置操作基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkConfigOperations() {
	suite.Run("ConfigOperations", func() {
		// 测试配置列表查询性能
		suite.Run("ListConfigs", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs?page=1&limit=50", nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试配置键查询性能（缓存命中）
		suite.Run("GetConfigByKey_CacheHit", func() {
			b := suite.T().(*testing.B)
			configKey := suite.testConfigs[0].Key

			// 预热缓存
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/configs/key/"+configKey, nil)
			req.Header.Set("Authorization", "Bearer "+suite.adminToken)
			suite.Router.ServeHTTP(w, req)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs/key/"+configKey, nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试配置键查询性能（缓存未命中）
		suite.Run("GetConfigByKey_CacheMiss", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// 清除缓存确保缓存未命中
				suite.Redis.FlushAll(context.Background())

				configKey := suite.testConfigs[i%len(suite.testConfigs)].Key
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs/key/"+configKey, nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试配置创建性能
		suite.Run("CreateConfig", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				configData := map[string]interface{}{
					"key":         fmt.Sprintf("benchmark.config.%d.%d", b.N, i),
					"value":       fmt.Sprintf("benchmark-value-%d-%d", b.N, i),
					"type":        "string",
					"namespace":   "benchmark",
					"environment": "test",
					"description": fmt.Sprintf("Benchmark config %d-%d", b.N, i),
				}

				body, _ := json.Marshal(configData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusCreated, w.Code)
			}
		})
	})
}

// BenchmarkPluginOperations 插件操作基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkPluginOperations() {
	suite.Run("PluginOperations", func() {
		// 测试插件列表查询性能
		suite.Run("ListPlugins", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/plugins?page=1&limit=20", nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试插件搜索性能
		suite.Run("SearchPlugins", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/plugins/search?q=benchmark", nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试单个插件查询性能
		suite.Run("GetPlugin", func() {
			b := suite.T().(*testing.B)
			pluginID := suite.testPlugins[0].ID
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/plugins/%d", pluginID), nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试插件创建性能
		suite.Run("CreatePlugin", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				pluginData := map[string]interface{}{
					"name":        fmt.Sprintf("Benchmark Plugin %d_%d", b.N, i),
					"description": fmt.Sprintf("Plugin for benchmark %d_%d", b.N, i),
					"version":     "1.0.0",
					"author":      "Benchmark Team",
					"category":    "benchmark",
					"tags":        "benchmark,test",
					"file_url":    fmt.Sprintf("https://example.com/benchmark_%d_%d.zip", b.N, i),
					"file_size":   1024,
				}

				body, _ := json.Marshal(pluginData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/plugins", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusCreated, w.Code)
			}
		})
	})
}

// BenchmarkConcurrentOperations 并发操作基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkConcurrentOperations() {
	suite.Run("ConcurrentOperations", func() {
		// 测试并发读取性能
		suite.Run("ConcurrentReads", func() {
			b := suite.T().(*testing.B)
			const concurrency = 10

			b.ResetTimer()
			b.SetParallelism(concurrency)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					// 随机选择一个配置进行读取
					configKey := suite.testConfigs[b.N%len(suite.testConfigs)].Key

					w := httptest.NewRecorder()
					req, _ := http.NewRequest("GET", "/api/v1/configs/key/"+configKey, nil)
					req.Header.Set("Authorization", "Bearer "+suite.adminToken)
					suite.Router.ServeHTTP(w, req)

					assert.Equal(suite.T(), http.StatusOK, w.Code)
				}
			})
		})

		// 测试并发写入性能
		suite.Run("ConcurrentWrites", func() {
			b := suite.T().(*testing.B)
			const concurrency = 5
			var counter int64

			b.ResetTimer()
			b.SetParallelism(concurrency)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					counter++

					configData := map[string]interface{}{
						"key":         fmt.Sprintf("concurrent.config.%d", counter),
						"value":       fmt.Sprintf("concurrent-value-%d", counter),
						"type":        "string",
						"namespace":   "concurrent",
						"environment": "benchmark",
						"description": fmt.Sprintf("Concurrent config %d", counter),
					}

					body, _ := json.Marshal(configData)
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+suite.adminToken)
					suite.Router.ServeHTTP(w, req)

					assert.Equal(suite.T(), http.StatusCreated, w.Code)
				}
			})
		})

		// 测试混合读写性能
		suite.Run("MixedReadWrite", func() {
			b := suite.T().(*testing.B)
			const concurrency = 8
			var writeCounter int64

			b.ResetTimer()
			b.SetParallelism(concurrency)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					// 80% 读操作，20% 写操作
					if b.N%5 == 0 {
						// 写操作
						writeCounter++

						configData := map[string]interface{}{
							"key":         fmt.Sprintf("mixed.config.%d", writeCounter),
							"value":       fmt.Sprintf("mixed-value-%d", writeCounter),
							"type":        "string",
							"namespace":   "mixed",
							"environment": "benchmark",
							"description": fmt.Sprintf("Mixed config %d", writeCounter),
						}

						body, _ := json.Marshal(configData)
						w := httptest.NewRecorder()
						req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Authorization", "Bearer "+suite.adminToken)
						suite.Router.ServeHTTP(w, req)

						assert.Equal(suite.T(), http.StatusCreated, w.Code)
					} else {
						// 读操作
						configKey := suite.testConfigs[b.N%len(suite.testConfigs)].Key

						w := httptest.NewRecorder()
						req, _ := http.NewRequest("GET", "/api/v1/configs/key/"+configKey, nil)
						req.Header.Set("Authorization", "Bearer "+suite.adminToken)
						suite.Router.ServeHTTP(w, req)

						assert.Equal(suite.T(), http.StatusOK, w.Code)
					}
				}
			})
		})
	})
}

// BenchmarkMemoryUsage 内存使用基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkMemoryUsage() {
	suite.Run("MemoryUsage", func() {
		// 测试大量数据查询的内存使用
		suite.Run("LargeDataQuery", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs?page=1&limit=1000", nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)

				// 强制垃圾回收
				if i%100 == 0 {
					runtime.GC()
				}
			}
		})

		// 测试缓存内存使用
		suite.Run("CacheMemoryUsage", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// 访问不同的配置以填充缓存
				configKey := suite.testConfigs[i%len(suite.testConfigs)].Key

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs/key/"+configKey, nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})
	})
}

// BenchmarkDatabaseOperations 数据库操作基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkDatabaseOperations() {
	suite.Run("DatabaseOperations", func() {
		// 测试复杂查询性能
		suite.Run("ComplexQuery", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/configs?namespace=test&environment=benchmark&page=1&limit=100", nil)
				req.Header.Set("Authorization", "Bearer "+suite.adminToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), http.StatusOK, w.Code)
			}
		})

		// 测试批量插入性能
		suite.Run("BatchInsert", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// 批量创建配置
				var wg sync.WaitGroup
				const batchSize = 10

				for j := 0; j < batchSize; j++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						configData := map[string]interface{}{
							"key":         fmt.Sprintf("batch.config.%d.%d", i, index),
							"value":       fmt.Sprintf("batch-value-%d-%d", i, index),
							"type":        "string",
							"namespace":   "batch",
							"environment": "benchmark",
							"description": fmt.Sprintf("Batch config %d-%d", i, index),
						}

						body, _ := json.Marshal(configData)
						w := httptest.NewRecorder()
						req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Authorization", "Bearer "+suite.adminToken)
						suite.Router.ServeHTTP(w, req)

						assert.Equal(suite.T(), http.StatusCreated, w.Code)
					}(j)
				}

				wg.Wait()
			}
		})
	})
}

// BenchmarkCacheOperations 缓存操作基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkCacheOperations() {
	suite.Run("CacheOperations", func() {
		// 测试缓存写入性能
		suite.Run("CacheWrite", func() {
			b := suite.T().(*testing.B)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("benchmark:cache:write:%d", i)
				value := fmt.Sprintf("cache-value-%d", i)

				err := suite.Redis.Set(context.Background(), key, value, time.Hour).Err()
				assert.NoError(suite.T(), err)
			}
		})

		// 测试缓存读取性能
		suite.Run("CacheRead", func() {
			b := suite.T().(*testing.B)

			// 预先写入数据
			for i := 0; i < 1000; i++ {
				key := fmt.Sprintf("benchmark:cache:read:%d", i)
				value := fmt.Sprintf("cache-value-%d", i)
				suite.Redis.Set(context.Background(), key, value, time.Hour)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("benchmark:cache:read:%d", i%1000)

				_, err := suite.Redis.Get(context.Background(), key).Result()
				assert.NoError(suite.T(), err)
			}
		})

		// 测试缓存删除性能
		suite.Run("CacheDelete", func() {
			b := suite.T().(*testing.B)

			// 预先写入数据
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("benchmark:cache:delete:%d", i)
				value := fmt.Sprintf("cache-value-%d", i)
				suite.Redis.Set(context.Background(), key, value, time.Hour)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("benchmark:cache:delete:%d", i)

				err := suite.Redis.Del(context.Background(), key).Err()
				assert.NoError(suite.T(), err)
			}
		})
	})
}

// 运行性能基准测试套件
func TestPerformanceBenchmarkSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能基准测试")
	}

	suite.Run(t, new(PerformanceBenchmarkSuite))
}

// 单独的基准测试函数，用于 go test -bench
func BenchmarkUserListAPI(b *testing.B) {
	// 这里可以添加独立的基准测试代码
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟用户列表查询
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(b, http.StatusOK, w.Code)
	}
	// 用于 go test -bench=BenchmarkUserListAPI
}

func BenchmarkConfigGetAPI(b *testing.B) {
	// 这里可以添加独立的基准测试代码
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟配置查询
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/key/benchmark:config:get", nil)
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(b, http.StatusOK, w.Code)
	}
	// 用于 go test -bench=BenchmarkConfigGetAPI
}

func BenchmarkPluginSearchAPI(b *testing.B) {
	// 这里可以添加独立的基准测试代码
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟插件搜索查询
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/plugins/search?name=benchmark", nil)
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(b, http.StatusOK, w.Code)
	}
	// 用于 go test -bench=BenchmarkPluginSearchAPI
}
