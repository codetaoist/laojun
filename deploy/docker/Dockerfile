# 多阶段构建 Dockerfile
# 第一阶段：构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go 工作区文件
COPY go.work go.work.sum ./
COPY go.mod go.sum ./

# 复制子模块的 go.mod 文件
COPY pkg/plugins/go.mod ./pkg/plugins/
COPY pkg/shared/go.mod ./pkg/shared/
COPY tools/debug/go.mod ./tools/debug/
COPY tools/plugin-cli/go.mod ./tools/plugin-cli/
COPY tools/swagger/go.mod ./tools/swagger/

# 配置 Go 工具链和代理
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,direct

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# Admin API 构建目标
FROM builder AS admin-api-builder
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o admin-api ./cmd/admin-api

# Config Center 构建目标
FROM builder AS config-center-builder
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o config-center ./cmd/config-center

# Marketplace API 构建目标
FROM builder AS marketplace-api-builder
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o marketplace-api ./cmd/marketplace-api

# Admin API 运行阶段
FROM alpine:latest AS admin-api

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=admin-api-builder /app/admin-api .

# 复制配置文件
COPY configs ./configs
COPY migrations ./migrations

# 创建必要的目录
RUN mkdir -p /app/logs /app/uploads /app/temp && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./admin-api"]

# Config Center 运行阶段
FROM alpine:latest AS config-center

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=config-center-builder /app/config-center .

# 复制配置文件
COPY configs ./configs
COPY migrations ./migrations

# 创建必要的目录
RUN mkdir -p /app/logs /app/uploads /app/temp && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 启动应用
CMD ["./config-center"]

# Marketplace API 运行阶段
FROM alpine:latest AS marketplace-api

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=marketplace-api-builder /app/marketplace-api .

# 复制配置文件
COPY configs ./configs
COPY migrations ./migrations

# 创建必要的目录
RUN mkdir -p /app/logs /app/uploads /app/temp && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8082

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# 启动应用
CMD ["./marketplace-api"]