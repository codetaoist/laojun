package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationConfig 验证配置
type ValidationConfig struct {
	SkipOnError bool // 遇到错误时是否跳过后续验证
}

// CustomValidator 自定义验证器
type CustomValidator struct {
	validator *validator.Validate
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationResponse 验证响应
type ValidationResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// NewCustomValidator 创建自定义验证器
func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// 注册自定义验证规则
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("slug", validateSlug)
	v.RegisterValidation("version", validateVersion)
	v.RegisterValidation("url_or_empty", validateURLOrEmpty)
	v.RegisterValidation("json_string", validateJSONString)
	v.RegisterValidation("role", validateRole)
	v.RegisterValidation("status", validateStatus)
	v.RegisterValidation("config_type", validateConfigType)
	v.RegisterValidation("plugin_category", validatePluginCategory)

	return &CustomValidator{validator: v}
}

// ValidateJSON JSON 验证中间件
func ValidateJSON(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data interface{}

		// 创建模型的新实例
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		data = reflect.New(modelType).Interface()

		// 绑定 JSON 数据
		if err := c.ShouldBindJSON(data); err != nil {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Error:   "invalid_json",
				Message: "Invalid JSON format",
				Errors: []ValidationError{{
					Field:   "json",
					Tag:     "json",
					Value:   "",
					Message: err.Error(),
				}},
			})
			c.Abort()
			return
		}

		// 验证数据
		validator := NewCustomValidator()
		if err := validator.Validate(data); err != nil {
			errors := formatValidationErrors(err)
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Error:   "validation_failed",
				Message: "Validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		// 将验证后的数据存储到上下文
		c.Set("validated_data", data)
		c.Next()
	}
}

// ValidateQuery 查询参数验证中间件
func ValidateQuery(rules map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var errors []ValidationError

		for param, rule := range rules {
			value := c.Query(param)

			if err := validateValue(param, value, rule); err != nil {
				errors = append(errors, *err)
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Error:   "validation_failed",
				Message: "Query parameter validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateParams 路径参数验证中间件
func ValidateParams(rules map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var errors []ValidationError

		for param, rule := range rules {
			value := c.Param(param)

			if err := validateValue(param, value, rule); err != nil {
				errors = append(errors, *err)
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Error:   "validation_failed",
				Message: "Path parameter validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateHeaders 请求头验证中间件
func ValidateHeaders(rules map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var errors []ValidationError

		for header, rule := range rules {
			value := c.GetHeader(header)

			if err := validateValue(header, value, rule); err != nil {
				errors = append(errors, *err)
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Error:   "validation_failed",
				Message: "Header validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Validate 验证数据
func (cv *CustomValidator) Validate(data interface{}) error {
	return cv.validator.Struct(data)
}

// validateValue 验证单个值
func validateValue(field, value, rule string) *ValidationError {
	rules := strings.Split(rule, "|")

	for _, r := range rules {
		parts := strings.Split(r, ":")
		tag := parts[0]
		var param string
		if len(parts) > 1 {
			param = parts[1]
		}

		if err := validateSingleRule(field, value, tag, param); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleRule 验证单个规则
func validateSingleRule(field, value, tag, param string) *ValidationError {
	switch tag {
	case "required":
		if value == "" {
			return &ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: fmt.Sprintf("%s is required", field),
			}
		}
	case "min":
		minLen, _ := strconv.Atoi(param)
		if len(value) < minLen {
			return &ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: fmt.Sprintf("%s must be at least %d characters", field, minLen),
			}
		}
	case "max":
		maxLen, _ := strconv.Atoi(param)
		if len(value) > maxLen {
			return &ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: fmt.Sprintf("%s must be at most %d characters", field, maxLen),
			}
		}
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if value != "" && !emailRegex.MatchString(value) {
			return &ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: fmt.Sprintf("%s must be a valid email address", field),
			}
		}
	case "uuid":
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		if value != "" && !uuidRegex.MatchString(value) {
			return &ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: fmt.Sprintf("%s must be a valid UUID", field),
			}
		}
	case "numeric":
		if value != "" {
			if _, err := strconv.Atoi(value); err != nil {
				return &ValidationError{
					Field:   field,
					Tag:     tag,
					Value:   value,
					Message: fmt.Sprintf("%s must be numeric", field),
				}
			}
		}
	case "alpha":
		alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
		if value != "" && !alphaRegex.MatchString(value) {
			return &ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: fmt.Sprintf("%s must contain only letters", field),
			}
		}
	case "alphanum":
		alphanumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
		if value != "" && !alphanumRegex.MatchString(value) {
			return &ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: fmt.Sprintf("%s must contain only letters and numbers", field),
			}
		}
	}

	return nil
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Tag:     e.Tag(),
				Value:   fmt.Sprintf("%v", e.Value()),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

// getErrorMessage 获取错误消息
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param())
	case "username":
		return fmt.Sprintf("%s must be a valid username (3-30 characters, letters, numbers, underscore, hyphen)", e.Field())
	case "password":
		return fmt.Sprintf("%s must be at least 8 characters with uppercase, lowercase, number and special character", e.Field())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", e.Field())
	case "slug":
		return fmt.Sprintf("%s must be a valid slug (lowercase letters, numbers, hyphen)", e.Field())
	case "version":
		return fmt.Sprintf("%s must be a valid semantic version", e.Field())
	case "url_or_empty":
		return fmt.Sprintf("%s must be a valid URL or empty", e.Field())
	case "json_string":
		return fmt.Sprintf("%s must be a valid JSON string", e.Field())
	case "role":
		return fmt.Sprintf("%s must be a valid role (admin, moderator, developer, user)", e.Field())
	case "status":
		return fmt.Sprintf("%s must be a valid status", e.Field())
	case "config_type":
		return fmt.Sprintf("%s must be a valid config type (string, number, boolean, json)", e.Field())
	case "plugin_category":
		return fmt.Sprintf("%s must be a valid plugin category", e.Field())
	default:
		return fmt.Sprintf("%s is invalid", e.Field())
	}
}

// 自定义验证函数

// validateUsername 验证用户名
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return usernameRegex.MatchString(username)
}

// validatePassword 验证密码
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	// 至少包含一个大写字母、一个小写字母、一个数字和一个特殊字符
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validatePhone 验证手机号
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	phoneRegex := regexp.MustCompile(`^(\+86)?1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// validateSlug 验证 slug
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	return slugRegex.MatchString(slug)
}

// validateVersion 验证版本号
func validateVersion(fl validator.FieldLevel) bool {
	version := fl.Field().String()
	versionRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(-[a-zA-Z0-9-]+)?(\+[a-zA-Z0-9-]+)?$`)
	return versionRegex.MatchString(version)
}

// validateURLOrEmpty 验证 URL 或空字符串
func validateURLOrEmpty(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true
	}

	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// validateJSONString 验证 JSON 字符串
func validateJSONString(fl validator.FieldLevel) bool {
	jsonStr := fl.Field().String()
	if jsonStr == "" {
		return true
	}

	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// validateRole 验证角色
func validateRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := []string{"admin", "moderator", "developer", "user"}

	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}

	return false
}

// validateStatus 验证状态
func validateStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"active", "inactive", "pending", "suspended", "deleted"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}

	return false
}

// validateConfigType 验证配置类型
func validateConfigType(fl validator.FieldLevel) bool {
	configType := fl.Field().String()
	validTypes := []string{"string", "number", "boolean", "json"}

	for _, validType := range validTypes {
		if configType == validType {
			return true
		}
	}

	return false
}

// validatePluginCategory 验证插件类别
func validatePluginCategory(fl validator.FieldLevel) bool {
	category := fl.Field().String()
	validCategories := []string{
		"authentication", "database", "api", "cache", "logging",
		"backup", "security", "notification", "workflow", "validation",
		"file", "analytics", "development", "scheduling", "monitoring",
		"configuration", "load-balancing", "messaging", "documentation", "performance",
	}

	for _, validCategory := range validCategories {
		if category == validCategory {
			return true
		}
	}

	return false
}
