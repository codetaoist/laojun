# 太上老君系统部署脚本 (PowerShell版本)
# 用途：管理系统的部署、启动、停止、更新等操作

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter(Position=1)]
    [string]$Service = ""
)

# 脚本配置
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$DockerDir = Join-Path $ProjectRoot "deploy\docker"
$ConfigDir = Join-Path $ProjectRoot "deploy\configs"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $ColorMap = @{
        "Red" = "Red"
        "Green" = "Green"
        "Yellow" = "Yellow"
        "Blue" = "Blue"
        "White" = "White"
    }
    
    Write-Host $Message -ForegroundColor $ColorMap[$Color]
}

function Log-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Log-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Log-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Log-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

# 检查依赖
function Check-Dependencies {
    Log-Info "检查系统依赖..."
    
    # 检查 Docker
    try {
        $dockerVersion = docker --version
        Log-Info "Docker 版本: $dockerVersion"
    }
    catch {
        Log-Error "Docker 未安装或未启动，请先安装并启动 Docker Desktop"
        exit 1
    }
    
    # 检查 Docker Compose
    try {
        $composeVersion = docker-compose --version
        Log-Info "Docker Compose 版本: $composeVersion"
    }
    catch {
        Log-Error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    }
    
    Log-Success "系统依赖检查完成"
}

# 检查环境配置
function Check-Environment {
    Log-Info "检查环境配置..."
    
    $envFile = Join-Path $ConfigDir ".env"
    $envTemplate = Join-Path $ConfigDir ".env.template"
    
    if (-not (Test-Path $envFile)) {
        if (Test-Path $envTemplate) {
            Log-Warning "未找到 .env 文件，正在从模板创建..."
            Copy-Item $envTemplate $envFile
            Log-Warning "请编辑 $envFile 文件配置您的环境变量"
            return $false
        }
        else {
            Log-Error "未找到环境配置文件"
            exit 1
        }
    }
    
    Log-Success "环境配置检查完成"
    return $true
}

# 构建镜像
function Build-Images {
    Log-Info "构建 Docker 镜像..."
    
    Set-Location $DockerDir
    
    # 构建所有服务镜像
    docker-compose build --no-cache
    
    if ($LASTEXITCODE -eq 0) {
        Log-Success "Docker 镜像构建完成"
    }
    else {
        Log-Error "Docker 镜像构建失败"
        exit 1
    }
}

# 启动服务
function Start-Services {
    Log-Info "启动服务..."
    
    Set-Location $DockerDir
    
    # 启动所有服务
    docker-compose up -d
    
    if ($LASTEXITCODE -eq 0) {
        # 等待服务启动
        Log-Info "等待服务启动..."
        Start-Sleep -Seconds 30
        
        # 检查服务状态
        Check-ServicesHealth
        
        Log-Success "服务启动完成"
        Log-Info "访问地址："
        Log-Info "  插件市场: http://localhost"
        Log-Info "  管理后台: http://localhost:8888"
        Log-Info "  管理API: http://localhost:8080"
        Log-Info "  插件市场API: http://localhost:8082"
        Log-Info "  配置中心: http://localhost:8081"
    }
    else {
        Log-Error "服务启动失败"
        exit 1
    }
}

# 停止服务
function Stop-Services {
    Log-Info "停止服务..."
    
    Set-Location $DockerDir
    docker-compose down
    
    if ($LASTEXITCODE -eq 0) {
        Log-Success "服务已停止"
    }
    else {
        Log-Error "服务停止失败"
    }
}

# 重启服务
function Restart-Services {
    Log-Info "重启服务..."
    Stop-Services
    Start-Services
}

# 检查服务健康状态
function Check-ServicesHealth {
    Log-Info "检查服务健康状态..."
    
    Set-Location $DockerDir
    
    # 检查容器状态
    $status = docker-compose ps
    
    if ($status -match "Up") {
        Log-Success "服务运行正常"
        docker-compose ps
    }
    else {
        Log-Error "部分服务未正常启动"
        docker-compose ps
        return $false
    }
    
    return $true
}

# 查看日志
function View-Logs {
    param([string]$ServiceName)
    
    Set-Location $DockerDir
    
    if ([string]::IsNullOrEmpty($ServiceName)) {
        docker-compose logs -f
    }
    else {
        docker-compose logs -f $ServiceName
    }
}

# 数据库迁移
function Migrate-Database {
    Log-Info "执行数据库迁移..."
    
    # 确保数据库服务运行
    Set-Location $DockerDir
    docker-compose up -d postgres redis
    
    # 等待数据库启动
    Start-Sleep -Seconds 10
    
    # 执行迁移
    Set-Location $ProjectRoot
    
    try {
        & make migrate-complete-up
        Log-Success "数据库迁移完成"
    }
    catch {
        Log-Error "数据库迁移失败: $_"
        exit 1
    }
}

# 备份数据
function Backup-Data {
    $backupDir = Join-Path $ProjectRoot "backups\$(Get-Date -Format 'yyyyMMdd_HHmmss')"
    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    
    Log-Info "备份数据到 $backupDir..."
    
    Set-Location $DockerDir
    
    try {
        # 备份数据库
        $dbBackup = Join-Path $backupDir "database.sql"
        docker-compose exec -T postgres pg_dump -U laojun laojun | Out-File -FilePath $dbBackup -Encoding UTF8
        
        # 备份上传文件
        $uploadsDir = Join-Path $ProjectRoot "uploads"
        if (Test-Path $uploadsDir) {
            Copy-Item -Path $uploadsDir -Destination $backupDir -Recurse
        }
        
        # 备份配置文件
        Copy-Item -Path $ConfigDir -Destination $backupDir -Recurse
        
        Log-Success "数据备份完成: $backupDir"
    }
    catch {
        Log-Error "数据备份失败: $_"
        exit 1
    }
}

# 更新系统
function Update-System {
    Log-Info "更新系统..."
    
    # 备份数据
    Backup-Data
    
    # 停止服务
    Stop-Services
    
    # 构建新镜像
    Build-Images
    
    # 启动服务
    Start-Services
    
    Log-Success "系统更新完成"
}

# 清理系统
function Cleanup-System {
    $response = Read-Host "这将删除所有容器、镜像和数据，确定要继续吗？(y/N)"
    
    if ($response -match "^[Yy]$") {
        Log-Info "清理系统..."
        
        Set-Location $DockerDir
        
        # 停止并删除容器
        docker-compose down -v --rmi all
        
        # 清理未使用的镜像
        docker system prune -f
        
        Log-Success "系统清理完成"
    }
    else {
        Log-Info "取消清理操作"
    }
}

# 显示帮助信息
function Show-Help {
    Write-Host "太上老君系统部署脚本 (PowerShell版本)"
    Write-Host ""
    Write-Host "用法: .\deploy.ps1 [命令] [服务名]"
    Write-Host ""
    Write-Host "命令:"
    Write-Host "  check       检查系统依赖和环境配置"
    Write-Host "  build       构建 Docker 镜像"
    Write-Host "  start       启动所有服务"
    Write-Host "  stop        停止所有服务"
    Write-Host "  restart     重启所有服务"
    Write-Host "  status      检查服务状态"
    Write-Host "  logs [服务] 查看日志（可指定具体服务）"
    Write-Host "  migrate     执行数据库迁移"
    Write-Host "  backup      备份数据"
    Write-Host "  update      更新系统（包含备份、重建、重启）"
    Write-Host "  cleanup     清理系统（删除所有数据）"
    Write-Host "  help        显示此帮助信息"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\deploy.ps1 start                # 启动所有服务"
    Write-Host "  .\deploy.ps1 logs admin-api       # 查看管理API日志"
    Write-Host "  .\deploy.ps1 backup               # 备份数据"
}

# 主函数
function Main {
    switch ($Command.ToLower()) {
        "check" {
            Check-Dependencies
            Check-Environment
        }
        "build" {
            Check-Dependencies
            Build-Images
        }
        "start" {
            Check-Dependencies
            if (Check-Environment) {
                Start-Services
            }
        }
        "stop" {
            Stop-Services
        }
        "restart" {
            Restart-Services
        }
        "status" {
            Check-ServicesHealth
        }
        "logs" {
            View-Logs $Service
        }
        "migrate" {
            Migrate-Database
        }
        "backup" {
            Backup-Data
        }
        "update" {
            Update-System
        }
        "cleanup" {
            Cleanup-System
        }
        "help" {
            Show-Help
        }
        default {
            Log-Error "未知命令: $Command"
            Show-Help
            exit 1
        }
    }
}

# 执行主函数
Main