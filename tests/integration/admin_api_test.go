package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	adminAPI "laojun/admin-api/internal/handler"
	adminModel "laojun/admin-api/internal/model"
	adminRepo "laojun/admin-api/internal/repository"
	adminService "laojun/admin-api/internal/service"
	sharedTesting "laojun/shared/testing"
)

// AdminAPIIntegrationTestSuite Admin API 集成测试套件
type AdminAPIIntegrationTestSuite struct {
	sharedTesting.TestSuite
	userHandler *adminAPI.UserHandler
	authHandler *adminAPI.AuthHandler
	httpHelper  *sharedTesting.HTTPTestHelper
	testUser    *adminModel.User
	authToken   string
}

func (suite *AdminAPIIntegrationTestSuite) SetupSuite() {
	suite.TestSuite.SetupSuite()

	// 设置数据库表
	err := suite.DB.AutoMigrate(&adminModel.User{})
	assert.NoError(suite.T(), err)

	// 创建仓库和服务层
	userRepo := adminRepo.NewUserRepository(suite.DB)
	authRepo := adminRepo.NewAuthRepository(suite.DB, suite.Redis)

	userService := adminService.NewUserService(userRepo)
	authService := adminService.NewAuthService(authRepo, userRepo)

	// 创建处理层
	suite.userHandler = adminAPI.NewUserHandler(userService)
	suite.authHandler = adminAPI.NewAuthHandler(authService)

	// 设置路由
	suite.setupRoutes()

	// 创建HTTP测试助手
	suite.httpHelper = sharedTesting.NewHTTPTestHelper(suite.Router)
}

func (suite *AdminAPIIntegrationTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 清理数据
	suite.DB.Exec("DELETE FROM users")

	// 创建测试用户
	suite.createTestUser()
}

func (suite *AdminAPIIntegrationTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()

	// 清理数据
	suite.DB.Exec("DELETE FROM users")
}

func (suite *AdminAPIIntegrationTestSuite) setupRoutes() {
	// 认证相关路由
	auth := suite.Router.Group("/api/v1/auth")
	{
		auth.POST("/login", suite.authHandler.Login)
		auth.POST("/refresh", suite.authHandler.RefreshToken)
		auth.POST("/logout", suite.authHandler.Logout)
	}

	// 用户相关路由
	users := suite.Router.Group("/api/v1/users")
	{
		users.POST("", suite.userHandler.CreateUser)
		users.GET("/:id", suite.userHandler.GetUser)
		users.PUT("/:id", suite.userHandler.UpdateUser)
		users.DELETE("/:id", suite.userHandler.DeleteUser)
		users.GET("", suite.userHandler.ListUsers)
	}
}

func (suite *AdminAPIIntegrationTestSuite) createTestUser() {
	suite.testUser = &adminModel.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "admin",
		Status:   "active",
	}

	// 直接插入数据
	err := suite.DB.Create(suite.testUser).Error
	assert.NoError(suite.T(), err)

	// 登录获取token
	suite.loginAndGetToken()
}

func (suite *AdminAPIIntegrationTestSuite) loginAndGetToken() {
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

// TestUserCRUD 测试用户CRUD操作
func (suite *AdminAPIIntegrationTestSuite) TestUserCRUD() {
	// 1. 创建用户
	suite.Run("CreateUser", func() {
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

		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "newuser", data["username"])
		assert.Equal(suite.T(), "newuser@example.com", data["email"])
		assert.Equal(suite.T(), "user", data["role"])

		// 验证数据库中的数据
		var user adminModel.User
		err := suite.DB.Where("username = ?", "newuser").First(&user).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "newuser", user.Username)
		assert.Equal(suite.T(), "newuser@example.com", user.Email)
	})

	// 2. 获取用户
	suite.Run("GetUser", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", suite.testUser.ID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "testuser", data["username"])
		assert.Equal(suite.T(), "test@example.com", data["email"])
	})

	// 3. 更新用户
	suite.Run("UpdateUser", func() {
		updateData := map[string]interface{}{
			"email": "updated@example.com",
			"role":  "moderator",
		}

		body, _ := json.Marshal(updateData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", suite.testUser.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "updated@example.com", data["email"])
		assert.Equal(suite.T(), "moderator", data["role"])

		// 验证数据库中的数据
		var user adminModel.User
		err := suite.DB.First(&user, suite.testUser.ID).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "updated@example.com", user.Email)
		assert.Equal(suite.T(), "moderator", user.Role)
	})

	// 4. 用户列表
	suite.Run("ListUsers", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])

		data := response["data"].(map[string]interface{})
		users := data["users"].([]interface{})
		assert.GreaterOrEqual(suite.T(), len(users), 1)

		total := data["total"].(float64)
		assert.GreaterOrEqual(suite.T(), int(total), 1)
	})

	// 5. 删除用户
	suite.Run("DeleteUser", func() {
		// 创建一个要删除的用户
		deleteUser := &adminModel.User{
			Username: "deleteuser",
			Email:    "delete@example.com",
			Password: "password123",
			Role:     "user",
			Status:   "active",
		}
		err := suite.DB.Create(deleteUser).Error
		assert.NoError(suite.T(), err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", deleteUser.ID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])

		// 验证用户已被删除
		var user adminModel.User
		err = suite.DB.First(&user, deleteUser.ID).Error
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
	})
}

// TestAuthentication 测试认证流程
func (suite *AdminAPIIntegrationTestSuite) TestAuthentication() {
	// 1. 登录
	suite.Run("Login", func() {
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
		assert.Equal(suite.T(), "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(suite.T(), data["access_token"])
		assert.NotEmpty(suite.T(), data["refresh_token"])
		assert.Equal(suite.T(), "Bearer", data["token_type"])
	})

	// 2. 错误的登录凭证
	suite.Run("LoginWithWrongCredentials", func() {
		loginData := map[string]interface{}{
			"username": "testuser",
			"password": "wrongpassword",
		}

		body, _ := json.Marshal(loginData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "error", response["status"])
		assert.Contains(suite.T(), response["error"].(string), "用户名或密码错误")
	})

	// 3. 刷新令牌
	suite.Run("RefreshToken", func() {
		// 首先登录获取refresh token
		loginData := map[string]interface{}{
			"username": "testuser",
			"password": "password123",
		}

		body, _ := json.Marshal(loginData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		suite.Router.ServeHTTP(w, req)

		var loginResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		loginData = loginResponse["data"].(map[string]interface{})
		refreshToken := loginData["refresh_token"].(string)

		// 使用refresh token获取新的access token
		refreshData := map[string]interface{}{
			"refresh_token": refreshToken,
		}

		body, _ = json.Marshal(refreshData)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(suite.T(), data["access_token"])
		assert.NotEmpty(suite.T(), data["refresh_token"])
	})

	// 4. 登出
	suite.Run("Logout", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(suite.T(), "success", response["status"])
	})
}

// TestValidation 测试输入验证
func (suite *AdminAPIIntegrationTestSuite) TestValidation() {
	// 1. 创建用户时的验证
	suite.Run("CreateUserValidation", func() {
		testCases := []struct {
			name           string
			userData       map[string]interface{}
			expectedStatus int
			expectedError  string
		}{
			{
				name: "缺少用户名",
				userData: map[string]interface{}{
					"email":    "test@example.com",
					"password": "password123",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "用户名不能为空",
			},
			{
				name: "无效的邮箱格式",
				userData: map[string]interface{}{
					"username": "testuser",
					"email":    "invalid-email",
					"password": "password123",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "邮箱格式无效",
			},
			{
				name: "密码太短",
				userData: map[string]interface{}{
					"username": "testuser",
					"email":    "test@example.com",
					"password": "123",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "密码长度至少6个字符",
			},
			{
				name: "重复的用户名",
				userData: map[string]interface{}{
					"username": "testuser", // 已存在的用户用户名
					"email":    "another@example.com",
					"password": "password123",
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "用户名已存在",
			},
		}

		for _, tc := range testCases {
			suite.Run(tc.name, func() {
				body, _ := json.Marshal(tc.userData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.authToken)
				suite.Router.ServeHTTP(w, req)

				assert.Equal(suite.T(), tc.expectedStatus, w.Code)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"].(string), tc.expectedError)
			})
		}
	})
}

// TestConcurrency 测试并发操作
func (suite *AdminAPIIntegrationTestSuite) TestConcurrency() {
	suite.Run("ConcurrentUserCreation", func() {
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				userData := map[string]interface{}{
					"username": fmt.Sprintf("concurrent_user_%d", index),
					"email":    fmt.Sprintf("concurrent_%d@example.com", index),
					"password": "password123",
					"role":     "user",
				}

				body, _ := json.Marshal(userData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+suite.authToken)
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

		// 验证所有用户都被创建
		var count int64
		suite.DB.Model(&adminModel.User{}).Where("username LIKE ?", "concurrent_user_%").Count(&count)
		assert.Equal(suite.T(), int64(numGoroutines), count)
	})
}

// TestErrorHandling 测试错误处理
func (suite *AdminAPIIntegrationTestSuite) TestErrorHandling() {
	// 1. 未授权访问
	suite.Run("UnauthorizedAccess", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		// 不设置Authorization头
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	// 2. 无效的token
	suite.Run("InvalidToken", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	// 3. 访问不存在的用户
	suite.Run("UserNotFound", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users/99999", nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(suite.T(), response["error"].(string), "用户不存在")
	})

	// 4. 无效的JSON格式
	suite.Run("InvalidJSON", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})
}

// TestPerformance 测试性能
func (suite *AdminAPIIntegrationTestSuite) TestPerformance() {
	suite.Run("UserListPerformance", func() {
		// 创建大量测试数据
		const numUsers = 1000
		users := make([]*adminModel.User, numUsers)

		for i := 0; i < numUsers; i++ {
			users[i] = &adminModel.User{
				Username: fmt.Sprintf("perf_user_%d", i),
				Email:    fmt.Sprintf("perf_%d@example.com", i),
				Password: "password123",
				Role:     "user",
				Status:   "active",
			}
		}

		// 批量插入
		err := suite.DB.CreateInBatches(users, 100).Error
		assert.NoError(suite.T(), err)

		// 测试查询性能
		start := time.Now()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users?page=1&limit=50", nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		suite.Router.ServeHTTP(w, req)

		duration := time.Since(start)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		assert.Less(suite.T(), duration, 1*time.Second, "查询时间应该少于1秒")

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		users_result := data["users"].([]interface{})
		assert.Len(suite.T(), users_result, 50)
	})
}

// 运行测试套件
func TestAdminAPIIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	suite.Run(t, new(AdminAPIIntegrationTestSuite))
}
