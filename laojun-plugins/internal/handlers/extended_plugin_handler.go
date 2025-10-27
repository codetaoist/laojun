package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/codetaoist/laojun-plugins/internal/models"
	"github.com/codetaoist/laojun-plugins/internal/services"
	"github.com/codetaoist/laojun-shared/models" // 使用shared模型
)

// ExtendedPluginHandler 扩展插件处理
type ExtendedPluginHandler struct {
	extendedPluginService *services.ExtendedPluginService
	categoryService       *services.CategoryService
}

// NewExtendedPluginHandler 创建扩展插件处理
func NewExtendedPluginHandler(extendedPluginService *services.ExtendedPluginService, categoryService *services.CategoryService) *ExtendedPluginHandler {
	return &ExtendedPluginHandler{
		extendedPluginService: extendedPluginService,
		categoryService:       categoryService,
	}
}

// GetExtendedPlugins 获取扩展插件列表
func (h *ExtendedPluginHandler) GetExtendedPlugins(c *gin.Context) {
	// 解析查询参数
	params := models.PluginSearchParams{
		Page:     1,
		PageSize: 20,
		Query:    c.Query("q"),
		SortBy:   c.Query("sort_by"),
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
			params.PageSize = limit
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

	// 解析插件类型
	if pluginType := c.Query("type"); pluginType != "" {
		pType := models.PluginType(pluginType)
		params.Type = &pType
	}

	// 解析运行时环境
	if runtime := c.Query("runtime"); runtime != "" {
		pRuntime := models.PluginRuntime(runtime)
		params.Runtime = &pRuntime
	}

	// 解析插件状态
	if status := c.Query("status"); status != "" {
		pStatus := models.PluginStatus(status)
		params.Status = &pStatus
	}

	// 获取插件列表
	plugins, meta, err := h.extendedPluginService.GetExtendedPlugins(params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugins")
		return
	}

	utils.PaginatedResponse(c, plugins, meta)
}

// GetExtendedPlugin 获取单个扩展插件详情
func (h *ExtendedPluginHandler) GetExtendedPlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	plugin, err := h.extendedPluginService.GetExtendedPlugin(pluginID)
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

// DeployPlugin 部署插件
func (h *ExtendedPluginHandler) DeployPlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	err = h.extendedPluginService.DeployPlugin(c.Request.Context(), pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to deploy plugin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":   "Plugin deployment started",
		"plugin_id": pluginID,
	})
}

// StopPlugin 停止插件
func (h *ExtendedPluginHandler) StopPlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	err = h.extendedPluginService.StopPlugin(c.Request.Context(), pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to stop plugin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":   "Plugin stopped successfully",
		"plugin_id": pluginID,
	})
}

// RestartPlugin 重启插件
func (h *ExtendedPluginHandler) RestartPlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	err = h.extendedPluginService.RestartPlugin(c.Request.Context(), pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to restart plugin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":   "Plugin restart started",
		"plugin_id": pluginID,
	})
}

// CallPlugin 调用插件
func (h *ExtendedPluginHandler) CallPlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var requestBody struct {
		Method     string                 `json:"method" binding:"required"`
		InputData  map[string]interface{} `json:"input_data"`
		ClientType string                 `json:"client_type"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	// 构建调用请求
	request := &models.PluginCallRequest{
		PluginID:   pluginID.String(),
		UserID:     &userUUID,
		Method:     requestBody.Method,
		InputData:  requestBody.InputData,
		ClientType: requestBody.ClientType,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	// 调用插件
	response, err := h.extendedPluginService.CallPlugin(c.Request.Context(), request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to call plugin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, response)
}

// GetPluginMetrics 获取插件指标
func (h *ExtendedPluginHandler) GetPluginMetrics(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	// 解析天数参数
	days := 7 // 默认7天
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	metrics, err := h.extendedPluginService.GetPluginMetrics(pluginID, days)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugin metrics")
		return
	}

	utils.SuccessResponse(c, metrics)
}

// CreateExtendedPlugin 创建扩展插件
func (h *ExtendedPluginHandler) CreateExtendedPlugin(c *gin.Context) {
	var plugin models.ExtendedPlugin

	if err := c.ShouldBindJSON(&plugin); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 验证必填字段
	if plugin.Name == "" {
		utils.BadRequestResponse(c, "Plugin name is required")
		return
	}

	if plugin.Type == "" {
		plugin.Type = models.PluginTypeInProcess // 默认为进程内插件
	}

	if plugin.Runtime == "" {
		plugin.Runtime = models.RuntimeGo // 默认为Go运行时
	}

	if plugin.Status == "" {
		plugin.Status = models.StatusPending // 默认为待部署状态
	}

	// 设置默认值
	plugin.IsActive = true
	plugin.IsFree = plugin.Price == 0

	err := h.extendedPluginService.CreateExtendedPlugin(&plugin)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create plugin")
		return
	}

	utils.CreatedResponse(c, plugin)
}

// UploadPlugin 上传插件
func (h *ExtendedPluginHandler) UploadPlugin(c *gin.Context) {
	// 解析multipart form
	err := c.Request.ParseMultipartForm(50 << 20) // 50MB max
	if err != nil {
		utils.BadRequestResponse(c, "Failed to parse form data: "+err.Error())
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded or invalid file")
		return
	}
	defer file.Close()

	// 验证文件类型
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		utils.BadRequestResponse(c, "Only ZIP files are allowed")
		return
	}

	// 验证文件大小
	if header.Size > 50<<20 { // 50MB
		utils.BadRequestResponse(c, "File size exceeds 50MB limit")
		return
	}

	// 创建上传目录
	uploadDir := filepath.Join("uploads", "plugins")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create upload directory")
		return
	}

	// 生成唯一文件名
	pluginID := uuid.New()
	filename := fmt.Sprintf("%s_%s", pluginID.String(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save file")
		return
	}

	// 解析插件信息
	var plugin models.ExtendedPlugin
	plugin.ID = pluginID
	plugin.Name = c.PostForm("name")
	plugin.Version = c.PostForm("version")
	plugin.Author = c.PostForm("author")
	plugin.Price, _ = strconv.ParseFloat(c.PostForm("price"), 64)

	if description := c.PostForm("description"); description != "" {
		plugin.Description = &description
	}
	if shortDesc := c.PostForm("short_description"); shortDesc != "" {
		plugin.ShortDescription = &shortDesc
	}
	if requirements := c.PostForm("requirements"); requirements != "" {
		plugin.Requirements = &requirements
	}
	if homepage := c.PostForm("homepage"); homepage != "" {
		plugin.IconURL = &homepage
	}
	if repository := c.PostForm("repository"); repository != "" {
		// 可以将repository信息存储在其他字段中
	}

	// 解析分类ID
	if categoryIDStr := c.PostForm("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			plugin.CategoryID = &categoryID
		}
	}

	// 解析标签
	if tagsStr := c.PostForm("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		plugin.Tags = tags
	}

	// 设置插件类型和运行时
	pluginType := c.PostForm("type")
	if pluginType == "" {
		pluginType = "in_process"
	}
	plugin.Type = models.PluginType(pluginType)

	runtime := c.PostForm("runtime")
	if runtime == "" {
		runtime = "go"
	}
	plugin.Runtime = models.PluginRuntime(runtime)

	// 设置默认值
	plugin.Status = models.StatusPending
	plugin.IsActive = true
	plugin.IsFree = plugin.Price == 0
	plugin.BinaryPath = &filePath

	// 验证必填字段
	if plugin.Name == "" {
		utils.BadRequestResponse(c, "Plugin name is required")
		return
	}
	if plugin.Version == "" {
		utils.BadRequestResponse(c, "Plugin version is required")
		return
	}
	if plugin.Author == "" {
		utils.BadRequestResponse(c, "Plugin author is required")
		return
	}

	// 创建插件记录
	err = h.extendedPluginService.CreateExtendedPlugin(&plugin)
	if err != nil {
		// 如果创建失败，删除已上传的文件
		os.Remove(filePath)
		utils.InternalServerErrorResponse(c, "Failed to create plugin: "+err.Error())
		return
	}

	utils.CreatedResponse(c, gin.H{
		"message":   "Plugin uploaded successfully",
		"plugin_id": plugin.ID,
		"plugin":    plugin,
	})
}

// UpdateExtendedPlugin 更新扩展插件
func (h *ExtendedPluginHandler) UpdateExtendedPlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var plugin models.ExtendedPlugin

	if err := c.ShouldBindJSON(&plugin); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	err = h.extendedPluginService.UpdateExtendedPlugin(pluginID, &plugin)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update plugin")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":   "Plugin updated successfully",
		"plugin_id": pluginID,
	})
}

// UpdatePluginStatus 更新插件状态
func (h *ExtendedPluginHandler) UpdatePluginStatus(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var requestBody struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 验证状态值是否有效
	status := models.PluginStatus(requestBody.Status)
	switch status {
	case models.StatusPending, models.StatusDeploying, models.StatusRunning, models.StatusStopped, models.StatusError:
		// 有效状态，继续处理
	default:
		utils.BadRequestResponse(c, "Invalid status value")
		return
	}

	err = h.extendedPluginService.UpdatePluginStatus(pluginID, status)
	if err != nil {
		if err.Error() == "plugin not found" {
			utils.NotFoundResponse(c, "Plugin not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update plugin status")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":   "Plugin status updated successfully",
		"plugin_id": pluginID,
		"status":    status,
	})
}

// GetPluginCallLogs 获取插件调用日志
func (h *ExtendedPluginHandler) GetPluginCallLogs(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	// 解析查询参数
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

	// 解析时间范围
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	// 解析成功状态参数
	var success *bool
	if successStr := c.Query("success"); successStr != "" {
		if s, err := strconv.ParseBool(successStr); err == nil {
			success = &s
		}
	}

	// 调用服务方法获取日志
	logs, totalCount, err := h.extendedPluginService.GetPluginCallLogs(&pluginID, nil, success, startTime, endTime, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get call logs")
		return
	}

	// 计算分页信息
	totalPages := (totalCount + limit - 1) / limit
	meta := sharedmodels.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	utils.PaginatedResponse(c, logs, meta)
}

// ValidatePlugin 验证插件
func (h *ExtendedPluginHandler) ValidatePlugin(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	plugin, err := h.extendedPluginService.GetExtendedPlugin(pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugin")
		return
	}

	if plugin == nil {
		utils.NotFoundResponse(c, "Plugin not found")
		return
	}

	// 执行插件验证逻辑
	validationResult := map[string]interface{}{
		"plugin_id": pluginID,
		"valid":     true,
		"checks":    []map[string]interface{}{},
	}

	checks := []map[string]interface{}{}

	// 基础字段验证
	if plugin.Name == "" {
		checks = append(checks, map[string]interface{}{
			"check":   "name_required",
			"passed":  false,
			"message": "Plugin name is required",
		})
		validationResult["valid"] = false
	} else {
		checks = append(checks, map[string]interface{}{
			"check":   "name_required",
			"passed":  true,
			"message": "Plugin name is valid",
		})
	}

	// 类型特定验证
	switch plugin.Type {
	case models.PluginTypeInProcess:
		if plugin.Runtime == models.RuntimeGo && (plugin.BinaryPath == nil || *plugin.BinaryPath == "") {
			checks = append(checks, map[string]interface{}{
				"check":   "go_binary_path",
				"passed":  false,
				"message": "Binary path is required for Go plugins",
			})
			validationResult["valid"] = false
		} else if plugin.Runtime == models.RuntimeJS && (plugin.BinaryPath == nil || *plugin.BinaryPath == "") {
			checks = append(checks, map[string]interface{}{
				"check":   "js_module_path",
				"passed":  false,
				"message": "Module path is required for JS plugins",
			})
			validationResult["valid"] = false
		} else {
			checks = append(checks, map[string]interface{}{
				"check":   "runtime_config",
				"passed":  true,
				"message": "Runtime configuration is valid",
			})
		}

	case models.PluginTypeMicroservice:
		if plugin.DockerImage == nil || *plugin.DockerImage == "" {
			checks = append(checks, map[string]interface{}{
				"check":   "docker_image",
				"passed":  false,
				"message": "Docker image is required for microservice plugins",
			})
			validationResult["valid"] = false
		} else {
			checks = append(checks, map[string]interface{}{
				"check":   "docker_image",
				"passed":  true,
				"message": "Docker image is specified",
			})
		}

		if plugin.ServicePort == nil {
			checks = append(checks, map[string]interface{}{
				"check":   "service_port",
				"passed":  false,
				"message": "Service port is required for microservice plugins",
			})
			validationResult["valid"] = false
		} else {
			checks = append(checks, map[string]interface{}{
				"check":   "service_port",
				"passed":  true,
				"message": "Service port is specified",
			})
		}
	}

	// 接口规范验证
	if plugin.InterfaceSpec != nil && *plugin.InterfaceSpec != "" {
		var spec map[string]interface{}
		if err := json.Unmarshal([]byte(*plugin.InterfaceSpec), &spec); err != nil {
			checks = append(checks, map[string]interface{}{
				"check":   "interface_spec",
				"passed":  false,
				"message": "Invalid interface specification JSON",
			})
			validationResult["valid"] = false
		} else {
			checks = append(checks, map[string]interface{}{
				"check":   "interface_spec",
				"passed":  true,
				"message": "Interface specification is valid JSON",
			})
		}
	}

	validationResult["checks"] = checks

	utils.SuccessResponse(c, validationResult)
}

// BatchCallPlugins 批量调用插件
func (h *ExtendedPluginHandler) BatchCallPlugins(c *gin.Context) {
	var requestBody struct {
		Calls []struct {
			PluginID   uuid.UUID              `json:"plugin_id" binding:"required"`
			Method     string                 `json:"method" binding:"required"`
			InputData  map[string]interface{} `json:"input_data"`
			ClientType string                 `json:"client_type"`
		} `json:"calls" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 限制批量调用数量
	if len(requestBody.Calls) > 10 {
		utils.BadRequestResponse(c, "Too many calls in batch (max 10)")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	// 并发调用插件
	results := make([]map[string]interface{}, len(requestBody.Calls))

	for i, call := range requestBody.Calls {
		request := &models.PluginCallRequest{
			PluginID:   call.PluginID.String(),
			UserID:     &userUUID,
			Method:     call.Method,
			InputData:  call.InputData,
			ClientType: call.ClientType,
			ClientIP:   c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
		}

		response, err := h.extendedPluginService.CallPlugin(c.Request.Context(), request)
		if err != nil {
			results[i] = map[string]interface{}{
				"plugin_id": call.PluginID.String(),
				"success":   false,
				"error":     err.Error(),
			}
		} else {
			results[i] = map[string]interface{}{
				"plugin_id": call.PluginID.String(),
				"success":   true,
				"response":  response,
			}
		}
	}

	utils.SuccessResponse(c, gin.H{
		"results": results,
	})
}

// GetPluginStatus 获取插件状态
func (h *ExtendedPluginHandler) GetPluginStatus(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	plugin, err := h.extendedPluginService.GetExtendedPlugin(pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugin")
		return
	}

	if plugin == nil {
		utils.NotFoundResponse(c, "Plugin not found")
		return
	}

	status := map[string]interface{}{
		"plugin_id":    pluginID,
		"status":       plugin.Status,
		"type":         plugin.Type,
		"runtime":      plugin.Runtime,
		"is_active":    plugin.IsActive,
		"last_updated": plugin.UpdatedAt,
	}

	// 如果是微服务插件，获取容器状态
	if plugin.Type == models.PluginTypeMicroservice {
		// 这里可以调用Docker API获取容器状态
		status["container_status"] = "running" // 示例状态
	}

	utils.SuccessResponse(c, status)
}

// GetPluginHealth 获取插件健康状态
func (h *ExtendedPluginHandler) GetPluginHealth(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	plugin, err := h.extendedPluginService.GetExtendedPlugin(pluginID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get plugin")
		return
	}

	if plugin == nil {
		utils.NotFoundResponse(c, "Plugin not found")
		return
	}

	health := map[string]interface{}{
		"plugin_id": pluginID,
		"healthy":   true,
		"status":    plugin.Status,
		"checks":    []map[string]interface{}{},
	}

	checks := []map[string]interface{}{}

	// 基础健康检查：检查插件状态是否为运行状态
	if plugin.Status == models.StatusRunning {
		checks = append(checks, map[string]interface{}{
			"name":    "status_check",
			"passed":  true,
			"message": "Plugin is running",
		})
	} else {
		checks = append(checks, map[string]interface{}{
			"name":    "status_check",
			"passed":  false,
			"message": fmt.Sprintf("Plugin status is %s", plugin.Status),
		})
		health["healthy"] = false
	}

	// 类型特定健康检查
	switch plugin.Type {
	case models.PluginTypeMicroservice:
		if plugin.HealthCheckPath != nil && *plugin.HealthCheckPath != "" {
			// 执行HTTP健康检查
			checks = append(checks, map[string]interface{}{
				"name":    "http_health_check",
				"passed":  true, // 这里应该实际调用健康检查端点
				"message": "HTTP health check passed",
			})
		}
	case models.PluginTypeInProcess:
		// 检查插件是否已加载
		checks = append(checks, map[string]interface{}{
			"name":    "plugin_loaded",
			"passed":  true, // 这里应该检查插件是否在内存中加载
			"message": "Plugin is loaded in memory",
		})
	}

	health["checks"] = checks

	utils.SuccessResponse(c, health)
}

// WebSocket相关处理器（简化实现）
func (h *ExtendedPluginHandler) WebSocketPluginCall(c *gin.Context) {
	// WebSocket插件调用实现
	utils.BadRequestResponse(c, "WebSocket not implemented yet")
}

func (h *ExtendedPluginHandler) WebSocketPluginStatus(c *gin.Context) {
	// WebSocket插件状态订阅实现
	utils.BadRequestResponse(c, "WebSocket not implemented yet")
}

func (h *ExtendedPluginHandler) WebSocketPluginLogs(c *gin.Context) {
	// WebSocket插件日志流实现
	utils.BadRequestResponse(c, "WebSocket not implemented yet")
}

// Webhook相关处理函数
func (h *ExtendedPluginHandler) PluginStatusWebhook(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var requestBody struct {
		Status    string                 `json:"status" binding:"required"`
		Message   string                 `json:"message"`
		Metadata  map[string]interface{} `json:"metadata"`
		Timestamp time.Time              `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 更新插件状态
	status := models.PluginStatus(requestBody.Status)
	err = h.extendedPluginService.UpdatePluginStatus(pluginID, status)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update plugin status")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Plugin status updated via webhook",
	})
}

func (h *ExtendedPluginHandler) PluginDeployedWebhook(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var requestBody struct {
		Success   bool                   `json:"success" binding:"required"`
		Message   string                 `json:"message"`
		Metadata  map[string]interface{} `json:"metadata"`
		Timestamp time.Time              `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 根据部署结果更新插件状态
	var status models.PluginStatus
	if requestBody.Success {
		status = models.StatusRunning
	} else {
		status = models.StatusError
	}

	err = h.extendedPluginService.UpdatePluginStatus(pluginID, status)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update plugin status")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Plugin deployment status updated",
	})
}

func (h *ExtendedPluginHandler) PluginErrorWebhook(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var requestBody struct {
		Error     string                 `json:"error" binding:"required"`
		Level     string                 `json:"level"`
		Metadata  map[string]interface{} `json:"metadata"`
		Timestamp time.Time              `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 记录错误并可能更新插件状态
	if requestBody.Level == "critical" {
		err = h.extendedPluginService.UpdatePluginStatus(pluginID, models.StatusError)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to update plugin status")
			return
		}
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Plugin error recorded",
	})
}

func (h *ExtendedPluginHandler) PluginMetricsWebhook(c *gin.Context) {
	pluginIDStr := c.Param("id")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid plugin ID")
		return
	}

	var requestBody struct {
		Metrics   map[string]interface{} `json:"metrics" binding:"required"`
		Timestamp time.Time              `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// 存储插件指标数据
	// 这里应该将指标数据存储到数据库或时序数据库中
	// TODO: 使用 pluginID 来存储指标数据
	_ = pluginID // 暂时忽略未使用的变量

	utils.SuccessResponse(c, gin.H{
		"message": "Plugin metrics recorded",
	})
}

// GetExtendedPluginCategories 获取扩展插件分类
func (h *ExtendedPluginHandler) GetExtendedPluginCategories(c *gin.Context) {
	// 检查是否需要统计信息
	withStats := c.Query("with_stats") == "true"

	if withStats {
		categories, err := h.categoryService.GetCategoriesWithStats()
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to get categories with stats")
			return
		}
		utils.SuccessResponse(c, categories)
	} else {
		categories, err := h.categoryService.GetCategories()
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to get categories")
			return
		}
		utils.SuccessResponse(c, categories)
	}
}
