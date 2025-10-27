import { request } from './api';
import { 
  ForumCategory, 
  ForumPost, 
  ForumReply, 
  BlogCategory, 
  BlogPost, 
  CodeSnippet,
  PaginationMeta 
} from '@/types';

// ========== 论坛相关 ==========

// 获取论坛分类
export const getForumCategories = (): Promise<ForumCategory[]> => {
  return request.get('/community/forum/categories');
};

// 获取论坛帖子列表
export const getForumPosts = (params: {
  category_id?: string;
  user_id?: string;
  query?: string;
  sort_by?: 'latest' | 'popular' | 'replies';
  page?: number;
  limit?: number;
}): Promise<{ data: ForumPost[]; meta: PaginationMeta }> => {
  return request.get('/community/forum/posts', { params });
};

// 获取单个论坛帖子
export const getForumPost = (id: number): Promise<{ data: ForumPost }> => {
  return request.get(`/community/forum/posts/${id}`);
};

// 创建论坛帖子（需要登录）
export const createForumPost = (data: {
  category_id: string;
  title: string;
  content: string;
}): Promise<ForumPost> => {
  return request.post('/community/forum/posts', data);
};

// 获取论坛回复列表
export const getForumReplies = (
  postId: number,
  params: {
    page?: number;
    limit?: number;
  }
): Promise<{ data: ForumReply[]; meta: PaginationMeta }> => {
  return request.get(`/community/forum/posts/${postId}/replies`, { params });
};

// 创建论坛回复（需要登录）
export const createForumReply = (postId: number, data: {
  content: string;
}): Promise<{ data: ForumReply }> => {
  return request.post(`/community/forum/posts/${postId}/replies`, data);
};

// ========== 博客相关 ==========

// 获取博客分类
export const getBlogCategories = (): Promise<BlogCategory[]> => {
  return request.get('/community/blog/categories');
};

// 获取博客文章列表
export const getBlogPosts = (params: {
  category_id?: string;
  user_id?: string;
  query?: string;
  sort_by?: 'latest' | 'popular' | 'views';
  page?: number;
  limit?: number;
}): Promise<{ data: BlogPost[]; meta: PaginationMeta }> => {
  return request.get('/community/blog/posts', { params });
};

// 获取单个博客文章
export const getBlogPost = (id: string): Promise<BlogPost> => {
  return request.get(`/community/blog/posts/${id}`);
};

// 创建博客文章（需要登录）
export const createBlogPost = (data: {
  category_id: string;
  title: string;
  content: string;
  summary?: string;
  cover_image?: string;
  tags?: string;
  is_published: boolean;
}): Promise<BlogPost> => {
  return request.post('/community/blog/posts', data);
};

// ========== 代码分享相关 ==========

// 获取代码片段列表
export const getCodeSnippets = (params: {
  language?: string;
  user_id?: string;
  query?: string;
  sort_by?: 'latest' | 'popular' | 'likes';
  page?: number;
  limit?: number;
}): Promise<{ data: CodeSnippet[]; meta: PaginationMeta }> => {
  return request.get('/community/code/snippets', { params });
};

// 创建代码片段（需要登录）
export const createCodeSnippet = (data: {
  title: string;
  description?: string;
  code: string;
  language: string;
  tags?: string;
  is_public: boolean;
}): Promise<CodeSnippet> => {
  return request.post('/community/code/snippets', data);
};

// ========== 通用操作 ==========

// 点赞/取消点赞（需要登录）
export const toggleLike = (
  target_type: 'forum_post' | 'forum_reply' | 'blog_post' | 'code_snippet',
  target_id: string
): Promise<{ is_liked: boolean; message: string }> => {
  return request.post('/community/like', {
    target_type,
    target_id
  });
};

// 获取社区统计数据
export const getCommunityStats = (): Promise<{
  total_posts: number;
  total_blogs: number;
  total_snippets: number;
  total_users: number;
  active_users: number;
}> => {
  return request.get('/community/stats');
};

// 统一导出社区服务
export const communityService = {
  // 论坛相关
  getForumCategories,
  getForumPosts,
  getForumPost,
  createForumPost,
  getForumReplies,
  createForumReply,
  
  // 博客相关
  getBlogCategories,
  getBlogPosts,
  getBlogPost,
  createBlogPost,
  
  // 代码分享相关
  getCodeSnippets,
  createCodeSnippet,
  
  // 通用操作
  toggleLike,
  getCommunityStats
};