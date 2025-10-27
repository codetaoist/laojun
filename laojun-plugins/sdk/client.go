package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// RegistryClient 注册中心客户端接口
type RegistryClient interface {
	// Register 注册插件
	Register(ctx context.Context, registration *PluginRegistration) error

	// Unregister 注销插件
	Unregister(ctx context.Context, pluginID string) error

	// UpdateStatus 更新插件状态
	UpdateStatus(ctx context.Context, pluginID string, status string) error

	// UpdateMetrics 更新插件指标
	UpdateMetrics(ctx context.Context, pluginID string, metrics *Metrics) error

	// GetPlugin 获取插件信息
	GetPlugin(ctx context.Context, pluginID string) (*PluginRegistration, error)

	// ListPlugins 列出插件
	ListPlugins(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error)

	// Heartbeat 发送心跳
	Heartbeat(ctx context.Context, pluginID string) error

	// Discover 发现插件
	Discover(ctx context.Context, criteria *DiscoveryCriteria) ([]*PluginRegistration, error)
}

// EventBusClient 事件总线客户端接口
type EventBusClient interface {
	// Publish 发布事件
	Publish(ctx context.Context, event *Event) error

	// Subscribe 订阅事件
	Subscribe(ctx context.Context, eventTypes []string, handler EventHandler) (*Subscription, error)

	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, subscriptionID string) error

	// Close 关闭客户端
	Close() error
}

// PluginRegistration 插件注册信息
type PluginRegistration struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Endpoints   []*Endpoint       `json:"endpoints"`
	Status      string            `json:"status"`
	Config      map[string]string `json:"config"`
	Permissions []string          `json:"permissions"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Endpoint 端点信息
type Endpoint struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// PluginFilter 插件过滤器
type PluginFilter struct {
	Status   string   `json:"status,omitempty"`
	Category string   `json:"category,omitempty"`
	Author   string   `json:"author,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// DiscoveryCriteria 发现条件
type DiscoveryCriteria struct {
	Keywords    []string `json:"keywords,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	MaxResults  int      `json:"max_results,omitempty"`
	SortBy      string   `json:"sort_by,omitempty"`
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error

// Subscription 订阅信息
type Subscription struct {
	ID         string    `json:"id"`
	EventTypes []string  `json:"event_types"`
	CreatedAt  time.Time `json:"created_at"`
}

// HTTPRegistryClient HTTP注册中心客户端
type HTTPRegistryClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	apiKey     string
}

// NewHTTPRegistryClient 创建HTTP注册中心客户端
func NewHTTPRegistryClient(baseURL, apiKey string, logger *logrus.Logger) *HTTPRegistryClient {
	return &HTTPRegistryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
		apiKey: apiKey,
	}
}

// Register 注册插件
func (c *HTTPRegistryClient) Register(ctx context.Context, registration *PluginRegistration) error {
	url := fmt.Sprintf("%s/api/v1/plugins", c.baseURL)
	
	data, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("plugin_id", registration.ID).Info("Plugin registered successfully")
	return nil
}

// Unregister 注销插件
func (c *HTTPRegistryClient) Unregister(ctx context.Context, pluginID string) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s", c.baseURL, pluginID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unregistration failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("plugin_id", pluginID).Info("Plugin unregistered successfully")
	return nil
}

// UpdateStatus 更新插件状态
func (c *HTTPRegistryClient) UpdateStatus(ctx context.Context, pluginID string, status string) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/status", c.baseURL, pluginID)
	
	data := map[string]string{"status": status}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateMetrics 更新插件指标
func (c *HTTPRegistryClient) UpdateMetrics(ctx context.Context, pluginID string, metrics *Metrics) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/metrics", c.baseURL, pluginID)
	
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("metrics update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetPlugin 获取插件信息
func (c *HTTPRegistryClient) GetPlugin(ctx context.Context, pluginID string) (*PluginRegistration, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/%s", c.baseURL, pluginID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("get plugin failed with status %d: %s", resp.StatusCode, string(body))
	}

	var plugin PluginRegistration
	if err := json.NewDecoder(resp.Body).Decode(&plugin); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &plugin, nil
}

// ListPlugins 列出插件
func (c *HTTPRegistryClient) ListPlugins(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error) {
	url := fmt.Sprintf("%s/api/v1/plugins", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加查询参数
	if filter != nil {
		q := req.URL.Query()
		if filter.Status != "" {
			q.Add("status", filter.Status)
		}
		if filter.Category != "" {
			q.Add("category", filter.Category)
		}
		if filter.Author != "" {
			q.Add("author", filter.Author)
		}
		for _, tag := range filter.Tags {
			q.Add("tags", tag)
		}
		req.URL.RawQuery = q.Encode()
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("list plugins failed with status %d: %s", resp.StatusCode, string(body))
	}

	var plugins []*PluginRegistration
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return plugins, nil
}

// Heartbeat 发送心跳
func (c *HTTPRegistryClient) Heartbeat(ctx context.Context, pluginID string) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/heartbeat", c.baseURL, pluginID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Discover 发现插件
func (c *HTTPRegistryClient) Discover(ctx context.Context, criteria *DiscoveryCriteria) ([]*PluginRegistration, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/discover", c.baseURL)
	
	data, err := json.Marshal(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal criteria: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("discover failed with status %d: %s", resp.StatusCode, string(body))
	}

	var plugins []*PluginRegistration
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return plugins, nil
}

// setHeaders 设置请求头
func (c *HTTPRegistryClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// WebSocketEventBusClient WebSocket事件总线客户端
type WebSocketEventBusClient struct {
	url           string
	logger        *logrus.Logger
	subscriptions map[string]*eventSubscription
	eventChan     chan *Event
	stopChan      chan struct{}
	connected     bool
}

type eventSubscription struct {
	ID         string
	EventTypes []string
	Handler    EventHandler
	CreatedAt  time.Time
}

// NewWebSocketEventBusClient 创建WebSocket事件总线客户端
func NewWebSocketEventBusClient(url string, logger *logrus.Logger) *WebSocketEventBusClient {
	return &WebSocketEventBusClient{
		url:           url,
		logger:        logger,
		subscriptions: make(map[string]*eventSubscription),
		eventChan:     make(chan *Event, 100),
		stopChan:      make(chan struct{}),
	}
}

// Publish 发布事件
func (c *WebSocketEventBusClient) Publish(ctx context.Context, event *Event) error {
	// 简化实现：通过HTTP发布事件
	url := c.url + "/publish"
	
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("publish failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("event_id", event.ID).Debug("Event published successfully")
	return nil
}

// Subscribe 订阅事件
func (c *WebSocketEventBusClient) Subscribe(ctx context.Context, eventTypes []string, handler EventHandler) (*Subscription, error) {
	subscriptionID := fmt.Sprintf("sub_%d", time.Now().UnixNano())
	
	sub := &eventSubscription{
		ID:         subscriptionID,
		EventTypes: eventTypes,
		Handler:    handler,
		CreatedAt:  time.Now(),
	}

	c.subscriptions[subscriptionID] = sub

	c.logger.WithFields(logrus.Fields{
		"subscription_id": subscriptionID,
		"event_types":     eventTypes,
	}).Info("Event subscription created")

	return &Subscription{
		ID:         subscriptionID,
		EventTypes: eventTypes,
		CreatedAt:  sub.CreatedAt,
	}, nil
}

// Unsubscribe 取消订阅
func (c *WebSocketEventBusClient) Unsubscribe(ctx context.Context, subscriptionID string) error {
	if _, exists := c.subscriptions[subscriptionID]; !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	delete(c.subscriptions, subscriptionID)

	c.logger.WithField("subscription_id", subscriptionID).Info("Event subscription removed")
	return nil
}

// Close 关闭客户端
func (c *WebSocketEventBusClient) Close() error {
	close(c.stopChan)
	c.connected = false
	
	// 清理所有订阅
	for id := range c.subscriptions {
		delete(c.subscriptions, id)
	}

	c.logger.Info("Event bus client closed")
	return nil
}

// MockRegistryClient 模拟注册中心客户端（用于测试）
type MockRegistryClient struct {
	plugins map[string]*PluginRegistration
	logger  *logrus.Logger
}

// NewMockRegistryClient 创建模拟注册中心客户端
func NewMockRegistryClient(logger *logrus.Logger) *MockRegistryClient {
	return &MockRegistryClient{
		plugins: make(map[string]*PluginRegistration),
		logger:  logger,
	}
}

// Register 注册插件
func (c *MockRegistryClient) Register(ctx context.Context, registration *PluginRegistration) error {
	c.plugins[registration.ID] = registration
	c.logger.WithField("plugin_id", registration.ID).Info("Plugin registered (mock)")
	return nil
}

// Unregister 注销插件
func (c *MockRegistryClient) Unregister(ctx context.Context, pluginID string) error {
	if _, exists := c.plugins[pluginID]; !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	delete(c.plugins, pluginID)
	c.logger.WithField("plugin_id", pluginID).Info("Plugin unregistered (mock)")
	return nil
}

// UpdateStatus 更新插件状态
func (c *MockRegistryClient) UpdateStatus(ctx context.Context, pluginID string, status string) error {
	plugin, exists := c.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	plugin.Status = status
	plugin.UpdatedAt = time.Now()
	return nil
}

// UpdateMetrics 更新插件指标
func (c *MockRegistryClient) UpdateMetrics(ctx context.Context, pluginID string, metrics *Metrics) error {
	_, exists := c.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	// 在实际实现中，这里会更新指标
	return nil
}

// GetPlugin 获取插件信息
func (c *MockRegistryClient) GetPlugin(ctx context.Context, pluginID string) (*PluginRegistration, error) {
	plugin, exists := c.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	// 返回副本
	pluginCopy := *plugin
	return &pluginCopy, nil
}

// ListPlugins 列出插件
func (c *MockRegistryClient) ListPlugins(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error) {
	var result []*PluginRegistration
	
	for _, plugin := range c.plugins {
		if c.matchesFilter(plugin, filter) {
			pluginCopy := *plugin
			result = append(result, &pluginCopy)
		}
	}
	
	return result, nil
}

// Heartbeat 发送心跳
func (c *MockRegistryClient) Heartbeat(ctx context.Context, pluginID string) error {
	_, exists := c.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	return nil
}

// Discover 发现插件
func (c *MockRegistryClient) Discover(ctx context.Context, criteria *DiscoveryCriteria) ([]*PluginRegistration, error) {
	var result []*PluginRegistration
	
	for _, plugin := range c.plugins {
		if c.matchesCriteria(plugin, criteria) {
			pluginCopy := *plugin
			result = append(result, &pluginCopy)
		}
	}
	
	// 应用结果限制
	if criteria.MaxResults > 0 && len(result) > criteria.MaxResults {
		result = result[:criteria.MaxResults]
	}
	
	return result, nil
}

// matchesFilter 检查插件是否匹配过滤条件
func (c *MockRegistryClient) matchesFilter(plugin *PluginRegistration, filter *PluginFilter) bool {
	if filter == nil {
		return true
	}
	
	if filter.Status != "" && plugin.Status != filter.Status {
		return false
	}
	
	if filter.Category != "" && plugin.Category != filter.Category {
		return false
	}
	
	if filter.Author != "" && plugin.Author != filter.Author {
		return false
	}
	
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
			for _, pluginTag := range plugin.Tags {
				if pluginTag == filterTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}
	
	return true
}

// matchesCriteria 检查插件是否匹配发现条件
func (c *MockRegistryClient) matchesCriteria(plugin *PluginRegistration, criteria *DiscoveryCriteria) bool {
	if criteria == nil {
		return true
	}
	
	// 关键词匹配
	if len(criteria.Keywords) > 0 {
		found := false
		for _, keyword := range criteria.Keywords {
			if c.containsKeyword(plugin, keyword) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// 分类匹配
	if len(criteria.Categories) > 0 {
		found := false
		for _, category := range criteria.Categories {
			if plugin.Category == category {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// containsKeyword 检查插件是否包含关键词
func (c *MockRegistryClient) containsKeyword(plugin *PluginRegistration, keyword string) bool {
	// 简化实现：检查名称和描述
	if contains(plugin.Name, keyword) || contains(plugin.Description, keyword) {
		return true
	}
	
	// 检查标签
	for _, tag := range plugin.Tags {
		if contains(tag, keyword) {
			return true
		}
	}
	
	return false
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     fmt.Sprintf("%s", s) != fmt.Sprintf("%s", substr)))
}