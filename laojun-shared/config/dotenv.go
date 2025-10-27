package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadDotenv 统一加载 .env 文件，支持环境分层覆盖
// 加载顺序（若存在则依次加载并覆盖同名变量）：
// 1) 项目根目录/.env
// 2) 项目根目录/.env.<APP_ENV>（APP_ENV 或 ENV 环境变量）
// 3) 项目根目录/.env.local（本地开发覆盖）
// Fallback：若上述均不存在，尝试当前工作目录/.env
// 注意：为最佳努力，不会因为 .env 缺失而报错
func LoadDotenv() {
	root := findProjectRoot()
	// 解析环境名
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = strings.TrimSpace(os.Getenv("ENV"))
	}
	if env != "" {
		env = strings.ToLower(env)
	}

	// 构造候选文件（仅项目根目录）
	candidates := []string{filepath.Join(root, ".env")}
	if env != "" {
		candidates = append(candidates, filepath.Join(root, ".env."+env))
	}
	candidates = append(candidates, filepath.Join(root, ".env.local"))

	loaded := make([]string, 0, len(candidates))
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := loadEnvFile(p); err != nil {
				fmt.Printf("Warning: failed to load %s: %v\n", p, err)
				continue
			}
			loaded = append(loaded, p)
		}
	}

	if len(loaded) > 0 {
		fmt.Println("Loaded env files:", strings.Join(loaded, ", "))
		return
	}

	// Fallback: 当前工作目录 .env
	if _, err := os.Stat(".env"); err == nil {
		_ = loadEnvFile(".env")
		fmt.Println("Loaded env file from CWD: .env")
		return
	}

	fmt.Println("No .env file found; using system environment")
}

// findProjectRoot 通过向上查找标识文件定位项目根目录
func findProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	dir := cwd
	for {
		markers := []string{"go.work", ".git", "go.mod"}
		for _, m := range markers {
			if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
				// 优先返回包含 go.work 的目录（多模块工作区）
				if m == "go.work" {
					return dir
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir { // 到达根目录
			break
		}
		dir = parent
	}

	return cwd
}

// loadEnvFile 解析简单的 KEY=VALUE 格式并注入到进程环境
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		// 去掉包裹引号
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		// 展开环境变量引用（如 ${VAR} 或 $VAR）
		val = os.ExpandEnv(val)

		_ = os.Setenv(key, val)
	}

	return scanner.Err()
}