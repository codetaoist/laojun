import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { message } from 'antd';
import { downloadService } from '@/services/download';
import { DownloadRecord, InstallationStatus } from '@/types';

export interface DownloadTask {
  id: string;
  pluginId: string;
  pluginName: string;
  pluginIcon?: string;
  type: 'download' | 'install' | 'update' | 'uninstall';
  status: 'pending' | 'downloading' | 'installing' | 'completed' | 'failed' | 'paused' | 'cancelled';
  progress: number;
  message: string;
  filePath?: string;
  error?: string;
  startTime: Date;
  endTime?: Date;
  fileSize?: number;
  downloadSpeed?: number;
}

export interface DownloadState {
  // 当前任务
  tasks: DownloadTask[];
  // 下载历史
  downloadHistory: DownloadRecord[];
  // 安装状态
  installationStatuses: Record<string, InstallationStatus>;
  // 加载状态
  loading: boolean;
  error: string | null;
  
  // 任务管理
  addTask: (task: Omit<DownloadTask, 'id' | 'startTime'>) => string;
  updateTask: (taskId: string, updates: Partial<DownloadTask>) => void;
  removeTask: (taskId: string) => void;
  clearCompletedTasks: () => void;
  cancelTask: (taskId: string) => void;
  retryTask: (taskId: string) => void;
  
  // 下载操作
  downloadPlugin: (pluginId: string, pluginName: string, pluginIcon?: string) => Promise<void>;
  installPlugin: (pluginId: string, pluginName: string, filePath: string, pluginIcon?: string) => Promise<void>;
  updatePlugin: (pluginId: string, pluginName: string, pluginIcon?: string) => Promise<void>;
  uninstallPlugin: (pluginId: string, pluginName: string) => Promise<void>;
  
  // 状态查询
  getTasksByPlugin: (pluginId: string) => DownloadTask[];
  getActiveTasksCount: () => number;
  isPluginDownloading: (pluginId: string) => boolean;
  isPluginInstalling: (pluginId: string) => boolean;
  getInstallationStatus: (pluginId: string) => InstallationStatus;
  
  // 历史记录
  loadDownloadHistory: () => void;
  clearDownloadHistory: () => void;
  
  // 工具方法
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
  reset: () => void;
}

const initialState = {
  tasks: [],
  downloadHistory: [],
  installationStatuses: {},
  loading: false,
  error: null,
};

export const useDownloadStore = create<DownloadState>()(
  persist(
    (set, get) => ({
      ...initialState,

      // 任务管理
      addTask: (taskData) => {
        const taskId = `${taskData.type}-${taskData.pluginId}-${Date.now()}`;
        const task: DownloadTask = {
          ...taskData,
          id: taskId,
          startTime: new Date(),
        };
        
        set((state) => ({
          tasks: [...state.tasks, task],
        }));
        
        return taskId;
      },

      updateTask: (taskId, updates) => {
        set((state) => ({
          tasks: state.tasks.map(task =>
            task.id === taskId ? { ...task, ...updates } : task
          ),
        }));
      },

      removeTask: (taskId) => {
        set((state) => ({
          tasks: state.tasks.filter(task => task.id !== taskId),
        }));
      },

      clearCompletedTasks: () => {
        set((state) => ({
          tasks: state.tasks.filter(task => 
            task.status !== 'completed' && task.status !== 'failed' && task.status !== 'cancelled'
          ),
        }));
      },

      cancelTask: (taskId) => {
        const { tasks, updateTask } = get();
        const task = tasks.find(t => t.id === taskId);
        
        if (task) {
          if (task.type === 'download' && task.status === 'downloading') {
            downloadService.cancelDownload(task.pluginId);
          }
          
          updateTask(taskId, {
            status: 'cancelled',
            message: '已取消',
            endTime: new Date(),
          });
        }
      },

      retryTask: (taskId) => {
        const { tasks, removeTask, downloadPlugin, installPlugin, updatePlugin } = get();
        const task = tasks.find(t => t.id === taskId);
        
        if (task) {
          removeTask(taskId);
          
          switch (task.type) {
            case 'download':
              downloadPlugin(task.pluginId, task.pluginName, task.pluginIcon);
              break;
            case 'install':
              if (task.filePath) {
                installPlugin(task.pluginId, task.pluginName, task.filePath, task.pluginIcon);
              }
              break;
            case 'update':
              updatePlugin(task.pluginId, task.pluginName, task.pluginIcon);
              break;
          }
        }
      },

      // 下载操作
      downloadPlugin: async (pluginId, pluginName, pluginIcon) => {
        const { addTask, updateTask, setError } = get();
        
        try {
          const taskId = addTask({
            pluginId,
            pluginName,
            pluginIcon,
            type: 'download',
            status: 'pending',
            progress: 0,
            message: '准备下载...',
          });

          updateTask(taskId, { status: 'downloading', message: '正在下载...' });

          const filePath = await downloadService.downloadPlugin({
            pluginId,
            onProgress: (progress, message, speed, fileSize) => {
              updateTask(taskId, {
                progress,
                message: message || `下载中... ${progress}%`,
                downloadSpeed: speed,
                fileSize,
              });
            },
            onComplete: (filePath) => {
              updateTask(taskId, {
                status: 'completed',
                progress: 100,
                message: '下载完成',
                filePath,
                endTime: new Date(),
              });
              message.success(`${pluginName} 下载完成`);
            },
            onError: (error) => {
              updateTask(taskId, {
                status: 'failed',
                message: '下载失败',
                error: error.message,
                endTime: new Date(),
              });
              message.error(`${pluginName} 下载失败: ${error.message}`);
            },
          });

        } catch (error) {
          setError(error instanceof Error ? error.message : '下载失败');
          message.error(`${pluginName} 下载失败`);
        }
      },

      installPlugin: async (pluginId, pluginName, filePath, pluginIcon) => {
        const { addTask, updateTask, setError } = get();
        
        try {
          const taskId = addTask({
            pluginId,
            pluginName,
            pluginIcon,
            type: 'install',
            status: 'pending',
            progress: 0,
            message: '准备安装...',
            filePath,
          });

          updateTask(taskId, { status: 'installing', message: '正在安装...' });

          await downloadService.installPlugin({
            pluginId,
            filePath,
            onProgress: (status, message) => {
              const progressMap: Record<InstallationStatus, number> = {
                not_installed: 0,
                installing: 50,
                installed: 100,
                failed: 0,
                updating: 75,
              };

              updateTask(taskId, {
                progress: progressMap[status],
                message: message || `安装中... ${status}`,
              });

              // 更新安装状态
              set((state) => ({
                installationStatuses: {
                  ...state.installationStatuses,
                  [pluginId]: status,
                },
              }));
            },
            onComplete: () => {
              updateTask(taskId, {
                status: 'completed',
                progress: 100,
                message: '安装完成',
                endTime: new Date(),
              });
              
              set((state) => ({
                installationStatuses: {
                  ...state.installationStatuses,
                  [pluginId]: 'installed',
                },
              }));
              
              message.success(`${pluginName} 安装完成`);
            },
            onError: (error) => {
              updateTask(taskId, {
                status: 'failed',
                message: '安装失败',
                error: error.message,
                endTime: new Date(),
              });
              
              set((state) => ({
                installationStatuses: {
                  ...state.installationStatuses,
                  [pluginId]: 'failed',
                },
              }));
              
              message.error(`${pluginName} 安装失败: ${error.message}`);
            },
          });

        } catch (error) {
          setError(error instanceof Error ? error.message : '安装失败');
          message.error(`${pluginName} 安装失败`);
        }
      },

      updatePlugin: async (pluginId, pluginName, pluginIcon) => {
        const { addTask, updateTask, setError } = get();
        
        try {
          const taskId = addTask({
            pluginId,
            pluginName,
            pluginIcon,
            type: 'update',
            status: 'pending',
            progress: 0,
            message: '准备更新...',
          });

          updateTask(taskId, { status: 'downloading', message: '正在更新...' });

          await downloadService.updatePlugin({
            pluginId,
            onProgress: (progress, message) => {
              updateTask(taskId, {
                progress,
                message: message || `更新中... ${progress}%`,
              });
            },
            onComplete: () => {
              updateTask(taskId, {
                status: 'completed',
                progress: 100,
                message: '更新完成',
                endTime: new Date(),
              });
              message.success(`${pluginName} 更新完成`);
            },
            onError: (error) => {
              updateTask(taskId, {
                status: 'failed',
                message: '更新失败',
                error: error.message,
                endTime: new Date(),
              });
              message.error(`${pluginName} 更新失败: ${error.message}`);
            },
          });

        } catch (error) {
          setError(error instanceof Error ? error.message : '更新失败');
          message.error(`${pluginName} 更新失败`);
        }
      },

      uninstallPlugin: async (pluginId, pluginName) => {
        const { addTask, updateTask, setError } = get();
        
        try {
          const taskId = addTask({
            pluginId,
            pluginName,
            type: 'uninstall',
            status: 'pending',
            progress: 0,
            message: '准备卸载...',
          });

          updateTask(taskId, { status: 'installing', message: '正在卸载...' });

          await downloadService.uninstallPlugin({
            pluginId,
            onProgress: (progress, message) => {
              updateTask(taskId, {
                progress,
                message: message || `卸载中... ${progress}%`,
              });
            },
            onComplete: () => {
              updateTask(taskId, {
                status: 'completed',
                progress: 100,
                message: '卸载完成',
                endTime: new Date(),
              });
              
              set((state) => ({
                installationStatuses: {
                  ...state.installationStatuses,
                  [pluginId]: 'not_installed',
                },
              }));
              
              message.success(`${pluginName} 卸载完成`);
            },
            onError: (error) => {
              updateTask(taskId, {
                status: 'failed',
                message: '卸载失败',
                error: error.message,
                endTime: new Date(),
              });
              message.error(`${pluginName} 卸载失败: ${error.message}`);
            },
          });

        } catch (error) {
          setError(error instanceof Error ? error.message : '卸载失败');
          message.error(`${pluginName} 卸载失败`);
        }
      },

      // 状态查询
      getTasksByPlugin: (pluginId) => {
        return get().tasks.filter(task => task.pluginId === pluginId);
      },

      getActiveTasksCount: () => {
        return get().tasks.filter(task => 
          task.status === 'downloading' || task.status === 'installing'
        ).length;
      },

      isPluginDownloading: (pluginId) => {
        return get().tasks.some(task => 
          task.pluginId === pluginId && 
          task.type === 'download' && 
          task.status === 'downloading'
        );
      },

      isPluginInstalling: (pluginId) => {
        return get().tasks.some(task => 
          task.pluginId === pluginId && 
          (task.type === 'install' || task.type === 'update' || task.type === 'uninstall') && 
          task.status === 'installing'
        );
      },

      getInstallationStatus: (pluginId) => {
        return get().installationStatuses[pluginId] || 'not_installed';
      },

      // 历史记录
      loadDownloadHistory: () => {
        try {
          const history = downloadService.getDownloadHistory();
          set({ downloadHistory: history });
        } catch (error) {
          console.error('加载下载历史失败:', error);
        }
      },

      clearDownloadHistory: () => {
        set({ downloadHistory: [] });
        message.success('下载历史已清除');
      },

      // 工具方法
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),
      clearError: () => set({ error: null }),
      
      reset: () => set(initialState),
    }),
    {
      name: 'marketplace-download-store',
      partialize: (state) => ({
        downloadHistory: state.downloadHistory,
        installationStatuses: state.installationStatuses,
      }),
    }
  )
);