# 太上老君系统 - 统一启动脚本

param(
    [ValidateSet("docker", "local", "help")]
    [string]$Environment = "help",
    [string]$Profile = "basic",
    [string]$Service = "all",
    [switch]$Build = $false,
    [switch]$Stop = $false,
    [switch]$Logs = $false
)

$ErrorActionPreference = "Stop"

# 设置工作目录
Set-Location $PSScriptRoot

function Show-Help {
    Write-Host "太上老君系统 - 统一启动脚本" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "用法:" -ForegroundColor Yellow
    Write-Host "  .\start.ps1 -Environment <docker|local> [选项]" -ForegroundColor White
    Write-Host ""
    Write-Host "环境选项:" -ForegroundColor Yellow
    Write-Host "  docker    启动Docker环境" -ForegroundColor White
    Write-Host "  local     启动本地环境" -ForegroundColor White
    Write-Host "  help      显示此帮助信息" -ForegroundColor White
    Write-Host ""
    Write-Host "Docker环境选项:" -ForegroundColor Yellow
    Write-Host "  -Profile <basic|dev-tools|all>  Docker配置文件 (默认: basic)" -ForegroundColor White
    Write-Host "  -Build                          重新构建镜像" -ForegroundColor White
    Write-Host "  -Stop                           停止服务" -ForegroundColor White
    Write-Host "  -Logs                           查看日志" -ForegroundColor White
    Write-Host ""
    Write-Host "本地环境选项:" -ForegroundColor Yellow
    Write-Host "  -Service <all|config-center|admin-api|marketplace-api>  启动的服务 (默认: all)" -ForegroundColor White
    Write-Host "  -Build                          重新构建服务" -ForegroundColor White
    Write-Host "  -Stop                           停止服务" -ForegroundColor White
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Yellow
    Write-Host "  .\start.ps1 -Environment docker -Profile dev-tools" -ForegroundColor Green
    Write-Host "  .\start.ps1 -Environment local -Service all -Build" -ForegroundColor Green
    Write-Host "  .\start.ps1 -Environment docker -Stop" -ForegroundColor Green
    Write-Host ""
    Write-Host "其他脚本:" -ForegroundColor Yellow
    Write-Host "  .\test.ps1                      运行测试" -ForegroundColor White
    Write-Host "  .\scripts\testing\verify-config.ps1  验证配置" -ForegroundColor White
}

switch ($Environment.ToLower()) {
    "docker" {
        Write-Host "启动Docker环境..." -ForegroundColor Cyan
        $params = @()
        if ($Profile) { $params += "-Profile", $Profile }
        if ($Build) { $params += "-Build" }
        if ($Stop) { $params += "-Stop" }
        if ($Logs) { $params += "-Logs" }
        
        & ".\scripts\deployment\start-docker.ps1" @params
    }
    "local" {
        Write-Host "启动本地环境..." -ForegroundColor Cyan
        $params = @()
        if ($Service) { $params += "-Service", $Service }
        if ($Build) { $params += "-Build" }
        if ($Stop) { $params += "-Stop" }
        
        & ".\scripts\deployment\start-local.ps1" @params
    }
    "help" {
        Show-Help
    }
    default {
        Write-Host "错误: 未知环境 '$Environment'" -ForegroundColor Red
        Write-Host ""
        Show-Help
        exit 1
    }
}