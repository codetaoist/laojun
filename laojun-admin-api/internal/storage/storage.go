package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Storage struct {
	basePath string
}

func NewStorage(basePath string) *Storage {
	return &Storage{
		basePath: basePath,
	}
}

func (s *Storage) Save(filename string, data io.Reader) error {
	fullPath := filepath.Join(s.basePath, filename)
	
	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	return err
}

func (s *Storage) Load(filename string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, filename)
	return os.Open(fullPath)
}

func (s *Storage) Delete(filename string) error {
	fullPath := filepath.Join(s.basePath, filename)
	return os.Remove(fullPath)
}

func (s *Storage) Exists(filename string) bool {
	fullPath := filepath.Join(s.basePath, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// ConfigStorage 配置存储接口
type ConfigStorage interface {
	Get(ctx interface{}, service, environment, key string) (*ConfigItem, error)
	Set(ctx interface{}, item *ConfigItem) error
	Delete(ctx interface{}, service, environment, key string) error
	List(ctx interface{}, service, environment string) ([]*ConfigItem, error)
	GetByTags(ctx interface{}, tags map[string]string) ([]*ConfigItem, error)
	GetHistory(ctx interface{}, service, environment, key string, limit int) ([]*ConfigItem, error)
	Backup(ctx interface{}, service, environment string) ([]byte, error)
	Restore(ctx interface{}, service, environment string, data []byte) error
	Watch(ctx interface{}, service, environment string) (<-chan *ConfigItem, error)
}

// ConfigItem 配置项
type ConfigItem struct {
	Key         string            `json:"key"`
	Value       interface{}       `json:"value"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Service     string            `json:"service"`
	Environment string            `json:"environment"`
	Tags        map[string]string `json:"tags"`
	CreatedBy   string            `json:"created_by"`
	UpdatedBy   string            `json:"updated_by"`
}