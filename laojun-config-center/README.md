# Laojun Config Center

太上老君配置中心

## 概述

本仓库包含太上老君平台的配置中心服务，提供集中化的配置管理、配置热更新、配置版本控制等功能。

## 功能特性

- 集中化配置管理
- 配置热更新
- 配置版本控制
- 配置权限管理
- 配置审计
- 多环境支持

## 快速开始

```bash
# 构建
go build -o bin/config-center ./cmd/config-center

# 运行
./bin/config-center
```

## 配置格式

支持 JSON、YAML、TOML 等多种配置格式。

## API 接口

- GET /api/v1/configs/{key} - 获取配置
- PUT /api/v1/configs/{key} - 更新配置
- DELETE /api/v1/configs/{key} - 删除配置