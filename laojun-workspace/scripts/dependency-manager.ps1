# Laojun Dependency Manager
param(
    [ValidateSet("sync", "update", "check", "graph", "clean")]
    [string]$Action = "check",
    [string]$Module = "",
    [switch]$DryRun,
    [switch]$Verbose,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun Dependency Manager ===" -ForegroundColor Green

# Workspace root directory
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# Module configurations with dependency relationships
$modules = @{
    "laojun-shared" = @{
        Path = "laojun-shared"
        Type = "library"
        Dependencies = @()
        Dependents = @("laojun-plugins", "laojun-config-center", "laojun-admin-api", "laojun-marketplace-api")
    }
    "laojun-plugins" = @{
        Path = "laojun-plugins"
        Type = "library"
        Dependencies = @("laojun-shared")
        Dependents = @("laojun-config-center", "laojun-admin-api", "laojun-marketplace-api")
    }
    "laojun-config-center" = @{
        Path = "laojun-config-center"
        Type = "service"
        Dependencies = @("laojun-shared", "laojun-plugins")
        Dependents = @()
    }
    "laojun-admin-api" = @{
        Path = "laojun-admin-api"
        Type = "service"
        Dependencies = @("laojun-shared", "laojun-plugins")
        Dependents = @()
    }
    "laojun-marketplace-api" = @{
        Path = "laojun-marketplace-api"
        Type = "service"
        Dependencies = @("laojun-shared", "laojun-plugins")
        Dependents = @()
    }
}

function Get-ModuleVersion {
    param($modulePath)
    
    $goModPath = Join-Path $modulePath "go.mod"
    if (-not (Test-Path $goModPath)) {
        return $null
    }
    
    $content = Get-Content $goModPath -Raw
    if ($content -match "module\s+([^\s\r\n]+)") {
        return $matches[1]
    }
    
    return $null
}

function Get-ModuleDependencies {
    param($modulePath)
    
    $goModPath = Join-Path $modulePath "go.mod"
    if (-not (Test-Path $goModPath)) {
        return @()
    }
    
    $dependencies = @()
    $content = Get-Content $goModPath
    $inRequireBlock = $false
    
    foreach ($line in $content) {
        $line = $line.Trim()
        
        if ($line -eq "require (") {
            $inRequireBlock = $true
            continue
        }
        
        if ($inRequireBlock -and $line -eq ")") {
            $inRequireBlock = $false
            continue
        }
        
        if ($inRequireBlock -or ($line -match "^require\s+")) {
            if ($line -match "github\.com/[^/]+/laojun-[^\s]+") {
                $dep = $matches[0]
                if ($line -match "$dep\s+v?([^\s]+)") {
                    $dependencies += @{
                        Module = $dep
                        Version = $matches[1]
                        Line = $line
                    }
                }
            }
        }
    }
    
    return $dependencies
}

function Sync-WorkspaceDependencies {
    Write-Host "Syncing workspace dependencies..." -ForegroundColor Cyan
    
    Set-Location $workspaceRoot
    
    # Update go.work file
    $workspaceModules = @()
    foreach ($moduleName in $modules.Keys) {
        $modulePath = Join-Path $parentDir $modules[$moduleName].Path
        if (Test-Path $modulePath) {
            $relativePath = "..\$($modules[$moduleName].Path)"
            $workspaceModules += $relativePath
        }
    }
    
    $moduleLines = $workspaceModules | ForEach-Object { "    $_" }
    $workContent = @"
go 1.21

use (
$($moduleLines -join "`n")
)
"@
    
    if (-not $DryRun) {
        Set-Content -Path "go.work" -Value $workContent -Encoding ASCII
        Write-Host "✓ Updated go.work file" -ForegroundColor Green
    } else {
        Write-Host "Would update go.work file" -ForegroundColor Yellow
    }
    
    # Sync workspace
    if (-not $DryRun) {
        try {
            go work sync
            Write-Host "✓ Workspace synchronized" -ForegroundColor Green
        } catch {
            Write-Warning "Failed to sync workspace: $_"
        }
    }
}

function Update-ModuleDependencies {
    param($targetModule)
    
    $modulesToUpdate = if ($targetModule) {
        @($targetModule)
    } else {
        $modules.Keys
    }
    
    foreach ($moduleName in $modulesToUpdate) {
        $moduleConfig = $modules[$moduleName]
        $modulePath = Join-Path $parentDir $moduleConfig.Path
        
        if (-not (Test-Path $modulePath)) {
            Write-Warning "Module path not found: $moduleName"
            continue
        }
        
        Write-Host "Updating dependencies for $moduleName..." -ForegroundColor Cyan
        
        Set-Location $modulePath
        
        # Update internal dependencies
        foreach ($depName in $moduleConfig.Dependencies) {
            $depConfig = $modules[$depName]
            $depPath = Join-Path $parentDir $depConfig.Path
            
            if (Test-Path $depPath) {
                $depModule = Get-ModuleVersion $depPath
                if ($depModule) {
                    Write-Host "  Updating $depName..." -ForegroundColor Gray
                    
                    if (-not $DryRun) {
                        try {
                            go get $depModule@latest
                            Write-Host "  ✓ Updated $depName" -ForegroundColor Green
                        } catch {
                            Write-Warning "  Failed to update ${depName}: ${_}"
                        }
                    } else {
                        Write-Host "  Would update $depName" -ForegroundColor Yellow
                    }
                }
            }
        }
        
        # Update external dependencies
        if (-not $DryRun) {
            try {
                go get -u ./...
                go mod tidy
                Write-Host "✓ Updated external dependencies for $moduleName" -ForegroundColor Green
            } catch {
                Write-Warning "Failed to update external dependencies for ${moduleName}: ${_}"
            }
        }
    }
}

function Check-DependencyConsistency {
    Write-Host "Checking dependency consistency..." -ForegroundColor Cyan
    
    $issues = @()
    
    foreach ($moduleName in $modules.Keys) {
        $moduleConfig = $modules[$moduleName]
        $modulePath = Join-Path $parentDir $moduleConfig.Path
        
        if (-not (Test-Path $modulePath)) {
            $issues += "Module path not found: $moduleName"
            continue
        }
        
        $dependencies = Get-ModuleDependencies $modulePath
        
        foreach ($dep in $dependencies) {
            $depModuleName = $dep.Module -replace "github\.com/[^/]+/", ""
            
            if ($modules.ContainsKey($depModuleName)) {
                $expectedDep = $moduleConfig.Dependencies -contains $depModuleName
                if (-not $expectedDep) {
                    $issues += "$moduleName has unexpected dependency on $depModuleName"
                }
                
                # Check version consistency
                $depPath = Join-Path $parentDir $modules[$depModuleName].Path
                $actualModule = Get-ModuleVersion $depPath
                if ($actualModule -and $dep.Module -ne $actualModule) {
                    $issues += "$moduleName references incorrect module path: $($dep.Module) (expected: $actualModule)"
                }
            }
        }
        
        # Check for missing dependencies
        foreach ($expectedDep in $moduleConfig.Dependencies) {
            $found = $dependencies | Where-Object { $_.Module -match $expectedDep }
            if (-not $found) {
                $issues += "$moduleName missing dependency on $expectedDep"
            }
        }
    }
    
    if ($issues.Count -eq 0) {
        Write-Host "✓ All dependencies are consistent" -ForegroundColor Green
    } else {
        Write-Host "Found $($issues.Count) dependency issues:" -ForegroundColor Red
        foreach ($issue in $issues) {
            Write-Host "  - $issue" -ForegroundColor Red
        }
    }
    
    return $issues.Count -eq 0
}

function Show-DependencyGraph {
    Write-Host "Dependency Graph:" -ForegroundColor Cyan
    Write-Host ""
    
    # Show dependency tree
    function Show-ModuleDeps {
        param($moduleName, $indent = 0)
        
        $prefix = "  " * $indent
        $moduleConfig = $modules[$moduleName]
        $typeColor = if ($moduleConfig.Type -eq "library") { "Blue" } else { "Magenta" }
        
        Write-Host "$prefix$moduleName ($($moduleConfig.Type))" -ForegroundColor $typeColor
        
        foreach ($dep in $moduleConfig.Dependencies) {
            Show-ModuleDeps $dep ($indent + 1)
        }
    }
    
    # Show root modules (services)
    $services = $modules.Keys | Where-Object { $modules[$_].Type -eq "service" }
    foreach ($service in $services) {
        Show-ModuleDeps $service
        Write-Host ""
    }
    
    # Show dependency statistics
    Write-Host "Statistics:" -ForegroundColor Yellow
    $libraries = $modules.Keys | Where-Object { $modules[$_].Type -eq "library" }
    $services = $modules.Keys | Where-Object { $modules[$_].Type -eq "service" }
    
    Write-Host "  Libraries: $($libraries.Count)" -ForegroundColor Blue
    Write-Host "  Services: $($services.Count)" -ForegroundColor Magenta
    
    foreach ($lib in $libraries) {
        $dependents = $modules[$lib].Dependents
        Write-Host "  $lib used by $($dependents.Count) modules" -ForegroundColor Gray
    }
}

function Clean-DependencyCache {
    Write-Host "Cleaning dependency cache..." -ForegroundColor Cyan
    
    foreach ($moduleName in $modules.Keys) {
        $modulePath = Join-Path $parentDir $modules[$moduleName].Path
        
        if (Test-Path $modulePath) {
            Set-Location $modulePath
            
            Write-Host "Cleaning $moduleName..." -ForegroundColor Gray
            
            if (-not $DryRun) {
                try {
                    go clean -modcache
                    go mod download
                    Write-Host "✓ Cleaned $moduleName" -ForegroundColor Green
                } catch {
                    Write-Warning "Failed to clean ${moduleName}: ${_}"
                }
            } else {
                Write-Host "Would clean $moduleName" -ForegroundColor Yellow
            }
        }
    }
}

# Main logic
Set-Location $workspaceRoot

if ($Verbose) {
    Write-Host "Action: $Action" -ForegroundColor Gray
    Write-Host "Module: $(if ($Module) { $Module } else { 'all' })" -ForegroundColor Gray
    Write-Host "Dry Run: $DryRun" -ForegroundColor Gray
    Write-Host ""
}

switch ($Action) {
    "sync" {
        Sync-WorkspaceDependencies
    }
    "update" {
        if ($Module -and -not $modules.ContainsKey($Module)) {
            Write-Error "Unknown module: $Module"
            exit 1
        }
        Update-ModuleDependencies $Module
    }
    "check" {
        $consistent = Check-DependencyConsistency
        if (-not $consistent) {
            exit 1
        }
    }
    "graph" {
        Show-DependencyGraph
    }
    "clean" {
        Clean-DependencyCache
    }
}

Write-Host ""
Write-Host "Dependency management completed successfully" -ForegroundColor Green