# Check Local Docker Images Script
param(
    [switch]$Help
)

function Write-Info { param($Message) Write-Host "[INFO] $Message" -ForegroundColor Blue }
function Write-Success { param($Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-Warning { param($Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
function Write-Error { param($Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

function Show-Help {
    Write-Host @"
Local Docker Images Check Tool

Usage:
    .\check-local-images.ps1 [parameters]

Parameters:
    -Help           Show this help information

Features:
    - Check local Docker images
    - Show available image versions
    - Generate environment configuration file

"@ -ForegroundColor Cyan
}

function Test-LocalImages {
    Write-Info "Checking local Docker images..."
    
    $requiredImages = @{
        "postgres" = @("postgres:15-alpine", "postgres:14-alpine", "postgres:13-alpine", "postgres:latest")
        "redis" = @("redis:7-alpine", "redis:6-alpine", "redis:alpine", "redis:latest")
        "nginx" = @("nginx:alpine", "nginx:latest", "nginx:stable-alpine")
    }
    
    $availableImages = @{}
    $missingImages = @{}
    
    try {
        $localImages = docker images --format "{{.Repository}}:{{.Tag}}" 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to get local images"
        }
    }
    catch {
        Write-Error "Cannot get local Docker images list"
        return $false
    }
    
    Write-Info "Local images list:"
    $localImages | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
    Write-Host ""
    
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
            $availableImages[$imageType] = $availableVersions[0]
            Write-Success "Found $imageType image: $($availableVersions[0])"
            if ($availableVersions.Count -gt 1) {
                Write-Info "  Other available versions: $($availableVersions[1..($availableVersions.Count-1)] -join ', ')"
            }
        } else {
            $missingImages[$imageType] = $requiredImages[$imageType][0]
            Write-Warning "Missing $imageType image, recommended: $($requiredImages[$imageType][0])"
        }
    }
    
    return @{
        "Available" = $availableImages
        "Missing" = $missingImages
        "HasMissing" = $missingImages.Count -gt 0
    }
}

function Create-EnvFile {
    param($ImageInfo)
    
    Write-Info "Creating environment configuration file..."
    
    $envContent = @"
# Laojun System - Local Images Configuration
# Generated at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')

# Image Configuration
"@
    
    foreach ($imageType in $ImageInfo.Available.Keys) {
        $imageName = $ImageInfo.Available[$imageType]
        $envVarName = switch ($imageType) {
            "postgres" { "POSTGRES_IMAGE" }
            "redis" { "REDIS_IMAGE" }
            "nginx" { "NGINX_IMAGE" }
        }
        $envContent += "`n$envVarName=$imageName"
    }
    
    foreach ($imageType in $ImageInfo.Missing.Keys) {
        $imageName = $ImageInfo.Missing[$imageType]
        $envVarName = switch ($imageType) {
            "postgres" { "POSTGRES_IMAGE" }
            "redis" { "REDIS_IMAGE" }
            "nginx" { "NGINX_IMAGE" }
        }
        $envContent += "`n$envVarName=$imageName"
    }
    
    $envContent += @"

# Database Configuration
POSTGRES_DB=laojun
POSTGRES_USER=laojun
POSTGRES_PASSWORD=laojun123

# Redis Configuration
REDIS_PASSWORD=redis123

# Application Configuration
SERVER_MODE=development
LOG_LEVEL=info
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRE_HOURS=24
"@
    
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $projectRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
    $dockerDir = Join-Path $projectRoot "deploy\docker"
    $envFile = Join-Path $dockerDir ".env"
    
    $envContent | Out-File -FilePath $envFile -Encoding UTF8 -Force
    
    Write-Success "Environment configuration file created: $envFile"
    
    Write-Info "Image configuration summary:"
    foreach ($imageType in $ImageInfo.Available.Keys) {
        Write-Info "  $imageType`: $($ImageInfo.Available[$imageType]) (local available)"
    }
    foreach ($imageType in $ImageInfo.Missing.Keys) {
        Write-Warning "  $imageType`: $($ImageInfo.Missing[$imageType]) (need download)"
    }
}

function Main {
    Write-Host "Local Docker Images Check Tool" -ForegroundColor Cyan
    Write-Host "==============================" -ForegroundColor Cyan
    
    if ($Help) {
        Show-Help
        return
    }
    
    try {
        $imageInfo = Test-LocalImages
        
        if ($imageInfo) {
            Create-EnvFile -ImageInfo $imageInfo
            
            Write-Host ""
            if ($imageInfo.HasMissing) {
                Write-Warning "Some images are missing, will download during deployment"
            } else {
                Write-Success "All images are available locally, can deploy offline!"
            }
        }
    }
    catch {
        Write-Error "Check failed: $($_.Exception.Message)"
        exit 1
    }
}

Main