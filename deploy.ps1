# Laojun System Docker Deployment Script
# Features:
# 1. Service separation architecture
# 2. Plugin marketplace as default homepage (port 80)
# 3. Admin backend accessible via port 8888
# 4. Unified environment variable management
# 5. Simplified network configuration

param(
    [Parameter(Position=0)]
    [ValidateSet("start", "stop", "restart", "build", "logs", "status", "clean", "help")]
    [string]$Action = "help",
    
    [Parameter(Position=1)]
    [ValidateSet("dev", "staging", "prod")]
    [string]$Environment = "dev",
    
    [switch]$Force
)

# Set paths
$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$COMPOSE_FILE = "$SCRIPT_DIR\deploy\docker\docker-compose.yml"
$ENV_FILE = "$SCRIPT_DIR\deploy\configs\.env"
$LOG_FILE = "$SCRIPT_DIR\logs\deploy.log"

# Create log directory
if (-not (Test-Path (Split-Path $LOG_FILE))) {
    New-Item -ItemType Directory -Path (Split-Path $LOG_FILE) -Force | Out-Null
}

# Logging function
function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] $Message"
    Write-Host $logMessage
    Add-Content -Path $LOG_FILE -Value $logMessage
}

# Error handling function
function Handle-Error {
    param([string]$Message)
    Write-Log "ERROR: $Message"
    Write-Host "ERROR: $Message" -ForegroundColor Red
    exit 1
}

# Check if Docker is running
function Test-Docker {
    try {
        docker version | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Check if docker-compose is available
function Test-DockerCompose {
    try {
        docker-compose version | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Show help information
function Show-Help {
    Write-Host @"
Laojun System Deployment Script

Usage: .\deploy.ps1 [action] [environment] [options]

Actions:
  start     Start services
  stop      Stop services
  restart   Restart services
  build     Build images
  logs      View logs
  status    View status
  clean     Clean resources
  help      Show help

Environments:
  dev       Development environment (default)
  staging   Staging environment
  prod      Production environment

Options:
  -Force    Force execution, skip confirmation

Examples:
  .\deploy.ps1 start dev
  .\deploy.ps1 build prod
  .\deploy.ps1 logs
  .\deploy.ps1 clean -Force
"@
}

# Build images
function Build-Images {
    Write-Log "Building Docker images..."
    
    try {
        docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE build
        Write-Log "Image build completed"
    }
    catch {
        Handle-Error "Image build failed: $_"
    }
}

# Start services
function Start-Services {
    Write-Log "Starting Laojun system services..."
    
    try {
        # Create network if it doesn't exist
        $networkExists = docker network ls --filter name=laojun-network --format "{{.Name}}" | Where-Object { $_ -eq "laojun-network" }
        if (-not $networkExists) {
            Write-Log "Creating Docker network..."
            docker network create laojun-network
        }
        
        # Start services
        docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d
        
        Write-Log "Services started successfully"
        
        # Wait for services to start
        Write-Log "Waiting for services to start..."
        Start-Sleep -Seconds 30
        
        # Show service status
        Show-Status
        
        Write-Log "System deployment completed!"
        Write-Host @"

Laojun System Deployed Successfully!

Access URLs:
  Plugin Marketplace (Default): http://localhost
  Admin Backend:               http://localhost:8888
  
API URLs:
  Admin API:                   http://localhost:8080
  Config Center:               http://localhost:8081
  Marketplace API:             http://localhost:8082

Log file: $LOG_FILE
"@
    }
    catch {
        Handle-Error "Service startup failed: $_"
    }
}

# Stop services
function Stop-Services {
    Write-Log "Stopping Laojun system services..."
    
    try {
        docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE down
        Write-Log "Services stopped successfully"
    }
    catch {
        Handle-Error "Service stop failed: $_"
    }
}

# Restart services
function Restart-Services {
    Write-Log "Restarting Laojun system services..."
    Stop-Services
    Start-Services
}

# Show logs
function Show-Logs {
    Write-Log "Displaying service logs..."
    
    try {
        docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE logs -f --tail=100
    }
    catch {
        Handle-Error "Log viewing failed: $_"
    }
}

# Show status
function Show-Status {
    Write-Log "Checking service status..."
    
    try {
        Write-Host "`n=== Container Status ==="
        docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE ps
        
        Write-Host "`n=== Health Check ==="
        $containers = @("laojun-nginx", "laojun-admin-api", "laojun-marketplace-api", "laojun-config-center")
        foreach ($container in $containers) {
            $health = docker inspect --format='{{.State.Health.Status}}' $container 2>$null
            if ($health) {
                Write-Host "$container : $health"
            } else {
                Write-Host "$container : Not running"
            }
        }
    }
    catch {
        Handle-Error "Status check failed: $_"
    }
}

# Clean resources
function Clean-Resources {
    if (-not $Force) {
        $confirm = Read-Host "Are you sure you want to clean unused Docker resources? (y/N)"
        if ($confirm -ne "y" -and $confirm -ne "Y") {
            Write-Log "Clean operation cancelled"
            return
        }
    }
    
    Write-Log "Cleaning Docker resources..."
    
    try {
        # Stop services
        docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE down -v --remove-orphans
        
        # Clean unused images
        docker image prune -f
        
        # Clean unused containers
        docker container prune -f
        
        # Clean unused networks
        docker network prune -f
        
        Write-Log "Resource cleanup completed"
    }
    catch {
        Handle-Error "Resource cleanup failed: $_"
    }
}

# Main function
function Main {
    Write-Log "Starting deployment script, action: $Action"
    
    # Check Docker environment
    if (-not (Test-Docker)) {
        Handle-Error "Docker is not running or not installed"
    }
    
    if (-not (Test-DockerCompose)) {
        Handle-Error "docker-compose is not installed or not available"
    }
    
    # Check required files
    if (-not (Test-Path $COMPOSE_FILE)) {
        Handle-Error "Docker Compose file not found: $COMPOSE_FILE"
    }
    
    if (-not (Test-Path $ENV_FILE)) {
        Handle-Error "Environment file not found: $ENV_FILE"
    }
    
    # Execute corresponding action
    switch ($Action) {
        "start" { Start-Services }
        "stop" { Stop-Services }
        "restart" { Restart-Services }
        "build" { Build-Images }
        "logs" { Show-Logs }
        "status" { Show-Status }
        "clean" { Clean-Resources }
        "help" { Show-Help }
        default { Show-Help }
    }
}

# Execute main function
try {
    Main
}
catch {
    Handle-Error "Script execution failed: $_"
}