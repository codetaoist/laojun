package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CommunityHandler 社区处理
type CommunityHandler struct {
	communityService *services.CommunityService
}

// NewCommunityHandler 创建社区处理
func NewCommunityHandler(communityService *services.CommunityService) *CommunityHandler {
	return &CommunityHandler{
		communityService: communityService,
	}
}

// ========== 论坛相关 ==========

// GetForumCategories 获取论坛分类
func (h *CommunityHandler) GetForumCategories(c *gin.Context) {
	categories, err := h.communityService.GetForumCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取论坛分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

// CreateForumPost 创建论坛帖子
func (h *CommunityHandler) CreateForumPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var req struct {
		CategoryID uuid.UUID `json:"category_id" binding:"required"`
		Title      string    `json:"title" binding:"required,max=200"`
		Content    string    `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	post := &models.ForumPost{
		CategoryID: req.CategoryID,
		UserID:     userID.(uuid.UUID),
		Title:      req.Title,
		Content:    req.Content,
	}

	err := h.communityService.CreateForumPost(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建帖子失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "帖子创建成功",
		"data":    post,
	})
}

// GetForumPosts 获取论坛帖子列表
func (h *CommunityHandler) GetForumPosts(c *gin.Context) {
	var params services.ForumSearchParams

	// 解析查询参数
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			params.CategoryID = &categoryID
		}
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			params.UserID = &userID
		}
	}

	params.Query = c.Query("query")
	params.SortBy = c.DefaultQuery("sort_by", "latest")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	params.Page = page
	params.Limit = limit

	posts, meta, err := h.communityService.GetForumPosts(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取帖子列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": posts,
		"meta": meta,
	})
}

// GetForumPost 获取单个论坛帖子
func (h *CommunityHandler) GetForumPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的帖子ID",
		})
		return
	}

	post, err := h.communityService.GetForumPost(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取帖子失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": post,
	})
}

// CreateForumReply 创建论坛回复
func (h *CommunityHandler) CreateForumReply(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	postIDStr := c.Param("id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的帖子ID",
		})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	reply := &models.ForumReply{
		PostID:  postID,
		UserID:  userID.(uuid.UUID),
		Content: req.Content,
	}

	err = h.communityService.CreateForumReply(reply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建回复失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "回复创建成功",
		"data":    reply,
	})
}

// GetForumReplies 获取论坛回复列表
func (h *CommunityHandler) GetForumReplies(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的帖子ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	replies, meta, err := h.communityService.GetForumReplies(postID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取回复列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": replies,
		"meta": meta,
	})
}

// ========== 博客相关 ==========

// GetBlogCategories 获取博客分类
func (h *CommunityHandler) GetBlogCategories(c *gin.Context) {
	categories, err := h.communityService.GetBlogCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取博客分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

// CreateBlogPost 创建博客文章
func (h *CommunityHandler) CreateBlogPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var req struct {
		CategoryID uuid.UUID `json:"category_id"`
		Title      string    `json:"title" binding:"required,max=200"`
		Content    string    `json:"content" binding:"required"`
		Summary    *string   `json:"summary"`
		CoverImage *string   `json:"cover_image"`
		Tags       *string   `json:"tags"`
		IsPublished bool     `json:"is_published"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	post := &models.BlogPost{
		CategoryID:  req.CategoryID,
		UserID:      userID.(uuid.UUID),
		Title:       req.Title,
		Content:     req.Content,
		Summary:     req.Summary,
		CoverImage:  req.CoverImage,
		Tags:        req.Tags,
		IsPublished: req.IsPublished,
	}

	err := h.communityService.CreateBlogPost(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建博客文章失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "博客文章创建成功",
		"data":    post,
	})
}

// GetBlogPosts 获取博客文章列表
func (h *CommunityHandler) GetBlogPosts(c *gin.Context) {
	var params services.BlogSearchParams

	// 解析查询参数
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			params.CategoryID = &categoryID
		}
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			params.UserID = &userID
		}
	}

	params.Query = c.Query("query")
	params.SortBy = c.DefaultQuery("sort_by", "latest")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	params.Page = page
	params.Limit = limit

	posts, meta, err := h.communityService.GetBlogPosts(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取博客文章列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": posts,
		"meta": meta,
	})
}

// GetBlogPost 获取单个博客文章
func (h *CommunityHandler) GetBlogPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的文章ID",
		})
		return
	}

	post, err := h.communityService.GetBlogPost(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取文章失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": post,
	})
}

// ========== 代码分享相关 ==========

// CreateCodeSnippet 创建代码片段
func (h *CommunityHandler) CreateCodeSnippet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var req struct {
		Title       string  `json:"title" binding:"required,max=200"`
		Description *string `json:"description"`
		Code        string  `json:"code" binding:"required"`
		Language    string  `json:"language" binding:"required"`
		Tags        *string `json:"tags"`
		IsPublic    bool    `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	snippet := &models.CodeSnippet{
		UserID:      userID.(uuid.UUID),
		Title:       req.Title,
		Description: req.Description,
		Code:        req.Code,
		Language:    req.Language,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
	}

	err := h.communityService.CreateCodeSnippet(snippet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建代码片段失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "代码片段创建成功",
		"data":    snippet,
	})
}

// GetCodeSnippets 获取代码片段列表
func (h *CommunityHandler) GetCodeSnippets(c *gin.Context) {
	var params services.CodeSearchParams

	// 解析查询参数
	params.Language = c.Query("language")
	params.Query = c.Query("query")
	params.SortBy = c.DefaultQuery("sort_by", "latest")

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			params.UserID = &userID
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	params.Page = page
	params.Limit = limit

	snippets, meta, err := h.communityService.GetCodeSnippets(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取代码片段列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": snippets,
		"meta": meta,
	})
}

// ========== 通用功能 ==========

// ToggleLike 切换点赞状态
func (h *CommunityHandler) ToggleLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var req struct {
		TargetType string    `json:"target_type" binding:"required,oneof=forum_post forum_reply blog_post code_snippet"`
		TargetID   uuid.UUID `json:"target_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	isLiked, err := h.communityService.ToggleLike(userID.(uuid.UUID), req.TargetType, req.TargetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "操作失败",
		})
		return
	}

	message := "取消点赞成功"
	if isLiked {
		message = "点赞成功"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  message,
		"is_liked": isLiked,
	})
}

// GetCommunityStats 获取社区统计信息
func (h *CommunityHandler) GetCommunityStats(c *gin.Context) {
	stats, err := h.communityService.GetCommunityStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取社区统计信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
