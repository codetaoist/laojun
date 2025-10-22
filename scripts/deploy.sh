#!/bin/bash

# 太上老君系统部署脚本
# Deploy script for Taishang Laojun System

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查必要的工具
check_requirements() {
    log_info "检查部署环境..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        log_error "Git 未安装，请先安装 Git"
        exit 1
    fi
    
    log_success "环境检查通过"
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    mkdir -p logs
    mkdir -p uploads/plugins
    mkdir -p deployments/ssl
    mkdir -p data/postgres
    mkdir -p data/redis
    
    log_success "目录创建完成"
}

# 检查环境配置文件
check_env_file() {
    log_info "检查环境配置文件..."
    
    if [ ! -f ".env.production" ]; then
        log_warning ".env.production 文件不存在，从示例文件复制..."
        cp .env.production.example .env.production
        log_warning "请编辑 .env.production 文件，设置正确的生产环境配置"
        exit 1
    fi
    
    log_success "环境配置文件检查完成"
}

# 构建前端项目
build_frontend() {
    log_info "构建前端项目..."
    
    # 构建 Admin 后台
    log_info "构建 Admin 后台..."
    cd web/admin
    if [ ! -d "node_modules" ]; then
        npm install
    fi
    npm run build
    cd ../..
    
    # 构建 Marketplace
    log_info "构建 Marketplace..."
    cd web/marketplace
    if [ ! -d "node_modules" ]; then
        npm install
    fi
    npm run build
    cd ../..
    
    log_success "前端项目构建完成"
}

# 构建后端服务
build_backend() {
    log_info "构建后端服务..."
    
    # 确保 Go 模块已下载
    go mod download
    
    # 构建各个服务
    log_info "构建 Config Center..."
    CGO_ENABLED=0 GOOS=linux go build -o build/config-center ./cmd/config-center
    
    log_info "构建 Admin API..."
    CGO_ENABLED=0 GOOS=linux go build -o build/admin-api ./cmd/admin-api
    
    log_info "构建 Marketplace API..."
    CGO_ENABLED=0 GOOS=linux go build -o build/marketplace-api ./cmd/marketplace-api
    
    log_success "后端服务构建完成"
}

# 数据库迁移
run_migrations() {
    log_info "运行数据库迁移..."
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    sleep 10
    
    # 运行迁移
    docker-compose -f deployments/docker-compose.prod.yml exec -T postgres psql -U ${DB_USER:-laojun} -d ${DB_NAME:-laojun_prod} -f /docker-entrypoint-initdb.d/migration.sql || true
    
    log_success "数据库迁移完成"
}

# 部署服务
deploy_services() {
    log_info "部署服务..."
    
    # 加载环境变量
    export $(cat .env.production | grep -v '^#' | xargs)
    
    # 停止现有服务
    log_info "停止现有服务..."
    docker-compose -f deployments/docker-compose.prod.yml down || true
    
    # 构建并启动服务
    log_info "构建并启动服务..."
    docker-compose -f deployments/docker-compose.prod.yml up --build -d
    
    log_success "服务部署完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待服务启动
    sleep 30
    
    # 检查各个服务
    services=("config-center:8090" "admin-api:8080" "marketplace-api:8082")
    
    for service in "${services[@]}"; do
        IFS=':' read -r name port <<< "$service"
        log_info "检查 $name 服务..."
        
        if curl -f http://localhost:$port/health &> /dev/null; then
            log_success "$name 服务运行正常"
        else
            log_warning "$name 服务可能未正常启动，请检查日志"
        fi
    done
    
    # 检查前端服务
    if curl -f http://localhost:3000 &> /dev/null; then
        log_success "Admin 前端服务运行正常"
    else
        log_warning "Admin 前端服务可能未正常启动"
    fi
    
    if curl -f http://localhost:3001 &> /dev/null; then
        log_success "Marketplace 前端服务运行正常"
    else
        log_warning "Marketplace 前端服务可能未正常启动"
    fi
}

# 显示部署信息
show_deployment_info() {
    log_success "部署完成！"
    echo ""
    echo "服务访问地址："
    echo "  - Admin 后台: http://localhost:3000"
    echo "  - Marketplace: http://localhost:3001"
    echo "  - Admin API: http://localhost:8080"
    echo "  - Marketplace API: http://localhost:8082"
    echo "  - Config Center: http://localhost:8090"
    echo "  - Grafana 监控: http://localhost:3002"
    echo "  - Prometheus: http://localhost:9090"
    echo ""
    echo "查看服务状态："
    echo "  docker-compose -f deployments/docker-compose.prod.yml ps"
    echo ""
    echo "查看服务日志："
    echo "  docker-compose -f deployments/docker-compose.prod.yml logs -f [service-name]"
    echo ""
    echo "停止服务："
    echo "  docker-compose -f deployments/docker-compose.prod.yml down"
}

# 主函数
main() {
    log_info "开始部署太上老君系统..."
    
    check_requirements
    create_directories
    check_env_file
    build_frontend
    build_backend
    deploy_services
    run_migrations
    health_check
    show_deployment_info
    
    log_success "部署流程完成！"
}

# 脚本参数处理
case "${1:-}" in
    "check")
        check_requirements
        ;;
    "build")
        build_frontend
        build_backend
        ;;
    "deploy")
        deploy_services
        ;;
    "health")
        health_check
        ;;
    *)
        main
        ;;
esac