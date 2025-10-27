import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { User, LoginCredentials, LoginResponse } from '../types';
import { AuthService } from '../api';

// 认证状态接口
export interface AuthState {
  // 状态
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // 操作
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
  getCurrentUser: () => Promise<void>;
  updateUser: (userData: Partial<User>) => void;
  clearError: () => void;
  setLoading: (loading: boolean) => void;
}

// 创建认证 store
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // 初始状态
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // 登录
      login: async (credentials: LoginCredentials) => {
        set({ isLoading: true, error: null });
        
        try {
          // 这里需要注入 AuthService 实例
          const authService = (window as any).__AUTH_SERVICE__ as AuthService;
          if (!authService) {
            throw new Error('AuthService not initialized');
          }

          const response: LoginResponse = await authService.login(credentials);
          
          set({
            user: response.user,
            token: response.token,
            refreshToken: response.refreshToken,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });

          // 存储 token 到 localStorage
          localStorage.setItem('token', response.token);
          localStorage.setItem('refreshToken', response.refreshToken);
        } catch (error: any) {
          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            error: error.message || '登录失败',
          });
          throw error;
        }
      },

      // 登出
      logout: async () => {
        set({ isLoading: true });
        
        try {
          const authService = (window as any).__AUTH_SERVICE__ as AuthService;
          if (authService) {
            await authService.logout();
          }
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          // 清除状态
          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,
          });

          // 清除本地存储
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
        }
      },

      // 刷新认证
      refreshAuth: async () => {
        const { refreshToken } = get();
        if (!refreshToken) {
          throw new Error('No refresh token available');
        }

        set({ isLoading: true, error: null });

        try {
          const authService = (window as any).__AUTH_SERVICE__ as AuthService;
          if (!authService) {
            throw new Error('AuthService not initialized');
          }

          const response: LoginResponse = await authService.refreshToken(refreshToken);
          
          set({
            user: response.user,
            token: response.token,
            refreshToken: response.refreshToken,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });

          // 更新本地存储
          localStorage.setItem('token', response.token);
          localStorage.setItem('refreshToken', response.refreshToken);
        } catch (error: any) {
          // 刷新失败，清除认证状态
          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            error: error.message || '认证已过期，请重新登录',
          });

          // 清除本地存储
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          
          throw error;
        }
      },

      // 获取当前用户信息
      getCurrentUser: async () => {
        const { token } = get();
        if (!token) {
          return;
        }

        set({ isLoading: true, error: null });

        try {
          const authService = (window as any).__AUTH_SERVICE__ as AuthService;
          if (!authService) {
            throw new Error('AuthService not initialized');
          }

          const user = await authService.getCurrentUser();
          
          set({
            user,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
        } catch (error: any) {
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
            error: error.message || '获取用户信息失败',
          });
          throw error;
        }
      },

      // 更新用户信息
      updateUser: (userData: Partial<User>) => {
        const { user } = get();
        if (user) {
          set({
            user: { ...user, ...userData },
          });
        }
      },

      // 清除错误
      clearError: () => {
        set({ error: null });
      },

      // 设置加载状态
      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// 认证 Hook
export const useAuth = () => {
  const store = useAuthStore();
  
  return {
    ...store,
    // 便捷方法
    isAdmin: store.user?.role === 'admin',
    isDeveloper: store.user?.role === 'developer',
    isUser: store.user?.role === 'user',
  };
};

// 初始化认证状态
export const initializeAuth = async () => {
  const { token, getCurrentUser, refreshAuth } = useAuthStore.getState();
  
  if (token) {
    try {
      // 尝试获取当前用户信息
      await getCurrentUser();
    } catch (error) {
      // 如果获取失败，尝试刷新 token
      try {
        await refreshAuth();
      } catch (refreshError) {
        // 刷新也失败，清除认证状态
        console.error('Failed to refresh auth:', refreshError);
      }
    }
  }
};

// 检查 token 是否即将过期
export const checkTokenExpiration = () => {
  const { token, refreshAuth } = useAuthStore.getState();
  
  if (!token) {
    return;
  }

  try {
    // 解析 JWT token
    const payload = JSON.parse(atob(token.split('.')[1]));
    const currentTime = Date.now() / 1000;
    const timeUntilExpiry = payload.exp - currentTime;

    // 如果 token 在 5 分钟内过期，尝试刷新
    if (timeUntilExpiry < 300) {
      refreshAuth().catch(error => {
        console.error('Failed to refresh token:', error);
      });
    }
  } catch (error) {
    console.error('Failed to parse token:', error);
  }
};

// 设置定时检查 token 过期
export const setupTokenExpirationCheck = () => {
  // 每分钟检查一次
  setInterval(checkTokenExpiration, 60000);
};