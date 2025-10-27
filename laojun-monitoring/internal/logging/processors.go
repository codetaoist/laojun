package logging

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FilterProcessor 过滤处理器
type FilterProcessor struct {
	name    string
	config  FilterProcessorConfig
	logger  *zap.Logger
	filters []Filter
}

// FilterProcessorConfig 过滤处理器配置
type FilterProcessorConfig struct {
	Rules []FilterRule `mapstructure:"rules"`
}

// FilterRule 过滤规则
type FilterRule struct {
	Field    string      `mapstructure:"field"`
	Operator string      `mapstructure:"operator"` // eq, ne, contains, regex, gt, lt, gte, lte
	Value    interface{} `mapstructure:"value"`
	Action   string      `mapstructure:"action"`   // drop, keep
}

// Filter 过滤器接口
type Filter interface {
	Match(entry *LogEntry) bool
	Action() string
}

// NewFilterProcessor 创建过滤处理器
func NewFilterProcessor(name string, config FilterProcessorConfig, logger *zap.Logger) *FilterProcessor {
	fp := &FilterProcessor{
		name:    name,
		config:  config,
		logger:  logger,
		filters: make([]Filter, 0),
	}
	
	// 创建过滤器
	for _, rule := range config.Rules {
		filter := NewRuleFilter(rule)
		fp.filters = append(fp.filters, filter)
	}
	
	return fp
}

// Name 返回处理器名称
func (fp *FilterProcessor) Name() string {
	return fp.name
}

// Process 处理日志条目
func (fp *FilterProcessor) Process(entry *LogEntry) *LogEntry {
	for _, filter := range fp.filters {
		if filter.Match(entry) {
			switch filter.Action() {
			case "drop":
				return nil // 丢弃日志
			case "keep":
				return entry // 保留日志
			}
		}
	}
	
	return entry
}

// RuleFilter 规则过滤器
type RuleFilter struct {
	rule FilterRule
}

// NewRuleFilter 创建规则过滤器
func NewRuleFilter(rule FilterRule) *RuleFilter {
	return &RuleFilter{rule: rule}
}

// Match 匹配规则
func (rf *RuleFilter) Match(entry *LogEntry) bool {
	value := rf.getFieldValue(entry, rf.rule.Field)
	if value == nil {
		return false
	}
	
	switch rf.rule.Operator {
	case "eq":
		return rf.equals(value, rf.rule.Value)
	case "ne":
		return !rf.equals(value, rf.rule.Value)
	case "contains":
		return rf.contains(value, rf.rule.Value)
	case "regex":
		return rf.regex(value, rf.rule.Value)
	case "gt":
		return rf.greater(value, rf.rule.Value)
	case "lt":
		return rf.less(value, rf.rule.Value)
	case "gte":
		return rf.greaterEqual(value, rf.rule.Value)
	case "lte":
		return rf.lessEqual(value, rf.rule.Value)
	default:
		return false
	}
}

// Action 返回动作
func (rf *RuleFilter) Action() string {
	return rf.rule.Action
}

// getFieldValue 获取字段值
func (rf *RuleFilter) getFieldValue(entry *LogEntry, field string) interface{} {
	switch field {
	case "level":
		return string(entry.Level)
	case "message":
		return entry.Message
	case "source":
		return entry.Source
	case "service":
		return entry.Service
	case "trace_id":
		return entry.TraceID
	case "span_id":
		return entry.SpanID
	default:
		// 检查字段和标签
		if value, exists := entry.Fields[field]; exists {
			return value
		}
		if value, exists := entry.Tags[field]; exists {
			return value
		}
		return nil
	}
}

// equals 相等比较
func (rf *RuleFilter) equals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// contains 包含比较
func (rf *RuleFilter) contains(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Contains(aStr, bStr)
}

// regex 正则表达式匹配
func (rf *RuleFilter) regex(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	
	matched, err := regexp.MatchString(bStr, aStr)
	if err != nil {
		return false
	}
	
	return matched
}

// greater 大于比较
func (rf *RuleFilter) greater(a, b interface{}) bool {
	aNum, aOk := rf.toFloat64(a)
	bNum, bOk := rf.toFloat64(b)
	
	if !aOk || !bOk {
		return false
	}
	
	return aNum > bNum
}

// less 小于比较
func (rf *RuleFilter) less(a, b interface{}) bool {
	aNum, aOk := rf.toFloat64(a)
	bNum, bOk := rf.toFloat64(b)
	
	if !aOk || !bOk {
		return false
	}
	
	return aNum < bNum
}

// greaterEqual 大于等于比较
func (rf *RuleFilter) greaterEqual(a, b interface{}) bool {
	aNum, aOk := rf.toFloat64(a)
	bNum, bOk := rf.toFloat64(b)
	
	if !aOk || !bOk {
		return false
	}
	
	return aNum >= bNum
}

// lessEqual 小于等于比较
func (rf *RuleFilter) lessEqual(a, b interface{}) bool {
	aNum, aOk := rf.toFloat64(a)
	bNum, bOk := rf.toFloat64(b)
	
	if !aOk || !bOk {
		return false
	}
	
	return aNum <= bNum
}

// toFloat64 转换为float64
func (rf *RuleFilter) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// EnrichProcessor 丰富处理器
type EnrichProcessor struct {
	name   string
	config EnrichProcessorConfig
	logger *zap.Logger
}

// EnrichProcessorConfig 丰富处理器配置
type EnrichProcessorConfig struct {
	Fields map[string]string `mapstructure:"fields"` // 静态字段
	Tags   map[string]string `mapstructure:"tags"`   // 静态标签
	Rules  []EnrichRule      `mapstructure:"rules"`  // 动态规则
}

// EnrichRule 丰富规则
type EnrichRule struct {
	Condition EnrichCondition `mapstructure:"condition"`
	Fields    map[string]string `mapstructure:"fields"`
	Tags      map[string]string `mapstructure:"tags"`
}

// EnrichCondition 丰富条件
type EnrichCondition struct {
	Field    string      `mapstructure:"field"`
	Operator string      `mapstructure:"operator"`
	Value    interface{} `mapstructure:"value"`
}

// NewEnrichProcessor 创建丰富处理器
func NewEnrichProcessor(name string, config EnrichProcessorConfig, logger *zap.Logger) *EnrichProcessor {
	return &EnrichProcessor{
		name:   name,
		config: config,
		logger: logger,
	}
}

// Name 返回处理器名称
func (ep *EnrichProcessor) Name() string {
	return ep.name
}

// Process 处理日志条目
func (ep *EnrichProcessor) Process(entry *LogEntry) *LogEntry {
	if entry == nil {
		return nil
	}
	
	// 确保字段和标签映射已初始化
	if entry.Fields == nil {
		entry.Fields = make(map[string]interface{})
	}
	if entry.Tags == nil {
		entry.Tags = make(map[string]string)
	}
	
	// 添加静态字段
	for k, v := range ep.config.Fields {
		entry.Fields[k] = v
	}
	
	// 添加静态标签
	for k, v := range ep.config.Tags {
		entry.Tags[k] = v
	}
	
	// 应用动态规则
	for _, rule := range ep.config.Rules {
		if ep.matchCondition(entry, rule.Condition) {
			// 添加规则字段
			for k, v := range rule.Fields {
				entry.Fields[k] = v
			}
			
			// 添加规则标签
			for k, v := range rule.Tags {
				entry.Tags[k] = v
			}
		}
	}
	
	return entry
}

// matchCondition 匹配条件
func (ep *EnrichProcessor) matchCondition(entry *LogEntry, condition EnrichCondition) bool {
	filter := NewRuleFilter(FilterRule{
		Field:    condition.Field,
		Operator: condition.Operator,
		Value:    condition.Value,
	})
	
	return filter.Match(entry)
}

// ParseProcessor 解析处理器
type ParseProcessor struct {
	name   string
	config ParseProcessorConfig
	logger *zap.Logger
	parser Parser
}

// ParseProcessorConfig 解析处理器配置
type ParseProcessorConfig struct {
	Type    string            `mapstructure:"type"`    // json, regex, grok, csv
	Pattern string            `mapstructure:"pattern"` // 解析模式
	Fields  []string          `mapstructure:"fields"`  // 字段名称
	Options map[string]string `mapstructure:"options"` // 解析选项
}

// Parser 解析器接口
type Parser interface {
	Parse(message string) (map[string]interface{}, error)
}

// NewParseProcessor 创建解析处理器
func NewParseProcessor(name string, config ParseProcessorConfig, logger *zap.Logger) *ParseProcessor {
	pp := &ParseProcessor{
		name:   name,
		config: config,
		logger: logger,
	}
	
	// 创建解析器
	switch config.Type {
	case "json":
		pp.parser = NewJSONParser()
	case "regex":
		pp.parser = NewRegexParser(config.Pattern, config.Fields)
	case "csv":
		pp.parser = NewCSVParser(config.Fields, config.Options)
	default:
		pp.parser = NewJSONParser()
	}
	
	return pp
}

// Name 返回处理器名称
func (pp *ParseProcessor) Name() string {
	return pp.name
}

// Process 处理日志条目
func (pp *ParseProcessor) Process(entry *LogEntry) *LogEntry {
	if entry == nil {
		return nil
	}
	
	// 解析消息
	fields, err := pp.parser.Parse(entry.Message)
	if err != nil {
		pp.logger.Debug("Failed to parse message", 
			zap.String("message", entry.Message), 
			zap.Error(err))
		return entry
	}
	
	// 确保字段映射已初始化
	if entry.Fields == nil {
		entry.Fields = make(map[string]interface{})
	}
	
	// 添加解析的字段
	for k, v := range fields {
		entry.Fields[k] = v
	}
	
	return entry
}

// JSONParser JSON解析器
type JSONParser struct{}

// NewJSONParser 创建JSON解析器
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Parse 解析JSON消息
func (jp *JSONParser) Parse(message string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(message), &result)
	return result, err
}

// RegexParser 正则表达式解析器
type RegexParser struct {
	pattern *regexp.Regexp
	fields  []string
}

// NewRegexParser 创建正则表达式解析器
func NewRegexParser(pattern string, fields []string) *RegexParser {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	
	return &RegexParser{
		pattern: regex,
		fields:  fields,
	}
}

// Parse 解析正则表达式消息
func (rp *RegexParser) Parse(message string) (map[string]interface{}, error) {
	if rp.pattern == nil {
		return nil, fmt.Errorf("invalid regex pattern")
	}
	
	matches := rp.pattern.FindStringSubmatch(message)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found")
	}
	
	result := make(map[string]interface{})
	
	// 跳过第一个匹配（完整匹配）
	for i, match := range matches[1:] {
		if i < len(rp.fields) {
			result[rp.fields[i]] = match
		} else {
			result[fmt.Sprintf("field_%d", i)] = match
		}
	}
	
	return result, nil
}

// CSVParser CSV解析器
type CSVParser struct {
	fields    []string
	separator string
}

// NewCSVParser 创建CSV解析器
func NewCSVParser(fields []string, options map[string]string) *CSVParser {
	separator := ","
	if sep, exists := options["separator"]; exists {
		separator = sep
	}
	
	return &CSVParser{
		fields:    fields,
		separator: separator,
	}
}

// Parse 解析CSV消息
func (cp *CSVParser) Parse(message string) (map[string]interface{}, error) {
	parts := strings.Split(message, cp.separator)
	result := make(map[string]interface{})
	
	for i, part := range parts {
		if i < len(cp.fields) {
			result[cp.fields[i]] = strings.TrimSpace(part)
		} else {
			result[fmt.Sprintf("field_%d", i)] = strings.TrimSpace(part)
		}
	}
	
	return result, nil
}

// TransformProcessor 转换处理器
type TransformProcessor struct {
	name        string
	config      TransformProcessorConfig
	logger      *zap.Logger
	transforms  []Transform
}

// TransformProcessorConfig 转换处理器配置
type TransformProcessorConfig struct {
	Rules []TransformRule `mapstructure:"rules"`
}

// TransformRule 转换规则
type TransformRule struct {
	Field     string `mapstructure:"field"`
	Operation string `mapstructure:"operation"` // lowercase, uppercase, trim, replace, remove
	Options   map[string]string `mapstructure:"options"`
}

// Transform 转换接口
type Transform interface {
	Apply(entry *LogEntry) *LogEntry
}

// NewTransformProcessor 创建转换处理器
func NewTransformProcessor(name string, config TransformProcessorConfig, logger *zap.Logger) *TransformProcessor {
	tp := &TransformProcessor{
		name:       name,
		config:     config,
		logger:     logger,
		transforms: make([]Transform, 0),
	}
	
	// 创建转换器
	for _, rule := range config.Rules {
		transform := NewRuleTransform(rule)
		tp.transforms = append(tp.transforms, transform)
	}
	
	return tp
}

// Name 返回处理器名称
func (tp *TransformProcessor) Name() string {
	return tp.name
}

// Process 处理日志条目
func (tp *TransformProcessor) Process(entry *LogEntry) *LogEntry {
	if entry == nil {
		return nil
	}
	
	result := entry
	
	for _, transform := range tp.transforms {
		result = transform.Apply(result)
		if result == nil {
			break
		}
	}
	
	return result
}

// RuleTransform 规则转换器
type RuleTransform struct {
	rule TransformRule
}

// NewRuleTransform 创建规则转换器
func NewRuleTransform(rule TransformRule) *RuleTransform {
	return &RuleTransform{rule: rule}
}

// Apply 应用转换
func (rt *RuleTransform) Apply(entry *LogEntry) *LogEntry {
	value := rt.getFieldValue(entry, rt.rule.Field)
	if value == nil {
		return entry
	}
	
	var newValue interface{}
	
	switch rt.rule.Operation {
	case "lowercase":
		newValue = strings.ToLower(fmt.Sprintf("%v", value))
	case "uppercase":
		newValue = strings.ToUpper(fmt.Sprintf("%v", value))
	case "trim":
		newValue = strings.TrimSpace(fmt.Sprintf("%v", value))
	case "replace":
		old := rt.rule.Options["old"]
		new := rt.rule.Options["new"]
		newValue = strings.ReplaceAll(fmt.Sprintf("%v", value), old, new)
	case "remove":
		rt.removeField(entry, rt.rule.Field)
		return entry
	default:
		return entry
	}
	
	rt.setFieldValue(entry, rt.rule.Field, newValue)
	
	return entry
}

// getFieldValue 获取字段值
func (rt *RuleTransform) getFieldValue(entry *LogEntry, field string) interface{} {
	switch field {
	case "level":
		return string(entry.Level)
	case "message":
		return entry.Message
	case "source":
		return entry.Source
	case "service":
		return entry.Service
	case "trace_id":
		return entry.TraceID
	case "span_id":
		return entry.SpanID
	default:
		if value, exists := entry.Fields[field]; exists {
			return value
		}
		if value, exists := entry.Tags[field]; exists {
			return value
		}
		return nil
	}
}

// setFieldValue 设置字段值
func (rt *RuleTransform) setFieldValue(entry *LogEntry, field string, value interface{}) {
	switch field {
	case "level":
		if str, ok := value.(string); ok {
			entry.Level = LogLevel(str)
		}
	case "message":
		entry.Message = fmt.Sprintf("%v", value)
	case "source":
		entry.Source = fmt.Sprintf("%v", value)
	case "service":
		entry.Service = fmt.Sprintf("%v", value)
	case "trace_id":
		entry.TraceID = fmt.Sprintf("%v", value)
	case "span_id":
		entry.SpanID = fmt.Sprintf("%v", value)
	default:
		if entry.Fields == nil {
			entry.Fields = make(map[string]interface{})
		}
		entry.Fields[field] = value
	}
}

// removeField 移除字段
func (rt *RuleTransform) removeField(entry *LogEntry, field string) {
	switch field {
	case "trace_id":
		entry.TraceID = ""
	case "span_id":
		entry.SpanID = ""
	default:
		if entry.Fields != nil {
			delete(entry.Fields, field)
		}
		if entry.Tags != nil {
			delete(entry.Tags, field)
		}
	}
}

// RateLimitProcessor 限流处理器
type RateLimitProcessor struct {
	name   string
	config RateLimitProcessorConfig
	logger *zap.Logger
	
	mu       sync.Mutex
	counters map[string]*RateCounter
}

// RateLimitProcessorConfig 限流处理器配置
type RateLimitProcessorConfig struct {
	Rate     int           `mapstructure:"rate"`     // 每秒允许的日志数量
	Burst    int           `mapstructure:"burst"`    // 突发允许的日志数量
	Window   time.Duration `mapstructure:"window"`   // 时间窗口
	GroupBy  string        `mapstructure:"group_by"` // 分组字段
	Action   string        `mapstructure:"action"`   // drop, sample
}

// RateCounter 速率计数器
type RateCounter struct {
	count      int
	lastReset  time.Time
	window     time.Duration
	rate       int
	burst      int
}

// NewRateLimitProcessor 创建限流处理器
func NewRateLimitProcessor(name string, config RateLimitProcessorConfig, logger *zap.Logger) *RateLimitProcessor {
	if config.Window == 0 {
		config.Window = time.Second
	}
	if config.Action == "" {
		config.Action = "drop"
	}
	
	return &RateLimitProcessor{
		name:     name,
		config:   config,
		logger:   logger,
		counters: make(map[string]*RateCounter),
	}
}

// Name 返回处理器名称
func (rlp *RateLimitProcessor) Name() string {
	return rlp.name
}

// Process 处理日志条目
func (rlp *RateLimitProcessor) Process(entry *LogEntry) *LogEntry {
	if entry == nil {
		return nil
	}
	
	// 获取分组键
	groupKey := rlp.getGroupKey(entry)
	
	rlp.mu.Lock()
	defer rlp.mu.Unlock()
	
	// 获取或创建计数器
	counter, exists := rlp.counters[groupKey]
	if !exists {
		counter = &RateCounter{
			lastReset: time.Now(),
			window:    rlp.config.Window,
			rate:      rlp.config.Rate,
			burst:     rlp.config.Burst,
		}
		rlp.counters[groupKey] = counter
	}
	
	// 检查是否需要重置计数器
	now := time.Now()
	if now.Sub(counter.lastReset) >= counter.window {
		counter.count = 0
		counter.lastReset = now
	}
	
	// 检查限流
	if counter.count >= counter.rate {
		switch rlp.config.Action {
		case "drop":
			return nil
		case "sample":
			// 简单采样：每N个保留1个
			if counter.count%10 == 0 {
				counter.count++
				return entry
			}
			return nil
		}
	}
	
	counter.count++
	return entry
}

// getGroupKey 获取分组键
func (rlp *RateLimitProcessor) getGroupKey(entry *LogEntry) string {
	if rlp.config.GroupBy == "" {
		return "default"
	}
	
	switch rlp.config.GroupBy {
	case "level":
		return string(entry.Level)
	case "source":
		return entry.Source
	case "service":
		return entry.Service
	default:
		if value, exists := entry.Fields[rlp.config.GroupBy]; exists {
			return fmt.Sprintf("%v", value)
		}
		if value, exists := entry.Tags[rlp.config.GroupBy]; exists {
			return value
		}
		return "default"
	}
}