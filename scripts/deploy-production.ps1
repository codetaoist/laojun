# 生产环境部署脚本 (PowerShell版本)
# 使用方法: .\deploy-production.ps1 [action]
# Actions: build, deploy, stop, restart, logs, status

param(
    [Parameter(Position=0)]
    [ValidateSet("build", "deploy", "stop", "restart", "logs", "status")]
    [string]$Action = "deploy"
)

# 配置变量
$ProjectName = "laojun"
$ComposeFile = "deployments/docker-compose.prod.yml"
$EnvFile = ".env.prod"
$BackupDir = "backups"

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# 检查必要条件
function Test-Requirements {
    Write-Info "检查部署环境..."
    
    # 检查Docker
    try {
        docker --version | Out-Null
    }
    catch {
        Write-Error "Docker 未安装或未启动"
        exit 1
    }
    
    # 检查Docker Compose
    try {
        docker-compose --version | Out-Null
    }
    catch {
        Write-Error "Docker Compose 未安装"
        exit 1
    }
    
    # 检查环境变量文件
    if (-not (Test-Path $EnvFile)) {
        Write-Error "环境变量文件 $EnvFile 不存在"
        Write-Info "请复制 .env.production.example 到 $EnvFile 并配置相应的值"
        exit 1
    }
    
    # 检查docker-compose文件
    if (-not (Test-Path $ComposeFile)) {
        Write-Error "Docker Compose 文件 $ComposeFile 不存在"
        exit 1
    }
    
    Write-Success "环境检查通过"
}

# 创建备份
function New-Backup {
    Write-Info "创建数据备份..."
    
    if (-not (Test-Path $BackupDir)) {
        New-Item -ItemType Directory -Path $BackupDir | Out-Null
    }
    
    $BackupFile = "$BackupDir/backup-$(Get-Date -Format 'yyyyMMdd-HHmmss').sql"
    
    # 检查PostgreSQL容器是否运行
    $PostgresStatus = docker-compose -f $ComposeFile --env-file $EnvFile ps postgres
    if ($PostgresStatus -match "Up") {
        Write-Info "备份数据库到 $BackupFile"
        # 这里需要根据实际环境变量配置数据库备份
        Write-Warning "请手动配置数据库备份逻辑"
    }
    else {
        Write-Warning "PostgreSQL 容器未运行，跳过备份"
    }
}

# 构建镜像
function Build-Images {
    Write-Info "构建Docker镜像..."
    
    docker-compose -f $ComposeFile --env-file $EnvFile build --no-cache
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "镜像构建完成"
    }
    else {
        Write-Error "镜像构建失败"
        exit 1
    }
}

# 部署服务
function Deploy-Services {
    Write-Info "部署服务..."
    
    # 先启动基础服务
    Write-Info "启动数据库和缓存服务..."
    docker-compose -f $ComposeFile --env-file $EnvFile up -d postgres redis
    
    # 等待数据库启动
    Write-Info "等待数据库启动..."
    Start-Sleep -Seconds 10
    
    # 启动后端服务
    Write-Info "启动后端服务..."
    docker-compose -f $ComposeFile --env-file $EnvFile up -d config-center
    Start-Sleep -Seconds 5
    docker-compose -f $ComposeFile --env-file $EnvFile up -d admin-api marketplace-api
    
    # 启动前端服务
    Write-Info "启动前端服务..."
    docker-compose -f $ComposeFile --env-file $EnvFile up -d admin-web marketplace-web
    
    # 启动反向代理
    Write-Info "启动Nginx反向代理..."
    docker-compose -f $ComposeFile --env-file $EnvFile up -d nginx
    
    # 启动监控服务（可选）
    Write-Info "启动监控服务..."
    docker-compose -f $ComposeFile --env-file $EnvFile up -d prometheus grafana
    
    Write-Success "服务部署完成"
}

# 停止服务
function Stop-Services {
    Write-Info "停止所有服务..."
    docker-compose -f $ComposeFile --env-file $EnvFile down
    Write-Success "服务已停止"
}

# 重启服务
function Restart-Services {
    Write-Info "重启服务..."
    Stop-Services
    Deploy-Services
}

# 查看日志
function Show-Logs {
    docker-compose -f $ComposeFile --env-file $EnvFile logs -f
}

# 查看状态
function Get-Status {
    Write-Info "服务状态:"
    docker-compose -f $ComposeFile --env-file $EnvFile ps
    
    Write-Info "健康检查:"
    
    try {
        $ConfigCenterStatus = Invoke-WebRequest -Uri "http://localhost:8090/health" -UseBasicParsing -TimeoutSec 5
        Write-Host "Config Center: $($ConfigCenterStatus.StatusCode)" -ForegroundColor Green
    }
    catch {
        Write-Host "Config Center: 无法连接" -ForegroundColor Red
    }
    
    try {
        $AdminApiStatus = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
        Write-Host "Admin API: $($AdminApiStatus.StatusCode)" -ForegroundColor Green
    }
    catch {
        Write-Host "Admin API: 无法连接" -ForegroundColor Red
    }
    
    try {
        $MarketplaceApiStatus = Invoke-WebRequest -Uri "http://localhost:8082/health" -UseBasicParsing -TimeoutSec 5
        Write-Host "Marketplace API: $($MarketplaceApiStatus.StatusCode)" -ForegroundColor Green
    }
    catch {
        Write-Host "Marketplace API: 无法连接" -ForegroundColor Red
    }
}

# 主函数
function Main {
    switch ($Action) {
        "build" {
            Test-Requirements
            Build-Images
        }
        "deploy" {
            Test-Requirements
            New-Backup
            Build-Images
            Deploy-Services
            Get-Status
        }
        "stop" {
            Stop-Services
        }
        "restart" {
            Test-Requirements
            Restart-Services
            Get-Status
        }
        "logs" {
            Show-Logs
        }
        "status" {
            Get-Status
        }
        default {
            Write-Host "使用方法: .\deploy-production.ps1 [build|deploy|stop|restart|logs|status]"
            Write-Host ""
            Write-Host "Actions:"
            Write-Host "  build   - 仅构建Docker镜像"
            Write-Host "  deploy  - 完整部署（默认）"
            Write-Host "  stop    - 停止所有服务"
            Write-Host "  restart - 重启所有服务"
            Write-Host "  logs    - 查看服务日志"
            Write-Host "  status  - 查看服务状态"
        }
    }
}

# 执行主函数
Main