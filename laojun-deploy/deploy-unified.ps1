#!/usr/bin/env pwsh
<#
.SYNOPSIS
    太上老君微服务平台 - 统一部署脚本

.DESCRIPTION
    支持多平台（Docker、Kubernetes、Helm）和多环境（local、dev、staging、production）的智能部署脚本

.PARAMETER Platform
    部署平台: docker, kubernetes, helm

.PARAMETER Environment
    部署环境: local, dev, staging, production

.PARAMETER Action
    操作类型: deploy, start, stop, restart, status, logs

.PARAMETER Services
    指定要部署的服务，默认部署所有服务

.PARAMETER DryRun
    预览模式，只显示将要执行的命令，不实际执行

.EXAMPLE
    .\deploy-unified.ps1 -Platform docker -Environment local -Action deploy
    .\deploy-unified.ps1 -Platform kubernetes -Environment production -Action deploy
    .\deploy-unified.ps1 -Platform helm -Environment staging -Action restart
    .\deploy-unified.ps1 -Platform docker -Environment dev -Action logs -Services "admin-api,gateway"

.NOTES
    Author: TaiShang LaoJun Team
    Version: 2.0.0
    Created: 2024-10-25
#>

param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("docker", "kubernetes", "helm")]
    [string]$Platform,

    [Parameter(Mandatory = $true)]
    [ValidateSet("local", "dev", "staging", "production")]
    [string]$Environment,

    [Parameter(Mandatory = $false)]
    [ValidateSet("deploy", "start", "stop", "restart", "status", "logs", "cleanup")]
    [string]$Action = "deploy",

    [Parameter(Mandatory = $false)]
    [string]$Services = "",

    [Parameter(Mandatory = $false)]
    [switch]$DryRun,

    [Parameter(Mandatory = $false)]
    [switch]$Verbose,

    [Parameter(Mandatory = $false)]
    [switch]$Force
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $colors = @{
        "Red" = [ConsoleColor]::Red
        "Green" = [ConsoleColor]::Green
        "Yellow" = [ConsoleColor]::Yellow
        "Blue" = [ConsoleColor]::Blue
        "Cyan" = [ConsoleColor]::Cyan
        "Magenta" = [ConsoleColor]::Magenta
        "White" = [ConsoleColor]::White
    }
    
    Write-Host $Message -ForegroundColor $colors[$Color]
}

function Write-Header {
    param([string]$Title)
    Write-ColorOutput "`n" + "="*80 "Cyan"
    Write-ColorOutput "  $Title" "Cyan"
    Write-ColorOutput "="*80 "Cyan"
}

function Write-Step {
    param([string]$Message)
    Write-ColorOutput "🚀 $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" "Red"
}

# 检查依赖
function Test-Dependencies {
    Write-Step "检查系统依赖..."
    
    $dependencies = @()
    
    switch ($Platform) {
        "docker" {
            $dependencies += @("docker", "docker-compose")
        }
        "kubernetes" {
            $dependencies += @("kubectl")
        }
        "helm" {
            $dependencies += @("helm", "kubectl")
        }
    }
    
    foreach ($dep in $dependencies) {
        try {
            $null = Get-Command $dep -ErrorAction Stop
            Write-ColorOutput "✅ $dep 已安装" "Green"
        }
        catch {
            Write-Error "$dep 未安装或不在 PATH 中"
            return $false
        }
    }
    
    return $true
}

# 加载环境配置
function Get-EnvironmentConfig {
    param([string]$EnvName)
    
    $configPath = ".\environments\$EnvName"
    if (-not (Test-Path $configPath)) {
        Write-Warning "环境配置目录不存在: $configPath"
        return @{}
    }
    
    $config = @{
        "ConfigPath" = $configPath
        "EnvFile" = "$configPath\.env"
        "ConfigFile" = "$configPath\config.yaml"
    }
    
    return $config
}

# Docker 部署函数
function Invoke-DockerDeploy {
    param(
        [string]$EnvConfig,
        [string]$Action,
        [string]$Services
    )
    
    $platformPath = ".\platforms\docker"
    Push-Location $platformPath
    
    try {
        $composeFile = "docker-compose.yml"
        $envComposeFile = "docker-compose.$Environment.yml"
        
        $composeArgs = @("-f", $composeFile)
        if (Test-Path $envComposeFile) {
            $composeArgs += @("-f", $envComposeFile)
        }
        
        # 设置环境变量文件
        if (Test-Path $EnvConfig.EnvFile) {
            $composeArgs += @("--env-file", $EnvConfig.EnvFile)
        }
        
        switch ($Action) {
            "deploy" {
                Write-Step "部署 Docker 服务..."
                if ($DryRun) {
                    Write-ColorOutput "DRY RUN: docker-compose $($composeArgs -join ' ') up -d $Services" "Yellow"
                } else {
                    & docker-compose @composeArgs up -d $Services.Split(',')
                }
            }
            "start" {
                Write-Step "启动 Docker 服务..."
                if ($DryRun) {
                    Write-ColorOutput "DRY RUN: docker-compose $($composeArgs -join ' ') start $Services" "Yellow"
                } else {
                    & docker-compose @composeArgs start $Services.Split(',')
                }
            }
            "stop" {
                Write-Step "停止 Docker 服务..."
                if ($DryRun) {
                    Write-ColorOutput "DRY RUN: docker-compose $($composeArgs -join ' ') stop $Services" "Yellow"
                } else {
                    & docker-compose @composeArgs stop $Services.Split(',')
                }
            }
            "restart" {
                Write-Step "重启 Docker 服务..."
                if ($DryRun) {
                    Write-ColorOutput "DRY RUN: docker-compose $($composeArgs -join ' ') restart $Services" "Yellow"
                } else {
                    & docker-compose @composeArgs restart $Services.Split(',')
                }
            }
            "status" {
                Write-Step "查看 Docker 服务状态..."
                & docker-compose @composeArgs ps
            }
            "logs" {
                Write-Step "查看 Docker 服务日志..."
                if ($Services) {
                    & docker-compose @composeArgs logs -f $Services.Split(',')
                } else {
                    & docker-compose @composeArgs logs -f
                }
            }
            "cleanup" {
                Write-Step "清理 Docker 资源..."
                if ($DryRun) {
                    Write-ColorOutput "DRY RUN: docker-compose $($composeArgs -join ' ') down -v --remove-orphans" "Yellow"
                } else {
                    & docker-compose @composeArgs down -v --remove-orphans
                }
            }
        }
    }
    finally {
        Pop-Location
    }
}

# Kubernetes 部署函数
function Invoke-KubernetesDeploy {
    param(
        [string]$EnvConfig,
        [string]$Action,
        [string]$Services
    )
    
    $platformPath = ".\platforms\kubernetes"
    
    switch ($Action) {
        "deploy" {
            Write-Step "部署到 Kubernetes..."
            if ($DryRun) {
                Write-ColorOutput "DRY RUN: kubectl apply -f $platformPath" "Yellow"
            } else {
                & kubectl apply -f $platformPath
            }
        }
        "stop" {
            Write-Step "从 Kubernetes 删除资源..."
            if ($DryRun) {
                Write-ColorOutput "DRY RUN: kubectl delete -f $platformPath" "Yellow"
            } else {
                & kubectl delete -f $platformPath
            }
        }
        "status" {
            Write-Step "查看 Kubernetes 资源状态..."
            & kubectl get all -n taishanglaojun
        }
        "logs" {
            Write-Step "查看 Kubernetes 日志..."
            if ($Services) {
                foreach ($service in $Services.Split(',')) {
                    & kubectl logs -f deployment/$service -n taishanglaojun
                }
            } else {
                & kubectl logs -f -l app=taishanglaojun -n taishanglaojun
            }
        }
    }
}

# Helm 部署函数
function Invoke-HelmDeploy {
    param(
        [string]$EnvConfig,
        [string]$Action,
        [string]$Services
    )
    
    $platformPath = ".\platforms\helm"
    $releaseName = "taishanglaojun-$Environment"
    
    switch ($Action) {
        "deploy" {
            Write-Step "使用 Helm 部署..."
            $valuesFile = "$platformPath\values-$Environment.yaml"
            if (-not (Test-Path $valuesFile)) {
                $valuesFile = "$platformPath\taishanglaojun\values.yaml"
            }
            
            if ($DryRun) {
                Write-ColorOutput "DRY RUN: helm upgrade --install $releaseName $platformPath\taishanglaojun -f $valuesFile" "Yellow"
            } else {
                & helm upgrade --install $releaseName "$platformPath\taishanglaojun" -f $valuesFile
            }
        }
        "stop" {
            Write-Step "卸载 Helm 发布..."
            if ($DryRun) {
                Write-ColorOutput "DRY RUN: helm uninstall $releaseName" "Yellow"
            } else {
                & helm uninstall $releaseName
            }
        }
        "status" {
            Write-Step "查看 Helm 发布状态..."
            & helm status $releaseName
        }
        "logs" {
            Write-Step "查看 Helm 部署日志..."
            & kubectl logs -f -l app.kubernetes.io/instance=$releaseName
        }
    }
}

# 主函数
function Main {
    Write-Header "太上老君微服务平台 - 统一部署脚本 v2.0.0"
    
    Write-ColorOutput "📋 部署配置:" "Cyan"
    Write-ColorOutput "   平台: $Platform" "White"
    Write-ColorOutput "   环境: $Environment" "White"
    Write-ColorOutput "   操作: $Action" "White"
    if ($Services) {
        Write-ColorOutput "   服务: $Services" "White"
    }
    if ($DryRun) {
        Write-ColorOutput "   模式: 预览模式 (DRY RUN)" "Yellow"
    }
    Write-ColorOutput ""
    
    # 检查依赖
    if (-not (Test-Dependencies)) {
        Write-Error "依赖检查失败，请安装缺失的工具"
        exit 1
    }
    
    # 加载环境配置
    $envConfig = Get-EnvironmentConfig -EnvName $Environment
    
    # 根据平台执行相应的部署操作
    try {
        switch ($Platform) {
            "docker" {
                Invoke-DockerDeploy -EnvConfig $envConfig -Action $Action -Services $Services
            }
            "kubernetes" {
                Invoke-KubernetesDeploy -EnvConfig $envConfig -Action $Action -Services $Services
            }
            "helm" {
                Invoke-HelmDeploy -EnvConfig $envConfig -Action $Action -Services $Services
            }
        }
        
        Write-Step "部署操作完成！"
        
        if ($Action -eq "deploy" -and -not $DryRun) {
            Write-ColorOutput "`n🎉 部署成功！访问地址:" "Green"
            Write-ColorOutput "   管理后台: http://localhost:8888" "Cyan"
            Write-ColorOutput "   插件市场: http://localhost" "Cyan"
            Write-ColorOutput "   API文档: http://localhost:8080/swagger" "Cyan"
        }
    }
    catch {
        Write-Error "部署过程中发生错误: $($_.Exception.Message)"
        exit 1
    }
}

# 执行主函数
Main