# Migration Files Maintenance Script
# Comprehensive tool for maintaining migration file quality and consistency

param(
    [switch]$CheckDuplicates = $false,
    [switch]$ValidateSyntax = $false,
    [switch]$CheckNaming = $false,
    [switch]$GenerateReport = $false,
    [switch]$All = $false,
    [switch]$Verbose = $false
)

$ProjectRoot = "d:\laojun"
$MigrationsPath = "$ProjectRoot\db\migrations"
$DockerMigrationsPath = "$ProjectRoot\deploy\docker\init-db\migrations"

if ($All) {
    $CheckDuplicates = $true
    $ValidateSyntax = $true
    $CheckNaming = $true
    $GenerateReport = $true
}

Write-Host "=== Migration Files Maintenance ===" -ForegroundColor Green
Write-Host "Project Root: $ProjectRoot" -ForegroundColor Gray
Write-Host "Migrations Path: $MigrationsPath" -ForegroundColor Gray
Write-Host ""

# Function to check file naming convention
function Test-NamingConvention {
    param($Directory)
    
    Write-Host "Checking naming convention in: $Directory" -ForegroundColor Yellow
    
    $files = Get-ChildItem -Path $Directory -Filter "*.sql" | Where-Object { $_.Name -notmatch "final|README" }
    $issues = @()
    
    foreach ($file in $files) {
        $name = $file.Name
        
        # Check if follows pattern: NNN_description.up.sql or NNN_description.down.sql
        if ($name -notmatch '^\d{3}_[a-z_]+\.(up|down)\.sql$' -and $name -ne "000_init_database.sql") {
            $issues += "❌ $name - Invalid naming pattern"
        } else {
            if ($Verbose) {
                Write-Host "  ✅ $name" -ForegroundColor Green
            }
        }
    }
    
    if ($issues.Count -eq 0) {
        Write-Host "✅ All files follow naming convention" -ForegroundColor Green
    } else {
        Write-Host "❌ Found $($issues.Count) naming issues:" -ForegroundColor Red
        $issues | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
    }
    
    return $issues
}

# Function to validate SQL syntax
function Test-SqlSyntax {
    param($Directory)
    
    Write-Host "Validating SQL syntax in: $Directory" -ForegroundColor Yellow
    
    $files = Get-ChildItem -Path $Directory -Filter "*.sql"
    $issues = @()
    
    foreach ($file in $files) {
        $content = Get-Content $file.FullName -Raw
        
        # Basic syntax checks
        $openParens = ($content -split '\(' | Measure-Object).Count - 1
        $closeParens = ($content -split '\)' | Measure-Object).Count - 1
        
        if ($openParens -ne $closeParens) {
            $issues += "❌ $($file.Name) - Mismatched parentheses"
        }
        
        # Check for common SQL errors
        if ($content -match '(?i)CREATE\s+TABLE\s+\w+\s*\(\s*\)') {
            $issues += "❌ $($file.Name) - Empty table definition"
        }
        
        if ($Verbose -and $issues.Count -eq 0) {
            Write-Host "  ✅ $($file.Name)" -ForegroundColor Green
        }
    }
    
    if ($issues.Count -eq 0) {
        Write-Host "✅ All SQL files have valid syntax" -ForegroundColor Green
    } else {
        Write-Host "❌ Found $($issues.Count) syntax issues:" -ForegroundColor Red
        $issues | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
    }
    
    return $issues
}

# Function to check for duplicate indexes
function Test-DuplicateIndexes {
    param($Directory)
    
    Write-Host "Checking for duplicate indexes in: $Directory" -ForegroundColor Yellow
    
    $files = Get-ChildItem -Path $Directory -Filter "*.sql"
    $allIndexes = @()
    
    foreach ($file in $files) {
        $content = Get-Content $file.FullName -Raw
        $matches = [regex]::Matches($content, 'CREATE\s+(?:UNIQUE\s+)?INDEX\s+IF\s+NOT\s+EXISTS\s+(\w+)', [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)
        
        foreach ($match in $matches) {
            $allIndexes += @{
                Name = $match.Groups[1].Value
                File = $file.Name
            }
        }
    }
    
    $duplicates = $allIndexes | Group-Object Name | Where-Object { $_.Count -gt 1 }
    
    if ($duplicates.Count -eq 0) {
        Write-Host "✅ No duplicate indexes found" -ForegroundColor Green
    } else {
        Write-Host "❌ Found $($duplicates.Count) duplicate indexes:" -ForegroundColor Red
        foreach ($dup in $duplicates) {
            Write-Host "  Index: $($dup.Name)" -ForegroundColor Yellow
            $dup.Group | ForEach-Object { Write-Host "    - $($_.File)" -ForegroundColor Gray }
        }
    }
    
    return $duplicates
}

# Function to check migration pairs
function Test-MigrationPairs {
    param($Directory)
    
    Write-Host "Checking migration up/down pairs in: $Directory" -ForegroundColor Yellow
    
    $upFiles = Get-ChildItem -Path $Directory -Filter "*up.sql" | ForEach-Object { $_.BaseName -replace '\.up$', '' }
    $downFiles = Get-ChildItem -Path $Directory -Filter "*down.sql" | ForEach-Object { $_.BaseName -replace '\.down$', '' }
    
    $missingDown = $upFiles | Where-Object { $_ -notin $downFiles }
    $missingUp = $downFiles | Where-Object { $_ -notin $upFiles }
    
    $issues = @()
    
    if ($missingDown.Count -gt 0) {
        $missingDown | ForEach-Object { $issues += "❌ Missing down file for: $_.up.sql" }
    }
    
    if ($missingUp.Count -gt 0) {
        $missingUp | ForEach-Object { $issues += "❌ Missing up file for: $_.down.sql" }
    }
    
    if ($issues.Count -eq 0) {
        Write-Host "✅ All migration files have matching pairs" -ForegroundColor Green
    } else {
        Write-Host "❌ Found $($issues.Count) pairing issues:" -ForegroundColor Red
        $issues | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
    }
    
    return $issues
}

# Main execution
$allIssues = @()

if ($CheckNaming) {
    Write-Host "1. Checking Naming Convention" -ForegroundColor Cyan
    $namingIssues = Test-NamingConvention $MigrationsPath
    $allIssues += $namingIssues
    Write-Host ""
}

if ($ValidateSyntax) {
    Write-Host "2. Validating SQL Syntax" -ForegroundColor Cyan
    $syntaxIssues = Test-SqlSyntax $MigrationsPath
    $allIssues += $syntaxIssues
    Write-Host ""
}

if ($CheckDuplicates) {
    Write-Host "3. Checking for Duplicate Indexes" -ForegroundColor Cyan
    $duplicateIssues = Test-DuplicateIndexes $MigrationsPath
    $allIssues += $duplicateIssues
    Write-Host ""
    
    Write-Host "4. Checking Migration Pairs" -ForegroundColor Cyan
    $pairIssues = Test-MigrationPairs $MigrationsPath
    $allIssues += $pairIssues
    Write-Host ""
}

# Generate summary report
if ($GenerateReport) {
    Write-Host "=== MAINTENANCE SUMMARY ===" -ForegroundColor Green
    
    $totalFiles = (Get-ChildItem -Path $MigrationsPath -Filter "*.sql" | Measure-Object).Count
    Write-Host "Total migration files: $totalFiles" -ForegroundColor Gray
    
    if ($allIssues.Count -eq 0) {
        Write-Host "🎉 All checks passed! Migration files are in excellent condition." -ForegroundColor Green
    } else {
        Write-Host "⚠️  Found $($allIssues.Count) issues that need attention:" -ForegroundColor Yellow
        $allIssues | ForEach-Object { Write-Host "  $_" -ForegroundColor Yellow }
    }
    
    Write-Host ""
    Write-Host "RECOMMENDATIONS:" -ForegroundColor Cyan
    Write-Host "1. Run this script regularly to maintain migration quality" -ForegroundColor Gray
    Write-Host "2. Always create both .up.sql and .down.sql files" -ForegroundColor Gray
    Write-Host "3. Follow the naming convention: NNN_description.(up|down).sql" -ForegroundColor Gray
    Write-Host "4. Avoid duplicate index creation across migration files" -ForegroundColor Gray
    Write-Host "5. Test migrations in a development environment before deployment" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Maintenance completed!" -ForegroundColor Green