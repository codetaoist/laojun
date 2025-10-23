package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codetaoist/laojun/internal/clients"
	"github.com/codetaoist/laojun/internal/handlers"
	"github.com/codetaoist/laojun/internal/models"
	"github.com/codetaoist/laojun/internal/services"
	"github.com/codetaoist/laojun/pkg/shared/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MarketplaceAdminIntegrationTestSuite 插件市场与总后台集成测试套件
type MarketplaceAdminIntegrationTestSuite struct {
	suite.Suite
	db                    *database.DB
	marketplaceClient     *clients.MarketplaceClient
	adminPluginService    *services.AdminPluginService
	adminPluginHandler    *handlers.AdminPluginHandler
	marketplaceHandler    *handlers.MarketplaceHandler
	router                *gin.Engine
	testPlugin            *models.Plugin
	testDeveloper         *models.Developer
	testCategory          *models.Category
}

// SetupSuite 设置测试套件
func (suite *MarketplaceAdminIntegrationTestSuite) SetupSuite() {
	// 设置测试数据库
	suite.db = setupTestDatabase()
	
	// 创建插件市场客户端
	suite.marketplaceClient = clients.NewMarketplaceClient(
		"http://localhost:8081", // 测试插件市场API地址
		"test-api-key",
	)
	
	// 创建服务
	suite.adminPluginService = services.NewAdminPluginService(suite.db, suite.marketplaceClient)
	
	// 创建处理器
	suite.adminPluginHandler = handlers.NewAdminPluginHandler(suite.adminPluginService, nil)
	suite.marketplaceHandler = handlers.NewMarketplaceHandler(nil, nil)
	
	// 设置路由
	suite.router = gin.New()
	suite.setupRoutes()
	
	// 创建测试数据
	suite.createTestData()
}

// TearDownSuite 清理测试套件
func (suite *MarketplaceAdminIntegrationTestSuite) TearDownSuite() {
	// 清理测试数据
	suite.cleanupTestData()
	
	// 关闭数据库连接
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

// setupRoutes 设置测试路由
func (suite *MarketplaceAdminIntegrationTestSuite) setupRoutes() {
	// 插件市场API路由
	marketplaceAPI := suite.router.Group("/marketplace/api/v1")
	{
		marketplaceAPI.GET("/plugins", suite.marketplaceHandler.GetPlugins)
		marketplaceAPI.GET("/plugins/:id", suite.marketplaceHandler.GetPlugin)
		marketplaceAPI.POST("/plugins", suite.marketplaceHandler.PublishPlugin)
		marketplaceAPI.PUT("/plugins/:id", suite.marketplaceHandler.UpdatePlugin)
		marketplaceAPI.DELETE("/plugins/:id", suite.marketplaceHandler.DeletePlugin)
		marketplaceAPI.POST("/plugins/:id/install", suite.marketplaceHandler.InstallPlugin)
		marketplaceAPI.DELETE("/plugins/:id/uninstall", suite.marketplaceHandler.UninstallPlugin)
	}
	
	// 总后台API路由
	adminAPI := suite.router.Group("/admin/api/v1")
	{
		adminAPI.GET("/plugins", suite.adminPluginHandler.GetPluginsForAdmin)
		adminAPI.GET("/plugins/:id", suite.adminPluginHandler.GetPluginForAdmin)
		adminAPI.PUT("/plugins/:id/status", suite.adminPluginHandler.UpdatePluginStatus)
		adminAPI.POST("/plugins/:id/sync-from-marketplace", suite.adminPluginHandler.SyncPluginFromMarketplace)
		adminAPI.POST("/plugins/:id/sync-to-marketplace", suite.adminPluginHandler.SyncPluginToMarketplace)
		adminAPI.GET("/marketplace/plugins", suite.adminPluginHandler.SearchMarketplacePlugins)
		adminAPI.POST("/marketplace/plugins/:id/install", suite.adminPluginHandler.InstallMarketplacePlugin)
	}
}

// createTestData 创建测试数据
func (suite *MarketplaceAdminIntegrationTestSuite) createTestData() {
	// 创建测试分类
	suite.testCategory = &models.Category{
		ID:          uuid.New(),
		Name:        "测试分类",
		Description: "用于测试的插件分类",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	suite.db.Create(suite.testCategory)
	
	// 创建测试开发者
	suite.testDeveloper = &models.Developer{
		ID:        uuid.New(),
		Name:      "测试开发者",
		Email:     "test@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	suite.db.Create(suite.testDeveloper)
	
	// 创建测试插件
	suite.testPlugin = &models.Plugin{
		ID:          uuid.New(),
		Name:        "测试插件",
		Description: "用于测试的插件",
		Version:     "1.0.0",
		Status:      "published",
		CategoryID:  suite.testCategory.ID,
		DeveloperID: suite.testDeveloper.ID,
		Price:       0.0,
		Downloads:   100,
		Rating:      4.5,
		ReviewCount: 10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	suite.db.Create(suite.testPlugin)
}

// cleanupTestData 清理测试数据
func (suite *MarketplaceAdminIntegrationTestSuite) cleanupTestData() {
	if suite.testPlugin != nil {
		suite.db.Delete(suite.testPlugin)
	}
	if suite.testDeveloper != nil {
		suite.db.Delete(suite.testDeveloper)
	}
	if suite.testCategory != nil {
		suite.db.Delete(suite.testCategory)
	}
}

// TestPluginSyncFromMarketplace 测试从插件市场同步插件
func (suite *MarketplaceAdminIntegrationTestSuite) TestPluginSyncFromMarketplace() {
	// 准备请求
	url := fmt.Sprintf("/admin/api/v1/plugins/%s/sync-from-marketplace", suite.testPlugin.ID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
}

// TestPluginSyncToMarketplace 测试同步插件到插件市场
func (suite *MarketplaceAdminIntegrationTestSuite) TestPluginSyncToMarketplace() {
	// 准备请求
	url := fmt.Sprintf("/admin/api/v1/plugins/%s/sync-to-marketplace", suite.testPlugin.ID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
}

// TestAdminPluginStatusUpdate 测试管理员更新插件状态
func (suite *MarketplaceAdminIntegrationTestSuite) TestAdminPluginStatusUpdate() {
	// 准备请求数据
	requestData := map[string]interface{}{
		"status": "suspended",
		"reason": "违反平台规则",
	}
	jsonData, _ := json.Marshal(requestData)
	
	// 准备请求
	url := fmt.Sprintf("/admin/api/v1/plugins/%s/status", suite.testPlugin.ID)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	// 验证数据库中的状态已更新
	var updatedPlugin models.Plugin
	suite.db.First(&updatedPlugin, "id = ?", suite.testPlugin.ID)
	assert.Equal(suite.T(), "suspended", updatedPlugin.Status)
}

// TestMarketplacePluginSearch 测试插件市场搜索
func (suite *MarketplaceAdminIntegrationTestSuite) TestMarketplacePluginSearch() {
	// 准备请求
	url := "/admin/api/v1/marketplace/plugins?search=测试&page=1&limit=10"
	req, _ := http.NewRequest("GET", url, nil)
	
	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	// 验证返回的数据结构
	data, exists := response["data"]
	assert.True(suite.T(), exists)
	
	dataMap, ok := data.(map[string]interface{})
	assert.True(suite.T(), ok)
	
	_, hasPlugins := dataMap["plugins"]
	assert.True(suite.T(), hasPlugins)
	
	_, hasMeta := dataMap["meta"]
	assert.True(suite.T(), hasMeta)
}

// TestPluginInstallFromMarketplace 测试从插件市场安装插件
func (suite *MarketplaceAdminIntegrationTestSuite) TestPluginInstallFromMarketplace() {
	// 准备请求数据
	requestData := map[string]interface{}{
		"version": "1.0.0",
	}
	jsonData, _ := json.Marshal(requestData)
	
	// 准备请求
	url := fmt.Sprintf("/admin/api/v1/marketplace/plugins/%s/install", suite.testPlugin.ID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
}

// TestBatchPluginSync 测试批量插件同步
func (suite *MarketplaceAdminIntegrationTestSuite) TestBatchPluginSync() {
	// 准备请求数据
	requestData := map[string]interface{}{
		"plugin_ids": []string{suite.testPlugin.ID.String()},
		"direction":  "to_marketplace",
	}
	jsonData, _ := json.Marshal(requestData)
	
	// 准备请求
	url := "/admin/api/v1/plugins/batch-sync"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
}

// TestAdminDashboardStats 测试管理员仪表板统计
func (suite *MarketplaceAdminIntegrationTestSuite) TestAdminDashboardStats() {
	// 准备请求
	url := "/admin/api/v1/dashboard/stats"
	req, _ := http.NewRequest("GET", url, nil)
	
	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	// 验证统计数据结构
	data, exists := response["data"]
	assert.True(suite.T(), exists)
	
	statsMap, ok := data.(map[string]interface{})
	assert.True(suite.T(), ok)
	
	// 验证必要的统计字段
	expectedFields := []string{
		"total_plugins", "today_plugins", "pending_reviews",
		"total_downloads", "total_revenue", "active_developers",
	}
	
	for _, field := range expectedFields {
		_, exists := statsMap[field]
		assert.True(suite.T(), exists, fmt.Sprintf("Missing field: %s", field))
	}
}

// TestPluginConfigManagement 测试插件配置管理
func (suite *MarketplaceAdminIntegrationTestSuite) TestPluginConfigManagement() {
	// 测试获取插件配置
	url := fmt.Sprintf("/admin/api/v1/plugins/%s/config", suite.testPlugin.ID)
	req, _ := http.NewRequest("GET", url, nil)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 测试更新插件配置
	configData := map[string]interface{}{
		"config": map[string]interface{}{
			"enabled":     true,
			"max_users":   1000,
			"debug_mode":  false,
			"api_timeout": 30,
		},
	}
	jsonData, _ := json.Marshal(configData)
	
	req, _ = http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
}

// TestPluginStatsAndLogs 测试插件统计和日志
func (suite *MarketplaceAdminIntegrationTestSuite) TestPluginStatsAndLogs() {
	// 测试获取插件统计
	statsURL := fmt.Sprintf("/admin/api/v1/plugins/%s/stats", suite.testPlugin.ID)
	req, _ := http.NewRequest("GET", statsURL, nil)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var statsResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &statsResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", statsResponse["status"])
	
	// 测试获取插件日志
	logsURL := fmt.Sprintf("/admin/api/v1/plugins/%s/logs?page=1&limit=10", suite.testPlugin.ID)
	req, _ = http.NewRequest("GET", logsURL, nil)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var logsResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &logsResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", logsResponse["status"])
}

// setupTestDatabase 设置测试数据库
func setupTestDatabase() *database.DB {
	// 这里应该设置测试数据库连接
	// 为了示例，返回nil，实际使用时需要配置真实的测试数据库
	return nil
}

// TestMarketplaceAdminIntegration 运行集成测试
func TestMarketplaceAdminIntegration(t *testing.T) {
	suite.Run(t, new(MarketplaceAdminIntegrationTestSuite))
}

// BenchmarkPluginSync 插件同步性能测试
func BenchmarkPluginSync(b *testing.B) {
	// 设置基准测试环境
	suite := &MarketplaceAdminIntegrationTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 执行插件同步操作
		url := fmt.Sprintf("/admin/api/v1/plugins/%s/sync-to-marketplace", suite.testPlugin.ID)
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkMarketplaceSearch 插件市场搜索性能测试
func BenchmarkMarketplaceSearch(b *testing.B) {
	// 设置基准测试环境
	suite := &MarketplaceAdminIntegrationTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 执行搜索操作
		url := "/admin/api/v1/marketplace/plugins?search=test&page=1&limit=20"
		req, _ := http.NewRequest("GET", url, nil)
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}