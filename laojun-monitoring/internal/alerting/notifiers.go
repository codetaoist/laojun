package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Notifier 通知器接口
type Notifier interface {
	// Send 发送通知
	Send(ctx context.Context, alerts []*Alert) error
	
	// Name 返回通知器名称
	Name() string
	
	// IsEnabled 返回是否启用
	IsEnabled() bool
	
	// SetEnabled 设置启用状态
	SetEnabled(enabled bool)
	
	// GetConfig 获取配置
	GetConfig() interface{}
	
	// UpdateConfig 更新配置
	UpdateConfig(config interface{}) error
}

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	logger  *zap.Logger
	config  EmailNotifierConfig
	enabled bool
}

// EmailNotifierConfig 邮件通知器配置
type EmailNotifierConfig struct {
	SMTPHost     string   `mapstructure:"smtp_host"`
	SMTPPort     int      `mapstructure:"smtp_port"`
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	From         string   `mapstructure:"from"`
	To           []string `mapstructure:"to"`
	Subject      string   `mapstructure:"subject"`
	Template     string   `mapstructure:"template"`
	TLS          bool     `mapstructure:"tls"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(logger *zap.Logger, config EmailNotifierConfig) *EmailNotifier {
	if config.Subject == "" {
		config.Subject = "Alert Notification"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	return &EmailNotifier{
		logger:  logger,
		config:  config,
		enabled: true,
	}
}

// Name 返回通知器名称
func (e *EmailNotifier) Name() string {
	return "email"
}

// IsEnabled 返回是否启用
func (e *EmailNotifier) IsEnabled() bool {
	return e.enabled
}

// SetEnabled 设置启用状态
func (e *EmailNotifier) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// GetConfig 获取配置
func (e *EmailNotifier) GetConfig() interface{} {
	return e.config
}

// UpdateConfig 更新配置
func (e *EmailNotifier) UpdateConfig(config interface{}) error {
	if cfg, ok := config.(EmailNotifierConfig); ok {
		e.config = cfg
		return nil
	}
	return fmt.Errorf("invalid config type for email notifier")
}

// Send 发送邮件通知
func (e *EmailNotifier) Send(ctx context.Context, alerts []*Alert) error {
	if !e.enabled {
		return nil
	}
	
	if len(alerts) == 0 {
		return nil
	}
	
	// 构建邮件内容
	subject := e.buildSubject(alerts)
	body := e.buildBody(alerts)
	
	// 发送邮件
	return e.sendEmail(subject, body)
}

// buildSubject 构建邮件主题
func (e *EmailNotifier) buildSubject(alerts []*Alert) string {
	firingCount := 0
	resolvedCount := 0
	
	for _, alert := range alerts {
		switch alert.State {
		case AlertStateFiring:
			firingCount++
		case AlertStateResolved:
			resolvedCount++
		}
	}
	
	if firingCount > 0 && resolvedCount > 0 {
		return fmt.Sprintf("[ALERT] %d firing, %d resolved", firingCount, resolvedCount)
	} else if firingCount > 0 {
		return fmt.Sprintf("[ALERT] %d firing", firingCount)
	} else if resolvedCount > 0 {
		return fmt.Sprintf("[RESOLVED] %d resolved", resolvedCount)
	}
	
	return e.config.Subject
}

// buildBody 构建邮件正文
func (e *EmailNotifier) buildBody(alerts []*Alert) string {
	var buf bytes.Buffer
	
	buf.WriteString("Alert Notification\n")
	buf.WriteString("==================\n\n")
	
	// 按状态分组
	firingAlerts := make([]*Alert, 0)
	resolvedAlerts := make([]*Alert, 0)
	
	for _, alert := range alerts {
		switch alert.State {
		case AlertStateFiring:
			firingAlerts = append(firingAlerts, alert)
		case AlertStateResolved:
			resolvedAlerts = append(resolvedAlerts, alert)
		}
	}
	
	// 输出firing告警
	if len(firingAlerts) > 0 {
		buf.WriteString("FIRING ALERTS:\n")
		buf.WriteString("--------------\n")
		for _, alert := range firingAlerts {
			buf.WriteString(fmt.Sprintf("Rule: %s\n", alert.RuleName))
			buf.WriteString(fmt.Sprintf("Value: %.2f\n", alert.Value))
			buf.WriteString(fmt.Sprintf("Started: %s\n", alert.ActiveAt.Format(time.RFC3339)))
			
			if len(alert.Labels) > 0 {
				buf.WriteString("Labels:\n")
				for k, v := range alert.Labels {
					buf.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
				}
			}
			
			if len(alert.Annotations) > 0 {
				buf.WriteString("Annotations:\n")
				for k, v := range alert.Annotations {
					buf.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
				}
			}
			
			buf.WriteString("\n")
		}
	}
	
	// 输出resolved告警
	if len(resolvedAlerts) > 0 {
		buf.WriteString("RESOLVED ALERTS:\n")
		buf.WriteString("----------------\n")
		for _, alert := range resolvedAlerts {
			buf.WriteString(fmt.Sprintf("Rule: %s\n", alert.RuleName))
			buf.WriteString(fmt.Sprintf("Resolved: %s\n", alert.ResolvedAt.Format(time.RFC3339)))
			buf.WriteString("\n")
		}
	}
	
	return buf.String()
}

// sendEmail 发送邮件
func (e *EmailNotifier) sendEmail(subject, body string) error {
	// 构建邮件消息
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", strings.Join(e.config.To, ","), subject, body)
	
	// SMTP认证
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
	
	// 发送邮件
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)
	return smtp.SendMail(addr, auth, e.config.From, e.config.To, []byte(msg))
}

// WebhookNotifier Webhook通知器
type WebhookNotifier struct {
	logger     *zap.Logger
	config     WebhookNotifierConfig
	enabled    bool
	httpClient *http.Client
}

// WebhookNotifierConfig Webhook通知器配置
type WebhookNotifierConfig struct {
	URL         string            `mapstructure:"url"`
	Method      string            `mapstructure:"method"`
	Headers     map[string]string `mapstructure:"headers"`
	Template    string            `mapstructure:"template"`
	Timeout     time.Duration     `mapstructure:"timeout"`
	MaxRetries  int               `mapstructure:"max_retries"`
	RetryDelay  time.Duration     `mapstructure:"retry_delay"`
}

// NewWebhookNotifier 创建Webhook通知器
func NewWebhookNotifier(logger *zap.Logger, config WebhookNotifierConfig) *WebhookNotifier {
	if config.Method == "" {
		config.Method = "POST"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}
	
	return &WebhookNotifier{
		logger:  logger,
		config:  config,
		enabled: true,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name 返回通知器名称
func (w *WebhookNotifier) Name() string {
	return "webhook"
}

// IsEnabled 返回是否启用
func (w *WebhookNotifier) IsEnabled() bool {
	return w.enabled
}

// SetEnabled 设置启用状态
func (w *WebhookNotifier) SetEnabled(enabled bool) {
	w.enabled = enabled
}

// GetConfig 获取配置
func (w *WebhookNotifier) GetConfig() interface{} {
	return w.config
}

// UpdateConfig 更新配置
func (w *WebhookNotifier) UpdateConfig(config interface{}) error {
	if cfg, ok := config.(WebhookNotifierConfig); ok {
		w.config = cfg
		w.httpClient.Timeout = cfg.Timeout
		return nil
	}
	return fmt.Errorf("invalid config type for webhook notifier")
}

// Send 发送Webhook通知
func (w *WebhookNotifier) Send(ctx context.Context, alerts []*Alert) error {
	if !w.enabled {
		return nil
	}
	
	if len(alerts) == 0 {
		return nil
	}
	
	// 构建请求体
	payload := w.buildPayload(alerts)
	
	// 发送请求
	return w.sendRequest(ctx, payload)
}

// buildPayload 构建请求负载
func (w *WebhookNotifier) buildPayload(alerts []*Alert) map[string]interface{} {
	payload := map[string]interface{}{
		"alerts":    alerts,
		"timestamp": time.Now().Unix(),
		"count":     len(alerts),
	}
	
	// 统计信息
	firingCount := 0
	resolvedCount := 0
	
	for _, alert := range alerts {
		switch alert.State {
		case AlertStateFiring:
			firingCount++
		case AlertStateResolved:
			resolvedCount++
		}
	}
	
	payload["firing_count"] = firingCount
	payload["resolved_count"] = resolvedCount
	
	return payload
}

// sendRequest 发送HTTP请求
func (w *WebhookNotifier) sendRequest(ctx context.Context, payload map[string]interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	var lastErr error
	
	for i := 0; i <= w.config.MaxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(w.config.RetryDelay):
			}
		}
		
		req, err := http.NewRequestWithContext(ctx, w.config.Method, w.config.URL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}
		
		// 设置默认Content-Type
		req.Header.Set("Content-Type", "application/json")
		
		// 设置自定义头部
		for k, v := range w.config.Headers {
			req.Header.Set(k, v)
		}
		
		resp, err := w.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to send request: %w", err)
			continue
		}
		
		resp.Body.Close()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		
		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	
	return lastErr
}

// SlackNotifier Slack通知器
type SlackNotifier struct {
	logger  *zap.Logger
	config  SlackNotifierConfig
	enabled bool
	webhook *WebhookNotifier
}

// SlackNotifierConfig Slack通知器配置
type SlackNotifierConfig struct {
	WebhookURL string `mapstructure:"webhook_url"`
	Channel    string `mapstructure:"channel"`
	Username   string `mapstructure:"username"`
	IconEmoji  string `mapstructure:"icon_emoji"`
	Title      string `mapstructure:"title"`
	Timeout    time.Duration `mapstructure:"timeout"`
}

// NewSlackNotifier 创建Slack通知器
func NewSlackNotifier(logger *zap.Logger, config SlackNotifierConfig) *SlackNotifier {
	if config.Username == "" {
		config.Username = "AlertManager"
	}
	if config.IconEmoji == "" {
		config.IconEmoji = ":warning:"
	}
	if config.Title == "" {
		config.Title = "Alert Notification"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	// 创建底层webhook通知器
	webhookConfig := WebhookNotifierConfig{
		URL:     config.WebhookURL,
		Method:  "POST",
		Timeout: config.Timeout,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	
	return &SlackNotifier{
		logger:  logger,
		config:  config,
		enabled: true,
		webhook: NewWebhookNotifier(logger, webhookConfig),
	}
}

// Name 返回通知器名称
func (s *SlackNotifier) Name() string {
	return "slack"
}

// IsEnabled 返回是否启用
func (s *SlackNotifier) IsEnabled() bool {
	return s.enabled
}

// SetEnabled 设置启用状态
func (s *SlackNotifier) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// GetConfig 获取配置
func (s *SlackNotifier) GetConfig() interface{} {
	return s.config
}

// UpdateConfig 更新配置
func (s *SlackNotifier) UpdateConfig(config interface{}) error {
	if cfg, ok := config.(SlackNotifierConfig); ok {
		s.config = cfg
		return nil
	}
	return fmt.Errorf("invalid config type for slack notifier")
}

// Send 发送Slack通知
func (s *SlackNotifier) Send(ctx context.Context, alerts []*Alert) error {
	if !s.enabled {
		return nil
	}
	
	if len(alerts) == 0 {
		return nil
	}
	
	// 构建Slack消息
	message := s.buildSlackMessage(alerts)
	
	// 使用webhook发送
	return s.sendSlackMessage(ctx, message)
}

// buildSlackMessage 构建Slack消息
func (s *SlackNotifier) buildSlackMessage(alerts []*Alert) map[string]interface{} {
	firingCount := 0
	resolvedCount := 0
	
	for _, alert := range alerts {
		switch alert.State {
		case AlertStateFiring:
			firingCount++
		case AlertStateResolved:
			resolvedCount++
		}
	}
	
	// 构建颜色
	color := "good"
	if firingCount > 0 {
		color = "danger"
	}
	
	// 构建标题
	title := s.config.Title
	if firingCount > 0 && resolvedCount > 0 {
		title = fmt.Sprintf("%s - %d firing, %d resolved", title, firingCount, resolvedCount)
	} else if firingCount > 0 {
		title = fmt.Sprintf("%s - %d firing", title, firingCount)
	} else if resolvedCount > 0 {
		title = fmt.Sprintf("%s - %d resolved", title, resolvedCount)
	}
	
	// 构建字段
	fields := make([]map[string]interface{}, 0)
	
	for _, alert := range alerts {
		field := map[string]interface{}{
			"title": alert.RuleName,
			"value": fmt.Sprintf("State: %s\nValue: %.2f", alert.State, alert.Value),
			"short": true,
		}
		fields = append(fields, field)
	}
	
	attachment := map[string]interface{}{
		"color":  color,
		"title":  title,
		"fields": fields,
		"ts":     time.Now().Unix(),
	}
	
	message := map[string]interface{}{
		"username":    s.config.Username,
		"icon_emoji":  s.config.IconEmoji,
		"attachments": []map[string]interface{}{attachment},
	}
	
	if s.config.Channel != "" {
		message["channel"] = s.config.Channel
	}
	
	return message
}

// sendSlackMessage 发送Slack消息
func (s *SlackNotifier) sendSlackMessage(ctx context.Context, message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: s.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}
	
	return nil
}

// NotifierManager 通知器管理器
type NotifierManager struct {
	notifiers map[string]Notifier
	logger    *zap.Logger
}

// NewNotifierManager 创建通知器管理器
func NewNotifierManager(logger *zap.Logger) *NotifierManager {
	return &NotifierManager{
		notifiers: make(map[string]Notifier),
		logger:    logger,
	}
}

// Register 注册通知器
func (nm *NotifierManager) Register(name string, notifier Notifier) {
	nm.notifiers[name] = notifier
	nm.logger.Info("Registered notifier", zap.String("name", name))
}

// Unregister 注销通知器
func (nm *NotifierManager) Unregister(name string) {
	delete(nm.notifiers, name)
	nm.logger.Info("Unregistered notifier", zap.String("name", name))
}

// Get 获取通知器
func (nm *NotifierManager) Get(name string) (Notifier, bool) {
	notifier, exists := nm.notifiers[name]
	return notifier, exists
}

// List 列出所有通知器
func (nm *NotifierManager) List() []string {
	names := make([]string, 0, len(nm.notifiers))
	for name := range nm.notifiers {
		names = append(names, name)
	}
	return names
}

// SendToAll 发送到所有启用的通知器
func (nm *NotifierManager) SendToAll(ctx context.Context, alerts []*Alert) error {
	var lastErr error
	
	for name, notifier := range nm.notifiers {
		if !notifier.IsEnabled() {
			continue
		}
		
		if err := notifier.Send(ctx, alerts); err != nil {
			nm.logger.Error("Failed to send notification", 
				zap.String("notifier", name), 
				zap.Error(err))
			lastErr = err
		}
	}
	
	return lastErr
}

// SendToSpecific 发送到指定通知器
func (nm *NotifierManager) SendToSpecific(ctx context.Context, notifierNames []string, alerts []*Alert) error {
	var lastErr error
	
	for _, name := range notifierNames {
		notifier, exists := nm.notifiers[name]
		if !exists {
			nm.logger.Warn("Notifier not found", zap.String("name", name))
			continue
		}
		
		if !notifier.IsEnabled() {
			continue
		}
		
		if err := notifier.Send(ctx, alerts); err != nil {
			nm.logger.Error("Failed to send notification", 
				zap.String("notifier", name), 
				zap.Error(err))
			lastErr = err
		}
	}
	
	return lastErr
}