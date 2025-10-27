package registry

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DiscoveryService 插件发现服务接口
type DiscoveryService interface {
	// FindPlugins 查找插件
	FindPlugins(ctx context.Context, query *DiscoveryQuery) (*DiscoveryResult, error)

	// RecommendPlugins 推荐插件
	RecommendPlugins(ctx context.Context, userID string, context *RecommendationContext) (*RecommendationResult, error)

	// GetSimilarPlugins 获取相似插件
	GetSimilarPlugins(ctx context.Context, pluginID string, limit int) ([]*PluginRegistration, error)

	// GetPopularPlugins 获取热门插件
	GetPopularPlugins(ctx context.Context, category string, limit int) ([]*PluginRegistration, error)

	// SearchPlugins 搜索插件
	SearchPlugins(ctx context.Context, searchTerm string, filters *SearchFilters) (*SearchResult, error)

	// GetPluginsByCategory 按分类获取插件
	GetPluginsByCategory(ctx context.Context, category string, sortBy string) ([]*PluginRegistration, error)
}

// DiscoveryQuery 发现查询
type DiscoveryQuery struct {
	Keywords             []string          `json:"keywords"`
	Categories           []string          `json:"categories"`
	RequiredPermissions  []string          `json:"required_permissions"`
	ExcludedPlugins      []string          `json:"excluded_plugins"`
	MinRating            float64           `json:"min_rating"`
	MaxResults           int               `json:"max_results"`
	SortBy               string            `json:"sort_by"` // relevance, popularity, rating, date
	IncludeExperimental  bool              `json:"include_experimental"`
	UserPreferences      map[string]string `json:"user_preferences"`
}

// DiscoveryResult 发现结果
type DiscoveryResult struct {
	Plugins      []*ScoredPlugin `json:"plugins"`
	TotalCount   int             `json:"total_count"`
	SearchTime   time.Duration   `json:"search_time"`
	Suggestions  []string        `json:"suggestions"`
	Categories   []string        `json:"categories"`
}

// ScoredPlugin 评分插件
type ScoredPlugin struct {
	Plugin     *PluginRegistration `json:"plugin"`
	Score      float64             `json:"score"`
	Reasons    []string            `json:"reasons"`
	Highlights []string            `json:"highlights"`
}

// RecommendationContext 推荐上下文
type RecommendationContext struct {
	UserHistory      []string          `json:"user_history"`       // 用户使用过的插件
	CurrentWorkflow  string            `json:"current_workflow"`   // 当前工作流
	UserPreferences  map[string]string `json:"user_preferences"`   // 用户偏好
	ProjectType      string            `json:"project_type"`       // 项目类型
	TeamSize         int               `json:"team_size"`          // 团队规模
	UsagePatterns    []string          `json:"usage_patterns"`     // 使用模式
}

// RecommendationResult 推荐结果
type RecommendationResult struct {
	Recommendations []*RecommendedPlugin `json:"recommendations"`
	Explanations    []string             `json:"explanations"`
	Confidence      float64              `json:"confidence"`
	GeneratedAt     time.Time            `json:"generated_at"`
}

// RecommendedPlugin 推荐插件
type RecommendedPlugin struct {
	Plugin      *PluginRegistration `json:"plugin"`
	Confidence  float64             `json:"confidence"`
	Reason      string              `json:"reason"`
	Benefits    []string            `json:"benefits"`
	Alternatives []string           `json:"alternatives"`
}

// SearchFilters 搜索过滤器
type SearchFilters struct {
	Categories      []string    `json:"categories"`
	Tags            []string    `json:"tags"`
	Authors         []string    `json:"authors"`
	MinRating       float64     `json:"min_rating"`
	MaxPrice        float64     `json:"max_price"`
	FreeOnly        bool        `json:"free_only"`
	RecentlyUpdated bool        `json:"recently_updated"`
	Status          []string    `json:"status"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Plugins       []*PluginRegistration `json:"plugins"`
	TotalCount    int                   `json:"total_count"`
	SearchTime    time.Duration         `json:"search_time"`
	Facets        map[string][]Facet    `json:"facets"`
	DidYouMean    string                `json:"did_you_mean,omitempty"`
}

// Facet 搜索面
type Facet struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// DefaultDiscoveryService 默认发现服务实现
type DefaultDiscoveryService struct {
	registry PluginRegistry
	logger   *logrus.Logger
}

// NewDefaultDiscoveryService 创建默认发现服务
func NewDefaultDiscoveryService(registry PluginRegistry, logger *logrus.Logger) *DefaultDiscoveryService {
	return &DefaultDiscoveryService{
		registry: registry,
		logger:   logger,
	}
}

// FindPlugins 查找插件
func (s *DefaultDiscoveryService) FindPlugins(ctx context.Context, query *DiscoveryQuery) (*DiscoveryResult, error) {
	startTime := time.Now()
	
	s.logger.WithField("query", query).Debug("Finding plugins")

	// 获取所有活跃插件
	filter := &PluginFilter{
		Status: StatusActive,
	}
	
	plugins, err := s.registry.ListPlugins(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	// 评分和过滤
	var scoredPlugins []*ScoredPlugin
	for _, plugin := range plugins {
		if s.shouldExcludePlugin(plugin, query) {
			continue
		}

		score, reasons, highlights := s.calculateRelevanceScore(plugin, query)
		if score > 0 {
			scoredPlugins = append(scoredPlugins, &ScoredPlugin{
				Plugin:     plugin,
				Score:      score,
				Reasons:    reasons,
				Highlights: highlights,
			})
		}
	}

	// 排序
	s.sortScoredPlugins(scoredPlugins, query.SortBy)

	// 应用结果限制
	if query.MaxResults > 0 && len(scoredPlugins) > query.MaxResults {
		scoredPlugins = scoredPlugins[:query.MaxResults]
	}

	// 生成建议和分类
	suggestions := s.generateSuggestions(query, plugins)
	categories := s.extractCategories(plugins)

	return &DiscoveryResult{
		Plugins:     scoredPlugins,
		TotalCount:  len(scoredPlugins),
		SearchTime:  time.Since(startTime),
		Suggestions: suggestions,
		Categories:  categories,
	}, nil
}

// RecommendPlugins 推荐插件
func (s *DefaultDiscoveryService) RecommendPlugins(ctx context.Context, userID string, recCtx *RecommendationContext) (*RecommendationResult, error) {
	s.logger.WithField("user_id", userID).Debug("Generating plugin recommendations")

	// 获取所有活跃插件
	filter := &PluginFilter{
		Status: StatusActive,
	}
	
	plugins, err := s.registry.ListPlugins(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	var recommendations []*RecommendedPlugin
	var explanations []string

	// 基于用户历史的协同过滤推荐
	if len(recCtx.UserHistory) > 0 {
		collabRecs := s.generateCollaborativeRecommendations(plugins, recCtx)
		recommendations = append(recommendations, collabRecs...)
		if len(collabRecs) > 0 {
			explanations = append(explanations, "基于您使用过的插件推荐")
		}
	}

	// 基于内容的推荐
	contentRecs := s.generateContentBasedRecommendations(plugins, recCtx)
	recommendations = append(recommendations, contentRecs...)
	if len(contentRecs) > 0 {
		explanations = append(explanations, "基于您的项目类型和偏好推荐")
	}

	// 热门插件推荐
	popularRecs := s.generatePopularityRecommendations(plugins, recCtx)
	recommendations = append(recommendations, popularRecs...)
	if len(popularRecs) > 0 {
		explanations = append(explanations, "当前热门插件推荐")
	}

	// 去重和排序
	recommendations = s.deduplicateRecommendations(recommendations)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Confidence > recommendations[j].Confidence
	})

	// 限制结果数量
	if len(recommendations) > 10 {
		recommendations = recommendations[:10]
	}

	confidence := s.calculateOverallConfidence(recommendations)

	return &RecommendationResult{
		Recommendations: recommendations,
		Explanations:    explanations,
		Confidence:      confidence,
		GeneratedAt:     time.Now(),
	}, nil
}

// GetSimilarPlugins 获取相似插件
func (s *DefaultDiscoveryService) GetSimilarPlugins(ctx context.Context, pluginID string, limit int) ([]*PluginRegistration, error) {
	// 获取目标插件
	targetPlugin, err := s.registry.GetPlugin(ctx, pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target plugin: %w", err)
	}

	// 获取所有活跃插件
	filter := &PluginFilter{
		Status: StatusActive,
	}
	
	plugins, err := s.registry.ListPlugins(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	// 计算相似度
	type similarPlugin struct {
		plugin     *PluginRegistration
		similarity float64
	}

	var similarities []similarPlugin
	for _, plugin := range plugins {
		if plugin.ID == pluginID {
			continue
		}

		similarity := s.calculateSimilarity(targetPlugin, plugin)
		if similarity > 0.3 { // 相似度阈值
			similarities = append(similarities, similarPlugin{
				plugin:     plugin,
				similarity: similarity,
			})
		}
	}

	// 按相似度排序
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].similarity > similarities[j].similarity
	})

	// 提取结果
	var result []*PluginRegistration
	for i, sim := range similarities {
		if i >= limit {
			break
		}
		result = append(result, sim.plugin)
	}

	return result, nil
}

// GetPopularPlugins 获取热门插件
func (s *DefaultDiscoveryService) GetPopularPlugins(ctx context.Context, category string, limit int) ([]*PluginRegistration, error) {
	filter := &PluginFilter{
		Status: StatusActive,
	}
	
	if category != "" {
		filter.Category = category
	}

	plugins, err := s.registry.ListPlugins(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	// 按热门度排序（这里简化为按请求数排序）
	sort.Slice(plugins, func(i, j int) bool {
		iMetrics := plugins[i].Metrics
		jMetrics := plugins[j].Metrics
		
		if iMetrics == nil && jMetrics == nil {
			return false
		}
		if iMetrics == nil {
			return false
		}
		if jMetrics == nil {
			return true
		}
		
		return iMetrics.RequestCount > jMetrics.RequestCount
	})

	// 应用限制
	if limit > 0 && len(plugins) > limit {
		plugins = plugins[:limit]
	}

	return plugins, nil
}

// SearchPlugins 搜索插件
func (s *DefaultDiscoveryService) SearchPlugins(ctx context.Context, searchTerm string, filters *SearchFilters) (*SearchResult, error) {
	startTime := time.Now()
	
	s.logger.WithFields(logrus.Fields{
		"search_term": searchTerm,
		"filters":     filters,
	}).Debug("Searching plugins")

	// 构建过滤器
	filter := &PluginFilter{
		Status: StatusActive,
	}
	
	if filters != nil {
		if len(filters.Categories) > 0 {
			filter.Category = filters.Categories[0] // 简化处理
		}
		if len(filters.Authors) > 0 {
			filter.Author = filters.Authors[0] // 简化处理
		}
	}

	plugins, err := s.registry.ListPlugins(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	// 文本搜索
	var matchedPlugins []*PluginRegistration
	searchTermLower := strings.ToLower(searchTerm)
	
	for _, plugin := range plugins {
		if s.matchesSearchTerm(plugin, searchTermLower) {
			if filters == nil || s.matchesSearchFilters(plugin, filters) {
				matchedPlugins = append(matchedPlugins, plugin)
			}
		}
	}

	// 生成搜索面
	facets := s.generateSearchFacets(matchedPlugins)

	return &SearchResult{
		Plugins:    matchedPlugins,
		TotalCount: len(matchedPlugins),
		SearchTime: time.Since(startTime),
		Facets:     facets,
	}, nil
}

// GetPluginsByCategory 按分类获取插件
func (s *DefaultDiscoveryService) GetPluginsByCategory(ctx context.Context, category string, sortBy string) ([]*PluginRegistration, error) {
	filter := &PluginFilter{
		Status:   StatusActive,
		Category: category,
	}

	plugins, err := s.registry.ListPlugins(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	// 排序
	switch sortBy {
	case "name":
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Name < plugins[j].Name
		})
	case "date":
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].RegisteredAt.After(plugins[j].RegisteredAt)
		})
	case "popularity":
		sort.Slice(plugins, func(i, j int) bool {
			iCount := int64(0)
			jCount := int64(0)
			if plugins[i].Metrics != nil {
				iCount = plugins[i].Metrics.RequestCount
			}
			if plugins[j].Metrics != nil {
				jCount = plugins[j].Metrics.RequestCount
			}
			return iCount > jCount
		})
	}

	return plugins, nil
}

// 辅助方法

func (s *DefaultDiscoveryService) shouldExcludePlugin(plugin *PluginRegistration, query *DiscoveryQuery) bool {
	// 检查排除列表
	for _, excludeID := range query.ExcludedPlugins {
		if plugin.ID == excludeID {
			return true
		}
	}

	// 检查分类过滤
	if len(query.Categories) > 0 {
		found := false
		for _, category := range query.Categories {
			if plugin.Category == category {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

func (s *DefaultDiscoveryService) calculateRelevanceScore(plugin *PluginRegistration, query *DiscoveryQuery) (float64, []string, []string) {
	score := 0.0
	var reasons []string
	var highlights []string

	// 关键词匹配
	for _, keyword := range query.Keywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(strings.ToLower(plugin.Name), keywordLower) {
			score += 10.0
			reasons = append(reasons, "名称匹配关键词")
			highlights = append(highlights, plugin.Name)
		}
		if strings.Contains(strings.ToLower(plugin.Description), keywordLower) {
			score += 5.0
			reasons = append(reasons, "描述匹配关键词")
		}
		for _, tag := range plugin.Tags {
			if strings.Contains(strings.ToLower(tag), keywordLower) {
				score += 3.0
				reasons = append(reasons, "标签匹配关键词")
			}
		}
	}

	// 权限匹配
	for _, requiredPerm := range query.RequiredPermissions {
		for _, pluginPerm := range plugin.Permissions {
			if pluginPerm == requiredPerm {
				score += 8.0
				reasons = append(reasons, "具备所需权限")
				break
			}
		}
	}

	// 基础分数
	if score == 0 {
		score = 1.0 // 基础分数
	}

	return score, reasons, highlights
}

func (s *DefaultDiscoveryService) sortScoredPlugins(plugins []*ScoredPlugin, sortBy string) {
	switch sortBy {
	case "popularity":
		sort.Slice(plugins, func(i, j int) bool {
			iCount := int64(0)
			jCount := int64(0)
			if plugins[i].Plugin.Metrics != nil {
				iCount = plugins[i].Plugin.Metrics.RequestCount
			}
			if plugins[j].Plugin.Metrics != nil {
				jCount = plugins[j].Plugin.Metrics.RequestCount
			}
			return iCount > jCount
		})
	case "date":
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Plugin.RegisteredAt.After(plugins[j].Plugin.RegisteredAt)
		})
	default: // relevance
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Score > plugins[j].Score
		})
	}
}

func (s *DefaultDiscoveryService) generateSuggestions(query *DiscoveryQuery, plugins []*PluginRegistration) []string {
	// 简化实现：基于现有插件生成建议
	var suggestions []string
	
	categoryCount := make(map[string]int)
	for _, plugin := range plugins {
		categoryCount[plugin.Category]++
	}

	for category, count := range categoryCount {
		if count > 5 { // 如果某个分类有足够多的插件
			suggestions = append(suggestions, category)
		}
	}

	return suggestions
}

func (s *DefaultDiscoveryService) extractCategories(plugins []*PluginRegistration) []string {
	categorySet := make(map[string]bool)
	for _, plugin := range plugins {
		categorySet[plugin.Category] = true
	}

	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}

	sort.Strings(categories)
	return categories
}

func (s *DefaultDiscoveryService) generateCollaborativeRecommendations(plugins []*PluginRegistration, ctx *RecommendationContext) []*RecommendedPlugin {
	// 简化的协同过滤实现
	var recommendations []*RecommendedPlugin

	for _, plugin := range plugins {
		// 检查是否已在用户历史中
		inHistory := false
		for _, historyID := range ctx.UserHistory {
			if plugin.ID == historyID {
				inHistory = true
				break
			}
		}

		if !inHistory {
			// 基于相似用户的使用模式推荐
			confidence := 0.6 // 简化的置信度计算
			recommendations = append(recommendations, &RecommendedPlugin{
				Plugin:     plugin,
				Confidence: confidence,
				Reason:     "其他用户也使用了这个插件",
				Benefits:   []string{"提高工作效率", "增强功能"},
			})
		}
	}

	return recommendations
}

func (s *DefaultDiscoveryService) generateContentBasedRecommendations(plugins []*PluginRegistration, ctx *RecommendationContext) []*RecommendedPlugin {
	var recommendations []*RecommendedPlugin

	for _, plugin := range plugins {
		confidence := 0.0

		// 基于项目类型匹配
		if ctx.ProjectType != "" && plugin.Category == ctx.ProjectType {
			confidence += 0.4
		}

		// 基于用户偏好匹配
		for prefKey, prefValue := range ctx.UserPreferences {
			for _, tag := range plugin.Tags {
				if strings.Contains(tag, prefValue) {
					confidence += 0.2
					break
				}
			}
		}

		if confidence > 0.3 {
			recommendations = append(recommendations, &RecommendedPlugin{
				Plugin:     plugin,
				Confidence: confidence,
				Reason:     "匹配您的项目类型和偏好",
				Benefits:   []string{"符合项目需求", "提升开发体验"},
			})
		}
	}

	return recommendations
}

func (s *DefaultDiscoveryService) generatePopularityRecommendations(plugins []*PluginRegistration, ctx *RecommendationContext) []*RecommendedPlugin {
	var recommendations []*RecommendedPlugin

	// 按热门度排序
	sort.Slice(plugins, func(i, j int) bool {
		iCount := int64(0)
		jCount := int64(0)
		if plugins[i].Metrics != nil {
			iCount = plugins[i].Metrics.RequestCount
		}
		if plugins[j].Metrics != nil {
			jCount = plugins[j].Metrics.RequestCount
		}
		return iCount > jCount
	})

	// 取前5个热门插件
	for i, plugin := range plugins {
		if i >= 5 {
			break
		}

		recommendations = append(recommendations, &RecommendedPlugin{
			Plugin:     plugin,
			Confidence: 0.5,
			Reason:     "当前热门插件",
			Benefits:   []string{"广泛使用", "社区支持好"},
		})
	}

	return recommendations
}

func (s *DefaultDiscoveryService) deduplicateRecommendations(recommendations []*RecommendedPlugin) []*RecommendedPlugin {
	seen := make(map[string]bool)
	var result []*RecommendedPlugin

	for _, rec := range recommendations {
		if !seen[rec.Plugin.ID] {
			seen[rec.Plugin.ID] = true
			result = append(result, rec)
		}
	}

	return result
}

func (s *DefaultDiscoveryService) calculateOverallConfidence(recommendations []*RecommendedPlugin) float64 {
	if len(recommendations) == 0 {
		return 0.0
	}

	total := 0.0
	for _, rec := range recommendations {
		total += rec.Confidence
	}

	return total / float64(len(recommendations))
}

func (s *DefaultDiscoveryService) calculateSimilarity(plugin1, plugin2 *PluginRegistration) float64 {
	similarity := 0.0

	// 分类相似度
	if plugin1.Category == plugin2.Category {
		similarity += 0.3
	}

	// 标签相似度
	commonTags := 0
	for _, tag1 := range plugin1.Tags {
		for _, tag2 := range plugin2.Tags {
			if tag1 == tag2 {
				commonTags++
				break
			}
		}
	}
	
	if len(plugin1.Tags) > 0 || len(plugin2.Tags) > 0 {
		maxTags := len(plugin1.Tags)
		if len(plugin2.Tags) > maxTags {
			maxTags = len(plugin2.Tags)
		}
		similarity += 0.4 * float64(commonTags) / float64(maxTags)
	}

	// 权限相似度
	commonPerms := 0
	for _, perm1 := range plugin1.Permissions {
		for _, perm2 := range plugin2.Permissions {
			if perm1 == perm2 {
				commonPerms++
				break
			}
		}
	}
	
	if len(plugin1.Permissions) > 0 || len(plugin2.Permissions) > 0 {
		maxPerms := len(plugin1.Permissions)
		if len(plugin2.Permissions) > maxPerms {
			maxPerms = len(plugin2.Permissions)
		}
		similarity += 0.3 * float64(commonPerms) / float64(maxPerms)
	}

	return similarity
}

func (s *DefaultDiscoveryService) matchesSearchTerm(plugin *PluginRegistration, searchTerm string) bool {
	// 检查名称
	if strings.Contains(strings.ToLower(plugin.Name), searchTerm) {
		return true
	}

	// 检查描述
	if strings.Contains(strings.ToLower(plugin.Description), searchTerm) {
		return true
	}

	// 检查标签
	for _, tag := range plugin.Tags {
		if strings.Contains(strings.ToLower(tag), searchTerm) {
			return true
		}
	}

	// 检查作者
	if strings.Contains(strings.ToLower(plugin.Author), searchTerm) {
		return true
	}

	return false
}

func (s *DefaultDiscoveryService) matchesSearchFilters(plugin *PluginRegistration, filters *SearchFilters) bool {
	// 检查分类
	if len(filters.Categories) > 0 {
		found := false
		for _, category := range filters.Categories {
			if plugin.Category == category {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查标签
	if len(filters.Tags) > 0 {
		found := false
		for _, filterTag := range filters.Tags {
			for _, pluginTag := range plugin.Tags {
				if pluginTag == filterTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查作者
	if len(filters.Authors) > 0 {
		found := false
		for _, author := range filters.Authors {
			if plugin.Author == author {
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

func (s *DefaultDiscoveryService) generateSearchFacets(plugins []*PluginRegistration) map[string][]Facet {
	facets := make(map[string][]Facet)

	// 分类面
	categoryCount := make(map[string]int)
	for _, plugin := range plugins {
		categoryCount[plugin.Category]++
	}
	
	var categoryFacets []Facet
	for category, count := range categoryCount {
		categoryFacets = append(categoryFacets, Facet{
			Value: category,
			Count: count,
		})
	}
	facets["categories"] = categoryFacets

	// 作者面
	authorCount := make(map[string]int)
	for _, plugin := range plugins {
		authorCount[plugin.Author]++
	}
	
	var authorFacets []Facet
	for author, count := range authorCount {
		authorFacets = append(authorFacets, Facet{
			Value: author,
			Count: count,
		})
	}
	facets["authors"] = authorFacets

	return facets
}