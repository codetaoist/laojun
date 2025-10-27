package config

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// DefaultConfigValidator 默认配置验证器
type DefaultConfigValidator struct {
	rules  map[string][]*ValidationRule
	mu     sync.RWMutex
	logger *logrus.Logger
}

// NewDefaultConfigValidator 创建默认配置验证器
func NewDefaultConfigValidator() *DefaultConfigValidator {
	return &DefaultConfigValidator{
		rules:  make(map[string][]*ValidationRule),
		logger: logrus.New(),
	}
}

// AddRule 添加验证规则
func (v *DefaultConfigValidator) AddRule(ctx context.Context, key string, rule *ValidationRule) error {
	if key == "" {
		return ErrInvalidConfigKey
	}
	
	if rule == nil {
		return fmt.Errorf("validation rule cannot be nil")
	}
	
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.rules[key] == nil {
		v.rules[key] = make([]*ValidationRule, 0)
	}
	
	v.rules[key] = append(v.rules[key], rule)
	
	v.logger.WithFields(logrus.Fields{
		"key":  key,
		"type": rule.Type,
		"name": rule.Name,
	}).Debug("Validation rule added")
	
	return nil
}

// RemoveRule 移除验证规则
func (v *DefaultConfigValidator) RemoveRule(ctx context.Context, key, ruleName string) error {
	if key == "" {
		return ErrInvalidConfigKey
	}
	
	if ruleName == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	
	v.mu.Lock()
	defer v.mu.Unlock()
	
	rules, exists := v.rules[key]
	if !exists {
		return fmt.Errorf("no rules found for key: %s", key)
	}
	
	for i, rule := range rules {
		if rule.Name == ruleName {
			v.rules[key] = append(rules[:i], rules[i+1:]...)
			
			// 如果没有规则了，删除键
			if len(v.rules[key]) == 0 {
				delete(v.rules, key)
			}
			
			v.logger.WithFields(logrus.Fields{
				"key":  key,
				"name": ruleName,
			}).Debug("Validation rule removed")
			
			return nil
		}
	}
	
	return fmt.Errorf("rule not found: %s", ruleName)
}

// Validate 验证配置值
func (v *DefaultConfigValidator) Validate(ctx context.Context, key, value string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	rules, exists := v.rules[key]
	if !exists {
		// 没有规则，验证通过
		return nil
	}
	
	for _, rule := range rules {
		if err := v.validateRule(key, value, rule); err != nil {
			v.logger.WithFields(logrus.Fields{
				"key":   key,
				"value": value,
				"rule":  rule.Name,
				"error": err,
			}).Warn("Config validation failed")
			
			return fmt.Errorf("validation failed for rule '%s': %w", rule.Name, err)
		}
	}
	
	v.logger.WithFields(logrus.Fields{
		"key":   key,
		"value": value,
		"rules": len(rules),
	}).Debug("Config validation passed")
	
	return nil
}

// ValidateMultiple 批量验证配置
func (v *DefaultConfigValidator) ValidateMultiple(ctx context.Context, configs map[string]string) error {
	for key, value := range configs {
		if err := v.Validate(ctx, key, value); err != nil {
			return fmt.Errorf("validation failed for key '%s': %w", key, err)
		}
	}
	
	return nil
}

// GetRules 获取指定键的验证规则
func (v *DefaultConfigValidator) GetRules(ctx context.Context, key string) ([]*ValidationRule, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	rules, exists := v.rules[key]
	if !exists {
		return []*ValidationRule{}, nil
	}
	
	// 返回副本
	result := make([]*ValidationRule, len(rules))
	copy(result, rules)
	
	return result, nil
}

// ListKeys 列出所有有验证规则的键
func (v *DefaultConfigValidator) ListKeys(ctx context.Context) ([]string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	keys := make([]string, 0, len(v.rules))
	for key := range v.rules {
		keys = append(keys, key)
	}
	
	return keys, nil
}

// validateRule 验证单个规则
func (v *DefaultConfigValidator) validateRule(key, value string, rule *ValidationRule) error {
	switch rule.Type {
	case ValidationTypeRequired:
		return v.validateRequired(value, rule)
	case ValidationTypeRegex:
		return v.validateRegex(value, rule)
	case ValidationTypeRange:
		return v.validateRange(value, rule)
	case ValidationTypeEnum:
		return v.validateEnum(value, rule)
	case ValidationTypeLength:
		return v.validateLength(value, rule)
	case ValidationTypeCustom:
		return v.validateCustom(key, value, rule)
	default:
		return fmt.Errorf("unsupported validation type: %s", rule.Type)
	}
}

// validateRequired 验证必填
func (v *DefaultConfigValidator) validateRequired(value string, rule *ValidationRule) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

// validateRegex 验证正则表达式
func (v *DefaultConfigValidator) validateRegex(value string, rule *ValidationRule) error {
	pattern, ok := rule.Parameters["pattern"]
	if !ok {
		return fmt.Errorf("regex pattern not specified")
	}
	
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	if !regex.MatchString(value) {
		return fmt.Errorf("value does not match pattern: %s", pattern)
	}
	
	return nil
}

// validateRange 验证数值范围
func (v *DefaultConfigValidator) validateRange(value string, rule *ValidationRule) error {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("value is not a valid number: %w", err)
	}
	
	if minStr, ok := rule.Parameters["min"]; ok {
		min, err := strconv.ParseFloat(minStr, 64)
		if err != nil {
			return fmt.Errorf("invalid min value: %w", err)
		}
		if num < min {
			return fmt.Errorf("value %g is less than minimum %g", num, min)
		}
	}
	
	if maxStr, ok := rule.Parameters["max"]; ok {
		max, err := strconv.ParseFloat(maxStr, 64)
		if err != nil {
			return fmt.Errorf("invalid max value: %w", err)
		}
		if num > max {
			return fmt.Errorf("value %g is greater than maximum %g", num, max)
		}
	}
	
	return nil
}

// validateEnum 验证枚举值
func (v *DefaultConfigValidator) validateEnum(value string, rule *ValidationRule) error {
	valuesStr, ok := rule.Parameters["values"]
	if !ok {
		return fmt.Errorf("enum values not specified")
	}
	
	values := strings.Split(valuesStr, ",")
	for _, v := range values {
		if strings.TrimSpace(v) == value {
			return nil
		}
	}
	
	return fmt.Errorf("value '%s' is not in allowed values: %s", value, valuesStr)
}

// validateLength 验证字符串长度
func (v *DefaultConfigValidator) validateLength(value string, rule *ValidationRule) error {
	length := len(value)
	
	if minStr, ok := rule.Parameters["min"]; ok {
		min, err := strconv.Atoi(minStr)
		if err != nil {
			return fmt.Errorf("invalid min length: %w", err)
		}
		if length < min {
			return fmt.Errorf("value length %d is less than minimum %d", length, min)
		}
	}
	
	if maxStr, ok := rule.Parameters["max"]; ok {
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return fmt.Errorf("invalid max length: %w", err)
		}
		if length > max {
			return fmt.Errorf("value length %d is greater than maximum %d", length, max)
		}
	}
	
	return nil
}

// validateCustom 自定义验证
func (v *DefaultConfigValidator) validateCustom(key, value string, rule *ValidationRule) error {
	// 这里可以实现自定义验证逻辑
	// 例如，调用外部验证服务或执行复杂的业务逻辑验证
	
	// 示例：检查是否为有效的URL
	if rule.Name == "url" {
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return fmt.Errorf("value must be a valid URL")
		}
	}
	
	// 示例：检查是否为有效的邮箱
	if rule.Name == "email" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			return fmt.Errorf("value must be a valid email address")
		}
	}
	
	// 示例：检查是否为有效的IP地址
	if rule.Name == "ip" {
		parts := strings.Split(value, ".")
		if len(parts) != 4 {
			return fmt.Errorf("value must be a valid IP address")
		}
		for _, part := range parts {
			num, err := strconv.Atoi(part)
			if err != nil || num < 0 || num > 255 {
				return fmt.Errorf("value must be a valid IP address")
			}
		}
	}
	
	return nil
}

// CompositeConfigValidator 组合配置验证器
type CompositeConfigValidator struct {
	validators []ConfigValidator
	mu         sync.RWMutex
	logger     *logrus.Logger
}

// NewCompositeConfigValidator 创建组合配置验证器
func NewCompositeConfigValidator(validators ...ConfigValidator) *CompositeConfigValidator {
	return &CompositeConfigValidator{
		validators: validators,
		logger:     logrus.New(),
	}
}

// AddValidator 添加验证器
func (v *CompositeConfigValidator) AddValidator(validator ConfigValidator) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.validators = append(v.validators, validator)
}

// RemoveValidator 移除验证器
func (v *CompositeConfigValidator) RemoveValidator(index int) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if index < 0 || index >= len(v.validators) {
		return fmt.Errorf("invalid validator index: %d", index)
	}
	
	v.validators = append(v.validators[:index], v.validators[index+1:]...)
	return nil
}

// AddRule 添加验证规则
func (v *CompositeConfigValidator) AddRule(ctx context.Context, key string, rule *ValidationRule) error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	// 添加到所有验证器
	for _, validator := range v.validators {
		if err := validator.AddRule(ctx, key, rule); err != nil {
			return err
		}
	}
	
	return nil
}

// RemoveRule 移除验证规则
func (v *CompositeConfigValidator) RemoveRule(ctx context.Context, key, ruleName string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	// 从所有验证器中移除
	for _, validator := range v.validators {
		if err := validator.RemoveRule(ctx, key, ruleName); err != nil {
			// 忽略不存在的错误
			if !strings.Contains(err.Error(), "not found") {
				return err
			}
		}
	}
	
	return nil
}

// Validate 验证配置值
func (v *CompositeConfigValidator) Validate(ctx context.Context, key, value string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	// 所有验证器都必须通过
	for i, validator := range v.validators {
		if err := validator.Validate(ctx, key, value); err != nil {
			return fmt.Errorf("validator %d failed: %w", i, err)
		}
	}
	
	return nil
}

// ValidateMultiple 批量验证配置
func (v *CompositeConfigValidator) ValidateMultiple(ctx context.Context, configs map[string]string) error {
	for key, value := range configs {
		if err := v.Validate(ctx, key, value); err != nil {
			return fmt.Errorf("validation failed for key '%s': %w", key, err)
		}
	}
	
	return nil
}

// GetRules 获取指定键的验证规则
func (v *CompositeConfigValidator) GetRules(ctx context.Context, key string) ([]*ValidationRule, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	var allRules []*ValidationRule
	
	// 收集所有验证器的规则
	for _, validator := range v.validators {
		rules, err := validator.GetRules(ctx, key)
		if err != nil {
			return nil, err
		}
		allRules = append(allRules, rules...)
	}
	
	return allRules, nil
}

// ListKeys 列出所有有验证规则的键
func (v *CompositeConfigValidator) ListKeys(ctx context.Context) ([]string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	keySet := make(map[string]bool)
	
	// 收集所有验证器的键
	for _, validator := range v.validators {
		keys, err := validator.ListKeys(ctx)
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			keySet[key] = true
		}
	}
	
	// 转换为切片
	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	
	return keys, nil
}