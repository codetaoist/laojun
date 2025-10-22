package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// IndexManager 索引管理器
type IndexManager struct {
	db *sql.DB
}

// NewIndexManager 创建新的索引管理器
func NewIndexManager(db *sql.DB) *IndexManager {
	return &IndexManager{db: db}
}

// IndexInfo 索引信息
type IndexInfo struct {
	IndexName   string `json:"index_name"`
	TableName   string `json:"table_name"`
	ColumnNames string `json:"column_names"`
	IsUnique    bool   `json:"is_unique"`
	IndexSize   string `json:"index_size"`
	CreatedAt   string `json:"created_at"`
}

// IndexStats 索引统计信息
type IndexStats struct {
	IndexName    string `json:"index_name"`
	TableName    string `json:"table_name"`
	IndexScans   int64  `json:"index_scans"`
	TupleReads   int64  `json:"tuple_reads"`
	TupleFetches int64  `json:"tuple_fetches"`
	IndexSize    string `json:"index_size"`
}

// ApplyCompositeIndexes 应用复合索引
func (im *IndexManager) ApplyCompositeIndexes(ctx context.Context) error {
	log.Println("开始应用复合索引...")

	// 读取复合索引SQL文件
	compositeIndexSQL := `
-- 用户认证和会话管理复合索引
CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_user_expires 
ON lj_user_sessions(user_id, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_token_expires 
ON lj_user_sessions(token_hash, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_user_active 
ON lj_user_sessions(user_id, expires_at, last_used_at);

CREATE INDEX IF NOT EXISTS idx_lj_jwt_keys_active_expires 
ON lj_jwt_keys(is_active, expires_at);

-- 权限查询优化复合索引
CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_group_device 
ON lj_user_group_permissions(group_id, device_type_id, permission_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_perm_granted 
ON lj_user_group_permissions(permission_id, granted, group_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_user_device 
ON lj_user_device_permissions(user_id, device_type_id, permission_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_user_valid 
ON lj_user_device_permissions(user_id, granted, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_cleanup 
ON lj_user_device_permissions(expires_at) WHERE expires_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_lj_permission_inheritance_parent_type 
ON lj_permission_inheritance(parent_permission_id, inheritance_type);

-- 扩展权限查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_device_module 
ON lj_extended_permissions(device_type_id, module_id, resource, action);

CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_module_element 
ON lj_extended_permissions(module_id, element_type, element_code);

CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_resource_action 
ON lj_extended_permissions(resource, action, device_type_id);

-- 层级结构查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_menus_parent_sort 
ON lj_menus(parent_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_menus_visible_sort 
ON lj_menus(is_hidden, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_modules_parent_active_sort 
ON lj_modules(parent_id, is_active, sort_order);

-- 状态和时间范围查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_users_active_created 
ON lj_users(is_active, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_users_active_login 
ON lj_users(is_active, last_login_at) WHERE last_login_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_lj_device_types_active_sort 
ON lj_device_types(is_active, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_user_groups_active_created 
ON lj_user_groups(is_active, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_user_created 
ON lj_user_group_members(user_id, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_group_created 
ON lj_user_group_members(group_id, created_at);

-- 权限模板查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_permission_templates_system_created 
ON lj_permission_templates(is_system, created_at);

-- 角色和权限关联复合索引
CREATE INDEX IF NOT EXISTS idx_lj_roles_system_created 
ON lj_roles(is_system, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_permissions_resource_action 
ON lj_permissions(resource, action);

-- 性能监控和统计查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_users_stats 
ON lj_users(created_at, is_active);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_stats 
ON lj_user_sessions(created_at, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_stats 
ON lj_user_device_permissions(permission_id, device_type_id, created_at);
`

	// 分割SQL语句并执行
	statements := strings.Split(compositeIndexSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		log.Printf("执行索引创建: %s", strings.Split(stmt, "\n")[0])
		if _, err := im.db.ExecContext(ctx, stmt); err != nil {
			log.Printf("创建索引失败: %v", err)
			return fmt.Errorf("创建索引失败: %w", err)
		}
	}

	log.Println("复合索引应用完成")
	return nil
}

// GetIndexInfo 获取索引信息
func (im *IndexManager) GetIndexInfo(ctx context.Context) ([]IndexInfo, error) {
	query := `
	SELECT 
		i.indexname as index_name,
		i.tablename as table_name,
		array_to_string(array_agg(a.attname ORDER BY a.attnum), ', ') as column_names,
		i.indexdef LIKE '%UNIQUE%' as is_unique,
		pg_size_pretty(pg_relation_size(c.oid)) as index_size,
		obj_description(c.oid) as created_at
	FROM pg_indexes i
	JOIN pg_class c ON c.relname = i.indexname
	JOIN pg_index idx ON idx.indexrelid = c.oid
	JOIN pg_attribute a ON a.attrelid = idx.indrelid AND a.attnum = ANY(idx.indkey)
	WHERE i.schemaname = 'public' 
		AND i.tablename LIKE 'lj_%'
		AND i.indexname LIKE 'idx_%'
	GROUP BY i.indexname, i.tablename, i.indexdef, c.oid
	ORDER BY i.tablename, i.indexname;
	`

	rows, err := im.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询索引信息失败: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var index IndexInfo
		var createdAt sql.NullString

		err := rows.Scan(
			&index.IndexName,
			&index.TableName,
			&index.ColumnNames,
			&index.IsUnique,
			&index.IndexSize,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描索引信息失败: %w", err)
		}

		if createdAt.Valid {
			index.CreatedAt = createdAt.String
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

// GetIndexStats 获取索引统计信息
func (im *IndexManager) GetIndexStats(ctx context.Context) ([]IndexStats, error) {
	query := `
	SELECT 
		i.indexrelname as index_name,
		i.relname as table_name,
		i.idx_scan as index_scans,
		i.idx_tup_read as tuple_reads,
		i.idx_tup_fetch as tuple_fetches,
		pg_size_pretty(pg_relation_size(c.oid)) as index_size
	FROM pg_stat_user_indexes i
	JOIN pg_class c ON c.relname = i.indexrelname
	WHERE i.relname LIKE 'lj_%'
		AND i.indexrelname LIKE 'idx_%'
	ORDER BY i.idx_scan DESC, i.relname, i.indexrelname;
	`

	rows, err := im.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询索引统计失败: %w", err)
	}
	defer rows.Close()

	var stats []IndexStats
	for rows.Next() {
		var stat IndexStats

		err := rows.Scan(
			&stat.IndexName,
			&stat.TableName,
			&stat.IndexScans,
			&stat.TupleReads,
			&stat.TupleFetches,
			&stat.IndexSize,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描索引统计失败: %w", err)
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// AnalyzeIndexUsage 分析索引使用情况
func (im *IndexManager) AnalyzeIndexUsage(ctx context.Context) (map[string]interface{}, error) {
	// 查询未使用的索引
	unusedIndexQuery := `
	SELECT 
		schemaname,
		tablename,
		indexname,
		pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size
	FROM pg_stat_user_indexes 
	WHERE idx_scan = 0 
		AND schemaname = 'public'
		AND tablename LIKE 'lj_%'
		AND indexname LIKE 'idx_%'
	ORDER BY pg_relation_size(indexname::regclass) DESC;
	`

	// 查询使用频率最高的索引
	mostUsedQuery := `
	SELECT 
		schemaname,
		tablename,
		indexname,
		idx_scan,
		pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size
	FROM pg_stat_user_indexes 
	WHERE schemaname = 'public'
		AND tablename LIKE 'lj_%'
		AND indexname LIKE 'idx_%'
	ORDER BY idx_scan DESC
	LIMIT 10;
	`

	// 查询索引总体统计
	overallStatsQuery := `
	SELECT 
		COUNT(*) as total_indexes,
		COUNT(CASE WHEN idx_scan = 0 THEN 1 END) as unused_indexes,
		COUNT(CASE WHEN idx_scan > 0 THEN 1 END) as used_indexes,
		pg_size_pretty(SUM(pg_relation_size(indexname::regclass))) as total_size
	FROM pg_stat_user_indexes 
	WHERE schemaname = 'public'
		AND tablename LIKE 'lj_%'
		AND indexname LIKE 'idx_%';
	`

	result := make(map[string]interface{})

	// 获取未使用的索引
	rows, err := im.db.QueryContext(ctx, unusedIndexQuery)
	if err != nil {
		return nil, fmt.Errorf("查询未使用索引失败: %w", err)
	}
	defer rows.Close()

	var unusedIndexes []map[string]interface{}
	for rows.Next() {
		var schema, table, index, size string
		if err := rows.Scan(&schema, &table, &index, &size); err != nil {
			return nil, err
		}
		unusedIndexes = append(unusedIndexes, map[string]interface{}{
			"schema": schema,
			"table":  table,
			"index":  index,
			"size":   size,
		})
	}
	result["unused_indexes"] = unusedIndexes

	// 获取使用频率最高的索引
	rows, err = im.db.QueryContext(ctx, mostUsedQuery)
	if err != nil {
		return nil, fmt.Errorf("查询最常用索引失败: %w", err)
	}
	defer rows.Close()

	var mostUsedIndexes []map[string]interface{}
	for rows.Next() {
		var schema, table, index, size string
		var scans int64
		if err := rows.Scan(&schema, &table, &index, &scans, &size); err != nil {
			return nil, err
		}
		mostUsedIndexes = append(mostUsedIndexes, map[string]interface{}{
			"schema": schema,
			"table":  table,
			"index":  index,
			"scans":  scans,
			"size":   size,
		})
	}
	result["most_used_indexes"] = mostUsedIndexes

	// 获取总体统计
	var totalIndexes, unusedCount, usedCount int
	var totalSize string
	err = im.db.QueryRowContext(ctx, overallStatsQuery).Scan(
		&totalIndexes, &unusedCount, &usedCount, &totalSize,
	)
	if err != nil {
		return nil, fmt.Errorf("查询索引总体统计失败: %w", err)
	}

	result["overall_stats"] = map[string]interface{}{
		"total_indexes":  totalIndexes,
		"unused_indexes": unusedCount,
		"used_indexes":   usedCount,
		"total_size":     totalSize,
		"usage_rate":     float64(usedCount) / float64(totalIndexes) * 100,
	}

	return result, nil
}

// DropUnusedIndexes 删除未使用的索引
func (im *IndexManager) DropUnusedIndexes(ctx context.Context, dryRun bool) ([]string, error) {
	query := `
	SELECT indexname
	FROM pg_stat_user_indexes 
	WHERE idx_scan = 0 
		AND schemaname = 'public'
		AND tablename LIKE 'lj_%'
		AND indexname LIKE 'idx_%'
		AND indexname NOT LIKE '%_pkey'
		AND indexname NOT LIKE '%_unique%'
	ORDER BY pg_relation_size(indexname::regclass) DESC;
	`

	rows, err := im.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询未使用索引失败: %w", err)
	}
	defer rows.Close()

	var droppedIndexes []string
	for rows.Next() {
		var indexName string
		if err := rows.Scan(&indexName); err != nil {
			return nil, err
		}

		if dryRun {
			droppedIndexes = append(droppedIndexes, indexName+" (dry run)")
			log.Printf("将删除未使用索引 (dry run): %s", indexName)
		} else {
			dropSQL := fmt.Sprintf("DROP INDEX IF EXISTS %s;", indexName)
			if _, err := im.db.ExecContext(ctx, dropSQL); err != nil {
				log.Printf("删除索引失败 %s: %v", indexName, err)
				continue
			}
			droppedIndexes = append(droppedIndexes, indexName)
			log.Printf("已删除未使用索引: %s", indexName)
		}
	}

	return droppedIndexes, nil
}

// ReindexTable 重建表索引
func (im *IndexManager) ReindexTable(ctx context.Context, tableName string) error {
	log.Printf("开始重建表索引: %s", tableName)

	reindexSQL := fmt.Sprintf("REINDEX TABLE %s;", tableName)
	if _, err := im.db.ExecContext(ctx, reindexSQL); err != nil {
		return fmt.Errorf("重建表索引失败: %w", err)
	}

	log.Printf("表索引重建完成: %s", tableName)
	return nil
}

// UpdateIndexStatistics 更新索引统计信息
func (im *IndexManager) UpdateIndexStatistics(ctx context.Context) error {
	log.Println("开始更新索引统计信息...")

	// 获取所有lj_开头的表
	query := `
	SELECT tablename 
	FROM pg_tables 
	WHERE schemaname = 'public' 
		AND tablename LIKE 'lj_%'
	ORDER BY tablename;
	`

	rows, err := im.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	// 为每个表更新统计信息
	for _, table := range tables {
		analyzeSQL := fmt.Sprintf("ANALYZE %s;", table)
		if _, err := im.db.ExecContext(ctx, analyzeSQL); err != nil {
			log.Printf("更新表统计信息失败: %s: %v", table, err)
			continue
		}
		log.Printf("已更新表统计信息: %s", table)
	}

	log.Println("索引统计信息更新完成")
	return nil
}
