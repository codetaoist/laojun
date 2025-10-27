import { ApiClient } from './client';
import {
  User,
  Plugin,
  PluginCategory,
  LoginRequest,
  LoginResponse,
  PaginatedResponse,
  ListParams,
  Statistics,
  Notification,
} from '@/types';

// 认证服务
export class AuthService {
  constructor(private apiClient: ApiClient) {}

  // 登录
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    return this.apiClient.post<LoginResponse>('/auth/login', credentials);
  }

  // 登出
  async logout(): Promise<void> {
    return this.apiClient.post('/auth/logout');
  }

  // 刷新token
  async refreshToken(refreshToken: string): Promise<LoginResponse> {
    return this.apiClient.post<LoginResponse>('/auth/refresh', { refreshToken });
  }

  // 获取当前用户信息
  async getCurrentUser(): Promise<User> {
    return this.apiClient.get<User>('/auth/me');
  }

  // 修改密码
  async changePassword(oldPassword: string, newPassword: string): Promise<void> {
    return this.apiClient.post('/auth/change-password', {
      oldPassword,
      newPassword,
    });
  }

  // 忘记密码
  async forgotPassword(email: string): Promise<void> {
    return this.apiClient.post('/auth/forgot-password', { email });
  }

  // 重置密码
  async resetPassword(token: string, newPassword: string): Promise<void> {
    return this.apiClient.post('/auth/reset-password', {
      token,
      newPassword,
    });
  }
}

// 用户服务
export class UserService {
  constructor(private apiClient: ApiClient) {}

  // 获取用户列表
  async getUsers(params: ListParams = {}): Promise<PaginatedResponse<User>> {
    return this.apiClient.get<PaginatedResponse<User>>('/users', { params });
  }

  // 获取用户详情
  async getUser(id: string): Promise<User> {
    return this.apiClient.get<User>(`/users/${id}`);
  }

  // 创建用户
  async createUser(userData: Partial<User>): Promise<User> {
    return this.apiClient.post<User>('/users', userData);
  }

  // 更新用户
  async updateUser(id: string, userData: Partial<User>): Promise<User> {
    return this.apiClient.put<User>(`/users/${id}`, userData);
  }

  // 删除用户
  async deleteUser(id: string): Promise<void> {
    return this.apiClient.delete(`/users/${id}`);
  }

  // 批量删除用户
  async deleteUsers(ids: string[]): Promise<void> {
    return this.apiClient.post('/users/batch-delete', { ids });
  }

  // 更新用户状态
  async updateUserStatus(id: string, status: string): Promise<User> {
    return this.apiClient.patch<User>(`/users/${id}/status`, { status });
  }
}

// 插件服务
export class PluginService {
  constructor(private apiClient: ApiClient) {}

  // 获取插件列表
  async getPlugins(params: ListParams = {}): Promise<PaginatedResponse<Plugin>> {
    return this.apiClient.get<PaginatedResponse<Plugin>>('/plugins', { params });
  }

  // 获取插件详情
  async getPlugin(id: string): Promise<Plugin> {
    return this.apiClient.get<Plugin>(`/plugins/${id}`);
  }

  // 创建插件
  async createPlugin(pluginData: Partial<Plugin>): Promise<Plugin> {
    return this.apiClient.post<Plugin>('/plugins', pluginData);
  }

  // 更新插件
  async updatePlugin(id: string, pluginData: Partial<Plugin>): Promise<Plugin> {
    return this.apiClient.put<Plugin>(`/plugins/${id}`, pluginData);
  }

  // 删除插件
  async deletePlugin(id: string): Promise<void> {
    return this.apiClient.delete(`/plugins/${id}`);
  }

  // 上传插件文件
  async uploadPlugin(file: File, onProgress?: (progress: number) => void): Promise<{ url: string }> {
    return this.apiClient.upload<{ url: string }>('/plugins/upload', file, onProgress);
  }

  // 审核插件
  async reviewPlugin(id: string, action: 'approve' | 'reject', reason?: string): Promise<Plugin> {
    return this.apiClient.post<Plugin>(`/plugins/${id}/review`, { action, reason });
  }

  // 获取我的插件
  async getMyPlugins(params: ListParams = {}): Promise<PaginatedResponse<Plugin>> {
    return this.apiClient.get<PaginatedResponse<Plugin>>('/plugins/my', { params });
  }

  // 搜索插件
  async searchPlugins(query: string, params: ListParams = {}): Promise<PaginatedResponse<Plugin>> {
    return this.apiClient.get<PaginatedResponse<Plugin>>('/plugins/search', {
      params: { ...params, q: query },
    });
  }

  // 获取热门插件
  async getPopularPlugins(limit: number = 10): Promise<Plugin[]> {
    return this.apiClient.get<Plugin[]>('/plugins/popular', { params: { limit } });
  }

  // 获取最新插件
  async getLatestPlugins(limit: number = 10): Promise<Plugin[]> {
    return this.apiClient.get<Plugin[]>('/plugins/latest', { params: { limit } });
  }

  // 下载插件
  async downloadPlugin(id: string): Promise<{ downloadUrl: string }> {
    return this.apiClient.post<{ downloadUrl: string }>(`/plugins/${id}/download`);
  }

  // 收藏插件
  async favoritePlugin(id: string): Promise<void> {
    return this.apiClient.post(`/plugins/${id}/favorite`);
  }

  // 取消收藏
  async unfavoritePlugin(id: string): Promise<void> {
    return this.apiClient.delete(`/plugins/${id}/favorite`);
  }

  // 获取收藏列表
  async getFavorites(params: ListParams = {}): Promise<PaginatedResponse<Plugin>> {
    return this.apiClient.get<PaginatedResponse<Plugin>>('/plugins/favorites', { params });
  }
}

// 分类服务
export class CategoryService {
  constructor(private apiClient: ApiClient) {}

  // 获取分类列表
  async getCategories(): Promise<PluginCategory[]> {
    return this.apiClient.get<PluginCategory[]>('/categories');
  }

  // 获取分类详情
  async getCategory(id: string): Promise<PluginCategory> {
    return this.apiClient.get<PluginCategory>(`/categories/${id}`);
  }

  // 创建分类
  async createCategory(categoryData: Partial<PluginCategory>): Promise<PluginCategory> {
    return this.apiClient.post<PluginCategory>('/categories', categoryData);
  }

  // 更新分类
  async updateCategory(id: string, categoryData: Partial<PluginCategory>): Promise<PluginCategory> {
    return this.apiClient.put<PluginCategory>(`/categories/${id}`, categoryData);
  }

  // 删除分类
  async deleteCategory(id: string): Promise<void> {
    return this.apiClient.delete(`/categories/${id}`);
  }

  // 获取分类下的插件
  async getCategoryPlugins(id: string, params: ListParams = {}): Promise<PaginatedResponse<Plugin>> {
    return this.apiClient.get<PaginatedResponse<Plugin>>(`/categories/${id}/plugins`, { params });
  }
}

// 统计服务
export class StatisticsService {
  constructor(private apiClient: ApiClient) {}

  // 获取总体统计
  async getOverallStats(): Promise<Statistics> {
    return this.apiClient.get<Statistics>('/statistics/overall');
  }

  // 获取用户统计
  async getUserStats(period: string = '30d'): Promise<any> {
    return this.apiClient.get(`/statistics/users`, { params: { period } });
  }

  // 获取插件统计
  async getPluginStats(period: string = '30d'): Promise<any> {
    return this.apiClient.get(`/statistics/plugins`, { params: { period } });
  }

  // 获取下载统计
  async getDownloadStats(period: string = '30d'): Promise<any> {
    return this.apiClient.get(`/statistics/downloads`, { params: { period } });
  }
}

// 通知服务
export class NotificationService {
  constructor(private apiClient: ApiClient) {}

  // 获取通知列表
  async getNotifications(params: ListParams = {}): Promise<PaginatedResponse<Notification>> {
    return this.apiClient.get<PaginatedResponse<Notification>>('/notifications', { params });
  }

  // 标记通知为已读
  async markAsRead(id: string): Promise<void> {
    return this.apiClient.patch(`/notifications/${id}/read`);
  }

  // 标记所有通知为已读
  async markAllAsRead(): Promise<void> {
    return this.apiClient.patch('/notifications/read-all');
  }

  // 删除通知
  async deleteNotification(id: string): Promise<void> {
    return this.apiClient.delete(`/notifications/${id}`);
  }

  // 获取未读通知数量
  async getUnreadCount(): Promise<{ count: number }> {
    return this.apiClient.get<{ count: number }>('/notifications/unread-count');
  }
}

// 文件服务
export class FileService {
  constructor(private apiClient: ApiClient) {}

  // 上传文件
  async uploadFile(file: File, onProgress?: (progress: number) => void): Promise<{ url: string }> {
    return this.apiClient.upload<{ url: string }>('/files/upload', file, onProgress);
  }

  // 上传图片
  async uploadImage(file: File, onProgress?: (progress: number) => void): Promise<{ url: string }> {
    return this.apiClient.upload<{ url: string }>('/files/upload/image', file, onProgress);
  }

  // 删除文件
  async deleteFile(url: string): Promise<void> {
    return this.apiClient.delete('/files', { params: { url } });
  }
}

// 服务工厂类
export class ApiServices {
  public auth: AuthService;
  public user: UserService;
  public plugin: PluginService;
  public category: CategoryService;
  public statistics: StatisticsService;
  public notification: NotificationService;
  public file: FileService;

  constructor(apiClient: ApiClient) {
    this.auth = new AuthService(apiClient);
    this.user = new UserService(apiClient);
    this.plugin = new PluginService(apiClient);
    this.category = new CategoryService(apiClient);
    this.statistics = new StatisticsService(apiClient);
    this.notification = new NotificationService(apiClient);
    this.file = new FileService(apiClient);
  }
}

// 创建API服务实例
export const createApiServices = (apiClient: ApiClient): ApiServices => {
  return new ApiServices(apiClient);
};