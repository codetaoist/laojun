package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/codetaoist/laojun-shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeveloperHandler 开发者处理器
type DeveloperHandler struct {
	developerService *services.DeveloperService
}

// NewDeveloperHandler 创建开发者处理器
func NewDeveloperHandler(developerService *services.DeveloperService) *DeveloperHandler {
	return &DeveloperHandler{
		developerService: developerService,
	}
}

// RegisterDeveloperRequest 注册开发者请求
type RegisterDeveloperRequest struct {
	CompanyName *string `json:"company_name"`
	Website     *string `json:"website"`
	Description *string `json:"description"`
}

// UpdateDeveloperRequest 更新开发者请求
type UpdateDeveloperRequest struct {
	CompanyName *string `json:"company_name"`
	Website     *string `json:"website"`
	Description *string `json:"description"`
}

// RegisterDeveloper 注册开发者
func (h *DeveloperHandler) RegisterDeveloper(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	// 绑定请求体
	var req RegisterDeveloperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// 验证网站URL格式
	if req.Website != nil && *req.Website != "" {
		website := strings.TrimSpace(*req.Website)
		if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
			website = "https://" + website
		}
		req.Website = &website
	}

	// 注册开发者
	developer, err := h.developerService.RegisterDeveloper(userUUID, req.CompanyName, req.Website, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "already a developer") {
			utils.ErrorResponse(c, http.StatusConflict, "User is already a developer")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register developer: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Developer registered successfully",
		"data":    developer,
	})
}

// GetDeveloper 获取开发者信息
func (h *DeveloperHandler) GetDeveloper(c *gin.Context) {
	// 解析开发者ID
	developerIDStr := c.Param("id")
	developerID, err := uuid.Parse(developerIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid developer ID format")
		return
	}

	// 获取开发者信息
	developer, err := h.developerService.GetDeveloper(developerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developer: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    developer,
	})
}

// GetMyDeveloperProfile 获取当前用户的开发者信息
func (h *DeveloperHandler) GetMyDeveloperProfile(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	// 获取开发者信息
	developer, err := h.developerService.GetDeveloperByUserID(userUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer profile not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developer profile: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    developer,
	})
}

// UpdateDeveloper 更新开发者信息
func (h *DeveloperHandler) UpdateDeveloper(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	// 获取开发者信息以验证权限
	developer, err := h.developerService.GetDeveloperByUserID(userUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer profile not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developer profile: "+err.Error())
		return
	}

	// 绑定请求体
	var req UpdateDeveloperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// 验证网站URL格式
	if req.Website != nil && *req.Website != "" {
		website := strings.TrimSpace(*req.Website)
		if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
			website = "https://" + website
		}
		req.Website = &website
	}

	// 更新开发者信息
	updatedDeveloper, err := h.developerService.UpdateDeveloper(developer.ID, req.CompanyName, req.Website, req.Description)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update developer: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Developer updated successfully",
		"data":    updatedDeveloper,
	})
}

// GetDeveloperStats 获取开发者统计信息
func (h *DeveloperHandler) GetDeveloperStats(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	// 获取开发者信息
	developer, err := h.developerService.GetDeveloperByUserID(userUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer profile not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developer profile: "+err.Error())
		return
	}

	// 获取统计信息
	stats, err := h.developerService.GetDeveloperStats(developer.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developer stats: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetDeveloperPlugins 获取开发者的插件列表
func (h *DeveloperHandler) GetDeveloperPlugins(c *gin.Context) {
	// 解析开发者ID
	developerIDStr := c.Param("id")
	developerID, err := uuid.Parse(developerIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid developer ID format")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 获取开发者插件列表
	plugins, meta, err := h.developerService.GetDeveloperPlugins(developerID, page, limit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developer plugins: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plugins,
		"meta":    meta,
	})
}

// GetMyPlugins 获取当前开发者的插件列表
func (h *DeveloperHandler) GetMyPlugins(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	// 获取开发者信息
	developer, err := h.developerService.GetDeveloperByUserID(userUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer profile not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developer profile: "+err.Error())
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 获取开发者插件列表
	plugins, meta, err := h.developerService.GetDeveloperPlugins(developer.ID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get plugins: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plugins,
		"meta":    meta,
	})
}

// VerifyDeveloper 验证开发者（管理员操作）
func (h *DeveloperHandler) VerifyDeveloper(c *gin.Context) {
	// 解析开发者ID
	developerIDStr := c.Param("id")
	developerID, err := uuid.Parse(developerIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid developer ID format")
		return
	}

	// 验证开发者
	err = h.developerService.VerifyDeveloper(developerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify developer: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Developer verified successfully",
	})
}

// DeactivateDeveloper 停用开发者（管理员操作）
func (h *DeveloperHandler) DeactivateDeveloper(c *gin.Context) {
	// 解析开发者ID
	developerIDStr := c.Param("id")
	developerID, err := uuid.Parse(developerIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid developer ID format")
		return
	}

	// 停用开发者
	err = h.developerService.DeactivateDeveloper(developerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Developer not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to deactivate developer: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Developer deactivated successfully",
	})
}

// GetDevelopers 获取开发者列表（管理员功能）
func (h *DeveloperHandler) GetDevelopers(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 解析verified参数
	var verified *bool
	if verifiedStr := c.Query("verified"); verifiedStr != "" {
		if v, err := strconv.ParseBool(verifiedStr); err == nil {
			verified = &v
		}
	}

	// 获取开发者列表
	developers, meta, err := h.developerService.GetDevelopers(page, limit, verified)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get developers: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    developers,
		"meta":    meta,
	})
}
