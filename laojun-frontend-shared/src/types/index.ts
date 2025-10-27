// 通用API响应类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
  timestamp: string;
}

// 分页响应类型
export interface PaginatedResponse<T = any> {
  data: T[];
  meta: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}

// 用户相关类型
export interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  role: UserRole;
  status: UserStatus;
  createdAt: string;
  updatedAt: string;
}

export enum UserRole {
  ADMIN = 'admin',
  DEVELOPER = 'developer',
  USER = 'user',
}

export enum UserStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  BANNED = 'banned',
}

// 认证相关类型
export interface LoginRequest {
  username: string;
  password: string;
  captcha?: string;
}

// 登录凭据类型别名
export type LoginCredentials = LoginRequest;

export interface LoginResponse {
  token: string;
  refreshToken: string;
  user: User;
  expiresIn: number;
}

// 插件相关类型
export interface Plugin {
  id: string;
  name: string;
  description: string;
  version: string;
  author: string;
  authorId: string;
  category: PluginCategory;
  tags: string[];
  downloadCount: number;
  rating: number;
  reviewCount: number;
  price: number;
  isPaid: boolean;
  status: PluginStatus;
  screenshots: string[];
  icon?: string;
  readme?: string;
  changelog?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PluginCategory {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  parentId?: string;
  sortOrder: number;
}

export enum PluginStatus {
  DRAFT = 'draft',
  PENDING = 'pending',
  APPROVED = 'approved',
  REJECTED = 'rejected',
  SUSPENDED = 'suspended',
}

// 路由相关类型
export interface RouteConfig {
  path: string;
  element: React.ComponentType;
  children?: RouteConfig[];
  meta?: RouteMeta;
}

export interface RouteMeta {
  title?: string;
  requireAuth?: boolean;
  roles?: UserRole[];
  icon?: string;
  hideInMenu?: boolean;
}

// 主题配置
export interface ThemeConfig {
  primaryColor: string;
  borderRadius: number;
  colorBgContainer: string;
}

// 应用配置类型
export interface AppConfig {
  name: string;
  version: string;
  apiBaseUrl: string;
  enableDevTools: boolean;
  theme: ThemeConfig;
}

// 错误类型
export interface AppError {
  code: string;
  message: string;
  details?: any;
  timestamp: string;
}

// 通用列表查询参数
export interface ListParams {
  page?: number;
  limit?: number;
  search?: string;
  sort?: string;
  order?: 'asc' | 'desc';
  [key: string]: any;
}

// 文件上传相关
export interface UploadFile {
  uid: string;
  name: string;
  status: 'uploading' | 'done' | 'error';
  url?: string;
  response?: any;
  error?: any;
}

// 通知类型
export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  read: boolean;
  createdAt: string;
}

// 统计数据类型
export interface Statistics {
  totalUsers: number;
  totalPlugins: number;
  totalDownloads: number;
  totalRevenue: number;
  growthRate: {
    users: number;
    plugins: number;
    downloads: number;
    revenue: number;
  };
}

// 所有类型都在当前文件中定义，无需额外导出