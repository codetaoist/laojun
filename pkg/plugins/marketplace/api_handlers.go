package marketplace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// SetupAPIRoutes 设置API路由
func (m *Marketplace) SetupAPIRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// 插件相关路由
	api.HandleFunc("/plugins", m.handleListPlugins).Methods("GET")
	api.HandleFunc("/plugins", m.handlePublishPlugin).Methods("POST")
	api.HandleFunc("/plugins/{id}", m.handleGetPlugin).Methods("GET")
	api.HandleFunc("/plugins/{id}", m.handleUpdatePlugin).Methods("PUT")
	api.HandleFunc("/plugins/{id}", m.handleDeletePlugin).Methods("DELETE")
	api.HandleFunc("/plugins/{id}/versions", m.handleListVersions).Methods("GET")
	api.HandleFunc("/plugins/{id}/versions/{version}", m.handleGetVersion).Methods("GET")
	api.HandleFunc("/plugins/{id}/download", m.handleDownloadPlugin).Methods("GET")
	api.HandleFunc("/plugins/{id}/download/{version}", m.handleDownloadPluginVersion).Methods("GET")

	// 搜索和分类路由
	api.HandleFunc("/search", m.handleSearchPlugins).Methods("GET")
	api.HandleFunc("/categories", m.handleListCategories).Methods("GET")
	api.HandleFunc("/categories/{category}/plugins", m.handleGetPluginsByCategory).Methods("GET")
	api.HandleFunc("/tags", m.handleListTags).Methods("GET")
	api.HandleFunc("/tags/{tag}/plugins", m.handleGetPluginsByTag).Methods("GET")

	// 评论和评分路由
	api.HandleFunc("/plugins/{id}/reviews", m.handleListReviews).Methods("GET")
	api.HandleFunc("/plugins/{id}/reviews", m.handleAddReview).Methods("POST")
	api.HandleFunc("/plugins/{id}/reviews/{reviewId}", m.handleUpdateReview).Methods("PUT")
	api.HandleFunc("/plugins/{id}/reviews/{reviewId}", m.handleDeleteReview).Methods("DELETE")

	// 统计信息
	api.HandleFunc("/stats", m.handleGetStats).Methods("GET")
	api.HandleFunc("/plugins/{id}/stats", m.handleGetPluginStats).Methods("GET")

	// 下载管理
	api.HandleFunc("/downloads", m.handleListDownloads).Methods("GET")
	api.HandleFunc("/downloads/{taskId}", m.handleGetDownloadProgress).Methods("GET")
	api.HandleFunc("/downloads/{taskId}/cancel", m.handleCancelDownload).Methods("POST")

	// 验证
	api.HandleFunc("/validate", m.handleValidatePlugin).Methods("POST")

	// 管理员路由
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(m.adminAuthMiddleware)
	admin.HandleFunc("/plugins/{id}/approve", m.handleApprovePlugin).Methods("POST")
	admin.HandleFunc("/plugins/{id}/reject", m.handleRejectPlugin).Methods("POST")
	admin.HandleFunc("/plugins/{id}/featured", m.handleSetFeatured).Methods("POST")
	admin.HandleFunc("/users", m.handleListUsers).Methods("GET")
	admin.HandleFunc("/users/{userId}/ban", m.handleBanUser).Methods("POST")
}

// 插件相关路由处理函数
// handleListPlugins 列出插件
func (m *Marketplace) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	category := r.URL.Query().Get("category")
	author := r.URL.Query().Get("author")
	status := r.URL.Query().Get("status")

	plugins := m.registry.ListPlugins()

	// 过滤
	var filtered []*PluginEntry
	for _, plugin := range plugins {
		if category != "" && plugin.Category != category {
			continue
		}
		if author != "" && plugin.Author != author {
			continue
		}
		if status != "" && plugin.Status != status {
			continue
		}
		filtered = append(filtered, plugin)
	}

	// 分页
	total := len(filtered)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		filtered = []*PluginEntry{}
	} else if end > total {
		filtered = filtered[start:]
	} else {
		filtered = filtered[start:end]
	}

	response := PaginatedResponse{
		Items:      filtered,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	m.sendJSONResponse(w, http.StatusOK, response)
}

// handlePublishPlugin 发布插件
func (m *Marketplace) handlePublishPlugin(w http.ResponseWriter, r *http.Request) {
	// 解析multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// 获取插件文件
	file, header, err := r.FormFile("plugin")
	if err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Plugin file is required")
		return
	}
	defer file.Close()

	// 获取清单数据
	manifestData := r.FormValue("manifest")
	if manifestData == "" {
		m.sendErrorResponse(w, http.StatusBadRequest, "Manifest data is required")
		return
	}

	var manifest PluginEntry
	if err := json.Unmarshal([]byte(manifestData), &manifest); err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Invalid manifest format")
		return
	}

	// 保存临时文件
	tempPath := filepath.Join(m.config.TempDir, header.Filename)
	tempFile, err := m.storage.Store(header.Filename, file)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, "Failed to save plugin file")
		return
	}

	// 验证插件
	validationResult, err := m.validator.ValidatePlugin(tempPath, &manifest)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, "Validation failed")
		return
	}

	if !validationResult.Valid {
		m.sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Plugin validation failed",
			Data:    validationResult,
		})
		return
	}

	// 发布插件
	err = m.PublishPlugin(&manifest, tempFile)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Plugin published successfully",
		Data:    manifest,
	})
}

// handleGetPlugin 获取插件详情
func (m *Marketplace) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	m.sendJSONResponse(w, http.StatusOK, plugin)
}

// handleUpdatePlugin 更新插件
func (m *Marketplace) handleUpdatePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	// 更新字段
	if name, ok := updates["name"].(string); ok {
		plugin.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		plugin.Description = description
	}
	if tags, ok := updates["tags"].([]interface{}); ok {
		plugin.Tags = make([]string, len(tags))
		for i, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				plugin.Tags[i] = tagStr
			}
		}
	}

	plugin.UpdatedAt = time.Now()

	err := m.registry.RegisterPlugin(plugin)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, plugin)
}

// handleDeletePlugin 删除插件
func (m *Marketplace) handleDeletePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	err := m.UnpublishPlugin(pluginID)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Plugin deleted successfully",
	})
}

// handleDownloadPlugin 下载插件
func (m *Marketplace) handleDownloadPlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]
	version := r.URL.Query().Get("version")

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	// 如果没有指定版本，使用最新版本
	if version == "" {
		version = plugin.Version
	}

	// 查找版本信息
	var versionInfo *VersionInfo
	for _, v := range plugin.Versions {
		if v.Version == version {
			versionInfo = &v
			break
		}
	}

	if versionInfo == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Version not found")
		return
	}

	// 开始下载任务
	taskID, err := m.DownloadPlugin(pluginID, version)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 更新下载计数
	plugin.Downloads++
	plugin.UpdatedAt = time.Now()
	m.registry.RegisterPlugin(plugin)

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"task_id":      taskID,
			"download_url": versionInfo.DownloadURL,
		},
	})
}

// handleSearchPlugins 搜索插件
func (m *Marketplace) handleSearchPlugins(w http.ResponseWriter, r *http.Request) {
	query := SearchQuery{
		Query:    r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
		Type:     r.URL.Query().Get("type"),
		Runtime:  r.URL.Query().Get("runtime"),
		Author:   r.URL.Query().Get("author"),
		License:  r.URL.Query().Get("license"),
	}

	if minRating := r.URL.Query().Get("min_rating"); minRating != "" {
		if rating, err := strconv.ParseFloat(minRating, 64); err == nil {
			query.MinRating = rating
		}
	}

	if tags := r.URL.Query().Get("tags"); tags != "" {
		query.Tags = strings.Split(tags, ",")
	}

	results := m.SearchPlugins(query)

	// 分页
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	total := len(results)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		results = []*PluginEntry{}
	} else if end > total {
		results = results[start:]
	} else {
		results = results[start:end]
	}

	response := PaginatedResponse{
		Items:      results,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	m.sendJSONResponse(w, http.StatusOK, response)
}

// handleListCategories 列出分类
func (m *Marketplace) handleListCategories(w http.ResponseWriter, r *http.Request) {
	categories := m.registry.GetCategories()
	m.sendJSONResponse(w, http.StatusOK, categories)
}

// handleGetPluginsByCategory 按分类获取插件
func (m *Marketplace) handleGetPluginsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]

	plugins := m.registry.GetPluginsByCategory(category)
	m.sendJSONResponse(w, http.StatusOK, plugins)
}

// handleListTags 列出标签
func (m *Marketplace) handleListTags(w http.ResponseWriter, r *http.Request) {
	tags := m.registry.GetTags()
	m.sendJSONResponse(w, http.StatusOK, tags)
}

// handleGetPluginsByTag 按标签获取插件
func (m *Marketplace) handleGetPluginsByTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tag := vars["tag"]

	plugins := m.registry.GetPluginsByTag(tag)
	m.sendJSONResponse(w, http.StatusOK, plugins)
}

// handleAddReview 添加评论
func (m *Marketplace) handleAddReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	var review Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	review.ID = fmt.Sprintf("review_%d", time.Now().Unix())
	review.CreatedAt = time.Now()

	err := m.AddReview(pluginID, review)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusCreated, review)
}

// handleGetStats 获取统计信息
func (m *Marketplace) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := m.GetStats()
	m.sendJSONResponse(w, http.StatusOK, stats)
}

// handleValidatePlugin 验证插件
func (m *Marketplace) handleValidatePlugin(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("plugin")
	if err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Plugin file is required")
		return
	}
	defer file.Close()

	// 保存临时文件
	tempPath := filepath.Join(m.config.TempDir, fmt.Sprintf("validate_%d.zip", time.Now().Unix()))
	tempFile, err := m.storage.Store(filepath.Base(tempPath), file)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer tempFile.Close()

	// 验证插件
	result, err := m.validator.ValidatePlugin(tempPath, nil)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, result)
}

// 管理员处理器

// handleApprovePlugin 批准插件
func (m *Marketplace) handleApprovePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	err := m.registry.UpdatePluginStatus(pluginID, "approved")
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Plugin approved successfully",
	})
}

// handleRejectPlugin 拒绝插件
func (m *Marketplace) handleRejectPlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	var data struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&data)

	err := m.registry.UpdatePluginStatus(pluginID, "rejected")
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Plugin rejected",
		Data:    map[string]string{"reason": data.Reason},
	})
}

// 工具方法

// sendJSONResponse 发送JSON响应
func (m *Marketplace) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// sendErrorResponse 发送错误响应
func (m *Marketplace) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	m.sendJSONResponse(w, statusCode, response)
}

// adminAuthMiddleware 管理员认证中间件
func (m *Marketplace) adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 简单的认证检查
		token := r.Header.Get("Authorization")
		if token == "" {
			m.sendErrorResponse(w, http.StatusUnauthorized, "Authorization required")
			return
		}

		// 这里应该验证token的有效性		// 简化实现，实际应用中需要更复杂的认证逻辑
		if !strings.HasPrefix(token, "Bearer ") {
			m.sendErrorResponse(w, http.StatusUnauthorized, "Invalid token format")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware CORS中间件
func (m *Marketplace) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 日志中间件
func (m *Marketplace) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter以捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		fmt.Printf("[%s] %s %s %d %v\n",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter 包装ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// 其他处理方法
// handleListVersions 列出插件版本
func (m *Marketplace) handleListVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	m.sendJSONResponse(w, http.StatusOK, plugin.Versions)
}

// handleGetVersion 获取特定版本信息
func (m *Marketplace) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]
	version := vars["version"]

	versionInfo, err := m.registry.GetPluginVersion(pluginID, version)
	if err != nil {
		m.sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, versionInfo)
}

// handleDownloadPluginVersion 下载特定版本
func (m *Marketplace) handleDownloadPluginVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]
	version := vars["version"]

	taskID, err := m.DownloadPlugin(pluginID, version)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"task_id": taskID},
	})
}

// handleListDownloads 列出下载任务
func (m *Marketplace) handleListDownloads(w http.ResponseWriter, r *http.Request) {
	downloads := m.downloader.ListDownloads()
	m.sendJSONResponse(w, http.StatusOK, downloads)
}

// handleGetDownloadProgress 获取下载进度
func (m *Marketplace) handleGetDownloadProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	progress, err := m.downloader.GetDownloadProgress(taskID)
	if err != nil {
		m.sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, progress)
}

// handleCancelDownload 取消下载
func (m *Marketplace) handleCancelDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	err := m.downloader.CancelDownload(taskID)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Download cancelled",
	})
}

// handleListReviews 列出评论
func (m *Marketplace) handleListReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	m.sendJSONResponse(w, http.StatusOK, plugin.Reviews)
}

// handleUpdateReview 更新评论
func (m *Marketplace) handleUpdateReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]
	reviewID := vars["reviewId"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	// 查找并更新评论
	for i, review := range plugin.Reviews {
		if review.ID == reviewID {
			if comment, ok := updates["comment"].(string); ok {
				plugin.Reviews[i].Comment = comment
			}
			if rating, ok := updates["rating"].(float64); ok {
				plugin.Reviews[i].Rating = rating
			}
			plugin.Reviews[i].UpdatedAt = time.Now()

			// 重新计算平均评分
			m.updatePluginRating(plugin)

			m.sendJSONResponse(w, http.StatusOK, plugin.Reviews[i])
			return
		}
	}

	m.sendErrorResponse(w, http.StatusNotFound, "Review not found")
}

// handleDeleteReview 删除评论
func (m *Marketplace) handleDeleteReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]
	reviewID := vars["reviewId"]

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	// 查找并删除评论
	for i, review := range plugin.Reviews {
		if review.ID == reviewID {
			plugin.Reviews = append(plugin.Reviews[:i], plugin.Reviews[i+1:]...)

			// 重新计算平均评分
			m.updatePluginRating(plugin)

			m.sendJSONResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Message: "Review deleted",
			})
			return
		}
	}

	m.sendErrorResponse(w, http.StatusNotFound, "Review not found")
}

// handleGetPluginStats 获取插件统计
func (m *Marketplace) handleGetPluginStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	stats := map[string]interface{}{
		"downloads":     plugin.Downloads,
		"rating":        plugin.Rating,
		"review_count":  len(plugin.Reviews),
		"version_count": len(plugin.Versions),
		"created_at":    plugin.CreatedAt,
		"updated_at":    plugin.UpdatedAt,
	}

	m.sendJSONResponse(w, http.StatusOK, stats)
}

// handleSetFeatured 设置推荐插件
func (m *Marketplace) handleSetFeatured(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	var data struct {
		Featured bool `json:"featured"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		m.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	plugin := m.registry.GetPlugin(pluginID)
	if plugin == nil {
		m.sendErrorResponse(w, http.StatusNotFound, "Plugin not found")
		return
	}

	plugin.Featured = data.Featured
	plugin.UpdatedAt = time.Now()

	err := m.registry.RegisterPlugin(plugin)
	if err != nil {
		m.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Featured status updated",
	})
}

// handleListUsers 列出用户（管理员功能）
func (m *Marketplace) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// 这里应该从用户管理系统获取用户列表
	// 简化实现，实际应调用用户管理系统的 API
	users := []map[string]interface{}{
		{
			"id":       "user1",
			"username": "developer1",
			"email":    "dev1@example.com",
			"status":   "active",
		},
	}

	m.sendJSONResponse(w, http.StatusOK, users)
}

// handleBanUser 封禁用户
func (m *Marketplace) handleBanUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var data struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&data)

	// 这里应该实现用户封禁逻辑
	// 简化实现，实际应调用用户管理系统的 API
	_ = userID
	_ = data.Reason

	m.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "User banned successfully",
	})
}

// updatePluginRating 更新插件评分
func (m *Marketplace) updatePluginRating(plugin *PluginEntry) {
	if len(plugin.Reviews) == 0 {
		plugin.Rating = 0
		return
	}

	total := 0.0
	for _, review := range plugin.Reviews {
		total += review.Rating
	}

	plugin.Rating = total / float64(len(plugin.Reviews))
}
