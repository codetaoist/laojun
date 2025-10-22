#!/bin/bash

# 生产环境部署脚本
# 使用方法: ./deploy-production.sh [action]
# Actions: build, deploy, stop, restart, logs, status

set -e

# 配置变量
PROJECT_NAME="laojun"
COMPOSE_FILE="deployments/docker-compose.prod.yml"
ENV_FILE=".env.prod"
BACKUP_DIR="backups"

# 颜色输出
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

# 检查必要条件
check_requirements() {
    log_info "检查部署环境..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    # 检查环境变量文件
    if [ ! -f "$ENV_FILE" ]; then
        log_error "环境变量文件 $ENV_FILE 不存在"
        log_info "请复制 .env.production.example 到 $ENV_FILE 并配置相应的值"
        exit 1
    fi
    
    # 检查docker-compose文件
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Docker Compose 文件 $COMPOSE_FILE 不存在"
        exit 1
    fi
    
    log_success "环境检查通过"
}

# 创建备份
create_backup() {
    log_info "创建数据备份..."
    
    mkdir -p "$BACKUP_DIR"
    BACKUP_FILE="$BACKUP_DIR/backup-$(date +%Y%m%d-%H%M%S).sql"
    
    # 备份数据库
    if docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps postgres | grep -q "Up"; then
        docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" exec -T postgres \
            pg_dump -U "$DB_USER" "$DB_NAME" > "$BACKUP_FILE"
        log_success "数据库备份完成: $BACKUP_FILE"
    else
        log_warning "PostgreSQL 容器未运行，跳过备份"
    fi
}

# 构建镜像
build_images() {
    log_info "构建Docker镜像..."
    
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" build --no-cache
    
    log_success "镜像构建完成"
}

# 部署服务
deploy_services() {
    log_info "部署服务..."
    
    # 先启动基础服务
    log_info "启动数据库和缓存服务..."
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d postgres redis
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    sleep 10
    
    # 启动后端服务
    log_info "启动后端服务..."
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d config-center
    sleep 5
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d admin-api marketplace-api
    
    # 启动前端服务
    log_info "启动前端服务..."
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d admin-web marketplace-web
    
    # 启动反向代理
    log_info "启动Nginx反向代理..."
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d nginx
    
    # 启动监控服务（可选）
    log_info "启动监控服务..."
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d prometheus grafana
    
    log_success "服务部署完成"
}

# 停止服务
stop_services() {
    log_info "停止所有服务..."
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" down
    log_success "服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    stop_services
    deploy_services
}

# 查看日志
view_logs() {
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" logs -f
}

# 查看状态
check_status() {
    log_info "服务状态:"
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps
    
    log_info "健康检查:"
    echo "Config Center: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/health || echo "无法连接")"
    echo "Admin API: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health || echo "无法连接")"
    echo "Marketplace API: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/health || echo "无法连接")"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "build")
            check_requirements
            build_images
            ;;
        "deploy")
            check_requirements
            create_backup
            build_images
            deploy_services
            check_status
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            check_requirements
            restart_services
            check_status
            ;;
        "logs")
            view_logs
            ;;
        "status")
            check_status
            ;;
        *)
            echo "使用方法: $0 [build|deploy|stop|restart|logs|status]"
            echo ""
            echo "Actions:"
            echo "  build   - 仅构建Docker镜像"
            echo "  deploy  - 完整部署（默认）"
            echo "  stop    - 停止所有服务"
            echo "  restart - 重启所有服务"
            echo "  logs    - 查看服务日志"
            echo "  status  - 查看服务状态"
            exit 1
            ;;
    esac
}

main "$@"