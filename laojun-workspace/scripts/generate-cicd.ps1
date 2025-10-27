# Laojun CI/CD Configuration Generator
param(
    [string]$Repository = "all",
    [switch]$Force,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun CI/CD Config Generator ===" -ForegroundColor Green

# Workspace root directory
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# Repository configurations
$repositories = @(
    @{ Name = "laojun-shared"; Type = "go-library" },
    @{ Name = "laojun-plugins"; Type = "go-library" },
    @{ Name = "laojun-config-center"; Type = "go-service" },
    @{ Name = "laojun-admin-api"; Type = "go-service" },
    @{ Name = "laojun-marketplace-api"; Type = "go-service" },
    @{ Name = "laojun-admin-web"; Type = "node-frontend" },
    @{ Name = "laojun-marketplace-web"; Type = "node-frontend" }
)

function New-GoLibraryCI {
    return @"
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: '1.21'

jobs:
  lint:
    name: Code Quality
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: `${{ env.GO_VERSION }}
    - run: go mod download
    - run: go vet ./...

  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: `${{ env.GO_VERSION }}
    - run: go mod download
    - run: go test -v -race -coverprofile=coverage.out ./...

  build:
    name: Build Verification
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: `${{ env.GO_VERSION }}
    - run: go mod download
    - run: go build ./...
"@
}

function New-GoServiceCI {
    return @"
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: '1.21'
  REGISTRY: ghcr.io
  IMAGE_NAME: `${{ github.repository }}

jobs:
  lint:
    name: Code Quality
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: `${{ env.GO_VERSION }}
    - run: go mod download
    - run: go vet ./...

  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: `${{ env.GO_VERSION }}
    - run: go mod download
    - run: go test -v -race -coverprofile=coverage.out ./...
      env:
        DATABASE_URL: "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
        REDIS_URL: "redis://localhost:6379"

  build-and-push:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: [lint, test]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
    - uses: actions/checkout@v4
    - uses: docker/setup-buildx-action@v3
    - uses: docker/login-action@v3
      with:
        registry: `${{ env.REGISTRY }}
        username: `${{ github.actor }}
        password: `${{ secrets.GITHUB_TOKEN }}
    - uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: `${{ env.REGISTRY }}/`${{ env.IMAGE_NAME }}:latest
"@
}

function New-NodeFrontendCI {
    return @"
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  NODE_VERSION: '18'

jobs:
  lint:
    name: Code Quality
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: `${{ env.NODE_VERSION }}
        cache: 'npm'
    - run: npm ci
    - run: npm run lint
    - run: npm run type-check

  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: `${{ env.NODE_VERSION }}
        cache: 'npm'
    - run: npm ci
    - run: npm run test:unit -- --coverage

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: `${{ env.NODE_VERSION }}
        cache: 'npm'
    - run: npm ci
    - run: npm run build
    - uses: actions/upload-artifact@v3
      with:
        name: build-artifacts
        path: dist/
"@
}

function New-CICDConfig {
    param($repo)
    
    $repoPath = Join-Path $parentDir $repo.Name
    if (-not (Test-Path $repoPath)) {
        Write-Warning "Repository path not found: $($repo.Name)"
        return $false
    }
    
    $workflowDir = Join-Path $repoPath ".github\workflows"
    $configFile = Join-Path $workflowDir "ci.yml"
    
    # Check if config already exists
    if ((Test-Path $configFile) -and -not $Force) {
        Write-Host "$($repo.Name): CI/CD config already exists, use -Force to overwrite" -ForegroundColor Yellow
        return $true
    }
    
    # Get template content
    $content = switch ($repo.Type) {
        "go-library" { New-GoLibraryCI }
        "go-service" { New-GoServiceCI }
        "node-frontend" { New-NodeFrontendCI }
        default { 
            Write-Warning "Unknown repository type: $($repo.Type)"
            return $false
        }
    }
    
    try {
        # Create directory
        if (-not (Test-Path $workflowDir)) {
            New-Item -Path $workflowDir -ItemType Directory -Force | Out-Null
        }
        
        # Write config file
        Set-Content -Path $configFile -Value $content -Encoding UTF8
        
        Write-Host "$($repo.Name): Created CI/CD config ($($repo.Type))" -ForegroundColor Green
        if ($Verbose) {
            Write-Host "  Path: $configFile" -ForegroundColor Gray
        }
        
        return $true
        
    } catch {
        Write-Error "$($repo.Name): Failed to create CI/CD config: $_"
        return $false
    }
}

# Main logic
Set-Location $workspaceRoot

# Filter repositories
$targetRepos = if ($Repository -eq "all") {
    $repositories
} else {
    $repositories | Where-Object { $_.Name -eq $Repository }
}

if (-not $targetRepos) {
    Write-Error "No matching repository found: $Repository"
    exit 1
}

Write-Host "Generating CI/CD configurations..." -ForegroundColor Cyan
Write-Host ""

$successCount = 0
foreach ($repo in $targetRepos) {
    if (New-CICDConfig $repo) {
        $successCount++
    }
}

Write-Host ""
Write-Host "CI/CD configuration generation completed: $successCount/$($targetRepos.Count)" -ForegroundColor Green

if ($successCount -gt 0) {
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "1. Review generated configuration files" -ForegroundColor White
    Write-Host "2. Adjust configurations based on project requirements" -ForegroundColor White
    Write-Host "3. Configure necessary GitHub Secrets" -ForegroundColor White
    Write-Host "4. Commit and push to GitHub" -ForegroundColor White
}