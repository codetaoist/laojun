# Laojun 监控体系启动脚本
# 用于启动完整的监控、日志和追踪系统

param(
    [string]$Environment = "development",
    [switch]$SkipBuild = $false,
    [switch]$Detached = $true,
    [string[]]$Services = @()
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 获取脚本目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$DockerDir = Join-Path $ProjectRoot "etc\docker"

Write-Host "=== Laojun 监控体系启动脚本 ===" -ForegroundColor Green
Write-Host "项目根目录: $ProjectRoot" -ForegroundColor Yellow
Write-Host "环境: $Environment" -ForegroundColor Yellow
Write-Host "跳过构建: $SkipBuild" -ForegroundColor Yellow

# 检查 Docker 是否运行
try {
    docker version | Out-Null
    Write-Host "✓ Docker 运行正常" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker 未运行或未安装" -ForegroundColor Red
    exit 1
}

# 检查 Docker Compose 是否可用
try {
    docker-compose version | Out-Null
    Write-Host "✓ Docker Compose 可用" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker Compose 未安装" -ForegroundColor Red
    exit 1
}

# 切换到 Docker 目录
Set-Location $DockerDir

# 创建必要的目录
$Directories = @(
    "..\..\var\log\laojun",
    "..\..\var\lib\laojun\data",
    "..\..\var\lib\laojun\plugins",
    "..\prometheus\data",
    "..\grafana\data",
    "..\alertmanager\data",
    "..\loki\data"
)

foreach ($Dir in $Directories) {
    $FullPath = Join-Path $DockerDir $Dir
    if (-not (Test-Path $FullPath)) {
        New-Item -ItemType Directory -Path $FullPath -Force | Out-Null
        Write-Host "✓ 创建目录: $Dir" -ForegroundColor Green
    }
}

# 设置环境变量
$env:COMPOSE_PROJECT_NAME = "laojun"
$env:COMPOSE_FILE = "docker-compose.yml"

# 定义服务组
$InfraServices = @("postgres", "redis", "minio")
$MonitoringServices = @("prometheus", "grafana", "alertmanager", "node-exporter", "cadvisor")
$LoggingServices = @("loki", "promtail")
$TracingServices = @("jaeger")
$ExporterServices = @("postgres-exporter", "redis-exporter")
$AppServices = @("config-center", "admin-api", "marketplace-api")
$ProxyServices = @("nginx")

$AllServices = $InfraServices + $MonitoringServices + $LoggingServices + $TracingServices + $ExporterServices + $AppServices + $ProxyServices

# 如果指定了特定服务，只启动这些服务
if ($Services.Count -gt 0) {
    $ServicesToStart = $Services
    Write-Host "启动指定服务: $($ServicesToStart -join ', ')" -ForegroundColor Yellow
} else {
    $ServicesToStart = $AllServices
    Write-Host "启动所有服务" -ForegroundColor Yellow
}

try {
    # 停止现有服务
    Write-Host "`n=== 停止现有服务 ===" -ForegroundColor Cyan
    docker-compose down --remove-orphans

    # 构建镜像（如果需要）
    if (-not $SkipBuild) {
        Write-Host "`n=== 构建应用镜像 ===" -ForegroundColor Cyan
        $BuildServices = $AppServices | Where-Object { $ServicesToStart -contains $_ }
        if ($BuildServices.Count -gt 0) {
            docker-compose build $BuildServices
            Write-Host "✓ 应用镜像构建完成" -ForegroundColor Green
        }
    }

    # 启动基础设施服务
    Write-Host "`n=== 启动基础设施服务 ===" -ForegroundColor Cyan
    $InfraToStart = $InfraServices | Where-Object { $ServicesToStart -contains $_ }
    if ($InfraToStart.Count -gt 0) {
        if ($Detached) {
            docker-compose up -d $InfraToStart
        } else {
            Start-Job -ScriptBlock { docker-compose up $InfraToStart }
        }
        
        # 等待基础设施服务启动
        Write-Host "等待基础设施服务启动..." -ForegroundColor Yellow
        Start-Sleep -Seconds 30
        
        # 检查服务健康状态
        foreach ($Service in $InfraToStart) {
            $HealthCheck = docker-compose ps --filter "health=healthy" --services | Where-Object { $_ -eq $Service }
            if ($HealthCheck) {
                Write-Host "✓ $Service 健康检查通过" -ForegroundColor Green
            } else {
                Write-Host "⚠ $Service 健康检查未通过，继续启动..." -ForegroundColor Yellow
            }
        }
    }

    # 启动监控服务
    Write-Host "`n=== 启动监控服务 ===" -ForegroundColor Cyan
    $MonitoringToStart = ($MonitoringServices + $ExporterServices) | Where-Object { $ServicesToStart -contains $_ }
    if ($MonitoringToStart.Count -gt 0) {
        if ($Detached) {
            docker-compose up -d $MonitoringToStart
        } else {
            Start-Job -ScriptBlock { docker-compose up $MonitoringToStart }
        }
        Start-Sleep -Seconds 20
    }

    # 启动日志和追踪服务
    Write-Host "`n=== 启动日志和追踪服务 ===" -ForegroundColor Cyan
    $LogTracingToStart = ($LoggingServices + $TracingServices) | Where-Object { $ServicesToStart -contains $_ }
    if ($LogTracingToStart.Count -gt 0) {
        if ($Detached) {
            docker-compose up -d $LogTracingToStart
        } else {
            Start-Job -ScriptBlock { docker-compose up $LogTracingToStart }
        }
        Start-Sleep -Seconds 15
    }

    # 启动应用服务
    Write-Host "`n=== 启动应用服务 ===" -ForegroundColor Cyan
    $AppsToStart = $AppServices | Where-Object { $ServicesToStart -contains $_ }
    if ($AppsToStart.Count -gt 0) {
        if ($Detached) {
            docker-compose up -d $AppsToStart
        } else {
            Start-Job -ScriptBlock { docker-compose up $AppsToStart }
        }
        Start-Sleep -Seconds 20
    }

    # 启动代理服务
    Write-Host "`n=== 启动代理服务 ===" -ForegroundColor Cyan
    $ProxyToStart = $ProxyServices | Where-Object { $ServicesToStart -contains $_ }
    if ($ProxyToStart.Count -gt 0) {
        if ($Detached) {
            docker-compose up -d $ProxyToStart
        } else {
            Start-Job -ScriptBlock { docker-compose up $ProxyToStart }
        }
    }

    # 显示服务状态
    Write-Host "`n=== 服务状态 ===" -ForegroundColor Cyan
    docker-compose ps

    # 显示访问地址
    Write-Host "`n=== 服务访问地址 ===" -ForegroundColor Green
    Write-Host "🌐 应用服务:" -ForegroundColor Yellow
    Write-Host "  - 管理后台 API: http://localhost:8080" -ForegroundColor White
    Write-Host "  - 插件市场 API: http://localhost:8081" -ForegroundColor White
    Write-Host "  - 配置中心: http://localhost:8090" -ForegroundColor White
    Write-Host "  - Nginx 代理: http://localhost" -ForegroundColor White
    
    Write-Host "`n📊 监控服务:" -ForegroundColor Yellow
    Write-Host "  - Prometheus: http://localhost:9090" -ForegroundColor White
    Write-Host "  - Grafana: http://localhost:3000 (admin/admin123)" -ForegroundColor White
    Write-Host "  - AlertManager: http://localhost:9093" -ForegroundColor White
    
    Write-Host "`n📋 日志和追踪:" -ForegroundColor Yellow
    Write-Host "  - Loki: http://localhost:3100" -ForegroundColor White
    Write-Host "  - Jaeger: http://localhost:16686" -ForegroundColor White
    
    Write-Host "`n🗄️ 基础设施:" -ForegroundColor Yellow
    Write-Host "  - PostgreSQL: localhost:5432" -ForegroundColor White
    Write-Host "  - Redis: localhost:6379" -ForegroundColor White
    Write-Host "  - MinIO: http://localhost:9001 (minioadmin/minioadmin123)" -ForegroundColor White

    Write-Host "`n✅ 监控体系启动完成！" -ForegroundColor Green
    Write-Host "使用 'docker-compose logs -f [service]' 查看服务日志" -ForegroundColor Cyan
    Write-Host "使用 'docker-compose down' 停止所有服务" -ForegroundColor Cyan

} catch {
    Write-Host "`n❌ 启动失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "查看详细日志: docker-compose logs" -ForegroundColor Yellow
    exit 1
}

# 如果不是分离模式，等待用户输入
if (-not $Detached) {
    Write-Host "`n按任意键停止服务..." -ForegroundColor Yellow
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    docker-compose down
}