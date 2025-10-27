# Laojun 安全扫描脚本
param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("all", "code", "dependencies", "secrets", "containers")]
    [string]$ScanType = "all",
    
    [Parameter(Mandatory=$false)]
    [switch]$FailOnIssues
)

Write-Host "=== Laojun 安全扫描 ===" -ForegroundColor Cyan
Write-Host ""

$ErrorCount = 0

# 代码安全扫描 (Gosec)
function Invoke-CodeScan {
    Write-Host "执行代码安全扫描 (Gosec)..." -ForegroundColor Green
    
    if (!(Get-Command gosec -ErrorAction SilentlyContinue)) {
        Write-Host "  警告: gosec 未安装，跳过代码安全扫描" -ForegroundColor Yellow
        Write-Host "  安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" -ForegroundColor Yellow
        return
    }
    
    try {
        gosec -conf security/configs/gosec/gosec.json ./...
        Write-Host "  代码安全扫描完成" -ForegroundColor Green
    } catch {
        Write-Host "  代码安全扫描发现问题" -ForegroundColor Red
        $script:ErrorCount++
    }
}

# 依赖漏洞扫描 (Trivy)
function Invoke-DependencyScan {
    Write-Host "执行依赖漏洞扫描 (Trivy)..." -ForegroundColor Green
    
    if (!(Get-Command trivy -ErrorAction SilentlyContinue)) {
        Write-Host "  警告: trivy 未安装，跳过依赖漏洞扫描" -ForegroundColor Yellow
        Write-Host "  安装说明: https://aquasecurity.github.io/trivy/latest/getting-started/installation/" -ForegroundColor Yellow
        return
    }
    
    try {
        trivy fs --config security/configs/trivy/trivy.yaml .
        Write-Host "  依赖漏洞扫描完成" -ForegroundColor Green
    } catch {
        Write-Host "  依赖漏洞扫描发现问题" -ForegroundColor Red
        $script:ErrorCount++
    }
}

# 密钥泄露扫描 (Trivy)
function Invoke-SecretScan {
    Write-Host "执行密钥泄露扫描..." -ForegroundColor Green
    
    if (!(Get-Command trivy -ErrorAction SilentlyContinue)) {
        Write-Host "  警告: trivy 未安装，跳过密钥泄露扫描" -ForegroundColor Yellow
        return
    }
    
    try {
        trivy fs --scanners secret --format json --output security/reports/secrets-report.json .
        Write-Host "  密钥泄露扫描完成" -ForegroundColor Green
    } catch {
        Write-Host "  密钥泄露扫描发现问题" -ForegroundColor Red
        $script:ErrorCount++
    }
}

# 容器镜像扫描 (Trivy)
function Invoke-ContainerScan {
    Write-Host "执行容器镜像扫描..." -ForegroundColor Green
    
    if (!(Get-Command trivy -ErrorAction SilentlyContinue)) {
        Write-Host "  警告: trivy 未安装，跳过容器镜像扫描" -ForegroundColor Yellow
        return
    }
    
    $dockerfiles = Get-ChildItem -Recurse -Name "Dockerfile*"
    if ($dockerfiles.Count -eq 0) {
        Write-Host "  未找到 Dockerfile，跳过容器镜像扫描" -ForegroundColor Yellow
        return
    }
    
    foreach ($dockerfile in $dockerfiles) {
        try {
            Write-Host "  扫描 $dockerfile..." -ForegroundColor Gray
            trivy config $dockerfile
        } catch {
            Write-Host "  容器配置扫描发现问题: $dockerfile" -ForegroundColor Red
            $script:ErrorCount++
        }
    }
}

# 执行扫描
switch ($ScanType) {
    "all" {
        Invoke-CodeScan
        Invoke-DependencyScan
        Invoke-SecretScan
        Invoke-ContainerScan
    }
    "code" {
        Invoke-CodeScan
    }
    "dependencies" {
        Invoke-DependencyScan
    }
    "secrets" {
        Invoke-SecretScan
    }
    "containers" {
        Invoke-ContainerScan
    }
}

# 生成报告摘要
Write-Host ""
Write-Host "安全扫描完成！" -ForegroundColor Green
Write-Host "报告位置: security/reports/" -ForegroundColor Cyan

if ($ErrorCount -gt 0) {
    Write-Host "发现 $ErrorCount 个安全问题" -ForegroundColor Red
    if ($FailOnIssues) {
        exit 1
    }
} else {
    Write-Host "未发现安全问题" -ForegroundColor Green
}