# Laojun Dockerfile Generator
param(
    [string]$Repository = "all",
    [switch]$Force,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun Dockerfile Generator ===" -ForegroundColor Green

# Workspace root directory
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# Repository configurations
$repositories = @(
    @{ Name = "laojun-config-center"; Type = "go-service" },
    @{ Name = "laojun-admin-api"; Type = "go-service" },
    @{ Name = "laojun-marketplace-api"; Type = "go-service" },
    @{ Name = "laojun-admin-web"; Type = "node-frontend" },
    @{ Name = "laojun-marketplace-web"; Type = "node-frontend" }
)

function New-GoServiceDockerfile {
    param($repoName)
    
    return @"
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files if they exist
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
"@
}

function New-NodeFrontendDockerfile {
    param($repoName)
    
    return @"
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of nginx directories
RUN chown -R appuser:appgroup /var/cache/nginx /var/run /var/log/nginx /usr/share/nginx/html

USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
"@
}

function New-NginxConfig {
    return @"
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    server {
        listen 8080;
        server_name localhost;
        root /usr/share/nginx/html;
        index index.html index.htm;

        # Security headers
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header Referrer-Policy "no-referrer-when-downgrade" always;
        add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

        # Handle client-side routing
        location / {
            try_files $uri $uri/ /index.html;
        }

        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
"@
}

function New-DockerIgnore {
    return @"
# Git
.git
.gitignore

# Documentation
README.md
*.md

# Development files
.env
.env.local
.env.development
.env.test
.env.production

# Dependencies
node_modules
vendor

# Build artifacts
dist
build
target
*.exe
*.dll
*.so
*.dylib

# Test files
coverage
*.test
test-results

# IDE files
.vscode
.idea
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Logs
*.log
logs

# Temporary files
tmp
temp
*.tmp
*.temp

# Docker
Dockerfile*
docker-compose*
.dockerignore
"@
}

function New-DockerConfig {
    param($repo)
    
    $repoPath = Join-Path $parentDir $repo.Name
    if (-not (Test-Path $repoPath)) {
        Write-Warning "Repository path not found: $($repo.Name)"
        return $false
    }
    
    $dockerFile = Join-Path $repoPath "Dockerfile"
    $dockerIgnoreFile = Join-Path $repoPath ".dockerignore"
    
    # Check if Dockerfile already exists
    if ((Test-Path $dockerFile) -and -not $Force) {
        Write-Host "$($repo.Name): Dockerfile already exists, use -Force to overwrite" -ForegroundColor Yellow
        return $true
    }
    
    try {
        # Generate Dockerfile content
        $dockerContent = switch ($repo.Type) {
            "go-service" { New-GoServiceDockerfile $repo.Name }
            "node-frontend" { New-NodeFrontendDockerfile $repo.Name }
            default { 
                Write-Warning "Unknown repository type: $($repo.Type)"
                return $false
            }
        }
        
        # Write Dockerfile
        Set-Content -Path $dockerFile -Value $dockerContent -Encoding UTF8
        
        # Write .dockerignore
        $dockerIgnoreContent = New-DockerIgnore
        Set-Content -Path $dockerIgnoreFile -Value $dockerIgnoreContent -Encoding UTF8
        
        # For frontend projects, also create nginx.conf
        if ($repo.Type -eq "node-frontend") {
            $nginxConfigFile = Join-Path $repoPath "nginx.conf"
            $nginxContent = New-NginxConfig
            Set-Content -Path $nginxConfigFile -Value $nginxContent -Encoding UTF8
        }
        
        Write-Host "$($repo.Name): Created Docker configuration ($($repo.Type))" -ForegroundColor Green
        if ($Verbose) {
            Write-Host "  Dockerfile: $dockerFile" -ForegroundColor Gray
            Write-Host "  .dockerignore: $dockerIgnoreFile" -ForegroundColor Gray
            if ($repo.Type -eq "node-frontend") {
                Write-Host "  nginx.conf: $(Join-Path $repoPath 'nginx.conf')" -ForegroundColor Gray
            }
        }
        
        return $true
        
    } catch {
        Write-Error "$($repo.Name): Failed to create Docker configuration: $_"
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

Write-Host "Generating Docker configurations..." -ForegroundColor Cyan
Write-Host ""

$successCount = 0
foreach ($repo in $targetRepos) {
    if (New-DockerConfig $repo) {
        $successCount++
    }
}

Write-Host ""
Write-Host "Docker configuration generation completed: $successCount/$($targetRepos.Count)" -ForegroundColor Green

if ($successCount -gt 0) {
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "1. Review generated Dockerfile and configurations" -ForegroundColor White
    Write-Host "2. Adjust configurations based on project requirements" -ForegroundColor White
    Write-Host "3. Test Docker builds locally" -ForegroundColor White
    Write-Host "4. Configure container registry credentials" -ForegroundColor White
}