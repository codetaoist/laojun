package marketplace

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DownloadConfig 下载配置
type DownloadConfig struct {
	MaxConcurrentDownloads int           `json:"max_concurrent_downloads"`
	DownloadTimeout        time.Duration `json:"download_timeout"`
	RetryAttempts          int           `json:"retry_attempts"`
	RetryDelay             time.Duration `json:"retry_delay"`
	TempDir                string        `json:"temp_dir"`
	VerifyChecksum         bool          `json:"verify_checksum"`
	AllowedHosts           []string      `json:"allowed_hosts"`
	MaxFileSize            int64         `json:"max_file_size"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	PluginID      string    `json:"plugin_id"`
	Version       string    `json:"version"`
	Status        string    `json:"status"` // downloading, completed, failed, cancelled
	Progress      float64   `json:"progress"`
	BytesTotal    int64     `json:"bytes_total"`
	BytesReceived int64     `json:"bytes_received"`
	Speed         int64     `json:"speed"` // bytes per second
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Error         string    `json:"error,omitempty"`
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID          string            `json:"id"`
	PluginID    string            `json:"plugin_id"`
	Version     string            `json:"version"`
	URL         string            `json:"url"`
	Checksum    string            `json:"checksum"`
	Destination string            `json:"destination"`
	Progress    *DownloadProgress `json:"progress"`
	Cancel      chan bool         `json:"-"`
	Done        chan bool         `json:"-"`
}

// Downloader 插件下载器
type Downloader struct {
	config    DownloadConfig
	tasks     map[string]*DownloadTask
	semaphore chan struct{}
	mu        sync.RWMutex
	client    *http.Client
}

// NewDownloader 创建新的下载器
func NewDownloader(config DownloadConfig) *Downloader {
	if config.MaxConcurrentDownloads <= 0 {
		config.MaxConcurrentDownloads = 3
	}
	if config.DownloadTimeout <= 0 {
		config.DownloadTimeout = 30 * time.Minute
	}
	if config.RetryAttempts <= 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}
	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 1024 * 1024 * 1024 // 1GB
	}

	return &Downloader{
		config:    config,
		tasks:     make(map[string]*DownloadTask),
		semaphore: make(chan struct{}, config.MaxConcurrentDownloads),
		client: &http.Client{
			Timeout: config.DownloadTimeout,
		},
	}
}

// DownloadPlugin 下载插件
func (d *Downloader) DownloadPlugin(pluginID, version, url, checksum, destination string) (string, error) {
	// 验证URL
	if !d.isURLAllowed(url) {
		return "", errors.New("URL not allowed")
	}

	// 创建下载任务
	taskID := fmt.Sprintf("%s-%s-%d", pluginID, version, time.Now().Unix())
	task := &DownloadTask{
		ID:          taskID,
		PluginID:    pluginID,
		Version:     version,
		URL:         url,
		Checksum:    checksum,
		Destination: destination,
		Progress: &DownloadProgress{
			PluginID:  pluginID,
			Version:   version,
			Status:    "downloading",
			StartTime: time.Now(),
		},
		Cancel: make(chan bool, 1),
		Done:   make(chan bool, 1),
	}

	d.mu.Lock()
	d.tasks[taskID] = task
	d.mu.Unlock()

	// 异步下载
	go d.downloadTask(task)

	return taskID, nil
}

// GetDownloadProgress 获取下载进度
func (d *Downloader) GetDownloadProgress(taskID string) (*DownloadProgress, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	task, exists := d.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task.Progress, nil
}

// CancelDownload 取消下载
func (d *Downloader) CancelDownload(taskID string) error {
	d.mu.RLock()
	task, exists := d.tasks[taskID]
	d.mu.RUnlock()

	if !exists {
		return errors.New("task not found")
	}

	select {
	case task.Cancel <- true:
		return nil
	default:
		return errors.New("task already cancelled or completed")
	}
}

// ListDownloads 列出所有下载任务进度
func (d *Downloader) ListDownloads() []*DownloadProgress {
	d.mu.RLock()
	defer d.mu.RUnlock()

	progresses := make([]*DownloadProgress, 0, len(d.tasks))
	for _, task := range d.tasks {
		progresses = append(progresses, task.Progress)
	}

	return progresses
}

// CleanupCompletedTasks 清理已完成的任务
func (d *Downloader) CleanupCompletedTasks() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for taskID, task := range d.tasks {
		if task.Progress.Status == "completed" || task.Progress.Status == "failed" || task.Progress.Status == "cancelled" {
			delete(d.tasks, taskID)
		}
	}
}

// VerifyFile 验证文件
func (d *Downloader) VerifyFile(filePath, expectedChecksum string) error {
	if !d.config.VerifyChecksum || expectedChecksum == "" {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 根据校验和长度判断算法
	var hash string
	if len(expectedChecksum) == 32 {
		// MD5
		hasher := md5.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return fmt.Errorf("failed to calculate MD5: %v", err)
		}
		hash = hex.EncodeToString(hasher.Sum(nil))
	} else if len(expectedChecksum) == 64 {
		// SHA256
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return fmt.Errorf("failed to calculate SHA256: %v", err)
		}
		hash = hex.EncodeToString(hasher.Sum(nil))
	} else {
		return errors.New("unsupported checksum format")
	}

	if hash != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, hash)
	}

	return nil
}

// GetDownloadStats 获取下载统计
func (d *Downloader) GetDownloadStats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := map[string]interface{}{
		"total_tasks":   len(d.tasks),
		"downloading":   0,
		"completed":     0,
		"failed":        0,
		"cancelled":     0,
		"total_bytes":   int64(0),
		"average_speed": int64(0),
	}

	var totalSpeed int64
	var speedCount int

	for _, task := range d.tasks {
		switch task.Progress.Status {
		case "downloading":
			stats["downloading"] = stats["downloading"].(int) + 1
		case "completed":
			stats["completed"] = stats["completed"].(int) + 1
		case "failed":
			stats["failed"] = stats["failed"].(int) + 1
		case "cancelled":
			stats["cancelled"] = stats["cancelled"].(int) + 1
		}

		stats["total_bytes"] = stats["total_bytes"].(int64) + task.Progress.BytesTotal

		if task.Progress.Speed > 0 {
			totalSpeed += task.Progress.Speed
			speedCount++
		}
	}

	if speedCount > 0 {
		stats["average_speed"] = totalSpeed / int64(speedCount)
	}

	return stats
}

// 私有方法

// downloadTask 执行下载任务
func (d *Downloader) downloadTask(task *DownloadTask) {
	// 获取信号量，确保并发下载不超过最大并发数
	d.semaphore <- struct{}{}
	defer func() { <-d.semaphore }()

	var err error
	for attempt := 0; attempt < d.config.RetryAttempts; attempt++ {
		select {
		case <-task.Cancel:
			task.Progress.Status = "cancelled"
			task.Progress.EndTime = time.Now()
			task.Done <- true
			return
		default:
		}

		err = d.performDownload(task)
		if err == nil {
			break
		}

		if attempt < d.config.RetryAttempts-1 {
			time.Sleep(d.config.RetryDelay)
		}
	}

	if err != nil {
		task.Progress.Status = "failed"
		task.Progress.Error = err.Error()
	} else {
		task.Progress.Status = "completed"
		task.Progress.Progress = 100.0
	}

	task.Progress.EndTime = time.Now()
	task.Done <- true
}

// performDownload 执行实际下载
func (d *Downloader) performDownload(task *DownloadTask) error {
	// 创建临时文件
	tempFile := filepath.Join(d.config.TempDir, fmt.Sprintf("%s.tmp", task.ID))
	defer os.Remove(tempFile)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", task.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 发送请求
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 检查文件大小是否超过最大允许值
	if resp.ContentLength > d.config.MaxFileSize {
		return fmt.Errorf("file too large: %d bytes", resp.ContentLength)
	}

	task.Progress.BytesTotal = resp.ContentLength

	// 创建临时文件
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	// 下载文件并更新进度
	err = d.downloadWithProgress(resp.Body, file, task)
	if err != nil {
		return err
	}

	// 验证校验和
	if err := d.VerifyFile(tempFile, task.Checksum); err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(task.Destination), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// 移动文件到最终位置
	if err := os.Rename(tempFile, task.Destination); err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}

	return nil
}

// downloadWithProgress 带进度的下载
func (d *Downloader) downloadWithProgress(src io.Reader, dst io.Writer, task *DownloadTask) error {
	buffer := make([]byte, 32*1024) // 32KB buffer
	var written int64
	lastUpdate := time.Now()

	for {
		select {
		case <-task.Cancel:
			return errors.New("download cancelled")
		default:
		}

		n, err := src.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			written += int64(n)
			task.Progress.BytesReceived = written

			// 更新进度
			now := time.Now()
			if now.Sub(lastUpdate) >= time.Second {
				if task.Progress.BytesTotal > 0 {
					task.Progress.Progress = float64(written) / float64(task.Progress.BytesTotal) * 100
				}

				// 计算速度
				duration := now.Sub(task.Progress.StartTime).Seconds()
				if duration > 0 {
					task.Progress.Speed = int64(float64(written) / duration)
				}

				lastUpdate = now
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// isURLAllowed 检查URL是否被允许下载
func (d *Downloader) isURLAllowed(url string) bool {
	if len(d.config.AllowedHosts) == 0 {
		return true
	}

	for _, host := range d.config.AllowedHosts {
		if strings.Contains(url, host) {
			return true
		}
	}

	return false
}

// BatchDownload 批量下载
func (d *Downloader) BatchDownload(downloads []struct {
	PluginID    string
	Version     string
	URL         string
	Checksum    string
	Destination string
}) ([]string, error) {
	taskIDs := make([]string, 0, len(downloads))

	for _, download := range downloads {
		taskID, err := d.DownloadPlugin(
			download.PluginID,
			download.Version,
			download.URL,
			download.Checksum,
			download.Destination,
		)
		if err != nil {
			return taskIDs, fmt.Errorf("failed to start download for %s: %v", download.PluginID, err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

// WaitForDownload 等待下载完成
func (d *Downloader) WaitForDownload(taskID string) error {
	d.mu.RLock()
	task, exists := d.tasks[taskID]
	d.mu.RUnlock()

	if !exists {
		return errors.New("task not found")
	}

	<-task.Done

	if task.Progress.Status == "failed" {
		return errors.New(task.Progress.Error)
	}

	return nil
}

// PauseDownload 暂停下载
func (d *Downloader) PauseDownload(taskID string) error {
	// 简单实现：取消当前下载
	return d.CancelDownload(taskID)
}

// ResumeDownload 恢复下载
func (d *Downloader) ResumeDownload(taskID string) error {
	d.mu.RLock()
	task, exists := d.tasks[taskID]
	d.mu.RUnlock()

	if !exists {
		return errors.New("task not found")
	}

	if task.Progress.Status != "cancelled" {
		return errors.New("task is not paused")
	}

	// 重新开始下载
	newTaskID, err := d.DownloadPlugin(
		task.PluginID,
		task.Version,
		task.URL,
		task.Checksum,
		task.Destination,
	)
	if err != nil {
		return err
	}

	// 更新任务ID
	d.mu.Lock()
	delete(d.tasks, taskID)
	d.mu.Unlock()

	_ = newTaskID // 可以返回新的任务ID
	return nil
}

// GetActiveDownloads 获取活跃的下载任务
func (d *Downloader) GetActiveDownloads() []*DownloadProgress {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var active []*DownloadProgress
	for _, task := range d.tasks {
		if task.Progress.Status == "downloading" {
			active = append(active, task.Progress)
		}
	}

	return active
}

// SetDownloadLimit 设置下载限制
func (d *Downloader) SetDownloadLimit(limit int) {
	if limit <= 0 {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// 创建新的信号量通道
	newSemaphore := make(chan struct{}, limit)

	// 迁移现有的信号量
	for i := 0; i < len(d.semaphore) && i < limit; i++ {
		newSemaphore <- struct{}{}
	}

	d.semaphore = newSemaphore
	d.config.MaxConcurrentDownloads = limit
}
