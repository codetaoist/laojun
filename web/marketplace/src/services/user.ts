import { request } from './api';
import { User, PaginatedResponse, Plugin, Purchase } from '@/types';

// 用户资料更新请求
export interface UpdateProfileRequest {
  full_name?: string;
  avatar?: string;
  bio?: string;
}

// 修改密码请求
export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

// 用户统计信息
export interface UserStats {
  user_id: string;
  plugins_count: number;
  favorites_count: number;
  reviews_count: number;
}

// 用户相关 API
export const userService = {
  // 获取用户资料
  getProfile: (): Promise<User> => {
    return request.get('/user/profile');
  },

  // 更新用户资料
  updateProfile: (data: UpdateProfileRequest): Promise<User> => {
    return request.put('/user/profile', data);
  },

  // 修改密码
  changePassword: (data: ChangePasswordRequest): Promise<void> => {
    return request.post('/user/change-password', data);
  },

  // 获取用户统计信息
  getUserStats: (): Promise<UserStats> => {
    return request.get('/user/stats');
  },

  // 退出登录
  logout: (): Promise<void> => {
    return request.post('/user/logout');
  },

  // 获取用户收藏的插件
  getFavoritePlugins: (page?: number, limit?: number): Promise<PaginatedResponse<Plugin>> => {
    return request.get('/user/favorites', { 
      params: { page, limit } 
    });
  },

  // 获取用户购买记录
  getPurchases: (page?: number, limit?: number): Promise<PaginatedResponse<Purchase>> => {
    return request.get('/user/purchases', { 
      params: { page, limit } 
    });
  },

  // 切换插件收藏状态
  toggleFavorite: (pluginId: string): Promise<{ is_favorited: boolean; message: string }> => {
    return request.post(`/plugins/${pluginId}/favorite`);
  },

  // 购买插件
  purchasePlugin: (pluginId: string): Promise<{ message: string }> => {
    return request.post(`/plugins/${pluginId}/purchase`);
  },
};

export default userService;