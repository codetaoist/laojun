# 太上老君系统自动化部署脚本 (PowerShell版本)
# 使用方法: .\deploy.ps1 [环境] [操作]
# 环境: dev|staging|prod (默认: prod)
# 操作: build|deploy|restart|stop|logs|backup (默认: deploy)

param(
    [string]$Environment = "prod",
    [string]$Action = "deploy"
)

# 配置变量
$ProjectName = "laojun"
$ComposeFile = "../docker/docker-compose.$Environment.yml"
$EnvFile = "../configs/.env.$Environment"

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

# 检查依赖
function Test-Dependencies {
    Write-Info "检查系统依赖..."
    
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error "Docker 未安装，请先安装 Docker Desktop"
        exit 1
    }
    
    if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue)) {
        Write-Error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    }
    
    Write-Success "系统依赖检查完成"
}

# 检查配置文件
function Test-Config {
    Write-Info "检查配置文件..."
    
    if (-not (Test-Path $ComposeFile)) {
        Write-Error "Docker Compose 文件不存在: $ComposeFile"
        exit 1
    }
    
    if (-not (Test-Path $EnvFile)) {
        Write-Warning "环境配置文件不存在: $EnvFile，将使用默认配置"
    }
    
    Write-Success "配置文件检查完成"
}

# 构建镜像
function Build-Images {
    Write-Info "构建 Docker 镜像..."
    
    docker-compose -f $ComposeFile --env-file $EnvFile build --no-cache
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Docker 镜像构建完成"
    } else {
        Write-Error "Docker 镜像构建失败"
        exit 1
    }
}

# 数据库迁移
function Invoke-Migrations {
    Write-Info "执行数据库迁移..."
    
    # 等待数据库启动
    docker-compose -f $ComposeFile --env-file $EnvFile up -d postgres redis
    Start-Sleep -Seconds 10
    
    # 运行迁移
    docker-compose -f $ComposeFile --env-file $EnvFile run --rm laojun-app /app/bin/db-migrate up
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "数据库迁移完成"
    } else {
        Write-Error "数据库迁移失败"
        exit 1
    }
}

# 部署应用
function Deploy-App {
    Write-Info "部署应用..."
    
    # 停止旧容器
    docker-compose -f $ComposeFile --env-file $EnvFile down
    
    # 构建镜像
    Build-Images
    
    # 启动服务
    docker-compose -f $ComposeFile --env-file $EnvFile up -d
    
    if ($LASTEXITCODE -eq 0) {
        # 等待服务启动
        Write-Info "等待服务启动..."
        Start-Sleep -Seconds 30
        
        # 检查服务状态
        Test-Health
        
        Write-Success "应用部署完成"
    } else {
        Write-Error "应用部署失败"
        exit 1
    }
}

# 重启服务
function Restart-Services {
    Write-Info "重启服务..."
    
    docker-compose -f $ComposeFile --env-file $EnvFile restart
    
    if ($LASTEXITCODE -eq 0) {
        # 等待服务启动
        Start-Sleep -Seconds 15
        Test-Health
        
        Write-Success "服务重启完成"
    } else {
        Write-Error "服务重启失败"
        exit 1
    }
}

# 停止服务
function Stop-Services {
    Write-Info "停止服务..."
    
    docker-compose -f $ComposeFile --env-file $EnvFile down
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "服务已停止"
    } else {
        Write-Error "服务停止失败"
        exit 1
    }
}

# 查看日志
function Show-Logs {
    Write-Info "显示服务日志..."
    
    docker-compose -f $ComposeFile --env-file $EnvFile logs -f --tail=100
}

# 健康检查
function Test-Health {
    Write-Info "检查服务健康状态..."
    
    $maxAttempts = 30
    $attempt = 1
    
    while ($attempt -le $maxAttempts) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost/health" -TimeoutSec 5 -ErrorAction Stop
            if ($response.StatusCode -eq 200) {
                Write-Success "服务健康检查通过"
                return $true
            }
        }
        catch {
            # 忽略错误，继续尝试
        }
        
        Write-Info "等待服务启动... ($attempt/$maxAttempts)"
        Start-Sleep -Seconds 5
        $attempt++
    }
    
    Write-Error "服务健康检查失败"
    return $false
}

# 备份数据
function Backup-Data {
    Write-Info "备份数据..."
    
    $backupDir = ".\backups\$(Get-Date -Format 'yyyyMMdd_HHmmss')"
    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    
    # 备份数据库
    docker-compose -f $ComposeFile --env-file $EnvFile exec -T postgres pg_dump -U laojun laojun | Out-File -FilePath "$backupDir\database.sql" -Encoding UTF8
    
    # 备份上传文件
    if (Test-Path ".\uploads") {
        Copy-Item -Path ".\uploads" -Destination $backupDir -Recurse
    }
    
    # 压缩备份
    Compress-Archive -Path "$backupDir\*" -DestinationPath "$backupDir.zip"
    Remove-Item -Path $backupDir -Recurse -Force
    
    Write-Success "数据备份完成: $backupDir.zip"
}

# 清理旧镜像
function Remove-OldImages {
    Write-Info "清理旧的 Docker 镜像..."
    
    docker image prune -f
    docker system prune -f
    
    Write-Success "镜像清理完成"
}

# 显示帮助信息
function Show-Help {
    Write-Host "太上老君系统部署脚本 (PowerShell版本)"
    Write-Host ""
    Write-Host "使用方法:"
    Write-Host "  .\deploy.ps1 [环境] [操作]"
    Write-Host ""
    Write-Host "环境:"
    Write-Host "  dev      开发环境"
    Write-Host "  staging  测试环境"
    Write-Host "  prod     生产环境 (默认)"
    Write-Host ""
    Write-Host "操作:"
    Write-Host "  build    构建镜像"
    Write-Host "  deploy   部署应用 (默认)"
    Write-Host "  restart  重启服务"
    Write-Host "  stop     停止服务"
    Write-Host "  logs     查看日志"
    Write-Host "  backup   备份数据"
    Write-Host "  health   健康检查"
    Write-Host "  cleanup  清理镜像"
    Write-Host "  help     显示帮助"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\deploy.ps1 prod deploy    # 部署到生产环境"
    Write-Host "  .\deploy.ps1 dev logs       # 查看开发环境日志"
    Write-Host "  .\deploy.ps1 prod backup    # 备份生产环境数据"
}

# 主函数
function Main {
    switch ($Action.ToLower()) {
        "build" {
            Test-Dependencies
            Test-Config
            Build-Images
        }
        "deploy" {
            Test-Dependencies
            Test-Config
            Deploy-App
        }
        "restart" {
            Test-Dependencies
            Test-Config
            Restart-Services
        }
        "stop" {
            Test-Dependencies
            Test-Config
            Stop-Services
        }
        "logs" {
            Test-Dependencies
            Test-Config
            Show-Logs
        }
        "backup" {
            Test-Dependencies
            Test-Config
            Backup-Data
        }
        "health" {
            Test-Health
        }
        "cleanup" {
            Test-Dependencies
            Remove-OldImages
        }
        "help" {
            Show-Help
        }
        default {
            Write-Error "未知操作: $Action"
            Show-Help
            exit 1
        }
    }
}

# 执行主函数
Main