import { create } from 'zustand';
import { Notification } from '../types';
import { NotificationService } from '../api';

// 通知状态接口
export interface NotificationState {
  // 状态
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;
  error: string | null;

  // 操作
  fetchNotifications: () => Promise<void>;
  markAsRead: (id: string) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  deleteNotification: (id: string) => Promise<void>;
  addNotification: (notification: Omit<Notification, 'id' | 'createdAt'>) => void;
  clearError: () => void;
  setLoading: (loading: boolean) => void;
}

// 创建通知 store
export const useNotificationStore = create<NotificationState>((set) => ({
  // 初始状态
  notifications: [],
  unreadCount: 0,
  isLoading: false,
  error: null,

  // 获取通知列表
  fetchNotifications: async () => {
    set({ isLoading: true, error: null });

    try {
      const notificationService = (window as any).__NOTIFICATION_SERVICE__ as NotificationService;
      if (!notificationService) {
        throw new Error('NotificationService not initialized');
      }

      const response = await notificationService.getNotifications({ limit: 50 });
      const unreadCountResponse = await notificationService.getUnreadCount();

      set({
        notifications: response.data,
        unreadCount: unreadCountResponse.count,
        isLoading: false,
        error: null,
      });
    } catch (error: any) {
      set({
        isLoading: false,
        error: error.message || '获取通知失败',
      });
      throw error;
    }
  },

  // 标记为已读
  markAsRead: async (id: string) => {
    try {
      const notificationService = (window as any).__NOTIFICATION_SERVICE__ as NotificationService;
      if (!notificationService) {
        throw new Error('NotificationService not initialized');
      }

      await notificationService.markAsRead(id);

      set(state => ({
        notifications: state.notifications.map(notification =>
          notification.id === id
            ? { ...notification, read: true }
            : notification
        ),
        unreadCount: Math.max(0, state.unreadCount - 1),
      }));
    } catch (error: any) {
      set({ error: error.message || '标记已读失败' });
      throw error;
    }
  },

  // 标记所有为已读
  markAllAsRead: async () => {
    try {
      const notificationService = (window as any).__NOTIFICATION_SERVICE__ as NotificationService;
      if (!notificationService) {
        throw new Error('NotificationService not initialized');
      }

      await notificationService.markAllAsRead();

      set(state => ({
        notifications: state.notifications.map(notification => ({
          ...notification,
          read: true,
        })),
        unreadCount: 0,
      }));
    } catch (error: any) {
      set({ error: error.message || '标记全部已读失败' });
      throw error;
    }
  },

  // 删除通知
  deleteNotification: async (id: string) => {
    try {
      const notificationService = (window as any).__NOTIFICATION_SERVICE__ as NotificationService;
      if (!notificationService) {
        throw new Error('NotificationService not initialized');
      }

      await notificationService.deleteNotification(id);

      set(state => {
        const notification = state.notifications.find(n => n.id === id);
        const wasUnread = notification && !notification.read;

        return {
          notifications: state.notifications.filter(n => n.id !== id),
          unreadCount: wasUnread ? Math.max(0, state.unreadCount - 1) : state.unreadCount,
        };
      });
    } catch (error: any) {
      set({ error: error.message || '删除通知失败' });
      throw error;
    }
  },

  // 添加本地通知
  addNotification: (notification: Omit<Notification, 'id' | 'createdAt'>) => {
    const newNotification: Notification = {
      ...notification,
      id: `local_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      createdAt: new Date().toISOString(),
    };

    set(state => ({
      notifications: [newNotification, ...state.notifications],
      unreadCount: notification.read ? state.unreadCount : state.unreadCount + 1,
    }));
  },

  // 清除错误
  clearError: () => {
    set({ error: null });
  },

  // 设置加载状态
  setLoading: (loading: boolean) => {
    set({ isLoading: loading });
  },
}));

// 通知 Hook
export const useNotification = () => {
  const store = useNotificationStore();
  
  return {
    ...store,
    // 便捷方法
    hasUnread: store.unreadCount > 0,
    getUnreadNotifications: () => store.notifications.filter(n => !n.read),
    getReadNotifications: () => store.notifications.filter(n => n.read),
  };
};

// 通知类型枚举
export const NotificationType = {
  INFO: 'info' as const,
  SUCCESS: 'success' as const,
  WARNING: 'warning' as const,
  ERROR: 'error' as const,
};

// 便捷的通知创建方法
export const createNotification = {
  info: (title: string, message: string) => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: NotificationType.INFO,
      title,
      message,
      read: false,
    });
  },

  success: (title: string, message: string) => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: NotificationType.SUCCESS,
      title,
      message,
      read: false,
    });
  },

  warning: (title: string, message: string) => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: NotificationType.WARNING,
      title,
      message,
      read: false,
    });
  },

  error: (title: string, message: string) => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: NotificationType.ERROR,
      title,
      message,
      read: false,
    });
  },
};

// 初始化通知
export const initializeNotifications = async () => {
  const { fetchNotifications } = useNotificationStore.getState();
  
  try {
    await fetchNotifications();
  } catch (error) {
    console.error('Failed to initialize notifications:', error);
  }
};

// 设置定时刷新通知
export const setupNotificationPolling = (interval: number = 30000) => {
  const { fetchNotifications } = useNotificationStore.getState();
  
  const poll = () => {
    fetchNotifications().catch(error => {
      console.error('Failed to poll notifications:', error);
    });
  };
  
  // 立即执行一次
  poll();
  
  // 设置定时器
  const timer = setInterval(poll, interval);
  
  return () => {
    clearInterval(timer);
  };
};