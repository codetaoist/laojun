# Laojun 开发工具脚本
# 提供常用的开发任务快捷方式

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("start", "stop", "restart", "logs", "status", "db", "lint", "format", "generate")]
    [string]$Action,
    
    [string]$Service = "all",
    [string]$Environment = "development",
    [switch]$Follow,
    [switch]$Verbose,
    [switch]$Fix
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 开发工具 ===" -ForegroundColor Green

# 工作区根目录
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# 服务配置
$services = @(
    @{
        Name = "config-center"
        Path = "laojun-config-center"
        Type = "go"
        Port = 8080
        Command = "go run cmd/server/main.go"
        Description = "配置中心服务"
        Dependencies = @()
    },
    @{
        Name = "admin-api"
        Path = "laojun-admin-api"
        Type = "go"
        Port = 8081
        Command = "go run cmd/server/main.go"
        Description = "管理后端 API"
        Dependencies = @("config-center")
    },
    @{
        Name = "marketplace-api"
        Path = "laojun-marketplace-api"
        Type = "go"
        Port = 8082
        Command = "go run cmd/server/main.go"
        Description = "市场后端 API"
        Dependencies = @("config-center")
    },
    @{
        Name = "admin-web"
        Path = "laojun-admin-web"
        Type = "node"
        Port = 3000
        Command = "npm run dev"
        Description = "管理前端"
        Dependencies = @("admin-api")
    },
    @{
        Name = "marketplace-web"
        Path = "laojun-marketplace-web"
        Type = "node"
        Port = 3001
        Command = "npm run dev"
        Description = "市场前端"
        Dependencies = @("marketplace-api")
    }
)

# 全局进程跟踪
$global:ServiceProcesses = @{}

function Test-ServiceRunning {
    param($service)
    
    try {
        $connection = Test-NetConnection -ComputerName "localhost" -Port $service.Port -InformationLevel Quiet -WarningAction SilentlyContinue
        return $connection
    } catch {
        return $false
    }
}

function Start-Service {
    param($service)
    
    $servicePath = Join-Path $parentDir $service.Path
    if (-not (Test-Path $servicePath)) {
        Write-Warning "$($service.Name): 服务路径不存在 ($($service.Path))"
        return $false
    }
    
    # 检查服务是否已运行
    if (Test-ServiceRunning $service) {
        Write-Host "$($service.Name): 服务已在端口 $($service.Port) 运行" -ForegroundColor Yellow
        return $true
    }
    
    # 检查依赖服务
    foreach ($dep in $service.Dependencies) {
        $depService = $services | Where-Object { $_.Name -eq $dep }
        if ($depService -and -not (Test-ServiceRunning $depService)) {
            Write-Host "$($service.Name): 启动依赖服务 $dep..." -ForegroundColor Cyan
            if (-not (Start-Service $depService)) {
                Write-Error "$($service.Name): 依赖服务 $dep 启动失败"
                return $false
            }
            Start-Sleep -Seconds 2
        }
    }
    
    Write-Host "$($service.Name): 启动服务..." -ForegroundColor Cyan
    Set-Location $servicePath
    
    try {
        # 根据服务类型执行不同的启动命令
        if ($service.Type -eq "go") {
            # 检查 Go 模块
            if (-not (Test-Path "go.mod")) {
                Write-Error "$($service.Name): go.mod 文件不存在"
                return $false
            }
            
            # 下载依赖
            Write-Host "$($service.Name): 下载 Go 依赖..." -ForegroundColor Gray
            go mod download
            if ($LASTEXITCODE -ne 0) {
                Write-Error "$($service.Name): 下载依赖失败"
                return $false
            }
            
        } elseif ($service.Type -eq "node") {
            # 检查 package.json
            if (-not (Test-Path "package.json")) {
                Write-Error "$($service.Name): package.json 文件不存在"
                return $false
            }
            
            # 检查 node_modules
            if (-not (Test-Path "node_modules")) {
                Write-Host "$($service.Name): 安装 npm 依赖..." -ForegroundColor Gray
                npm install
                if ($LASTEXITCODE -ne 0) {
                    Write-Error "$($service.Name): 安装依赖失败"
                    return $false
                }
            }
        }
        
        # 启动服务进程
        $processArgs = @{
            FilePath = "powershell"
            ArgumentList = @("-Command", $service.Command)
            WorkingDirectory = $servicePath
            PassThru = $true
        }
        
        $process = Start-Process @processArgs
        $global:ServiceProcesses[$service.Name] = $process
        
        # 等待服务启动
        Write-Host "$($service.Name): 等待服务启动..." -ForegroundColor Gray
        $maxWait = 30
        $waited = 0
        
        while ($waited -lt $maxWait) {
            Start-Sleep -Seconds 1
            $waited++
            
            if (Test-ServiceRunning $service) {
                Write-Host "$($service.Name): ✓ 服务已启动 (端口 $($service.Port))" -ForegroundColor Green
                return $true
            }
            
            # 检查进程是否还在运行
            if ($process.HasExited) {
                Write-Error "$($service.Name): 服务进程意外退出 (退出码: $($process.ExitCode))"
                return $false
            }
        }
        
        Write-Warning "$($service.Name): 服务启动超时，但进程仍在运行"
        return $false
        
    } catch {
        Write-Error "$($service.Name): 启动失败: $_"
        return $false
    }
}

function Stop-Service {
    param($service)
    
    Write-Host "$($service.Name): 停止服务..." -ForegroundColor Cyan
    
    # 停止跟踪的进程
    if ($global:ServiceProcesses.ContainsKey($service.Name)) {
        $process = $global:ServiceProcesses[$service.Name]
        if (-not $process.HasExited) {
            $process.Kill()
            $process.WaitForExit(5000)
        }
        $global:ServiceProcesses.Remove($service.Name)
    }
    
    # 查找并停止占用端口的进程
    try {
        $netstat = netstat -ano | Select-String ":$($service.Port)\s"
        foreach ($line in $netstat) {
            if ($line -match '\s+(\d+)$') {
                $pid = $matches[1]
                $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
                if ($process) {
                    Write-Host "$($service.Name): 停止进程 $pid ($($process.ProcessName))" -ForegroundColor Yellow
                    Stop-Process -Id $pid -Force
                }
            }
        }
        
        Write-Host "$($service.Name): ✓ 服务已停止" -ForegroundColor Green
        return $true
        
    } catch {
        Write-Warning "$($service.Name): 停止服务时出现错误: $_"
        return $false
    }
}

function Get-ServiceStatus {
    param($service)
    
    $isRunning = Test-ServiceRunning $service
    $processInfo = ""
    
    if ($global:ServiceProcesses.ContainsKey($service.Name)) {
        $process = $global:ServiceProcesses[$service.Name]
        if (-not $process.HasExited) {
            $processInfo = " (PID: $($process.Id))"
        }
    }
    
    $status = if ($isRunning) { "运行中$processInfo" } else { "已停止" }
    $color = if ($isRunning) { "Green" } else { "Red" }
    
    return @{
        Status = $status
        Color = $color
        Running = $isRunning
    }
}

function Show-ServiceLogs {
    param($service, $follow = $false)
    
    $servicePath = Join-Path $parentDir $service.Path
    $logPath = Join-Path $servicePath "logs"
    
    if (-not (Test-Path $logPath)) {
        Write-Warning "$($service.Name): 日志目录不存在 ($logPath)"
        return
    }
    
    $logFiles = Get-ChildItem -Path $logPath -Filter "*.log" | Sort-Object LastWriteTime -Descending
    if (-not $logFiles) {
        Write-Warning "$($service.Name): 未找到日志文件"
        return
    }
    
    $latestLog = $logFiles[0].FullName
    Write-Host "$($service.Name): 显示日志 ($($logFiles[0].Name))" -ForegroundColor Cyan
    
    if ($follow) {
        Get-Content -Path $latestLog -Tail 50 -Wait
    } else {
        Get-Content -Path $latestLog -Tail 100
    }
}

function Invoke-Linting {
    param($repos)
    
    Write-Host "代码检查..." -ForegroundColor Cyan
    
    foreach ($repo in $repos) {
        $repoPath = Join-Path $parentDir $repo.Path
        if (-not (Test-Path $repoPath)) {
            continue
        }
        
        Write-Host "$($repo.Name): 检查代码..." -ForegroundColor Yellow
        Set-Location $repoPath
        
        if ($repo.Type -eq "go") {
            # Go 代码检查
            Write-Host "  运行 go vet..." -ForegroundColor Gray
            go vet ./...
            
            Write-Host "  运行 go fmt 检查..." -ForegroundColor Gray
            $fmtOutput = go fmt ./...
            if ($fmtOutput) {
                Write-Warning "  发现格式问题: $fmtOutput"
            }
            
            # 如果安装了 golangci-lint
            if (Get-Command "golangci-lint" -ErrorAction SilentlyContinue) {
                Write-Host "  运行 golangci-lint..." -ForegroundColor Gray
                golangci-lint run
            }
            
        } elseif ($repo.Type -eq "node") {
            # Node.js 代码检查
            if (Test-Path "package.json") {
                $packageJson = Get-Content "package.json" | ConvertFrom-Json
                
                if ($packageJson.scripts.lint) {
                    Write-Host "  运行 npm run lint..." -ForegroundColor Gray
                    npm run lint
                }
                
                if ($packageJson.scripts.typecheck) {
                    Write-Host "  运行类型检查..." -ForegroundColor Gray
                    npm run typecheck
                }
            }
        }
        
        Write-Host "  ✓ 检查完成" -ForegroundColor Green
    }
}

function Invoke-Formatting {
    param($repos)
    
    Write-Host "代码格式化..." -ForegroundColor Cyan
    
    foreach ($repo in $repos) {
        $repoPath = Join-Path $parentDir $repo.Path
        if (-not (Test-Path $repoPath)) {
            continue
        }
        
        Write-Host "$($repo.Name): 格式化代码..." -ForegroundColor Yellow
        Set-Location $repoPath
        
        if ($repo.Type -eq "go") {
            # Go 代码格式化
            Write-Host "  运行 go fmt..." -ForegroundColor Gray
            go fmt ./...
            
            # 如果安装了 goimports
            if (Get-Command "goimports" -ErrorAction SilentlyContinue) {
                Write-Host "  运行 goimports..." -ForegroundColor Gray
                goimports -w .
            }
            
        } elseif ($repo.Type -eq "node") {
            # Node.js 代码格式化
            if (Test-Path "package.json") {
                $packageJson = Get-Content "package.json" | ConvertFrom-Json
                
                if ($packageJson.scripts.format) {
                    Write-Host "  运行 npm run format..." -ForegroundColor Gray
                    npm run format
                } elseif (Get-Command "prettier" -ErrorAction SilentlyContinue) {
                    Write-Host "  运行 prettier..." -ForegroundColor Gray
                    prettier --write .
                }
            }
        }
        
        Write-Host "  ✓ 格式化完成" -ForegroundColor Green
    }
}

# 主逻辑
Set-Location $workspaceRoot

# 过滤服务
$targetServices = if ($Service -eq "all") {
    $services
} else {
    $services | Where-Object { $_.Name -eq $Service }
}

if (-not $targetServices -and $Service -ne "all") {
    Write-Error "未找到服务: $Service"
    Write-Host "可用服务:" -ForegroundColor Yellow
    foreach ($svc in $services) {
        Write-Host "  - $($svc.Name) : $($svc.Description)" -ForegroundColor White
    }
    exit 1
}

switch ($Action) {
    "start" {
        Write-Host "启动服务..." -ForegroundColor Cyan
        foreach ($svc in $targetServices) {
            Start-Service $svc
        }
    }
    
    "stop" {
        Write-Host "停止服务..." -ForegroundColor Cyan
        # 反向停止（先停止依赖者）
        $reversedServices = $targetServices | Sort-Object { $_.Dependencies.Count } -Descending
        foreach ($svc in $reversedServices) {
            Stop-Service $svc
        }
    }
    
    "restart" {
        Write-Host "重启服务..." -ForegroundColor Cyan
        foreach ($svc in $targetServices) {
            Stop-Service $svc
            Start-Sleep -Seconds 1
            Start-Service $svc
        }
    }
    
    "status" {
        Write-Host "`n服务状态:" -ForegroundColor Cyan
        Write-Host ("{0,-20} {1,-15} {2,-10} {3}" -f "服务", "状态", "端口", "描述") -ForegroundColor White
        Write-Host ("-" * 60) -ForegroundColor Gray
        
        foreach ($svc in $services) {
            $status = Get-ServiceStatus $svc
            Write-Host ("{0,-20} {1,-15} {2,-10} {3}" -f $svc.Name, $status.Status, $svc.Port, $svc.Description) -ForegroundColor $status.Color
        }
    }
    
    "logs" {
        if ($Service -eq "all") {
            Write-Error "查看日志需要指定具体服务"
            exit 1
        }
        
        Show-ServiceLogs $targetServices[0] $Follow
    }
    
    "db" {
        Write-Host "数据库管理功能开发中..." -ForegroundColor Yellow
        Write-Host "将来会支持:" -ForegroundColor Gray
        Write-Host "  - 数据库迁移" -ForegroundColor Gray
        Write-Host "  - 数据备份恢复" -ForegroundColor Gray
        Write-Host "  - 数据库连接测试" -ForegroundColor Gray
    }
    
    "lint" {
        $repos = $services | Where-Object { $_.Name -in $targetServices.Name }
        Invoke-Linting $repos
    }
    
    "format" {
        $repos = $services | Where-Object { $_.Name -in $targetServices.Name }
        Invoke-Formatting $repos
    }
    
    "generate" {
        Write-Host "代码生成功能开发中..." -ForegroundColor Yellow
        Write-Host "将来会支持:" -ForegroundColor Gray
        Write-Host "  - API 文档生成" -ForegroundColor Gray
        Write-Host "  - 数据库模型生成" -ForegroundColor Gray
        Write-Host "  - 客户端 SDK 生成" -ForegroundColor Gray
    }
}

Write-Host "`n开发工具执行完成！" -ForegroundColor Green