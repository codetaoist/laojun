import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/zh-cn';
import router from './router';
import ErrorBoundary from './components/ErrorBoundary';
import './index.css';

// 扩展 dayjs 插件
dayjs.extend(relativeTime);

// 设置 dayjs 中文语言
dayjs.locale('zh-cn');

// Ant Design 主题配置
const theme = {
  token: {
    colorPrimary: '#1890ff',
    borderRadius: 8,
    wireframe: false,
  },
  components: {
    Layout: {
      headerBg: '#fff',
      headerHeight: 64,
      headerPadding: '0 24px',
    },
    Card: {
      borderRadiusLG: 12,
    },
    Button: {
      borderRadius: 6,
    },
    Input: {
      borderRadius: 6,
    },
  },
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <ConfigProvider
        locale={zhCN}
        theme={theme}
        componentSize="middle"
      >
        <App>
          <RouterProvider router={router} />
        </App>
      </ConfigProvider>
    </ErrorBoundary>
  </React.StrictMode>
);