# Makefile for laojun-shared library

.PHONY: help build test lint clean codegen check-api examples install-tools

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  build         - 构建所有工具"
	@echo "  test          - 运行所有测试"
	@echo "  lint          - 运行代码检查"
	@echo "  check-api     - 检查API规范"
	@echo "  examples      - 运行所有示例"
	@echo "  codegen       - 生成新模块代码模板"
	@echo "  clean         - 清理构建文件"
	@echo "  install-tools - 安装开发工具"

# 构建工具
build:
	@echo "🔨 构建代码生成工具..."
	@go build -o bin/codegen.exe ./tools/codegen
	@echo "🔨 构建API检查工具..."
	@go build -o bin/linter.exe ./tools/linter
	@echo "✅ 构建完成"

# 运行测试
test:
	@echo "🧪 运行单元测试..."
	@go test ./...
	@echo "🧪 运行集成测试..."
	@go test ./test/...
	@echo "✅ 测试完成"

# 代码检查
lint:
	@echo "🔍 运行golangci-lint..."
	@golangci-lint run
	@echo "✅ 代码检查完成"

# API规范检查
check-api: build
	@echo "📋 检查API规范..."
	@./bin/linter.exe -dir .
	@echo "✅ API规范检查完成"

# 运行示例
examples:
	@echo "🚀 运行缓存示例..."
	@go run examples/cache_example.go
	@echo ""
	@echo "🚀 运行工具示例..."
	@go run examples/utils_example.go
	@echo ""
	@echo "🚀 运行健康检查示例..."
	@go run examples/health_example.go
	@echo ""
	@echo "🚀 运行日志示例..."
	@go run examples/logger_example.go
	@echo "✅ 所有示例运行完成"

# 生成新模块代码模板
codegen: build
	@if [ -z "$(PACKAGE)" ]; then \
		echo "❌ 请指定包名: make codegen PACKAGE=mypackage"; \
		exit 1; \
	fi
	@echo "📝 生成 $(PACKAGE) 模块代码模板..."
	@./bin/codegen.exe -package $(PACKAGE) -output .
	@echo "✅ 代码模板生成完成"

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	@rm -rf bin/
	@go clean -cache
	@echo "✅ 清理完成"

# 安装开发工具
install-tools:
	@echo "📦 安装golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "📦 安装goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "📦 安装govulncheck..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "✅ 开发工具安装完成"

# 格式化代码
fmt:
	@echo "🎨 格式化代码..."
	@go fmt ./...
	@goimports -w .
	@echo "✅ 代码格式化完成"

# 安全检查
security:
	@echo "🔒 运行安全检查..."
	@govulncheck ./...
	@echo "✅ 安全检查完成"

# 完整检查（包含所有检查项）
check-all: fmt lint check-api test security
	@echo "✅ 所有检查完成"

# 发布前检查
pre-release: clean check-all examples
	@echo "🎉 发布前检查完成，可以发布！"

# 初始化新模块开发环境
init-module:
	@if [ -z "$(PACKAGE)" ]; then \
		echo "❌ 请指定包名: make init-module PACKAGE=mypackage"; \
		exit 1; \
	fi
	@echo "🚀 初始化 $(PACKAGE) 模块开发环境..."
	@make codegen PACKAGE=$(PACKAGE)
	@echo "📝 请编辑以下文件完成模块开发:"
	@echo "  - $(PACKAGE)/interface.go"
	@echo "  - $(PACKAGE)/implementation.go"
	@echo "  - $(PACKAGE)/errors.go"
	@echo "  - $(PACKAGE)/interface_test.go"
	@echo "  - examples/$(PACKAGE)_example.go"
	@echo "✅ 模块初始化完成"

# 模块开发完成后的验证
verify-module:
	@if [ -z "$(PACKAGE)" ]; then \
		echo "❌ 请指定包名: make verify-module PACKAGE=mypackage"; \
		exit 1; \
	fi
	@echo "🔍 验证 $(PACKAGE) 模块..."
	@go test ./$(PACKAGE)/...
	@./bin/linter.exe -dir ./$(PACKAGE)
	@go run examples/$(PACKAGE)_example.go
	@echo "✅ $(PACKAGE) 模块验证完成"