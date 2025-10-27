import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { PaginatedResponse } from '../types';

// API客户端配置接口
export interface ApiClientConfig {
  baseURL: string;
  timeout?: number;
  enableLogging?: boolean;
  onUnauthorized?: () => void;
  onError?: (error: any) => void;
}

// 请求元数据
interface RequestMetadata {
  startTime: Date;
  requestId: string;
}

// 生成请求ID
const generateRequestId = (): string => {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
};

// 统一API客户端类
export class ApiClient {
  private instance: AxiosInstance;
  private config: ApiClientConfig;

  constructor(config: ApiClientConfig) {
    this.config = {
      timeout: 30000,
      enableLogging: true,
      ...config,
    };

    this.instance = axios.create({
      baseURL: this.config.baseURL,
      timeout: this.config.timeout,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    // 请求拦截器
    this.instance.interceptors.request.use(
      (config) => {
        // 添加认证token
        const token = this.getToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // 添加请求ID和时间戳
        const requestId = generateRequestId();
        config.headers['X-Request-ID'] = requestId;
        
        // 添加元数据用于日志记录
        (config as any).metadata = {
          startTime: new Date(),
          requestId,
        } as RequestMetadata;

        if (this.config.enableLogging) {
          console.log(`[API] ${config.method?.toUpperCase()} ${config.url} - Request ID: ${requestId}`);
        }

        return config;
      },
      (error) => {
        console.error('[API] Request error:', error);
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.instance.interceptors.response.use(
      (response: AxiosResponse<any>) => {
        const metadata = (response.config as any).metadata as RequestMetadata;
        
        if (this.config.enableLogging && metadata) {
          const duration = new Date().getTime() - metadata.startTime.getTime();
          console.log(
            `[API] ${response.config.method?.toUpperCase()} ${response.config.url} - ` +
            `${response.status} - ${duration}ms - Request ID: ${metadata.requestId}`
          );
        }

        return this.handleResponse(response);
      },
      (error) => {
        const metadata = (error.config as any)?.metadata as RequestMetadata;
        
        if (this.config.enableLogging && metadata) {
          const duration = new Date().getTime() - metadata.startTime.getTime();
          console.error(
            `[API] ${error.config?.method?.toUpperCase()} ${error.config?.url} - ` +
            `Error - ${duration}ms - Request ID: ${metadata.requestId}`,
            error
          );
        }

        return this.handleError(error);
      }
    );
  }

  private handleResponse(response: AxiosResponse<any>): any {
    const payload = response.data;

    // 分页响应：{ data: [], meta: { ... } }
    if (this.isPaginatedResponse(payload)) {
      return payload as PaginatedResponse;
    }

    // 后端统一包裹：{ code: number, message: string, data?: any, timestamp: string }
    if (this.isApiResponse(payload)) {
      if (payload.code >= 400) {
        const errorMessage = payload.message || '请求失败';
        throw new Error(errorMessage);
      }
      return payload.data !== undefined ? payload.data : payload;
    }

    // 兼容旧版：{ success: boolean, data?: any, message?: string }
    if (this.isLegacyResponse(payload)) {
      if (!payload.success) {
        const errorMessage = payload.message || '请求失败';
        throw new Error(errorMessage);
      }
      return payload.data !== undefined ? payload.data : payload;
    }

    // 直接返回数据
    return payload;
  }

  private handleError(error: any): Promise<never> {
    if (!error.response) {
      // 网络错误
      const networkError = new Error('网络连接失败，请检查网络设置');
      this.config.onError?.(networkError);
      return Promise.reject(networkError);
    }

    const { status, data } = error.response;
    let errorMessage = '';

    switch (status) {
      case 401:
        errorMessage = data?.message || '认证失败，请重新登录';
        this.handleUnauthorized();
        break;
      case 403:
        errorMessage = data?.message || '没有权限访问该资源';
        break;
      case 404:
        errorMessage = data?.message || '请求的资源不存在';
        break;
      case 422:
        errorMessage = data?.message || '请求参数验证失败';
        break;
      case 500:
        errorMessage = data?.message || '服务器内部错误';
        break;
      default:
        errorMessage = data?.message || `请求失败 (${status})`;
    }

    const apiError = new Error(errorMessage);
    (apiError as any).status = status;
    (apiError as any).data = data;

    this.config.onError?.(apiError);
    return Promise.reject(apiError);
  }

  private handleUnauthorized(): void {
    // 清除认证信息
    this.clearToken();
    
    // 执行未授权回调
    this.config.onUnauthorized?.();
    
    // 如果不在登录页面，则重定向到登录页
    if (!window.location.pathname.includes('/login')) {
      window.location.href = '/login';
    }
  }

  private isPaginatedResponse(payload: any): boolean {
    return payload && 
           typeof payload === 'object' && 
           Array.isArray(payload.data) && 
           payload.meta &&
           typeof payload.meta.page === 'number';
  }

  private isApiResponse(payload: any): boolean {
    return payload && 
           typeof payload === 'object' && 
           typeof payload.code === 'number' && 
           typeof payload.message === 'string';
  }

  private isLegacyResponse(payload: any): boolean {
    return payload && 
           typeof payload === 'object' && 
           'success' in payload;
  }

  private getToken(): string | null {
    return localStorage.getItem('token');
  }

  private clearToken(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
  }

  // 公共API方法
  public get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.get(url, config);
  }

  public post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.post(url, data, config);
  }

  public put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.put(url, data, config);
  }

  public patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.patch(url, data, config);
  }

  public delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.delete(url, config);
  }

  // 文件上传
  public upload<T = any>(url: string, file: File, onProgress?: (progress: number) => void): Promise<T> {
    const formData = new FormData();
    formData.append('file', file);

    return this.instance.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  }

  // 获取原始axios实例（用于特殊需求）
  public getInstance(): AxiosInstance {
    return this.instance;
  }

  // 更新配置
  public updateConfig(config: Partial<ApiClientConfig>): void {
    this.config = { ...this.config, ...config };
    this.instance.defaults.baseURL = this.config.baseURL;
    this.instance.defaults.timeout = this.config.timeout;
  }
}

// 创建默认API客户端实例
export const createApiClient = (config: ApiClientConfig): ApiClient => {
  return new ApiClient(config);
};

// 导出类型
export type { RequestMetadata };