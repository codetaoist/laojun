package marketplace

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Marketplace 插件市场
type Marketplace struct {
	config     MarketplaceConfig
	storage    Storage
	registry   *PluginRegistry
	downloader *Downloader
	validator  *Validator
	indexer    *Indexer
	cache      *Cache
	mu         sync.RWMutex
}

// MarketplaceConfig 市场配置
type MarketplaceConfig struct {
	StoragePath     string            `json:"storage_path"`
	RegistryURL     string            `json:"registry_url"`
	CacheSize       int               `json:"cache_size"`
	CacheTTL        time.Duration     `json:"cache_ttl"`
	MaxFileSize     int64             `json:"max_file_size"`
	AllowedTypes    []string          `json:"allowed_types"`
	RequireAuth     bool              `json:"require_auth"`
	EnableSigning   bool              `json:"enable_signing"`
	TrustedSources  []string          `json:"trusted_sources"`
	ScanTimeout     time.Duration     `json:"scan_timeout"`
	DownloadTimeout time.Duration     `json:"download_timeout"`
	Mirrors         []string          `json:"mirrors"`
	Metadata        map[string]string `json:"metadata"`
}

// PluginRegistry 插件注册中心
type PluginRegistry struct {
	plugins    map[string]*PluginEntry
	categories map[string][]string
	tags       map[string][]string
	mu         sync.RWMutex
}

// PluginEntry 插件条目
type PluginEntry struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Author        string                 `json:"author"`
	Version       string                 `json:"version"`
	Type          string                 `json:"type"`
	Runtime       string                 `json:"runtime"`
	Category      string                 `json:"category"`
	Tags          []string               `json:"tags"`
	License       string                 `json:"license"`
	Homepage      string                 `json:"homepage"`
	Repository    string                 `json:"repository"`
	Documentation string                 `json:"documentation"`
	Icon          string                 `json:"icon"`
	Screenshots   []string               `json:"screenshots"`
	Dependencies  []Dependency           `json:"dependencies"`
	Versions      []VersionInfo          `json:"versions"`
	Downloads     int64                  `json:"downloads"`
	Rating        float64                `json:"rating"`
	Reviews       []Review               `json:"reviews"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	PublishedAt   time.Time              `json:"published_at"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	Version     string            `json:"version"`
	Description string            `json:"description"`
	DownloadURL string            `json:"download_url"`
	Checksum    string            `json:"checksum"`
	Size        int64             `json:"size"`
	PublishedAt time.Time         `json:"published_at"`
	Deprecated  bool              `json:"deprecated"`
	PreRelease  bool              `json:"pre_release"`
	Changes     []string          `json:"changes"`
	Metadata    map[string]string `json:"metadata"`
}

// Dependency 依赖信息
type Dependency struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

// Review 评论
type Review struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	Helpful   int       `json:"helpful"`
}

// SearchQuery 搜索查询
type SearchQuery struct {
	Query     string   `json:"query"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	Type      string   `json:"type"`
	Runtime   string   `json:"runtime"`
	Author    string   `json:"author"`
	License   string   `json:"license"`
	MinRating float64  `json:"min_rating"`
	SortBy    string   `json:"sort_by"`
	Order     string   `json:"order"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Plugins    []*PluginEntry `json:"plugins"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	Query      SearchQuery    `json:"query"`
}

// NewMarketplace 创建新的插件市场
func NewMarketplace(config MarketplaceConfig) (*Marketplace, error) {
	storage, err := NewStorage(config.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %v", err)
	}

	registry := NewPluginRegistry()
	downloader := NewDownloader(config.DownloadTimeout)
	validator := NewValidator(config.AllowedTypes)
	indexer := NewIndexer()
	cache := NewCache(config.CacheSize, config.CacheTTL)

	marketplace := &Marketplace{
		config:     config,
		storage:    storage,
		registry:   registry,
		downloader: downloader,
		validator:  validator,
		indexer:    indexer,
		cache:      cache,
	}

	// 加载现有插件
	if err := marketplace.loadPlugins(); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %v", err)
	}

	return marketplace, nil
}

// PublishPlugin 发布插件
func (m *Marketplace) PublishPlugin(pluginData []byte, manifest *PluginManifest, authorID string) (*PluginEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证插件
	if err := m.validator.ValidatePlugin(pluginData, manifest); err != nil {
		return nil, fmt.Errorf("plugin validation failed: %v", err)
	}

	// 检查插件是否已存在
	if existing := m.registry.GetPlugin(manifest.ID); existing != nil {
		// 检查版本是否已存在
		if m.versionExists(existing, manifest.Version) {
			return nil, errors.New("plugin version already exists")
		}
	}

	// 计算校验和
	checksum := m.calculateChecksum(pluginData)

	// 存储插件文件
	filename := fmt.Sprintf("%s-%s.tar.gz", manifest.ID, manifest.Version)
	if err := m.storage.Store(filename, pluginData); err != nil {
		return nil, fmt.Errorf("failed to store plugin: %v", err)
	}

	// 创建或更新插件条目
	entry := m.createPluginEntry(manifest, authorID, checksum, int64(len(pluginData)))

	// 注册插件
	if err := m.registry.RegisterPlugin(entry); err != nil {
		return nil, fmt.Errorf("failed to register plugin: %v", err)
	}

	// 更新索引
	m.indexer.IndexPlugin(entry)

	// 清除缓存
	m.cache.Clear()

	return entry, nil
}

// GetPlugin 获取插件信息
func (m *Marketplace) GetPlugin(pluginID string) (*PluginEntry, error) {
	// 尝试从缓存获取插件信息
	if cached := m.cache.Get(fmt.Sprintf("plugin:%s", pluginID)); cached != nil {
		if entry, ok := cached.(*PluginEntry); ok {
			return entry, nil
		}
	}

	m.mu.RLock()
	entry := m.registry.GetPlugin(pluginID)
	m.mu.RUnlock()

	if entry == nil {
		return nil, errors.New("plugin not found")
	}

	// 缓存结果
	m.cache.Set(fmt.Sprintf("plugin:%s", pluginID), entry)

	return entry, nil
}

// DownloadPlugin 下载插件
func (m *Marketplace) DownloadPlugin(pluginID, version string) ([]byte, error) {
	entry, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	// 查找指定版本
	var versionInfo *VersionInfo
	for _, v := range entry.Versions {
		if v.Version == version {
			versionInfo = &v
			break
		}
	}

	if versionInfo == nil {
		return nil, errors.New("version not found")
	}

	// 从存储获取插件文件
	filename := fmt.Sprintf("%s-%s.tar.gz", pluginID, version)
	data, err := m.storage.Retrieve(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve plugin: %v", err)
	}

	// 验证校验和
	if checksum := m.calculateChecksum(data); checksum != versionInfo.Checksum {
		return nil, errors.New("checksum mismatch")
	}

	// 更新下载计数
	m.mu.Lock()
	entry.Downloads++
	m.mu.Unlock()

	return data, nil
}

// SearchPlugins 搜索插件
func (m *Marketplace) SearchPlugins(query SearchQuery) (*SearchResult, error) {
	// 尝试从缓存获取搜索结果
	cacheKey := fmt.Sprintf("search:%s", m.generateSearchKey(query))
	if cached := m.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*SearchResult); ok {
			return result, nil
		}
	}

	m.mu.RLock()
	plugins := m.registry.SearchPlugins(query)
	m.mu.RUnlock()

	// 排序
	m.sortPlugins(plugins, query.SortBy, query.Order)

	// 分页
	total := len(plugins)
	start := query.Offset
	end := start + query.Limit
	if end > total {
		end = total
	}

	if start > total {
		start = total
	}

	pagedPlugins := plugins[start:end]

	result := &SearchResult{
		Plugins:    pagedPlugins,
		Total:      total,
		Page:       query.Offset/query.Limit + 1,
		PageSize:   query.Limit,
		TotalPages: (total + query.Limit - 1) / query.Limit,
		Query:      query,
	}

	// 缓存结果
	m.cache.Set(cacheKey, result)

	return result, nil
}

// InstallPlugin 安装插件
func (m *Marketplace) InstallPlugin(pluginID, version, targetPath string) error {
	// 下载插件
	data, err := m.DownloadPlugin(pluginID, version)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %v", err)
	}

	// 解压到目标路径
	if err := m.extractPlugin(data, targetPath); err != nil {
		return fmt.Errorf("failed to extract plugin: %v", err)
	}

	return nil
}

// UninstallPlugin 卸载插件
func (m *Marketplace) UninstallPlugin(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 从注册表移除
	if err := m.registry.UnregisterPlugin(pluginID); err != nil {
		return fmt.Errorf("failed to unregister plugin: %v", err)
	}

	// 从索引移除插件
	m.indexer.RemovePlugin(pluginID)

	// 清除缓存
	m.cache.Clear()

	return nil
}

// UpdatePlugin 更新插件
func (m *Marketplace) UpdatePlugin(pluginID string, pluginData []byte, manifest *PluginManifest, authorID string) error {
	// 检查插件是否存在
	existing, err := m.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("plugin not found: %v", err)
	}

	// 检查权限
	if existing.Author != authorID {
		return errors.New("insufficient permissions")
	}

	// 发布新版本
	_, err = m.PublishPlugin(pluginData, manifest, authorID)
	return err
}

// GetCategories 获取分类列表
func (m *Marketplace) GetCategories() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	categories := make(map[string]int)
	for _, entry := range m.registry.plugins {
		categories[entry.Category]++
	}

	return categories
}

// GetTags 获取标签列表
func (m *Marketplace) GetTags() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tags := make(map[string]int)
	for _, entry := range m.registry.plugins {
		for _, tag := range entry.Tags {
			tags[tag]++
		}
	}

	return tags
}

// GetStats 获取市场统计信息
func (m *Marketplace) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalPlugins := len(m.registry.plugins)
	totalDownloads := int64(0)
	avgRating := 0.0
	ratingCount := 0

	for _, entry := range m.registry.plugins {
		totalDownloads += entry.Downloads
		if entry.Rating > 0 {
			avgRating += entry.Rating
			ratingCount++
		}
	}

	if ratingCount > 0 {
		avgRating /= float64(ratingCount)
	}

	return map[string]interface{}{
		"total_plugins":   totalPlugins,
		"total_downloads": totalDownloads,
		"average_rating":  avgRating,
		"categories":      m.GetCategories(),
		"tags":            m.GetTags(),
		"cache_stats":     m.cache.GetStats(),
		"storage_stats":   m.storage.GetStats(),
	}
}

// AddReview 添加评论
func (m *Marketplace) AddReview(pluginID, userID, username string, rating int, comment string) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entry := m.registry.GetPlugin(pluginID)
	if entry == nil {
		return errors.New("plugin not found")
	}

	// 检查用户是否已经评论过
	for i, review := range entry.Reviews {
		if review.UserID == userID {
			// 更新现有评论
			entry.Reviews[i] = Review{
				ID:        review.ID,
				UserID:    userID,
				Username:  username,
				Rating:    rating,
				Comment:   comment,
				CreatedAt: time.Now(),
				Helpful:   review.Helpful,
			}
			m.updatePluginRating(entry)
			return nil
		}
	}

	// 添加新评论
	review := Review{
		ID:        m.generateReviewID(),
		UserID:    userID,
		Username:  username,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: time.Now(),
		Helpful:   0,
	}

	entry.Reviews = append(entry.Reviews, review)
	m.updatePluginRating(entry)

	return nil
}

// 辅助方法

func (m *Marketplace) loadPlugins() error {
	// 从存储加载插件信息
	// 这里应该从持久化存储（如数据库）加载
	return nil
}

func (m *Marketplace) versionExists(entry *PluginEntry, version string) bool {
	for _, v := range entry.Versions {
		if v.Version == version {
			return true
		}
	}
	return false
}

func (m *Marketplace) calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (m *Marketplace) createPluginEntry(manifest *PluginManifest, authorID, checksum string, size int64) *PluginEntry {
	now := time.Now()

	versionInfo := VersionInfo{
		Version:     manifest.Version,
		Description: manifest.Description,
		DownloadURL: fmt.Sprintf("/api/v1/plugins/%s/versions/%s/download", manifest.ID, manifest.Version),
		Checksum:    checksum,
		Size:        size,
		PublishedAt: now,
		Deprecated:  false,
		PreRelease:  strings.Contains(manifest.Version, "-"),
		Changes:     manifest.Changes,
		Metadata:    manifest.Metadata,
	}

	// 检查是否是新插件
	if existing := m.registry.GetPlugin(manifest.ID); existing != nil {
		// 更新现有插件
		existing.Version = manifest.Version
		existing.Description = manifest.Description
		existing.UpdatedAt = now
		existing.Versions = append(existing.Versions, versionInfo)
		return existing
	}

	// 创建新插件
	return &PluginEntry{
		ID:            manifest.ID,
		Name:          manifest.Name,
		Description:   manifest.Description,
		Author:        authorID,
		Version:       manifest.Version,
		Type:          manifest.Type,
		Runtime:       manifest.Runtime,
		Category:      manifest.Category,
		Tags:          manifest.Tags,
		License:       manifest.License,
		Homepage:      manifest.Homepage,
		Repository:    manifest.Repository,
		Documentation: manifest.Documentation,
		Icon:          manifest.Icon,
		Screenshots:   manifest.Screenshots,
		Dependencies:  manifest.Dependencies,
		Versions:      []VersionInfo{versionInfo},
		Downloads:     0,
		Rating:        0,
		Reviews:       []Review{},
		CreatedAt:     now,
		UpdatedAt:     now,
		PublishedAt:   now,
		Status:        "published",
		Metadata:      manifest.Metadata,
	}
}

func (m *Marketplace) sortPlugins(plugins []*PluginEntry, sortBy, order string) {
	sort.Slice(plugins, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = plugins[i].Name < plugins[j].Name
		case "downloads":
			less = plugins[i].Downloads < plugins[j].Downloads
		case "rating":
			less = plugins[i].Rating < plugins[j].Rating
		case "created":
			less = plugins[i].CreatedAt.Before(plugins[j].CreatedAt)
		case "updated":
			less = plugins[i].UpdatedAt.Before(plugins[j].UpdatedAt)
		default:
			less = plugins[i].Name < plugins[j].Name
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

func (m *Marketplace) generateSearchKey(query SearchQuery) string {
	data, _ := json.Marshal(query)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}

func (m *Marketplace) extractPlugin(data []byte, targetPath string) error {
	// 创建目标目录
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return err
	}

	// 解压tar.gz文件
	reader := strings.NewReader(string(data))
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(targetPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()
		}
	}

	return nil
}

func (m *Marketplace) updatePluginRating(entry *PluginEntry) {
	if len(entry.Reviews) == 0 {
		entry.Rating = 0
		return
	}

	total := 0
	for _, review := range entry.Reviews {
		total += review.Rating
	}

	entry.Rating = float64(total) / float64(len(entry.Reviews))
}

func (m *Marketplace) generateReviewID() string {
	return fmt.Sprintf("review_%d", time.Now().UnixNano())
}

// ServeHTTP 提供HTTP API
func (m *Marketplace) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/api/v1/plugins/"):
		m.handleGetPlugin(w, r)
	case r.Method == "POST" && r.URL.Path == "/api/v1/plugins":
		m.handlePublishPlugin(w, r)
	case r.Method == "GET" && r.URL.Path == "/api/v1/plugins":
		m.handleSearchPlugins(w, r)
	case r.Method == "GET" && r.URL.Path == "/api/v1/stats":
		m.handleGetStats(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (m *Marketplace) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	// 实现获取插件的HTTP处理
	w.Header().Set("Content-Type", "application/json")
	// ... 具体实现
}

func (m *Marketplace) handlePublishPlugin(w http.ResponseWriter, r *http.Request) {
	// 实现发布插件的HTTP处理
	w.Header().Set("Content-Type", "application/json")
	// ... 具体实现
}

func (m *Marketplace) handleSearchPlugins(w http.ResponseWriter, r *http.Request) {
	// 实现搜索插件的HTTP处理
	w.Header().Set("Content-Type", "application/json")
	// ... 具体实现
}

func (m *Marketplace) handleGetStats(w http.ResponseWriter, r *http.Request) {
	// 实现获取统计信息的HTTP处理
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.GetStats())
}
