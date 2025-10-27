package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Client 配置中心客户端 (已弃用，请使用 HTTPConfigClient)
// Deprecated: Use HTTPConfigClient instead
type Client struct {
	baseURL     string
	httpClient  *http.Client
	cache       map[string]interface{}
	cacheMutex  sync.RWMutex
	logger      *logrus.Logger
	service     string
	environment string
}

// ClientConfig 客户端配置 (已弃用，请使用 ConfigOptions)
// Deprecated: Use ConfigOptions instead
type ClientConfig struct {
	BaseURL     string
	Service     string
	Environment string
	Timeout     time.Duration
	APIKey      string
}

// NewClient 创建配置客户端 (已弃用，请使用 NewHTTPConfigClient)
// Deprecated: Use NewHTTPConfigClient instead
func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		cache:       make(map[string]interface{}),
		logger:      logrus.New(),
		service:     config.Service,
		environment: config.Environment,
	}
}

// HTTPConfigClient HTTP配置中心客户端
type HTTPConfigClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	options    *ConfigOptions
	
	// 认证信息
	token    string
	username string
	password string
	
	// 连接池
	mu sync.RWMutex
}

// NewHTTPConfigClient 创建HTTP配置中心客户端
func NewHTTPConfigClient(baseURL string, options *ConfigOptions) *HTTPConfigClient {
	timeout := 30 * time.Second
	if options != nil && options.Timeout > 0 {
		timeout = options.Timeout
	}
	
	return &HTTPConfigClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:  logrus.New(),
		options: options,
	}
}

// SetAuth 设置认证信息
func (c *HTTPConfigClient) SetAuth(token, username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.token = token
	c.username = username
	c.password = password
}

// Get 获取配置
func (c *Client) Get(ctx context.Context, key string) (interface{}, error) {
	// 先检查缓存
	c.cacheMutex.RLock()
	if value, exists := c.cache[key]; exists {
		c.cacheMutex.RUnlock()
		return value, nil
	}
	c.cacheMutex.RUnlock()

	// 从配置中心获取
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s/%s", c.baseURL, c.service, c.environment, key)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("config not found: %s", key)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Data struct {
			Value interface{} `json:"value"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 更新缓存
	c.cacheMutex.Lock()
	c.cache[key] = response.Data.Value
	c.cacheMutex.Unlock()

	return response.Data.Value, nil
}

// GetString 获取字符串配置
func (c *Client) GetString(ctx context.Context, key string, defaultValue string) string {
	value, err := c.Get(ctx, key)
	if err != nil {
		c.logger.Warnf("Failed to get config %s: %v, using default: %s", key, err, defaultValue)
		return defaultValue
	}

	if str, ok := value.(string); ok {
		return str
	}

	c.logger.Warnf("Config %s is not a string, using default: %s", key, defaultValue)
	return defaultValue
}

// GetInt 获取整数配置
func (c *Client) GetInt(ctx context.Context, key string, defaultValue int) int {
	value, err := c.Get(ctx, key)
	if err != nil {
		c.logger.Warnf("Failed to get config %s: %v, using default: %d", key, err, defaultValue)
		return defaultValue
	}

	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		// 尝试解析字符串为整数
		if i, err := fmt.Sscanf(v, "%d", &defaultValue); err == nil && i == 1 {
			return defaultValue
		}
	}

	c.logger.Warnf("Config %s is not an integer, using default: %d", key, defaultValue)
	return defaultValue
}

// GetBool 获取布尔配置
func (c *Client) GetBool(ctx context.Context, key string, defaultValue bool) bool {
	value, err := c.Get(ctx, key)
	if err != nil {
		c.logger.Warnf("Failed to get config %s: %v, using default: %t", key, err, defaultValue)
		return defaultValue
	}

	if b, ok := value.(bool); ok {
		return b
	}

	c.logger.Warnf("Config %s is not a boolean, using default: %t", key, defaultValue)
	return defaultValue
}

// GetDuration 获取时间间隔配置
func (c *Client) GetDuration(ctx context.Context, key string, defaultValue time.Duration) time.Duration {
	value, err := c.Get(ctx, key)
	if err != nil {
		c.logger.Warnf("Failed to get config %s: %v, using default: %s", key, err, defaultValue)
		return defaultValue
	}

	if str, ok := value.(string); ok {
		if duration, err := time.ParseDuration(str); err == nil {
			return duration
		}
	}

	c.logger.Warnf("Config %s is not a valid duration, using default: %s", key, defaultValue)
	return defaultValue
}

// Set 设置配置
func (c *Client) Set(ctx context.Context, key string, value interface{}, valueType string) error {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s/%s", c.baseURL, c.service, c.environment, key)

	payload := map[string]interface{}{
		"value": value,
		"type":  valueType,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 更新缓存
	c.cacheMutex.Lock()
	c.cache[key] = value
	c.cacheMutex.Unlock()

	return nil
}

// HTTPConfigClient 方法实现

// Get 获取配置
func (c *HTTPConfigClient) Get(ctx context.Context, key string) (*ConfigItem, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s", c.baseURL, url.PathEscape(key))
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrConfigNotFound
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	var item ConfigItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &item, nil
}

// Set 设置配置
func (c *HTTPConfigClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	url := fmt.Sprintf("%s/api/v1/configs/%s", c.baseURL, url.PathEscape(key))
	
	payload := map[string]interface{}{
		"value": value,
	}
	if ttl > 0 {
		payload["ttl"] = int64(ttl.Seconds())
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Delete 删除配置
func (c *HTTPConfigClient) Delete(ctx context.Context, key string) error {
	url := fmt.Sprintf("%s/api/v1/configs/%s", c.baseURL, url.PathEscape(key))
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return ErrConfigNotFound
	}
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Exists 检查配置是否存在
func (c *HTTPConfigClient) Exists(ctx context.Context, key string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s/exists", c.baseURL, url.PathEscape(key))
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK, nil
}

// List 列出配置键
func (c *HTTPConfigClient) List(ctx context.Context, prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/configs", c.baseURL)
	if prefix != "" {
		url += "?prefix=" + url.QueryEscape(prefix)
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Keys []string `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Keys, nil
}

// GetMultiple 批量获取配置
func (c *HTTPConfigClient) GetMultiple(ctx context.Context, keys []string) (map[string]*ConfigItem, error) {
	url := fmt.Sprintf("%s/api/v1/configs/batch", c.baseURL)
	
	payload := map[string]interface{}{
		"keys": keys,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	var result map[string]*ConfigItem
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result, nil
}

// SetMultiple 批量设置配置
func (c *HTTPConfigClient) SetMultiple(ctx context.Context, configs map[string]string, ttl time.Duration) error {
	url := fmt.Sprintf("%s/api/v1/configs/batch", c.baseURL)
	
	payload := map[string]interface{}{
		"configs": configs,
	}
	if ttl > 0 {
		payload["ttl"] = int64(ttl.Seconds())
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GetHistory 获取配置历史
func (c *HTTPConfigClient) GetHistory(ctx context.Context, key string, limit int) ([]*ConfigHistory, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s/history", c.baseURL, url.PathEscape(key))
	if limit > 0 {
		url += "?limit=" + strconv.Itoa(limit)
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	var history []*ConfigHistory
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return history, nil
}

// GetVersion 获取指定版本的配置
func (c *HTTPConfigClient) GetVersion(ctx context.Context, key string, version int64) (*ConfigItem, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s/versions/%d", c.baseURL, url.PathEscape(key), version)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrConfigNotFound
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	var item ConfigItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &item, nil
}

// Health 获取健康状态
func (c *HTTPConfigClient) Health(ctx context.Context) (*ConfigHealth, error) {
	url := fmt.Sprintf("%s/api/v1/health", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	var health ConfigHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &health, nil
}

// Close 关闭客户端
func (c *HTTPConfigClient) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}

// setAuthHeaders 设置认证头
func (c *HTTPConfigClient) setAuthHeaders(req *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
}

// Delete 删除配置
func (c *Client) Delete(ctx context.Context, key string) error {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s/%s", c.baseURL, c.service, c.environment, key)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 从缓存中删除
	c.cacheMutex.Lock()
	delete(c.cache, key)
	c.cacheMutex.Unlock()

	return nil
}

// List 列出所有配置
func (c *Client) List(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s", c.baseURL, c.service, c.environment)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Data []struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result := make(map[string]interface{})
	for _, item := range response.Data {
		result[item.Key] = item.Value
	}

	// 更新缓存
	c.cacheMutex.Lock()
	for key, value := range result {
		c.cache[key] = value
	}
	c.cacheMutex.Unlock()

	return result, nil
}

// RefreshCache 刷新缓存
func (c *Client) RefreshCache(ctx context.Context) error {
	configs, err := c.List(ctx)
	if err != nil {
		return err
	}

	c.cacheMutex.Lock()
	c.cache = configs
	c.cacheMutex.Unlock()

	c.logger.Infof("Cache refreshed with %d configs", len(configs))
	return nil
}

// ClearCache 清空缓存
func (c *Client) ClearCache() {
	c.cacheMutex.Lock()
	c.cache = make(map[string]interface{})
	c.cacheMutex.Unlock()
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config center unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
