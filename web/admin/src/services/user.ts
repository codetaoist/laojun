import { request } from './api';
import { User, PaginatedResponse, QueryParams } from '@/types';

// 与后端模型对齐的请求类型
export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  avatar?: string;
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  avatar?: string;
  isActive?: boolean;
}

// 后端响应模型（局部定义用于映射）
interface BackendUserResponse {
  id: string;
  username: string;
  email: string;
  avatar?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string | null;
  roles?: Array<{
    id: string;
    name: string;
    display_name?: string;
    description?: string | null;
    is_system: boolean;
    created_at: string;
    updated_at: string;
  }>;
}

interface BackendUserListResponse {
  users: BackendUserResponse[];
  total: number;
  page: number;
  size: number;
}

// 将后端用户响应映射为前端 User 类型
const mapUser = (u: BackendUserResponse): User => ({
  id: u.id,
  username: u.username,
  email: u.email,
  name: u.username, // 后端无 name 字段，先用 username 兜底
  avatar: u.avatar ?? undefined,
  status: u.is_active ? 'active' : 'inactive',
  roles: (u.roles || []).map(r => ({
    id: r.id,
    name: r.name,
    description: r.description ?? undefined,
    permissions: [],
    isSystem: !!r.is_system,
    createdAt: r.created_at,
    updatedAt: r.updated_at,
  })),
  createdAt: u.created_at,
  updatedAt: u.updated_at,
  lastLoginAt: u.last_login_at || undefined,
});

// 将前端更新请求映射为后端字段命名（snake_case）
const toBackendUpdatePayload = (data: UpdateUserRequest): Record<string, any> => {
  const payload: Record<string, any> = {};
  if (typeof data.username !== 'undefined') payload.username = data.username;
  if (typeof data.email !== 'undefined') payload.email = data.email;
  if (typeof data.avatar !== 'undefined') payload.avatar = data.avatar;
  if (typeof data.bio !== 'undefined') payload.bio = data.bio;
  if (typeof data.isActive !== 'undefined') payload.is_active = data.isActive;
  return payload;
};

export const userService = {
  // 获取用户列表（分页+搜索），并映射为前端分页结构
  getUsers: async (params?: QueryParams): Promise<PaginatedResponse<User>> => {
    // 后端参数为 page/size/search
    const query = {
      page: params?.page ?? 1,
      size: params?.pageSize ?? params?.limit ?? 10,
      search: params?.search ?? '',
    };
    const res: BackendUserListResponse = await request.get('/users', { params: query });
    return {
      items: (res.users || []).map(mapUser),
      total: res.total,
      page: res.page,
      pageSize: res.size,
      totalPages: Math.ceil(res.total / (res.size || 1)),
    };
  },

  // 获取用户详情并映射
  getUser: async (id: string): Promise<User> => {
    const u: BackendUserResponse = await request.get(`/users/${id}`);
    return mapUser(u);
  },

  // 创建用户
  createUser: async (data: CreateUserRequest): Promise<User> => {
    const u: BackendUserResponse = await request.post('/users', data);
    return mapUser(u);
  },

  // 更新用户（映射 isActive => is_active）
  updateUser: async (id: string, data: UpdateUserRequest): Promise<User> => {
    const payload = toBackendUpdatePayload(data);
    const u: BackendUserResponse = await request.put(`/users/${id}`, payload);
    return mapUser(u);
  },

  // 删除用户
  deleteUser: (id: string): Promise<void> => {
    return request.delete(`/users/${id}`);
  },

  // 重置用户密码（管理员功能，无需旧密码）
  resetPassword: (id: string, newPassword: string): Promise<void> => {
    return request.post(`/users/${id}/reset-password`, { new_password: newPassword });
  },

  // 设置启用/禁用状态
  setActive: async (id: string, isActive: boolean): Promise<User> => {
    return userService.updateUser(id, { isActive });
  },

  // 为用户分配角色
  assignRoles: async (id: string, roleIds: string[]): Promise<void> => {
    return request.post(`/users/${id}/roles`, { 
      user_id: id,
      role_ids: roleIds 
    });
  },
};