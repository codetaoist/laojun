package sdk

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors 验证错误集合
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}

// HasErrors 检查是否有错误
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// AddError 添加错误
func (e *ValidationErrors) AddError(field, message string, value interface{}) {
	*e = append(*e, &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// Validator 验证器接口
type Validator interface {
	Validate(value interface{}) ValidationErrors
}

// StringValidator 字符串验证器
type StringValidator struct {
	MinLength int
	MaxLength int
	Pattern   *regexp.Regexp
	Required  bool
}

// Validate 验证字符串
func (v *StringValidator) Validate(value interface{}) ValidationErrors {
	var errors ValidationErrors
	
	str, ok := value.(string)
	if !ok {
		errors.AddError("", "value must be a string", value)
		return errors
	}
	
	if v.Required && str == "" {
		errors.AddError("", "field is required", value)
		return errors
	}
	
	if str == "" && !v.Required {
		return errors
	}
	
	if v.MinLength > 0 && len(str) < v.MinLength {
		errors.AddError("", fmt.Sprintf("minimum length is %d", v.MinLength), value)
	}
	
	if v.MaxLength > 0 && len(str) > v.MaxLength {
		errors.AddError("", fmt.Sprintf("maximum length is %d", v.MaxLength), value)
	}
	
	if v.Pattern != nil && !v.Pattern.MatchString(str) {
		errors.AddError("", "does not match required pattern", value)
	}
	
	return errors
}

// NumberValidator 数字验证器
type NumberValidator struct {
	Min      *float64
	Max      *float64
	Required bool
}

// Validate 验证数字
func (v *NumberValidator) Validate(value interface{}) ValidationErrors {
	var errors ValidationErrors
	
	if value == nil {
		if v.Required {
			errors.AddError("", "field is required", value)
		}
		return errors
	}
	
	var num float64
	var ok bool
	
	switch val := value.(type) {
	case int:
		num = float64(val)
		ok = true
	case int32:
		num = float64(val)
		ok = true
	case int64:
		num = float64(val)
		ok = true
	case float32:
		num = float64(val)
		ok = true
	case float64:
		num = val
		ok = true
	case string:
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			num = parsed
			ok = true
		}
	}
	
	if !ok {
		errors.AddError("", "value must be a number", value)
		return errors
	}
	
	if v.Min != nil && num < *v.Min {
		errors.AddError("", fmt.Sprintf("minimum value is %f", *v.Min), value)
	}
	
	if v.Max != nil && num > *v.Max {
		errors.AddError("", fmt.Sprintf("maximum value is %f", *v.Max), value)
	}
	
	return errors
}

// ArrayValidator 数组验证器
type ArrayValidator struct {
	MinItems    int
	MaxItems    int
	ItemValidator Validator
	Required    bool
}

// Validate 验证数组
func (v *ArrayValidator) Validate(value interface{}) ValidationErrors {
	var errors ValidationErrors
	
	if value == nil {
		if v.Required {
			errors.AddError("", "field is required", value)
		}
		return errors
	}
	
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		errors.AddError("", "value must be an array", value)
		return errors
	}
	
	length := rv.Len()
	
	if v.MinItems > 0 && length < v.MinItems {
		errors.AddError("", fmt.Sprintf("minimum items is %d", v.MinItems), value)
	}
	
	if v.MaxItems > 0 && length > v.MaxItems {
		errors.AddError("", fmt.Sprintf("maximum items is %d", v.MaxItems), value)
	}
	
	// 验证每个元素
	if v.ItemValidator != nil {
		for i := 0; i < length; i++ {
			item := rv.Index(i).Interface()
			itemErrors := v.ItemValidator.Validate(item)
			for _, err := range itemErrors {
				err.Field = fmt.Sprintf("[%d].%s", i, err.Field)
				errors = append(errors, err)
			}
		}
	}
	
	return errors
}

// ConfigValidator 配置验证器
type ConfigValidator struct {
	Schema *ConfigSchema
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator(schema *ConfigSchema) *ConfigValidator {
	return &ConfigValidator{Schema: schema}
}

// Validate 验证配置
func (v *ConfigValidator) Validate(config map[string]interface{}) ValidationErrors {
	var errors ValidationErrors
	
	if v.Schema == nil {
		return errors
	}
	
	// 验证必需字段
	for name, prop := range v.Schema.Properties {
		value, exists := config[name]
		
		if prop.Required && !exists {
			errors.AddError(name, "field is required", nil)
			continue
		}
		
		if !exists {
			continue
		}
		
		// 验证字段
		fieldErrors := v.validateProperty(name, value, prop)
		errors = append(errors, fieldErrors...)
	}
	
	// 检查未知字段
	for name := range config {
		if _, exists := v.Schema.Properties[name]; !exists {
			errors.AddError(name, "unknown field", config[name])
		}
	}
	
	return errors
}

// validateProperty 验证属性
func (v *ConfigValidator) validateProperty(name string, value interface{}, prop *ConfigProperty) ValidationErrors {
	var errors ValidationErrors
	
	// 类型验证
	if !v.isValidType(value, prop.Type) {
		errors.AddError(name, fmt.Sprintf("expected type %s", prop.Type), value)
		return errors
	}
	
	// 根据类型进行具体验证
	switch prop.Type {
	case "string":
		validator := &StringValidator{
			Required: prop.Required,
		}
		if prop.MinLength != nil {
			validator.MinLength = *prop.MinLength
		}
		if prop.MaxLength != nil {
			validator.MaxLength = *prop.MaxLength
		}
		if prop.Pattern != "" {
			if pattern, err := regexp.Compile(prop.Pattern); err == nil {
				validator.Pattern = pattern
			}
		}
		fieldErrors := validator.Validate(value)
		for _, err := range fieldErrors {
			err.Field = name
			errors = append(errors, err)
		}
		
	case "number", "integer":
		validator := &NumberValidator{
			Required: prop.Required,
		}
		if prop.Minimum != nil {
			validator.Min = prop.Minimum
		}
		if prop.Maximum != nil {
			validator.Max = prop.Maximum
		}
		fieldErrors := validator.Validate(value)
		for _, err := range fieldErrors {
			err.Field = name
			errors = append(errors, err)
		}
		
	case "array":
		validator := &ArrayValidator{
			Required: prop.Required,
		}
		if prop.MinItems != nil {
			validator.MinItems = *prop.MinItems
		}
		if prop.MaxItems != nil {
			validator.MaxItems = *prop.MaxItems
		}
		fieldErrors := validator.Validate(value)
		for _, err := range fieldErrors {
			if err.Field == "" {
				err.Field = name
			} else {
				err.Field = name + "." + err.Field
			}
			errors = append(errors, err)
		}
	}
	
	// 枚举值验证
	if len(prop.Enum) > 0 {
		valid := false
		for _, enumValue := range prop.Enum {
			if reflect.DeepEqual(value, enumValue) {
				valid = true
				break
			}
		}
		if !valid {
			errors.AddError(name, "value not in allowed enum values", value)
		}
	}
	
	return errors
}

// isValidType 检查类型是否有效
func (v *ConfigValidator) isValidType(value interface{}, expectedType string) bool {
	if value == nil {
		return true // null值在其他地方处理
	}
	
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			return true
		}
		return false
	case "integer":
		switch value.(type) {
		case int, int32, int64:
			return true
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
	case "object":
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Map
	default:
		return true
	}
}

// FileUtils 文件工具
type FileUtils struct{}

// NewFileUtils 创建文件工具
func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

// Exists 检查文件是否存在
func (f *FileUtils) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir 检查是否为目录
func (f *FileUtils) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile 检查是否为文件
func (f *FileUtils) IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ReadJSON 读取JSON文件
func (f *FileUtils) ReadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", path, err)
	}
	
	return nil
}

// WriteJSON 写入JSON文件
func (f *FileUtils) WriteJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	
	return nil
}

// CopyFile 复制文件
func (f *FileUtils) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()
	
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destFile.Close()
	
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	
	return nil
}

// GetFileHash 获取文件哈希值
func (f *FileUtils) GetFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()
	
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// StringUtils 字符串工具
type StringUtils struct{}

// NewStringUtils 创建字符串工具
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// GenerateID 生成随机ID
func (s *StringUtils) GenerateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// Slugify 生成URL友好的字符串
func (s *StringUtils) Slugify(text string) string {
	// 转换为小写
	text = strings.ToLower(text)
	
	// 替换空格和特殊字符为连字符
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	text = reg.ReplaceAllString(text, "-")
	
	// 移除首尾的连字符
	text = strings.Trim(text, "-")
	
	return text
}

// TruncateString 截断字符串
func (s *StringUtils) TruncateString(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	
	if maxLength <= 3 {
		return text[:maxLength]
	}
	
	return text[:maxLength-3] + "..."
}

// SanitizeString 清理字符串
func (s *StringUtils) SanitizeString(text string) string {
	// 移除控制字符
	reg := regexp.MustCompile(`[\x00-\x1f\x7f]`)
	text = reg.ReplaceAllString(text, "")
	
	// 移除多余的空白字符
	text = strings.TrimSpace(text)
	reg = regexp.MustCompile(`\s+`)
	text = reg.ReplaceAllString(text, " ")
	
	return text
}

// TimeUtils 时间工具
type TimeUtils struct{}

// NewTimeUtils 创建时间工具
func NewTimeUtils() *TimeUtils {
	return &TimeUtils{}
}

// FormatDuration 格式化持续时间
func (t *TimeUtils) FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		return fmt.Sprintf("%.1fd", d.Hours()/24)
	}
}

// ParseDuration 解析持续时间字符串
func (t *TimeUtils) ParseDuration(s string) (time.Duration, error) {
	// 支持更多格式
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	
	// 尝试解析数字+单位格式
	reg := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([a-zA-Z]+)$`)
	matches := reg.FindStringSubmatch(s)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}
	
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", matches[1])
	}
	
	unit := strings.ToLower(matches[2])
	switch unit {
	case "s", "sec", "second", "seconds":
		return time.Duration(value * float64(time.Second)), nil
	case "m", "min", "minute", "minutes":
		return time.Duration(value * float64(time.Minute)), nil
	case "h", "hour", "hours":
		return time.Duration(value * float64(time.Hour)), nil
	case "d", "day", "days":
		return time.Duration(value * float64(24*time.Hour)), nil
	default:
		return 0, fmt.Errorf("unknown duration unit: %s", unit)
	}
}

// IsExpired 检查时间是否已过期
func (t *TimeUtils) IsExpired(timestamp time.Time, ttl time.Duration) bool {
	return time.Since(timestamp) > ttl
}

// LoggerUtils 日志工具
type LoggerUtils struct{}

// NewLoggerUtils 创建日志工具
func NewLoggerUtils() *LoggerUtils {
	return &LoggerUtils{}
}

// CreateLogger 创建日志记录器
func (l *LoggerUtils) CreateLogger(pluginID string, level logrus.Level) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(level)
	
	// 设置格式化器
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	
	// 添加插件ID字段
	logger = logger.WithField("plugin_id", pluginID).Logger
	
	return logger
}

// CreateFileLogger 创建文件日志记录器
func (l *LoggerUtils) CreateFileLogger(pluginID, logFile string, level logrus.Level) (*logrus.Logger, error) {
	logger := l.CreateLogger(pluginID, level)
	
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	logger.SetOutput(file)
	
	return logger, nil
}

// HTTPUtils HTTP工具
type HTTPUtils struct{}

// NewHTTPUtils 创建HTTP工具
func NewHTTPUtils() *HTTPUtils {
	return &HTTPUtils{}
}

// ParseContentType 解析Content-Type
func (h *HTTPUtils) ParseContentType(contentType string) (mediaType string, params map[string]string) {
	parts := strings.Split(contentType, ";")
	mediaType = strings.TrimSpace(parts[0])
	
	params = make(map[string]string)
	for i := 1; i < len(parts); i++ {
		param := strings.TrimSpace(parts[i])
		if idx := strings.Index(param, "="); idx > 0 {
			key := strings.TrimSpace(param[:idx])
			value := strings.TrimSpace(param[idx+1:])
			// 移除引号
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			params[key] = value
		}
	}
	
	return mediaType, params
}

// BuildContentType 构建Content-Type
func (h *HTTPUtils) BuildContentType(mediaType string, params map[string]string) string {
	if len(params) == 0 {
		return mediaType
	}
	
	var parts []string
	parts = append(parts, mediaType)
	
	for key, value := range params {
		if strings.Contains(value, " ") || strings.Contains(value, ";") {
			value = fmt.Sprintf(`"%s"`, value)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	
	return strings.Join(parts, "; ")
}

// GetClientIP 获取客户端IP
func (h *HTTPUtils) GetClientIP(headers map[string]string) string {
	// 检查常见的代理头
	proxyHeaders := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP",
	}
	
	for _, header := range proxyHeaders {
		if ip := headers[header]; ip != "" {
			// X-Forwarded-For可能包含多个IP，取第一个
			if idx := strings.Index(ip, ","); idx > 0 {
				ip = strings.TrimSpace(ip[:idx])
			}
			return ip
		}
	}
	
	// 如果没有代理头，返回空字符串
	return ""
}

// Utils 工具集合
type Utils struct {
	File   *FileUtils
	String *StringUtils
	Time   *TimeUtils
	Logger *LoggerUtils
	HTTP   *HTTPUtils
}

// NewUtils 创建工具集合
func NewUtils() *Utils {
	return &Utils{
		File:   NewFileUtils(),
		String: NewStringUtils(),
		Time:   NewTimeUtils(),
		Logger: NewLoggerUtils(),
		HTTP:   NewHTTPUtils(),
	}
}