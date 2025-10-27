# Laojun Frontend Shared - 统一前端架构库

Laojun Frontend Shared 是太上老君微服务架构中的前端共享库，提供统一的组件、工具函数、API 客户端和状态管理解决方案，确保前端项目的一致性和可维护性。

## 功能概述

### 核心功能
- **通用组件库**: 基于 Ant Design 的业务组件封装
- **API 客户端**: 统一的 HTTP 请求封装和错误处理
- **路由管理**: 可复用的路由配置和权限控制
- **状态管理**: 跨应用的状态管理解决方案
- **工具函数**: 常用的业务工具函数和类型定义

### 技术特性
- **TypeScript**: 完整的类型定义和类型安全
- **模块化设计**: 支持按需导入，减少包体积
- **多格式输出**: 支持 ESM、CJS 和 UMD 格式
- **Tree Shaking**: 优化的构建配置，支持摇树优化
- **开发友好**: 完整的 TypeScript 声明文件

## 模块导出说明

### 主要导出模块
```typescript
// 主入口
import { version, config } from '@laojun/frontend-shared'

// API 模块
import { 
  createApiClient, 
  ApiClient, 
  RequestConfig,
  ResponseData 
} from '@laojun/frontend-shared/api'

// 组件模块
import { 
  LaojunButton, 
  LaojunTable, 
  LaojunForm,
  LaojunModal,
  LaojunUpload 
} from '@laojun/frontend-shared/components'

// 路由模块
import { 
  createRouter, 
  RouteConfig, 
  AuthGuard,
  PermissionGuard 
} from '@laojun/frontend-shared/router'

// 状态管理模块
import { 
  createStore, 
  useGlobalStore, 
  StoreConfig 
} from '@laojun/frontend-shared/stores'

// 工具函数模块
import { 
  formatDate, 
  validateForm, 
  debounce, 
  throttle,
  storage 
} from '@laojun/frontend-shared/utils'
```

### API 客户端接口
```typescript
// API 客户端配置
interface ApiClientConfig {
  baseURL: string
  timeout?: number
  headers?: Record<string, string>
  interceptors?: {
    request?: RequestInterceptor[]
    response?: ResponseInterceptor[]
  }
}

// 请求配置
interface RequestConfig {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  params?: Record<string, any>
  data?: any
  headers?: Record<string, string>
}

// 响应数据格式
interface ResponseData<T = any> {
  code: number
  message: string
  data: T
  timestamp: number
}
```

### 组件接口
```typescript
// 通用按钮组件
interface LaojunButtonProps {
  type?: 'primary' | 'secondary' | 'danger'
  size?: 'small' | 'medium' | 'large'
  loading?: boolean
  disabled?: boolean
  onClick?: () => void
  children: React.ReactNode
}

// 通用表格组件
interface LaojunTableProps<T = any> {
  dataSource: T[]
  columns: ColumnConfig<T>[]
  loading?: boolean
  pagination?: PaginationConfig
  rowSelection?: RowSelectionConfig<T>
  onRow?: (record: T) => React.HTMLAttributes<HTMLTableRowElement>
}

// 通用表单组件
interface LaojunFormProps {
  initialValues?: Record<string, any>
  onFinish?: (values: any) => void
  onFinishFailed?: (errorInfo: any) => void
  layout?: 'horizontal' | 'vertical' | 'inline'
  children: React.ReactNode
}
```

## 详细依赖关系

### 内部依赖
- 被以下项目使用:
  - `laojun-admin-web`: 管理后台前端
  - `laojun-marketplace-web`: 插件市场前端
  - 其他前端项目

### 外部依赖
```json
{
  "核心依赖": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "typescript": "^5.0.0"
  },
  "UI组件": {
    "antd": "^5.12.0",
    "@ant-design/icons": "^5.2.0"
  },
  "HTTP客户端": {
    "axios": "^1.6.0"
  },
  "状态管理": {
    "zustand": "^4.4.0"
  },
  "路由": {
    "react-router-dom": "^6.8.0"
  },
  "工具库": {
    "dayjs": "^1.11.0",
    "lodash-es": "^4.17.21",
    "classnames": "^2.3.0"
  }
}
```

### 开发依赖
```json
{
  "构建工具": {
    "vite": "^5.0.0",
    "rollup": "^4.0.0"
  },
  "类型检查": {
    "@types/react": "^18.2.0",
    "@types/lodash-es": "^4.17.12"
  },
  "代码质量": {
    "eslint": "^8.45.0",
    "prettier": "^3.0.0"
  }
}
```

## 完整启动配置

### 环境要求
- Node.js >= 18.0.0
- pnpm >= 8.0.0

### 开发环境配置
```bash
# .env.development
VITE_NODE_ENV=development
VITE_BUILD_TARGET=development
```

### 构建配置
```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  build: {
    lib: {
      entry: {
        index: resolve(__dirname, 'src/index.ts'),
        api: resolve(__dirname, 'src/api/index.ts'),
        components: resolve(__dirname, 'src/components/index.ts'),
        router: resolve(__dirname, 'src/router/index.ts'),
        stores: resolve(__dirname, 'src/stores/index.ts'),
        utils: resolve(__dirname, 'src/utils/index.ts')
      },
      formats: ['es', 'cjs']
    },
    rollupOptions: {
      external: ['react', 'react-dom', 'antd'],
      output: {
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
          antd: 'antd'
        }
      }
    }
  }
})
```

### 启动步骤
```bash
# 1. 安装依赖
pnpm install

# 2. 开发模式（监听文件变化）
pnpm dev

# 3. 构建库文件
pnpm build

# 4. 生成类型声明文件
pnpm build:types

# 5. 发布到 npm
pnpm publish
```

## 使用方法

### 安装
```bash
# 使用 pnpm
pnpm add @laojun/frontend-shared

# 使用 npm
npm install @laojun/frontend-shared

# 使用 yarn
yarn add @laojun/frontend-shared
```

### 基本使用
```typescript
// 1. 创建 API 客户端
import { createApiClient } from '@laojun/frontend-shared/api'

const apiClient = createApiClient({
  baseURL: 'https://api.laojun.com',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 2. 使用通用组件
import { LaojunButton, LaojunTable } from '@laojun/frontend-shared/components'

function MyComponent() {
  return (
    <div>
      <LaojunButton type="primary" onClick={() => console.log('clicked')}>
        点击我
      </LaojunButton>
      <LaojunTable 
        dataSource={data} 
        columns={columns}
        loading={loading}
      />
    </div>
  )
}

// 3. 使用工具函数
import { formatDate, debounce } from '@laojun/frontend-shared/utils'

const formattedDate = formatDate(new Date(), 'YYYY-MM-DD HH:mm:ss')
const debouncedSearch = debounce(searchFunction, 300)
```

### 高级使用
```typescript
// 1. 自定义 API 拦截器
import { createApiClient } from '@laojun/frontend-shared/api'

const apiClient = createApiClient({
  baseURL: 'https://api.laojun.com',
  interceptors: {
    request: [
      (config) => {
        // 添加认证 token
        const token = localStorage.getItem('token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      }
    ],
    response: [
      (response) => {
        // 统一处理响应
        if (response.data.code !== 200) {
          throw new Error(response.data.message)
        }
        return response.data.data
      }
    ]
  }
})

// 2. 创建全局状态管理
import { createStore } from '@laojun/frontend-shared/stores'

interface AppState {
  user: User | null
  theme: 'light' | 'dark'
  setUser: (user: User) => void
  setTheme: (theme: 'light' | 'dark') => void
}

const useAppStore = createStore<AppState>((set) => ({
  user: null,
  theme: 'light',
  setUser: (user) => set({ user }),
  setTheme: (theme) => set({ theme })
}))
```

## 测试方法

### 单元测试
```bash
# 运行所有测试
pnpm test

# 运行测试覆盖率
pnpm test:coverage

# 监听模式
pnpm test:watch
```

### 组件测试
```typescript
// 组件测试示例
import { render, screen, fireEvent } from '@testing-library/react'
import { LaojunButton } from '../components/LaojunButton'

describe('LaojunButton', () => {
  it('should render correctly', () => {
    render(<LaojunButton>Test Button</LaojunButton>)
    expect(screen.getByText('Test Button')).toBeInTheDocument()
  })

  it('should handle click events', () => {
    const handleClick = jest.fn()
    render(<LaojunButton onClick={handleClick}>Click Me</LaojunButton>)
    
    fireEvent.click(screen.getByText('Click Me'))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })
})
```

### API 测试
```typescript
// API 客户端测试
import { createApiClient } from '../api'

describe('ApiClient', () => {
  it('should make GET requests', async () => {
    const client = createApiClient({ baseURL: 'https://api.test.com' })
    const response = await client.get('/users')
    
    expect(response).toBeDefined()
  })

  it('should handle errors', async () => {
    const client = createApiClient({ baseURL: 'https://api.test.com' })
    
    await expect(client.get('/invalid')).rejects.toThrow()
  })
})
```

## 部署注意事项

### 构建输出
```
dist/
├── index.js              # CJS 格式主入口
├── index.esm.js          # ESM 格式主入口
├── index.d.ts            # TypeScript 声明文件
├── api/                  # API 模块
│   ├── index.js
│   ├── index.esm.js
│   └── index.d.ts
├── components/           # 组件模块
├── router/              # 路由模块
├── stores/              # 状态管理模块
└── utils/               # 工具函数模块
```

### 版本管理
```bash
# 发布补丁版本
pnpm version patch

# 发布次要版本
pnpm version minor

# 发布主要版本
pnpm version major

# 发布预发布版本
pnpm version prerelease
```

### 依赖项目更新
```bash
# 在使用该库的项目中更新
pnpm update @laojun/frontend-shared

# 或指定版本
pnpm add @laojun/frontend-shared@latest
```

## 性能指标

### 包大小
- **完整包**: < 500KB (gzipped)
- **API 模块**: < 50KB (gzipped)
- **组件模块**: < 200KB (gzipped)
- **工具模块**: < 30KB (gzipped)

### 构建时间
- **开发构建**: < 5s
- **生产构建**: < 30s
- **类型检查**: < 10s

## 开发指南

### 目录结构
```
src/
├── api/                 # API 客户端
│   ├── client.ts
│   ├── interceptors.ts
│   └── types.ts
├── components/          # 通用组件
│   ├── Button/
│   ├── Table/
│   ├── Form/
│   └── index.ts
├── router/             # 路由管理
│   ├── guards.ts
│   ├── config.ts
│   └── types.ts
├── stores/             # 状态管理
│   ├── create.ts
│   ├── persist.ts
│   └── types.ts
├── utils/              # 工具函数
│   ├── date.ts
│   ├── validation.ts
│   ├── storage.ts
│   └── index.ts
└── index.ts            # 主入口
```

### 开发规范
- 使用 TypeScript 严格模式
- 遵循 ESLint 和 Prettier 规则
- 组件必须有完整的 TypeScript 类型定义
- 工具函数必须有单元测试
- 导出的 API 必须有文档注释

### 贡献指南
1. Fork 项目并创建功能分支
2. 编写代码并添加测试
3. 确保所有测试通过
4. 提交 Pull Request
5. 等待代码审查和合并

## 故障排除

### 常见问题
1. **类型错误**: 确保安装了正确版本的 TypeScript
2. **构建失败**: 检查 Node.js 版本和依赖安装
3. **导入错误**: 确保使用正确的导入路径
4. **样式问题**: 确保正确导入了 Ant Design 样式

### 调试方法
```bash
# 查看详细构建信息
pnpm build --verbose

# 分析包大小
pnpm analyze

# 检查类型
pnpm type-check
```