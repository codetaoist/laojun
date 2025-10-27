package main

import (
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/observability"
)

func main() {
	fmt.Println("=== 配置验证和默认值处理示例 ===")

	// 测试1: 默认配置验证
	testDefaultConfig()

	// 测试2: 无效配置验证
	testInvalidConfigs()

	// 测试3: 边界值测试
	testBoundaryValues()

	// 测试4: 默认值应用
	testDefaultValueApplication()

	// 测试5: 复杂配置验证
	testComplexConfigValidation()

	// 测试6: 导出配置验证
	testExportConfigValidation()

	fmt.Println("=== 配置验证示例完成! ===")
}

// testDefaultConfig 测试默认配置
func testDefaultConfig() {
	fmt.Println("\n--- 测试1: 默认配置验证 ---")

	config := observability.DefaultConfig()
	
	fmt.Printf("默认配置:\n")
	fmt.Printf("  服务名称: %s\n", config.ServiceName)
	fmt.Printf("  服务版本: %s\n", config.ServiceVersion)
	fmt.Printf("  环境: %s\n", config.Environment)
	fmt.Printf("  采样率: %.1f\n", config.SampleRate)
	fmt.Printf("  超时: %v\n", config.Timeout)
	fmt.Printf("  缓冲区大小: %d\n", config.BufferSize)
	fmt.Printf("  刷新周期: %v\n", config.FlushPeriod)

	if err := config.Validate(); err != nil {
		log.Printf("默认配置验证失败: %v", err)
	} else {
		fmt.Println("✓ 默认配置验证通过")
	}
}

// testInvalidConfigs 测试无效配置
func testInvalidConfigs() {
	fmt.Println("\n--- 测试2: 无效配置验证 ---")

	testCases := []struct {
		name   string
		config *observability.Config
	}{
		{
			name: "空服务名称",
			config: &observability.Config{
				ServiceName:    "",
				ServiceVersion: "1.0.0",
				Environment:    "development",
			},
		},
		{
			name: "无效服务名称格式",
			config: &observability.Config{
				ServiceName:    "-invalid-name-",
				ServiceVersion: "1.0.0",
				Environment:    "development",
			},
		},
		{
			name: "服务名称过长",
			config: &observability.Config{
				ServiceName:    "this-is-a-very-long-service-name-that-exceeds-the-maximum-allowed-length-of-one-hundred-characters",
				ServiceVersion: "1.0.0",
				Environment:    "development",
			},
		},
		{
			name: "空服务版本",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "",
				Environment:    "development",
			},
		},
		{
			name: "无效环境名称",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "invalid-env",
			},
		},
		{
			name: "负采样率",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     -0.1,
			},
		},
		{
			name: "采样率超过1",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.5,
			},
		},
		{
			name: "负超时",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        -1 * time.Second,
			},
		},
		{
			name: "超时过长",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        10 * time.Minute,
			},
		},
		{
			name: "缓冲区大小为0",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     0,
			},
		},
		{
			name: "缓冲区大小过大",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     200000,
			},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("测试: %s\n", tc.name)
		if err := tc.config.Validate(); err != nil {
			fmt.Printf("  ✓ 预期验证失败: %v\n", err)
		} else {
			fmt.Printf("  ✗ 意外验证通过\n")
		}
	}
}

// testBoundaryValues 测试边界值
func testBoundaryValues() {
	fmt.Println("\n--- 测试3: 边界值测试 ---")

	testCases := []struct {
		name   string
		config *observability.Config
		valid  bool
	}{
		{
			name: "最小有效采样率",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     0.0,
				Timeout:        1 * time.Second,
				BufferSize:     1,
				FlushPeriod:    1 * time.Second,
			},
			valid: true,
		},
		{
			name: "最大有效采样率",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        5 * time.Minute,
				BufferSize:     100000,
				FlushPeriod:    1 * time.Hour,
			},
			valid: true,
		},
		{
			name: "最小服务名称长度",
			config: &observability.Config{
				ServiceName:    "a",
				ServiceVersion: "1",
				Environment:    "dev",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     1000,
				FlushPeriod:    10 * time.Second,
			},
			valid: true,
		},
		{
			name: "最大服务名称长度",
			config: &observability.Config{
				ServiceName:    "a123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     1000,
				FlushPeriod:    10 * time.Second,
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("测试: %s\n", tc.name)
		err := tc.config.Validate()
		if tc.valid {
			if err == nil {
				fmt.Printf("  ✓ 验证通过\n")
			} else {
				fmt.Printf("  ✗ 意外验证失败: %v\n", err)
			}
		} else {
			if err != nil {
				fmt.Printf("  ✓ 预期验证失败: %v\n", err)
			} else {
				fmt.Printf("  ✗ 意外验证通过\n")
			}
		}
	}
}

// testDefaultValueApplication 测试默认值应用
func testDefaultValueApplication() {
	fmt.Println("\n--- 测试4: 默认值应用 ---")

	// 创建空配置
	config := &observability.Config{}
	
	fmt.Println("应用默认值前:")
	fmt.Printf("  服务名称: '%s'\n", config.ServiceName)
	fmt.Printf("  服务版本: '%s'\n", config.ServiceVersion)
	fmt.Printf("  环境: '%s'\n", config.Environment)
	fmt.Printf("  采样率: %.1f\n", config.SampleRate)
	fmt.Printf("  超时: %v\n", config.Timeout)
	fmt.Printf("  缓冲区大小: %d\n", config.BufferSize)
	fmt.Printf("  刷新周期: %v\n", config.FlushPeriod)

	// 应用默认值
	config.ApplyDefaults()

	fmt.Println("\n应用默认值后:")
	fmt.Printf("  服务名称: '%s'\n", config.ServiceName)
	fmt.Printf("  服务版本: '%s'\n", config.ServiceVersion)
	fmt.Printf("  环境: '%s'\n", config.Environment)
	fmt.Printf("  采样率: %.1f\n", config.SampleRate)
	fmt.Printf("  超时: %v\n", config.Timeout)
	fmt.Printf("  缓冲区大小: %d\n", config.BufferSize)
	fmt.Printf("  刷新周期: %v\n", config.FlushPeriod)

	if err := config.Validate(); err != nil {
		fmt.Printf("✗ 应用默认值后验证失败: %v\n", err)
	} else {
		fmt.Println("✓ 应用默认值后验证通过")
	}

	// 测试部分配置的默认值应用
	fmt.Println("\n测试部分配置:")
	partialConfig := &observability.Config{
		ServiceName:    "my-service",
		ServiceVersion: "2.0.0",
		// 其他字段为空，应该应用默认值
	}

	partialConfig.ApplyDefaults()
	fmt.Printf("  服务名称: '%s' (保持原值)\n", partialConfig.ServiceName)
	fmt.Printf("  服务版本: '%s' (保持原值)\n", partialConfig.ServiceVersion)
	fmt.Printf("  环境: '%s' (应用默认值)\n", partialConfig.Environment)
	fmt.Printf("  采样率: %.1f (应用默认值)\n", partialConfig.SampleRate)
}

// testComplexConfigValidation 测试复杂配置验证
func testComplexConfigValidation() {
	fmt.Println("\n--- 测试5: 复杂配置验证 ---")

	// 测试资源属性验证
	config := &observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		SampleRate:     0.5,
		Timeout:        30 * time.Second,
		BufferSize:     1000,
		FlushPeriod:    10 * time.Second,
		ResourceAttributes: map[string]string{
			"team":        "backend",
			"region":      "us-west-2",
			"datacenter":  "dc1",
			"":            "empty-key",  // 无效：空键
		},
	}

	fmt.Println("测试资源属性验证:")
	if err := config.Validate(); err != nil {
		fmt.Printf("  ✓ 预期验证失败: %v\n", err)
	} else {
		fmt.Printf("  ✗ 意外验证通过\n")
	}

	// 修复配置
	delete(config.ResourceAttributes, "")
	config.ResourceAttributes["component"] = "api-server"

	if err := config.Validate(); err != nil {
		fmt.Printf("  ✗ 修复后验证失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 修复后验证通过\n")
	}

	// 测试ValidateAndApplyDefaults方法
	fmt.Println("\n测试ValidateAndApplyDefaults:")
	emptyConfig := &observability.Config{}
	if err := emptyConfig.ValidateAndApplyDefaults(); err != nil {
		fmt.Printf("  ✗ ValidateAndApplyDefaults失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ ValidateAndApplyDefaults成功\n")
		fmt.Printf("    服务名称: %s\n", emptyConfig.ServiceName)
		fmt.Printf("    环境: %s\n", emptyConfig.Environment)
	}
}

// testExportConfigValidation 测试导出配置验证
func testExportConfigValidation() {
	fmt.Println("\n--- 测试6: 导出配置验证 ---")

	testCases := []struct {
		name   string
		config *observability.Config
		valid  bool
	}{
		{
			name: "有效导出配置",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     1000,
				FlushPeriod:    10 * time.Second,
				Export: &observability.ExportConfig{
					Endpoints: map[string]string{
						"jaeger": "http://localhost:14268/api/traces",
						"otlp":   "http://localhost:4317",
					},
					Formats:      []string{"json", "otlp"},
					Timeout:      10 * time.Second,
					MaxRetries:   3,
					RetryDelay:   1 * time.Second,
					BatchSize:    100,
					BatchTimeout: 5 * time.Second,
					MaxQueueSize: 1000,
				},
			},
			valid: true,
		},
		{
			name: "无效URL端点",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     1000,
				FlushPeriod:    10 * time.Second,
				Export: &observability.ExportConfig{
					Endpoints: map[string]string{
						"invalid": "not-a-valid-url",
					},
					Formats:      []string{"json"},
					Timeout:      10 * time.Second,
					MaxRetries:   3,
					RetryDelay:   1 * time.Second,
					BatchSize:    100,
					BatchTimeout: 5 * time.Second,
					MaxQueueSize: 1000,
				},
			},
			valid: false,
		},
		{
			name: "不支持的导出格式",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     1000,
				FlushPeriod:    10 * time.Second,
				Export: &observability.ExportConfig{
					Endpoints: map[string]string{
						"jaeger": "http://localhost:14268/api/traces",
					},
					Formats:      []string{"unsupported-format"},
					Timeout:      10 * time.Second,
					MaxRetries:   3,
					RetryDelay:   1 * time.Second,
					BatchSize:    100,
					BatchTimeout: 5 * time.Second,
					MaxQueueSize: 1000,
				},
			},
			valid: false,
		},
		{
			name: "批量大小超出范围",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     1000,
				FlushPeriod:    10 * time.Second,
				Export: &observability.ExportConfig{
					Endpoints: map[string]string{
						"jaeger": "http://localhost:14268/api/traces",
					},
					Formats:      []string{"json"},
					Timeout:      10 * time.Second,
					MaxRetries:   3,
					RetryDelay:   1 * time.Second,
					BatchSize:    50000, // 超出最大值
					BatchTimeout: 5 * time.Second,
					MaxQueueSize: 1000,
				},
			},
			valid: false,
		},
		{
			name: "空端点配置",
			config: &observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "development",
				SampleRate:     1.0,
				Timeout:        30 * time.Second,
				BufferSize:     1000,
				FlushPeriod:    10 * time.Second,
				Export: &observability.ExportConfig{
					Endpoints:    map[string]string{}, // 空端点
					Formats:      []string{"json"},
					Timeout:      10 * time.Second,
					MaxRetries:   3,
					RetryDelay:   1 * time.Second,
					BatchSize:    100,
					BatchTimeout: 5 * time.Second,
					MaxQueueSize: 1000,
				},
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("测试: %s\n", tc.name)
		err := tc.config.Validate()
		if tc.valid {
			if err == nil {
				fmt.Printf("  ✓ 验证通过\n")
			} else {
				fmt.Printf("  ✗ 意外验证失败: %v\n", err)
			}
		} else {
			if err != nil {
				fmt.Printf("  ✓ 预期验证失败: %v\n", err)
			} else {
				fmt.Printf("  ✗ 意外验证通过\n")
			}
		}
	}

	// 测试导出配置默认值应用
	fmt.Println("\n测试导出配置默认值应用:")
	exportConfig := &observability.ExportConfig{}
	
	fmt.Println("应用默认值前:")
	fmt.Printf("  超时: %v\n", exportConfig.Timeout)
	fmt.Printf("  最大重试: %d\n", exportConfig.MaxRetries)
	fmt.Printf("  批量大小: %d\n", exportConfig.BatchSize)
	fmt.Printf("  端点数量: %d\n", len(exportConfig.Endpoints))

	exportConfig.ApplyDefaults()

	fmt.Println("应用默认值后:")
	fmt.Printf("  超时: %v\n", exportConfig.Timeout)
	fmt.Printf("  最大重试: %d\n", exportConfig.MaxRetries)
	fmt.Printf("  批量大小: %d\n", exportConfig.BatchSize)
	fmt.Printf("  端点数量: %d\n", len(exportConfig.Endpoints))
	fmt.Printf("  格式数量: %d\n", len(exportConfig.Formats))

	if err := exportConfig.Validate(); err != nil {
		fmt.Printf("  ✗ 应用默认值后验证失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 应用默认值后验证通过\n")
	}
}