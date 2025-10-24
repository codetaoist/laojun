# 太上老君系统快速参考

## 🚀 一键启动

```powershell
# Windows (推荐)
cd d:\laojun\deploy\scripts
.\one-click-deploy.ps1

# Linux/macOS
cd /path/to/laojun/deploy/scripts
./deploy.sh start
```

## 📋 常用命令

### 服务管理
```powershell
.\deploy.ps1 start      # 启动所有服务
.\deploy.ps1 stop       # 停止所有服务
.\deploy.ps1 restart    # 重启所有服务
.\deploy.ps1 status     # 检查服务状态
```

### 日志查看
```powershell
.\deploy.ps1 logs                # 查看所有日志
.\deploy.ps1 logs admin-api      # 查看管理API日志
.\deploy.ps1 logs marketplace-api # 查看插件市场API日志
```

### 数据管理
```powershell
.\deploy.ps1 migrate    # 数据库迁移
.\deploy.ps1 backup     # 备份数据
.\deploy.ps1 update     # 更新系统
```

## 🌐 访问地址

| 服务 | 地址 | 描述 |
|------|------|------|
| 插件市场 | http://localhost | 主页面 |
| 管理后台 | http://localhost:8888 | 系统管理 |
| 管理API | http://localhost:8080 | 后台API |
| 插件市场API | http://localhost:8082 | 市场API |
| 配置中心 | http://localhost:8081 | 配置管理 |

## ⚙️ 环境配置

### 开发环境
```powershell
.\one-click-deploy.ps1
```

### 生产环境
```powershell
.\one-click-deploy.ps1 -Production
```

### 清理重新部署
```powershell
.\one-click-deploy.ps1 -Clean
```

## 🔧 故障排除

### 端口被占用
```powershell
# 查看端口占用
netstat -ano | findstr :80

# 结束进程
taskkill /PID <进程ID> /F
```

### 服务异常
```powershell
# 查看服务状态
docker-compose ps

# 重启特定服务
docker-compose restart admin-api

# 查看错误日志
docker-compose logs admin-api
```

### 清理系统
```powershell
# 清理Docker缓存
docker system prune -f

# 完全重置
.\deploy.ps1 cleanup
```

## 📊 监控检查

```powershell
# 健康检查
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health

# 资源使用
docker stats

# 服务状态
docker-compose ps
```

## 🔐 默认配置

### 开发环境密码
- 数据库: `laojun123`
- Redis: `redis123`
- JWT密钥: `dev-jwt-secret-key`

### 生产环境
⚠️ **必须修改** `deploy/configs/.env` 中的密码配置

## 📁 重要目录

```
deploy/
├── configs/           # 环境配置
│   ├── .env.template  # 配置模板
│   ├── .env.development # 开发环境
│   └── .env.production  # 生产环境
├── docker/            # Docker配置
│   └── docker-compose.yml
├── scripts/           # 部署脚本
│   ├── deploy.ps1     # Windows脚本
│   ├── deploy.sh      # Linux脚本
│   └── one-click-deploy.ps1 # 一键部署
└── docs/              # 文档
    ├── deployment-guide.md # 完整指南
    └── quick-reference.md  # 快速参考
```

## 🆘 获取帮助

```powershell
# 查看脚本帮助
.\deploy.ps1 help
.\one-click-deploy.ps1 -Help

# 查看完整文档
notepad deploy\docs\deployment-guide.md
```

---

💡 **提示**: 首次部署建议使用 `.\one-click-deploy.ps1` 进行一键启动