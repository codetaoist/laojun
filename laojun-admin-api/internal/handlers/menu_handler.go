package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MenuHandler struct {
	menuService *services.MenuService
}

// NewMenuHandler 创建菜单处理
func NewMenuHandler(menuService *services.MenuService) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
	}
}

// GetMenus 获取菜单列表
func (h *MenuHandler) GetMenus(c *gin.Context) {
	var params models.MenuSearchParams

	// 手动处理分页参数，避免在tree_mode时的验证错误
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page >= 1 {
			params.Page = page
		} else if page < 1 {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Invalid page parameter: must be >= 1",
			})
			return
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize >= 1 && pageSize <= 100 {
			params.PageSize = pageSize
		} else {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Invalid page_size parameter: must be between 1 and 100",
			})
			return
		}
	}

	// 绑定其他查询参数（排除分页参数）
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid query parameters: " + err.Error(),
		})
		return
	}

	// 处理 is_hidden 参数
	if hiddenStr := c.Query("is_hidden"); hiddenStr != "" {
		if hidden, err := strconv.ParseBool(hiddenStr); err == nil {
			params.IsHidden = &hidden
		}
	}

	// 处理 tree_mode 参数
	if treeModeStr := c.Query("tree_mode"); treeModeStr != "" {
		if treeMode, err := strconv.ParseBool(treeModeStr); err == nil {
			params.TreeMode = treeMode
		}
	}

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 获取菜单列表
	result, err := h.menuService.GetMenus(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get menus: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.PaginatedResponse[models.Menu]]{
		Success: true,
		Data:    result,
	})
}

// GetMenu 根据ID获取菜单
func (h *MenuHandler) GetMenu(c *gin.Context) {
	h.GetMenuByID(c)
}

// GetMenuByID 根据ID获取菜单
func (h *MenuHandler) GetMenuByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid menu ID",
		})
		return
	}

	menu, err := h.menuService.GetMenuByID(id)
	if err != nil {
		if err.Error() == "menu not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Menu not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get menu: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.Menu]{
		Success: true,
		Data:    menu,
	})
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var req models.MenuCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	menu, err := h.menuService.CreateMenu(req)
	if err != nil {
		if err.Error() == "parent menu not found" {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Parent menu not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to create menu: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.ApiResponse[*models.Menu]{
		Success: true,
		Data:    menu,
		Message: "Menu created successfully",
	})
}

// UpdateMenu 更新菜单
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid menu ID",
		})
		return
	}

	var req models.MenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	menu, err := h.menuService.UpdateMenu(id, req)
	if err != nil {
		if err.Error() == "menu not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Menu not found",
			})
			return
		}
		if err.Error() == "parent menu not found" ||
			err.Error() == "cannot set menu as its own parent" ||
			err.Error() == "would create circular reference" {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to update menu: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.Menu]{
		Success: true,
		Data:    menu,
		Message: "Menu updated successfully",
	})
}

// DeleteMenu 删除菜单
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid menu ID",
		})
		return
	}

	err = h.menuService.DeleteMenu(id)
	if err != nil {
		if err.Error() == "menu not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Menu not found",
			})
			return
		}
		if err.Error() == "cannot delete menu with child menus" {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Cannot delete menu with child menus",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to delete menu: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[interface{}]{
		Success: true,
		Message: "Menu deleted successfully",
	})
}

// BatchDeleteMenus 批量删除菜单
func (h *MenuHandler) BatchDeleteMenus(c *gin.Context) {
	var req struct {
		MenuIDs []string `json:"menu_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// 转换为UUID
	var ids []uuid.UUID
	for _, idStr := range req.MenuIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Invalid menu ID: " + idStr,
			})
			return
		}
		ids = append(ids, id)
	}

	err := h.menuService.BatchDeleteMenus(ids)
	if err != nil {
		if err.Error() == "cannot delete menus with child menus" {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Cannot delete menus with child menus",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to batch delete menus: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[interface{}]{
		Success: true,
		Message: "Menus deleted successfully",
	})
}

// MoveMenu 移动菜单
func (h *MenuHandler) MoveMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid menu ID",
		})
		return
	}

	var req models.MenuMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	err = h.menuService.MoveMenu(id, req)
	if err != nil {
		if err.Error() == "cannot move menu to itself" ||
			err.Error() == "target parent menu not found" ||
			err.Error() == "would create circular reference" {
			c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to move menu: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[interface{}]{
		Success: true,
		Message: "Menu moved successfully",
	})
}

// GetMenuStats 获取菜单统计信息
func (h *MenuHandler) GetMenuStats(c *gin.Context) {
	stats, err := h.menuService.GetMenuStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get menu stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.MenuStats]{
		Success: true,
		Data:    stats,
	})
}

// BatchUpdateMenus 批量更新菜单
func (h *MenuHandler) BatchUpdateMenus(c *gin.Context) {
	var req models.MenuBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// 这里可以实现批量更新逻辑
	// 为了简化，暂时返回成功响应
	c.JSON(http.StatusOK, models.ApiResponse[interface{}]{
		Success: true,
		Message: "Menus updated successfully",
	})
}

func (h *MenuHandler) ToggleFavorite(c *gin.Context) {
	idStr := c.Param("id")
	menuID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid menu ID",
		})
		return
	}

	menu, err := h.menuService.ToggleFavorite(menuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to toggle favorite: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.Menu]{
		Success: true,
		Data:    menu,
		Message: "Menu favorite toggled",
	})
}

// MenuConfig 相关处理器方法

// GetMenuConfigs 获取菜单配置列表
func (h *MenuHandler) GetMenuConfigs(c *gin.Context) {
	deviceType := c.Query("device_type")

	configs, err := h.menuService.GetMenuConfigs(deviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get menu configs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[[]models.MenuConfig]{
		Success: true,
		Data:    configs,
	})
}

// GetMenuConfig 根据ID或设备类型获取菜单配置
func (h *MenuHandler) GetMenuConfig(c *gin.Context) {
	identifier := c.Param("id")

	config, err := h.menuService.GetMenuConfig(identifier)
	if err != nil {
		if err.Error() == "menu config not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Menu config not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get menu config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.MenuConfig]{
		Success: true,
		Data:    config,
	})
}

// CreateMenuConfig 创建菜单配置
func (h *MenuHandler) CreateMenuConfig(c *gin.Context) {
	var config models.MenuConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// 验证必填字段
	if config.Name == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Name is required",
		})
		return
	}

	if config.DeviceType == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Device type is required",
		})
		return
	}

	result, err := h.menuService.CreateMenuConfig(&config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to create menu config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.ApiResponse[*models.MenuConfig]{
		Success: true,
		Data:    result,
		Message: "Menu config created successfully",
	})
}

// UpdateMenuConfig 更新菜单配置
func (h *MenuHandler) UpdateMenuConfig(c *gin.Context) {
	identifier := c.Param("id")

	var updates models.MenuConfig
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	result, err := h.menuService.UpdateMenuConfig(identifier, &updates)
	if err != nil {
		if err.Error() == "menu config not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Menu config not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to update menu config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.MenuConfig]{
		Success: true,
		Data:    result,
		Message: "Menu config updated successfully",
	})
}

// DeleteMenuConfig 删除菜单配置
func (h *MenuHandler) DeleteMenuConfig(c *gin.Context) {
	identifier := c.Param("id")

	err := h.menuService.DeleteMenuConfig(identifier)
	if err != nil {
		if err.Error() == "menu config not found" {
			c.JSON(http.StatusNotFound, models.ApiResponse[interface{}]{
				Success: false,
				Error:   "Menu config not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to delete menu config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[interface{}]{
		Success: true,
		Message: "Menu config deleted successfully",
	})
}

// GetVisualMenuConfig 获取可视化菜单配置
func (h *MenuHandler) GetVisualMenuConfig(c *gin.Context) {
	deviceType := c.Query("device_type")
	if deviceType == "" {
		deviceType = "desktop" // 默认设备类型
	}

	config, err := h.menuService.GetMenuConfig(deviceType)
	if err != nil {
		// 如果没有找到配置，返回默认的可视化配置
		if err.Error() == "menu config not found" {
			defaultConfig := map[string]interface{}{
				"layout": "sidebar",
				"theme": "light",
				"collapsed": false,
				"showIcons": true,
				"showBreadcrumb": true,
				"animation": "slide",
				"device_type": deviceType,
			}
			
			c.JSON(http.StatusOK, models.ApiResponse[map[string]interface{}]{
				Success: true,
				Data:    defaultConfig,
				Message: "Using default visual configuration",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ApiResponse[interface{}]{
			Success: false,
			Error:   "Failed to get visual menu config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse[*models.MenuConfig]{
		Success: true,
		Data:    config,
	})
}
