# Laojun 清理脚本
# 用于清理所有模块的构建产物和缓存

param(
    [switch]$Deep,
    [switch]$Verbose,
    [string]$Target = "all"
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 清理 ===" -ForegroundColor Green

# 工作区根目录
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# 所有仓库列表
$repositories = @(
    @{
        Name = "laojun-shared"
        Path = "laojun-shared"
        Type = "go"
        Description = "共享组件库"
    },
    @{
        Name = "laojun-plugins"
        Path = "laojun-plugins"
        Type = "go"
        Description = "插件系统"
    },
    @{
        Name = "laojun-config-center"
        Path = "laojun-config-center"
        Type = "go"
        Description = "配置中心"
    },
    @{
        Name = "laojun-admin-api"
        Path = "laojun-admin-api"
        Type = "go"
        Description = "管理后端 API"
    },
    @{
        Name = "laojun-marketplace-api"
        Path = "laojun-marketplace-api"
        Type = "go"
        Description = "市场后端 API"
    },
    @{
        Name = "laojun-admin-web"
        Path = "laojun-admin-web"
        Type = "node"
        Description = "管理前端"
    },
    @{
        Name = "laojun-marketplace-web"
        Path = "laojun-marketplace-web"
        Type = "node"
        Description = "市场前端"
    }
)

# 清理统计
$cleanStats = @{
    Success = 0
    Failed = 0
    Skipped = 0
    TotalSize = 0
}

function Get-DirectorySize {
    param($path)
    if (Test-Path $path) {
        $size = (Get-ChildItem -Path $path -Recurse -File | Measure-Object -Property Length -Sum).Sum
        return if ($size) { $size } else { 0 }
    }
    return 0
}

function Format-FileSize {
    param($bytes)
    if ($bytes -ge 1GB) {
        return "{0:N2} GB" -f ($bytes / 1GB)
    } elseif ($bytes -ge 1MB) {
        return "{0:N2} MB" -f ($bytes / 1MB)
    } elseif ($bytes -ge 1KB) {
        return "{0:N2} KB" -f ($bytes / 1KB)
    } else {
        return "$bytes bytes"
    }
}

function Clean-Repository {
    param($repo)
    
    $repoPath = Join-Path $parentDir $repo.Path
    if (-not (Test-Path $repoPath)) {
        Write-Warning "仓库路径不存在: $($repo.Path)"
        $cleanStats.Skipped++
        return $false
    }
    
    Write-Host "清理 $($repo.Name) - $($repo.Description)..." -ForegroundColor Cyan
    Set-Location $repoPath
    
    try {
        $cleanedSize = 0
        
        if ($repo.Type -eq "go") {
            # Go 项目清理
            
            # 清理构建产物
            $binDir = "bin"
            if (Test-Path $binDir) {
                $size = Get-DirectorySize $binDir
                Write-Host "  清理构建产物 ($binDir)..." -ForegroundColor Yellow
                Remove-Item -Path $binDir -Recurse -Force
                $cleanedSize += $size
                if ($Verbose) {
                    Write-Host "    已清理: $(Format-FileSize $size)" -ForegroundColor Gray
                }
            }
            
            # 清理测试覆盖率文件
            $coverageFiles = Get-ChildItem -Name "*.out" -Recurse
            foreach ($file in $coverageFiles) {
                if (Test-Path $file) {
                    $size = (Get-Item $file).Length
                    Write-Host "  清理覆盖率文件 ($file)..." -ForegroundColor Yellow
                    Remove-Item -Path $file -Force
                    $cleanedSize += $size
                }
            }
            
            # Go 缓存清理
            if ($Deep) {
                Write-Host "  清理 Go 缓存..." -ForegroundColor Yellow
                go clean -cache -testcache -modcache
            } else {
                Write-Host "  清理 Go 构建缓存..." -ForegroundColor Yellow
                go clean -cache -testcache
            }
            
        } elseif ($repo.Type -eq "node") {
            # Node.js 项目清理
            
            # 清理 node_modules（深度清理时）
            if ($Deep) {
                $nodeModulesDir = "node_modules"
                if (Test-Path $nodeModulesDir) {
                    $size = Get-DirectorySize $nodeModulesDir
                    Write-Host "  清理 node_modules..." -ForegroundColor Yellow
                    Remove-Item -Path $nodeModulesDir -Recurse -Force
                    $cleanedSize += $size
                    if ($Verbose) {
                        Write-Host "    已清理: $(Format-FileSize $size)" -ForegroundColor Gray
                    }
                }
            }
            
            # 清理构建产物
            $distDir = "dist"
            if (Test-Path $distDir) {
                $size = Get-DirectorySize $distDir
                Write-Host "  清理构建产物 ($distDir)..." -ForegroundColor Yellow
                Remove-Item -Path $distDir -Recurse -Force
                $cleanedSize += $size
                if ($Verbose) {
                    Write-Host "    已清理: $(Format-FileSize $size)" -ForegroundColor Gray
                }
            }
            
            # 清理其他构建目录
            $buildDirs = @("build", ".next", ".nuxt")
            foreach ($dir in $buildDirs) {
                if (Test-Path $dir) {
                    $size = Get-DirectorySize $dir
                    Write-Host "  清理构建目录 ($dir)..." -ForegroundColor Yellow
                    Remove-Item -Path $dir -Recurse -Force
                    $cleanedSize += $size
                    if ($Verbose) {
                        Write-Host "    已清理: $(Format-FileSize $size)" -ForegroundColor Gray
                    }
                }
            }
        }
        
        # 清理通用临时文件
        $tempPatterns = @("*.tmp", "*.temp", "*.log", ".DS_Store", "Thumbs.db")
        foreach ($pattern in $tempPatterns) {
            $tempFiles = Get-ChildItem -Name $pattern -Recurse -ErrorAction SilentlyContinue
            foreach ($file in $tempFiles) {
                if (Test-Path $file) {
                    $size = (Get-Item $file).Length
                    Write-Host "  清理临时文件 ($file)..." -ForegroundColor Yellow
                    Remove-Item -Path $file -Force
                    $cleanedSize += $size
                }
            }
        }
        
        $cleanStats.TotalSize += $cleanedSize
        Write-Host "  ✓ 清理完成 ($(Format-FileSize $cleanedSize))" -ForegroundColor Green
        $cleanStats.Success++
        return $true
        
    } catch {
        Write-Error "清理失败: $_"
        $cleanStats.Failed++
        return $false
    }
}

# 开始清理
Set-Location $workspaceRoot

if ($Target -eq "all") {
    Write-Host "清理所有仓库..." -ForegroundColor Yellow
    if ($Deep) {
        Write-Host "执行深度清理（包括依赖缓存）..." -ForegroundColor Yellow
    }
    Write-Host ""
    
    foreach ($repo in $repositories) {
        Clean-Repository $repo
        Write-Host ""
    }
    
    # 清理工作区临时文件
    Write-Host "清理工作区..." -ForegroundColor Cyan
    $workspaceTempFiles = Get-ChildItem -Path $workspaceRoot -Name "*.tmp" -ErrorAction SilentlyContinue
    foreach ($file in $workspaceTempFiles) {
        Remove-Item -Path $file -Force
        Write-Host "  清理: $file" -ForegroundColor Yellow
    }
    
} else {
    # 清理指定仓库
    $targetRepo = $repositories | Where-Object { $_.Name -eq $Target }
    if ($targetRepo) {
        Clean-Repository $targetRepo
    } else {
        Write-Error "未找到仓库: $Target"
        Write-Host "可用仓库:" -ForegroundColor Yellow
        foreach ($repo in $repositories) {
            Write-Host "  - $($repo.Name)" -ForegroundColor White
        }
        exit 1
    }
}

# 清理总结
Write-Host "=== 清理总结 ===" -ForegroundColor Green
Write-Host "成功: $($cleanStats.Success)" -ForegroundColor Green
Write-Host "失败: $($cleanStats.Failed)" -ForegroundColor Red
Write-Host "跳过: $($cleanStats.Skipped)" -ForegroundColor Yellow
Write-Host "总清理大小: $(Format-FileSize $cleanStats.TotalSize)" -ForegroundColor Cyan

if ($cleanStats.Failed -gt 0) {
    Write-Host "存在清理失败的仓库，请检查错误信息。" -ForegroundColor Red
    exit 1
} else {
    Write-Host "清理完成！" -ForegroundColor Green
    if ($Deep) {
        Write-Host "建议运行 .\scripts\sync-deps.ps1 重新下载依赖。" -ForegroundColor Yellow
    }
}