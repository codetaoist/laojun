import zhCN from './zh-CN';
import enUS from './en-US';

export type Locale = 'zh-CN' | 'en-US';

export const locales = {
  'zh-CN': zhCN,
  'en-US': enUS,
};

export const localeNames = {
  'zh-CN': '简体中文',
  'en-US': 'English',
};

export const defaultLocale: Locale = 'zh-CN';

export const getLocale = (): Locale => {
  // 从 localStorage 获取用户设置的语言
  const savedLocale = localStorage.getItem('locale') as Locale;
  if (savedLocale && locales[savedLocale]) {
    return savedLocale;
  }

  // 从浏览器语言获取
  const browserLocale = navigator.language;
  if (browserLocale.startsWith('zh')) {
    return 'zh-CN';
  }
  if (browserLocale.startsWith('en')) {
    return 'en-US';
  }

  return defaultLocale;
};

export const setLocale = (locale: Locale): void => {
  localStorage.setItem('locale', locale);
  // 触发自定义事件通知其他组件语言已更改
  window.dispatchEvent(new CustomEvent('localeChange', { detail: locale }));
};

export const t = (key: string, params?: Record<string, any>): string => {
  const locale = getLocale();
  const messages = locales[locale];
  
  // 支持嵌套键，如 'common.ok'
  const keys = key.split('.');
  let value: any = messages;
  
  for (const k of keys) {
    if (value && typeof value === 'object' && k in value) {
      value = value[k];
    } else {
      // 如果找不到翻译，返回键名
      return key;
    }
  }
  
  if (typeof value !== 'string') {
    return key;
  }
  
  // 替换参数
  if (params) {
    return value.replace(/\{\{(\w+)\}\}/g, (match, paramKey) => {
      return params[paramKey] !== undefined ? String(params[paramKey]) : match;
    });
  }
  
  return value;
};

// React Hook for internationalization
import { useState, useEffect } from 'react';

export const useTranslation = () => {
  const [locale, setLocaleState] = useState<Locale>(getLocale());

  useEffect(() => {
    const handleLocaleChange = (event: CustomEvent<Locale>) => {
      setLocaleState(event.detail);
    };

    window.addEventListener('localeChange', handleLocaleChange as EventListener);
    
    return () => {
      window.removeEventListener('localeChange', handleLocaleChange as EventListener);
    };
  }, []);

  const changeLocale = (newLocale: Locale) => {
    setLocale(newLocale);
    setLocaleState(newLocale);
  };

  return {
    locale,
    t,
    changeLocale,
    locales: localeNames,
  };
};

export default {
  locales,
  localeNames,
  defaultLocale,
  getLocale,
  setLocale,
  t,
  useTranslation,
};