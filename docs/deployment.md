# 部署指南

本文档详细介绍了 Laojun 项目的各种部署方式，包括本地开发、Docker 容器化部署和 Kubernetes 集群部署。

## 目录

- [环境要求](#环境要求)
- [本地部署](#本地部署)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [配置管理](#配置管理)
- [监控和日志](#监控和日志)
- [安全配置](#安全配置)
- [故障排除](#故障排除)

## 环境要求

### 基础要求

- **操作系统**: Linux, macOS, Windows
- **Go 版本**: 1.21 或更高
- **数据库**: PostgreSQL 13+ 或 MySQL 8.0+
- **缓存**: Redis 6.0+
- **内存**: 最小 2GB，推荐 4GB+
- **存储**: 最小 10GB 可用空间

### 可选组件

- **Docker**: 20.10+ (容器化部署)
- **Kubernetes**: 1.20+ (集群部署)
- **Nginx**: 1.18+ (反向代理)
- **Prometheus**: 2.30+ (监控)
- **Grafana**: 8.0+ (可视化)

## 本地部署

### 1. 准备环境

```bash
# 安装 Go
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装 PostgreSQL
sudo apt-get update
sudo apt-get install postgresql postgresql-contrib

# 安装 Redis
sudo apt-get install redis-server
```

### 2. 克隆项目

```bash
git clone https://github.com/codetaoist/laojun.git
cd laojun
```

### 3. 配置数据库

```bash
# 创建数据库用户和数据库
sudo -u postgres psql
CREATE USER laojun WITH PASSWORD 'laojun123';
CREATE DATABASE laojun OWNER laojun;
GRANT ALL PRIVILEGES ON DATABASE laojun TO laojun;
\q
```

### 4. 配置应用

```bash
# 复制配置文件
cp configs/config.example.yaml configs/config.yaml

# 编辑配置文件
vim configs/config.yaml
```

配置示例：

```yaml
app:
  name: "Laojun"
  version: "1.0.0"
  environment: "development"
  debug: true

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  username: "laojun"
  password: "laojun123"
  database: "laojun"
  ssl_mode: "disable"
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0
  pool_size: 10
```

### 5. 安装依赖

```bash
go mod download
```

### 6. 运行数据库迁移

```bash
go run cmd/migrate/main.go
```

### 7. 启动服务

```bash
# 开发模式
go run cmd/api/main.go

# 或使用构建脚本
.\scripts\build.ps1 -Target api -Environment dev
.\build\laojun-api.exe
```

### 8. 验证部署

```bash
# 健康检查
curl http://localhost:8080/health

# API 文档
open http://localhost:8080/swagger/
```

## Docker 部署

### 1. 构建镜像

```bash
# 构建应用镜像
docker build -t laojun:latest .

# 或使用构建脚本
.\scripts\build.ps1 -Target docker -Environment prod
```

### 2. 使用 Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: laojun
      POSTGRES_USER: laojun
      POSTGRES_PASSWORD: laojun123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U laojun"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  laojun-api:
    image: laojun:latest
    ports:
      - "8080:8080"
    environment:
      - LAOJUN_ENVIRONMENT=production
      - LAOJUN_DATABASE_HOST=postgres
      - LAOJUN_REDIS_HOST=redis
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - laojun-api
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f laojun-api
```

### 4. 使用部署脚本

```bash
# Docker 部署
.\scripts\deploy.ps1 -Environment prod -Target docker -Monitor

# 查看部署状态
.\scripts\deploy.ps1 --status -Target docker
```

## Kubernetes 部署

### 1. 准备集群

```bash
# 检查集群连接
kubectl cluster-info

# 创建命名空间
kubectl create namespace laojun
```

### 2. 配置存储

```yaml
# k8s/storage.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: postgres-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /data/postgres
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - node1
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: laojun
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: local-storage
```

### 3. 部署数据库

```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: laojun
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:13
        env:
        - name: POSTGRES_DB
          value: "laojun"
        - name: POSTGRES_USER
          value: "laojun"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: laojun
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
  type: ClusterIP
```

### 4. 部署应用

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: laojun-api
  namespace: laojun
spec:
  replicas: 3
  selector:
    matchLabels:
      app: laojun-api
  template:
    metadata:
      labels:
        app: laojun-api
    spec:
      containers:
      - name: laojun-api
        image: laojun:latest
        ports:
        - containerPort: 8080
        env:
        - name: LAOJUN_ENVIRONMENT
          value: "production"
        - name: LAOJUN_DATABASE_HOST
          value: "postgres-service"
        - name: LAOJUN_REDIS_HOST
          value: "redis-service"
        volumeMounts:
        - name: config-volume
          mountPath: /app/configs
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config-volume
        configMap:
          name: laojun-config
---
apiVersion: v1
kind: Service
metadata:
  name: laojun-api-service
  namespace: laojun
spec:
  selector:
    app: laojun-api
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 5. 配置 Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: laojun-ingress
  namespace: laojun
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - api.laojun.dev
    secretName: laojun-tls
  rules:
  - host: api.laojun.dev
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: laojun-api-service
            port:
              number: 80
```

### 6. 使用 Kustomize 部署

```bash
# 应用所有配置
kubectl apply -k k8s/

# 或使用部署脚本
.\scripts\deploy.ps1 -Environment prod -Target k8s -Version v1.0.0 -Monitor
```

### 7. 验证部署

```bash
# 检查 Pod 状态
kubectl get pods -n laojun

# 检查服务状态
kubectl get services -n laojun

# 查看日志
kubectl logs -f deployment/laojun-api -n laojun

# 端口转发测试
kubectl port-forward service/laojun-api-service 8080:80 -n laojun
```

## 配置管理

### 环境变量

支持通过环境变量覆盖配置：

```bash
export LAOJUN_ENVIRONMENT=production
export LAOJUN_DATABASE_HOST=postgres.example.com
export LAOJUN_DATABASE_PASSWORD=secret
export LAOJUN_REDIS_HOST=redis.example.com
```

### ConfigMap (Kubernetes)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: laojun-config
  namespace: laojun
data:
  config.yaml: |
    app:
      name: "Laojun"
      environment: "production"
    server:
      port: 8080
    database:
      host: "postgres-service"
      port: 5432
```

### Secret (Kubernetes)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: laojun-secret
  namespace: laojun
type: Opaque
data:
  database-password: bGFvanVuMTIz  # base64 encoded
  jwt-secret: c2VjcmV0a2V5MTIz      # base64 encoded
```

## 监控和日志

### Prometheus 监控

```yaml
# k8s/monitoring.yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: laojun-metrics
  namespace: laojun
spec:
  selector:
    matchLabels:
      app: laojun-api
  endpoints:
  - port: metrics
    path: /metrics
```

### 日志收集

```yaml
# k8s/logging.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: laojun
data:
  fluent-bit.conf: |
    [INPUT]
        Name tail
        Path /var/log/containers/*laojun*.log
        Parser docker
        Tag kube.*
    
    [OUTPUT]
        Name elasticsearch
        Match *
        Host elasticsearch.logging.svc.cluster.local
        Port 9200
        Index laojun-logs
```

### Grafana 仪表板

导入预配置的仪表板：

```bash
# 导入仪表板
kubectl apply -f k8s/grafana-dashboard.yaml
```

## 安全配置

### TLS/SSL 配置

```yaml
# k8s/tls-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: laojun-tls
  namespace: laojun
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTi...  # base64 encoded certificate
  tls.key: LS0tLS1CRUdJTi...  # base64 encoded private key
```

### 网络策略

```yaml
# k8s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: laojun-network-policy
  namespace: laojun
spec:
  podSelector:
    matchLabels:
      app: laojun-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
```

### Pod 安全策略

```yaml
# k8s/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: laojun-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

## 故障排除

### 常见问题

#### 1. 数据库连接失败

```bash
# 检查数据库服务状态
kubectl get pods -l app=postgres -n laojun

# 查看数据库日志
kubectl logs -l app=postgres -n laojun

# 测试数据库连接
kubectl exec -it deployment/laojun-api -n laojun -- \
  psql -h postgres-service -U laojun -d laojun
```

#### 2. 应用启动失败

```bash
# 查看 Pod 状态
kubectl describe pod <pod-name> -n laojun

# 查看应用日志
kubectl logs <pod-name> -n laojun

# 检查配置
kubectl get configmap laojun-config -n laojun -o yaml
```

#### 3. 服务无法访问

```bash
# 检查服务状态
kubectl get services -n laojun

# 检查 Ingress 配置
kubectl get ingress -n laojun

# 测试服务连通性
kubectl exec -it deployment/laojun-api -n laojun -- \
  curl http://laojun-api-service/health
```

### 性能调优

#### 1. 资源限制调整

```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

#### 2. 数据库连接池优化

```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h
```

#### 3. 缓存配置优化

```yaml
redis:
  pool_size: 20
  min_idle_conns: 5
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
```

### 备份和恢复

#### 数据库备份

```bash
# 创建备份
kubectl exec -it deployment/postgres -n laojun -- \
  pg_dump -U laojun laojun > backup.sql

# 恢复备份
kubectl exec -i deployment/postgres -n laojun -- \
  psql -U laojun laojun < backup.sql
```

#### 配置备份

```bash
# 备份所有配置
kubectl get all,configmap,secret -n laojun -o yaml > laojun-backup.yaml

# 恢复配置
kubectl apply -f laojun-backup.yaml
```

## 升级和回滚

### 滚动升级

```bash
# 更新镜像
kubectl set image deployment/laojun-api laojun-api=laojun:v1.1.0 -n laojun

# 查看升级状态
kubectl rollout status deployment/laojun-api -n laojun
```

### 回滚

```bash
# 查看历史版本
kubectl rollout history deployment/laojun-api -n laojun

# 回滚到上一版本
kubectl rollout undo deployment/laojun-api -n laojun

# 回滚到指定版本
kubectl rollout undo deployment/laojun-api --to-revision=2 -n laojun
```

---

更多部署相关问题，请参考 [故障排除文档](troubleshooting.md) 或提交 [Issue](https://github.com/codetaoist/laojun/issues)。