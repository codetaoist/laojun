package query

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

// PrometheusQuerier Prometheus查询器
// 用于查询Prometheus服务器的指标数据
type PrometheusQuerier struct {
	client api.Client
	api    v1.API
	config PrometheusConfig
	logger *logrus.Logger
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	URL     string        `yaml:"url" json:"url"`
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	Auth    *AuthConfig   `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// NewPrometheusQuerier 创建Prometheus查询器
func NewPrometheusQuerier(config PrometheusConfig, logger *logrus.Logger) (*PrometheusQuerier, error) {
	clientConfig := api.Config{
		Address: config.URL,
	}

	// 设置超时
	if config.Timeout > 0 {
		clientConfig.RoundTripper = &http.Transport{
			ResponseHeaderTimeout: config.Timeout,
		}
	}

	// 设置认证
	if config.Auth != nil {
		switch config.Auth.Type {
		case "basic":
			if config.Auth.Username != "" && config.Auth.Password != "" {
				clientConfig.RoundTripper = &basicAuthRoundTripper{
					username: config.Auth.Username,
					password: config.Auth.Password,
					next:     clientConfig.RoundTripper,
				}
			}
		case "bearer":
			if config.Auth.Token != "" {
				clientConfig.RoundTripper = &bearerTokenRoundTripper{
					token: config.Auth.Token,
					next:  clientConfig.RoundTripper,
				}
			}
		}
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	return &PrometheusQuerier{
		client: client,
		api:    v1.NewAPI(client),
		config: config,
		logger: logger,
	}, nil
}

// Query 执行指标查询
func (pq *PrometheusQuerier) Query(ctx context.Context, query string) (*QueryResponse, error) {
	pq.logger.WithFields(logrus.Fields{
		"query": query,
	}).Debug("Executing Prometheus query")

	result, warnings, err := pq.api.Query(ctx, query, time.Now())
	if err != nil {
		pq.logger.WithFields(logrus.Fields{
			"query": query,
			"error": err,
		}).Error("Prometheus query failed")
		return &QueryResponse{
			Error: err.Error(),
		}, fmt.Errorf("prometheus query failed: %w", err)
	}

	if len(warnings) > 0 {
		pq.logger.WithFields(logrus.Fields{
			"query":    query,
			"warnings": warnings,
		}).Warn("Prometheus query returned warnings")
	}

	return pq.convertResult(result), nil
}

// QueryRange 执行范围查询
func (pq *PrometheusQuerier) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResponse, error) {
	pq.logger.WithFields(logrus.Fields{
		"query": query,
		"start": start,
		"end":   end,
		"step":  step,
	}).Debug("Executing Prometheus range query")

	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	result, warnings, err := pq.api.QueryRange(ctx, query, r)
	if err != nil {
		pq.logger.WithFields(logrus.Fields{
			"query": query,
			"error": err,
		}).Error("Prometheus range query failed")
		return nil, fmt.Errorf("prometheus range query failed: %w", err)
	}

	if len(warnings) > 0 {
		pq.logger.WithFields(logrus.Fields{
			"query":    query,
			"warnings": warnings,
		}).Warn("Prometheus range query returned warnings")
	}

	return pq.convertResult(result), nil
}

// IsHealthy 检查Prometheus连接是否健康
func (pq *PrometheusQuerier) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := pq.api.Query(ctx, "up", time.Now())
	return err == nil
}

// Close 关闭查询器
func (pq *PrometheusQuerier) Close() error {
	// Prometheus客户端不需要显式关闭
	return nil
}

// convertResult 转换Prometheus结果为通用格式
func (pq *PrometheusQuerier) convertResult(value model.Value) *QueryResponse {
	response := &QueryResponse{
		Results: []QueryResult{},
	}

	switch v := value.(type) {
	case model.Vector:
		for _, sample := range v {
			result := QueryResult{
				Value:     float64(sample.Value),
				Labels:    convertLabels(sample.Metric),
				Timestamp: sample.Timestamp.Time(),
			}
			response.Results = append(response.Results, result)
		}
	case model.Matrix:
		for _, sampleStream := range v {
			// 对于矩阵数据，我们取最后一个值
			if len(sampleStream.Values) > 0 {
				lastValue := sampleStream.Values[len(sampleStream.Values)-1]
				result := QueryResult{
					Value:     float64(lastValue.Value),
					Labels:    convertLabels(sampleStream.Metric),
					Timestamp: lastValue.Timestamp.Time(),
				}
				response.Results = append(response.Results, result)
			}
		}
	case *model.Scalar:
		result := QueryResult{
			Value:     float64(v.Value),
			Labels:    map[string]string{},
			Timestamp: v.Timestamp.Time(),
		}
		response.Results = append(response.Results, result)
	}

	return response
}

// convertLabels 转换Prometheus标签为map
func convertLabels(metric model.Metric) map[string]string {
	labels := make(map[string]string)
	for k, v := range metric {
		labels[string(k)] = string(v)
	}
	return labels
}

// basicAuthRoundTripper HTTP基础认证
type basicAuthRoundTripper struct {
	username string
	password string
	next     http.RoundTripper
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.username, rt.password)
	if rt.next == nil {
		rt.next = http.DefaultTransport
	}
	return rt.next.RoundTrip(req)
}

// bearerTokenRoundTripper Bearer Token认证
type bearerTokenRoundTripper struct {
	token string
	next  http.RoundTripper
}

func (rt *bearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+rt.token)
	if rt.next == nil {
		rt.next = http.DefaultTransport
	}
	return rt.next.RoundTrip(req)
}
