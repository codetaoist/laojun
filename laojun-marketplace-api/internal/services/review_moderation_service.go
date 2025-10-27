package services

import (
	"database/sql"
	"fmt"
	"time"

	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/logger"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// ModerationStatus 审核状态枚举
type ModerationStatus string

const (
	ModerationStatusPending  ModerationStatus = "pending"  // 待审核
	ModerationStatusApproved ModerationStatus = "approved" // 已通过
	ModerationStatusRejected ModerationStatus = "rejected" // 已拒绝
	ModerationStatusHidden   ModerationStatus = "hidden"   // 已隐藏
	ModerationStatusFlagged  ModerationStatus = "flagged"  // 已标记
)

// ModerationAction 审核动作枚举
type ModerationAction string

const (
	ModerationActionApprove ModerationAction = "approve" // 通过
	ModerationActionReject  ModerationAction = "reject"  // 拒绝
	ModerationActionHide    ModerationAction = "hide"    // 隐藏
	ModerationActionFlag    ModerationAction = "flag"    // 标记
	ModerationActionRestore ModerationAction = "restore" // 恢复
)

// ReviewModeration 评价审核记录
type ReviewModeration struct {
	ID         uuid.UUID        `json:"id" db:"id"`
	ReviewID   uuid.UUID        `json:"review_id" db:"review_id"`
	ModeratorID uuid.UUID       `json:"moderator_id" db:"moderator_id"`
	Action     ModerationAction `json:"action" db:"action"`
	Status     ModerationStatus `json:"status" db:"status"`
	Reason     string           `json:"reason" db:"reason"`
	Note       *string          `json:"note" db:"note"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	
	// 关联数据
	Review    *models.Review `json:"review,omitempty"`
	Moderator *models.User   `json:"moderator,omitempty"`
}

// ReviewFlag 评价举报
type ReviewFlag struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ReviewID  uuid.UUID `json:"review_id" db:"review_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Reason    string    `json:"reason" db:"reason"`
	Category  string    `json:"category" db:"category"` // spam, inappropriate, fake, etc.
	Note      *string   `json:"note" db:"note"`
	Status    string    `json:"status" db:"status"`     // pending, resolved, dismissed
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	
	// 关联数据
	Review *models.Review `json:"review,omitempty"`
	User   *models.User   `json:"user,omitempty"`
}

// ModerationRequest 审核请求
type ModerationRequest struct {
	ReviewID uuid.UUID        `json:"review_id" binding:"required"`
	Action   ModerationAction `json:"action" binding:"required"`
	Reason   string           `json:"reason" binding:"required"`
	Note     string           `json:"note"`
}

// FlagRequest 举报请求
type FlagRequest struct {
	ReviewID uuid.UUID `json:"review_id" binding:"required"`
	Reason   string    `json:"reason" binding:"required"`
	Category string    `json:"category" binding:"required"`
	Note     string    `json:"note"`
}

// ModerationStats 审核统计
type ModerationStats struct {
	PendingReviews  int `json:"pending_reviews"`
	ApprovedReviews int `json:"approved_reviews"`
	RejectedReviews int `json:"rejected_reviews"`
	HiddenReviews   int `json:"hidden_reviews"`
	FlaggedReviews  int `json:"flagged_reviews"`
	TotalFlags      int `json:"total_flags"`
	PendingFlags    int `json:"pending_flags"`
}

// ReviewModerationService 评价审核服务
type ReviewModerationService struct {
	db *shareddb.DB
}

// NewReviewModerationService 创建评价审核服务
func NewReviewModerationService(db *shareddb.DB) *ReviewModerationService {
	return &ReviewModerationService{db: db}
}

// ModerateReview 审核评价
func (s *ReviewModerationService) ModerateReview(moderatorID uuid.UUID, req ModerationRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 检查评价是否存在
	var reviewExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM mp_reviews WHERE id = $1)", req.ReviewID).Scan(&reviewExists)
	if err != nil {
		return fmt.Errorf("failed to check review: %w", err)
	}
	
	if !reviewExists {
		return fmt.Errorf("review not found")
	}

	// 确定新状态
	var newStatus ModerationStatus
	switch req.Action {
	case ModerationActionApprove:
		newStatus = ModerationStatusApproved
	case ModerationActionReject:
		newStatus = ModerationStatusRejected
	case ModerationActionHide:
		newStatus = ModerationStatusHidden
	case ModerationActionFlag:
		newStatus = ModerationStatusFlagged
	case ModerationActionRestore:
		newStatus = ModerationStatusApproved
	default:
		return fmt.Errorf("invalid moderation action")
	}

	// 更新评价状态
	_, err = tx.Exec("UPDATE mp_reviews SET moderation_status = $1, updated_at = $2 WHERE id = $3",
		newStatus, time.Now(), req.ReviewID)
	if err != nil {
		return fmt.Errorf("failed to update review status: %w", err)
	}

	// 创建审核记录
	moderation := &ReviewModeration{
		ID:          uuid.New(),
		ReviewID:    req.ReviewID,
		ModeratorID: moderatorID,
		Action:      req.Action,
		Status:      newStatus,
		Reason:      req.Reason,
		CreatedAt:   time.Now(),
	}

	if req.Note != "" {
		moderation.Note = &req.Note
	}

	insertQuery := `
		INSERT INTO mp_review_moderations 
		(id, review_id, moderator_id, action, status, reason, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(insertQuery,
		moderation.ID, moderation.ReviewID, moderation.ModeratorID,
		moderation.Action, moderation.Status, moderation.Reason,
		moderation.Note, moderation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create moderation record: %w", err)
	}

	return tx.Commit()
}

// FlagReview 举报评价
func (s *ReviewModerationService) FlagReview(userID uuid.UUID, req FlagRequest) error {
	// 检查是否已经举报过
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM mp_review_flags WHERE review_id = $1 AND user_id = $2 AND status = 'pending')"
	err := s.db.QueryRow(checkQuery, req.ReviewID, userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existing flag: %w", err)
	}

	if exists {
		return fmt.Errorf("review already flagged by user")
	}

	// 检查评价是否存在
	var reviewExists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM mp_reviews WHERE id = $1)", req.ReviewID).Scan(&reviewExists)
	if err != nil {
		return fmt.Errorf("failed to check review: %w", err)
	}
	
	if !reviewExists {
		return fmt.Errorf("review not found")
	}

	// 创建举报记录
	flag := &ReviewFlag{
		ID:        uuid.New(),
		ReviewID:  req.ReviewID,
		UserID:    userID,
		Reason:    req.Reason,
		Category:  req.Category,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if req.Note != "" {
		flag.Note = &req.Note
	}

	insertQuery := `
		INSERT INTO mp_review_flags 
		(id, review_id, user_id, reason, category, note, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.Exec(insertQuery,
		flag.ID, flag.ReviewID, flag.UserID, flag.Reason,
		flag.Category, flag.Note, flag.Status, flag.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create flag: %w", err)
	}

	return nil
}

// GetPendingReviews 获取待审核评价列表
func (s *ReviewModerationService) GetPendingReviews(page, limit int) ([]models.Review, models.PaginationMeta, error) {
	var reviews []models.Review
	var totalCount int

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM mp_reviews WHERE moderation_status = 'pending'"
	err := s.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取待审核评价列表
	query := `
		SELECT r.id, r.user_id, r.plugin_id, r.rating, r.content,
		       r.moderation_status, r.created_at, r.updated_at,
		       u.username, u.email,
		       p.name as plugin_name
		FROM mp_reviews r
		LEFT JOIN mp_users u ON r.user_id = u.id
		LEFT JOIN mp_plugins p ON r.plugin_id = p.id
		WHERE r.moderation_status = 'pending'
		ORDER BY r.created_at ASC
		LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var review models.Review
		var username, email, pluginName sql.NullString

		err := rows.Scan(
			&review.ID, &review.UserID, &review.PluginID, &review.Rating,
			&review.Content, &review.ModerationStatus, &review.CreatedAt, &review.UpdatedAt,
			&username, &email, &pluginName,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 设置用户信息
		if username.Valid {
			review.User = &models.User{
				ID:       review.UserID,
				Username: username.String,
			}
			if email.Valid {
				review.User.Email = email.String
			}
		}

		// 设置插件名称
		if pluginName.Valid {
			review.PluginName = pluginName.String
		}

		reviews = append(reviews, review)
	}

	// 计算分页信息
	totalPages := (totalCount + limit - 1) / limit
	meta := models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return reviews, meta, nil
}

// GetFlaggedReviews 获取被举报的评价列表
func (s *ReviewModerationService) GetFlaggedReviews(page, limit int) ([]ReviewFlag, models.PaginationMeta, error) {
	var flags []ReviewFlag
	var totalCount int

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM mp_review_flags WHERE status = 'pending'"
	err := s.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取举报列表
	query := `
		SELECT f.id, f.review_id, f.user_id, f.reason, f.category,
		       f.note, f.status, f.created_at,
		       r.content as review_content, r.rating,
		       u.username
		FROM mp_review_flags f
		LEFT JOIN mp_reviews r ON f.review_id = r.id
		LEFT JOIN mp_users u ON f.user_id = u.id
		WHERE f.status = 'pending'
		ORDER BY f.created_at ASC
		LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var flag ReviewFlag
		var reviewContent sql.NullString
		var rating sql.NullInt32
		var username sql.NullString

		err := rows.Scan(
			&flag.ID, &flag.ReviewID, &flag.UserID, &flag.Reason,
			&flag.Category, &flag.Note, &flag.Status, &flag.CreatedAt,
			&reviewContent, &rating, &username,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 设置评价信息
		if reviewContent.Valid {
			flag.Review = &models.Review{
				ID:      flag.ReviewID,
				Content: reviewContent.String,
			}
			if rating.Valid {
				flag.Review.Rating = int(rating.Int32)
			}
		}

		// 设置用户信息
		if username.Valid {
			flag.User = &models.User{
				ID:       flag.UserID,
				Username: username.String,
			}
		}

		flags = append(flags, flag)
	}

	// 计算分页信息
	totalPages := (totalCount + limit - 1) / limit
	meta := models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return flags, meta, nil
}

// ResolveFlaggedReview 处理举报的评价
func (s *ReviewModerationService) ResolveFlaggedReview(flagID uuid.UUID, moderatorID uuid.UUID, action string, note string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取举报信息
	var flag ReviewFlag
	flagQuery := "SELECT review_id, user_id FROM mp_review_flags WHERE id = $1 AND status = 'pending'"
	err = tx.QueryRow(flagQuery, flagID).Scan(&flag.ReviewID, &flag.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("flag not found or already resolved")
		}
		return fmt.Errorf("failed to get flag: %w", err)
	}

	// 更新举报状态
	flagStatus := "resolved"
	if action == "dismiss" {
		flagStatus = "dismissed"
	}

	_, err = tx.Exec("UPDATE mp_review_flags SET status = $1, updated_at = $2 WHERE id = $3",
		flagStatus, time.Now(), flagID)
	if err != nil {
		return fmt.Errorf("failed to update flag status: %w", err)
	}

	// 如果需要对评价采取行动
	if action == "hide" || action == "reject" {
		var moderationAction ModerationAction
		if action == "hide" {
			moderationAction = ModerationActionHide
		} else {
			moderationAction = ModerationActionReject
		}

		// 创建审核记录
		modReq := ModerationRequest{
			ReviewID: flag.ReviewID,
			Action:   moderationAction,
			Reason:   "Flagged by users",
			Note:     note,
		}

		err = s.moderateReviewInTx(tx, moderatorID, modReq)
		if err != nil {
			return fmt.Errorf("failed to moderate review: %w", err)
		}
	}

	return tx.Commit()
}

// moderateReviewInTx 在事务中审核评价
func (s *ReviewModerationService) moderateReviewInTx(tx *sql.Tx, moderatorID uuid.UUID, req ModerationRequest) error {
	// 确定新状态
	var newStatus ModerationStatus
	switch req.Action {
	case ModerationActionApprove:
		newStatus = ModerationStatusApproved
	case ModerationActionReject:
		newStatus = ModerationStatusRejected
	case ModerationActionHide:
		newStatus = ModerationStatusHidden
	case ModerationActionFlag:
		newStatus = ModerationStatusFlagged
	case ModerationActionRestore:
		newStatus = ModerationStatusApproved
	default:
		return fmt.Errorf("invalid moderation action")
	}

	// 更新评价状态
	_, err := tx.Exec("UPDATE mp_reviews SET moderation_status = $1, updated_at = $2 WHERE id = $3",
		newStatus, time.Now(), req.ReviewID)
	if err != nil {
		return fmt.Errorf("failed to update review status: %w", err)
	}

	// 创建审核记录
	moderation := &ReviewModeration{
		ID:          uuid.New(),
		ReviewID:    req.ReviewID,
		ModeratorID: moderatorID,
		Action:      req.Action,
		Status:      newStatus,
		Reason:      req.Reason,
		CreatedAt:   time.Now(),
	}

	if req.Note != "" {
		moderation.Note = &req.Note
	}

	insertQuery := `
		INSERT INTO mp_review_moderations 
		(id, review_id, moderator_id, action, status, reason, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(insertQuery,
		moderation.ID, moderation.ReviewID, moderation.ModeratorID,
		moderation.Action, moderation.Status, moderation.Reason,
		moderation.Note, moderation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create moderation record: %w", err)
	}

	return nil
}

// GetModerationStats 获取审核统计
func (s *ReviewModerationService) GetModerationStats() (*ModerationStats, error) {
	stats := &ModerationStats{}

	// 获取评价审核统计
	reviewQuery := `
		SELECT 
			COUNT(CASE WHEN moderation_status = 'pending' THEN 1 END) as pending_reviews,
			COUNT(CASE WHEN moderation_status = 'approved' THEN 1 END) as approved_reviews,
			COUNT(CASE WHEN moderation_status = 'rejected' THEN 1 END) as rejected_reviews,
			COUNT(CASE WHEN moderation_status = 'hidden' THEN 1 END) as hidden_reviews,
			COUNT(CASE WHEN moderation_status = 'flagged' THEN 1 END) as flagged_reviews
		FROM mp_reviews`

	err := s.db.QueryRow(reviewQuery).Scan(
		&stats.PendingReviews, &stats.ApprovedReviews, &stats.RejectedReviews,
		&stats.HiddenReviews, &stats.FlaggedReviews,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get review stats: %w", err)
	}

	// 获取举报统计
	flagQuery := `
		SELECT 
			COUNT(*) as total_flags,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_flags
		FROM mp_review_flags`

	err = s.db.QueryRow(flagQuery).Scan(&stats.TotalFlags, &stats.PendingFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to get flag stats: %w", err)
	}

	return stats, nil
}

// GetModerationHistory 获取审核历史
func (s *ReviewModerationService) GetModerationHistory(reviewID uuid.UUID) ([]ReviewModeration, error) {
	var moderations []ReviewModeration

	query := `
		SELECT m.id, m.review_id, m.moderator_id, m.action, m.status,
		       m.reason, m.note, m.created_at,
		       u.username
		FROM mp_review_moderations m
		LEFT JOIN mp_users u ON m.moderator_id = u.id
		WHERE m.review_id = $1
		ORDER BY m.created_at DESC`

	rows, err := s.db.Query(query, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var moderation ReviewModeration
		var username sql.NullString

		err := rows.Scan(
			&moderation.ID, &moderation.ReviewID, &moderation.ModeratorID,
			&moderation.Action, &moderation.Status, &moderation.Reason,
			&moderation.Note, &moderation.CreatedAt, &username,
		)
		if err != nil {
			return nil, err
		}

		// 设置审核员信息
		if username.Valid {
			moderation.Moderator = &models.User{
				ID:       moderation.ModeratorID,
				Username: username.String,
			}
		}

		moderations = append(moderations, moderation)
	}

	return moderations, nil
}