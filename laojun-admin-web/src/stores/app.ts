import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AppState {
  // 主题设置
  theme: 'light' | 'dark' | 'auto';
  
  // 侧边栏状态
  sidebarCollapsed: boolean;
  
  // 语言设置
  language: string;
  
  // 面包屑导航
  breadcrumbs: Array<{ title: string; path?: string }>;
  
  // 页面标题
  pageTitle: string;
  
  // 全局加载状态
  globalLoading: boolean;
  
  // 操作
  setTheme: (theme: 'light' | 'dark' | 'auto') => void;
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  setLanguage: (language: string) => void;
  setBreadcrumbs: (breadcrumbs: Array<{ title: string; path?: string }>) => void;
  setPageTitle: (title: string) => void;
  setGlobalLoading: (loading: boolean) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      // 初始状态
      theme: 'light',
      sidebarCollapsed: false,
      language: 'zh-CN',
      breadcrumbs: [],
      pageTitle: '太上老君管理后台',
      globalLoading: false,

      // 设置主题
      setTheme: (theme) => {
        set({ theme });
        
        // 应用主题到 document
        const root = document.documentElement;
        if (theme === 'dark') {
          root.classList.add('dark');
        } else if (theme === 'light') {
          root.classList.remove('dark');
        } else {
          // auto 模式，根据系统偏好设置
          const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
          if (prefersDark) {
            root.classList.add('dark');
          } else {
            root.classList.remove('dark');
          }
        }
      },

      // 切换侧边栏
      toggleSidebar: () => {
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed }));
      },

      // 设置侧边栏状态
      setSidebarCollapsed: (collapsed) => {
        set({ sidebarCollapsed: collapsed });
      },

      // 设置语言
      setLanguage: (language) => {
        set({ language });
        // 这里可以添加国际化逻辑
      },

      // 设置面包屑
      setBreadcrumbs: (breadcrumbs) => {
        set({ breadcrumbs });
      },

      // 设置页面标题
      setPageTitle: (title) => {
        set({ pageTitle: title });
        document.title = `${title} - 太上老君管理后台`;
      },

      // 设置全局加载状态
      setGlobalLoading: (loading) => {
        set({ globalLoading: loading });
      },
    }),
    {
      name: 'app-storage',
      partialize: (state) => ({
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed,
        language: state.language,
      }),
    }
  )
);