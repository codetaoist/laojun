import { request } from './api';

// 验证码配置接口
export interface CaptchaConfig {
  enabled: boolean;
  type: string;
}

// 验证码接口
export interface Captcha {
  image: string;
  key: string;
}

// 认证相关 API
export const authService = {
  // 获取验证码配置
  getCaptchaConfig: async (): Promise<CaptchaConfig> => {
    try {
      const response = await request.get('/auth/captcha/config');
      return response;
    } catch (error) {
      console.error('获取验证码配置失败:', error);
      // 默认启用验证码作为后备方案
      return { enabled: true, type: 'image' };
    }
  },

  // 获取验证码
  getCaptcha: (): Promise<Captcha> => {
    return request.get('/auth/captcha');
  },
};

export default authService;