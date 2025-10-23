package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/codetaoist/laojun/internal/models"
	"github.com/google/uuid"
)

// MarketplaceClient 插件市场API客户端
type MarketplaceClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewMarketplaceClient 创建插件市场API客户端
func NewMarketplaceClient(baseURL, apiKey string) *MarketplaceClient {
	return &MarketplaceClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UpdatePluginStatus 更新插件状态
func (c *MarketplaceClient) UpdatePluginStatus(pluginID uuid.UUID, status string) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/status", c.baseURL, pluginID)

	payload := map[string]interface{}{
		"status": status,
	}

	return c.makeRequest("PUT", url, payload, nil)
}

// GetPlugin 获取插件信息
func (c *MarketplaceClient) GetPlugin(pluginID uuid.UUID) (*models.Plugin, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/%s", c.baseURL, pluginID)

	var plugin models.Plugin
	if err := c.makeRequest("GET", url, nil, &plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

// SyncPlugin 同步插件到市场
func (c *MarketplaceClient) SyncPlugin(plugin *models.Plugin) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/sync", c.baseURL, plugin.ID)

	return c.makeRequest("POST", url, plugin, nil)
}

// GetPluginList 获取插件列表
func (c *MarketplaceClient) GetPluginList(params models.MarketplaceSearchParams) (*models.MarketplacePluginListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/plugins", c.baseURL)

	// 构建查询参数
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if params.Search != "" {
		q.Add("search", params.Search)
	}
	if params.Category != "" {
		q.Add("category", params.Category)
	}
	if params.Tag != "" {
		q.Add("tag", params.Tag)
	}
	if params.SortBy != "" {
		q.Add("sort_by", params.SortBy)
	}
	if params.SortOrder != "" {
		q.Add("sort_order", params.SortOrder)
	}
	q.Add("page", fmt.Sprintf("%d", params.Page))
	q.Add("limit", fmt.Sprintf("%d", params.Limit))
	req.URL.RawQuery = q.Encode()

	var response models.MarketplacePluginListResponse
	if err := c.makeRequestWithReq(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// InstallPlugin 安装插件
func (c *MarketplaceClient) InstallPlugin(pluginID uuid.UUID, version string) (*models.PluginInstallResponse, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/install", c.baseURL, pluginID)

	payload := map[string]interface{}{
		"version": version,
	}

	var response models.PluginInstallResponse
	if err := c.makeRequest("POST", url, payload, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// UninstallPlugin 卸载插件
func (c *MarketplaceClient) UninstallPlugin(pluginID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/uninstall", c.baseURL, pluginID)

	return c.makeRequest("DELETE", url, nil, nil)
}

// GetCategories 获取分类列表
func (c *MarketplaceClient) GetCategories() ([]models.Category, error) {
	url := fmt.Sprintf("%s/api/v1/categories", c.baseURL)

	var categories []models.Category
	if err := c.makeRequest("GET", url, nil, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// GetTags 获取标签列表
func (c *MarketplaceClient) GetTags() ([]models.Tag, error) {
	url := fmt.Sprintf("%s/api/v1/tags", c.baseURL)

	var tags []models.Tag
	if err := c.makeRequest("GET", url, nil, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// DownloadPlugin 下载插件
func (c *MarketplaceClient) DownloadPlugin(pluginID uuid.UUID, version string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/download", c.baseURL, pluginID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if version != "" {
		q := req.URL.Query()
		q.Add("version", version)
		req.URL.RawQuery = q.Encode()
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// PublishPlugin 发布插件
func (c *MarketplaceClient) PublishPlugin(plugin *models.Plugin) error {
	url := fmt.Sprintf("%s/api/v1/plugins", c.baseURL)

	return c.makeRequest("POST", url, plugin, nil)
}

// UpdatePlugin 更新插件
func (c *MarketplaceClient) UpdatePlugin(plugin *models.Plugin) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s", c.baseURL, plugin.ID)

	return c.makeRequest("PUT", url, plugin, nil)
}

// DeletePlugin 删除插件
func (c *MarketplaceClient) DeletePlugin(pluginID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s", c.baseURL, pluginID)

	return c.makeRequest("DELETE", url, nil, nil)
}

// GetPluginVersions 获取插件版本列表
func (c *MarketplaceClient) GetPluginVersions(pluginID uuid.UUID) ([]models.PluginVersion, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/versions", c.baseURL, pluginID)

	var versions []models.PluginVersion
	if err := c.makeRequest("GET", url, nil, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// GetPluginReviews 获取插件评论
func (c *MarketplaceClient) GetPluginReviews(pluginID uuid.UUID, page, limit int) (*models.PluginReviewListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/reviews", c.baseURL, pluginID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("page", fmt.Sprintf("%d", page))
	q.Add("limit", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()

	var response models.PluginReviewListResponse
	if err := c.makeRequestWithReq(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// makeRequest 发起HTTP请求
func (c *MarketplaceClient) makeRequest(method, url string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.makeRequestWithReq(req, result)
}

// makeRequestWithReq 使用现有请求发起HTTP请求
func (c *MarketplaceClient) makeRequestWithReq(req *http.Request, result interface{}) error {
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// setHeaders 设置请求头
func (c *MarketplaceClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TaiShangLaoJun-Admin/1.0")

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// HealthCheck 健康检查
func (c *MarketplaceClient) HealthCheck() error {
	url := fmt.Sprintf("%s/api/v1/health", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}