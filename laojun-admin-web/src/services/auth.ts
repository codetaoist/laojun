import { request } from './api';
import { LoginRequest, LoginResponse, User } from '@/types';

// 基于 TextEncoder/TextDecoder 的 UTF-8 Base64 编解码，兼容中文
const base64EncodeUtf8 = (str: string): string => {
  const encoder = new TextEncoder();
  const bytes = encoder.encode(str);
  let binary = '';
  bytes.forEach(b => { binary += String.fromCharCode(b); });
  return btoa(binary);
};

const base64DecodeUtf8 = (b64: string): string => {
  const binary = atob(b64);
  const bytes = new Uint8Array([...binary].map(ch => ch.charCodeAt(0)));
  const decoder = new TextDecoder();
  return decoder.decode(bytes);
};

// 在非开发模式下，避免 /auth/profile 被并发或重复调用
let inflightProfile: Promise<User> | null = null;

export const authService = {
  // 登录（支持本地固定凭据校验）
  login: (data: LoginRequest): Promise<LoginResponse> => {
    const username = (data.username || '').trim();
    const password = data.password || '';

    // 仅在开发模式且显式开启时启用本地固定凭据：admin/password
    const devEnabled = (import.meta.env?.VITE_ENABLE_DEV_LOGIN === 'true');
    if (import.meta.env?.DEV && devEnabled && username === 'admin' && password === 'password') {
      const now = Date.now();
      const expiresIn = 3600; // 1小时
      const expMs = now + expiresIn * 1000;

      const user: User = {
        id: 'dev-admin',
        username: 'admin',
        email: 'admin@laojun.local',
        name: '系统管理员',
        avatar: '',
        status: 'active',
        roles: [
          {
            id: 'role-super-admin',
            name: 'super_admin',
            description: '系统超级管理员',
            permissions: [],
            isSystem: true,
            createdAt: new Date(now).toISOString(),
            updatedAt: new Date(now).toISOString(),
          },
        ],
        createdAt: new Date(now).toISOString(),
        updatedAt: new Date(now).toISOString(),
        lastLoginAt: new Date(now).toISOString(),
      };

      const payload = { user, exp: expMs };
      const token = 'dev.' + base64EncodeUtf8(JSON.stringify(payload));
      const refreshToken = 'dev.refresh.' + Math.random().toString(36).slice(2);
      const expiresAt = new Date(expMs).toISOString();

      return Promise.resolve({ token, refreshToken, user, expiresAt });
    }

    // 走后端接口
    return request.post('/login', data);
  },

  // 登出
  logout: (): Promise<void> => {
    return request.post('/auth/logout');
  },

  // 刷新 token (暂时不可用)
  refreshToken: (refreshToken: string): Promise<LoginResponse> => {
    // return request.post('/auth/refresh', { refreshToken });
    return Promise.reject(new Error('Refresh token not implemented'));
  },

  // 获取当前用户信息（本地令牌优先；并发请求去重）
  getCurrentUser: (): Promise<User> => {
    const token = localStorage.getItem('token') || '';
    if (token.startsWith('dev.')) {
      try {
        const payload = JSON.parse(base64DecodeUtf8(token.slice(4)));
        return Promise.resolve(payload.user as User);
      } catch (e) {
        return Promise.reject(new Error('Invalid local token'));
      }
    }
    // 非开发模式下，通过单例 Promise 去重
    if (!inflightProfile) {
      inflightProfile = request.get<User>('/auth/profile')
        .finally(() => { inflightProfile = null; });
    }
    return inflightProfile;
  },

  // 修改密码 (暂时不可用)
  changePassword: (data: {
    oldPassword: string;
    newPassword: string;
  }): Promise<void> => {
    // return request.post('/auth/change-password', data);
    return Promise.reject(new Error('Change password not implemented'));
  },

  // 获取验证码配置
  getCaptchaConfig: async (): Promise<{ enabled: boolean; type: string }> => {
    try {
      const result = await request.get('/auth/captcha/config');
      return result.data;
    } catch (error) {
      console.error('获取验证码配置失败:', error)
      // 默认返回启用状态，保持向后兼容
      return { enabled: true, type: 'image' };
    }
  },

  // 获取验证码配置
  getCaptchaConfig: async (): Promise<{ enabled: boolean; type: string }> => {
    try {
      const result = await request.get('/auth/captcha/config');
      console.log('验证码配置API响应:', result);
      // 后端返回格式: { success: true, data: { enabled: boolean, type: string } }
      return result.data || result;
    } catch (error) {
      console.error('获取验证码配置失败:', error);
      throw error;
    }
  },

  // 获取验证码
  getCaptcha: async (): Promise<{ image: string; key: string }> => {
    console.log('调用验证码API...')
    try {
      const result = await request.get('/auth/captcha');
      console.log('验证码API原始响应:', result)
      return result;
    } catch (error) {
      console.error('验证码API调用失败:', error)
      throw error;
    }
  },

  // 验证 token 有效性（本地令牌优先；复用用户信息请求）
  validateToken: (): Promise<boolean> => {
    const token = localStorage.getItem('token') || '';
    if (token.startsWith('dev.')) {
      try {
        const payload = JSON.parse(base64DecodeUtf8(token.slice(4)));
        return Promise.resolve(typeof payload.exp === 'number' && Date.now() < payload.exp);
      } catch (e) {
        return Promise.resolve(false);
      }
    }
    // 后端未提供 /validate，直接复用 getCurrentUser 的结果，避免重复发起 /auth/profile
    return authService.getCurrentUser().then(() => true).catch(() => false);
  },
};