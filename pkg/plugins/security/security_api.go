package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// SecurityAPI 安全管理API
type SecurityAPI struct {
	securityManager *SecurityManager
	auditLogger     *AuditLogger
	router          *mux.Router
}

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	APIResponse
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// NewSecurityAPI 创建新的安全API
func NewSecurityAPI(securityManager *SecurityManager, auditLogger *AuditLogger) *SecurityAPI {
	api := &SecurityAPI{
		securityManager: securityManager,
		auditLogger:     auditLogger,
		router:          mux.NewRouter(),
	}

	api.setupRoutes()
	return api
}

// GetRouter 获取路由器
func (api *SecurityAPI) GetRouter() *mux.Router {
	return api.router
}

// setupRoutes 设置路由
func (api *SecurityAPI) setupRoutes() {
	// 中间件
	api.router.Use(api.corsMiddleware)
	api.router.Use(api.loggingMiddleware)

	// 认证相关
	auth := api.router.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", api.handleLogin).Methods("POST")
	auth.HandleFunc("/logout", api.handleLogout).Methods("POST")
	auth.HandleFunc("/refresh", api.handleRefreshToken).Methods("POST")
	auth.HandleFunc("/validate", api.handleValidateToken).Methods("POST")

	// 用户管理（需要认证）
	users := api.router.PathPrefix("/users").Subrouter()
	users.Use(api.authMiddleware)
	users.HandleFunc("", api.handleCreateUser).Methods("POST")
	users.HandleFunc("", api.handleListUsers).Methods("GET")
	users.HandleFunc("/{id}", api.handleGetUser).Methods("GET")
	users.HandleFunc("/{id}", api.handleUpdateUser).Methods("PUT")
	users.HandleFunc("/{id}", api.handleDeleteUser).Methods("DELETE")
	users.HandleFunc("/{id}/roles", api.handleAssignRole).Methods("POST")
	users.HandleFunc("/{id}/roles/{roleId}", api.handleRevokeRole).Methods("DELETE")
	users.HandleFunc("/{id}/permissions", api.handleGrantPermission).Methods("POST")
	users.HandleFunc("/{id}/permissions/{permissionId}", api.handleRevokePermission).Methods("DELETE")

	// 角色管理
	roles := api.router.PathPrefix("/roles").Subrouter()
	roles.Use(api.authMiddleware)
	roles.HandleFunc("", api.handleCreateRole).Methods("POST")
	roles.HandleFunc("", api.handleListRoles).Methods("GET")
	roles.HandleFunc("/{id}", api.handleGetRole).Methods("GET")
	roles.HandleFunc("/{id}", api.handleUpdateRole).Methods("PUT")
	roles.HandleFunc("/{id}", api.handleDeleteRole).Methods("DELETE")

	// 权限管理
	permissions := api.router.PathPrefix("/permissions").Subrouter()
	permissions.Use(api.authMiddleware)
	permissions.HandleFunc("", api.handleCreatePermission).Methods("POST")
	permissions.HandleFunc("", api.handleListPermissions).Methods("GET")
	permissions.HandleFunc("/{id}", api.handleGetPermission).Methods("GET")
	permissions.HandleFunc("/{id}", api.handleUpdatePermission).Methods("PUT")
	permissions.HandleFunc("/{id}", api.handleDeletePermission).Methods("DELETE")
	permissions.HandleFunc("/check", api.handleCheckPermission).Methods("POST")

	// 令牌管理
	tokens := api.router.PathPrefix("/tokens").Subrouter()
	tokens.Use(api.authMiddleware)
	tokens.HandleFunc("", api.handleListTokens).Methods("GET")
	tokens.HandleFunc("/{id}", api.handleRevokeToken).Methods("DELETE")

	// 资源限制管理
	resources := api.router.PathPrefix("/resources").Subrouter()
	resources.Use(api.authMiddleware)
	resources.HandleFunc("/limits", api.handleSetResourceLimit).Methods("POST")
	resources.HandleFunc("/limits/{pluginId}", api.handleGetResourceLimit).Methods("GET")
	resources.HandleFunc("/usage/{pluginId}", api.handleGetResourceUsage).Methods("GET")
	resources.HandleFunc("/usage", api.handleGetAllResourceUsage).Methods("GET")
	resources.HandleFunc("/alerts", api.handleGetResourceAlerts).Methods("GET")
	resources.HandleFunc("/thresholds", api.handleSetResourceThreshold).Methods("POST")
	resources.HandleFunc("/thresholds", api.handleGetResourceThresholds).Methods("GET")

	// 审计日志
	audit := api.router.PathPrefix("/audit").Subrouter()
	audit.Use(api.authMiddleware)
	audit.HandleFunc("/logs", api.handleQueryAuditLogs).Methods("GET")
	audit.HandleFunc("/logs", api.handleQueryAuditLogsPost).Methods("POST")
	audit.HandleFunc("/metrics", api.handleGetAuditMetrics).Methods("GET")
	audit.HandleFunc("/export", api.handleExportAuditLogs).Methods("GET")

	// 系统统计
	stats := api.router.PathPrefix("/stats").Subrouter()
	stats.Use(api.authMiddleware)
	stats.HandleFunc("/security", api.handleGetSecurityStats).Methods("GET")
	stats.HandleFunc("/system", api.handleGetSystemStats).Methods("GET")
}

// 认证相关处理函数
// handleLogin 处理登录
func (api *SecurityAPI) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 这里应该验证用户名和密码
	// 为了演示，我们假设验证成功
	userID := "user_" + req.Username

	// 创建访问令牌
	token, err := api.securityManager.CreateToken(userID, "access", []string{"read", "write"})
	if err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	// 记录审计日志
	api.auditLogger.LogInfo("authentication", "login", "user", map[string]interface{}{
		"user_id":    userID,
		"ip_address": api.getClientIP(r),
	})

	api.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"token":      token.Token,
		"expires_at": token.ExpiresAt,
	})
}

// handleLogout 处理登出
func (api *SecurityAPI) handleLogout(w http.ResponseWriter, r *http.Request) {
	tokenString := api.extractToken(r)
	if tokenString == "" {
		api.sendErrorResponse(w, http.StatusBadRequest, "No token provided")
		return
	}

	// 验证并撤销令牌
	context, err := api.securityManager.ValidateToken(tokenString)
	if err != nil {
		api.sendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	if err := api.securityManager.RevokeToken(context.TokenID); err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}

	// 记录审计日志
	api.auditLogger.LogInfo("authentication", "logout", "user", map[string]interface{}{
		"user_id":    context.UserID,
		"ip_address": api.getClientIP(r),
	})

	api.sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// handleRefreshToken 处理令牌刷新
func (api *SecurityAPI) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 验证刷新令牌
	context, err := api.securityManager.ValidateToken(req.RefreshToken)
	if err != nil {
		api.sendErrorResponse(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// 创建新的访问令牌
	token, err := api.securityManager.CreateToken(context.UserID, "access", []string{"read", "write"})
	if err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	api.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"token":      token.Token,
		"expires_at": token.ExpiresAt,
	})
}

// handleValidateToken 处理令牌验证
func (api *SecurityAPI) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	tokenString := api.extractToken(r)
	if tokenString == "" {
		api.sendErrorResponse(w, http.StatusBadRequest, "No token provided")
		return
	}

	context, err := api.securityManager.ValidateToken(tokenString)
	if err != nil {
		api.sendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	api.sendJSONResponse(w, http.StatusOK, context)
}

// 用户管理处理函数
// handleCreateUser 处理创建用户
func (api *SecurityAPI) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.securityManager.CreateUser(&user); err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.sendJSONResponse(w, http.StatusCreated, user)
}

// handleListUsers 处理列出用户
func (api *SecurityAPI) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// 这里应该实现用户列表功能
	// 为了演示，返回空列表
	api.sendJSONResponse(w, http.StatusOK, []User{})
}

// handleGetUser 处理获取用户
func (api *SecurityAPI) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := api.securityManager.GetUser(userID)
	if err != nil {
		api.sendErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	api.sendJSONResponse(w, http.StatusOK, user)
}

// handleUpdateUser 处理更新用户
func (api *SecurityAPI) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user.ID = userID
	if err := api.securityManager.UpdateUser(&user); err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.sendJSONResponse(w, http.StatusOK, user)
}

// handleDeleteUser 处理删除用户
func (api *SecurityAPI) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if err := api.securityManager.DeleteUser(userID); err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

// handleAssignRole 处理分配角色
func (api *SecurityAPI) handleAssignRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req struct {
		RoleID string `json:"role_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.securityManager.AssignRole(userID, req.RoleID); err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Role assigned successfully",
	})
}

// handleRevokeRole 处理撤销角色
func (api *SecurityAPI) handleRevokeRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	roleID := vars["roleId"]

	if err := api.securityManager.RevokeRole(userID, roleID); err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Role revoked successfully",
	})
}

// 权限管理处理函数
// handleCheckPermission 处理权限检查
func (api *SecurityAPI) handleCheckPermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	hasPermission := api.securityManager.CheckPermission(req.UserID, req.Resource, req.Action)

	api.sendJSONResponse(w, http.StatusOK, map[string]bool{
		"has_permission": hasPermission,
	})
}

// 资源管理处理函数
// handleSetResourceLimit 处理设置资源限制
func (api *SecurityAPI) handleSetResourceLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PluginID string         `json:"plugin_id"`
		Limit    *ResourceLimit `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.securityManager.SetResourceLimit(req.PluginID, req.Limit); err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Resource limit set successfully",
	})
}

// handleGetResourceLimit 处理获取资源限制
func (api *SecurityAPI) handleGetResourceLimit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["pluginId"]

	limit, err := api.securityManager.GetResourceLimit(pluginID)
	if err != nil {
		api.sendErrorResponse(w, http.StatusNotFound, "Resource limit not found")
		return
	}

	api.sendJSONResponse(w, http.StatusOK, limit)
}

// handleGetResourceUsage 处理获取资源使用情况
func (api *SecurityAPI) handleGetResourceUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["pluginId"]

	usage, err := api.securityManager.CheckResourceUsage(pluginID)
	if err != nil {
		api.sendErrorResponse(w, http.StatusNotFound, "Resource usage not found")
		return
	}

	api.sendJSONResponse(w, http.StatusOK, usage)
}

// handleGetAllResourceUsage 处理获取所有资源使用情况
func (api *SecurityAPI) handleGetAllResourceUsage(w http.ResponseWriter, r *http.Request) {
	usage := api.securityManager.resourceMonitor.GetAllResourceUsage()
	api.sendJSONResponse(w, http.StatusOK, usage)
}

// handleGetResourceAlerts 处理获取资源告警
func (api *SecurityAPI) handleGetResourceAlerts(w http.ResponseWriter, r *http.Request) {
	pluginID := r.URL.Query().Get("plugin_id")
	resolvedStr := r.URL.Query().Get("resolved")

	resolved := false
	if resolvedStr == "true" {
		resolved = true
	}

	alerts := api.securityManager.resourceMonitor.GetAlerts(pluginID, resolved)
	api.sendJSONResponse(w, http.StatusOK, alerts)
}

// handleSetResourceThreshold 处理设置资源阈值
func (api *SecurityAPI) handleSetResourceThreshold(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type     string  `json:"type"`
		Warning  float64 `json:"warning"`
		Critical float64 `json:"critical"`
		Enabled  bool    `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	api.securityManager.resourceMonitor.SetThreshold(req.Type, req.Warning, req.Critical, req.Enabled)

	api.sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Resource threshold set successfully",
	})
}

// handleGetResourceThresholds 处理获取资源阈值
func (api *SecurityAPI) handleGetResourceThresholds(w http.ResponseWriter, r *http.Request) {
	thresholds := api.securityManager.resourceMonitor.GetThresholds()
	api.sendJSONResponse(w, http.StatusOK, thresholds)
}

// 审计日志处理函数
// handleQueryAuditLogs 处理查询审计日志
func (api *SecurityAPI) handleQueryAuditLogs(w http.ResponseWriter, r *http.Request) {
	query := &AuditQuery{}

	// 解析查询参数
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query.EndTime = &endTime
		}
	}

	if levelStr := r.URL.Query().Get("level"); levelStr != "" {
		if level, err := strconv.Atoi(levelStr); err == nil {
			auditLevel := AuditLogLevel(level)
			query.Level = &auditLevel
		}
	}

	query.Category = r.URL.Query().Get("category")
	query.Action = r.URL.Query().Get("action")
	query.Resource = r.URL.Query().Get("resource")
	query.UserID = r.URL.Query().Get("user_id")
	query.PluginID = r.URL.Query().Get("plugin_id")
	query.Result = r.URL.Query().Get("result")

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			query.Offset = offset
		}
	}

	events, total, err := api.auditLogger.Query(query)
	if err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := PaginatedResponse{
		APIResponse: APIResponse{
			Success: true,
			Data:    events,
		},
		Total:  total,
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleQueryAuditLogsPost 处理POST查询审计日志
func (api *SecurityAPI) handleQueryAuditLogsPost(w http.ResponseWriter, r *http.Request) {
	var query AuditQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	events, total, err := api.auditLogger.Query(&query)
	if err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := PaginatedResponse{
		APIResponse: APIResponse{
			Success: true,
			Data:    events,
		},
		Total:  total,
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetAuditMetrics 处理获取审计指标
func (api *SecurityAPI) handleGetAuditMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := api.auditLogger.GetMetrics()
	api.sendJSONResponse(w, http.StatusOK, metrics)
}

// handleExportAuditLogs 处理导出审计日志
func (api *SecurityAPI) handleExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	format := r.URL.Query().Get("format")

	if format == "" {
		format = "json"
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid start_time format")
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		api.sendErrorResponse(w, http.StatusBadRequest, "Invalid end_time format")
		return
	}

	data, err := api.auditLogger.ExportLogs(startTime, endTime, format)
	if err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 设置响应头
	filename := fmt.Sprintf("audit_logs_%s_%s.%s",
		startTime.Format("20060102"),
		endTime.Format("20060102"),
		format)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	w.Write(data)
}

// 统计处理函数
// handleGetSecurityStats 处理获取安全统计
func (api *SecurityAPI) handleGetSecurityStats(w http.ResponseWriter, r *http.Request) {
	stats := api.securityManager.GetStats()
	api.sendJSONResponse(w, http.StatusOK, stats)
}

// handleGetSystemStats 处理获取系统统计
func (api *SecurityAPI) handleGetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats, err := api.securityManager.resourceMonitor.GetSystemStats()
	if err != nil {
		api.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.sendJSONResponse(w, http.StatusOK, stats)
}

// 中间件处理函数
// corsMiddleware CORS中间件
func (api *SecurityAPI) corsMiddleware(next http.Handler) http.Handler {
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
func (api *SecurityAPI) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter以捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// 记录API访问日志
		api.auditLogger.LogInfo("api", r.Method, r.URL.Path, map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapped.statusCode,
			"duration_ms": duration.Milliseconds(),
			"ip_address":  api.getClientIP(r),
			"user_agent":  r.UserAgent(),
		})
	})
}

// authMiddleware 认证中间件
func (api *SecurityAPI) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := api.extractToken(r)
		if tokenString == "" {
			api.sendErrorResponse(w, http.StatusUnauthorized, "No token provided")
			return
		}

		context, err := api.securityManager.ValidateToken(tokenString)
		if err != nil {
			api.sendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// 将安全上下文添加到请求上下文
		r = r.WithContext(r.Context())
		r.Header.Set("X-User-ID", context.UserID)
		r.Header.Set("X-Username", context.Username)

		next.ServeHTTP(w, r)
	})
}

// 辅助方法

// sendJSONResponse 发送JSON响应
func (api *SecurityAPI) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse 发送错误响应
func (api *SecurityAPI) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// extractToken 提取令牌
func (api *SecurityAPI) extractToken(r *http.Request) string {
	// 从Authorization头提取令牌
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 从查询参数提取令牌
	return r.URL.Query().Get("token")
}

// getClientIP 获取客户端IP
func (api *SecurityAPI) getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头，获取第一个IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	parts := strings.Split(r.RemoteAddr, ":")
	return parts[0]
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

// 其他未实现的处理器方法（为了完整性）

func (api *SecurityAPI) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	// 实现创建角色逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleListRoles(w http.ResponseWriter, r *http.Request) {
	// 实现列出角色逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleGetRole(w http.ResponseWriter, r *http.Request) {
	// 实现获取角色逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	// 实现更新角色逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	// 实现删除角色逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleCreatePermission(w http.ResponseWriter, r *http.Request) {
	// 实现创建权限逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleListPermissions(w http.ResponseWriter, r *http.Request) {
	// 实现列出权限逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleGetPermission(w http.ResponseWriter, r *http.Request) {
	// 实现获取权限逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleUpdatePermission(w http.ResponseWriter, r *http.Request) {
	// 实现更新权限逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleDeletePermission(w http.ResponseWriter, r *http.Request) {
	// 实现删除权限逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleGrantPermission(w http.ResponseWriter, r *http.Request) {
	// 实现授予权限逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleRevokePermission(w http.ResponseWriter, r *http.Request) {
	// 实现撤销权限逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}

func (api *SecurityAPI) handleListTokens(w http.ResponseWriter, r *http.Request) {
	// 实现列出令牌逻辑
	api.sendErrorResponse(w, http.StatusNotImplemented, "Not implemented")
}
