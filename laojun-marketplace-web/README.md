# Laojun Marketplace Web - 插件市场前端

Laojun Marketplace Web 是太上老君微服务架构中的插件市场前端应用，基于 React 18 + TypeScript + Ant Design 5 构建，为用户提供插件浏览、搜索、购买和管理的完整体验。

## 功能概述

### 核心功能
- **插件浏览**: 插件列表展示、分类筛选、搜索功能
- **插件详情**: 详细信息展示、评价系统、版本历史
- **用户系统**: 用户注册登录、个人中心、收藏管理
- **购买流程**: 插件购买、支付集成、订单管理
- **开发者中心**: 插件上传、版本管理、收益统计

### 用户体验特性
- **响应式设计**: 支持桌面端和移动端访问
- **搜索优化**: 智能搜索、标签筛选、排序功能
- **个性化推荐**: 基于用户行为的插件推荐
- **社交功能**: 评价评论、分享功能、用户互动
- **多语言支持**: 国际化支持，多语言切换

### 技术特性
- **现代化架构**: React 18 + TypeScript + Vite 构建
- **组件化设计**: 基于 Ant Design 5 的企业级 UI 组件
- **状态管理**: Zustand 轻量级状态管理
- **路由管理**: React Router v6 声明式路由
- **样式方案**: Less + Sass 混合样式解决方案

## 核心接口说明

### API 接口
```typescript
// 插件相关接口
interface PluginAPI {
  getPlugins(params: PluginSearchParams): Promise<PluginListResponse>
  getPluginDetail(id: string): Promise<PluginDetail>
  getPluginVersions(id: string): Promise<PluginVersion[]>
  purchasePlugin(id: string, versionId: string): Promise<PurchaseResult>
  downloadPlugin(id: string, versionId: string): Promise<DownloadInfo>
}

// 用户相关接口
interface UserAPI {
  login(credentials: LoginRequest): Promise<LoginResponse>
  register(userData: RegisterRequest): Promise<User>
  getUserProfile(): Promise<UserProfile>
  updateProfile(data: UpdateProfileRequest): Promise<UserProfile>
  getUserPurchases(): Promise<Purchase[]>
  getUserFavorites(): Promise<Plugin[]>
}

// 评价相关接口
interface ReviewAPI {
  getPluginReviews(pluginId: string, params: ReviewParams): Promise<ReviewListResponse>
  createReview(pluginId: string, review: CreateReviewRequest): Promise<Review>
  updateReview(reviewId: string, review: UpdateReviewRequest): Promise<Review>
  deleteReview(reviewId: string): Promise<void>
}

// 开发者相关接口
interface DeveloperAPI {
  uploadPlugin(pluginData: PluginUploadRequest): Promise<Plugin>
  updatePlugin(id: string, data: PluginUpdateRequest): Promise<Plugin>
  getPluginStats(id: string): Promise<PluginStats>
  getEarnings(): Promise<EarningsData>
}
```

### 组件接口
```typescript
// 插件卡片组件
interface PluginCardProps {
  plugin: Plugin
  showPrice?: boolean
  showRating?: boolean
  onFavorite?: (pluginId: string) => void
  onPurchase?: (pluginId: string) => void
}

// 搜索组件
interface SearchBarProps {
  placeholder?: string
  onSearch: (keyword: string) => void
  onFilter: (filters: SearchFilters) => void
  suggestions?: string[]
}

// 评价组件
interface ReviewListProps {
  pluginId: string
  reviews: Review[]
  loading?: boolean
  onSubmitReview: (review: CreateReviewRequest) => void
}

// 购买组件
interface PurchaseModalProps {
  plugin: Plugin
  visible: boolean
  onConfirm: (paymentMethod: PaymentMethod) => void
  onCancel: () => void
}
```

### 数据模型
```typescript
// 插件模型
interface Plugin {
  id: string
  name: string
  description: string
  category: PluginCategory
  developer: Developer
  price: number
  rating: number
  downloadCount: number
  tags: string[]
  screenshots: string[]
  versions: PluginVersion[]
  createdAt: string
  updatedAt: string
}

// 用户模型
interface User {
  id: string
  username: string
  email: string
  avatar?: string
  role: UserRole
  profile: UserProfile
  purchases: Purchase[]
  favorites: string[]
}

// 评价模型
interface Review {
  id: string
  userId: string
  pluginId: string
  rating: number
  content: string
  createdAt: string
  user: {
    username: string
    avatar?: string
  }
}
```

## 详细依赖关系

### 内部依赖
- `@laojun/frontend-shared`: 共享组件库和工具函数
- 后端 API 服务:
  - `laojun-marketplace-api`: 插件市场 API 服务
  - `laojun-gateway`: API 网关服务
  - `laojun-plugins`: 插件管理服务

### 外部依赖
```json
{
  "核心框架": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.1"
  },
  "UI组件": {
    "antd": "^5.12.8",
    "@ant-design/icons": "^5.2.6"
  },
  "状态管理": {
    "zustand": "^4.4.7"
  },
  "HTTP客户端": {
    "axios": "^1.6.2"
  },
  "工具库": {
    "dayjs": "^1.11.10",
    "lodash-es": "^4.17.21"
  },
  "样式处理": {
    "less": "^4.2.0",
    "sass-embedded": "^1.93.2"
  }
}
```

## 完整启动配置

### 环境要求
- Node.js >= 18.0.0
- pnpm >= 8.0.0

### 环境变量
```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8081/api/v1
VITE_GATEWAY_URL=http://localhost:8000
VITE_APP_TITLE=太上老君插件市场
VITE_APP_VERSION=1.0.0
VITE_UPLOAD_URL=http://localhost:8081/api/v1/upload
VITE_CDN_URL=http://localhost:8081/static

# .env.production
VITE_API_BASE_URL=https://marketplace-api.laojun.com/api/v1
VITE_GATEWAY_URL=https://gateway.laojun.com
VITE_APP_TITLE=太上老君插件市场
VITE_APP_VERSION=1.0.0
VITE_UPLOAD_URL=https://marketplace-api.laojun.com/api/v1/upload
VITE_CDN_URL=https://cdn.laojun.com
```

### 配置文件示例
```typescript
// src/config/index.ts
export const config = {
  api: {
    baseURL: import.meta.env.VITE_API_BASE_URL,
    uploadURL: import.meta.env.VITE_UPLOAD_URL,
    timeout: 10000,
  },
  app: {
    title: import.meta.env.VITE_APP_TITLE,
    version: import.meta.env.VITE_APP_VERSION,
  },
  cdn: {
    baseURL: import.meta.env.VITE_CDN_URL,
  },
  auth: {
    tokenKey: 'laojun_marketplace_token',
    refreshTokenKey: 'laojun_marketplace_refresh_token',
  },
  payment: {
    alipay: {
      appId: 'your_alipay_app_id',
    },
    wechat: {
      appId: 'your_wechat_app_id',
    }
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

## 功能模块详解

### 插件浏览模块
```typescript
// 插件列表页面
const PluginListPage: React.FC = () => {
  const [plugins, setPlugins] = useState<Plugin[]>([])
  const [loading, setLoading] = useState(false)
  const [filters, setFilters] = useState<SearchFilters>({})
  
  const fetchPlugins = async (params: PluginSearchParams) => {
    setLoading(true)
    try {
      const response = await pluginAPI.getPlugins(params)
      setPlugins(response.data)
    } catch (error) {
      message.error('获取插件列表失败')
    } finally {
      setLoading(false)
    }
  }
  
  return (
    <div className="plugin-list-page">
      <SearchBar onSearch={handleSearch} onFilter={setFilters} />
      <PluginGrid plugins={plugins} loading={loading} />
      <Pagination {...paginationProps} />
    </div>
  )
}
```

### 用户认证模块
```typescript
// 登录页面
const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { login } = useAuthStore()
  
  const handleLogin = async (values: LoginForm) => {
    try {
      const response = await userAPI.login(values)
      login(response.user, response.token)
      message.success('登录成功')
      navigate('/')
    } catch (error) {
      message.error('登录失败，请检查用户名和密码')
    }
  }
  
  return (
    <div className="login-page">
      <Form onFinish={handleLogin}>
        <Form.Item name="username" rules={[{ required: true }]}>
          <Input placeholder="用户名" />
        </Form.Item>
        <Form.Item name="password" rules={[{ required: true }]}>
          <Input.Password placeholder="密码" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit">
            登录
          </Button>
        </Form.Item>
      </Form>
    </div>
  )
}
```

### 支付集成模块
```typescript
// 支付组件
const PaymentModal: React.FC<PaymentModalProps> = ({ 
  plugin, 
  visible, 
  onConfirm, 
  onCancel 
}) => {
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay')
  
  const handlePayment = async () => {
    try {
      const result = await pluginAPI.purchasePlugin(plugin.id, {
        paymentMethod,
        amount: plugin.price
      })
      
      // 跳转到支付页面
      window.location.href = result.paymentUrl
    } catch (error) {
      message.error('支付失败，请重试')
    }
  }
  
  return (
    <Modal title="购买插件" visible={visible} onCancel={onCancel}>
      <div className="payment-info">
        <h3>{plugin.name}</h3>
        <p>价格: ¥{plugin.price}</p>
      </div>
      <Radio.Group 
        value={paymentMethod} 
        onChange={(e) => setPaymentMethod(e.target.value)}
      >
        <Radio value="alipay">支付宝</Radio>
        <Radio value="wechat">微信支付</Radio>
        <Radio value="balance">账户余额</Radio>
      </Radio.Group>
      <div className="payment-actions">
        <Button onClick={onCancel}>取消</Button>
        <Button type="primary" onClick={handlePayment}>
          确认支付
        </Button>
      </div>
    </Modal>
  )
}
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

### 端到端测试
```typescript
// E2E 测试示例
describe('Plugin Purchase Flow', () => {
  it('should complete plugin purchase successfully', async () => {
    // 1. 访问插件详情页
    await page.goto('/plugins/test-plugin-id')
    
    // 2. 点击购买按钮
    await page.click('[data-testid="purchase-button"]')
    
    // 3. 选择支付方式
    await page.click('[data-testid="payment-alipay"]')
    
    // 4. 确认支付
    await page.click('[data-testid="confirm-payment"]')
    
    // 5. 验证跳转到支付页面
    await expect(page.url()).toContain('alipay.com')
  })
})
```

### 性能测试
```bash
# 使用 Lighthouse 进行性能测试
pnpm lighthouse

# 使用 Bundle Analyzer 分析包大小
pnpm analyze
```

## 部署注意事项

### 依赖服务
- **后端 API**: 确保 `laojun-marketplace-api` 服务正常运行
- **网关服务**: 确保 `laojun-gateway` 服务正常运行
- **插件服务**: 确保 `laojun-plugins` 服务正常运行
- **支付服务**: 确保支付网关配置正确

### 资源需求
- **内存**: 最小 512MB，推荐 1GB
- **CPU**: 最小 1 核，推荐 2 核
- **存储**: 最小 200MB，推荐 1GB
- **网络**: 稳定的网络连接，支持 HTTPS

### 部署配置
```nginx
# Nginx 配置示例
server {
    listen 80;
    server_name marketplace.laojun.com;
    
    location / {
        root /var/www/laojun-marketplace-web/dist;
        try_files $uri $uri/ /index.html;
        
        # 静态资源缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }
    
    location /api {
        proxy_pass http://laojun-marketplace-api:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### CDN 配置
```javascript
// CDN 资源配置
const cdnConfig = {
  images: 'https://cdn.laojun.com/images/',
  plugins: 'https://cdn.laojun.com/plugins/',
  assets: 'https://cdn.laojun.com/assets/'
}
```

## 性能指标

### 基准数据
- **首屏加载时间**: < 2s (3G 网络)
- **页面切换时间**: < 300ms
- **搜索响应时间**: < 500ms
- **图片加载时间**: < 1s
- **包大小**: < 3MB (gzipped)

### 性能优化策略
- **代码分割**: 路由级别和组件级别的懒加载
- **图片优化**: WebP 格式、懒加载、响应式图片
- **缓存策略**: 浏览器缓存、CDN 缓存、API 缓存
- **预加载**: 关键资源预加载、预取
- **压缩**: Gzip/Brotli 压缩、代码压缩

## 开发指南

### 目录结构
```
src/
├── components/          # 通用组件
│   ├── PluginCard/     # 插件卡片
│   ├── SearchBar/      # 搜索栏
│   ├── ReviewList/     # 评价列表
│   └── PaymentModal/   # 支付弹窗
├── pages/              # 页面组件
│   ├── Home/          # 首页
│   ├── PluginList/    # 插件列表
│   ├── PluginDetail/  # 插件详情
│   ├── UserCenter/    # 用户中心
│   └── Developer/     # 开发者中心
├── stores/             # 状态管理
│   ├── authStore.ts   # 认证状态
│   ├── pluginStore.ts # 插件状态
│   └── userStore.ts   # 用户状态
├── services/           # API 服务
│   ├── pluginAPI.ts   # 插件 API
│   ├── userAPI.ts     # 用户 API
│   └── paymentAPI.ts  # 支付 API
├── utils/              # 工具函数
├── hooks/              # 自定义 Hooks
├── types/              # TypeScript 类型定义
└── styles/             # 样式文件
```

### 开发规范
- 使用 TypeScript 严格模式
- 遵循 ESLint 和 Prettier 规则
- 组件使用函数式组件 + Hooks
- 状态管理使用 Zustand
- 样式使用 CSS Modules 或 styled-components
- API 调用统一使用 axios 实例

### 代码示例
```typescript
// 自定义 Hook 示例
const usePluginList = (initialParams: PluginSearchParams) => {
  const [plugins, setPlugins] = useState<Plugin[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  
  const fetchPlugins = useCallback(async (params: PluginSearchParams) => {
    setLoading(true)
    try {
      const response = await pluginAPI.getPlugins(params)
      setPlugins(response.data)
      setTotal(response.total)
    } catch (error) {
      message.error('获取插件列表失败')
    } finally {
      setLoading(false)
    }
  }, [])
  
  useEffect(() => {
    fetchPlugins(initialParams)
  }, [initialParams, fetchPlugins])
  
  return { plugins, loading, total, refetch: fetchPlugins }
}
```

## 故障排除

### 常见问题
1. **构建失败**: 检查 Node.js 版本和依赖安装
2. **API 调用失败**: 检查后端服务状态和网络连接
3. **支付失败**: 检查支付配置和网络环境
4. **图片加载失败**: 检查 CDN 配置和图片路径
5. **路由不工作**: 检查 Nginx 配置和 history 模式设置

### 调试方法
```bash
# 开发环境调试
pnpm dev --debug

# 查看网络请求
# 使用浏览器开发者工具的 Network 面板

# 查看状态管理
# 使用 Zustand DevTools

# 性能分析
pnpm build --analyze
```

### 监控和日志
```typescript
// 错误监控配置
import * as Sentry from '@sentry/react'

Sentry.init({
  dsn: 'your-sentry-dsn',
  environment: process.env.NODE_ENV,
  integrations: [
    new Sentry.BrowserTracing(),
  ],
  tracesSampleRate: 1.0,
})

// 用户行为追踪
import { analytics } from './utils/analytics'

analytics.track('plugin_view', {
  pluginId: plugin.id,
  pluginName: plugin.name,
  category: plugin.category
})
```