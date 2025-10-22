package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	configAPI "laojun/config-center/internal/handler"
	configModel "laojun/config-center/internal/model"
	configRepo "laojun/config-center/internal/repository"
	configService "laojun/config-center/internal/service"
	sharedTesting "laojun/shared/testing"
)

// ConfigCenterIntegrationTestSuite Config Center 集成测试套件
type ConfigCenterIntegrationTestSuite struct {
	sharedTesting.TestSuite
	configHandler *configAPI.ConfigHandler
	httpHelper    *sharedTesting.HTTPTestHelper
	testConfig    *configModel.Config
}

func (suite *ConfigCenterIntegrationTestSuite) SetupSuite() {
	suite.TestSuite.SetupSuite()
	
	// 设置数据库表
	err := suite.DB.AutoMigrate(&configModel.Config{})
	assert.NoError(suite.T(), err)
	
	// 创建仓库和服务
	configRepo := configRepo.NewConfigRepository(suite.DB)
	cacheService := configService.NewCacheService(suite.Redis)
	configSvc := configService.NewConfigService(configRepo, cacheService)
	
	// 创建处理
	suite.configHandler = configAPI.NewConfigHandler(configSvc)
	
	// 设置路由
	suite.setupRoutes()
	
	// 创建HTTP测试助手
	suite.httpHelper = sharedTesting.NewHTTPTestHelper(suite.Router)
}

func (suite *ConfigCenterIntegrationTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()
	
	// 清理数据库和缓存
	suite.DB.Exec("DELETE FROM configs")
	suite.Redis.FlushAll(context.Background())
	
	// 创建测试配置
	suite.createTestConfig()
}

func (suite *ConfigCenterIntegrationTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()
	
	// 清理数据库和缓存
	suite.DB.Exec("DELETE FROM configs")
	suite.Redis.FlushAll(context.Background())
}

func (suite *ConfigCenterIntegrationTestSuite) setupRoutes() {
	// 配置相关路由
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
}

func (suite *ConfigCenterIntegrationTestSuite) createTestConfig() {
	suite.testConfig = &configModel.Config{
		Key:         "test.database.host",
		Value:       "localhost",
		Type:        "string",
		Namespace:   "default",
		Environment: "development",
		Description: "Database host configuration",
		IsEncrypted: false,
		Version:     1,
		Status:      "active",
	}
	
	// 直接插入数据到数据库
	err := suite.DB.Create(suite.testConfig).Error
	assert.NoError(suite.T(), err)
}

// TestConfigCRUD 测试配置CRUD操作
func (suite *ConfigCenterIntegrationTestSuite) TestConfigCRUD() {
	// 1. 创建配置
	suite.Run("CreateConfig", func() {
		configData := map[string]interface{}{
			"key":         "test.redis.host",
			"value":       "redis://localhost:6379",
			"type":        "string",
			"namespace":   "default",
			"environment": "development",
			"description": "Redis host configuration",
		}
		
		body, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "test.redis.host", data["key"])
		assert.Equal(suite.T(), "redis://localhost:6379", data["value"])
		assert.Equal(suite.T(), "string", data["type"])
		assert.Equal(suite.T(), "default", data["namespace"])
		assert.Equal(suite.T(), "development", data["environment"])
		
		// 验证数据库中的数据
		var config configModel.Config
		err := suite.DB.Where("key = ?", "test.redis.host").First(&config).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "test.redis.host", config.Key)
		assert.Equal(suite.T(), "redis://localhost:6379", config.Value)
	})
	
	// 2. 获取配置
	suite.Run("GetConfig", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/configs/%d", suite.testConfig.ID), nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "test.database.host", data["key"])
		assert.Equal(suite.T(), "localhost", data["value"])
	})
	
	// 3. 通过key获取配置
	suite.Run("GetConfigByKey", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/key/test.database.host", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "test.database.host", data["key"])
		assert.Equal(suite.T(), "localhost", data["value"])
	})
	
	// 4. 更新配置
	suite.Run("UpdateConfig", func() {
		updateData := map[string]interface{}{
			"value":       "updated-localhost",
			"description": "Updated database host configuration",
		}
		
		body, _ := json.Marshal(updateData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/configs/%d", suite.testConfig.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "updated-localhost", data["value"])
		assert.Equal(suite.T(), "Updated database host configuration", data["description"])
		
		// 验证数据库中的数据
		var config configModel.Config
		err := suite.DB.First(&config, suite.testConfig.ID).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "updated-localhost", config.Value)
		assert.Equal(suite.T(), "Updated database host configuration", config.Description)
		assert.Equal(suite.T(), uint(2), config.Version) // 版本应该递增
	})
	
	// 5. 配置列表
	suite.Run("ListConfigs", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs?page=1&limit=10", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		data := response["data"].(map[string]interface{})
		configs := data["configs"].([]interface{})
		assert.GreaterOrEqual(suite.T(), len(configs), 1)
		
		total := data["total"].(float64)
		assert.GreaterOrEqual(suite.T(), int(total), 1)
	})
	
	// 6. 按命名空间获取配置
	suite.Run("GetConfigsByNamespace", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/namespace/default", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		data := response["data"].([]interface{})
		assert.GreaterOrEqual(suite.T(), len(data), 1)
		
		config := data[0].(map[string]interface{})
		assert.Equal(suite.T(), "default", config["namespace"])
	})
	
	// 7. 按环境获取配置
	suite.Run("GetConfigsByEnvironment", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/environment/development", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		data := response["data"].([]interface{})
		assert.GreaterOrEqual(suite.T(), len(data), 1)
		
		config := data[0].(map[string]interface{})
		assert.Equal(suite.T(), "development", config["environment"])
	})
	
	// 8. 删除配置
	suite.Run("DeleteConfig", func() {
		// 创建一个要删除的配置
		deleteConfig := &configModel.Config{
			Key:         "test.delete.config",
			Value:       "delete-value",
			Type:        "string",
			Namespace:   "default",
			Environment: "development",
			Description: "Config to be deleted",
			Version:     1,
			Status:      "active",
		}
		err := suite.DB.Create(deleteConfig).Error
		assert.NoError(suite.T(), err)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/configs/%d", deleteConfig.ID), nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
		
		// 验证配置已被删除
		var config configModel.Config
		err = suite.DB.First(&config, deleteConfig.ID).Error
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
	})
}

// TestCaching 测试缓存功能
func (suite *ConfigCenterIntegrationTestSuite) TestCaching() {
	// 1. 测试缓存写入
	suite.Run("CacheWrite", func() {
		// 第一次请求，应该从数据库读取并写入缓存
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/key/test.database.host", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// 验证缓存中有数据
		cacheKey := fmt.Sprintf("config:key:%s", "test.database.host")
		exists := suite.Redis.Exists(context.Background(), cacheKey).Val()
		assert.Equal(suite.T(), int64(1), exists)
	})
	
	// 2. 测试缓存读取
	suite.Run("CacheRead", func() {
		// 先设置缓存数据
		cacheKey := fmt.Sprintf("config:key:%s", "test.database.host")
		cacheValue := `{"id":1,"key":"test.database.host","value":"cached-value","type":"string"}`
		suite.Redis.Set(context.Background(), cacheKey, cacheValue, time.Hour)
		
		// 请求应该从缓存读取数据
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/key/test.database.host", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "cached-value", data["value"])
	})
	
	// 3. 测试缓存失效
	suite.Run("CacheInvalidation", func() {
		// 先确保缓存中有数据
		cacheKey := fmt.Sprintf("config:key:%s", "test.database.host")
		suite.Redis.Set(context.Background(), cacheKey, "cached-data", time.Hour)
		
		// 更新配置，应该清除缓存数据		updateData := map[string]interface{}{
			"value": "new-value",
		}
		
		body, _ := json.Marshal(updateData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/configs/%d", suite.testConfig.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// 验证缓存已被清除
		exists := suite.Redis.Exists(context.Background(), cacheKey).Val()
		assert.Equal(suite.T(), int64(0), exists)
	})
}

// TestValidation 测试输入验证
func (suite *ConfigCenterIntegrationTestSuite) TestValidation() {
	suite.Run("CreateConfigValidation", func() {
		testCases := []struct {
			name           string
			configData     map[string]interface{}
			expectedStatus int
			expectedError  string
		}{
			{
				name: "缺少配置键",
				configData: map[string]interface{}{
					"value":       "test-value",
					"type":        "string",
					"namespace":   "default",
					"environment": "development",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "配置键不能为空",
			},
			{
				name: "缺少配置值",	
				configData: map[string]interface{}{
					"key":         "test.missing.value",
					"type":        "string",
					"namespace":   "default",
					"environment": "development",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "配置值不能为空",
			},
			{
				name: "无效的配置类",
				configData: map[string]interface{}{
					"key":         "test.invalid.type",
					"value":       "test-value",
					"type":        "invalid-type",
					"namespace":   "default",
					"environment": "development",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "无效的配置类",
			},
			{
				name: "重复的配置键",
				configData: map[string]interface{}{
					"key":         "test.database.host", // 已存在的配置键
					"value":       "another-value",
					"type":        "string",
					"namespace":   "default",
					"environment": "development",
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "配置键已存在",
			},
		}
		
		for _, tc := range testCases {
			suite.Run(tc.name, func() {
				body, _ := json.Marshal(tc.configData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				suite.Router.ServeHTTP(w, req)
				
				assert.Equal(suite.T(), tc.expectedStatus, w.Code)
				
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			})
		}
	})
}

// TestConfigTypes 测试不同配置类型
func (suite *ConfigCenterIntegrationTestSuite) TestConfigTypes() {
	testCases := []struct {
		name       string
		configType string
		value      interface{}
		expected   interface{}
	}{
		{
			name:       "字符串类型",
			configType: "string",
			value:      "test-string-value",
			expected:   "test-string-value",
		},
		{
			name:       "整数类型",
			configType: "int",
			value:      "123",
			expected:   "123",
		},
		{
			name:       "布尔类型",
			configType: "bool",
			value:      "true",
			expected:   "true",
		},
		{
			name:       "JSON类型",
			configType: "json",
			value:      `{"key": "value", "number": 123}`,
			expected:   `{"key": "value", "number": 123}`,
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			configData := map[string]interface{}{
				"key":         fmt.Sprintf("test.type.%s", tc.configType),
				"value":       tc.value,
				"type":        tc.configType,
				"namespace":   "default",
				"environment": "development",
				"description": fmt.Sprintf("Test %s type config", tc.configType),
			}
			
			// 创建配置
			body, _ := json.Marshal(configData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			suite.Router.ServeHTTP(w, req)
			
			assert.Equal(suite.T(), http.StatusCreated, w.Code)
			
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			assert.Equal(suite.T(), tc.expected, data["value"])
			assert.Equal(suite.T(), tc.configType, data["type"])
		})
	}
}

// TestConcurrency 测试并发操作
func (suite *ConfigCenterIntegrationTestSuite) TestConcurrency() {
	suite.Run("ConcurrentConfigCreation", func() {
		const numGoroutines = 10
		results := make(chan error, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				configData := map[string]interface{}{
					"key":         fmt.Sprintf("concurrent.config.%d", index),
					"value":       fmt.Sprintf("value-%d", index),
					"type":        "string",
					"namespace":   "default",
					"environment": "development",
					"description": fmt.Sprintf("Concurrent config %d", index),
				}
				
				body, _ := json.Marshal(configData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				suite.Router.ServeHTTP(w, req)
				
				if w.Code != http.StatusCreated {
					results <- fmt.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
				} else {
					results <- nil
				}
			}(i)
		}
		
		// 等待所有goroutine完成
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(suite.T(), err)
		}
		
		// 验证所有配置都被创建
		var count int64
		suite.DB.Model(&configModel.Config{}).Where("key LIKE ?", "concurrent.config.%").Count(&count)
		assert.Equal(suite.T(), int64(numGoroutines), count)
	})
}

// TestErrorHandling 测试错误处理
func (suite *ConfigCenterIntegrationTestSuite) TestErrorHandling() {
	// 1. 访问不存在的配置
	suite.Run("ConfigNotFound", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/99999", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(suite.T(), response["error"].(string), "配置不存在")
	})
	
	// 2. 通过不存在的key获取配置
	suite.Run("ConfigKeyNotFound", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/key/nonexistent.key", nil)
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(suite.T(), response["error"].(string), "配置不存在")
	})
	
	// 3. 无效的JSON格式
	suite.Run("InvalidJSON", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		suite.Router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})
}

// TestPerformance 测试性能
func (suite *ConfigCenterIntegrationTestSuite) TestPerformance() {
	suite.Run("ConfigListPerformance", func() {
		// 创建大量测试数据
		const numConfigs = 1000
		configs := make([]*configModel.Config, numConfigs)
		
		for i := 0; i < numConfigs; i++ {
			configs[i] = &configModel.Config{
				Key:         fmt.Sprintf("perf.config.%d", i),
				Value:       fmt.Sprintf("value-%d", i),
				Type:        "string",
				Namespace:   "default",
				Environment: "development",
				Description: fmt.Sprintf("Performance test config %d", i),
				Version:     1,
				Status:      "active",
			}
		}
		
		// 批量插入
		err := suite.DB.CreateInBatches(configs, 100).Error
		assert.NoError(suite.T(), err)
		
		// 测试查询性能
		start := time.Now()
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs?page=1&limit=50", nil)
		suite.Router.ServeHTTP(w, req)
		
		duration := time.Since(start)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		assert.Less(suite.T(), duration, 1*time.Second, "查询时间应该少于1秒")
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		configs_result := data["configs"].([]interface{})
		assert.Len(suite.T(), configs_result, 50)
	})
	
	suite.Run("CachePerformance", func() {
		// 测试缓存性能
		const numRequests = 100
		
		// 预热缓存
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configs/key/test.database.host", nil)
		suite.Router.ServeHTTP(w, req)
		
		// 测试缓存读取性能
		start := time.Now()
		
		for i := 0; i < numRequests; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/configs/key/test.database.host", nil)
			suite.Router.ServeHTTP(w, req)
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		}
		
		duration := time.Since(start)
		avgDuration := duration / numRequests
		
		assert.Less(suite.T(), avgDuration, 10*time.Millisecond, "平均缓存读取时间应该少于10毫秒")
	})
}

// 运行测试套件
func TestConfigCenterIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}
	
	suite.Run(t, new(ConfigCenterIntegrationTestSuite))
}
