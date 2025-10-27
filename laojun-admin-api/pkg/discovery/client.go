package discovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client 服务发现客户端
type Client struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
	TTL     int               `json:"ttl"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name    string            `json:"name"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
	TTL     int               `json:"ttl"`
}

// NewClient 创建服务发现客户端
func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Register 注册服务
func (c *Client) Register(req *RegisterRequest) (*ServiceInstance, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/services", c.baseURL)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Message string           `json:"message"`
		Service *ServiceInstance `json:"service"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("Service registered successfully",
		zap.String("service", req.Name),
		zap.String("id", result.Service.ID),
		zap.String("address", fmt.Sprintf("%s:%d", req.Address, req.Port)))

	return result.Service, nil
}

// Deregister 注销服务
func (c *Client) Deregister(serviceID string) error {
	url := fmt.Sprintf("%s/api/v1/services/%s", c.baseURL, serviceID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deregistration failed with status: %d", resp.StatusCode)
	}

	c.logger.Info("Service deregistered successfully", zap.String("id", serviceID))
	return nil
}

// UpdateHealth 更新健康状态
func (c *Client) UpdateHealth(serviceID string, status string) error {
	data, err := json.Marshal(map[string]string{"status": status})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/services/%s/health", c.baseURL, serviceID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health update failed with status: %d", resp.StatusCode)
	}

	return nil
}