#!/bin/bash

# 太上老君微服务平台 Helm 部署脚本
# 作者: TaiShang LaoJun Team
# 版本: 1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
CHART_NAME="taishanglaojun"
RELEASE_NAME="taishanglaojun"
NAMESPACE="taishanglaojun"
MONITORING_NAMESPACE="taishanglaojun-monitoring"
CHART_PATH="./taishanglaojun"
VALUES_FILE="values.yaml"

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
    log_info "检查依赖..."
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查 helm
    if ! command -v helm &> /dev/null; then
        log_error "helm 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查 Kubernetes 连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 检查文件
check_files() {
    log_info "检查文件..."
    
    if [ ! -d "$CHART_PATH" ]; then
        log_error "Chart 目录不存在: $CHART_PATH"
        exit 1
    fi
    
    if [ ! -f "$CHART_PATH/Chart.yaml" ]; then
        log_error "Chart.yaml 文件不存在"
        exit 1
    fi
    
    if [ ! -f "$CHART_PATH/$VALUES_FILE" ]; then
        log_error "values.yaml 文件不存在"
        exit 1
    fi
    
    log_success "文件检查通过"
}

# 创建命名空间
create_namespaces() {
    log_info "创建命名空间..."
    
    # 创建主命名空间
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        kubectl create namespace $NAMESPACE
        kubectl label namespace $NAMESPACE name=$NAMESPACE
        log_success "创建命名空间: $NAMESPACE"
    else
        log_info "命名空间已存在: $NAMESPACE"
    fi
    
    # 创建监控命名空间
    if ! kubectl get namespace $MONITORING_NAMESPACE &> /dev/null; then
        kubectl create namespace $MONITORING_NAMESPACE
        kubectl label namespace $MONITORING_NAMESPACE name=$MONITORING_NAMESPACE
        log_success "创建命名空间: $MONITORING_NAMESPACE"
    else
        log_info "命名空间已存在: $MONITORING_NAMESPACE"
    fi
}

# 安装或升级 Chart
install_chart() {
    log_info "安装/升级 Helm Chart..."
    
    # 更新依赖
    helm dependency update $CHART_PATH
    
    # 安装或升级
    helm upgrade --install $RELEASE_NAME $CHART_PATH \
        --namespace $NAMESPACE \
        --create-namespace \
        --values $CHART_PATH/$VALUES_FILE \
        --timeout 10m \
        --wait
    
    log_success "Chart 安装/升级完成"
}

# 验证部署
verify_deployment() {
    log_info "验证部署..."
    
    # 等待 Pod 就绪
    log_info "等待 Pod 就绪..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=taishanglaojun-platform -n $NAMESPACE --timeout=300s
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=taishanglaojun-monitoring -n $MONITORING_NAMESPACE --timeout=300s
    
    # 检查服务状态
    log_info "检查服务状态..."
    kubectl get pods -n $NAMESPACE
    kubectl get pods -n $MONITORING_NAMESPACE
    kubectl get svc -n $NAMESPACE
    kubectl get svc -n $MONITORING_NAMESPACE
    
    log_success "部署验证完成"
}

# 显示访问信息
show_access_info() {
    log_info "访问信息:"
    echo ""
    echo "请在 hosts 文件中添加以下条目:"
    echo "127.0.0.1 taishanglaojun.local"
    echo "127.0.0.1 api.taishanglaojun.local"
    echo "127.0.0.1 monitoring.taishanglaojun.local"
    echo "127.0.0.1 prometheus.taishanglaojun.local"
    echo "127.0.0.1 grafana.taishanglaojun.local"
    echo "127.0.0.1 jaeger.taishanglaojun.local"
    echo "127.0.0.1 kibana.taishanglaojun.local"
    echo ""
    echo "服务访问地址:"
    echo "- 服务发现 API: http://api.taishanglaojun.local/discovery"
    echo "- 监控 API: http://api.taishanglaojun.local/monitoring"
    echo "- Prometheus: http://prometheus.taishanglaojun.local"
    echo "- Grafana: http://grafana.taishanglaojun.local (admin/admin123)"
    echo "- Jaeger: http://jaeger.taishanglaojun.local"
    echo "- Kibana: http://kibana.taishanglaojun.local"
    echo "- 监控面板: http://monitoring.taishanglaojun.local (admin/admin123)"
    echo ""
}

# 卸载
uninstall() {
    log_info "卸载 Helm Chart..."
    
    helm uninstall $RELEASE_NAME -n $NAMESPACE || true
    
    # 删除命名空间
    kubectl delete namespace $NAMESPACE || true
    kubectl delete namespace $MONITORING_NAMESPACE || true
    
    log_success "卸载完成"
}

# 查看状态
status() {
    log_info "查看部署状态..."
    
    echo "Helm Release 状态:"
    helm status $RELEASE_NAME -n $NAMESPACE
    
    echo ""
    echo "Pod 状态:"
    kubectl get pods -n $NAMESPACE
    kubectl get pods -n $MONITORING_NAMESPACE
    
    echo ""
    echo "Service 状态:"
    kubectl get svc -n $NAMESPACE
    kubectl get svc -n $MONITORING_NAMESPACE
    
    echo ""
    echo "Ingress 状态:"
    kubectl get ingress -n $NAMESPACE
    kubectl get ingress -n $MONITORING_NAMESPACE
}

# 查看日志
logs() {
    local component=${2:-"discovery"}
    local namespace=$NAMESPACE
    
    if [[ "$component" == "prometheus" || "$component" == "grafana" || "$component" == "jaeger" ]]; then
        namespace=$MONITORING_NAMESPACE
    fi
    
    log_info "查看 $component 日志..."
    kubectl logs -f deployment/taishanglaojun-$component -n $namespace
}

# 端口转发
port_forward() {
    local service=${2:-"discovery"}
    local port=${3:-"8081"}
    local namespace=$NAMESPACE
    
    if [[ "$service" == "prometheus" || "$service" == "grafana" || "$service" == "jaeger" ]]; then
        namespace=$MONITORING_NAMESPACE
    fi
    
    log_info "端口转发 $service:$port..."
    kubectl port-forward svc/taishanglaojun-$service $port:$port -n $namespace
}

# 主函数
main() {
    case "${1:-install}" in
        "install")
            check_dependencies
            check_files
            create_namespaces
            install_chart
            verify_deployment
            show_access_info
            ;;
        "upgrade")
            check_dependencies
            check_files
            install_chart
            verify_deployment
            ;;
        "uninstall")
            uninstall
            ;;
        "status")
            status
            ;;
        "logs")
            logs "$@"
            ;;
        "port-forward")
            port_forward "$@"
            ;;
        "help"|"-h"|"--help")
            echo "用法: $0 [命令]"
            echo ""
            echo "命令:"
            echo "  install       安装平台 (默认)"
            echo "  upgrade       升级平台"
            echo "  uninstall     卸载平台"
            echo "  status        查看状态"
            echo "  logs [组件]   查看日志"
            echo "  port-forward [服务] [端口]  端口转发"
            echo "  help          显示帮助"
            echo ""
            echo "示例:"
            echo "  $0 install"
            echo "  $0 logs discovery"
            echo "  $0 port-forward prometheus 9090"
            ;;
        *)
            log_error "未知命令: $1"
            echo "使用 '$0 help' 查看帮助"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"