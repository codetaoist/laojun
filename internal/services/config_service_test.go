package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun/config-center/internal/model"
	sharedTesting "github.com/codetaoist/laojun/shared/testing"
)

// MockConfigRepository 模拟配置仓库
type MockConfigRepository struct {
	mock.Mock
}

func (m *MockConfigRepository) Create(ctx context.Context, config *model.Config) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockConfigRepository) GetByID(ctx context.Context, id uint) (*model.Config, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Config), args.Error(1)
}

func (m *MockConfigRepository) GetByKey(ctx context.Context, key string) (*model.Config, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Config), args.Error(1)
}

func (m *MockConfigRepository) Update(ctx context.Context, config *model.Config) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockConfigRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConfigRepository) List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.Config, int64, error) {
	args := m.Called(ctx, filters, offset, limit)
	return args.Get(0).([]*model.Config), args.Get(1).(int64), args.Error(2)
}

func (m *MockConfigRepository) GetByNamespace(ctx context.Context, namespace string) ([]*model.Config, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]*model.Config), args.Error(1)
}

func (m *MockConfigRepository) GetByEnvironment(ctx context.Context, environment string) ([]*model.Config, error) {
	args := m.Called(ctx, environment)
	return args.Get(0).([]*model.Config), args.Error(1)
}

// MockCacheService 模拟缓存服务
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// ConfigServiceTestSuite 配置服务测试套件
type ConfigServiceTestSuite struct {
	sharedTesting.TestSuite
	configRepo    *MockConfigRepository
	cacheService  *MockCacheService
	configService *ConfigService
}

func (suite *ConfigServiceTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 创建模拟仓库和缓存服务
	suite.configRepo = new(MockConfigRepository)
	suite.cacheService = new(MockCacheService)

	// 创建配置服务
	suite.configService = &ConfigService{
		configRepo:   suite.configRepo,
		cacheService: suite.cacheService,
	}
}

func (suite *ConfigServiceTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()
	suite.configRepo.AssertExpectations(suite.T())
	suite.cacheService.AssertExpectations(suite.T())
}

// TestCreateConfig 测试创建配置
func (suite *ConfigServiceTestSuite) TestCreateConfig() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		input       *CreateConfigRequest
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功创建配置",
			input: &CreateConfigRequest{
				Key:         "app.database.host",
				Value:       "localhost",
				Type:        "string",
				Namespace:   "default",
				Environment: "development",
				Description: "数据库主机地址",
			},
			setupMocks: func() {
				// 检查配置键不存在
				suite.configRepo.On("GetByKey", ctx, "app.database.host").Return((*model.Config)(nil), gorm.ErrRecordNotFound)
				// 创建配置
				suite.configRepo.On("Create", ctx, mock.AnythingOfType("*model.Config")).Return(nil)
				// 清除缓存
				suite.cacheService.On("DeletePattern", ctx, "config:*").Return(nil)
			},
			expectError: false,
		},
		{
			name: "配置键已存在",
			input: &CreateConfigRequest{
				Key:         "existing.key",
				Value:       "value",
				Type:        "string",
				Namespace:   "default",
				Environment: "development",
			},
			setupMocks: func() {
				existingConfig := &model.Config{
					ID:          1,
					Key:         "existing.key",
					Value:       "old_value",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
				}
				suite.configRepo.On("GetByKey", ctx, "existing.key").Return(existingConfig, nil)
			},
			expectError: true,
			errorMsg:    "配置键已存在",
		},
		{
			name: "无效的配置类型",
			input: &CreateConfigRequest{
				Key:         "app.config",
				Value:       "invalid_json",
				Type:        "json",
				Namespace:   "default",
				Environment: "development",
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "无效的JSON格式",
		},
		{
			name: "配置键为空",
			input: &CreateConfigRequest{
				Key:         "",
				Value:       "value",
				Type:        "string",
				Namespace:   "default",
				Environment: "development",
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "配置键不能为空",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configRepo.ExpectedCalls = nil
			suite.configRepo.Calls = nil
			suite.cacheService.ExpectedCalls = nil
			suite.cacheService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			config, err := suite.configService.CreateConfig(ctx, tc.input)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), config)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), config)
				assert.Equal(suite.T(), tc.input.Key, config.Key)
				assert.Equal(suite.T(), tc.input.Value, config.Value)
				assert.Equal(suite.T(), tc.input.Type, config.Type)
				assert.Equal(suite.T(), tc.input.Namespace, config.Namespace)
				assert.Equal(suite.T(), tc.input.Environment, config.Environment)
			}
		})
	}
}

// TestGetConfig 测试获取配置
func (suite *ConfigServiceTestSuite) TestGetConfig() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		configID    uint
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "成功获取配置",
			configID: 1,
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
				suite.configRepo.On("GetByID", ctx, uint(1)).Return(config, nil)
			},
			expectError: false,
		},
		{
			name:     "配置不存在",
			configID: 999,
			setupMocks: func() {
				suite.configRepo.On("GetByID", ctx, uint(999)).Return((*model.Config)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "配置不存在",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configRepo.ExpectedCalls = nil
			suite.configRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			config, err := suite.configService.GetConfig(ctx, tc.configID)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), config)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), config)
				assert.Equal(suite.T(), tc.configID, config.ID)
			}
		})
	}
}

// TestGetConfigByKey 测试通过键获取配置
func (suite *ConfigServiceTestSuite) TestGetConfigByKey() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		configKey   string
		setupMocks  func()
		expectError bool
		errorMsg    string
		useCache    bool
	}{
		{
			name:      "从数据库获取配置",
			configKey: "app.database.host",
			setupMocks: func() {
				// 缓存中没有数据
				suite.cacheService.On("Get", ctx, "config:app.database.host").Return("", gorm.ErrRecordNotFound)

				config := &model.Config{
					ID:          1,
					Key:         "app.database.host",
					Value:       "localhost",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
				}
				suite.configRepo.On("GetByKey", ctx, "app.database.host").Return(config, nil)

				// 设置缓存
				configJSON, _ := json.Marshal(config)
				suite.cacheService.On("Set", ctx, "config:app.database.host", string(configJSON), 5*time.Minute).Return(nil)
			},
			expectError: false,
			useCache:    false,
		},
		{
			name:      "从缓存获取配置",
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
				configJSON, _ := json.Marshal(config)

				// 缓存中有数据
				suite.cacheService.On("Get", ctx, "config:app.database.host").Return(string(configJSON), nil)
			},
			expectError: false,
			useCache:    true,
		},
		{
			name:      "配置不存在",
			configKey: "nonexistent.key",
			setupMocks: func() {
				// 缓存中没有数据
				suite.cacheService.On("Get", ctx, "config:nonexistent.key").Return("", gorm.ErrRecordNotFound)
				// 数据库中也没有数据
				suite.configRepo.On("GetByKey", ctx, "nonexistent.key").Return((*model.Config)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "配置不存在",
			useCache:    false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configRepo.ExpectedCalls = nil
			suite.configRepo.Calls = nil
			suite.cacheService.ExpectedCalls = nil
			suite.cacheService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			config, err := suite.configService.GetConfigByKey(ctx, tc.configKey)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), config)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), config)
				assert.Equal(suite.T(), tc.configKey, config.Key)
			}
		})
	}
}

// TestUpdateConfig 测试更新配置
func (suite *ConfigServiceTestSuite) TestUpdateConfig() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		configID    uint
		input       *UpdateConfigRequest
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "成功更新配置",
			configID: 1,
			input: &UpdateConfigRequest{
				Value:       "new_value",
				Description: "更新后的描述",
			},
			setupMocks: func() {
				existingConfig := &model.Config{
					ID:          1,
					Key:         "app.database.host",
					Value:       "old_value",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
					Description: "旧描述",
				}
				suite.configRepo.On("GetByID", ctx, uint(1)).Return(existingConfig, nil)
				suite.configRepo.On("Update", ctx, mock.AnythingOfType("*model.Config")).Return(nil)

				// 清除缓存
				suite.cacheService.On("Delete", ctx, "config:app.database.host").Return(nil)
				suite.cacheService.On("DeletePattern", ctx, "config:*").Return(nil)
			},
			expectError: false,
		},
		{
			name:     "配置不存在",
			configID: 999,
			input: &UpdateConfigRequest{
				Value: "new_value",
			},
			setupMocks: func() {
				suite.configRepo.On("GetByID", ctx, uint(999)).Return((*model.Config)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "配置不存在",
		},
		{
			name:     "无效的JSON格式",
			configID: 1,
			input: &UpdateConfigRequest{
				Value: "invalid_json",
			},
			setupMocks: func() {
				existingConfig := &model.Config{
					ID:          1,
					Key:         "app.config",
					Value:       `{"key": "value"}`,
					Type:        "json",
					Namespace:   "default",
					Environment: "development",
				}
				suite.configRepo.On("GetByID", ctx, uint(1)).Return(existingConfig, nil)
			},
			expectError: true,
			errorMsg:    "无效的JSON格式",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configRepo.ExpectedCalls = nil
			suite.configRepo.Calls = nil
			suite.cacheService.ExpectedCalls = nil
			suite.cacheService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			config, err := suite.configService.UpdateConfig(ctx, tc.configID, tc.input)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), config)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), config)
				assert.Equal(suite.T(), tc.configID, config.ID)
				if tc.input.Value != "" {
					assert.Equal(suite.T(), tc.input.Value, config.Value)
				}
				if tc.input.Description != "" {
					assert.Equal(suite.T(), tc.input.Description, config.Description)
				}
			}
		})
	}
}

// TestDeleteConfig 测试删除配置
func (suite *ConfigServiceTestSuite) TestDeleteConfig() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		configID    uint
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "成功删除配置",
			configID: 1,
			setupMocks: func() {
				config := &model.Config{
					ID:          1,
					Key:         "app.database.host",
					Value:       "localhost",
					Type:        "string",
					Namespace:   "default",
					Environment: "development",
				}
				suite.configRepo.On("GetByID", ctx, uint(1)).Return(config, nil)
				suite.configRepo.On("Delete", ctx, uint(1)).Return(nil)

				// 清除缓存
				suite.cacheService.On("Delete", ctx, "config:app.database.host").Return(nil)
				suite.cacheService.On("DeletePattern", ctx, "config:*").Return(nil)
			},
			expectError: false,
		},
		{
			name:     "配置不存在",
			configID: 999,
			setupMocks: func() {
				suite.configRepo.On("GetByID", ctx, uint(999)).Return((*model.Config)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "配置不存在",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configRepo.ExpectedCalls = nil
			suite.configRepo.Calls = nil
			suite.cacheService.ExpectedCalls = nil
			suite.cacheService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			err := suite.configService.DeleteConfig(ctx, tc.configID)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestListConfigs 测试配置列表
func (suite *ConfigServiceTestSuite) TestListConfigs() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		filters     map[string]interface{}
		offset      int
		limit       int
		setupMocks  func()
		expectLen   int
		expectTotal int64
	}{
		{
			name:    "成功获取配置列表",
			filters: map[string]interface{}{"namespace": "default"},
			offset:  0,
			limit:   10,
			setupMocks: func() {
				configs := []*model.Config{
					{ID: 1, Key: "app.database.host", Value: "localhost", Namespace: "default"},
					{ID: 2, Key: "app.database.port", Value: "5432", Namespace: "default"},
					{ID: 3, Key: "app.redis.host", Value: "localhost", Namespace: "default"},
				}
				suite.configRepo.On("List", ctx, map[string]interface{}{"namespace": "default"}, 0, 10).Return(configs, int64(3), nil)
			},
			expectLen:   3,
			expectTotal: 3,
		},
		{
			name:    "空配置列表",
			filters: map[string]interface{}{},
			offset:  0,
			limit:   10,
			setupMocks: func() {
				configs := []*model.Config{}
				suite.configRepo.On("List", ctx, map[string]interface{}{}, 0, 10).Return(configs, int64(0), nil)
			},
			expectLen:   0,
			expectTotal: 0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.configRepo.ExpectedCalls = nil
			suite.configRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			configs, total, err := suite.configService.ListConfigs(ctx, tc.filters, tc.offset, tc.limit)

			assert.NoError(suite.T(), err)
			assert.Len(suite.T(), configs, tc.expectLen)
			assert.Equal(suite.T(), tc.expectTotal, total)
		})
	}
}

// TestValidateConfigValue 测试配置值验证
func (suite *ConfigServiceTestSuite) TestValidateConfigValue() {
	testCases := []struct {
		name        string
		configType  string
		value       string
		expectError bool
	}{
		{"有效字符", "string", "hello world", false},
		{"有效整数", "int", "123", false},
		{"无效整数", "int", "not_a_number", true},
		{"有效浮点", "float", "123.45", false},
		{"无效浮点", "float", "not_a_float", true},
		{"有效布尔", "bool", "true", false},
		{"无效布尔", "bool", "not_a_bool", true},
		{"有效JSON", "json", `{"key": "value"}`, false},
		{"无效JSON", "json", `{invalid json}`, true},
		{"有效YAML", "yaml", "key: value", false},
		{"无效YAML", "yaml", "key: value\n  invalid: yaml: content", true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.configService.validateConfigValue(tc.configType, tc.value)
			if tc.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestGetConfigsByNamespace 测试按命名空间获取配置
func (suite *ConfigServiceTestSuite) TestGetConfigsByNamespace() {
	ctx := context.Background()
	namespace := "production"

	configs := []*model.Config{
		{ID: 1, Key: "app.database.host", Value: "prod-db", Namespace: namespace},
		{ID: 2, Key: "app.redis.host", Value: "prod-redis", Namespace: namespace},
	}

	suite.configRepo.On("GetByNamespace", ctx, namespace).Return(configs, nil)

	result, err := suite.configService.GetConfigsByNamespace(ctx, namespace)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), namespace, result[0].Namespace)
	assert.Equal(suite.T(), namespace, result[1].Namespace)
}

// TestGetConfigsByEnvironment 测试按环境获取配置
func (suite *ConfigServiceTestSuite) TestGetConfigsByEnvironment() {
	ctx := context.Background()
	environment := "production"

	configs := []*model.Config{
		{ID: 1, Key: "app.database.host", Value: "prod-db", Environment: environment},
		{ID: 2, Key: "app.redis.host", Value: "prod-redis", Environment: environment},
	}

	suite.configRepo.On("GetByEnvironment", ctx, environment).Return(configs, nil)

	result, err := suite.configService.GetConfigsByEnvironment(ctx, environment)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), environment, result[0].Environment)
	assert.Equal(suite.T(), environment, result[1].Environment)
}

// 运行测试套件
func TestConfigServiceSuite(t *testing.T) {
	suite.Run(t, new(ConfigServiceTestSuite))
}

// 基准测试
func BenchmarkConfigService_GetConfigByKey(b *testing.B) {
	// 设置基准测试环境
	configRepo := new(MockConfigRepository)
	cacheService := new(MockCacheService)
	configService := &ConfigService{
		configRepo:   configRepo,
		cacheService: cacheService,
	}

	ctx := context.Background()
	configKey := "app.database.host"

	config := &model.Config{
		ID:          1,
		Key:         configKey,
		Value:       "localhost",
		Type:        "string",
		Namespace:   "default",
		Environment: "development",
	}

	// 设置模拟 - 从缓存获取配置
	configJSON, _ := json.Marshal(config)
	cacheService.On("Get", ctx, "config:"+configKey).Return(string(configJSON), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configService.GetConfigByKey(ctx, configKey)
	}
}

func BenchmarkConfigService_CreateConfig(b *testing.B) {
	// 设置基准测试环境
	configRepo := new(MockConfigRepository)
	cacheService := new(MockCacheService)
	configService := &ConfigService{
		configRepo:   configRepo,
		cacheService: cacheService,
	}

	ctx := context.Background()
	req := &CreateConfigRequest{
		Key:         "bench.config",
		Value:       "benchmark_value",
		Type:        "string",
		Namespace:   "default",
		Environment: "development",
	}

	// 设置模拟
	configRepo.On("GetByKey", ctx, "bench.config").Return((*model.Config)(nil), gorm.ErrRecordNotFound)
	configRepo.On("Create", ctx, mock.AnythingOfType("*model.Config")).Return(nil)
	cacheService.On("DeletePattern", ctx, "config:*").Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configService.CreateConfig(ctx, req)
	}
}
