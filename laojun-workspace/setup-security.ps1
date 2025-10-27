# Laojun 项目安全扫描和漏洞检测配置脚本
# 作者: Laojun Team
# 版本: 1.0.0

Write-Host "=== Laojun 安全扫描和漏洞检测配置 ===" -ForegroundColor Green
Write-Host ""

# 创建安全配置目录结构
Write-Host "创建安全配置目录结构..." -ForegroundColor Yellow

$SecurityDir = "security"
$ConfigDir = "$SecurityDir/configs"
$ScriptsDir = "$SecurityDir/scripts"
$ReportsDir = "$SecurityDir/reports"

# 创建目录
@($SecurityDir, $ConfigDir, $ScriptsDir, $ReportsDir, 
  "$ConfigDir/sonar", "$ConfigDir/trivy", "$ConfigDir/gosec") | ForEach-Object {
    if (!(Test-Path $_)) {
        New-Item -ItemType Directory -Path $_ -Force | Out-Null
        Write-Host "  创建目录: $_" -ForegroundColor Gray
    }
}

# 创建 SonarQube 配置
Write-Host "生成 SonarQube 配置..." -ForegroundColor Yellow

$sonarContent = @(
    "# SonarQube 项目配置",
    "sonar.projectKey=laojun",
    "sonar.projectName=Laojun",
    "sonar.projectVersion=1.0.0",
    "",
    "# 源码配置",
    "sonar.sources=.",
    "sonar.exclusions=**/*_test.go,**/vendor/**,**/node_modules/**,**/dist/**,**/build/**",
    "sonar.tests=.",
    "sonar.test.inclusions=**/*_test.go",
    "",
    "# Go 语言配置",
    "sonar.go.coverage.reportPaths=coverage.out",
    "sonar.go.tests.reportPaths=test-report.json",
    "",
    "# 质量门配置",
    "sonar.qualitygate.wait=true",
    "",
    "# 分析配置",
    "sonar.sourceEncoding=UTF-8",
    "sonar.scm.provider=git"
)

$sonarContent | Out-File -FilePath "$ConfigDir/sonar/sonar-project.properties" -Encoding UTF8

# 创建 Trivy 配置
Write-Host "生成 Trivy 配置..." -ForegroundColor Yellow

$trivyContent = @(
    "# Trivy 漏洞扫描配置",
    "format: json",
    "output: security/reports/trivy-report.json",
    "severity: HIGH,CRITICAL",
    "ignore-unfixed: false",
    "exit-code: 1",
    "",
    "# 扫描类型",
    "scanners:",
    "  - vuln",
    "  - secret",
    "  - config",
    "",
    "# 忽略规则",
    "ignorefile: .trivyignore",
    "",
    "# 缓存配置",
    "cache-dir: .trivy-cache"
)

$trivyContent | Out-File -FilePath "$ConfigDir/trivy/trivy.yaml" -Encoding UTF8

# 创建 .trivyignore 文件
$trivyIgnoreContent = @(
    "# Trivy 忽略规则",
    "# 忽略测试文件中的漏洞",
    "**/*_test.go",
    "",
    "# 忽略示例代码",
    "**/examples/**",
    "",
    "# 忽略第三方依赖（如果需要）",
    "# vendor/**"
)

$trivyIgnoreContent | Out-File -FilePath ".trivyignore" -Encoding UTF8

# 创建 Gosec 配置
Write-Host "生成 Gosec 配置..." -ForegroundColor Yellow

$gosecConfig = @{
    "global" = @{
        "nosec" = $false
        "fmt" = "json"
        "out" = "security/reports/gosec-report.json"
        "stdout" = $false
        "verbose" = "text"
        "severity" = "medium"
        "confidence" = "medium"
    }
    "rules" = @{
        "G101" = @{ "enabled" = $true }
        "G102" = @{ "enabled" = $true }
        "G103" = @{ "enabled" = $true }
        "G104" = @{ "enabled" = $true }
        "G106" = @{ "enabled" = $true }
        "G107" = @{ "enabled" = $true }
        "G108" = @{ "enabled" = $true }
        "G109" = @{ "enabled" = $true }
        "G110" = @{ "enabled" = $true }
        "G201" = @{ "enabled" = $true }
        "G202" = @{ "enabled" = $true }
        "G203" = @{ "enabled" = $true }
        "G204" = @{ "enabled" = $true }
        "G301" = @{ "enabled" = $true }
        "G302" = @{ "enabled" = $true }
        "G303" = @{ "enabled" = $true }
        "G304" = @{ "enabled" = $true }
        "G305" = @{ "enabled" = $true }
        "G306" = @{ "enabled" = $true }
        "G307" = @{ "enabled" = $true }
        "G401" = @{ "enabled" = $true }
        "G402" = @{ "enabled" = $true }
        "G403" = @{ "enabled" = $true }
        "G404" = @{ "enabled" = $true }
        "G501" = @{ "enabled" = $true }
        "G502" = @{ "enabled" = $true }
        "G503" = @{ "enabled" = $true }
        "G504" = @{ "enabled" = $true }
        "G505" = @{ "enabled" = $true }
        "G601" = @{ "enabled" = $true }
    }
}

$gosecConfig | ConvertTo-Json -Depth 3 | Out-File -FilePath "$ConfigDir/gosec/gosec.json" -Encoding UTF8

Write-Host ""
Write-Host "=== 安全配置完成 ===" -ForegroundColor Green
Write-Host ""
Write-Host "已创建以下配置文件:" -ForegroundColor Cyan
Write-Host "  - $ConfigDir/sonar/sonar-project.properties" -ForegroundColor Gray
Write-Host "  - $ConfigDir/trivy/trivy.yaml" -ForegroundColor Gray
Write-Host "  - .trivyignore" -ForegroundColor Gray
Write-Host "  - $ConfigDir/gosec/gosec.json" -ForegroundColor Gray
Write-Host ""
Write-Host "下一步:" -ForegroundColor Yellow
Write-Host "  1. 安装安全扫描工具 (gosec, trivy)" -ForegroundColor Gray
Write-Host "  2. 配置 SonarQube 服务器" -ForegroundColor Gray
Write-Host "  3. 运行安全扫描脚本" -ForegroundColor Gray
Write-Host ""