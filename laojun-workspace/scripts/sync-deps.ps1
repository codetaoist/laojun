# Laojun 依赖同步脚本
# 用于同步所有模块的依赖和版本

param(
    [switch]$UpdateAll,
    [switch]$Verbose,
    [string]$Target = "all"
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 依赖同步 ===" -ForegroundColor Green

# 工作区根目录
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# Go 模块列表（按依赖顺序）
$goModules = @(
    @{
        Name = "laojun-shared"
        Path = "laojun-shared"
        Description = "共享组件库"
        IsShared = $true
    },
    @{
        Name = "laojun-plugins"
        Path = "laojun-plugins"
        Description = "插件系统"
        Dependencies = @("laojun-shared")
    },
    @{
        Name = "laojun-config-center"
        Path = "laojun-config-center"
        Description = "配置中心"
        Dependencies = @("laojun-shared")
    },
    @{
        Name = "laojun-admin-api"
        Path = "laojun-admin-api"
        Description = "管理后端 API"
        Dependencies = @("laojun-shared", "laojun-plugins")
    },
    @{
        Name = "laojun-marketplace-api"
        Path = "laojun-marketplace-api"
        Description = "市场后端 API"
        Dependencies = @("laojun-shared", "laojun-plugins")
    }
)

# 同步统计
$syncStats = @{
    Success = 0
    Failed = 0
    Skipped = 0
}

function Sync-ModuleDeps {
    param($module)
    
    $modulePath = Join-Path $parentDir $module.Path
    if (-not (Test-Path $modulePath)) {
        Write-Warning "模块路径不存在: $($module.Path)"
        $syncStats.Skipped++
        return $false
    }
    
    $goModPath = Join-Path $modulePath "go.mod"
    if (-not (Test-Path $goModPath)) {
        Write-Warning "不是 Go 模块: $($module.Path)"
        $syncStats.Skipped++
        return $false
    }
    
    Write-Host "同步 $($module.Name) - $($module.Description)..." -ForegroundColor Cyan
    Set-Location $modulePath
    
    try {
        # 清理模块缓存
        Write-Host "  清理模块缓存..." -ForegroundColor Yellow
        go clean -modcache
        
        # 下载依赖
        Write-Host "  下载依赖..." -ForegroundColor Yellow
        go mod download
        
        # 整理依赖
        Write-Host "  整理依赖..." -ForegroundColor Yellow
        go mod tidy
        
        # 验证依赖
        Write-Host "  验证依赖..." -ForegroundColor Yellow
        go mod verify
        
        # 更新依赖（如果指定）
        if ($UpdateAll) {
            Write-Host "  更新所有依赖..." -ForegroundColor Yellow
            
            # 获取当前依赖列表
            $depsOutput = go list -m -u all
            $outdatedDeps = $depsOutput | Where-Object { $_ -match "\[.*\]" }
            
            if ($outdatedDeps) {
                Write-Host "    发现可更新的依赖:" -ForegroundColor Cyan
                foreach ($dep in $outdatedDeps) {
                    Write-Host "      $dep" -ForegroundColor Gray
                }
                
                # 更新依赖
                go get -u ./...
                go mod tidy
            } else {
                Write-Host "    所有依赖都是最新版本" -ForegroundColor Green
            }
        }
        
        # 检查本地依赖（laojun-* 模块）
        if ($module.Dependencies) {
            Write-Host "  检查本地依赖..." -ForegroundColor Yellow
            
            foreach ($dep in $module.Dependencies) {
                $depPath = Join-Path $parentDir $dep
                if (Test-Path $depPath) {
                    Write-Host "    ✓ $dep (本地)" -ForegroundColor Green
                } else {
                    Write-Warning "    ✗ $dep (缺失)"
                }
            }
        }
        
        # 显示模块信息
        if ($Verbose) {
            Write-Host "  模块信息:" -ForegroundColor Yellow
            $modInfo = go list -m
            Write-Host "    $modInfo" -ForegroundColor Gray
            
            Write-Host "  直接依赖:" -ForegroundColor Yellow
            $directDeps = go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all | Where-Object { $_ -ne "" }
            foreach ($dep in $directDeps) {
                if ($dep -ne $modInfo) {
                    Write-Host "    $dep" -ForegroundColor Gray
                }
            }
        }
        
        Write-Host "  ✓ 依赖同步完成" -ForegroundColor Green
        $syncStats.Success++
        return $true
        
    } catch {
        Write-Error "依赖同步失败: $_"
        $syncStats.Failed++
        return $false
    }
}

# 开始同步
Set-Location $workspaceRoot

# 首先同步工作区
Write-Host "同步 Go 工作区..." -ForegroundColor Yellow
try {
    go work sync
    Write-Host "✓ Go 工作区同步完成" -ForegroundColor Green
} catch {
    Write-Warning "Go 工作区同步失败: $_"
}

Write-Host ""

if ($Target -eq "all") {
    Write-Host "同步所有模块依赖..." -ForegroundColor Yellow
    
    foreach ($module in $goModules) {
        Sync-ModuleDeps $module
        Write-Host ""
    }
} else {
    # 同步指定模块
    $targetModule = $goModules | Where-Object { $_.Name -eq $Target }
    if ($targetModule) {
        Sync-ModuleDeps $targetModule
    } else {
        Write-Error "未找到模块: $Target"
        Write-Host "可用模块:" -ForegroundColor Yellow
        foreach ($module in $goModules) {
            Write-Host "  - $($module.Name)" -ForegroundColor White
        }
        exit 1
    }
}

# 同步总结
Write-Host "=== 依赖同步总结 ===" -ForegroundColor Green
Write-Host "成功: $($syncStats.Success)" -ForegroundColor Green
Write-Host "失败: $($syncStats.Failed)" -ForegroundColor Red
Write-Host "跳过: $($syncStats.Skipped)" -ForegroundColor Yellow

if ($syncStats.Failed -gt 0) {
    Write-Host "存在依赖同步失败的模块，请检查错误信息。" -ForegroundColor Red
    exit 1
} else {
    Write-Host "所有模块依赖同步成功！" -ForegroundColor Green
    Write-Host "建议运行 .\scripts\build-all.ps1 验证构建。" -ForegroundColor Yellow
}