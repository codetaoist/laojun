import { request } from './api';
import { PaginatedResponse, QueryParams } from '@/types';

// 设备类型枚举
export enum DeviceType {
  PC = 'pc',
  WEB = 'web',
  MOBILE = 'mobile',
  WATCH = 'watch',
  IOT = 'iot',
  ROBOT = 'robot'
}

// 菜单相关类型定义
export interface Menu {
  id: string;
  title: string;
  path?: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sortOrder: number;
  isHidden: boolean;
  isFavorite: boolean;
  deviceTypes?: string; // JSON字符串，存储适配的设备类型
  permissions?: string; // JSON字符串，存储权限要求
  customIcon?: string; // 自定义图标URL
  description?: string; // 菜单描述
  keywords?: string; // 搜索关键词
  level?: number; // 菜单层级
  children?: Menu[];
  createdAt: string;
  updatedAt: string;
}

// 创建菜单请求
export interface CreateMenuRequest {
  title: string;
  path?: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sortOrder?: number;
  isHidden?: boolean;
  isFavorite?: boolean;
  deviceTypes?: string;
  permissions?: string;
  customIcon?: string;
  description?: string;
  keywords?: string;
}

// 更新菜单请求
export interface UpdateMenuRequest {
  title?: string;
  path?: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sortOrder?: number;
  isHidden?: boolean;
  isFavorite?: boolean;
  deviceTypes?: string;
  permissions?: string;
  customIcon?: string;
  description?: string;
  keywords?: string;
}

// 菜单搜索参数
export interface MenuSearchParams extends QueryParams {
  title?: string;
  path?: string;
  parentId?: string;
  isHidden?: boolean;
  tree?: boolean; // 是否返回树形结构
}

// 菜单移动请求
export interface MenuMoveRequest {
  targetParentId?: string;
  targetPosition: number;
}

// 批量更新菜单请求
export interface MenuBatchUpdateRequest {
  ids: string[];
  updates: Partial<UpdateMenuRequest>;
}

// 菜单统计信息
export interface MenuStats {
  totalMenus: number;
  visibleMenus: number;
  hiddenMenus: number;
  topLevelMenus: number;
  maxDepth: number;
}

// 后端菜单响应模型
interface BackendMenuResponse {
  id: string;
  title: string;
  path?: string | null;
  icon?: string | null;
  component?: string | null;
  parent_id?: string | null;
  sort_order: number;
  is_hidden: boolean;
  is_favorite: boolean;
  device_types?: string | null;
  permissions?: string | null;
  custom_icon?: string | null;
  description?: string | null;
  keywords?: string | null;
  level?: number;
  created_at: string;
  updated_at: string;
  children?: BackendMenuResponse[];
}

interface BackendMenuListResponse {
  data: BackendMenuResponse[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

// 将后端菜单响应映射为前端 Menu 类型
const mapMenu = (m: BackendMenuResponse): Menu => ({
  id: m.id,
  title: m.title,
  path: m.path ?? undefined,
  icon: m.icon ?? undefined,
  component: m.component ?? undefined,
  parentId: m.parent_id ?? undefined,
  sortOrder: m.sort_order,
  isHidden: m.is_hidden,
  isFavorite: m.is_favorite,
  deviceTypes: m.device_types ?? undefined,
  permissions: m.permissions ?? undefined,
  customIcon: m.custom_icon ?? undefined,
  description: m.description ?? undefined,
  keywords: m.keywords ?? undefined,
  level: m.level,
  children: m.children ? m.children.map(mapMenu) : undefined,
  createdAt: m.created_at,
  updatedAt: m.updated_at,
});

// 将前端创建请求映射为后端字段命名
const toBackendCreatePayload = (data: CreateMenuRequest): Record<string, any> => {
  const payload: Record<string, any> = {
    title: data.title,
  };
  if (typeof data.path !== 'undefined') payload.path = data.path;
  if (typeof data.icon !== 'undefined') payload.icon = data.icon;
  if (typeof data.component !== 'undefined') payload.component = data.component;
  if (typeof data.parentId !== 'undefined') payload.parent_id = data.parentId;
  if (typeof data.sortOrder !== 'undefined') payload.sort_order = data.sortOrder;
  if (typeof data.isHidden !== 'undefined') payload.is_hidden = data.isHidden;
  if (typeof data.isFavorite !== 'undefined') payload.is_favorite = data.isFavorite;
  if (typeof data.deviceTypes !== 'undefined') payload.device_types = data.deviceTypes;
  if (typeof data.permissions !== 'undefined') payload.permissions = data.permissions;
  if (typeof data.customIcon !== 'undefined') payload.custom_icon = data.customIcon;
  if (typeof data.description !== 'undefined') payload.description = data.description;
  if (typeof data.keywords !== 'undefined') payload.keywords = data.keywords;
  return payload;
};

// 将前端更新请求映射为后端字段命名
const toBackendUpdatePayload = (data: UpdateMenuRequest): Record<string, any> => {
  const payload: Record<string, any> = {};
  if (typeof data.title !== 'undefined') payload.title = data.title;
  if (typeof data.path !== 'undefined') payload.path = data.path;
  if (typeof data.icon !== 'undefined') payload.icon = data.icon;
  if (typeof data.component !== 'undefined') payload.component = data.component;
  if (typeof data.parentId !== 'undefined') payload.parent_id = data.parentId;
  if (typeof data.sortOrder !== 'undefined') payload.sort_order = data.sortOrder;
  if (typeof data.isHidden !== 'undefined') payload.is_hidden = data.isHidden;
  if (typeof data.isFavorite !== 'undefined') payload.is_favorite = data.isFavorite;
  if (typeof data.deviceTypes !== 'undefined') payload.device_types = data.deviceTypes;
  if (typeof data.permissions !== 'undefined') payload.permissions = data.permissions;
  if (typeof data.customIcon !== 'undefined') payload.custom_icon = data.customIcon;
  if (typeof data.description !== 'undefined') payload.description = data.description;
  if (typeof data.keywords !== 'undefined') payload.keywords = data.keywords;
  return payload;
};

// 将前端移动请求映射为后端字段命名
const toBackendMovePayload = (data: MenuMoveRequest): Record<string, any> => {
  const payload: Record<string, any> = {
    target_position: data.targetPosition,
  };
  if (typeof data.targetParentId !== 'undefined') payload.target_parent_id = data.targetParentId;
  return payload;
};

export const menuService = {
  // 获取菜单列表
  getMenus: async (params?: MenuSearchParams): Promise<PaginatedResponse<Menu>> => {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.pageSize) queryParams.append('size', params.pageSize.toString());
    if (params?.title) queryParams.append('title', params.title);
    if (params?.path) queryParams.append('path', params.path);
    if (params?.parentId) queryParams.append('parent_id', params.parentId);
    if (typeof params?.isHidden !== 'undefined') queryParams.append('is_hidden', params.isHidden.toString());
    if (typeof params?.tree !== 'undefined') queryParams.append('tree_mode', params.tree.toString());
    if (params?.search) queryParams.append('search', params.search);

    const response = await request.get<BackendMenuListResponse>(`/menus?${queryParams.toString()}`);
    return {
      items: response.data.map(mapMenu),
      total: response.total,
      page: response.page,
      pageSize: response.limit,
      totalPages: response.total_pages,
    };
  },

  // 获取单个菜单
  getMenu: async (id: string): Promise<Menu> => {
    const response = await request.get<BackendMenuResponse>(`/menus/${id}`);
    return mapMenu(response);
  },

  // 创建菜单
  createMenu: async (data: CreateMenuRequest): Promise<Menu> => {
    const payload = toBackendCreatePayload(data);
    const response = await request.post<BackendMenuResponse>('/menus', payload);
    return mapMenu(response);
  },

  // 更新菜单
  updateMenu: async (id: string, data: UpdateMenuRequest): Promise<Menu> => {
    const payload = toBackendUpdatePayload(data);
    const response = await request.put<BackendMenuResponse>(`/menus/${id}`, payload);
    return mapMenu(response);
  },

  // 删除菜单
  deleteMenu: (id: string): Promise<void> => {
    return request.delete(`/menus/${id}`);
  },

  // 批量删除菜单
  batchDeleteMenus: (ids: string[]): Promise<void> => {
    return request.delete('/menus/batch', { data: { ids } });
  },

  // 移动菜单
  moveMenu: async (id: string, data: MenuMoveRequest): Promise<Menu> => {
    const payload = toBackendMovePayload(data);
    const response = await request.post<BackendMenuResponse>(`/menus/${id}/move`, payload);
    return mapMenu(response);
  },

  // 获取菜单统计信息
  getMenuStats: async (): Promise<MenuStats> => {
    const response = await request.get<BackendMenuStatsResponse>('/menus/stats');
    return {
      totalMenus: response.total_menus ?? 0,
      visibleMenus: response.visible_menus ?? 0,
      hiddenMenus: response.hidden_menus ?? 0,
      // 后端暂未提供顶层菜单统计，先用总数兜底或按需计算
      topLevelMenus: 0,
      maxDepth: response.max_depth ?? 0,
    };
  },

  // 批量更新菜单
  batchUpdateMenus: async (data: MenuBatchUpdateRequest): Promise<void> => {
    const payload = {
      ids: data.ids,
      updates: toBackendUpdatePayload(data.updates),
    };
    await request.put('/menus/batch', payload);
  },
};

// 图标库相关类型和服务
export interface IconLibraryItem {
  id: string;
  name: string;
  iconType: 'antd' | 'custom' | 'svg' | 'font';
  iconData: string;
  category?: string;
  tags: string[];
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface IconSearchParams extends QueryParams {
  category?: string;
  iconType?: string;
  tags?: string;
  search?: string;
}

export const iconService = {
  // 获取图标库列表
  getIcons: async (params?: IconSearchParams): Promise<PaginatedResponse<IconLibraryItem>> => {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.pageSize) queryParams.append('size', params.pageSize.toString());
    if (params?.category) queryParams.append('category', params.category);
    if (params?.iconType) queryParams.append('icon_type', params.iconType);
    if (params?.tags) queryParams.append('tags', params.tags);
    if (params?.search) queryParams.append('search', params.search);

    const response = await request.get<any>(`/icons?${queryParams.toString()}`);
    return response;
  },

  // 上传自定义图标
  uploadIcon: async (file: File, name: string, category?: string, tags?: string[]): Promise<IconLibraryItem> => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('name', name);
    if (category) formData.append('category', category);
    if (tags) formData.append('tags', JSON.stringify(tags));

    const response = await request.post<IconLibraryItem>('/icons/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    return response;
  },

  // 删除图标
  deleteIcon: async (id: string): Promise<void> => {
    await request.delete(`/icons/${id}`);
  },
};

// 菜单配置相关类型和服务
export interface MenuConfig {
  id: string;
  name: string;
  description?: string;
  deviceType: DeviceType;
  config: Record<string, any>;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface MenuConfigRequest {
  name: string;
  description?: string;
  deviceType: DeviceType;
  config: Record<string, any>;
  isActive?: boolean;
}

export const menuConfigService = {
  // 获取配置列表
  getConfigs: async (deviceType?: DeviceType): Promise<MenuConfig[]> => {
    const queryParams = deviceType ? `?device_type=${deviceType}` : '';
    const response = await request.get<MenuConfig[]>(`/menu-configs${queryParams}`);
    return response;
  },

  // 获取单个配置
  getConfig: async (id: string): Promise<MenuConfig> => {
    const response = await request.get<MenuConfig>(`/menu-configs/${id}`);
    return response;
  },

  // 创建配置
  createConfig: async (data: MenuConfigRequest): Promise<MenuConfig> => {
    const response = await request.post<MenuConfig>('/menu-configs', data);
    return response;
  },

  // 更新配置
  updateConfig: async (id: string, data: Partial<MenuConfigRequest>): Promise<MenuConfig> => {
    const response = await request.put<MenuConfig>(`/menu-configs/${id}`, data);
    return response;
  },

  // 删除配置
  deleteConfig: async (id: string): Promise<void> => {
    await request.delete(`/menu-configs/${id}`);
  },

  // 导出配置
  exportConfig: async (id: string): Promise<Blob> => {
    const response = await request.get(`/menu-configs/${id}/export`, { responseType: 'blob' });
    return response;
  },

  // 导入配置
  importConfig: async (file: File): Promise<MenuConfig> => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await request.post<MenuConfig>('/menu-configs/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    return response;
  },
};

// 拖拽排序相关类型和服务
export interface DragSortRequest {
  menuId: string;
  newParentId?: string;
  newSortOrder: number;
  targetMenuId?: string;
  position: 'before' | 'after' | 'inside';
}

export interface BatchOperationRequest {
  menuIds: string[];
  operation: 'delete' | 'hide' | 'show' | 'favorite' | 'unfavorite';
}

export const menuOperationService = {
  // 拖拽排序
  dragSort: async (data: DragSortRequest): Promise<void> => {
    const payload = {
      menu_id: data.menuId,
      new_parent_id: data.newParentId,
      new_sort_order: data.newSortOrder,
      target_menu_id: data.targetMenuId,
      position: data.position,
    };
    await request.post('/menus/drag-sort', payload);
  },

  // 批量操作
  batchOperation: async (data: BatchOperationRequest): Promise<void> => {
    const payload = {
      menu_ids: data.menuIds,
      operation: data.operation,
    };
    await request.post('/menus/batch-operation', payload);
  },

  // 切换收藏状态
  toggleFavorite: async (id: string): Promise<void> => {
    await request.post(`/menus/${id}/toggle-favorite`);
  },
};