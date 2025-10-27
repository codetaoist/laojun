import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { AppState } from '@/types';

interface AppStore extends AppState {
  // 设置主题
  setTheme: (theme: AppState['theme']) => void;
  
  // 设置语言
  setLanguage: (language: string) => void;
  
  // 添加搜索历史
  addSearchHistory: (query: string) => void;
  
  // 清除搜索历史
  clearSearchHistory: () => void;
  
  // 设置视图模式
  setViewMode: (mode: AppState['viewMode']) => void;
  
  // 设置每页显示数量
  setPageSize: (size: number) => void;
  
  // 重置应用状态
  reset: () => void;
}

const initialState: AppState = {
  theme: 'auto',
  language: 'zh-CN',
  searchHistory: [],
  viewMode: 'grid',
  pageSize: 20,
};

export const useAppStore = create<AppStore>()(
  persist(
    (set, get) => ({
      ...initialState,

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

      setLanguage: (language) => {
        set({ language });
      },

      addSearchHistory: (query) => {
        const { searchHistory } = get();
        const trimmedQuery = query.trim();
        
        if (!trimmedQuery || searchHistory.includes(trimmedQuery)) {
          return;
        }

        const newHistory = [trimmedQuery, ...searchHistory.slice(0, 9)]; // 保留最近 10 条
        set({ searchHistory: newHistory });
      },

      clearSearchHistory: () => {
        set({ searchHistory: [] });
      },

      setViewMode: (viewMode) => {
        set({ viewMode });
      },

      setPageSize: (pageSize) => {
        set({ pageSize });
      },

      reset: () => {
        set(initialState);
      },
    }),
    {
      name: 'marketplace-app-store',
      partialize: (state) => ({
        theme: state.theme,
        language: state.language,
        searchHistory: state.searchHistory,
        viewMode: state.viewMode,
        pageSize: state.pageSize,
      }),
    }
  )
);

// 监听系统主题变化
if (typeof window !== 'undefined') {
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  
  const handleThemeChange = () => {
    const { theme, setTheme } = useAppStore.getState();
    if (theme === 'auto') {
      setTheme('auto'); // 触发主题更新
    }
  };

  mediaQuery.addEventListener('change', handleThemeChange);
}

export default useAppStore;