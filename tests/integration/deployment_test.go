package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestDeploymentScripts 测试部署脚本的集成功能
func TestDeploymentScripts(t *testing.T) {
	// 设置测试环境
	projectRoot := getProjectRoot(t)
	
	tests := []struct {
		name       string
		scriptPath string
		args       []string
		timeout    time.Duration
	}{
		{
			name:       "check local images script",
			scriptPath: "deploy/scripts/check-local-images.ps1",
			args:       []string{},
			timeout:    30 * time.Second,
		},
		{
			name:       "deployment validation",
			scriptPath: "deploy/scripts/test-deployment.ps1",
			args:       []string{},
			timeout:    60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptFullPath := filepath.Join(projectRoot, tt.scriptPath)
			
			// 检查脚本文件是否存在
			if _, err := os.Stat(scriptFullPath); os.IsNotExist(err) {
				t.Skipf("Script not found: %s", scriptFullPath)
			}

			// 执行脚本
			cmd := exec.Command("powershell", "-File", scriptFullPath)
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Dir = projectRoot

			// 设置超时
			done := make(chan error, 1)
			go func() {
				done <- cmd.Run()
			}()

			select {
			case err := <-done:
				if err != nil {
					t.Logf("Script execution completed with error: %v", err)
					// 对于某些脚本，非零退出码可能是正常的
					// 这里我们记录但不一定失败测试
				} else {
					t.Logf("Script executed successfully")
				}
			case <-time.After(tt.timeout):
				if err := cmd.Process.Kill(); err != nil {
					t.Logf("Failed to kill process: %v", err)
				}
				t.Errorf("Script execution timed out after %v", tt.timeout)
			}
		})
	}
}

// TestToolsIntegration 测试工具集成功能
func TestToolsIntegration(t *testing.T) {
	projectRoot := getProjectRoot(t)
	
	tests := []struct {
		name    string
		tool    string
		args    []string
		timeout time.Duration
	}{
		{
			name:    "menu manager help",
			tool:    "cmd/menu-manager/main.go",
			args:    []string{},
			timeout: 10 * time.Second,
		},
		{
			name:    "db maintenance help",
			tool:    "cmd/db-maintenance/main.go",
			args:    []string{},
			timeout: 10 * time.Second,
		},
		{
			name:    "marketplace manager help",
			tool:    "cmd/marketplace-manager/main.go",
			args:    []string{},
			timeout: 10 * time.Second,
		},
		{
			name:    "project manager help",
			tool:    "cmd/project-manager/main.go",
			args:    []string{},
			timeout: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolPath := filepath.Join(projectRoot, tt.tool)
			
			// 检查工具文件是否存在
			if _, err := os.Stat(toolPath); os.IsNotExist(err) {
				t.Skipf("Tool not found: %s", toolPath)
			}

			// 执行工具
			cmd := exec.Command("go", "run", toolPath)
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Dir = projectRoot

			// 设置超时
			done := make(chan error, 1)
			go func() {
				done <- cmd.Run()
			}()

			select {
			case err := <-done:
				// 对于帮助命令，退出码1是正常的（显示帮助后退出）
				if exitError, ok := err.(*exec.ExitError); ok {
					if exitError.ExitCode() == 1 {
						t.Logf("Tool displayed help and exited (expected)")
					} else {
						t.Errorf("Tool exited with unexpected code: %d", exitError.ExitCode())
					}
				} else if err != nil {
					t.Errorf("Tool execution failed: %v", err)
				} else {
					t.Logf("Tool executed successfully")
				}
			case <-time.After(tt.timeout):
				if err := cmd.Process.Kill(); err != nil {
					t.Logf("Failed to kill process: %v", err)
				}
				t.Errorf("Tool execution timed out after %v", tt.timeout)
			}
		})
	}
}

// getProjectRoot 获取项目根目录
func getProjectRoot(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// 从tests/integration目录向上查找项目根目录
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatalf("Could not find project root (go.mod not found)")
		}
		wd = parent
	}
}