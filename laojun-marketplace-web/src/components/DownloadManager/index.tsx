import React, { useState, useEffect } from 'react';
import {
  Drawer,
  List,
  Progress,
  Button,
  Typography,
  Space,
  Tag,
  Avatar,
  Divider,
  Empty,
  Tooltip,
  Modal,
} from 'antd';
import {
  DownloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  DeleteOutlined,
  FolderOpenOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { downloadService, DownloadProgressCallback, InstallationCallback } from '@/services/download';
import { DownloadRecord, InstallationStatus } from '@/types';

const { Title, Text } = Typography;

interface DownloadTask {
  id: string;
  pluginId: string;
  pluginName: string;
  pluginIcon?: string;
  type: 'download' | 'install' | 'update';
  status: 'pending' | 'downloading' | 'installing' | 'completed' | 'failed' | 'paused';
  progress: number;
  message: string;
  filePath?: string;
  error?: string;
  startTime: Date;
  endTime?: Date;
}

interface DownloadManagerProps {
  visible: boolean;
  onClose: () => void;
}

const DownloadManager: React.FC<DownloadManagerProps> = ({ visible, onClose }) => {
  const [tasks, setTasks] = useState<DownloadTask[]>([]);
  const [downloadHistory, setDownloadHistory] = useState<DownloadRecord[]>([]);
  const [activeTab, setActiveTab] = useState<'current' | 'history'>('current');

  useEffect(() => {
    if (visible) {
      loadDownloadHistory();
    }
  }, [visible]);

  const loadDownloadHistory = () => {
    const history = downloadService.getDownloadHistory();
    setDownloadHistory(history);
  };

  const addDownloadTask = (
    pluginId: string,
    pluginName: string,
    pluginIcon?: string,
    type: 'download' | 'install' | 'update' = 'download'
  ) => {
    const newTask: DownloadTask = {
      id: `${type}-${pluginId}-${Date.now()}`,
      pluginId,
      pluginName,
      pluginIcon,
      type,
      status: 'pending',
      progress: 0,
      message: '准备中...',
      startTime: new Date(),
    };

    setTasks(prev => [...prev, newTask]);
    return newTask.id;
  };

  const updateTask = (taskId: string, updates: Partial<DownloadTask>) => {
    setTasks(prev => prev.map(task => 
      task.id === taskId ? { ...task, ...updates } : task
    ));
  };

  const removeTask = (taskId: string) => {
    setTasks(prev => prev.filter(task => task.id !== taskId));
  };

  const startDownload = async (pluginId: string, pluginName: string, pluginIcon?: string) => {
    const taskId = addDownloadTask(pluginId, pluginName, pluginIcon, 'download');

    const onProgress: DownloadProgressCallback = (progress, message) => {
      updateTask(taskId, {
        status: 'downloading',
        progress,
        message,
      });
    };

    try {
      const filePath = await downloadService.downloadPlugin({
        pluginId,
        onProgress,
        onComplete: (filePath) => {
          updateTask(taskId, {
            status: 'completed',
            progress: 100,
            message: '下载完成',
            filePath,
            endTime: new Date(),
          });
        },
        onError: (error) => {
          updateTask(taskId, {
            status: 'failed',
            message: '下载失败',
            error: error.message,
            endTime: new Date(),
          });
        },
      });

      // 下载完成后自动开始安装
      setTimeout(() => {
        startInstallation(pluginId, pluginName, filePath, pluginIcon);
      }, 1000);

    } catch (error) {
      updateTask(taskId, {
        status: 'failed',
        message: '下载失败',
        error: error instanceof Error ? error.message : '未知错误',
        endTime: new Date(),
      });
    }
  };

  const startInstallation = async (
    pluginId: string, 
    pluginName: string, 
    filePath: string, 
    pluginIcon?: string
  ) => {
    const taskId = addDownloadTask(pluginId, pluginName, pluginIcon, 'install');

    const onProgress: InstallationCallback = (status, message) => {
      const progressMap: Record<InstallationStatus, number> = {
        not_installed: 0,
        installing: 50,
        installed: 100,
        failed: 0,
        updating: 75,
      };

      updateTask(taskId, {
        status: status === 'installed' ? 'completed' : 'installing',
        progress: progressMap[status],
        message: message || `状态: ${status}`,
      });
    };

    try {
      await downloadService.installPlugin({
        pluginId,
        filePath,
        onProgress,
        onComplete: () => {
          updateTask(taskId, {
            status: 'completed',
            progress: 100,
            message: '安装完成',
            endTime: new Date(),
          });
        },
        onError: (error) => {
          updateTask(taskId, {
            status: 'failed',
            message: '安装失败',
            error: error.message,
            endTime: new Date(),
          });
        },
      });
    } catch (error) {
      updateTask(taskId, {
        status: 'failed',
        message: '安装失败',
        error: error instanceof Error ? error.message : '未知错误',
        endTime: new Date(),
      });
    }
  };

  const cancelTask = (taskId: string, pluginId: string) => {
    const task = tasks.find(t => t.id === taskId);
    if (task?.type === 'download' && task.status === 'downloading') {
      downloadService.cancelDownload(pluginId);
    }
    removeTask(taskId);
  };

  const retryTask = (task: DownloadTask) => {
    removeTask(task.id);
    if (task.type === 'download') {
      startDownload(task.pluginId, task.pluginName, task.pluginIcon);
    } else if (task.type === 'install' && task.filePath) {
      startInstallation(task.pluginId, task.pluginName, task.filePath, task.pluginIcon);
    }
  };

  const clearCompleted = () => {
    setTasks(prev => prev.filter(task => 
      task.status !== 'completed' && task.status !== 'failed'
    ));
  };

  const getStatusIcon = (status: DownloadTask['status']) => {
    switch (status) {
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'failed':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />;
      case 'downloading':
      case 'installing':
        return <DownloadOutlined style={{ color: '#1890ff' }} />;
      case 'paused':
        return <PauseCircleOutlined style={{ color: '#faad14' }} />;
      default:
        return <PlayCircleOutlined style={{ color: '#d9d9d9' }} />;
    }
  };

  const getStatusColor = (status: DownloadTask['status']) => {
    switch (status) {
      case 'completed':
        return 'success';
      case 'failed':
        return 'error';
      case 'downloading':
      case 'installing':
        return 'processing';
      case 'paused':
        return 'warning';
      default:
        return 'default';
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDuration = (startTime: Date, endTime?: Date) => {
    const end = endTime || new Date();
    const duration = Math.floor((end.getTime() - startTime.getTime()) / 1000);
    
    if (duration < 60) return `${duration}秒`;
    if (duration < 3600) return `${Math.floor(duration / 60)}分${duration % 60}秒`;
    return `${Math.floor(duration / 3600)}小时${Math.floor((duration % 3600) / 60)}分`;
  };

  const renderCurrentTasks = () => {
    if (tasks.length === 0) {
      return (
        <Empty
          description="暂无下载任务"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      );
    }

    return (
      <>
        <div className="flex justify-between items-center mb-4">
          <Text strong>当前任务 ({tasks.length})</Text>
          <Button size="small" onClick={clearCompleted}>
            清除已完成
          </Button>
        </div>
        
        <List
          dataSource={tasks}
          renderItem={(task) => (
            <List.Item
              actions={[
                task.status === 'failed' && (
                  <Tooltip title="重试">
                    <Button
                      type="text"
                      icon={<ReloadOutlined />}
                      onClick={() => retryTask(task)}
                    />
                  </Tooltip>
                ),
                <Tooltip title="取消">
                  <Button
                    type="text"
                    icon={<DeleteOutlined />}
                    onClick={() => cancelTask(task.id, task.pluginId)}
                  />
                </Tooltip>,
              ].filter(Boolean)}
            >
              <List.Item.Meta
                avatar={
                  <Avatar 
                    src={task.pluginIcon} 
                    icon={getStatusIcon(task.status)}
                    size={40}
                  />
                }
                title={
                  <Space>
                    <Text strong>{task.pluginName}</Text>
                    <Tag color={getStatusColor(task.status)}>
                      {task.type === 'download' ? '下载' : '安装'}
                    </Tag>
                  </Space>
                }
                description={
                  <div>
                    <div>{task.message}</div>
                    {(task.status === 'downloading' || task.status === 'installing') && (
                      <Progress 
                        percent={task.progress} 
                        size="small" 
                        className="mt-2"
                        status={task.status === 'failed' ? 'exception' : 'active'}
                      />
                    )}
                    {task.error && (
                      <Text type="danger" className="text-xs">
                        错误: {task.error}
                      </Text>
                    )}
                    <div className="text-xs text-gray-500 mt-1">
                      开始时间: {task.startTime.toLocaleTimeString()}
                      {task.endTime && ` | 耗时: ${formatDuration(task.startTime, task.endTime)}`}
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </>
    );
  };

  const renderDownloadHistory = () => {
    if (downloadHistory.length === 0) {
      return (
        <Empty
          description="暂无下载历史"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      );
    }

    return (
      <>
        <div className="flex justify-between items-center mb-4">
          <Text strong>下载历史 ({downloadHistory.length})</Text>
          <Button size="small" icon={<ReloadOutlined />} onClick={loadDownloadHistory}>
            刷新
          </Button>
        </div>
        
        <List
          dataSource={downloadHistory}
          renderItem={(record) => (
            <List.Item
              actions={[
                <Tooltip title="打开文件位置">
                  <Button
                    type="text"
                    icon={<FolderOpenOutlined />}
                    onClick={() => {
                      // 这里可以调用系统API打开文件位置
                      console.log('打开文件位置:', record.filePath);
                    }}
                  />
                </Tooltip>,
              ]}
            >
              <List.Item.Meta
                avatar={
                  <Avatar 
                    icon={<CheckCircleOutlined />}
                    style={{ backgroundColor: '#52c41a' }}
                  />
                }
                title={
                  <Space>
                    <Text strong>插件 {record.pluginId}</Text>
                    <Tag>v{record.version}</Tag>
                  </Space>
                }
                description={
                  <div>
                    <div>文件大小: {formatFileSize(record.fileSize)}</div>
                    <div className="text-xs text-gray-500">
                      下载时间: {new Date(record.downloadedAt).toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-500">
                      文件路径: {record.filePath}
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </>
    );
  };

  return (
    <Drawer
      title="下载管理器"
      placement="right"
      width={480}
      open={visible}
      onClose={onClose}
      extra={
        <Space>
          <Button
            type={activeTab === 'current' ? 'primary' : 'default'}
            size="small"
            onClick={() => setActiveTab('current')}
          >
            当前任务
          </Button>
          <Button
            type={activeTab === 'history' ? 'primary' : 'default'}
            size="small"
            onClick={() => setActiveTab('history')}
          >
            下载历史
          </Button>
        </Space>
      }
    >
      {activeTab === 'current' ? renderCurrentTasks() : renderDownloadHistory()}
    </Drawer>
  );
};

export default DownloadManager;

// 导出便捷的Hook
export const useDownloadManager = () => {
  const [visible, setVisible] = useState(false);
  const [managerRef, setManagerRef] = useState<{
    startDownload: (pluginId: string, pluginName: string, pluginIcon?: string) => void;
  } | null>(null);

  const show = () => setVisible(true);
  const hide = () => setVisible(false);

  const DownloadManagerComponent = () => (
    <DownloadManager
      visible={visible}
      onClose={hide}
      ref={(ref: any) => setManagerRef(ref)}
    />
  );

  return {
    show,
    hide,
    visible,
    DownloadManagerComponent,
    startDownload: managerRef?.startDownload,
  };
};