# 项目目录结构说明

## 根目录结构

```
├── .env.*                    # 环境配置文件模板
├── Makefile                  # 构建和部署脚本
├── README.md                 # 项目说明文档
├── go.mod/go.work           # Go 模块配置
├── docker-compose.test.yml   # 测试环境配置
├── api/                     # API 规范文档
├── bin/                     # 编译后的二进制文件
├── cmd/                     # 应用程序入口点
├── configs/                 # 配置文件
├── db/                      # 数据库迁移文件
├── deployments/             # 部署配置
├── docker/                  # Docker 配置
├── docs/                    # 项目文档
├── examples/                # 示例代码
├── internal/                # 内部应用代码
├── k8s/                     # Kubernetes 配置
├── laojun/                  # 项目特定配置
├── pkg/                     # 可重用的库代码
├── scripts/                 # 数据库脚本和工具
├── sql/                     # SQL 脚本文件
├── tests/                   # 测试文件
├── third_party/             # 第三方依赖
├── tools/                   # 开发工具
├── var/                     # 运行时数据
└── web/                     # 前端应用
```

## 目录清理说明

### 已删除的文件
- 删除了16个临时测试和调试文件（如 `check_*.go`, `test_*.go`, `debug_*.go`）
- 删除了临时测试结果文件 `test_results.json`
- 合并了重复的 `test/` 目录到 `tests/` 目录

### 重新组织的文件
- 将数据库相关脚本移动到 `scripts/` 目录
- 将所有SQL文件整理到 `sql/` 目录
- 将测试文件统一到 `tests/` 目录

### 保留的文件
- 环境配置文件（development, production, staging）- 用于不同部署环境
- Docker配置文件 - 分别用于开发和部署
- 核心配置文件（Makefile, go.mod, README.md）

## 目录用途说明

- **cmd/**: 应用程序的主要入口点，每个子目录包含一个可执行程序
- **internal/**: 应用程序的内部代码，不对外暴露
- **pkg/**: 可以被外部应用程序使用的库代码
- **scripts/**: 数据库迁移、修复和维护脚本
- **sql/**: 数据库相关的SQL脚本文件
- **tests/**: 所有测试文件，包括单元测试、集成测试和性能测试
- **web/**: 前端应用程序（admin 和 marketplace）
- **docs/**: 项目文档和API文档
- **configs/**: 应用程序配置文件
- **deployments/**: 部署相关的配置文件
- **docker/**: Docker 容器配置
- **k8s/**: Kubernetes 部署配置

这种结构遵循了 Go 项目的标准布局，提高了代码的可维护性和可读性。