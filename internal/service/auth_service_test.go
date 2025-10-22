package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun/admin-api/internal/model"
	sharedTesting "github.com/codetaoist/laojun/shared/testing"
)

// MockAuthRepository 模拟认证仓库
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthRepository) CreateSession(ctx context.Context, session *model.UserSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockAuthRepository) GetSession(ctx context.Context, token string) (*model.UserSession, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserSession), args.Error(1)
}

func (m *MockAuthRepository) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthRepository) DeleteUserSessions(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// AuthServiceTestSuite 认证服务测试套件
type AuthServiceTestSuite struct {
	sharedTesting.TestSuite
	authRepo    *MockAuthRepository
	authService *AuthService
	jwtSecret   string
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 创建模拟仓库
	suite.authRepo = new(MockAuthRepository)
	suite.jwtSecret = "test-jwt-secret-key"

	// 创建认证服务
	suite.authService = &AuthService{
		authRepo:  suite.authRepo,
		jwtSecret: suite.jwtSecret,
	}
}

func (suite *AuthServiceTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()
	suite.authRepo.AssertExpectations(suite.T())
}

// TestLogin 测试用户登录
func (suite *AuthServiceTestSuite) TestLogin() {
	ctx := context.Background()

	// 创建测试用户密码哈希
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	testCases := []struct {
		name        string
		input       *LoginRequest
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "邮箱登录成功",
			input: &LoginRequest{
				Email:    "test@example.com",
				Password: password,
			},
			setupMocks: func() {
				user := &model.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "user",
					Status:       "active",
				}
				suite.authRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
				suite.authRepo.On("UpdateLastLogin", ctx, uint(1)).Return(nil)
				suite.authRepo.On("CreateSession", ctx, mock.AnythingOfType("*model.UserSession")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "用户名登录成功",
			input: &LoginRequest{
				Username: "testuser",
				Password: password,
			},
			setupMocks: func() {
				user := &model.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "user",
					Status:       "active",
				}
				suite.authRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
				suite.authRepo.On("UpdateLastLogin", ctx, uint(1)).Return(nil)
				suite.authRepo.On("CreateSession", ctx, mock.AnythingOfType("*model.UserSession")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "用户不存在或密码错误",
			input: &LoginRequest{
				Email:    "nonexistent@example.com",
				Password: password,
			},
			setupMocks: func() {
				suite.authRepo.On("GetByEmail", ctx, "nonexistent@example.com").Return((*model.User)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "用户不存在或密码错误",
		},
		{
			name: "密码错误",
			input: &LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func() {
				user := &model.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "user",
					Status:       "active",
				}
				suite.authRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
			},
			expectError: true,
			errorMsg:    "用户不存在或密码错误",
		},
		{
			name: "用户已禁用",
			input: &LoginRequest{
				Email:    "test@example.com",
				Password: password,
			},
			setupMocks: func() {
				user := &model.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "user",
					Status:       "disabled",
				}
				suite.authRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
			},
			expectError: true,
			errorMsg:    "用户账户已被禁用",
		},
		{
			name: "缺少登录凭据",
			input: &LoginRequest{
				Password: password,
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "请提供邮箱或用户名",
		},
		{
			name: "密码为空",
			input: &LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "密码不能为空",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.authRepo.ExpectedCalls = nil
			suite.authRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			response, err := suite.authService.Login(ctx, tc.input)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), response)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), response)
				assert.NotEmpty(suite.T(), response.AccessToken)
				assert.NotEmpty(suite.T(), response.RefreshToken)
				assert.NotNil(suite.T(), response.User)
				assert.Greater(suite.T(), response.ExpiresIn, int64(0))
			}
		})
	}
}

// TestRefreshToken 测试刷新令牌
func (suite *AuthServiceTestSuite) TestRefreshToken() {
	ctx := context.Background()

	// 创建有效的刷新令牌和过期刷新令牌
	validRefreshToken := suite.generateTestRefreshToken(1, "testuser")
	expiredRefreshToken := suite.generateExpiredTestRefreshToken(1, "testuser")

	testCases := []struct {
		name        string
		token       string
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:  "成功刷新令牌",
			token: validRefreshToken,
			setupMocks: func() {
				session := &model.UserSession{
					ID:           1,
					UserID:       1,
					RefreshToken: validRefreshToken,
					ExpiresAt:    time.Now().Add(24 * time.Hour),
					IsActive:     true,
				}
				user := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Role:     "user",
					Status:   "active",
				}
				suite.authRepo.On("GetSession", ctx, validRefreshToken).Return(session, nil)
				suite.authRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
				suite.authRepo.On("DeleteSession", ctx, validRefreshToken).Return(nil)
				suite.authRepo.On("CreateSession", ctx, mock.AnythingOfType("*model.UserSession")).Return(nil)
			},
			expectError: false,
		},
		{
			name:  "令牌不存在",
			token: "invalid-token",
			setupMocks: func() {
				suite.authRepo.On("GetSession", ctx, "invalid-token").Return((*model.UserSession)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "无效的刷新令牌",
		},
		{
			name:  "令牌已过期",
			token: expiredRefreshToken,
			setupMocks: func() {
				session := &model.UserSession{
					ID:           1,
					UserID:       1,
					RefreshToken: expiredRefreshToken,
					ExpiresAt:    time.Now().Add(-1 * time.Hour), // 已过期
					IsActive:     true,
				}
				suite.authRepo.On("GetSession", ctx, expiredRefreshToken).Return(session, nil)
			},
			expectError: true,
			errorMsg:    "刷新令牌已过期",
		},
		{
			name:  "令牌已失效",
			token: validRefreshToken,
			setupMocks: func() {
				session := &model.UserSession{
					ID:           1,
					UserID:       1,
					RefreshToken: validRefreshToken,
					ExpiresAt:    time.Now().Add(24 * time.Hour),
					IsActive:     false, // 已失效
				}
				suite.authRepo.On("GetSession", ctx, validRefreshToken).Return(session, nil)
			},
			expectError: true,
			errorMsg:    "刷新令牌已失效",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.authRepo.ExpectedCalls = nil
			suite.authRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			response, err := suite.authService.RefreshToken(ctx, tc.token)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), response)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), response)
				assert.NotEmpty(suite.T(), response.AccessToken)
				assert.NotEmpty(suite.T(), response.RefreshToken)
				assert.NotEqual(suite.T(), tc.token, response.RefreshToken) // 新的刷新令牌应该不同
			}
		})
	}
}

// TestLogout 测试用户登出
func (suite *AuthServiceTestSuite) TestLogout() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		token       string
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:  "成功登出",
			token: "valid-refresh-token",
			setupMocks: func() {
				suite.authRepo.On("DeleteSession", ctx, "valid-refresh-token").Return(nil)
			},
			expectError: false,
		},
		{
			name:  "令牌不存在（仍然成功）",
			token: "nonexistent-token",
			setupMocks: func() {
				suite.authRepo.On("DeleteSession", ctx, "nonexistent-token").Return(gorm.ErrRecordNotFound)
			},
			expectError: false, // 登出操作应该是幂等的
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.authRepo.ExpectedCalls = nil
			suite.authRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			err := suite.authService.Logout(ctx, tc.token)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestValidateToken 测试令牌验证
func (suite *AuthServiceTestSuite) TestValidateToken() {
	ctx := context.Background()

	// 创建有效的访问令牌和过期令牌
	validAccessToken := suite.generateTestAccessToken(1, "testuser", "user")
	expiredAccessToken := suite.generateExpiredTestAccessToken(1, "testuser", "user")

	testCases := []struct {
		name        string
		token       string
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:  "有效令牌",
			token: validAccessToken,
			setupMocks: func() {
				user := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Role:     "user",
					Status:   "active",
				}
				suite.authRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
			},
			expectError: false,
		},
		{
			name:        "无效令牌格式",
			token:       "invalid-token",
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "无效的访问令牌",
		},
		{
			name:        "令牌已过期",
			token:       expiredAccessToken,
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "令牌已过期",
		},
		{
			name:  "用户不存在",
			token: validAccessToken,
			setupMocks: func() {
				suite.authRepo.On("GetByID", ctx, uint(1)).Return((*model.User)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "用户不存在",
		},
		{
			name:  "用户已禁用",
			token: validAccessToken,
			setupMocks: func() {
				user := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Role:     "user",
					Status:   "disabled",
				}
				suite.authRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
			},
			expectError: true,
			errorMsg:    "用户账户已被禁用",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.authRepo.ExpectedCalls = nil
			suite.authRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			user, err := suite.authService.ValidateToken(ctx, tc.token)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), user)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), user)
				assert.Equal(suite.T(), uint(1), user.ID)
				assert.Equal(suite.T(), "testuser", user.Username)
			}
		})
	}
}

// TestLogoutAllSessions 测试登出所有会话
func (suite *AuthServiceTestSuite) TestLogoutAllSessions() {
	ctx := context.Background()
	userID := uint(1)

	testCases := []struct {
		name        string
		setupMocks  func()
		expectError bool
	}{
		{
			name: "成功登出所有会话",
			setupMocks: func() {
				suite.authRepo.On("DeleteUserSessions", ctx, userID).Return(nil)
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.authRepo.ExpectedCalls = nil
			suite.authRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			err := suite.authService.LogoutAllSessions(ctx, userID)

			if tc.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// 辅助方法：生成测试访问令牌
func (suite *AuthServiceTestSuite) generateTestAccessToken(userID uint, username, role string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"type":     "access",
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(suite.jwtSecret))
	return tokenString
}

// 辅助方法：生成过期的测试访问令牌
func (suite *AuthServiceTestSuite) generateExpiredTestAccessToken(userID uint, username, role string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"type":     "access",
		"exp":      time.Now().Add(-1 * time.Hour).Unix(), // 已过期
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(suite.jwtSecret))
	return tokenString
}

// 辅助方法：生成测试刷新令牌
func (suite *AuthServiceTestSuite) generateTestRefreshToken(userID uint, username string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"type":     "refresh",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(suite.jwtSecret))
	return tokenString
}

// 辅助方法：生成过期的测试刷新令牌
func (suite *AuthServiceTestSuite) generateExpiredTestRefreshToken(userID uint, username string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"type":     "refresh",
		"exp":      time.Now().Add(-1 * time.Hour).Unix(), // 已过期
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(suite.jwtSecret))
	return tokenString
}

// 运行测试套件
func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// 基准测试
func BenchmarkAuthService_Login(b *testing.B) {
	// 设置基准测试环境
	authRepo := new(MockAuthRepository)
	authService := &AuthService{
		authRepo:  authRepo,
		jwtSecret: "test-jwt-secret",
	}

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		ID:           1,
		Username:     "benchuser",
		Email:        "bench@example.com",
		PasswordHash: string(hashedPassword),
		Role:         "user",
		Status:       "active",
	}

	req := &LoginRequest{
		Email:    "bench@example.com",
		Password: password,
	}

	// 设置模拟
	authRepo.On("GetByEmail", ctx, "bench@example.com").Return(user, nil)
	authRepo.On("UpdateLastLogin", ctx, uint(1)).Return(nil)
	authRepo.On("CreateSession", ctx, mock.AnythingOfType("*model.UserSession")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authService.Login(ctx, req)
	}
}

func BenchmarkAuthService_ValidateToken(b *testing.B) {
	// 设置基准测试环境
	authRepo := new(MockAuthRepository)
	authService := &AuthService{
		authRepo:  authRepo,
		jwtSecret: "test-jwt-secret",
	}

	ctx := context.Background()

	// 创建测试令牌
	claims := jwt.MapClaims{
		"user_id":  1,
		"username": "benchuser",
		"role":     "user",
		"type":     "access",
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-jwt-secret"))

	user := &model.User{
		ID:       1,
		Username: "benchuser",
		Email:    "bench@example.com",
		Role:     "user",
		Status:   "active",
	}

	// 设置模拟
	authRepo.On("GetByID", ctx, uint(1)).Return(user, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authService.ValidateToken(ctx, tokenString)
	}
}
