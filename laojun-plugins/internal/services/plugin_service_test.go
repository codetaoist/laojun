package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun-plugins/internal/models"
	"github.com/codetaoist/laojun-shared/testing"
)

// MockPluginRepository 模拟插件仓库
type MockPluginRepository struct {
	mock.Mock
}

func (m *MockPluginRepository) Create(ctx context.Context, plugin *model.Plugin) error {
	args := m.Called(ctx, plugin)
	return args.Error(0)
}

func (m *MockPluginRepository) GetByID(ctx context.Context, id uint) (*model.Plugin, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Plugin), args.Error(1)
}

func (m *MockPluginRepository) GetByName(ctx context.Context, name string) (*model.Plugin, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Plugin), args.Error(1)
}

func (m *MockPluginRepository) Update(ctx context.Context, plugin *model.Plugin) error {
	args := m.Called(ctx, plugin)
	return args.Error(0)
}

func (m *MockPluginRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPluginRepository) List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.Plugin, int64, error) {
	args := m.Called(ctx, filters, offset, limit)
	return args.Get(0).([]*model.Plugin), args.Get(1).(int64), args.Error(2)
}

func (m *MockPluginRepository) GetByCategory(ctx context.Context, category string) ([]*model.Plugin, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]*model.Plugin), args.Error(1)
}

func (m *MockPluginRepository) GetByAuthor(ctx context.Context, authorID uint) ([]*model.Plugin, error) {
	args := m.Called(ctx, authorID)
	return args.Get(0).([]*model.Plugin), args.Error(1)
}

func (m *MockPluginRepository) Search(ctx context.Context, query string, filters map[string]interface{}, offset, limit int) ([]*model.Plugin, int64, error) {
	args := m.Called(ctx, query, filters, offset, limit)
	return args.Get(0).([]*model.Plugin), args.Get(1).(int64), args.Error(2)
}

func (m *MockPluginRepository) UpdateDownloadCount(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPluginRepository) UpdateRating(ctx context.Context, id uint, rating float64) error {
	args := m.Called(ctx, id, rating)
	return args.Error(0)
}

// MockStorageService 模拟存储服务
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) UploadFile(ctx context.Context, filename string, data []byte) (string, error) {
	args := m.Called(ctx, filename, data)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) DeleteFile(ctx context.Context, filename string) error {
	args := m.Called(ctx, filename)
	return args.Error(0)
}

func (m *MockStorageService) GetFileURL(ctx context.Context, filename string) (string, error) {
	args := m.Called(ctx, filename)
	return args.String(0), args.Error(1)
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

// PluginServiceTestSuite 插件服务测试套件
type PluginServiceTestSuite struct {
	sharedTesting.TestSuite
	pluginRepo     *MockPluginRepository
	storageService *MockStorageService
	cacheService   *MockCacheService
	pluginService  *PluginService
}

func (suite *PluginServiceTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 创建模拟仓库和服务
	suite.pluginRepo = new(MockPluginRepository)
	suite.storageService = new(MockStorageService)
	suite.cacheService = new(MockCacheService)

	// 创建插件服务
	suite.pluginService = &PluginService{
		pluginRepo:     suite.pluginRepo,
		storageService: suite.storageService,
		cacheService:   suite.cacheService,
	}
}

func (suite *PluginServiceTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()
	suite.pluginRepo.AssertExpectations(suite.T())
	suite.storageService.AssertExpectations(suite.T())
	suite.cacheService.AssertExpectations(suite.T())
}

// TestCreatePlugin 测试创建插件
func (suite *PluginServiceTestSuite) TestCreatePlugin() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		input       *CreatePluginRequest
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功创建插件",
			input: &CreatePluginRequest{
				Name:        "test-plugin",
				DisplayName: "Test Plugin",
				Description: "A test plugin",
				Version:     "1.0.0",
				Category:    "utility",
				AuthorID:    1,
				Tags:        []string{"test", "utility"},
				FileData:    []byte("plugin content"),
				FileName:    "test-plugin.zip",
			},
			setupMocks: func() {
				// 检查插件名不存在
				suite.pluginRepo.On("GetByName", ctx, "test-plugin").Return((*model.Plugin)(nil), gorm.ErrRecordNotFound)

				// 上传文件
				suite.storageService.On("UploadFile", ctx, "test-plugin.zip", []byte("plugin content")).Return("plugins/test-plugin.zip", nil)

				// 创建插件
				suite.pluginRepo.On("Create", ctx, mock.AnythingOfType("*model.Plugin")).Return(nil)

				// 清除缓存
				suite.cacheService.On("DeletePattern", ctx, "plugin:*").Return(nil)
			},
			expectError: false,
		},
		{
			name: "插件名已存在",
			input: &CreatePluginRequest{
				Name:        "existing-plugin",
				DisplayName: "Existing Plugin",
				Description: "An existing plugin",
				Version:     "1.0.0",
				Category:    "utility",
				AuthorID:    1,
			},
			setupMocks: func() {
				existingPlugin := &model.Plugin{
					ID:          1,
					Name:        "existing-plugin",
					DisplayName: "Existing Plugin",
					Version:     "1.0.0",
				}
				suite.pluginRepo.On("GetByName", ctx, "existing-plugin").Return(existingPlugin, nil)
			},
			expectError: true,
			errorMsg:    "插件名已存在",
		},
		{
			name: "无效的插件名",
			input: &CreatePluginRequest{
				Name:        "",
				DisplayName: "Test Plugin",
				Description: "A test plugin",
				Version:     "1.0.0",
				Category:    "utility",
				AuthorID:    1,
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "插件名不能为空",
		},
		{
			name: "无效的版本号",
			input: &CreatePluginRequest{
				Name:        "test-plugin",
				DisplayName: "Test Plugin",
				Description: "A test plugin",
				Version:     "invalid-version",
				Category:    "utility",
				AuthorID:    1,
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "无效的版本号格式",
		},
		{
			name: "文件上传失败",
			input: &CreatePluginRequest{
				Name:        "test-plugin",
				DisplayName: "Test Plugin",
				Description: "A test plugin",
				Version:     "1.0.0",
				Category:    "utility",
				AuthorID:    1,
				FileData:    []byte("plugin content"),
				FileName:    "test-plugin.zip",
			},
			setupMocks: func() {
				// 检查插件名不存在
				suite.pluginRepo.On("GetByName", ctx, "test-plugin").Return((*model.Plugin)(nil), gorm.ErrRecordNotFound)

				// 文件上传失败
				suite.storageService.On("UploadFile", ctx, "test-plugin.zip", []byte("plugin content")).Return("", fmt.Errorf("上传失败"))
			},
			expectError: true,
			errorMsg:    "上传失败",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.pluginRepo.ExpectedCalls = nil
			suite.pluginRepo.Calls = nil
			suite.storageService.ExpectedCalls = nil
			suite.storageService.Calls = nil
			suite.cacheService.ExpectedCalls = nil
			suite.cacheService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			plugin, err := suite.pluginService.CreatePlugin(ctx, tc.input)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), plugin)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), plugin)
				assert.Equal(suite.T(), tc.input.Name, plugin.Name)
				assert.Equal(suite.T(), tc.input.DisplayName, plugin.DisplayName)
				assert.Equal(suite.T(), tc.input.Version, plugin.Version)
				assert.Equal(suite.T(), tc.input.Category, plugin.Category)
				assert.Equal(suite.T(), tc.input.AuthorID, plugin.AuthorID)
			}
		})
	}
}

// TestGetPlugin 测试获取插件
func (suite *PluginServiceTestSuite) TestGetPlugin() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		pluginID    uint
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "成功获取插件",
			pluginID: 1,
			setupMocks: func() {
				plugin := &model.Plugin{
					ID:          1,
					Name:        "test-plugin",
					DisplayName: "Test Plugin",
					Description: "A test plugin",
					Version:     "1.0.0",
					Category:    "utility",
					AuthorID:    1,
					Status:      "published",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				suite.pluginRepo.On("GetByID", ctx, uint(1)).Return(plugin, nil)
			},
			expectError: false,
		},
		{
			name:     "插件不存在",
			pluginID: 999,
			setupMocks: func() {
				suite.pluginRepo.On("GetByID", ctx, uint(999)).Return((*model.Plugin)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "插件不存在",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.pluginRepo.ExpectedCalls = nil
			suite.pluginRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			plugin, err := suite.pluginService.GetPlugin(ctx, tc.pluginID)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), plugin)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), plugin)
				assert.Equal(suite.T(), tc.pluginID, plugin.ID)
			}
		})
	}
}

// TestUpdatePlugin 测试更新插件
func (suite *PluginServiceTestSuite) TestUpdatePlugin() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		pluginID    uint
		input       *UpdatePluginRequest
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "成功更新插件",
			pluginID: 1,
			input: &UpdatePluginRequest{
				DisplayName: "Updated Plugin",
				Description: "Updated description",
				Version:     "1.1.0",
			},
			setupMocks: func() {
				existingPlugin := &model.Plugin{
					ID:          1,
					Name:        "test-plugin",
					DisplayName: "Test Plugin",
					Description: "A test plugin",
					Version:     "1.0.0",
					Category:    "utility",
					AuthorID:    1,
					Status:      "published",
				}
				suite.pluginRepo.On("GetByID", ctx, uint(1)).Return(existingPlugin, nil)
				suite.pluginRepo.On("Update", ctx, mock.AnythingOfType("*model.Plugin")).Return(nil)

				// 清除缓存
				suite.cacheService.On("Delete", ctx, "plugin:1").Return(nil)
				suite.cacheService.On("DeletePattern", ctx, "plugin:*").Return(nil)
			},
			expectError: false,
		},
		{
			name:     "插件不存在",
			pluginID: 999,
			input: &UpdatePluginRequest{
				DisplayName: "Updated Plugin",
			},
			setupMocks: func() {
				suite.pluginRepo.On("GetByID", ctx, uint(999)).Return((*model.Plugin)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "插件不存在",
		},
		{
			name:     "无效的版本号",
			pluginID: 1,
			input: &UpdatePluginRequest{
				Version: "invalid-version",
			},
			setupMocks: func() {
				existingPlugin := &model.Plugin{
					ID:      1,
					Name:    "test-plugin",
					Version: "1.0.0",
				}
				suite.pluginRepo.On("GetByID", ctx, uint(1)).Return(existingPlugin, nil)
			},
			expectError: true,
			errorMsg:    "无效的版本号格式",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.pluginRepo.ExpectedCalls = nil
			suite.pluginRepo.Calls = nil
			suite.cacheService.ExpectedCalls = nil
			suite.cacheService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			plugin, err := suite.pluginService.UpdatePlugin(ctx, tc.pluginID, tc.input)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), plugin)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), plugin)
				assert.Equal(suite.T(), tc.pluginID, plugin.ID)
				if tc.input.DisplayName != "" {
					assert.Equal(suite.T(), tc.input.DisplayName, plugin.DisplayName)
				}
				if tc.input.Description != "" {
					assert.Equal(suite.T(), tc.input.Description, plugin.Description)
				}
				if tc.input.Version != "" {
					assert.Equal(suite.T(), tc.input.Version, plugin.Version)
				}
			}
		})
	}
}

// TestDeletePlugin 测试删除插件
func (suite *PluginServiceTestSuite) TestDeletePlugin() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		pluginID    uint
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "成功删除插件",
			pluginID: 1,
			setupMocks: func() {
				plugin := &model.Plugin{
					ID:       1,
					Name:     "test-plugin",
					FilePath: "plugins/test-plugin.zip",
				}
				suite.pluginRepo.On("GetByID", ctx, uint(1)).Return(plugin, nil)
				suite.pluginRepo.On("Delete", ctx, uint(1)).Return(nil)

				// 删除文件
				suite.storageService.On("DeleteFile", ctx, "plugins/test-plugin.zip").Return(nil)

				// 清除缓存
				suite.cacheService.On("Delete", ctx, "plugin:1").Return(nil)
				suite.cacheService.On("DeletePattern", ctx, "plugin:*").Return(nil)
			},
			expectError: false,
		},
		{
			name:     "插件不存在",
			pluginID: 999,
			setupMocks: func() {
				suite.pluginRepo.On("GetByID", ctx, uint(999)).Return((*model.Plugin)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "插件不存在",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.pluginRepo.ExpectedCalls = nil
			suite.pluginRepo.Calls = nil
			suite.storageService.ExpectedCalls = nil
			suite.storageService.Calls = nil
			suite.cacheService.ExpectedCalls = nil
			suite.cacheService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			err := suite.pluginService.DeletePlugin(ctx, tc.pluginID)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestListPlugins 测试插件列表
func (suite *PluginServiceTestSuite) TestListPlugins() {
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
			name:    "成功获取插件列表",
			filters: map[string]interface{}{"category": "utility"},
			offset:  0,
			limit:   10,
			setupMocks: func() {
				plugins := []*model.Plugin{
					{ID: 1, Name: "plugin1", Category: "utility", Status: "published"},
					{ID: 2, Name: "plugin2", Category: "utility", Status: "published"},
					{ID: 3, Name: "plugin3", Category: "utility", Status: "published"},
				}
				suite.pluginRepo.On("List", ctx, map[string]interface{}{"category": "utility"}, 0, 10).Return(plugins, int64(3), nil)
			},
			expectLen:   3,
			expectTotal: 3,
		},
		{
			name:    "空列表",
			filters: map[string]interface{}{},
			offset:  0,
			limit:   10,
			setupMocks: func() {
				plugins := []*model.Plugin{}
				suite.pluginRepo.On("List", ctx, map[string]interface{}{}, 0, 10).Return(plugins, int64(0), nil)
			},
			expectLen:   0,
			expectTotal: 0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.pluginRepo.ExpectedCalls = nil
			suite.pluginRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			plugins, total, err := suite.pluginService.ListPlugins(ctx, tc.filters, tc.offset, tc.limit)

			assert.NoError(suite.T(), err)
			assert.Len(suite.T(), plugins, tc.expectLen)
			assert.Equal(suite.T(), tc.expectTotal, total)
		})
	}
}

// TestSearchPlugins 测试搜索插件
func (suite *PluginServiceTestSuite) TestSearchPlugins() {
	ctx := context.Background()
	query := "test"
	filters := map[string]interface{}{"category": "utility"}

	plugins := []*model.Plugin{
		{ID: 1, Name: "test-plugin", DisplayName: "Test Plugin", Category: "utility"},
		{ID: 2, Name: "another-test", DisplayName: "Another Test", Category: "utility"},
	}

	suite.pluginRepo.On("Search", ctx, query, filters, 0, 10).Return(plugins, int64(2), nil)

	result, total, err := suite.pluginService.SearchPlugins(ctx, query, filters, 0, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
}

// TestDownloadPlugin 测试下载插件
func (suite *PluginServiceTestSuite) TestDownloadPlugin() {
	ctx := context.Background()
	pluginID := uint(1)

	plugin := &model.Plugin{
		ID:       1,
		Name:     "test-plugin",
		FilePath: "plugins/test-plugin.zip",
		Status:   "published",
	}

	suite.pluginRepo.On("GetByID", ctx, pluginID).Return(plugin, nil)
	suite.storageService.On("GetFileURL", ctx, "plugins/test-plugin.zip").Return("https://storage.example.com/plugins/test-plugin.zip", nil)
	suite.pluginRepo.On("UpdateDownloadCount", ctx, pluginID).Return(nil)

	// 清除缓存
	suite.cacheService.On("Delete", ctx, "plugin:1").Return(nil)

	downloadURL, err := suite.pluginService.DownloadPlugin(ctx, pluginID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "https://storage.example.com/plugins/test-plugin.zip", downloadURL)
}

// TestValidateVersion 测试版本号验证
func (suite *PluginServiceTestSuite) TestValidateVersion() {
	testCases := []struct {
		name        string
		version     string
		expectError bool
	}{
		{"有效版本号 三位", "1.0.0", false},
		{"有效版本号 带预发布", "1.0.0-alpha", false},
		{"有效版本号 带构建号", "1.0.0+build.1", false},
		{"有效版本号 完整格式", "1.0.0-alpha.1+build.1", false},
		{"无效版本号 空字符串", "", true},
		{"无效版本号 只有数字", "1", true},
		{"无效版本号 两位", "1.0", true},
		{"无效版本号 非数字", "a.b.c", true},
		{"无效版本号 负数", "-1.0.0", true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.pluginService.validateVersion(tc.version)
			if tc.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestGetPluginsByCategory 测试按分类获取插件
func (suite *PluginServiceTestSuite) TestGetPluginsByCategory() {
	ctx := context.Background()
	category := "utility"

	plugins := []*model.Plugin{
		{ID: 1, Name: "plugin1", Category: category, Status: "published"},
		{ID: 2, Name: "plugin2", Category: category, Status: "published"},
	}

	suite.pluginRepo.On("GetByCategory", ctx, category).Return(plugins, nil)

	result, err := suite.pluginService.GetPluginsByCategory(ctx, category)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), category, result[0].Category)
	assert.Equal(suite.T(), category, result[1].Category)
}

// TestGetPluginsByAuthor 测试按作者获取插件
func (suite *PluginServiceTestSuite) TestGetPluginsByAuthor() {
	ctx := context.Background()
	authorID := uint(1)

	plugins := []*model.Plugin{
		{ID: 1, Name: "plugin1", AuthorID: authorID, Status: "published"},
		{ID: 2, Name: "plugin2", AuthorID: authorID, Status: "published"},
	}

	suite.pluginRepo.On("GetByAuthor", ctx, authorID).Return(plugins, nil)

	result, err := suite.pluginService.GetPluginsByAuthor(ctx, authorID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), authorID, result[0].AuthorID)
	assert.Equal(suite.T(), authorID, result[1].AuthorID)
}

// TestUpdatePluginRating 测试更新插件评分
func (suite *PluginServiceTestSuite) TestUpdatePluginRating() {
	ctx := context.Background()
	pluginID := uint(1)
	rating := 4.5

	plugin := &model.Plugin{
		ID:     1,
		Name:   "test-plugin",
		Status: "published",
	}

	suite.pluginRepo.On("GetByID", ctx, pluginID).Return(plugin, nil)
	suite.pluginRepo.On("UpdateRating", ctx, pluginID, rating).Return(nil)

	// 清除缓存
	suite.cacheService.On("Delete", ctx, "plugin:1").Return(nil)

	err := suite.pluginService.UpdatePluginRating(ctx, pluginID, rating)

	assert.NoError(suite.T(), err)
}

// 运行测试套件
func TestPluginServiceSuite(t *testing.T) {
	suite.Run(t, new(PluginServiceTestSuite))
}

// 基准测试
func BenchmarkPluginService_GetPlugin(b *testing.B) {
	// 设置基准测试环境
	pluginRepo := new(MockPluginRepository)
	storageService := new(MockStorageService)
	cacheService := new(MockCacheService)
	pluginService := &PluginService{
		pluginRepo:     pluginRepo,
		storageService: storageService,
		cacheService:   cacheService,
	}

	ctx := context.Background()
	pluginID := uint(1)

	plugin := &model.Plugin{
		ID:          1,
		Name:        "test-plugin",
		DisplayName: "Test Plugin",
		Version:     "1.0.0",
		Category:    "utility",
		Status:      "published",
	}

	// 设置模拟
	pluginRepo.On("GetByID", ctx, pluginID).Return(plugin, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pluginService.GetPlugin(ctx, pluginID)
	}
}

func BenchmarkPluginService_SearchPlugins(b *testing.B) {
	// 设置基准测试环境
	pluginRepo := new(MockPluginRepository)
	storageService := new(MockStorageService)
	cacheService := new(MockCacheService)
	pluginService := &PluginService{
		pluginRepo:     pluginRepo,
		storageService: storageService,
		cacheService:   cacheService,
	}

	ctx := context.Background()
	query := "test"
	filters := map[string]interface{}{"category": "utility"}

	plugins := []*model.Plugin{
		{ID: 1, Name: "test-plugin", Category: "utility"},
		{ID: 2, Name: "another-test", Category: "utility"},
	}

	// 设置模拟
	pluginRepo.On("Search", ctx, query, filters, 0, 10).Return(plugins, int64(2), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pluginService.SearchPlugins(ctx, query, filters, 0, 10)
	}
}
