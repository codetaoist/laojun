import { request } from './api';
import {
  Plugin,
  SearchParams,
  PaginatedResponse,
  Category,
  Review,
  DownloadStats,
  InstallationStatus,
  PluginVersion,
  Purchase,
} from '@/types';

// 将前端搜索参数映射为后端所需的查询键
const mapSearchParams = (params: SearchParams | undefined): Record<string, any> | undefined => {
  if (!params) return undefined;
  const mapped: Record<string, any> = {};
  if (params.query) mapped.query = params.query; // admin-api 使用 query
  if (params.category) mapped.category_id = params.category;
  if (params.featured !== undefined) mapped.featured = params.featured;
  if (params.rating !== undefined) mapped.min_rating = params.rating;
  if (params.priceRange && params.priceRange.length === 2) {
    mapped.min_price = params.priceRange[0];
    mapped.max_price = params.priceRange[1];
  }
  if (params.sortBy) {
    // 映射排序字段到后端支持的键
    switch (params.sortBy) {
      case 'downloads':
        mapped.sort_by = 'downloads';
        break;
      case 'rating':
        mapped.sort_by = 'rating';
        break;
      case 'updated':
        // admin-api 不区分 updated，使用 created_at 近似
        mapped.sort_by = 'created_at';
        break;
      case 'created':
        mapped.sort_by = 'created_at';
        break;
      case 'name':
        mapped.sort_by = 'name';
        break;
      default:
        mapped.sort_by = params.sortBy;
    }
  }
  if (params.sortOrder) mapped.sort_order = params.sortOrder;
  if (params.page) mapped.page = params.page;
  if (params.pageSize) mapped.limit = params.pageSize;
  return mapped;
};

// 插件相关 API
// 简单分类缓存，减少重复请求
let categoriesCache: { data: Category[]; timestamp: number } | null = null;
const CATEGORIES_TTL = 5 * 60 * 1000; // 5分钟

export const pluginService = {
  // 获取插件列表
  getPlugins: (params: GetPluginsParams): Promise<PaginatedResponse<Plugin>> => {
    return request.get('/plugins', { params });
  },

  // 获取特色插件
  getFeaturedPlugins: async (limit: number = 8): Promise<Plugin[]> => {
    const response = await request.get('/plugins', { params: { featured: true, limit } });
    return response.data || [];
  },

  // 获取最新插件
  getLatestPlugins: async (limit: number = 8): Promise<Plugin[]> => {
    const response = await request.get('/plugins', { params: { sort: 'latest', limit } });
    return response.data || [];
  },

  // 获取分类列表
  getCategories: async (): Promise<Category[]> => {
    const response = await request.get('/categories');
    return response.data || [];
  },

  // 获取插件详情
  getPluginDetail: (id: string): Promise<PluginDetail> => {
    return request.get(`/plugins/${id}`);
  },

  // 获取插件评论列表
  getPluginReviews: (id: string, page?: number, limit?: number): Promise<PaginatedResponse<Review>> => {
    return request.get(`/plugins/${id}/reviews`, { params: { page, limit } });
  },

  // 创建评论（需登录）
  createReview: (pluginId: string, payload: CreateReviewPayload): Promise<{ message: string } & Review> => {
    return request.post(`/plugins/${pluginId}/reviews`, payload);
  },

  // 获取扩展插件列表
  getExtendedPlugins: (params: GetPluginsParams): Promise<PaginatedResponse<Plugin>> => {
    return request.get('/marketplace/plugins', { params });
  },

  // 搜索扩展插件
  searchExtendedPlugins: (keyword: string): Promise<Plugin[]> => {
    return request.get('/marketplace/plugins/search', { params: { keyword } });
  },

  // 下载扩展插件
  downloadExtendedPlugin: (id: string): Promise<{ message: string }> => {
    return request.post(`/marketplace/plugins/${id}/download`);
  },

  // 安装扩展插件
  installExtendedPlugin: (id: string): Promise<{ message: string }> => {
    return request.post(`/marketplace/plugins/${id}/install`);
  },

  // 更新扩展插件
  updateExtendedPlugin: (id: string): Promise<{ message: string }> => {
    return request.post(`/marketplace/plugins/${id}/update`);
  },

  // 获取用户收藏的扩展插件
  getFavoritePlugins: (): Promise<Plugin[]> => {
    return request.get('/marketplace/plugins/favorites');
  },

  // 获取扩展插件统计信息
  getExtendedPluginStats: (): Promise<{ total: number, installed: number, updated: number }> => {
    return request.get('/marketplace/plugins/stats');
  },
};

export default pluginService;