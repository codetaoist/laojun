package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/codetaoist/laojun-admin-api/internal/services"
)

// PluginAuditHandler 插件审核处理器
type PluginAuditHandler struct {
	auditService *services.PluginAuditService
	logger       *logrus.Logger
}

// NewPluginAuditHandler 创建插件审核处理器
func NewPluginAuditHandler(auditService *services.PluginAuditService, logger *logrus.Logger) *PluginAuditHandler {
	return &PluginAuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// SubmitPluginForAudit 提交插件审核
// @Summary 提交插件审核
// @Description 开发者提交插件进行审核
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param request body models.PluginAuditSubmissionRequest true "审核提交请求"
// @Success 200 {object} models.PluginAuditRecord
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/plugin-audit/submit [post]
func (h *PluginAuditHandler) SubmitPluginForAudit(c *gin.Context) {
	var req models.PluginAuditSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 从JWT token中获取开发者ID
	developerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := developerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	auditRecord, err := h.auditService.SubmitPluginForAudit(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to submit plugin for audit")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit plugin for audit"})
		return
	}

	c.JSON(http.StatusOK, auditRecord)
}

// AssignAuditor 分配审核员
// @Summary 分配审核员
// @Description 管理员为插件审核分配审核员
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param id path string true "审核记录ID"
// @Param request body models.PluginAuditAssignmentRequest true "审核员分配请求"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/plugin-audit/{id}/assign [post]
func (h *PluginAuditHandler) AssignAuditor(c *gin.Context) {
	auditRecordIDStr := c.Param("id")
	auditRecordID, err := uuid.Parse(auditRecordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit record ID"})
		return
	}

	var req models.PluginAuditAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// TODO: 验证管理员权限
	// if !h.hasAdminPermission(c) {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	//     return
	// }

	err = h.auditService.AssignAuditor(c.Request.Context(), auditRecordID, req.AuditorID, req.Notes)
	if err != nil {
		h.logger.WithError(err).Error("Failed to assign auditor")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign auditor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Auditor assigned successfully"})
}

// SubmitAuditReview 提交审核结果
// @Summary 提交审核结果
// @Description 审核员提交插件审核结果
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param id path string true "审核记录ID"
// @Param request body models.PluginAuditReviewRequest true "审核结果请求"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/plugin-audit/{id}/review [post]
func (h *PluginAuditHandler) SubmitAuditReview(c *gin.Context) {
	auditRecordIDStr := c.Param("id")
	auditRecordID, err := uuid.Parse(auditRecordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit record ID"})
		return
	}

	var req models.PluginAuditReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 从JWT token中获取审核员ID
	auditorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := auditorID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// TODO: 验证审核员权限和分配
	// if !h.isAssignedAuditor(c, auditRecordID, userID) {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "Not assigned to this audit"})
	//     return
	// }

	err = h.auditService.SubmitAuditReview(c.Request.Context(), auditRecordID, &req, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to submit audit review")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit audit review"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audit review submitted successfully"})
}

// GetAuditRecords 获取审核记录列表
// @Summary 获取审核记录列表
// @Description 获取插件审核记录列表，支持筛选和分页
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param status query []string false "审核状态筛选"
// @Param priority query []string false "优先级筛选"
// @Param auditor_id query string false "审核员ID筛选"
// @Param developer_id query string false "开发者ID筛选"
// @Param submission_type query []string false "提交类型筛选"
// @Param date_from query string false "开始日期"
// @Param date_to query string false "结束日期"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页大小" default(20)
// @Param sort_by query string false "排序字段" default(submitted_at)
// @Param sort_order query string false "排序方向" default(desc)
// @Success 200 {object} models.PluginAuditListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/plugin-audit/records [get]
func (h *PluginAuditHandler) GetAuditRecords(c *gin.Context) {
	req := &models.PluginAuditListRequest{
		Page:      1,
		Size:      20,
		SortBy:    "submitted_at",
		SortOrder: "desc",
	}

	// 解析查询参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			req.Size = size
		}
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	// 解析状态筛选
	if statusList := c.QueryArray("status"); len(statusList) > 0 {
		for _, status := range statusList {
			req.Status = append(req.Status, models.PluginAuditStatus(status))
		}
	}

	// 解析优先级筛选
	if priorityList := c.QueryArray("priority"); len(priorityList) > 0 {
		for _, priority := range priorityList {
			req.Priority = append(req.Priority, models.PluginAuditPriority(priority))
		}
	}

	// 解析提交类型筛选
	if submissionTypeList := c.QueryArray("submission_type"); len(submissionTypeList) > 0 {
		for _, submissionType := range submissionTypeList {
			req.SubmissionType = append(req.SubmissionType, models.PluginSubmissionType(submissionType))
		}
	}

	// 解析审核员ID筛选
	if auditorIDStr := c.Query("auditor_id"); auditorIDStr != "" {
		if auditorID, err := uuid.Parse(auditorIDStr); err == nil {
			req.AuditorID = &auditorID
		}
	}

	// 解析开发者ID筛选
	if developerIDStr := c.Query("developer_id"); developerIDStr != "" {
		if developerID, err := uuid.Parse(developerIDStr); err == nil {
			req.DeveloperID = &developerID
		}
	}

	// TODO: 解析日期筛选
	// if dateFromStr := c.Query("date_from"); dateFromStr != "" {
	//     if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
	//         req.DateFrom = &dateFrom
	//     }
	// }

	response, err := h.auditService.GetAuditRecords(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit records")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit records"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAuditRecord 获取单个审核记录详情
// @Summary 获取审核记录详情
// @Description 获取指定插件审核记录的详细信息
// @Tags 插件审核
// @Accept json
// @Produce json
// @Param id path string true "审核记录ID"
// @Success 200 {object} models.PluginAuditResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/plugin-audit/{id} [get]
func (h *PluginAuditHandler) GetAuditRecord(c *gin.Context) {
	auditRecordIDStr := c.Param("id")
	auditRecordID, err := uuid.Parse(auditRecordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit record ID"})
		return
	}

	// TODO: 从服务层获取审核记录详情
	// record, err := h.auditService.GetAuditRecord(c.Request.Context(), auditRecordID)
	// if err != nil {
	//     h.logger.WithError(err).Error("Failed to get audit record")
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit record"})
	//     return
	// }

	// 临时返回空记录
	_ = auditRecordID
	c.JSON(http.StatusOK, gin.H{"message": "Audit record details (placeholder)"})
}

// GetAuditStatistics 获取审核统计信息
// @Summary 获取审核统计信息
// @Description 获取插件审核的统计信息和指标
// @Tags 插件审核
// @Accept json
// @Produce json
// @Success 200 {object} models.PluginAuditStatistics
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/plugin-audit/statistics [get]
func (h *PluginAuditHandler) GetAuditStatistics(c *gin.Context) {
	// TODO: 验证管理员权限
	// if !h.hasAdminPermission(c) {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	//     return
	// }

	statistics, err := h.auditService.GetAuditStatistics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit statistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit statistics"})
		return
	}

	c.JSON(http.StatusOK, statistics)
}

// VerifyDeveloper 开发者认证申请
// @Summary 开发者认证申请
// @Description 开发者提交认证申请
// @Tags 开发者认证
// @Accept json
// @Produce json
// @Param request body models.DeveloperVerificationRequest true "开发者认证请求"
// @Success 200 {object} models.DeveloperProfile
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/developer/verify [post]
func (h *PluginAuditHandler) VerifyDeveloper(c *gin.Context) {
	var req models.DeveloperVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 从JWT token中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	profile, err := h.auditService.VerifyDeveloper(c.Request.Context(), uid, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to verify developer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify developer"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateAuditorProfile 更新审核员档案
// @Summary 更新审核员档案
// @Description 管理员更新审核员档案信息
// @Tags 审核员管理
// @Accept json
// @Produce json
// @Param id path string true "审核员ID"
// @Param request body models.AuditorProfileUpdateRequest true "审核员档案更新请求"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/auditor/{id}/profile [put]
func (h *PluginAuditHandler) UpdateAuditorProfile(c *gin.Context) {
	auditorIDStr := c.Param("id")
	auditorID, err := uuid.Parse(auditorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid auditor ID"})
		return
	}

	var req models.AuditorProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// TODO: 验证管理员权限
	// if !h.hasAdminPermission(c) {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	//     return
	// }

	err = h.auditService.UpdateAuditorProfile(c.Request.Context(), auditorID, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update auditor profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update auditor profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Auditor profile updated successfully"})
}

// 私有方法

// hasAdminPermission 检查管理员权限
func (h *PluginAuditHandler) hasAdminPermission(c *gin.Context) bool {
	// TODO: 实现权限检查逻辑
	return true
}

// isAssignedAuditor 检查是否为分配的审核员
func (h *PluginAuditHandler) isAssignedAuditor(c *gin.Context, auditRecordID, auditorID uuid.UUID) bool {
	// TODO: 实现审核员分配检查逻辑
	return true
}