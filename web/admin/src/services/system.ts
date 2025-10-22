import { request } from './api';

// 后端系统设置模型（下划线风格）
interface BackendSystemSetting {
  id: string;
  key: string;
  value: string;
  type: 'string' | 'number' | 'boolean' | 'json';
  category: string;
  description?: string;
  is_public: boolean;
  is_sensitive?: boolean;
  created_at?: string;
  updated_at?: string;
}

// 前端系统配置模型（驼峰风格）
export interface SystemConfig {
  id: string;
  key: string;
  value: string;
  type: 'string' | 'number' | 'boolean' | 'json';
  category: string;
  description?: string;
  isPublic: boolean;
  isSensitive?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

// 审计日志条目（与前端使用对齐）
export interface AuditLog {
  id: string;
  userId?: string;
  targetId?: string;
  targetType?: string;
  action: string;
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  description?: string;
  oldData?: string;
  newData?: string;
  ipAddress?: string;
  userAgent?: string;
  createdAt: string;
}

// 性能指标
export interface Metrics {
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
}

const mapSetting = (s: BackendSystemSetting): SystemConfig => ({
  id: s.id,
  key: s.key,
  value: s.value,
  type: s.type,
  category: s.category,
  description: s.description,
  isPublic: !!s.is_public,
  isSensitive: !!s.is_sensitive,
  createdAt: s.created_at,
  updatedAt: s.updated_at,
});

export const systemService = {
  // 获取所有系统配置
  async getConfigs(): Promise<SystemConfig[]> {
    const res = await request.get<BackendSystemSetting[]>(`/system/configs`);
    return (res || []).map(mapSetting);
  },

  // 保存系统配置（批量）
  async saveConfigs(settings: SystemConfig[]): Promise<void> {
    // 将驼峰转换为后端下划线 & 保持 value 为字符串（JSON文本）
    const payload: BackendSystemSetting[] = settings.map(s => ({
      id: s.id,
      key: s.key,
      value: s.value,
      type: s.type,
      category: s.category,
      description: s.description,
      is_public: !!s.isPublic,
      is_sensitive: !!s.isSensitive,
      updated_at: s.updatedAt,
      created_at: s.createdAt,
    }));
    await request.post(`/system/configs`, payload);
  },

  // 获取审计日志（分页）
  async getLogs(params: { page?: number; pageSize?: number; level?: string; module?: string; startDate?: string; endDate?: string; }): Promise<{ data: AuditLog[]; total: number; }> {
    const query = {
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 10,
      level: params.level,
      module: params.module,
      startDate: params.startDate,
      endDate: params.endDate,
    };
    const res = await request.get<{ data: any[]; total: number }>(`/system/logs`, { params: query });
    const logs: AuditLog[] = (res?.data || []).map((l: any) => ({
      id: l.id,
      userId: l.user_id,
      targetId: l.target_id,
      targetType: l.target_type,
      action: l.action,
      level: l.level,
      description: l.description,
      oldData: l.old_data,
      newData: l.new_data,
      ipAddress: l.ip_address,
      userAgent: l.user_agent,
      createdAt: l.created_at,
    }));
    return { data: logs, total: res?.total || 0 };
  },

  // 清理审计日志
  async clearLogs(params: { level?: string; module?: string; before?: string; }): Promise<number> {
    const res = await request.delete<{ affected: number }>(`/system/logs`, { params });
    return (res as any)?.affected ?? 0;
  },

  // 获取性能指标
  async getMetrics(): Promise<Metrics> {
    return await request.get<Metrics>(`/system/metrics`);
  }
};