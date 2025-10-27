package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-marketplace-api/internal/models"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IconHandler struct {
	iconService *services.IconService
}

func NewIconHandler(iconService *services.IconService) *IconHandler {
	return &IconHandler{
		iconService: iconService,
	}
}

// GetIcons 获取图标列表
func (h *IconHandler) GetIcons(c *gin.Context) {
	var params models.IconSearchParams

	// 绑定查询参数
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid query parameters: " + err.Error(),
		})
		return
	}

	// 处理is_active参数
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			params.IsActive = &isActive
		}
	}

	// 获取图标列表
	result, err := h.iconService.GetIcons(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get icons: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.PaginatedResponse[models.Icon]]{
		Success: true,
		Data:    result,
	})
}

// GetIcon 根据ID获取图标
func (h *IconHandler) GetIcon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid icon ID",
		})
		return
	}

	icon, err := h.iconService.GetIconByID(id)
	if err != nil {
		if err.Error() == "icon not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Icon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get icon: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.Icon]{
		Success: true,
		Data:    icon,
	})
}

// CreateIcon 创建图标
func (h *IconHandler) CreateIcon(c *gin.Context) {
	var req models.CreateIconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	icon, err := h.iconService.CreateIcon(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to create icon: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.ApiResponse[*models.Icon]{
		Success: true,
		Data:    icon,
	})
}

// UpdateIcon 更新图标
func (h *IconHandler) UpdateIcon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid icon ID",
		})
		return
	}

	var req models.UpdateIconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	icon, err := h.iconService.UpdateIcon(id, req)
	if err != nil {
		if err.Error() == "icon not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Icon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to update icon: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.Icon]{
		Success: true,
		Data:    icon,
	})
}

// DeleteIcon 删除图标
func (h *IconHandler) DeleteIcon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid icon ID",
		})
		return
	}

	err = h.iconService.DeleteIcon(id)
	if err != nil {
		if err.Error() == "icon not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Icon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to delete icon: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[interface{}]{
		Success: true,
		Data:    nil,
	})
}

// GetIconStats 获取图标统计信息
func (h *IconHandler) GetIconStats(c *gin.Context) {
	stats, err := h.iconService.GetIconStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get icon stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.IconStats]{
		Success: true,
		Data:    stats,
	})
}

// GetIconCategories 获取图标分类列表
func (h *IconHandler) GetIconCategories(c *gin.Context) {
	categories, err := h.iconService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get icon categories: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[[]string]{
		Success: true,
		Data:    categories,
	})
}
