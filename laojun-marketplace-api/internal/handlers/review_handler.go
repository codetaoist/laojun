package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/codetaoist/laojun-shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReviewHandler 评论处理
type ReviewHandler struct {
	reviewService *services.ReviewService
}

// NewReviewHandler 创建评论处理
func NewReviewHandler(reviewService *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// CreateReviewRequest 创建评论请求
type CreateReviewRequest struct {
	Rating  float64 `json:"rating" binding:"required,min=1,max=5"`
	Comment string  `json:"comment" binding:"required,min=10,max=1000"`
}

// UpdateReviewRequest 更新评论请求
type UpdateReviewRequest struct {
	Rating  float64 `json:"rating" binding:"required,min=1,max=5"`
	Comment string  `json:"comment" binding:"required,min=10,max=1000"`
}

// GetPluginReviews 获取插件的评论列表
// @Summary 获取插件的评论列表
// @Description 获取指定插件的评论列表，支持分页和排序
// @Tags 插件審核
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Param sort_by query string false "排序字段" default(rating)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} models.PaginatedResponse[models.Review]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /plugins/{id}/reviews [get]
func (h *ReviewHandler) GetPluginReviews(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	// 解析查询参数
	params := models.ReviewSearchParams{
		Page:   1,
		Limit:  20,
		SortBy: c.Query("sort_by"),
	}

	// 解析页码
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// 解析每页数量
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			params.Limit = limit
		}
	}

	// 获取评论列表
	reviews, meta, err := h.reviewService.GetPluginReviews(pluginID, params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get reviews")
		return
	}

	// 如果需要统计信息
	if c.Query("with_stats") == "true" {
		stats, err := h.reviewService.GetReviewStats(pluginID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get review stats")
			return
		}

		response := gin.H{
			"reviews": reviews,
			"meta":    meta,
			"stats":   stats,
		}
		utils.SuccessResponse(c, response)
	} else {
		utils.PaginatedResponse(c, reviews, meta)
	}
}

// CreateReview 创建评论
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plugin ID")
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// 验证评分
	if req.Rating < 1 || req.Rating > 5 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Rating must be between 1 and 5")
		return
	}

	// 清理评论内容
	req.Comment = utils.SanitizeString(req.Comment)

	review, err := h.reviewService.CreateReview(userID.(uuid.UUID), pluginID, int(req.Rating), req.Comment)
	if err != nil {
		if err.Error() == "user has already reviewed this plugin" {
			utils.ErrorResponse(c, http.StatusBadRequest, "You have already reviewed this plugin")
			return
		}
		if err.Error() == "user must purchase the plugin before reviewing" {
			utils.ErrorResponse(c, http.StatusBadRequest, "You must purchase the plugin before reviewing")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create review")
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Code:      http.StatusCreated,
		Message:   "Review created successfully",
		Data:      review,
		Timestamp: time.Now(),
	})
}

// UpdateReview 更新评论
func (h *ReviewHandler) UpdateReview(c *gin.Context) {

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID")
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// 验证评分
	if req.Rating < 1 || req.Rating > 5 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Rating must be between 1 and 5")
		return
	}

	// 清理评论内容
	req.Comment = utils.SanitizeString(req.Comment)

	review, err := h.reviewService.UpdateReview(reviewID, int(req.Rating), req.Comment)
	if err != nil {
		if err.Error() == "review not found or not owned by user" {
			utils.ErrorResponse(c, http.StatusNotFound, "Review not found or you don't have permission to update it")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update review")
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:      http.StatusOK,
		Message:   "Review updated successfully",
		Data:      review,
		Timestamp: time.Now(),
	})
}

// DeleteReview 删除评论
func (h *ReviewHandler) DeleteReview(c *gin.Context) {

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID")
		return
	}

	err = h.reviewService.DeleteReview(reviewID)
	if err != nil {
		if err.Error() == "review not found or not owned by user" {
			utils.ErrorResponse(c, http.StatusNotFound, "Review not found or you don't have permission to delete it")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete review")
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
