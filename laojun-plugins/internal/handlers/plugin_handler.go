package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/codetaoist/laojun-plugins/internal/services"
)

// PluginHandler 插件处理
type PluginHandler struct {
	pluginService *services.PluginService
}

// NewPluginHandler 创建插件处理
func NewPluginHandler(pluginService *services.PluginService) *PluginHandler {
	return &PluginHandler{
		pluginService: pluginService,
	}
}

// GetPlugins 获取插件列表
func (h *PluginHandler) GetPlugins(c *gin.Context) {
	// 解析查询参数（兼容多种前端键名）
	query := c.Query("q")
	if query == "" {
		query = c.Query("keyword")
	}
	if query == "" {
		query = c.Query("query") // 前端 mapSearchParams 可能使用 query
	}
	params := models.PluginSearchParams{
		Page:   1,
		Limit:  20,
		Query:  query,
		SortBy: c.Query("sort_by"),
	}

	// 兼容 sort 参数（latest、downloads、rating、name）
	if sort := c.Query("sort"); sort != "" {
		switch sort {
		case "latest":
			params.SortBy = "created_at"
		case "downloads":
			params.SortBy = "downloads"
		case "rating":
			params.SortBy = "rating"
		case "name":
			params.SortBy = "name"
		default:
			params.SortBy = sort
		}
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

	// 解析分类ID
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			params.CategoryID = &categoryID
		}
	}

	// 解析是否精选插件
	if featuredStr := c.Query("featured"); featuredStr != "" {
		if featured, err := strconv.ParseBool(featuredStr); err == nil {
			params.Featured = &featured
		}
	}

	// 解析价格范围
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil && minPrice >= 0 {
			params.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && maxPrice >= 0 {
			params.MaxPrice = &maxPrice
		}
	}

	// 解析最低评分
	if minRatingStr := c.Query("min_rating"); minRatingStr != "" {
		if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil && minRating >= 0 && minRating <= 5 {
			params.MinRating = &minRating
		}
	}

	// 获取插件列表
	plugins, meta, err := h.pluginService.GetPlugins(params)
	if err != nil {
		// 记录详细错误信息
		logger.Error("Failed to get plugins: ", err.Error())
		utils.InternalServerErrorResponse(c, "Failed to get plugins")
		return
	}

	utils.PaginatedResponse(c, plugins, meta)
}

// GetPlugin 获取单个插件详情
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	plugin, err := h.pluginService.GetPlugin(id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugin")
		return
	}

	if plugin == nil {
		utils.NotFoundResponse(c, "Plugin not found")
		return
	}

	utils.SuccessResponse(c, plugin)
}

// GetPluginsByCategory 根据分类获取插件
func (h *PluginHandler) GetPluginsByCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category ID")
		return
	}

	// 解析查询参数
	params := models.PluginSearchParams{
		Page:       1,
		Limit:      20,
		CategoryID: &categoryID,
		Query:      c.Query("q"),
		SortBy:     c.Query("sort_by"),
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

	// 获取插件列表
	plugins, meta, err := h.pluginService.GetPluginsByCategory(categoryID, params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugins")
		return
	}

	utils.PaginatedResponse(c, plugins, meta)
}

// GetUserFavorites 获取用户收藏的插件
func (h *PluginHandler) GetUserFavorites(c *gin.Context) {
	// 获取用户ID
	userUUID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.UnauthorizedResponse(c)
		return
	}

	// 解析分页参数
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	plugins, meta, err := h.pluginService.GetUserFavorites(userUUID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user favorites")
		return
	}

	utils.PaginatedResponse(c, plugins, meta)
}

// PurchasePlugin 购买插件
func (h *PluginHandler) PurchasePlugin(c *gin.Context) {
	// 获取用户ID
	userUUID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.UnauthorizedResponse(c)
		return
	}

	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	err = h.pluginService.PurchasePlugin(userUUID, pluginID)
	if err != nil {
		if err.Error() == "plugin already purchased" {
			utils.BadRequestResponse(c, "Plugin already purchased")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to purchase plugin")
		return
	}

	utils.SuccessResponseWithMessage(c, "Plugin purchased successfully", nil)
}

// GetUserPurchases 获取用户购买的插件
func (h *PluginHandler) GetUserPurchases(c *gin.Context) {
	// 获取用户ID
	userUUID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.UnauthorizedResponse(c)
		return
	}

	// 解析分页参数
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	purchases, meta, err := h.pluginService.GetUserPurchases(userUUID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user purchases")
		return
	}

	utils.PaginatedResponse(c, purchases, meta)
}

// GetMarketplacePluginStats 市场插件统计（简化聚合）
func (h *PluginHandler) GetMarketplacePluginStats(c *gin.Context) {
	params := models.PluginSearchParams{
		Page:  1,
		Limit: 1,
	}
	_, meta, err := h.pluginService.GetPlugins(params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugin stats")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     meta.Total,
		"installed": 0,
		"updated":   0,
	})
}

// CreatePlugin 创建插件
func (h *PluginHandler) CreatePlugin(c *gin.Context) {
	var plugin models.Plugin
	if err := c.ShouldBindJSON(&plugin); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if plugin.Name == "" || plugin.Description == nil || *plugin.Description == "" || plugin.Author == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Name, description, and author are required")
		return
	}

	if err := h.pluginService.CreatePlugin(&plugin); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create plugin")
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Code:      http.StatusCreated,
		Message:   "Plugin created successfully",
		Data:      plugin,
		Timestamp: time.Now(),
	})
}

// UpdatePlugin 更新插件
func (h *PluginHandler) UpdatePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plugin ID")
		return
	}

	var plugin models.Plugin
	if err := c.ShouldBindJSON(&plugin); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if plugin.Name == "" || plugin.Description == nil || *plugin.Description == "" || plugin.Author == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Name, description, and author are required")
		return
	}

	if err := h.pluginService.UpdatePlugin(id, &plugin); err != nil {
		if err.Error() == "plugin not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Plugin not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update plugin")
		return
	}

	utils.SuccessResponseWithMessage(c, "Plugin updated successfully", nil)
}

// DeletePlugin 删除插件
func (h *PluginHandler) DeletePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plugin ID")
		return
	}

	if err := h.pluginService.DeletePlugin(id); err != nil {
		if err.Error() == "plugin not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Plugin not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete plugin")
		return
	}

	utils.SuccessResponseWithMessage(c, "Plugin deleted successfully", nil)
}

// UpdatePluginStatus 更新插件状态
func (h *PluginHandler) UpdatePluginStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plugin ID")
		return
	}

	var request struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.pluginService.UpdatePluginStatus(id, request.IsActive); err != nil {
		if err.Error() == "plugin not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Plugin not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update plugin status")
		return
	}

	utils.SuccessResponseWithMessage(c, "Plugin status updated successfully", nil)
}

// ToggleFavorite 切换收藏状态
func (h *PluginHandler) ToggleFavorite(c *gin.Context) {
	// 获取用户ID
	userUUID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.UnauthorizedResponse(c)
		return
	}

	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	isFavorited, err := h.pluginService.ToggleFavorite(userUUID, pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to toggle favorite")
		return
	}

	message := "Plugin removed from favorites"
	if isFavorited {
		message = "Plugin added to favorites"
	}

	utils.SuccessResponseWithMessage(c, message, gin.H{
		"is_favorited": isFavorited,
	})
}
