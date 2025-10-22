package handler

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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun/config-center/internal/model"
	"github.com/codetaoist/laojun/config-center/internal/service"
	sharedTesting "github.com/codetaoist/laojun/shared/testing"
)

// MockConfigService 模拟配置服务
type MockConfigService struct {
	mock.Mock
}

func (m *MockConfigService) CreateConfig(ctx context.Context, req *service.CreateConfigRequest) (*model.Config, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Config), args.Error(1)
}

func (m *MockConfigService) GetConfig(ctx context.Context, id uint) (*model.Config, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Config), args.Error(1)
}

func (m *MockConfigService) GetConfigByKey(ctx context.Context, key string) (*model.Config, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Config), args.Error(1)
}

func (m *MockConfigService) UpdateConfig(ctx context.Context, id uint, req *service.UpdateConfigRequest) (*model.Config, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Config), args.Error(1)
}

func (m *MockConfigService) DeleteConfig(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConfigService) ListConfigs(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.Config, int64, error) {
	args := m.Called(ctx, filters, offset, limit)
	return args.Get(0).([]*model.Config), args.Get(1).(int64), args.Error(2)
}

func (m *MockConfigService) GetConfigsByNamespace(ctx context.Context, namespace string) ([]*model.Config, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]*model.Config), args.Error(1)
}

func (m *MockConfigService) GetConfigsByEnvironment(ctx context.Context, environment string) ([]*model.Config, error) {
	args := m.Called(ctx, environment)
	return args.Get(0).([]*model.Config), args.Error(1)
}

// ConfigHandlerTestSuite 配置处理器测试套
type ConfigHandlerTestSuite struct {
	sharedTesting.TestSuite
	configService *MockConfigService
	handler       *ConfigHandler
	httpHelper    *sharedTesting.HTTPTestHelper
}

func (suite *ConfigHandlerTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 创建模拟服务
	suite.configService = new(MockConfigService)

	// 创建处理器
	suite.handler = &ConfigHandler{
		configService: suite.configService,
	}

	// 设置路由
	suite.Router.POST("/configs", suite.handler.CreateConfig)
	suite.Router.GET("/configs/:id", suite.handler.GetConfig)
	suite.Router.GET("/configs/key/:key", suite.handler.GetConfigByKey)
	suite.Router.PUT("/configs/:id", suite.handler.UpdateConfig)
	suite.Router.DELETE("/configs/:id", suite.handler.DeleteConfig)
	suite.Router.GET("/configs", suite.handler.ListConfigs)
	suite.Router.GET("/configs/namespace/:namespace", suite.handler.GetConfigsByNamespace)
	suite.Router.GET("/configs/environment/:environment", suite.handler.GetConfigsByEnvironment)

	// 创建HTTP测试助手
	suite.httpHelper = sharedTesting.NewHTTPTestHelper(suite.Router)
}

func (suite *ConfigHandlerTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()
	suite.configService.AssertExpectations(suite.T())
}

// TestCreateConfig 测试创建配置
func (suite *ConfigHandlerTestSuite) TestCreateConfig() {
	testCases := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功创建配置",
			requestBody: map[string]interface{}{
				"key":         "app.database.host",
				"value":       "localhost",
				"type":        "string",
				"namespace":   "default",
				"environment": "development",
				"description": "数据库主机地址",
			},
			setupMocks: func() {
				req := &service.CreateConfigRequest{
					Key:         "app.database.host",
					Value:       "localhost",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
					Description: "数据库主机地址",
				}
				config := &model.Config{
					ID:          1,
					Key:         "app.database.host",
					Value:       "localhost",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
					Description: "数据库主机地址",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				suite.configService.On("CreateConfig", mock.Anything, req).Return(config, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "无效的请求体",
			requestBody: map[string]interface{}{
				"key": "", // 空键
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "配置键不能为空",
		},
		{
			name: "配置键已存在",
			requestBody: map[string]interface{}{
				"key":         "existing.key",
				"value":       "value",
				"type":        "string",
				"namespace":   "default",
				"environment": "development",
			},
			setupMocks: func() {
				req := &service.CreateConfigRequest{
					Key:         "existing.key",
					Value:       "value",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
				}
				suite.configService.On("CreateConfig", mock.Anything, req).Return((*model.Config)(nil), fmt.Errorf("配置键已存在"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "配置键已存在",
		},
		{
			name:           "无效的JSON",
			requestBody:    "invalid json",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configService.ExpectedCalls = nil
			suite.configService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 准备请求体
			var body []byte
			if str, ok := tc.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tc.requestBody)
			}

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			suite.Router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			} else if tc.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])
			}
		})
	}
}

// TestGetConfig 测试获取配置
func (suite *ConfigHandlerTestSuite) TestGetConfig() {
	testCases := []struct {
		name           string
		configID       string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "成功获取配置",
			configID: "1",
			setupMocks: func() {
				config := &model.Config{
					ID:          1,
					Key:         "app.database.host",
					Value:       "localhost",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
					Description: "数据库主机地址",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				suite.configService.On("GetConfig", mock.Anything, uint(1)).Return(config, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "配置不存在",
			configID: "999",
			setupMocks: func() {
				suite.configService.On("GetConfig", mock.Anything, uint(999)).Return((*model.Config)(nil), gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "配置不存在",
		},
		{
			name:           "无效的配置ID",
			configID:       "invalid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的配置ID",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configService.ExpectedCalls = nil
			suite.configService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/configs/"+tc.configID, nil)
			suite.Router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			} else if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])
			}
		})
	}
}

// TestGetConfigByKey 测试通过键获取配置
func (suite *ConfigHandlerTestSuite) TestGetConfigByKey() {
	testCases := []struct {
		name           string
		configKey      string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "成功获取配置",
			configKey: "app.database.host",
			setupMocks: func() {
				config := &model.Config{
					ID:          1,
					Key:         "app.database.host",
					Value:       "localhost",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
				}
				suite.configService.On("GetConfigByKey", mock.Anything, "app.database.host").Return(config, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "配置不存在",
			configKey: "nonexistent.key",
			setupMocks: func() {
				suite.configService.On("GetConfigByKey", mock.Anything, "nonexistent.key").Return((*model.Config)(nil), gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "配置不存在",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configService.ExpectedCalls = nil
			suite.configService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/configs/key/"+tc.configKey, nil)
			suite.Router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			} else if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])
			}
		})
	}
}

// TestUpdateConfig 测试更新配置
func (suite *ConfigHandlerTestSuite) TestUpdateConfig() {
	testCases := []struct {
		name           string
		configID       string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "成功更新配置",
			configID: "1",
			requestBody: map[string]interface{}{
				"value":       "new_value",
				"description": "更新后的描述",
			},
			setupMocks: func() {
				req := &service.UpdateConfigRequest{
					Value:       "new_value",
					Description: "更新后的描述",
				}
				config := &model.Config{
					ID:          1,
					Key:         "app.database.host",
					Value:       "new_value",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
					Description: "更新后的描述",
					UpdatedAt:   time.Now(),
				}
				suite.configService.On("UpdateConfig", mock.Anything, uint(1), req).Return(config, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "配置不存在",
			configID: "999",
			requestBody: map[string]interface{}{
				"value": "new_value",
			},
			setupMocks: func() {
				req := &service.UpdateConfigRequest{
					Value: "new_value",
				}
				suite.configService.On("UpdateConfig", mock.Anything, uint(999), req).Return((*model.Config)(nil), gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "配置不存在",
		},
		{
			name:           "无效的配置ID",
			configID:       "invalid",
			requestBody:    map[string]interface{}{"value": "new_value"},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的配置ID",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configService.ExpectedCalls = nil
			suite.configService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 准备请求体
			body, _ := json.Marshal(tc.requestBody)

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/configs/"+tc.configID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			suite.Router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			} else if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])
			}
		})
	}
}

// TestDeleteConfig 测试删除配置
func (suite *ConfigHandlerTestSuite) TestDeleteConfig() {
	testCases := []struct {
		name           string
		configID       string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "成功删除配置",
			configID: "1",
			setupMocks: func() {
				suite.configService.On("DeleteConfig", mock.Anything, uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "配置不存在",
			configID: "999",
			setupMocks: func() {
				suite.configService.On("DeleteConfig", mock.Anything, uint(999)).Return(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "配置不存在",
		},
		{
			name:           "无效的配置ID",
			configID:       "invalid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的配置ID",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configService.ExpectedCalls = nil
			suite.configService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/configs/"+tc.configID, nil)
			suite.Router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			} else if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
			}
		})
	}
}

// TestListConfigs 测试配置列表
func (suite *ConfigHandlerTestSuite) TestListConfigs() {
	testCases := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "成功获取配置列表",
			queryParams: "?namespace=default&page=1&limit=10",
			setupMocks: func() {
				configs := []*model.Config{
					{ID: 1, Key: "app.database.host", Value: "localhost", Namespace: "default"},
					{ID: 2, Key: "app.database.port", Value: "5432", Namespace: "default"},
				}
				filters := map[string]interface{}{"namespace": "default"}
				suite.configService.On("ListConfigs", mock.Anything, filters, 0, 10).Return(configs, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "空列表",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				configs := []*model.Config{}
				filters := map[string]interface{}{}
				suite.configService.On("ListConfigs", mock.Anything, filters, 0, 10).Return(configs, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:        "带环境过滤",
			queryParams: "?environment=production&page=1&limit=5",
			setupMocks: func() {
				configs := []*model.Config{
					{ID: 1, Key: "app.database.host", Value: "prod-db", Environment: "production"},
				}
				filters := map[string]interface{}{"environment": "production"}
				suite.configService.On("ListConfigs", mock.Anything, filters, 0, 5).Return(configs, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configService.ExpectedCalls = nil
			suite.configService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/configs"+tc.queryParams, nil)
			suite.Router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])

				data := response["data"].(map[string]interface{})
				configs := data["configs"].([]interface{})
				assert.Len(suite.T(), configs, tc.expectedCount)
			}
		})
	}
}

// TestGetConfigsByNamespace 测试按命名空间获取配置
func (suite *ConfigHandlerTestSuite) TestGetConfigsByNamespace() {
	namespace := "production"
	configs := []*model.Config{
		{ID: 1, Key: "app.database.host", Value: "prod-db", Namespace: namespace},
		{ID: 2, Key: "app.redis.host", Value: "prod-redis", Namespace: namespace},
	}

	suite.configService.On("GetConfigsByNamespace", mock.Anything, namespace).Return(configs, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/namespace/"+namespace, nil)
	suite.Router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "success", response["status"])

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

// TestGetConfigsByEnvironment 测试按环境获取配置
func (suite *ConfigHandlerTestSuite) TestGetConfigsByEnvironment() {
	environment := "production"
	configs := []*model.Config{
		{ID: 1, Key: "app.database.host", Value: "prod-db", Environment: environment},
		{ID: 2, Key: "app.redis.host", Value: "prod-redis", Environment: environment},
	}

	suite.configService.On("GetConfigsByEnvironment", mock.Anything, environment).Return(configs, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/environment/"+environment, nil)
	suite.Router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "success", response["status"])

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

// TestValidation 测试输入验证
func (suite *ConfigHandlerTestSuite) TestValidation() {
	testCases := []struct {
		name           string
		endpoint       string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "创建配置 - 缺少必填字段",
			endpoint: "/configs",
			method:   "POST",
			requestBody: map[string]interface{}{
				"value": "test_value",
				// 缺少 key 字段
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "配置键不能为空",
		},
		{
			name:     "创建配置 - 无效的配置类型",
			endpoint: "/configs",
			method:   "POST",
			requestBody: map[string]interface{}{
				"key":   "test.key",
				"value": "test_value",
				"type":  "invalid_type",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的配置类型",
		},
		{
			name:        "更新配置 - 空的更新内容",
			endpoint:    "/configs/1",
			method:      "PUT",
			requestBody: map[string]interface{}{
				// 空的更新内容
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "至少需要提供一个更新字段",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 准备请求体
			body, _ := json.Marshal(tc.requestBody)

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			suite.Router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			}
		})
	}
}

// 运行测试套件
func TestConfigHandlerSuite(t *testing.T) {
	suite.Run(t, new(ConfigHandlerTestSuite))
}

// 基准测试
func BenchmarkConfigHandler_GetConfig(b *testing.B) {
	// 设置基准测试环境
	gin.SetMode(gin.TestMode)

	configService := new(MockConfigService)
	handler := &ConfigHandler{
		configService: configService,
	}

	router := gin.New()
	router.GET("/configs/:id", handler.GetConfig)

	config := &model.Config{
		ID:          1,
		Key:         "app.database.host",
		Value:       "localhost",
		Type:        "string",
		Namespace:   "default",
		Environment: "development",
	}

	// 设置模拟
	configService.On("GetConfig", mock.Anything, uint(1)).Return(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/configs/1", nil)
		router.ServeHTTP(w, req)
	}
}
