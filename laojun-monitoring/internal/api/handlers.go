package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/codetaoist/laojun-monitoring/internal/alerting"
	"github.com/codetaoist/laojun-monitoring/internal/collectors"
	"github.com/codetaoist/laojun-monitoring/internal/logging"
	"github.com/codetaoist/laojun-monitoring/internal/metrics"
	"github.com/codetaoist/laojun-monitoring/internal/storage"
)

// Handler API处理器
type Handler struct {
	logger          *zap.Logger
	metricRegistry  *metrics.MetricRegistry
	storageManager  *storage.StorageManager
	alertManager    *alerting.AlertManager
	collectorManager *collectors.CollectorManager
	pipelineManager *logging.PipelineManager
}

// NewHandler 创建新的API处理器
func NewHandler(
	logger *zap.Logger,
	metricRegistry *metrics.MetricRegistry,
	storageManager *storage.StorageManager,
	alertManager *alerting.AlertManager,
	collectorManager *collectors.CollectorManager,
	pipelineManager *logging.PipelineManager,
) *Handler {
	return &Handler{
		logger:          logger,
		metricRegistry:  metricRegistry,
		storageManager:  storageManager,
		alertManager:    alertManager,
		collectorManager: collectorManager,
		pipelineManager: pipelineManager,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// 健康检查
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
	router.HandleFunc("/ready", h.ReadinessCheck).Methods("GET")
	
	// 指标相关
	metricsRouter := router.PathPrefix("/api/v1/metrics").Subrouter()
	metricsRouter.HandleFunc("", h.ListMetrics).Methods("GET")
	metricsRouter.HandleFunc("/{name}", h.GetMetric).Methods("GET")
	metricsRouter.HandleFunc("/{name}", h.UpdateMetric).Methods("POST")
	metricsRouter.HandleFunc("/{name}", h.DeleteMetric).Methods("DELETE")
	metricsRouter.HandleFunc("/query", h.QueryMetrics).Methods("GET", "POST")
	metricsRouter.HandleFunc("/query_range", h.QueryRangeMetrics).Methods("GET", "POST")
	
	// 告警相关
	alertsRouter := router.PathPrefix("/api/v1/alerts").Subrouter()
	alertsRouter.HandleFunc("", h.ListAlerts).Methods("GET")
	alertsRouter.HandleFunc("/{id}", h.GetAlert).Methods("GET")
	alertsRouter.HandleFunc("/rules", h.ListAlertRules).Methods("GET")
	alertsRouter.HandleFunc("/rules", h.CreateAlertRule).Methods("POST")
	alertsRouter.HandleFunc("/rules/{id}", h.GetAlertRule).Methods("GET")
	alertsRouter.HandleFunc("/rules/{id}", h.UpdateAlertRule).Methods("PUT")
	alertsRouter.HandleFunc("/rules/{id}", h.DeleteAlertRule).Methods("DELETE")
	
	// 收集器相关
	collectorsRouter := router.PathPrefix("/api/v1/collectors").Subrouter()
	collectorsRouter.HandleFunc("", h.ListCollectors).Methods("GET")
	collectorsRouter.HandleFunc("/{name}", h.GetCollector).Methods("GET")
	collectorsRouter.HandleFunc("/{name}/start", h.StartCollector).Methods("POST")
	collectorsRouter.HandleFunc("/{name}/stop", h.StopCollector).Methods("POST")
	collectorsRouter.HandleFunc("/{name}/stats", h.GetCollectorStats).Methods("GET")
	
	// 日志管道相关
	pipelinesRouter := router.PathPrefix("/api/v1/pipelines").Subrouter()
	pipelinesRouter.HandleFunc("", h.ListPipelines).Methods("GET")
	pipelinesRouter.HandleFunc("/{name}", h.GetPipeline).Methods("GET")
	pipelinesRouter.HandleFunc("/{name}/start", h.StartPipeline).Methods("POST")
	pipelinesRouter.HandleFunc("/{name}/stop", h.StopPipeline).Methods("POST")
	pipelinesRouter.HandleFunc("/{name}/stats", h.GetPipelineStats).Methods("GET")
	
	// 系统状态
	statusRouter := router.PathPrefix("/api/v1/status").Subrouter()
	statusRouter.HandleFunc("", h.GetSystemStatus).Methods("GET")
	statusRouter.HandleFunc("/stats", h.GetSystemStats).Methods("GET")
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// ReadinessCheck 就绪检查
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ready := true
	checks := make(map[string]bool)
	
	// 检查存储
	if h.storageManager != nil {
		checks["storage"] = h.storageManager.IsHealthy()
		if !checks["storage"] {
			ready = false
		}
	}
	
	// 检查告警管理器
	if h.alertManager != nil {
		checks["alerting"] = h.alertManager.IsRunning()
		if !checks["alerting"] {
			ready = false
		}
	}
	
	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}
	
	response := map[string]interface{}{
		"ready":     ready,
		"checks":    checks,
		"timestamp": time.Now().Unix(),
	}
	
	h.writeJSONResponse(w, status, response)
}

// ListMetrics 列出所有指标
func (h *Handler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	if h.metricRegistry == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Metric registry not available")
		return
	}
	
	metrics := h.metricRegistry.ListMetrics()
	
	response := map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetMetric 获取指定指标
func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.metricRegistry == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Metric registry not available")
		return
	}
	
	metric, err := h.metricRegistry.GetMetric(name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Metric not found: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, metric)
}

// UpdateMetric 更新指标值
func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.metricRegistry == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Metric registry not available")
		return
	}
	
	var request struct {
		Value  float64           `json:"value"`
		Labels map[string]string `json:"labels,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	
	if err := h.metricRegistry.UpdateMetricValue(name, request.Value, request.Labels); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update metric: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteMetric 删除指标
func (h *Handler) DeleteMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.metricRegistry == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Metric registry not available")
		return
	}
	
	if err := h.metricRegistry.UnregisterMetric(name); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete metric: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// QueryMetrics 查询指标
func (h *Handler) QueryMetrics(w http.ResponseWriter, r *http.Request) {
	if h.storageManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Storage manager not available")
		return
	}
	
	query := r.URL.Query().Get("query")
	if query == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Query parameter is required")
		return
	}
	
	timeParam := r.URL.Query().Get("time")
	var queryTime time.Time
	if timeParam != "" {
		if timestamp, err := strconv.ParseInt(timeParam, 10, 64); err == nil {
			queryTime = time.Unix(timestamp, 0)
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid time parameter")
			return
		}
	} else {
		queryTime = time.Now()
	}
	
	storage := h.storageManager.GetStorage()
	if storage == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Storage not available")
		return
	}
	
	result, err := storage.Query(r.Context(), query, queryTime)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, result)
}

// QueryRangeMetrics 范围查询指标
func (h *Handler) QueryRangeMetrics(w http.ResponseWriter, r *http.Request) {
	if h.storageManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Storage manager not available")
		return
	}
	
	query := r.URL.Query().Get("query")
	if query == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Query parameter is required")
		return
	}
	
	startParam := r.URL.Query().Get("start")
	endParam := r.URL.Query().Get("end")
	stepParam := r.URL.Query().Get("step")
	
	if startParam == "" || endParam == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Start and end parameters are required")
		return
	}
	
	start, err := strconv.ParseInt(startParam, 10, 64)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid start parameter")
		return
	}
	
	end, err := strconv.ParseInt(endParam, 10, 64)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid end parameter")
		return
	}
	
	step := time.Minute // 默认步长
	if stepParam != "" {
		if stepDuration, err := time.ParseDuration(stepParam); err == nil {
			step = stepDuration
		}
	}
	
	queryRange := storage.QueryRange{
		Start: time.Unix(start, 0),
		End:   time.Unix(end, 0),
		Step:  step,
	}
	
	storage := h.storageManager.GetStorage()
	if storage == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Storage not available")
		return
	}
	
	result, err := storage.QueryRange(r.Context(), query, queryRange)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Range query failed: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, result)
}

// ListAlerts 列出所有告警
func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	if h.alertManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Alert manager not available")
		return
	}
	
	alerts := h.alertManager.GetAlerts()
	
	response := map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAlert 获取指定告警
func (h *Handler) GetAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if h.alertManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Alert manager not available")
		return
	}
	
	alert := h.alertManager.GetAlert(id)
	if alert == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Alert not found")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, alert)
}

// ListAlertRules 列出所有告警规则
func (h *Handler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	if h.alertManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Alert manager not available")
		return
	}
	
	rules := h.alertManager.ListRules()
	
	response := map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// CreateAlertRule 创建告警规则
func (h *Handler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	if h.alertManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Alert manager not available")
		return
	}
	
	var rule alerting.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	
	if err := h.alertManager.AddRule(&rule); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create rule: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusCreated, rule)
}

// GetAlertRule 获取指定告警规则
func (h *Handler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if h.alertManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Alert manager not available")
		return
	}
	
	rule := h.alertManager.GetRule(id)
	if rule == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Alert rule not found")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, rule)
}

// UpdateAlertRule 更新告警规则
func (h *Handler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if h.alertManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Alert manager not available")
		return
	}
	
	var rule alerting.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	
	rule.ID = id
	
	if err := h.alertManager.UpdateRule(&rule); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update rule: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, rule)
}

// DeleteAlertRule 删除告警规则
func (h *Handler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if h.alertManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Alert manager not available")
		return
	}
	
	if err := h.alertManager.RemoveRule(id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete rule: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListCollectors 列出所有收集器
func (h *Handler) ListCollectors(w http.ResponseWriter, r *http.Request) {
	if h.collectorManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Collector manager not available")
		return
	}
	
	collectors := h.collectorManager.ListCollectors()
	
	response := map[string]interface{}{
		"collectors": collectors,
		"count":      len(collectors),
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetCollector 获取指定收集器
func (h *Handler) GetCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.collectorManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Collector manager not available")
		return
	}
	
	collector, err := h.collectorManager.GetCollector(name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Collector not found: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, collector)
}

// StartCollector 启动收集器
func (h *Handler) StartCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.collectorManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Collector manager not available")
		return
	}
	
	if err := h.collectorManager.StartCollector(name); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start collector: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "started"})
}

// StopCollector 停止收集器
func (h *Handler) StopCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.collectorManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Collector manager not available")
		return
	}
	
	if err := h.collectorManager.StopCollector(name); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop collector: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// GetCollectorStats 获取收集器统计
func (h *Handler) GetCollectorStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.collectorManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Collector manager not available")
		return
	}
	
	stats, err := h.collectorManager.GetCollectorStats(name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Collector stats not found: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, stats)
}

// ListPipelines 列出所有管道
func (h *Handler) ListPipelines(w http.ResponseWriter, r *http.Request) {
	if h.pipelineManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Pipeline manager not available")
		return
	}
	
	pipelines := h.pipelineManager.ListPipelines()
	
	response := map[string]interface{}{
		"pipelines": pipelines,
		"count":     len(pipelines),
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetPipeline 获取指定管道
func (h *Handler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.pipelineManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Pipeline manager not available")
		return
	}
	
	pipeline, err := h.pipelineManager.GetPipeline(name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Pipeline not found: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, pipeline)
}

// StartPipeline 启动管道
func (h *Handler) StartPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.pipelineManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Pipeline manager not available")
		return
	}
	
	pipeline, err := h.pipelineManager.GetPipeline(name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Pipeline not found: %v", err))
		return
	}
	
	if err := pipeline.Start(); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start pipeline: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "started"})
}

// StopPipeline 停止管道
func (h *Handler) StopPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.pipelineManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Pipeline manager not available")
		return
	}
	
	pipeline, err := h.pipelineManager.GetPipeline(name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Pipeline not found: %v", err))
		return
	}
	
	if err := pipeline.Stop(); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop pipeline: %v", err))
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// GetPipelineStats 获取管道统计
func (h *Handler) GetPipelineStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if h.pipelineManager == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Pipeline manager not available")
		return
	}
	
	pipeline, err := h.pipelineManager.GetPipeline(name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Pipeline not found: %v", err))
		return
	}
	
	stats := pipeline.GetStats()
	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GetSystemStatus 获取系统状态
func (h *Handler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(time.Now()).String(), // 这里应该是实际的启动时间
		"version":   "1.0.0",
	}
	
	// 添加各组件状态
	components := make(map[string]interface{})
	
	if h.metricRegistry != nil {
		components["metrics"] = map[string]interface{}{
			"status": "running",
			"count":  len(h.metricRegistry.ListMetrics()),
		}
	}
	
	if h.storageManager != nil {
		components["storage"] = map[string]interface{}{
			"status":  "running",
			"healthy": h.storageManager.IsHealthy(),
		}
	}
	
	if h.alertManager != nil {
		components["alerting"] = map[string]interface{}{
			"status":  "running",
			"running": h.alertManager.IsRunning(),
		}
	}
	
	if h.collectorManager != nil {
		collectors := h.collectorManager.ListCollectors()
		components["collectors"] = map[string]interface{}{
			"status": "running",
			"count":  len(collectors),
		}
	}
	
	if h.pipelineManager != nil {
		pipelines := h.pipelineManager.ListPipelines()
		components["pipelines"] = map[string]interface{}{
			"status": "running",
			"count":  len(pipelines),
		}
	}
	
	status["components"] = components
	
	h.writeJSONResponse(w, http.StatusOK, status)
}

// GetSystemStats 获取系统统计
func (h *Handler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})
	
	// 指标统计
	if h.metricRegistry != nil {
		stats["metrics"] = map[string]interface{}{
			"total_count": len(h.metricRegistry.ListMetrics()),
		}
	}
	
	// 告警统计
	if h.alertManager != nil {
		alertStats := h.alertManager.GetStats()
		stats["alerts"] = alertStats
	}
	
	// 收集器统计
	if h.collectorManager != nil {
		collectorStats := h.collectorManager.GetAllStats()
		stats["collectors"] = collectorStats
	}
	
	// 管道统计
	if h.pipelineManager != nil {
		pipelineStats := h.pipelineManager.GetStats()
		stats["pipelines"] = pipelineStats
	}
	
	// 存储统计
	if h.storageManager != nil {
		storageStats := h.storageManager.GetStats()
		stats["storage"] = storageStats
	}
	
	h.writeJSONResponse(w, http.StatusOK, stats)
}

// writeJSONResponse 写入JSON响应
func (h *Handler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// writeErrorResponse 写入错误响应
func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.logger.Error("API error", 
		zap.Int("status_code", statusCode), 
		zap.String("message", message))
	
	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Unix(),
	}
	
	h.writeJSONResponse(w, statusCode, response)
}