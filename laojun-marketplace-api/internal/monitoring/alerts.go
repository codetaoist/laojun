package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// Alert 告警信息
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Level       AlertLevel             `json:"level"`
	Status      AlertStatus            `json:"status"`
	Message     string                 `json:"message"`
	Description string                 `json:"description"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	StartsAt    time.Time              `json:"starts_at"`
	EndsAt      *time.Time             `json:"ends_at,omitempty"`
	GeneratorURL string                `json:"generator_url,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Expression  string                 `json:"expression"`
	Level       AlertLevel             `json:"level"`
	Duration    time.Duration          `json:"duration"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NotificationChannel 通知渠道接口
type NotificationChannel interface {
	Name() string
	Send(ctx context.Context, alert *Alert) error
}

// AlertManager 告警管理器
type AlertManager struct {
	rules       map[string]*AlertRule
	channels    map[string]NotificationChannel
	activeAlerts map[string]*Alert
	logger      *zap.Logger
	mu          sync.RWMutex
}

// NewAlertManager 创建告警管理器
func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		rules:        make(map[string]*AlertRule),
		channels:     make(map[string]NotificationChannel),
		activeAlerts: make(map[string]*Alert),
		logger:       logger,
	}
}

// AddRule 添加告警规则
func (am *AlertManager) AddRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.rules[rule.ID] = rule
}

// RemoveRule 移除告警规则
func (am *AlertManager) RemoveRule(ruleID string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.rules, ruleID)
}

// AddChannel 添加通知渠道
func (am *AlertManager) AddChannel(channel NotificationChannel) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.channels[channel.Name()] = channel
}

// RemoveChannel 移除通知渠道
func (am *AlertManager) RemoveChannel(channelName string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.channels, channelName)
}

// FireAlert 触发告警
func (am *AlertManager) FireAlert(alert *Alert) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert.Status = AlertStatusFiring
	alert.StartsAt = time.Now()
	am.activeAlerts[alert.ID] = alert

	am.logger.Warn("Alert fired",
		zap.String("alert_id", alert.ID),
		zap.String("name", alert.Name),
		zap.String("level", string(alert.Level)),
		zap.String("message", alert.Message),
	)

	// 发送通知
	return am.sendNotifications(context.Background(), alert)
}

// ResolveAlert 解决告警
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.EndsAt = &now

	am.logger.Info("Alert resolved",
		zap.String("alert_id", alert.ID),
		zap.String("name", alert.Name),
		zap.Duration("duration", now.Sub(alert.StartsAt)),
	)

	// 发送解决通知
	if err := am.sendNotifications(context.Background(), alert); err != nil {
		return err
	}

	delete(am.activeAlerts, alertID)
	return nil
}

// GetActiveAlerts 获取活跃告警
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// sendNotifications 发送通知
func (am *AlertManager) sendNotifications(ctx context.Context, alert *Alert) error {
	var wg sync.WaitGroup
	var errors []error
	var mu sync.Mutex

	for _, channel := range am.channels {
		wg.Add(1)
		go func(ch NotificationChannel) {
			defer wg.Done()
			if err := ch.Send(ctx, alert); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("channel %s: %w", ch.Name(), err))
				mu.Unlock()
			}
		}(channel)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}
	return nil
}

// WebhookChannel Webhook通知渠道
type WebhookChannel struct {
	name   string
	url    string
	client *http.Client
}

// NewWebhookChannel 创建Webhook通知渠道
func NewWebhookChannel(name, url string) *WebhookChannel {
	return &WebhookChannel{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (wc *WebhookChannel) Name() string {
	return wc.name
}

func (wc *WebhookChannel) Send(ctx context.Context, alert *Alert) error {
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", wc.url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := wc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// EmailChannel 邮件通知渠道
type EmailChannel struct {
	name     string
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
}

// NewEmailChannel 创建邮件通知渠道
func NewEmailChannel(name, smtpHost string, smtpPort int, username, password, from string, to []string) *EmailChannel {
	return &EmailChannel{
		name:     name,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

func (ec *EmailChannel) Name() string {
	return ec.name
}

func (ec *EmailChannel) Send(ctx context.Context, alert *Alert) error {
	// 这里应该实现实际的邮件发送逻辑
	// 为了简化，这里只记录日志
	fmt.Printf("Sending email alert: %s to %v\n", alert.Name, ec.to)
	return nil
}

// SlackChannel Slack通知渠道
type SlackChannel struct {
	name      string
	webhookURL string
	client    *http.Client
}

// NewSlackChannel 创建Slack通知渠道
func NewSlackChannel(name, webhookURL string) *SlackChannel {
	return &SlackChannel{
		name:       name,
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (sc *SlackChannel) Name() string {
	return sc.name
}

func (sc *SlackChannel) Send(ctx context.Context, alert *Alert) error {
	color := "good"
	switch alert.Level {
	case AlertLevelWarning:
		color = "warning"
	case AlertLevelCritical:
		color = "danger"
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":      color,
				"title":      alert.Name,
				"text":       alert.Message,
				"fields": []map[string]interface{}{
					{
						"title": "Level",
						"value": string(alert.Level),
						"short": true,
					},
					{
						"title": "Status",
						"value": string(alert.Status),
						"short": true,
					},
					{
						"title": "Time",
						"value": alert.StartsAt.Format(time.RFC3339),
						"short": true,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sc.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// PredefinedAlerts 预定义告警
func GetPredefinedAlerts() []*AlertRule {
	return []*AlertRule{
		{
			ID:         "high_error_rate",
			Name:       "High Error Rate",
			Expression: "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m]) > 0.05",
			Level:      AlertLevelWarning,
			Duration:   2 * time.Minute,
			Labels: map[string]string{
				"service": "marketplace-api",
				"type":    "error_rate",
			},
			Annotations: map[string]string{
				"summary":     "High error rate detected",
				"description": "Error rate is above 5% for more than 2 minutes",
			},
			Enabled: true,
		},
		{
			ID:         "high_response_time",
			Name:       "High Response Time",
			Expression: "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2",
			Level:      AlertLevelWarning,
			Duration:   5 * time.Minute,
			Labels: map[string]string{
				"service": "marketplace-api",
				"type":    "latency",
			},
			Annotations: map[string]string{
				"summary":     "High response time detected",
				"description": "95th percentile response time is above 2 seconds",
			},
			Enabled: true,
		},
		{
			ID:         "database_connection_pool_exhausted",
			Name:       "Database Connection Pool Exhausted",
			Expression: "db_connections_in_use / db_connections_max > 0.9",
			Level:      AlertLevelCritical,
			Duration:   1 * time.Minute,
			Labels: map[string]string{
				"service": "marketplace-api",
				"type":    "database",
			},
			Annotations: map[string]string{
				"summary":     "Database connection pool nearly exhausted",
				"description": "More than 90% of database connections are in use",
			},
			Enabled: true,
		},
		{
			ID:         "redis_connection_failed",
			Name:       "Redis Connection Failed",
			Expression: "redis_up == 0",
			Level:      AlertLevelCritical,
			Duration:   30 * time.Second,
			Labels: map[string]string{
				"service": "marketplace-api",
				"type":    "redis",
			},
			Annotations: map[string]string{
				"summary":     "Redis connection failed",
				"description": "Unable to connect to Redis server",
			},
			Enabled: true,
		},
	}
}