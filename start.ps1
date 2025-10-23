# Laojun System Quick Start Script
Write-Host "Starting Laojun System..." -ForegroundColor Green

# Check if Docker is running
try {
    docker version | Out-Null
}
catch {
    Write-Host "Docker is not running, please start Docker Desktop first" -ForegroundColor Red
    exit 1
}

# Set paths
$COMPOSE_FILE = ".\deploy\docker\docker-compose.yml"
$ENV_FILE = ".\deploy\configs\.env"

# Check if files exist
if (-not (Test-Path $COMPOSE_FILE)) {
    Write-Host "Docker Compose file not found: $COMPOSE_FILE" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $ENV_FILE)) {
    Write-Host "Environment file not found: $ENV_FILE" -ForegroundColor Red
    exit 1
}

Write-Host "Building and starting services..." -ForegroundColor Yellow

try {
    # Build and start services
    docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d --build
    
    Write-Host "Waiting for services to start..." -ForegroundColor Yellow
    Start-Sleep -Seconds 30
    
    # Check service status
    Write-Host "Checking service status..." -ForegroundColor Yellow
    docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE ps
    
    Write-Host ""
    Write-Host "Laojun System started successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Plugin Marketplace: http://localhost" -ForegroundColor Green
    Write-Host "Admin Backend: http://localhost:8888" -ForegroundColor Green
    Write-Host ""
}
catch {
    Write-Host "Startup failed: $_" -ForegroundColor Red
    exit 1
}