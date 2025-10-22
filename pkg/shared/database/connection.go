package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// 主数据库配置
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string

	// 连接池配置
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生存时间
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
	// 读写分离配置
	ReadReplicas []ReplicaConfig

	// 日志配置
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration

	// 性能配置
	PrepareStmt                              bool // 预编译语句
	DisableForeignKeyConstraintWhenMigrating bool
}

// ReplicaConfig 从库配置
type ReplicaConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	db     *gorm.DB
	config DatabaseConfig
}

// NewDatabaseManager 创建新的数据库管理器
func NewDatabaseManager(config DatabaseConfig) (*DatabaseManager, error) {
	// 设置默认值
	setDefaultConfig(&config)

	// 创建主数据库连接
	dsn := buildDSN(config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode, config.TimeZone)

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             config.SlowThreshold,
				LogLevel:                  config.LogLevel,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		PrepareStmt:                              config.PrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: config.DisableForeignKeyConstraintWhenMigrating,
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	configureConnectionPool(sqlDB, config)

	// 配置读写分离
	if len(config.ReadReplicas) > 0 {
		if err := configureReadWriteSplit(db, config); err != nil {
			return nil, fmt.Errorf("failed to configure read-write split: %w", err)
		}
	}

	return &DatabaseManager{
		db:     db,
		config: config,
	}, nil
}

// GetDB 获取数据库连接
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// Close 关闭数据库连接
func (dm *DatabaseManager) Close() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping 检查数据库连接
func (dm *DatabaseManager) Ping() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetStats 获取连接池统计信息
func (dm *DatabaseManager) GetStats() sql.DBStats {
	sqlDB, _ := dm.db.DB()
	return sqlDB.Stats()
}

// Transaction 执行事务
func (dm *DatabaseManager) Transaction(fn func(*gorm.DB) error) error {
	return dm.db.Transaction(fn)
}

// TransactionWithContext 带上下文的事务
func (dm *DatabaseManager) TransactionWithContext(ctx context.Context, fn func(*gorm.DB) error) error {
	return dm.db.WithContext(ctx).Transaction(fn)
}

// HealthCheck 健康检查
func (dm *DatabaseManager) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return dm.db.WithContext(ctx).Exec("SELECT 1").Error
}

// GetConnectionInfo 获取连接信息
func (dm *DatabaseManager) GetConnectionInfo() map[string]interface{} {
	stats := dm.GetStats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// OptimizeQueries 查询优化
func (dm *DatabaseManager) OptimizeQueries() *gorm.DB {
	return dm.db.Session(&gorm.Session{
		PrepareStmt: true,
	})
}

// ReadOnlyDB 获取只读数据库连接
func (dm *DatabaseManager) ReadOnlyDB() *gorm.DB {
	return dm.db.Clauses(dbresolver.Read)
}

// WriteDB 获取写数据库连接
func (dm *DatabaseManager) WriteDB() *gorm.DB {
	return dm.db.Clauses(dbresolver.Write)
}

// 辅助函数

// setDefaultConfig 设置默认配置
func setDefaultConfig(config *DatabaseConfig) {
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 100
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Hour
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 10 * time.Minute
	}
	if config.LogLevel == 0 {
		config.LogLevel = logger.Info
	}
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 200 * time.Millisecond
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.TimeZone == "" {
		config.TimeZone = "UTC"
	}
}

// buildDSN 构建数据源名
func buildDSN(host string, port int, user, password, dbname, sslmode, timezone string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, port, user, password, dbname, sslmode, timezone)
}

// configureConnectionPool 配置连接池
func configureConnectionPool(sqlDB *sql.DB, config DatabaseConfig) {
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
}

// configureReadWriteSplit 配置读写分离
func configureReadWriteSplit(db *gorm.DB, config DatabaseConfig) error {
	var replicas []gorm.Dialector

	for _, replica := range config.ReadReplicas {
		dsn := buildDSN(replica.Host, replica.Port, replica.User, replica.Password,
			replica.DBName, replica.SSLMode, replica.TimeZone)
		replicas = append(replicas, postgres.Open(dsn))
	}

	return db.Use(dbresolver.Register(dbresolver.Config{
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{}, // 随机选择从库
	}))
}

// QueryOptimizer 查询优化工具
type QueryOptimizer struct {
	db *gorm.DB
}

// NewQueryOptimizer 创建查询优化工具
func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

// OptimizeSelect 优化 SELECT 查询
func (qo *QueryOptimizer) OptimizeSelect() *gorm.DB {
	return qo.db.Session(&gorm.Session{
		PrepareStmt: true,
	})
}

// BatchInsert 批量插入优化
func (qo *QueryOptimizer) BatchInsert(data interface{}, batchSize int) error {
	return qo.db.CreateInBatches(data, batchSize).Error
}

// BatchUpdate 批量更新优化
func (qo *QueryOptimizer) BatchUpdate(model interface{}, updates map[string]interface{}, batchSize int) error {
	// 这里需要根据具体的业务逻辑实现批量更新
	// 可以使用原生 SQL 或者分批处理
	return qo.db.Model(model).Updates(updates).Error
}

// ExplainQuery 分析查询计划
func (qo *QueryOptimizer) ExplainQuery(query string, args ...interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)

	rows, err := qo.db.Raw(explainQuery, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var result map[string]interface{}
		if err := qo.db.ScanRows(rows, &result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// IndexManager 索引管理工具
type IndexManager struct {
	db *gorm.DB
}

// NewIndexManager 创建索引管理工具
func NewIndexManager(db *gorm.DB) *IndexManager {
	return &IndexManager{db: db}
}

// CreateIndex 创建索引
func (im *IndexManager) CreateIndex(tableName, indexName string, columns []string, unique bool) error {
	var indexType string
	if unique {
		indexType = "UNIQUE"
	}

	query := fmt.Sprintf("CREATE %s INDEX IF NOT EXISTS %s ON %s (%s)",
		indexType, indexName, tableName, joinColumns(columns))

	return im.db.Exec(query).Error
}

// DropIndex 删除索引
func (im *IndexManager) DropIndex(indexName string) error {
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)
	return im.db.Exec(query).Error
}

// ListIndexes 列出表的所有索引
func (im *IndexManager) ListIndexes(tableName string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			indexname as index_name,
			indexdef as index_definition
		FROM pg_indexes 
		WHERE tablename = ?
	`

	if err := im.db.Raw(query, tableName).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// AnalyzeIndexUsage 分析索引使用情况
func (im *IndexManager) AnalyzeIndexUsage(tableName string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_tup_read,
			idx_tup_fetch,
			idx_scan
		FROM pg_stat_user_indexes 
		WHERE tablename = ?
		ORDER BY idx_scan DESC
	`

	if err := im.db.Raw(query, tableName).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// joinColumns 连接列名
func joinColumns(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

// ConnectionMonitor 连接监控工具
type ConnectionMonitor struct {
	db *gorm.DB
}

// NewConnectionMonitor 创建连接监控工具
func NewConnectionMonitor(db *gorm.DB) *ConnectionMonitor {
	return &ConnectionMonitor{db: db}
}

// MonitorConnections 监控连接状态
func (cm *ConnectionMonitor) MonitorConnections() (map[string]interface{}, error) {
	var result map[string]interface{}

	query := `
		SELECT 
			count(*) as total_connections,
			count(*) FILTER (WHERE state = 'active') as active_connections,
			count(*) FILTER (WHERE state = 'idle') as idle_connections,
			count(*) FILTER (WHERE state = 'idle in transaction') as idle_in_transaction
		FROM pg_stat_activity 
		WHERE datname = current_database()
	`

	if err := cm.db.Raw(query).Scan(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// GetLongRunningQueries 获取长时间运行的查询
func (cm *ConnectionMonitor) GetLongRunningQueries(threshold time.Duration) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			pid,
			usename,
			application_name,
			client_addr,
			state,
			query_start,
			now() - query_start as duration,
			query
		FROM pg_stat_activity 
		WHERE state = 'active' 
		AND now() - query_start > interval '%d seconds'
		ORDER BY query_start
	`

	if err := cm.db.Raw(fmt.Sprintf(query, int(threshold.Seconds()))).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// KillConnection 终止连接
func (cm *ConnectionMonitor) KillConnection(pid int) error {
	query := "SELECT pg_terminate_backend(?)"
	return cm.db.Exec(query, pid).Error
}
