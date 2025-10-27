# Kubernetes 部署指南

本文档详细说明如何在 Kubernetes 集群中部署太上老君系统。

## 环境要求

### 集群要求
- **Kubernetes**: 1.20+ 
- **节点数量**: 最少 3 个节点
- **内存**: 每节点最少 4GB
- **存储**: 支持 PersistentVolume
- **网络**: CNI 网络插件

### 工具要求
- kubectl 客户端
- Helm 3.0+ (可选)
- Kustomize (可选)

## 快速开始

### 1. 准备集群

```bash
# 检查集群连接
kubectl cluster-info

# 创建命名空间
kubectl create namespace laojun

# 设置默认命名空间
kubectl config set-context --current --namespace=laojun
```

### 2. 创建配置和密钥

```bash
# 创建数据库密钥
kubectl create secret generic postgres-secret \
  --from-literal=password=your_secure_password \
  -n laojun

# 创建应用配置
kubectl create configmap laojun-config \
  --from-literal=DB_HOST=postgres-service \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_NAME=laojun \
  --from-literal=REDIS_HOST=redis-service \
  --from-literal=REDIS_PORT=6379 \
  -n laojun
```

### 3. 部署存储

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

### 4. 部署数据库

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
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - laojun
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - laojun
          initialDelaySeconds: 5
          periodSeconds: 5
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

### 5. 部署 Redis

```yaml
# k8s/redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: laojun
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:6-alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: laojun
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
  type: ClusterIP
```

### 6. 部署应用

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
        - name: DB_USER
          value: "laojun"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        envFrom:
        - configMapRef:
            name: laojun-config
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
  type: ClusterIP
```

### 7. 配置 Ingress

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

## 使用 Kustomize 部署

### 1. 创建 Kustomization

```yaml
# k8s/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: laojun

resources:
- storage.yaml
- postgres.yaml
- redis.yaml
- deployment.yaml
- ingress.yaml

configMapGenerator:
- name: laojun-config
  literals:
  - DB_HOST=postgres-service
  - DB_PORT=5432
  - DB_NAME=laojun
  - REDIS_HOST=redis-service
  - REDIS_PORT=6379

secretGenerator:
- name: postgres-secret
  literals:
  - password=your_secure_password

images:
- name: laojun
  newTag: v1.0.0
```

### 2. 执行部署

```bash
# 使用 Kustomize 部署
kubectl apply -k k8s/

# 或使用部署脚本
./scripts/deploy-k8s.sh
```

## 监控和日志

### 1. 部署监控

```yaml
# k8s/monitoring.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: laojun
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'laojun-api'
      static_configs:
      - targets: ['laojun-api-service:80']
      metrics_path: /metrics
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: laojun
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        args:
        - '--config.file=/etc/prometheus/prometheus.yml'
        - '--storage.tsdb.path=/prometheus'
        - '--web.console.libraries=/etc/prometheus/console_libraries'
        - '--web.console.templates=/etc/prometheus/consoles'
      volumes:
      - name: config
        configMap:
          name: prometheus-config
```

### 2. 日志收集

```yaml
# k8s/logging.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: laojun
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
    
    [INPUT]
        Name              tail
        Path              /var/log/containers/*laojun*.log
        Parser            docker
        Tag               kube.*
        Refresh_Interval  5
    
    [OUTPUT]
        Name  stdout
        Match *
```

## 运维操作

### 1. 扩缩容

```bash
# 扩容应用
kubectl scale deployment laojun-api --replicas=5 -n laojun

# 查看扩容状态
kubectl get pods -n laojun -l app=laojun-api
```

### 2. 滚动更新

```bash
# 更新镜像
kubectl set image deployment/laojun-api laojun-api=laojun:v1.1.0 -n laojun

# 查看更新状态
kubectl rollout status deployment/laojun-api -n laojun

# 查看更新历史
kubectl rollout history deployment/laojun-api -n laojun

# 回滚到上一版本
kubectl rollout undo deployment/laojun-api -n laojun

# 回滚到指定版本
kubectl rollout undo deployment/laojun-api --to-revision=2 -n laojun
```

### 3. 故障排除

```bash
# 查看 Pod 状态
kubectl get pods -n laojun

# 查看 Pod 日志
kubectl logs -f deployment/laojun-api -n laojun

# 进入 Pod 调试
kubectl exec -it deployment/laojun-api -n laojun -- /bin/sh

# 查看事件
kubectl get events -n laojun --sort-by=.metadata.creationTimestamp

# 描述资源详情
kubectl describe pod <pod-name> -n laojun
```

### 4. 数据备份

```bash
# 备份数据库
kubectl exec -it deployment/postgres -n laojun -- \
  pg_dump -U laojun laojun > backup.sql

# 恢复数据库
kubectl exec -i deployment/postgres -n laojun -- \
  psql -U laojun -d laojun < backup.sql
```

## 高可用配置

### 1. 多副本部署

```yaml
# 确保关键组件多副本
spec:
  replicas: 3  # API 服务
  
# 配置反亲和性
spec:
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - laojun-api
              topologyKey: kubernetes.io/hostname
```

### 2. 资源限制

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### 3. 健康检查

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

## 安全配置

### 1. RBAC 配置

```yaml
# k8s/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: laojun-sa
  namespace: laojun
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: laojun-role
  namespace: laojun
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: laojun-rolebinding
  namespace: laojun
subjects:
- kind: ServiceAccount
  name: laojun-sa
  namespace: laojun
roleRef:
  kind: Role
  name: laojun-role
  apiGroup: rbac.authorization.k8s.io
```

### 2. 网络策略

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
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

## 性能优化

### 1. 资源调优

- **CPU**: 根据负载调整 requests 和 limits
- **内存**: 监控内存使用，避免 OOM
- **存储**: 使用 SSD 存储提升性能

### 2. 网络优化

- **Service Mesh**: 考虑使用 Istio 进行流量管理
- **负载均衡**: 配置合适的负载均衡策略
- **缓存**: 合理使用 Redis 缓存

### 3. 监控告警

- **Prometheus**: 收集应用指标
- **Grafana**: 可视化监控面板
- **AlertManager**: 配置告警规则

## 故障恢复

### 1. 备份策略

- **数据库**: 定期备份 PostgreSQL
- **配置**: 备份 Kubernetes 配置文件
- **镜像**: 保留历史版本镜像

### 2. 灾难恢复

- **多区域部署**: 跨可用区部署
- **数据同步**: 配置数据库主从复制
- **自动故障转移**: 配置自动故障转移机制

更多 Kubernetes 部署问题，请参考官方文档或提交 Issue。