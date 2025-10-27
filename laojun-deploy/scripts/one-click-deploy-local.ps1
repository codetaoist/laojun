# 太上老君系统一键部署脚本 (本地镜像优先版本)
# 功能：
# 1. 检查本地Docker镜像
# 2. 优先使用本地镜像，避免网络拉取问题
# 3. 自动选择最佳镜像版本
# 4. 完整的部署流程管理

param(
    [switch]$Production,    # 生产环境部署
    [switch]$Clean,         # 清理现有容器
    [switch]$Help,          # 显示帮助
    [switch]$CheckOnly,     # 仅检查镜像，不部署
    [switch]$UseLocal       # 强制使用本地镜像配置
)

# 脚本路径配置
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$DockerDir = Join-Path $ProjectRoot "deploy\docker"
$ConfigDir = Join-Path $ProjectRoot "configs"

# 颜色输出函数
function Write-Info { param($Message) Write-Host "[INFO] $Message" -ForegroundColor Blue }
function Write-Success { param($Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-Warning { param($Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
function Write-Error { param($Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

# 显示帮助信息
function Show-Help {
    Write-Host @"
太上老君系统一键部署脚本 (本地镜像优先版本)

用法:
    .\one-click-deploy-local.ps1 [参数]

参数:
    -Production     生产环境部署 (默认为开发环境)
    -Clean          清理现有容器和数据
    -Help           显示此帮助信息
    -CheckOnly      仅检查本地镜像，不执行部署
    -UseLocal       强制使用本地镜像配置文件

示例:
    .\one-click-deploy-local.ps1                    # 开发环境部署
    .\one-click-deploy-local.ps1 -Production        # 生产环境部署
    .\one-click-deploy-local.ps1 -Clean             # 清理后重新部署
    .\one-click-deploy-local.ps1 -CheckOnly         # 仅检查镜像
    .\one-click-deploy-local.ps1 -UseLocal          # 使用本地镜像配置

特点:
    - 自动检查本地Docker镜像
    - 优先使用本地镜像，避免网络问题
    - 智能选择最佳镜像版本
    - 完整的健康检查和错误处理

"@ -ForegroundColor Cyan
}

# 检查Docker环境
function Test-DockerEnvironment {
    Write-Info "Checking Docker environment..."
    
    # 检查Docker是否安装
    try {
        $dockerVersion = docker --version 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "Docker not found"
        }
        Write-Success "Docker found: $dockerVersion"
    }
    catch {
        Write-Error "Docker is not installed or not accessible"
        Write-Info "Please install Docker Desktop and ensure it's running"
        exit 1
    }
    
    # 检查Docker Compose
    try {
        $composeVersion = docker compose version 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "Docker Compose not found"
        }
        Write-Success "Docker Compose found: $composeVersion"
    }
    catch {
        Write-Error "Docker Compose is not available"
        Write-Info "Please ensure Docker Desktop includes Docker Compose"
        exit 1
    }
    
    # 检查Docker服务状态
    try {
        docker info >$null 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "Docker daemon not running"
        }
        Write-Success "Docker daemon is running"
    }
    catch {
        Write-Error "Docker daemon is not running"
        Write-Info "Please start Docker Desktop"
        exit 1
    }
}

# 检查本地Docker镜像
function Test-LocalImages {
    Write-Info "Checking local Docker images..."
    
    # 定义所需镜像
    $requiredImages = @{
        "postgres" = @("postgres:15-alpine", "postgres:14-alpine", "postgres:13-alpine", "postgres:latest")
        "redis" = @("redis:7-alpine", "redis:6-alpine", "redis:alpine", "redis:latest")
        "nginx" = @("nginx:alpine", "nginx:latest", "nginx:stable-alpine")
    }
    
    $availableImages = @{}
    $missingImages = @{}
    
    # 获取本地所有镜像
    try {
        $localImages = docker images --format "{{.Repository}}:{{.Tag}}" 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to get local images"
        }
    }
    catch {
        Write-Error "Failed to retrieve local Docker images"
        return $false
    }
    
    # 检查每个所需镜像
    foreach ($imageType in $requiredImages.Keys) {
        $found = $false
        $availableVersions = @()
        
        foreach ($version in $requiredImages[$imageType]) {
            if ($localImages -contains $version) {
                $availableVersions += $version
                $found = $true
            }
        }
        
        if ($found) {
            $availableImages[$imageType] = $availableVersions[0]  # 使用第一个找到的版本
            Write-Success "Found $imageType image: $($availableVersions[0])"
            if ($availableVersions.Count -gt 1) {
                Write-Info "  Other available versions: $($availableVersions[1..($availableVersions.Count-1)] -join ', ')"
            }
        } else {
            $missingImages[$imageType] = $requiredImages[$imageType][0]  # 使用推荐版本
            Write-Warning "Missing $imageType image, will use: $($requiredImages[$imageType][0])"
        }
    }
    
    return @{
        "Available" = $availableImages
        "Missing" = $missingImages
        "HasMissing" = $missingImages.Count -gt 0
    }
}

# 创建本地镜像环境配置
function Set-LocalImageEnvironment {
    param($ImageInfo)
    
    Write-Info "Creating local image environment configuration..."
    
    $envContent = @"
# 太上老君系统 - 自动生成的本地镜像配置
# 生成时间: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')

# ===========================================
# 镜像配置 (基于本地可用镜像)
# ===========================================

"@
    
    # 添加可用镜像配置
    foreach ($imageType in $ImageInfo.Available.Keys) {
        $imageName = $ImageInfo.Available[$imageType]
        $envVarName = switch ($imageType) {
            "postgres" { "POSTGRES_IMAGE" }
            "redis" { "REDIS_IMAGE" }
            "nginx" { "NGINX_IMAGE" }
        }
        $envContent += "$envVarName=$imageName`n"
    }
    
    # 添加缺失镜像配置
    foreach ($imageType in $ImageInfo.Missing.Keys) {
        $imageName = $ImageInfo.Missing[$imageType]
        $envVarName = switch ($imageType) {
            "postgres" { "POSTGRES_IMAGE" }
            "redis" { "REDIS_IMAGE" }
            "nginx" { "NGINX_IMAGE" }
        }
        $envContent += "$envVarName=$imageName`n"
    }
    
    # 添加其他配置
    $envContent += @"

# ===========================================
# 数据库配置
# ===========================================
POSTGRES_DB=laojun
POSTGRES_USER=laojun
POSTGRES_PASSWORD=laojun123

# ===========================================
# Redis配置
# ===========================================
REDIS_PASSWORD=redis123

# ===========================================
# 应用配置
# ===========================================
SERVER_MODE=development
LOG_LEVEL=info

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRE_HOURS=24
"@
    
    # 写入环境配置文件
    $envFile = Join-Path $DockerDir ".env"
    $envContent | Out-File -FilePath $envFile -Encoding UTF8 -Force
    
    Write-Success "Local image environment configuration created: .env"
    
    # 显示配置摘要
    Write-Info "Image configuration summary:"
    foreach ($imageType in $ImageInfo.Available.Keys) {
        Write-Info "  $imageType`: $($ImageInfo.Available[$imageType]) (local)"
    }
    foreach ($imageType in $ImageInfo.Missing.Keys) {
        Write-Warning "  $imageType`: $($ImageInfo.Missing[$imageType]) (will download)"
    }
}

# 检查端口占用
function Test-Ports {
    Write-Info "Checking port availability..."
    
    $requiredPorts = @(80, 8080, 8081, 8082, 8888, 5432, 6379)
    $occupiedPorts = @()
    
    foreach ($port in $requiredPorts) {
        $connection = Test-NetConnection -ComputerName localhost -Port $port -WarningAction SilentlyContinue
        if ($connection.TcpTestSucceeded) {
            $occupiedPorts += $port
        }
    }
    
    if ($occupiedPorts.Count -gt 0) {
        Write-Warning "Following ports are occupied: $($occupiedPorts -join ', ')"
        Write-Info "If you need to stop processes, please handle manually or use -Clean parameter"
    } else {
        Write-Success "All required ports are available"
    }
}

# 清理现有容器
function Clear-Containers {
    Write-Info "Cleaning existing containers..."
    
    Push-Location $DockerDir
    try {
        # 尝试使用本地配置文件
        if (Test-Path "docker-compose.local.yml") {
            Write-Info "Using local Docker Compose configuration..."
            docker compose -f docker-compose.local.yml down --volumes --remove-orphans 2>$null
        } else {
            docker compose down --volumes --remove-orphans 2>$null
        }
        
        # 清理未使用的镜像和网络
        docker system prune -f >$null 2>&1
        
        Write-Success "Container cleanup completed"
    }
    catch {
        Write-Warning "Warning occurred during cleanup, but can continue"
    }
    finally {
        Pop-Location
    }
}

# 拉取缺失镜像
function Get-MissingImages {
    param($ImageInfo)
    
    if (-not $ImageInfo.HasMissing) {
        Write-Success "All required images are available locally"
        return $true
    }
    
    Write-Info "Pulling missing images..."
    
    foreach ($imageType in $ImageInfo.Missing.Keys) {
        $imageName = $ImageInfo.Missing[$imageType]
        Write-Info "Pulling $imageName..."
        
        try {
            docker pull $imageName
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Successfully pulled $imageName"
            } else {
                Write-Warning "Failed to pull $imageName, will try during compose"
            }
        }
        catch {
            Write-Warning "Error pulling $imageName`: $($_.Exception.Message)"
        }
    }
    
    return $true
}

# 构建和启动服务
function Start-Services {
    Write-Info "Building and starting services..."
    
    Push-Location $DockerDir
    try {
        # 选择配置文件
        $composeFile = if ($UseLocal -or (Test-Path "docker-compose.local.yml")) {
            "docker-compose.local.yml"
        } else {
            "docker-compose.yml"
        }
        
        Write-Info "Using Docker Compose file: $composeFile"
        
        # 构建应用镜像
        Write-Info "Building application images..."
        docker compose -f $composeFile build --no-cache
        
        # 启动服务
        Write-Info "Starting services..."
        docker compose -f $composeFile up -d
        
        Write-Success "Services started successfully"
    }
    catch {
        Write-Error "Service startup failed: $($_.Exception.Message)"
        exit 1
    }
    finally {
        Pop-Location
    }
}

# 等待服务就绪
function Wait-ServicesReady {
    Write-Info "Waiting for services to be ready..."
    
    $maxWait = 120  # 最大等待时间（秒）
    $waited = 0
    $interval = 5
    
    Push-Location $DockerDir
    try {
        while ($waited -lt $maxWait) {
            Start-Sleep $interval
            $waited += $interval
            
            # 检查关键服务的健康状态
            $healthyServices = 0
            $totalServices = 0
            
            try {
                $containers = docker compose ps --format json | ConvertFrom-Json
                
                foreach ($container in $containers) {
                    $totalServices++
                    if ($container.Health -eq "healthy" -or $container.State -eq "running") {
                        $healthyServices++
                    }
                }
                
                Write-Info "Service status: $healthyServices/$totalServices ready (wait time: ${waited}s)"
                
                if ($healthyServices -eq $totalServices -and $totalServices -gt 0) {
                    Write-Success "All services are ready"
                    return
                }
            }
            catch {
                Write-Warning "Error checking service status, continuing to wait..."
            }
        }
        
        Write-Warning "Service startup timeout, but may still be initializing"
    }
    finally {
        Pop-Location
    }
}

# 显示访问信息
function Show-AccessInfo {
    Write-Host @"

部署完成！

访问地址:
   插件市场 (首页):    http://localhost
   管理后台:           http://localhost:8888
   API文档:            http://localhost:8080/swagger
   配置中心:           http://localhost:8081

管理命令:
   查看服务状态:       docker compose ps
   查看日志:           docker compose logs -f
   停止服务:           docker compose down
   重启服务:           docker compose restart

监控命令:
   查看资源使用:       docker stats
   查看容器详情:       docker compose ps -a

提示:
   - 首次启动可能需要几分钟初始化数据库
   - 如遇问题请查看日志: docker compose logs
   - 更多帮助请参考: deploy/docs/quick-reference.md

"@ -ForegroundColor Green
}

# 主函数
function Main {
    Write-Host "太上老君系统一键部署 (本地镜像优先版本)" -ForegroundColor Cyan
    Write-Host "============================================" -ForegroundColor Cyan
    
    if ($Help) {
        Show-Help
        return
    }
    
    try {
        # 1. 检查环境
        Test-DockerEnvironment
        
        # 2. 检查本地镜像
        $imageInfo = Test-LocalImages
        
        if ($CheckOnly) {
            Write-Info "Image check completed. Use -UseLocal to deploy with local images."
            return
        }
        
        # 3. 检查端口
        Test-Ports
        
        # 4. 清理（如果需要）
        if ($Clean) {
            Clear-Containers
        }
        
        # 5. 设置本地镜像环境
        Set-LocalImageEnvironment -ImageInfo $imageInfo
        
        # 6. 拉取缺失镜像
        Get-MissingImages -ImageInfo $imageInfo
        
        # 7. 启动服务
        Start-Services
        
        # 8. 等待就绪
        Wait-ServicesReady
        
        # 9. 显示访问信息
        Show-AccessInfo
        
    }
    catch {
        Write-Error "部署失败: $($_.Exception.Message)"
        Write-Info "请检查错误信息并重试，或使用 -Help 获取帮助"
        exit 1
    }
}

# 执行主函数
Main