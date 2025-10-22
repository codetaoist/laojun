package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Client 配置中心客户
type Client struct {
	baseURL     string
	httpClient  *http.Client
	cache       map[string]interface{}
	cacheMutex  sync.RWMutex
	logger      *logrus.Logger
	service     string
	environment string
}

// ClientConfig 客户端配置
type ClientConfig struct {
	BaseURL     string
	Service     string
	Environment string
	Timeout     time.Duration
	APIKey      string
}

// NewClient 创建配置客户
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
