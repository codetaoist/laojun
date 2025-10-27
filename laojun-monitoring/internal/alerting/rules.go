package alerting

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RuleEngine 规则引擎
type RuleEngine struct {
	mu      sync.RWMutex
	rules   map[string]*AlertRule
	querier Querier
	logger  *zap.Logger
	
	// 规则评估统计
	evaluations map[string]*RuleEvaluationStats
}

// RuleGroupManager 规则组管理器
type RuleGroupManager struct {
	mu     sync.RWMutex
	groups map[string]*RuleGroup
	engine *RuleEngine
	logger *zap.Logger
}

// RuleGroup 规则组
type RuleGroup struct {
	Name     string       `json:"name"`
	Interval time.Duration `json:"interval"`
	Rules    []*AlertRule `json:"rules"`
}

// Sample 样本数据
type Sample struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Series 时间序列
type Series struct {
	Labels  map[string]string `json:"labels"`
	Samples []Sample         `json:"samples"`
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(logger *zap.Logger, querier Querier) *RuleEngine {
	return &RuleEngine{
		rules:       make(map[string]*AlertRule),
		querier:     querier,
		logger:      logger,
		evaluations: make(map[string]*RuleEvaluationStats),
	}
}

// AddRule 添加规则
func (re *RuleEngine) AddRule(rule *AlertRule) error {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	// 验证规则
	if err := re.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}
	
	re.rules[rule.Name] = rule
	re.evaluations[rule.Name] = &RuleEvaluationStats{}
	
	re.logger.Info("Added alert rule", 
		zap.String("name", rule.Name),
		zap.String("expr", rule.Expr))
	
	return nil
}

// RemoveRule 移除规则
func (re *RuleEngine) RemoveRule(name string) {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	delete(re.rules, name)
	delete(re.evaluations, name)
	
	re.logger.Info("Removed alert rule", zap.String("name", name))
}

// GetRule 获取规则
func (re *RuleEngine) GetRule(name string) (*AlertRule, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	rule, exists := re.rules[name]
	return rule, exists
}

// ListRules 列出所有规则
func (re *RuleEngine) ListRules() []*AlertRule {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	rules := make([]*AlertRule, 0, len(re.rules))
	for _, rule := range re.rules {
		rules = append(rules, rule)
	}
	
	return rules
}

// EvaluateRule 评估单个规则
func (re *RuleEngine) EvaluateRule(ctx context.Context, ruleName string, timestamp time.Time) ([]*Alert, error) {
	re.mu.RLock()
	rule, exists := re.rules[ruleName]
	re.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleName)
	}
	
	start := time.Now()
	alerts, err := re.evaluateRuleInternal(ctx, rule, timestamp)
	duration := time.Since(start)
	
	// 更新统计信息
	re.mu.Lock()
	stats := re.evaluations[ruleName]
	stats.TotalEvaluations++
	stats.LastEvaluation = timestamp
	stats.EvaluationTime = duration
	
	if err != nil {
		stats.Errors++
		stats.LastError = err.Error()
		stats.LastErrorTime = timestamp
	}
	re.mu.Unlock()
	
	return alerts, err
}

// EvaluateAllRules 评估所有规则
func (re *RuleEngine) EvaluateAllRules(ctx context.Context, timestamp time.Time) ([]*Alert, error) {
	re.mu.RLock()
	rules := make([]*AlertRule, 0, len(re.rules))
	for _, rule := range re.rules {
		rules = append(rules, rule)
	}
	re.mu.RUnlock()
	
	var allAlerts []*Alert
	var lastErr error
	
	for _, rule := range rules {
		alerts, err := re.EvaluateRule(ctx, rule.Name, timestamp)
		if err != nil {
			re.logger.Error("Failed to evaluate rule", 
				zap.String("rule", rule.Name), 
				zap.Error(err))
			lastErr = err
			continue
		}
		
		allAlerts = append(allAlerts, alerts...)
	}
	
	return allAlerts, lastErr
}

// GetRuleStats 获取规则统计信息
func (re *RuleEngine) GetRuleStats(ruleName string) (*RuleEvaluationStats, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	stats, exists := re.evaluations[ruleName]
	return stats, exists
}

// GetAllRuleStats 获取所有规则统计信息
func (re *RuleEngine) GetAllRuleStats() map[string]*RuleEvaluationStats {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	stats := make(map[string]*RuleEvaluationStats)
	for name, stat := range re.evaluations {
		stats[name] = stat
	}
	
	return stats
}

// validateRule 验证规则
func (re *RuleEngine) validateRule(rule *AlertRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	
	if rule.Expr == "" {
		return fmt.Errorf("rule expression cannot be empty")
	}
	
	if rule.For < 0 {
		return fmt.Errorf("rule for duration cannot be negative")
	}
	
	// 验证表达式语法
	if err := re.validateExpression(rule.Expr); err != nil {
		return fmt.Errorf("invalid expression: %w", err)
	}
	
	return nil
}

// validateExpression 验证表达式语法
func (re *RuleEngine) validateExpression(expr string) error {
	// 简单的表达式验证
	// 支持的格式: metric_name{label="value"} operator threshold
	
	// 检查是否包含基本组件
	if !strings.Contains(expr, ">") && !strings.Contains(expr, "<") && 
	   !strings.Contains(expr, ">=") && !strings.Contains(expr, "<=") && 
	   !strings.Contains(expr, "==") && !strings.Contains(expr, "!=") {
		return fmt.Errorf("expression must contain a comparison operator")
	}
	
	return nil
}

// evaluateRuleInternal 内部规则评估逻辑
func (re *RuleEngine) evaluateRuleInternal(ctx context.Context, rule *AlertRule, timestamp time.Time) ([]*Alert, error) {
	// 解析表达式
	query, operator, threshold, err := re.parseExpression(rule.Expr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}
	
	// 执行查询
	result, err := re.querier.Query(ctx, query, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	
	// 评估结果
	alerts := make([]*Alert, 0)
	
	switch result.Type {
	case "vector":
		if series, ok := result.Result.([]Series); ok {
			for _, s := range series {
				if len(s.Samples) == 0 {
					continue
				}
				
				// 取最新样本
				sample := s.Samples[len(s.Samples)-1]
				
				// 评估条件
				if re.evaluateCondition(sample.Value, operator, threshold) {
					alert := &Alert{
				ID:          re.generateAlertID(rule.Name, s.Labels),
				RuleID:      rule.Name,
				RuleName:    rule.Name,
				Name:        rule.Name,
				State:       AlertStateFiring,
				Status:      AlertStateFiring,
				Severity:    rule.Severity,
				Message:     fmt.Sprintf("Alert %s is firing", rule.Name),
				Value:       sample.Value,
				Labels:      s.Labels,
				Annotations: rule.Annotations,
				ActiveAt:    timestamp,
				StartsAt:    timestamp,
				UpdatedAt:   timestamp,
			}
					
					alerts = append(alerts, alert)
				}
			}
		}
	case "scalar":
		if value, ok := result.Result.(float64); ok {
			if re.evaluateCondition(value, operator, threshold) {
				alert := &Alert{
					ID:          re.generateAlertID(rule.Name, nil),
					RuleID:      rule.Name,
					RuleName:    rule.Name,
					Name:        rule.Name,
					State:       AlertStateFiring,
					Status:      AlertStateFiring,
					Severity:    rule.Severity,
					Message:     fmt.Sprintf("Alert %s is firing", rule.Name),
					Value:       value,
					Labels:      make(map[string]string),
					Annotations: rule.Annotations,
					ActiveAt:    timestamp,
					StartsAt:    timestamp,
					UpdatedAt:   timestamp,
				}
				
				alerts = append(alerts, alert)
			}
		}
	}
	
	return alerts, nil
}

// parseExpression 解析表达式
func (re *RuleEngine) parseExpression(expr string) (query, operator string, threshold float64, err error) {
	// 支持的操作符
	operators := []string{">=", "<=", "==", "!=", ">", "<"}
	
	for _, op := range operators {
		if strings.Contains(expr, op) {
			parts := strings.Split(expr, op)
			if len(parts) != 2 {
				continue
			}
			
			query = strings.TrimSpace(parts[0])
			operator = op
			
			thresholdStr := strings.TrimSpace(parts[1])
			threshold, err = strconv.ParseFloat(thresholdStr, 64)
			if err != nil {
				return "", "", 0, fmt.Errorf("invalid threshold value: %s", thresholdStr)
			}
			
			return query, operator, threshold, nil
		}
	}
	
	return "", "", 0, fmt.Errorf("no valid operator found in expression")
}

// evaluateCondition 评估条件
func (re *RuleEngine) evaluateCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

// generateAlertID 生成告警ID
func (re *RuleEngine) generateAlertID(ruleName string, labels map[string]string) string {
	if len(labels) == 0 {
		return ruleName
	}
	
	var labelPairs []string
	for k, v := range labels {
		labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", k, v))
	}
	
	return fmt.Sprintf("%s{%s}", ruleName, strings.Join(labelPairs, ","))
}

// RuleGroup 规则组
// 移除重复的RuleGroup和RuleGroupManager定义，使用上面的定义

// NewRuleGroupManager 创建规则组管理器
func NewRuleGroupManager(logger *zap.Logger, engine *RuleEngine) *RuleGroupManager {
	return &RuleGroupManager{
		groups: make(map[string]*RuleGroup),
		engine: engine,
		logger: logger,
	}
}

// AddGroup 添加规则组
func (rgm *RuleGroupManager) AddGroup(group *RuleGroup) error {
	rgm.mu.Lock()
	defer rgm.mu.Unlock()
	
	// 验证规则组
	if group.Name == "" {
		return fmt.Errorf("rule group name cannot be empty")
	}
	
	if group.Interval <= 0 {
		group.Interval = 30 * time.Second // 默认间隔
	}
	
	// 添加组中的所有规则到引擎
	for _, rule := range group.Rules {
		if err := rgm.engine.AddRule(rule); err != nil {
			return fmt.Errorf("failed to add rule %s: %w", rule.Name, err)
		}
	}
	
	rgm.groups[group.Name] = group
	
	rgm.logger.Info("Added rule group", 
		zap.String("name", group.Name),
		zap.Int("rules", len(group.Rules)),
		zap.Duration("interval", group.Interval))
	
	return nil
}

// RemoveGroup 移除规则组
func (rgm *RuleGroupManager) RemoveGroup(name string) {
	rgm.mu.Lock()
	defer rgm.mu.Unlock()
	
	group, exists := rgm.groups[name]
	if !exists {
		return
	}
	
	// 从引擎中移除所有规则
	for _, rule := range group.Rules {
		rgm.engine.RemoveRule(rule.Name)
	}
	
	delete(rgm.groups, name)
	
	rgm.logger.Info("Removed rule group", zap.String("name", name))
}

// GetGroup 获取规则组
func (rgm *RuleGroupManager) GetGroup(name string) (*RuleGroup, bool) {
	rgm.mu.RLock()
	defer rgm.mu.RUnlock()
	
	group, exists := rgm.groups[name]
	return group, exists
}

// ListGroups 列出所有规则组
func (rgm *RuleGroupManager) ListGroups() []*RuleGroup {
	rgm.mu.RLock()
	defer rgm.mu.RUnlock()
	
	groups := make([]*RuleGroup, 0, len(rgm.groups))
	for _, group := range rgm.groups {
		groups = append(groups, group)
	}
	
	return groups
}

// EvaluateGroup 评估规则组
func (rgm *RuleGroupManager) EvaluateGroup(ctx context.Context, groupName string, timestamp time.Time) ([]*Alert, error) {
	rgm.mu.RLock()
	group, exists := rgm.groups[groupName]
	rgm.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("rule group not found: %s", groupName)
	}
	
	var allAlerts []*Alert
	var lastErr error
	
	for _, rule := range group.Rules {
		alerts, err := rgm.engine.EvaluateRule(ctx, rule.Name, timestamp)
		if err != nil {
			rgm.logger.Error("Failed to evaluate rule in group", 
				zap.String("group", groupName),
				zap.String("rule", rule.Name), 
				zap.Error(err))
			lastErr = err
			continue
		}
		
		allAlerts = append(allAlerts, alerts...)
	}
	
	return allAlerts, lastErr
}

// TemplateEngine 模板引擎
type TemplateEngine struct {
	templates map[string]string
	mu        sync.RWMutex
}

// NewTemplateEngine 创建模板引擎
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]string),
	}
}

// AddTemplate 添加模板
func (te *TemplateEngine) AddTemplate(name, template string) {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	te.templates[name] = template
}

// RenderTemplate 渲染模板
func (te *TemplateEngine) RenderTemplate(name string, data map[string]interface{}) (string, error) {
	te.mu.RLock()
	template, exists := te.templates[name]
	te.mu.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("template not found: %s", name)
	}
	
	// 简单的模板替换
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	
	return result, nil
}

// ExpressionValidator 表达式验证器
type ExpressionValidator struct {
	patterns map[string]*regexp.Regexp
}

// NewExpressionValidator 创建表达式验证器
func NewExpressionValidator() *ExpressionValidator {
	patterns := map[string]*regexp.Regexp{
		"metric_name": regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`),
		"label_name":  regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`),
		"label_value": regexp.MustCompile(`^.*$`), // 允许任何值
	}
	
	return &ExpressionValidator{
		patterns: patterns,
	}
}

// ValidateMetricName 验证指标名称
func (ev *ExpressionValidator) ValidateMetricName(name string) bool {
	return ev.patterns["metric_name"].MatchString(name)
}

// ValidateLabelName 验证标签名称
func (ev *ExpressionValidator) ValidateLabelName(name string) bool {
	return ev.patterns["label_name"].MatchString(name)
}

// ValidateExpression 验证完整表达式
func (ev *ExpressionValidator) ValidateExpression(expr string) error {
	if expr == "" {
		return fmt.Errorf("expression cannot be empty")
	}
	
	// 基本语法检查
	if !strings.Contains(expr, ">") && !strings.Contains(expr, "<") && 
	   !strings.Contains(expr, ">=") && !strings.Contains(expr, "<=") && 
	   !strings.Contains(expr, "==") && !strings.Contains(expr, "!=") {
		return fmt.Errorf("expression must contain a comparison operator")
	}
	
	return nil
}