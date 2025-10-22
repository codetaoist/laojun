package utils

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) (bool, []string) {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "Password must be at least 8 characters long")
	}

	if len(password) > 128 {
		errors = append(errors, "Password must be no more than 128 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "Password must contain at least one digit")
	}
	if !hasSpecial {
		errors = append(errors, "Password must contain at least one special character")
	}

	return len(errors) == 0, errors
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) (bool, []string) {
	var errors []string

	if len(username) < 3 {
		errors = append(errors, "Username must be at least 3 characters long")
	}

	if len(username) > 50 {
		errors = append(errors, "Username must be no more than 50 characters long")
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		errors = append(errors, "Username can only contain letters, numbers, underscores, and hyphens")
	}

	return len(errors) == 0, errors
}

// ValidateUUID 验证UUID格式
func ValidateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// SanitizeString 清理字符串（移除前后空格，防止XSS攻击）
func SanitizeString(input string) string {
	// 移除前后空格
	input = strings.TrimSpace(input)

	// 基本的HTML标签清理（简单实现）
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	input = strings.ReplaceAll(input, "&", "&amp;")

	return input
}

// ValidateRequired 验证必填字段
func ValidateRequired(value string, fieldName string) (bool, string) {
	if strings.TrimSpace(value) == "" {
		return false, fieldName + " is required"
	}
	return true, ""
}

// ValidateStringLength 验证字符串长度
func ValidateStringLength(value string, fieldName string, min, max int) (bool, string) {
	length := len(strings.TrimSpace(value))
	if length < min {
		return false, fieldName + " must be at least " + string(rune(min)) + " characters long"
	}
	if length > max {
		return false, fieldName + " must be no more than " + string(rune(max)) + " characters long"
	}
	return true, ""
}

// ValidatePositiveNumber 验证正数
func ValidatePositiveNumber(value float64, fieldName string) (bool, string) {
	if value <= 0 {
		return false, fieldName + " must be a positive number"
	}
	return true, ""
}

// ValidateRating 验证评分范围
func ValidateRating(rating float64) (bool, string) {
	if rating < 1 || rating > 5 {
		return false, "Rating must be between 1 and 5"
	}
	return true, ""
}
