# 太上老君系统 Docker 部署 Makefile

# 默认环境
ENV ?= dev
COMPOSE_FILE = docker-compose.yml
COMPOSE_ENV_FILE = docker-compose.$(ENV).yml

# 项目名称
PROJECT_NAME = laojun

# Docker Compose 命令
DOCKER_COMPOSE = docker-compose -p $(PROJECT_NAME) -f $(COMPOSE_FILE)

# 如果环境特定文件存在，则添加
ifneq (,$(wildcard $(COMPOSE_ENV_FILE)))
    DOCKER_COMPOSE += -f $(COMPOSE_ENV_FILE)
endif

# 颜色定义
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

.PHONY: help build up down restart logs clean prune status health backup restore

# 默认目标
.DEFAULT_GOAL := help

## 显示帮助信息
help:
	@echo "$(GREEN)太上老君系统 Docker 部署命令$(NC)"
	@echo ""
	@echo "$(YELLOW)基础命令:$(NC)"
	@echo "  make build ENV=dev|prod     构建所有镜像"
	@echo "  make up ENV=dev|prod        启动所有服务"
	@echo "  make down                   停止所有服务"
	@echo "  make restart ENV=dev|prod   重启所有服务"
	@echo "  make logs                   查看所有服务日志"
	@echo "  make status                 查看服务状态"
	@echo ""
	@echo "$(YELLOW)开发命令:$(NC)"
	@echo "  make dev                    启动开发环境"
	@echo "  make dev-tools              启动开发工具"
	@echo "  make db-tools               启动数据库管理工具"
	@echo ""
	@echo "$(YELLOW)生产命令:$(NC)"
	@echo "  make prod                   启动生产环境"
	@echo "  make prod-build             构建生产镜像"
	@echo ""
	@echo "$(YELLOW)维护命令:$(NC)"
	@echo "  make clean                  清理未使用的镜像和容器"
	@echo "  make prune                  深度清理Docker资源"
	@echo "  make health                 检查服务健康状态"
	@echo "  make backup                 备份数据"
	@echo "  make restore FILE=xxx       恢复数据"
	@echo ""
	@echo "$(YELLOW)示例:$(NC)"
	@echo "  make dev                    # 启动开发环境"
	@echo "  make prod                   # 启动生产环境"
	@echo "  make logs SERVICE=postgres  # 查看PostgreSQL日志"

## 构建所有镜像
build:
	@echo "$(GREEN)构建镜像 (环境: $(ENV))...$(NC)"
	$(DOCKER_COMPOSE) build --no-cache

## 启动所有服务
up:
	@echo "$(GREEN)启动服务 (环境: $(ENV))...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)服务启动完成$(NC)"
	@make status

## 停止所有服务
down:
	@echo "$(YELLOW)停止所有服务...$(NC)"
	$(DOCKER_COMPOSE) down

## 重启所有服务
restart:
	@echo "$(YELLOW)重启服务 (环境: $(ENV))...$(NC)"
	$(DOCKER_COMPOSE) restart
	@make status

## 查看服务日志
logs:
ifdef SERVICE
	$(DOCKER_COMPOSE) logs -f $(SERVICE)
else
	$(DOCKER_COMPOSE) logs -f
endif

## 查看服务状态
status:
	@echo "$(GREEN)服务状态:$(NC)"
	$(DOCKER_COMPOSE) ps

## 检查服务健康状态
health:
	@echo "$(GREEN)检查服务健康状态:$(NC)"
	@docker ps --filter "name=$(PROJECT_NAME)" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

## 启动开发环境
dev:
	@echo "$(GREEN)启动开发环境...$(NC)"
	@make up ENV=dev
	@echo "$(GREEN)开发环境启动完成$(NC)"
	@echo "$(YELLOW)访问地址:$(NC)"
	@echo "  插件市场: http://localhost"
	@echo "  管理后台: http://localhost:8888"
	@echo "  API文档: http://localhost:8080/swagger"
	@echo "  数据库管理: http://localhost:8090 (需要启动 db-tools)"
	@echo "  Redis管理: http://localhost:8091 (需要启动 db-tools)"

## 启动开发工具
dev-tools:
	@echo "$(GREEN)启动开发工具...$(NC)"
	$(DOCKER_COMPOSE) --profile dev-tools up -d

## 启动数据库管理工具
db-tools:
	@echo "$(GREEN)启动数据库管理工具...$(NC)"
	$(DOCKER_COMPOSE) --profile db-tools up -d
	@echo "$(GREEN)数据库管理工具启动完成$(NC)"
	@echo "$(YELLOW)访问地址:$(NC)"
	@echo "  Adminer (数据库): http://localhost:8090"
	@echo "  Redis Commander: http://localhost:8091"

## 启动生产环境
prod:
	@echo "$(GREEN)启动生产环境...$(NC)"
	@make up ENV=prod
	@echo "$(GREEN)生产环境启动完成$(NC)"

## 构建生产镜像
prod-build:
	@echo "$(GREEN)构建生产镜像...$(NC)"
	@make build ENV=prod

## 清理未使用的镜像和容器
clean:
	@echo "$(YELLOW)清理未使用的Docker资源...$(NC)"
	docker system prune -f
	docker image prune -f

## 深度清理Docker资源
prune:
	@echo "$(RED)深度清理Docker资源 (包括数据卷)...$(NC)"
	@echo "$(RED)警告: 这将删除所有未使用的Docker资源!$(NC)"
	@read -p "确认继续? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		docker system prune -a -f --volumes; \
	else \
		echo ""; \
		echo "操作已取消"; \
	fi

## 备份数据
backup:
	@echo "$(GREEN)备份数据...$(NC)"
	@mkdir -p ./backups
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	docker exec $(PROJECT_NAME)_postgres_1 pg_dump -U laojun laojun > ./backups/postgres_$$TIMESTAMP.sql; \
	docker exec $(PROJECT_NAME)_redis_1 redis-cli -a $${REDIS_PASSWORD:-redis123} --rdb /data/dump_$$TIMESTAMP.rdb; \
	docker cp $(PROJECT_NAME)_redis_1:/data/dump_$$TIMESTAMP.rdb ./backups/; \
	echo "$(GREEN)备份完成: ./backups/postgres_$$TIMESTAMP.sql, ./backups/dump_$$TIMESTAMP.rdb$(NC)"

## 恢复数据
restore:
ifndef FILE
	@echo "$(RED)错误: 请指定备份文件$(NC)"
	@echo "使用方法: make restore FILE=./backups/postgres_20240101_120000.sql"
else
	@echo "$(GREEN)恢复数据: $(FILE)$(NC)"
	@if [[ "$(FILE)" == *.sql ]]; then \
		docker exec -i $(PROJECT_NAME)_postgres_1 psql -U laojun laojun < $(FILE); \
	elif [[ "$(FILE)" == *.rdb ]]; then \
		docker cp $(FILE) $(PROJECT_NAME)_redis_1:/data/dump.rdb; \
		docker restart $(PROJECT_NAME)_redis_1; \
	else \
		echo "$(RED)错误: 不支持的文件格式$(NC)"; \
	fi
endif

## 进入容器shell
shell:
ifdef SERVICE
	$(DOCKER_COMPOSE) exec $(SERVICE) /bin/sh
else
	@echo "$(RED)错误: 请指定服务名称$(NC)"
	@echo "使用方法: make shell SERVICE=postgres"
endif

## 查看配置
config:
	@echo "$(GREEN)Docker Compose 配置:$(NC)"
	$(DOCKER_COMPOSE) config

## 拉取最新镜像
pull:
	@echo "$(GREEN)拉取最新镜像...$(NC)"
	$(DOCKER_COMPOSE) pull

## 更新服务
update:
	@echo "$(GREEN)更新服务...$(NC)"
	@make pull
	@make build
	@make restart