# Laojun Platform Makefile
# 用于构建、测试和部署 Laojun 平台

# 变量定义
PROJECT_NAME := laojun
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOMOD := $(GO) mod

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 目录定义
SRC_DIR := ./usr/src/laojun
BUILD_DIR := ./build
DIST_DIR := ./dist
DOCKER_DIR := ./docker
SCRIPTS_DIR := ./scripts

# 服务列表
SERVICES := admin-api config-center marketplace-api

# 测试相关
TEST_TIMEOUT := 30m
COVERAGE_DIR := ./coverage
BENCHMARK_TIME := 10s

# Docker 相关
DOCKER_REGISTRY := registry.example.com
DOCKER_NAMESPACE := laojun

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Laojun Platform Makefile"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 环境检查
.PHONY: check-env
check-env: ## 检查开发环境
	@echo "检查开发环境..."
	@which go > /dev/null || (echo "错误: Go 未安装" && exit 1)
	@which docker > /dev/null || (echo "错误: Docker 未安装" && exit 1)
	@which docker-compose > /dev/null || (echo "错误: Docker Compose 未安装" && exit 1)
	@echo "环境检查通过"

# 依赖管理
.PHONY: deps
deps: ## 下载和整理依赖
	@echo "下载依赖..."
	@cd $(SRC_DIR) && $(GOMOD) download
	@cd $(SRC_DIR) && $(GOMOD) tidy
	@echo "依赖管理完成"

# 代码格式化和检查
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	@cd $(SRC_DIR) && $(GOFMT) -s -w .
	@echo "代码格式化完成"

# 构建相关
.PHONY: build
build: deps fmt ## 构建所有服务
	@echo "构建所有服务..."
	@mkdir -p $(BUILD_DIR)
	@for service in $(SERVICES); do \
echo "构建 $$service..."; \
cd $(SRC_DIR)/$$service && $(GOBUILD) $(LDFLAGS) -o ../../$(BUILD_DIR)/$$service ./cmd/main.go; \
done
	@echo "构建完成"

# 测试相关
.PHONY: test
test: ## 运行所有测试
	@echo "运行所有测试..."
	@mkdir -p $(COVERAGE_DIR)
	@cd $(SRC_DIR) && $(GOTEST) -timeout $(TEST_TIMEOUT) -race -coverprofile=../../$(COVERAGE_DIR)/coverage.out ./...
	@echo "测试完成"

# 快速测试
.PHONY: quick-test
quick-test: ## 快速测试
	@echo "快速测试完成"

# 本地构建目标（非 Docker）
.PHONY: build-admin-api-local
build-admin-api-local:
	@echo "构建 admin-api（本地）..."
	@go build -o build/admin-api.exe ./cmd/admin-api

.PHONY: build-marketplace-api-local
build-marketplace-api-local:
	@echo "构建 marketplace-api（本地）..."
	@go build -o build/marketplace-api.exe ./cmd/marketplace-api

.PHONY: build-config-center-local
build-config-center-local:
	@echo "构建 config-center（本地）..."
	@go build -o build/config-center.exe ./cmd/config-center

.PHONY: build-migrate-local
build-migrate-local:
	@echo "构建 migrate（本地）..."
	@go build -o build/migrate.exe ./cmd/migrate

.PHONY: build-db-complete-migrate-local
build-db-complete-migrate-local:
	@echo "构建 db-complete-migrate（本地）..."
	@go build -o build/db-complete-migrate.exe ./cmd/db-complete-migrate

.PHONY: build-local
build-local: build-admin-api-local build-marketplace-api-local build-config-center-local build-migrate-local build-db-complete-migrate-local
	@echo "本地构建完成"

# 数据库迁移相关命令
.PHONY: migrate-complete-up
migrate-complete-up: ## 执行完整数据库迁移
	@echo "执行完整数据库迁移..."
	@go run cmd/db-complete-migrate/main.go -action=up

.PHONY: migrate-complete-status
migrate-complete-status: ## 查看完整迁移状态
	@echo "查看完整迁移状态..."
	@go run cmd/db-complete-migrate/main.go -action=status

.PHONY: migrate-complete-reset
migrate-complete-reset: ## 重置数据库（危险操作）
	@echo "重置数据库..."
	@go run cmd/db-complete-migrate/main.go -action=reset

.PHONY: migrate-generate
migrate-generate: ## 生成完整迁移文件
	@echo "生成完整迁移文件..."
	@go run scripts/generate_complete_migration.go

# 清理
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf $(COVERAGE_DIR)
	@echo "清理完成"