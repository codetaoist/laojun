# 太上老君系统部署入口脚本 (PowerShell)
# 使用方法: .\deploy.ps1 [环境] [操作]
# 环境: dev|staging|prod (默认: prod)
# 操作: build|deploy|restart|stop|logs|backup|health|cleanup|help (默认: deploy)

param(
    [string]$Environment = "prod",
    [string]$Action = "deploy"
)

# 颜色定义
$Green = "Green"
$Blue = "Blue"

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

# 显示帮助信息
function Show-Help {
    Write-Host "太上老君系统部署脚本" -ForegroundColor $Blue
    Write-Host ""
    Write-Host "使用方法:"
    Write-Host "  .\deploy.ps1 [环境] [操作]"
    Write-Host ""
    Write-Host "环境:"
    Write-Host "  dev      - 开发环境"
    Write-Host "  staging  - 预发布环境"
    Write-Host "  prod     - 生产环境 (默认)"
    Write-Host ""
    Write-Host "操作:"
    Write-Host "  build    - 构建镜像"
    Write-Host "  deploy   - 部署服务 (默认)"
    Write-Host "  restart  - 重启服务"
    Write-Host "  stop     - 停止服务"
    Write-Host "  logs     - 查看日志"
    Write-Host "  backup   - 备份数据"
    Write-Host "  health   - 健康检查"
    Write-Host "  cleanup  - 清理资源"
    Write-Host "  help     - 显示帮助"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\deploy.ps1 prod deploy    # 部署生产环境"
    Write-Host "  .\deploy.ps1 dev build      # 构建开发环境镜像"
    Write-Host "  .\deploy.ps1 prod logs      # 查看生产环境日志"
}

# 检查参数
if ($Environment -eq "help" -or $Action -eq "help") {
    Show-Help
    exit 0
}

# 检查部署目录
if (-not (Test-Path "deploy")) {
    Write-Error "错误: 找不到 deploy 目录"
    Write-Error "请确保在项目根目录运行此脚本"
    exit 1
}

# 切换到部署脚本目录
Set-Location "deploy/scripts"

# 调用实际的部署脚本
Write-Info "调用部署脚本: .\deploy.ps1 $Environment $Action"
& .\deploy.ps1 $Environment $Action