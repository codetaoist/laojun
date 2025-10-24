#!/bin/bash

# 太上老君系统部署脚本
# 用途：管理系统的部署、启动、停止、更新等操作

set -e

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOCKER_DIR="$PROJECT_ROOT/deploy/docker"
CONFIG_DIR="$PROJECT_ROOT/deploy/configs"

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

# 检查环境配置
check_environment() {
    log_info "检查环境配置..."
    
    if [ ! -f "$CONFIG_DIR/.env" ]; then
        if [ -f "$CONFIG_DIR/.env.template" ]; then
            log_warning "未找到 .env 文件，正在从模板创建..."
            cp "$CONFIG_DIR/.env.template" "$CONFIG_DIR/.env"
            log_warning "请编辑 $CONFIG_DIR/.env 文件配置您的环境变量"
            return 1
        else
            log_error "未找到环境配置文件"
            exit 1
        fi
    fi
    
    log_success "环境配置检查完成"
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    cd "$DOCKER_DIR"
    
    # 构建所有服务镜像
    docker-compose build --no-cache
    
    log_success "Docker 镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    cd "$DOCKER_DIR"
    
    # 启动所有服务
    docker-compose up -d
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 30
    
    # 检查服务状态
    check_services_health
    
    log_success "服务启动完成"
    log_info "访问地址："
    log_info "  插件市场: http://localhost"
    log_info "  管理后台: http://localhost:8888"
    log_info "  管理API: http://localhost:8080"
    log_info "  插件市场API: http://localhost:8082"
    log_info "  配置中心: http://localhost:8081"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    
    cd "$DOCKER_DIR"
    docker-compose down
    
    log_success "服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    stop_services
    start_services
}

# 检查服务健康状态
check_services_health() {
    log_info "检查服务健康状态..."
    
    cd "$DOCKER_DIR"
    
    # 检查容器状态
    if docker-compose ps | grep -q "Up"; then
        log_success "服务运行正常"
        docker-compose ps
    else
        log_error "部分服务未正常启动"
        docker-compose ps
        return 1
    fi
}

# 查看日志
view_logs() {
    local service=$1
    cd "$DOCKER_DIR"
    
    if [ -z "$service" ]; then
        docker-compose logs -f
    else
        docker-compose logs -f "$service"
    fi
}

# 数据库迁移
migrate_database() {
    log_info "执行数据库迁移..."
    
    cd "$PROJECT_ROOT"
    
    # 确保数据库服务运行
    cd "$DOCKER_DIR"
    docker-compose up -d postgres redis
    
    # 等待数据库启动
    sleep 10
    
    # 执行迁移
    cd "$PROJECT_ROOT"
    make migrate-complete-up
    
    log_success "数据库迁移完成"
}

# 备份数据
backup_data() {
    local backup_dir="$PROJECT_ROOT/backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    log_info "备份数据到 $backup_dir..."
    
    cd "$DOCKER_DIR"
    
    # 备份数据库
    docker-compose exec -T postgres pg_dump -U laojun laojun > "$backup_dir/database.sql"
    
    # 备份上传文件
    if [ -d "$PROJECT_ROOT/uploads" ]; then
        cp -r "$PROJECT_ROOT/uploads" "$backup_dir/"
    fi
    
    # 备份配置文件
    cp -r "$CONFIG_DIR" "$backup_dir/"
    
    log_success "数据备份完成: $backup_dir"
}

# 更新系统
update_system() {
    log_info "更新系统..."
    
    # 备份数据
    backup_data
    
    # 停止服务
    stop_services
    
    # 构建新镜像
    build_images
    
    # 启动服务
    start_services
    
    log_success "系统更新完成"
}

# 清理系统
cleanup_system() {
    log_warning "这将删除所有容器、镜像和数据，确定要继续吗？(y/N)"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        log_info "清理系统..."
        
        cd "$DOCKER_DIR"
        
        # 停止并删除容器
        docker-compose down -v --rmi all
        
        # 清理未使用的镜像
        docker system prune -f
        
        log_success "系统清理完成"
    else
        log_info "取消清理操作"
    fi
}

# 显示帮助信息
show_help() {
    echo "太上老君系统部署脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  check       检查系统依赖和环境配置"
    echo "  build       构建 Docker 镜像"
    echo "  start       启动所有服务"
    echo "  stop        停止所有服务"
    echo "  restart     重启所有服务"
    echo "  status      检查服务状态"
    echo "  logs [服务] 查看日志（可指定具体服务）"
    echo "  migrate     执行数据库迁移"
    echo "  backup      备份数据"
    echo "  update      更新系统（包含备份、重建、重启）"
    echo "  cleanup     清理系统（删除所有数据）"
    echo "  help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start                # 启动所有服务"
    echo "  $0 logs admin-api       # 查看管理API日志"
    echo "  $0 backup               # 备份数据"
}

# 主函数
main() {
    case "${1:-help}" in
        check)
            check_dependencies
            check_environment
            ;;
        build)
            check_dependencies
            build_images
            ;;
        start)
            check_dependencies
            check_environment
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status)
            check_services_health
            ;;
        logs)
            view_logs "$2"
            ;;
        migrate)
            migrate_database
            ;;
        backup)
            backup_data
            ;;
        update)
            update_system
            ;;
        cleanup)
            cleanup_system
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"