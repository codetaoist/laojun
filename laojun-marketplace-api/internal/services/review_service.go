package services

import (
	"database/sql"
	"fmt"
	"time"

	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// ReviewService 评论服务
type ReviewService struct {
	db *shareddb.DB
}

// NewReviewService 创建评论服务
func NewReviewService(db *shareddb.DB) *ReviewService {
	return &ReviewService{db: db}
}

// Review 评论模型
type Review struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	PluginID  uuid.UUID `json:"plugin_id" db:"plugin_id"`
	Rating    int       `json:"rating" db:"rating"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	User *models.User `json:"user,omitempty"`
}

// ReviewStats 评论统计信息
type ReviewStats struct {
	TotalReviews  int     `json:"total_reviews"`
	AverageRating float64 `json:"average_rating"`
	FiveStar      int     `json:"five_star"`
	FourStar      int     `json:"four_star"`
	ThreeStar     int     `json:"three_star"`
	TwoStar       int     `json:"two_star"`
	OneStar       int     `json:"one_star"`
}

// GetPluginReviews 获取插件的评论列�?
func (s *ReviewService) GetPluginReviews(pluginID uuid.UUID, params models.ReviewSearchParams) ([]models.Review, models.PaginationMeta, error) {
	var reviews []models.Review
	var totalCount int

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM mp_reviews WHERE plugin_id = $1"
	err := s.db.QueryRow(countQuery, pluginID).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 计算偏移�?
	offset := (params.Page - 1) * params.Limit

	// 构建排序
	orderBy := "r.created_at DESC"
	if params.SortBy != "" {
		switch params.SortBy {
		case "rating":
			orderBy = "r.rating DESC"
		case "created_at":
			orderBy = "r.created_at DESC"
		case "helpful":
			orderBy = "r.helpful_count DESC"
		}
	}

	// 获取评论列表
	query := fmt.Sprintf(`
		SELECT 
			r.id, r.user_id, r.plugin_id, r.rating, r.content, 
			r.helpful_count, r.created_at, r.updated_at,
			u.username, u.email
		FROM mp_reviews r
		JOIN mp_users u ON r.user_id = u.id
		WHERE r.plugin_id = $1
		ORDER BY %s
		LIMIT $2 OFFSET $3`, orderBy)

	rows, err := s.db.Query(query, pluginID, params.Limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var review models.Review
		var username, email string

		err := rows.Scan(
			&review.ID, &review.UserID, &review.PluginID, &review.Rating,
			&review.Content, &review.HelpfulCount, &review.CreatedAt, &review.UpdatedAt,
			&username, &email,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 设置用户信息
		if review.UserID != nil {
			review.User = &models.User{
				ID:       *review.UserID,
				Username: username,
				Email:    email,
			}
		}

		reviews = append(reviews, review)
	}

	// 计算分页信息
	totalPages := (totalCount + params.Limit - 1) / params.Limit

	meta := models.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return reviews, meta, nil
}

// CreateReview 创建评论
func (s *ReviewService) CreateReview(userID, pluginID uuid.UUID, rating int, comment string) (*Review, error) {
	// 检查用户是否已经评论过这个插件
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM mp_reviews WHERE user_id = $1 AND plugin_id = $2)"
	err := s.db.QueryRow(checkQuery, userID, pluginID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check review existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("user has already reviewed this plugin")
	}

	// 创建评论
	reviewID := uuid.New()
	insertQuery := `
		INSERT INTO mp_reviews (id, user_id, plugin_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, user_id, plugin_id, rating, comment, created_at, updated_at`

	var review Review
	err = s.db.QueryRow(insertQuery, reviewID, userID, pluginID, rating, comment).Scan(
		&review.ID, &review.UserID, &review.PluginID, &review.Rating,
		&review.Comment, &review.CreatedAt, &review.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// 更新插件的平均评�?
	err = s.updatePluginRating(pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to update plugin rating: %w", err)
	}

	return &review, nil
}

// GetReview 获取评论
func (s *ReviewService) GetReview(reviewID uuid.UUID) (*Review, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.plugin_id, r.rating, r.comment, r.created_at, r.updated_at,
			u.username, u.email,
			p.name as plugin_name
		FROM mp_reviews r
		JOIN mp_users u ON r.user_id = u.id
		JOIN mp_plugins p ON r.plugin_id = p.id
		WHERE r.id = $1`

	var review Review
	var user models.User
	var pluginName string
	err := s.db.QueryRow(query, reviewID).Scan(
		&review.ID, &review.UserID, &review.PluginID, &review.Rating,
		&review.Comment, &review.CreatedAt, &review.UpdatedAt,
		&user.Username, &user.Email, &pluginName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("review not found")
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	review.User = &user
	return &review, nil
}

// GetPluginReviewsList 获取插件的评论列表（简化版�?
func (s *ReviewService) GetPluginReviewsList(pluginID uuid.UUID, page, limit int) ([]Review, *models.PaginationMeta, error) {
	// 计算偏移�?
	offset := (page - 1) * limit

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM mp_reviews WHERE plugin_id = $1"
	err := s.db.QueryRow(countQuery, pluginID).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count reviews: %w", err)
	}

	// 获取评论列表，按创建时间降序排列
	query := `
		SELECT r.id, r.user_id, r.plugin_id, r.rating, r.comment, r.is_verified, 
		       r.created_at, r.updated_at, u.username, u.avatar_url
		FROM mp_reviews r
		JOIN mp_users u ON r.user_id = u.id
		WHERE r.plugin_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, pluginID, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		var user models.User

		err := rows.Scan(
			&review.ID, &review.UserID, &review.PluginID, &review.Rating,
			&review.Comment, &review.CreatedAt, &review.UpdatedAt,
			&user.Username, &user.Email,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan review: %w", err)
		}

		review.User = &user
		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate reviews: %w", err)
	}

	// 创建分页元数据
	meta := &models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	return reviews, meta, nil
}

// GetUserReviews 获取用户的评论列表（简化版）
func (s *ReviewService) GetUserReviews(userID uuid.UUID, page, limit int) ([]Review, *models.PaginationMeta, error) {
	// 计算偏移量
	offset := (page - 1) * limit

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM mp_reviews WHERE user_id = $1"
	err := s.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count reviews: %w", err)
	}

	// 获取评论列表
	query := `
		SELECT 
			r.id, r.user_id, r.plugin_id, r.rating, r.comment, r.created_at, r.updated_at,
			p.name as plugin_name, p.icon_url as plugin_icon
		FROM mp_reviews r
		JOIN mp_plugins p ON r.plugin_id = p.id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		var pluginName, pluginIcon string

		err := rows.Scan(
			&review.ID, &review.UserID, &review.PluginID, &review.Rating,
			&review.Comment, &review.CreatedAt, &review.UpdatedAt,
			&pluginName, &pluginIcon,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan review: %w", err)
		}

		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate reviews: %w", err)
	}

	// 创建分页元数据
	meta := &models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	return reviews, meta, nil
}

// UpdateReview 更新评论
func (s *ReviewService) UpdateReview(reviewID uuid.UUID, rating int, comment string) (*Review, error) {
	updateQuery := `
		UPDATE mp_reviews
		SET rating = $2, comment = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, plugin_id, rating, comment, created_at, updated_at`

	var review Review
	err := s.db.QueryRow(updateQuery, reviewID, rating, comment).Scan(
		&review.ID, &review.UserID, &review.PluginID, &review.Rating,
		&review.Comment, &review.CreatedAt, &review.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("review not found")
		}
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	// 更新插件的平均评分
	err = s.updatePluginRating(review.PluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to update plugin rating: %w", err)
	}

	return &review, nil
}

// DeleteReview 删除评论
func (s *ReviewService) DeleteReview(reviewID uuid.UUID) error {
	// 先获取评论信息以便更新插件评分
	var pluginID uuid.UUID
	selectQuery := "SELECT plugin_id FROM mp_reviews WHERE id = $1"
	err := s.db.QueryRow(selectQuery, reviewID).Scan(&pluginID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("review not found")
		}
		return fmt.Errorf("failed to get review: %w", err)
	}

	// 删除评论
	deleteQuery := "DELETE FROM mp_reviews WHERE id = $1"
	result, err := s.db.Exec(deleteQuery, reviewID)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	// 更新插件的平均评分
	err = s.updatePluginRating(pluginID)
	if err != nil {
		return fmt.Errorf("failed to update plugin rating: %w", err)
	}

	return nil
}

// updatePluginRating 更新插件的平均评分
func (s *ReviewService) updatePluginRating(pluginID uuid.UUID) error {
	query := `
		UPDATE mp_plugins
		SET rating = (
			SELECT COALESCE(AVG(rating), 0)
			FROM mp_reviews
			WHERE plugin_id = $1
		)
		WHERE id = $1`

	_, err := s.db.Exec(query, pluginID)
	if err != nil {
		return fmt.Errorf("failed to update plugin rating: %w", err)
	}

	return nil
}

// GetReviewStats 获取评论统计信息
func (s *ReviewService) GetReviewStats(pluginID uuid.UUID) (*ReviewStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_reviews,
			COALESCE(AVG(rating), 0) as average_rating,
			COUNT(CASE WHEN rating = 5 THEN 1 END) as five_star,
			COUNT(CASE WHEN rating = 4 THEN 1 END) as four_star,
			COUNT(CASE WHEN rating = 3 THEN 1 END) as three_star,
			COUNT(CASE WHEN rating = 2 THEN 1 END) as two_star,
			COUNT(CASE WHEN rating = 1 THEN 1 END) as one_star
		FROM mp_reviews
		WHERE plugin_id = $1`

	var stats ReviewStats
	err := s.db.QueryRow(query, pluginID).Scan(
		&stats.TotalReviews, &stats.AverageRating,
		&stats.FiveStar, &stats.FourStar, &stats.ThreeStar,
		&stats.TwoStar, &stats.OneStar,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get review stats: %w", err)
	}

	return &stats, nil
}
