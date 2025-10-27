# 太上老君微服务平台 Helm Charts

本目录包含太上老君微服务平台的 Helm Charts，提供了简化的 Kubernetes 部署方案。

## 📋 目录结构

```
helm/
├── taishanglaojun/                    # 主 Chart
│   ├── Chart.yaml                     # Chart 元数据
│   ├── values.yaml                    # 默认配置值
│   ├── charts/                        # 子 Charts
│   │   ├── taishanglaojun-platform/   # 平台服务 Chart
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   └── templates/             # 模板文件
│   │   │       ├── _helpers.tpl
│   │   │       ├── namespace.yaml
│   │   │       ├── discovery-deployment.yaml
│   │   │       ├── discovery-service.yaml
│   │   │       ├── monitoring-deployment.yaml
│   │   │       └── monitoring-service.yaml
│   │   └── taishanglaojun-monitoring/ # 监控栈 Chart
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       └── templates/
│   │           ├── _helpers.tpl
│   │           ├── namespace.yaml
│   │           ├── prometheus-deployment.yaml
│   │           ├── grafana-deployment.yaml
│   │           └── jaeger-deployment.yaml
│   └── templates/                     # 主 Chart 模板
├── deploy.sh                         # Linux/macOS 部署脚本
├── deploy.ps1                        # Windows PowerShell 部署脚本
└── README.md                         # 本文档
```

## 🚀 快速开始

### 前提条件

1. **Kubernetes 集群**: 版本 1.20+
2. **Helm**: 版本 3.0+
3. **kubectl**: 已配置并能连接到集群
4. **Ingress Controller**: 推荐使用 NGINX Ingress Controller

### 安装 Helm

#### Linux/macOS
```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

#### Windows
```powershell
# 使用 Chocolatey
choco install kubernetes-helm

# 或使用 Scoop
scoop install helm
```

### 部署平台

#### 使用部署脚本 (推荐)

**Linux/macOS:**
```bash
# 给脚本执行权限
chmod +x deploy.sh

# 安装平台
./deploy.sh install

# 查看状态
./deploy.sh status

# 升级平台
./deploy.sh upgrade

# 卸载平台
./deploy.sh uninstall
```

**Windows PowerShell:**
```powershell
# 安装平台
.\deploy.ps1 install

# 查看状态
.\deploy.ps1 status

# 升级平台
.\deploy.ps1 upgrade

# 卸载平台
.\deploy.ps1 uninstall
```

#### 手动部署

```bash
# 1. 添加依赖并更新
helm dependency update ./taishanglaojun

# 2. 安装 Chart
helm install taishanglaojun ./taishanglaojun \
  --namespace taishanglaojun \
  --create-namespace \
  --wait

# 3. 查看状态
helm status taishanglaojun -n taishanglaojun
```

## ⚙️ 配置说明

### 主要配置参数

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `global.domain` | 全局域名 | `taishanglaojun.local` |
| `global.imageRegistry` | 镜像仓库地址 | `""` |
| `global.storageClass` | 存储类 | `""` |
| `platform.enabled` | 启用平台服务 | `true` |
| `monitoring.enabled` | 启用监控栈 | `true` |

### 平台服务配置

#### 服务发现组件
```yaml
platform:
  discovery:
    enabled: true
    replicaCount: 2
    image:
      repository: taishanglaojun/laojun-discovery
      tag: "1.0.0"
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 200m
        memory: 256Mi
    autoscaling:
      enabled: false
      minReplicas: 2
      maxReplicas: 10
```

#### 监控组件
```yaml
platform:
  monitoring:
    enabled: true
    replicaCount: 2
    image:
      repository: taishanglaojun/laojun-monitoring
      tag: "1.0.0"
    resources:
      requests:
        cpu: 200m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

### 监控栈配置

#### Prometheus
```yaml
monitoring:
  prometheus:
    enabled: true
    persistence:
      enabled: true
      size: 50Gi
    retention: "30d"
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

#### Grafana
```yaml
monitoring:
  grafana:
    enabled: true
    admin:
      username: admin
      password: admin123
    persistence:
      enabled: true
      size: 10Gi
```

### 自定义配置

创建自定义的 values 文件：

```yaml
# custom-values.yaml
global:
  domain: "mycompany.com"
  imageRegistry: "registry.mycompany.com"

platform:
  discovery:
    replicaCount: 3
    resources:
      requests:
        cpu: 200m
        memory: 256Mi

monitoring:
  prometheus:
    persistence:
      size: 100Gi
    retention: "90d"
```

使用自定义配置部署：
```bash
helm install taishanglaojun ./taishanglaojun \
  --namespace taishanglaojun \
  --create-namespace \
  --values custom-values.yaml
```

## 🌐 访问配置

### 配置 Hosts 文件

在本地 hosts 文件中添加以下条目：

**Linux/macOS**: `/etc/hosts`
**Windows**: `C:\Windows\System32\drivers\etc\hosts`

```
127.0.0.1 taishanglaojun.local
127.0.0.1 api.taishanglaojun.local
127.0.0.1 monitoring.taishanglaojun.local
127.0.0.1 prometheus.taishanglaojun.local
127.0.0.1 grafana.taishanglaojun.local
127.0.0.1 jaeger.taishanglaojun.local
127.0.0.1 kibana.taishanglaojun.local
```

### 服务访问地址

| 服务 | 地址 | 凭据 |
|------|------|------|
| 服务发现 API | http://api.taishanglaojun.local/discovery | - |
| 监控 API | http://api.taishanglaojun.local/monitoring | - |
| Prometheus | http://prometheus.taishanglaojun.local | admin/admin123 |
| Grafana | http://grafana.taishanglaojun.local | admin/admin123 |
| Jaeger | http://jaeger.taishanglaojun.local | - |
| Kibana | http://kibana.taishanglaojun.local | - |

## 🔧 管理操作

### 查看日志

```bash
# 使用部署脚本
./deploy.sh logs discovery
./deploy.sh logs monitoring
./deploy.sh logs prometheus

# 使用 kubectl
kubectl logs -f deployment/taishanglaojun-discovery -n taishanglaojun
kubectl logs -f deployment/taishanglaojun-prometheus -n taishanglaojun-monitoring
```

### 端口转发

```bash
# 使用部署脚本
./deploy.sh port-forward discovery 8081
./deploy.sh port-forward prometheus 9090

# 使用 kubectl
kubectl port-forward svc/taishanglaojun-discovery 8081:8081 -n taishanglaojun
kubectl port-forward svc/taishanglaojun-prometheus 9090:9090 -n taishanglaojun-monitoring
```

### 扩缩容

```bash
# 扩展服务发现组件
kubectl scale deployment taishanglaojun-discovery --replicas=3 -n taishanglaojun

# 或使用 Helm 升级
helm upgrade taishanglaojun ./taishanglaojun \
  --namespace taishanglaojun \
  --set platform.discovery.replicaCount=3
```

### 更新镜像

```bash
# 更新服务发现组件镜像
helm upgrade taishanglaojun ./taishanglaojun \
  --namespace taishanglaojun \
  --set platform.discovery.image.tag=1.1.0
```

## 🔄 升级和回滚

### 升级

```bash
# 更新依赖
helm dependency update ./taishanglaojun

# 升级 Release
helm upgrade taishanglaojun ./taishanglaojun \
  --namespace taishanglaojun \
  --wait
```

### 回滚

```bash
# 查看历史版本
helm history taishanglaojun -n taishanglaojun

# 回滚到上一个版本
helm rollback taishanglaojun -n taishanglaojun

# 回滚到指定版本
helm rollback taishanglaojun 1 -n taishanglaojun
```

## 🚨 故障排除

### 常见问题

1. **Pod 无法启动**
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   kubectl logs <pod-name> -n <namespace>
   ```

2. **Helm 安装失败**
   ```bash
   helm install taishanglaojun ./taishanglaojun --debug --dry-run
   ```

3. **依赖更新失败**
   ```bash
   helm dependency update ./taishanglaojun --debug
   ```

4. **Ingress 不工作**
   ```bash
   kubectl get ingress -n taishanglaojun
   kubectl describe ingress <ingress-name> -n taishanglaojun
   ```

### 调试命令

```bash
# 验证 Chart 语法
helm lint ./taishanglaojun

# 渲染模板但不安装
helm template taishanglaojun ./taishanglaojun

# 调试安装过程
helm install taishanglaojun ./taishanglaojun --debug --dry-run

# 查看 Release 详情
helm get all taishanglaojun -n taishanglaojun
```

## 📊 监控和观测

### Prometheus 指标

Chart 自动配置了以下监控指标：

- 应用程序指标 (通过 `/metrics` 端点)
- Kubernetes 集群指标
- 节点和 Pod 资源指标
- 自定义业务指标

### Grafana 仪表板

预配置的仪表板包括：

- 系统资源监控
- 应用性能监控
- Kubernetes 集群监控
- 业务指标监控

### 告警规则

内置告警规则涵盖：

- 系统资源告警 (CPU、内存、磁盘)
- 应用程序告警 (错误率、响应时间)
- Kubernetes 告警 (Pod 重启、节点状态)

## 🔒 安全配置

### RBAC

Chart 自动创建必要的 RBAC 资源：

- ServiceAccount
- ClusterRole
- ClusterRoleBinding

### 网络策略

配置了网络策略来限制 Pod 间通信：

- 平台服务之间的通信
- 监控组件的数据收集
- 外部访问控制

### 密钥管理

支持多种密钥管理方式：

- Kubernetes Secret
- 外部密钥管理系统 (如 Vault)
- 环境变量注入

## 🏗️ 自定义开发

### 添加新组件

1. 在相应的子 Chart 中添加模板文件
2. 更新 `values.yaml` 配置
3. 添加必要的依赖关系
4. 更新文档

### 创建自定义 Chart

```bash
# 创建新的子 Chart
helm create charts/my-component

# 添加到主 Chart 依赖
# 在 Chart.yaml 中添加:
dependencies:
  - name: my-component
    version: "1.0.0"
    repository: "file://./charts/my-component"
```

## 📚 参考文档

- [Helm 官方文档](https://helm.sh/docs/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [Chart 最佳实践](https://helm.sh/docs/chart_best_practices/)
- [模板开发指南](https://helm.sh/docs/chart_template_guide/)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进 Helm Charts。

### 开发流程

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

### 测试

```bash
# 语法检查
helm lint ./taishanglaojun

# 模板渲染测试
helm template taishanglaojun ./taishanglaojun

# 安装测试
helm install test-release ./taishanglaojun --dry-run
```

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../LICENSE) 文件。