# 太上老君微服务平台 Helm 部署脚本 (PowerShell)
# 作者: TaiShang LaoJun Team
# 版本: 1.0.0

param(
    [Parameter(Position=0)]
    [string]$Command = "install",
    
    [Parameter(Position=1)]
    [string]$Component = "discovery",
    
    [Parameter(Position=2)]
    [string]$Port = "8081"
)

# 配置变量
$ChartName = "taishanglaojun"
$ReleaseName = "taishanglaojun"
$Namespace = "taishanglaojun"
$MonitoringNamespace = "taishanglaojun-monitoring"
$ChartPath = ".\taishanglaojun"
$ValuesFile = "values.yaml"

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    # 检查 kubectl
    try {
        kubectl version --client | Out-Null
    }
    catch {
        Write-Error "kubectl 未安装或不在 PATH 中"
        exit 1
    }
    
    # 检查 helm
    try {
        helm version | Out-Null
    }
    catch {
        Write-Error "helm 未安装或不在 PATH 中"
        exit 1
    }
    
    # 检查 Kubernetes 连接
    try {
        kubectl cluster-info | Out-Null
    }
    catch {
        Write-Error "无法连接到 Kubernetes 集群"
        exit 1
    }
    
    Write-Success "依赖检查通过"
}

# 检查文件
function Test-Files {
    Write-Info "检查文件..."
    
    if (-not (Test-Path $ChartPath)) {
        Write-Error "Chart 目录不存在: $ChartPath"
        exit 1
    }
    
    if (-not (Test-Path "$ChartPath\Chart.yaml")) {
        Write-Error "Chart.yaml 文件不存在"
        exit 1
    }
    
    if (-not (Test-Path "$ChartPath\$ValuesFile")) {
        Write-Error "values.yaml 文件不存在"
        exit 1
    }
    
    Write-Success "文件检查通过"
}

# 创建命名空间
function New-Namespaces {
    Write-Info "创建命名空间..."
    
    # 创建主命名空间
    try {
        kubectl get namespace $Namespace | Out-Null
        Write-Info "命名空间已存在: $Namespace"
    }
    catch {
        kubectl create namespace $Namespace
        kubectl label namespace $Namespace name=$Namespace
        Write-Success "创建命名空间: $Namespace"
    }
    
    # 创建监控命名空间
    try {
        kubectl get namespace $MonitoringNamespace | Out-Null
        Write-Info "命名空间已存在: $MonitoringNamespace"
    }
    catch {
        kubectl create namespace $MonitoringNamespace
        kubectl label namespace $MonitoringNamespace name=$MonitoringNamespace
        Write-Success "创建命名空间: $MonitoringNamespace"
    }
}

# 安装或升级 Chart
function Install-Chart {
    Write-Info "安装/升级 Helm Chart..."
    
    # 更新依赖
    helm dependency update $ChartPath
    
    # 安装或升级
    helm upgrade --install $ReleaseName $ChartPath `
        --namespace $Namespace `
        --create-namespace `
        --values "$ChartPath\$ValuesFile" `
        --timeout 10m `
        --wait
    
    Write-Success "Chart 安装/升级完成"
}

# 验证部署
function Test-Deployment {
    Write-Info "验证部署..."
    
    # 等待 Pod 就绪
    Write-Info "等待 Pod 就绪..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=taishanglaojun-platform -n $Namespace --timeout=300s
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=taishanglaojun-monitoring -n $MonitoringNamespace --timeout=300s
    
    # 检查服务状态
    Write-Info "检查服务状态..."
    kubectl get pods -n $Namespace
    kubectl get pods -n $MonitoringNamespace
    kubectl get svc -n $Namespace
    kubectl get svc -n $MonitoringNamespace
    
    Write-Success "部署验证完成"
}

# 显示访问信息
function Show-AccessInfo {
    Write-Info "访问信息:"
    Write-Host ""
    Write-Host "请在 hosts 文件中添加以下条目:"
    Write-Host "127.0.0.1 taishanglaojun.local"
    Write-Host "127.0.0.1 api.taishanglaojun.local"
    Write-Host "127.0.0.1 monitoring.taishanglaojun.local"
    Write-Host "127.0.0.1 prometheus.taishanglaojun.local"
    Write-Host "127.0.0.1 grafana.taishanglaojun.local"
    Write-Host "127.0.0.1 jaeger.taishanglaojun.local"
    Write-Host "127.0.0.1 kibana.taishanglaojun.local"
    Write-Host ""
    Write-Host "服务访问地址:"
    Write-Host "- 服务发现 API: http://api.taishanglaojun.local/discovery"
    Write-Host "- 监控 API: http://api.taishanglaojun.local/monitoring"
    Write-Host "- Prometheus: http://prometheus.taishanglaojun.local"
    Write-Host "- Grafana: http://grafana.taishanglaojun.local (admin/admin123)"
    Write-Host "- Jaeger: http://jaeger.taishanglaojun.local"
    Write-Host "- Kibana: http://kibana.taishanglaojun.local"
    Write-Host "- 监控面板: http://monitoring.taishanglaojun.local (admin/admin123)"
    Write-Host ""
}

# 卸载
function Uninstall-Chart {
    Write-Info "卸载 Helm Chart..."
    
    try {
        helm uninstall $ReleaseName -n $Namespace
    }
    catch {
        Write-Warning "Helm release 卸载失败或不存在"
    }
    
    # 删除命名空间
    try {
        kubectl delete namespace $Namespace
    }
    catch {
        Write-Warning "命名空间删除失败或不存在: $Namespace"
    }
    
    try {
        kubectl delete namespace $MonitoringNamespace
    }
    catch {
        Write-Warning "命名空间删除失败或不存在: $MonitoringNamespace"
    }
    
    Write-Success "卸载完成"
}

# 查看状态
function Get-Status {
    Write-Info "查看部署状态..."
    
    Write-Host "Helm Release 状态:"
    helm status $ReleaseName -n $Namespace
    
    Write-Host ""
    Write-Host "Pod 状态:"
    kubectl get pods -n $Namespace
    kubectl get pods -n $MonitoringNamespace
    
    Write-Host ""
    Write-Host "Service 状态:"
    kubectl get svc -n $Namespace
    kubectl get svc -n $MonitoringNamespace
    
    Write-Host ""
    Write-Host "Ingress 状态:"
    kubectl get ingress -n $Namespace
    kubectl get ingress -n $MonitoringNamespace
}

# 查看日志
function Get-Logs {
    param([string]$ComponentName)
    
    $TargetNamespace = $Namespace
    
    if ($ComponentName -in @("prometheus", "grafana", "jaeger")) {
        $TargetNamespace = $MonitoringNamespace
    }
    
    Write-Info "查看 $ComponentName 日志..."
    kubectl logs -f deployment/taishanglaojun-$ComponentName -n $TargetNamespace
}

# 端口转发
function Start-PortForward {
    param([string]$ServiceName, [string]$ServicePort)
    
    $TargetNamespace = $Namespace
    
    if ($ServiceName -in @("prometheus", "grafana", "jaeger")) {
        $TargetNamespace = $MonitoringNamespace
    }
    
    Write-Info "端口转发 ${ServiceName}:${ServicePort}..."
    kubectl port-forward svc/taishanglaojun-$ServiceName ${ServicePort}:${ServicePort} -n $TargetNamespace
}

# 显示帮助
function Show-Help {
    Write-Host "用法: .\deploy.ps1 [命令] [参数]"
    Write-Host ""
    Write-Host "命令:"
    Write-Host "  install       安装平台 (默认)"
    Write-Host "  upgrade       升级平台"
    Write-Host "  uninstall     卸载平台"
    Write-Host "  status        查看状态"
    Write-Host "  logs [组件]   查看日志"
    Write-Host "  port-forward [服务] [端口]  端口转发"
    Write-Host "  help          显示帮助"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\deploy.ps1 install"
    Write-Host "  .\deploy.ps1 logs discovery"
    Write-Host "  .\deploy.ps1 port-forward prometheus 9090"
}

# 主函数
switch ($Command.ToLower()) {
    "install" {
        Test-Dependencies
        Test-Files
        New-Namespaces
        Install-Chart
        Test-Deployment
        Show-AccessInfo
    }
    "upgrade" {
        Test-Dependencies
        Test-Files
        Install-Chart
        Test-Deployment
    }
    "uninstall" {
        Uninstall-Chart
    }
    "status" {
        Get-Status
    }
    "logs" {
        Get-Logs -ComponentName $Component
    }
    "port-forward" {
        Start-PortForward -ServiceName $Component -ServicePort $Port
    }
    "help" {
        Show-Help
    }
    default {
        Write-Error "未知命令: $Command"
        Write-Host "使用 '.\deploy.ps1 help' 查看帮助"
        exit 1
    }
}