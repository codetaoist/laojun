import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import enUS from 'antd/locale/en_US';
import { router } from './router';
import { useAppStore } from './stores/app';
import { setMessageApi } from './services/api';
import ErrorBoundary from './components/ErrorBoundary';
import AccessibilityProvider from './components/AccessibilityProvider';
import AccessibilityPanel from './components/AccessibilityPanel';
import { useTranslation } from './locales';
import './index.css';

// 主应用组件
const MainApp: React.FC = () => {
  const { theme: appTheme } = useAppStore();
  const { locale } = useTranslation();

  // 主题配置
  const themeConfig = {
    algorithm: appTheme === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      colorPrimary: '#1890ff',
      borderRadius: 6,
    },
  };

  // 根据当前语言设置 Ant Design 的 locale
  const antdLocale = locale === 'en-US' ? enUS : zhCN;

  return (
    <AccessibilityProvider>
      <ConfigProvider
        locale={antdLocale}
        theme={themeConfig}
      >
        <App>
          <AppContent />
          <AccessibilityPanel />
        </App>
      </ConfigProvider>
    </AccessibilityProvider>
  );
};

// 应用内容组件
const AppContent: React.FC = () => {
  const { message } = App.useApp();

  // 设置全局message API
  React.useEffect(() => {
    setMessageApi(message);
  }, [message]);

  return (
    <ErrorBoundary>
      <RouterProvider router={router} />
    </ErrorBoundary>
  );
};

// 渲染应用
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <MainApp />
  </React.StrictMode>
);