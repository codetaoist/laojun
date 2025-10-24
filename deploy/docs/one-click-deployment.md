# 太上老君系统 - 一键部署说明

## 🚀 快速开始

### 方法一：双击运行（最简单）
1. 双击 `一键部署.bat` 文件
2. 等待部署完成
3. 访问 http://localhost

### 方法二：PowerShell运行
```powershell
# 开发环境部署（默认）
.\deploy\scripts\one-click-deploy.ps1

# 生产环境部署
.\deploy\scripts\one-click-deploy.ps1 -Production

# 清理重新部署
.\deploy\scripts\one-click-deploy.ps1 -Clean

# 查看帮助
.\deploy\scripts\one-click-deploy.ps1 -Help
```

## 📋 前置要求

1. **Docker Desktop**
   - 下载地址：https://www.docker.com/products/docker-desktop
   - 确保Docker服务正在运行

2. **系统要求**
   - Windows 10/11
   - 至少 4GB 可用内存
   - 至少 10GB 可用磁盘空间

## 🌐 访问地址

部署完成后，您可以通过以下地址访问：

| 服务 | 地址 | 说明 |
|------|------|------|
| 插件市场（主页） | http://localhost | 系统主页面 |
| 管理后台 | http://localhost:8888 | 系统管理界面 |
| API文档 | http://localhost:8080/swagger | API接口文档 |
| 配置中心 | http://localhost:8081 | 配置管理 |

## 🧪 测试部署

```powershell
# 基础测试
.\deploy\scripts\test-deployment.ps1

# 详细测试
.\deploy\scripts\test-deployment.ps1 -Detailed
```

## 🔧 常用命令

### 服务管理
```powershell
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 启动服务
docker compose up -d
```

### 故障排除
```powershell
# 查看特定服务日志
docker compose logs admin-api
docker compose logs marketplace-api
docker compose logs config-center

# 重启特定服务
docker compose restart admin-api

# 查看容器资源使用
docker stats

# 清理系统
docker system prune -f
```

## ❗ 常见问题

### 1. 端口被占用
```powershell
# 查看端口占用
netstat -ano | findstr :80
netstat -ano | findstr :8080

# 停止占用进程
taskkill /PID <进程ID> /F
```

### 2. Docker服务未启动
- 启动Docker Desktop
- 等待Docker完全启动后重试

### 3. 内存不足
- 关闭其他应用程序
- 在Docker Desktop中增加内存分配

### 4. 服务启动慢
- 首次启动需要下载镜像，请耐心等待
- 可以查看日志了解启动进度：`docker compose logs -f`

## 📞 获取帮助

1. 查看详细部署指南：`deploy/docs/deployment-guide.md`
2. 查看快速参考：`deploy/docs/quick-reference.md`
3. 运行测试脚本：`.\deploy\scripts\test-deployment.ps1`

## 🔄 更新系统

```powershell
# 停止服务
docker compose down

# 拉取最新代码
git pull

# 重新部署
.\deploy\scripts\one-click-deploy.ps1 -Clean
```

---

**提示**：首次部署建议使用 `一键部署.bat` 进行双击运行，这是最简单的方式！