# Configuration Verification Script
param(
    [switch]$Detailed
)

$ErrorActionPreference = "Continue"

Write-Host "Laojun System - Configuration Verification" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# Check configuration files
Write-Host "`n1. Checking configuration files..." -ForegroundColor Yellow

$configFiles = @(
    ".env.local",
    ".env.docker", 
    "configs/config-center.local.yaml",
    "configs/admin-api.local.yaml",
    "configs/marketplace-api.local.yaml",
    "configs/database.local.yaml",
    "configs/config-center.docker.yaml",
    "configs/admin-api.docker.yaml",
    "configs/database.docker.yaml"
)

$configOK = 0
foreach ($file in $configFiles) {
    if (Test-Path $file) {
        Write-Host "  [OK] $file" -ForegroundColor Green
        $configOK++
    } else {
        Write-Host "  [MISSING] $file" -ForegroundColor Red
    }
}

Write-Host "Configuration files: $configOK/$($configFiles.Count) found" -ForegroundColor Cyan

# Check port configuration
Write-Host "`n2. Port configuration..." -ForegroundColor Yellow

Write-Host "Local development ports:" -ForegroundColor Cyan
Write-Host "  - config-center: 8093" -ForegroundColor White
Write-Host "  - admin-api: 8080" -ForegroundColor White  
Write-Host "  - marketplace-api: 8082" -ForegroundColor White
Write-Host "  - PostgreSQL: 5432" -ForegroundColor White
Write-Host "  - Redis: 6379" -ForegroundColor White

Write-Host "`nDocker environment ports:" -ForegroundColor Cyan
Write-Host "  - nginx: 80, 443" -ForegroundColor White
Write-Host "  - postgres: 5432" -ForegroundColor White
Write-Host "  - redis: 6379" -ForegroundColor White
Write-Host "  - adminer: 8090" -ForegroundColor White
Write-Host "  - redis-commander: 8091" -ForegroundColor White

# Check port usage
Write-Host "`n3. Checking port usage..." -ForegroundColor Yellow

$ports = @(8080, 8082, 8093, 5432, 6379)
$portsOK = 0
foreach ($port in $ports) {
    try {
        $connection = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
        if ($connection) {
            Write-Host "  [USED] Port $port is in use" -ForegroundColor Yellow
        } else {
            Write-Host "  [FREE] Port $port is available" -ForegroundColor Green
            $portsOK++
        }
    } catch {
        Write-Host "  [FREE] Port $port is available" -ForegroundColor Green
        $portsOK++
    }
}

Write-Host "Available ports: $portsOK/$($ports.Count)" -ForegroundColor Cyan

# Check executable files
Write-Host "`n4. Checking executable files..." -ForegroundColor Yellow

$binFiles = @(
    "bin/config-center.exe",
    "bin/admin-api.exe", 
    "bin/marketplace-api.exe"
)

$binOK = 0
foreach ($file in $binFiles) {
    if (Test-Path $file) {
        Write-Host "  [OK] $file" -ForegroundColor Green
        $binOK++
    } else {
        Write-Host "  [MISSING] $file (needs build)" -ForegroundColor Red
    }
}

Write-Host "Executable files: $binOK/$($binFiles.Count) found" -ForegroundColor Cyan

# Check startup scripts
Write-Host "`n5. Checking startup scripts..." -ForegroundColor Yellow

$scripts = @(
    "start-local.ps1",
    "start-docker.ps1"
)

$scriptsOK = 0
foreach ($script in $scripts) {
    if (Test-Path $script) {
        Write-Host "  [OK] $script" -ForegroundColor Green
        $scriptsOK++
    } else {
        Write-Host "  [MISSING] $script" -ForegroundColor Red
    }
}

Write-Host "Startup scripts: $scriptsOK/$($scripts.Count) found" -ForegroundColor Cyan

Write-Host "`nConfiguration verification completed!" -ForegroundColor Green
Write-Host "`nUsage:" -ForegroundColor Cyan
Write-Host "  Local development: .\start-local.ps1" -ForegroundColor White
Write-Host "  Docker environment: .\start-docker.ps1" -ForegroundColor White
Write-Host "  View documentation: docs/configuration-guide.md" -ForegroundColor White

# Summary
$totalChecks = $configFiles.Count + $binFiles.Count + $scripts.Count
$totalOK = $configOK + $binOK + $scriptsOK

Write-Host "`nSummary: $totalOK/$totalChecks items OK" -ForegroundColor $(if ($totalOK -eq $totalChecks) { "Green" } else { "Yellow" })

if ($totalOK -lt $totalChecks) {
    Write-Host "Some items are missing. Run build commands if needed." -ForegroundColor Yellow
}