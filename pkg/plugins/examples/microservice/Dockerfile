# 微服务插件 Dockerfile 示例
# 基于 Go 的轻量级镜像
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o plugin ./cmd/plugin

# 运行时镜像
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN addgroup -g 1001 -S plugin && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G plugin -g plugin plugin

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/plugin .

# 复制配置文件
COPY --from=builder /app/config/ ./config/

# 设置文件权限
RUN chown -R plugin:plugin /app

# 切换到非 root 用户
USER plugin

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 设置环境变量
ENV PLUGIN_ENV=production
ENV PLUGIN_PORT=8080
ENV PLUGIN_LOG_LEVEL=info

# 启动命令
CMD ["./plugin"]