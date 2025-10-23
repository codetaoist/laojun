package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun/internal/services"
	"github.com/codetaoist/laojun/pkg/shared/models"
	"github.com/codetaoist/laojun/pkg/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminPluginHandler 总后台插件管理处理器
type AdminPluginHandler struct {
	adminPluginService  *services.AdminPluginService
}

// NewAdminPluginHandler 创建新的管理员插件处理器
func NewAdminPluginHandler(
	adminPluginService *services.AdminPluginService,
) *AdminPluginHandler {
	return &AdminPluginHandler{
		adminPluginService: adminPluginService,
	}
}

// GetPluginsForAdmin 获取管理员插件列表
// @Summary 获取管理员插件列表
// @Description 获取总后台管理的插件列表，支持筛选和分页
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Param status query []string false "插件状态过滤"
// @Param category query string false "分类过滤"
// @Param developer query string false "开发者过滤"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default("created_at")
// @Param sort_order query string false "排序方向" default("DESC")
// @Success 200 {object} models.PaginatedResponse[models.Plugin]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins [get]
func (h *AdminPluginHandler) GetPluginsForAdmin(c *gin.Context) {
	var params models.AdminPluginSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:      400,
			Message:   "Invalid query parameters",
			Timestamp: time.Now(),
		})
		return
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	// 调用服务获取插件列表
	plugins, total, err := h.adminPluginService.GetPluginsForAdmin(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:      500,
			Message:   "Failed to get plugins: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 构建响应
	response := models.PaginatedResponse{
		Data: plugins,
		Meta: *total,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Success",
		Data:      response,
		Timestamp: time.Now(),
	})
}

// GetPluginForAdmin 获取插件详情（管理员视图）
// @Summary 获取插件详情（管理员视图）
// @Description 获取插件的详细信息，包括管理员专用的统计和配置信息
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} models.ApiResponse[models.AdminPluginDetail]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins/{id} [get]
func (h *AdminPluginHandler) GetPluginForAdmin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	pluginDetail, err := h.adminPluginService.GetPluginForAdmin(pluginID)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get plugin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, pluginDetail)
}

// UpdatePluginStatus 更新插件状态
// @Summary 更新插件状态
// @Description 管理员更新插件状态（启用、禁用、下架等）
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Param request body models.UpdatePluginStatusRequest true "状态更新请求"
// @Success 200 {object} models.ApiResponse[any]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins/{id}/status [patch]
func (h *AdminPluginHandler) UpdatePluginStatus(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var request models.UpdatePluginStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 获取当前管理员ID
	adminID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid admin ID format")
		return
	}

	err = h.adminPluginService.UpdatePluginStatus(pluginID, request.Status, request.Reason, adminUUID)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update plugin status: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Plugin status updated successfully"})
}

// GetPluginStats 获取插件统计信息
// @Summary 获取插件统计信息
// @Description 获取插件的详细统计信息，包括下载量、评分、收入等
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} models.ApiResponse[models.PluginStats]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins/{id}/stats [get]
func (h *AdminPluginHandler) GetPluginStats(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	stats, err := h.adminPluginService.GetPluginStats(pluginID)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get plugin stats: "+err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}

// GetPluginConfig 获取插件配置
// @Summary 获取插件配置
// @Description 获取插件的配置信息
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} models.ApiResponse[models.PluginConfig]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins/{id}/config [get]
func (h *AdminPluginHandler) GetPluginConfig(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	config, err := h.adminPluginService.GetPluginConfig(pluginID)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get plugin config: "+err.Error())
		return
	}

	utils.SuccessResponse(c, config)
}

// UpdatePluginConfig 更新插件配置
// @Summary 更新插件配置
// @Description 更新插件的配置信息
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Param request body models.UpdatePluginConfigRequest true "配置更新请求"
// @Success 200 {object} models.ApiResponse[any]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins/{id}/config [put]
func (h *AdminPluginHandler) UpdatePluginConfig(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var request models.UpdatePluginConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 获取当前管理员ID
	adminID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid admin ID format")
		return
	}

	err = h.adminPluginService.UpdatePluginConfig(pluginID, request.Config, adminUUID)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update plugin config: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Plugin config updated successfully"})
}

// SyncPluginFromMarketplace 从插件市场同步插件
func (h *AdminPluginHandler) SyncPluginFromMarketplace(c *gin.Context) {
	pluginIDStr := c.Param("id")
	_, err := uuid.Parse(pluginIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:      400,
			Message:   "Invalid plugin ID",
			Timestamp: time.Now(),
		})
		return
	}

	var request models.SyncPluginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: time.Now(),
		})
		return
	}

	// TODO: 实现从市场同步插件逻辑
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Marketplace sync feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// SyncPluginToMarketplace 同步插件到插件市场
func (h *AdminPluginHandler) SyncPluginToMarketplace(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Sync to marketplace feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// BatchSyncPlugins 批量同步插件
func (h *AdminPluginHandler) BatchSyncPlugins(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Batch sync feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// GetDashboardStats 获取仪表板统计
// @Summary 获取仪表板统计
// @Description 获取总后台插件管理的仪表板统计信息
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse[models.AdminDashboardStats]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins/dashboard-stats [get]
func (h *AdminPluginHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminPluginService.GetDashboardStats()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get dashboard stats: "+err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}

// GetPluginLogs 获取插件日志
// @Summary 获取插件日志
// @Description 获取插件的操作日志
// @Tags 总后台插件管理
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(50)
// @Param level query string false "日志级别过滤"
// @Success 200 {object} models.PaginatedResponse[models.PluginLog]
// @Failure 400 {object} models.ApiResponse[any]
// @Failure 404 {object} models.ApiResponse[any]
// @Failure 500 {object} models.ApiResponse[any]
// @Router /admin/plugins/{id}/logs [get]
func (h *AdminPluginHandler) GetPluginLogs(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	// 解析分页参数
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	level := c.Query("level")

	logs, meta, err := h.adminPluginService.GetPluginLogs(pluginID, page, limit, level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:      500,
			Message:   "Failed to get plugin logs: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	response := models.PaginatedResponse{
		Data: logs,
		Meta: *meta,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Success",
		Data:      response,
		Timestamp: time.Now(),
	})
}

// SearchMarketplacePlugins 搜索插件市场插件
func (h *AdminPluginHandler) SearchMarketplacePlugins(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Marketplace search feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// GetMarketplacePlugin 获取插件市场插件详情
func (h *AdminPluginHandler) GetMarketplacePlugin(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Marketplace plugin details feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// GetMarketplaceCategories 获取插件市场分类
func (h *AdminPluginHandler) GetMarketplaceCategories(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Marketplace categories feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// GetMarketplaceTags 获取插件市场标签
func (h *AdminPluginHandler) GetMarketplaceTags(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Marketplace tags feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// InstallMarketplacePlugin 安装插件市场插件
func (h *AdminPluginHandler) InstallMarketplacePlugin(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Plugin installation feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// UninstallMarketplacePlugin 卸载插件市场插件
func (h *AdminPluginHandler) UninstallMarketplacePlugin(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Plugin uninstallation feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// PublishToMarketplace 发布插件到市场
func (h *AdminPluginHandler) PublishToMarketplace(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Publish to marketplace feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// UpdateMarketplacePlugin 更新插件市场插件
func (h *AdminPluginHandler) UpdateMarketplacePlugin(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Update marketplace plugin feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// DeleteFromMarketplace 从插件市场删除插件
func (h *AdminPluginHandler) DeleteFromMarketplace(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      200,
		Message:   "Delete from marketplace feature not implemented yet",
		Timestamp: time.Now(),
	})
}

// 占位符方法 - 这些方法需要根据具体需求实现
func (h *AdminPluginHandler) GetSystemConfig(c *gin.Context)        { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) UpdateSystemConfig(c *gin.Context)     { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetSystemHealth(c *gin.Context)        { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetSystemMetrics(c *gin.Context)       { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) ClearCache(c *gin.Context)             { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetCacheStats(c *gin.Context)          { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetDevelopers(c *gin.Context)          { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetDeveloper(c *gin.Context)           { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) UpdateDeveloperStatus(c *gin.Context)  { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetDeveloperPlugins(c *gin.Context)    { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetDeveloperStats(c *gin.Context)      { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetReviewQueue(c *gin.Context)         { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetMyReviewTasks(c *gin.Context)       { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) AssignReviewer(c *gin.Context)         { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) ReviewPlugin(c *gin.Context)           { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) BatchReviewPlugins(c *gin.Context)     { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetPluginReports(c *gin.Context)       { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetDownloadReports(c *gin.Context)     { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetRevenueReports(c *gin.Context)      { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetUserReports(c *gin.Context)         { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetPerformanceReports(c *gin.Context)  { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetSystemLogs(c *gin.Context)          { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetAuditLogs(c *gin.Context)           { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) ExportLogs(c *gin.Context)             { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) CreateBackup(c *gin.Context)           { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) GetBackupList(c *gin.Context)          { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) RestoreBackup(c *gin.Context)          { utils.NotImplementedResponse(c) }
func (h *AdminPluginHandler) DeleteBackup(c *gin.Context)           { utils.NotImplementedResponse(c) }

// 页面处理方法

// PluginListPage 插件列表页面
func (h *AdminPluginHandler) PluginListPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/plugins/list.html", gin.H{
		"title": "插件管理",
	})
}

// PluginDetailPage 插件详情页面
func (h *AdminPluginHandler) PluginDetailPage(c *gin.Context) {
	pluginID := c.Param("id")
	c.HTML(http.StatusOK, "admin/plugins/detail.html", gin.H{
		"title":    "插件详情",
		"pluginID": pluginID,
	})
}

// PluginConfigPage 插件配置页面
func (h *AdminPluginHandler) PluginConfigPage(c *gin.Context) {
	pluginID := c.Param("id")
	c.HTML(http.StatusOK, "admin/plugins/config.html", gin.H{
		"title":    "插件配置",
		"pluginID": pluginID,
	})
}

// PluginLogsPage 插件日志页面
func (h *AdminPluginHandler) PluginLogsPage(c *gin.Context) {
	pluginID := c.Param("id")
	c.HTML(http.StatusOK, "admin/plugins/logs.html", gin.H{
		"title":    "插件日志",
		"pluginID": pluginID,
	})
}

// DashboardPage 仪表板页面
func (h *AdminPluginHandler) DashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"title": "总后台仪表板",
	})
}

// MarketplacePage 插件市场页面
func (h *AdminPluginHandler) MarketplacePage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/marketplace/index.html", gin.H{
		"title": "插件市场",
	})
}

// MarketplacePluginPage 插件市场插件详情页面
func (h *AdminPluginHandler) MarketplacePluginPage(c *gin.Context) {
	pluginID := c.Param("id")
	c.HTML(http.StatusOK, "admin/marketplace/plugin.html", gin.H{
		"title":    "插件详情",
		"pluginID": pluginID,
	})
}

// DevelopersPage 开发者管理页面
func (h *AdminPluginHandler) DevelopersPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/developers/list.html", gin.H{
		"title": "开发者管理",
	})
}

// DeveloperDetailPage 开发者详情页面
func (h *AdminPluginHandler) DeveloperDetailPage(c *gin.Context) {
	developerID := c.Param("id")
	c.HTML(http.StatusOK, "admin/developers/detail.html", gin.H{
		"title":       "开发者详情",
		"developerID": developerID,
	})
}

// ReviewQueuePage 审核队列页面
func (h *AdminPluginHandler) ReviewQueuePage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/reviews/queue.html", gin.H{
		"title": "审核队列",
	})
}

// MyReviewTasksPage 我的审核任务页面
func (h *AdminPluginHandler) MyReviewTasksPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/reviews/my-tasks.html", gin.H{
		"title": "我的审核任务",
	})
}

// ReviewDetailPage 审核详情页面
func (h *AdminPluginHandler) ReviewDetailPage(c *gin.Context) {
	reviewID := c.Param("id")
	c.HTML(http.StatusOK, "admin/reviews/detail.html", gin.H{
		"title":    "审核详情",
		"reviewID": reviewID,
	})
}

// ReportsPage 报告页面
func (h *AdminPluginHandler) ReportsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/reports/index.html", gin.H{
		"title": "报告管理",
	})
}

// PluginReportsPage 插件报告页面
func (h *AdminPluginHandler) PluginReportsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/reports/plugins.html", gin.H{
		"title": "插件报告",
	})
}

// DownloadReportsPage 下载报告页面
func (h *AdminPluginHandler) DownloadReportsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/reports/downloads.html", gin.H{
		"title": "下载报告",
	})
}

// RevenueReportsPage 收入报告页面
func (h *AdminPluginHandler) RevenueReportsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/reports/revenue.html", gin.H{
		"title": "收入报告",
	})
}

// SystemConfigPage 系统配置页面
func (h *AdminPluginHandler) SystemConfigPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/system/config.html", gin.H{
		"title": "系统配置",
	})
}

// SystemHealthPage 系统健康页面
func (h *AdminPluginHandler) SystemHealthPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/system/health.html", gin.H{
		"title": "系统健康",
	})
}

// SystemLogsPage 系统日志页面
func (h *AdminPluginHandler) SystemLogsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/system/logs.html", gin.H{
		"title": "系统日志",
	})
}

// BackupPage 备份页面
func (h *AdminPluginHandler) BackupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/system/backup.html", gin.H{
		"title": "系统备份",
	})
}