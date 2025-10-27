export type ReviewStatus = 
  | 'pending'
  | 'in_review'
  | 'approved'
  | 'rejected'
  | 'needs_modification'
  | 'appealed'
  | 'appeal_approved'
  | 'appeal_rejected'
  | 'withdrawn';

export type ReviewType = 
  | 'initial'
  | 'update'
  | 'appeal'
  | 'security'
  | 'compliance'
  | 'quality';

export type ReviewResult = 
  | 'approved'
  | 'rejected'
  | 'needs_modification'
  | 'pending';

export type ReviewPriority = 
  | 'low'
  | 'normal'
  | 'high'
  | 'urgent';

export type AppealStatus = 
  | 'pending'
  | 'in_review'
  | 'approved'
  | 'rejected';

export interface PluginReview {
  id: string;
  pluginId: string;
  pluginVersionId?: string;
  reviewType: ReviewType;
  status: ReviewStatus;
  result: ReviewResult;
  priority: ReviewPriority;
  reviewerId?: string;
  assignedAt?: string;
  reviewedAt?: string;
  notes?: string;
  rejectionReason?: string;
  autoReviewScore?: number;
  autoReviewResult?: string;
  estimatedReviewTime?: number;
  actualReviewTime?: number;
  createdAt: string;
  updatedAt: string;
  
  // 关联数据
  plugin?: Plugin;
  pluginVersion?: PluginVersion;
  reviewer?: User;
}

export interface DeveloperAppeal {
  id: string;
  pluginId: string;
  pluginReviewId: string;
  developerId: string;
  reason: string;
  additionalInfo?: string;
  status: AppealStatus;
  processedBy?: string;
  processedAt?: string;
  decision?: 'approved' | 'rejected';
  processingNotes?: string;
  createdAt: string;
  updatedAt: string;
  
  // 关联数据
  plugin?: Plugin;
  pluginReview?: PluginReview;
  developer?: User;
  processor?: User;
}

export interface ReviewerWorkload {
  id: string;
  reviewerId: string;
  currentLoad: number;
  maxCapacity: number;
  averageReviewTime: number;
  completedReviews: number;
  pendingReviews: number;
  lastAssignedAt?: string;
  isActive: boolean;
  specialties: string[];
  createdAt: string;
  updatedAt: string;
  
  // 关联数据
  reviewer?: User;
}

export interface ReviewConfig {
  id: string;
  autoReviewEnabled: boolean;
  autoReviewThreshold: number;
  maxReviewTime: number;
  priorityWeights: Record<string, number>;
  reviewerAssignmentStrategy: 'round_robin' | 'workload_based' | 'specialty_based';
  notificationSettings: Record<string, any>;
  qualityThresholds: Record<string, number>;
  escalationRules: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface ReviewTemplate {
  id: string;
  name: string;
  description?: string;
  reviewType: ReviewType;
  checklistItems: ChecklistItem[];
  isActive: boolean;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  
  // 关联数据
  creator?: User;
}

export interface ChecklistItem {
  id: string;
  title: string;
  description?: string;
  isRequired: boolean;
  category: string;
  weight: number;
  order: number;
}

export interface AutoReviewLog {
  id: string;
  pluginId: string;
  pluginVersionId?: string;
  score: number;
  result: string;
  details: Record<string, any>;
  executedAt: string;
  executionTime: number;
  
  // 关联数据
  plugin?: Plugin;
  pluginVersion?: PluginVersion;
}

export interface PluginVersionReview {
  id: string;
  pluginId: string;
  pluginVersionId: string;
  reviewId: string;
  isCurrentVersion: boolean;
  createdAt: string;
  
  // 关联数据
  plugin?: Plugin;
  pluginVersion?: PluginVersion;
  review?: PluginReview;
}

export interface ReviewStats {
  totalReviews: number;
  pendingReviews: number;
  approvedReviews: number;
  rejectedReviews: number;
  averageReviewTime: number;
  reviewerCount: number;
  appealCount: number;
  autoReviewCount: number;
  qualityScore: number;
  trendsData: TrendData[];
}

export interface TrendData {
  date: string;
  reviews: number;
  approvals: number;
  rejections: number;
  averageTime: number;
}

export interface QueueStats {
  totalInQueue: number;
  highPriority: number;
  normalPriority: number;
  lowPriority: number;
  urgentPriority: number;
  averageWaitTime: number;
  oldestInQueue: string;
  estimatedProcessingTime: number;
}

export interface PerformanceMetrics {
  reviewerPerformance: ReviewerPerformance[];
  systemPerformance: SystemPerformance;
  qualityMetrics: QualityMetrics;
}

export interface ReviewerPerformance {
  reviewerId: string;
  reviewerName: string;
  totalReviews: number;
  averageReviewTime: number;
  approvalRate: number;
  qualityScore: number;
  efficiency: number;
  specialties: string[];
}

export interface SystemPerformance {
  throughput: number;
  averageProcessingTime: number;
  queueLength: number;
  autoReviewAccuracy: number;
  escalationRate: number;
  appealRate: number;
}

export interface QualityMetrics {
  overallQualityScore: number;
  codeQualityScore: number;
  securityScore: number;
  performanceScore: number;
  usabilityScore: number;
  documentationScore: number;
}

export interface ReviewSuggestion {
  category: string;
  title: string;
  description: string;
  severity: 'info' | 'warning' | 'error';
  autoFixable: boolean;
  recommendation: string;
}

export interface NotificationSettings {
  emailEnabled: boolean;
  smsEnabled: boolean;
  pushEnabled: boolean;
  reviewAssigned: boolean;
  reviewCompleted: boolean;
  appealCreated: boolean;
  appealProcessed: boolean;
  urgentReview: boolean;
  deadlineReminder: boolean;
  reminderIntervals: number[];
}

export interface ReviewCalendarEvent {
  id: string;
  pluginId: string;
  pluginName: string;
  reviewerId: string;
  reviewerName: string;
  scheduledAt: string;
  estimatedDuration: number;
  priority: ReviewPriority;
  status: 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
  notes?: string;
}

export interface ReviewRule {
  id: string;
  name: string;
  description: string;
  condition: string;
  action: string;
  priority: number;
  isActive: boolean;
  category: string;
  parameters: Record<string, any>;
}

// 基础类型（从其他文件引用）
export interface Plugin {
  id: string;
  name: string;
  description: string;
  version: string;
  developerId: string;
  categoryId: string;
  status: string;
  reviewStatus: ReviewStatus;
  reviewPriority: ReviewPriority;
  autoReviewScore?: number;
  autoReviewResult?: string;
  reviewNotes?: string;
  reviewedAt?: string;
  reviewerId?: string;
  submittedForReviewAt?: string;
  rejectionReason?: string;
  appealCount: number;
  lastAppealAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PluginVersion {
  id: string;
  pluginId: string;
  version: string;
  changelog: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  avatar?: string;
  role: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

// 请求和响应类型
export interface ReviewQueueParams {
  status?: ReviewStatus;
  priority?: ReviewPriority;
  reviewerId?: string;
  dateRange?: [string, string];
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface ReviewRequest {
  result: ReviewResult;
  notes: string;
  rejectionReason?: string;
  checklistResults?: Record<string, boolean>;
  qualityScores?: Record<string, number>;
  estimatedFixTime?: number;
}

export interface BatchReviewRequest {
  pluginIds: string[];
  result: ReviewResult;
  notes: string;
  rejectionReason?: string;
}

export interface AppealRequest {
  reason: string;
  additionalInfo?: string;
  attachments?: string[];
}

export interface AppealProcessRequest {
  decision: 'approved' | 'rejected';
  notes: string;
  actionRequired?: string;
}

export interface AssignReviewerRequest {
  reviewerId: string;
  priority?: ReviewPriority;
  deadline?: string;
  notes?: string;
}

// 分页响应类型
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// API响应类型
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
  code?: number;
}