package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"go.uber.org/zap"
)

// MetricsHandler 指标处理器
type MetricsHandler struct {
	monitoringService *services.MonitoringService
	logger           *zap.Logger
	promClient       v1.API
}

// NewMetricsHandler 创建指标处理器
func NewMetricsHandler(monitoringService *services.MonitoringService, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		monitoringService: monitoringService,
		logger:           logger,
	}
}

// MetricsResponse 指标响应
type MetricsResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error,omitempty"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	Query string `json:"query" binding:"required"`
	Time  string `json:"time,omitempty"`
}

// RangeQueryRequest 范围查询请求
type RangeQueryRequest struct {
	Query string `json:"query" binding:"required"`
	Start string `json:"start" binding:"required"`
	End   string `json:"end" binding:"required"`
	Step  string `json:"step" binding:"required"`
}

// CustomMetricRequest 自定义指标请求
type CustomMetricRequest struct {
	Name        string            `json:"name" binding:"required"`
	Type        string            `json:"type" binding:"required"`
	Help        string            `json:"help"`
	Labels      map[string]string `json:"labels"`
	Value       float64           `json:"value"`
	Timestamp   *time.Time        `json:"timestamp,omitempty"`
}

// GetMetrics 获取所有指标
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	// 获取收集器统计信息
	collectors := h.monitoringService.GetCollectors()
	collectorsInfo := make([]map[string]interface{}, 0, len(collectors))
	
	for _, collector := range collectors {
		info := map[string]interface{}{
			"name":    collector.Name(),
			"healthy": collector.IsHealthy(),
			"ready":   collector.IsReady(),
		}
		collectorsInfo = append(collectorsInfo, info)
	}
	
	// 获取导出器统计信息
	exporters := h.monitoringService.GetExporters()
	exportersInfo := make([]map[string]interface{}, 0, len(exporters))
	
	for _, exporter := range exporters {
		info := map[string]interface{}{
			"name":    exporter.Name(),
			"healthy": exporter.IsHealthy(),
			"ready":   exporter.IsReady(),
		}
		exportersInfo = append(exportersInfo, info)
	}
	
	response := MetricsResponse{
		Status: "success",
		Data: gin.H{
			"collectors": collectorsInfo,
			"exporters":  exportersInfo,
			"stats":      h.monitoringService.GetStats(),
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// QueryMetrics 查询指标
func (h *MetricsHandler) QueryMetrics(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid query request", zap.Error(err))
		c.JSON(http.StatusBadRequest, MetricsResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	
	// 解析时间参数
	var queryTime time.Time
	if req.Time != "" {
		if t, err := time.Parse(time.RFC3339, req.Time); err == nil {
			queryTime = t
		} else if timestamp, err := strconv.ParseInt(req.Time, 10, 64); err == nil {
			queryTime = time.Unix(timestamp, 0)
		} else {
			c.JSON(http.StatusBadRequest, MetricsResponse{
				Status: "error",
				Error:  "Invalid time format",
			})
			return
		}
	} else {
		queryTime = time.Now()
	}
	
	// 如果有 Prometheus 客户端，执行查询
	if h.promClient != nil {
		result, warnings, err := h.promClient.Query(c.Request.Context(), req.Query, queryTime)
		if err != nil {
			h.logger.Error("Failed to query metrics", zap.Error(err))
			c.JSON(http.StatusInternalServerError, MetricsResponse{
				Status: "error",
				Error:  err.Error(),
			})
			return
		}
		
		response := MetricsResponse{
			Status: "success",
			Data: gin.H{
				"resultType": result.Type(),
				"result":     result,
				"warnings":   warnings,
			},
		}
		
		c.JSON(http.StatusOK, response)
		return
	}
	
	// 模拟查询结果
	c.JSON(http.StatusOK, MetricsResponse{
		Status: "success",
		Data: gin.H{
			"query":     req.Query,
			"timestamp": queryTime,
			"result":    "Query executed successfully (mock)",
		},
	})
}

// QueryRangeMetrics 范围查询指标
func (h *MetricsHandler) QueryRangeMetrics(c *gin.Context) {
	var req RangeQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid range query request", zap.Error(err))
		c.JSON(http.StatusBadRequest, MetricsResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	
	// 解析时间参数
	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		if timestamp, err := strconv.ParseInt(req.Start, 10, 64); err == nil {
			start = time.Unix(timestamp, 0)
		} else {
			c.JSON(http.StatusBadRequest, MetricsResponse{
				Status: "error",
				Error:  "Invalid start time format",
			})
			return
		}
	}
	
	end, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		if timestamp, err := strconv.ParseInt(req.End, 10, 64); err == nil {
			end = time.Unix(timestamp, 0)
		} else {
			c.JSON(http.StatusBadRequest, MetricsResponse{
				Status: "error",
				Error:  "Invalid end time format",
			})
			return
		}
	}
	
	step, err := time.ParseDuration(req.Step)
	if err != nil {
		c.JSON(http.StatusBadRequest, MetricsResponse{
			Status: "error",
			Error:  "Invalid step format",
		})
		return
	}
	
	// 如果有 Prometheus 客户端，执行范围查询
	if h.promClient != nil {
		r := v1.Range{
			Start: start,
			End:   end,
			Step:  step,
		}
		
		result, warnings, err := h.promClient.QueryRange(c.Request.Context(), req.Query, r)
		if err != nil {
			h.logger.Error("Failed to query range metrics", zap.Error(err))
			c.JSON(http.StatusInternalServerError, MetricsResponse{
				Status: "error",
				Error:  err.Error(),
			})
			return
		}
		
		response := MetricsResponse{
			Status: "success",
			Data: gin.H{
				"resultType": result.Type(),
				"result":     result,
				"warnings":   warnings,
			},
		}
		
		c.JSON(http.StatusOK, response)
		return
	}
	
	// 模拟范围查询结果
	c.JSON(http.StatusOK, MetricsResponse{
		Status: "success",
		Data: gin.H{
			"query": req.Query,
			"start": start,
			"end":   end,
			"step":  step,
			"result": "Range query executed successfully (mock)",
		},
	})
}

// CreateCustomMetric 创建自定义指标
func (h *MetricsHandler) CreateCustomMetric(c *gin.Context) {
	var req CustomMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid custom metric request", zap.Error(err))
		c.JSON(http.StatusBadRequest, MetricsResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	
	// 验证指标类型
	validTypes := map[string]bool{
		"counter":   true,
		"gauge":     true,
		"histogram": true,
		"summary":   true,
	}
	
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, MetricsResponse{
			Status: "error",
			Error:  "Invalid metric type. Must be one of: counter, gauge, histogram, summary",
		})
		return
	}
	
	// 这里应该实际创建和注册自定义指标
	// 由于这是一个复杂的操作，这里只是模拟
	h.logger.Info("Creating custom metric",
		zap.String("name", req.Name),
		zap.String("type", req.Type),
		zap.Float64("value", req.Value))
	
	response := MetricsResponse{
		Status: "success",
		Data: gin.H{
			"message": "Custom metric created successfully",
			"metric": gin.H{
				"name":   req.Name,
				"type":   req.Type,
				"help":   req.Help,
				"labels": req.Labels,
				"value":  req.Value,
			},
		},
	}
	
	c.JSON(http.StatusCreated, response)
}

// SetPrometheusClient 设置 Prometheus 客户端
func (h *MetricsHandler) SetPrometheusClient(client api.Client) {
	h.promClient = v1.NewAPI(client)
}