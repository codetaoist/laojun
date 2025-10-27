import { request } from './api';

// 系统配置相关接口
export interface SystemConfig {
  id: string;
  key: string;
  value: string;
  description?: string;
  category: string;
  type: 'string' | 'number' | 'boolean' | 'json';
  isPublic: boolean;
  updatedAt: string;
}

// 系统日志相关接口
export interface SystemLog {
  id: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: string;
  source: string;
  details?: any;
}

// 系统指标相关接口
export interface SystemMetrics {
  cpu: {
    usage: number;
    cores: number;
  };
  memory: {
    used: number;
    total: number;
    usage: number;
  };
  disk: {
    used: number;
    total: number;
    usage: number;
  };
  network: {
    bytesIn: number;
    bytesOut: number;
  };
  uptime: number;
  timestamp: string;
}

// 配置管理API
export const configAPI = {
  // 获取配置列表
  getConfigs: (): Promise<SystemConfig[]> => {
    return request.get('/system/configs');
  },

  // 获取单个配置
  getConfig: (id: string): Promise<SystemConfig> => {
    return request.get(`/system/configs/${id}`);
  },

  // 创建配置
  createConfig: (config: Omit<SystemConfig, 'id' | 'updatedAt'>): Promise<SystemConfig> => {
    return request.post('/system/configs', config);
  },

  // 更新配置
  updateConfig: (id: string, config: Partial<Omit<SystemConfig, 'id' | 'updatedAt'>>): Promise<SystemConfig> => {
    return request.put(`/system/configs/${id}`, config);
  },

  // 删除配置
  deleteConfig: (id: string): Promise<void> => {
    return request.delete(`/system/configs/${id}`);
  },

  // 批量更新配置
  batchUpdateConfigs: (configs: Array<{ id: string; value: string }>): Promise<SystemConfig[]> => {
    return request.put('/system/configs/batch', { configs });
  },
};

// 日志管理API
export const logAPI = {
  // 获取日志列表
  getLogs: (params?: {
    level?: string;
    source?: string;
    startTime?: string;
    endTime?: string;
    page?: number;
    pageSize?: number;
  }): Promise<{
    logs: SystemLog[];
    total: number;
    page: number;
    pageSize: number;
  }> => {
    return request.get('/system/logs', { params });
  },

  // 清理日志
  clearLogs: (params?: {
    level?: string;
    source?: string;
    beforeTime?: string;
  }): Promise<{ deletedCount: number }> => {
    return request.delete('/system/logs', { params });
  },

  // 导出日志
  exportLogs: async (
    params?: { level?: string; source?: string; startTime?: string; endTime?: string }
  ): Promise<Blob> => {
    try {
      return await request.get('/system/logs/export', { params, responseType: 'blob' });
    } catch (error) {
      const data = await logAPI.getLogs({
        level: params?.level,
        source: params?.source,
        startTime: params?.startTime,
        endTime: params?.endTime,
        page: 1,
        pageSize: 1000,
      });
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      return blob;
    }
  },
};

// 性能监控API
export const metricsAPI = {
  // 获取当前系统指标
  getCurrentMetrics: (): Promise<SystemMetrics> => {
    return request.get('/system/metrics');
  },

  // 获取历史指标数据
  getHistoryMetrics: (params?: {
    startTime?: string;
    endTime?: string;
    interval?: string; // '1m', '5m', '1h', '1d'
  }): Promise<SystemMetrics[]> => {
    return request.get('/system/metrics/history', { params });
  },
};

// 系统信息API
export const systemAPI = {
  // 获取系统信息
  getSystemInfo: (): Promise<{
    version: string;
    buildTime: string;
    gitCommit: string;
    goVersion: string;
    platform: string;
    arch: string;
  }> => {
    return request.get('/system/info');
  },

  // 健康检查
  healthCheck: (): Promise<{ status: string; version: string }> => {
    return request.get('/health');
  },
};