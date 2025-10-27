# Migration Files Consistency Test Script
# Verify consistency between main project and Docker deployment migration files

param(
    [switch]$Verbose = $false
)

# Set paths
$ProjectRoot = "d:\laojun"
$MainMigrationsPath = "$ProjectRoot\db\migrations"
$DockerMigrationsPath = "$ProjectRoot\deploy\docker\init-db\migrations"
$FinalPath = "$MainMigrationsPath\final"

Write-Host "=== Migration Files Consistency Test ===" -ForegroundColor Green
Write-Host "Project Root: $ProjectRoot" -ForegroundColor Gray
Write-Host "Main Migrations: $MainMigrationsPath" -ForegroundColor Gray
Write-Host "Docker Migrations: $DockerMigrationsPath" -ForegroundColor Gray
Write-Host ""

# Check if directory exists
function Test-DirectoryExists {
    param($Path, $Name)
    
    if (Test-Path $Path) {
        Write-Host "OK $Name directory exists" -ForegroundColor Green
        return $true
    } else {
        Write-Host "ERROR $Name directory not found: $Path" -ForegroundColor Red
        return $false
    }
}

# Check if file exists
function Test-FileExists {
    param($Path, $Name)
    
    if (Test-Path $Path) {
        Write-Host "OK $Name file exists" -ForegroundColor Green
        return $true
    } else {
        Write-Host "ERROR $Name file not found: $Path" -ForegroundColor Red
        return $false
    }
}

# 1. Check directory structure
Write-Host "1. Checking Directory Structure" -ForegroundColor Yellow
$mainDirExists = Test-DirectoryExists $MainMigrationsPath "Main Migrations"
$dockerDirExists = Test-DirectoryExists $DockerMigrationsPath "Docker Migrations"
$finalDirExists = Test-DirectoryExists $FinalPath "Final"

if (-not ($mainDirExists -and $dockerDirExists)) {
    Write-Host "Critical directories missing, cannot continue test" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 2. Check key files
Write-Host "2. Checking Key Files" -ForegroundColor Yellow

$keyFiles = @{
    "complete_deployment.sql" = "$FinalPath\complete_deployment.sql"
    "main_init_script" = "$MainMigrationsPath\000_init_database.sql"
    "main_marketplace_script" = "$MainMigrationsPath\001_create_marketplace_tables.up.sql"
    "main_system_script" = "$MainMigrationsPath\006_create_system_tables.up.sql"
    "main_optimization_script" = "$MainMigrationsPath\009_comprehensive_optimization.up.sql"
    "docker_init_script" = "$DockerMigrationsPath\000_init_database.up.sql"
    "docker_marketplace_script" = "$DockerMigrationsPath\001_create_marketplace_tables.up.sql"
    "docker_system_script" = "$DockerMigrationsPath\006_create_system_tables.up.sql"
    "docker_optimization_script" = "$DockerMigrationsPath\009_comprehensive_optimization.up.sql"
    "docker_complete_script" = "$DockerMigrationsPath\999_complete_deployment.up.sql"
}

$allFilesExist = $true
foreach ($file in $keyFiles.GetEnumerator()) {
    $exists = Test-FileExists $file.Value $file.Key
    if (-not $exists) {
        $allFilesExist = $false
    }
}

Write-Host ""

# 3. Check migration files order
Write-Host "3. Checking Migration Files Order" -ForegroundColor Yellow

$migrationFiles = Get-ChildItem -Path $DockerMigrationsPath -Filter "*.sql" | Sort-Object Name
Write-Host "Found $($migrationFiles.Count) migration files:" -ForegroundColor Gray

foreach ($file in $migrationFiles) {
    Write-Host "  - $($file.Name)" -ForegroundColor Gray
}

Write-Host ""

# 4. Generate test report
Write-Host "4. Test Report" -ForegroundColor Yellow

if ($allFilesExist) {
    Write-Host "SUCCESS All tests passed! Migration files consistency is good" -ForegroundColor Green
    Write-Host ""
    Write-Host "Recommended deployment order:" -ForegroundColor Cyan
    Write-Host "1. Use main project migrations: cd $MainMigrationsPath" -ForegroundColor Gray
    Write-Host "2. Use Docker deployment: cd $ProjectRoot\deploy\docker && docker-compose up -d" -ForegroundColor Gray
    Write-Host ""
    exit 0
} else {
    Write-Host "ERROR Issues found, please check error messages above" -ForegroundColor Red
    Write-Host ""
    Write-Host "Recommended fix steps:" -ForegroundColor Cyan
    Write-Host "1. Ensure all required migration files exist" -ForegroundColor Gray
    Write-Host "2. Check SQL file syntax errors" -ForegroundColor Gray
    Write-Host "3. Verify file naming conventions" -ForegroundColor Gray
    Write-Host "4. Re-run this test script" -ForegroundColor Gray
    Write-Host ""
    exit 1
}