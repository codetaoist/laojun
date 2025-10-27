import { request } from './api';
import { Permission, PaginatedResponse, QueryParams } from '@/types';

// 权限创建请求接口
export interface CreatePermissionRequest {
  name: string;
  resource: string;
  action: string;
  description?: string;
}

// 权限更新请求接口
export interface UpdatePermissionRequest {
  name?: string;
  resource?: string;
  action?: string;
  description?: string;
}

// 权限模板接口
export interface PermissionTemplate {
  id: string;
  name: string;
  description: string;
  permissionIds: string[];
  isSystem: boolean;
  createdAt: string;
  updatedAt: string;
}

// 权限模板创建请求接口
export interface CreatePermissionTemplateRequest {
  name: string;
  description: string;
  permissionIds: string[];
}

// 权限模板更新请求接口
export interface UpdatePermissionTemplateRequest {
  name?: string;
  description?: string;
  permissionIds?: string[];
}

// 权限统计信息接口
export interface PermissionStats {
  totalPermissions: number;
  systemPermissions: number;
  customPermissions: number;
  byResource: { [key: string]: number };
  byAction: { [key: string]: number };
  recentlyCreated: Permission[];
  mostUsed: Permission[];
}

// 权限查询参数接口
export interface PermissionQueryParams extends QueryParams {
  resource?: string;
  action?: string;
  isSystem?: boolean;
}

// 当前用户权限检查响应
export interface UserPermissionCheckResponse {
  hasPermission: boolean;
  reason?: string;
  module?: string;
  resource?: string;
  action?: string;
  deviceType?: string;
}

export const permissionService = {
  // 获取权限列表
  getPermissions: (params?: PermissionQueryParams): Promise<PaginatedResponse<Permission>> => {
    return request.get<PaginatedResponse<Permission>>('/permissions', { params });
  },

  // 获取权限详情
  getPermission: (id: string): Promise<Permission> => {
    return request.get<Permission>(`/permissions/${id}`);
  },

  // 创建权限
  createPermission: (data: CreatePermissionRequest): Promise<Permission> => {
    return request.post<Permission>('/permissions', data);
  },

  // 更新权限
  updatePermission: (id: string, data: UpdatePermissionRequest): Promise<Permission> => {
    return request.put<Permission>(`/permissions/${id}`, data);
  },

  // 删除权限
  deletePermission: (id: string): Promise<void> => {
    return request.delete(`/permissions/${id}`);
  },

  // 批量删除权限
  batchDeletePermissions: (ids: string[]): Promise<void> => {
    return request.post('/permissions/batch-delete', { ids });
  },

  // 获取权限统计信息
  getPermissionStats: (): Promise<PermissionStats> => {
    return request.get<PermissionStats>('/permissions/stats');
  },

  // 获取所有资源列表
  getResources: (): Promise<string[]> => {
    return request.get<string[]>('/permissions/resources');
  },

  // 获取所有操作列表
  getActions: (): Promise<string[]> => {
    return request.get<string[]>('/permissions/actions');
  },

  // 检查权限是否被使用
  checkPermissionUsage: (id: string): Promise<{
    isUsed: boolean;
    usedByRoles: Array<{ id: string; name: string; displayName?: string }>;
    usedByUsers: Array<{ id: string; username: string; name: string }>;
  }> => {
    return request.get(`/permissions/${id}/usage`);
  },

  // 检查当前用户是否拥有某项权限
  checkCurrentUserPermission: async (params: { deviceType?: string; module: string; resource: string; action: string; }): Promise<UserPermissionCheckResponse> => {
    const query = {
      device_type: params.deviceType || 'web',
      module: params.module,
      resource: params.resource,
      action: params.action,
    };
    const res = await request.get<UserPermissionCheckResponse>('/permissions/check', { params: query });
    return res;
  },

  // 权限模板相关接口
  
  // 获取权限模板列表
  getPermissionTemplates: (params?: QueryParams): Promise<PaginatedResponse<PermissionTemplate>> => {
    return request.get<PaginatedResponse<PermissionTemplate>>('/permission-templates', { params });
  },

  // 获取权限模板详情
  getPermissionTemplate: (id: string): Promise<PermissionTemplate> => {
    return request.get<PermissionTemplate>(`/permission-templates/${id}`);
  },

  // 创建权限模板
  createPermissionTemplate: (data: CreatePermissionTemplateRequest): Promise<PermissionTemplate> => {
    return request.post<PermissionTemplate>('/permission-templates', data);
  },

  // 更新权限模板
  updatePermissionTemplate: (id: string, data: UpdatePermissionTemplateRequest): Promise<PermissionTemplate> => {
    return request.put<PermissionTemplate>(`/permission-templates/${id}`, data);
  },

  // 删除权限模板
  deletePermissionTemplate: (id: string): Promise<void> => {
    return request.delete(`/permission-templates/${id}`);
  },

  // 应用权限模板到角色
  applyTemplateToRole: (templateId: string, roleId: string): Promise<void> => {
    return request.post(`/permission-templates/${templateId}/apply`, { role_id: roleId });
  },

  // 从角色创建权限模板
  createTemplateFromRole: (roleId: string, templateData: { name: string; description: string }): Promise<PermissionTemplate> => {
    return request.post(`/roles/${roleId}/create-template`, templateData);
  },

  // 导出权限配置
  exportPermissions: (format: 'json' | 'csv' | 'excel' = 'json'): Promise<Blob> => {
    return request.get(`/permissions/export?format=${format}`, {
      responseType: 'blob'
    });
  },

  // 导入权限配置
  importPermissions: (file: File): Promise<{
    success: number;
    failed: number;
    errors: string[];
  }> => {
    const formData = new FormData();
    formData.append('file', file);
    return request.post('/permissions/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
  },

  // 同步系统权限
  syncSystemPermissions: (): Promise<{
    added: number;
    updated: number;
    removed: number;
  }> => {
    return request.post('/permissions/sync-system');
  }
};