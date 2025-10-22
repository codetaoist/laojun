# etc/laojun

用于存放系统级配置文件（开发/部署示例模板）：
- `admin-api.yaml`：后端API服务配置
- `database.yaml`：数据库、缓存、时序与对象存储配置
- `nginx/`：Web服务器配置（示例）
- `docker/`：Docker Compose 等部署示例

注意：此目录为仓库中的开发态/示例配置位置。生产安装应将文件落位到系统 `/etc/laojun/`。