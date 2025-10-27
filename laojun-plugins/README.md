# Laojun Plugins

太上老君插件系统

## 概述

本仓库包含太上老君平台的插件系统，提供插件开发 SDK、运行时环境、插件注册中心等功能。

## 功能特性

- 插件开发 SDK
- 插件运行时环境
- 插件注册中心
- 插件生命周期管理
- 插件通信机制

## 目录结构

```
laojun-plugins/
├── sdk/            # 插件开发 SDK
├── runtime/        # 插件运行时
├── registry/       # 插件注册中心
├── examples/       # 示例插件
└── docs/           # 开发文档
```

## 插件开发

参考 `examples/` 目录下的示例插件，使用 SDK 开发自定义插件。

## 快速开始

```bash
# 构建插件运行时
go build -o bin/plugin-runtime ./runtime

# 运行插件注册中心
go build -o bin/plugin-registry ./registry
```