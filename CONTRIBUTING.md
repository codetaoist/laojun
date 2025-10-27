# 贡献指南

感谢您对太上老君微服务平台的关注！我们欢迎所有形式的贡献，包括但不限于代码、文档、测试、反馈和建议。

## 📋 目录

- [贡献方式](#贡献方式)
- [开发环境搭建](#开发环境搭建)
- [代码规范](#代码规范)
- [提交流程](#提交流程)
- [测试要求](#测试要求)
- [文档贡献](#文档贡献)
- [社区参与](#社区参与)
- [行为准则](#行为准则)

## 🤝 贡献方式

### 您可以通过以下方式为项目做出贡献：

- 🐛 **报告 Bug**: 发现问题并提交详细的 Issue
- 💡 **功能建议**: 提出新功能或改进建议
- 💻 **代码贡献**: 修复 Bug 或实现新功能
- 📚 **文档改进**: 完善文档、教程和示例
- 🧪 **测试用例**: 编写和改进测试用例
- 🌍 **国际化**: 翻译文档和界面
- 🎨 **UI/UX**: 改进用户界面和体验
- 📢 **推广**: 分享项目，撰写博客文章

## 🛠️ 开发环境搭建

### 前置要求

| 工具 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.21+ | 主要开发语言 |
| Node.js | 18+ | 前端开发语言 |
| PostgreSQL | 13+ | 主数据库 |
| Redis | 6+ | 缓存数据库 |
| Docker | 20.10+ | 容器化开发 |
| Kubernetes | 1.25+ | 本地测试集群 |
| Git | 2.30+ | 版本控制 |
| Make | 任意版本 | 构建工具 |

### 环境搭建步骤

1. **Fork 项目**
```bash
# 在 GitHub 上 Fork 项目到您的账户
# 然后克隆到本地
git clone https://github.com/YOUR_USERNAME/laojun.git
cd laojun

# 添加上游仓库
git remote add upstream https://github.com/codetaoist/laojun.git
```

2. **安装依赖**
```bash
# 安装 Go 依赖
go mod download

# 安装开发工具
make install-tools
```

3. **启动开发环境**
```bash
# 启动本地开发环境
make dev-up

# 或者使用 Docker Compose
docker-compose -f docker-compose.dev.yml up -d
```

4. **验证环境**
```bash
# 运行测试
make test

# 检查代码质量
make lint

# 构建项目
make build
```

### 开发工具配置

#### VS Code 配置

创建 `.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

#### GoLand 配置

1. 启用 Go Modules 支持
2. 配置代码格式化工具为 `goimports`
3. 启用 `golangci-lint` 作为代码检查工具

## 📝 代码规范

### Go 代码规范

我们遵循标准的 Go 代码规范，并有一些额外的要求：

#### 1. 代码格式化

```bash
# 使用 goimports 格式化代码
goimports -w .

# 使用 gofmt 格式化
gofmt -w .
```

#### 2. 命名规范

```go
// ✅ 好的命名
type ServiceRegistry interface {
    RegisterService(service *Service) error
    DeregisterService(serviceID string) error
}

type HTTPServer struct {
    port   int
    router *gin.Engine
}

// ❌ 不好的命名
type sr interface {
    reg(s *Service) error
    dereg(id string) error
}

type server struct {
    p int
    r *gin.Engine
}
```

#### 3. 错误处理

```go
// ✅ 好的错误处理
func (s *ServiceRegistry) RegisterService(service *Service) error {
    if service == nil {
        return errors.New("service cannot be nil")
    }
    
    if err := s.validateService(service); err != nil {
        return fmt.Errorf("service validation failed: %w", err)
    }
    
    if err := s.consul.Register(service); err != nil {
        return fmt.Errorf("failed to register service to consul: %w", err)
    }
    
    return nil
}

// ❌ 不好的错误处理
func (s *ServiceRegistry) RegisterService(service *Service) error {
    s.consul.Register(service)  // 忽略错误
    return nil
}
```

#### 4. 注释规范

```go
// ServiceRegistry 定义了服务注册和发现的接口
// 它提供了服务的注册、注销、发现和健康检查功能
type ServiceRegistry interface {
    // RegisterService 注册一个新服务到注册中心
    // 如果服务已存在，将更新其信息
    RegisterService(service *Service) error
    
    // DeregisterService 从注册中心注销指定的服务
    // serviceID 是服务的唯一标识符
    DeregisterService(serviceID string) error
}
```

### 项目结构规范

```
taishanglaojun/
├── cmd/                    # 主程序入口
│   ├── discovery/         # 服务发现服务
│   └── monitoring/        # 监控服务
├── internal/              # 内部包（不对外暴露）
│   ├── discovery/         # 服务发现逻辑
│   ├── monitoring/        # 监控逻辑
│   ├── config/           # 配置管理
│   └── common/           # 通用工具
├── pkg/                   # 公共包（可对外暴露）
│   ├── client/           # 客户端 SDK
│   ├── types/            # 类型定义
│   └── utils/            # 工具函数
├── api/                   # API 定义
│   ├── proto/            # gRPC 协议定义
│   └── openapi/          # OpenAPI 规范
├── deployments/           # 部署配置
├── docs/                  # 文档
├── examples/              # 示例代码
├── scripts/               # 脚本文件
└── tests/                 # 测试文件
```

## 🔄 提交流程

### 1. 创建分支

```bash
# 同步上游代码
git fetch upstream
git checkout main
git merge upstream/main

# 创建功能分支
git checkout -b feature/your-feature-name

# 或者修复分支
git checkout -b fix/issue-number
```

### 2. 开发和测试

```bash
# 开发您的功能
# ...

# 运行测试
make test

# 运行代码检查
make lint

# 运行安全扫描
make security-scan
```

### 3. 提交代码

#### 提交信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**类型 (type)**:
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式化（不影响功能）
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

**示例**:
```bash
# 新功能
git commit -m "feat(discovery): add service health check endpoint"

# Bug 修复
git commit -m "fix(monitoring): resolve memory leak in metrics collector"

# 文档更新
git commit -m "docs: update API documentation for service registration"

# 重构
git commit -m "refactor(config): simplify configuration loading logic"
```

### 4. 推送和创建 PR

```bash
# 推送到您的 Fork
git push origin feature/your-feature-name

# 在 GitHub 上创建 Pull Request
```

### Pull Request 要求

#### PR 标题格式
```
<type>[optional scope]: <description>
```

#### PR 描述模板
```markdown
## 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 文档更新
- [ ] 代码重构
- [ ] 性能优化
- [ ] 其他

## 变更描述
简要描述此 PR 的变更内容和目的。

## 相关 Issue
Fixes #(issue number)

## 测试
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试通过

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 通过了所有 CI 检查

## 截图（如适用）
如果有 UI 变更，请提供截图。

## 其他说明
任何其他需要说明的内容。
```

## 🧪 测试要求

### 测试类型

1. **单元测试**: 测试单个函数或方法
2. **集成测试**: 测试组件间的交互
3. **端到端测试**: 测试完整的用户场景
4. **性能测试**: 测试系统性能
5. **安全测试**: 测试安全漏洞

### 测试规范

#### 单元测试示例

```go
func TestServiceRegistry_RegisterService(t *testing.T) {
    tests := []struct {
        name    string
        service *Service
        wantErr bool
    }{
        {
            name: "valid service",
            service: &Service{
                ID:      "test-service-001",
                Name:    "test-service",
                Address: "127.0.0.1",
                Port:    8080,
            },
            wantErr: false,
        },
        {
            name:    "nil service",
            service: nil,
            wantErr: true,
        },
        {
            name: "invalid port",
            service: &Service{
                ID:      "test-service-002",
                Name:    "test-service",
                Address: "127.0.0.1",
                Port:    0,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            registry := NewServiceRegistry()
            err := registry.RegisterService(tt.service)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("RegisterService() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### 集成测试示例

```go
func TestServiceDiscovery_Integration(t *testing.T) {
    // 启动测试环境
    testEnv := setupTestEnvironment(t)
    defer testEnv.Cleanup()

    // 注册服务
    service := &Service{
        ID:      "integration-test-service",
        Name:    "test-service",
        Address: "127.0.0.1",
        Port:    8080,
    }
    
    err := testEnv.Registry.RegisterService(service)
    require.NoError(t, err)

    // 发现服务
    instances, err := testEnv.Registry.DiscoverService("test-service")
    require.NoError(t, err)
    require.Len(t, instances, 1)
    assert.Equal(t, service.ID, instances[0].ID)
}
```

### 测试覆盖率

- 新代码的测试覆盖率应达到 **80%** 以上
- 核心功能的测试覆盖率应达到 **90%** 以上

```bash
# 运行测试并生成覆盖率报告
make test-coverage

# 查看覆盖率报告
go tool cover -html=coverage.out
```

## 📚 文档贡献

### 文档类型

1. **API 文档**: 接口说明和示例
2. **用户指南**: 使用教程和最佳实践
3. **开发者文档**: 架构设计和开发指南
4. **运维文档**: 部署和运维指南

### 文档规范

#### Markdown 规范

```markdown
# 一级标题

## 二级标题

### 三级标题

#### 四级标题

**粗体文本**

*斜体文本*

`行内代码`

```go
// 代码块
func main() {
    fmt.Println("Hello, World!")
}
```

> 引用文本

- 无序列表项 1
- 无序列表项 2

1. 有序列表项 1
2. 有序列表项 2

[链接文本](https://example.com)

![图片描述](image.png)
```

#### 代码示例规范

```markdown
### 示例：注册服务

```go
package main

import (
    "context"
    "log"
    
    "github.com/codetaoist/laojun/pkg/client"
)

func main() {
    // 创建客户端
    client := client.New(&client.Config{
        Address: "http://localhost:8080",
        Timeout: 30, // 30秒超时
    })
    
    // 注册服务
    err := client.RegisterService(context.Background(), &client.Service{
        ID:      "my-service-001",
        Name:    "my-service",
        Address: "192.168.1.100",
        Port:    8080,
        Tags:    []string{"api", "v1"},
        Health: &client.HealthCheck{
            HTTP:     "http://192.168.1.100:8080/health",
            Interval: "10s",
            Timeout:  "3s",
        },
    })
    
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }
    
    log.Println("Service registered successfully")
}
```

**输出**:
```
Service registered successfully
```
```

## 👥 社区参与

### 沟通渠道

- 💬 [GitHub Discussions](https://github.com/codetaoist/laojun/discussions)
- 📧 [邮件列表](mailto:dev@codetaoist.com)
- 💬 [微信群](https://weixin.qq.com/codetaoist)
- 📱 [钉钉群](https://dingtalk.com/codetaoist)

### 参与方式

1. **参与讨论**: 在 GitHub Discussions 中参与技术讨论
2. **回答问题**: 帮助其他用户解决问题
3. **分享经验**: 撰写博客文章或教程
4. **组织活动**: 参与或组织技术分享会
5. **推广项目**: 在社交媒体上分享项目

### 成为维护者

如果您想成为项目维护者，需要满足以下条件：

1. **持续贡献**: 至少 6 个月的持续贡献
2. **代码质量**: 提交的代码质量高，遵循项目规范
3. **社区参与**: 积极参与社区讨论和帮助其他用户
4. **技术能力**: 对项目架构和技术栈有深入理解
5. **责任心**: 愿意承担维护项目的责任

## 📜 行为准则

### 我们的承诺

为了营造一个开放和友好的环境，我们承诺：

- 尊重所有参与者，无论其经验水平、性别、性取向、残疾、外貌、身材、种族、民族、年龄、宗教或国籍
- 使用友好和包容的语言
- 尊重不同的观点和经验
- 优雅地接受建设性批评
- 关注对社区最有利的事情
- 对其他社区成员表现出同理心

### 不可接受的行为

以下行为被认为是不可接受的：

- 使用性化的语言或图像，以及不受欢迎的性关注或性骚扰
- 恶意评论、侮辱/贬损评论，以及个人或政治攻击
- 公开或私下骚扰
- 未经明确许可，发布他人的私人信息，如物理或电子地址
- 在专业环境中可能被认为不合适的其他行为

### 执行

项目维护者有权利和责任删除、编辑或拒绝不符合本行为准则的评论、提交、代码、wiki 编辑、问题和其他贡献。

## 🎁 贡献者权益

### 认可方式

1. **贡献者列表**: 在 README 中列出所有贡献者
2. **发布说明**: 在版本发布说明中感谢贡献者
3. **社交媒体**: 在官方社交媒体上宣传重要贡献
4. **会议演讲**: 邀请优秀贡献者参与技术会议

### 奖励机制

1. **贡献徽章**: 根据贡献类型和数量颁发不同徽章
2. **纪念品**: 为活跃贡献者提供项目纪念品
3. **技术交流**: 邀请参与内部技术交流会
4. **职业机会**: 优先推荐相关工作机会

## 📞 联系我们

如果您有任何问题或建议，请通过以下方式联系我们：

- 📧 **邮箱**: [dev@codetaoist.com](mailto:dev@codetaoist.com)
- 💬 **GitHub**: [创建 Issue](https://github.com/codetaoist/laojun/issues/new)
- 🌐 **官网**: [https://codetaoist.com](https://codetaoist.com)

---

## 📚 相关资源

- [项目主页](https://github.com/codetaoist/laojun)
- [文档网站](https://docs.codetaoist.com)
- [API 文档](docs/api.md)
- [架构设计](docs/architecture.md)
- [部署指南](docs/deployment.md)

---

<div align="center">
  <p>🏮 感谢您对太上老君微服务平台的贡献！🏮</p>
  <p>让我们一起构建更好的微服务治理平台！</p>
</div>