# Laojun 监控服务启动脚本

Write-Host "启动 Laojun 监控服务..." -ForegroundColor Green

# 检查 Docker 是否运行
try {
    docker info | Out-Null
} catch {
    Write-Host "错误: Docker 未运行，请先启动 Docker" -ForegroundColor Red
    exit 1
}

# 创建网络
docker network create laojun 2>$null

# 启动监控服务
docker-compose -f monitoring/docker-compose.yml up -d

Write-Host "监控服务启动完成！" -ForegroundColor Green
Write-Host "访问地址："
Write-Host "  Grafana:      http://localhost:3000 (admin/admin123)" -ForegroundColor Cyan
Write-Host "  Prometheus:   http://localhost:9090" -ForegroundColor Cyan
Write-Host "  Jaeger:       http://localhost:16686" -ForegroundColor Cyan
Write-Host "  AlertManager: http://localhost:9093" -ForegroundColor Cyan