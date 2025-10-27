package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PrometheusStorage Prometheus存储实现
type PrometheusStorage struct {
	client   api.Client
	queryAPI v1.API
	config   PrometheusStorageConfig
}

// PrometheusStorageConfig Prometheus存储配置
type PrometheusStorageConfig struct {
	URL      string        `mapstructure:"url"`
	Username string        `mapstructure:"username"`
	Password string        `mapstructure:"password"`
	Timeout  time.Duration `mapstructure:"timeout"`
	
	// 推送网关配置（用于写入）
	PushGatewayURL string `mapstructure:"push_gateway_url"`
	JobName        string `mapstructure:"job_name"`
	Instance       string `mapstructure:"instance"`
}

// PrometheusStorageFactory Prometheus存储工厂
type PrometheusStorageFactory struct{}

// Create 创建Prometheus存储实例
func (f *PrometheusStorageFactory) Create(config StorageConfig) (Storage, error) {
	promConfig := PrometheusStorageConfig{
		Timeout:  30 * time.Second,
		JobName:  "laojun-monitoring",
		Instance: "localhost",
	}

	// 解析配置
	if url, ok := config.Config["url"].(string); ok {
		promConfig.URL = url
	}
	if username, ok := config.Config["username"].(string); ok {
		promConfig.Username = username
	}
	if password, ok := config.Config["password"].(string); ok {
		promConfig.Password = password
	}
	if timeout, ok := config.Config["timeout"].(string); ok {
		if d, err := time.ParseDuration(timeout); err == nil {
			promConfig.Timeout = d
		}
	}
	if pushGatewayURL, ok := config.Config["push_gateway_url"].(string); ok {
		promConfig.PushGatewayURL = pushGatewayURL
	}
	if jobName, ok := config.Config["job_name"].(string); ok {
		promConfig.JobName = jobName
	}
	if instance, ok := config.Config["instance"].(string); ok {
		promConfig.Instance = instance
	}

	return NewPrometheusStorage(promConfig)
}

// SupportedTypes 支持的存储类型
func (f *PrometheusStorageFactory) SupportedTypes() []string {
	return []string{"prometheus"}
}

// NewPrometheusStorage 创建Prometheus存储
func NewPrometheusStorage(config PrometheusStorageConfig) (*PrometheusStorage, error) {
	if config.URL == "" {
		return nil, NewStorageError("config", "prometheus", 
			fmt.Errorf("prometheus URL is required"))
	}

	// 创建Prometheus客户端
	clientConfig := api.Config{
		Address: config.URL,
	}

	if config.Username != "" && config.Password != "" {
		clientConfig.RoundTripper = &basicAuthRoundTripper{
			username: config.Username,
			password: config.Password,
			next:     http.DefaultTransport,
		}
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, NewStorageError("client", "prometheus", err)
	}

	queryAPI := v1.NewAPI(client)

	return &PrometheusStorage{
		client:   client,
		queryAPI: queryAPI,
		config:   config,
	}, nil
}

// basicAuthRoundTripper HTTP基本认证
type basicAuthRoundTripper struct {
	username string
	password string
	next     http.RoundTripper
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.username, rt.password)
	return rt.next.RoundTrip(req)
}

// Write 写入指标数据（通过Push Gateway）
func (s *PrometheusStorage) Write(ctx context.Context, samples []MetricSample) error {
	if s.config.PushGatewayURL == "" {
		return NewStorageError("write", "prometheus", 
			fmt.Errorf("push gateway URL not configured"))
	}

	// 按指标名称分组
	metricGroups := make(map[string][]MetricSample)
	for _, sample := range samples {
		metricGroups[sample.MetricName] = append(metricGroups[sample.MetricName], sample)
	}

	// 为每个指标组推送数据
	for metricName, metricSamples := range metricGroups {
		if err := s.pushMetricGroup(ctx, metricName, metricSamples); err != nil {
			return err
		}
	}

	return nil
}

// pushMetricGroup 推送指标组数据
func (s *PrometheusStorage) pushMetricGroup(ctx context.Context, metricName string, samples []MetricSample) error {
	// 构建Prometheus格式的数据
	var lines []string
	
	for _, sample := range samples {
		line := s.formatSample(sample)
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	data := strings.Join(lines, "\n")

	// 构建Push Gateway URL
	pushURL := fmt.Sprintf("%s/metrics/job/%s/instance/%s",
		strings.TrimRight(s.config.PushGatewayURL, "/"),
		url.QueryEscape(s.config.JobName),
		url.QueryEscape(s.config.Instance))

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", pushURL, strings.NewReader(data))
	if err != nil {
		return NewStorageError("request", "prometheus", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: s.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return NewStorageError("push", "prometheus", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewStorageError("push", "prometheus", 
			fmt.Errorf("push failed with status %d: %s", resp.StatusCode, string(body)))
	}

	return nil
}

// formatSample 格式化样本为Prometheus格式
func (s *PrometheusStorage) formatSample(sample MetricSample) string {
	metricName := sample.MetricName
	
	// 构建标签字符串
	var labelParts []string
	for k, v := range sample.Labels {
		labelParts = append(labelParts, fmt.Sprintf(`%s="%s"`, k, v))
	}

	var labelStr string
	if len(labelParts) > 0 {
		labelStr = "{" + strings.Join(labelParts, ",") + "}"
	}

	// 格式化值
	valueStr := strconv.FormatFloat(sample.Value, 'f', -1, 64)
	
	// 格式化时间戳（毫秒）
	timestamp := sample.Timestamp.UnixNano() / int64(time.Millisecond)
	
	return fmt.Sprintf("%s%s %s %d", metricName, labelStr, valueStr, timestamp)
}

// Query 查询指标数据
func (s *PrometheusStorage) Query(ctx context.Context, query string, timestamp time.Time) ([]QueryResult, error) {
	result, warnings, err := s.queryAPI.Query(ctx, query, timestamp)
	if err != nil {
		return nil, NewQueryError("query", query, err)
	}

	if len(warnings) > 0 {
		// 记录警告但不返回错误
		fmt.Printf("Prometheus query warnings: %v\n", warnings)
	}

	return s.convertResult(result), nil
}

// QueryRange 范围查询
func (s *PrometheusStorage) QueryRange(ctx context.Context, query string, queryRange QueryRange) ([]QueryResult, error) {
	// 计算步长（默认为范围的1/100）
	step := queryRange.End.Sub(queryRange.Start) / 100
	if step < time.Second {
		step = time.Second
	}

	promRange := v1.Range{
		Start: queryRange.Start,
		End:   queryRange.End,
		Step:  step,
	}

	result, warnings, err := s.queryAPI.QueryRange(ctx, query, promRange)
	if err != nil {
		return nil, NewQueryError("query_range", query, err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Prometheus query range warnings: %v\n", warnings)
	}

	return s.convertResult(result), nil
}

// convertResult 转换Prometheus结果
func (s *PrometheusStorage) convertResult(result model.Value) []QueryResult {
	var results []QueryResult

	switch v := result.(type) {
	case model.Vector:
		for _, sample := range v {
			results = append(results, QueryResult{
				MetricName: string(sample.Metric["__name__"]),
				Labels:     s.convertLabels(sample.Metric),
				Samples: []MetricSample{{
					MetricName: string(sample.Metric["__name__"]),
					Labels:     s.convertLabels(sample.Metric),
					Value:      float64(sample.Value),
					Timestamp:  sample.Timestamp.Time(),
				}},
			})
		}

	case model.Matrix:
		for _, sampleStream := range v {
			var samples []MetricSample
			for _, pair := range sampleStream.Values {
				samples = append(samples, MetricSample{
					MetricName: string(sampleStream.Metric["__name__"]),
					Labels:     s.convertLabels(sampleStream.Metric),
					Value:      float64(pair.Value),
					Timestamp:  pair.Timestamp.Time(),
				})
			}

			if len(samples) > 0 {
				results = append(results, QueryResult{
					MetricName: string(sampleStream.Metric["__name__"]),
					Labels:     s.convertLabels(sampleStream.Metric),
					Samples:    samples,
				})
			}
		}

	case *model.Scalar:
		results = append(results, QueryResult{
			MetricName: "scalar",
			Labels:     make(map[string]string),
			Samples: []MetricSample{{
				MetricName: "scalar",
				Labels:     make(map[string]string),
				Value:      float64(v.Value),
				Timestamp:  v.Timestamp.Time(),
			}},
		})
	}

	return results
}

// convertLabels 转换标签
func (s *PrometheusStorage) convertLabels(metric model.Metric) map[string]string {
	labels := make(map[string]string)
	for k, v := range metric {
		if k != "__name__" {
			labels[string(k)] = string(v)
		}
	}
	return labels
}

// LabelValues 获取标签值
func (s *PrometheusStorage) LabelValues(ctx context.Context, labelName string) ([]string, error) {
	values, warnings, err := s.queryAPI.LabelValues(ctx, labelName, nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, NewQueryError("label_values", labelName, err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Prometheus label values warnings: %v\n", warnings)
	}

	result := make([]string, len(values))
	for i, v := range values {
		result[i] = string(v)
	}

	return result, nil
}

// LabelNames 获取标签名称
func (s *PrometheusStorage) LabelNames(ctx context.Context) ([]string, error) {
	names, warnings, err := s.queryAPI.LabelNames(ctx, nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, NewQueryError("label_names", "", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Prometheus label names warnings: %v\n", warnings)
	}

	result := make([]string, len(names))
	for i, name := range names {
		result[i] = string(name)
	}

	return result, nil
}

// MetricNames 获取指标名称
func (s *PrometheusStorage) MetricNames(ctx context.Context) ([]string, error) {
	// 使用特殊查询获取所有指标名称
	result, warnings, err := s.queryAPI.Query(ctx, "group by (__name__) ({__name__=~\".+\"})", time.Now())
	if err != nil {
		return nil, NewQueryError("metric_names", "", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Prometheus metric names warnings: %v\n", warnings)
	}

	var names []string
	if vector, ok := result.(model.Vector); ok {
		for _, sample := range vector {
			if name, exists := sample.Metric["__name__"]; exists {
				names = append(names, string(name))
			}
		}
	}

	return names, nil
}

// DeleteExpiredData 删除过期数据（Prometheus不支持）
func (s *PrometheusStorage) DeleteExpiredData(ctx context.Context, before time.Time) error {
	// Prometheus通过配置的retention policy自动删除过期数据
	// 这里返回nil表示操作成功（实际上是no-op）
	return nil
}

// Health 健康检查
func (s *PrometheusStorage) Health(ctx context.Context) error {
	// 执行简单查询来检查连接
	_, _, err := s.queryAPI.Query(ctx, "up", time.Now())
	if err != nil {
		return NewStorageError("health_check", "prometheus", err)
	}

	return nil
}

// Close 关闭存储
func (s *PrometheusStorage) Close() error {
	// Prometheus客户端不需要显式关闭
	return nil
}

// GetStats 获取存储统计信息
func (s *PrometheusStorage) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type":             "prometheus",
		"url":              s.config.URL,
		"push_gateway_url": s.config.PushGatewayURL,
		"job_name":         s.config.JobName,
		"instance":         s.config.Instance,
	}
}