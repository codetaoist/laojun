# Test All Local Services Script

$ErrorActionPreference = "Continue"

Write-Host "Testing All Local Services..." -ForegroundColor Cyan
Write-Host "==============================" -ForegroundColor Cyan

# Set environment variables
$env:ENV_FILE = ".env.local"
$env:CONFIG_CENTER_CONFIG = "configs/config-center.local.yaml"
$env:ADMIN_API_CONFIG = "configs/admin-api.local.yaml"
$env:MARKETPLACE_API_CONFIG = "configs/marketplace-api.local.yaml"
$env:DATABASE_CONFIG = "configs/database.local.yaml"

Write-Host "Environment variables set" -ForegroundColor Green

# Services configuration
$services = @(
    @{
        Name = "config-center"
        Executable = "bin/config-center.exe"
        Port = 8093
        HealthEndpoint = "http://localhost:8093/health"
    },
    @{
        Name = "admin-api"
        Executable = "bin/admin-api.exe"
        Port = 8080
        HealthEndpoint = "http://localhost:8080/health"
    },
    @{
        Name = "marketplace-api"
        Executable = "bin/marketplace-api.exe"
        Port = 8082
        HealthEndpoint = "http://localhost:8082/health"
    }
)

$processes = @()
$results = @()

# Start all services
Write-Host "`nStarting all services..." -ForegroundColor Yellow

foreach ($service in $services) {
    Write-Host "Starting $($service.Name)..." -ForegroundColor Cyan
    
    try {
        $process = Start-Process -FilePath $service.Executable -PassThru -WindowStyle Hidden
        
        if ($process) {
            $processes += @{
                Name = $service.Name
                Process = $process
                Port = $service.Port
                HealthEndpoint = $service.HealthEndpoint
            }
            Write-Host "  [OK] $($service.Name) started with PID: $($process.Id)" -ForegroundColor Green
        } else {
            Write-Host "  [FAIL] Failed to start $($service.Name)" -ForegroundColor Red
            $results += @{
                Service = $service.Name
                Status = "Failed to start"
                Port = $service.Port
            }
        }
    } catch {
        Write-Host "  [ERROR] Error starting $($service.Name): $($_.Exception.Message)" -ForegroundColor Red
        $results += @{
            Service = $service.Name
            Status = "Error: $($_.Exception.Message)"
            Port = $service.Port
        }
    }
}

# Wait for services to initialize
Write-Host "`nWaiting for services to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Test each service
Write-Host "`nTesting services..." -ForegroundColor Yellow

foreach ($proc in $processes) {
    Write-Host "Testing $($proc.Name)..." -ForegroundColor Cyan
    
    $status = "Unknown"
    
    # Check if process is still running
    if (!$proc.Process.HasExited) {
        Write-Host "  [OK] Process is running" -ForegroundColor Green
        
        # Test port connectivity
        try {
            $tcpClient = New-Object System.Net.Sockets.TcpClient
            $tcpClient.ConnectAsync("localhost", $proc.Port).Wait(2000)
            
            if ($tcpClient.Connected) {
                Write-Host "  [OK] Port $($proc.Port) is accessible" -ForegroundColor Green
                $tcpClient.Close()
                
                # Test health endpoint
                try {
                    $response = Invoke-WebRequest -Uri $proc.HealthEndpoint -TimeoutSec 5 -ErrorAction SilentlyContinue
                    if ($response.StatusCode -eq 200) {
                        Write-Host "  [OK] Health check passed" -ForegroundColor Green
                        $status = "Running and healthy"
                    } else {
                        Write-Host "  [WARN] Health check returned status: $($response.StatusCode)" -ForegroundColor Yellow
                        $status = "Running but health check failed"
                    }
                } catch {
                    Write-Host "  [WARN] Health endpoint not available" -ForegroundColor Yellow
                    $status = "Running but no health endpoint"
                }
            } else {
                Write-Host "  [FAIL] Port $($proc.Port) is not accessible" -ForegroundColor Red
                $status = "Running but port not accessible"
            }
        } catch {
            Write-Host "  [FAIL] Cannot connect to port $($proc.Port)" -ForegroundColor Red
            $status = "Running but port connection failed"
        }
    } else {
        Write-Host "  [FAIL] Process has exited" -ForegroundColor Red
        $status = "Process exited"
    }
    
    $results += @{
        Service = $proc.Name
        Status = $status
        Port = $proc.Port
        PID = $proc.Process.Id
    }
}

# Stop all services
Write-Host "`nStopping all services..." -ForegroundColor Yellow

foreach ($proc in $processes) {
    try {
        if (!$proc.Process.HasExited) {
            $proc.Process.Kill()
            Write-Host "  [OK] Stopped $($proc.Name)" -ForegroundColor Green
        }
    } catch {
        Write-Host "  [WARN] Error stopping $($proc.Name): $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

# Display results
Write-Host "`nTest Results Summary:" -ForegroundColor Cyan
Write-Host "=====================" -ForegroundColor Cyan

foreach ($result in $results) {
    $color = switch ($result.Status) {
        "Running and healthy" { "Green" }
        { $_ -like "Running*" } { "Yellow" }
        default { "Red" }
    }
    
    Write-Host "$($result.Service) (Port $($result.Port)): $($result.Status)" -ForegroundColor $color
}

$healthyServices = ($results | Where-Object { $_.Status -eq "Running and healthy" }).Count
$totalServices = $results.Count

Write-Host "`nOverall Result: $healthyServices/$totalServices services are fully functional" -ForegroundColor $(if ($healthyServices -eq $totalServices) { "Green" } else { "Yellow" })

Write-Host "`nAll services test completed!" -ForegroundColor Cyan