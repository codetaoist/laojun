# Laojun 批量构建脚本
# 用于构建所有 Go 模块

param(
    [switch]$Clean,
    [switch]$Verbose,
    [string]$Target = "all"
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 批量构建 ===" -ForegroundColor Green

# 工作区根目录
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# Go 模块列表（按依赖顺序）
$goModules = @(
    @{
        Name = "laojun-shared"
        Path = "laojun-shared"
        HasBinary = $false
        Description = "共享组件库"
    },
    @{
        Name = "laojun-plugins"
        Path = "laojun-plugins" 
        HasBinary = $true
        BinaryPath = "cmd/plugin-manager"
        Description = "插件系统"
    },
    @{
        Name = "laojun-config-center"
        Path = "laojun-config-center"
        HasBinary = $true
        BinaryPath = "cmd/config-center"
        Description = "配置中心"
    },
    @{
        Name = "laojun-admin-api"
        Path = "laojun-admin-api"
        HasBinary = $true
        BinaryPath = "cmd/admin-api"
        Description = "管理后端 API"
    },
    @{
        Name = "laojun-marketplace-api"
        Path = "laojun-marketplace-api"
        HasBinary = $true
        BinaryPath = "cmd/marketplace-api"
        Description = "市场后端 API"
    }
)

# 构建统计
$buildStats = @{
    Success = 0
    Failed = 0
    Skipped = 0
}

function Build-Module {
    param($module)
    
    $modulePath = Join-Path $parentDir $module.Path
    if (-not (Test-Path $modulePath)) {
        Write-Warning "模块路径不存在: $($module.Path)"
        $buildStats.Skipped++
        return $false
    }
    
    $goModPath = Join-Path $modulePath "go.mod"
    if (-not (Test-Path $goModPath)) {
        Write-Warning "不是 Go 模块: $($module.Path)"
        $buildStats.Skipped++
        return $false
    }
    
    Write-Host "构建 $($module.Name) - $($module.Description)..." -ForegroundColor Cyan
    Set-Location $modulePath
    
    try {
        # 清理构建缓存
        if ($Clean) {
            Write-Host "  清理构建缓存..." -ForegroundColor Yellow
            go clean -cache -modcache -testcache
        }
        
        # 下载依赖
        Write-Host "  下载依赖..." -ForegroundColor Yellow
        go mod download
        go mod tidy
        
        # 构建
        if ($module.HasBinary -and $module.BinaryPath) {
            Write-Host "  构建二进制文件..." -ForegroundColor Yellow
            $binaryName = Split-Path -Leaf $module.BinaryPath
            $outputPath = "bin\$binaryName.exe"
            
            # 创建 bin 目录
            $binDir = "bin"
            if (-not (Test-Path $binDir)) {
                New-Item -ItemType Directory -Path $binDir -Force | Out-Null
            }
            
            $buildCmd = "go build -o $outputPath ./$($module.BinaryPath)"
            if ($Verbose) {
                Write-Host "    执行: $buildCmd" -ForegroundColor Gray
            }
            
            Invoke-Expression $buildCmd
            
            if (Test-Path $outputPath) {
                $fileInfo = Get-Item $outputPath
                Write-Host "  ✓ 构建成功: $outputPath ($([math]::Round($fileInfo.Length / 1MB, 2)) MB)" -ForegroundColor Green
            } else {
                throw "二进制文件未生成"
            }
        } else {
            # 只验证编译
            Write-Host "  验证编译..." -ForegroundColor Yellow
            go build ./...
            Write-Host "  ✓ 编译验证成功" -ForegroundColor Green
        }
        
        $buildStats.Success++
        return $true
        
    } catch {
        Write-Error "构建失败: $_"
        $buildStats.Failed++
        return $false
    }
}

# 开始构建
Set-Location $workspaceRoot

if ($Target -eq "all") {
    Write-Host "构建所有模块..." -ForegroundColor Yellow
    
    foreach ($module in $goModules) {
        Build-Module $module
        Write-Host ""
    }
} else {
    # 构建指定模块
    $targetModule = $goModules | Where-Object { $_.Name -eq $Target }
    if ($targetModule) {
        Build-Module $targetModule
    } else {
        Write-Error "未找到模块: $Target"
        Write-Host "可用模块:" -ForegroundColor Yellow
        foreach ($module in $goModules) {
            Write-Host "  - $($module.Name)" -ForegroundColor White
        }
        exit 1
    }
}

# 构建总结
Write-Host "=== 构建总结 ===" -ForegroundColor Green
Write-Host "成功: $($buildStats.Success)" -ForegroundColor Green
Write-Host "失败: $($buildStats.Failed)" -ForegroundColor Red
Write-Host "跳过: $($buildStats.Skipped)" -ForegroundColor Yellow

if ($buildStats.Failed -gt 0) {
    Write-Host "存在构建失败的模块，请检查错误信息。" -ForegroundColor Red
    exit 1
} else {
    Write-Host "所有模块构建成功！" -ForegroundColor Green
}