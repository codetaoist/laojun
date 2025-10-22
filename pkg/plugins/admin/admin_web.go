package admin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// AdminWeb Admin Web管理界面
type AdminWeb struct {
	config          *AdminWebConfig
	pluginManager   PluginManagerInterface
	securityManager SecurityManagerInterface
	marketPlace     MarketplaceInterface
	router          *mux.Router
	templates       *template.Template
	upgrader        websocket.Upgrader
	wsClients       map[*websocket.Conn]bool
	wsHub           chan []byte
}

// AdminWebConfig Admin Web配置
type AdminWebConfig struct {
	Port           int      `json:"port"`
	Host           string   `json:"host"`
	StaticPath     string   `json:"static_path"`
	TemplatePath   string   `json:"template_path"`
	EnableAuth     bool     `json:"enable_auth"`
	SessionTimeout int      `json:"session_timeout"`
	MaxFileSize    int64    `json:"max_file_size"`
	AllowedOrigins []string `json:"allowed_origins"`
	EnableHTTPS    bool     `json:"enable_https"`
	CertFile       string   `json:"cert_file"`
	KeyFile        string   `json:"key_file"`
}

// PluginManagerInterface 插件管理器接口
type PluginManagerInterface interface {
	ListPlugins() ([]PluginInfo, error)
	GetPlugin(id string) (*PluginInfo, error)
	InstallPlugin(path string) error
	UninstallPlugin(id string) error
	EnablePlugin(id string) error
	DisablePlugin(id string) error
	GetPluginConfig(id string) (map[string]interface{}, error)
	SetPluginConfig(id string, config map[string]interface{}) error
	GetPluginLogs(id string, limit int) ([]LogEntry, error)
	GetPluginMetrics(id string) (*PluginMetrics, error)
}

// SecurityManagerInterface 安全管理器接口
type SecurityManagerInterface interface {
	ValidateToken(token string) (*SecurityContext, error)
	CheckPermission(userID, resource, action string) bool
	GetResourceUsage(pluginID string) (*ResourceUsage, error)
	GetSecurityStats() *SecurityStats
}

// MarketplaceInterface 插件市场接口
type MarketplaceInterface interface {
	SearchPlugins(query string) ([]MarketplacePlugin, error)
	GetPlugin(id string) (*MarketplacePlugin, error)
	DownloadPlugin(id, version string) (string, error)
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Status       string                 `json:"status"`
	Enabled      bool                   `json:"enabled"`
	Config       map[string]interface{} `json:"config"`
	InstallTime  time.Time              `json:"install_time"`
	UpdateTime   time.Time              `json:"update_time"`
	Dependencies []string               `json:"dependencies"`
	Permissions  []string               `json:"permissions"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// PluginMetrics 插件指标
type PluginMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	DiskUsage   int64   `json:"disk_usage"`
	NetworkIn   int64   `json:"network_in"`
	NetworkOut  int64   `json:"network_out"`
	Requests    int64   `json:"requests"`
	Errors      int64   `json:"errors"`
	Uptime      int64   `json:"uptime"`
}

// SecurityContext 安全上下文
type SecurityContext struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenID     string   `json:"token_id"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	PluginID    string    `json:"plugin_id"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemoryBytes int64     `json:"memory_bytes"`
	DiskBytes   int64     `json:"disk_bytes"`
	NetworkIn   int64     `json:"network_in"`
	NetworkOut  int64     `json:"network_out"`
	Timestamp   time.Time `json:"timestamp"`
}

// SecurityStats 安全统计
type SecurityStats struct {
	TotalUsers         int `json:"total_users"`
	ActiveSessions     int `json:"active_sessions"`
	FailedLogins       int `json:"failed_logins"`
	SecurityAlerts     int `json:"security_alerts"`
	ResourceViolations int `json:"resource_violations"`
}

// MarketplacePlugin 市场插件
type MarketplacePlugin struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Rating      float64   `json:"rating"`
	Downloads   int       `json:"downloads"`
	UpdateTime  time.Time `json:"update_time"`
	Size        int64     `json:"size"`
	Screenshots []string  `json:"screenshots"`
}

// DashboardData 仪表板数据
type DashboardData struct {
	TotalPlugins    int            `json:"total_plugins"`
	EnabledPlugins  int            `json:"enabled_plugins"`
	DisabledPlugins int            `json:"disabled_plugins"`
	SystemMetrics   *SystemMetrics `json:"system_metrics"`
	SecurityStats   *SecurityStats `json:"security_stats"`
	RecentLogs      []LogEntry     `json:"recent_logs"`
	TopPlugins      []PluginInfo   `json:"top_plugins"`
	Alerts          []Alert        `json:"alerts"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIn   int64   `json:"network_in"`
	NetworkOut  int64   `json:"network_out"`
	Uptime      int64   `json:"uptime"`
}

// Alert 告警
type Alert struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Level    string    `json:"level"`
	Message  string    `json:"message"`
	Source   string    `json:"source"`
	Time     time.Time `json:"time"`
	Resolved bool      `json:"resolved"`
}

// NewAdminWeb 创建新的Admin Web实例
func NewAdminWeb(config *AdminWebConfig, pluginManager PluginManagerInterface, securityManager SecurityManagerInterface, marketplace MarketplaceInterface) *AdminWeb {
	admin := &AdminWeb{
		config:          config,
		pluginManager:   pluginManager,
		securityManager: securityManager,
		marketPlace:     marketplace,
		router:          mux.NewRouter(),
		wsClients:       make(map[*websocket.Conn]bool),
		wsHub:           make(chan []byte, 256),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该检查Origin
			},
		},
	}

	admin.loadTemplates()
	admin.setupRoutes()
	admin.startWebSocketHub()

	return admin
}

// Start 启动Admin Web服务
func (aw *AdminWeb) Start() error {
	addr := fmt.Sprintf("%s:%d", aw.config.Host, aw.config.Port)

	fmt.Printf("Admin Web interface starting on %s\n", addr)

	if aw.config.EnableHTTPS {
		return http.ListenAndServeTLS(addr, aw.config.CertFile, aw.config.KeyFile, aw.router)
	}

	return http.ListenAndServe(addr, aw.router)
}

// loadTemplates 加载模板
func (aw *AdminWeb) loadTemplates() {
	templatePath := aw.config.TemplatePath
	if templatePath == "" {
		templatePath = "./templates"
	}

	// 创建模板函数
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatBytes": func(bytes int64) string {
			const unit = 1024
			if bytes < unit {
				return fmt.Sprintf("%d B", bytes)
			}
			div, exp := int64(unit), 0
			for n := bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
		},
		"formatPercent": func(value float64) string {
			return fmt.Sprintf("%.1f%%", value)
		},
	}

	aw.templates = template.Must(template.New("").Funcs(funcMap).ParseGlob(filepath.Join(templatePath, "*.html")))
}

// setupRoutes 设置路由
func (aw *AdminWeb) setupRoutes() {
	// 中间件
	aw.router.Use(aw.corsMiddleware)
	aw.router.Use(aw.loggingMiddleware)

	// 静态文件服务
	staticPath := aw.config.StaticPath
	if staticPath == "" {
		staticPath = "./static"
	}
	aw.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// 认证路由
	aw.router.HandleFunc("/login", aw.handleLogin).Methods("GET", "POST")
	aw.router.HandleFunc("/logout", aw.handleLogout).Methods("POST")

	// 需要认证的路由
	protected := aw.router.PathPrefix("/").Subrouter()
	if aw.config.EnableAuth {
		protected.Use(aw.authMiddleware)
	}

	// 主页路由
	protected.HandleFunc("/", aw.handleDashboard).Methods("GET")
	protected.HandleFunc("/dashboard", aw.handleDashboard).Methods("GET")

	// 插件管理
	plugins := protected.PathPrefix("/plugins").Subrouter()
	plugins.HandleFunc("", aw.handlePluginList).Methods("GET")
	plugins.HandleFunc("/install", aw.handlePluginInstall).Methods("GET", "POST")
	plugins.HandleFunc("/upload", aw.handlePluginUpload).Methods("POST")
	plugins.HandleFunc("/{id}", aw.handlePluginDetail).Methods("GET")
	plugins.HandleFunc("/{id}/config", aw.handlePluginConfig).Methods("GET", "POST")
	plugins.HandleFunc("/{id}/logs", aw.handlePluginLogs).Methods("GET")
	plugins.HandleFunc("/{id}/metrics", aw.handlePluginMetrics).Methods("GET")
	plugins.HandleFunc("/{id}/enable", aw.handlePluginEnable).Methods("POST")
	plugins.HandleFunc("/{id}/disable", aw.handlePluginDisable).Methods("POST")
	plugins.HandleFunc("/{id}/uninstall", aw.handlePluginUninstall).Methods("POST")

	// 插件市场
	market := protected.PathPrefix("/market").Subrouter()
	market.HandleFunc("", aw.handleMarketplace).Methods("GET")
	market.HandleFunc("/search", aw.handleMarketSearch).Methods("GET")
	market.HandleFunc("/{id}", aw.handleMarketDetail).Methods("GET")
	market.HandleFunc("/{id}/install", aw.handleMarketInstall).Methods("POST")

	// 系统监控
	monitoring := protected.PathPrefix("/monitoring").Subrouter()
	monitoring.HandleFunc("", aw.handleMonitoring).Methods("GET")
	monitoring.HandleFunc("/resources", aw.handleResourceMonitoring).Methods("GET")
	monitoring.HandleFunc("/security", aw.handleSecurityMonitoring).Methods("GET")
	monitoring.HandleFunc("/logs", aw.handleSystemLogs).Methods("GET")

	// 系统设置
	settings := protected.PathPrefix("/settings").Subrouter()
	settings.HandleFunc("", aw.handleSettings).Methods("GET", "POST")
	settings.HandleFunc("/users", aw.handleUserManagement).Methods("GET")
	settings.HandleFunc("/security", aw.handleSecuritySettings).Methods("GET", "POST")

	// API路由
	api := protected.PathPrefix("/api").Subrouter()
	api.HandleFunc("/dashboard", aw.handleAPIDashboard).Methods("GET")
	api.HandleFunc("/plugins", aw.handleAPIPlugins).Methods("GET")
	api.HandleFunc("/plugins/{id}/metrics", aw.handleAPIPluginMetrics).Methods("GET")
	api.HandleFunc("/system/metrics", aw.handleAPISystemMetrics).Methods("GET")
	api.HandleFunc("/alerts", aw.handleAPIAlerts).Methods("GET")

	// WebSocket
	protected.HandleFunc("/ws", aw.handleWebSocket)
}

// 页面处理函数
// handleDashboard 处理仪表板页面
func (aw *AdminWeb) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := aw.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	aw.renderTemplate(w, "dashboard.html", data)
}

// handlePluginList 处理插件列表页面
func (aw *AdminWeb) handlePluginList(w http.ResponseWriter, r *http.Request) {
	plugins, err := aw.pluginManager.ListPlugins()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":   "插件管理",
		"Plugins": plugins,
	}

	aw.renderTemplate(w, "plugin_list.html", data)
}

// handlePluginDetail 处理插件详情页面
func (aw *AdminWeb) handlePluginDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	plugin, err := aw.pluginManager.GetPlugin(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}

	// 获取插件指标
	metrics, _ := aw.pluginManager.GetPluginMetrics(pluginID)

	// 获取资源使用情况
	resourceUsage, _ := aw.securityManager.GetResourceUsage(pluginID)

	data := map[string]interface{}{
		"Title":         "插件详情",
		"Plugin":        plugin,
		"Metrics":       metrics,
		"ResourceUsage": resourceUsage,
	}

	aw.renderTemplate(w, "plugin_detail.html", data)
}

// handlePluginConfig 处理插件配置页面
func (aw *AdminWeb) handlePluginConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	if r.Method == "POST" {
		// 处理配置更新
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		config := make(map[string]interface{})
		for key, values := range r.Form {
			if len(values) > 0 {
				config[key] = values[0]
			}
		}

		if err := aw.pluginManager.SetPluginConfig(pluginID, config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/plugins/%s", pluginID), http.StatusSeeOther)
		return
	}

	// 获取当前配置
	config, err := aw.pluginManager.GetPluginConfig(pluginID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	plugin, err := aw.pluginManager.GetPlugin(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title":  "插件配置",
		"Plugin": plugin,
		"Config": config,
	}

	aw.renderTemplate(w, "plugin_config.html", data)
}

// handlePluginInstall 处理插件安装页面
func (aw *AdminWeb) handlePluginInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		pluginPath := r.FormValue("plugin_path")
		if pluginPath == "" {
			http.Error(w, "Plugin path is required", http.StatusBadRequest)
			return
		}

		if err := aw.pluginManager.InstallPlugin(pluginPath); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/plugins", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Title": "安装插件",
	}

	aw.renderTemplate(w, "plugin_install.html", data)
}

// handlePluginUpload 处理插件上传
func (aw *AdminWeb) handlePluginUpload(w http.ResponseWriter, r *http.Request) {
	// 解析multipart表单
	if err := r.ParseMultipartForm(aw.config.MaxFileSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("plugin_file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存文件并安装插件
	pluginPath := filepath.Join(aw.config.PluginDir, header.Filename)
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := os.Create(pluginPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := aw.pluginManager.InstallPlugin(pluginPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 这里应该实现文件保存和插件安装逻辑
	_ = header // 使用文件头信息，如文件名等
	// 临时响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Plugin uploaded successfully",
	})
}

// handleMarketplace 处理插件市场页面
func (aw *AdminWeb) handleMarketplace(w http.ResponseWriter, r *http.Request) {
	// 获取热门插件
	plugins, err := aw.marketPlace.SearchPlugins("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":   "插件市场",
		"Plugins": plugins,
	}

	aw.renderTemplate(w, "marketplace.html", data)
}

// handleMarketSearch 处理市场搜索
func (aw *AdminWeb) handleMarketSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	plugins, err := aw.marketPlace.SearchPlugins(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugins)
}

// handleMonitoring 处理监控页面
func (aw *AdminWeb) handleMonitoring(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "系统监控",
	}

	aw.renderTemplate(w, "monitoring.html", data)
}

// handleSettings 处理设置页面
func (aw *AdminWeb) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 处理设置更新
		// 这里应该实现设置更新逻辑
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Title": "系统设置",
	}

	aw.renderTemplate(w, "settings.html", data)
}

// API处理函数
// handleAPIDashboard 处理仪表板API
func (aw *AdminWeb) handleAPIDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := aw.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleAPIPlugins 处理插件API
func (aw *AdminWeb) handleAPIPlugins(w http.ResponseWriter, r *http.Request) {
	plugins, err := aw.pluginManager.ListPlugins()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugins)
}

// handleAPIPluginMetrics 处理插件指标API
func (aw *AdminWeb) handleAPIPluginMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	metrics, err := aw.pluginManager.GetPluginMetrics(pluginID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// WebSocket处理函数
// handleWebSocket 处理WebSocket连接
func (aw *AdminWeb) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := aw.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	aw.wsClients[conn] = true

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(aw.wsClients, conn)
			break
		}
	}
}

// startWebSocketHub 启动WebSocket中心
func (aw *AdminWeb) startWebSocketHub() {
	go func() {
		for {
			select {
			case message := <-aw.wsHub:
				for client := range aw.wsClients {
					err := client.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						client.Close()
						delete(aw.wsClients, client)
					}
				}
			}
		}
	}()
}

// BroadcastMessage 广播消息
func (aw *AdminWeb) BroadcastMessage(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	select {
	case aw.wsHub <- data:
	default:
	}
}

// 辅助方法

// getDashboardData 获取仪表板数据
func (aw *AdminWeb) getDashboardData() (*DashboardData, error) {
	plugins, err := aw.pluginManager.ListPlugins()
	if err != nil {
		return nil, err
	}

	enabledCount := 0
	for _, plugin := range plugins {
		if plugin.Enabled {
			enabledCount++
		}
	}

	securityStats := aw.securityManager.GetSecurityStats()

	data := &DashboardData{
		TotalPlugins:    len(plugins),
		EnabledPlugins:  enabledCount,
		DisabledPlugins: len(plugins) - enabledCount,
		SecurityStats:   securityStats,
		TopPlugins:      plugins[:min(len(plugins), 5)],
		SystemMetrics: &SystemMetrics{
			CPUUsage:    45.2,
			MemoryUsage: 67.8,
			DiskUsage:   23.4,
			NetworkIn:   1024 * 1024,
			NetworkOut:  2048 * 1024,
			Uptime:      3600 * 24 * 7,
		},
		RecentLogs: []LogEntry{},
		Alerts:     []Alert{},
	}

	return data, nil
}

// renderTemplate 渲染模板
func (aw *AdminWeb) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	if err := aw.templates.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 中间件函数

// corsMiddleware CORS中间件
func (aw *AdminWeb) corsMiddleware(next http.Handler) http.Handler {
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
func (aw *AdminWeb) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		fmt.Printf("[%s] %s %s - %v\n",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			duration)
	})
}

// authMiddleware 认证中间件
func (aw *AdminWeb) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查会话或令牌
		// 这里应该实现实际的认证逻辑
		next.ServeHTTP(w, r)
	})
}

// 其他处理器方法的占位符实现
func (aw *AdminWeb) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 处理登录逻辑
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	aw.renderTemplate(w, "login.html", map[string]interface{}{"Title": "登录"})
}

func (aw *AdminWeb) handleLogout(w http.ResponseWriter, r *http.Request) {
	// 处理登出逻辑
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (aw *AdminWeb) handlePluginLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := aw.pluginManager.GetPluginLogs(pluginID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (aw *AdminWeb) handlePluginMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	metrics, err := aw.pluginManager.GetPluginMetrics(pluginID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (aw *AdminWeb) handlePluginEnable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	if err := aw.pluginManager.EnablePlugin(pluginID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (aw *AdminWeb) handlePluginDisable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	if err := aw.pluginManager.DisablePlugin(pluginID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (aw *AdminWeb) handlePluginUninstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	if err := aw.pluginManager.UninstallPlugin(pluginID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (aw *AdminWeb) handleMarketDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	plugin, err := aw.marketPlace.GetPlugin(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title":  "插件详情",
		"Plugin": plugin,
	}

	aw.renderTemplate(w, "market_detail.html", data)
}

func (aw *AdminWeb) handleMarketInstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["id"]

	version := r.FormValue("version")
	if version == "" {
		version = "latest"
	}

	// 下载并安装插件
	pluginPath, err := aw.marketPlace.DownloadPlugin(pluginID, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := aw.pluginManager.InstallPlugin(pluginPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (aw *AdminWeb) handleResourceMonitoring(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "资源监控",
	}
	aw.renderTemplate(w, "resource_monitoring.html", data)
}

func (aw *AdminWeb) handleSecurityMonitoring(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "安全监控",
	}
	aw.renderTemplate(w, "security_monitoring.html", data)
}

func (aw *AdminWeb) handleSystemLogs(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "系统日志",
	}
	aw.renderTemplate(w, "system_logs.html", data)
}

func (aw *AdminWeb) handleUserManagement(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "用户管理",
	}
	aw.renderTemplate(w, "user_management.html", data)
}

func (aw *AdminWeb) handleSecuritySettings(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "安全设置",
	}
	aw.renderTemplate(w, "security_settings.html", data)
}

func (aw *AdminWeb) handleAPISystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := &SystemMetrics{
		CPUUsage:    45.2,
		MemoryUsage: 67.8,
		DiskUsage:   23.4,
		NetworkIn:   1024 * 1024,
		NetworkOut:  2048 * 1024,
		Uptime:      3600 * 24 * 7,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (aw *AdminWeb) handleAPIAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := []Alert{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
