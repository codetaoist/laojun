#!/bin/bash

# 太上老君系统部署入口脚本
# 使用方法: ./deploy.sh [环境] [操作]
# 环境: dev|staging|prod (默认: prod)
# 操作: build|deploy|restart|stop|logs|backup|health|cleanup|help (默认: deploy)

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "太上老君系统部署脚本"
    echo ""
    echo "使用方法:"
    echo "  ./deploy.sh [环境] [操作]"
    echo ""
    echo "环境:"
    echo "  dev      - 开发环境"
    echo "  staging  - 预发布环境"
    echo "  prod     - 生产环境 (默认)"
    echo ""
    echo "操作:"
    echo "  build    - 构建镜像"
    echo "  deploy   - 部署服务 (默认)"
    echo "  restart  - 重启服务"
    echo "  stop     - 停止服务"
    echo "  logs     - 查看日志"
    echo "  backup   - 备份数据"
    echo "  health   - 健康检查"
    echo "  cleanup  - 清理资源"
    echo "  help     - 显示帮助"
    echo ""
    echo "示例:"
    echo "  ./deploy.sh prod deploy    # 部署生产环境"
    echo "  ./deploy.sh dev build      # 构建开发环境镜像"
    echo "  ./deploy.sh prod logs      # 查看生产环境日志"
}

# 检查参数
if [[ "$1" == "help" || "$1" == "-h" || "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# 检查部署目录
if [[ ! -d "$SCRIPT_DIR/deploy" ]]; then
    echo "错误: 找不到 deploy 目录"
    echo "请确保在项目根目录运行此脚本"
    exit 1
fi

# 切换到部署脚本目录
cd "$SCRIPT_DIR/deploy/scripts"

# 调用实际的部署脚本
log_info "调用部署脚本: ./deploy.sh $@"
exec ./deploy.sh "$@"