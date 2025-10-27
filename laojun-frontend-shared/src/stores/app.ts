import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { AppConfig, ThemeConfig } from '../types';

// 应用状态接口
export interface AppState {
  // 配置
  config: AppConfig;
  
  // UI 状态
  sidebarCollapsed: boolean;
  loading: boolean;
  
  // 主题
  theme: 'light' | 'dark';
  themeConfig: ThemeConfig;
  
  // 语言
  locale: string;
  
  // 操作
  updateConfig: (config: Partial<AppConfig>) => void;
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  setLoading: (loading: boolean) => void;
  setTheme: (theme: 'light' | 'dark') => void;
  updateThemeConfig: (config: Partial<ThemeConfig>) => void;
  setLocale: (locale: string) => void;
}

// 默认配置
const defaultConfig: AppConfig = {
  name: 'Laojun Platform',
  version: '1.0.0',
  apiBaseUrl: (typeof process !== 'undefined' && process.env?.REACT_APP_API_BASE_URL) || 'http://localhost:8080',
  enableDevTools: typeof process !== 'undefined' && process.env?.NODE_ENV === 'development',
  theme: {
    primaryColor: '#1890ff',
    borderRadius: 6,
    colorBgContainer: '#ffffff',
  },
};

// 创建应用 store
export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      // 初始状态
      config: defaultConfig,
      sidebarCollapsed: false,
      loading: false,
      theme: 'light',
      themeConfig: defaultConfig.theme,
      locale: 'zh-CN',

      // 更新配置
      updateConfig: (newConfig: Partial<AppConfig>) => {
        set(state => ({
          config: { ...state.config, ...newConfig },
        }));
      },

      // 切换侧边栏
      toggleSidebar: () => {
        set(state => ({
          sidebarCollapsed: !state.sidebarCollapsed,
        }));
      },

      // 设置侧边栏状态
      setSidebarCollapsed: (collapsed: boolean) => {
        set({ sidebarCollapsed: collapsed });
      },

      // 设置加载状态
      setLoading: (loading: boolean) => {
        set({ loading });
      },

      // 设置主题
      setTheme: (theme: 'light' | 'dark') => {
        set({ theme });
        
        // 更新 HTML 类名
        document.documentElement.setAttribute('data-theme', theme);
      },

      // 更新主题配置
      updateThemeConfig: (newThemeConfig: Partial<ThemeConfig>) => {
        set(state => ({
          themeConfig: { ...state.themeConfig, ...newThemeConfig },
          config: {
            ...state.config,
            theme: { ...state.config.theme, ...newThemeConfig },
          },
        }));
      },

      // 设置语言
      setLocale: (locale: string) => {
        set({ locale });
      },
    }),
    {
      name: 'app-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
        themeConfig: state.themeConfig,
        locale: state.locale,
      }),
    }
  )
);

// 应用 Hook
export const useApp = () => {
  const store = useAppStore();
  
  return {
    ...store,
    // 便捷方法
    isDarkTheme: store.theme === 'dark',
    isLightTheme: store.theme === 'light',
  };
};

// 初始化应用状态
export const initializeApp = () => {
  const { theme } = useAppStore.getState();
  
  // 设置初始主题
  document.documentElement.setAttribute('data-theme', theme);
  
  // 监听系统主题变化
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handleThemeChange = (e: MediaQueryListEvent) => {
    const { setTheme } = useAppStore.getState();
    setTheme(e.matches ? 'dark' : 'light');
  };
  
  mediaQuery.addEventListener('change', handleThemeChange);
  
  return () => {
    mediaQuery.removeEventListener('change', handleThemeChange);
  };
};