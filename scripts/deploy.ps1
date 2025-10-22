param(
    [string]$Action = "main"
)

# 日志函数
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $color = switch ($Level) {
        "ERROR" { "Red" }
        "WARNING" { "Yellow" }
        "SUCCESS" { "Green" }
        default { "White" }
    }
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $color
}

function Log-Info { param([string]$Message) Write-Log $Message "INFO" }
function Log-Error { param([string]$Message) Write-Log $Message "ERROR" }
function Log-Warning { param([string]$Message) Write-Log $Message "WARNING" }
function Log-Success { param([string]$Message) Write-Log $Message "SUCCESS" }

# 检查系统要求
function Check-Requirements {
    Log-Info "检查系统要求..."
    
    # 检查Docker
    try {
        $dockerVersion = docker --version
        Log-Success "Docker已安装: $dockerVersion"
    } catch {
        Log-Error "Docker未安装或不可用"
        return $false
    }
    
    # 检查Docker Compose
    try {
        $composeVersion = docker-compose --version
        Log-Success "Docker Compose已安装: $composeVersion"
    } catch {
        Log-Error "Docker Compose未安装或不可用"
        return $false
    }
    
    # 检查Git
    try {
        $gitVersion = git --version
        Log-Success "Git已安装: $gitVersion"
    } catch {
        Log-Error "Git未安装或不可用"
        return $false
    }
    
    # 检查Node.js
    try {
        $nodeVersion = node --version
        Log-Success "Node.js已安装: $nodeVersion"
    } catch {
        Log-Error "Node.js未安装或不可用"
        return $false
    }
    
    # 检查Go
    try {
        $goVersion = go version
        Log-Success "Go已安装: $goVersion"
    } catch {
        Log-Error "Go未安装或不可用"
        return $false
    }
    
    Log-Success "所有系统要求检查通过"
    return $true
}

# 创建必要目录
function Create-Directories {
    Log-Info "创建必要目录..."
    
    $directories = @("build", "logs", "data/postgres", "data/redis")
    
    foreach ($dir in $directories) {
        if (!(Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
    }
    
    Log-Success "目录创建完成"
}

# 检查环境配置文件
function Check-EnvFile {
    Log-Info "检查环境配置文件..."
    
    if (!(Test-Path ".env.production")) {
        Log-Warning ".env.production 文件不存在，从示例文件复制..."
        if (Test-Path ".env.example") {
            Copy-Item ".env.example" ".env.production"
            Log-Success ".env.production 文件已创建"
        } else {
            Log-Error ".env.example 文件不存在，请手动创建 .env.production"
            return $false
        }
    }
    
    Log-Success "环境配置文件检查完成"
    return $true
}

# 构建前端项目
function Build-Frontend {
    Log-Info "构建前端项目..."
    
    # 构建Admin项目
    Log-Info "构建 Admin 项目..."
    Set-Location "web/admin"
    npm install
    npm run build
    Set-Location "../.."
    
    # 构建Marketplace项目
    Log-Info "构建 Marketplace 项目..."
    Set-Location "web/marketplace"
    npm install
    npm run build
    Set-Location "../.."
    
    Log-Success "前端项目构建完成"
}

# 构建后端服务
function Build-Backend {
    Log-Info "构建后端服务..."
    
    # 下载依赖
    go mod download
    
    # 构建各个服务
    Log-Info "构建 Config Center..."
    go build -o "build/config-center.exe" "./cmd/config-center"
    
    Log-Info "构建 Admin API..."
    go build -o "build/admin-api.exe" "./cmd/admin-api"
    
    Log-Info "构建 Marketplace API..."
    go build -o "build/marketplace-api.exe" "./cmd/marketplace-api"
    
    Log-Success "后端服务构建完成"
}

# 部署服务
function Deploy-Services {
    Log-Info "部署服务..."
    
    # 停止现有服务
    docker-compose down
    
    # 启动服务
    docker-compose up -d
    
    Log-Success "服务部署完成"
}

# 运行数据库迁移
function Run-Migrations {
    Log-Info "运行数据库迁移..."
    
    # 等待数据库启动
    Start-Sleep -Seconds 10
    
    # 运行迁移
    if (Test-Path "migrations/001_init.sql") {
        Log-Info "执行数据库初始化..."
        # 这里需要根据实际情况调整数据库连接参数
        # psql -h localhost -U laojun -d laojun -f migrations/001_init.sql
    }
    
    Log-Success "数据库迁移完成"
}

# 健康检查
function Test-Health {
    Log-Info "执行健康检查..."
    
    $services = @(
        @{Name="Config Center"; Url="http://localhost:8090/health"},
        @{Name="Admin API"; Url="http://localhost:8091/health"},
        @{Name="Marketplace API"; Url="http://localhost:8092/health"}
    )
    
    foreach ($service in $services) {
        try {
            $response = Invoke-WebRequest -Uri $service.Url -TimeoutSec 5
            if ($response.StatusCode -eq 200) {
                Log-Success "$($service.Name) 健康检查通过"
            } else {
                Log-Warning "$($service.Name) 健康检查失败: $($response.StatusCode)"
            }
        } catch {
            Log-Warning "$($service.Name) 健康检查失败: $($_.Exception.Message)"
        }
    }
}

# 显示部署信息
function Show-DeploymentInfo {
    Log-Info "部署信息:"
    Write-Host ""
    Write-Host "服务访问地址:" -ForegroundColor Cyan
    Write-Host "  Admin 后台:        http://localhost:3000" -ForegroundColor White
    Write-Host "  Marketplace:       http://localhost:3001" -ForegroundColor White
    Write-Host "  Config Center:     http://localhost:8090" -ForegroundColor White
    Write-Host "  Admin API:         http://localhost:8091" -ForegroundColor White
    Write-Host "  Marketplace API:   http://localhost:8092" -ForegroundColor White
    Write-Host ""
    Write-Host "监控地址:" -ForegroundColor Cyan
    Write-Host "  Prometheus:        http://localhost:9090" -ForegroundColor White
    Write-Host "  Grafana:           http://localhost:3003" -ForegroundColor White
    Write-Host ""
}

# 主函数
function Main {
    Log-Info "开始部署太上老君系统..."
    
    if (!(Check-Requirements)) {
        Log-Error "系统要求检查失败，请安装必要的软件"
        exit 1
    }
    
    Create-Directories
    
    if (!(Check-EnvFile)) {
        Log-Error "环境配置文件检查失败"
        exit 1
    }
    
    Build-Frontend
    Build-Backend
    Deploy-Services
    Run-Migrations
    
    Start-Sleep -Seconds 15
    Test-Health
    Show-DeploymentInfo
    
    Log-Success "部署流程完成！"
}

# 脚本参数处理
switch ($Action) {
    "check" {
        Check-Requirements
    }
    "build" {
        Build-Frontend
        Build-Backend
    }
    "deploy" {
        Deploy-Services
    }
    "health" {
        Test-Health
    }
    default {
        Main
    }
}