# 微服务插件示例

这是一个微服务类型的插件示例，展示了如何创建独立运行的服务插件。

## 功能特性

- 独立的 HTTP 服务
- 健康检查端点
- 数据处理 API
- Docker 容器化部署
- 事件发送机制

## 文件结构

```
microservice-plugin/
├── plugin.json      # 插件配置文件
├── server.js        # 服务器主文件
├── package.json     # Node.js 依赖配置
├── Dockerfile       # Docker 构建文件
└── README.md        # 说明文档
```

## API 端点

### 健康检查
- **GET** `/health`
- 返回服务健康状态

### 数据处理
- **POST** `/api/process`
- 参数: `{ "data": "要处理的文本" }`
- 返回处理后的数据

### 插件信息
- **GET** `/api/info`
- 返回插件基本信息

## 部署方式

### 1. Docker 部署

```bash
# 构建镜像
docker build -t microservice-plugin:1.0.0 .

# 运行容器
docker run -p 8080:8080 microservice-plugin:1.0.0
```

### 2. 本地开发

```bash
# 安装依赖
npm install

# 启动服务
npm start

# 开发模式（自动重启）
npm run dev
```

## 配置说明

### plugin.json 关键配置

- `type`: "microservice" - 微服务类型
- `runtime`: "docker" - Docker 运行时
- `docker_image`: Docker 镜像名称
- `service_port`: 服务端口
- `health_check_path`: 健康检查路径

### 资源限制

- 内存限制: 512MB
- CPU 限制: 50%
- 磁盘限制: 1GB

## 使用示例

```bash
# 健康检查
curl http://localhost:8080/health

# 处理数据
curl -X POST http://localhost:8080/api/process \
  -H "Content-Type: application/json" \
  -d '{"data": "hello world"}'
```

## 扩展开发

基于此示例，您可以：

1. 添加更多 API 端点
2. 集成数据库
3. 实现认证授权
4. 添加日志记录
5. 集成消息队列
6. 实现缓存机制

## 许可证

MIT License