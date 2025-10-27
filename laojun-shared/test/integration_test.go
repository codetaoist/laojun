package test

import (
	"context"
	"testing"
	"time"

	"github.com/codetaoist/laojun-shared/cache"
	"github.com/codetaoist/laojun-shared/crypto"
	"github.com/codetaoist/laojun-shared/health"
	"github.com/codetaoist/laojun-shared/logger"
	"github.com/codetaoist/laojun-shared/utils"
	"github.com/codetaoist/laojun-shared/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheManager(t *testing.T) {
	// 测试内存缓存
	config := &cache.CacheConfig{
		Type:              cache.CacheTypeMemory,
		DefaultExpiration: time.Minute,
		MemoryCleanup:     time.Minute * 5,
	}

	manager, err := cache.NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// 测试基本操作
	err = manager.Set(context.Background(), "test_key", "test_value", time.Minute)
	assert.NoError(t, err)

	value, err := manager.Get(context.Background(), "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)

	exists, err := manager.Exists(context.Background(), "test_key")
	assert.NoError(t, err)
	assert.True(t, exists > 0)

	// 测试JSON操作
	testData := map[string]interface{}{
		"name": "test",
		"age":  25,
	}
	err = manager.SetJSON(context.Background(), "json_key", testData, time.Minute)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = manager.GetJSON(context.Background(), "json_key", &result)
	assert.NoError(t, err)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(25), result["age"])
}

func TestValidator(t *testing.T) {
	v := validator.New()

	// 测试结构体验证
	type TestStruct struct {
		Email    string `json:"email" validate:"required,email"`
		Phone    string `json:"phone" validate:"required,phone"`
		Password string `json:"password" validate:"required,password"`
		Username string `json:"username" validate:"required,username"`
	}

	// 有效数据
	validData := TestStruct{
		Email:    "test@example.com",
		Phone:    "13800138000",
		Password: "Password123!",
		Username: "testuser",
	}

	err := v.Validate(validData)
	assert.NoError(t, err)

	// 无效数据
	invalidData := TestStruct{
		Email:    "invalid-email",
		Phone:    "123",
		Password: "weak",
		Username: "a",
	}

	err = v.Validate(invalidData)
	assert.Error(t, err)

	errors := v.ValidateStruct(invalidData)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "email")
	assert.Contains(t, errors, "phone")
	assert.Contains(t, errors, "password")
	assert.Contains(t, errors, "username")
}

func TestCrypto(t *testing.T) {
	helper := crypto.NewEncryptionHelper("test-secret-key")

	// 测试加密解密
	plaintext := "Hello, World!"
	encrypted, err := helper.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := helper.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// 测试密码哈希
	password := "mypassword"
	hashedPassword, err := helper.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, password, hashedPassword)

	// 验证密码
	isValid := helper.CheckPassword(password, hashedPassword)
	assert.True(t, isValid)

	// 错误密码
	isValid = helper.CheckPassword("wrongpassword", hashedPassword)
	assert.False(t, isValid)

	// 测试SHA256哈希
	hash := crypto.SHA256Hash("test")
	assert.NotEmpty(t, hash)
	assert.Equal(t, 64, len(hash)) // SHA256 hex string length

	// 测试随机密钥生成
	key, err := crypto.GenerateRandomKey(32)
	assert.NoError(t, err)
	assert.Equal(t, 64, len(key)) // hex string is double the byte length
}

func TestHealthChecker(t *testing.T) {
	h := health.New(health.Config{
		Timeout: 5 * time.Second,
	})

	// 添加自定义检查器
	customChecker := health.NewCustomChecker("test-checker", func(ctx context.Context) (health.Status, string, error) {
		return health.StatusHealthy, "Test checker is healthy", nil
	})
	h.AddChecker(customChecker)

	// 执行健康检查
	report := h.Check(context.Background())
	assert.Equal(t, health.StatusHealthy, report.Status)
	assert.Contains(t, report.Checks, "test-checker")
	assert.Equal(t, health.StatusHealthy, report.Checks["test-checker"].Status)
}

func TestUtils(t *testing.T) {
	// 测试字符串工具
	assert.True(t, utils.String.IsEmpty(""))
	assert.False(t, utils.String.IsEmpty("test"))
	assert.Equal(t, "test_string", utils.String.ToSnakeCase("TestString"))
	assert.Equal(t, "testString", utils.String.ToCamelCase("test_string"))
	assert.True(t, utils.String.Contains("hello world", "world"))

	// 测试切片工具
	slice := []string{"a", "b", "c"}
	assert.True(t, utils.Slice.Contains(slice, "b"))
	assert.False(t, utils.Slice.Contains(slice, "d"))

	unique := utils.Slice.Unique([]string{"a", "b", "a", "c", "b"})
	assert.Equal(t, []string{"a", "b", "c"}, unique)

	// 测试时间工具
	now := time.Now()
	formatted := utils.Time.FormatTime(now, "2006-01-02")
	assert.Contains(t, formatted, "-")

	isToday := utils.Time.IsToday(now)
	assert.True(t, isToday)

	// 测试数字工具
	assert.Equal(t, 5.0, utils.Number.Max(3, 5))
	assert.Equal(t, 3.0, utils.Number.Min(3, 5))
	assert.Equal(t, 5.0, utils.Number.Abs(-5))
	assert.True(t, utils.Number.IsEven(4))
	assert.False(t, utils.Number.IsEven(5))

	// 测试转换工具
	assert.Equal(t, "123", utils.Convert.ToString(123))
	intVal, err := utils.Convert.ToInt("123")
	assert.NoError(t, err)
	assert.Equal(t, 123, intVal)
	floatVal, err := utils.Convert.ToFloat64("123.45")
	assert.NoError(t, err)
	assert.Equal(t, 123.45, floatVal)
	boolVal, err := utils.Convert.ToBool("true")
	assert.NoError(t, err)
	assert.True(t, boolVal)

	// 测试JSON工具
	data := map[string]interface{}{"name": "test", "age": 25}
	jsonStr, err := utils.JSON.ToJSON(data)
	assert.NoError(t, err)
	assert.Contains(t, jsonStr, "test")

	var result map[string]interface{}
	err = utils.JSON.FromJSON(jsonStr, &result)
	assert.NoError(t, err)
	assert.Equal(t, "test", result["name"])

	// 测试验证工具
	assert.True(t, utils.Validate.IsValidEmail("test@example.com"))
	assert.False(t, utils.Validate.IsValidEmail("invalid-email"))
	assert.True(t, utils.Validate.IsValidURL("https://example.com"))
	assert.False(t, utils.Validate.IsValidURL("invalid-url"))

	// 测试加密工具
	uuid := utils.Crypto.GenerateUUID()
	assert.NotEmpty(t, uuid)
	assert.Equal(t, 36, len(uuid)) // UUID length with hyphens

	hex, err := utils.Crypto.GenerateRandomHex(16)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(hex)) // 16 hex characters

	// 测试分页工具
	pagination := utils.NewPagination(1, 10)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 10, pagination.PageSize)
	assert.Equal(t, 0, pagination.GetOffset())
	assert.Equal(t, 10, pagination.GetLimit())
}

func TestLogger(t *testing.T) {
	config := logger.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	log := logger.New(config)
	assert.NotNil(t, log)

	// 测试日志记录（这里只是确保不会panic）
	log.Info("Test info message")
	log.Error("Test error message")
	log.Debug("Test debug message") // 由于level是info，这个不会输出
}