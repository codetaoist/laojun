import React from 'react';
import { Layout } from 'antd';
import { ErrorBoundary } from '../ErrorBoundary';
import { LoadingSpinner } from '../Loading';

const { Header, Content, Footer, Sider } = Layout;

// 基础布局组件属性
export interface BaseLayoutProps {
  children: React.ReactNode;
  header?: React.ReactNode;
  sidebar?: React.ReactNode;
  footer?: React.ReactNode;
  loading?: boolean;
  className?: string;
  style?: React.CSSProperties;
  siderCollapsed?: boolean;
  siderCollapsible?: boolean;
  onSiderCollapse?: (collapsed: boolean) => void;
  siderWidth?: number;
  siderCollapsedWidth?: number;
  headerHeight?: number;
  footerHeight?: number;
}

// 基础布局组件
export const BaseLayout: React.FC<BaseLayoutProps> = ({
  children,
  header,
  sidebar,
  footer,
  loading = false,
  className,
  style,
  siderCollapsed = false,
  siderCollapsible = false,
  onSiderCollapse,
  siderWidth = 256,
  siderCollapsedWidth = 80,
  headerHeight = 64,
  footerHeight = 48,
}) => {
  if (loading) {
    return <LoadingSpinner size="large" tip="页面加载中..." />;
  }

  return (
    <ErrorBoundary>
      <Layout className={className} style={{ minHeight: '100vh', ...style }}>
        {/* 侧边栏 */}
        {sidebar && (
          <Sider
            collapsed={siderCollapsed}
            collapsible={siderCollapsible}
            onCollapse={onSiderCollapse}
            width={siderWidth}
            collapsedWidth={siderCollapsedWidth}
            style={{
              overflow: 'auto',
              height: '100vh',
              position: 'fixed',
              left: 0,
              top: 0,
              bottom: 0,
            }}
          >
            {sidebar}
          </Sider>
        )}

        {/* 主要内容区域 */}
        <Layout
          style={{
            marginLeft: sidebar ? (siderCollapsed ? siderCollapsedWidth : siderWidth) : 0,
            transition: 'margin-left 0.2s',
          }}
        >
          {/* 头部 */}
          {header && (
            <Header
              style={{
                position: 'fixed',
                top: 0,
                right: 0,
                left: sidebar ? (siderCollapsed ? siderCollapsedWidth : siderWidth) : 0,
                zIndex: 1000,
                height: headerHeight,
                lineHeight: `${headerHeight}px`,
                padding: 0,
                background: '#fff',
                borderBottom: '1px solid #f0f0f0',
                transition: 'left 0.2s',
              }}
            >
              {header}
            </Header>
          )}

          {/* 内容区域 */}
          <Content
            style={{
              marginTop: header ? headerHeight : 0,
              marginBottom: footer ? footerHeight : 0,
              padding: '24px',
              background: '#f0f2f5',
              minHeight: 'calc(100vh - 64px)',
            }}
          >
            <ErrorBoundary>
              {children}
            </ErrorBoundary>
          </Content>

          {/* 底部 */}
          {footer && (
            <Footer
              style={{
                position: 'fixed',
                bottom: 0,
                right: 0,
                left: sidebar ? (siderCollapsed ? siderCollapsedWidth : siderWidth) : 0,
                zIndex: 1000,
                height: footerHeight,
                lineHeight: `${footerHeight}px`,
                padding: '0 24px',
                background: '#fff',
                borderTop: '1px solid #f0f0f0',
                textAlign: 'center',
                transition: 'left 0.2s',
              }}
            >
              {footer}
            </Footer>
          )}
        </Layout>
      </Layout>
    </ErrorBoundary>
  );
};

// 简单布局组件（无侧边栏）
export interface SimpleLayoutProps {
  children: React.ReactNode;
  header?: React.ReactNode;
  footer?: React.ReactNode;
  loading?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

export const SimpleLayout: React.FC<SimpleLayoutProps> = ({
  children,
  header,
  footer,
  loading = false,
  className,
  style,
}) => {
  if (loading) {
    return <LoadingSpinner size="large" tip="页面加载中..." />;
  }

  return (
    <ErrorBoundary>
      <Layout className={className} style={{ minHeight: '100vh', ...style }}>
        {/* 头部 */}
        {header && (
          <Header
            style={{
              position: 'fixed',
              top: 0,
              left: 0,
              right: 0,
              zIndex: 1000,
              height: 64,
              lineHeight: '64px',
              padding: 0,
              background: '#fff',
              borderBottom: '1px solid #f0f0f0',
            }}
          >
            {header}
          </Header>
        )}

        {/* 内容区域 */}
        <Content
          style={{
            marginTop: header ? 64 : 0,
            marginBottom: footer ? 48 : 0,
            padding: '24px',
            background: '#f0f2f5',
            minHeight: 'calc(100vh - 64px)',
          }}
        >
          <ErrorBoundary>
            {children}
          </ErrorBoundary>
        </Content>

        {/* 底部 */}
        {footer && (
          <Footer
            style={{
              position: 'fixed',
              bottom: 0,
              left: 0,
              right: 0,
              zIndex: 1000,
              height: 48,
              lineHeight: '48px',
              padding: '0 24px',
              background: '#fff',
              borderTop: '1px solid #f0f0f0',
              textAlign: 'center',
            }}
          >
            {footer}
          </Footer>
        )}
      </Layout>
    </ErrorBoundary>
  );
};

// 居中布局组件
export interface CenteredLayoutProps {
  children: React.ReactNode;
  maxWidth?: number | string;
  loading?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

export const CenteredLayout: React.FC<CenteredLayoutProps> = ({
  children,
  maxWidth = 1200,
  loading = false,
  className,
  style,
}) => {
  if (loading) {
    return <LoadingSpinner size="large" tip="页面加载中..." />;
  }

  return (
    <ErrorBoundary>
      <div
        className={className}
        style={{
          minHeight: '100vh',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'flex-start',
          padding: '24px',
          background: '#f0f2f5',
          ...style,
        }}
      >
        <div
          style={{
            width: '100%',
            maxWidth,
            background: '#fff',
            borderRadius: 8,
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
            overflow: 'hidden',
          }}
        >
          {children}
        </div>
      </div>
    </ErrorBoundary>
  );
};