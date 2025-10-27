package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codetaoist/laojun-marketplace-api/internal/events"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DataSyncManager 数据同步管理器
type DataSyncManager struct {
	db           *gorm.DB
	redisClient  *redis.Client
	eventManager *events.EventManager
	logger       *logrus.Logger
	syncInterval time.Duration
	stopCh       chan struct{}
}

// SyncConfig 同步配置
type SyncConfig struct {
	Enabled      bool          `json:"enabled"`
	Interval     time.Duration `json:"interval"`
	BatchSize    int           `json:"batch_size"`
	RetryCount   int           `json:"retry_count"`
	RetryDelay   time.Duration `json:"retry_delay"`
	EnabledTypes []string      `json:"enabled_types"`
}

// SyncRecord 同步记录
type SyncRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"index" json:"type"`
	SourceID    string    `gorm:"index" json:"source_id"`
	TargetID    string    `json:"target_id"`
	Status      string    `gorm:"index" json:"status"` // pending, syncing, success, failed
	Data        string    `json:"data"`                // JSON格式的数据
	Error       string    `json:"error"`
	RetryCount  int       `json:"retry_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	SyncedAt    *time.Time `json:"synced_at"`
}

// NewDataSyncManager 创建数据同步管理器
func NewDataSyncManager(
	db *gorm.DB,
	redisClient *redis.Client,
	eventManager *events.EventManager,
	logger *logrus.Logger,
	syncInterval time.Duration,
) *DataSyncManager {
	return &DataSyncManager{
		db:           db,
		redisClient:  redisClient,
		eventManager: eventManager,
		logger:       logger,
		syncInterval: syncInterval,
		stopCh:       make(chan struct{}),
	}
}

// Start 启动数据同步管理器
func (dsm *DataSyncManager) Start(ctx context.Context) error {
	dsm.logger.Info("Starting data sync manager")

	// 创建同步记录表
	if err := dsm.db.AutoMigrate(&SyncRecord{}); err != nil {
		return fmt.Errorf("failed to migrate sync record table: %w", err)
	}

	// 启动同步任务
	go dsm.syncWorker(ctx)

	// 启动清理任务
	go dsm.cleanupWorker(ctx)

	dsm.logger.Info("Data sync manager started successfully")
	return nil
}

// Stop 停止数据同步管理器
func (dsm *DataSyncManager) Stop(ctx context.Context) error {
	dsm.logger.Info("Stopping data sync manager")
	close(dsm.stopCh)
	return nil
}

// SyncUserData 同步用户数据
func (dsm *DataSyncManager) SyncUserData(ctx context.Context, userID string, userData map[string]interface{}) error {
	return dsm.createSyncRecord(ctx, "user", userID, userData)
}

// SyncPluginData 同步插件数据
func (dsm *DataSyncManager) SyncPluginData(ctx context.Context, pluginID string, pluginData map[string]interface{}) error {
	return dsm.createSyncRecord(ctx, "plugin", pluginID, pluginData)
}

// SyncPaymentData 同步支付数据
func (dsm *DataSyncManager) SyncPaymentData(ctx context.Context, orderID string, paymentData map[string]interface{}) error {
	return dsm.createSyncRecord(ctx, "payment", orderID, paymentData)
}

// SyncReviewData 同步评价数据
func (dsm *DataSyncManager) SyncReviewData(ctx context.Context, reviewID string, reviewData map[string]interface{}) error {
	return dsm.createSyncRecord(ctx, "review", reviewID, reviewData)
}

// SyncDeveloperData 同步开发者数据
func (dsm *DataSyncManager) SyncDeveloperData(ctx context.Context, developerID string, developerData map[string]interface{}) error {
	return dsm.createSyncRecord(ctx, "developer", developerID, developerData)
}

// createSyncRecord 创建同步记录
func (dsm *DataSyncManager) createSyncRecord(ctx context.Context, syncType, sourceID string, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal sync data: %w", err)
	}

	record := &SyncRecord{
		Type:      syncType,
		SourceID:  sourceID,
		Status:    "pending",
		Data:      string(dataJSON),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := dsm.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}

	dsm.logger.Infof("Created sync record: type=%s, source_id=%s, record_id=%d", syncType, sourceID, record.ID)
	return nil
}

// syncWorker 同步工作器
func (dsm *DataSyncManager) syncWorker(ctx context.Context) {
	ticker := time.NewTicker(dsm.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			dsm.logger.Info("Sync worker stopped due to context cancellation")
			return
		case <-dsm.stopCh:
			dsm.logger.Info("Sync worker stopped")
			return
		case <-ticker.C:
			if err := dsm.processPendingSyncRecords(ctx); err != nil {
				dsm.logger.Errorf("Failed to process pending sync records: %v", err)
			}
		}
	}
}

// processPendingSyncRecords 处理待同步记录
func (dsm *DataSyncManager) processPendingSyncRecords(ctx context.Context) error {
	var records []SyncRecord
	if err := dsm.db.WithContext(ctx).
		Where("status IN ?", []string{"pending", "failed"}).
		Where("retry_count < ?", 3).
		Order("created_at ASC").
		Limit(100).
		Find(&records).Error; err != nil {
		return fmt.Errorf("failed to fetch pending sync records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	dsm.logger.Infof("Processing %d pending sync records", len(records))

	for _, record := range records {
		if err := dsm.processSyncRecord(ctx, &record); err != nil {
			dsm.logger.Errorf("Failed to process sync record %d: %v", record.ID, err)
			dsm.updateSyncRecordStatus(ctx, record.ID, "failed", err.Error(), record.RetryCount+1)
		} else {
			now := time.Now()
			dsm.updateSyncRecordStatus(ctx, record.ID, "success", "", record.RetryCount)
			dsm.db.WithContext(ctx).Model(&SyncRecord{}).Where("id = ?", record.ID).Update("synced_at", &now)
		}
	}

	return nil
}

// processSyncRecord 处理单个同步记录
func (dsm *DataSyncManager) processSyncRecord(ctx context.Context, record *SyncRecord) error {
	dsm.logger.Infof("Processing sync record: id=%d, type=%s, source_id=%s", record.ID, record.Type, record.SourceID)

	// 更新状态为同步中
	dsm.updateSyncRecordStatus(ctx, record.ID, "syncing", "", record.RetryCount)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(record.Data), &data); err != nil {
		return fmt.Errorf("failed to unmarshal sync data: %w", err)
	}

	switch record.Type {
	case "user":
		return dsm.syncUserToExternalSystems(ctx, record.SourceID, data)
	case "plugin":
		return dsm.syncPluginToExternalSystems(ctx, record.SourceID, data)
	case "payment":
		return dsm.syncPaymentToExternalSystems(ctx, record.SourceID, data)
	case "review":
		return dsm.syncReviewToExternalSystems(ctx, record.SourceID, data)
	case "developer":
		return dsm.syncDeveloperToExternalSystems(ctx, record.SourceID, data)
	default:
		return fmt.Errorf("unknown sync type: %s", record.Type)
	}
}

// syncUserToExternalSystems 同步用户数据到外部系统
func (dsm *DataSyncManager) syncUserToExternalSystems(ctx context.Context, userID string, data map[string]interface{}) error {
	dsm.logger.Infof("Syncing user %s to external systems", userID)

	// 同步到插件系统
	if err := dsm.syncToPluginSystem(ctx, "user", userID, data); err != nil {
		dsm.logger.Errorf("Failed to sync user to plugin system: %v", err)
	}

	// 同步到监控系统
	if err := dsm.syncToMonitoringSystem(ctx, "user", userID, data); err != nil {
		dsm.logger.Errorf("Failed to sync user to monitoring system: %v", err)
	}

	// 同步到管理后台
	if err := dsm.syncToAdminSystem(ctx, "user", userID, data); err != nil {
		dsm.logger.Errorf("Failed to sync user to admin system: %v", err)
	}

	return nil
}

// syncPluginToExternalSystems 同步插件数据到外部系统
func (dsm *DataSyncManager) syncPluginToExternalSystems(ctx context.Context, pluginID string, data map[string]interface{}) error {
	dsm.logger.Infof("Syncing plugin %s to external systems", pluginID)

	// 同步到插件运行时系统
	if err := dsm.syncToPluginRuntime(ctx, pluginID, data); err != nil {
		dsm.logger.Errorf("Failed to sync plugin to runtime system: %v", err)
	}

	// 同步到监控系统
	if err := dsm.syncToMonitoringSystem(ctx, "plugin", pluginID, data); err != nil {
		dsm.logger.Errorf("Failed to sync plugin to monitoring system: %v", err)
	}

	// 同步到管理后台
	if err := dsm.syncToAdminSystem(ctx, "plugin", pluginID, data); err != nil {
		dsm.logger.Errorf("Failed to sync plugin to admin system: %v", err)
	}

	return nil
}

// syncPaymentToExternalSystems 同步支付数据到外部系统
func (dsm *DataSyncManager) syncPaymentToExternalSystems(ctx context.Context, orderID string, data map[string]interface{}) error {
	dsm.logger.Infof("Syncing payment %s to external systems", orderID)

	// 同步到财务系统
	if err := dsm.syncToFinancialSystem(ctx, orderID, data); err != nil {
		dsm.logger.Errorf("Failed to sync payment to financial system: %v", err)
	}

	// 同步到监控系统
	if err := dsm.syncToMonitoringSystem(ctx, "payment", orderID, data); err != nil {
		dsm.logger.Errorf("Failed to sync payment to monitoring system: %v", err)
	}

	// 同步到管理后台
	if err := dsm.syncToAdminSystem(ctx, "payment", orderID, data); err != nil {
		dsm.logger.Errorf("Failed to sync payment to admin system: %v", err)
	}

	return nil
}

// syncReviewToExternalSystems 同步评价数据到外部系统
func (dsm *DataSyncManager) syncReviewToExternalSystems(ctx context.Context, reviewID string, data map[string]interface{}) error {
	dsm.logger.Infof("Syncing review %s to external systems", reviewID)

	// 同步到搜索系统
	if err := dsm.syncToSearchSystem(ctx, reviewID, data); err != nil {
		dsm.logger.Errorf("Failed to sync review to search system: %v", err)
	}

	// 同步到监控系统
	if err := dsm.syncToMonitoringSystem(ctx, "review", reviewID, data); err != nil {
		dsm.logger.Errorf("Failed to sync review to monitoring system: %v", err)
	}

	// 同步到管理后台
	if err := dsm.syncToAdminSystem(ctx, "review", reviewID, data); err != nil {
		dsm.logger.Errorf("Failed to sync review to admin system: %v", err)
	}

	return nil
}

// syncDeveloperToExternalSystems 同步开发者数据到外部系统
func (dsm *DataSyncManager) syncDeveloperToExternalSystems(ctx context.Context, developerID string, data map[string]interface{}) error {
	dsm.logger.Infof("Syncing developer %s to external systems", developerID)

	// 同步到插件系统
	if err := dsm.syncToPluginSystem(ctx, "developer", developerID, data); err != nil {
		dsm.logger.Errorf("Failed to sync developer to plugin system: %v", err)
	}

	// 同步到监控系统
	if err := dsm.syncToMonitoringSystem(ctx, "developer", developerID, data); err != nil {
		dsm.logger.Errorf("Failed to sync developer to monitoring system: %v", err)
	}

	// 同步到管理后台
	if err := dsm.syncToAdminSystem(ctx, "developer", developerID, data); err != nil {
		dsm.logger.Errorf("Failed to sync developer to admin system: %v", err)
	}

	return nil
}

// 以下是具体的同步实现方法（占位符实现）

func (dsm *DataSyncManager) syncToPluginSystem(ctx context.Context, dataType, id string, data map[string]interface{}) error {
	dsm.logger.Infof("Should sync %s %s to plugin system", dataType, id)
	// TODO: 实现插件系统同步逻辑
	return nil
}

func (dsm *DataSyncManager) syncToPluginRuntime(ctx context.Context, pluginID string, data map[string]interface{}) error {
	dsm.logger.Infof("Should sync plugin %s to runtime system", pluginID)
	// TODO: 实现插件运行时同步逻辑
	return nil
}

func (dsm *DataSyncManager) syncToMonitoringSystem(ctx context.Context, dataType, id string, data map[string]interface{}) error {
	dsm.logger.Infof("Should sync %s %s to monitoring system", dataType, id)
	// TODO: 实现监控系统同步逻辑
	return nil
}

func (dsm *DataSyncManager) syncToAdminSystem(ctx context.Context, dataType, id string, data map[string]interface{}) error {
	dsm.logger.Infof("Should sync %s %s to admin system", dataType, id)
	// TODO: 实现管理后台同步逻辑
	return nil
}

func (dsm *DataSyncManager) syncToFinancialSystem(ctx context.Context, orderID string, data map[string]interface{}) error {
	dsm.logger.Infof("Should sync payment %s to financial system", orderID)
	// TODO: 实现财务系统同步逻辑
	return nil
}

func (dsm *DataSyncManager) syncToSearchSystem(ctx context.Context, reviewID string, data map[string]interface{}) error {
	dsm.logger.Infof("Should sync review %s to search system", reviewID)
	// TODO: 实现搜索系统同步逻辑
	return nil
}

// updateSyncRecordStatus 更新同步记录状态
func (dsm *DataSyncManager) updateSyncRecordStatus(ctx context.Context, recordID uint, status, errorMsg string, retryCount int) {
	updates := map[string]interface{}{
		"status":      status,
		"error":       errorMsg,
		"retry_count": retryCount,
		"updated_at":  time.Now(),
	}

	if err := dsm.db.WithContext(ctx).Model(&SyncRecord{}).Where("id = ?", recordID).Updates(updates).Error; err != nil {
		dsm.logger.Errorf("Failed to update sync record status: %v", err)
	}
}

// cleanupWorker 清理工作器
func (dsm *DataSyncManager) cleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // 每天清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			dsm.logger.Info("Cleanup worker stopped due to context cancellation")
			return
		case <-dsm.stopCh:
			dsm.logger.Info("Cleanup worker stopped")
			return
		case <-ticker.C:
			dsm.cleanupOldRecords(ctx)
		}
	}
}

// cleanupOldRecords 清理旧记录
func (dsm *DataSyncManager) cleanupOldRecords(ctx context.Context) {
	// 删除30天前的成功记录
	cutoffTime := time.Now().AddDate(0, 0, -30)
	result := dsm.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", "success", cutoffTime).
		Delete(&SyncRecord{})

	if result.Error != nil {
		dsm.logger.Errorf("Failed to cleanup old sync records: %v", result.Error)
	} else if result.RowsAffected > 0 {
		dsm.logger.Infof("Cleaned up %d old sync records", result.RowsAffected)
	}
}

// GetSyncStats 获取同步统计信息
func (dsm *DataSyncManager) GetSyncStats(ctx context.Context) (map[string]interface{}, error) {
	var stats []struct {
		Type   string `json:"type"`
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	if err := dsm.db.WithContext(ctx).
		Model(&SyncRecord{}).
		Select("type, status, count(*) as count").
		Group("type, status").
		Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get sync stats: %w", err)
	}

	result := make(map[string]interface{})
	for _, stat := range stats {
		if result[stat.Type] == nil {
			result[stat.Type] = make(map[string]int64)
		}
		result[stat.Type].(map[string]int64)[stat.Status] = stat.Count
	}

	return result, nil
}