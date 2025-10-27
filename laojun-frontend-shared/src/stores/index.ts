// 导出认证相关
export { useAuthStore, useAuth, initializeAuth, checkTokenExpiration } from './auth';
export type { AuthState } from './auth';

// 导出应用相关
export { useAppStore, useApp, initializeApp } from './app';
export type { AppState } from './app';

// 导出通知相关
export { 
  useNotificationStore, 
  useNotification, 
  createNotification,
  initializeNotifications,
  setupNotificationPolling,
  NotificationType
} from './notification';
export type { 
  NotificationState
} from './notification';

// 导入所需的依赖
import { useAuth } from './auth';
import { useApp } from './app';
import { useNotification } from './notification';
import { useAuthStore } from './auth';
import { useNotificationStore } from './notification';
import { initializeApp } from './app';
import { initializeAuth } from './auth';
import { initializeNotifications } from './notification';

// 组合 Hook - 用于初始化所有 stores
export const useStores = () => {
  const auth = useAuth();
  const app = useApp();
  const notification = useNotification();

  return {
    auth,
    app,
    notification,
  };
};

// 初始化所有 stores
export const initializeStores = async () => {
  // 初始化应用状态
  const cleanupApp = initializeApp();
  
  // 初始化认证状态
  await initializeAuth();
  
  // 初始化通知（如果用户已登录）
  const authStore = useAuthStore.getState();
  if (authStore.isAuthenticated) {
    await initializeNotifications();
  }
  
  return () => {
    cleanupApp();
  };
};

// 重置所有 stores（用于登出等场景）
export const resetStores = () => {
  // 重置认证状态
  const authStore = useAuthStore.getState();
  authStore.logout();
  
  // 重置通知状态
  useNotificationStore.setState({
    notifications: [],
    unreadCount: 0,
    isLoading: false,
    error: null,
  });
  
  // 应用状态保持不变（主题、语言等用户偏好）
};