// 用户相关类型
export interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  avatar?: string;
  status: 'active' | 'inactive' | 'banned';
  roles: Role[];
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
}

// 角色相关类型
export interface Role {
  id: string;
  name: string;
  displayName?: string;
  description?: string;
  permissions: Permission[];
  isSystem: boolean;
  createdAt: string;
  updatedAt: string;
}

// 权限相关类型
export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
  isSystem: boolean;
}

// 菜单相关类型
export interface MenuItem {
  id: string;
  name: string;
  path?: string;
  icon?: string;
  component?: string;
  parentId?: string;
  order: number;
  visible: boolean;
  permissions: string[];
  children?: MenuItem[];
}

// 菜单管理类型（与后端模型对齐）
export interface Menu {
  id: string;
  title: string;
  path?: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sortOrder: number;
  isHidden: boolean;
  children?: Menu[];
  createdAt: string;
  updatedAt: string;
}

// 插件相关类型
export interface Plugin {
  id: string;
  name: string;
  version: string;
  description?: string;
  author?: string;
  status: 'installed' | 'active' | 'inactive' | 'error';
  dependencies?: string[];
  permissions?: string[];
  routes?: PluginRoute[];
  menus?: PluginMenu[];
  components?: PluginComponent[];
  installedAt: string;
  updatedAt: string;
}

export interface PluginRoute {
  path: string;
  component: string;
  permissions?: string[];
}

export interface PluginMenu {
  id: string;
  label: string;
  icon?: string;
  path?: string;
  component?: string;
  permissions?: string[];
  order?: number;
}

export interface PluginComponent {
  name: string;
  path: string;
  permissions?: string[];
}

// API 响应类型
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  code?: string;
  timestamp: number;
}

export interface PaginatedResponse<T = any> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 查询参数类型
export interface QueryParams {
  page?: number;
  pageSize?: number;
  search?: string;
  sort?: string;
  order?: 'asc' | 'desc';
  [key: string]: any;
}

// 登录相关类型
export interface LoginRequest {
  username: string;
  password: string;
  captcha?: string;
  captcha_key?: string;
}

export interface LoginResponse {
  token: string;
  user: User;
  expiresAt: string;
  refreshToken?: string;
}

// 系统配置类型
export interface SystemConfig {
  siteName: string;
  siteDescription?: string;
  logo?: string;
  favicon?: string;
  theme: 'light' | 'dark' | 'auto';
  language: string;
  timezone: string;
  enableRegistration: boolean;
  enableCaptcha: boolean;
  sessionTimeout: number;
  maxLoginAttempts: number;
  lockoutDuration: number;
}

// 系统状态类型
export interface SystemStatus {
  cpu: {
    usage: number;
    cores: number;
  };
  memory: {
    used: number;
    total: number;
    usage: number;
  };
  disk: {
    used: number;
    total: number;
    usage: number;
  };
  network: {
    bytesIn: number;
    bytesOut: number;
  };
  uptime: number;
  timestamp: number;
}