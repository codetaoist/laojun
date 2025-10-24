# 太上老君系统 - 本地开发环境启动脚本

param(
    [string]$Service = "all",
    [switch]$Build = $false,
    [switch]$Stop = $false
)

$ErrorActionPreference = "Stop"

# 设置工作目录
Set-Location $PSScriptRoot

# 颜色输出函数
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    } else {
        $input | Write-Output
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Info($message) {
    Write-ColorOutput Cyan "ℹ️  $message"
}

function Write-Success($message) {
    Write-ColorOutput Green "✅ $message"
}

function Write-Error($message) {
    Write-ColorOutput Red "❌ $message"
}

function Write-Warning($message) {
    Write-ColorOutput Yellow "⚠️  $message"
}

# 检查环境
function Test-Environment {
    Write-Info "检查本地开发环境..."
    
    # 检查Go版本
    try {
        $goVersion = go version
        Write-Success "Go环境: $goVersion"
    } catch {
        Write-Error "Go环境未安装或不在PATH中"
        exit 1
    }
    
    # 检查PostgreSQL连接
    try {
        $pgResult = & psql -U laojun -d laojun_local -c 'SELECT version();' 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "PostgreSQL连接正常"
        } else {
            Write-Warning "PostgreSQL数据库 laojun_local 不存在，将尝试创建"
            Create-LocalDatabase
        }
    } catch {
        Write-Error "PostgreSQL连接失败，请确保PostgreSQL服务运行且配置正确"
        Write-Info "请检查以下配置："
        Write-Info "- PostgreSQL服务运行在 localhost:5432"
        Write-Info "- 用户 laojun 存在且密码为 laojun123"
        Write-Info "- 数据库 laojun_local 存在"
        exit 1
    }
    
    # 检查Redis连接
    try {
        $redisResult = & redis-cli -a redis123 ping 2>$null
        if ($redisResult -eq "PONG") {
            Write-Success "Redis连接正常"
        } else {
            Write-Error "Redis连接失败"
            exit 1
        }
    } catch {
        Write-Error "Redis连接失败，请确保Redis服务运行在 localhost:6379"
        exit 1
    }
}

# 创建本地数据库
function Create-LocalDatabase {
    Write-Info "创建本地数据库..."
    try {
        & psql -U postgres -c 'CREATE DATABASE laojun_local;' 2>$null
        & psql -U postgres -c 'CREATE USER laojun WITH PASSWORD ''laojun123'';' 2>$null
        & psql -U postgres -c 'GRANT ALL PRIVILEGES ON DATABASE laojun_local TO laojun;' 2>$null
        Write-Success "数据库创建成功"
    } catch {
        Write-Warning "数据库创建失败，可能已存在"
    }
}

# 设置环境变量
function Set-LocalEnvironment {
    Write-Info "设置本地开发环境变量..."
    if (Test-Path ".env.local") {
        Copy-Item ".env.local" ".env" -Force
        Write-Success "环境变量文件已设置为本地开发模式"
    } else {
        Write-Error "本地环境变量文件 .env.local 不存在"
        exit 1
    }
}

# 构建服务
function Build-Services {
    Write-Info "构建Go服务..."
    
    $services = @("config-center", "admin-api", "marketplace-api")
    
    foreach ($svc in $services) {
        Write-Info "构建 $svc..."
        go build -o "bin/$svc.exe" "./cmd/$svc"
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$svc 构建成功"
        } else {
            Write-Error "$svc 构建失败"
            exit 1
        }
    }
}

# 启动服务
function Start-Service($serviceName) {
    $configFile = "configs/$serviceName.local.yaml"
    $binFile = "bin/$serviceName.exe"
    
    if (!(Test-Path $configFile)) {
        Write-Error "配置文件不存在: $configFile"
        return $false
    }
    
    if (!(Test-Path $binFile)) {
        Write-Error "可执行文件不存在: $binFile，请先运行构建"
        return $false
    }
    
    Write-Info "启动 $serviceName..."
    $env:CONFIG_FILE = $configFile
    Start-Process -FilePath $binFile -WindowStyle Normal
    Write-Success "$serviceName 已启动"
    return $true
}

# 停止服务
function Stop-Services {
    Write-Info "停止本地服务..."
    
    $processes = @("config-center", "admin-api", "marketplace-api")
    
    foreach ($proc in $processes) {
        $running = Get-Process -Name $proc -ErrorAction SilentlyContinue
        if ($running) {
            Stop-Process -Name $proc -Force
            Write-Success "已停止 $proc"
        }
    }
}

# 主逻辑
Write-Info "太上老君系统 - 本地开发环境管理"
Write-Info "========================================"

if ($Stop) {
    Stop-Services
    exit 0
}

# 设置环境
Set-LocalEnvironment

# 检查环境
Test-Environment

# 构建服务
if ($Build) {
    Build-Services
}

# 启动服务
switch ($Service.ToLower()) {
    "all" {
        Write-Info "启动所有服务..."
        Start-Service "config-center"
        Start-Sleep 2
        Start-Service "admin-api"
        Start-Sleep 2
        Start-Service "marketplace-api"
    }
    "config-center" {
        Start-Service "config-center"
    }
    "admin-api" {
        Start-Service "admin-api"
    }
    "marketplace-api" {
        Start-Service "marketplace-api"
    }
    default {
        Write-Error "未知服务: $Service"
        Write-Info "可用服务: all, config-center, admin-api, marketplace-api"
        exit 1
    }
}

Write-Success "本地开发环境启动完成！"
Write-Info ""
Write-Info "服务访问地址："
Write-Info "- 配置中心: http://localhost:8093"
Write-Info "- 管理API: http://localhost:8080"
Write-Info "- 市场API: http://localhost:8082"
Write-Info ""
Write-Info "使用 './start-local.ps1 -Stop' 停止所有服务"