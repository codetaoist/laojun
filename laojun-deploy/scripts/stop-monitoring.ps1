# Laojun 监控体系停止脚本
# 用于安全停止所有监控、日志和追踪服务

param(
    [switch]$RemoveVolumes = $false,
    [switch]$RemoveImages = $false,
    [string[]]$Services = @()
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 获取脚本目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$DockerDir = Join-Path $ProjectRoot "etc\docker"

Write-Host "=== Laojun 监控体系停止脚本 ===" -ForegroundColor Red
Write-Host "项目根目录: $ProjectRoot" -ForegroundColor Yellow
Write-Host "删除数据卷: $RemoveVolumes" -ForegroundColor Yellow
Write-Host "删除镜像: $RemoveImages" -ForegroundColor Yellow

# 检查 Docker 是否运行
try {
    docker version | Out-Null
    Write-Host "✓ Docker 运行正常" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker 未运行或未安装" -ForegroundColor Red
    exit 1
}

# 切换到 Docker 目录
Set-Location $DockerDir

# 设置环境变量
$env:COMPOSE_PROJECT_NAME = "laojun"
$env:COMPOSE_FILE = "docker-compose.yml"

try {
    # 显示当前运行的服务
    Write-Host "`n=== 当前运行的服务 ===" -ForegroundColor Cyan
    $RunningServices = docker-compose ps --services --filter "status=running"
    if ($RunningServices) {
        Write-Host "运行中的服务:" -ForegroundColor Yellow
        $RunningServices | ForEach-Object { Write-Host "  - $_" -ForegroundColor White }
    } else {
        Write-Host "没有运行中的服务" -ForegroundColor Yellow
    }

    # 如果指定了特定服务，只停止这些服务
    if ($Services.Count -gt 0) {
        Write-Host "`n=== 停止指定服务 ===" -ForegroundColor Cyan
        Write-Host "停止服务: $($Services -join ', ')" -ForegroundColor Yellow
        docker-compose stop $Services
        docker-compose rm -f $Services
    } else {
        # 优雅停止所有服务
        Write-Host "`n=== 优雅停止所有服务 ===" -ForegroundColor Cyan
        
        # 首先停止应用服务
        Write-Host "停止应用服务..." -ForegroundColor Yellow
        $AppServices = @("nginx", "admin-api", "marketplace-api", "config-center")
        docker-compose stop $AppServices
        
        # 然后停止监控和日志服务
        Write-Host "停止监控和日志服务..." -ForegroundColor Yellow
        $MonitoringServices = @("grafana", "prometheus", "alertmanager", "loki", "promtail", "jaeger")
        docker-compose stop $MonitoringServices
        
        # 停止导出器
        Write-Host "停止导出器..." -ForegroundColor Yellow
        $ExporterServices = @("node-exporter", "cadvisor", "postgres-exporter", "redis-exporter")
        docker-compose stop $ExporterServices
        
        # 最后停止基础设施服务
        Write-Host "停止基础设施服务..." -ForegroundColor Yellow
        $InfraServices = @("minio", "redis", "postgres")
        docker-compose stop $InfraServices
        
        # 移除容器
        Write-Host "移除容器..." -ForegroundColor Yellow
        docker-compose rm -f
    }

    # 移除数据卷（如果指定）
    if ($RemoveVolumes) {
        Write-Host "`n=== 移除数据卷 ===" -ForegroundColor Red
        Write-Host "⚠️  警告：这将删除所有持久化数据！" -ForegroundColor Red
        $Confirmation = Read-Host "确认删除数据卷？(yes/no)"
        if ($Confirmation -eq "yes") {
            docker-compose down -v
            Write-Host "✓ 数据卷已删除" -ForegroundColor Green
        } else {
            Write-Host "取消删除数据卷" -ForegroundColor Yellow
        }
    }

    # 移除镜像（如果指定）
    if ($RemoveImages) {
        Write-Host "`n=== 移除镜像 ===" -ForegroundColor Red
        Write-Host "⚠️  警告：这将删除所有相关镜像！" -ForegroundColor Red
        $Confirmation = Read-Host "确认删除镜像？(yes/no)"
        if ($Confirmation -eq "yes") {
            # 获取项目相关的镜像
            $ProjectImages = docker images --filter "label=com.docker.compose.project=laojun" -q
            if ($ProjectImages) {
                docker rmi $ProjectImages -f
                Write-Host "✓ 项目镜像已删除" -ForegroundColor Green
            }
            
            # 删除构建的应用镜像
            $AppImages = @("laojun/config-center", "laojun/admin-api", "laojun/marketplace-api")
            foreach ($Image in $AppImages) {
                $ImageId = docker images $Image -q
                if ($ImageId) {
                    docker rmi $Image -f
                    Write-Host "✓ 删除镜像: $Image" -ForegroundColor Green
                }
            }
        } else {
            Write-Host "取消删除镜像" -ForegroundColor Yellow
        }
    }

    # 清理网络
    Write-Host "`n=== 清理网络 ===" -ForegroundColor Cyan
    $Networks = docker network ls --filter "name=laojun" -q
    if ($Networks) {
        docker network rm $Networks 2>$null
        Write-Host "✓ 网络已清理" -ForegroundColor Green
    }

    # 清理未使用的资源
    Write-Host "`n=== 清理未使用的资源 ===" -ForegroundColor Cyan
    docker system prune -f
    Write-Host "✓ 未使用的资源已清理" -ForegroundColor Green

    # 显示最终状态
    Write-Host "`n=== 最终状态 ===" -ForegroundColor Cyan
    $RemainingContainers = docker-compose ps -q
    if ($RemainingContainers) {
        Write-Host "剩余容器:" -ForegroundColor Yellow
        docker-compose ps
    } else {
        Write-Host "所有容器已停止" -ForegroundColor Green
    }

    # 显示磁盘使用情况
    Write-Host "`n=== Docker 磁盘使用情况 ===" -ForegroundColor Cyan
    docker system df

    Write-Host "`n✅ 监控体系停止完成！" -ForegroundColor Green
    
    if (-not $RemoveVolumes) {
        Write-Host "💡 提示：数据卷已保留，下次启动时数据将恢复" -ForegroundColor Cyan
        Write-Host "   如需完全清理，请使用 -RemoveVolumes 参数" -ForegroundColor Cyan
    }

} catch {
    Write-Host "`n❌ 停止失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "查看详细信息: docker-compose logs" -ForegroundColor Yellow
    exit 1
}

Write-Host "`n使用 './start-monitoring.ps1' 重新启动监控体系" -ForegroundColor Cyan