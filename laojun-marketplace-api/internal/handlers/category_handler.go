package handlers

import (
	"net/http"
	"time"

	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/codetaoist/laojun-shared/logger"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/codetaoist/laojun-shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CategoryHandler 分类处理
type CategoryHandler struct {
	categoryService *services.CategoryService
}

// NewCategoryHandler 创建分类处理
func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories 获取所有分类
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	// 检查是否需要统计信息
	withStats := c.Query("with_stats") == "true"

	if withStats {
		categories, err := h.categoryService.GetCategoriesWithStats()
		if err != nil {
			// 记录详细错误信息
			logger.Error("Failed to get categories with stats: ", err.Error())
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get categories with stats")
			return
		}
		utils.SuccessResponse(c, categories)
	} else {
		categories, err := h.categoryService.GetCategories()
		if err != nil {
			// 记录详细错误信息
			logger.Error("Failed to get categories: ", err.Error())
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get categories")
			return
		}
		utils.SuccessResponse(c, categories)
	}
}

// GetCategory 获取单个分类详情
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	category, err := h.categoryService.GetCategory(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get category")
		return
	}

	if category == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
		return
	}

	utils.SuccessResponse(c, category)
}

// CreateCategory 创建分类
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if category.Name == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Category name is required")
		return
	}

	var description, icon string
	if category.Description != nil {
		description = *category.Description
	}
	if category.Icon != nil {
		icon = *category.Icon
	}
	createdCategory, err := h.categoryService.CreateCategory(category.Name, description, icon)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create category")
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Code:      http.StatusCreated,
		Message:   "Category created successfully",
		Data:      createdCategory,
		Timestamp: time.Now(),
	})
}

// UpdateCategory 更新分类
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if category.Name == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Category name is required")
		return
	}

	var description, icon string
	if category.Description != nil {
		description = *category.Description
	}
	if category.Icon != nil {
		icon = *category.Icon
	}
	updatedCategory, err := h.categoryService.UpdateCategory(id, category.Name, description, icon)
	if err != nil {
		if err.Error() == "category not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update category")
		return
	}

	utils.SuccessResponseWithMessage(c, "Category updated successfully", updatedCategory)
}

// DeleteCategory 删除分类
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := h.categoryService.DeleteCategory(id); err != nil {
		if err.Error() == "category not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
			return
		}
		// 检查是否是因为有插件使用此分类而无法删除
		if err.Error() != "" && err.Error() != "category not found" {
			utils.ErrorResponse(c, http.StatusConflict, err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	utils.SuccessResponseWithMessage(c, "Category deleted successfully", nil)
}
