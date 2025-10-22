import { request } from './api';
import { Role, PaginatedResponse, QueryParams, Permission } from '@/types';

export interface CreateRoleRequest {
  name: string;
  displayName: string;
  description?: string;
}

export interface UpdateRoleRequest {
  displayName?: string;
  description?: string;
}

export interface AssignPermissionRequest {
  roleId: string;
  permissionIds: string[];
}

// 后端角色响应模型
interface BackendRole {
  id: string;
  name: string;
  display_name: string;
  description?: string | null;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

// 后端角色分页响应模型
interface BackendRoleListResponse {
  data: BackendRole[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

const mapRole = (r: BackendRole): Role => ({
  id: r.id,
  name: r.name,
  displayName: r.display_name,
  description: r.description ?? undefined,
  permissions: [],
  isSystem: !!r.is_system,
  createdAt: r.created_at,
  updatedAt: r.updated_at,
});

export const roleService = {
  // 获取角色列表（分页+搜索），并映射为前端分页结构
  getRoles: async (params?: QueryParams): Promise<PaginatedResponse<Role>> => {
    const query = {
      page: params?.page ?? 1,
      size: params?.pageSize ?? (params as any)?.limit ?? 10,
      search: params?.search ?? '',
    };
    const res: BackendRoleListResponse = await request.get('/roles', { params: query });
    return {
      items: (res.data || []).map(mapRole),
      total: res.total,
      page: res.page,
      pageSize: res.limit,
      totalPages: res.total_pages,
    };
  },

  // 获取角色详情并映射
  getRole: async (id: string): Promise<Role> => {
    const r: BackendRole = await request.get(`/roles/${id}`);
    return mapRole(r);
  },

  // 创建角色（映射后端字段）
  createRole: async (data: CreateRoleRequest): Promise<Role> => {
    const r: BackendRole = await request.post('/roles', data);
    return mapRole(r);
  },

  // 更新角色（映射后端字段）
  updateRole: async (id: string, data: UpdateRoleRequest): Promise<Role> => {
    const r: BackendRole = await request.put(`/roles/${id}`, data);
    return mapRole(r);
  },

  // 删除角色
  deleteRole: (id: string): Promise<void> => {
    return request.delete(`/roles/${id}`);
  },

  // 批量删除角色（后端尚未实现，暂不使用）
  batchDeleteRoles: (ids: string[]): Promise<void> => {
    return request.post('/roles/batch-delete', { ids });
  },

  // 为角色分配权限（后端接口待接通）
  assignPermissions: (data: AssignPermissionRequest): Promise<void> => {
    return request.post(`/roles/${data.roleId}/permissions`, { 
      permissionIds: data.permissionIds 
    });
  },

  // 获取角色的权限（返回权限ID数组）
  getRolePermissions: async (roleId: string): Promise<string[]> => {
    const perms = await request.get<{ id: string }[]>(`/roles/${roleId}/permissions`);
    return (perms || []).map(p => p.id);
  },

  // 获取所有可用权限（基础权限结构）
  getAvailablePermissions: (params?: { search?: string; resource?: string; action?: string }): Promise<Permission[]> => {
    return request.get<Permission[]>('/permissions', { params });
  },
};