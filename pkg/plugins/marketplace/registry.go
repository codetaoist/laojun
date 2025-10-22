package marketplace

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// NewPluginRegistry 创建新的插件注册中心
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:    make(map[string]*PluginEntry),
		categories: make(map[string][]string),
		tags:       make(map[string][]string),
	}
}

// RegisterPlugin 注册插件
func (pr *PluginRegistry) RegisterPlugin(entry *PluginEntry) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if entry.ID == "" {
		return errors.New("plugin ID cannot be empty")
	}

	// 检查插件是否已存在
	if existing, exists := pr.plugins[entry.ID]; exists {
		// 更新现有插件
		existing.Name = entry.Name
		existing.Description = entry.Description
		existing.Version = entry.Version
		existing.UpdatedAt = time.Now()

		// 合并版本信息
		versionExists := false
		for i, v := range existing.Versions {
			if v.Version == entry.Version {
				existing.Versions[i] = entry.Versions[0]
				versionExists = true
				break
			}
		}

		if !versionExists {
			existing.Versions = append(existing.Versions, entry.Versions...)
		}

		// 更新分类和标签索引
		pr.updateCategoryIndex(entry.ID, entry.Category)
		pr.updateTagIndex(entry.ID, entry.Tags)

		return nil
	}

	// 注册新插件
	pr.plugins[entry.ID] = entry

	// 更新索引
	pr.updateCategoryIndex(entry.ID, entry.Category)
	pr.updateTagIndex(entry.ID, entry.Tags)

	return nil
}

// UnregisterPlugin 注销插件
func (pr *PluginRegistry) UnregisterPlugin(pluginID string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return errors.New("plugin not found")
	}

	// 从索引中移除
	pr.removeCategoryIndex(pluginID, entry.Category)
	pr.removeTagIndex(pluginID, entry.Tags)

	// 删除插件
	delete(pr.plugins, pluginID)

	return nil
}

// GetPlugin 获取插件
func (pr *PluginRegistry) GetPlugin(pluginID string) *PluginEntry {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return pr.plugins[pluginID]
}

// ListPlugins 列出所有插件
func (pr *PluginRegistry) ListPlugins() []*PluginEntry {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := make([]*PluginEntry, 0, len(pr.plugins))
	for _, entry := range pr.plugins {
		plugins = append(plugins, entry)
	}

	return plugins
}

// SearchPlugins 搜索插件
func (pr *PluginRegistry) SearchPlugins(query SearchQuery) []*PluginEntry {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var results []*PluginEntry

	for _, entry := range pr.plugins {
		if pr.matchesQuery(entry, query) {
			results = append(results, entry)
		}
	}

	return results
}

// GetPluginsByCategory 按分类获取插件
func (pr *PluginRegistry) GetPluginsByCategory(category string) []*PluginEntry {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	pluginIDs, exists := pr.categories[category]
	if !exists {
		return []*PluginEntry{}
	}

	plugins := make([]*PluginEntry, 0, len(pluginIDs))
	for _, id := range pluginIDs {
		if entry, exists := pr.plugins[id]; exists {
			plugins = append(plugins, entry)
		}
	}

	return plugins
}

// GetPluginsByTag 按标签获取插件
func (pr *PluginRegistry) GetPluginsByTag(tag string) []*PluginEntry {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	pluginIDs, exists := pr.tags[tag]
	if !exists {
		return []*PluginEntry{}
	}

	plugins := make([]*PluginEntry, 0, len(pluginIDs))
	for _, id := range pluginIDs {
		if entry, exists := pr.plugins[id]; exists {
			plugins = append(plugins, entry)
		}
	}

	return plugins
}

// GetCategories 获取所有分类
func (pr *PluginRegistry) GetCategories() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	categories := make([]string, 0, len(pr.categories))
	for category := range pr.categories {
		categories = append(categories, category)
	}

	sort.Strings(categories)
	return categories
}

// GetTags 获取所有标签
func (pr *PluginRegistry) GetTags() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	tags := make([]string, 0, len(pr.tags))
	for tag := range pr.tags {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags
}

// GetStats 获取注册表统计信息
func (pr *PluginRegistry) GetStats() map[string]interface{} {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	totalDownloads := int64(0)
	totalReviews := 0
	avgRating := 0.0
	ratingCount := 0

	typeCount := make(map[string]int)
	runtimeCount := make(map[string]int)
	statusCount := make(map[string]int)

	for _, entry := range pr.plugins {
		totalDownloads += entry.Downloads
		totalReviews += len(entry.Reviews)

		if entry.Rating > 0 {
			avgRating += entry.Rating
			ratingCount++
		}

		typeCount[entry.Type]++
		runtimeCount[entry.Runtime]++
		statusCount[entry.Status]++
	}

	if ratingCount > 0 {
		avgRating /= float64(ratingCount)
	}

	return map[string]interface{}{
		"total_plugins":        len(pr.plugins),
		"total_downloads":      totalDownloads,
		"total_reviews":        totalReviews,
		"average_rating":       avgRating,
		"categories":           len(pr.categories),
		"tags":                 len(pr.tags),
		"type_distribution":    typeCount,
		"runtime_distribution": runtimeCount,
		"status_distribution":  statusCount,
	}
}

// UpdatePluginStatus 更新插件状态
func (pr *PluginRegistry) UpdatePluginStatus(pluginID, status string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return errors.New("plugin not found")
	}

	entry.Status = status
	entry.UpdatedAt = time.Now()

	return nil
}

// AddPluginVersion 添加插件版本
func (pr *PluginRegistry) AddPluginVersion(pluginID string, version VersionInfo) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return errors.New("plugin not found")
	}

	// 检查版本是否已存在
	for i, v := range entry.Versions {
		if v.Version == version.Version {
			entry.Versions[i] = version
			entry.UpdatedAt = time.Now()
			return nil
		}
	}

	// 添加新版本
	entry.Versions = append(entry.Versions, version)
	entry.Version = version.Version // 更新当前版本
	entry.UpdatedAt = time.Now()

	// 按版本号排序
	sort.Slice(entry.Versions, func(i, j int) bool {
		return pr.compareVersions(entry.Versions[i].Version, entry.Versions[j].Version) > 0
	})

	return nil
}

// RemovePluginVersion 移除插件版本
func (pr *PluginRegistry) RemovePluginVersion(pluginID, version string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return errors.New("plugin not found")
	}

	// 查找并移除版本
	for i, v := range entry.Versions {
		if v.Version == version {
			entry.Versions = append(entry.Versions[:i], entry.Versions[i+1:]...)
			entry.UpdatedAt = time.Now()

			// 如果移除的是当前版本，更新为最新版本
			if entry.Version == version && len(entry.Versions) > 0 {
				entry.Version = entry.Versions[0].Version
			}

			return nil
		}
	}

	return errors.New("version not found")
}

// GetPluginVersion 获取特定版本信息
func (pr *PluginRegistry) GetPluginVersion(pluginID, version string) (*VersionInfo, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return nil, errors.New("plugin not found")
	}

	for _, v := range entry.Versions {
		if v.Version == version {
			return &v, nil
		}
	}

	return nil, errors.New("version not found")
}

// GetLatestVersion 获取最新版本信息
func (pr *PluginRegistry) GetLatestVersion(pluginID string) (*VersionInfo, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return nil, errors.New("plugin not found")
	}

	if len(entry.Versions) == 0 {
		return nil, errors.New("no versions available")
	}

	// 返回第一个版本（应该是最新的，因为已排序）
	return &entry.Versions[0], nil
}

// ExportRegistry 导出注册信息
func (pr *PluginRegistry) ExportRegistry() ([]byte, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	data := struct {
		Plugins    map[string]*PluginEntry `json:"plugins"`
		Categories map[string][]string     `json:"categories"`
		Tags       map[string][]string     `json:"tags"`
		ExportedAt time.Time               `json:"exported_at"`
	}{
		Plugins:    pr.plugins,
		Categories: pr.categories,
		Tags:       pr.tags,
		ExportedAt: time.Now(),
	}

	return json.Marshal(data)
}

// ImportRegistry 导入注册信息
func (pr *PluginRegistry) ImportRegistry(data []byte) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	var importData struct {
		Plugins    map[string]*PluginEntry `json:"plugins"`
		Categories map[string][]string     `json:"categories"`
		Tags       map[string][]string     `json:"tags"`
		ExportedAt time.Time               `json:"exported_at"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %v", err)
	}

	// 合并插件
	for id, entry := range importData.Plugins {
		pr.plugins[id] = entry
	}

	// 合并分类
	for category, pluginIDs := range importData.Categories {
		if existing, exists := pr.categories[category]; exists {
			pr.categories[category] = pr.mergeStringSlices(existing, pluginIDs)
		} else {
			pr.categories[category] = pluginIDs
		}
	}

	// 合并标签
	for tag, pluginIDs := range importData.Tags {
		if existing, exists := pr.tags[tag]; exists {
			pr.tags[tag] = pr.mergeStringSlices(existing, pluginIDs)
		} else {
			pr.tags[tag] = pluginIDs
		}
	}

	return nil
}

// 私有方法

// matchesQuery 检查插件是否匹配查询条件
func (pr *PluginRegistry) matchesQuery(entry *PluginEntry, query SearchQuery) bool {
	// 文本搜索
	if query.Query != "" {
		searchText := strings.ToLower(query.Query)
		if !strings.Contains(strings.ToLower(entry.Name), searchText) &&
			!strings.Contains(strings.ToLower(entry.Description), searchText) &&
			!strings.Contains(strings.ToLower(entry.Author), searchText) {
			return false
		}
	}

	// 分类过滤
	if query.Category != "" && entry.Category != query.Category {
		return false
	}

	// 类型过滤
	if query.Type != "" && entry.Type != query.Type {
		return false
	}

	// 运行时环境过滤
	if query.Runtime != "" && entry.Runtime != query.Runtime {
		return false
	}

	// 作者过滤
	if query.Author != "" && entry.Author != query.Author {
		return false
	}

	// 许可证过滤
	if query.License != "" && entry.License != query.License {
		return false
	}

	// 评分过滤
	if query.MinRating > 0 && entry.Rating < query.MinRating {
		return false
	}

	// 标签过滤
	if len(query.Tags) > 0 {
		for _, queryTag := range query.Tags {
			found := false
			for _, entryTag := range entry.Tags {
				if entryTag == queryTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// updateCategoryIndex 更新分类索引
func (pr *PluginRegistry) updateCategoryIndex(pluginID, category string) {
	if category == "" {
		return
	}

	if pluginIDs, exists := pr.categories[category]; exists {
		// 检查是否已存在
		for _, id := range pluginIDs {
			if id == pluginID {
				return
			}
		}
		pr.categories[category] = append(pluginIDs, pluginID)
	} else {
		pr.categories[category] = []string{pluginID}
	}
}

// updateTagIndex 更新标签索引
func (pr *PluginRegistry) updateTagIndex(pluginID string, tags []string) {
	for _, tag := range tags {
		if tag == "" {
			continue
		}

		if pluginIDs, exists := pr.tags[tag]; exists {
			// 检查是否已存在
			found := false
			for _, id := range pluginIDs {
				if id == pluginID {
					found = true
					break
				}
			}
			if !found {
				pr.tags[tag] = append(pluginIDs, pluginID)
			}
		} else {
			pr.tags[tag] = []string{pluginID}
		}
	}
}

// removeCategoryIndex 移除分类索引
func (pr *PluginRegistry) removeCategoryIndex(pluginID, category string) {
	if category == "" {
		return
	}

	if pluginIDs, exists := pr.categories[category]; exists {
		for i, id := range pluginIDs {
			if id == pluginID {
				pr.categories[category] = append(pluginIDs[:i], pluginIDs[i+1:]...)
				break
			}
		}

		// 如果分类为空，删除分类
		if len(pr.categories[category]) == 0 {
			delete(pr.categories, category)
		}
	}
}

// removeTagIndex 移除标签索引
func (pr *PluginRegistry) removeTagIndex(pluginID string, tags []string) {
	for _, tag := range tags {
		if tag == "" {
			continue
		}

		if pluginIDs, exists := pr.tags[tag]; exists {
			for i, id := range pluginIDs {
				if id == pluginID {
					pr.tags[tag] = append(pluginIDs[:i], pluginIDs[i+1:]...)
					break
				}
			}

			// 如果标签为空，删除标签
			if len(pr.tags[tag]) == 0 {
				delete(pr.tags, tag)
			}
		}
	}
}

// compareVersions 比较版本号
func (pr *PluginRegistry) compareVersions(v1, v2 string) int {
	// 简单的版本比较实现
	// 实际应用中应该使用更复杂的语义版本比较库
	if v1 == v2 {
		return 0
	}
	if v1 > v2 {
		return 1
	}
	return -1
}

// mergeStringSlices 合并字符串切片，去重
func (pr *PluginRegistry) mergeStringSlices(slice1, slice2 []string) []string {
	merged := make([]string, len(slice1))
	copy(merged, slice1)

	for _, item := range slice2 {
		found := false
		for _, existing := range merged {
			if existing == item {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, item)
		}
	}

	return merged
}

// ValidatePluginEntry 验证插件条目
func (pr *PluginRegistry) ValidatePluginEntry(entry *PluginEntry) error {
	if entry.ID == "" {
		return errors.New("plugin ID is required")
	}

	if entry.Name == "" {
		return errors.New("plugin name is required")
	}

	if entry.Version == "" {
		return errors.New("plugin version is required")
	}

	if entry.Author == "" {
		return errors.New("plugin author is required")
	}

	if entry.Type == "" {
		return errors.New("plugin type is required")
	}

	if entry.Runtime == "" {
		return errors.New("plugin runtime is required")
	}

	// 验证版本信息
	if len(entry.Versions) == 0 {
		return errors.New("at least one version is required")
	}

	for _, version := range entry.Versions {
		if version.Version == "" {
			return errors.New("version number is required")
		}
		if version.DownloadURL == "" {
			return errors.New("download URL is required")
		}
		if version.Checksum == "" {
			return errors.New("checksum is required")
		}
	}

	return nil
}

// GetPluginDependencies 获取插件依赖
func (pr *PluginRegistry) GetPluginDependencies(pluginID string) ([]*PluginEntry, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return nil, errors.New("plugin not found")
	}

	dependencies := make([]*PluginEntry, 0, len(entry.Dependencies))
	for _, dep := range entry.Dependencies {
		if depEntry, exists := pr.plugins[dep.ID]; exists {
			dependencies = append(dependencies, depEntry)
		}
	}

	return dependencies, nil
}

// CheckDependencies 检查依赖关系
func (pr *PluginRegistry) CheckDependencies(pluginID string) error {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	entry, exists := pr.plugins[pluginID]
	if !exists {
		return errors.New("plugin not found")
	}

	for _, dep := range entry.Dependencies {
		if dep.Optional {
			continue
		}

		depEntry, exists := pr.plugins[dep.ID]
		if !exists {
			return fmt.Errorf("required dependency not found: %s", dep.ID)
		}

		// 检查版本兼容性
		if dep.Version != "" && !pr.isVersionCompatible(depEntry.Version, dep.Version) {
			return fmt.Errorf("dependency version mismatch: %s requires %s, found %s",
				dep.ID, dep.Version, depEntry.Version)
		}
	}

	return nil
}

// isVersionCompatible 检查版本兼容性
func (pr *PluginRegistry) isVersionCompatible(current, required string) bool {
	// 简单的版本兼容性检查
	// 实际应用中应该使用更复杂的语义版本兼容性检查库
	return current >= required
}
