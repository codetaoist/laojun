# Laojun 工作区环境设置脚本
# 用于初始化开发环境和克隆所有子仓库

param(
    [switch]$SkipClone,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 工作区环境设置 ===" -ForegroundColor Green

# 检查 Go 版本
Write-Host "检查 Go 环境..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "Go 版本: $goVersion" -ForegroundColor Green
    
    # 检查 Go 版本是否满足要求 (1.21+)
    if ($goVersion -match "go(\d+)\.(\d+)") {
        $major = [int]$matches[1]
        $minor = [int]$matches[2]
        if ($major -lt 1 -or ($major -eq 1 -and $minor -lt 21)) {
            Write-Warning "建议使用 Go 1.21 或更高版本"
        }
    }
} catch {
    Write-Error "Go 未安装或不在 PATH 中。请先安装 Go 1.21+"
    exit 1
}

# 工作区根目录
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

Write-Host "工作区根目录: $workspaceRoot" -ForegroundColor Cyan
Write-Host "父目录: $parentDir" -ForegroundColor Cyan

# 子仓库列表
$repositories = @(
    "laojun-shared",
    "laojun-admin-api", 
    "laojun-marketplace-api",
    "laojun-plugins",
    "laojun-config-center",
    "laojun-admin-web",
    "laojun-marketplace-web",
    "laojun-deploy",
    "laojun-docs"
)

# 检查子仓库是否存在
Write-Host "检查子仓库..." -ForegroundColor Yellow
$missingRepos = @()
foreach ($repo in $repositories) {
    $repoPath = Join-Path $parentDir $repo
    if (-not (Test-Path $repoPath)) {
        $missingRepos += $repo
        Write-Warning "仓库不存在: $repo"
    } else {
        Write-Host "✓ $repo" -ForegroundColor Green
    }
}

if ($missingRepos.Count -gt 0 -and -not $SkipClone) {
    Write-Host "发现缺失的仓库，需要手动创建或克隆这些仓库:" -ForegroundColor Red
    foreach ($repo in $missingRepos) {
        Write-Host "  - $repo" -ForegroundColor Red
    }
    Write-Host "请确保所有子仓库都存在后重新运行此脚本。" -ForegroundColor Yellow
}

# 同步 Go 工作区
Write-Host "同步 Go 工作区..." -ForegroundColor Yellow
Set-Location $workspaceRoot
try {
    go work sync
    Write-Host "✓ Go 工作区同步完成" -ForegroundColor Green
} catch {
    Write-Warning "Go 工作区同步失败: $_"
}

# 下载依赖
Write-Host "下载所有模块依赖..." -ForegroundColor Yellow
foreach ($repo in $repositories) {
    $repoPath = Join-Path $parentDir $repo
    if (Test-Path $repoPath) {
        $goModPath = Join-Path $repoPath "go.mod"
        if (Test-Path $goModPath) {
            Write-Host "下载 $repo 的依赖..." -ForegroundColor Cyan
            Set-Location $repoPath
            try {
                go mod download
                Write-Host "✓ $repo 依赖下载完成" -ForegroundColor Green
            } catch {
                Write-Warning "$repo 依赖下载失败: $_"
            }
        }
    }
}

Set-Location $workspaceRoot

Write-Host "=== 环境设置完成 ===" -ForegroundColor Green
Write-Host "下一步:" -ForegroundColor Yellow
Write-Host "  1. 运行 .\scripts\build-all.ps1 构建所有模块" -ForegroundColor White
Write-Host "  2. 运行 .\scripts\test-all.ps1 运行测试" -ForegroundColor White