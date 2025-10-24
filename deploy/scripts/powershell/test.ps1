# 太上老君系统 - 统一测试脚本

param(
    [ValidateSet("config", "services", "all", "help")]
    [string]$Type = "help",
    [switch]$Detailed = $false
)

$ErrorActionPreference = "Stop"

# 设置工作目录
Set-Location $PSScriptRoot

function Show-Help {
    Write-Host "太上老君系统 - 统一测试脚本" -ForegroundColor Cyan
    Write-Host "==============================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "用法:" -ForegroundColor Yellow
    Write-Host "  .\test.ps1 -Type <config|services|all> [选项]" -ForegroundColor White
    Write-Host ""
    Write-Host "测试类型:" -ForegroundColor Yellow
    Write-Host "  config     验证配置文件" -ForegroundColor White
    Write-Host "  services   测试所有服务" -ForegroundColor White
    Write-Host "  all        运行所有测试" -ForegroundColor White
    Write-Host "  help       显示此帮助信息" -ForegroundColor White
    Write-Host ""
    Write-Host "选项:" -ForegroundColor Yellow
    Write-Host "  -Detailed  显示详细信息" -ForegroundColor White
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Yellow
    Write-Host "  .\test.ps1 -Type config" -ForegroundColor Green
    Write-Host "  .\test.ps1 -Type services" -ForegroundColor Green
    Write-Host "  .\test.ps1 -Type all -Detailed" -ForegroundColor Green
    Write-Host ""
    Write-Host "直接运行测试脚本:" -ForegroundColor Yellow
    Write-Host "  .\scripts\testing\verify-config.ps1" -ForegroundColor White
    Write-Host "  .\scripts\testing\test-all-services.ps1" -ForegroundColor White
}

switch ($Type.ToLower()) {
    "config" {
        Write-Host "验证配置文件..." -ForegroundColor Cyan
        $params = @()
        if ($Detailed) { $params += "-Detailed" }
        
        & ".\scripts\testing\verify-config.ps1" @params
    }
    "services" {
        Write-Host "测试所有服务..." -ForegroundColor Cyan
        & ".\scripts\testing\test-all-services.ps1"
    }
    "all" {
        Write-Host "运行所有测试..." -ForegroundColor Cyan
        Write-Host ""
        
        Write-Host "1. 验证配置文件..." -ForegroundColor Yellow
        $params = @()
        if ($Detailed) { $params += "-Detailed" }
        & ".\scripts\testing\verify-config.ps1" @params
        
        Write-Host ""
        Write-Host "2. 测试所有服务..." -ForegroundColor Yellow
        & ".\scripts\testing\test-all-services.ps1"
    }
    "help" {
        Show-Help
    }
    default {
        Write-Host "错误: 未知测试类型 '$Type'" -ForegroundColor Red
        Write-Host ""
        Show-Help
        exit 1
    }
}