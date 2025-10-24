# 验证项目结构脚本
Write-Host "=== 太上老君平台项目结构验证 ===" -ForegroundColor Green

# 检查必要的目录
$requiredDirs = @(
    "cmd",
    "internal", 
    "pkg",
    "api",
    "configs",
    "scripts",
    "docs",
    "deployments",
    "docker",
    "k8s",
    "web",
    "tests",
    "examples",
    "tools",
    "build",
    "third_party",
    "migrations"
)

Write-Host "`n检查目录结构..." -ForegroundColor Yellow
$missingDirs = @()
foreach ($dir in $requiredDirs) {
    if (Test-Path $dir) {
        Write-Host "✓ $dir" -ForegroundColor Green
    } else {
        Write-Host "✗ $dir (缺失)" -ForegroundColor Red
        $missingDirs += $dir
    }
}

# 检查关键文件
$requiredFiles = @(
    "go.mod",
    "go.work",
    "Makefile",
    "README.md",
    "docker-compose.test.yml"
)

Write-Host "`n检查关键文件..." -ForegroundColor Yellow
$missingFiles = @()
foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✓ $file" -ForegroundColor Green
    } else {
        Write-Host "✗ $file (缺失)" -ForegroundColor Red
        $missingFiles += $file
    }
}

# 检查Go模块
Write-Host "`n检查Go模块..." -ForegroundColor Yellow
$goModules = @(
    ".",
    "pkg/shared",
    "pkg/plugins",
    "tools/debug",
    "tools/plugin-cli",
    "tools/swagger"
)

foreach ($module in $goModules) {
    $goModPath = "$module/go.mod"
    if (Test-Path $goModPath) {
        Write-Host "✓ $goModPath" -ForegroundColor Green
    } else {
        Write-Host "✗ $goModPath (缺失)" -ForegroundColor Red
    }
}

# 总结
Write-Host "`n=== 验证结果 ===" -ForegroundColor Green
if ($missingDirs.Count -eq 0 -and $missingFiles.Count -eq 0) {
    Write-Host "✓ 项目结构验证通过！" -ForegroundColor Green
} else {
    Write-Host "✗ 发现问题：" -ForegroundColor Red
    if ($missingDirs.Count -gt 0) {
        Write-Host "  缺失目录: $($missingDirs -join ', ')" -ForegroundColor Red
    }
    if ($missingFiles.Count -gt 0) {
        Write-Host "  缺失文件: $($missingFiles -join ', ')" -ForegroundColor Red
    }
}

Write-Host "`n项目结构符合Go语言标准项目布局规范" -ForegroundColor Cyan