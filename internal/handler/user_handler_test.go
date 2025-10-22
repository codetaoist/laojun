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

	"github.com/codetaoist/laojun/admin-api/internal/model"
	"github.com/codetaoist/laojun/admin-api/internal/service"
	sharedTesting "github.com/codetaoist/laojun/shared/testing"
)

// MockUserService 模拟用户服务
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *service.CreateUserRequest) (*model.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint, req *service.UpdateUserRequest) (*model.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

// UserHandlerTestSuite 用户处理器测试套
type UserHandlerTestSuite struct {
	sharedTesting.TestSuite
	userService *MockUserService
	userHandler *UserHandler
	router      *gin.Engine
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	suite.userService = new(MockUserService)

	// 创建处理器
	suite.userHandler = &UserHandler{
		userService: suite.userService,
	}

	// 创建路由
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *UserHandlerTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()
	suite.userService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", suite.userHandler.CreateUser)
			users.GET("/:id", suite.userHandler.GetUser)
			users.PUT("/:id", suite.userHandler.UpdateUser)
			users.DELETE("/:id", suite.userHandler.DeleteUser)
			users.GET("", suite.userHandler.ListUsers)
		}
	}
}

// TestCreateUser 测试创建用户
func (suite *UserHandlerTestSuite) TestCreateUser() {
	testCases := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功创建用户",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
				"role":     "user",
			},
			setupMocks: func() {
				req := &service.CreateUserRequest{
					Username: "testuser",
					Email:    "test@example.com",
					Password: "password123",
					Role:     "user",
				}
				user := &model.User{
					ID:        1,
					Username:  "testuser",
					Email:     "test@example.com",
					Role:      "user",
					Status:    "active",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.userService.On("CreateUser", mock.Anything, req).Return(user, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "无效的请求体",
			requestBody: map[string]interface{}{
				"username": "",
				"email":    "invalid-email",
				"password": "123",
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "用户名已存在",
			requestBody: map[string]interface{}{
				"username": "existinguser",
				"email":    "test@example.com",
				"password": "password123",
				"role":     "user",
			},
			setupMocks: func() {
				req := &service.CreateUserRequest{
					Username: "existinguser",
					Email:    "test@example.com",
					Password: "password123",
					Role:     "user",
				}
				suite.userService.On("CreateUser", mock.Anything, req).Return((*model.User)(nil), fmt.Errorf("用户名已存在"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "用户名已存在",
		},
		{
			name:           "无效的JSON格式",
			requestBody:    "invalid json",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userService.ExpectedCalls = nil
			suite.userService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 准备请求
			var body []byte
			if str, ok := tc.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tc.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// 执行请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"], tc.expectedError)
			}

			if w.Code == http.StatusCreated {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])
			}
		})
	}
}

// TestGetUser 测试获取用户
func (suite *UserHandlerTestSuite) TestGetUser() {
	testCases := []struct {
		name           string
		userID         string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功获取用户",
			userID: "1",
			setupMocks: func() {
				user := &model.User{
					ID:        1,
					Username:  "testuser",
					Email:     "test@example.com",
					Role:      "user",
					Status:    "active",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.userService.On("GetUser", mock.Anything, uint(1)).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "用户不存在",
			userID: "999",
			setupMocks: func() {
				suite.userService.On("GetUser", mock.Anything, uint(999)).Return((*model.User)(nil), fmt.Errorf("用户不存在"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "用户不存在",
		},
		{
			name:           "无效的用户ID",
			userID:         "invalid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userService.ExpectedCalls = nil
			suite.userService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 准备请求
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tc.userID, nil)

			// 执行请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"], tc.expectedError)
			}

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])
			}
		})
	}
}

// TestUpdateUser 测试更新用户
func (suite *UserHandlerTestSuite) TestUpdateUser() {
	testCases := []struct {
		name           string
		userID         string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功更新用户",
			userID: "1",
			requestBody: map[string]interface{}{
				"email":  "newemail@example.com",
				"role":   "admin",
				"status": "active",
			},
			setupMocks: func() {
				req := &service.UpdateUserRequest{
					Email:  "newemail@example.com",
					Role:   "admin",
					Status: "active",
				}
				user := &model.User{
					ID:        1,
					Username:  "testuser",
					Email:     "newemail@example.com",
					Role:      "admin",
					Status:    "active",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.userService.On("UpdateUser", mock.Anything, uint(1), req).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "用户不存在",
			userID: "999",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			setupMocks: func() {
				req := &service.UpdateUserRequest{
					Email: "test@example.com",
				}
				suite.userService.On("UpdateUser", mock.Anything, uint(999), req).Return((*model.User)(nil), fmt.Errorf("用户不存在"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "用户不存在",
		},
		{
			name:           "无效的用户ID",
			userID:         "invalid",
			requestBody:    map[string]interface{}{},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "无效的邮箱格式",
			userID: "1",
			requestBody: map[string]interface{}{
				"email": "invalid-email",
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userService.ExpectedCalls = nil
			suite.userService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 准备请求
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+tc.userID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// 执行请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"], tc.expectedError)
			}

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])
			}
		})
	}
}

// TestDeleteUser 测试删除用户
func (suite *UserHandlerTestSuite) TestDeleteUser() {
	testCases := []struct {
		name           string
		userID         string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功删除用户",
			userID: "1",
			setupMocks: func() {
				suite.userService.On("DeleteUser", mock.Anything, uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "用户不存在",
			userID: "999",
			setupMocks: func() {
				suite.userService.On("DeleteUser", mock.Anything, uint(999)).Return(fmt.Errorf("用户不存在"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "用户不存在",
		},
		{
			name:           "无效的用户ID",
			userID:         "invalid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userService.ExpectedCalls = nil
			suite.userService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 准备请求
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+tc.userID, nil)

			// 执行请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"], tc.expectedError)
			}

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
			}
		})
	}
}

// TestListUsers 测试用户列表
func (suite *UserHandlerTestSuite) TestListUsers() {
	testCases := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "成功获取用户列表",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				users := []*model.User{
					{ID: 1, Username: "user1", Email: "user1@example.com"},
					{ID: 2, Username: "user2", Email: "user2@example.com"},
					{ID: 3, Username: "user3", Email: "user3@example.com"},
				}
				suite.userService.On("ListUsers", mock.Anything, 0, 10).Return(users, int64(3), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "默认分页参数",
			queryParams: "",
			setupMocks: func() {
				users := []*model.User{}
				suite.userService.On("ListUsers", mock.Anything, 0, 20).Return(users, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "无效的分页参数",
			queryParams: "?page=invalid&limit=10",
			setupMocks: func() {
				users := []*model.User{}
				suite.userService.On("ListUsers", mock.Anything, 0, 10).Return(users, int64(0), nil)
			},
			expectedStatus: http.StatusOK, // 应该使用默认值
		},
		{
			name:        "超出限制的limit",
			queryParams: "?page=1&limit=1000",
			setupMocks: func() {
				users := []*model.User{}
				suite.userService.On("ListUsers", mock.Anything, 0, 100).Return(users, int64(0), nil) // 应该被限制为100
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userService.ExpectedCalls = nil
			suite.userService.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 准备请求
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tc.queryParams, nil)

			// 执行请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(suite.T(), response["error"], tc.expectedError)
			}

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "success", response["status"])
				assert.NotNil(suite.T(), response["data"])

				data := response["data"].(map[string]interface{})
				assert.NotNil(suite.T(), data["users"])
				assert.NotNil(suite.T(), data["total"])
				assert.NotNil(suite.T(), data["page"])
				assert.NotNil(suite.T(), data["limit"])
			}
		})
	}
}

// TestValidation 测试输入验证
func (suite *UserHandlerTestSuite) TestValidation() {
	testCases := []struct {
		name        string
		requestBody map[string]interface{}
		endpoint    string
		method      string
		expectError bool
		errorField  string
	}{
		{
			name: "用户名不能为空",
			requestBody: map[string]interface{}{
				"username": "",
				"email":    "test@example.com",
				"password": "password123",
			},
			endpoint:    "/api/v1/users",
			method:      http.MethodPost,
			expectError: true,
			errorField:  "username",
		},
		{
			name: "邮箱格式无效",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"email":    "invalid-email",
				"password": "password123",
			},
			endpoint:    "/api/v1/users",
			method:      http.MethodPost,
			expectError: true,
			errorField:  "email",
		},
		{
			name: "密码太短",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "123",
			},
			endpoint:    "/api/v1/users",
			method:      http.MethodPost,
			expectError: true,
			errorField:  "password",
		},
		{
			name: "无效的角色值",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
				"role":     "invalid_role",
			},
			endpoint:    "/api/v1/users",
			method:      http.MethodPost,
			expectError: true,
			errorField:  "role",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 准备请求
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// 执行请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			if tc.expectError {
				assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(suite.T(), "error", response["status"])
				assert.Contains(suite.T(), response["error"], tc.errorField)
			}
		})
	}
}

// 运行测试套件
func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

// 基准测试
func BenchmarkUserHandler_CreateUser(b *testing.B) {
	gin.SetMode(gin.TestMode)

	userService := new(MockUserService)
	userHandler := &UserHandler{userService: userService}

	router := gin.New()
	router.POST("/users", userHandler.CreateUser)

	req := &service.CreateUserRequest{
		Username: "benchuser",
		Email:    "bench@example.com",
		Password: "password123",
		Role:     "user",
	}

	user := &model.User{
		ID:       1,
		Username: "benchuser",
		Email:    "bench@example.com",
		Role:     "user",
		Status:   "active",
	}

	userService.On("CreateUser", mock.Anything, req).Return(user, nil)

	requestBody := map[string]interface{}{
		"username": "benchuser",
		"email":    "bench@example.com",
		"password": "password123",
		"role":     "user",
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
