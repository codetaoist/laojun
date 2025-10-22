package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/laojun/internal/database"
	"github.com/gin-gonic/gin"
)

// @Description 提供索引管理功能，包括应用复合索引、获取索引信息和统计等
type IndexHandler struct {
	indexManager *database.IndexManager
}

// @Description 创建新的索引处理函数
// @Param indexManager body database.IndexManager true "索引管理器"
// @Success 200 {object} IndexHandler "索引处理函数"
// @Router /api/v1/indexes [post]
func NewIndexHandler(indexManager *database.IndexManager) *IndexHandler {
	return &IndexHandler{
		indexManager: indexManager,
	}
}

// ApplyCompositeIndexes 应用复合索引
// @Summary 应用复合索引
// @Description 应用所有复合索引以优化查询性能
// @Tags 索引管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "应用成功"
// @Failure 500 {object} map[string]interface{} "应用失败"
// @Router /api/v1/indexes/apply [post]
func (h *IndexHandler) ApplyCompositeIndexes(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.indexManager.ApplyCompositeIndexes(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "应用复合索引失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "复合索引应用成功",
		"status":  "success",
	})
}

// GetIndexInfo 获取索引信息
// @Summary 获取索引信息
// @Description 获取所有索引的详细信息
// @Tags 索引管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "索引信息"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /api/v1/indexes/info [get]
func (h *IndexHandler) GetIndexInfo(c *gin.Context) {
	ctx := c.Request.Context()

	indexes, err := h.indexManager.GetIndexInfo(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取索引信息失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取索引信息成功",
		"data":    indexes,
		"count":   len(indexes),
	})
}

// GetIndexStats 获取索引统计信息
// @Summary 获取索引统计信息
// @Description 获取索引使用统计和性能数据
// @Tags 索引管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "索引统计信息"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /api/v1/indexes/stats [get]
func (h *IndexHandler) GetIndexStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.indexManager.GetIndexStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取索引统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取索引统计成功",
		"data":    stats,
		"count":   len(stats),
	})
}

// AnalyzeIndexUsage 分析索引使用情况
// @Summary 分析索引使用情况
// @Description 分析索引使用情况，包括未使用索引和使用频率最高的索引
// @Tags 索引管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "索引使用分析"
// @Failure 500 {object} map[string]interface{} "分析失败"
// @Router /api/v1/indexes/analyze [get]
func (h *IndexHandler) AnalyzeIndexUsage(c *gin.Context) {
	ctx := c.Request.Context()

	analysis, err := h.indexManager.AnalyzeIndexUsage(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "分析索引使用情况失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "索引使用分析完成",
		"data":    analysis,
	})
}

// DropUnusedIndexes 删除未使用的索引
// @Summary 删除未使用的索引
// @Description 删除未使用的索引以释放存储空间
// @Tags 索引管理
// @Accept json
// @Produce json
// @Param dry_run query bool false "是否为试运行模式"
// @Success 200 {object} map[string]interface{} "删除结果"
// @Failure 500 {object} map[string]interface{} "删除失败"
// @Router /api/v1/indexes/cleanup [delete]
func (h *IndexHandler) DropUnusedIndexes(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取dry_run参数
	dryRunStr := c.DefaultQuery("dry_run", "true")
	dryRun, err := strconv.ParseBool(dryRunStr)
	if err != nil {
		dryRun = true // 默认为试运行模式
	}

	droppedIndexes, err := h.indexManager.DropUnusedIndexes(ctx, dryRun)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除未使用索引失败",
			"details": err.Error(),
		})
		return
	}

	message := "未使用索引删除完成"
	if dryRun {
		message = "未使用索引分析完成（试运行模式）"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         message,
		"dropped_indexes": droppedIndexes,
		"count":           len(droppedIndexes),
		"dry_run":         dryRun,
	})
}

// ReindexTable 重建表索引
// @Summary 重建表索引
// @Description 重建指定表的所有索引以优化查询性能
// @Tags 索引管理
// @Accept json
// @Produce json
// @Param table_name path string true "表名"
// @Success 200 {object} map[string]interface{} "重建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "重建失败"
// @Router /api/v1/indexes/reindex/{table_name} [post]
func (h *IndexHandler) ReindexTable(c *gin.Context) {
	ctx := c.Request.Context()
	tableName := c.Param("table_name")

	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "表名不能为空",
		})
		return
	}

	// 验证表名格式（只允许lj_开头的表）
	if len(tableName) < 3 || tableName[:3] != "lj_" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "只能重建lj_开头的表索引",
		})
		return
	}

	err := h.indexManager.ReindexTable(ctx, tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "重建表索引失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "表索引重建完成",
		"table_name": tableName,
	})
}

// UpdateIndexStatistics 更新索引统计信息
// @Summary 更新索引统计信息
// @Description 更新所有表的索引统计信息，包括使用频率、大小等
// @Tags 索引管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "更新失败"
// @Router /api/v1/indexes/update-stats [post]
func (h *IndexHandler) UpdateIndexStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.indexManager.UpdateIndexStatistics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新索引统计信息失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "索引统计信息更新成功",
	})
}

// GetIndexRecommendations 获取索引优化建议
// @Summary 获取索引优化建议
// @Description 基于查询模式分析，提供索引优化建议
// @Tags 索引管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "优化建议"
// @Failure 500 {object} map[string]interface{} "分析失败"
// @Router /api/v1/indexes/recommendations [get]
func (h *IndexHandler) GetIndexRecommendations(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取索引使用分析
	analysis, err := h.indexManager.AnalyzeIndexUsage(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取索引分析失败",
			"details": err.Error(),
		})
		return
	}

	// 生成优化建议
	recommendations := generateIndexRecommendations(analysis)

	c.JSON(http.StatusOK, gin.H{
		"message":         "索引优化建议生成完成",
		"recommendations": recommendations,
		"analysis":        analysis,
	})
}

// generateIndexRecommendations 生成索引优化建议
func generateIndexRecommendations(analysis map[string]interface{}) []map[string]interface{} {
	var recommendations []map[string]interface{}

	// 检查总体统计
	if overallStats, ok := analysis["overall_stats"].(map[string]interface{}); ok {
		if usageRate, ok := overallStats["usage_rate"].(float64); ok {
			if usageRate < 80 {
				recommendations = append(recommendations, map[string]interface{}{
					"type":        "warning",
					"category":    "索引使用",
					"title":       "索引使用率偏低",
					"description": "当前索引使用率为 " + strconv.FormatFloat(usageRate, 'f', 1, 64) + "%，建议清理未使用的索引",
					"action":      "运行索引清理功能删除未使用的索引",
					"priority":    "medium",
				})
			}
		}

		if unusedCount, ok := overallStats["unused_indexes"].(int); ok && unusedCount > 5 {
			recommendations = append(recommendations, map[string]interface{}{
				"type":        "warning",
				"category":    "未使用索引",
				"title":       "存在大量未使用索引",
				"description": "发现 " + strconv.Itoa(unusedCount) + " 个未使用的索引，占用存储空间",
				"action":      "建议删除未使用的索引以释放存储空间",
				"priority":    "high",
			})
		}
	}

	// 检查未使用的索引
	if unusedIndexes, ok := analysis["unused_indexes"].([]map[string]interface{}); ok {
		if len(unusedIndexes) > 0 {
			recommendations = append(recommendations, map[string]interface{}{
				"type":        "info",
				"category":    "存储优化",
				"title":       "可优化存储空间",
				"description": "删除未使用的索引可以释放存储空间并提高写入性能",
				"action":      "使用索引清理功能删除未使用的索引",
				"priority":    "low",
			})
		}
	}

	// 检查高频使用的索引
	if mostUsedIndexes, ok := analysis["most_used_indexes"].([]map[string]interface{}); ok {
		if len(mostUsedIndexes) > 0 {
			recommendations = append(recommendations, map[string]interface{}{
				"type":        "success",
				"category":    "性能优化",
				"title":       "索引使用良好",
				"description": "发现高频使用的索引，说明索引策略有效",
				"action":      "继续监控索引使用情况，定期更新统计信息",
				"priority":    "low",
			})
		}
	}

	// 添加通用建议
	recommendations = append(recommendations, map[string]interface{}{
		"type":        "info",
		"category":    "维护建议",
		"title":       "定期维护索引",
		"description": "建议定期更新索引统计信息以保持查询优化器的准确",
		"action":      "每周运行一次索引统计信息更新",
		"priority":    "low",
	})

	return recommendations
}
