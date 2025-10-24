#!/bin/bash

# 太上老君系统 Docker 容器启动脚本
# 用于初始化容器环境和启动应用

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [ "${APP_DEBUG}" = "true" ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# 检查必要的环境变量
check_env_vars() {
    log_info "检查环境变量..."
    
    local required_vars=(
        "DB_HOST"
        "DB_PORT"
        "DB_USER"
        "DB_PASSWORD"
        "DB_NAME"
        "REDIS_HOST"
        "REDIS_PORT"
        "JWT_SECRET"
    )
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            log_error "必需的环境变量 $var 未设置"
            exit 1
        fi
    done
    
    log_info "环境变量检查完成"
}

# 等待数据库就绪
wait_for_db() {
    log_info "等待数据库连接..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
            log_info "数据库连接成功"
            return 0
        fi
        
        log_warn "数据库连接失败，重试 $attempt/$max_attempts..."
        sleep 2
        ((attempt++))
    done
    
    log_error "数据库连接超时"
    exit 1
}

# 等待Redis就绪
wait_for_redis() {
    log_info "等待Redis连接..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" ping >/dev/null 2>&1; then
            log_info "Redis连接成功"
            return 0
        fi
        
        log_warn "Redis连接失败，重试 $attempt/$max_attempts..."
        sleep 2
        ((attempt++))
    done
    
    log_error "Redis连接超时"
    exit 1
}

# 运行数据库迁移
run_migrations() {
    log_info "运行数据库迁移..."
    
    if [ -f "./migrate" ]; then
        ./migrate -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE" -path ./migrations up
        log_info "数据库迁移完成"
    else
        log_warn "未找到迁移工具，跳过数据库迁移"
    fi
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    local dirs=(
        "/app/logs"
        "/app/uploads"
        "/app/temp"
        "/app/cache"
    )
    
    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_debug "创建目录: $dir"
        fi
    done
    
    # 设置目录权限
    chown -R appuser:appgroup /app/logs /app/uploads /app/temp /app/cache 2>/dev/null || true
    
    log_info "目录创建完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 检查应用端口
    local port=${APP_PORT:-8080}
    if ! nc -z localhost "$port" 2>/dev/null; then
        log_error "应用端口 $port 不可访问"
        return 1
    fi
    
    # 检查健康检查端点
    if command -v curl >/dev/null 2>&1; then
        if ! curl -f "http://localhost:$port/health" >/dev/null 2>&1; then
            log_error "健康检查端点不可访问"
            return 1
        fi
    fi
    
    log_info "健康检查通过"
    return 0
}

# 信号处理
cleanup() {
    log_info "接收到停止信号，正在清理..."
    
    # 停止应用进程
    if [ ! -z "$APP_PID" ]; then
        kill -TERM "$APP_PID" 2>/dev/null || true
        wait "$APP_PID" 2>/dev/null || true
    fi
    
    log_info "清理完成"
    exit 0
}

# 设置信号处理
trap cleanup SIGTERM SIGINT

# 主函数
main() {
    log_info "启动太上老君系统容器..."
    log_info "应用环境: ${APP_ENV:-production}"
    log_info "调试模式: ${APP_DEBUG:-false}"
    
    # 执行初始化步骤
    check_env_vars
    create_directories
    
    # 等待依赖服务
    wait_for_db
    wait_for_redis
    
    # 运行迁移（仅在主服务中）
    if [ "${RUN_MIGRATIONS:-false}" = "true" ]; then
        run_migrations
    fi
    
    # 启动应用
    log_info "启动应用: $@"
    exec "$@" &
    APP_PID=$!
    
    # 等待应用启动
    sleep 5
    
    # 执行健康检查
    if ! health_check; then
        log_error "应用启动失败"
        exit 1
    fi
    
    log_info "应用启动成功，PID: $APP_PID"
    
    # 等待应用进程
    wait "$APP_PID"
}

# 如果脚本被直接执行
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi