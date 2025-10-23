package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun/pkg/shared/database"
	"github.com/codetaoist/laojun/pkg/shared/models"
	"github.com/google/uuid"
)

// AdminPluginService 总后台插件管理服务
type AdminPluginService struct {
	db                *database.DB
	marketplaceClient MarketplaceAPIClient
}

// NewAdminPluginService 创建总后台插件管理服务
func NewAdminPluginService(db *database.DB, marketplaceClient MarketplaceAPIClient) *AdminPluginService {
	return &AdminPluginService{
		db:                db,
		marketplaceClient: marketplaceClient,
	}
}

// GetPluginsForAdmin 获取管理员插件列表
func (s *AdminPluginService) GetPluginsForAdmin(params models.AdminPluginSearchParams) ([]models.Plugin, *models.PaginationMeta, error) {
	var plugins []models.Plugin
	var total int64

	// 构建查询条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if params.Query != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIndex, argIndex+1))
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	if params.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *params.CategoryID)
		argIndex++
	}

	if params.DeveloperID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.developer_id = $%d", argIndex))
		args = append(args, *params.DeveloperID)
		argIndex++
	}

	if params.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.status = $%d", argIndex))
		args = append(args, *params.Status)
		argIndex++
	}

	if params.Featured != nil && *params.Featured {
		whereConditions = append(whereConditions, fmt.Sprintf("p.is_featured = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	if params.Free != nil && *params.Free {
		whereConditions = append(whereConditions, fmt.Sprintf("p.is_free = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 计算总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM plugins p WHERE %s", whereClause)
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count plugins: %w", err)
	}

	// 构建排序
	sortBy := "p.created_at"
	if params.SortBy != "" {
		switch params.SortBy {
		case "name":
			sortBy = "p.name"
		case "rating":
			sortBy = "p.rating"
		case "downloads":
			sortBy = "p.download_count"
		case "updated_at":
			sortBy = "p.updated_at"
		}
	}

	sortOrder := "DESC"
	if params.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// 分页
	limit := 20
	if params.Limit > 0 {
		limit = params.Limit
	}
	offset := 0
	if params.Page > 1 {
		offset = (params.Page - 1) * limit
	}

	// 查询数据
	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.short_description, p.author, p.developer_id,
		       p.version, p.category_id, p.price, p.is_free, p.is_featured, p.is_active,
		       p.download_count, p.rating, p.review_count, p.icon_url, p.banner_url,
		       p.screenshots, p.tags, p.requirements, p.changelog, p.created_at, p.updated_at
		FROM plugins p 
		WHERE %s 
		ORDER BY %s %s 
		LIMIT $%d OFFSET $%d`,
		whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	rows, err := s.db.Query(dataQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query plugins: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Plugin
		var screenshots, tags sql.NullString

		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.ShortDescription, &p.Author, &p.DeveloperID,
			&p.Version, &p.CategoryID, &p.Price, &p.IsFree, &p.IsFeatured, &p.IsActive,
			&p.DownloadCount, &p.Rating, &p.ReviewCount, &p.IconURL, &p.BannerURL,
			&screenshots, &tags, &p.Requirements, &p.Changelog, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan plugin: %w", err)
		}

		// 处理JSON字段
		if screenshots.Valid {
			// 这里需要解析JSON，简化处理
			p.Screenshots = []string{}
		}
		if tags.Valid {
			// 这里需要解析JSON，简化处理
			p.Tags = []string{}
		}

		plugins = append(plugins, p)
	}

	// 计算分页信息
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := &models.PaginationMeta{
		Page:       params.Page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return plugins, meta, nil
}

// GetPluginForAdmin 获取管理员插件详情
func (s *AdminPluginService) GetPluginForAdmin(pluginID uuid.UUID) (*models.AdminPluginDetail, error) {
	query := `
		SELECT p.id, p.name, p.description, p.short_description, p.author, p.developer_id,
		       p.version, p.category_id, p.price, p.is_free, p.is_featured, p.is_active,
		       p.download_count, p.rating, p.review_count, p.icon_url, p.banner_url,
		       p.screenshots, p.tags, p.requirements, p.changelog, p.created_at, p.updated_at
		FROM plugins p 
		WHERE p.id = $1`

	var detail models.AdminPluginDetail
	var screenshots, tags sql.NullString

	err := s.db.QueryRow(query, pluginID).Scan(
		&detail.ID, &detail.Name, &detail.Description, &detail.ShortDescription, &detail.Author, &detail.DeveloperID,
		&detail.Version, &detail.CategoryID, &detail.Price, &detail.IsFree, &detail.IsFeatured, &detail.IsActive,
		&detail.DownloadCount, &detail.Rating, &detail.ReviewCount, &detail.IconURL, &detail.BannerURL,
		&screenshots, &tags, &detail.Requirements, &detail.Changelog, &detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found")
		}
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}

	// 处理JSON字段
	if screenshots.Valid {
		detail.Screenshots = []string{}
	}
	if tags.Valid {
		detail.Tags = []string{}
	}

	return &detail, nil
}

// UpdatePluginStatus 更新插件状态
func (s *AdminPluginService) UpdatePluginStatus(pluginID uuid.UUID, status string, reason string, adminID uuid.UUID) error {
	query := `
		UPDATE plugins 
		SET is_active = $2, updated_at = $3
		WHERE id = $1`

	isActive := status == "approved"
	_, err := s.db.Exec(query, pluginID, isActive, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	// 同步状态到插件市场
	if s.marketplaceClient != nil {
		if err := s.syncStatusToMarketplace(pluginID, status); err != nil {
			// 记录错误但不阻止操作
			fmt.Printf("Failed to sync status to marketplace: %v\n", err)
		}
	}

	return nil
}

// GetPluginStats 获取插件统计信息
func (s *AdminPluginService) GetPluginStats(pluginID uuid.UUID) (*models.PluginStats, error) {
	stats := &models.PluginStats{
		PluginID:  pluginID,
		UpdatedAt: time.Now(),
	}

	// 查询基础统计信息
	query := `
		SELECT download_count, rating, review_count
		FROM plugins 
		WHERE id = $1`

	err := s.db.QueryRow(query, pluginID).Scan(
		&stats.TotalDownloads,
		&stats.AverageRating,
		&stats.TotalReviews,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin stats: %w", err)
	}

	// 这里可以添加更多统计查询
	stats.MonthlyDownloads = stats.TotalDownloads / 12 // 简化计算
	stats.WeeklyDownloads = stats.MonthlyDownloads / 4
	stats.DailyDownloads = stats.WeeklyDownloads / 7

	return stats, nil
}

// GetPluginConfig 获取插件配置
func (s *AdminPluginService) GetPluginConfig(pluginID uuid.UUID) (*models.PluginConfig, error) {
	// 简化实现，返回默认配置
	config := &models.PluginConfig{
		ID:          uuid.New(),
		PluginID:    pluginID,
		ConfigKey:   "default",
		ConfigValue: map[string]interface{}{},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return config, nil
}

// UpdatePluginConfig 更新插件配置
func (s *AdminPluginService) UpdatePluginConfig(pluginID uuid.UUID, config map[string]interface{}, adminID uuid.UUID) error {
	// 简化实现
	return nil
}

// GetDashboardStats 获取仪表板统计
func (s *AdminPluginService) GetDashboardStats() (*models.AdminDashboardStats, error) {
	stats := &models.AdminDashboardStats{
		UpdatedAt: time.Now(),
	}

	// 查询插件统计
	pluginQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN is_active = true THEN 1 END) as approved,
			COUNT(CASE WHEN is_active = false THEN 1 END) as rejected
		FROM plugins`

	err := s.db.QueryRow(pluginQuery).Scan(
		&stats.TotalPlugins,
		&stats.ApprovedPlugins,
		&stats.RejectedPlugins,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin stats: %w", err)
	}

	// 查询下载统计
	downloadQuery := `
		SELECT COALESCE(SUM(download_count), 0) as total_downloads
		FROM plugins`

	err = s.db.QueryRow(downloadQuery).Scan(&stats.TotalDownloads)
	if err != nil {
		return nil, fmt.Errorf("failed to get download stats: %w", err)
	}

	stats.MonthlyDownloads = stats.TotalDownloads / 12 // 简化计算

	return stats, nil
}

// GetPluginLogs 获取插件日志
func (s *AdminPluginService) GetPluginLogs(pluginID uuid.UUID, page, limit int, level string) ([]models.PluginLog, *models.PaginationMeta, error) {
	// 简化实现，返回空日志
	logs := []models.PluginLog{}
	meta := &models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      0,
		TotalPages: 0,
	}

	return logs, meta, nil
}

// syncStatusToMarketplace 同步状态到插件市场
func (s *AdminPluginService) syncStatusToMarketplace(pluginID uuid.UUID, status string) error {
	if s.marketplaceClient == nil {
		return fmt.Errorf("marketplace client not configured")
	}
	return s.marketplaceClient.UpdatePluginStatus(pluginID, status)
}

// MarketplaceAPIClient 插件市场API客户端接口
type MarketplaceAPIClient interface {
	UpdatePluginStatus(pluginID uuid.UUID, status string) error
	GetPlugin(pluginID uuid.UUID) (*models.Plugin, error)
	SyncPlugin(plugin *models.Plugin) error
}