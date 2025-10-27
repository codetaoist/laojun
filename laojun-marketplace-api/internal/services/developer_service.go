package services

import (
	"database/sql"
	"fmt"
	"time"

	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// DeveloperService 开发者服务
type DeveloperService struct {
	db *shareddb.DB
}

// NewDeveloperService 创建开发者服务
func NewDeveloperService(db *shareddb.DB) *DeveloperService {
	return &DeveloperService{db: db}
}

// Developer 开发者模型
type Developer struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	CompanyName *string   `json:"company_name" db:"company_name"`
	Website     *string   `json:"website" db:"website"`
	Description *string   `json:"description" db:"description"`
	IsVerified  bool      `json:"is_verified" db:"is_verified"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	User *models.User `json:"user,omitempty"`
}

// DeveloperStats 开发者统计信息
type DeveloperStats struct {
	TotalPlugins   int     `json:"total_plugins"`
	ActivePlugins  int     `json:"active_plugins"`
	TotalDownloads int     `json:"total_downloads"`
	TotalRevenue   float64 `json:"total_revenue"`
	AverageRating  float64 `json:"average_rating"`
	TotalReviews   int     `json:"total_reviews"`
}

// RegisterDeveloper 注册开发者
func (s *DeveloperService) RegisterDeveloper(userID uuid.UUID, companyName, website, description *string) (*Developer, error) {
	// 检查用户是否已经是开发者
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM mp_developers WHERE user_id = $1)"
	err := s.db.QueryRow(checkQuery, userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check developer existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("user is already a developer")
	}

	// 创建开发者记录
	developerID := uuid.New()
	insertQuery := `
		INSERT INTO mp_developers (id, user_id, company_name, website, description, is_verified, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, false, true, NOW(), NOW())
		RETURNING id, user_id, company_name, website, description, is_verified, is_active, created_at, updated_at`

	var developer Developer
	err = s.db.QueryRow(insertQuery, developerID, userID, companyName, website, description).Scan(
		&developer.ID, &developer.UserID, &developer.CompanyName, &developer.Website,
		&developer.Description, &developer.IsVerified, &developer.IsActive,
		&developer.CreatedAt, &developer.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create developer: %w", err)
	}

	return &developer, nil
}

// GetDeveloper 获取开发者信息
func (s *DeveloperService) GetDeveloper(developerID uuid.UUID) (*Developer, error) {
	query := `
		SELECT d.id, d.user_id, d.company_name, d.website, d.description,
			   d.is_verified, d.is_active, d.created_at, d.updated_at,
			   u.username, u.email
		FROM mp_developers d
		JOIN mp_users u ON d.user_id = u.id
		WHERE d.id = $1`

	var developer Developer
	var user models.User
	err := s.db.QueryRow(query, developerID).Scan(
		&developer.ID, &developer.UserID, &developer.CompanyName, &developer.Website,
		&developer.Description, &developer.IsVerified, &developer.IsActive,
		&developer.CreatedAt, &developer.UpdatedAt,
		&user.Username, &user.Email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("developer not found")
		}
		return nil, fmt.Errorf("failed to get developer: %w", err)
	}

	developer.User = &user
	return &developer, nil
}

// GetDeveloperByUserID 根据用户ID获取开发者信息
func (s *DeveloperService) GetDeveloperByUserID(userID uuid.UUID) (*Developer, error) {
	query := `
		SELECT 
			d.id, d.user_id, d.company_name, d.website, d.description, 
			d.is_verified, d.is_active, d.created_at, d.updated_at,
			u.username, u.email
		FROM mp_developers d
		JOIN mp_users u ON d.user_id = u.id
		WHERE d.user_id = $1`

	var developer Developer
	var user models.User
	err := s.db.QueryRow(query, userID).Scan(
		&developer.ID, &developer.UserID, &developer.CompanyName, &developer.Website,
		&developer.Description, &developer.IsVerified, &developer.IsActive,
		&developer.CreatedAt, &developer.UpdatedAt,
		&user.Username, &user.Email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("developer not found")
		}
		return nil, fmt.Errorf("failed to get developer: %w", err)
	}

	developer.User = &user
	return &developer, nil
}

// UpdateDeveloper 更新开发者信息
func (s *DeveloperService) UpdateDeveloper(developerID uuid.UUID, companyName, website, description *string) (*Developer, error) {
	updateQuery := `
		UPDATE mp_developers
		SET company_name = $2, website = $3, description = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, company_name, website, description, is_verified, is_active, created_at, updated_at`

	var developer Developer
	err := s.db.QueryRow(updateQuery, developerID, companyName, website, description).Scan(
		&developer.ID, &developer.UserID, &developer.CompanyName, &developer.Website,
		&developer.Description, &developer.IsVerified, &developer.IsActive,
		&developer.CreatedAt, &developer.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("developer not found")
		}
		return nil, fmt.Errorf("failed to update developer: %w", err)
	}

	return &developer, nil
}

// GetDeveloperStats 获取开发者统计信息
func (s *DeveloperService) GetDeveloperStats(developerID uuid.UUID) (*DeveloperStats, error) {
	query := `
		SELECT 
			d.id, d.company_name, d.website, d.description, d.is_verified,
			COUNT(p.id) as total_plugins,
			COUNT(CASE WHEN p.is_active = true THEN 1 END) as active_plugins,
			COALESCE(SUM(pu.amount), 0) as total_revenue,
			COUNT(DISTINCT pu.id) as total_sales,
			COALESCE(AVG(r.rating), 0) as avg_rating,
			COUNT(r.id) as total_reviews
		FROM mp_developers d
		LEFT JOIN mp_plugins p ON d.id = p.developer_id
		LEFT JOIN mp_purchases pu ON p.id = pu.plugin_id AND pu.status = 'completed'
		LEFT JOIN mp_reviews r ON p.id = r.plugin_id
		WHERE d.id = $1
		GROUP BY d.id, d.company_name, d.website, d.description, d.is_verified`

	var stats DeveloperStats
	var id uuid.UUID
	var companyName, website, description sql.NullString
	var isVerified bool
	var totalSales int

	err := s.db.QueryRow(query, developerID).Scan(
		&id, &companyName, &website, &description, &isVerified,
		&stats.TotalPlugins, &stats.ActivePlugins, &stats.TotalRevenue,
		&totalSales, &stats.AverageRating, &stats.TotalReviews,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有插件，返回零值统�?
			return &DeveloperStats{}, nil
		}
		return nil, fmt.Errorf("failed to get developer stats: %w", err)
	}

	return &stats, nil
}

// GetDeveloperPlugins 获取开发者的插件列表
func (s *DeveloperService) GetDeveloperPlugins(developerID uuid.UUID, page, limit int) ([]models.Plugin, *models.PaginationMeta, error) {
	// 计算偏移�?
	offset := (page - 1) * limit

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM mp_plugins WHERE developer_id = $1"
	err := s.db.QueryRow(countQuery, developerID).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count plugins: %w", err)
	}

	// 获取插件列表
	query := `
		SELECT 
			p.id, p.name, p.description, p.short_description, p.author, p.developer_id, 
			p.version, p.icon_url, p.price, p.rating, p.download_count, p.is_featured, 
			p.created_at, p.updated_at, p.category_id, c.name as category_name, 
			c.icon as category_icon, c.color as category_color
		FROM mp_plugins p
		LEFT JOIN mp_categories c ON p.category_id = c.id
		WHERE p.developer_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, developerID, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query plugins: %w", err)
	}
	defer rows.Close()

	var plugins []models.Plugin
	for rows.Next() {
		var plugin models.Plugin
		var categoryName, categoryIcon, categoryColor sql.NullString

		err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.ShortDescription,
			&plugin.Author, &plugin.DeveloperID, &plugin.Version, &plugin.IconURL, &plugin.Price,
			&plugin.Rating, &plugin.DownloadCount, &plugin.IsFeatured, &plugin.CreatedAt, &plugin.UpdatedAt,
			&plugin.CategoryID, &categoryName, &categoryIcon, &categoryColor,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan plugin: %w", err)
		}

		plugins = append(plugins, plugin)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate plugins: %w", err)
	}

	// 创建分页元数�?
	meta := &models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	return plugins, meta, nil
}

// VerifyDeveloper 验证开发者（管理员操作）
func (s *DeveloperService) VerifyDeveloper(developerID uuid.UUID) error {
	updateQuery := "UPDATE mp_developers SET is_verified = true, updated_at = NOW() WHERE id = $1"
	result, err := s.db.Exec(updateQuery, developerID)
	if err != nil {
		return fmt.Errorf("failed to verify developer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("developer not found")
	}

	return nil
}

// DeactivateDeveloper 停用开发者（管理员操作）
func (s *DeveloperService) DeactivateDeveloper(developerID uuid.UUID) error {
	updateQuery := "UPDATE mp_developers SET is_active = false, updated_at = NOW() WHERE id = $1"
	result, err := s.db.Exec(updateQuery, developerID)
	if err != nil {
		return fmt.Errorf("failed to deactivate developer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("developer not found")
	}

	return nil
}

// GetDevelopers 获取开发者列表（管理员功能）
func (s *DeveloperService) GetDevelopers(page, limit int, verified *bool) ([]Developer, *models.PaginationMeta, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if verified != nil {
		whereClause += fmt.Sprintf(" AND d.is_verified = $%d", argIndex)
		args = append(args, *verified)
		argIndex++
	}

	// 计算偏移�?
	offset := (page - 1) * limit

	// 获取总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM mp_developers d %s", whereClause)
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count developers: %w", err)
	}

	// 获取开发者列�?
	query := fmt.Sprintf(`
		SELECT 
			d.id, d.user_id, d.company_name, d.website, d.description, 
			d.is_verified, d.is_active, d.created_at, d.updated_at,
			u.username, u.email
		FROM mp_developers d
		JOIN mp_users u ON d.user_id = u.id
		%s
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query developers: %w", err)
	}
	defer rows.Close()

	var developers []Developer
	for rows.Next() {
		var developer Developer
		var user models.User

		err := rows.Scan(
			&developer.ID, &developer.UserID, &developer.CompanyName, &developer.Website,
			&developer.Description, &developer.IsVerified, &developer.IsActive,
			&developer.CreatedAt, &developer.UpdatedAt,
			&user.Username, &user.Email,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan developer: %w", err)
		}

		developer.User = &user
		developers = append(developers, developer)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate developers: %w", err)
	}

	// 创建分页元数�?
	meta := &models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	return developers, meta, nil
}
