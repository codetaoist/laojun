import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User } from '@/types';
import { authService } from '@/services/auth';

interface AuthState {
  // 状态
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  expiresAt: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // 操作
  login: (username: string, password: string, captcha?: string, captchaKey?: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
  updateUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
  checkAuth: () => Promise<boolean>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // 初始状态
      user: null,
      token: null,
      refreshToken: null,
      expiresAt: null,
      isAuthenticated: false,
      isLoading: false,

      // 登录
      login: async (username: string, password: string, captcha?: string, captchaKey?: string) => {
        set({ isLoading: true });
        try {
          const response = await authService.login({ username, password, captcha, captcha_key: captchaKey });
          
          set({
            user: response.user,
            token: response.token,
            refreshToken: response.refreshToken ?? null,
            expiresAt: response.expiresAt ?? null,
            isAuthenticated: true,
            isLoading: false,
          });

          // 存储到 localStorage
          localStorage.setItem('token', response.token);
          if (response.refreshToken) {
            localStorage.setItem('refreshToken', response.refreshToken);
          } else {
            localStorage.removeItem('refreshToken');
          }
          if (response.expiresAt) {
            localStorage.setItem('expiresAt', response.expiresAt);
          } else {
            localStorage.removeItem('expiresAt');
          }
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      // 登出
      logout: async () => {
        try {
          await authService.logout();
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          // 清除状态和本地存储
          set({
            user: null,
            token: null,
            refreshToken: null,
            expiresAt: null,
            isAuthenticated: false,
            isLoading: false,
          });
          
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('expiresAt');
        }
      },

      // 刷新认证（占位，当前不启用）
      refreshAuth: async () => {
        const { refreshToken } = get();
        if (!refreshToken) {
          throw new Error('No refresh token available');
        }

        // 若后端提供刷新接口，启用下面逻辑
        // const response = await authService.refreshToken(refreshToken);
        // set({
        //   user: response.user,
        //   token: response.token,
        //   refreshToken: response.refreshToken ?? null,
        //   expiresAt: response.expiresAt ?? null,
        //   isAuthenticated: true,
        // });
        // localStorage.setItem('token', response.token);
        // if (response.refreshToken) {
        //   localStorage.setItem('refreshToken', response.refreshToken);
        // } else {
        //   localStorage.removeItem('refreshToken');
        // }
        // if (response.expiresAt) {
        //   localStorage.setItem('expiresAt', response.expiresAt);
        // } else {
        //   localStorage.removeItem('expiresAt');
        // }

        throw new Error('Refresh token not implemented');
      },

      // 更新用户信息
      updateUser: (user: User) => {
        set({ user });
      },

      // 设置加载状态
      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },

      // 检查认证状态：以获取当前用户成功与否为准（若已过期则直接登出）
      checkAuth: async () => {
        const { token, expiresAt } = get();
        if (!token) {
          return false;
        }

        if (expiresAt && Date.parse(expiresAt) <= Date.now()) {
          await get().logout();
          return false;
        }

        try {
          const user = await authService.getCurrentUser();
          set({ user, isAuthenticated: true });
          return true;
        } catch (error) {
          // 认证失败，清除状态
          await get().logout();
          return false;
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        user: state.user,
        expiresAt: state.expiresAt,
      }),
    }
  )
);