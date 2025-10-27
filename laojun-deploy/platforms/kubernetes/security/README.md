# Kubernetes 安全策略配置

本目录包含太上老君微服务平台在 Kubernetes 环境下的安全策略配置文件。

## 📁 文件说明

### 🔐 核心安全配置

| 文件 | 用途 | 说明 |
|------|------|------|
| `rbac.yaml` | 基于角色的访问控制 | 定义服务账户、角色和权限绑定 |
| `network-policies.yaml` | 网络安全策略 | 控制 Pod 间的网络访问规则 |
| `pod-security-policy.yaml` | Pod 安全策略 | 定义 Pod 的安全标准和限制 |
| `secrets-management.yaml` | 密钥管理 | 管理敏感信息和配置 |
| `security-policies.yaml` | 综合安全策略 | 使用 Kyverno 等工具的安全策略 |

## 🚀 使用方法

### 应用所有安全策略

```bash
# 应用所有安全配置
kubectl apply -f .

# 或者分别应用
kubectl apply -f rbac.yaml
kubectl apply -f network-policies.yaml
kubectl apply -f pod-security-policy.yaml
kubectl apply -f secrets-management.yaml
kubectl apply -f security-policies.yaml
```

### 验证安全策略

```bash
# 检查 RBAC 配置
kubectl get serviceaccounts,roles,rolebindings -n taishanglaojun

# 检查网络策略
kubectl get networkpolicies -n taishanglaojun

# 检查 Pod 安全策略
kubectl get podsecuritypolicies

# 检查密钥
kubectl get secrets -n taishanglaojun
```

## 🔒 安全特性

### 网络安全
- **默认拒绝策略**: 默认拒绝所有入站和出站流量
- **最小权限原则**: 只允许必要的网络连接
- **服务隔离**: 不同服务间的网络隔离

### 访问控制
- **服务账户隔离**: 每个服务使用独立的服务账户
- **最小权限**: 只授予必要的 Kubernetes API 权限
- **命名空间隔离**: 严格的命名空间边界

### Pod 安全
- **非特权运行**: 禁止特权容器
- **只读根文件系统**: 增强容器安全性
- **安全上下文**: 强制安全配置

### 密钥管理
- **加密存储**: 所有敏感信息加密存储
- **访问控制**: 严格的密钥访问权限
- **轮换策略**: 定期轮换敏感凭据

## ⚠️ 注意事项

### 部署前准备

1. **确保 Kubernetes 版本兼容性**
   ```bash
   kubectl version --short
   ```

2. **检查必要的准入控制器**
   ```bash
   kubectl get validatingadmissionwebhooks
   kubectl get mutatingadmissionwebhooks
   ```

3. **安装 Kyverno（如果使用 security-policies.yaml）**
   ```bash
   kubectl create -f https://github.com/kyverno/kyverno/releases/latest/download/install.yaml
   ```

### 配置自定义

1. **更新密钥值**: 在部署前设置 `secrets-management.yaml` 中的实际密钥值
2. **调整网络策略**: 根据实际网络拓扑调整网络策略
3. **自定义 RBAC**: 根据服务需求调整权限配置

### 故障排除

```bash
# 检查 Pod 是否因安全策略被拒绝
kubectl get events --sort-by=.metadata.creationTimestamp

# 检查网络策略是否阻止连接
kubectl describe networkpolicy -n taishanglaojun

# 检查 RBAC 权限问题
kubectl auth can-i <verb> <resource> --as=system:serviceaccount:taishanglaojun:<service-account>
```

## 📚 相关文档

- [Kubernetes 安全最佳实践](https://kubernetes.io/docs/concepts/security/)
- [网络策略文档](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [RBAC 授权](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Pod 安全策略](https://kubernetes.io/docs/concepts/policy/pod-security-policy/)

---

**⚠️ 重要提醒**: 这些安全策略会严格限制集群中的操作。在生产环境部署前，请在测试环境中充分验证。