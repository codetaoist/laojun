# 太上老君系统 - Docker环境启动脚本

param(
    [string]$Profile = "basic",
    [switch]$Build = $false,
    [switch]$Stop = $false,
    [switch]$Logs = $false
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

# Docker Compose文件路径
$ComposeFile = "deploy/docker/docker-compose.minimal.yml"

# 检查Docker环境
function Test-DockerEnvironment {
    Write-Info "检查Docker环境..."
    
    try {
        $dockerVersion = docker --version
        Write-Success "Docker: $dockerVersion"
    } catch {
        Write-Error "Docker未安装或不在PATH中"
        exit 1
    }
    
    try {
        $composeVersion = docker-compose --version
        Write-Success "Docker Compose: $composeVersion"
    } catch {
        Write-Error "Docker Compose未安装或不在PATH中"
        exit 1
    }
    
    if (!(Test-Path $ComposeFile)) {
        Write-Error "Docker Compose配置文件不存在: $ComposeFile"
        exit 1
    }
}

# 设置环境变量
function Set-DockerEnvironment {
    Write-Info "设置Docker环境变量..."
    if (Test-Path ".env.docker") {
        Copy-Item ".env.docker" ".env" -Force
        Write-Success "环境变量文件已设置为Docker模式"
    } else {
        Write-Error "Docker环境变量文件 .env.docker 不存在"
        exit 1
    }
}

# 构建镜像
function Build-Images {
    Write-Info "构建Docker镜像..."
    docker-compose -f $ComposeFile build
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Docker镜像构建成功"
    } else {
        Write-Error "Docker镜像构建失败"
        exit 1
    }
}

# 启动服务
function Start-Services($profile) {
    Write-Info "启动Docker服务 (Profile: $profile)..."
    
    switch ($profile.ToLower()) {
        "basic" {
            docker-compose -f $ComposeFile up -d
        }
        "dev-tools" {
            docker-compose -f $ComposeFile --profile dev-tools up -d
        }
        "all" {
            docker-compose -f $ComposeFile --profile dev-tools up -d
        }
        default {
            Write-Error "未知Profile: $profile"
            Write-Info "可用Profile: basic, dev-tools, all"
            exit 1
        }
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Docker服务启动成功"
    } else {
        Write-Error "Docker服务启动失败"
        exit 1
    }
}

# 停止服务
function Stop-Services {
    Write-Info "停止Docker服务..."
    docker-compose -f $ComposeFile --profile dev-tools down
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Docker服务已停止"
    } else {
        Write-Error "停止Docker服务失败"
        exit 1
    }
}

# 查看日志
function Show-Logs {
    Write-Info "显示Docker服务日志..."
    docker-compose -f $ComposeFile logs -f
}

# 显示服务状态
function Show-Status {
    Write-Info "Docker服务状态："
    docker-compose -f $ComposeFile --profile dev-tools ps
}

# 主逻辑
Write-Info "太上老君系统 - Docker环境管理"
Write-Info "====================================="

if ($Logs) {
    Show-Logs
    exit 0
}

if ($Stop) {
    Stop-Services
    exit 0
}

# 检查环境
Test-DockerEnvironment

# 设置环境
Set-DockerEnvironment

# 构建镜像
if ($Build) {
    Build-Images
}

# 启动服务
Start-Services $Profile

# 等待服务启动
Write-Info "等待服务启动..."
Start-Sleep 5

# 显示状态
Show-Status

Write-Success "Docker环境启动完成！"
Write-Info ""
Write-Info "服务访问地址："
Write-Info "- Web服务: http://localhost"

if ($Profile -eq "dev-tools" -or $Profile -eq "all") {
    Write-Info "- 数据库管理: http://localhost:8090"
    Write-Info "- Redis管理: http://localhost:8091"
}

Write-Info ""
Write-Info "常用命令："
Write-Info "- 查看日志: './start-docker.ps1 -Logs'"
Write-Info "- 停止服务: './start-docker.ps1 -Stop'"
Write-Info "- 重新构建: './start-docker.ps1 -Build'"