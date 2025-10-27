// 插件相关类型
export interface Plugin {
  id: string;
  name: string;
  description: string;
  version: string;
  author: string;
  authorEmail?: string;
  category: string;
  tags: string[];
  icon?: string;
  screenshots?: string[];
  downloadUrl: string;
  documentationUrl?: string;
  repositoryUrl?: string;
  license: string;
  price: number; // 0 表示免费
  currency: string;
  downloads: number;
  rating: number;
  reviewCount: number;
  size: number; // 文件大小，单位：字节
  requirements: {
    minVersion: string;
    maxVersion?: string;
    dependencies?: string[];
  };
  status: 'active' | 'inactive' | 'pending' | 'rejected';
  featured: boolean;
  createdAt: string;
  updatedAt: string;
  publishedAt?: string;
}

// 插件分类
export interface Category {
  id: string;
  name: string;
  description: string;
  icon?: string;
  pluginCount: number;
  parentId?: string;
  children?: Category[];
}

// 插件评论
export interface Review {
  id: string;
  pluginId: string;
  userId: string;
  userName: string;
  userAvatar?: string;
  rating: number;
  title: string;
  content: string;
  helpful: number;
  createdAt: string;
  updatedAt: string;
}

// 用户信息
export interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  avatar?: string;
  bio?: string;
  website?: string;
  location?: string;
  joinedAt: string;
}

// 开发者信息
export interface Developer {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  bio?: string;
  website?: string;
  verified: boolean;
  pluginCount: number;
  totalDownloads: number;
  averageRating: number;
  joinedAt: string;
}

// 搜索参数
export interface SearchParams {
  query?: string;
  category?: string;
  tags?: string[];
  author?: string;
  priceRange?: [number, number];
  rating?: number;
  sortBy?: 'relevance' | 'downloads' | 'rating' | 'updated' | 'created' | 'name';
  sortOrder?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
  featured?: boolean;
  free?: boolean;
}

// 分页响应
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// API 响应
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  code?: string;
}

// 下载统计
export interface DownloadStats {
  pluginId: string;
  totalDownloads: number;
  dailyDownloads: { date: string; count: number }[];
  weeklyDownloads: { week: string; count: number }[];
  monthlyDownloads: { month: string; count: number }[];
}

// 安装状态
export interface InstallationStatus {
  pluginId: string;
  status: 'installing' | 'installed' | 'failed' | 'updating' | 'uninstalling';
  progress?: number;
  error?: string;
  installedVersion?: string;
  availableVersion?: string;
}

// 购买记录
export interface Purchase {
  id: string;
  pluginId: string;
  userId: string;
  amount: number;
  currency: string;
  status: 'pending' | 'completed' | 'failed' | 'refunded';
  paymentMethod: string;
  transactionId?: string;
  createdAt: string;
  completedAt?: string;
}

// 收藏
export interface Favorite {
  id: string;
  pluginId: string;
  userId: string;
  createdAt: string;
}

// 插件版本历史
export interface PluginVersion {
  id: string;
  pluginId: string;
  version: string;
  changelog: string;
  downloadUrl: string;
  size: number;
  requirements: {
    minVersion: string;
    maxVersion?: string;
    dependencies?: string[];
  };
  status: 'active' | 'deprecated' | 'beta';
  createdAt: string;
}

// 应用状态
export interface AppState {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  searchHistory: string[];
  viewMode: 'grid' | 'list';
  pageSize: number;
}

// 购物车项目
export interface CartItem {
  pluginId: string;
  plugin: Plugin;
  quantity: number;
  addedAt: string;
}

// 购物车状态
export interface CartState {
  items: CartItem[];
  total: number;
  currency: string;
}

// 订单状态
export type OrderStatus = 'pending' | 'processing' | 'completed' | 'cancelled' | 'refunded';

// 支付状态
export type PaymentStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled' | 'refunded';

// 支付方式
export type PaymentMethod = 'credit-card' | 'paypal' | 'alipay' | 'wechat' | 'bank-transfer';

// 账单信息
export interface BillingInfo {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  address: string;
  city: string;
  state: string;
  zipCode: string;
  country: string;
}

// 订单项目
export interface OrderItem {
  id: string;
  pluginId: string;
  plugin: Plugin;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
}

// 订单
export interface Order {
  id: string;
  userId: string;
  items: OrderItem[];
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  currency: string;
  status: OrderStatus;
  paymentStatus: PaymentStatus;
  paymentMethod: PaymentMethod;
  billingInfo: BillingInfo;
  notes?: string;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
  cancelledAt?: string;
}

// 支付信息
export interface PaymentInfo {
  orderId: string;
  amount: number;
  currency: string;
  method: PaymentMethod;
  transactionId?: string;
  gatewayResponse?: any;
  status: PaymentStatus;
  createdAt: string;
  completedAt?: string;
  failureReason?: string;
}

// 退款信息
export interface RefundInfo {
  id: string;
  orderId: string;
  paymentId: string;
  amount: number;
  currency: string;
  reason: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  createdAt: string;
  completedAt?: string;
}

// 订单状态管理
export interface OrderState {
  orders: Order[];
  currentOrder?: Order;
  loading: boolean;
  error?: string;
}

// 下载记录
export interface DownloadRecord {
  id: string;
  userId: string;
  pluginId: string;
  orderId?: string;
  downloadUrl: string;
  downloadedAt: string;
  ipAddress: string;
  userAgent: string;
}

// ========== 社区相关类型 ==========

// 论坛分类
export interface ForumCategory {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  color: string;
  sort_order: number;
  parent_id?: string;
  is_active: boolean;
  post_count: number;
  last_post_id?: string;
  created_at: string;
  updated_at: string;
}

// 论坛帖子
export interface ForumPost {
  id: string;
  category_id: string;
  user_id: string;
  title: string;
  content: string;
  likes_count: number;
  replies_count: number;
  views_count: number;
  is_pinned: boolean;
  is_locked: boolean;
  created_at: string;
  updated_at: string;
  // 关联数据
  username: string;
  avatar_url?: string;
  category_name: string;
}

// 论坛回复
export interface ForumReply {
  id: string;
  post_id: string;
  user_id: string;
  user?: User;
  content: string;
  content_type: 'markdown' | 'html' | 'text';
  like_count: number;
  created_at: string;
  updated_at: string;
}

// 博客分类
export interface BlogCategory {
  id: string;
  name: string;
  description?: string;
  color: string;
  sort_order: number;
  is_active: boolean;
  post_count: number;
  created_at: string;
  updated_at: string;
}

// 博客文章
export interface BlogPost {
  id: string;
  title: string;
  content: string;
  summary?: string;
  cover_image?: string;
  author_id: string;
  author?: User;
  category_id?: string;
  category?: BlogCategory;
  tags?: string;
  is_published: boolean;
  view_count: number;
  like_count: number;
  comment_count: number;
  published_at?: string;
  created_at: string;
  updated_at: string;
}

// 博客评论
export interface BlogComment {
  id: string;
  post_id: string;
  user_id: string;
  user?: User;
  content: string;
  like_count: number;
  created_at: string;
  updated_at: string;
}

// 代码片段
export interface CodeSnippet {
  id: string;
  title: string;
  description?: string;
  code: string;
  language: string;
  user_id: string;
  user?: User;
  tags?: string;
  is_public: boolean;
  view_count: number;
  like_count: number;
  fork_count: number;
  created_at: string;
  updated_at: string;
}

// 点赞记录
export interface Like {
  id: string;
  user_id: string;
  target_type: 'forum_post' | 'forum_reply' | 'blog_post' | 'blog_comment' | 'code_snippet';
  target_id: string;
  created_at: string;
}

// 用户关注
export interface UserFollow {
  id: string;
  follower_id: string;
  following_id: string;
  created_at: string;
}

// 书签收藏
export interface Bookmark {
  id: string;
  user_id: string;
  target_type: 'forum_post' | 'blog_post' | 'code_snippet';
  target_id: string;
  created_at: string;
}

// 用户积分
export interface UserPoints {
  id: string;
  user_id: string;
  total_points: number;
  available_points: number;
  level: number;
  updated_at: string;
}

// 积分记录
export interface PointRecord {
  id: string;
  user_id: string;
  type: 'earn' | 'spend';
  points: number;
  reason: string;
  reference_type?: string;
  reference_id?: string;
  created_at: string;
}

// 徽章
export interface Badge {
  id: string;
  name: string;
  description: string;
  icon: string;
  color: string;
  type: 'achievement' | 'contribution' | 'special';
  condition_type: string;
  condition_value: number;
  is_active: boolean;
  created_at: string;
}

// 用户徽章
export interface UserBadge {
  id: string;
  user_id: string;
  badge_id: string;
  badge?: Badge;
  earned_at: string;
}

// 私信
export interface Message {
  id: string;
  sender_id: string;
  sender?: User;
  receiver_id: string;
  receiver?: User;
  title: string;
  content: string;
  is_read: boolean;
  created_at: string;
}

// 通知
export interface Notification {
  id: string;
  user_id: string;
  type: 'like' | 'reply' | 'follow' | 'mention' | 'system';
  title: string;
  content?: string;
  target_type?: string;
  target_id?: string;
  is_read: boolean;
  created_at: string;
}

// 分页元数据
export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}