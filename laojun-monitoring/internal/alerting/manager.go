package alerting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/query"
	"github.com/sirupsen/logrus"
)

// AlertManager 告警管理器
type AlertManager struct {
	mu       sync.RWMutex
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *logrus.Logger
	config   AlertingConfig
	querier  query.MetricQuerier

	// 告警规则和状态
	rules       map[string]*AlertRule
	alerts      map[string]*Alert
	silences    map[string]*Silence
	receivers   map[string]*Receiver
	
	// 统计信息
	stats ManagerStats
}

// AlertingConfig 告警配置
type AlertingConfig struct {
	Enabled            bool          `yaml:"enabled"`
	EvaluationInterval time.Duration `yaml:"evaluation_interval"`
	Rules              []AlertRule   `yaml:"rules"`
	Receivers          []Receiver    `yaml:"receivers"`
	Routes             []Route       `yaml:"routes"`
}

// 类型定义已移至types.go文件

// Route 路由规则
type Route struct {
	Matchers map[string]string `yaml:"matchers" json:"matchers"`
	Receiver string            `yaml:"receiver" json:"receiver"`
}

// ManagerStats 管理器统计信息
type ManagerStats struct {
	RuleCount       int       `json:"rule_count"`
	AlertCount      int       `json:"alert_count"`
	SilenceCount    int       `json:"silence_count"`
	EvaluationCount int64     `json:"evaluation_count"`
	LastEvaluation  time.Time `json:"last_evaluation"`
	ErrorCount      int64     `json:"error_count"`
	LastError       string    `json:"last_error,omitempty"`
}

// NewAlertManager 创建新的告警管理器
func NewAlertManager(config AlertingConfig, querier query.MetricQuerier, logger *logrus.Logger) *AlertManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	am := &AlertManager{
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		config:    config,
		querier:   querier,
		rules:     make(map[string]*AlertRule),
		alerts:    make(map[string]*Alert),
		silences:  make(map[string]*Silence),
		receivers: make(map[string]*Receiver),
	}
	
	// 初始化规则和接收器
	for i := range config.Rules {
		rule := &config.Rules[i]
		am.rules[rule.ID] = rule
	}
	
	for i := range config.Receivers {
		receiver := &config.Receivers[i]
		am.receivers[receiver.Name] = receiver
	}
	
	return am
}

// Start 启动告警管理器
func (am *AlertManager) Start() error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	if am.running {
		return nil
	}
	
	if !am.config.Enabled {
		am.logger.Info("Alert manager is disabled")
		return nil
	}
	
	am.running = true
	am.logger.Info("Starting alert manager")
	
	go am.evaluationLoop()
	go am.cleanupLoop()
	
	return nil
}

// Stop 停止告警管理器
func (am *AlertManager) Stop() error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	if !am.running {
		return nil
	}
	
	am.running = false
	am.cancel()
	am.logger.Info("Stopping alert manager")
	
	return nil
}

// IsRunning 检查管理器是否运行中
func (am *AlertManager) IsRunning() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.running
}

// GetStats 获取管理器统计信息
func (am *AlertManager) GetStats() ManagerStats {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	am.stats.RuleCount = len(am.rules)
	am.stats.AlertCount = len(am.alerts)
	am.stats.SilenceCount = len(am.silences)
	
	return am.stats
}

// AddRule 添加告警规则
func (am *AlertManager) AddRule(rule *AlertRule) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().Unix())
	}
	
	am.rules[rule.ID] = rule
	am.logger.WithField("rule_id", rule.ID).Info("Alert rule added")
	
	return nil
}

// RemoveRule 移除告警规则
func (am *AlertManager) RemoveRule(ruleID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	delete(am.rules, ruleID)
	
	// 移除相关的告警
	for alertID, alert := range am.alerts {
		if alert.RuleID == ruleID {
			delete(am.alerts, alertID)
		}
	}
	
	am.logger.WithField("rule_id", ruleID).Info("Alert rule removed")
	
	return nil
}

// GetRules 获取所有规则
func (am *AlertManager) GetRules() []*AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}
	
	return rules
}

// GetAlerts 获取所有告警
func (am *AlertManager) GetAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}
	
	return alerts
}

// CreateSilence 创建静默规则
func (am *AlertManager) CreateSilence(silence *Silence) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	if silence.ID == "" {
		silence.ID = fmt.Sprintf("silence_%d", time.Now().Unix())
	}
	
	am.silences[silence.ID] = silence
	am.logger.WithField("silence_id", silence.ID).Info("Silence created")
	
	return nil
}

// RemoveSilence 移除静默规则
func (am *AlertManager) RemoveSilence(silenceID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	delete(am.silences, silenceID)
	am.logger.WithField("silence_id", silenceID).Info("Silence removed")
	
	return nil
}

// evaluationLoop 评估循环
func (am *AlertManager) evaluationLoop() {
	ticker := time.NewTicker(am.config.EvaluationInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.evaluate()
		}
	}
}

// evaluate 执行一次评估
func (am *AlertManager) evaluate() {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.stats.EvaluationCount++
	am.stats.LastEvaluation = time.Now()
	
	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}
		
		// 这里应该执行实际的查询和评估
		// 为了演示，我们创建一些模拟告警
		am.evaluateRule(rule)
	}
	
	// 检查告警是否应该被解决
	am.checkResolvedAlerts()
	
	am.logger.WithFields(logrus.Fields{
		"evaluation_count": am.stats.EvaluationCount,
		"rule_count":       len(am.rules),
		"alert_count":      len(am.alerts),
	}).Debug("Alert evaluation completed")
}

// evaluateRule 评估单个规则
func (am *AlertManager) evaluateRule(rule *AlertRule) {
	// 模拟查询执行和结果评估
	// 在实际实现中，这里会执行 Prometheus 查询或其他指标查询
	
	alertID := fmt.Sprintf("%s_%d", rule.ID, time.Now().Unix())
	
	// 检查是否已存在相同的告警
	existingAlert := am.findExistingAlert(rule.ID)
	if existingAlert != nil {
		// 更新现有告警
		existingAlert.UpdatedAt = time.Now()
		return
	}
	
	// 创建新告警（模拟条件触发）
	if am.shouldTriggerAlert(rule) {
		alert := &Alert{
			ID:          alertID,
			RuleID:      rule.ID,
			Name:        rule.Name,
			Status:      "firing",
			Severity:    rule.Severity,
			Message:     fmt.Sprintf("Alert %s is firing", rule.Name),
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
			StartsAt:    time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// 检查是否被静默
		if am.isAlertSilenced(alert) {
			alert.Status = "silenced"
		}
		
		am.alerts[alertID] = alert
		
		// 发送告警通知
		am.sendAlert(alert)
		
		am.logger.WithFields(logrus.Fields{
			"alert_id":  alertID,
			"rule_id":   rule.ID,
			"severity":  rule.Severity,
			"status":    alert.Status,
		}).Info("Alert triggered")
	}
}

// shouldTriggerAlert 判断是否应该触发告警
func (am *AlertManager) shouldTriggerAlert(rule *AlertRule) bool {
	// 使用查询器执行指标查询
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	response, err := am.querier.Query(ctx, rule.Query)
	if err != nil {
		am.logger.WithFields(logrus.Fields{
			"rule_id": rule.ID,
			"query":   rule.Query,
			"error":   err.Error(),
		}).Error("Failed to execute query for alert rule")
		
		// 更新错误统计
		am.mu.Lock()
		am.stats.ErrorCount++
		am.stats.LastError = err.Error()
		am.mu.Unlock()
		
		return false
	}
	
	// 如果查询返回结果，说明条件满足
	return len(response.Results) > 0
}

// findExistingAlert 查找现有告警
func (am *AlertManager) findExistingAlert(ruleID string) *Alert {
	for _, alert := range am.alerts {
		if alert.RuleID == ruleID && alert.Status == "firing" {
			return alert
		}
	}
	return nil
}

// isAlertSilenced 检查告警是否被静默
func (am *AlertManager) isAlertSilenced(alert *Alert) bool {
	now := time.Now()
	
	for _, silence := range am.silences {
		if now.Before(silence.StartsAt) || now.After(silence.EndsAt) {
			continue
		}
		
		// 检查标签匹配
		if am.matchLabels(alert.Labels, silence.Matchers) {
			return true
		}
	}
	
	return false
}

// matchLabels 检查标签是否匹配
func (am *AlertManager) matchLabels(alertLabels, matchers map[string]string) bool {
	for key, value := range matchers {
		if alertLabels[key] != value {
			return false
		}
	}
	return true
}

// checkResolvedAlerts 检查已解决的告警
func (am *AlertManager) checkResolvedAlerts() {
	now := time.Now()
	
	for alertID, alert := range am.alerts {
		if alert.Status == "firing" {
			// 检查告警是否应该被解决
			// 这里是模拟逻辑，实际实现会重新评估条件
			if am.shouldResolveAlert(alert) {
				alert.Status = "resolved"
				alert.EndsAt = &now
				alert.UpdatedAt = now
				
				am.logger.WithField("alert_id", alertID).Info("Alert resolved")
			}
		}
		
		// 清理旧的已解决告警
		if alert.Status == "resolved" && alert.EndsAt != nil {
			if now.Sub(*alert.EndsAt) > 24*time.Hour {
				delete(am.alerts, alertID)
			}
		}
	}
}

// shouldResolveAlert 判断告警是否应该被解决
func (am *AlertManager) shouldResolveAlert(alert *Alert) bool {
	// 查找对应的规则
	rule, exists := am.rules[alert.RuleID]
	if !exists {
		// 如果规则不存在，解决告警
		return true
	}
	
	// 重新评估规则条件
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	response, err := am.querier.Query(ctx, rule.Query)
	if err != nil {
		am.logger.WithFields(logrus.Fields{
			"alert_id": alert.ID,
			"rule_id":  alert.RuleID,
			"query":    rule.Query,
			"error":    err.Error(),
		}).Error("Failed to execute query for alert resolution check")
		
		// 查询失败时，如果告警持续时间超过规则持续时间的2倍，则解决告警
		return time.Since(alert.StartsAt) > rule.Duration*2
	}
	
	// 如果查询没有返回结果，说明条件不再满足，应该解决告警
	return len(response.Results) == 0
}

// sendAlert 发送告警通知
func (am *AlertManager) sendAlert(alert *Alert) {
	// 这里应该根据路由规则发送到相应的接收器
	// 为了演示，我们只记录日志
	
	am.logger.WithFields(logrus.Fields{
		"alert_id":  alert.ID,
		"name":      alert.Name,
		"severity":  alert.Severity,
		"status":    alert.Status,
		"message":   alert.Message,
	}).Info("Alert notification sent")
}

// cleanupLoop 清理循环
func (am *AlertManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.cleanup()
		}
	}
}

// cleanup 清理过期数据
func (am *AlertManager) cleanup() {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	now := time.Now()
	
	// 清理过期的静默规则
	for silenceID, silence := range am.silences {
		if now.After(silence.EndsAt) {
			delete(am.silences, silenceID)
			am.logger.WithField("silence_id", silenceID).Debug("Expired silence removed")
		}
	}
}

// Health 检查管理器健康状态
func (am *AlertManager) Health() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	status := "healthy"
	if !am.running {
		status = "stopped"
	} else if time.Since(am.stats.LastEvaluation) > am.config.EvaluationInterval*2 {
		status = "unhealthy"
	}
	
	return map[string]interface{}{
		"status":            status,
		"running":           am.running,
		"rule_count":        len(am.rules),
		"alert_count":       len(am.alerts),
		"silence_count":     len(am.silences),
		"evaluation_count":  am.stats.EvaluationCount,
		"last_evaluation":   am.stats.LastEvaluation,
		"error_count":       am.stats.ErrorCount,
		"last_error":        am.stats.LastError,
		"evaluation_interval": am.config.EvaluationInterval.String(),
	}
}

// IsHealthy 检查告警管理器是否健康
func (am *AlertManager) IsHealthy() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	return am.running
}

// IsReady 检查告警管理器是否准备就绪
func (am *AlertManager) IsReady() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.running
}

// ListRules 获取所有规则列表
func (am *AlertManager) ListRules() []*AlertRule {
	return am.GetRules()
}

// GetRule 根据ID获取规则
func (am *AlertManager) GetRule(ruleID string) *AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.rules[ruleID]
}

// UpdateRule 更新规则
func (am *AlertManager) UpdateRule(rule *AlertRule) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	if _, exists := am.rules[rule.ID]; !exists {
		return fmt.Errorf("rule with ID %s not found", rule.ID)
	}
	
	am.rules[rule.ID] = rule
	return nil
}

// GetAlert 根据ID获取告警
func (am *AlertManager) GetAlert(alertID string) *Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.alerts[alertID]
}