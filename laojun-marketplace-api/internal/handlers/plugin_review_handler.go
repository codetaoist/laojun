package handlers

import (
	"github.com/codetaoist/laojun-marketplace-api/internal/models"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/codetaoist/laojun-shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PluginReviewHandler 插件审核处理器
type PluginReviewHandler struct {
	reviewService *services.PluginReviewService
}

// NewPluginReviewHandler 创建插件审核处理器
func NewPluginReviewHandler(reviewService *services.PluginReviewService) *PluginReviewHandler {
	return &PluginReviewHandler{
		reviewService: reviewService,
	}
}

// GetReviewQueue 获取审核队列
// @Summary 获取审核队列
// @Description 获取待审核的插件列表
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param status query []string false "审核状态过滤"
// @Param priority query []string false "优先级过滤"
// @Param review_type query []string false "审核类型过滤"
// @Param category_id query string false "分类ID过滤"
// @Param reviewer_id query string false "审核员ID过滤"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default("submitted_for_review_at")
// @Param sort_order query string false "排序方向" default("DESC")
// @Success 200 {object} models.PaginatedResponse[models.Plugin]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/queue [get]
func (h *PluginReviewHandler) GetReviewQueue(c *gin.Context) {
	var params models.ReviewQueueParams

	// 解析查询参数
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters: "+err.Error())
		return
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.SortBy == "" {
		params.SortBy = "submitted_for_review_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "DESC"
	}

	plugins, meta, err := h.reviewService.GetReviewQueue(params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get review queue: "+err.Error())
		return
	}

	response := models.PaginatedResponse[models.Plugin]{
		Data:       plugins,
		Total:      meta.Total,
		Page:       meta.Page,
		Limit:      meta.Limit,
		TotalPages: meta.TotalPages,
	}

	utils.SuccessResponse(c, response)
}

// GetMyReviewTasks 获取我的审核任务
// @Summary 获取我的审核任务
// @Description 获取当前审核员分配的审核任务
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param status query []string false "状态过滤"
// @Param priority query []string false "优先级过滤"
// @Success 200 {object} models.PaginatedResponse[models.Plugin]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 401 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/my-tasks [get]
func (h *PluginReviewHandler) GetMyReviewTasks(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var reviewerUUID uuid.UUID
	switch v := userID.(type) {
	case uuid.UUID:
		reviewerUUID = v
	case string:
		var err error
		reviewerUUID, err = uuid.Parse(v)
		if err != nil {
			utils.BadRequestResponse(c, "Invalid user ID format")
			return
		}
	default:
		utils.BadRequestResponse(c, "Invalid user ID type")
		return
	}

	var params models.ReviewQueueParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters: "+err.Error())
		return
	}

	// 设置审核员过滤，只显示分配给当前用户的任务
	params.ReviewerID = &reviewerUUID
	// 只显示待审核的任务
	if len(params.Status) == 0 {
		params.Status = []models.PluginReviewStatus{models.ReviewStatusPending}
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	plugins, meta, err := h.reviewService.GetReviewQueue(params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get my review tasks: "+err.Error())
		return
	}

	response := models.PaginatedResponse[models.Plugin]{
		Data:       plugins,
		Total:      meta.Total,
		Page:       meta.Page,
		Limit:      meta.Limit,
		TotalPages: meta.TotalPages,
	}

	utils.SuccessResponse(c, response)
}

// AssignReviewer 分配审核员
// @Summary 分配审核员
// @Description 为插件分配审核员
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Param request body AssignReviewerRequest true "分配审核员请求"
// @Success 200 {object} models.ApiResponse[any]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/{id}/assign [post]
func (h *PluginReviewHandler) AssignReviewer(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var request struct {
		ReviewerID uuid.UUID `json:"reviewer_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	err = h.reviewService.AssignReviewer(pluginID, request.ReviewerID)
	if err != nil {
		if err.Error() == "plugin not found or not in pending status" {
			utils.NotFoundResponse(c, "Plugin not found or not in pending status")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to assign reviewer: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Reviewer assigned successfully"})
}

// ReviewPlugin 审核插件
// @Summary 审核插件
// @Description 对插件进行审核
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param request body models.ReviewRequest true "审核请求"
// @Success 200 {object} models.ApiResponse[models.PluginReview]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/review [post]
func (h *PluginReviewHandler) ReviewPlugin(c *gin.Context) {
	var request models.ReviewRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 获取当前用户ID（审核员ID）
	reviewerID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	reviewerUUID, ok := reviewerID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	review, err := h.reviewService.ReviewPlugin(reviewerUUID, request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to review plugin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, review)
}

// BatchReviewPlugins 批量审核插件
// @Summary 批量审核插件
// @Description 批量审核多个插件
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param request body models.BatchReviewRequest true "批量审核请求"
// @Success 200 {object} models.ApiResponse[[]models.PluginReview]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/batch [post]
func (h *PluginReviewHandler) BatchReviewPlugins(c *gin.Context) {
	var request models.BatchReviewRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 获取当前用户ID（审核员ID）
	reviewerID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	reviewerUUID, ok := reviewerID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	reviews, err := h.reviewService.BatchReviewPlugins(reviewerUUID, request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to batch review plugins: "+err.Error())
		return
	}

	utils.SuccessResponse(c, reviews)
}

// GetPluginReviewHistory 获取插件审核历史
// @Summary 获取插件审核历史
// @Description 获取指定插件的审核历史记录
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} models.ApiResponse[[]models.PluginReview]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/{id}/history [get]
func (h *PluginReviewHandler) GetPluginReviewHistory(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	reviews, err := h.reviewService.GetPluginReviewHistory(pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get review history: "+err.Error())
		return
	}

	utils.SuccessResponse(c, reviews)
}

// CreateAppeal 创建申诉
// @Summary 创建申诉
// @Description 开发者创建插件审核申诉
// @Tags 插件申诉
// @Accept json
// @Produce json
// @Param request body models.AppealRequest true "申诉请求"
// @Success 200 {object} models.ApiResponse[models.DeveloperAppeal]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 403 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /developer/appeals [post]
func (h *PluginReviewHandler) CreateAppeal(c *gin.Context) {
	var request models.AppealRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 获取当前用户ID（开发者）
	developerID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	developerUUID, ok := developerID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	appeal, err := h.reviewService.CreateAppeal(developerUUID, request)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		if err.Error() == "plugin does not belong to this developer" {
			utils.ForbiddenResponse(c, "Plugin does not belong to this developer")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create appeal: "+err.Error())
		return
	}

	utils.SuccessResponse(c, appeal)
}

// ProcessAppeal 处理申诉
// @Summary 处理申诉
// @Description 管理员处理开发者插件审核申诉
// @Tags 插件申诉
// @Accept json
// @Produce json
// @Param request body models.AppealProcessRequest true "申诉处理请求"
// @Success 200 {object} models.ApiResponse[models.DeveloperAppeal]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/appeals/process [post]
func (h *PluginReviewHandler) ProcessAppeal(c *gin.Context) {
	var request models.AppealProcessRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 获取当前用户ID（管理员ID）
	adminID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	appeal, err := h.reviewService.ProcessAppeal(adminUUID, request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to process appeal: "+err.Error())
		return
	}

	utils.SuccessResponse(c, appeal)
}

// GetAppeal 获取申诉详情
// @Summary 获取申诉详情
// @Description 获取指定申诉的详细信息
// @Tags 插件申诉
// @Accept json
// @Produce json
// @Param id path string true "申诉ID"
// @Success 200 {object} models.ApiResponse[models.DeveloperAppeal]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/appeals/{id} [get]
func (h *PluginReviewHandler) GetAppeal(c *gin.Context) {
	appealIDStr := c.Param("id")
	appealID, err := uuid.Parse(appealIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid appeal ID")
		return
	}

	appeal, err := h.reviewService.GetAppeal(appealID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			utils.NotFoundResponse(c, "Appeal not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get appeal: "+err.Error())
		return
	}

	utils.SuccessResponse(c, appeal)
}

// GetAppeals 获取申诉列表
// @Summary 获取申诉列表
// @Description 获取申诉列表，支持分页和过滤
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param status query []string false "申诉状态过滤"
// @Param plugin_id query string false "插件ID过滤"
// @Param developer_id query string false "开发者ID过滤"
// @Param date_from query string false "开始日期过滤"
// @Param date_to query string false "结束日期过滤"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default("created_at")
// @Param sort_order query string false "排序方向" default("DESC")
// @Success 200 {object} models.PaginatedResponse[models.DeveloperAppeal]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugin-review/appeals [get]
func (h *PluginReviewHandler) GetAppeals(c *gin.Context) {
	var params models.AppealListParams

	// 解析查询参数
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters: "+err.Error())
		return
	}

	// 设置默认值
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

	// 获取申诉列表
	appeals, meta, err := h.reviewService.GetAppeals(params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get appeals: "+err.Error())
		return
	}

	// 返回分页结果
	response := models.PaginatedResponse[models.DeveloperAppeal]{
		Data:       appeals,
		Total:      meta.Total,
		Page:       meta.Page,
		Limit:      meta.Limit,
		TotalPages: meta.TotalPages,
	}

	utils.SuccessResponse(c, response)
}

// GetReviewStats 获取审核统计
// @Summary 获取审核统计
// @Description 获取插件审核的统计信息
// @Tags 插件审核
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse[models.ReviewStats]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/stats [get]
func (h *PluginReviewHandler) GetReviewStats(c *gin.Context) {
	stats, err := h.reviewService.GetReviewStats()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get review stats: "+err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}

// AutoReviewPlugin 自动审核插件
// @Summary 自动审核插件
// @Description 对插件进行自动审核
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} models.ApiResponse[models.AutoReviewLog]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/{id}/auto [post]
func (h *PluginReviewHandler) AutoReviewPlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	log, err := h.reviewService.AutoReview(pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to auto review plugin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, log)
}

// GetReviewerWorkload 获取审核员工作負載
// @Summary 获取审核员工作負載
// @Description 获取审核员的工作負載統計
// @Tags 插件審核
// @Accept json
// @Produce json
// @Param reviewer_id query string false "審核員ID"
// @Param date_from query string false "開始日期"
// @Param date_to query string false "結束日期"
// @Success 200 {object} models.ApiResponse[[]models.ReviewerWorkload]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/reviews/workload [get]
func (h *PluginReviewHandler) GetReviewerWorkload(c *gin.Context) {
	// 这里可以实现获取审核员工作負載的邏輯
	// 暂时返回空數據
	utils.SuccessResponse(c, []models.ReviewerWorkload{})
}

// GetMyReviews 获取我的审核任务
// @Summary 获取我的审核任务
// @Description 获取当前审核员的审核任务列表
// @Tags 插件審核
// @Accept json
// @Produce json
// @Param status query string false "狀態過濾"
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(20)
// @Success 200 {object} models.PaginatedResponse[models.Plugin]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /reviewer/reviews [get]
func (h *PluginReviewHandler) GetMyReviews(c *gin.Context) {
	// 获取当前用户ID（审核员ID）
	reviewerID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	reviewerUUID, ok := reviewerID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	var params models.ReviewQueueParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters: "+err.Error())
		return
	}

	// 设置审核员ID
	params.ReviewerID = &reviewerUUID

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	plugins, meta, err := h.reviewService.GetReviewQueue(params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get my reviews: "+err.Error())
		return
	}

	response := models.PaginatedResponse[models.Plugin]{
		Data:       plugins,
		Total:      meta.Total,
		Page:       meta.Page,
		Limit:      meta.Limit,
		TotalPages: meta.TotalPages,
	}

	utils.SuccessResponse(c, response)
}
