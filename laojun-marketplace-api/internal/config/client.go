package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"github.com/codetaoist/laojun-shared/config"
)

// Client 配置中心客户端
type Client struct {
	configCenterURL string
	httpClient      *http.Client
	logger          *zap.Logger
	serviceName     string
	environment     string
}

// NewClient 创建配置中心客户端
func NewClient(configCenterURL, serviceName, environment string, logger *zap.Logger) *Client {
	return &Client{
		configCenterURL: configCenterURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:      logger,
		serviceName: serviceName,
		environment: environment,
	}
}

// LoadConfig 从配置中心加载配置
func (c *Client) LoadConfig(ctx context.Context) (*config.Config, error) {
	// 构建配置文件名
	configFile := fmt.Sprintf("%s.%s.yaml", c.serviceName, c.environment)
	
	// 从配置中心获取配置
	url := fmt.Sprintf("%s/api/v1/configs/%s", c.configCenterURL, configFile)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 如果配置中心不可用，回退到本地配置
		c.logger.Warn("Config center unavailable, falling back to local config",
			zap.Int("status_code", resp.StatusCode),
			zap.String("config_file", configFile))
		return c.loadLocalConfig()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// 解析配置
	var cfg config.Config
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	c.logger.Info("Config loaded from config center",
		zap.String("config_file", configFile),
		zap.String("service", c.serviceName))

	return &cfg, nil
}

// loadLocalConfig 加载本地配置作为回退
func (c *Client) loadLocalConfig() (*config.Config, error) {
	// 尝试加载本地配置文件
	configFiles := []string{
		fmt.Sprintf("./configs/%s.%s.yaml", c.serviceName, c.environment),  // 相对于服务目录
		fmt.Sprintf("./configs/%s.yaml", c.serviceName),                    // 相对于服务目录
		"./configs/config.yaml",                                            // 相对于服务目录
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			c.logger.Info("Loading local config file", zap.String("file", configFile))
			return config.LoadConfig()
		}
	}

	return nil, fmt.Errorf("no local config file found")
}

// WatchConfig 监听配置变化
func (c *Client) WatchConfig(ctx context.Context, callback func(*config.Config)) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cfg, err := c.LoadConfig(ctx)
			if err != nil {
				c.logger.Error("Failed to reload config", zap.Error(err))
				continue
			}
			
			if callback != nil {
				callback(cfg)
			}
		}
	}
}

// GetConfigVersion 获取配置版本信息
func (c *Client) GetConfigVersion(ctx context.Context) (string, error) {
	configFile := fmt.Sprintf("%s.%s.yaml", c.serviceName, c.environment)
	url := fmt.Sprintf("%s/api/v1/configs/%s/version", c.configCenterURL, configFile)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get version failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var result struct {
		Version string `json:"version"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal version: %w", err)
	}

	return result.Version, nil
}