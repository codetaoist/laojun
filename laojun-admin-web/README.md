# Laojun Admin Web - 后台管理系统

Laojun Admin Web 是太上老君微服务架构中的后台管理系统前端，基于 React 18 + TypeScript + Ant Design 5 构建，提供完整的系统管理、用户管理、插件管理和监控功能。

## 功能概述

### 核心功能
- **用户管理**: 用户账户管理、权限分配、角色管理
- **插件管理**: 插件审核、上架管理、版本控制、统计分析
- **系统监控**: 实时监控面板、性能指标、告警管理
- **配置管理**: 系统配置、参数设置、环境管理
- **数据分析**: 用户行为分析、插件使用统计、收入分析

### 技术特性
- **现代化架构**: React 18 + TypeScript + Vite 构建
- **组件化设计**: 基于 Ant Design 5 的企业级 UI 组件
- **状态管理**: Zustand 轻量级状态管理
- **路由管理**: React Router v6 声明式路由
- **开发体验**: ESLint + TypeScript 严格类型检查

## 核心接口说明

### API 接口
```typescript
// 用户管理接口
interface UserAPI {
  getUsers(params: UserSearchParams): Promise<UserListResponse>
  createUser(data: CreateUserRequest): Promise<User>
  updateUser(id: string, data: UpdateUserRequest): Promise<User>
  deleteUser(id: string): Promise<void>
}

// 插件管理接口
interface PluginAPI {
  getPlugins(params: PluginSearchParams): Promise<PluginListResponse>
  reviewPlugin(id: string, action: ReviewAction): Promise<void>
  updatePluginStatus(id: string, status: PluginStatus): Promise<void>
  getPluginStats(id: string): Promise<PluginStats>
}

// 系统监控接口
interface MonitoringAPI {
  getSystemMetrics(): Promise<SystemMetrics>
  getAlerts(params: AlertParams): Promise<AlertListResponse>
  acknowledgeAlert(id: string): Promise<void>
}
```

### 组件接口
```typescript
// 用户表格组件
interface UserTableProps {
  dataSource: User[]
  loading?: boolean
  pagination?: PaginationConfig
  onEdit: (user: User) => void
  onDelete: (userId: string) => void
}

// 插件审核组件
interface PluginReviewProps {
  plugin: Plugin
  onApprove: (id: string) => void
  onReject: (id: string, reason: string) => void
}
```

## 详细依赖关系

### 内部依赖
- `@laojun/frontend-shared`: 共享组件库和工具函数
- 后端 API 服务:
  - `laojun-admin-api`: 管理后台 API 服务
  - `laojun-gateway`: API 网关服务

### 外部依赖
```json
{
  "核心框架": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.8.0"
  },
  "UI组件": {
    "antd": "^5.12.0",
    "@ant-design/icons": "^5.2.0"
  },
  "状态管理": {
    "zustand": "^4.4.0"
  },
  "HTTP客户端": {
    "axios": "^1.6.0"
  },
  "工具库": {
    "dayjs": "^1.11.0",
    "classnames": "^2.3.0",
    "ahooks": "^3.7.0"
  },
  "拖拽功能": {
    "react-dnd": "^16.0.1",
    "react-dnd-html5-backend": "^16.0.1"
  }
}
```

## 完整启动配置

### 环境要求
- Node.js >= 16.0.0
- pnpm >= 8.0.0

### 环境变量
```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_GATEWAY_URL=http://localhost:8000
VITE_APP_TITLE=太上老君管理后台
VITE_APP_VERSION=1.0.0

# .env.production
VITE_API_BASE_URL=https://api.laojun.com/api/v1
VITE_GATEWAY_URL=https://gateway.laojun.com
VITE_APP_TITLE=太上老君管理后台
VITE_APP_VERSION=1.0.0
```

### 配置文件示例
```typescript
// src/config/index.ts
export const config = {
  api: {
    baseURL: import.meta.env.VITE_API_BASE_URL,
    timeout: 10000,
  },
  app: {
    title: import.meta.env.VITE_APP_TITLE,
    version: import.meta.env.VITE_APP_VERSION,
  },
  auth: {
    tokenKey: 'laojun_admin_token',
    refreshTokenKey: 'laojun_admin_refresh_token',
  }
}
```

### 启动步骤
```bash
# 1. 安装依赖
pnpm install

# 2. 启动开发服务器
pnpm dev

# 3. 构建生产版本
pnpm build

# 4. 预览生产版本
pnpm preview
```

## 测试方法

### 单元测试
```bash
# 运行单元测试
pnpm test

# 运行测试覆盖率
pnpm test:coverage

# 监听模式运行测试
pnpm test:watch
```

### 集成测试
```bash
# 运行端到端测试
pnpm test:e2e

# 运行组件测试
pnpm test:component
```

### 代码质量检查
```bash
# ESLint 检查
pnpm lint

# 自动修复 ESLint 问题
pnpm lint:fix

# TypeScript 类型检查
pnpm type-check
```

## 部署注意事项

### 依赖服务
- **后端 API**: 确保 `laojun-admin-api` 服务正常运行
- **网关服务**: 确保 `laojun-gateway` 服务正常运行
- **认证服务**: 确保用户认证系统正常工作

### 资源需求
- **内存**: 最小 512MB，推荐 1GB
- **CPU**: 最小 1 核，推荐 2 核
- **存储**: 最小 100MB，推荐 500MB
- **网络**: 稳定的网络连接，支持 HTTPS

### 部署配置
```nginx
# Nginx 配置示例
server {
    listen 80;
    server_name admin.laojun.com;
    
    location / {
        root /var/www/laojun-admin-web/dist;
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://laojun-gateway:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 性能指标

### 基准数据
- **首屏加载时间**: < 2s (3G 网络)
- **页面切换时间**: < 500ms
- **API 响应时间**: < 1s (95% 请求)
- **内存使用**: < 100MB (运行时)
- **包大小**: < 2MB (gzipped)

### 性能优化
- **代码分割**: 路由级别的懒加载
- **资源压缩**: Gzip/Brotli 压缩
- **缓存策略**: 静态资源长期缓存
- **CDN 加速**: 静态资源 CDN 分发

## 开发指南

### 目录结构
```
src/
├── components/          # 通用组件
├── pages/              # 页面组件
├── stores/             # 状态管理
├── services/           # API 服务
├── utils/              # 工具函数
├── hooks/              # 自定义 Hooks
├── types/              # TypeScript 类型定义
└── config/             # 配置文件
```

### 开发规范
- 使用 TypeScript 严格模式
- 遵循 ESLint 规则
- 组件使用函数式组件 + Hooks
- 状态管理使用 Zustand
- API 调用统一使用 axios 实例

## 故障排除

### 常见问题
1. **构建失败**: 检查 Node.js 版本和依赖安装
2. **API 调用失败**: 检查后端服务状态和网络连接
3. **路由不工作**: 检查 Nginx 配置和 history 模式设置
4. **样式问题**: 检查 Less 编译和 Ant Design 主题配置

### 日志查看
```bash
# 开发环境日志
console.log() 输出到浏览器控制台

# 生产环境日志
# 配置日志收集服务，如 Sentry 或 LogRocket
```