# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o laojun-monitoring \
    ./cmd/main.go

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata curl

# 创建非root用户
RUN addgroup -g 1001 -S monitoring && \
    adduser -u 1001 -S monitoring -G monitoring

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/laojun-monitoring .

# 复制配置文件
COPY --chown=monitoring:monitoring config/ ./config/

# 创建数据目录
RUN mkdir -p ./data ./logs && \
    chown -R monitoring:monitoring ./data ./logs

# 切换到非root用户
USER monitoring

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# 设置环境变量
ENV GIN_MODE=release
ENV LOG_LEVEL=info

# 启动命令
CMD ["./laojun-monitoring", "-config", "config/config.yaml"]