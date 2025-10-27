# 太上老君微服务平台 Kubernetes 部署

本目录包含太上老君微服务平台在 Kubernetes 环境中的完整部署配置。

## 📋 目录结构

```
k8s/
├── namespace.yaml      # 命名空间配置
├── rbac.yaml          # RBAC 权限配置
├── configmaps.yaml    # 配置文件映射
├── services.yaml      # 服务配置
├── deployments.yaml   # 部署配置
├── ingress.yaml       # 入口配置和网络策略
├── deploy.sh          # Linux/macOS 部署脚本
├── deploy.ps1         # Windows PowerShell 部署脚本
└── README.md          # 本文档
```

## 🚀 快速开始

### 前提条件

1. **Kubernetes 集群**: 版本 1.20+
2. **kubectl**: 已配置并能连接到集群
3. **Ingress Controller**: 推荐使用 NGINX Ingress Controller
4. **存储**: 支持动态存储卷分配（可选）

### 安装 NGINX Ingress Controller

```bash
# 使用 Helm 安装
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace

# 或使用 kubectl 安装
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml
```

### 部署平台

#### Linux/macOS

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

#### Windows PowerShell

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
# 1. 创建命名空间
kubectl apply -f namespace.yaml

# 2. 创建 RBAC 配置
kubectl apply -f rbac.yaml

# 3. 创建配置映射
kubectl apply -f configmaps.yaml

# 4. 创建服务
kubectl apply -f services.yaml

# 5. 创建部署
kubectl apply -f deployments.yaml

# 6. 创建入口
kubectl apply -f ingress.yaml
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

| 服务 | 地址 | 说明 |
|------|------|------|
| 服务发现 API | http://api.taishanglaojun.local/discovery | 服务注册与发现 |
| 监控 API | http://api.taishanglaojun.local/monitoring | 监控数据接口 |
| Prometheus | http://prometheus.taishanglaojun.local | 指标收集 |
| Grafana | http://grafana.taishanglaojun.local | 可视化面板 |
| Jaeger | http://jaeger.taishanglaojun.local | 链路追踪 |
| Kibana | http://kibana.taishanglaojun.local | 日志分析 |
| 监控面板 | http://monitoring.taishanglaojun.local | 统一监控入口 |

### 默认凭据

| 服务 | 用户名 | 密码 |
|------|--------|------|
| Grafana | admin | admin123 |
| 监控面板 | admin | admin123 |

## 🏗️ 架构说明

### 命名空间

- `taishanglaojun`: 主要业务服务
- `taishanglaojun-monitoring`: 监控相关服务

### 核心组件

#### 业务服务 (taishanglaojun 命名空间)

- **laojun-discovery**: 服务发现组件
- **laojun-monitoring**: 监控服务组件
- **redis**: 缓存服务
- **consul**: 服务注册中心

#### 监控服务 (taishanglaojun-monitoring 命名空间)

- **prometheus**: 指标收集和存储
- **grafana**: 监控数据可视化
- **jaeger**: 分布式链路追踪
- **influxdb**: 时序数据库
- **elasticsearch**: 搜索引擎
- **kibana**: 日志分析和可视化

### 网络策略

配置了网络策略来限制 Pod 间的通信：

- 主命名空间的服务可以访问监控命名空间
- 监控命名空间可以抓取主命名空间的指标
- 限制了不必要的跨命名空间通信

## 🔧 配置说明

### 资源配置

| 组件 | CPU 请求 | CPU 限制 | 内存请求 | 内存限制 |
|------|----------|----------|----------|----------|
| laojun-discovery | 100m | 200m | 128Mi | 256Mi |
| laojun-monitoring | 200m | 500m | 256Mi | 512Mi |
| redis | 100m | 200m | 128Mi | 256Mi |
| consul | 200m | 500m | 256Mi | 512Mi |
| prometheus | 500m | 1000m | 512Mi | 1Gi |
| grafana | 200m | 500m | 256Mi | 512Mi |

### 健康检查

所有服务都配置了 liveness 和 readiness 探针：

- **Liveness Probe**: 检测容器是否存活
- **Readiness Probe**: 检测容器是否准备好接收流量

### 持久化存储

当前配置使用 `emptyDir` 卷，数据在 Pod 重启时会丢失。生产环境建议：

1. 使用 PersistentVolume 和 PersistentVolumeClaim
2. 配置适当的存储类
3. 设置备份策略

## 📊 监控和观测

### Prometheus 指标

- 自动发现 Kubernetes 服务
- 收集节点和 Pod 指标
- 支持自定义指标

### Grafana 仪表板

- 系统资源监控
- 应用性能监控
- Kubernetes 集群监控

### Jaeger 链路追踪

- 分布式请求追踪
- 性能分析
- 依赖关系图

### 日志聚合

- Elasticsearch 存储日志
- Kibana 分析和可视化
- 结构化日志搜索

## 🔒 安全配置

### RBAC

- 为每个服务配置了最小权限
- 使用 ServiceAccount 进行身份验证
- 限制了 API 访问权限

### 网络安全

- 配置了 NetworkPolicy 限制流量
- 使用 TLS 加密内部通信（可选）
- 监控服务需要基本认证

### 密钥管理

- 使用 Kubernetes Secret 存储敏感信息
- 避免在配置文件中硬编码密码
- 支持外部密钥管理系统集成

## 🚨 故障排除

### 常见问题

1. **Pod 无法启动**
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   kubectl logs <pod-name> -n <namespace>
   ```

2. **服务无法访问**
   ```bash
   kubectl get svc -n <namespace>
   kubectl describe svc <service-name> -n <namespace>
   ```

3. **Ingress 不工作**
   ```bash
   kubectl get ingress -n <namespace>
   kubectl describe ingress <ingress-name> -n <namespace>
   ```

### 日志查看

```bash
# 查看所有 Pod 日志
kubectl logs -l app.kubernetes.io/part-of=taishanglaojun-platform --all-containers=true

# 查看特定服务日志
kubectl logs deployment/laojun-discovery -n taishanglaojun

# 实时查看日志
kubectl logs -f deployment/laojun-monitoring -n taishanglaojun
```

### 性能调优

1. **资源调整**: 根据实际负载调整 CPU 和内存限制
2. **副本数量**: 根据流量调整 Deployment 副本数
3. **存储优化**: 使用高性能存储类
4. **网络优化**: 配置适当的网络插件

## 🔄 升级和维护

### 滚动更新

```bash
# 更新镜像
kubectl set image deployment/laojun-discovery laojun-discovery=taishanglaojun/laojun-discovery:v2.0.0 -n taishanglaojun

# 查看更新状态
kubectl rollout status deployment/laojun-discovery -n taishanglaojun

# 回滚更新
kubectl rollout undo deployment/laojun-discovery -n taishanglaojun
```

### 备份和恢复

```bash
# 备份配置
kubectl get all,configmap,secret -n taishanglaojun -o yaml > backup.yaml

# 恢复配置
kubectl apply -f backup.yaml
```

### 监控告警

配置 Prometheus AlertManager 规则：

- CPU 使用率过高
- 内存使用率过高
- Pod 重启频繁
- 服务不可用

## 📚 参考文档

- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [Prometheus Operator](https://prometheus-operator.dev/)
- [Grafana 文档](https://grafana.com/docs/)
- [Jaeger 文档](https://www.jaegertracing.io/docs/)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进部署配置。

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../LICENSE) 文件。