package marketplace

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Storage 存储接口
type Storage interface {
	Store(filename string, data []byte) error
	Retrieve(filename string) ([]byte, error)
	Delete(filename string) error
	Exists(filename string) bool
	List() ([]string, error)
	GetStats() map[string]interface{}
	Cleanup(maxAge time.Duration) error
}

// FileStorage 文件存储实现
type FileStorage struct {
	basePath string
	mu       sync.RWMutex
	stats    StorageStats
}

// StorageStats 存储统计信息
type StorageStats struct {
	TotalFiles   int64     `json:"total_files"`
	TotalSize    int64     `json:"total_size"`
	LastAccess   time.Time `json:"last_access"`
	LastModified time.Time `json:"last_modified"`
	Operations   struct {
		Stores    int64 `json:"stores"`
		Retrieves int64 `json:"retrieves"`
		Deletes   int64 `json:"deletes"`
		Errors    int64 `json:"errors"`
	} `json:"operations"`
}

// NewStorage 创建新的存储实例
func NewStorage(basePath string) (Storage, error) {
	// 确保基础路径存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	storage := &FileStorage{
		basePath: basePath,
		stats:    StorageStats{},
	}

	// 初始化统计信息
	if err := storage.updateStats(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage stats: %v", err)
	}

	return storage, nil
}

// Store 存储文件
func (fs *FileStorage) Store(filename string, data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 验证文件名是否安全
	if err := fs.validateFilename(filename); err != nil {
		fs.stats.Operations.Errors++
		return err
	}

	// 创建完整路径
	fullPath := filepath.Join(fs.basePath, filename)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fs.stats.Operations.Errors++
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 创建临时文件
	tempPath := fullPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		fs.stats.Operations.Errors++
		return fmt.Errorf("failed to create temp file: %v", err)
	}

	// 写入数据
	_, err = file.Write(data)
	file.Close()

	if err != nil {
		os.Remove(tempPath)
		fs.stats.Operations.Errors++
		return fmt.Errorf("failed to write data: %v", err)
	}

	// 原子性重命名
	if err := os.Rename(tempPath, fullPath); err != nil {
		os.Remove(tempPath)
		fs.stats.Operations.Errors++
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	// 更新统计信息
	fs.stats.Operations.Stores++
	fs.stats.TotalSize += int64(len(data))
	fs.stats.LastModified = time.Now()

	return nil
}

// Retrieve 检索文件内容
func (fs *FileStorage) Retrieve(filename string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// 验证文件名是否安全
	if err := fs.validateFilename(filename); err != nil {
		fs.stats.Operations.Errors++
		return nil, err
	}

	fullPath := filepath.Join(fs.basePath, filename)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}

	// 读取文件
	data, err := os.ReadFile(fullPath)
	if err != nil {
		fs.stats.Operations.Errors++
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// 更新统计信息
	fs.stats.Operations.Retrieves++
	fs.stats.LastAccess = time.Now()

	return data, nil
}

// Delete 删除文件
func (fs *FileStorage) Delete(filename string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 验证文件名是否安全
	if err := fs.validateFilename(filename); err != nil {
		fs.stats.Operations.Errors++
		return err
	}

	fullPath := filepath.Join(fs.basePath, filename)

	// 获取文件信息
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}
	if err != nil {
		fs.stats.Operations.Errors++
		return fmt.Errorf("failed to stat file: %v", err)
	}

	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		fs.stats.Operations.Errors++
		return fmt.Errorf("failed to delete file: %v", err)
	}

	// 更新统计信息
	fs.stats.Operations.Deletes++
	fs.stats.TotalSize -= info.Size()
	fs.stats.TotalFiles--

	return nil
}

// Exists 检查文件是否存在
func (fs *FileStorage) Exists(filename string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if err := fs.validateFilename(filename); err != nil {
		return false
	}

	fullPath := filepath.Join(fs.basePath, filename)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

// List 列出所有文件
func (fs *FileStorage) List() ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var files []string

	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// 获取相对路径
			relPath, err := filepath.Rel(fs.basePath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	return files, nil
}

// GetStats 获取存储统计信息
func (fs *FileStorage) GetStats() map[string]interface{} {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return map[string]interface{}{
		"total_files":   fs.stats.TotalFiles,
		"total_size":    fs.stats.TotalSize,
		"last_access":   fs.stats.LastAccess,
		"last_modified": fs.stats.LastModified,
		"operations":    fs.stats.Operations,
		"base_path":     fs.basePath,
	}
}

// Cleanup 清理过期文件
func (fs *FileStorage) Cleanup(maxAge time.Duration) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var deletedCount int64
	var deletedSize int64

	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				return err
			}
			deletedCount++
			deletedSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cleanup failed: %v", err)
	}

	// 更新统计信息
	fs.stats.TotalFiles -= deletedCount
	fs.stats.TotalSize -= deletedSize

	return nil
}

// validateFilename 验证文件名是否安全
func (fs *FileStorage) validateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if filepath.IsAbs(filename) {
		return fmt.Errorf("filename cannot be absolute path")
	}

	// 检查路径遍历攻击
	cleanPath := filepath.Clean(filename)
	if cleanPath != filename || filepath.IsAbs(cleanPath) {
		return fmt.Errorf("invalid filename: %s", filename)
	}

	return nil
}

// updateStats 更新统计信息
func (fs *FileStorage) updateStats() error {
	var totalFiles int64
	var totalSize int64

	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return err
	}

	fs.stats.TotalFiles = totalFiles
	fs.stats.TotalSize = totalSize

	return nil
}

// S3Storage S3存储实现（示例）
type S3Storage struct {
	bucket    string
	region    string
	accessKey string
	secretKey string
	mu        sync.RWMutex
	stats     StorageStats
}

// NewS3Storage 创建S3存储实例
func NewS3Storage(bucket, region, accessKey, secretKey string) Storage {
	return &S3Storage{
		bucket:    bucket,
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
		stats:     StorageStats{},
	}
}

// Store S3存储实现
func (s3 *S3Storage) Store(filename string, data []byte) error {
	// 这里应该实现S3上传逻辑
	// 为了示例，返回未实现错误
	return fmt.Errorf("S3 storage not implemented")
}

// Retrieve S3检索实现
func (s3 *S3Storage) Retrieve(filename string) ([]byte, error) {
	// 这里应该实现S3下载逻辑
	return nil, fmt.Errorf("S3 storage not implemented")
}

// Delete S3删除实现
func (s3 *S3Storage) Delete(filename string) error {
	// 这里应该实现S3删除逻辑
	return fmt.Errorf("S3 storage not implemented")
}

// Exists S3存在检查实现
func (s3 *S3Storage) Exists(filename string) bool {
	// 这里应该实现S3存在检查逻辑
	return false
}

// List S3列表实现
func (s3 *S3Storage) List() ([]string, error) {
	// 这里应该实现S3列表逻辑
	return nil, fmt.Errorf("S3 storage not implemented")
}

// GetStats S3统计信息实现
func (s3 *S3Storage) GetStats() map[string]interface{} {
	s3.mu.RLock()
	defer s3.mu.RUnlock()

	return map[string]interface{}{
		"type":          "s3",
		"bucket":        s3.bucket,
		"region":        s3.region,
		"total_files":   s3.stats.TotalFiles,
		"total_size":    s3.stats.TotalSize,
		"last_access":   s3.stats.LastAccess,
		"last_modified": s3.stats.LastModified,
		"operations":    s3.stats.Operations,
	}
}

// Cleanup S3清理实现
func (s3 *S3Storage) Cleanup(maxAge time.Duration) error {
	// 这里应该实现S3清理逻辑
	return fmt.Errorf("S3 storage not implemented")
}

// CachedStorage 带缓存的存储实现
type CachedStorage struct {
	backend Storage
	cache   map[string]CacheEntry
	maxSize int
	mu      sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data        []byte
	Timestamp   time.Time
	AccessCount int64
}

// NewCachedStorage 创建带缓存的存储
func NewCachedStorage(backend Storage, maxSize int) Storage {
	return &CachedStorage{
		backend: backend,
		cache:   make(map[string]CacheEntry),
		maxSize: maxSize,
	}
}

// Store 缓存存储实现
func (cs *CachedStorage) Store(filename string, data []byte) error {
	// 存储到后端
	if err := cs.backend.Store(filename, data); err != nil {
		return err
	}

	// 更新缓存
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.cache[filename] = CacheEntry{
		Data:        data,
		Timestamp:   time.Now(),
		AccessCount: 0,
	}

	// 检查缓存大小是否超过最大缓存大小
	cs.evictIfNeeded()

	return nil
}

// Retrieve 缓存检索实现
func (cs *CachedStorage) Retrieve(filename string) ([]byte, error) {
	// 先检查缓存
	cs.mu.Lock()
	if entry, exists := cs.cache[filename]; exists {
		entry.AccessCount++
		cs.cache[filename] = entry
		cs.mu.Unlock()
		return entry.Data, nil
	}
	cs.mu.Unlock()

	// 从后端获取数据
	data, err := cs.backend.Retrieve(filename)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	cs.mu.Lock()
	cs.cache[filename] = CacheEntry{
		Data:        data,
		Timestamp:   time.Now(),
		AccessCount: 1,
	}
	cs.evictIfNeeded()
	cs.mu.Unlock()

	return data, nil
}

// Delete 缓存删除实现
func (cs *CachedStorage) Delete(filename string) error {
	// 从后端删除
	if err := cs.backend.Delete(filename); err != nil {
		return err
	}

	// 从缓存删除
	cs.mu.Lock()
	delete(cs.cache, filename)
	cs.mu.Unlock()

	return nil
}

// Exists 缓存存在检查实现
func (cs *CachedStorage) Exists(filename string) bool {
	// 先检查缓存
	cs.mu.RLock()
	_, exists := cs.cache[filename]
	cs.mu.RUnlock()

	if exists {
		return true
	}

	// 检查后端是否存在
	return cs.backend.Exists(filename)
}

// List 缓存列表实现
func (cs *CachedStorage) List() ([]string, error) {
	return cs.backend.List()
}

// GetStats 缓存统计信息实现
func (cs *CachedStorage) GetStats() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	backendStats := cs.backend.GetStats()

	cacheSize := 0
	for _, entry := range cs.cache {
		cacheSize += len(entry.Data)
	}

	return map[string]interface{}{
		"backend": backendStats,
		"cache": map[string]interface{}{
			"entries":    len(cs.cache),
			"size_bytes": cacheSize,
			"max_size":   cs.maxSize,
		},
	}
}

// Cleanup 缓存清理实现
func (cs *CachedStorage) Cleanup(maxAge time.Duration) error {
	// 清理后端
	if err := cs.backend.Cleanup(maxAge); err != nil {
		return err
	}

	// 清理缓存
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for filename, entry := range cs.cache {
		if entry.Timestamp.Before(cutoff) {
			delete(cs.cache, filename)
		}
	}

	return nil
}

// evictIfNeeded 根据需要驱逐缓存条目
func (cs *CachedStorage) evictIfNeeded() {
	if len(cs.cache) <= cs.maxSize {
		return
	}

	// 找到最少使用的条目
	var oldestKey string
	var oldestTime time.Time
	var minAccess int64 = -1

	for key, entry := range cs.cache {
		if minAccess == -1 || entry.AccessCount < minAccess {
			minAccess = entry.AccessCount
			oldestKey = key
			oldestTime = entry.Timestamp
		} else if entry.AccessCount == minAccess && entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(cs.cache, oldestKey)
	}
}

// CalculateChecksum 计算文件校验和
func CalculateChecksum(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// CalculateFileChecksum 计算文件校验和
func CalculateFileChecksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
