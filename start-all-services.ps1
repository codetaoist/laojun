# 启动所有Laojun服务的PowerShell脚本
# 使用统一端口配置

param(
    [switch]$Build = $false,
    [switch]$Stop = $false,
    [switch]$Status = $false
)

# 服务配置
$services = @(
    @{
        Name = "Discovery"
        Path = "laojun-discovery"
        Executable = "discovery.exe"
        BuildCmd = "go build -o discovery.exe ./cmd"
    },
    @{
        Name = "Config Center"
        Path = "laojun-config-center"
        Executable = "config-center.exe"
        BuildCmd = "go build -o config-center.exe ./cmd/config-center"
    },
    @{
        Name = "Gateway"
        Path = "laojun-gateway"
        Executable = "gateway.exe"
        BuildCmd = "go build -o gateway.exe ./cmd"
    },
    @{
        Name = "Admin API"
        Path = "laojun-admin-api"
        Executable = "admin-api.exe"
        BuildCmd = "go build -o admin-api.exe ./cmd/admin-api"
    },
    @{
        Name = "Marketplace API"
        Path = "laojun-marketplace-api"
        Executable = "marketplace-api.exe"
        BuildCmd = "go build -o marketplace-api.exe ./cmd/marketplace-api"
    },
    @{
        Name = "Plugins"
        Path = "laojun-plugins"
        Executable = "plugin-manager.exe"
        BuildCmd = "go build -o plugin-manager.exe ./cmd/plugin-manager"
    },
    @{
        Name = "Monitoring"
        Path = "laojun-monitoring"
        Executable = "monitoring.exe"
        BuildCmd = "go build -o monitoring.exe ./cmd"
    }
)

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# 检查服务状态
function Get-ServiceStatus {
    param([string]$ProcessName)
    
    $process = Get-Process -Name $ProcessName.Replace(".exe", "") -ErrorAction SilentlyContinue
    if ($process) {
        return @{
            Running = $true
            PID = $process.Id
            StartTime = $process.StartTime
        }
    }
    return @{ Running = $false }
}

# 停止服务
function Stop-ServiceProcess {
    param([string]$ProcessName)
    
    $process = Get-Process -Name $ProcessName.Replace(".exe", "") -ErrorAction SilentlyContinue
    if ($process) {
        Write-ColorOutput "Stopping $ProcessName (PID: $($process.Id))" "Yellow"
        Stop-Process -Id $process.Id -Force
        Start-Sleep -Seconds 2
        return $true
    }
    return $false
}

# 构建服务
function Build-ServiceApp {
    param([hashtable]$Service)
    
    Write-ColorOutput "Building $($Service.Name)..." "Cyan"
    
    if (!(Test-Path $Service.Path)) {
        Write-ColorOutput "Error: Directory $($Service.Path) does not exist" "Red"
        return $false
    }
    
    Push-Location $Service.Path
    try {
        $result = Invoke-Expression $Service.BuildCmd 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ $($Service.Name) built successfully" "Green"
            return $true
        } else {
            Write-ColorOutput "✗ $($Service.Name) build failed: $result" "Red"
            return $false
        }
    }
    finally {
        Pop-Location
    }
}

# 启动服务
function Start-ServiceApp {
    param([hashtable]$Service)
    
    $executablePath = Join-Path $Service.Path $Service.Executable
    
    if (!(Test-Path $executablePath)) {
        Write-ColorOutput "Error: Executable $executablePath does not exist" "Red"
        return $false
    }
    
    # 检查是否已经运行
    $status = Get-ServiceStatus $Service.Executable
    if ($status.Running) {
        Write-ColorOutput "$($Service.Name) is already running (PID: $($status.PID))" "Yellow"
        return $true
    }
    
    Write-ColorOutput "Starting $($Service.Name)..." "Cyan"
    
    try {
        Push-Location $Service.Path
        Start-Process -FilePath ".\$($Service.Executable)" -WindowStyle Minimized
        Start-Sleep -Seconds 3
        
        $status = Get-ServiceStatus $Service.Executable
        if ($status.Running) {
            Write-ColorOutput "✓ $($Service.Name) started successfully (PID: $($status.PID))" "Green"
            return $true
        } else {
            Write-ColorOutput "✗ $($Service.Name) failed to start" "Red"
            return $false
        }
    }
    finally {
        Pop-Location
    }
}

# 显示服务状态
if ($Status) {
    Write-ColorOutput "`n=== Laojun Service Status ===" "Magenta"
    Write-ColorOutput "Service Name`t`tStatus`t`tPID`tStart Time" "White"
    Write-ColorOutput "----------------------------------------" "White"
    
    foreach ($service in $services) {
        $status = Get-ServiceStatus $service.Executable
        $statusText = if ($status.Running) { "Running" } else { "Stopped" }
        $color = if ($status.Running) { "Green" } else { "Red" }
        $pid = if ($status.Running) { $status.PID } else { "-" }
        $startTime = if ($status.Running -and $status.StartTime) { $status.StartTime.ToString("HH:mm:ss") } else { "-" }
        
        Write-Host "$($service.Name.PadRight(20))" -NoNewline
        Write-Host "$statusText".PadRight(12) -ForegroundColor $color -NoNewline
        Write-Host "$pid".PadRight(8) -NoNewline
        Write-Host $startTime
    }
    exit
}

# 停止所有服务
if ($Stop) {
    Write-ColorOutput "`n=== Stop All Laojun Services ===" "Magenta"
    
    $stoppedCount = 0
    foreach ($service in $services) {
        if (Stop-ServiceProcess $service.Executable) {
            $stoppedCount++
        }
    }
    
    Write-ColorOutput "`nStopped $stoppedCount services" "Green"
    exit
}

# 构建服务
if ($Build) {
    Write-ColorOutput "`n=== Build All Laojun Services ===" "Magenta"
    
    $buildSuccessCount = 0
    foreach ($service in $services) {
        if (Build-ServiceApp $service) {
            $buildSuccessCount++
        }
    }
    
    Write-ColorOutput "`nBuild completed: $buildSuccessCount/$($services.Count) services built successfully" "Green"
    
    if ($buildSuccessCount -lt $services.Count) {
        Write-ColorOutput "Some services failed to build, please check error messages" "Red"
        exit 1
    }
}

# 启动所有服务
Write-ColorOutput "`n=== Start All Laojun Services ===" "Magenta"

$startSuccessCount = 0
foreach ($service in $services) {
    if (Start-ServiceApp $service) {
        $startSuccessCount++
    }
    Start-Sleep -Seconds 2  # 服务间启动间隔
}

Write-ColorOutput "`nStartup completed: $startSuccessCount/$($services.Count) services started successfully" "Green"

if ($startSuccessCount -eq $services.Count) {
    Write-ColorOutput "`nAll services started successfully!" "Green"
    Write-ColorOutput "Use 'powershell .\start-all-services.ps1 -Status' to check service status" "Cyan"
    Write-ColorOutput "Use 'powershell .\start-all-services.ps1 -Stop' to stop all services" "Cyan"
} else {
    Write-ColorOutput "`nSome services failed to start, please check error messages" "Red"
    exit 1
}