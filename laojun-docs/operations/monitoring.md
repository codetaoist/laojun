# Laojun 监控体系

## 概述

Laojun 监控体系是一个完整的可观测性解决方案，包含监控、日志、追踪和告警四个核心组件，为系统提供全方位的运行状态监控和问题诊断能力。

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Laojun 监控体系                          │
├─────────────────────────────────────────────────────────────────┤
│  应用层                                                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                │
│  │ Admin API   │ │Marketplace  │ │Config Center│                │
│  │             │ │    API      │ │             │                │
│  └─────────────┘ └─────────────┘ └─────────────┘                │
├─────────────────────────────────────────────────────────────────┤
│  监控层                                                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                │
│  │ Prometheus  │ │   Grafana   │ │AlertManager │                │
│  │   (指标)    │ │  (可视化)   │ │   (告警)    │                │
│  └─────────────┘ └─────────────┘ └─────────────┘                │
├─────────────────────────────────────────────────────────────────┤
│  日志层                                                         │
│  ┌─────────────┐ ┌─────────────┐                                │
│  │    Loki     │ │  Promtail   │                                │
│  │  (日志存储) │ │ (日志收集)  │                                │
│  └─────────────┘ └─────────────┘                                │
├─────────────────────────────────────────────────────────────────┤
│  追踪层                                                         │
│  ┌─────────────┐                                                │
│  │   Jaeger    │                                                │
│  │ (分布式追踪)│                                                │
│  └─────────────┘                                                │
├─────────────────────────────────────────────────────────────────┤
│  基础设施层                                                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                │
│  │ PostgreSQL  │ │    Redis    │ │    MinIO    │                │
│  │             │ │             │ │             │                │
│  └─────────────┘ └─────────────┘ └─────────────┘                │
└─────────────────────────────────────────────────────────────────┘
```

## 组件说明

### 监控组件

#### Prometheus
- **功能**: 时间序列数据库，负责收集和存储指标数据
- **端口**: 9090
- **配置**: `/etc/prometheus/prometheus.yml`
- **数据保留**: 200小时
- **特性**:
  - 自动服务发现
  - 多维度数据模型
  - 强大的查询语言 PromQL
  - 内置告警规则引擎

#### Grafana
- **功能**: 数据可视化和仪表板平台
- **端口**: 3000
- **默认账号**: admin/admin123
- **数据源**: Prometheus, Loki, Jaeger, PostgreSQL, Redis
- **仪表板**:
  - 系统监控 (Node Exporter)
  - 应用监控 (Laojun Services)
  - 业务指标 (Business Metrics)

#### AlertManager
- **功能**: 告警管理和通知
- **端口**: 9093
- **配置**: `/etc/alertmanager/config.yml`
- **通知方式**:
  - 邮件通知
  - 钉钉机器人 (可选)
  - 企业微信 (可选)
  - Webhook

### 日志组件

#### Loki
- **功能**: 日志聚合和存储
- **端口**: 3100
- **特性**:
  - 高效的日志压缩
  - 标签索引
  - 与 Grafana 深度集成
  - 支持 LogQL 查询语言

#### Promtail
- **功能**: 日志收集和转发
- **配置**: `/etc/promtail/config.yml`
- **收集范围**:
  - 系统日志
  - 容器日志
  - 应用日志
  - Nginx 日志
  - 数据库日志

### 追踪组件

#### Jaeger
- **功能**: 分布式追踪
- **端口**: 16686 (UI), 14268 (收集器)
- **特性**:
  - 请求链路追踪
  - 性能分析
  - 依赖关系图
  - 与监控指标关联

### 导出器

#### Node Exporter
- **功能**: 系统指标收集
- **端口**: 9100
- **指标**: CPU, 内存, 磁盘, 网络

#### cAdvisor
- **功能**: 容器指标收集
- **端口**: 8080
- **指标**: 容器资源使用情况

#### PostgreSQL Exporter
- **功能**: PostgreSQL 数据库指标
- **端口**: 9187

#### Redis Exporter
- **功能**: Redis 缓存指标
- **端口**: 9121

## 快速开始

### 启动监控体系

```powershell
# 启动完整监控体系
.\scripts\start-monitoring.ps1

# 启动特定服务
.\scripts\start-monitoring.ps1 -Services @("prometheus", "grafana")

# 跳过构建直接启动
.\scripts\start-monitoring.ps1 -SkipBuild

# 前台运行（用于调试）
.\scripts\start-monitoring.ps1 -Detached:$false
```

### 停止监控体系

```powershell
# 停止所有服务
.\scripts\stop-monitoring.ps1

# 停止并删除数据卷
.\scripts\stop-monitoring.ps1 -RemoveVolumes

# 停止并删除镜像
.\scripts\stop-monitoring.ps1 -RemoveImages

# 停止特定服务
.\scripts\stop-monitoring.ps1 -Services @("grafana", "prometheus")
```

### 访问地址

| 服务 | 地址 | 账号 | 说明 |
|------|------|------|------|
| Grafana | http://localhost:3000 | admin/admin123 | 监控仪表板 |
| Prometheus | http://localhost:9090 | - | 指标查询 |
| AlertManager | http://localhost:9093 | - | 告警管理 |
| Jaeger | http://localhost:16686 | - | 分布式追踪 |
| MinIO | http://localhost:9001 | minioadmin/minioadmin123 | 对象存储 |

## 监控指标

### 系统指标
- CPU 使用率
- 内存使用率
- 磁盘使用率
- 网络 I/O
- 文件描述符

### 应用指标
- HTTP 请求数量
- HTTP 响应时间
- HTTP 错误率
- Goroutine 数量
- 内存分配

### 业务指标
- 用户注册数量
- 插件下载数量
- 配置更新数量
- API 调用统计

### 基础设施指标
- 数据库连接数
- 数据库查询时间
- Redis 连接数
- Redis 命令执行时间
- 消息队列长度

## 告警规则

### 系统级告警
- 服务不可用
- CPU 使用率过高 (>80%)
- 内存使用率过高 (>85%)
- 磁盘空间不足 (<10%)
- HTTP 错误率过高 (>5%)

### 应用级告警
- 响应时间过长 (>2s)
- Goroutine 泄漏 (>1000)
- 数据库连接池耗尽
- Redis 连接失败

### 业务级告警
- 用户注册异常
- 插件下载失败率过高
- 配置更新失败

## 日志管理

### 日志级别
- **ERROR**: 错误信息，需要立即处理
- **WARN**: 警告信息，需要关注
- **INFO**: 一般信息，正常运行状态
- **DEBUG**: 调试信息，开发环境使用

### 日志格式
```json
{
  "time": "2024-01-01T12:00:00Z",
  "level": "info",
  "service": "admin-api",
  "request_id": "req-123456",
  "user_id": "user-789",
  "message": "User login successful",
  "duration": "150ms",
  "status": 200
}
```

### 日志查询
```logql
# 查询特定服务的错误日志
{job="laojun", service="admin-api"} |= "ERROR"

# 查询特定用户的操作日志
{job="laojun"} | json | user_id="user-123"

# 查询响应时间超过1秒的请求
{job="laojun"} | json | duration > 1s
```

## 性能优化

### Prometheus 优化
- 调整抓取间隔
- 配置数据保留策略
- 使用记录规则预计算
- 启用压缩存储

### Grafana 优化
- 使用变量减少查询
- 配置查询缓存
- 优化仪表板刷新间隔
- 使用模板变量

### Loki 优化
- 合理设置标签
- 配置日志保留策略
- 使用并行查询
- 启用压缩

## 故障排除

### 常见问题

#### Prometheus 无法抓取指标
1. 检查目标服务是否运行
2. 验证网络连接
3. 检查防火墙设置
4. 查看 Prometheus 日志

#### Grafana 仪表板无数据
1. 检查数据源配置
2. 验证查询语句
3. 检查时间范围
4. 查看 Grafana 日志

#### 告警不触发
1. 检查告警规则语法
2. 验证 AlertManager 配置
3. 检查通知渠道设置
4. 查看告警历史

#### 日志收集异常
1. 检查 Promtail 配置
2. 验证日志文件权限
3. 检查 Loki 连接
4. 查看 Promtail 日志

### 调试命令

```powershell
# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f prometheus
docker-compose logs -f grafana
docker-compose logs -f loki

# 检查配置文件
docker-compose config

# 重启特定服务
docker-compose restart prometheus

# 查看资源使用情况
docker stats
```

## 扩展配置

### 添加新的监控目标
1. 在应用中暴露 `/metrics` 端点
2. 更新 `prometheus.yml` 配置
3. 重启 Prometheus 服务
4. 在 Grafana 中创建仪表板

### 自定义告警规则
1. 编辑 `alerts.yml` 文件
2. 添加新的告警规则
3. 重新加载 Prometheus 配置
4. 在 AlertManager 中配置通知

### 集成外部系统
- **钉钉机器人**: 配置 Webhook URL
- **企业微信**: 配置 API 密钥
- **邮件通知**: 配置 SMTP 服务器
- **短信通知**: 集成短信网关

## 安全考虑

### 访问控制
- 配置 Grafana 用户权限
- 启用 Prometheus 认证
- 设置网络访问限制
- 使用 HTTPS 加密传输

### 数据保护
- 定期备份监控数据
- 配置数据加密
- 设置访问日志
- 监控异常访问

## 维护指南

### 定期维护
- 清理过期数据
- 更新组件版本
- 检查磁盘空间
- 备份配置文件

### 容量规划
- 监控数据增长趋势
- 评估存储需求
- 规划硬件升级
- 优化查询性能

## 参考资料

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana 官方文档](https://grafana.com/docs/)
- [Loki 官方文档](https://grafana.com/docs/loki/)
- [Jaeger 官方文档](https://www.jaegertracing.io/docs/)
- [AlertManager 官方文档](https://prometheus.io/docs/alerting/latest/alertmanager/)