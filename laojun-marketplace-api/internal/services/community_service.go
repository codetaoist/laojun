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

// CommunityService 社区服务
type CommunityService struct {
	db *shareddb.DB
}

// NewCommunityService 创建社区服务
func NewCommunityService(db *shareddb.DB) *CommunityService {
	return &CommunityService{db: db}
}

// ========== 论坛相关 ==========

// ForumSearchParams 论坛搜索参数
type ForumSearchParams struct {
	CategoryID *uuid.UUID `json:"category_id"`
	Query      string     `json:"query"`
	UserID     *uuid.UUID `json:"user_id"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	SortBy     string     `json:"sort_by"` // latest, popular, replies
}

// CreateForumPost 创建论坛帖子
func (s *CommunityService) CreateForumPost(post *models.ForumPost) error {
	query := `
		INSERT INTO mp_forum_posts (id, category_id, user_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	now := time.Now()
	post.ID = uuid.New()
	post.CreatedAt = now
	post.UpdatedAt = now

	_, err := s.db.Exec(query, post.ID, post.CategoryID, post.UserID, post.Title, post.Content, post.CreatedAt, post.UpdatedAt)
	return err
}

// GetForumPosts 获取论坛帖子列表
func (s *CommunityService) GetForumPosts(params ForumSearchParams) ([]models.ForumPost, models.PaginationMeta, error) {
	var posts []models.ForumPost
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

	if params.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.user_id = $%d", argIndex))
		args = append(args, *params.UserID)
		argIndex++
	}

	if params.Query != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(p.title ILIKE $%d OR p.content ILIKE $%d)", argIndex, argIndex+1))
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mp_forum_posts p 
		WHERE %s`, whereClause)

	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 构建排序
	orderBy := "p.created_at DESC"
	switch params.SortBy {
	case "popular":
		orderBy = "p.likes_count DESC, p.created_at DESC"
	case "replies":
		orderBy = "p.replies_count DESC, p.created_at DESC"
	}

	// 获取数据
	offset := (params.Page - 1) * params.Limit
	query := fmt.Sprintf(`
		SELECT id, category_id, user_id, title, content, 
		       COALESCE(likes_count, 0), COALESCE(replies_count, 0), COALESCE(views_count, 0), 
		       COALESCE(is_pinned, false), COALESCE(is_locked, false), 
		       created_at, updated_at
		FROM mp_forum_posts p
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
		var post models.ForumPost
		err := rows.Scan(
			&post.ID, &post.CategoryID, &post.UserID, &post.Title, &post.Content,
			&post.LikesCount, &post.RepliesCount, &post.ViewsCount, &post.IsPinned, &post.IsLocked,
			&post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}
		// 设置默认值
		post.Username = "匿名用户"
		post.CategoryName = "未分类"
		posts = append(posts, post)
	}

	meta := models.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      totalCount,
		TotalPages: (totalCount + params.Limit - 1) / params.Limit,
	}

	return posts, meta, nil
}

// GetForumPost 获取单个论坛帖子
func (s *CommunityService) GetForumPost(id uuid.UUID) (*models.ForumPost, error) {
	query := `
		SELECT p.id, p.category_id, p.user_id, p.title, p.content,
		       COALESCE(p.likes_count, 0), COALESCE(p.replies_count, 0), COALESCE(p.views_count, 0),
		       COALESCE(p.is_pinned, false), COALESCE(p.is_locked, false), p.created_at, p.updated_at,
		       COALESCE(u.username, '匿名用户') AS username, u.avatar,
		       COALESCE(c.name, '未分类') AS category_name
		FROM mp_forum_posts p
		LEFT JOIN mp_users u ON p.user_id = u.id
		LEFT JOIN mp_forum_categories c ON p.category_id = c.id
		WHERE p.id = $1`

	var post models.ForumPost
	var avatarNS sql.NullString
	err := s.db.QueryRow(query, id).Scan(
		&post.ID, &post.CategoryID, &post.UserID, &post.Title, &post.Content,
		&post.LikesCount, &post.RepliesCount, &post.ViewsCount, &post.IsPinned, &post.IsLocked,
		&post.CreatedAt, &post.UpdatedAt, &post.Username, &avatarNS, &post.CategoryName,
	)
	if err != nil {
		return nil, err
	}
	if avatarNS.Valid {
		post.AvatarURL = &avatarNS.String
	} else {
		post.AvatarURL = nil
	}

	// 增加浏览次数
	s.db.Exec("UPDATE mp_forum_posts SET views_count = views_count + 1 WHERE id = $1", id)

	return &post, nil
}

// CreateForumReply 创建论坛回复
func (s *CommunityService) CreateForumReply(reply *models.ForumReply) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 插入回复
	query := `
		INSERT INTO mp_forum_replies (id, post_id, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	now := time.Now()
	reply.ID = uuid.New()
	reply.CreatedAt = now
	reply.UpdatedAt = now

	_, err = tx.Exec(query, reply.ID, reply.PostID, reply.UserID, reply.Content, reply.CreatedAt, reply.UpdatedAt)
	if err != nil {
		return err
	}

	// 更新帖子回复数
	_, err = tx.Exec("UPDATE mp_forum_posts SET replies_count = replies_count + 1 WHERE id = $1", reply.PostID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetForumReplies 获取论坛回复列表
func (s *CommunityService) GetForumReplies(postID uuid.UUID, page, limit int) ([]models.ForumReply, models.PaginationMeta, error) {
	var replies []models.ForumReply
	var totalCount int

	// 获取总数
	err := s.db.QueryRow("SELECT COUNT(*) FROM mp_forum_replies WHERE post_id = $1", postID).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 获取数据
	offset := (page - 1) * limit
	query := `
		SELECT r.id, r.post_id, r.user_id, r.content, COALESCE(r.likes_count, 0), r.created_at, r.updated_at,
		       COALESCE(u.username, '匿名用户') AS username, u.avatar
		FROM mp_forum_replies r
		LEFT JOIN mp_users u ON r.user_id = u.id
		WHERE r.post_id = $1
		ORDER BY r.created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, postID, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var reply models.ForumReply
		var avatarNS sql.NullString
		err := rows.Scan(
			&reply.ID, &reply.PostID, &reply.UserID, &reply.Content, &reply.LikesCount,
			&reply.CreatedAt, &reply.UpdatedAt, &reply.Username, &avatarNS,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}
		if avatarNS.Valid {
			reply.AvatarURL = &avatarNS.String
		} else {
			reply.AvatarURL = nil
		}
		replies = append(replies, reply)
	}

	meta := models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: (totalCount + limit - 1) / limit,
	}

	return replies, meta, nil
}

// ========== 博客相关 ==========

// BlogSearchParams 博客搜索参数
type BlogSearchParams struct {
	CategoryID *uuid.UUID `json:"category_id"`
	Query      string     `json:"query"`
	UserID     *uuid.UUID `json:"user_id"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	SortBy     string     `json:"sort_by"` // latest, popular, views
}

// CreateBlogPost 创建博客文章
func (s *CommunityService) CreateBlogPost(post *models.BlogPost) error {
	query := `
		INSERT INTO mp_blog_posts (id, category_id, user_id, title, content, summary, cover_image, 
		                          tags, is_published, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	now := time.Now()
	post.ID = uuid.New()
	post.CreatedAt = now
	post.UpdatedAt = now

	_, err := s.db.Exec(query, post.ID, post.CategoryID, post.UserID, post.Title, post.Content,
		post.Summary, post.CoverImage, post.Tags, post.IsPublished, post.CreatedAt, post.UpdatedAt)
	return err
}

// GetBlogPosts 获取博客文章列表
func (s *CommunityService) GetBlogPosts(params BlogSearchParams) ([]models.BlogPost, models.PaginationMeta, error) {
	var posts []models.BlogPost
	var totalCount int

	// 构建查询条件
	whereConditions := []string{"p.is_published = true"}
	args := []interface{}{}
	argIndex := 1

	if params.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *params.CategoryID)
		argIndex++
	}

	if params.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.user_id = $%d", argIndex))
		args = append(args, *params.UserID)
		argIndex++
	}

	if params.Query != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(p.title ILIKE $%d OR p.content ILIKE $%d OR p.tags ILIKE $%d)", argIndex, argIndex+1, argIndex+2))
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		argIndex += 3
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mp_blog_posts p 
		WHERE %s`, whereClause)

	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 构建排序
	orderBy := "p.created_at DESC"
	switch params.SortBy {
	case "popular":
		orderBy = "p.likes_count DESC, p.created_at DESC"
	case "views":
		orderBy = "p.views_count DESC, p.created_at DESC"
	}

	// 获取数据
	offset := (params.Page - 1) * params.Limit
	query := fmt.Sprintf(`
		SELECT id, category_id, user_id, title, summary, cover_image, tags,
		       COALESCE(likes_count, 0), COALESCE(comments_count, 0), COALESCE(views_count, 0), 
		       created_at, updated_at
		FROM mp_blog_posts p
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
		var post models.BlogPost
		err := rows.Scan(
			&post.ID, &post.CategoryID, &post.UserID, &post.Title, &post.Summary, &post.CoverImage, &post.Tags,
			&post.LikesCount, &post.CommentsCount, &post.ViewsCount, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}
		// 设置默认值
		post.Username = "匿名用户"
		post.CategoryName = "未分类"
		posts = append(posts, post)
	}

	meta := models.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      totalCount,
		TotalPages: (totalCount + params.Limit - 1) / params.Limit,
	}

	return posts, meta, nil
}

// GetBlogPost 获取单个博客文章
func (s *CommunityService) GetBlogPost(id uuid.UUID) (*models.BlogPost, error) {
	query := `
		SELECT p.id, p.category_id, p.user_id, p.title, p.content, p.summary, p.cover_image, p.tags,
		       COALESCE(p.likes_count, 0), COALESCE(p.comments_count, 0), COALESCE(p.views_count, 0), p.created_at, p.updated_at,
		       COALESCE(u.username, '匿名用户') AS username, u.avatar,
		       COALESCE(c.name, '未分类') AS category_name
		FROM mp_blog_posts p
		LEFT JOIN mp_users u ON p.user_id = u.id
		LEFT JOIN mp_blog_categories c ON p.category_id = c.id
		WHERE p.id = $1 AND p.is_published = true`

	var post models.BlogPost
	var avatarNS sql.NullString
	var summaryNS sql.NullString
	var coverNS sql.NullString
	var tagsNS sql.NullString
	err := s.db.QueryRow(query, id).Scan(
		&post.ID, &post.CategoryID, &post.UserID, &post.Title, &post.Content, &summaryNS, &coverNS, &tagsNS,
		&post.LikesCount, &post.CommentsCount, &post.ViewsCount, &post.CreatedAt, &post.UpdatedAt,
		&post.Username, &avatarNS, &post.CategoryName,
	)
	if err != nil {
		return nil, err
	}
	if avatarNS.Valid {
		post.AvatarURL = &avatarNS.String
	} else {
		post.AvatarURL = nil
	}
	if summaryNS.Valid {
		post.Summary = &summaryNS.String
	} else {
		post.Summary = nil
	}
	if coverNS.Valid {
		post.CoverImage = &coverNS.String
	} else {
		post.CoverImage = nil
	}
	if tagsNS.Valid {
		post.Tags = &tagsNS.String
	} else {
		post.Tags = nil
	}

	// 增加浏览次数
	s.db.Exec("UPDATE mp_blog_posts SET views_count = views_count + 1 WHERE id = $1", id)

	return &post, nil
}

// ========== 代码分享相关 ==========

// CodeSearchParams 代码搜索参数
type CodeSearchParams struct {
	Language string     `json:"language"`
	Query    string     `json:"query"`
	UserID   *uuid.UUID `json:"user_id"`
	Page     int        `json:"page"`
	Limit    int        `json:"limit"`
	SortBy   string     `json:"sort_by"` // latest, popular, views
}

// CreateCodeSnippet 创建代码片段
func (s *CommunityService) CreateCodeSnippet(snippet *models.CodeSnippet) error {
	query := `
		INSERT INTO mp_code_snippets (id, user_id, title, description, code, language, tags, 
		                             is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	now := time.Now()
	snippet.ID = uuid.New()
	snippet.CreatedAt = now
	snippet.UpdatedAt = now

	_, err := s.db.Exec(query, snippet.ID, snippet.UserID, snippet.Title, snippet.Description,
		snippet.Code, snippet.Language, snippet.Tags, snippet.IsPublic, snippet.CreatedAt, snippet.UpdatedAt)
	return err
}

// GetCodeSnippets 获取代码片段列表
func (s *CommunityService) GetCodeSnippets(params CodeSearchParams) ([]models.CodeSnippet, models.PaginationMeta, error) {
	var snippets []models.CodeSnippet
	var totalCount int

	// 构建查询条件
	whereConditions := []string{"s.is_public = true"}
	args := []interface{}{}
	argIndex := 1

	if params.Language != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("s.language = $%d", argIndex))
		args = append(args, params.Language)
		argIndex++
	}

	if params.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("s.user_id = $%d", argIndex))
		args = append(args, *params.UserID)
		argIndex++
	}

	if params.Query != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(s.title ILIKE $%d OR s.description ILIKE $%d OR s.tags ILIKE $%d)", argIndex, argIndex+1, argIndex+2))
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		argIndex += 3
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mp_code_snippets s 
		WHERE %s`, whereClause)

	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 构建排序
	orderBy := "s.created_at DESC"
	switch params.SortBy {
	case "popular":
		orderBy = "s.likes_count DESC, s.created_at DESC"
	case "views":
		orderBy = "s.views_count DESC, s.created_at DESC"
	}

	// 获取数据
	offset := (params.Page - 1) * params.Limit
	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, language, tags,
		       COALESCE(likes_count, 0), COALESCE(views_count, 0), 
		       created_at, updated_at
		FROM mp_code_snippets s
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
		var snippet models.CodeSnippet
		err := rows.Scan(
			&snippet.ID, &snippet.UserID, &snippet.Title, &snippet.Description, &snippet.Language, &snippet.Tags,
			&snippet.LikesCount, &snippet.ViewsCount, &snippet.CreatedAt, &snippet.UpdatedAt,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}
		// 设置默认�?
		snippet.Username = "匿名用户"
		snippets = append(snippets, snippet)
	}

	meta := models.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      totalCount,
		TotalPages: (totalCount + params.Limit - 1) / params.Limit,
	}

	return snippets, meta, nil
}

// ========== 通用功能 ==========

// ToggleLike 切换点赞状态
func (s *CommunityService) ToggleLike(userID uuid.UUID, targetType string, targetID uuid.UUID) (bool, error) {
	// 检查是否已点赞
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM mp_likes WHERE user_id = $1 AND target_type = $2 AND target_id = $3)",
		userID, targetType, targetID).Scan(&exists)
	if err != nil {
		return false, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	if exists {
		// 取消点赞
		_, err = tx.Exec("DELETE FROM mp_likes WHERE user_id = $1 AND target_type = $2 AND target_id = $3",
			userID, targetType, targetID)
		if err != nil {
			return false, err
		}

		// 更新计数
		err = s.updateLikeCount(tx, targetType, targetID, -1)
		if err != nil {
			return false, err
		}

		tx.Commit()
		return false, nil
	} else {
		// 添加点赞
		_, err = tx.Exec("INSERT INTO mp_likes (id, user_id, target_type, target_id, created_at) VALUES ($1, $2, $3, $4, $5)",
			uuid.New(), userID, targetType, targetID, time.Now())
		if err != nil {
			return false, err
		}

		// 更新计数
		err = s.updateLikeCount(tx, targetType, targetID, 1)
		if err != nil {
			return false, err
		}

		tx.Commit()
		return true, nil
	}
}

// updateLikeCount 更新点赞计数
func (s *CommunityService) updateLikeCount(tx *sql.Tx, targetType string, targetID uuid.UUID, delta int) error {
	var query string
	switch targetType {
	case "forum_post":
		query = "UPDATE mp_forum_posts SET likes_count = likes_count + $1 WHERE id = $2"
	case "forum_reply":
		query = "UPDATE mp_forum_replies SET likes_count = likes_count + $1 WHERE id = $2"
	case "blog_post":
		query = "UPDATE mp_blog_posts SET likes_count = likes_count + $1 WHERE id = $2"
	case "code_snippet":
		query = "UPDATE mp_code_snippets SET likes_count = likes_count + $1 WHERE id = $2"
	default:
		return fmt.Errorf("unsupported target type: %s", targetType)
	}

	_, err := tx.Exec(query, delta, targetID)
	return err
}

// GetForumCategories 获取论坛分类
func (s *CommunityService) GetForumCategories() ([]models.ForumCategory, error) {
	query := `
		SELECT id, name, description, icon, sort_order, created_at
		FROM mp_forum_categories
		ORDER BY sort_order ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.ForumCategory
	for rows.Next() {
		var category models.ForumCategory
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.Icon, &category.SortOrder, &category.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetBlogCategories 获取博客分类
func (s *CommunityService) GetBlogCategories() ([]models.BlogCategory, error) {
	query := `
		SELECT id, name, description, color, sort_order, created_at
		FROM mp_blog_categories
		ORDER BY sort_order ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.BlogCategory
	for rows.Next() {
		var category models.BlogCategory
		var color *string
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &color, &category.SortOrder, &category.CreatedAt)
		if err != nil {
			return nil, err
		}
		// 设置默认图标，因为数据库中没有icon字段
		defaultIcon := "BookOutlined"
		category.Icon = &defaultIcon
		categories = append(categories, category)
	}

	return categories, nil
}

// CommunityStats 社区统计信息
type CommunityStats struct {
	TotalForumPosts   int             `json:"total_forum_posts"`
	TotalForumReplies int             `json:"total_forum_replies"`
	TotalBlogPosts    int             `json:"total_blog_posts"`
	TotalCodeSnippets int             `json:"total_code_snippets"`
	TotalUsers        int             `json:"total_users"`
	TotalLikes        int             `json:"total_likes"`
	ActiveUsers       int             `json:"active_users"` // 最近天活跃用户数
	RecentPosts       int             `json:"recent_posts"` // 最近天新帖子数
	PopularCategories []CategoryStats `json:"popular_categories"`
}

// CategoryStats 分类统计信息
type CategoryStats struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PostCount int    `json:"post_count"`
}

// GetCommunityStats 获取社区统计信息
func (s *CommunityService) GetCommunityStats() (*CommunityStats, error) {
	stats := &CommunityStats{}

	// 获取论坛帖子总数
	err := s.db.QueryRow("SELECT COUNT(*) FROM mp_forum_posts").Scan(&stats.TotalForumPosts)
	if err != nil {
		// 如果表不存在，设置为0
		stats.TotalForumPosts = 0
	}

	// 获取论坛回复总数
	err = s.db.QueryRow("SELECT COUNT(*) FROM mp_forum_replies").Scan(&stats.TotalForumReplies)
	if err != nil {
		stats.TotalForumReplies = 0
	}

	// 获取博客文章总数
	err = s.db.QueryRow("SELECT COUNT(*) FROM mp_blog_posts").Scan(&stats.TotalBlogPosts)
	if err != nil {
		stats.TotalBlogPosts = 0
	}

	// 获取代码片段总数
	err = s.db.QueryRow("SELECT COUNT(*) FROM mp_code_snippets").Scan(&stats.TotalCodeSnippets)
	if err != nil {
		stats.TotalCodeSnippets = 0
	}

	// 获取用户总数
	err = s.db.QueryRow("SELECT COUNT(*) FROM mp_users").Scan(&stats.TotalUsers)
	if err != nil {
		stats.TotalUsers = 0
	}

	// 获取点赞总数
	err = s.db.QueryRow("SELECT COUNT(*) FROM mp_likes").Scan(&stats.TotalLikes)
	if err != nil {
		stats.TotalLikes = 0
	}

	// 设置默认值
	stats.ActiveUsers = 0
	stats.RecentPosts = 0

	// 获取热门分类（简化版本）
	categoriesQuery := `
		SELECT id, name, 0 as post_count
		FROM mp_forum_categories
		WHERE is_active = true
		ORDER BY sort_order ASC
		LIMIT 5`

	rows, err := s.db.Query(categoriesQuery)
	if err != nil {
		// 如果查询失败，返回空数组
		stats.PopularCategories = []CategoryStats{}
	} else {
		defer rows.Close()
		var categories []CategoryStats
		for rows.Next() {
			var cat CategoryStats
			err := rows.Scan(&cat.ID, &cat.Name, &cat.PostCount)
			if err != nil {
				continue
			}
			categories = append(categories, cat)
		}
		stats.PopularCategories = categories
	}

	return stats, nil
}
