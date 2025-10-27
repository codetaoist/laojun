package query

import (
	"fmt"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/config"
	"github.com/sirupsen/logrus"
)

// QuerierFactory 查询器工厂
type QuerierFactory struct {
	logger *logrus.Logger
}

// NewQuerierFactory 创建查询器工厂
func NewQuerierFactory(logger *logrus.Logger) *QuerierFactory {
	return &QuerierFactory{
		logger: logger,
	}
}

// CreateQuerier 根据配置创建查询器
func (f *QuerierFactory) CreateQuerier(queryConfig config.QueryConfig) (MetricQuerier, error) {
	switch queryConfig.Type {
	case "prometheus":
		return f.createPrometheusQuerier(queryConfig.Prometheus)
	case "memory":
		return f.createMemoryQuerier()
	default:
		f.logger.WithField("type", queryConfig.Type).Warn("Unknown querier type, falling back to memory querier")
		return f.createMemoryQuerier()
	}
}

// createPrometheusQuerier 创建Prometheus查询器
func (f *QuerierFactory) createPrometheusQuerier(config config.PrometheusQueryConfig) (MetricQuerier, error) {
	prometheusConfig := PrometheusConfig{
		URL:     config.URL,
		Timeout: config.Timeout,
	}

	// 转换认证配置
	if config.Auth != nil {
		prometheusConfig.Auth = &AuthConfig{
			Type:     config.Auth.Type,
			Username: config.Auth.Username,
			Password: config.Auth.Password,
			Token:    config.Auth.Token,
		}
	}

	// 设置默认超时
	if prometheusConfig.Timeout == 0 {
		prometheusConfig.Timeout = 30 * time.Second
	}

	f.logger.WithFields(logrus.Fields{
		"url":     prometheusConfig.URL,
		"timeout": prometheusConfig.Timeout,
	}).Info("Creating Prometheus querier")

	return NewPrometheusQuerier(prometheusConfig, f.logger)
}

// createMemoryQuerier 创建内存查询器
func (f *QuerierFactory) createMemoryQuerier() (MetricQuerier, error) {
	f.logger.Info("Creating memory querier")
	return NewMemoryQuerier(f.logger), nil
}

// ValidateConfig 验证查询配置
func ValidateQueryConfig(config config.QueryConfig) error {
	switch config.Type {
	case "prometheus":
		if config.Prometheus.URL == "" {
			return fmt.Errorf("prometheus URL is required")
		}
	case "memory":
		// 内存查询器不需要额外配置
	case "":
		return fmt.Errorf("query type is required")
	default:
		return fmt.Errorf("unsupported query type: %s", config.Type)
	}
	return nil
}