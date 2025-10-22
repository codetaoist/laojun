package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun/admin-api/internal/model"
	sharedTesting "github.com/codetaoist/laojun/shared/testing"
)

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
	sharedTesting.TestSuite
	userRepo    *MockUserRepository
	userService *UserService
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()

	// 创建模拟仓库
	suite.userRepo = new(MockUserRepository)

	// 创建用户服务
	suite.userService = &UserService{
		userRepo: suite.userRepo,
	}
}

func (suite *UserServiceTestSuite) TearDownTest() {
	suite.TestSuite.TearDownTest()
	suite.userRepo.AssertExpectations(suite.T())
}

// TestCreateUser 测试创建用户
func (suite *UserServiceTestSuite) TestCreateUser() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		input       *CreateUserRequest
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功创建用户",
			input: &CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Role:     "user",
			},
			setupMocks: func() {
				// 检查用户名不存在
				suite.userRepo.On("GetByUsername", ctx, "testuser").Return((*model.User)(nil), gorm.ErrRecordNotFound)
				// 检查邮箱不存在
				suite.userRepo.On("GetByEmail", ctx, "test@example.com").Return((*model.User)(nil), gorm.ErrRecordNotFound)
				// 创建用户
				suite.userRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "用户名已存在",
			input: &CreateUserRequest{
				Username: "existinguser",
				Email:    "test@example.com",
				Password: "password123",
				Role:     "user",
			},
			setupMocks: func() {
				existingUser := &model.User{
					ID:       1,
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				suite.userRepo.On("GetByUsername", ctx, "existinguser").Return(existingUser, nil)
			},
			expectError: true,
			errorMsg:    "用户名已存在",
		},
		{
			name: "邮箱已存在",
			input: &CreateUserRequest{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "password123",
				Role:     "user",
			},
			setupMocks: func() {
				// 用户名不存在
				suite.userRepo.On("GetByUsername", ctx, "testuser").Return((*model.User)(nil), gorm.ErrRecordNotFound)
				// 邮箱已存在
				existingUser := &model.User{
					ID:       1,
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				suite.userRepo.On("GetByEmail", ctx, "existing@example.com").Return(existingUser, nil)
			},
			expectError: true,
			errorMsg:    "邮箱已存在",
		},
		{
			name: "密码太短",
			input: &CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123",
				Role:     "user",
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "密码长度至少为8个字符",
		},
		{
			name: "无效的邮箱格式",
			input: &CreateUserRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
				Role:     "user",
			},
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "邮箱格式无效",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userRepo.ExpectedCalls = nil
			suite.userRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			user, err := suite.userService.CreateUser(ctx, tc.input)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), user)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), user)
				assert.Equal(suite.T(), tc.input.Username, user.Username)
				assert.Equal(suite.T(), tc.input.Email, user.Email)
				assert.Equal(suite.T(), tc.input.Role, user.Role)
				assert.NotEmpty(suite.T(), user.PasswordHash)
				assert.NotEqual(suite.T(), tc.input.Password, user.PasswordHash) // 密码应该被哈希存储
			}
		})
	}
}

// TestGetUser 测试获取用户
func (suite *UserServiceTestSuite) TestGetUser() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		userID      uint
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:   "成功获取用户",
			userID: 1,
			setupMocks: func() {
				user := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Role:     "user",
					Status:   "active",
				}
				suite.userRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
			},
			expectError: false,
		},
		{
			name:   "用户不存在",
			userID: 999,
			setupMocks: func() {
				suite.userRepo.On("GetByID", ctx, uint(999)).Return((*model.User)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "用户不存在",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userRepo.ExpectedCalls = nil
			suite.userRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			user, err := suite.userService.GetUser(ctx, tc.userID)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), user)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), user)
				assert.Equal(suite.T(), tc.userID, user.ID)
			}
		})
	}
}

// TestUpdateUser 测试更新用户
func (suite *UserServiceTestSuite) TestUpdateUser() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		userID      uint
		input       *UpdateUserRequest
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:   "成功更新用户",
			userID: 1,
			input: &UpdateUserRequest{
				Email:  "newemail@example.com",
				Role:   "admin",
				Status: "active",
			},
			setupMocks: func() {
				existingUser := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "old@example.com",
					Role:     "user",
					Status:   "active",
				}
				suite.userRepo.On("GetByID", ctx, uint(1)).Return(existingUser, nil)
				suite.userRepo.On("GetByEmail", ctx, "newemail@example.com").Return((*model.User)(nil), gorm.ErrRecordNotFound)
				suite.userRepo.On("Update", ctx, mock.AnythingOfType("*model.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "用户不存在",
			userID: 999,
			input: &UpdateUserRequest{
				Email: "test@example.com",
			},
			setupMocks: func() {
				suite.userRepo.On("GetByID", ctx, uint(999)).Return((*model.User)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "用户不存在",
		},
		{
			name:   "邮箱已被其他用户使用",
			userID: 1,
			input: &UpdateUserRequest{
				Email: "existing@example.com",
			},
			setupMocks: func() {
				currentUser := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "old@example.com",
				}
				otherUser := &model.User{
					ID:       2,
					Username: "otheruser",
					Email:    "existing@example.com",
				}
				suite.userRepo.On("GetByID", ctx, uint(1)).Return(currentUser, nil)
				suite.userRepo.On("GetByEmail", ctx, "existing@example.com").Return(otherUser, nil)
			},
			expectError: true,
			errorMsg:    "邮箱已被其他用户使用",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userRepo.ExpectedCalls = nil
			suite.userRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			user, err := suite.userService.UpdateUser(ctx, tc.userID, tc.input)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				assert.Nil(suite.T(), user)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), user)
				assert.Equal(suite.T(), tc.userID, user.ID)
				if tc.input.Email != "" {
					assert.Equal(suite.T(), tc.input.Email, user.Email)
				}
				if tc.input.Role != "" {
					assert.Equal(suite.T(), tc.input.Role, user.Role)
				}
				if tc.input.Status != "" {
					assert.Equal(suite.T(), tc.input.Status, user.Status)
				}
			}
		})
	}
}

// TestDeleteUser 测试删除用户
func (suite *UserServiceTestSuite) TestDeleteUser() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		userID      uint
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:   "成功删除用户",
			userID: 1,
			setupMocks: func() {
				user := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Status:   "active",
				}
				suite.userRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
				suite.userRepo.On("Delete", ctx, uint(1)).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "用户不存在",
			userID: 999,
			setupMocks: func() {
				suite.userRepo.On("GetByID", ctx, uint(999)).Return((*model.User)(nil), gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "用户不存在",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userRepo.ExpectedCalls = nil
			suite.userRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			err := suite.userService.DeleteUser(ctx, tc.userID)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestListUsers 测试用户列表
func (suite *UserServiceTestSuite) TestListUsers() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		offset      int
		limit       int
		setupMocks  func()
		expectLen   int
		expectTotal int64
	}{
		{
			name:   "成功获取用户列表",
			offset: 0,
			limit:  10,
			setupMocks: func() {
				users := []*model.User{
					{ID: 1, Username: "user1", Email: "user1@example.com"},
					{ID: 2, Username: "user2", Email: "user2@example.com"},
					{ID: 3, Username: "user3", Email: "user3@example.com"},
				}
				suite.userRepo.On("List", ctx, 0, 10).Return(users, int64(3), nil)
			},
			expectLen:   3,
			expectTotal: 3,
		},
		{
			name:   "空用户列表",
			offset: 0,
			limit:  10,
			setupMocks: func() {
				users := []*model.User{}
				suite.userRepo.On("List", ctx, 0, 10).Return(users, int64(0), nil)
			},
			expectLen:   0,
			expectTotal: 0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 重置模拟对象
			suite.userRepo.ExpectedCalls = nil
			suite.userRepo.Calls = nil

			// 设置模拟
			tc.setupMocks()

			// 执行测试
			users, total, err := suite.userService.ListUsers(ctx, tc.offset, tc.limit)

			assert.NoError(suite.T(), err)
			assert.Len(suite.T(), users, tc.expectLen)
			assert.Equal(suite.T(), tc.expectTotal, total)
		})
	}
}

// TestValidatePassword 测试密码验证
func (suite *UserServiceTestSuite) TestValidatePassword() {
	testCases := []struct {
		name     string
		password string
		valid    bool
	}{
		{"有效密码", "password123", true},
		{"密码太短", "123", false},
		{"密码为空", "", false},
		{"长密码有效", "this-is-a-very-long-password-that-should-be-valid", true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.userService.validatePassword(tc.password)
			if tc.valid {
				assert.NoError(suite.T(), err)
			} else {
				assert.Error(suite.T(), err)
			}
		})
	}
}

// TestValidateEmail 测试邮箱验证
func (suite *UserServiceTestSuite) TestValidateEmail() {
	testCases := []struct {
		name  string
		email string
		valid bool
	}{
		{"有效邮箱", "test@example.com", true},
		{"无效邮箱格式1", "invalid-email", false},
		{"无效邮箱格式2", "@example.com", false},
		{"无效邮箱格式3", "test@", false},
		{"邮箱为空", "", false},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.userService.validateEmail(tc.email)
			if tc.valid {
				assert.NoError(suite.T(), err)
			} else {
				assert.Error(suite.T(), err)
			}
		})
	}
}

// 运行测试套件
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// 基准测试
func BenchmarkUserService_CreateUser(b *testing.B) {
	// 设置基准测试环境
	userRepo := new(MockUserRepository)
	userService := &UserService{userRepo: userRepo}

	ctx := context.Background()
	req := &CreateUserRequest{
		Username: "benchuser",
		Email:    "bench@example.com",
		Password: "password123",
		Role:     "user",
	}

	// 设置模拟
	userRepo.On("GetByUsername", ctx, "benchuser").Return((*model.User)(nil), gorm.ErrRecordNotFound)
	userRepo.On("GetByEmail", ctx, "bench@example.com").Return((*model.User)(nil), gorm.ErrRecordNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userService.CreateUser(ctx, req)
	}
}
