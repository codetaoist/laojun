package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	mathrand "math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// StringUtils 字符串工具
type StringUtils struct{}

// IsEmpty 检查字符串是否为空
func (s StringUtils) IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// IsNotEmpty 检查字符串是否不为空
func (s StringUtils) IsNotEmpty(str string) bool {
	return !s.IsEmpty(str)
}

// Truncate 截断字符串
func (s StringUtils) Truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// ToCamelCase 转换为驼峰命名
func (s StringUtils) ToCamelCase(str string) string {
	words := strings.Fields(strings.ReplaceAll(str, "_", " "))
	if len(words) == 0 {
		return ""
	}
	
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += strings.Title(strings.ToLower(words[i]))
	}
	return result
}

// ToSnakeCase 转换为蛇形命名
func (s StringUtils) ToSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// Contains 检查字符串是否包含子字符串（忽略大小写）
func (s StringUtils) Contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// Reverse 反转字符串
func (s StringUtils) Reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// RandomString 生成随机字符串
func (s StringUtils) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[mathrand.Int63()%int64(len(charset))]
	}
	return string(b)
}

// SliceUtils 切片工具
type SliceUtils struct{}

// Contains 检查切片是否包含元素
func (s SliceUtils) Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ContainsInt 检查整数切片是否包含元素
func (s SliceUtils) ContainsInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Unique 去重字符串切片
func (s SliceUtils) Unique(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// UniqueInt 去重整数切片
func (s SliceUtils) UniqueInt(slice []int) []int {
	keys := make(map[int]bool)
	var result []int
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// Reverse 反转切片
func (s SliceUtils) Reverse(slice []string) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// ReverseInt 反转整数切片
func (s SliceUtils) ReverseInt(slice []int) []int {
	result := make([]int, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// MapUtils 映射工具
type MapUtils struct{}

// Keys 获取map的所有键
func (m MapUtils) Keys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// Values 获取map的所有值
func (m MapUtils) Values(data map[string]interface{}) []interface{} {
	values := make([]interface{}, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}

// Merge 合并两个map
func (m MapUtils) Merge(map1, map2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for k, v := range map1 {
		result[k] = v
	}
	
	for k, v := range map2 {
		result[k] = v
	}
	
	return result
}

// TimeUtils 时间工具
type TimeUtils struct{}

// FormatTime 格式化时间
func (t TimeUtils) FormatTime(timeVal time.Time, format string) string {
	switch format {
	case "date":
		return timeVal.Format("2006-01-02")
	case "datetime":
		return timeVal.Format("2006-01-02 15:04:05")
	case "time":
		return timeVal.Format("15:04:05")
	case "iso":
		return timeVal.Format(time.RFC3339)
	default:
		return timeVal.Format(format)
	}
}

// ParseTime 解析时间字符串
func (t TimeUtils) ParseTime(timeStr, format string) (time.Time, error) {
	switch format {
	case "date":
		return time.Parse("2006-01-02", timeStr)
	case "datetime":
		return time.Parse("2006-01-02 15:04:05", timeStr)
	case "time":
		return time.Parse("15:04:05", timeStr)
	case "iso":
		return time.Parse(time.RFC3339, timeStr)
	default:
		return time.Parse(format, timeStr)
	}
}

// IsToday 检查是否是今天
func (t TimeUtils) IsToday(timeVal time.Time) bool {
	now := time.Now()
	return timeVal.Year() == now.Year() && timeVal.YearDay() == now.YearDay()
}

// DaysBetween 计算两个日期之间的天数
func (t TimeUtils) DaysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}

// StartOfDay 获取一天的开始时间
func (t TimeUtils) StartOfDay(timeVal time.Time) time.Time {
	return time.Date(timeVal.Year(), timeVal.Month(), timeVal.Day(), 0, 0, 0, 0, timeVal.Location())
}

// EndOfDay 获取一天的结束时间
func (t TimeUtils) EndOfDay(timeVal time.Time) time.Time {
	return time.Date(timeVal.Year(), timeVal.Month(), timeVal.Day(), 23, 59, 59, 999999999, timeVal.Location())
}

// NumberUtils 数字工具
type NumberUtils struct{}

// Round 四舍五入到指定小数位
func (n NumberUtils) Round(num float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(num*ratio) / ratio
}

// Max 返回两个数的最大值
func (n NumberUtils) Max(a, b float64) float64 {
	return math.Max(a, b)
}

// Min 返回两个数的最小值
func (n NumberUtils) Min(a, b float64) float64 {
	return math.Min(a, b)
}

// Abs 返回绝对值
func (n NumberUtils) Abs(num float64) float64 {
	return math.Abs(num)
}

// IsEven 检查是否为偶数
func (n NumberUtils) IsEven(num int) bool {
	return num%2 == 0
}

// IsOdd 检查是否为奇数
func (n NumberUtils) IsOdd(num int) bool {
	return num%2 != 0
}

// ConvertUtils 转换工具
type ConvertUtils struct{}

// ToString 转换为字符串
func (c ConvertUtils) ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToInt 转换为整数
func (c ConvertUtils) ToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// ToFloat64 转换为浮点数
func (c ConvertUtils) ToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// ToBool 转换为布尔值
func (c ConvertUtils) ToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// JSONUtils JSON工具
type JSONUtils struct{}

// ToJSON 转换为JSON字符串
func (j JSONUtils) ToJSON(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON 从JSON字符串解析
func (j JSONUtils) FromJSON(jsonStr string, target interface{}) error {
	return json.Unmarshal([]byte(jsonStr), target)
}

// PrettyJSON 格式化JSON
func (j JSONUtils) PrettyJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ValidateUtils 验证工具
type ValidateUtils struct{}

// IsValidEmail 验证邮箱格式
func (v ValidateUtils) IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhone 验证手机号格式（中国）
func (v ValidateUtils) IsValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// IsValidURL 验证URL格式
func (v ValidateUtils) IsValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// IsValidUUID 验证UUID格式
func (v ValidateUtils) IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// CryptoUtils 加密工具
type CryptoUtils struct{}

// GenerateUUID 生成UUID
func (c CryptoUtils) GenerateUUID() string {
	return uuid.New().String()
}

// GenerateRandomHex 生成随机十六进制字符串
func (c CryptoUtils) GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ReflectUtils 反射工具
type ReflectUtils struct{}

// GetStructFields 获取结构体字段名
func (r ReflectUtils) GetStructFields(obj interface{}) []string {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return nil
	}
	
	typ := val.Type()
	fields := make([]string, val.NumField())
	
	for i := 0; i < val.NumField(); i++ {
		fields[i] = typ.Field(i).Name
	}
	
	return fields
}

// IsZeroValue 检查是否为零值
func (r ReflectUtils) IsZeroValue(value interface{}) bool {
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}

// 全局工具实例
var (
	String   = StringUtils{}
	Slice    = SliceUtils{}
	Map      = MapUtils{}
	Time     = TimeUtils{}
	Number   = NumberUtils{}
	Convert  = ConvertUtils{}
	JSON     = JSONUtils{}
	Validate = ValidateUtils{}
	Crypto   = CryptoUtils{}
	Reflect  = ReflectUtils{}
)

// Pagination 分页工具
type Pagination struct {
	Page     int `json:"page" form:"page" binding:"min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100"`
}

// GetOffset 获取偏移量
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *Pagination) GetLimit() int {
	return p.PageSize
}

// NewPagination 创建分页对象
func NewPagination(page, pageSize int) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// CalculatePagination 计算分页信息
func CalculatePagination(total int64, page, pageSize int) map[string]interface{} {
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	
	return map[string]interface{}{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"has_next":    page < totalPages,
		"has_prev":    page > 1,
	}
}