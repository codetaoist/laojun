import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { App } from 'antd';
import { ApiResponse } from '@/types';

// 创建 axios 实例
const apiBase: string = (import.meta as any)?.env?.VITE_ADMIN_API_BASE || '/api/v1';
const api: AxiosInstance = axios.create({
  baseURL: apiBase,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: false,
});

// 消息实例
let messageApi: any = null;

// 设置消息实例
export const setMessageApi = (api: any) => {
  messageApi = api;
};

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 添加认证 token
    const token = localStorage.getItem('token');
    if (token) {
      (config.headers as any).Authorization = `Bearer ${token}`;
    }
    
    // 添加请求 ID 用于追踪
    (config.headers as any)['X-Request-ID'] = generateRequestId();
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response;
    
    // 兼容后端响应：只有在明确存在 success 且为 false 时才判定失败
    if (data && Object.prototype.hasOwnProperty.call(data, 'success') && (data as any).success === false) {
      if (messageApi) {
        messageApi.error((data as any).message || '请求失败');
      }
      return Promise.reject(new Error((data as any).message || '请求失败'));
    }
    
    return response;
  },
  (error) => {
    // 处理 HTTP 错误状态码
    if (error.response) {
      const { status, data } = error.response;
      
      switch (status) {
        case 401: {
          const isOnLogin = window.location.pathname === '/login';
          // 登录页的 401 多为"账号或密码错误"等业务错误，避免全局提示与整页刷新
          if (messageApi && !isOnLogin) {
            messageApi.error('登录已过期，请重新登录');
          }
          // 清除本地存储的认证信息
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          // 非登录页才进行整页跳转
          if (!isOnLogin) {
            window.location.href = '/login';
          }
          break;
        }
        case 403:
          if (messageApi) {
            messageApi.error('没有权限访问该资源');
          }
          break;
        case 404:
          if (messageApi) {
            messageApi.error('请求的资源不存在');
          }
          break;
        case 500:
          if (messageApi) {
            messageApi.error('服务器内部错误');
          }
          break;
        default:
          if (messageApi) {
            messageApi.error(data?.message || `请求失败 (${status})`);
          }
      }
    } else if (error.request) {
      if (messageApi) {
        messageApi.error('网络连接失败，请检查网络设置');
      }
    } else {
      if (messageApi) {
        messageApi.error('请求配置错误');
      }
    }
    
    return Promise.reject(error);
  }
);

// 生成请求 ID
function generateRequestId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

// 通用请求方法
export const request = {
  get: <T = any>(url: string, config?: AxiosRequestConfig) =>
    api.get<ApiResponse<T>>(url, config).then(res => res.data.data),
    
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    api.post<ApiResponse<T>>(url, data, config).then(res => res.data.data),
    
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    api.put<ApiResponse<T>>(url, data, config).then(res => res.data.data),
    
  delete: <T = any>(url: string, config?: AxiosRequestConfig) =>
    api.delete<ApiResponse<T>>(url, config).then(res => res.data.data),
    
  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    api.patch<ApiResponse<T>>(url, data, config).then(res => res.data.data),
};

export default api;