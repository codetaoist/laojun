import { request } from './api';

export interface ReviewQueueParams {
  status?: string;
  priority?: string;
  reviewerId?: string;
  dateRange?: [string, string];
  page?: number;
  pageSize?: number;
}

export interface ReviewRequest {
  result: 'approved' | 'rejected' | 'needs_modification';
  notes: string;
  rejectionReason?: string;
}

export interface BatchReviewRequest {
  pluginIds: string[];
  result: 'approved' | 'rejected' | 'needs_modification';
  notes: string;
  rejectionReason?: string;
}

export interface AppealRequest {
  reason: string;
  additionalInfo?: string;
}

export interface AppealProcessRequest {
  decision: 'approved' | 'rejected';
  notes: string;
}

export interface AssignReviewerRequest {
  reviewerId: string;
}

export const pluginReviewService = {
  // 获取审核队列
  getReviewQueue: (params: ReviewQueueParams = {}) => {
    return request.get('/plugin-review/queue', { params });
  },

  // 分配审核员
  assignReviewer: (pluginId: string, reviewerId: string) => {
    return request.post(`/plugin-review/assign/${pluginId}`, { reviewerId });
  },

  // 审核插件
  reviewPlugin: (pluginId: string, data: ReviewRequest) => {
    return request.post(`/plugin-review/review/${pluginId}`, data);
  },

  // 批量审核插件
  batchReviewPlugins: (data: BatchReviewRequest) => {
    return request.post('/plugin-review/batch-review', data);
  },

  // 获取插件审核历史
  getPluginReviewHistory: (pluginId: string) => {
    return request.get(`/plugin-review/history/${pluginId}`);
  },

  // 创建申诉
  createAppeal: (pluginId: string, data: AppealRequest) => {
    return request.post(`/plugin-review/appeal/${pluginId}`, data);
  },

  // 处理申诉
  processAppeal: (appealId: string, data: AppealProcessRequest) => {
    return request.post(`/plugin-review/appeal/${appealId}/process`, data);
  },

  // 获取申诉详情
  getAppeal: (appealId: string) => {
    return request.get(`/plugin-review/appeal/${appealId}`);
  },

  // 获取申诉列表
  getAppeals: (params: any = {}) => {
    return request.get('/plugin-review/appeals', { params });
  },

  // 获取审核统计
  getReviewStats: () => {
    return request.get('/plugin-review/stats');
  },

  // 自动审核插件
  autoReviewPlugin: (pluginId: string) => {
    return request.post(`/plugin-review/auto-review/${pluginId}`);
  },

  // 获取审核员工作负载
  getReviewerWorkload: () => {
    return request.get('/plugin-review/workload');
  },

  // 获取我的审核任务
  getMyReviewTasks: () => {
    return request.get('/plugin-review/my-tasks');
  },

  // 获取审核配置
  getReviewConfig: () => {
    return request.get('/plugin-review/config');
  },

  // 更新审核配置
  updateReviewConfig: (data: any) => {
    return request.put('/plugin-review/config', data);
  },

  // 获取审核模板
  getReviewTemplates: () => {
    return request.get('/plugin-review/templates');
  },

  // 创建审核模板
  createReviewTemplate: (data: any) => {
    return request.post('/plugin-review/templates', data);
  },

  // 更新审核模板
  updateReviewTemplate: (templateId: string, data: any) => {
    return request.put(`/plugin-review/templates/${templateId}`, data);
  },

  // 删除审核模板
  deleteReviewTemplate: (templateId: string) => {
    return request.delete(`/plugin-review/templates/${templateId}`);
  },

  // 获取自动审核日志
  getAutoReviewLogs: (params: any = {}) => {
    return request.get('/plugin-review/auto-review-logs', { params });
  },

  // 获取审核员列表
  getReviewers: () => {
    return request.get('/plugin-review/reviewers');
  },

  // 导出审核报告
  exportReviewReport: (params: any = {}) => {
    return request.get('/plugin-review/export', { 
      params,
      responseType: 'blob'
    });
  },

  // 获取审核趋势数据
  getReviewTrends: (params: any = {}) => {
    return request.get('/plugin-review/trends', { params });
  },

  // 获取插件质量评分
  getPluginQualityScore: (pluginId: string) => {
    return request.get(`/plugin-review/quality-score/${pluginId}`);
  },

  // 重新评估插件
  reassessPlugin: (pluginId: string) => {
    return request.post(`/plugin-review/reassess/${pluginId}`);
  },

  // 获取审核规则
  getReviewRules: () => {
    return request.get('/plugin-review/rules');
  },

  // 更新审核规则
  updateReviewRules: (data: any) => {
    return request.put('/plugin-review/rules', data);
  },

  // 测试审核规则
  testReviewRules: (pluginId: string, rules: any) => {
    return request.post(`/plugin-review/test-rules/${pluginId}`, { rules });
  },

  // 获取审核队列统计
  getQueueStats: () => {
    return request.get('/plugin-review/queue/stats');
  },

  // 优化审核队列
  optimizeQueue: () => {
    return request.post('/plugin-review/queue/optimize');
  },

  // 获取审核性能指标
  getPerformanceMetrics: (params: any = {}) => {
    return request.get('/plugin-review/metrics', { params });
  },

  // 获取审核建议
  getReviewSuggestions: (pluginId: string) => {
    return request.get(`/plugin-review/suggestions/${pluginId}`);
  },

  // 标记为紧急审核
  markAsUrgent: (pluginId: string) => {
    return request.post(`/plugin-review/mark-urgent/${pluginId}`);
  },

  // 取消紧急标记
  unmarkAsUrgent: (pluginId: string) => {
    return request.post(`/plugin-review/unmark-urgent/${pluginId}`);
  },

  // 获取审核通知设置
  getNotificationSettings: () => {
    return request.get('/plugin-review/notifications/settings');
  },

  // 更新审核通知设置
  updateNotificationSettings: (data: any) => {
    return request.put('/plugin-review/notifications/settings', data);
  },

  // 发送审核提醒
  sendReviewReminder: (pluginId: string) => {
    return request.post(`/plugin-review/reminder/${pluginId}`);
  },

  // 获取审核日历
  getReviewCalendar: (params: any = {}) => {
    return request.get('/plugin-review/calendar', { params });
  },

  // 预约审核时间
  scheduleReview: (pluginId: string, data: any) => {
    return request.post(`/plugin-review/schedule/${pluginId}`, data);
  },

  // 取消预约审核
  cancelScheduledReview: (pluginId: string) => {
    return request.delete(`/plugin-review/schedule/${pluginId}`);
  }
};

export default pluginReviewService;