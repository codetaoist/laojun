#!/bin/bash

# 太上老君系统自动化部署脚本
# 使用方法: ./deploy.sh [环境] [操作]
# 环境: dev|staging|prod (默认: prod)
# 操作: build|deploy|restart|stop|logs|backup (默认: deploy)

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="laojun"
ENVIRONMENT="${1:-prod}"
ACTION="${2:-deploy}"
COMPOSE_FILE="../docker/docker-compose.${ENVIRONMENT}.yml"
ENV_FILE="../configs/.env.${ENVIRONMENT}"

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

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "系统依赖检查完成"
}

# 检查配置文件
check_config() {
    log_info "检查配置文件..."
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Docker Compose 文件不存在: $COMPOSE_FILE"
        exit 1
    fi
    
    if [ ! -f "$ENV_FILE" ]; then
        log_warning "环境配置文件不存在: $ENV_FILE，将使用默认配置"
    fi
    
    log_success "配置文件检查完成"
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" build --no-cache
    
    log_success "Docker 镜像构建完成"
}

# 数据库迁移
run_migrations() {
    log_info "执行数据库迁移..."
    
    # 等待数据库启动
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d postgres redis
    sleep 10
    
    # 运行迁移
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" run --rm laojun-app /app/bin/db-migrate up
    
    log_success "数据库迁移完成"
}

# 部署应用
deploy_app() {
    log_info "部署应用..."
    
    # 停止旧容器
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" down
    
    # 构建镜像
    build_images
    
    # 启动服务
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 30
    
    # 检查服务状态
    check_health
    
    log_success "应用部署完成"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" restart
    
    # 等待服务启动
    sleep 15
    check_health
    
    log_success "服务重启完成"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" down
    
    log_success "服务已停止"
}

# 查看日志
show_logs() {
    log_info "显示服务日志..."
    
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" logs -f --tail=100
}

# 健康检查
check_health() {
    log_info "检查服务健康状态..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f http://localhost/health &> /dev/null; then
            log_success "服务健康检查通过"
            return 0
        fi
        
        log_info "等待服务启动... ($attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    log_error "服务健康检查失败"
    return 1
}

# 备份数据
backup_data() {
    log_info "备份数据..."
    
    local backup_dir="./backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # 备份数据库
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" exec -T postgres pg_dump -U laojun laojun > "$backup_dir/database.sql"
    
    # 备份上传文件
    if [ -d "./uploads" ]; then
        cp -r ./uploads "$backup_dir/"
    fi
    
    # 压缩备份
    tar -czf "$backup_dir.tar.gz" -C "$backup_dir" .
    rm -rf "$backup_dir"
    
    log_success "数据备份完成: $backup_dir.tar.gz"
}

# 清理旧镜像
cleanup_images() {
    log_info "清理旧的 Docker 镜像..."
    
    docker image prune -f
    docker system prune -f
    
    log_success "镜像清理完成"
}

# 显示帮助信息
show_help() {
    echo "太上老君系统部署脚本"
    echo ""
    echo "使用方法:"
    echo "  $0 [环境] [操作]"
    echo ""
    echo "环境:"
    echo "  dev      开发环境"
    echo "  staging  测试环境"
    echo "  prod     生产环境 (默认)"
    echo ""
    echo "操作:"
    echo "  build    构建镜像"
    echo "  deploy   部署应用 (默认)"
    echo "  restart  重启服务"
    echo "  stop     停止服务"
    echo "  logs     查看日志"
    echo "  backup   备份数据"
    echo "  health   健康检查"
    echo "  cleanup  清理镜像"
    echo "  help     显示帮助"
    echo ""
    echo "示例:"
    echo "  $0 prod deploy    # 部署到生产环境"
    echo "  $0 dev logs       # 查看开发环境日志"
    echo "  $0 prod backup    # 备份生产环境数据"
}

# 主函数
main() {
    cd "$SCRIPT_DIR"
    
    case "$ACTION" in
        "build")
            check_dependencies
            check_config
            build_images
            ;;
        "deploy")
            check_dependencies
            check_config
            deploy_app
            ;;
        "restart")
            check_dependencies
            check_config
            restart_services
            ;;
        "stop")
            check_dependencies
            check_config
            stop_services
            ;;
        "logs")
            check_dependencies
            check_config
            show_logs
            ;;
        "backup")
            check_dependencies
            check_config
            backup_data
            ;;
        "health")
            check_health
            ;;
        "cleanup")
            check_dependencies
            cleanup_images
            ;;
        "help")
            show_help
            ;;
        *)
            log_error "未知操作: $ACTION"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"