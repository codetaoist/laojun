# Index Optimization Script
# Identify and resolve duplicate index creation statements across migration files

param(
    [switch]$DryRun = $true,
    [switch]$Verbose = $false
)

$ProjectRoot = "d:\laojun"
$MigrationsPath = "$ProjectRoot\db\migrations"

Write-Host "=== Index Optimization Analysis ===" -ForegroundColor Green
Write-Host "Migrations Path: $MigrationsPath" -ForegroundColor Gray
Write-Host "Dry Run Mode: $DryRun" -ForegroundColor Gray
Write-Host ""

# Function to extract index names from CREATE INDEX statements
function Get-IndexesFromFile {
    param($FilePath)
    
    $content = Get-Content $FilePath -Raw
    $indexes = @()
    
    # Match CREATE INDEX statements
    $matches = [regex]::Matches($content, 'CREATE\s+(?:UNIQUE\s+)?INDEX\s+IF\s+NOT\s+EXISTS\s+(\w+)', [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)
    
    foreach ($match in $matches) {
        $indexes += @{
            Name = $match.Groups[1].Value
            File = Split-Path $FilePath -Leaf
            FullStatement = $match.Value
        }
    }
    
    return $indexes
}

# Get all migration files
$migrationFiles = Get-ChildItem -Path $MigrationsPath -Filter "*.sql" | Where-Object { $_.Name -notmatch "exported_" }

Write-Host "1. Analyzing Migration Files" -ForegroundColor Yellow

$allIndexes = @()
foreach ($file in $migrationFiles) {
    $indexes = Get-IndexesFromFile $file.FullName
    $allIndexes += $indexes
    
    if ($Verbose -and $indexes.Count -gt 0) {
        Write-Host "  $($file.Name): $($indexes.Count) indexes" -ForegroundColor Gray
    }
}

Write-Host "Total indexes found: $($allIndexes.Count)" -ForegroundColor Cyan
Write-Host ""

# Group by index name to find duplicates
Write-Host "2. Identifying Duplicate Indexes" -ForegroundColor Yellow

$groupedIndexes = $allIndexes | Group-Object Name
$duplicates = $groupedIndexes | Where-Object { $_.Count -gt 1 }

if ($duplicates.Count -eq 0) {
    Write-Host "No duplicate indexes found!" -ForegroundColor Green
    exit 0
}

Write-Host "Found $($duplicates.Count) duplicate index names:" -ForegroundColor Red
Write-Host ""

foreach ($duplicate in $duplicates) {
    Write-Host "Index: $($duplicate.Name)" -ForegroundColor Yellow
    foreach ($index in $duplicate.Group) {
        Write-Host "  - File: $($index.File)" -ForegroundColor Gray
    }
    Write-Host ""
}

# Recommendations
Write-Host "3. Optimization Recommendations" -ForegroundColor Yellow
Write-Host ""

Write-Host "RECOMMENDED ACTIONS:" -ForegroundColor Cyan
Write-Host "1. Keep indexes in the earliest migration file where they logically belong" -ForegroundColor Gray
Write-Host "2. Remove duplicate index creation from later migration files" -ForegroundColor Gray
Write-Host "3. Consider consolidating all indexes into 007_create_indexes.up.sql" -ForegroundColor Gray
Write-Host "4. Update corresponding .down.sql files" -ForegroundColor Gray
Write-Host ""

Write-Host "SPECIFIC DUPLICATES TO RESOLVE:" -ForegroundColor Cyan

# Show specific recommendations
$recommendations = @{
    "mp_users indexes" = @("001_create_marketplace_tables.up.sql", "007_create_indexes.up.sql", "009_comprehensive_optimization.up.sql")
    "mp_plugins indexes" = @("003_create_plugin_tables.up.sql", "005_add_is_featured_column.up.sql", "007_create_indexes.up.sql", "009_comprehensive_optimization.up.sql")
    "mp_categories indexes" = @("003_create_plugin_tables.up.sql", "007_create_indexes.up.sql")
    "mp_forum indexes" = @("001_create_marketplace_tables.up.sql", "007_create_indexes.up.sql", "009_comprehensive_optimization.up.sql")
}

foreach ($rec in $recommendations.GetEnumerator()) {
    Write-Host "$($rec.Key):" -ForegroundColor Yellow
    foreach ($file in $rec.Value) {
        Write-Host "  - $file" -ForegroundColor Gray
    }
    Write-Host ""
}

Write-Host "NEXT STEPS:" -ForegroundColor Cyan
Write-Host "1. Run with -DryRun:`$false to apply automatic fixes" -ForegroundColor Gray
Write-Host "2. Manually review and consolidate complex indexes" -ForegroundColor Gray
Write-Host "3. Test migration rollback functionality" -ForegroundColor Gray
Write-Host "4. Update Docker migration files accordingly" -ForegroundColor Gray

if (-not $DryRun) {
    Write-Host ""
    Write-Host "APPLYING FIXES..." -ForegroundColor Red
    Write-Host "This feature is not implemented yet for safety reasons." -ForegroundColor Yellow
    Write-Host "Please manually review and fix the duplicate indexes." -ForegroundColor Yellow
}