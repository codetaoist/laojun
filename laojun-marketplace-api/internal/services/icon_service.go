package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun-marketplace-api/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type IconService struct {
	db *sql.DB
}

func NewIconService(db *sql.DB) *IconService {
	return &IconService{
		db: db,
	}
}

// GetIcons 获取图标列表
func (s *IconService) GetIcons(params models.IconSearchParams) (*models.PaginatedResponse[models.Icon], error) {
	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, params.Category)
		argIndex++
	}

	if params.IconType != "" {
		conditions = append(conditions, fmt.Sprintf("icon_type = $%d", argIndex))
		args = append(args, params.IconType)
		argIndex++
	}

	if params.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *params.IsActive)
		argIndex++
	}

	if params.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR $%d = ANY(tags))", argIndex, argIndex))
		args = append(args, "%"+params.Keyword+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sys_icons %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count icons: %w", err)
	}

	// 查询数据
	offset := (params.Page - 1) * params.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, name, icon_type, icon_data, category, tags, is_active, created_at, updated_at
		FROM sys_icons %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.Limit, offset)

	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query icons: %w", err)
	}
	defer rows.Close()

	var icons []models.Icon
	for rows.Next() {
		var icon models.Icon
		err := rows.Scan(
			&icon.ID,
			&icon.Name,
			&icon.IconType,
			&icon.IconData,
			&icon.Category,
			pq.Array(&icon.Tags),
			&icon.IsActive,
			&icon.CreatedAt,
			&icon.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan icon: %w", err)
		}
		icons = append(icons, icon)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate icons: %w", err)
	}

	return &models.PaginatedResponse[models.Icon]{
		Data:       icons,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: (total + params.Limit - 1) / params.Limit,
	}, nil
}

// GetIconByID 根据ID获取图标
func (s *IconService) GetIconByID(id uuid.UUID) (*models.Icon, error) {
	query := `
		SELECT id, name, icon_type, icon_data, category, tags, is_active, created_at, updated_at
		FROM sys_icons
		WHERE id = $1
	`

	var icon models.Icon
	err := s.db.QueryRow(query, id).Scan(
		&icon.ID,
		&icon.Name,
		&icon.IconType,
		&icon.IconData,
		&icon.Category,
		pq.Array(&icon.Tags),
		&icon.IsActive,
		&icon.CreatedAt,
		&icon.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("icon not found")
		}
		return nil, fmt.Errorf("failed to get icon: %w", err)
	}

	return &icon, nil
}

// CreateIcon 创建图标
func (s *IconService) CreateIcon(req models.CreateIconRequest) (*models.Icon, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO sys_icons (id, name, icon_type, icon_data, category, tags, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, icon_type, icon_data, category, tags, is_active, created_at, updated_at
	`

	var icon models.Icon
	err := s.db.QueryRow(
		query,
		id,
		req.Name,
		req.IconType,
		req.IconData,
		req.Category,
		pq.Array(req.Tags),
		true,
		now,
		now,
	).Scan(
		&icon.ID,
		&icon.Name,
		&icon.IconType,
		&icon.IconData,
		&icon.Category,
		pq.Array(&icon.Tags),
		&icon.IsActive,
		&icon.CreatedAt,
		&icon.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create icon: %w", err)
	}

	return &icon, nil
}

// UpdateIcon 更新图标
func (s *IconService) UpdateIcon(id uuid.UUID, req models.UpdateIconRequest) (*models.Icon, error) {
	// 构建更新字段
	var setParts []string
	var args []interface{}
	argIndex := 1

	if req.Name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.IconType != "" {
		setParts = append(setParts, fmt.Sprintf("icon_type = $%d", argIndex))
		args = append(args, req.IconType)
		argIndex++
	}

	if req.IconData != "" {
		setParts = append(setParts, fmt.Sprintf("icon_data = $%d", argIndex))
		args = append(args, req.IconData)
		argIndex++
	}

	if req.Category != "" {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, req.Category)
		argIndex++
	}

	if req.Tags != nil {
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, pq.Array(req.Tags))
		argIndex++
	}

	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return s.GetIconByID(id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE sys_icons
		SET %s
		WHERE id = $%d
		RETURNING id, name, icon_type, icon_data, category, tags, is_active, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var icon models.Icon
	err := s.db.QueryRow(query, args...).Scan(
		&icon.ID,
		&icon.Name,
		&icon.IconType,
		&icon.IconData,
		&icon.Category,
		pq.Array(&icon.Tags),
		&icon.IsActive,
		&icon.CreatedAt,
		&icon.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("icon not found")
		}
		return nil, fmt.Errorf("failed to update icon: %w", err)
	}

	return &icon, nil
}

// DeleteIcon 删除图标
func (s *IconService) DeleteIcon(id uuid.UUID) error {
	query := "DELETE FROM sys_icons WHERE id = $1"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete icon: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("icon not found")
	}

	return nil
}

// GetIconStats 获取图标统计信息
func (s *IconService) GetIconStats() (*models.IconStats, error) {
	stats := &models.IconStats{
		IconTypes: make(map[string]int),
	}

	// 获取总数和活跃数
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive
		FROM sys_icons
	`
	err := s.db.QueryRow(query).Scan(&stats.TotalIcons, &stats.ActiveIcons, &stats.InactiveIcons)
	if err != nil {
		return nil, fmt.Errorf("failed to get icon counts: %w", err)
	}

	// 获取分类统计
	categoryQuery := `
		SELECT category, COUNT(*) as count
		FROM sys_icons
		WHERE is_active = true AND category IS NOT NULL AND category != ''
		GROUP BY category
		ORDER BY count DESC
	`
	rows, err := s.db.Query(categoryQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category models.IconCategory
		err := rows.Scan(&category.Category, &category.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		stats.Categories = append(stats.Categories, category)
	}

	// 获取图标类型统计
	typeQuery := `
		SELECT icon_type, COUNT(*) as count
		FROM sys_icons
		WHERE is_active = true
		GROUP BY icon_type
	`
	typeRows, err := s.db.Query(typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get icon type stats: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var iconType string
		var count int
		err := typeRows.Scan(&iconType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan icon type: %w", err)
		}
		stats.IconTypes[iconType] = count
	}

	return stats, nil
}

// GetCategories 获取所有分类
func (s *IconService) GetCategories() ([]string, error) {
	query := `
		SELECT DISTINCT category
		FROM sys_icons
		WHERE is_active = true AND category IS NOT NULL AND category != ''
		ORDER BY category
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}
