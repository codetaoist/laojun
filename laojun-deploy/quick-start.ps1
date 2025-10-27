#!/usr/bin/env pwsh
<#
.SYNOPSIS
    太上老君微服务平台 - 快速启动脚本

.DESCRIPTION
    一键启动本地开发环境，自动选择最佳部署方式

.PARAMETER Platform
    强制指定平台，默认自动检测

.PARAMETER Clean
    清理现有环境后重新部署

.EXAMPLE
    .\quick-start.ps1
    .\quick-start.ps1 -Platform docker
    .\quick-start.ps1 -Clean

.NOTES
    Author: TaiShang LaoJun Team
    Version: 1.0.0
#>

param(
    [Parameter(Mandatory = $false)]
    [ValidateSet("docker", "kubernetes", "helm", "auto")]
    [string]$Platform = "auto",

    [Parameter(Mandatory = $false)]
    [switch]$Clean,

    [Parameter(Mandatory = $false)]
    [switch]$Verbose
)

# 颜色输出函数
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    $colors = @{
        "Red" = [ConsoleColor]::Red; "Green" = [ConsoleColor]::Green
        "Yellow" = [ConsoleColor]::Yellow; "Blue" = [ConsoleColor]::Blue
        "Cyan" = [ConsoleColor]::Cyan; "White" = [ConsoleColor]::White
    }
    Write-Host $Message -ForegroundColor $colors[$Color]
}

function Write-Header {
    Write-ColorOutput "`n" + "🚀 "*20 "Cyan"
    Write-ColorOutput "   太上老君微服务平台 - 快速启动" "Cyan"
    Write-ColorOutput "🚀 "*20 + "`n" "Cyan"
}

function Write-Step {
    param([string]$Message)
    Write-ColorOutput "✨ $Message" "Green"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ️  $Message" "Blue"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" "Yellow"
}

# 自动检测最佳平台
function Get-BestPlatform {
    Write-Step "自动检测最佳部署平台..."
    
    # 检查 Docker
    try {
        $null = Get-Command docker -ErrorAction Stop
        $null = Get-Command docker-compose -ErrorAction Stop
        $dockerVersion = docker --version
        Write-Info "检测到 Docker: $dockerVersion"
        return "docker"
    }
    catch {
        Write-Warning "Docker 未安装或不可用"
    }
    
    # 检查 Kubernetes
    try {
        $null = Get-Command kubectl -ErrorAction Stop
        $k8sVersion = kubectl version --client --short 2>$null
        Write-Info "检测到 Kubernetes: $k8sVersion"
        return "kubernetes"
    }
    catch {
        Write-Warning "Kubernetes 未安装或不可用"
    }
    
    Write-Warning "未检测到可用的部署平台，请安装 Docker 或 Kubernetes"
    return $null
}

# 环境检查
function Test-Environment {
    Write-Step "检查运行环境..."
    
    # 检查 PowerShell 版本
    $psVersion = $PSVersionTable.PSVersion
    Write-Info "PowerShell 版本: $psVersion"
    
    # 检查操作系统
    $os = [System.Environment]::OSVersion.Platform
    Write-Info "操作系统: $os"
    
    # 检查网络连接
    try {
        $null = Test-NetConnection -ComputerName "github.com" -Port 443 -InformationLevel Quiet
        Write-Info "网络连接正常"
    }
    catch {
        Write-Warning "网络连接可能有问题"
    }
    
    return $true
}

# 清理环境
function Clear-Environment {
    param([string]$Platform)
    
    Write-Step "清理现有环境..."
    
    try {
        switch ($Platform) {
            "docker" {
                Write-Info "停止 Docker 容器..."
                & .\deploy-unified.ps1 -Platform docker -Environment local -Action cleanup -Force
            }
            "kubernetes" {
                Write-Info "清理 Kubernetes 资源..."
                & .\deploy-unified.ps1 -Platform kubernetes -Environment local -Action stop -Force
            }
            "helm" {
                Write-Info "卸载 Helm 发布..."
                & .\deploy-unified.ps1 -Platform helm -Environment local -Action stop -Force
            }
        }
        Write-ColorOutput "✅ 环境清理完成" "Green"
    }
    catch {
        Write-Warning "清理过程中出现警告: $($_.Exception.Message)"
    }
}

# 部署应用
function Deploy-Application {
    param([string]$Platform)
    
    Write-Step "开始部署太上老君微服务平台..."
    
    try {
        & .\deploy-unified.ps1 -Platform $Platform -Environment local -Action deploy -Verbose:$Verbose
        return $true
    }
    catch {
        Write-ColorOutput "❌ 部署失败: $($_.Exception.Message)" "Red"
        return $false
    }
}

# 显示访问信息
function Show-AccessInfo {
    Write-ColorOutput "`n🎉 部署成功！" "Green"
    Write-ColorOutput "="*50 "Cyan"
    Write-ColorOutput "📱 访问地址:" "Cyan"
    Write-ColorOutput "   🏠 插件市场（主页）: http://localhost" "White"
    Write-ColorOutput "   ⚙️  管理后台: http://localhost:8888" "White"
    Write-ColorOutput "   📚 API文档: http://localhost:8080/swagger" "White"
    Write-ColorOutput "   📊 监控面板: http://localhost:9090" "White"
    Write-ColorOutput "   📈 Grafana: http://localhost:3000 (admin/admin123)" "White"
    Write-ColorOutput "="*50 "Cyan"
    Write-ColorOutput "`n💡 提示:" "Yellow"
    Write-ColorOutput "   • 首次启动可能需要几分钟来下载镜像" "White"
    Write-ColorOutput "   • 使用 Ctrl+C 停止服务" "White"
    Write-ColorOutput "   • 查看日志: .\deploy-unified.ps1 -Platform $Platform -Environment local -Action logs" "White"
    Write-ColorOutput "   • 重启服务: .\deploy-unified.ps1 -Platform $Platform -Environment local -Action restart" "White"
}

# 主函数
function Main {
    Write-Header
    
    # 环境检查
    if (-not (Test-Environment)) {
        Write-ColorOutput "❌ 环境检查失败" "Red"
        exit 1
    }
    
    # 确定部署平台
    if ($Platform -eq "auto") {
        $Platform = Get-BestPlatform
        if (-not $Platform) {
            Write-ColorOutput "❌ 无法确定部署平台" "Red"
            exit 1
        }
    }
    
    Write-Info "使用部署平台: $Platform"
    
    # 清理环境（如果需要）
    if ($Clean) {
        Clear-Environment -Platform $Platform
    }
    
    # 部署应用
    if (Deploy-Application -Platform $Platform) {
        Show-AccessInfo
        
        # 等待用户输入
        Write-ColorOutput "`n按任意键查看服务状态..." "Yellow"
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        
        # 显示服务状态
        Write-Step "检查服务状态..."
        & .\deploy-unified.ps1 -Platform $Platform -Environment local -Action status
    }
    else {
        Write-ColorOutput "❌ 快速启动失败" "Red"
        exit 1
    }
}

# 执行主函数
Main