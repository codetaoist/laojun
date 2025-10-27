# Laojun 开发环境检查脚本
# 用于验证开发环境的完整性和配置

param(
    [switch]$Verbose,
    [switch]$Fix
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 开发环境检查 ===" -ForegroundColor Green

# 检查结果统计
$checkResults = @{
    Passed = 0
    Failed = 0
    Warnings = 0
    Fixed = 0
}

function Test-Command {
    param($command)
    try {
        Get-Command $command -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

function Test-Version {
    param($command, $minVersion, $versionArg = "--version")
    try {
        $output = & $command $versionArg 2>&1
        if ($LASTEXITCODE -ne 0) {
            return $false, "命令执行失败"
        }
        
        # 提取版本号（简单的正则匹配）
        if ($output -match "(\d+\.\d+(?:\.\d+)?)") {
            $version = [Version]$matches[1]
            $minVer = [Version]$minVersion
            if ($version -ge $minVer) {
                return $true, $version.ToString()
            } else {
                return $false, "版本过低: $($version.ToString()) < $minVersion"
            }
        } else {
            return $false, "无法解析版本号"
        }
    } catch {
        return $false, $_.Exception.Message
    }
}

function Write-CheckResult {
    param($name, $passed, $message = "", $warning = $false)
    
    if ($passed) {
        Write-Host "  ✓ $name" -ForegroundColor Green
        if ($message -and $Verbose) {
            Write-Host "    $message" -ForegroundColor Gray
        }
        $checkResults.Passed++
    } elseif ($warning) {
        Write-Host "  ⚠ $name" -ForegroundColor Yellow
        if ($message) {
            Write-Host "    $message" -ForegroundColor Yellow
        }
        $checkResults.Warnings++
    } else {
        Write-Host "  ✗ $name" -ForegroundColor Red
        if ($message) {
            Write-Host "    $message" -ForegroundColor Red
        }
        $checkResults.Failed++
    }
}

# 1. 检查基础工具
Write-Host "`n1. 基础工具检查" -ForegroundColor Cyan

# Go 环境
if (Test-Command "go") {
    $goVersionOk, $goVersionMsg = Test-Version "go" "1.19" "version"
    Write-CheckResult "Go 环境" $goVersionOk $goVersionMsg
    
    if ($goVersionOk) {
        # 检查 Go 环境变量
        $gopath = $env:GOPATH
        $goroot = go env GOROOT
        Write-CheckResult "GOROOT" ($goroot -ne "") $goroot
        
        if ($Verbose) {
            Write-Host "    GOPATH: $gopath" -ForegroundColor Gray
            Write-Host "    GOPROXY: $(go env GOPROXY)" -ForegroundColor Gray
            Write-Host "    GOSUMDB: $(go env GOSUMDB)" -ForegroundColor Gray
        }
    }
} else {
    Write-CheckResult "Go 环境" $false "Go 未安装或不在 PATH 中"
}

# Git 环境
if (Test-Command "git") {
    $gitVersionOk, $gitVersionMsg = Test-Version "git" "2.20"
    Write-CheckResult "Git 环境" $gitVersionOk $gitVersionMsg
    
    if ($gitVersionOk) {
        # 检查 Git 配置
        try {
            $gitUser = git config --global user.name 2>$null
            $gitEmail = git config --global user.email 2>$null
            Write-CheckResult "Git 用户配置" ($gitUser -and $gitEmail) "用户: $gitUser, 邮箱: $gitEmail"
        } catch {
            Write-CheckResult "Git 用户配置" $false "未配置用户信息" $true
        }
    }
} else {
    Write-CheckResult "Git 环境" $false "Git 未安装或不在 PATH 中"
}

# Node.js 环境（可选）
if (Test-Command "node") {
    $nodeVersionOk, $nodeVersionMsg = Test-Version "node" "16.0"
    Write-CheckResult "Node.js 环境" $nodeVersionOk $nodeVersionMsg
    
    if ($nodeVersionOk -and (Test-Command "npm")) {
        $npmVersionOk, $npmVersionMsg = Test-Version "npm" "8.0"
        Write-CheckResult "npm 环境" $npmVersionOk $npmVersionMsg
    }
} else {
    Write-CheckResult "Node.js 环境" $false "Node.js 未安装（前端开发需要）" $true
}

# Docker 环境（可选）
if (Test-Command "docker") {
    $dockerVersionOk, $dockerVersionMsg = Test-Version "docker" "20.0"
    Write-CheckResult "Docker 环境" $dockerVersionOk $dockerVersionMsg
    
    if ($dockerVersionOk) {
        try {
            docker info 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-CheckResult "Docker 服务" $true "Docker 守护进程运行正常"
            } else {
                Write-CheckResult "Docker 服务" $false "Docker 守护进程未运行" $true
            }
        } catch {
            Write-CheckResult "Docker 服务" $false "无法连接 Docker 守护进程" $true
        }
    }
} else {
    Write-CheckResult "Docker 环境" $false "Docker 未安装（容器化部署需要）" $true
}

# 2. 工作区结构检查
Write-Host "`n2. 工作区结构检查" -ForegroundColor Cyan

$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# 检查工作区文件
$workspaceFiles = @(
    @{ Path = "go.work"; Required = $true; Description = "Go 工作区配置" },
    @{ Path = "README.md"; Required = $true; Description = "工作区说明文档" },
    @{ Path = "scripts"; Required = $true; Description = "开发脚本目录" }
)

foreach ($file in $workspaceFiles) {
    $filePath = Join-Path $workspaceRoot $file.Path
    $exists = Test-Path $filePath
    Write-CheckResult $file.Description $exists $file.Path
}

# 检查子仓库
$repositories = @(
    "laojun-shared",
    "laojun-plugins", 
    "laojun-config-center",
    "laojun-admin-api",
    "laojun-marketplace-api",
    "laojun-admin-web",
    "laojun-marketplace-web"
)

Write-Host "`n3. 子仓库检查" -ForegroundColor Cyan

foreach ($repo in $repositories) {
    $repoPath = Join-Path $parentDir $repo
    $exists = Test-Path $repoPath
    Write-CheckResult $repo $exists $repoPath
    
    if ($exists -and $Verbose) {
        # 检查仓库状态
        Set-Location $repoPath
        try {
            $gitStatus = git status --porcelain 2>$null
            if ($LASTEXITCODE -eq 0) {
                if ($gitStatus) {
                    Write-Host "    有未提交的更改" -ForegroundColor Yellow
                } else {
                    Write-Host "    工作目录干净" -ForegroundColor Gray
                }
            }
        } catch {
            Write-Host "    非 Git 仓库或 Git 错误" -ForegroundColor Yellow
        }
    }
}

# 4. Go 工作区检查
Write-Host "`n4. Go 工作区检查" -ForegroundColor Cyan

Set-Location $workspaceRoot

if (Test-Path "go.work") {
    try {
        # 检查工作区同步状态
        go work sync 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-CheckResult "工作区同步" $true "go.work 同步正常"
        } else {
            Write-CheckResult "工作区同步" $false "工作区需要同步"
            if ($Fix) {
                Write-Host "    正在修复..." -ForegroundColor Yellow
                go work sync
                if ($LASTEXITCODE -eq 0) {
                    Write-Host "    ✓ 已修复" -ForegroundColor Green
                    $checkResults.Fixed++
                }
            }
        }
        
        # 检查模块状态
        $workContent = Get-Content "go.work" -Raw
        $moduleCount = ($workContent | Select-String "use \." -AllMatches).Matches.Count
        Write-CheckResult "工作区模块" ($moduleCount -gt 0) "包含 $moduleCount 个模块"
        
    } catch {
        Write-CheckResult "工作区验证" $false $_.Exception.Message
    }
} else {
    Write-CheckResult "Go 工作区文件" $false "go.work 文件不存在"
}

# 5. 网络连接检查
Write-Host "`n5. 网络连接检查" -ForegroundColor Cyan

$networkTests = @(
    @{ Host = "proxy.golang.org"; Port = 443; Description = "Go 模块代理" },
    @{ Host = "github.com"; Port = 443; Description = "GitHub" },
    @{ Host = "registry.npmjs.org"; Port = 443; Description = "npm 注册表" }
)

foreach ($test in $networkTests) {
    try {
        $tcpClient = New-Object System.Net.Sockets.TcpClient
        $tcpClient.ConnectAsync($test.Host, $test.Port).Wait(3000)
        if ($tcpClient.Connected) {
            Write-CheckResult $test.Description $true "$($test.Host):$($test.Port)"
            $tcpClient.Close()
        } else {
            Write-CheckResult $test.Description $false "连接超时" $true
        }
    } catch {
        Write-CheckResult $test.Description $false "连接失败" $true
    }
}

# 6. 磁盘空间检查
Write-Host "`n6. 系统资源检查" -ForegroundColor Cyan

try {
    $drive = Get-WmiObject -Class Win32_LogicalDisk | Where-Object { $_.DeviceID -eq (Split-Path $workspaceRoot -Qualifier) }
    $freeSpaceGB = [math]::Round($drive.FreeSpace / 1GB, 2)
    $totalSpaceGB = [math]::Round($drive.Size / 1GB, 2)
    $freeSpacePercent = [math]::Round(($drive.FreeSpace / $drive.Size) * 100, 1)
    
    $spaceOk = $freeSpaceGB -gt 5  # 至少 5GB 可用空间
    Write-CheckResult "磁盘空间" $spaceOk "$freeSpaceGB GB 可用 ($freeSpacePercent% of $totalSpaceGB GB)"
    
} catch {
    Write-CheckResult "磁盘空间" $false "无法获取磁盘信息" $true
}

# 内存检查
try {
    $memory = Get-WmiObject -Class Win32_ComputerSystem
    $totalMemoryGB = [math]::Round($memory.TotalPhysicalMemory / 1GB, 2)
    $memoryOk = $totalMemoryGB -gt 4  # 至少 4GB 内存
    Write-CheckResult "系统内存" $memoryOk "$totalMemoryGB GB"
} catch {
    Write-CheckResult "系统内存" $false "无法获取内存信息" $true
}

# 检查总结
Write-Host "`n=== 检查总结 ===" -ForegroundColor Green
Write-Host "通过: $($checkResults.Passed)" -ForegroundColor Green
Write-Host "失败: $($checkResults.Failed)" -ForegroundColor Red
Write-Host "警告: $($checkResults.Warnings)" -ForegroundColor Yellow
if ($Fix) {
    Write-Host "已修复: $($checkResults.Fixed)" -ForegroundColor Cyan
}

# 建议
if ($checkResults.Failed -gt 0) {
    Write-Host "`n建议:" -ForegroundColor Yellow
    Write-Host "- 安装缺失的工具和依赖" -ForegroundColor White
    Write-Host "- 配置必要的环境变量" -ForegroundColor White
    Write-Host "- 运行 .\scripts\setup.ps1 初始化环境" -ForegroundColor White
    if (-not $Fix) {
        Write-Host "- 使用 -Fix 参数自动修复部分问题" -ForegroundColor White
    }
}

if ($checkResults.Warnings -gt 0) {
    Write-Host "`n注意:" -ForegroundColor Yellow
    Write-Host "- 某些工具是可选的，但建议安装以获得完整的开发体验" -ForegroundColor White
    Write-Host "- 网络连接问题可能影响依赖下载" -ForegroundColor White
}

if ($checkResults.Failed -eq 0 -and $checkResults.Warnings -eq 0) {
    Write-Host "`n🎉 开发环境检查通过！可以开始开发了。" -ForegroundColor Green
} elseif ($checkResults.Failed -eq 0) {
    Write-Host "`n✅ 开发环境基本就绪，存在一些可选组件的警告。" -ForegroundColor Yellow
} else {
    Write-Host "`n❌ 开发环境存在问题，请根据上述建议进行修复。" -ForegroundColor Red
    exit 1
}