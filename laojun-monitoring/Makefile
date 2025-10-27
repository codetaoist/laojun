# 太上老君监控系统 Makefile
# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# 构建信息
BINARY_NAME = laojun-monitoring
MAIN_PATH = ./cmd/main.go
BUILD_DIR = ./build
DOCKER_IMAGE = laojun-monitoring
DOCKER_TAG ?= latest

# Go 相关变量
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

# 构建标志
LDFLAGS = -w -s \
	-X 'main.Version=$(VERSION)' \
	-X 'main.Commit=$(COMMIT)' \
	-X 'main.BuildTime=$(BUILD_TIME)'

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "太上老君监控系统 - 构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关
.PHONY: dev
dev: ## 启动开发环境
	@echo "启动开发环境..."
	go run $(MAIN_PATH) -config config/config.yaml

.PHONY: dev-watch
dev-watch: ## 启动开发环境（文件监控）
	@echo "启动开发环境（文件监控）..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "请先安装 air: go install github.com/cosmtrek/air@latest"; \
	fi

# 构建相关
.PHONY: build
build: clean ## 构建二进制文件
	@echo "构建 $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-all
build-all: clean ## 构建所有平台的二进制文件
	@echo "构建所有平台的二进制文件..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	
	# Linux ARM64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# Windows AMD64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	# macOS AMD64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	
	# macOS ARM64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	@echo "所有平台构建完成"

.PHONY: install
install: build ## 安装到系统
	@echo "安装 $(BINARY_NAME)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成"

# 测试相关
.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成完成: coverage.html"

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "运行竞态检测测试..."
	go test -race -v ./...

.PHONY: benchmark
benchmark: ## 运行基准测试
	@echo "运行基准测试..."
	go test -bench=. -benchmem ./...

# 代码质量
.PHONY: lint
lint: ## 代码检查
	@echo "运行代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "请先安装 golangci-lint"; \
	fi

.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

.PHONY: vet
vet: ## 静态分析
	@echo "运行静态分析..."
	go vet ./...

.PHONY: mod
mod: ## 整理依赖
	@echo "整理依赖..."
	go mod tidy
	go mod verify

# Docker 相关
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	docker run -d \
		--name $(DOCKER_IMAGE) \
		-p 8080:8080 \
		-v $(PWD)/config:/app/config \
		-v $(PWD)/data:/app/data \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-stop
docker-stop: ## 停止 Docker 容器
	@echo "停止 Docker 容器..."
	docker stop $(DOCKER_IMAGE) || true
	docker rm $(DOCKER_IMAGE) || true

.PHONY: docker-push
docker-push: docker-build ## 推送 Docker 镜像
	@echo "推送 Docker 镜像..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

# Docker Compose 相关
.PHONY: up
up: ## 启动完整环境
	@echo "启动完整环境..."
	docker-compose up -d

.PHONY: down
down: ## 停止完整环境
	@echo "停止完整环境..."
	docker-compose down

.PHONY: logs
logs: ## 查看日志
	@echo "查看日志..."
	docker-compose logs -f

.PHONY: restart
restart: down up ## 重启完整环境

# 部署相关
.PHONY: deploy-dev
deploy-dev: ## 部署到开发环境
	@echo "部署到开发环境..."
	@./scripts/deploy.sh dev

.PHONY: deploy-prod
deploy-prod: ## 部署到生产环境
	@echo "部署到生产环境..."
	@./scripts/deploy.sh prod

# 清理相关
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.prof

.PHONY: clean-docker
clean-docker: ## 清理 Docker 资源
	@echo "清理 Docker 资源..."
	docker system prune -f
	docker volume prune -f

.PHONY: clean-all
clean-all: clean clean-docker ## 清理所有文件

# 工具安装
.PHONY: tools
tools: ## 安装开发工具
	@echo "安装开发工具..."
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 初始化项目
.PHONY: init
init: tools mod ## 初始化项目
	@echo "初始化项目..."
	@mkdir -p data logs
	@cp config/config.yaml config/local.yaml 2>/dev/null || true
	@echo "项目初始化完成"

# 健康检查
.PHONY: health
health: ## 检查服务健康状态
	@echo "检查服务健康状态..."
	@curl -f http://localhost:8080/api/v1/health || echo "服务未运行"

# 版本信息
.PHONY: version
version: ## 显示版本信息
	@echo "版本: $(VERSION)"
	@echo "提交: $(COMMIT)"
	@echo "构建时间: $(BUILD_TIME)"

# 生成发布包
.PHONY: release
release: clean build-all ## 生成发布包
	@echo "生成发布包..."
	@mkdir -p release
	@cd $(BUILD_DIR) && \
		for binary in $(BINARY_NAME)-*; do \
			if [[ $$binary == *.exe ]]; then \
				zip -r ../release/$${binary%.exe}-$(VERSION).zip $$binary ../config/ ../README.md; \
			else \
				tar -czf ../release/$$binary-$(VERSION).tar.gz $$binary ../config/ ../README.md; \
			fi; \
		done
	@echo "发布包生成完成: release/"