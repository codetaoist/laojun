package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator interface defines validation methods
type Validator interface {
	Validate(data interface{}) error
	ValidateStruct(data interface{}) map[string]string
	RegisterValidation(tag string, fn validator.Func) error
	RegisterTranslation(tag, message string) error
}

// CustomValidator implements the Validator interface
type CustomValidator struct {
	validator    *validator.Validate
	translations map[string]string
}

// New creates a new validator instance
func New() *CustomValidator {
	v := validator.New()
	cv := &CustomValidator{
		validator:    v,
		translations: make(map[string]string),
	}

	// Register custom validations
	cv.registerCustomValidations()
	cv.registerDefaultTranslations()

	return cv
}

// Validate validates a struct and returns the first error
func (cv *CustomValidator) Validate(data interface{}) error {
	return cv.validator.Struct(data)
}

// ValidateStruct validates a struct and returns all validation errors as a map
func (cv *CustomValidator) ValidateStruct(data interface{}) map[string]string {
	errors := make(map[string]string)
	
	err := cv.validator.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := cv.getFieldName(err)
			message := cv.getErrorMessage(err)
			errors[field] = message
		}
	}
	
	return errors
}

// RegisterValidation registers a custom validation function
func (cv *CustomValidator) RegisterValidation(tag string, fn validator.Func) error {
	return cv.validator.RegisterValidation(tag, fn)
}

// RegisterTranslation registers a custom error message translation
func (cv *CustomValidator) RegisterTranslation(tag, message string) error {
	cv.translations[tag] = message
	return nil
}

// registerCustomValidations registers all custom validation rules
func (cv *CustomValidator) registerCustomValidations() {
	// Phone number validation
	cv.validator.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
		return matched
	})

	// Password strength validation
	cv.validator.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}
		
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber := regexp.MustCompile(`\d`).MatchString(password)
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
		
		return hasUpper && hasLower && hasNumber && hasSpecial
	})

	// Username validation
	cv.validator.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) < 3 || len(username) > 20 {
			return false
		}
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
		return matched
	})

	// Chinese name validation
	cv.validator.RegisterValidation("chinese_name", func(fl validator.FieldLevel) bool {
		name := fl.Field().String()
		matched, _ := regexp.MatchString(`^[\u4e00-\u9fa5]{2,10}$`, name)
		return matched
	})

	// ID card validation
	cv.validator.RegisterValidation("id_card", func(fl validator.FieldLevel) bool {
		idCard := fl.Field().String()
		matched, _ := regexp.MatchString(`^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`, idCard)
		return matched
	})

	// HTTP URL validation
	cv.validator.RegisterValidation("http_url", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, url)
		return matched
	})

	// IP address validation
	cv.validator.RegisterValidation("ip_addr", func(fl validator.FieldLevel) bool {
		ip := fl.Field().String()
		matched, _ := regexp.MatchString(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`, ip)
		return matched
	})

	// Port validation
	cv.validator.RegisterValidation("port", func(fl validator.FieldLevel) bool {
		port := fl.Field().Int()
		return port >= 1 && port <= 65535
	})
}

// registerDefaultTranslations registers default error message translations
func (cv *CustomValidator) registerDefaultTranslations() {
	cv.translations["required"] = "{field} is required"
	cv.translations["email"] = "{field} must be a valid email address"
	cv.translations["min"] = "{field} must be at least {param} characters"
	cv.translations["max"] = "{field} must be at most {param} characters"
	cv.translations["len"] = "{field} must be exactly {param} characters"
	cv.translations["gte"] = "{field} must be greater than or equal to {param}"
	cv.translations["lte"] = "{field} must be less than or equal to {param}"
	cv.translations["gt"] = "{field} must be greater than {param}"
	cv.translations["lt"] = "{field} must be less than {param}"
	cv.translations["oneof"] = "{field} must be one of: {param}"
	cv.translations["uuid"] = "{field} must be a valid UUID format"
	cv.translations["phone"] = "{field} must be a valid phone number"
	cv.translations["password"] = "{field} must contain uppercase, lowercase, numbers and special characters, at least 8 characters"
	cv.translations["username"] = "{field} can only contain letters, numbers and underscores, 3-20 characters"
	cv.translations["chinese_name"] = "{field} must be a valid Chinese name"
	cv.translations["id_card"] = "{field} must be a valid ID card number"
	cv.translations["http_url"] = "{field} must be a valid HTTP/HTTPS URL"
	cv.translations["ip_addr"] = "{field} must be a valid IP address"
	cv.translations["port"] = "{field} must be a valid port number (1-65535)"
}

// ValidateVar validates a single variable
func (cv *CustomValidator) ValidateVar(field interface{}, tag string) error {
	return cv.validator.Var(field, tag)
}

// ValidateEmail validates email format
func (cv *CustomValidator) ValidateEmail(email string) bool {
	return cv.validator.Var(email, "required,email") == nil
}

// ValidatePhone validates phone number format
func (cv *CustomValidator) ValidatePhone(phone string) bool {
	return cv.validator.Var(phone, "required,phone") == nil
}

// ValidatePassword validates password strength
func (cv *CustomValidator) ValidatePassword(password string) bool {
	return cv.validator.Var(password, "required,password") == nil
}

// ValidateUsername validates username format
func (cv *CustomValidator) ValidateUsername(username string) bool {
	return cv.validator.Var(username, "required,username") == nil
}

// Global validator instance
var DefaultValidator = New()

// getFieldName extracts the field name from validation error
func (cv *CustomValidator) getFieldName(err validator.FieldError) string {
	field := err.Field()
	
	// Convert to snake_case for JSON field names
	var result strings.Builder
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	
	return strings.ToLower(result.String())
}

// getErrorMessage generates error message from validation error
func (cv *CustomValidator) getErrorMessage(err validator.FieldError) string {
	tag := err.Tag()
	field := err.Field()
	param := err.Param()
	
	if template, exists := cv.translations[tag]; exists {
		message := strings.ReplaceAll(template, "{field}", field)
		message = strings.ReplaceAll(message, "{param}", param)
		return message
	}
	
	return fmt.Sprintf("%s validation failed for field %s", tag, field)
}