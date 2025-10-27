package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ModuleTemplate 模块模板数据
type ModuleTemplate struct {
	PackageName string
	ModuleName  string
	UpperName   string
}

// 接口模板
var interfaceTemplate = `package {{.PackageName}}

import (
	"context"
	"fmt"
	"time"
)

// {{.UpperName}} defines the interface for {{.PackageName}} operations.
// All implementations must be thread-safe.
type {{.UpperName}} interface {
	// TODO: Define your interface methods here
	// Example:
	// Get(ctx context.Context, key string) (interface{}, error)
	// Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// Config defines the configuration for {{.PackageName}}.
type Config struct {
	Enabled bool          ` + "`" + `yaml:"enabled" env:"{{.UpperName}}_ENABLED" default:"true"` + "`" + `
	Debug   bool          ` + "`" + `yaml:"debug" env:"{{.UpperName}}_DEBUG" default:"false"` + "`" + `
	Timeout time.Duration ` + "`" + `yaml:"timeout" env:"{{.UpperName}}_TIMEOUT" default:"30s"` + "`" + `
	
	// TODO: Add your specific configuration fields here
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	// TODO: Add your validation logic here
	return nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Debug:   false,
		Timeout: 30 * time.Second,
	}
}
`

// 实现模板
var implementationTemplate = `package {{.PackageName}}

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// {{.ModuleName}}Impl implements the {{.UpperName}} interface.
type {{.ModuleName}}Impl struct {
	config Config
	mu     sync.RWMutex
	// TODO: Add your implementation fields here
}

// New creates a new {{.PackageName}} instance.
func New(config Config) {{.UpperName}} {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}
	
	return &{{.ModuleName}}Impl{
		config: config,
	}
}

// TODO: Implement your interface methods here
// Example:
// func (impl *{{.ModuleName}}Impl) Get(ctx context.Context, key string) (interface{}, error) {
//     impl.mu.RLock()
//     defer impl.mu.RUnlock()
//     
//     // TODO: Implement your logic here
//     return nil, fmt.Errorf("not implemented")
// }

// Close closes the {{.PackageName}} instance and releases resources.
func (impl *{{.ModuleName}}Impl) Close() error {
	// TODO: Implement cleanup logic here
	return nil
}
`

// 错误定义模板
var errorsTemplate = `package {{.PackageName}}

import "errors"

// Common errors for {{.PackageName}} operations.
var (
	ErrNotFound     = errors.New("{{.PackageName}}: not found")
	ErrInvalidInput = errors.New("{{.PackageName}}: invalid input")
	ErrTimeout      = errors.New("{{.PackageName}}: operation timeout")
	ErrConnection   = errors.New("{{.PackageName}}: connection failed")
	ErrNotSupported = errors.New("{{.PackageName}}: operation not supported")
	
	// TODO: Add your specific errors here
)

// IsNotFound checks if the error is a "not found" error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsTimeout checks if the error is a timeout error.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// TODO: Add more error checking functions as needed
`

// 测试模板
var testTemplate = `package {{.PackageName}}

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	config := DefaultConfig()
	
	impl := New(config)
	if impl == nil {
		t.Fatal("New() returned nil")
	}
	
	// TODO: Add more tests
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid timeout",
			config: Config{
				Enabled: true,
				Timeout: -1,
			},
			wantErr: true,
		},
		// TODO: Add more test cases
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: Add more tests for your interface methods
`

// 示例模板
var exampleTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/{{.PackageName}}"
)

func main() {
	fmt.Println("=== {{.UpperName}} 示例 ===")
	
	// 1. 创建配置
	config := {{.PackageName}}.DefaultConfig()
	config.Debug = true
	
	// 2. 创建实例
	impl := {{.PackageName}}.New(config)
	defer func() {
		if closer, ok := impl.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				log.Printf("关闭失败: %v", err)
			}
		}
	}()
	
	// 3. 使用示例
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// TODO: Add your usage examples here
	fmt.Println("TODO: 添加使用示例")
	
	fmt.Println("=== 示例完成 ===")
}
`

func main() {
	var (
		packageName = flag.String("package", "", "Package name (required)")
		outputDir   = flag.String("output", ".", "Output directory")
	)
	flag.Parse()

	if *packageName == "" {
		log.Fatal("Package name is required")
	}

	// 创建模板数据
	data := ModuleTemplate{
		PackageName: *packageName,
		ModuleName:  strings.Title(*packageName),
		UpperName:   strings.ToUpper(*packageName),
	}

	// 创建输出目录
	pkgDir := filepath.Join(*outputDir, *packageName)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	// 生成文件
	files := map[string]string{
		"interface.go":      interfaceTemplate,
		"implementation.go": implementationTemplate,
		"errors.go":         errorsTemplate,
		"interface_test.go": testTemplate,
	}

	for filename, tmplContent := range files {
		if err := generateFile(filepath.Join(pkgDir, filename), tmplContent, data); err != nil {
			log.Fatalf("生成文件 %s 失败: %v", filename, err)
		}
		fmt.Printf("生成文件: %s\n", filepath.Join(pkgDir, filename))
	}

	// 生成示例文件
	exampleDir := filepath.Join(*outputDir, "examples")
	if err := os.MkdirAll(exampleDir, 0755); err != nil {
		log.Fatalf("创建示例目录失败: %v", err)
	}

	exampleFile := filepath.Join(exampleDir, fmt.Sprintf("%s_example.go", *packageName))
	if err := generateFile(exampleFile, exampleTemplate, data); err != nil {
		log.Fatalf("生成示例文件失败: %v", err)
	}
	fmt.Printf("生成示例文件: %s\n", exampleFile)

	fmt.Printf("\n✅ 成功生成 %s 模块的代码模板\n", *packageName)
	fmt.Println("\n下一步:")
	fmt.Printf("1. 编辑 %s/interface.go 定义接口方法\n", pkgDir)
	fmt.Printf("2. 编辑 %s/implementation.go 实现接口方法\n", pkgDir)
	fmt.Printf("3. 编辑 %s/errors.go 添加特定错误\n", pkgDir)
	fmt.Printf("4. 编辑 %s/interface_test.go 添加测试用例\n", pkgDir)
	fmt.Printf("5. 编辑 %s 添加使用示例\n", exampleFile)
}

func generateFile(filename, tmplContent string, data ModuleTemplate) error {
	tmpl, err := template.New("file").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("执行模板失败: %w", err)
	}

	return nil
}