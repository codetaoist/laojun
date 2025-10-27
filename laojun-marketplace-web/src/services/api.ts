import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

import { ApiResponse } from '@/types';

// 创建 axios 实例
const api: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 添加认证 token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // 添加请求时间戳
    (config as any).metadata = { startTime: new Date() };

    return config;
  },
  (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
  }
);

// 响应拦截器（兼容后端 APIResponse 与分页响应格式）
api.interceptors.response.use(
  (response: AxiosResponse<any>) => {
    // 计算请求耗时
    const endTime = new Date();
    const duration = endTime.getTime() - (((response.config as any).metadata?.startTime?.getTime()) || 0);
    console.log(`API ${response.config.method?.toUpperCase()} ${response.config.url} took ${duration}ms`);

    const payload = response.data;

    // 分页响应：{ data: [], meta: { ... } }
    const isPaginated = payload && typeof payload === 'object' && Array.isArray(payload.data) && payload.meta;
    if (isPaginated) {
      return payload;
    }

    // 后端统一包裹：{ code: number, message: string, data?: any, timestamp: string }
    const hasApiEnvelope = payload && typeof payload === 'object' && typeof payload.code === 'number' && typeof payload.message === 'string';
    if (hasApiEnvelope) {
      if (payload.code >= 400) {
        const errorMessage = payload.message || '请求失败';
        console.warn(errorMessage);
        return Promise.reject(new Error(errorMessage));
      }
      return payload.data !== undefined ? payload.data : payload;
    }

    // 兼容旧版：{ success: boolean, data?: any, message?: string }
    const hasSuccessFlag = payload && typeof payload === 'object' && 'success' in payload;
    if (hasSuccessFlag) {
      if (!payload.success) {
        const errorMessage = payload.message || '请求失败';
        console.warn(errorMessage);
        return Promise.reject(new Error(errorMessage));
      }
      return payload.data !== undefined ? payload.data : payload;
    }

    // 兜底：直接返回 payload
    return payload;
  },
  (error) => {
    console.error('Response error:', error);

    // 处理网络错误
    if (!error.response) {
      console.warn('网络连接失败，请检查网络设置');
      return Promise.reject(error);
    }

    const { status, data } = error.response;

    let errorMessage = '';
    
    switch (status) {
      case 401:
        errorMessage = data?.message || '用户名或密码错误';
        console.warn(errorMessage);
        // 清除本地存储的认证信息
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
        // 只有在非登录页面时才重定向到登录页
        if (!window.location.pathname.includes('/login')) {
          window.location.href = '/login';
        }
        break;

      case 403:
        errorMessage = '没有权限访问该资源';
        console.warn(errorMessage);
        break;

      case 404:
        errorMessage = '请求的资源不存在';
        console.warn(errorMessage);
        break;

      case 422:
        // 表单验证错误
        const validationErrors = data?.errors;
        if (validationErrors && typeof validationErrors === 'object') {
          const firstError = Object.values(validationErrors)[0] as any;
          errorMessage = Array.isArray(firstError) ? firstError[0] : firstError;
        } else {
          errorMessage = data?.message || '请求参数错误';
        }
        console.warn(errorMessage);
        break;

      case 429:
        errorMessage = '请求过于频繁，请稍后再试';
        console.warn(errorMessage);
        break;

      case 500:
        errorMessage = '服务器内部错误，请稍后再试';
        console.warn(errorMessage);
        break;

      default:
        errorMessage = data?.message || '请求失败';
        console.warn(errorMessage);
    }

    // 创建一个包含错误信息的Error对象
    const customError = new Error(errorMessage);
    customError.name = 'APIError';
    (customError as any).status = status;
    (customError as any).data = data;

    return Promise.reject(customError);
  }
);

// 通用请求方法封装
export const request = {
  get: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return api.get(url, config) as unknown as Promise<T>;
  },

  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return api.post(url, data, config) as unknown as Promise<T>;
  },

  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return api.put(url, data, config) as unknown as Promise<T>;
  },

  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return api.patch(url, data, config) as unknown as Promise<T>;
  },

  delete: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return api.delete(url, config) as unknown as Promise<T>;
  },
};

export default api;

// 文件上传（带进度）
export const uploadFile = (
  url: string,
  file: File,
  onProgress?: (progress: number) => void
): Promise<any> => {
  const formData = new FormData();
  formData.append('file', file);

  return api.post(url, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (progressEvent: ProgressEvent) => {
      if (onProgress && progressEvent.total) {
        const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        onProgress(progress);
      }
    },
  });
};

// 文件下载
export const downloadFile = (url: string, filename?: string): Promise<void> => {
  return api
    .get(url, { responseType: 'blob' })
    .then((response) => {
      const blob = new Blob([response.data]);
      const link = document.createElement('a');
      link.href = window.URL.createObjectURL(blob);
      link.download = filename || 'downloaded_file';
      link.click();
    });
};