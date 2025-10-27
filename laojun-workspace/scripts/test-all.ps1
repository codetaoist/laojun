# Laojun 批量测试脚本
# 用于运行所有模块的测试

param(
    [switch]$Coverage,
    [switch]$Verbose,
    [switch]$Short,
    [string]$Target = "all",
    [string]$TestPattern = ""
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 批量测试 ===" -ForegroundColor Green

# 工作区根目录
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# Go 模块列表
$goModules = @(
    @{
        Name = "laojun-shared"
        Path = "laojun-shared"
        Description = "共享组件库"
    },
    @{
        Name = "laojun-plugins"
        Path = "laojun-plugins"
        Description = "插件系统"
    },
    @{
        Name = "laojun-config-center"
        Path = "laojun-config-center"
        Description = "配置中心"
    },
    @{
        Name = "laojun-admin-api"
        Path = "laojun-admin-api"
        Description = "管理后端 API"
    },
    @{
        Name = "laojun-marketplace-api"
        Path = "laojun-marketplace-api"
        Description = "市场后端 API"
    }
)

# 测试统计
$testStats = @{
    TotalModules = 0
    PassedModules = 0
    FailedModules = 0
    SkippedModules = 0
    TotalTests = 0
    PassedTests = 0
    FailedTests = 0
    Coverage = @{}
}

function Test-Module {
    param($module)
    
    $modulePath = Join-Path $parentDir $module.Path
    if (-not (Test-Path $modulePath)) {
        Write-Warning "模块路径不存在: $($module.Path)"
        $testStats.SkippedModules++
        return $false
    }
    
    $goModPath = Join-Path $modulePath "go.mod"
    if (-not (Test-Path $goModPath)) {
        Write-Warning "不是 Go 模块: $($module.Path)"
        $testStats.SkippedModules++
        return $false
    }
    
    Write-Host "测试 $($module.Name) - $($module.Description)..." -ForegroundColor Cyan
    Set-Location $modulePath
    
    try {
        $testStats.TotalModules++
        
        # 构建测试命令
        $testCmd = "go test"
        
        if ($Verbose) {
            $testCmd += " -v"
        }
        
        if ($Short) {
            $testCmd += " -short"
        }
        
        if ($Coverage) {
            $coverageFile = "coverage.out"
            $testCmd += " -coverprofile=$coverageFile"
        }
        
        if ($TestPattern) {
            $testCmd += " -run `"$TestPattern`""
        }
        
        $testCmd += " ./..."
        
        if ($Verbose) {
            Write-Host "    执行: $testCmd" -ForegroundColor Gray
        }
        
        # 执行测试
        $output = Invoke-Expression $testCmd 2>&1
        
        # 解析测试结果
        $testResult = $output | Out-String
        
        # 提取测试统计信息
        if ($testResult -match "PASS") {
            Write-Host "  ✓ 测试通过" -ForegroundColor Green
            $testStats.PassedModules++
            
            # 提取测试数量
            if ($testResult -match "(\d+) tests") {
                $moduleTests = [int]$matches[1]
                $testStats.TotalTests += $moduleTests
                $testStats.PassedTests += $moduleTests
                Write-Host "    测试数量: $moduleTests" -ForegroundColor Gray
            }
            
        } elseif ($testResult -match "FAIL") {
            Write-Host "  ✗ 测试失败" -ForegroundColor Red
            $testStats.FailedModules++
            
            # 显示失败详情
            if ($Verbose) {
                Write-Host "失败详情:" -ForegroundColor Red
                Write-Host $testResult -ForegroundColor Red
            }
            
        } else {
            Write-Host "  - 无测试文件" -ForegroundColor Yellow
        }
        
        # 处理覆盖率
        if ($Coverage) {
            $coverageFile = "coverage.out"
            if (Test-Path $coverageFile) {
                try {
                    $coverageOutput = go tool cover -func=$coverageFile
                    $coverageLines = $coverageOutput -split "`n"
                    $totalLine = $coverageLines | Where-Object { $_ -match "total:" }
                    
                    if ($totalLine -match "(\d+\.\d+)%") {
                        $coveragePercent = [double]$matches[1]
                        $testStats.Coverage[$module.Name] = $coveragePercent
                        Write-Host "    覆盖率: $coveragePercent%" -ForegroundColor Cyan
                    }
                } catch {
                    Write-Warning "无法计算覆盖率: $_"
                }
            }
        }
        
        return $true
        
    } catch {
        Write-Error "测试执行失败: $_"
        $testStats.FailedModules++
        return $false
    }
}

# 开始测试
Set-Location $workspaceRoot

if ($Target -eq "all") {
    Write-Host "测试所有模块..." -ForegroundColor Yellow
    
    foreach ($module in $goModules) {
        Test-Module $module
        Write-Host ""
    }
} else {
    # 测试指定模块
    $targetModule = $goModules | Where-Object { $_.Name -eq $Target }
    if ($targetModule) {
        Test-Module $targetModule
    } else {
        Write-Error "未找到模块: $Target"
        Write-Host "可用模块:" -ForegroundColor Yellow
        foreach ($module in $goModules) {
            Write-Host "  - $($module.Name)" -ForegroundColor White
        }
        exit 1
    }
}

# 测试总结
Write-Host "=== 测试总结 ===" -ForegroundColor Green
Write-Host "模块统计:" -ForegroundColor Yellow
Write-Host "  总计: $($testStats.TotalModules)" -ForegroundColor White
Write-Host "  通过: $($testStats.PassedModules)" -ForegroundColor Green
Write-Host "  失败: $($testStats.FailedModules)" -ForegroundColor Red
Write-Host "  跳过: $($testStats.SkippedModules)" -ForegroundColor Yellow

if ($testStats.TotalTests -gt 0) {
    Write-Host "测试统计:" -ForegroundColor Yellow
    Write-Host "  总测试数: $($testStats.TotalTests)" -ForegroundColor White
    Write-Host "  通过测试: $($testStats.PassedTests)" -ForegroundColor Green
    Write-Host "  失败测试: $($testStats.FailedTests)" -ForegroundColor Red
}

if ($Coverage -and $testStats.Coverage.Count -gt 0) {
    Write-Host "覆盖率统计:" -ForegroundColor Yellow
    foreach ($module in $testStats.Coverage.Keys) {
        $coverage = $testStats.Coverage[$module]
        $color = if ($coverage -ge 80) { "Green" } elseif ($coverage -ge 60) { "Yellow" } else { "Red" }
        Write-Host "  $module`: $coverage%" -ForegroundColor $color
    }
    
    $avgCoverage = ($testStats.Coverage.Values | Measure-Object -Average).Average
    Write-Host "  平均覆盖率: $([math]::Round($avgCoverage, 2))%" -ForegroundColor Cyan
}

if ($testStats.FailedModules -gt 0) {
    Write-Host "存在测试失败的模块，请检查错误信息。" -ForegroundColor Red
    exit 1
} else {
    Write-Host "所有测试通过！" -ForegroundColor Green
}