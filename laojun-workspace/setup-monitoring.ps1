# Laojun 监控和运维工具配置脚本
param(
    [Parameter(Mandatory=$false)]
    [string]$Environment = "development"
)

# 配置变量
$MonitoringDir = "monitoring"
$ConfigDir = "$MonitoringDir/configs"

Write-Host "Laojun 监控和运维工具配置" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

# 创建监控目录结构
Write-Host "创建监控目录结构..." -ForegroundColor Green

$directories = @(
    $MonitoringDir,
    $ConfigDir,
    "$ConfigDir/prometheus",
    "$ConfigDir/grafana",
    "$ConfigDir/grafana/dashboards",
    "$ConfigDir/grafana/provisioning",
    "$ConfigDir/grafana/provisioning/dashboards",
    "$ConfigDir/grafana/provisioning/datasources",
    "$ConfigDir/loki",
    "$ConfigDir/alertmanager",
    "$MonitoringDir/data",
    "$MonitoringDir/data/prometheus",
    "$MonitoringDir/data/grafana",
    "$MonitoringDir/data/loki"
)

foreach ($dir in $directories) {
    if (!(Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "  创建目录: $dir" -ForegroundColor Gray
    }
}

Write-Host "生成配置文件..." -ForegroundColor Green

# 创建 Docker Compose 文件
Write-Host "  生成 Docker Compose 配置..." -ForegroundColor Gray
$dockerCompose = @"
version: '3.8'

networks:
  monitoring:
    driver: bridge
  laojun:
    external: true

volumes:
  prometheus_data:
  grafana_data:
  loki_data:

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: laojun-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./configs/prometheus/rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
    networks:
      - monitoring
      - laojun

  grafana:
    image: grafana/grafana:latest
    container_name: laojun-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./configs/grafana/provisioning:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
    networks:
      - monitoring

  loki:
    image: grafana/loki:latest
    container_name: laojun-loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    volumes:
      - ./configs/loki/loki.yml:/etc/loki/local-config.yaml
      - loki_data:/loki
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - monitoring

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: laojun-jaeger
    restart: unless-stopped
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - monitoring

  alertmanager:
    image: prom/alertmanager:latest
    container_name: laojun-alertmanager
    restart: unless-stopped
    ports:
      - "9093:9093"
    volumes:
      - ./configs/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
    networks:
      - monitoring
"@

Set-Content -Path "$MonitoringDir/docker-compose.yml" -Value $dockerCompose -Encoding UTF8

Write-Host "监控工具配置完成！" -ForegroundColor Green
Write-Host ""
Write-Host "下一步操作："
Write-Host "1. 手动创建配置文件（参考 monitoring/configs/ 目录）" -ForegroundColor Yellow
Write-Host "2. 启动监控服务: docker-compose -f monitoring/docker-compose.yml up -d" -ForegroundColor Yellow
Write-Host "3. 访问 Grafana: http://localhost:3000 (admin/admin123)" -ForegroundColor Yellow