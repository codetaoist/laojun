import api, { request } from './api';
import { Plugin } from '@/types';

export interface PluginListParams {
  page?: number;
  pageSize?: number;
  search?: string;
  status?: 'enabled' | 'disabled';
  isSystem?: boolean;
}

export interface PluginInstallRequest {
  file: File;
}

export interface PluginConfigRequest {
  pluginId: string;
  config: Record<string, any>;
}

export interface PluginStatusRequest {
  pluginId: string;
  status: 'enabled' | 'disabled';
}

export interface PluginStats {
  total: number;
  enabled: number;
  disabled: number;
  system: number;
}

// 获取插件列表
export const getPlugins = (params?: PluginListParams) => {
  return api.get<Plugin[]>('/plugins', { params });
};

// 获取插件详情
export const getPlugin = (id: string) => {
  return api.get<Plugin>(`/plugins/${id}`);
};

// 安装插件
export const installPlugin = (data: PluginInstallRequest) => {
  const formData = new FormData();
  formData.append('file', data.file);
  
  return api.post<Plugin>('/plugins/install', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
};

// 卸载插件
export const uninstallPlugin = (id: string) => {
  return api.delete(`/plugins/${id}`);
};

// 启用/禁用插件
export const updatePluginStatus = (data: PluginStatusRequest) => {
  return api.patch(`/plugins/${data.pluginId}/status`, {
    status: data.status,
  });
};

// 更新插件配置
export const updatePluginConfig = (data: PluginConfigRequest) => {
  return api.patch(`/plugins/${data.pluginId}/config`, {
    config: data.config,
  });
};

// 获取插件配置
export const getPluginConfig = (id: string) => {
  return api.get<Record<string, any>>(`/plugins/${id}/config`);
};

// 批量卸载插件
export const batchUninstallPlugins = (ids: string[]) => {
  return api.delete('/plugins/batch', {
    data: { ids },
  });
};

// 获取插件统计信息
export const getPluginStats = () => {
  return api.get<PluginStats>('/plugins/stats');
};

// 重启插件
export const restartPlugin = (id: string) => {
  return api.post(`/plugins/${id}/restart`);
};

// 获取插件日志
export const getPluginLogs = (id: string, params?: { page?: number; pageSize?: number }) => {
  return api.get<string[]>(`/plugins/${id}/logs`, { params });
};

// 验证插件包
export const validatePlugin = (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  
  return api.post<{ valid: boolean; errors?: string[] }>('/plugins/validate', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
};

// 获取插件市场列表
export const getMarketplacePlugins = (params?: {
  page?: number;
  pageSize?: number;
  search?: string;
  category?: string;
  sort?: 'name' | 'downloads' | 'rating' | 'updated';
}) => {
  return request.get<Plugin[]>('/marketplace/plugins', { params });
};

// 从市场安装插件
export const installFromMarketplace = (pluginId: string) => {
  return api.post<Plugin>(`/plugins/marketplace/${pluginId}/install`);
};

// 获取插件市场分类
export const getMarketplaceCategories = () => {
  return request.get<any[]>('/marketplace/categories');
};

export const pluginService = {
  getPlugins,
  getPlugin,
  installPlugin,
  uninstallPlugin,
  updatePluginStatus,
  updatePluginConfig,
  getPluginConfig,
  batchUninstallPlugins,
  getPluginStats,
  restartPlugin,
  getPluginLogs,
  validatePlugin,
  getMarketplacePlugins,
  installFromMarketplace,
  getMarketplaceCategories,
};