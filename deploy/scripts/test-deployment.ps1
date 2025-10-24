# 太上老君系统部署测试脚本
# 用途：测试部署是否成功

param(
    [switch]$Detailed  # 详细测试
)

# 颜色输出函数
function Write-Info { param([string]$msg) Write-Host "[INFO] $msg" -ForegroundColor Blue }
function Write-Success { param([string]$msg) Write-Host "[SUCCESS] $msg" -ForegroundColor Green }
function Write-Warning { param([string]$msg) Write-Host "[WARNING] $msg" -ForegroundColor Yellow }
function Write-Error { param([string]$msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }

# 测试HTTP端点
function Test-HttpEndpoint {
    param(
        [string]$Url,
        [string]$Name,
        [int]$TimeoutSeconds = 10
    )
    
    try {
        $response = Invoke-WebRequest -Uri $Url -TimeoutSec $TimeoutSeconds -UseBasicParsing -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            Write-Success "$Name 可访问 ($Url)"
            return $true
        } else {
            Write-Warning "$Name 返回状态码: $($response.StatusCode) ($Url)"
            return $false
        }
    }
    catch {
        Write-Error "$Name 无法访问 ($Url): $($_.Exception.Message)"
        return $false
    }
}

# 测试Docker容器状态
function Test-DockerContainers {
    Write-Info "检查Docker容器状态..."
    
    try {
        Push-Location (Join-Path $PSScriptRoot "deploy\docker")
        
        $containers = docker compose ps --format json | ConvertFrom-Json
        $runningCount = 0
        $totalCount = 0
        
        Write-Host "`n容器状态:" -ForegroundColor Cyan
        Write-Host "----------------------------------------"
        
        foreach ($container in $containers) {
            $totalCount++
            $status = $container.State
            $health = $container.Health
            
            $statusColor = switch ($status) {
                "running" { "Green" }
                "exited" { "Red" }
                default { "Yellow" }
            }
            
            $healthInfo = if ($health) { " ($health)" } else { "" }
            Write-Host "  $($container.Service): " -NoNewline
            Write-Host "$status$healthInfo" -ForegroundColor $statusColor
            
            if ($status -eq "running") {
                $runningCount++
            }
        }
        
        Write-Host "----------------------------------------"
        Write-Host "运行中: $runningCount/$totalCount" -ForegroundColor $(if ($runningCount -eq $totalCount) { "Green" } else { "Yellow" })
        
        return $runningCount -eq $totalCount -and $totalCount -gt 0
    }
    catch {
        Write-Error "检查容器状态失败: $($_.Exception.Message)"
        return $false
    }
    finally {
        Pop-Location
    }
}

# 测试服务端点
function Test-ServiceEndpoints {
    Write-Info "测试服务端点..."
    
    $endpoints = @(
        @{ Url = "http://localhost"; Name = "插件市场（主页）" },
        @{ Url = "http://localhost:8888"; Name = "管理后台" },
        @{ Url = "http://localhost:8080/health"; Name = "管理API健康检查" },
        @{ Url = "http://localhost:8082/health"; Name = "插件市场API健康检查" },
        @{ Url = "http://localhost:8081/health"; Name = "配置中心健康检查" }
    )
    
    $successCount = 0
    
    Write-Host "`n服务端点测试:" -ForegroundColor Cyan
    Write-Host "----------------------------------------"
    
    foreach ($endpoint in $endpoints) {
        if (Test-HttpEndpoint -Url $endpoint.Url -Name $endpoint.Name) {
            $successCount++
        }
    }
    
    Write-Host "----------------------------------------"
    Write-Host "可访问: $successCount/$($endpoints.Count)" -ForegroundColor $(if ($successCount -eq $endpoints.Count) { "Green" } else { "Yellow" })
    
    return $successCount -eq $endpoints.Count
}

# 测试数据库连接
function Test-DatabaseConnection {
    Write-Info "测试数据库连接..."
    
    try {
        Push-Location (Join-Path $PSScriptRoot "deploy\docker")
        
        $result = docker compose exec -T postgres pg_isready -U laojun -d laojun 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "数据库连接正常"
            return $true
        } else {
            Write-Error "数据库连接失败"
            return $false
        }
    }
    catch {
        Write-Error "测试数据库连接时出错: $($_.Exception.Message)"
        return $false
    }
    finally {
        Pop-Location
    }
}

# 测试Redis连接
function Test-RedisConnection {
    Write-Info "测试Redis连接..."
    
    try {
        Push-Location (Join-Path $PSScriptRoot "deploy\docker")
        
        $result = docker compose exec -T redis redis-cli ping 2>$null
        if ($result -match "PONG") {
            Write-Success "Redis连接正常"
            return $true
        } else {
            Write-Error "Redis连接失败"
            return $false
        }
    }
    catch {
        Write-Error "测试Redis连接时出错: $($_.Exception.Message)"
        return $false
    }
    finally {
        Pop-Location
    }
}

# 显示系统资源使用情况
function Show-ResourceUsage {
    Write-Info "系统资源使用情况..."
    
    try {
        Write-Host "`nDocker容器资源使用:" -ForegroundColor Cyan
        Write-Host "----------------------------------------"
        docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
        
        Write-Host "`nDocker系统信息:" -ForegroundColor Cyan
        Write-Host "----------------------------------------"
        docker system df
    }
    catch {
        Write-Warning "无法获取资源使用信息"
    }
}

# 主测试函数
function Main {
    Write-Host "🧪 太上老君系统部署测试" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    
    $allTestsPassed = $true
    
    # 基础测试
    $containerTest = Test-DockerContainers
    $endpointTest = Test-ServiceEndpoints
    $dbTest = Test-DatabaseConnection
    $redisTest = Test-RedisConnection
    
    $allTestsPassed = $containerTest -and $endpointTest -and $dbTest -and $redisTest
    
    # 详细测试
    if ($Detailed) {
        Show-ResourceUsage
    }
    
    # 测试结果总结
    Write-Host "`n" + "="*50 -ForegroundColor Cyan
    if ($allTestsPassed) {
        Write-Success "🎉 所有测试通过！系统部署成功！"
        Write-Host "`n📱 访问地址:" -ForegroundColor Green
        Write-Host "   插件市场（主页）: http://localhost" -ForegroundColor White
        Write-Host "   管理后台:        http://localhost:8888" -ForegroundColor White
        Write-Host "   API文档:         http://localhost:8080/swagger" -ForegroundColor White
    } else {
        Write-Warning "⚠️  部分测试失败，请检查服务状态"
        Write-Info "建议操作:"
        Write-Info "1. 等待几分钟让服务完全启动"
        Write-Info "2. 查看日志: docker compose logs"
        Write-Info "3. 重新运行测试: .\test-deployment.ps1"
    }
    
    Write-Host "`n💡 提示:" -ForegroundColor Cyan
    Write-Host "   - 使用 -Detailed 参数查看详细信息" -ForegroundColor White
    Write-Host "   - 查看日志: docker compose logs -f" -ForegroundColor White
    Write-Host "   - 重启服务: docker compose restart" -ForegroundColor White
}

# 执行主函数
Main