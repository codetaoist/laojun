#!/bin/bash

# 太上老君微服务平台 Kubernetes 部署脚本
# 使用方法: ./deploy.sh [install|uninstall|upgrade|status]

set -e

NAMESPACE_MAIN="taishanglaojun"
NAMESPACE_MONITORING="taishanglaojun-monitoring"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

# 检查kubectl是否可用
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未找到，请先安装 kubectl"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    log_success "kubectl 连接正常"
}

# 检查必要的文件
check_files() {
    local files=(
        "namespace.yaml"
        "rbac.yaml"
        "configmaps.yaml"
        "services.yaml"
        "deployments.yaml"
        "ingress.yaml"
    )
    
    for file in "${files[@]}"; do
        if [[ ! -f "$SCRIPT_DIR/$file" ]]; then
            log_error "文件 $file 不存在"
            exit 1
        fi
    done
    
    log_success "所有必要文件检查完成"
}

# 安装函数
install() {
    log_info "开始安装太上老君微服务平台..."
    
    # 创建命名空间
    log_info "创建命名空间..."
    kubectl apply -f "$SCRIPT_DIR/namespace.yaml"
    
    # 等待命名空间创建完成
    kubectl wait --for=condition=Active namespace/$NAMESPACE_MAIN --timeout=60s
    kubectl wait --for=condition=Active namespace/$NAMESPACE_MONITORING --timeout=60s
    
    # 创建 RBAC
    log_info "创建 RBAC 配置..."
    kubectl apply -f "$SCRIPT_DIR/rbac.yaml"
    
    # 创建 ConfigMaps
    log_info "创建 ConfigMaps..."
    kubectl apply -f "$SCRIPT_DIR/configmaps.yaml"
    
    # 创建 Services
    log_info "创建 Services..."
    kubectl apply -f "$SCRIPT_DIR/services.yaml"
    
    # 创建 Deployments
    log_info "创建 Deployments..."
    kubectl apply -f "$SCRIPT_DIR/deployments.yaml"
    
    # 等待部署完成
    log_info "等待部署完成..."
    kubectl wait --for=condition=available deployment --all -n $NAMESPACE_MAIN --timeout=300s
    kubectl wait --for=condition=available deployment --all -n $NAMESPACE_MONITORING --timeout=300s
    
    # 创建 Ingress
    log_info "创建 Ingress..."
    kubectl apply -f "$SCRIPT_DIR/ingress.yaml"
    
    log_success "太上老君微服务平台安装完成！"
    
    # 显示访问信息
    show_access_info
}

# 卸载函数
uninstall() {
    log_warning "开始卸载太上老君微服务平台..."
    
    # 删除 Ingress
    log_info "删除 Ingress..."
    kubectl delete -f "$SCRIPT_DIR/ingress.yaml" --ignore-not-found=true
    
    # 删除 Deployments
    log_info "删除 Deployments..."
    kubectl delete -f "$SCRIPT_DIR/deployments.yaml" --ignore-not-found=true
    
    # 删除 Services
    log_info "删除 Services..."
    kubectl delete -f "$SCRIPT_DIR/services.yaml" --ignore-not-found=true
    
    # 删除 ConfigMaps
    log_info "删除 ConfigMaps..."
    kubectl delete -f "$SCRIPT_DIR/configmaps.yaml" --ignore-not-found=true
    
    # 删除 RBAC
    log_info "删除 RBAC 配置..."
    kubectl delete -f "$SCRIPT_DIR/rbac.yaml" --ignore-not-found=true
    
    # 删除命名空间
    log_info "删除命名空间..."
    kubectl delete -f "$SCRIPT_DIR/namespace.yaml" --ignore-not-found=true
    
    log_success "太上老君微服务平台卸载完成！"
}

# 升级函数
upgrade() {
    log_info "开始升级太上老君微服务平台..."
    
    # 更新 ConfigMaps
    log_info "更新 ConfigMaps..."
    kubectl apply -f "$SCRIPT_DIR/configmaps.yaml"
    
    # 更新 Services
    log_info "更新 Services..."
    kubectl apply -f "$SCRIPT_DIR/services.yaml"
    
    # 滚动更新 Deployments
    log_info "滚动更新 Deployments..."
    kubectl apply -f "$SCRIPT_DIR/deployments.yaml"
    
    # 等待滚动更新完成
    log_info "等待滚动更新完成..."
    kubectl rollout status deployment --all -n $NAMESPACE_MAIN --timeout=300s
    kubectl rollout status deployment --all -n $NAMESPACE_MONITORING --timeout=300s
    
    # 更新 Ingress
    log_info "更新 Ingress..."
    kubectl apply -f "$SCRIPT_DIR/ingress.yaml"
    
    log_success "太上老君微服务平台升级完成！"
}

# 状态检查函数
status() {
    log_info "检查太上老君微服务平台状态..."
    
    echo
    log_info "命名空间状态:"
    kubectl get namespaces $NAMESPACE_MAIN $NAMESPACE_MONITORING 2>/dev/null || log_warning "命名空间不存在"
    
    echo
    log_info "Pod 状态:"
    kubectl get pods -n $NAMESPACE_MAIN 2>/dev/null || log_warning "主命名空间中没有 Pod"
    kubectl get pods -n $NAMESPACE_MONITORING 2>/dev/null || log_warning "监控命名空间中没有 Pod"
    
    echo
    log_info "Service 状态:"
    kubectl get services -n $NAMESPACE_MAIN 2>/dev/null || log_warning "主命名空间中没有 Service"
    kubectl get services -n $NAMESPACE_MONITORING 2>/dev/null || log_warning "监控命名空间中没有 Service"
    
    echo
    log_info "Ingress 状态:"
    kubectl get ingress -n $NAMESPACE_MAIN 2>/dev/null || log_warning "主命名空间中没有 Ingress"
    kubectl get ingress -n $NAMESPACE_MONITORING 2>/dev/null || log_warning "监控命名空间中没有 Ingress"
}

# 显示访问信息
show_access_info() {
    echo
    log_info "访问信息:"
    echo "请在 /etc/hosts 文件中添加以下条目:"
    echo "127.0.0.1 taishanglaojun.local"
    echo "127.0.0.1 api.taishanglaojun.local"
    echo "127.0.0.1 monitoring.taishanglaojun.local"
    echo "127.0.0.1 prometheus.taishanglaojun.local"
    echo "127.0.0.1 grafana.taishanglaojun.local"
    echo "127.0.0.1 jaeger.taishanglaojun.local"
    echo "127.0.0.1 kibana.taishanglaojun.local"
    echo
    echo "服务访问地址:"
    echo "- 服务发现: http://api.taishanglaojun.local/discovery"
    echo "- 监控服务: http://api.taishanglaojun.local/monitoring"
    echo "- Prometheus: http://prometheus.taishanglaojun.local"
    echo "- Grafana: http://grafana.taishanglaojun.local (admin/admin123)"
    echo "- Jaeger: http://jaeger.taishanglaojun.local"
    echo "- Kibana: http://kibana.taishanglaojun.local"
    echo
    echo "监控面板: http://monitoring.taishanglaojun.local (admin/admin123)"
}

# 显示帮助信息
show_help() {
    echo "太上老君微服务平台 Kubernetes 部署脚本"
    echo
    echo "使用方法:"
    echo "  $0 [命令]"
    echo
    echo "命令:"
    echo "  install    安装平台"
    echo "  uninstall  卸载平台"
    echo "  upgrade    升级平台"
    echo "  status     查看状态"
    echo "  help       显示帮助"
    echo
    echo "示例:"
    echo "  $0 install    # 安装平台"
    echo "  $0 status     # 查看状态"
    echo "  $0 upgrade    # 升级平台"
    echo "  $0 uninstall  # 卸载平台"
}

# 主函数
main() {
    case "${1:-help}" in
        install)
            check_kubectl
            check_files
            install
            ;;
        uninstall)
            check_kubectl
            uninstall
            ;;
        upgrade)
            check_kubectl
            check_files
            upgrade
            ;;
        status)
            check_kubectl
            status
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