import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User } from '@/types';
import { message } from 'antd';
import { request } from '@/services/api';

export interface LoginCredentials {
  username: string;
  password: string;
  remember?: boolean;
  captcha?: string;
  captcha_key?: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  fullName?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

interface AuthStore extends AuthState {
  // 登录
  login: (credentials: LoginCredentials) => Promise<boolean>;
  
  // 注册
  register: (data: RegisterData) => Promise<boolean>;
  
  // 退出登录
  logout: () => void;
  
  // 刷新token
  refreshAuth: () => Promise<boolean>;
  
  // 获取用户信息
  fetchUserProfile: () => Promise<void>;
  
  // 更新用户信息
  updateProfile: (data: Partial<User>) => Promise<boolean>;
  
  // 修改密码
  changePassword: (oldPassword: string, newPassword: string) => Promise<boolean>;
  
  // 清除错误
  clearError: () => void;
  
  // 设置用户信息
  setUser: (user: User | null) => void;
  
  // 设置认证状态
  setAuth: (token: string, refreshToken: string, user: User) => void;
  
  // 重置状态
  reset: () => void;
}

const initialState: AuthState = {
  user: null,
  token: null,
  refreshToken: null,
  isAuthenticated: false,
  loading: false,
  error: null,
};

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      ...initialState,

      login: async (credentials) => {
        set({ loading: true, error: null });
        
        try {
          // 使用统一 axios 客户端调用登录API
          const payload: any = await request.post('/auth/login', credentials);

          // 设置认证信息（兼容 { message, data } 与直接数据返回）
          const { token, user } = (payload?.data || payload);
          
          set({
            user,
            token,
            refreshToken: null, // 暂不使用refresh token
            isAuthenticated: true,
            loading: false,
            error: null,
          });

          // 存储到localStorage，供拦截器自动附加
          localStorage.setItem('token', token);
          // localStorage.setItem('refreshToken', refreshToken); // 暂不使用

          message.success('登录成功');
          return true;
        } catch (error: any) {
          const errorMessage = error?.message || '登录失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
          return false;
        }
      },

      register: async (data) => {
        set({ loading: true, error: null });
        
        try {
          // 验证密码确认
          if (data.password !== data.confirmPassword) {
            throw new Error('两次输入的密码不一致');
          }

          // 使用统一 axios 客户端调用注册API
          await request.post('/auth/register', {
            username: data.username,
            email: data.email,
            password: data.password,
            confirm_password: data.confirmPassword,
            full_name: data.fullName,
          });

          set({ loading: false, error: null });
          message.success('注册成功，请登录');
          return true;
        } catch (error: any) {
          const errorMessage = error?.message || '注册失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
          return false;
        }
      },

      logout: () => {
        // 清除本地存储
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
        
        // 重置状态
        set({
          user: null,
          token: null,
          refreshToken: null,
          isAuthenticated: false,
          loading: false,
          error: null,
        });

        message.success('已退出登录');
      },

      refreshAuth: async () => {
        const { token } = get();
        
        if (!token) {
          return false;
        }

        try {
          // 使用后端提供的 /api/v1/auth/refresh，传入当前 token
          const resp: any = await request.post('/auth/refresh', { token });

          const data = resp?.data || resp;
          const newToken = data?.token || data?.token;

          // 如果返回了新的过期时间，也可以按需处理
          set({ token: newToken, isAuthenticated: true });

          localStorage.setItem('token', newToken);

          return true;
        } catch (error) {
          console.error('刷新token失败:', error);
          get().logout();
          return false;
        }
      },

      fetchUserProfile: async () => {
        const { token } = get();
        
        if (!token) {
          return;
        }

        try {
          // 使用统一 axios 客户端（拦截器会自动附加 Authorization）
          const resp: any = await request.get('/user/profile');
          const user = resp?.data || resp;
          set({ user });
        } catch (error) {
          console.error('获取用户信息失败:', error);
        }
      },

      updateProfile: async (profileData) => {
        const { token } = get();
        
        if (!token) {
          return false;
        }

        set({ loading: true, error: null });

        try {
          // 使用统一 axios 客户端更新用户信息
          const resp: any = await request.put('/user/profile', profileData);
          const updatedUser = resp?.data || resp;
          set({ user: updatedUser, loading: false });
          message.success('用户信息更新成功');
          return true;
        } catch (error: any) {
          const errorMessage = error?.message || '更新用户信息失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
          return false;
        }
      },

      changePassword: async (oldPassword, newPassword) => {
        const { token } = get();
        
        if (!token) {
          return false;
        }

        set({ loading: true, error: null });

        try {
          // 使用统一 axios 客户端修改密码
          await request.post('/user/change-password', { oldPassword, newPassword });

          set({ loading: false });
          message.success('密码修改成功');
          return true;
        } catch (error: any) {
          const errorMessage = error?.message || '修改密码失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
          return false;
        }
      },

      clearError: () => {
        set({ error: null });
      },

      setUser: (user) => {
        set({ user });
      },

      setAuth: (token, refreshToken, user) => {
        set({
          token,
          refreshToken,
          user,
          isAuthenticated: true,
        });
        
        localStorage.setItem('token', token);
        localStorage.setItem('refreshToken', refreshToken);
      },

      reset: () => {
        set(initialState);
      },
    }),
    {
      name: 'marketplace-auth-store',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// 初始化时检查本地存储的token
if (typeof window !== 'undefined') {
  const token = localStorage.getItem('token');
  
  if (token) {
    // 设置token并尝试获取用户信息
    useAuthStore.setState({ token, isAuthenticated: true });
    useAuthStore.getState().fetchUserProfile().catch(() => {
      // 如果获取用户信息失败，清除认证状态
      useAuthStore.getState().logout();
    });
  }
}

export default useAuthStore;