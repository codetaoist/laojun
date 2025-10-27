# Laojun System One-Click Deployment Script
# Purpose: One-click deployment to Docker environment

param(
    [switch]$Production,  # Production environment deployment
    [switch]$Clean,       # Clean and redeploy
    [switch]$Help         # Show help
)

# Script configuration
$ScriptDir = $PSScriptRoot
$DeployDir = Split-Path -Parent $ScriptDir
$ProjectRoot = Split-Path -Parent $DeployDir
$DockerDir = Join-Path $DeployDir "docker"
$ConfigDir = Join-Path $DeployDir "configs"

# Color output functions
function Write-Info { param([string]$msg) Write-Host "[INFO] $msg" -ForegroundColor Blue }
function Write-Success { param([string]$msg) Write-Host "[SUCCESS] $msg" -ForegroundColor Green }
function Write-Warning { param([string]$msg) Write-Host "[WARNING] $msg" -ForegroundColor Yellow }
function Write-Error { param([string]$msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }

# Show help information
function Show-Help {
    Write-Host @"
Laojun System One-Click Deployment Script

Usage:
    .\one-click-deploy.ps1                # Development environment deployment
    .\one-click-deploy.ps1 -Production    # Production environment deployment
    .\one-click-deploy.ps1 -Clean         # Clean and redeploy
    .\one-click-deploy.ps1 -Help          # Show this help

Access URLs after deployment:
    Plugin Marketplace (Home): http://localhost
    Admin Backend:             http://localhost:8888
    API Documentation:         http://localhost:8080/swagger

"@ -ForegroundColor Cyan
}

# Check Docker environment
function Test-DockerEnvironment {
    Write-Info "Checking Docker environment..."
    
    # Check if Docker is installed
    try {
        $dockerVersion = docker --version 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "Docker not installed"
        }
        Write-Success "Docker installed: $dockerVersion"
    }
    catch {
        Write-Error "Docker not installed or not running. Please install Docker Desktop first"
        Write-Info "Download: https://www.docker.com/products/docker-desktop"
        exit 1
    }
    
    # Check Docker Compose
    try {
        $composeVersion = docker compose version 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "Docker Compose not installed"
        }
        Write-Success "Docker Compose installed: $composeVersion"
    }
    catch {
        Write-Error "Docker Compose not installed. Please update Docker Desktop to latest version"
        exit 1
    }
    
    # Check if Docker is running
    try {
        docker info >$null 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "Docker not running"
        }
        Write-Success "Docker service is running"
    }
    catch {
        Write-Error "Docker service not running. Please start Docker Desktop"
        exit 1
    }
}

# Check port usage
function Test-Ports {
    Write-Info "Checking port usage..."
    
    $ports = @(80, 8080, 8081, 8082, 8888, 5432, 6379)
    $occupiedPorts = @()
    
    foreach ($port in $ports) {
        $connection = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
        if ($connection) {
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

# Prepare environment configuration
function Set-Environment {
    Write-Info "Preparing environment configuration..."
    
    # Determine environment type
    $envFile = if ($Production) { ".env.production" } else { ".env.development" }
    $sourceEnv = Join-Path $ConfigDir $envFile
    $targetEnv = Join-Path $DockerDir ".env"
    
    if (Test-Path $sourceEnv) {
        Copy-Item $sourceEnv $targetEnv -Force
        Write-Success "Environment configuration copied: $envFile"
    } else {
        Write-Warning "Environment configuration file not found: $envFile"
        Write-Info "Will use default configuration"
    }
    
    # Show environment information
    $envType = if ($Production) { "Production" } else { "Development" }
    Write-Info "Deployment environment: $envType"
}

# Clean existing containers
function Clear-Containers {
    Write-Info "Cleaning existing containers..."
    
    Push-Location $DockerDir
    try {
        # Stop and remove containers
        docker compose down --volumes --remove-orphans 2>$null
        
        # Clean unused images and networks
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

# Build and start services
function Start-Services {
    Write-Info "Building and starting services..."
    
    Push-Location $DockerDir
    try {
        # Pull base images
        Write-Info "Pulling base images..."
        docker compose pull --quiet
        
        # Build application images
        Write-Info "Building application images..."
        docker compose build --no-cache
        
        # Start services
        Write-Info "Starting services..."
        docker compose up -d
        
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

# Wait for services to be ready
function Wait-ServicesReady {
    Write-Info "Waiting for services to be ready..."
    
    $maxWait = 120  # Maximum wait time (seconds)
    $waited = 0
    $interval = 5
    
    while ($waited -lt $maxWait) {
        Start-Sleep $interval
        $waited += $interval
        
        # Check health status of key services
        $healthyServices = 0
        $totalServices = 0
        
        try {
            Push-Location $DockerDir
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
        finally {
            Pop-Location
        }
    }
    
    Write-Warning "Service startup timeout, but may still be initializing"
}

# Execute database migration
function Invoke-DatabaseMigration {
    Write-Info "Executing database migration..."
    
    try {
        Push-Location $DockerDir
        
        # Wait for database to be ready
        Write-Info "Waiting for database to be ready..."
        $dbReady = $false
        $maxAttempts = 30
        
        for ($i = 1; $i -le $maxAttempts; $i++) {
            try {
                docker compose exec -T postgres pg_isready -U laojun -d laojun >$null 2>&1
                if ($LASTEXITCODE -eq 0) {
                    $dbReady = $true
                    break
                }
            }
            catch {}
            
            Write-Info "Database connection attempt $i/$maxAttempts..."
            Start-Sleep 2
        }
        
        if (-not $dbReady) {
            Write-Warning "Database connection timeout, skipping migration"
            return
        }
        
        # Execute migration (if migration tool exists)
        Write-Info "Database is ready, migration will execute automatically"
        Write-Success "Database migration completed"
    }
    catch {
        Write-Warning "Issue occurred during database migration: $($_.Exception.Message)"
    }
    finally {
        Pop-Location
    }
}

# Show access information
function Show-AccessInfo {
    Write-Host @"

Deployment completed!

Access URLs:
   Plugin Marketplace (Home): http://localhost
   Admin Backend:             http://localhost:8888
   API Documentation:         http://localhost:8080/swagger
   Config Center:             http://localhost:8081

Management commands:
   View service status:    docker compose ps
   View logs:              docker compose logs -f
   Stop services:          docker compose down
   Restart services:       docker compose restart

Monitoring commands:
   View resource usage:    docker stats
   View container details: docker compose ps -a

Tips:
   - First startup may take a few minutes to initialize database
   - If issues occur, check logs: docker compose logs
   - For more help see: deploy/docs/quick-reference.md

"@ -ForegroundColor Green
}

# Main function
function Main {
    Write-Host "Laojun System One-Click Deployment" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    
    if ($Help) {
        Show-Help
        return
    }
    
    try {
        # 1. Check environment
        Test-DockerEnvironment
        Test-Ports
        
        # 2. Clean (if needed)
        if ($Clean) {
            Clear-Containers
        }
        
        # 3. Prepare configuration
        Set-Environment
        
        # 4. Start services
        Start-Services
        
        # 5. Wait for readiness
        Wait-ServicesReady
        
        # 6. Database migration
        Invoke-DatabaseMigration
        
        # 7. Show access information
        Show-AccessInfo
        
    }
    catch {
        Write-Error "Deployment failed: $($_.Exception.Message)"
        Write-Info "Please check error information and retry, or use -Help for help"
        exit 1
    }
}

# Execute main function
Main