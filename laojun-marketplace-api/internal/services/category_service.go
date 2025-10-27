package services

import (
	"database/sql"
	"fmt"

	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// CategoryService 分类服务
type CategoryService struct {
	db *shareddb.DB
}

// NewCategoryService 创建分类服务
func NewCategoryService(db *shareddb.DB) *CategoryService {
	return &CategoryService{db: db}
}

// GetCategories 获取所有分类
func (s *CategoryService) GetCategories() ([]models.Category, error) {
	query := "SELECT id, name, description, icon, color, sort_order, is_active, created_at, updated_at FROM mp_categories ORDER BY name"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		var description, icon, color sql.NullString
		
		err := rows.Scan(
			&category.ID, &category.Name, &description,
			&icon, &color, &category.SortOrder, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		
		// 处理可能为NULL的字段
		if description.Valid {
			category.Description = &description.String
		}
		if icon.Valid {
			category.Icon = &icon.String
		}
		if color.Valid {
			category.Color = color.String
		} else {
			category.Color = "#1890ff" // 默认颜色
		}
		
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate categories: %w", err)
	}

	return categories, nil
}

// GetCategory 获取单个分类
func (s *CategoryService) GetCategory(categoryID uuid.UUID) (*models.Category, error) {
	query := "SELECT id, name, description, icon, color, sort_order, is_active, created_at, updated_at FROM mp_categories WHERE id = $1"

	var category models.Category
	var description, icon, color sql.NullString
	
	err := s.db.QueryRow(query, categoryID).Scan(
		&category.ID, &category.Name, &description,
		&icon, &color, &category.SortOrder, &category.IsActive,
		&category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// 处理可能为NULL的字段
	if description.Valid {
		category.Description = &description.String
	}
	if icon.Valid {
		category.Icon = &icon.String
	}
	if color.Valid {
		category.Color = color.String
	} else {
		category.Color = "#1890ff" // 默认颜色
	}

	return &category, nil
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(name, description, icon string) (*models.Category, error) {
	categoryID := uuid.New()
	query := `
		INSERT INTO mp_categories (id, name, description, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, name, description, icon, created_at, updated_at`

	var category models.Category
	err := s.db.QueryRow(query, categoryID, name, description, icon).Scan(
		&category.ID, &category.Name, &category.Description,
		&category.Icon, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(categoryID uuid.UUID, name, description, icon string) (*models.Category, error) {
	query := `
		UPDATE mp_categories
		SET name = $2, description = $3, icon = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, icon, created_at, updated_at`

	var category models.Category
	err := s.db.QueryRow(query, categoryID, name, description, icon).Scan(
		&category.ID, &category.Name, &category.Description,
		&category.Icon, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &category, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(categoryID uuid.UUID) error {
	// 检查是否有插件使用此分类
	var count int
	checkQuery := "SELECT COUNT(*) FROM mp_plugins WHERE category_id = $1"
	err := s.db.QueryRow(checkQuery, categoryID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete category: %d plugins are using this category", count)
	}

	// 删除分类
	deleteQuery := "DELETE FROM mp_categories WHERE id = $1"
	result, err := s.db.Exec(deleteQuery, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// GetCategoryPlugins 获取分类下的插件
func (s *CategoryService) GetCategoryPlugins(categoryID uuid.UUID) ([]models.Plugin, error) {
	query := `
		SELECT id, name, description, version, author, 
		       category_id, is_active, download_count, rating, created_at, updated_at
		FROM mp_plugins 
		WHERE category_id = $1 AND is_active = true
		ORDER BY download_count DESC
	`

	rows, err := s.db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query category plugins: %w", err)
	}
	defer rows.Close()

	var plugins []models.Plugin
	for rows.Next() {
		var plugin models.Plugin
		err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Version,
			&plugin.Author, &plugin.CategoryID, &plugin.IsActive, &plugin.DownloadCount, &plugin.Rating,
			&plugin.CreatedAt, &plugin.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate plugins: %w", err)
	}

	return plugins, nil
}

// GetCategoryWithStats 获取带统计信息的分类列表
func (s *CategoryService) GetCategoriesWithStats() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT c.id, c.name, c.description, c.icon, c.color, c.created_at, c.updated_at,
		       COUNT(p.id) as plugin_count,
		       COALESCE(AVG(p.rating), 0) as avg_rating,
		       COALESCE(SUM(p.download_count), 0) as total_downloads
		FROM categories c
		LEFT JOIN plugins p ON c.id = p.category_id
		GROUP BY c.id, c.name, c.description, c.icon, c.color, c.created_at, c.updated_at
		ORDER BY plugin_count DESC, c.name ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category
		var pluginCount int
		var avgRating float64
		var totalDownloads int

		err := rows.Scan(
			&category.ID, &category.Name, &category.Description,
			&category.Icon, &category.Color, &category.CreatedAt, &category.UpdatedAt,
			&pluginCount, &avgRating, &totalDownloads,
		)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"id":              category.ID,
			"name":            category.Name,
			"description":     category.Description,
			"icon":            category.Icon,
			"color":           category.Color,
			"created_at":      category.CreatedAt,
			"updated_at":      category.UpdatedAt,
			"plugin_count":    pluginCount,
			"avg_rating":      avgRating,
			"total_downloads": totalDownloads,
		}

		results = append(results, result)
	}

	return results, nil
}

// CheckCategoryExists 检查分类是否存在
func (s *CategoryService) CheckCategoryExists(id uuid.UUID) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM mp_categories WHERE id = $1)"
	err := s.db.QueryRow(query, id).Scan(&exists)
	return exists, err
}
