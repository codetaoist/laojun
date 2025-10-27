package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun-plugins/internal/models"
	"github.com/codetaoist/laojun-shared/models" // 使用shared模型
)

// PluginReviewService 插件审核服务
type PluginReviewService struct {
	db *database.DB
}

// NewPluginReviewService 创建插件审核服务
func NewPluginReviewService(db *database.DB) *PluginReviewService {
	return &PluginReviewService{db: db}
}

// GetReviewQueue 获取审核队列
func (s *PluginReviewService) GetReviewQueue(params models.ReviewQueueParams) ([]models.Plugin, sharedmodels.PaginationMeta, error) {
	var plugins []models.Plugin
	var total int

	// 构建查询条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	// 状态过滤
	if len(params.Status) > 0 {
		placeholders := make([]string, len(params.Status))
		for i, status := range params.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, string(status))
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("p.review_status IN (%s)", strings.Join(placeholders, ",")))
	}

	// 优先级过滤
	if len(params.Priority) > 0 {
		placeholders := make([]string, len(params.Priority))
		for i, priority := range params.Priority {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, string(priority))
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("p.review_priority IN (%s)", strings.Join(placeholders, ",")))
	}

	// 分类过滤
	if params.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *params.CategoryID)
		argIndex++
	}

	// 审核员过滤
	if params.ReviewerID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.reviewer_id = $%d", argIndex))
		args = append(args, *params.ReviewerID)
		argIndex++
	}

	// 日期过滤
	if params.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.submitted_for_review_at >= $%d", argIndex))
		args = append(args, *params.DateFrom)
		argIndex++
	}

	if params.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.submitted_for_review_at <= $%d", argIndex))
		args = append(args, *params.DateTo)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 排序
	sortBy := "p.submitted_for_review_at"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	sortOrder := "DESC"
	if params.SortOrder != "" {
		sortOrder = params.SortOrder
	}

	// 分页
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	offset := (params.Page - 1) * params.Limit

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mp_plugins p 
		LEFT JOIN mp_categories c ON p.category_id = c.id 
		WHERE %s`, whereClause)

	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, sharedmodels.PaginationMeta{}, fmt.Errorf("failed to count plugins: %w", err)
	}

	// 查询数据
	query := fmt.Sprintf(`
		SELECT 
			p.id, p.name, p.description, p.author, p.category_id, p.icon_url as icon, 
			p.screenshots, p.tags, p.price, p.is_active,
			p.download_count as downloads, p.rating, p.rating_count as review_count, p.created_at, p.updated_at,
			p.review_status, p.review_priority, p.auto_review_score,
			p.auto_review_result, p.review_notes, p.reviewed_at,
			p.reviewer_id, p.submitted_for_review_at, p.rejection_reason,
			p.appeal_count, p.last_appeal_at,
			c.id as category_table_id, c.name as category_name, c.description as category_description,
			c.icon as category_icon
		FROM mp_plugins p 
		LEFT JOIN mp_categories c ON p.category_id = c.id 
		WHERE %s 
		ORDER BY %s %s 
		LIMIT $%d OFFSET $%d`,
		whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, params.Limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, sharedmodels.PaginationMeta{}, fmt.Errorf("failed to query plugins: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var plugin models.Plugin
		var category models.Category
		var screenshotsJSON, tagsJSON sql.NullString
		var autoReviewResult, reviewNotes, rejectionReason sql.NullString

		err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Author,
			&plugin.CategoryID, &plugin.Icon, &screenshotsJSON, &tagsJSON,
			&plugin.Price, &plugin.IsActive,
			&plugin.Downloads, &plugin.Rating, &plugin.ReviewCount,
			&plugin.CreatedAt, &plugin.UpdatedAt,
			// 审核相关字段
			&plugin.ReviewStatus, &plugin.ReviewPriority, &plugin.AutoReviewScore,
			&autoReviewResult, &reviewNotes, &plugin.ReviewedAt,
			&plugin.ReviewerID, &plugin.SubmittedForReviewAt, &rejectionReason,
			&plugin.AppealCount, &plugin.LastAppealAt,
			// 分类信息
			&category.ID, &category.Name, &category.Description,
			&category.Icon,
		)
		if err != nil {
			return nil, sharedmodels.PaginationMeta{}, fmt.Errorf("failed to scan plugin: %w", err)
		}

		// 处理可能为NULL的字符串字段
		if autoReviewResult.Valid {
			plugin.AutoReviewResult = &autoReviewResult.String
		}
		if reviewNotes.Valid {
			plugin.ReviewNotes = &reviewNotes.String
		}
		if rejectionReason.Valid {
			plugin.RejectionReason = &rejectionReason.String
		}

		// 解析JSON字段
		if screenshotsJSON.Valid {
			json.Unmarshal([]byte(screenshotsJSON.String), &plugin.Screenshots)
		}
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &plugin.Tags)
		}

		// 设置分类信息
		if category.ID != uuid.Nil {
			plugin.Category = &category
		}

		plugins = append(plugins, plugin)
	}

	meta := sharedmodels.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: (total + params.Limit - 1) / params.Limit,
	}

	return plugins, meta, nil
}

// AssignReviewer 分配审核员
func (s *PluginReviewService) AssignReviewer(pluginID, reviewerID uuid.UUID) error {
	query := `
		UPDATE mp_plugins 
		SET reviewer_id = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND review_status = 'pending'`

	result, err := s.db.Exec(query, reviewerID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to assign reviewer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plugin not found or not in pending status")
	}

	return nil
}

// ReviewPlugin 审核插件
func (s *PluginReviewService) ReviewPlugin(reviewerID uuid.UUID, request models.ReviewRequest) (*models.PluginReview, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取插件当前状态
	var currentStatus string
	err = tx.QueryRow("SELECT review_status FROM mp_plugins WHERE id = $1", request.PluginID).Scan(&currentStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin status: %w", err)
	}

	// 更新插件状态
	newStatus := string(request.ReviewResult)
	if request.ReviewResult == models.ReviewResultPass {
		newStatus = "approved"
	} else if request.ReviewResult == models.ReviewResultFail {
		newStatus = "rejected"
	}

	// 更新插件状态
	updateQuery := `
		UPDATE mp_plugins 
		SET review_status = $1, review_notes = $2, reviewed_at = CURRENT_TIMESTAMP, 
			reviewer_id = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4`

	_, err = tx.Exec(updateQuery, newStatus, request.ReviewNotes, reviewerID, request.PluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to update plugin: %w", err)
	}

	// 创建审核记录
	reviewID := uuid.New()
	checklistJSON, _ := json.Marshal(request.ReviewChecklist)

	insertReviewQuery := `
		INSERT INTO mp_plugin_reviews 
		(id, plugin_id, reviewer_id, previous_status, new_status, review_type, 
		 review_result, priority, review_notes, review_checklist, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = tx.Exec(insertReviewQuery,
		reviewID, request.PluginID, reviewerID, currentStatus, newStatus,
		models.ReviewTypeManual, request.ReviewResult, models.ReviewPriorityNormal,
		request.ReviewNotes, checklistJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create review record: %w", err)
	}

	// 更新审核员工作負載
	err = s.updateReviewerWorkload(tx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update reviewer workload: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 返回审核记录
	review := &models.PluginReview{
		ID:              reviewID,
		PluginID:        request.PluginID,
		ReviewerID:      &reviewerID,
		PreviousStatus:  currentStatus,
		NewStatus:       newStatus,
		ReviewType:      models.ReviewTypeManual,
		ReviewResult:    request.ReviewResult,
		Priority:        models.ReviewPriorityNormal,
		ReviewNotes:     request.ReviewNotes,
		ReviewChecklist: request.ReviewChecklist,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return review, nil
}

// BatchReviewPlugins 批量审核插件
func (s *PluginReviewService) BatchReviewPlugins(reviewerID uuid.UUID, request models.BatchReviewRequest) ([]models.PluginReview, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var reviews []models.PluginReview

	for _, pluginID := range request.PluginIDs {
		// 获取插件当前状态
		var currentStatus string
		err = tx.QueryRow("SELECT review_status FROM mp_plugins WHERE id = $1", pluginID).Scan(&currentStatus)
		if err != nil {
			continue // 跳过不存在的插件
		}

		// 更新插件状态
		newStatus := string(request.ReviewResult)
		if request.ReviewResult == models.ReviewResultPass {
			newStatus = "approved"
		} else if request.ReviewResult == models.ReviewResultFail {
			newStatus = "rejected"
		}

		// 更新插件状态
		updateQuery := `
			UPDATE mp_plugins 
			SET review_status = $1, review_notes = $2, reviewed_at = CURRENT_TIMESTAMP, 
				reviewer_id = $3, updated_at = CURRENT_TIMESTAMP
			WHERE id = $4`

		_, err = tx.Exec(updateQuery, newStatus, request.ReviewNotes, reviewerID, pluginID)
		if err != nil {
			continue // 跳过更新失败的插件
		}

		// 创建审核记录
		reviewID := uuid.New()
		checklistJSON, _ := json.Marshal(request.ReviewChecklist)

		insertReviewQuery := `
			INSERT INTO mp_plugin_reviews 
			(id, plugin_id, reviewer_id, previous_status, new_status, review_type, 
			 review_result, priority, review_notes, review_checklist, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

		_, err = tx.Exec(insertReviewQuery,
			reviewID, pluginID, reviewerID, currentStatus, newStatus,
			models.ReviewTypeManual, request.ReviewResult, models.ReviewPriorityNormal,
			request.ReviewNotes, checklistJSON)
		if err != nil {
			continue // 跳过创建记录失败的插件
		}

		review := models.PluginReview{
			ID:              reviewID,
			PluginID:        pluginID,
			ReviewerID:      &reviewerID,
			PreviousStatus:  currentStatus,
			NewStatus:       newStatus,
			ReviewType:      models.ReviewTypeManual,
			ReviewResult:    request.ReviewResult,
			Priority:        models.ReviewPriorityNormal,
			ReviewNotes:     request.ReviewNotes,
			ReviewChecklist: request.ReviewChecklist,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		reviews = append(reviews, review)
	}

	// 更新审核员工作負載
	err = s.updateReviewerWorkload(tx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update reviewer workload: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return reviews, nil
}

// GetPluginReviewHistory 获取插件审核历史
func (s *PluginReviewService) GetPluginReviewHistory(pluginID uuid.UUID) ([]models.PluginReview, error) {
	query := `
		SELECT 
			pr.id, pr.plugin_id, pr.reviewer_id, pr.previous_status, pr.new_status,
			pr.review_type, pr.review_result, pr.priority, pr.review_notes,
			pr.review_checklist, pr.auto_review_data, pr.created_at, pr.updated_at,
			u.id as reviewer_id, u.username as reviewer_username, u.email as reviewer_email
		FROM mp_plugin_reviews pr
		LEFT JOIN ua_admin u ON pr.reviewer_id = u.id
		WHERE pr.plugin_id = $1
		ORDER BY pr.created_at DESC`

	rows, err := s.db.Query(query, pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to query review history: %w", err)
	}
	defer rows.Close()

	var reviews []models.PluginReview
	for rows.Next() {
		var review models.PluginReview
		var reviewer models.User
		var reviewerID sql.NullString
		var checklistJSON, autoReviewJSON sql.NullString

		err := rows.Scan(
			&review.ID, &review.PluginID, &review.ReviewerID, &review.PreviousStatus,
			&review.NewStatus, &review.ReviewType, &review.ReviewResult, &review.Priority,
			&review.ReviewNotes, &checklistJSON, &autoReviewJSON, &review.CreatedAt,
			&review.UpdatedAt, &reviewerID, &reviewer.Username, &reviewer.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}

		// 解析JSON字段
		if checklistJSON.Valid {
			json.Unmarshal([]byte(checklistJSON.String), &review.ReviewChecklist)
		}
		if autoReviewJSON.Valid {
			json.Unmarshal([]byte(autoReviewJSON.String), &review.AutoReviewData)
		}

		// 设置审核员信息
		if reviewerID.Valid {
			review.Reviewer = &reviewer
		}

		reviews = append(reviews, review)
	}

	return reviews, nil
}

// CreateAppeal 创建申诉
func (s *PluginReviewService) CreateAppeal(developerID uuid.UUID, request models.AppealRequest) (*models.DeveloperAppeal, error) {
	// 检查插件是否属于该开发人员
	var pluginDeveloperID uuid.UUID
	err := s.db.QueryRow("SELECT developer_id FROM mp_plugins WHERE id = $1", request.PluginID).Scan(&pluginDeveloperID)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	if pluginDeveloperID != developerID {
		return nil, fmt.Errorf("plugin does not belong to this developer")
	}

	// 创建申诉记录
	appealID := uuid.New()
	query := `
		INSERT INTO mp_developer_appeals 
		(id, plugin_id, developer_id, review_id, appeal_type, appeal_reason, 
		 appeal_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = s.db.Exec(query, appealID, request.PluginID, developerID,
		request.ReviewID, request.AppealType, request.AppealReason)
	if err != nil {
		return nil, fmt.Errorf("failed to create appeal: %w", err)
	}

	// 更新插件申诉计数
	updateQuery := `
		UPDATE mp_plugins 
		SET appeal_count = appeal_count + 1, last_appeal_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err = s.db.Exec(updateQuery, request.PluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to update plugin appeal count: %w", err)
	}

	appeal := &models.DeveloperAppeal{
		ID:           appealID,
		PluginID:     request.PluginID,
		DeveloperID:  developerID,
		ReviewID:     request.ReviewID,
		AppealType:   request.AppealType,
		AppealReason: request.AppealReason,
		AppealStatus: "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return appeal, nil
}

// ProcessAppeal 处理申诉
func (s *PluginReviewService) ProcessAppeal(adminID uuid.UUID, request models.AppealProcessRequest) (*models.DeveloperAppeal, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 更新申诉状态
	status := "rejected"
	if request.Approved {
		status = "approved"
	}

	updateQuery := `
		UPDATE mp_developer_appeals 
		SET appeal_status = $1, admin_id = $2, admin_reply = $3, 
			processed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4`

	_, err = tx.Exec(updateQuery, status, adminID, request.AdminReply, request.AppealID)
	if err != nil {
		return nil, fmt.Errorf("failed to update appeal: %w", err)
	}

	// 如果申诉通过，更新插件状态为待审核
	if request.Approved {
		var pluginID uuid.UUID
		err = tx.QueryRow("SELECT plugin_id FROM mp_developer_appeals WHERE id = $1", request.AppealID).Scan(&pluginID)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin ID: %w", err)
		}

		_, err = tx.Exec("UPDATE mp_plugins SET review_status = 'pending' WHERE id = $1", pluginID)
		if err != nil {
			return nil, fmt.Errorf("failed to update plugin status: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 获取更新后的申诉记录
	appeal, err := s.GetAppeal(request.AppealID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated appeal: %w", err)
	}

	return appeal, nil
}

// GetAppeal 获取申诉详情
func (s *PluginReviewService) GetAppeal(appealID uuid.UUID) (*models.DeveloperAppeal, error) {
	query := `
		SELECT 
			da.id, da.plugin_id, da.developer_id, da.review_id, da.appeal_type,
			da.appeal_reason, da.appeal_status, da.admin_id, da.admin_reply,
			da.processed_at, da.created_at, da.updated_at,
			p.name as plugin_name,
			d.username as developer_username,
			a.username as admin_username
		FROM mp_developer_appeals da
		LEFT JOIN mp_plugins p ON da.plugin_id = p.id
		LEFT JOIN ua_admin d ON da.developer_id = d.id
		LEFT JOIN ua_admin a ON da.admin_id = a.id
		WHERE da.id = $1`

	var appeal models.DeveloperAppeal
	var plugin models.Plugin
	var developer, admin models.User
	var adminID sql.NullString
	var processedAt sql.NullTime

	err := s.db.QueryRow(query, appealID).Scan(
		&appeal.ID, &appeal.PluginID, &appeal.DeveloperID, &appeal.ReviewID,
		&appeal.AppealType, &appeal.AppealReason, &appeal.AppealStatus,
		&appeal.AdminID, &appeal.AdminReply, &processedAt,
		&appeal.CreatedAt, &appeal.UpdatedAt,
		&plugin.Name, &developer.Username, &admin.Username,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get appeal: %w", err)
	}

	if processedAt.Valid {
		appeal.ProcessedAt = &processedAt.Time
	}

	appeal.Plugin = &plugin
	appeal.Developer = &developer
	if adminID.Valid {
		appeal.Admin = &admin
	}

	return &appeal, nil
}

// GetAppeals 获取申诉列表
func (s *PluginReviewService) GetAppeals(params models.AppealListParams) ([]models.DeveloperAppeal, sharedmodels.PaginationMeta, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 状态过滤
	if len(params.Status) > 0 {
		placeholders := make([]string, len(params.Status))
		for i, status := range params.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("da.appeal_status IN (%s)", strings.Join(placeholders, ",")))
	}

	// 插件ID过滤
	if params.PluginID != nil {
		conditions = append(conditions, fmt.Sprintf("da.plugin_id = $%d", argIndex))
		args = append(args, *params.PluginID)
		argIndex++
	}

	// 开发者ID过滤
	if params.DeveloperID != nil {
		conditions = append(conditions, fmt.Sprintf("da.developer_id = $%d", argIndex))
		args = append(args, *params.DeveloperID)
		argIndex++
	}

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 计算总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mp_developer_appeals da
		%s`, whereClause)

	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, sharedmodels.PaginationMeta{}, fmt.Errorf("failed to count appeals: %w", err)
	}

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "DESC"
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.Limit

	// 构建主查询
	query := fmt.Sprintf(`
		SELECT 
			da.id, da.plugin_id, da.developer_id, da.review_id, da.appeal_type,
			da.appeal_reason, da.appeal_status, da.admin_id, da.admin_reply,
			da.processed_at, da.created_at, da.updated_at,
			p.name as plugin_name,
			d.username as developer_username,
			a.username as admin_username
		FROM mp_developer_appeals da
		LEFT JOIN mp_plugins p ON da.plugin_id = p.id
		LEFT JOIN ua_admin d ON da.developer_id = d.id
		LEFT JOIN ua_admin a ON da.admin_id = a.id
		%s
		ORDER BY da.%s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, params.SortBy, params.SortOrder, argIndex, argIndex+1)

	args = append(args, params.Limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, sharedmodels.PaginationMeta{}, fmt.Errorf("failed to query appeals: %w", err)
	}
	defer rows.Close()

	var appeals []models.DeveloperAppeal
	for rows.Next() {
		var appeal models.DeveloperAppeal
		var plugin models.Plugin
		var developer, admin models.User
		var adminID sql.NullString
		var adminUsername sql.NullString
		var processedAt sql.NullTime

		err := rows.Scan(
			&appeal.ID, &appeal.PluginID, &appeal.DeveloperID, &appeal.ReviewID,
			&appeal.AppealType, &appeal.AppealReason, &appeal.AppealStatus,
			&adminID, &appeal.AdminReply, &processedAt,
			&appeal.CreatedAt, &appeal.UpdatedAt,
			&plugin.Name, &developer.Username, &adminUsername,
		)
		if err != nil {
			return nil, sharedmodels.PaginationMeta{}, fmt.Errorf("failed to scan appeal: %w", err)
		}

		if processedAt.Valid {
			appeal.ProcessedAt = &processedAt.Time
		}

		appeal.Plugin = &plugin
		appeal.Developer = &developer
		if adminID.Valid && adminUsername.Valid {
			admin.Username = adminUsername.String
			appeal.Admin = &admin
		}

		appeals = append(appeals, appeal)
	}

	if err = rows.Err(); err != nil {
		return nil, sharedmodels.PaginationMeta{}, fmt.Errorf("failed to iterate appeals: %w", err)
	}

	// 计算分页信息
	totalPages := (total + params.Limit - 1) / params.Limit
	meta := sharedmodels.PaginationMeta{
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	return appeals, meta, nil
}

// GetReviewStats 获取审核统计
func (s *PluginReviewService) GetReviewStats() (*models.ReviewStats, error) {
	stats := &models.ReviewStats{
		StatusBreakdown:   make(map[models.PluginReviewStatus]int),
		PriorityBreakdown: make(map[models.ReviewPriority]int),
		CategoryBreakdown: make(map[string]int),
		ReviewerWorkload:  make(map[string]models.ReviewerWorkload),
	}

	// 基础统计
	err := s.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN review_status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN review_status = 'approved' THEN 1 END) as approved,
			COUNT(CASE WHEN review_status = 'rejected' THEN 1 END) as rejected,
			COUNT(CASE WHEN review_status = 'suspended' THEN 1 END) as suspended
		FROM mp_plugins`).Scan(
		&stats.TotalPlugins, &stats.PendingReviews, &stats.ApprovedPlugins,
		&stats.RejectedPlugins, &stats.SuspendedPlugins)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	// 状态分布
	statusRows, err := s.db.Query("SELECT review_status, COUNT(*) FROM mp_plugins GROUP BY review_status")
	if err != nil {
		return nil, fmt.Errorf("failed to get status breakdown: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		statusRows.Scan(&status, &count)
		stats.StatusBreakdown[models.PluginReviewStatus(status)] = count
	}

	// 优先级分布
	priorityRows, err := s.db.Query("SELECT review_priority, COUNT(*) FROM mp_plugins WHERE review_priority IS NOT NULL GROUP BY review_priority")
	if err != nil {
		return nil, fmt.Errorf("failed to get priority breakdown: %w", err)
	}
	defer priorityRows.Close()

	for priorityRows.Next() {
		var priority string
		var count int
		priorityRows.Scan(&priority, &count)
		stats.PriorityBreakdown[models.ReviewPriority(priority)] = count
	}

	// 分类分布
	categoryRows, err := s.db.Query(`
		SELECT c.name, COUNT(p.id) 
		FROM mp_categories c 
		LEFT JOIN mp_plugins p ON c.id = p.category_id 
		GROUP BY c.name`)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown: %w", err)
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var category string
		var count int
		categoryRows.Scan(&category, &count)
		stats.CategoryBreakdown[category] = count
	}

	return stats, nil
}

// updateReviewerWorkload 更新审核员工作負載
func (s *PluginReviewService) updateReviewerWorkload(tx *sql.Tx, reviewerID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)

	// 检查今天的记录是否存在
	var exists bool
	err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM mp_reviewer_workload WHERE reviewer_id = $1 AND date = $2)",
		reviewerID, today).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// 更新现有记录
		_, err = tx.Exec(`
			UPDATE mp_reviewer_workload 
			SET reviews_completed = reviews_completed + 1, updated_at = CURRENT_TIMESTAMP 
			WHERE reviewer_id = $1 AND date = $2`, reviewerID, today)
	} else {
		// 创建新记录
		_, err = tx.Exec(`
			INSERT INTO mp_reviewer_workload 
			(id, reviewer_id, date, reviews_assigned, reviews_completed, reviews_pending, 
			 average_review_time, created_at, updated_at)
			VALUES ($1, $2, $3, 0, 1, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			uuid.New(), reviewerID, today)
	}

	return err
}

// AutoReview 自动审核插件
func (s *PluginReviewService) AutoReview(pluginID uuid.UUID) (*models.AutoReviewLog, error) {
	// 这里实现自动审核逻辑
	// 包括安全扫描、代码质量检测、性能检测等

	// 示例实现
	log := &models.AutoReviewLog{
		ID:          uuid.New(),
		PluginID:    pluginID,
		ReviewType:  "security_scan",
		CheckResult: models.ReviewResultPass,
		Score:       85.5,
		Details: map[string]interface{}{
			"security_score":    90,
			"quality_score":     85,
			"performance_score": 82,
		},
		CreatedAt: time.Now(),
	}

	// 保存自动审核日志
	detailsJSON, _ := json.Marshal(log.Details)
	query := `
		INSERT INTO mp_auto_review_logs 
		(id, plugin_id, review_type, check_result, score, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.Exec(query, log.ID, log.PluginID, log.ReviewType,
		log.CheckResult, log.Score, detailsJSON, log.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save auto review log: %w", err)
	}

	// 更新插件的自动审核结果
	updateQuery := `
		UPDATE mp_plugins 
		SET auto_review_score = $1, auto_review_result = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3`
	_, err = s.db.Exec(updateQuery, log.Score, log.CheckResult, pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to update plugin auto review result: %w", err)
	}

	return log, nil
}
