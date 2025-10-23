#!/bin/bash

# 太上老君系统服务器快速部署脚本
# 适用于全新的 Linux 服务器环境

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
PROJECT_NAME="laojun"
DOMAIN_NAME="${1:-your-domain.com}"
EMAIL="${2:-admin@your-domain.com}"

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

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户运行此脚本"
        exit 1
    fi
}

# 更新系统
update_system() {
    log_info "更新系统包..."
    
    if command -v apt-get &> /dev/null; then
        apt-get update && apt-get upgrade -y
        apt-get install -y curl wget git unzip
    elif command -v yum &> /dev/null; then
        yum update -y
        yum install -y curl wget git unzip
    else
        log_error "不支持的操作系统"
        exit 1
    fi
    
    log_success "系统更新完成"
}

# 安装 Docker
install_docker() {
    log_info "安装 Docker..."
    
    if command -v docker &> /dev/null; then
        log_warning "Docker 已安装，跳过安装步骤"
        return
    fi
    
    # 安装 Docker
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    
    # 启动 Docker 服务
    systemctl start docker
    systemctl enable docker
    
    # 添加当前用户到 docker 组
    usermod -aG docker $USER
    
    log_success "Docker 安装完成"
}

# 安装 Docker Compose
install_docker_compose() {
    log_info "安装 Docker Compose..."
    
    if command -v docker-compose &> /dev/null; then
        log_warning "Docker Compose 已安装，跳过安装步骤"
        return
    fi
    
    # 下载 Docker Compose
    DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    
    # 设置执行权限
    chmod +x /usr/local/bin/docker-compose
    
    # 创建软链接
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    
    log_success "Docker Compose 安装完成"
}

# 安装 Nginx
install_nginx() {
    log_info "安装 Nginx..."
    
    if command -v nginx &> /dev/null; then
        log_warning "Nginx 已安装，跳过安装步骤"
        return
    fi
    
    if command -v apt-get &> /dev/null; then
        apt-get install -y nginx
    elif command -v yum &> /dev/null; then
        yum install -y nginx
    fi
    
    # 启动 Nginx 服务
    systemctl start nginx
    systemctl enable nginx
    
    log_success "Nginx 安装完成"
}

# 安装 Certbot (Let's Encrypt)
install_certbot() {
    log_info "安装 Certbot..."
    
    if command -v certbot &> /dev/null; then
        log_warning "Certbot 已安装，跳过安装步骤"
        return
    fi
    
    if command -v apt-get &> /dev/null; then
        apt-get install -y certbot python3-certbot-nginx
    elif command -v yum &> /dev/null; then
        yum install -y certbot python3-certbot-nginx
    fi
    
    log_success "Certbot 安装完成"
}

# 配置防火墙
configure_firewall() {
    log_info "配置防火墙..."
    
    if command -v ufw &> /dev/null; then
        # Ubuntu/Debian
        ufw allow ssh
        ufw allow 80/tcp
        ufw allow 443/tcp
        ufw --force enable
    elif command -v firewall-cmd &> /dev/null; then
        # CentOS/RHEL
        firewall-cmd --permanent --add-service=ssh
        firewall-cmd --permanent --add-service=http
        firewall-cmd --permanent --add-service=https
        firewall-cmd --reload
    fi
    
    log_success "防火墙配置完成"
}

# 创建项目目录
create_project_directory() {
    log_info "创建项目目录..."
    
    PROJECT_DIR="/opt/$PROJECT_NAME"
    mkdir -p "$PROJECT_DIR"
    cd "$PROJECT_DIR"
    
    log_success "项目目录创建完成: $PROJECT_DIR"
}

# 配置 SSL 证书
setup_ssl() {
    log_info "配置 SSL 证书..."
    
    if [ "$DOMAIN_NAME" = "your-domain.com" ]; then
        log_warning "跳过 SSL 配置，请手动配置域名"
        return
    fi
    
    # 获取 SSL 证书
    certbot --nginx -d "$DOMAIN_NAME" --email "$EMAIL" --agree-tos --non-interactive
    
    # 设置自动续期
    (crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | crontab -
    
    log_success "SSL 证书配置完成"
}

# 创建部署用户
create_deploy_user() {
    log_info "创建部署用户..."
    
    if id "deploy" &>/dev/null; then
        log_warning "部署用户已存在，跳过创建"
        return
    fi
    
    useradd -m -s /bin/bash deploy
    usermod -aG docker deploy
    
    # 设置 sudo 权限
    echo "deploy ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
    
    log_success "部署用户创建完成"
}

# 显示部署信息
show_deployment_info() {
    log_success "服务器环境配置完成！"
    echo ""
    echo "=== 部署信息 ==="
    echo "项目目录: /opt/$PROJECT_NAME"
    echo "域名: $DOMAIN_NAME"
    echo "邮箱: $EMAIL"
    echo ""
    echo "=== 下一步操作 ==="
    echo "1. 将项目代码上传到服务器: /opt/$PROJECT_NAME"
    echo "2. 修改配置文件: .env.production"
    echo "3. 运行部署脚本: ./deploy.sh prod deploy"
    echo ""
    echo "=== 常用命令 ==="
    echo "查看服务状态: docker-compose -f docker-compose.prod.yml ps"
    echo "查看日志: docker-compose -f docker-compose.prod.yml logs -f"
    echo "重启服务: ./deploy.sh prod restart"
    echo ""
}

# 主函数
main() {
    echo "=== 太上老君系统服务器快速部署脚本 ==="
    echo "域名: $DOMAIN_NAME"
    echo "邮箱: $EMAIL"
    echo ""
    
    check_root
    update_system
    install_docker
    install_docker_compose
    install_nginx
    install_certbot
    configure_firewall
    create_deploy_user
    create_project_directory
    
    if [ "$DOMAIN_NAME" != "your-domain.com" ]; then
        setup_ssl
    fi
    
    show_deployment_info
}

# 显示帮助信息
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "太上老君系统服务器快速部署脚本"
    echo ""
    echo "使用方法:"
    echo "  $0 [域名] [邮箱]"
    echo ""
    echo "参数:"
    echo "  域名    你的域名 (默认: your-domain.com)"
    echo "  邮箱    SSL 证书邮箱 (默认: admin@your-domain.com)"
    echo ""
    echo "示例:"
    echo "  $0 example.com admin@example.com"
    echo ""
    exit 0
fi

# 执行主函数
main "$@"