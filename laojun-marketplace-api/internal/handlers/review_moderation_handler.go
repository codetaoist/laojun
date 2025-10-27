package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReviewModerationHandler 评价审核处理器
type ReviewModerationHandler struct {
	moderationService *services.ReviewModerationService
}

// NewReviewModerationHandler 创建评价审核处理器
func NewReviewModerationHandler(db *shareddb.DB) *ReviewModerationHandler {
	return &ReviewModerationHandler{
		moderationService: services.NewReviewModerationService(db),
	}
}

// ModerateReview 审核评价
// @Summary 审核评价
// @Description 管理员审核用户评价
// @Tags 评价审核
// @Accept json
// @Produce json
// @Param request body services.ModerationRequest true "审核请求"
// @Success 200 {object} map[string]interface{} "审核成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/reviews/moderate [post]
func (h *ReviewModerationHandler) ModerateReview(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	moderatorID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID格式错误"})
		return
	}

	// 检查管理员权限
	role, exists := c.Get("user_role")
	if !exists || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req services.ModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证审核动作
	validActions := map[services.ModerationAction]bool{
		services.ModerationActionApprove: true,
		services.ModerationActionReject:  true,
		services.ModerationActionHide:    true,
		services.ModerationActionFlag:    true,
		services.ModerationActionRestore: true,
	}

	if !validActions[req.Action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核动作"})
		return
	}

	err := h.moderationService.ModerateReview(moderatorID, req)
	if err != nil {
		logger.Error("Failed to moderate review", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "审核成功",
		"data": gin.H{
			"review_id": req.ReviewID,
			"action":    req.Action,
			"status":    "completed",
		},
	})
}

// FlagReview 举报评价
// @Summary 举报评价
// @Description 用户举报不当评价
// @Tags 评价审核
// @Accept json
// @Produce json
// @Param request body services.FlagRequest true "举报请求"
// @Success 200 {object} map[string]interface{} "举报成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/reviews/flag [post]
func (h *ReviewModerationHandler) FlagReview(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID格式错误"})
		return
	}

	var req services.FlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证举报类别
	validCategories := map[string]bool{
		"spam":         true,
		"inappropriate": true,
		"fake":         true,
		"offensive":    true,
		"misleading":   true,
		"other":        true,
	}

	if !validCategories[req.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的举报类别"})
		return
	}

	err := h.moderationService.FlagReview(uid, req)
	if err != nil {
		if err.Error() == "review already flagged by user" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "您已经举报过此评价"})
			return
		}
		if err.Error() == "review not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "评价不存在"})
			return
		}
		logger.Error("Failed to flag review", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "举报失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "举报成功",
		"data": gin.H{
			"review_id": req.ReviewID,
			"category":  req.Category,
			"status":    "pending",
		},
	})
}

// GetPendingReviews 获取待审核评价列表
// @Summary 获取待审核评价列表
// @Description 管理员获取待审核的评价列表
// @Tags 评价审核
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/reviews/pending [get]
func (h *ReviewModerationHandler) GetPendingReviews(c *gin.Context) {
	// 检查管理员权限
	role, exists := c.Get("user_role")
	if !exists || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	reviews, meta, err := h.moderationService.GetPendingReviews(page, limit)
	if err != nil {
		logger.Error("Failed to get pending reviews", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待审核评价失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    reviews,
		"meta":    meta,
	})
}

// GetFlaggedReviews 获取被举报的评价列表
// @Summary 获取被举报的评价列表
// @Description 管理员获取被举报的评价列表
// @Tags 评价审核
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/reviews/flagged [get]
func (h *ReviewModerationHandler) GetFlaggedReviews(c *gin.Context) {
	// 检查管理员权限
	role, exists := c.Get("user_role")
	if !exists || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	flags, meta, err := h.moderationService.GetFlaggedReviews(page, limit)
	if err != nil {
		logger.Error("Failed to get flagged reviews", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取被举报评价失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    flags,
		"meta":    meta,
	})
}

// ResolveFlaggedReview 处理举报的评价
// @Summary 处理举报的评价
// @Description 管理员处理被举报的评价
// @Tags 评价审核
// @Accept json
// @Produce json
// @Param flag_id path string true "举报ID"
// @Param request body map[string]interface{} true "处理请求"
// @Success 200 {object} map[string]interface{} "处理成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 404 {object} map[string]interface{} "举报不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/reviews/flags/{flag_id}/resolve [post]
func (h *ReviewModerationHandler) ResolveFlaggedReview(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	moderatorID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID格式错误"})
		return
	}

	// 检查管理员权限
	role, exists := c.Get("user_role")
	if !exists || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 解析举报ID
	flagIDStr := c.Param("flag_id")
	flagID, err := uuid.Parse(flagIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的举报ID"})
		return
	}

	// 解析请求参数
	var req struct {
		Action string `json:"action" binding:"required"` // dismiss, hide, reject
		Note   string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证处理动作
	validActions := map[string]bool{
		"dismiss": true,
		"hide":    true,
		"reject":  true,
	}

	if !validActions[req.Action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的处理动作"})
		return
	}

	err = h.moderationService.ResolveFlaggedReview(flagID, moderatorID, req.Action, req.Note)
	if err != nil {
		if err.Error() == "flag not found or already resolved" {
			c.JSON(http.StatusNotFound, gin.H{"error": "举报不存在或已处理"})
			return
		}
		logger.Error("Failed to resolve flagged review", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理举报失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "处理成功",
		"data": gin.H{
			"flag_id": flagID,
			"action":  req.Action,
			"status":  "resolved",
		},
	})
}

// GetModerationStats 获取审核统计
// @Summary 获取审核统计
// @Description 管理员获取评价审核统计信息
// @Tags 评价审核
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/reviews/stats [get]
func (h *ReviewModerationHandler) GetModerationStats(c *gin.Context) {
	// 检查管理员权限
	role, exists := c.Get("user_role")
	if !exists || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	stats, err := h.moderationService.GetModerationStats()
	if err != nil {
		logger.Error("Failed to get moderation stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    stats,
	})
}

// GetModerationHistory 获取评价审核历史
// @Summary 获取评价审核历史
// @Description 管理员获取特定评价的审核历史记录
// @Tags 评价审核
// @Accept json
// @Produce json
// @Param review_id path string true "评价ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/reviews/{review_id}/history [get]
func (h *ReviewModerationHandler) GetModerationHistory(c *gin.Context) {
	// 检查管理员权限
	role, exists := c.Get("user_role")
	if !exists || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 解析评价ID
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评价ID"})
		return
	}

	history, err := h.moderationService.GetModerationHistory(reviewID)
	if err != nil {
		logger.Error("Failed to get moderation history", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审核历史失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    history,
	})
}