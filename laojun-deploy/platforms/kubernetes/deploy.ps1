# 太上老君微服务平台 Kubernetes 部署脚本 (PowerShell)
# 使用方法: .\deploy.ps1 [install|uninstall|upgrade|status]

param(
    [Parameter(Position=0)]
    [ValidateSet("install", "uninstall", "upgrade", "status", "help")]
    [string]$Command = "help"
)

$NAMESPACE_MAIN = "taishanglaojun"
$NAMESPACE_MONITORING = "taishanglaojun-monitoring"
$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path

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

# 检查kubectl是否可用
function Test-Kubectl {
    try {
        $null = Get-Command kubectl -ErrorAction Stop
        $null = kubectl cluster-info 2>$null
        Write-Success "kubectl 连接正常"
        return $true
    }
    catch {
        Write-Error "kubectl 未找到或无法连接到 Kubernetes 集群"
        return $false
    }
}

# 检查必要的文件
function Test-Files {
    $files = @(
        "namespace.yaml",
        "rbac.yaml",
        "configmaps.yaml",
        "services.yaml",
        "deployments.yaml",
        "ingress.yaml"
    )
    
    foreach ($file in $files) {
        $filePath = Join-Path $SCRIPT_DIR $file
        if (-not (Test-Path $filePath)) {
            Write-Error "文件 $file 不存在"
            return $false
        }
    }
    
    Write-Success "所有必要文件检查完成"
    return $true
}

# 安装函数
function Install-Platform {
    Write-Info "开始安装太上老君微服务平台..."
    
    try {
        # 创建命名空间
        Write-Info "创建命名空间..."
        kubectl apply -f "$SCRIPT_DIR/namespace.yaml"
        
        # 等待命名空间创建完成
        kubectl wait --for=condition=Active namespace/$NAMESPACE_MAIN --timeout=60s
        kubectl wait --for=condition=Active namespace/$NAMESPACE_MONITORING --timeout=60s
        
        # 创建 RBAC
        Write-Info "创建 RBAC 配置..."
        kubectl apply -f "$SCRIPT_DIR/rbac.yaml"
        
        # 创建 ConfigMaps
        Write-Info "创建 ConfigMaps..."
        kubectl apply -f "$SCRIPT_DIR/configmaps.yaml"
        
        # 创建 Services
        Write-Info "创建 Services..."
        kubectl apply -f "$SCRIPT_DIR/services.yaml"
        
        # 创建 Deployments
        Write-Info "创建 Deployments..."
        kubectl apply -f "$SCRIPT_DIR/deployments.yaml"
        
        # 等待部署完成
        Write-Info "等待部署完成..."
        kubectl wait --for=condition=available deployment --all -n $NAMESPACE_MAIN --timeout=300s
        kubectl wait --for=condition=available deployment --all -n $NAMESPACE_MONITORING --timeout=300s
        
        # 创建 Ingress
        Write-Info "创建 Ingress..."
        kubectl apply -f "$SCRIPT_DIR/ingress.yaml"
        
        Write-Success "太上老君微服务平台安装完成！"
        
        # 显示访问信息
        Show-AccessInfo
    }
    catch {
        Write-Error "安装过程中发生错误: $($_.Exception.Message)"
        exit 1
    }
}

# 卸载函数
function Uninstall-Platform {
    Write-Warning "开始卸载太上老君微服务平台..."
    
    try {
        # 删除 Ingress
        Write-Info "删除 Ingress..."
        kubectl delete -f "$SCRIPT_DIR/ingress.yaml" --ignore-not-found=true
        
        # 删除 Deployments
        Write-Info "删除 Deployments..."
        kubectl delete -f "$SCRIPT_DIR/deployments.yaml" --ignore-not-found=true
        
        # 删除 Services
        Write-Info "删除 Services..."
        kubectl delete -f "$SCRIPT_DIR/services.yaml" --ignore-not-found=true
        
        # 删除 ConfigMaps
        Write-Info "删除 ConfigMaps..."
        kubectl delete -f "$SCRIPT_DIR/configmaps.yaml" --ignore-not-found=true
        
        # 删除 RBAC
        Write-Info "删除 RBAC 配置..."
        kubectl delete -f "$SCRIPT_DIR/rbac.yaml" --ignore-not-found=true
        
        # 删除命名空间
        Write-Info "删除命名空间..."
        kubectl delete -f "$SCRIPT_DIR/namespace.yaml" --ignore-not-found=true
        
        Write-Success "太上老君微服务平台卸载完成！"
    }
    catch {
        Write-Error "卸载过程中发生错误: $($_.Exception.Message)"
        exit 1
    }
}

# 升级函数
function Update-Platform {
    Write-Info "开始升级太上老君微服务平台..."
    
    try {
        # 更新 ConfigMaps
        Write-Info "更新 ConfigMaps..."
        kubectl apply -f "$SCRIPT_DIR/configmaps.yaml"
        
        # 更新 Services
        Write-Info "更新 Services..."
        kubectl apply -f "$SCRIPT_DIR/services.yaml"
        
        # 滚动更新 Deployments
        Write-Info "滚动更新 Deployments..."
        kubectl apply -f "$SCRIPT_DIR/deployments.yaml"
        
        # 等待滚动更新完成
        Write-Info "等待滚动更新完成..."
        kubectl rollout status deployment --all -n $NAMESPACE_MAIN --timeout=300s
        kubectl rollout status deployment --all -n $NAMESPACE_MONITORING --timeout=300s
        
        # 更新 Ingress
        Write-Info "更新 Ingress..."
        kubectl apply -f "$SCRIPT_DIR/ingress.yaml"
        
        Write-Success "太上老君微服务平台升级完成！"
    }
    catch {
        Write-Error "升级过程中发生错误: $($_.Exception.Message)"
        exit 1
    }
}

# 状态检查函数
function Get-PlatformStatus {
    Write-Info "检查太上老君微服务平台状态..."
    
    Write-Host ""
    Write-Info "命名空间状态:"
    try {
        kubectl get namespaces $NAMESPACE_MAIN $NAMESPACE_MONITORING
    }
    catch {
        Write-Warning "命名空间不存在"
    }
    
    Write-Host ""
    Write-Info "Pod 状态:"
    try {
        kubectl get pods -n $NAMESPACE_MAIN
    }
    catch {
        Write-Warning "主命名空间中没有 Pod"
    }
    
    try {
        kubectl get pods -n $NAMESPACE_MONITORING
    }
    catch {
        Write-Warning "监控命名空间中没有 Pod"
    }
    
    Write-Host ""
    Write-Info "Service 状态:"
    try {
        kubectl get services -n $NAMESPACE_MAIN
    }
    catch {
        Write-Warning "主命名空间中没有 Service"
    }
    
    try {
        kubectl get services -n $NAMESPACE_MONITORING
    }
    catch {
        Write-Warning "监控命名空间中没有 Service"
    }
    
    Write-Host ""
    Write-Info "Ingress 状态:"
    try {
        kubectl get ingress -n $NAMESPACE_MAIN
    }
    catch {
        Write-Warning "主命名空间中没有 Ingress"
    }
    
    try {
        kubectl get ingress -n $NAMESPACE_MONITORING
    }
    catch {
        Write-Warning "监控命名空间中没有 Ingress"
    }
}

# 显示访问信息
function Show-AccessInfo {
    Write-Host ""
    Write-Info "访问信息:"
    Write-Host "请在 C:\Windows\System32\drivers\etc\hosts 文件中添加以下条目:"
    Write-Host "127.0.0.1 taishanglaojun.local"
    Write-Host "127.0.0.1 api.taishanglaojun.local"
    Write-Host "127.0.0.1 monitoring.taishanglaojun.local"
    Write-Host "127.0.0.1 prometheus.taishanglaojun.local"
    Write-Host "127.0.0.1 grafana.taishanglaojun.local"
    Write-Host "127.0.0.1 jaeger.taishanglaojun.local"
    Write-Host "127.0.0.1 kibana.taishanglaojun.local"
    Write-Host ""
    Write-Host "服务访问地址:"
    Write-Host "- 服务发现: http://api.taishanglaojun.local/discovery"
    Write-Host "- 监控服务: http://api.taishanglaojun.local/monitoring"
    Write-Host "- Prometheus: http://prometheus.taishanglaojun.local"
    Write-Host "- Grafana: http://grafana.taishanglaojun.local (admin/admin123)"
    Write-Host "- Jaeger: http://jaeger.taishanglaojun.local"
    Write-Host "- Kibana: http://kibana.taishanglaojun.local"
    Write-Host ""
    Write-Host "监控面板: http://monitoring.taishanglaojun.local (admin/admin123)"
}

# 显示帮助信息
function Show-Help {
    Write-Host "太上老君微服务平台 Kubernetes 部署脚本"
    Write-Host ""
    Write-Host "使用方法:"
    Write-Host "  .\deploy.ps1 [命令]"
    Write-Host ""
    Write-Host "命令:"
    Write-Host "  install    安装平台"
    Write-Host "  uninstall  卸载平台"
    Write-Host "  upgrade    升级平台"
    Write-Host "  status     查看状态"
    Write-Host "  help       显示帮助"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\deploy.ps1 install    # 安装平台"
    Write-Host "  .\deploy.ps1 status     # 查看状态"
    Write-Host "  .\deploy.ps1 upgrade    # 升级平台"
    Write-Host "  .\deploy.ps1 uninstall  # 卸载平台"
}

# 主函数
switch ($Command) {
    "install" {
        if (-not (Test-Kubectl)) { exit 1 }
        if (-not (Test-Files)) { exit 1 }
        Install-Platform
    }
    "uninstall" {
        if (-not (Test-Kubectl)) { exit 1 }
        Uninstall-Platform
    }
    "upgrade" {
        if (-not (Test-Kubectl)) { exit 1 }
        if (-not (Test-Files)) { exit 1 }
        Update-Platform
    }
    "status" {
        if (-not (Test-Kubectl)) { exit 1 }
        Get-PlatformStatus
    }
    "help" {
        Show-Help
    }
    default {
        Write-Error "未知命令: $Command"
        Show-Help
        exit 1
    }
}