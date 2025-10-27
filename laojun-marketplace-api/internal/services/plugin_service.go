package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// PluginService 插件服务
type PluginService struct {
	db *shareddb.DB
}

// NewPluginService 创建插件服务
func NewPluginService(db *shareddb.DB) *PluginService {
	return &PluginService{db: db}
}

// GetPlugins 获取插件列表
func (s *PluginService) GetPlugins(params models.PluginSearchParams) ([]models.Plugin, models.PaginationMeta, error) {
	var plugins []models.Plugin
	var totalCount int

	// 构建查询条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if params.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *params.CategoryID)
		argIndex++
	}

	if params.Featured != nil && *params.Featured {
		whereConditions = append(whereConditions, fmt.Sprintf("p.is_featured = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	if params.MinPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.price >= $%d", argIndex))
		args = append(args, *params.MinPrice)
		argIndex++
	}

	if params.MaxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.price <= $%d", argIndex))
		args = append(args, *params.MaxPrice)
		argIndex++
	}

	if params.MinRating != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.rating >= $%d", argIndex))
		args = append(args, *params.MinRating)
		argIndex++
	}

	if params.Query != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIndex, argIndex+1))
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mp_plugins p 
		LEFT JOIN mp_categories c ON p.category_id = c.id 
		WHERE %s`, whereClause)

	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 构建排序
	orderBy := "p.created_at DESC"
	if params.SortBy != "" {
		switch params.SortBy {
		case "name":
			orderBy = "p.name ASC"
		case "price":
			orderBy = "p.price ASC"
		case "rating":
			orderBy = "p.rating DESC"
		case "downloads":
			orderBy = "p.download_count DESC"
		case "created_at":
			orderBy = "p.created_at DESC"
		}
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.Limit

	// 获取插件列表
	query := fmt.Sprintf(`
		SELECT 
			p.id, p.name, p.description, p.author, p.developer_id, p.version, p.icon_url, 
			p.price, p.rating, p.download_count, p.is_featured, p.created_at, p.updated_at,
			p.category_id, c.name as category_name, c.icon as category_icon, c.color as category_color
		FROM mp_plugins p
		LEFT JOIN mp_categories c ON p.category_id = c.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, params.Limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var plugin models.Plugin
		var categoryName, categoryIcon, categoryColor sql.NullString

		err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Author, &plugin.DeveloperID,
			&plugin.Version, &plugin.IconURL, &plugin.Price, &plugin.Rating,
			&plugin.DownloadCount, &plugin.IsFeatured, &plugin.CreatedAt, &plugin.UpdatedAt,
			&plugin.CategoryID, &categoryName, &categoryIcon, &categoryColor,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 设置分类信息
		if categoryName.Valid && plugin.CategoryID != nil {
			var icon *string
			if categoryIcon.Valid {
				icon = &categoryIcon.String
			}
			plugin.Category = &models.Category{
				ID:    *plugin.CategoryID,
				Name:  categoryName.String,
				Icon:  icon,
				Color: categoryColor.String,
			}
		}

		plugins = append(plugins, plugin)
	}

	// 计算分页信息
	totalPages := (totalCount + params.Limit - 1) / params.Limit

	meta := models.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return plugins, meta, nil
}

// GetPlugin 获取单个插件详情
func (s *PluginService) GetPlugin(id uuid.UUID) (*models.Plugin, error) {
	var plugin models.Plugin
	var categoryName, categoryIcon, categoryColor sql.NullString

	query := `
		SELECT 
			p.id, p.name, p.description, p.short_description, p.author, p.developer_id, 
			p.version, p.icon_url, p.price, p.rating, p.download_count, p.is_featured, 
			p.created_at, p.updated_at, p.category_id, c.name as category_name, 
			c.icon as category_icon, c.color as category_color
		FROM mp_plugins p
		LEFT JOIN mp_categories c ON p.category_id = c.id
		WHERE p.id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.ShortDescription, &plugin.Author, &plugin.DeveloperID,
		&plugin.Version, &plugin.IconURL, &plugin.Price, &plugin.Rating,
		&plugin.DownloadCount, &plugin.IsFeatured, &plugin.CreatedAt, &plugin.UpdatedAt,
		&plugin.CategoryID, &categoryName, &categoryIcon, &categoryColor,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 设置分类信息
	if categoryName.Valid && plugin.CategoryID != nil {
		var icon *string
		if categoryIcon.Valid {
			icon = &categoryIcon.String
		}
		plugin.Category = &models.Category{
			ID:    *plugin.CategoryID,
			Name:  categoryName.String,
			Icon:  icon,
			Color: categoryColor.String,
		}
	}

	return &plugin, nil
}

// GetPluginsByCategory 根据分类获取插件
func (s *PluginService) GetPluginsByCategory(categoryID uuid.UUID, params models.PluginSearchParams) ([]models.Plugin, models.PaginationMeta, error) {
	params.CategoryID = &categoryID
	return s.GetPlugins(params)
}

// ToggleFavorite 切换收藏状态
func (s *PluginService) ToggleFavorite(userID, pluginID uuid.UUID) (bool, error) {
	// 检查是否已收藏
	checkQuery := "SELECT EXISTS(SELECT 1 FROM mp_favorites WHERE user_id = $1 AND plugin_id = $2)"
	var isFavorited bool
	err := s.db.QueryRow(checkQuery, userID, pluginID).Scan(&isFavorited)
	if err != nil {
		return false, err
	}

	if isFavorited {
		// 取消收藏
		_, err = s.db.Exec("DELETE FROM mp_favorites WHERE user_id = $1 AND plugin_id = $2", userID, pluginID)
		return false, err
	} else {
		// 添加收藏
		_, err = s.db.Exec("INSERT INTO mp_favorites (user_id, plugin_id) VALUES ($1, $2)", userID, pluginID)
		return true, err
	}
}

// GetUserFavorites 获取用户收藏的插件
func (s *PluginService) GetUserFavorites(userID uuid.UUID, page, limit int) ([]models.Plugin, models.PaginationMeta, error) {
	var plugins []models.Plugin
	var totalCount int

	// 获取总数
	countQuery := `
		SELECT COUNT(*) 
		FROM mp_favorites uf 
		JOIN mp_plugins p ON uf.plugin_id = p.id 
		WHERE uf.user_id = $1`

	err := s.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取收藏的插件
	query := `
		SELECT 
			p.id, p.name, p.description, p.short_description, p.author, p.developer_id, 
			p.version, p.icon_url, p.price, p.rating, p.download_count, p.is_featured, 
			p.created_at, p.updated_at, p.category_id, c.name as category_name, 
			c.icon as category_icon, c.color as category_color
		FROM mp_favorites uf
		JOIN mp_plugins p ON uf.plugin_id = p.id
		LEFT JOIN mp_categories c ON p.category_id = c.id
		WHERE uf.user_id = $1
		ORDER BY uf.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var plugin models.Plugin
		var categoryName, categoryIcon, categoryColor sql.NullString

		err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.ShortDescription, &plugin.Author, &plugin.DeveloperID,
			&plugin.Version, &plugin.IconURL, &plugin.Price, &plugin.Rating,
			&plugin.DownloadCount, &plugin.IsFeatured, &plugin.CreatedAt, &plugin.UpdatedAt,
			&plugin.CategoryID, &categoryName, &categoryIcon, &categoryColor,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 设置分类信息
		if categoryName.Valid && plugin.CategoryID != nil {
			var icon *string
			if categoryIcon.Valid {
				icon = &categoryIcon.String
			}
			plugin.Category = &models.Category{
				ID:    *plugin.CategoryID,
				Name:  categoryName.String,
				Icon:  icon,
				Color: categoryColor.String,
			}
		}

		plugins = append(plugins, plugin)
	}

	// 计算分页信息
	totalPages := (totalCount + limit - 1) / limit
	meta := models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return plugins, meta, nil
}

// PurchasePlugin 购买插件
func (s *PluginService) PurchasePlugin(userID, pluginID uuid.UUID) error {
	// 检查是否已购买
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM mp_purchases WHERE user_id = $1 AND plugin_id = $2)"
	err := s.db.QueryRow(checkQuery, userID, pluginID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("plugin already purchased")
	}

	// 获取插件价格
	var price float64
	priceQuery := "SELECT price FROM mp_plugins WHERE id = $1"
	err = s.db.QueryRow(priceQuery, pluginID).Scan(&price)
	if err != nil {
		return err
	}

	// 创建购买记录
	purchaseID := uuid.New()
	insertQuery := `
		INSERT INTO mp_purchases (id, user_id, plugin_id, amount, status, created_at) 
		VALUES ($1, $2, $3, $4, 'completed', NOW())`

	_, err = s.db.Exec(insertQuery, purchaseID, userID, pluginID, price)
	if err != nil {
		return err
	}

	// 更新插件下载次数
	_, err = s.db.Exec("UPDATE mp_plugins SET download_count = download_count + 1 WHERE id = $1", pluginID)
	return err
}

// GetUserPurchases 获取用户购买的插件
func (s *PluginService) GetUserPurchases(userID uuid.UUID, page, limit int) ([]models.Purchase, models.PaginationMeta, error) {
	var purchases []models.Purchase
	var totalCount int

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM mp_purchases WHERE user_id = $1"
	err := s.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取购买记录
	query := `
		SELECT 
			pu.id, pu.user_id, pu.plugin_id, pu.amount, pu.status, pu.created_at,
			p.name as plugin_name, p.version as plugin_version, p.icon_url as plugin_icon
		FROM mp_purchases pu
		JOIN mp_plugins p ON pu.plugin_id = p.id
		WHERE pu.user_id = $1
		ORDER BY pu.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var purchase models.Purchase
		var pluginName, pluginVersion string
		var pluginIcon sql.NullString

		err := rows.Scan(
			&purchase.ID, &purchase.UserID, &purchase.PluginID,
			&purchase.Amount, &purchase.Status, &purchase.CreatedAt,
			&pluginName, &pluginVersion, &pluginIcon,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 设置插件信息
		var iconURL *string
		if pluginIcon.Valid {
			iconURL = &pluginIcon.String
		}
		purchase.Plugin = &models.Plugin{
			ID:      purchase.PluginID,
			Name:    pluginName,
			Version: pluginVersion,
			IconURL: iconURL,
		}

		purchases = append(purchases, purchase)
	}

	// 计算分页信息
	totalPages := (totalCount + limit - 1) / limit
	meta := models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return purchases, meta, nil
}

// CreatePlugin 创建插件
func (s *PluginService) CreatePlugin(plugin *models.Plugin) error {
	plugin.ID = uuid.New()
	plugin.CreatedAt = time.Now()
	plugin.UpdatedAt = time.Now()

	query := `
		INSERT INTO mp_plugins (
			id, name, description, short_description, author, version, 
			category_id, price, is_free, is_featured, is_active,
			icon_url, banner_url, screenshots, tags, requirements, changelog
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`

	_, err := s.db.Exec(query,
		plugin.ID, plugin.Name, plugin.Description, plugin.ShortDescription,
		plugin.Author, plugin.Version, plugin.CategoryID, plugin.Price,
		plugin.IsFree, plugin.IsFeatured, plugin.IsActive,
		plugin.IconURL, plugin.BannerURL, plugin.Screenshots, plugin.Tags,
		plugin.Requirements, plugin.Changelog,
	)

	return err
}

// UpdatePlugin 更新插件
func (s *PluginService) UpdatePlugin(id uuid.UUID, plugin *models.Plugin) error {
	plugin.UpdatedAt = time.Now()

	query := `
		UPDATE mp_plugins SET 
			name = $2, description = $3, short_description = $4, author = $5,
			version = $6, category_id = $7, price = $8, is_free = $9,
			is_featured = $10, is_active = $11, icon_url = $12, banner_url = $13,
			screenshots = $14, tags = $15, requirements = $16, changelog = $17,
			updated_at = $18
		WHERE id = $1`

	result, err := s.db.Exec(query,
		id, plugin.Name, plugin.Description, plugin.ShortDescription,
		plugin.Author, plugin.Version, plugin.CategoryID, plugin.Price,
		plugin.IsFree, plugin.IsFeatured, plugin.IsActive,
		plugin.IconURL, plugin.BannerURL, plugin.Screenshots, plugin.Tags,
		plugin.Requirements, plugin.Changelog, plugin.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

// DeletePlugin 删除插件
func (s *PluginService) DeletePlugin(id uuid.UUID) error {
	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除相关的收藏记录
	_, err = tx.Exec("DELETE FROM mp_favorites WHERE plugin_id = $1", id)
	if err != nil {
		return err
	}

	// 删除相关的购买记录
	_, err = tx.Exec("DELETE FROM mp_purchases WHERE plugin_id = $1", id)
	if err != nil {
		return err
	}

	// 删除相关的评论记录
	_, err = tx.Exec("DELETE FROM mp_reviews WHERE plugin_id = $1", id)
	if err != nil {
		return err
	}

	// 删除插件
	result, err := tx.Exec("DELETE FROM plugins WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plugin not found")
	}

	return tx.Commit()
}

// UpdatePluginStatus 更新插件状态
func (s *PluginService) UpdatePluginStatus(id uuid.UUID, isActive bool) error {
	query := "UPDATE mp_plugins SET is_active = $2, updated_at = NOW() WHERE id = $1"
	result, err := s.db.Exec(query, id, isActive)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

// UpdatePluginRating 更新插件评分
func (s *PluginService) UpdatePluginRating(pluginID uuid.UUID) error {
	query := `
		UPDATE mp_plugins 
		SET rating = (
			SELECT COALESCE(AVG(rating), 0) 
			FROM mp_reviews 
			WHERE plugin_id = $1
		), review_count = (
			SELECT COUNT(*) 
			FROM mp_reviews 
			WHERE plugin_id = $1
		), updated_at = NOW()
		WHERE id = $1`

	_, err := s.db.Exec(query, pluginID)
	return err
}
